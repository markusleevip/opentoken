package config

type NodeAgent struct {
	Enabled             bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	ServerWSURL         string   `mapstructure:"server-ws-url" json:"server-ws-url" yaml:"server-ws-url"`
	Token               string   `mapstructure:"token" json:"token" yaml:"token"`
	NodeName            string   `mapstructure:"node-name" json:"node-name" yaml:"node-name"`
	Models              []string `mapstructure:"models" json:"models" yaml:"models"`
	ReconnectBackoffSec int      `mapstructure:"reconnect-backoff-sec" json:"reconnect-backoff-sec" yaml:"reconnect-backoff-sec"`
	HandshakeTimeoutSec int      `mapstructure:"handshake-timeout-sec" json:"handshake-timeout-sec" yaml:"handshake-timeout-sec"`
}
