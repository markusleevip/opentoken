package model

import (
	"errors"
	"time"

	"opentoken-server/utils"

	"gorm.io/gorm"
)

// User
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UID          string `gorm:"type:varchar(12);not null;uniqueIndex" json:"uid"`
	Username     string `gorm:"type:varchar(64);not null;uniqueIndex" json:"username"`
	PasswordSHA1 string `gorm:"type:char(40);not null" json:"-"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.UID == "" {
		u.UID = utils.GenerateUID()
	}
	if u.PasswordSHA1 == "" {
		return errors.New("password is required")
	}
	return nil
}
