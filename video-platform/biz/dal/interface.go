package dal

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ==================== 接口定义（避免循环依赖） ====================

// DBProvider 提供数据库访问能力
type DBProvider interface {
	DB() *gorm.DB
}

// RedisClient Redis 最小接口（按需扩展）
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	// Sorted Set（排行榜用）
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd
	ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	ZScore(ctx context.Context, key string, member string) *redis.FloatCmd
}
