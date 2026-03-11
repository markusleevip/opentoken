package initialize

import (
	"fmt"
	"opentoken-node/config"
	"opentoken-node/global"
	"opentoken-node/initialize/internal"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GormMysql 初始化Mysql数据库
func GormMysql() *gorm.DB {
	m := global.GLB_CONFIG.Mysql
	return initMysqlDatabase(m)
}

// GormMysqlByConfig 通过传入配置初始化Mysql数据库
func GormMysqlByConfig(m config.Mysql) *gorm.DB {
	return initMysqlDatabase(m)
}

// initMysqlDatabase 初始化Mysql数据库的辅助函数
func initMysqlDatabase(m config.Mysql) *gorm.DB {
	if m.Dbname == "" {
		return nil
	}

	dsn := m.Dsn()
	fmt.Printf("Connecting to MySQL with DSN: %s\n", dsn)

	// 添加超时参数
	dsnWithTimeout := dsn + "&timeout=10s&readTimeout=30s&writeTimeout=30s"

	mysqlConfig := mysql.Config{
		DSN:                       dsnWithTimeout, // DSN data source name
		DefaultStringSize:         191,            // string 类型字段的默认长度
		SkipInitializeWithVersion: false,          // 根据版本自动配置
	}

	// 数据库配置
	general := m.GeneralDB
	if db, err := gorm.Open(mysql.New(mysqlConfig), internal.Gorm.Config(general)); err != nil {
		fmt.Printf("Failed to connect to MySQL: %v\n", err)
		panic(err)
	} else {
		fmt.Println("Successfully connected to MySQL database")
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, err := db.DB()
		if err != nil {
			fmt.Printf("Failed to get underlying sql.DB: %v\n", err)
			panic(err)
		}

		// 设置连接池参数
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)

		// 设置连接超时
		sqlDB.SetConnMaxLifetime(0)    // 连接永不过期
		sqlDB.SetConnMaxIdleTime(3600) // 空闲连接1小时后超时

		// 测试连接
		if err := sqlDB.Ping(); err != nil {
			fmt.Printf("Failed to ping MySQL: %v\n", err)
			panic(err)
		} else {
			fmt.Println("Successfully pinged MySQL database")
		}

		return db
	}
}
