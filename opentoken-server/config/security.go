package config

// Security 安全配置
type Security struct {
	APIKeyEncryptKey string `mapstructure:"api-key-encrypt-key" json:"api-key-encrypt-key" yaml:"api-key-encrypt-key"`
}
