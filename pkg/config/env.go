package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var ErrValidation = errors.New("validation error")

// EnvKeyRule is a function that validates an env key value.
type EnvKeyRule func(string) bool

type EnvRules map[string][]EnvKeyRule

// Env represents a configuration environment.
type Env struct {
	env         map[string]string
	required    []string
	withEnvVars bool
	envFiles    []string
	rules       EnvRules
}

// NewEnv creates a new Env instance.
func NewEnv() *Env {
	return &Env{
		env:         make(map[string]string),
		required:    make([]string, 0),
		withEnvVars: false,
		envFiles:    make([]string, 0),
		rules:       make(EnvRules),
	}
}

// Required adds a list of environment keys that must be set.
func (e *Env) Required(keys ...string) {
	e.required = append(e.required, keys...)
}

// WithEnvVars enables the loading of environment variables.
func (e *Env) WithEnvVars() {
	e.withEnvVars = true
}

// KeyRules adds a validation rules for an environment key.
func (e *Env) KeyRules(key string, rules ...EnvKeyRule) {
	e.rules[key] = rules
}

// AddEnvFiles adds environment files to the list of files to load.
func (e *Env) AddEnvFiles(filenames ...string) {
	e.envFiles = append(e.envFiles, filenames...)
}

// Load loads environment variables and env files added with AddEnvFiles.
func (e *Env) Load() error {
	if e.withEnvVars {
		e.loadEnvVars(e.required...)
	}

	if len(e.envFiles) != 0 {
		err := e.loadEnvFiles()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return e.validate()
			}
			return err
		}
	}

	return e.validate()
}

// ReadFile reads an environment file and returns a map of key-value pairs.
func (e *Env) ReadFile(filename string) (map[string]string, error) {
	env, err := godotenv.Read(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	return env, nil
}

// Get returns the value of an environment variable.
func (e *Env) Get(key string) string {
	return e.env[key]
}

// GetOrDefault returns the value of an environment variable or a default value if the variable is not set.
func (e *Env) GetOrDefault(key string, defaultValue string) string {
	value, ok := e.env[key]
	if !ok {
		return defaultValue
	}
	return value
}

// Set sets the value of an environment variable.
func (e *Env) Set(key string, value string) {
	e.env[key] = value
}

// validate checks if all required environment keys are set and if their values are valid according to the rules.
func (e *Env) validate() error {
	for _, field := range e.required {
		envVar, ok := e.env[field]
		if !ok || envVar == "" {
			return fmt.Errorf("%w: env key %s is required but was not found", ErrValidation, field)
		}
	}

	for key, rules := range e.rules {
		if err := e.runKeyRules(key, rules); err != nil {
			return err
		}
	}
	return nil
}

func (e *Env) runKeyRules(key string, rules []EnvKeyRule) error {
	value, ok := e.env[key]
	if !ok || value == "" {
		return fmt.Errorf("%w: env key %s have rules but was not found", ErrValidation, key)
	}
	for _, rule := range rules {
		if !rule(value) {
			return fmt.Errorf("%w: env key %s is invalid", ErrValidation, key)
		}
	}

	return nil
}

// append appends a map of key-value pairs to the environment.
func (e *Env) append(env map[string]string) {
	for key, value := range env {
		if _, ok := e.env[key]; ok {
			// Prevent overwriting existing values got from env vars
			continue
		}
		e.env[key] = value
	}
}

// loadEnvFiles loads environment files from the list of files to load.
func (e *Env) loadEnvFiles() error {
	for _, filename := range e.envFiles {
		env, err := e.ReadFile(filename)
		if err != nil {
			return err
		}
		e.append(env)
	}
	return nil
}

// loadEnvVars loads environment variables from the provided keys.
func (e *Env) loadEnvVars(keys ...string) {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			e.env[key] = value
		}
	}
}
