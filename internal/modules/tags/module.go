package tags

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/infra/messagebus"
	"recipes-desk/internal/infra/messagebus/rabbitmq"
	httphandler "recipes-desk/internal/modules/tags/adapter/in/http"
	"recipes-desk/internal/modules/tags/adapter/out/mongorepo"
	"recipes-desk/internal/modules/tags/application"
	"recipes-desk/internal/modules/tags/config"
)

// Module represents the tags module.
type Module struct {
	handler *httphandler.Handler
}

// NewModule creates a new tags module.
func NewModule(
	db *mongo.Database,
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	log *zerolog.Logger,
) *Module {
	config, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load tags config")
	}

	tagRepo := mongorepo.NewTags(db)
	err = tagRepo.InitIndexes(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tag indexes")
	}
	tagService := application.NewService(
		tagRepo,
		mongorepo.ObjectIDGenerator{},
		log,
		config.Pagination.MaxLimit,
		config.Pagination.DefaultLimit,
	)
	tagHandler := httphandler.NewHandler(tagService, log)

	eventHandler := application.NewEventHandler(tagService, log)

	setupConsumers(topology, consumer, eventHandler)

	return &Module{
		handler: tagHandler,
	}
}

// RegisterRoutes registers all tag routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	// Public routes
	public.GET("/tags", m.handler.Search)
	public.GET("/tags/:id", m.handler.GetByID)

	// Protected routes (admin only in future)
	protected.POST("/tags", m.handler.Create)
}

func setupConsumers(
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	handlers *application.EventHandler,
) {
	go func() {
		if err := consumeRecipeCreated(
			topology,
			consumer,
			handlers.HandleRecipeCreated,
		); err != nil {
			panic(err)
		}
	}()
	go func() {
		if err := consumeRecipeDeleted(
			topology,
			consumer,
			handlers.HandleRecipeDeleted,
		); err != nil {
			panic(err)
		}
	}()
}

// consumeRecipeCreated sets up a topology and consumer for recipes.created events.
func consumeRecipeCreated(
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	handler messagebus.Handler,
) error {
	exchangeCfg := rabbitmq.NewExchangeConfig("recipes.topic", rabbitmq.ExchangeTypeTopic)
	queueCfg := rabbitmq.NewQueueConfig("tags.recipes-created", rabbitmq.QueueTypeClassic)
	bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, "recipes.created")
	if err := topology.SetupTopologyWithDLQ(
		exchangeCfg,
		queueCfg,
		bindingCfg,
	); err != nil {
		return err
	}

	return consumer.Consume(
		context.Background(),
		queueCfg.Name,
		handler,
	)
}

// consumeRecipeDeleted sets up a topology and consumer for recipes.deleted events.
func consumeRecipeDeleted(
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	handler messagebus.Handler,
) error {
	exchangeCfg := rabbitmq.NewExchangeConfig("recipes.topic", rabbitmq.ExchangeTypeTopic)
	queueCfg := rabbitmq.NewQueueConfig("tags.recipes-deleted", rabbitmq.QueueTypeClassic)
	bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, "recipes.deleted")
	if err := topology.SetupTopologyWithDLQ(
		exchangeCfg,
		queueCfg,
		bindingCfg,
	); err != nil {
		return err
	}

	return consumer.Consume(
		context.Background(),
		queueCfg.Name,
		handler,
	)
}
