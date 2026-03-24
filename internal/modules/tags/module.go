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
	handler  *httphandler.Handler
	consumer *rabbitmq.Consumer
}

// NewModule creates a new tags module.
func NewModule(db *mongo.Database, conn *rabbitmq.Connection, log *zerolog.Logger) *Module {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load tags config")
	}

	tagRepo := mongorepo.NewTags(db)
	if err = tagRepo.InitIndexes(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tag indexes")
	}

	tagService := application.NewService(
		tagRepo,
		mongorepo.ObjectIDGenerator{},
		log,
		cfg.Pagination.MaxLimit,
		cfg.Pagination.DefaultLimit,
	)
	tagHandler := httphandler.NewHandler(tagService, log)

	topology := rabbitmq.NewTopology(conn, log)
	consumer, err := rabbitmq.NewConsumer(conn, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create RabbitMQ consumer")
	}

	eventHandler := application.NewEventHandler(tagService, log)
	setupConsumers(topology, consumer, eventHandler, log)

	return &Module{
		handler:  tagHandler,
		consumer: consumer,
	}
}

// Shutdown gracefully stops the module's consumers.
func (m *Module) Shutdown() error {
	return m.consumer.Close()
}

// RegisterRoutes registers all tag routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	public.GET("/tags", m.handler.Search)
	public.GET("/tags/:id", m.handler.GetByID)
	protected.POST("/tags", m.handler.Create)
}

func setupConsumers(
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	eventHandler *application.EventHandler,
	log *zerolog.Logger,
) {
	go func() {
		err := consumeRecipeCreated(
			topology,
			consumer,
			eventHandler.HandleRecipeCreated,
		)
		if err != nil {
			log.Error().Err(err).Msg("Recipe created consumer failed")
		}
	}()
	go func() {
		err := consumeRecipeDeleted(
			topology,
			consumer,
			eventHandler.HandleRecipeDeleted,
		)
		if err != nil {
			log.Error().Err(err).Msg("Recipe deleted consumer failed")
		}
	}()
}

func consumeRecipeCreated(
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	handler messagebus.Handler,
) error {
	exchangeCfg := rabbitmq.NewExchangeConfig("recipes.topic", rabbitmq.ExchangeTypeTopic)
	queueCfg := rabbitmq.NewQueueConfig("tags.recipes-created", rabbitmq.QueueTypeClassic)
	bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, "recipes.created")

	if err := topology.SetupTopologyWithDLQ(exchangeCfg, queueCfg, bindingCfg); err != nil {
		return err
	}

	return consumer.Consume(queueCfg.Name, handler)
}

func consumeRecipeDeleted(
	topology *rabbitmq.Topology,
	consumer messagebus.Consumer,
	handler messagebus.Handler,
) error {
	exchangeCfg := rabbitmq.NewExchangeConfig("recipes.topic", rabbitmq.ExchangeTypeTopic)
	queueCfg := rabbitmq.NewQueueConfig("tags.recipes-deleted", rabbitmq.QueueTypeClassic)
	bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, "recipes.deleted")

	if err := topology.SetupTopologyWithDLQ(exchangeCfg, queueCfg, bindingCfg); err != nil {
		return err
	}

	return consumer.Consume(queueCfg.Name, handler)
}
