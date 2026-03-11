package global

import (
	"opentoken-node/config"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GLB_CONFIG        config.Server
	GLB_VERSION       *string
	GLB_DB            *gorm.DB
	GLB_VP            *viper.Viper
	GLB_LOG           *zap.Logger
	GLB_ACTIVE_DBNAME *string
)
