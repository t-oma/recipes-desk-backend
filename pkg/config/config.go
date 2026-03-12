package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func LoadYAML[T any](filename string) (*T, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config T
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml data: %w", err)
	}

	return &config, nil
}

func LoadEnv(filename string) (map[string]string, error) {
	env, err := godotenv.Read(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	return env, nil
}

func RequireEnvKeys(env map[string]string, keys ...string) error {
	for _, field := range keys {
		if _, ok := env[field]; !ok {
			return fmt.Errorf("env variable %s is required but was not found", field)
		}
	}
	return nil
}
