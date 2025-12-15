package dal

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// initRedis 初始化 Redis 客户端
// 如果未配置或连接失败，返回 nil（降级模式）
func initRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		log.Println("[Redis] REDIS_ADDR 未配置，跳过 Redis 初始化（降级模式）")
		return nil
	}

	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			db = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis] 连接失败: %v（降级模式）", err)
		return nil
	}

	log.Printf("[Redis] 连接成功: %s", addr)
	return client
}
