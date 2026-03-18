//go:build integration
// +build integration

package rabbitmq_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/rabbitmq"
	"recipes-desk/pkg/testutils"
)

func TestIntegration_ChannelPool_Initialize(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	connManager := setupRabbitMQFromContainer(t, container)
	defer connManager.Stop()

	log := zerolog.New(zerolog.NewConsoleWriter())

	t.Run("initialize with valid size", func(t *testing.T) {
		pool, err := rabbitmq.NewChannelPool(connManager, 5, &log)
		require.NoError(t, err)
		defer pool.Close(context.Background())

		stats := pool.Stats()
		assert.Equal(t, 5, stats.Available)
		assert.Equal(t, 0, stats.InUse)
		assert.Equal(t, 5, stats.Total)
	})

	t.Run("initialize with zero size uses default", func(t *testing.T) {
		pool, err := rabbitmq.NewChannelPool(connManager, 0, &log)
		require.NoError(t, err)
		defer pool.Close(context.Background())

		stats := pool.Stats()
		assert.Equal(t, rabbitmq.ChannelPoolMinSize, stats.Available)
	})
}

func TestIntegration_ChannelPool_GetAndPut(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	connManager := setupRabbitMQFromContainer(t, container)
	defer connManager.Stop()

	log := zerolog.New(zerolog.NewConsoleWriter())
	poolSize := 3
	pool, err := rabbitmq.NewChannelPool(connManager, poolSize, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	t.Run("get channel from pool", func(t *testing.T) {
		ch, err := pool.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, ch)
		assert.False(t, ch.IsClosed())

		stats := pool.Stats()
		assert.Equal(t, 1, stats.InUse)
		assert.Equal(t, 2, stats.Available)

		// Return channel
		pool.Put(ch)

		stats = pool.Stats()
		assert.Equal(t, poolSize, stats.Available)
		assert.Equal(t, 0, stats.InUse)
	})

	t.Run("get multiple channels", func(t *testing.T) {
		var channels []*amqp.Channel

		// Get all channels
		for i := 0; i < poolSize; i++ {
			ch, err := pool.Get(context.Background())
			require.NoError(t, err)
			channels = append(channels, ch)
		}

		stats := pool.Stats()
		assert.Equal(t, 0, stats.Available)
		assert.Equal(t, poolSize, stats.InUse)

		// Return all channels
		for _, ch := range channels {
			pool.Put(ch)
		}

		stats = pool.Stats()
		assert.Equal(t, poolSize, stats.Available)
		assert.Equal(t, 0, stats.InUse)
	})

	t.Run("get with timeout when pool empty", func(t *testing.T) {
		// Get all channels
		var channels []*amqp.Channel
		for i := 0; i < 3; i++ {
			ch, err := pool.Get(context.Background())
			require.NoError(t, err)
			channels = append(channels, ch)
		}

		// Try to get one more - should timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := pool.Get(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// Return channels
		for _, ch := range channels {
			pool.Put(ch)
		}
	})
}

func TestIntegration_ChannelPool_ConcurrentAccess(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	connManager := setupRabbitMQFromContainer(t, container)
	defer connManager.Stop()

	log := zerolog.New(zerolog.NewConsoleWriter())
	pool, err := rabbitmq.NewChannelPool(connManager, 10, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	t.Run("concurrent get and put", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100
		goroutines := 10

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < iterations; j++ {
					ch, err := pool.Get(context.Background())
					require.NoError(t, err)

					// Simulate some work
					time.Sleep(time.Microsecond)

					pool.Put(ch)
				}
			}()
		}

		wg.Wait()

		// All channels should be returned
		stats := pool.Stats()
		assert.Equal(t, 10, stats.Available)
		assert.Equal(t, 0, stats.InUse)
	})
}

func TestIntegration_ChannelPool_ClosedChannel(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	connManager := setupRabbitMQFromContainer(t, container)
	defer connManager.Stop()

	log := zerolog.New(zerolog.NewConsoleWriter())
	pool, err := rabbitmq.NewChannelPool(connManager, 3, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	t.Run("get handles closed channel", func(t *testing.T) {
		// Get channel and close it manually
		ch, err := pool.Get(context.Background())
		require.NoError(t, err)

		ch.Close()
		assert.True(t, ch.IsClosed())

		// Put closed channel back (should be dropped)
		pool.Put(ch)

		// Get again - should return new channel
		newCh, err := pool.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, newCh)
		assert.False(t, newCh.IsClosed())

		pool.Put(newCh)
	})
}

func TestIntegration_ChannelPool_Close(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	connManager := setupRabbitMQFromContainer(t, container)
	defer connManager.Stop()

	log := zerolog.New(zerolog.NewConsoleWriter())
	pool, err := rabbitmq.NewChannelPool(connManager, 3, &log)
	require.NoError(t, err)

	t.Run("close pool", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := pool.Close(ctx)
		require.NoError(t, err)

		// Try to get channel from closed pool
		_, err = pool.Get(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, rabbitmq.ErrChannelPoolClosed)
	})

	t.Run("double close is safe", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := pool.Close(ctx)
		require.NoError(t, err) // Should not error
	})
}

func TestIntegration_ChannelPool_Stats(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	connManager := setupRabbitMQFromContainer(t, container)
	defer connManager.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	pool, err := rabbitmq.NewChannelPool(connManager, 5, &log)
	require.NoError(t, err)
	defer pool.Close(ctx)

	t.Run("stats reflect pool state", func(t *testing.T) {
		// Initial state
		stats := pool.Stats()
		assert.Equal(t, 5, stats.Total)
		assert.Equal(t, 5, stats.Available)
		assert.Equal(t, 0, stats.InUse)

		// Get 2 channels
		ch1, err := pool.Get(context.Background())
		require.NoError(t, err)
		ch2, err := pool.Get(context.Background())
		require.NoError(t, err)

		stats = pool.Stats()
		assert.Equal(t, 3, stats.Available)
		assert.Equal(t, 2, stats.InUse)

		// Return channels
		pool.Put(ch1)
		pool.Put(ch2)

		stats = pool.Stats()
		assert.Equal(t, 5, stats.Available)
		assert.Equal(t, 0, stats.InUse)
	})
}
