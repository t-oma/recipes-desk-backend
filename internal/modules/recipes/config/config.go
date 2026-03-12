package config

import "recipes-desk/pkg/config"

type Config struct{}

func Load() (*Config, error) {
	env := config.NewEnv()
	env.WithEnvVars()
	if err := env.Load(); err != nil {
		return nil, err
	}

	cfg, err := config.LoadYAML[Config](
		env.GetOrDefault("RECIPES_CONFIG_PATH", "./internal/modules/recipes/config/recipes.yaml"),
	)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
