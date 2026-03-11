package config

type UpstreamLLM struct {
	BaseURL         string            `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	APIKey          string            `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	ChatPath        string            `mapstructure:"chat-path" json:"chat-path" yaml:"chat-path"`
	TimeoutSec      int               `mapstructure:"timeout-sec" json:"timeout-sec" yaml:"timeout-sec"`
	ModelMap        map[string]string `mapstructure:"model-map" json:"model-map" yaml:"model-map"`
	InsecureSkipTLS bool              `mapstructure:"insecure-skip-tls" json:"insecure-skip-tls" yaml:"insecure-skip-tls"`
}
