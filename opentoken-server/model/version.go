package model

import (
	"time"
	"gorm.io/gorm"
)

// 版本管理 结构体  Version
type Version struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	VersionName *string        `json:"versionName" form:"versionName" gorm:"comment:版本名称;column:version_name;size:255;" binding:"required"` //版本名称
	VersionCode *string        `json:"versionCode" form:"versionCode" gorm:"comment:版本号;column:version_code;size:100;" binding:"required"`  //版本号
	Description *string        `json:"description" form:"description" gorm:"comment:版本描述;column:description;size:500;"`                     //版本描述
	VersionData *string        `json:"versionData" form:"versionData" gorm:"comment:版本数据JSON;column:version_data;type:text;"`               //版本数据
}

// TableName 版本管理 Version
func (Version) TableName() string {
	return "versions"
}
