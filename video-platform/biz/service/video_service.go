package service

import (
	"context"
	"fmt"
	"time"

	"video-platform/biz/dal"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/pkg/util"

	"github.com/redis/go-redis/v9"
)

const (
	hotVideosKey = "fanone:video:hot:zset"
	cacheTTL     = 5 * time.Minute
	topN         = 200
)

// VideoService 视频服务
type VideoService struct {
	store *dal.Store
}

// NewVideoService 创建视频服务实例
func NewVideoService(store *dal.Store) *VideoService {
	return &VideoService{store: store}
}

// CreateVideo 创建视频
func (s *VideoService) CreateVideo(ctx context.Context, video *model.Video) error {
	if err := db.CreateVideo(s.store, video); err != nil {
		return fmt.Errorf("创建视频记录失败: %w", err)
	}
	return nil
}

// GetVideoByID 根据 ID 获取视频
func (s *VideoService) GetVideoByID(ctx context.Context, videoID uint) (*model.Video, error) {
	video, err := db.GetVideoByID(s.store, videoID)
	if err != nil {
		return nil, fmt.Errorf("查询视频失败: %w", err)
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}
	return video, nil
}

// ListVideosByUser 获取用户发布的视频列表
func (s *VideoService) ListVideosByUser(ctx context.Context, userID uint, offset, limit int) ([]model.Video, int64, error) {
	return db.ListVideosByUser(s.store, userID, offset, limit)
}

// SearchVideosParams 搜索视频参数
type SearchVideosParams struct {
	Keywords  string
	Username  string
	FromDate  *time.Time
	ToDate    *time.Time
	SortByHot bool
}

// SearchVideos 搜索视频
func (s *VideoService) SearchVideos(ctx context.Context, params SearchVideosParams, offset, limit int) ([]model.Video, int64, error) {
	return db.SearchVideos(s.store, db.SearchVideosParams{
		Keywords:  params.Keywords,
		Username:  params.Username,
		FromDate:  params.FromDate,
		ToDate:    params.ToDate,
		SortByHot: params.SortByHot,
	}, offset, limit)
}

// ListVideoComments 获取视频评论列表
func (s *VideoService) ListVideoComments(ctx context.Context, videoID uint, offset, limit int) ([]db.VideoCommentWithUser, int64, error) {
	return db.ListTopLevelCommentsByVideo(s.store, videoID, offset, limit)
}

// GetHotVideos 获取热门视频列表
func (s *VideoService) GetHotVideos(ctx context.Context, offset, limit int) ([]model.Video, int64, error) {
	var total int64
	if err := s.store.DB().Model(&model.Video{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	redisCli := s.store.Redis()
	exists, err := redisCli.Exists(ctx, hotVideosKey).Result()
	if err != nil {
		return nil, 0, err
	}
	if exists == 0 {
		if err := s.rebuildHotVideosZSet(ctx, redisCli); err != nil {
			return nil, 0, err
		}
	}

	zs, err := redisCli.ZRevRangeWithScores(ctx, hotVideosKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, err
	}

	ids := make([]uint, 0, len(zs))
	for _, z := range zs {
		idStr, ok := z.Member.(string)
		if !ok {
			continue
		}
		id, err := util.ParseUint(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	videos, err := db.GetVideosByIDs(s.store, ids)
	if err != nil {
		return nil, 0, err
	}

	// 按热度排序
	byID := make(map[uint]model.Video, len(videos))
	for i := range videos {
		byID[videos[i].ID] = videos[i]
	}
	ordered := make([]model.Video, 0, len(ids))
	for _, id := range ids {
		if v, ok := byID[id]; ok {
			ordered = append(ordered, v)
		}
	}
	return ordered, total, nil
}

// rebuildHotVideosZSet 重建热榜缓存
func (s *VideoService) rebuildHotVideosZSet(ctx context.Context, redisCli dal.RedisClient) error {
	var top []model.Video
	if err := s.store.DB().
		Model(&model.Video{}).
		Order("(like_count*3 + comment_count*2 + visit_count) desc").
		Limit(topN).
		Find(&top).Error; err != nil {
		return err
	}

	zs := make([]redis.Z, 0, len(top))
	for i := range top {
		score := float64(top[i].LikeCount*3 + top[i].CommentCount*2 + top[i].VisitCount)
		zs = append(zs, redis.Z{
			Score:  score,
			Member: fmt.Sprintf("%d", top[i].ID),
		})
	}

	if err := redisCli.Del(ctx, hotVideosKey).Err(); err != nil {
		return err
	}
	if len(zs) > 0 {
		if err := redisCli.ZAdd(ctx, hotVideosKey, zs...).Err(); err != nil {
			return err
		}
	}
	if err := redisCli.Expire(ctx, hotVideosKey, cacheTTL).Err(); err != nil {
		return err
	}
	return nil
}
