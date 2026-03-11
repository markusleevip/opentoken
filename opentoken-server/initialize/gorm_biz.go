package initialize

import (
	"opentoken-server/global"
)

func bizModel() error {
	_ = global.GLB_DB
	return nil
}
