package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"recipes-desk/pkg/config"
)

var (
	ErrMissingJWTSecret = errors.New("JWT secret is required")
	ErrInvalidJWTSecret = errors.New("JWT secret must be at least 32 characters long")
)

const (
	DefaultJWTSecretLength  = 32
	DefaultJWTAccessExpiry  = 15 * time.Minute
	DefaultJWTRefreshExpiry = 168 * time.Hour
)

type (
	Config struct {
		JWT JWT `yaml:"jwt"`
	}
	JWT struct {
		Secret string
		Expiry JWTExpiry `yaml:"expiry"`
	}
	JWTExpiry struct {
		Access  time.Duration `yaml:"access"`
		Refresh time.Duration `yaml:"refresh"`
	}
)

func Load() (*Config, error) {
	cfg, err := config.LoadYAML[Config]("./internal/modules/auth/config/auth.yaml")
	if err != nil {
		return nil, err
	}

	cfg.JWT.Secret = os.Getenv("AUTH_JWT_SECRET")

	// Validation
	if cfg.JWT.Secret == "" {
		return nil, ErrMissingJWTSecret
	}
	if len(cfg.JWT.Secret) < DefaultJWTSecretLength {
		return nil, ErrInvalidJWTSecret
	}
	if cfg.JWT.Expiry.Access == 0 {
		fmt.Printf(
			"WARNING: JWT access expiry is not set, using default value: %s\n",
			DefaultJWTAccessExpiry,
		)
		cfg.JWT.Expiry.Access = DefaultJWTAccessExpiry
	}
	if cfg.JWT.Expiry.Refresh == 0 {
		fmt.Printf(
			"WARNING: JWT refresh expiry is not set, using default value: %s\n",
			DefaultJWTRefreshExpiry,
		)
		cfg.JWT.Expiry.Refresh = DefaultJWTRefreshExpiry
	}

	return cfg, nil
}
