package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
