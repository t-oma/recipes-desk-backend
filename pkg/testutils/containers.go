package testutils

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	rabbitmqtc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func SetupMongoContainer(t *testing.T, dbName string) (*mongo.Database, func()) {
	t.Helper()

	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongo:8",
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

// MongoDBTestContainer holds the MongoDB container, client, and database.
type MongoDBTestContainer struct {
	Container testcontainers.Container
	Client    *mongo.Client
	URI       string
	Host      string
	Port      int
}

// SetupMongoDBContainer starts a MongoDB container and returns a connected database.
func SetupMongoDBContainer(t *testing.T) (*MongoDBTestContainer, func()) {
	t.Helper()

	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:8",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("27017/tcp")),
	)
	require.NoError(t, err)

	uri, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "27017")
	require.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	cleanup := func() {
		if err = client.Disconnect(ctx); err != nil {
			t.Logf("Failed to disconnect MongoDB client: %v", err)
		}
		if err = container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MongoDB container: %v", err)
		}
	}

	return &MongoDBTestContainer{
		Container: container,
		Client:    client,
		URI:       uri,
		Host:      host,
		Port:      port.Int(),
	}, cleanup
}

// RabbitMQTestContainer holds the RabbitMQ container and connection info.
type RabbitMQTestContainer struct {
	Container *rabbitmqtc.RabbitMQContainer
	Host      string
	Port      int
}

// SetupRabbitMQContainer starts a RabbitMQ container for testing.
func SetupRabbitMQContainer(t *testing.T) (*RabbitMQTestContainer, func()) {
	t.Helper()

	ctx := context.Background()

	// Start RabbitMQ container with management UI
	rmqContainer, err := rabbitmqtc.Run(
		ctx,
		"rabbitmq:4.2-alpine",
		rabbitmqtc.WithAdminUsername("guest"),
		rabbitmqtc.WithAdminPassword("guest"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5672/tcp"),
		),
	)
	require.NoError(t, err)

	// Get connection info
	// host, err := rmqContainer.Host(ctx)
	// require.NoError(t, err)
	// Force IPv4 to avoid IPv6 issues in CI
	host := "127.0.0.1"

	port, err := rmqContainer.MappedPort(ctx, "5672")
	require.NoError(t, err)

	cleanup := func() {
		if err = rmqContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate RabbitMQ container: %v", err)
		}
	}

	return &RabbitMQTestContainer{
		Container: rmqContainer,
		Host:      host,
		Port:      port.Int(),
	}, cleanup
}

// NewTestLogger creates a logger for tests.
func NewTestLogger(t *testing.T, log ...uint8) *zerolog.Logger {
	t.Helper()
	var logger zerolog.Logger
	if len(log) > 0 && log[0] != 0 {
		logger = zerolog.New(zerolog.NewConsoleWriter())
	} else {
		logger = zerolog.Nop()
	}
	return &logger
}
