package config

import "recipes-desk/pkg/config"

type Config struct{}

func Load() (*Config, error) {
	cfg, err := config.LoadYAML[Config]("./internal/modules/recipes/config/recipes.yaml")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
