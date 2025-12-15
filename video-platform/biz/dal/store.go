package dal

import (
	"log"

	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Store 统一数据访问层，包含 DB 和 Redis 客户端
type Store struct {
	db    *gorm.DB
	redis *redis.Client // 可为 nil（降级模式）
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

// DB 获取数据库实例（用于复杂查询或传递给 db 包函数）
func (s *Store) DB() *gorm.DB {
	return s.db
}

// HasRedis 检查 Redis 是否可用
func (s *Store) HasRedis() bool {
	return s.redis != nil
}

// autoMigrate 自动迁移所有模型
func autoMigrate(gormDB *gorm.DB) {
	err := gormDB.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")
}

