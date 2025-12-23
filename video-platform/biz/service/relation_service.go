package service

import (
	"context"
	"fmt"
	"log"

	"video-platform/biz/dal"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
)

// RelationService 社交关系服务
type RelationService struct {
	store *dal.Store
}

// NewRelationService 创建社交关系服务实例
func NewRelationService(store *dal.Store) *RelationService {
	return &RelationService{store: store}
}

// FollowActionType 关注操作类型
type FollowActionType int32

const (
	FollowActionFollow   FollowActionType = 1 // 关注
	FollowActionUnfollow FollowActionType = 2 // 取关
)

// FollowUser 关注/取消关注用户
// 返回 delta: 1=新增关注, -1=取消关注, 0=幂等无操作
func (s *RelationService) FollowUser(ctx context.Context, followerID, followingID uint, actionType FollowActionType) (int64, error) {
	// 不能关注自己
	if followerID == followingID {
		return 0, ErrCannotFollowSelf
	}

	// 检查目标用户是否存在
	targetUser, err := db.GetUserByID(s.store, followingID)
	if err != nil {
		return 0, fmt.Errorf("查询目标用户失败: %w", err)
	}
	if targetUser == nil {
		return 0, ErrUserNotFound
	}

	var followDelta int64
	err = s.store.WithTx(func(txStore *dal.Store) error {
		// 查询关注记录（包含软删除）
		follow, err := db.GetFollowUnscoped(txStore, followerID, followingID)
		if err != nil {
			return err
		}

		if actionType == FollowActionFollow { // 关注
			if follow == nil {
				// 不存在，创建新记录
				newFollow := &model.Follow{
					FollowerID:  followerID,
					FollowingID: followingID,
				}
				if err := db.CreateFollow(txStore, newFollow); err != nil {
					return err
				}
				followDelta = 1
			} else if follow.DeletedAt.Valid {
				// 已软删除，恢复记录
				if err := db.RestoreFollow(txStore, follow.ID); err != nil {
					return err
				}
				followDelta = 1
			}
			// 已关注则幂等返回
		} else { // 取关
			if follow != nil && !follow.DeletedAt.Valid {
				// 存在且未删除，软删除
				if err := db.SoftDeleteFollow(txStore, follow.ID); err != nil {
					return err
				}
				followDelta = -1
			}
			// 不存在或已删除则幂等返回
		}

		// 更新用户关注/粉丝计数
		if followDelta != 0 {
			// 更新关注者的"关注数"
			if err := db.UpdateUserFollowingCount(txStore, followerID, followDelta); err != nil {
				return err
			}
			// 更新被关注者的"粉丝数"
			if err := db.UpdateUserFollowerCount(txStore, followingID, followDelta); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[社交模块][关注操作] 事务执行失败 follower_id=%d following_id=%d: %v", followerID, followingID, err)
		return 0, err
	}

	return followDelta, nil
}

// ListFollowings 获取用户关注列表
func (s *RelationService) ListFollowings(ctx context.Context, userID uint, offset, limit int) ([]model.User, int64, error) {
	return db.ListFollowings(s.store, userID, offset, limit)
}

// ListFollowers 获取用户粉丝列表
func (s *RelationService) ListFollowers(ctx context.Context, userID uint, offset, limit int) ([]model.User, int64, error) {
	return db.ListFollowers(s.store, userID, offset, limit)
}

// ListFriends 获取用户好友列表（互相关注）
func (s *RelationService) ListFriends(ctx context.Context, userID uint, offset, limit int) ([]model.User, int64, error) {
	return db.ListFriends(s.store, userID, offset, limit)
}
