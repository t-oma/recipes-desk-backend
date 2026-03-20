package config

import (
	"time"

	"recipes-desk/pkg/config"
)

const (
	DefaultServerTimeoutWrite = 10 * time.Second
	DefaultServerTimeoutRead  = 10 * time.Second
)

type (
	Config struct {
		App      App      `yaml:"app"`
		Server   Server   `yaml:"server"`
		MongoDB  MongoDB  // loaded from .env or environment variables
		RabbitMQ RabbitMQ // loaded from .env or environment variables
	}
	App struct {
		Environment string `yaml:"environment"`
	}
	Server struct {
		Port     int      `yaml:"port"`
		Timeouts Timeouts `yaml:"timeouts"`
	}
	Timeouts struct {
		Write time.Duration `yaml:"write"`
		Read  time.Duration `yaml:"read"`
	}
	MongoDB struct {
		URI      string
		Database string
	}
	RabbitMQ struct {
		Host     string
		Port     int
		User     string
		Password string
		VHost    string
	}
)

func Load() (*Config, error) {
	cfg, err := config.LoadYAML[Config]("./configs/config.yaml")
	if err != nil {
		return nil, err
	}

	env := config.NewEnv()
	env.Required("MONGO_URI")
	env.Required("MONGO_DATABASE")
	env.Required("RABBITMQ_HOST")
	env.Required("RABBITMQ_PORT")
	env.Required("RABBITMQ_USER")
	env.Required("RABBITMQ_PASSWORD")
	env.Required("RABBITMQ_VHOST")
	env.WithEnvVars()
	env.AddEnvFiles(".env")
	if err = env.Load(); err != nil {
		return nil, err
	}

	cfg.MongoDB.URI = env.Get("MONGO_URI")
	cfg.MongoDB.Database = env.Get("MONGO_DATABASE")

	cfg.RabbitMQ.Host = env.Get("RABBITMQ_HOST")
	cfg.RabbitMQ.Port = env.GetInt("RABBITMQ_PORT")
	cfg.RabbitMQ.User = env.Get("RABBITMQ_USER")
	cfg.RabbitMQ.Password = env.Get("RABBITMQ_PASSWORD")
	cfg.RabbitMQ.VHost = env.Get("RABBITMQ_VHOST")

	if cfg.Server.Timeouts.Write <= 0 {
		cfg.Server.Timeouts.Write = DefaultServerTimeoutWrite
	}
	if cfg.Server.Timeouts.Read <= 0 {
		cfg.Server.Timeouts.Read = DefaultServerTimeoutRead
	}

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
