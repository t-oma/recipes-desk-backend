package connection

import "time"

type ConnectionManagerConfig struct {
	Heartbeat         time.Duration `yaml:"heartbeat"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout"`
	MaxChannels       int           `yaml:"maxChannels"`
	PrefetchCount     int           `yaml:"prefetchCount"`
	EnableTLS         bool          `yaml:"enableTLS"` //nolint:tagliatelle // TLS is abbreviation
}
