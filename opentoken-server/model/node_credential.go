package model

import "time"

// NodeCredential 用于给 node 分配通信凭证（仅存哈希）
type NodeCredential struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name       string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"name"`
	TokenHash  string     `gorm:"type:char(64);not null;uniqueIndex" json:"-"`
	Enabled    bool       `gorm:"not null;default:true" json:"enabled"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

func (NodeCredential) TableName() string { return "node_credentials" }
