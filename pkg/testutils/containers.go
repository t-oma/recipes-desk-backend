package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	amqp "github.com/rabbitmq/amqp091-go"
	mongodbtc "github.com/testcontainers/testcontainers-go/modules/mongodb"
	rabbitmqtc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"recipes-desk/internal/infra/rabbitmq"
)

func SetupMongoContainer(t *testing.T, dbName string) (*mongo.Database, func()) {
	t.Helper()

	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodbtc.Run(ctx, "mongo:8",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("27017/tcp")),
	)
	require.NoError(t, err)

	// Get connection string
	connStr, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	require.NoError(t, err)

	// Check connection
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	db := client.Database(dbName)

	cleanup := func() {
		if err = client.Disconnect(ctx); err != nil {
			t.Logf("Failed to disconnect from MongoDB: %v", err)
		}
		if err = mongoContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MongoDB container: %v", err)
		}
	}

	return db, cleanup
}

// SetupRabbitMQContainer starts a RabbitMQ container for testing.
func SetupRabbitMQContainer(t *testing.T) (*amqp.Connection, func()) {
	t.Helper()

	ctx := context.Background()

	// Start RabbitMQ container with management plugin
	rmqContainer, err := rabbitmqtc.Run(ctx, "rabbitmq:4.2",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5672/tcp")),
	)
	require.NoError(t, err)

	// Get connection details
	host, err := rmqContainer.Host(ctx)
	require.NoError(t, err)
	port, err := rmqContainer.MappedPort(ctx, "5672")
	require.NoError(t, err)

	// Create RabbitMQ connection
	config := rabbitmq.Config{
		Host:     host,
		Port:     port.Int(),
		User:     "guest",
		Password: "guest",
		VHost:    "/",
	}

	conn, err := rabbitmq.Connect(config)
	require.NoError(t, err)

	cleanup := func() {
		if err = conn.Close(); err != nil {
			t.Logf("Failed to close RabbitMQ connection: %v", err)
		}
		if err = rmqContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate RabbitMQ container: %v", err)
		}
	}

	return conn, cleanup
}
