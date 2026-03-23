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
		App     App     `yaml:"app"`
		Server  Server  `yaml:"server"`
		MongoDB MongoDB // loaded from .env or environment variables
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
)

func Load() (*Config, error) {
	cfg, err := config.LoadYAML[Config]("./configs/config.yaml")
	if err != nil {
		return nil, err
	}

	env := config.NewEnv()
	env.Required("MONGO_URI")
	env.Required("MONGO_DATABASE")
	env.WithEnvVars()
	env.AddEnvFiles(".env")
	if err = env.Load(); err != nil {
		return nil, err
	}

	cfg.MongoDB.URI = env.Get("MONGO_URI")
	cfg.MongoDB.Database = env.Get("MONGO_DATABASE")

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
