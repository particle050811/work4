package db

import (
	"errors"

	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

// ==================== 用户相关数据库操作 ====================

// CreateUser 创建用户
func CreateUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// GetUserByID 根据 ID 获取用户
func GetUserByID(db *gorm.DB, id uint) (*model.User, error) {
	var user model.User
	err := db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(db *gorm.DB, user *model.User) error {
	return db.Save(user).Error
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(db *gorm.DB, userID uint, avatarURL string) error {
	return db.Model(&model.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error
}

// UserExists 检查用户名是否存在
func UserExists(db *gorm.DB, username string) (bool, error) {
	var count int64
	err := db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
