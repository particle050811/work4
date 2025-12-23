package model

import (
	"time"

	"gorm.io/gorm"
)

// Follow 用户关注关系
type Follow struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FollowerID uint           `gorm:"uniqueIndex:idx_follower_following;index;not null" json:"follower_id"` // 关注者（发起关注的人）
	FollowingID uint          `gorm:"uniqueIndex:idx_follower_following;index;not null" json:"following_id"` // 被关注者
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // 软删除支持取关后再次关注
}

func (Follow) TableName() string {
	return "follows"
}
