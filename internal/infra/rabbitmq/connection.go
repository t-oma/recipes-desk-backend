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

type ConnectionManager struct {
	config    ConnectionManagerConfig
	conn      *amqp.Connection
	mu        sync.RWMutex
	log       *zerolog.Logger
	done      chan struct{}
	isRunning bool
}

func NewConnectionManager(config ConnectionManagerConfig, log *zerolog.Logger) *ConnectionManager {
	return &ConnectionManager{
		config:    config,
		conn:      nil,
		mu:        sync.RWMutex{},
		log:       log,
		done:      make(chan struct{}),
		isRunning: false,
	}
}

func (cm *ConnectionManager) Connection() (*amqp.Connection, error) {
	if cm.IsConnected() {
		return cm.conn, nil
	}

	return nil, errors.New("connection not available")
}

func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.conn != nil && !cm.conn.IsClosed()
}

func (cm *ConnectionManager) Start(ctx context.Context) error {
	if err := cm.connect(); err != nil {
		return err
	}

	cm.mu.Lock()
	cm.isRunning = true
	cm.mu.Unlock()

	go cm.monitor(ctx)
	return nil
}

func (cm *ConnectionManager) Stop() error {
	cm.mu.Lock()
	if !cm.isRunning {
		cm.mu.Unlock()
		return nil
	}
	cm.isRunning = false
	close(cm.done)
	cm.mu.Unlock()

	return cm.Close()
}

func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.conn == nil {
		return nil
	}

	if err := cm.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	cm.conn = nil

	if cm.log != nil {
		cm.log.Info().Msg("RabbitMQ connection closed")
	}
	return nil
}

func (cm *ConnectionManager) connect() error {
	if cm.IsConnected() {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	conn, err := amqp.Dial(cm.config.URI())
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	cm.conn = conn

	if cm.log != nil {
		cm.log.Info().
			Str("host", cm.config.Host).
			Int("port", cm.config.Port).
			Msg("Connected to RabbitMQ")
	}

	return nil
}

func (cm *ConnectionManager) reconnect(ctx context.Context) {
	backoff := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
	}
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.done:
			return
		default:
		}

		cm.mu.Lock()
		if cm.IsConnected() {
			cm.mu.Unlock()
			return
		}
		cm.conn = nil
		cm.mu.Unlock()

		if err := cm.connect(); err != nil {
			delay := backoff[min(attempt, len(backoff)-1)]
			attempt++

			if cm.log != nil {
				cm.log.Error().
					Err(err).
					Int("attempt", attempt).
					Dur("delay", delay).
					Msg("Failed to reconnect, retrying...")
			}

			select {
			case <-ctx.Done():
				return
			case <-cm.done:
				return
			case <-time.After(delay):
				continue
			}
		}

		if cm.log != nil {
			cm.log.Info().Int("attempt", attempt+1).Msg("Reconnected to RabbitMQ")
		}
		return
	}
}

func (cm *ConnectionManager) monitor(ctx context.Context) {
	for {
		if !cm.IsConnected() {
			cm.reconnect(ctx)
			continue
		}

		cm.mu.RLock()
		closeChan := cm.conn.NotifyClose(make(chan *amqp.Error, 1))
		cm.mu.RUnlock()

		select {
		case <-ctx.Done():
			return
		case <-cm.done:
			return
		case err := <-closeChan:
			if cm.log != nil {
				cm.log.Error().Err(err).Msg("RabbitMQ connection closed, reconnecting...")
			}
			cm.reconnect(ctx)
		}
	}
}
