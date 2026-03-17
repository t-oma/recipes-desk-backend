package rabbitmq

import (
	"errors"
	"fmt"
	"time"
)

type ConnectionManagerConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	VHost             string
	Heartbeat         time.Duration `yaml:"heartbeat"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout"`
	MaxChannels       int           `yaml:"maxChannels"`
	PrefetchCount     int           `yaml:"prefetchCount"`
	EnableTLS         bool          `yaml:"enableTLS"` //nolint:tagliatelle // TLS is abbreviation
}

func (c ConnectionManagerConfig) URI() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.VHost,
	)
}

func (c ConnectionManagerConfig) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.User == "" {
		return errors.New("user is required")
	}
	if c.Password == "" {
		return errors.New("password is required")
	}
	if c.VHost == "" {
		return errors.New("vhost is required")
	}
	if c.Heartbeat == 0 {
		return errors.New("heartbeat is required")
	}
	if c.ConnectionTimeout == 0 {
		return errors.New("connection timeout is required")
	}
	if c.MaxChannels == 0 {
		return errors.New("max channels is required")
	}
	if c.PrefetchCount == 0 {
		return errors.New("prefetch count is required")
	}

	return nil
}
