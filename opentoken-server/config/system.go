package config

type System struct {
	DbType        string   `mapstructure:"db-type" json:"db-type" yaml:"db-type"`    // 数据库类型:mysql(默认)|sqlite|sqlserver|postgresql
	OssType       string   `mapstructure:"oss-type" json:"oss-type" yaml:"oss-type"` // Oss类型
	RouterPrefix  string   `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`
	Host          string   `mapstructure:"host" json:"host" yaml:"host"`                   // 监听地址，空表示0.0.0.0
	Port          int      `mapstructure:"port" json:"port" yaml:"port"`                   // 端口号
	AdminUser     []string `mapstructure:"admin-user" json:"admin-user" yaml:"admin-user"` // 系统管理员账号（按 username）
	LimitCountIP  int      `mapstructure:"iplimit-count" json:"iplimit-count" yaml:"iplimit-count"`
	LimitTimeIP   int      `mapstructure:"iplimit-time" json:"iplimit-time" yaml:"iplimit-time"`
	UseMultipoint bool     `mapstructure:"use-multipoint" json:"use-multipoint" yaml:"use-multipoint"`    // 多点登录拦截
	UseStrictAuth bool     `mapstructure:"use-strict-auth" json:"use-strict-auth" yaml:"use-strict-auth"` // 使用树形角色分配模式
}
