package config

import (
	"time"

	"recipes-desk/pkg/config"
)

type (
	Config struct {
		App     App    `yaml:"app"`
		Server  Server `yaml:"server"`
		MongoDB MongoDB
	}
	App struct {
		Environment string `yaml:"environment"`
	}
	Server struct {
		Port     string   `yaml:"port"`
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
	env, err := config.LoadEnv(".env")
	if err != nil {
		return nil, err
	}
	err = config.RequireEnvKeys(env, "MONGO_URI", "MONGO_DATABASE")
	if err != nil {
		return nil, err
	}

	cfg.MongoDB.URI = env["MONGO_URI"]
	cfg.MongoDB.Database = env["MONGO_DATABASE"]

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
