package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ChannelPoolMinSize = 10

// PoolStats represents the statistics of the channel pool.
type PoolStats struct {
	Available int
	InUse     int
	Total     int
}

// ChannelPool manages a pool of AMQP channels.
type ChannelPool struct {
	connManager *ConnectionManager
	pool        chan *amqp.Channel
	size        int
	mu          sync.RWMutex
	closed      bool
	log         *zerolog.Logger
}

// NewChannelPool creates a new channel pool.
func NewChannelPool(
	connManager *ConnectionManager,
	size int,
	log *zerolog.Logger,
) (*ChannelPool, error) {
	if size <= 0 {
		size = ChannelPoolMinSize
	}

	cp := &ChannelPool{
		connManager: connManager,
		pool:        make(chan *amqp.Channel, size),
		size:        size,
		log:         log,
		closed:      false,
		mu:          sync.RWMutex{},
	}

	if err := cp.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize channel pool: %w", err)
	}

	return cp, nil
}

// Get retrieves a channel from the pool.
func (p *ChannelPool) Get(ctx context.Context) (*amqp.Channel, error) {
	if p.IsClosed() {
		return nil, ErrChannelPoolClosed
	}

	select {
	case ch := <-p.pool:
		if ch.IsClosed() {
			return p.createChannel()
		}
		return ch, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for channel: %w", ctx.Err())
	}
}

// Put returns a channel to the pool.
func (p *ChannelPool) Put(ch *amqp.Channel) {
	if ch == nil || ch.IsClosed() {
		return
	}

	if p.IsClosed() {
		ch.Close()
		return
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	select {
	case p.pool <- ch:
	default:
		// Pool is full, close the channel
		ch.Close()
	}
}

// Close closes the pool and all channels.
func (p *ChannelPool) Close() error {
	if p.IsClosed() {
		return nil // Already closed
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.closePool()
	p.closed = true
	close(p.pool)

	if p.log != nil {
		p.log.Info().Msg("Channel pool closed")
	}

	return nil
}

// Stats returns the current pool statistics.
func (p *ChannelPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	available := len(p.pool)

	return PoolStats{
		Available: available,
		InUse:     p.size - available,
		Total:     p.size,
	}
}

func (p *ChannelPool) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.closed
}

// initialize populates the pool with channels.
func (p *ChannelPool) initialize() error {
	if p.IsClosed() {
		return ErrChannelPoolClosed
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.size; i++ {
		ch, err := p.createChannel()
		if err != nil {
			// Close already created channels
			p.closePool()
			return fmt.Errorf("failed to create channel %d: %w", i, err)
		}
		p.pool <- ch
	}

	if p.log != nil {
		p.log.Info().
			Int("size", p.size).
			Msg("Channel pool initialized")
	}

	return nil
}

// createChannel creates a new channel from the connection.
func (p *ChannelPool) createChannel() (*amqp.Channel, error) {
	conn, err := p.connManager.Connection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return ch, nil
}

// closePool closes all channels in the pool.
func (p *ChannelPool) closePool() {
	for {
		select {
		case ch := <-p.pool:
			if ch != nil {
				ch.Close()
			}
		default:
			return
		}
	}
}

var ErrChannelPoolClosed = errors.New("channel pool is closed")
