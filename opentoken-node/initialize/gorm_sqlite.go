package initialize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"opentoken-node/config"
	"opentoken-node/global"
	"opentoken-node/initialize/internal"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// GormSqlite 初始化Sqlite数据库
func GormSqlite() *gorm.DB {
	s := global.GLB_CONFIG.Sqlite
	return initSqliteDatabase(s)
}

// GormSqliteByConfig 通过传入配置初始化Sqlite数据库
func GormSqliteByConfig(s config.Sqlite) *gorm.DB {
	return initSqliteDatabase(s)
}

func initSqliteDatabase(s config.Sqlite) *gorm.DB {
	if strings.TrimSpace(s.DbPath) == "" {
		return nil
	}

	busyTimeout := s.BusyTimeoutMs
	if busyTimeout <= 0 {
		busyTimeout = 5000
	}

	dsn := s.DbPath
	if dir := filepath.Dir(dsn); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Failed to create SQLite directory %s: %v\n", dir, err)
			panic(err)
		}
	}

	general := config.GeneralDB{
		Prefix:       s.Prefix,
		LogMode:      s.LogMode,
		Singular:     s.Singular,
		LogZap:       s.LogZap,
		MaxIdleConns: s.MaxIdleConns,
		MaxOpenConns: s.MaxOpenConns,
	}

	fmt.Printf("Connecting to SQLite with DSN: %s\n", dsn)

	if db, err := gorm.Open(sqlite.Open(dsn), internal.Gorm.Config(general)); err != nil {
		fmt.Printf("Failed to connect to SQLite: %v\n", err)
		panic(err)
	} else {
		if err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", busyTimeout)).Error; err != nil {
			fmt.Printf("Failed to set PRAGMA busy_timeout: %v\n", err)
			panic(err)
		}
		if s.ForeignKeys {
			if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
				fmt.Printf("Failed to set PRAGMA foreign_keys=ON: %v\n", err)
				panic(err)
			}
		} else {
			if err := db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
				fmt.Printf("Failed to set PRAGMA foreign_keys=OFF: %v\n", err)
				panic(err)
			}
		}

		sqlDB, err := db.DB()
		if err != nil {
			fmt.Printf("Failed to get underlying sql.DB: %v\n", err)
			panic(err)
		}

		maxOpen := s.MaxOpenConns
		if maxOpen <= 0 {
			maxOpen = 1
		}
		maxIdle := s.MaxIdleConns
		if maxIdle < 0 {
			maxIdle = 0
		}
		if maxIdle == 0 {
			maxIdle = 1
		}

		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetMaxIdleConns(maxIdle)

		fmt.Println("Successfully connected to SQLite database")
		return db
	}
}
