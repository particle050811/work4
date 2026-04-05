package dal

import (
	"log"

	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Store 统一数据访问层，聚合 MySQL 和 Redis
type Store struct {
	db    *gorm.DB
	redis *redis.Client
}

var defaultStore *Store

// Init 初始化 Store（在 main.go 中调用）
func Init() {
	defaultStore = &Store{
		db:    db.InitMySQL(),
		redis: initRedis(), // 失败返回 nil，不 panic
	}
	autoMigrate(defaultStore.db)
}

// GetStore 获取全局 Store 实例
func GetStore() *Store {
	return defaultStore
}

// DB 获取数据库实例
func (s *Store) DB() *gorm.DB {
	return s.db
}

// Redis 获取 Redis 客户端
func (s *Store) Redis() RedisClient {
	return s.redis
}

// HasRedis 检查 Redis 是否可用
func (s *Store) HasRedis() bool {
	return s != nil && s.redis != nil
}

// WithTx 在事务中执行操作，返回带事务的 Store
func (s *Store) WithTx(fn func(txStore *Store) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txStore := &Store{db: tx, redis: s.redis}
		return fn(txStore)
	})
}

// autoMigrate 自动迁移所有模型
func autoMigrate(gormDB *gorm.DB) {
	err := gormDB.AutoMigrate(
		&model.User{},
		&model.Video{},
		&model.Comment{},
		&model.VideoLike{},
		&model.Follow{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")
}
