package initialize

import (
	"context"
	"fmt"
	"os"
	"time"

	"opentoken-server/global"
	"opentoken-server/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func renameTableIfNeeded(db *gorm.DB, from string, to string) error {
	m := db.Migrator()
	hasFrom := m.HasTable(from)
	hasTo := m.HasTable(to)
	if !hasFrom || hasTo {
		return nil
	}
	fmt.Printf("Renaming table %s -> %s\n", from, to)
	return m.RenameTable(from, to)
}

func migratePointsTableNames(db *gorm.DB) error {
	// points_* -> 去前缀
	if err := renameTableIfNeeded(db, "points_accounts", "accounts"); err != nil {
		return err
	}
	if err := renameTableIfNeeded(db, "points_rules", "rules"); err != nil {
		return err
	}
	if err := renameTableIfNeeded(db, "points_orders", "orders"); err != nil {
		return err
	}
	return nil
}

func Gorm() *gorm.DB {
	switch global.GLB_CONFIG.System.DbType {
	case "mysql":
		global.GLB_ACTIVE_DBNAME = &global.GLB_CONFIG.Mysql.Dbname
		return GormMysql()
	case "sqlite":
		global.GLB_ACTIVE_DBNAME = &global.GLB_CONFIG.Sqlite.DbPath
		return GormSqlite()
	default:
		global.GLB_ACTIVE_DBNAME = &global.GLB_CONFIG.Mysql.Dbname
		return GormMysql()
	}
}

func RegisterTables() {
	fmt.Println("register table")
	if global.GLB_DB == nil {
		global.GLB_LOG.Error("Database connection is nil")
		os.Exit(1)
	}

	db := global.GLB_DB

	// 获取底层的 SQL DB 来设置超时
	sqlDB, err := db.DB()
	if err != nil {
		global.GLB_LOG.Error("Failed to get underlying SQL DB", zap.Error(err))
		os.Exit(1)
	}

	// 设置 SQL 查询超时
	sqlDB.SetConnMaxLifetime(60) // 60秒

	fmt.Println("Starting table migration...")

	// 使用上下文超时来防止无限期挂起
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 逐个迁移表，便于定位问题
	fmt.Println("Migrating core tables...")
	db = db.WithContext(ctx)
	if err := db.AutoMigrate(&model.Version{}, &model.User{}, &model.NodeCredential{}); err != nil {
		global.GLB_LOG.Error("Core table migration failed", zap.Error(err))
		os.Exit(1)
	}
	fmt.Println("Core tables migrated successfully")

	fmt.Println("All tables migrated successfully")
	global.GLB_LOG.Info("register table success")

	err = bizModel()

	if err != nil {
		global.GLB_LOG.Error("register biz_table failed", zap.Error(err))
		os.Exit(1)
	}
	fmt.Println("Biz models registered successfully")
	global.GLB_LOG.Info("register table success")
}
