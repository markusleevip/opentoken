package initialize

import (
	"opentoken-node/global"
)

func bizModel() error {
	db := global.GLB_DB
	err := db.AutoMigrate()
	if err != nil {
		return err
	}
	return nil
}
