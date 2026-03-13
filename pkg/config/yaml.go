package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrReadFile  = errors.New("read config file")
	ErrUnmarshal = errors.New("unmarshal yaml data")
)

// LoadYAML loads a YAML configuration file.
// It returns an error if the file does not exist or if the YAML unmarshaling fails.
func LoadYAML[T any](filename string) (*T, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadFile, err)
	}

	var config T
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
	}

	return &config, nil
}
