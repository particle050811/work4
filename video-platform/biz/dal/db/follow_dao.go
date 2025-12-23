package db

import (
	"errors"

	"video-platform/biz/dal/model"

	"gorm.io/gorm"
)

// GetFollowUnscoped 查询关注记录（包含软删除）
func GetFollowUnscoped(store DBProvider, followerID, followingID uint) (*model.Follow, error) {
	var follow model.Follow
	err := store.DB().Unscoped().
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		First(&follow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &follow, nil
}

// CreateFollow 创建关注记录
func CreateFollow(store DBProvider, follow *model.Follow) error {
	return store.DB().Create(follow).Error
}

// RestoreFollow 恢复软删除的关注记录
func RestoreFollow(store DBProvider, followID uint) error {
	return store.DB().Unscoped().Model(&model.Follow{}).
		Where("id = ?", followID).
		Update("deleted_at", nil).Error
}

// SoftDeleteFollow 软删除关注记录
func SoftDeleteFollow(store DBProvider, followID uint) error {
	return store.DB().Delete(&model.Follow{}, followID).Error
}

// UpdateUserFollowingCount 更新用户关注数
func UpdateUserFollowingCount(store DBProvider, userID uint, delta int64) error {
	return store.DB().Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("following_count", gorm.Expr("following_count + ?", delta)).Error
}

// UpdateUserFollowerCount 更新用户粉丝数
func UpdateUserFollowerCount(store DBProvider, userID uint, delta int64) error {
	return store.DB().Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("follower_count", gorm.Expr("follower_count + ?", delta)).Error
}

// ListFollowings 获取用户关注列表
func ListFollowings(store DBProvider, userID uint, offset, limit int) ([]model.User, int64, error) {
	var total int64
	if err := store.DB().Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var follows []model.Follow
	if err := store.DB().
		Where("follower_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	if len(follows) == 0 {
		return []model.User{}, total, nil
	}

	// 获取被关注者 ID 列表
	followingIDs := make([]uint, 0, len(follows))
	for _, f := range follows {
		followingIDs = append(followingIDs, f.FollowingID)
	}

	// 查询用户信息
	var users []model.User
	if err := store.DB().Where("id IN ?", followingIDs).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 按关注顺序排列
	userMap := make(map[uint]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}
	ordered := make([]model.User, 0, len(followingIDs))
	for _, id := range followingIDs {
		if u, ok := userMap[id]; ok {
			ordered = append(ordered, u)
		}
	}

	return ordered, total, nil
}

// ListFollowers 获取用户粉丝列表
func ListFollowers(store DBProvider, userID uint, offset, limit int) ([]model.User, int64, error) {
	var total int64
	if err := store.DB().Model(&model.Follow{}).
		Where("following_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var follows []model.Follow
	if err := store.DB().
		Where("following_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	if len(follows) == 0 {
		return []model.User{}, total, nil
	}

	// 获取粉丝 ID 列表
	followerIDs := make([]uint, 0, len(follows))
	for _, f := range follows {
		followerIDs = append(followerIDs, f.FollowerID)
	}

	// 查询用户信息
	var users []model.User
	if err := store.DB().Where("id IN ?", followerIDs).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 按关注顺序排列
	userMap := make(map[uint]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}
	ordered := make([]model.User, 0, len(followerIDs))
	for _, id := range followerIDs {
		if u, ok := userMap[id]; ok {
			ordered = append(ordered, u)
		}
	}

	return ordered, total, nil
}

// ListFriends 获取用户好友列表（互相关注）
func ListFriends(store DBProvider, userID uint, offset, limit int) ([]model.User, int64, error) {
	// 好友 = 我关注的人中，同时也关注我的人
	// 使用子查询：找出 userID 关注的人中，也关注 userID 的人
	subQuery := store.DB().Model(&model.Follow{}).
		Select("following_id").
		Where("follower_id = ?", userID)

	var total int64
	if err := store.DB().Model(&model.Follow{}).
		Where("follower_id IN (?) AND following_id = ?", subQuery, userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var follows []model.Follow
	if err := store.DB().
		Where("follower_id IN (?) AND following_id = ?", subQuery, userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	if len(follows) == 0 {
		return []model.User{}, total, nil
	}

	// 获取好友 ID 列表（这些人关注了 userID，同时 userID 也关注他们）
	friendIDs := make([]uint, 0, len(follows))
	for _, f := range follows {
		friendIDs = append(friendIDs, f.FollowerID)
	}

	// 查询用户信息
	var users []model.User
	if err := store.DB().Where("id IN ?", friendIDs).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 按顺序排列
	userMap := make(map[uint]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}
	ordered := make([]model.User, 0, len(friendIDs))
	for _, id := range friendIDs {
		if u, ok := userMap[id]; ok {
			ordered = append(ordered, u)
		}
	}

	return ordered, total, nil
}
