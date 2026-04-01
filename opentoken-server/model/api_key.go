package model

import "time"

// APIKey 用于访问 LLM API 的用户凭证
type APIKey struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string     `gorm:"type:varchar(64);not null" json:"name"`
	TokenHash   string     `gorm:"type:char(64);not null;uniqueIndex" json:"-"`
	TokenPrefix string     `gorm:"type:varchar(20);not null" json:"token_prefix"` // 用于展示，如 "sk-abc12..."
	TokenEnc    string     `gorm:"type:varchar(255)" json:"-"`                      // 加密存储的完整 token
	Enabled     bool       `gorm:"not null;default:true" json:"enabled"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	Description string     `gorm:"type:varchar(255)" json:"description"`
}

func (APIKey) TableName() string { return "api_keys" }
