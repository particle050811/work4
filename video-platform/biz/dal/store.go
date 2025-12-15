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
	redis *redis.Client // 可为 nil（降级模式）
}

// 确保 Store 实现 StoreLike 接口
var _ StoreLike = (*Store)(nil)

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

// Redis 获取 Redis 客户端（实现 CacheProvider）
func (s *Store) Redis() RedisClient {
	return s.redis
}

// HasRedis 检查 Redis 是否可用
func (s *Store) HasRedis() bool {
	return s.redis != nil
}

// WithTx 在事务中执行操作，返回带事务的 Store
func (s *Store) WithTx(fn func(txStore *Store) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txStore := &Store{db: tx, redis: s.redis}
		return fn(txStore)
	})
}

// Close 关闭所有连接
func (s *Store) Close() error {
	// 关闭 MySQL
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
	// 关闭 Redis
	if s.redis != nil {
		s.redis.Close()
	}
	return nil
}

// autoMigrate 自动迁移所有模型
func autoMigrate(gormDB *gorm.DB) {
	err := gormDB.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")
}
