package model

import "time"

// NodeCredential 用于给 node 分配通信凭证
type NodeCredential struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"name"`
	TokenHash   string     `gorm:"type:char(64);not null;uniqueIndex" json:"-"`
	TokenPrefix string     `gorm:"type:varchar(20);not null" json:"token_prefix"` // 用于展示
	TokenEnc    string     `gorm:"type:varchar(255)" json:"-"`                      // 加密存储的完整 token
	Enabled     bool       `gorm:"not null;default:true" json:"enabled"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
}

func (NodeCredential) TableName() string { return "node_credentials" }
