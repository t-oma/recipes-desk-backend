package config

import (
	"errors"
	"time"

	"github.com/rs/zerolog"

	"recipes-desk/pkg/config"
)

var ErrInvalidJWTSecret = errors.New("JWT secret must be at least 32 characters long")

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

func Load(log *zerolog.Logger) (*Config, error) {
	env := config.NewEnv()
	env.WithEnvVars()
	env.Required("AUTH_JWT_SECRET")
	env.KeyRules("AUTH_JWT_SECRET", func(value string) bool {
		return len(value) >= DefaultJWTSecretLength
	})
	if err := env.Load(); err != nil {
		return nil, err
	}

	cfg, err := config.LoadYAML[Config](
		env.GetOrDefault("AUTH_CONFIG_PATH", "./internal/modules/auth/config/auth.yaml"),
	)
	if err != nil {
		return nil, err
	}

	cfg.JWT.Secret = env.Get("AUTH_JWT_SECRET")

	if cfg.JWT.Expiry.Access <= 0 {
		cfg.JWT.Expiry.Access = DefaultJWTAccessExpiry
		log.Warn().Dur("access expiry", DefaultJWTAccessExpiry).
			Msg("JWT access expiry is not set, using default value")
	}
	if cfg.JWT.Expiry.Refresh <= 0 {
		cfg.JWT.Expiry.Refresh = DefaultJWTRefreshExpiry
		log.Warn().Dur("refresh expiry", DefaultJWTRefreshExpiry).
			Msg("JWT refresh expiry is not set, using default value")
	}

	return cfg, nil
}
