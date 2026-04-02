package config

type Server struct {
	JWT       JWT     `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Zap       Zap     `mapstructure:"zap" json:"zap" yaml:"zap"`
	Email     Email   `mapstructure:"email" json:"email" yaml:"email"`
	System    System  `mapstructure:"system" json:"system" yaml:"system"`
	Captcha   Captcha `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	Security  Security `mapstructure:"security" json:"security" yaml:"security"`
	// auto
	AutoCode Autocode `mapstructure:"autocode" json:"autocode" yaml:"autocode"`
	// gorm
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Sqlite Sqlite `mapstructure:"sqlite" json:"sqlite" yaml:"sqlite"`
	// oss
	Local     Local     `mapstructure:"local" json:"local" yaml:"local"`
	Minio     Minio     `mapstructure:"minio" json:"minio" yaml:"minio"`

	Excel Excel `mapstructure:"excel" json:"excel" yaml:"excel"`
	// 跨域配置
	Cors CORS `mapstructure:"cors" json:"cors" yaml:"cors"`
}
