package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password       string         `gorm:"size:255;not null" json:"-"` // 密码不对外暴露
	AvatarURL      string         `gorm:"size:500" json:"avatar_url"`
	FollowingCount int64          `gorm:"default:0" json:"following_count"` // 关注数
	FollowerCount  int64          `gorm:"default:0" json:"follower_count"`  // 粉丝数
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // 软删除
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
