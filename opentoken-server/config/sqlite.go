package config

type Sqlite struct {
	DbPath        string `mapstructure:"db-path" json:"db-path" yaml:"db-path"`
	Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	LogMode       string `mapstructure:"log-mode" json:"log-mode" yaml:"log-mode"`
	Singular      bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	LogZap        bool   `mapstructure:"log-zap" json:"log-zap" yaml:"log-zap"`
	MaxIdleConns  int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"`
	MaxOpenConns  int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"`
	BusyTimeoutMs int    `mapstructure:"busy-timeout-ms" json:"busy-timeout-ms" yaml:"busy-timeout-ms"`
	ForeignKeys   bool   `mapstructure:"foreign-keys" json:"foreign-keys" yaml:"foreign-keys"`
}
