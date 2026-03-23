package config

import (
	"recipes-desk/pkg/config"
)

const (
	DefaultMaxLimit = 100
	DefaultLimit    = 20
)

type Config struct {
	Pagination Pagination `yaml:"pagination"`
}

type Pagination struct {
	MaxLimit     int `yaml:"max-limit"`
	DefaultLimit int `yaml:"default-limit"`
}

func Load() (*Config, error) {
	env := config.NewEnv()
	env.WithEnvVars()
	env.Optional("RECIPES_CONFIG_PATH").WithRules(config.IsPath(true))
	if err := env.Load(); err != nil {
		return nil, err
	}

	cfg, err := config.LoadYAML[Config](
		env.GetOrDefault("RECIPES_CONFIG_PATH", "./internal/modules/recipes/config/recipes.yaml"),
	)
	if err != nil {
		return nil, err
	}

	cfg.setDefaults()

	return cfg, nil
}

func (c *Config) setDefaults() {
	if c.Pagination.MaxLimit <= 0 {
		c.Pagination.MaxLimit = DefaultMaxLimit
	}
	if c.Pagination.DefaultLimit <= 0 {
		c.Pagination.DefaultLimit = DefaultLimit
	}
}
