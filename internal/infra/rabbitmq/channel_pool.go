package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cp.initialize(ctx); err != nil {
		return nil, fmt.Errorf("initialize channel pool: %w", err)
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

// Close closes the pool and all available channels.
// The context controls the timeout for closing channels.
// Channels that are currently in use will not be closed - they should be closed by their owners, for example by calling Put().
func (p *ChannelPool) Close(ctx context.Context) error {
	if p.IsClosed() {
		return nil // Already closed
	}

	// Capture stats before closing
	beforeStats := p.Stats()
	start := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Mark as closed first to prevent new Get() calls
	p.closed = true
	// Close all available channels (non-blocking)
	closedCount := p.closeChannels(ctx)
	// Close the pool channel
	close(p.pool)

	duration := time.Since(start)

	if p.log != nil {
		select {
		case <-ctx.Done():
			// Timeout occurred
			p.log.Warn().
				Int("total_channels", beforeStats.Total).
				Int("channels_closed", closedCount).
				Int("channels_remaining", beforeStats.Available-closedCount).
				Int("channels_in_use", beforeStats.InUse).
				Dur("timeout", duration).
				Msg("Channel pool closed with timeout")
		default:
			// Successfully closed
			p.log.Info().
				Int("total_channels", beforeStats.Total).
				Int("channels_closed", closedCount).
				Int("channels_in_use", beforeStats.InUse).
				Dur("duration", duration).
				Msg("Channel pool closed")
		}
	}

	// Return error if timeout occurred
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout closing channel pool: %w", ctx.Err())
	default:
		return nil
	}
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

// IsClosed returns true if the pool is closed.
func (p *ChannelPool) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.closed
}

// initialize populates the pool with channels.
func (p *ChannelPool) initialize(ctx context.Context) error {
	if p.IsClosed() {
		return ErrChannelPoolClosed
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.size; i++ {
		ch, err := p.createChannel()
		if err != nil {
			// Close already created channels
			p.closeChannels(ctx)
			return fmt.Errorf("create channel %d: %w", i, err)
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
		return nil, fmt.Errorf("get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	return ch, nil
}

// closeChannels closes all available channels in the pool.
// Returns the number of channels closed.
func (p *ChannelPool) closeChannels(ctx context.Context) int {
	closedCount := 0
	for {
		select {
		case <-ctx.Done():
			// Timeout occurred, return what we've closed so far
			return closedCount
		case ch, ok := <-p.pool:
			if !ok {
				// Pool channel is closed
				return closedCount
			}
			if ch != nil {
				ch.Close()
				closedCount++
			}
		default:
			// No more channels available in pool
			return closedCount
		}
	}
}

// ErrChannelPoolClosed is returned when the pool is closed.
var ErrChannelPoolClosed = errors.New("channel pool is closed")
