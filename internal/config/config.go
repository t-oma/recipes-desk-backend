package config

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App     App    `mapstructure:"app"`
	Server  Server `mapstructure:"server"`
	MongoDB MongoDB
	JWT     JWT `mapstructure:"jwt"`
}

type App struct {
	Environment string `mapstructure:"environment"`
}

type Server struct {
	Port string `mapstructure:"port"`
	// Timeouts
	WriteTimeout time.Duration `mapstructure:"write-timeout"`
	ReadTimeout  time.Duration `mapstructure:"read-timeout"`
}

type MongoDB struct {
	URI      string
	Database string
}

type JWT struct {
	Secret        string
	AccessExpiry  time.Duration `mapstructure:"access-expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh-expiry"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.AddConfigPath(".")
	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		slog.Warn("Warning: .env file not found, using environment variables only")
	}

	v.AddConfigPath("./configs/")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if err := v.MergeInConfig(); err != nil {
		return nil, err
	}

	v.SetDefault("app.environment", "development")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.write-timeout", "10s")
	v.SetDefault("server.read-timeout", "10s")
	v.SetDefault("jwt.access-expiry", "15m")
	v.SetDefault("jwt.refresh-expiry", "168h")

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	config.MongoDB.URI = v.GetString("MONGO_URI")
	if config.MongoDB.URI == "" {
		panic("MONGO_URI is required")
	}

	config.MongoDB.Database = v.GetString("MONGO_DATABASE")
	if config.MongoDB.Database == "" {
		panic("MONGO_DATABASE is required")
	}

	config.JWT.Secret = v.GetString("JWT_SECRET")
	if config.JWT.Secret == "" {
		panic("JWT_SECRET is required")
	}

	return &config, nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
