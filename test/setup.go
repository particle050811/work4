package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

const hotVideosKey = "fanone:video:hot:zset"

// RelationFixture 记录社交模块预置数据。
//
// 预置用户与关系如下：
// 1. relation_alice / RelationPass123
//   - 关注了 relation_bob、relation_carol
//   - 与 relation_bob 互相关注，因此好友列表应包含 relation_bob
//
// 2. relation_bob / RelationPass123
//   - 关注了 relation_alice
//   - 与 relation_alice 互相关注，因此好友列表应包含 relation_alice
//
// 3. relation_carol / RelationPass123
//   - 没有主动关注任何人
//   - 被 relation_alice 关注，因此只会出现在 relation_alice 的关注列表和自己的粉丝列表验证场景里
type RelationFixture struct {
	Alice SeedUser
	Bob   SeedUser
	Carol SeedUser
}

type SeedUser struct {
	Username     string
	Password     string
	UserID       string
	AccessToken  string
	RefreshToken string
}

func loginSeedUser(client *http.Client, baseURL, username, password string) (*SeedUser, error) {
	// 首次准备测试数据时，可临时打开下面这段注册逻辑，插入一次后再继续保持注释状态。
	//
	// registerResult := testRegister(client, baseURL, username, password)
	// if registerResult.Err != nil {
	// 	return nil, fmt.Errorf("注册测试用户 %s 失败: %w", username, registerResult.Err)
	// }
	// if registerResult.Data.Base.Code != 0 && !isUserAlreadyExists(registerResult.Data.Base.Msg) {
	// 	return nil, fmt.Errorf("注册测试用户 %s 失败: %s", username, registerResult.Data.Base.Msg)
	// }

	loginResult := testLogin(client, baseURL, username, password)
	if loginResult.Err != nil {
		return nil, fmt.Errorf("登录测试用户 %s 失败: %w", username, loginResult.Err)
	}
	if loginResult.Data.Base.Code != 0 {
		return nil, fmt.Errorf("登录测试用户 %s 失败: %s；如为首次运行，请先临时打开 setup.go 中的注册代码插入一次", username, loginResult.Data.Base.Msg)
	}

	return &SeedUser{
		Username:     username,
		Password:     password,
		UserID:       loginResult.Data.Data.ID,
		AccessToken:  loginResult.Data.AccessToken,
		RefreshToken: loginResult.Data.RefreshToken,
	}, nil
}

func isUserAlreadyExists(msg string) bool {
	lowerMsg := strings.ToLower(msg)
	return strings.Contains(lowerMsg, "已存在") ||
		strings.Contains(lowerMsg, "already exists") ||
		strings.Contains(lowerMsg, "duplicate")
}

func resetTestEnvironment() error {
	dsn, err := getConfigValue("DB_DSN")
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接测试数据库失败: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("测试数据库不可用: %w", err)
	}

	statements := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE `video_likes`",
		"TRUNCATE TABLE `comments`",
		"TRUNCATE TABLE `follows`",
		"TRUNCATE TABLE `videos`",
		"TRUNCATE TABLE `users`",
		"SET FOREIGN_KEY_CHECKS = 1",
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("执行 SQL 失败 [%s]: %w", stmt, err)
		}
	}

	if err := clearRedisCache(); err != nil {
		return err
	}

	return nil
}

func clearRedisCache() error {
	addr, err := getConfigValue("REDIS_ADDR")
	if err != nil {
		return err
	}

	password, _ := getConfigValueOptional("REDIS_PASSWORD")
	dbStr, _ := getConfigValueOptional("REDIS_DB")
	dbIndex := 0
	if strings.TrimSpace(dbStr) != "" {
		fmt.Sscanf(dbStr, "%d", &dbIndex)
	}

	cli := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIndex,
	})
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cli.Del(ctx, hotVideosKey).Err(); err != nil {
		return fmt.Errorf("清理 Redis 热榜缓存失败: %w", err)
	}
	return nil
}

func prepareRelationFixture(client *http.Client, baseURL string) (*RelationFixture, error) {
	users := []SeedUser{
		{Username: "relation_alice", Password: "RelationPass123"},
		{Username: "relation_bob", Password: "RelationPass123"},
		{Username: "relation_carol", Password: "RelationPass123"},
	}

	for i := range users {
		seedUser, err := loginSeedUser(client, baseURL, users[i].Username, users[i].Password)
		if err != nil {
			return nil, err
		}
		users[i] = *seedUser
	}

	// 首次准备社交测试数据时，可临时打开下面这段关注逻辑，插入一次后再继续保持注释状态。
	//
	// if err := ensureFollowRelation(client, baseURL, users[0].AccessToken, users[1].UserID, "Alice 关注 Bob"); err != nil {
	// 	return nil, err
	// }
	// if err := ensureFollowRelation(client, baseURL, users[1].AccessToken, users[0].UserID, "Bob 关注 Alice"); err != nil {
	// 	return nil, err
	// }
	// if err := ensureFollowRelation(client, baseURL, users[0].AccessToken, users[2].UserID, "Alice 关注 Carol"); err != nil {
	// 	return nil, err
	// }

	return &RelationFixture{
		Alice: users[0],
		Bob:   users[1],
		Carol: users[2],
	}, nil
}

func ensureFollowRelation(client *http.Client, baseURL, token, toUserID, scene string) error {
	res := testFollowUser(client, baseURL, token, toUserID)
	if res.Err != nil {
		return fmt.Errorf("预置 %s 失败: %w", scene, res.Err)
	}
	if res.Data.Base.Code != 0 && !isRelationAlreadyExists(res.Data.Base.Msg) {
		return fmt.Errorf("预置 %s 失败: %s", scene, res.Data.Base.Msg)
	}
	return nil
}

func isRelationAlreadyExists(msg string) bool {
	lowerMsg := strings.ToLower(msg)
	return strings.Contains(lowerMsg, "已关注") ||
		strings.Contains(lowerMsg, "already follow") ||
		strings.Contains(lowerMsg, "duplicate")
}

func getConfigValue(key string) (string, error) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value, nil
	}
	if value, ok := loadValueFromEnvFile(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("未找到配置 %s，请先设置环境变量或 video-platform/.env", key)
}

func getConfigValueOptional(key string) (string, bool) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value, true
	}
	return loadValueFromEnvFile(key)
}

func loadValueFromEnvFile(key string) (string, bool) {
	candidates := []string{
		"../video-platform/.env",
		".env",
	}

	for _, path := range candidates {
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			idx := strings.Index(line, "=")
			if idx <= 0 {
				continue
			}

			k := strings.TrimSpace(line[:idx])
			if k != key {
				continue
			}

			v := strings.TrimSpace(line[idx+1:])
			v = strings.Trim(v, `"'`)
			_ = file.Close()
			return v, true
		}
		_ = file.Close()
	}

	return "", false
}
