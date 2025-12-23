package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"video-platform/biz/dal"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
)

// InteractionService 互动服务
type InteractionService struct {
	store *dal.Store
}

// NewInteractionService 创建互动服务实例
func NewInteractionService(store *dal.Store) *InteractionService {
	return &InteractionService{store: store}
}

// LikeActionType 点赞操作类型
type LikeActionType int32

const (
	LikeActionAdd    LikeActionType = 1 // 点赞
	LikeActionCancel LikeActionType = 2 // 取消点赞
)

// LikeVideo 点赞/取消点赞视频
// 返回 delta: 1=新增点赞, -1=取消点赞, 0=幂等无操作
func (s *InteractionService) LikeVideo(ctx context.Context, userID, videoID uint, actionType LikeActionType) (int64, error) {
	// 检查视频是否存在
	video, err := db.GetVideoByID(s.store, videoID)
	if err != nil {
		return 0, fmt.Errorf("查询视频失败: %w", err)
	}
	if video == nil {
		return 0, ErrVideoNotFound
	}

	var likeDelta int64
	err = s.store.WithTx(func(txStore *dal.Store) error {
		// 查询点赞记录（包含软删除）
		like, err := db.GetVideoLikeUnscoped(txStore, userID, videoID)
		if err != nil {
			return err
		}

		if actionType == LikeActionAdd { // 点赞
			if like == nil {
				// 不存在，创建新记录
				newLike := &model.VideoLike{
					UserID:  userID,
					VideoID: videoID,
				}
				if err := db.CreateVideoLike(txStore, newLike); err != nil {
					return err
				}
				likeDelta = 1
			} else if like.DeletedAt.Valid {
				// 已软删除，恢复记录
				if err := db.RestoreVideoLike(txStore, like.ID); err != nil {
					return err
				}
				likeDelta = 1
			}
			// 已点赞则幂等返回，likeDelta = 0
		} else { // 取消点赞
			if like != nil && !like.DeletedAt.Valid {
				// 存在且未删除，软删除
				if err := db.SoftDeleteVideoLike(txStore, like.ID); err != nil {
					return err
				}
				likeDelta = -1
			}
			// 不存在或已删除则幂等返回，likeDelta = 0
		}

		// 更新视频点赞数
		if likeDelta != 0 {
			if err := db.IncreaseVideoLikeCount(txStore, videoID, likeDelta); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	// 更新 Redis 热榜缓存（失败不阻塞主流程）
	if likeDelta != 0 {
		scoreDelta := float64(likeDelta * 3) // 点赞权重为 3
		if err := s.store.Redis().ZIncrBy(ctx, hotVideosKey, scoreDelta, fmt.Sprintf("%d", videoID)).Err(); err != nil {
			log.Printf("[互动模块][点赞操作] 更新热榜缓存失败 video_id=%d: %v", videoID, err)
		}
	}

	return likeDelta, nil
}

// ListLikedVideos 获取用户点赞的视频列表
func (s *InteractionService) ListLikedVideos(ctx context.Context, userID uint, offset, limit int) ([]model.Video, int64, error) {
	videoIDs, total, err := db.ListVideoLikesByUser(s.store, userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("查询点赞列表失败: %w", err)
	}

	if len(videoIDs) == 0 {
		return []model.Video{}, total, nil
	}

	videos, err := db.GetVideosByIDs(s.store, videoIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("查询视频详情失败: %w", err)
	}

	// 按点赞顺序排列
	videoMap := make(map[uint]model.Video, len(videos))
	for _, v := range videos {
		videoMap[v.ID] = v
	}
	ordered := make([]model.Video, 0, len(videoIDs))
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			ordered = append(ordered, v)
		}
	}

	return ordered, total, nil
}

// PublishComment 发布评论
func (s *InteractionService) PublishComment(ctx context.Context, userID, videoID uint, content string) error {
	// 校验内容
	content = strings.TrimSpace(content)
	if content == "" {
		return ErrCommentEmpty
	}
	if len(content) > 1000 {
		return ErrCommentTooLong
	}

	// 检查视频是否存在
	video, err := db.GetVideoByID(s.store, videoID)
	if err != nil {
		return fmt.Errorf("查询视频失败: %w", err)
	}
	if video == nil {
		return ErrVideoNotFound
	}

	// 事务：创建评论 + 更新视频评论数
	err = s.store.WithTx(func(txStore *dal.Store) error {
		comment := &model.Comment{
			UserID:  userID,
			VideoID: videoID,
			Content: content,
		}
		if err := db.CreateComment(txStore, comment); err != nil {
			return err
		}
		return db.IncreaseVideoCommentCount(txStore, videoID, 1)
	})

	if err != nil {
		return err
	}

	// 更新 Redis 热榜缓存（评论权重为 2）
	if err := s.store.Redis().ZIncrBy(ctx, hotVideosKey, 2.0, fmt.Sprintf("%d", videoID)).Err(); err != nil {
		log.Printf("[互动模块][发布评论] 更新热榜缓存失败 video_id=%d: %v", videoID, err)
	}

	return nil
}

// ListUserComments 获取用户发表的评论列表
func (s *InteractionService) ListUserComments(ctx context.Context, userID uint, offset, limit int) ([]model.Comment, int64, error) {
	return db.ListCommentsByUser(s.store, userID, offset, limit)
}

// DeleteComment 删除评论
func (s *InteractionService) DeleteComment(ctx context.Context, userID, commentID uint) error {
	// 查询评论
	comment, err := db.GetCommentByID(s.store, commentID)
	if err != nil {
		return fmt.Errorf("查询评论失败: %w", err)
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	// 权限校验：只能删除自己的评论
	if comment.UserID != userID {
		return ErrNoPermission
	}

	videoID := comment.VideoID

	// 事务：软删除评论 + 更新视频评论数
	err = s.store.WithTx(func(txStore *dal.Store) error {
		if err := txStore.DB().Delete(&model.Comment{}, commentID).Error; err != nil {
			return err
		}
		return db.IncreaseVideoCommentCount(txStore, videoID, -1)
	})

	if err != nil {
		return err
	}

	// 更新 Redis 热榜缓存（评论权重为 -2）
	if err := s.store.Redis().ZIncrBy(ctx, hotVideosKey, -2.0, fmt.Sprintf("%d", videoID)).Err(); err != nil {
		log.Printf("[互动模块][删除评论] 更新热榜缓存失败 video_id=%d: %v", videoID, err)
	}

	return nil
}
