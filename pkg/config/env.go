package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/joho/godotenv"
)

var ErrValidation = errors.New("validation error")

const (
	SourceVar     envKeySource = "var"
	SourceFile    envKeySource = "file"
	SourceUnknown envKeySource = "unknown"
)

type (
	envKeySource string
	// EnvKeyRule is a function that validates an env key value.
	EnvKeyRule func(string) bool
	// EnvKey represents an environment variable.
	EnvKey struct { //nolint:recvcheck // intentionally mixed pointer and value receivers
		value    string
		optional bool
		isSet    bool
		source   envKeySource
		rules    []EnvKeyRule
	}
)

// NewEnvKey creates a new EnvKey.
func NewEnvKey(optional bool) *EnvKey {
	return &EnvKey{
		value:    "",
		optional: optional,
		isSet:    false,
		source:   SourceUnknown,
		rules:    []EnvKeyRule{},
	}
}

// WithRules adds validation rules to the EnvKey.
func (e *EnvKey) WithRules(rules ...EnvKeyRule) *EnvKey {
	e.rules = rules
	return e
}

// Set sets the value of the EnvKey.
func (e *EnvKey) Set(value string, source envKeySource) {
	e.value = value
	e.isSet = true
	e.source = source
}

// isValid checks if the EnvKey is valid.
func (e EnvKey) isValid() bool {
	if e.optional && e.value == "" {
		return true
	}
	if !e.optional && e.value == "" {
		return false
	}
	for _, rule := range e.rules {
		if !rule(e.value) {
			return false
		}
	}
	return true
}

// Env represents a configuration environment.
type Env struct {
	env         map[string]*EnvKey
	withEnvVars bool
	envFiles    []string
}

// NewEnv creates a new Env instance.
func NewEnv() *Env {
	return &Env{
		env:         make(map[string]*EnvKey),
		withEnvVars: false,
		envFiles:    make([]string, 0),
	}
}

// Required returns a required EnvKey.
func (e *Env) Required(key string) *EnvKey {
	e.env[key] = NewEnvKey(false)
	return e.env[key]
}

// Optional returns an optional EnvKey.
func (e *Env) Optional(key string) *EnvKey {
	e.env[key] = NewEnvKey(true)
	return e.env[key]
}

// WithEnvVars enables the loading of environment variables.
func (e *Env) WithEnvVars() {
	e.withEnvVars = true
}

// AddEnvFiles adds environment files to the list of files to load.
func (e *Env) AddEnvFiles(filenames ...string) {
	e.envFiles = append(e.envFiles, filenames...)
}

// Load loads environment variables and env files added with AddEnvFiles.
func (e *Env) Load() error {
	if e.withEnvVars {
		e.loadEnvVars()
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
	if envKey, ok := e.env[key]; ok {
		return envKey.value
	}
	return ""
}

// GetOrDefault returns the value of an environment variable or a default value if the variable is not set.
func (e *Env) GetOrDefault(key string, defaultValue string) string {
	envKey, ok := e.env[key]
	if !ok {
		return defaultValue
	} else if !envKey.isSet {
		return defaultValue
	}
	return envKey.value
}

// Debug returns a map of environment variables and their masked values.
func (e *Env) Debug() map[string]string {
	m := make(map[string]string)
	for key, envKey := range e.env {
		m[key] = strings.Repeat("*", utf8.RuneCountInString(envKey.value))
	}
	return m
}

// validate checks if all required environment keys are set and if their values are valid according to the rules.
func (e *Env) validate() error {
	for key, envKey := range e.env {
		if !envKey.isValid() {
			return fmt.Errorf("%w: env key %s is invalid", ErrValidation, key)
		}
	}

	return nil
}

// loadEnvFiles loads environment files from the list of files to load.
func (e *Env) loadEnvFiles() error {
	for _, filename := range e.envFiles {
		env, err := e.ReadFile(filename)
		if err != nil {
			return err
		}

		for key, value := range env {
			envKey, ok := e.env[key]
			if !ok {
				e.env[key] = NewEnvKey(false)
				e.env[key].Set(value, SourceFile)
				continue
			}
			if !envKey.isSet {
				e.env[key].Set(value, SourceFile)
				continue
			}
			if envKey.isSet && envKey.source == SourceFile {
				e.env[key].Set(value, SourceFile)
			}
		}
	}
	return nil
}

// loadEnvVars loads environment variables from the provided keys.
func (e *Env) loadEnvVars() {
	keys := make([]string, 0, len(e.env))
	for key := range e.env {
		keys = append(keys, key)
	}

	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			e.env[key].Set(value, SourceVar)
		}
	}
}

// IsPath returns a rule that checks if the value is a valid path.
func IsPath(mustExist bool) EnvKeyRule {
	return func(value string) bool {
		if mustExist {
			if _, err := os.Stat(value); err == nil {
				return true
			}
		} else {
			return fs.ValidPath(value)
		}
		return false
	}
}

// MinLength returns a rule that checks if the value is at least the given length.
func MinLength(minLength int) EnvKeyRule {
	return func(value string) bool {
		return len(value) >= minLength
	}
}
