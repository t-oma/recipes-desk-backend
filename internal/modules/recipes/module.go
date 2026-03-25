package recipes

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/infra/messagebus/rabbitmq"
	httphandler "recipes-desk/internal/modules/recipes/adapter/in/http"
	"recipes-desk/internal/modules/recipes/adapter/out/mongorepo"
	rmqpublisher "recipes-desk/internal/modules/recipes/adapter/out/rabbitmq"
	"recipes-desk/internal/modules/recipes/application"
	"recipes-desk/internal/modules/recipes/config"
)

// Module represents the recipes module.
type Module struct {
	handler  *httphandler.Handler
	producer *rabbitmq.Producer
}

// NewModule creates a new recipes module.
func NewModule(db *mongo.Database, conn *rabbitmq.Connection, log *zerolog.Logger) *Module {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load recipes config")
	}

	recipeRepo := mongorepo.NewRecipes(db)
	idGenerator := mongorepo.ObjectIDGenerator{}

	producer, err := rabbitmq.NewProducer(conn, log, rabbitmq.DefaultProducerConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create RabbitMQ producer")
	}

	publisher := rmqpublisher.NewPublisher(producer)

	recipeService := application.NewService(
		recipeRepo,
		idGenerator,
		publisher,
		log,
		cfg.Pagination.MaxLimit,
		cfg.Pagination.DefaultLimit,
	)
	recipeHandler := httphandler.NewHandler(recipeService, log)

	return &Module{
		handler:  recipeHandler,
		producer: producer,
	}
}

// Shutdown gracefully stops the module's producer.
func (m *Module) Shutdown() error {
	return m.producer.Close()
}

// RegisterRoutes registers all recipe routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	public.GET("/recipes", m.handler.Search)
	public.GET("/recipes/:id", m.handler.GetByID)
	protected.POST("/recipes", m.handler.Create)
	protected.PUT("/recipes/:id", m.handler.Update)
	protected.DELETE("/recipes/:id", m.handler.Delete)
}
