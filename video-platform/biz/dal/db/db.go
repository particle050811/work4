package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitMySQL 初始化 MySQL 连接
func InitMySQL() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN 环境变量未设置，请检查 .env 文件")
	}

	// 自动创建数据库（如果不存在）
	createDatabaseIfNotExists(dsn)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 获取底层 sql.DB 并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("[MySQL] 连接成功")
	return db
}

// createDatabaseIfNotExists 自动创建数据库
func createDatabaseIfNotExists(dsn string) {
	// 从 DSN 中提取数据库名和无数据库的 DSN
	// DSN 格式: user:pass@tcp(host:port)/dbname?params
	idx := strings.LastIndex(dsn, "/")
	if idx == -1 {
		return
	}

	baseDSN := dsn[:idx] + "/"
	dbPart := dsn[idx+1:]

	// 提取数据库名（去掉参数部分）
	dbName := dbPart
	if paramIdx := strings.Index(dbPart, "?"); paramIdx != -1 {
		dbName = dbPart[:paramIdx]
	}

	// 连接到 MySQL（不指定数据库）
	db, err := sql.Open("mysql", baseDSN)
	if err != nil {
		log.Printf("连接 MySQL 失败: %v", err)
		return
	}
	defer db.Close()

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		log.Printf("创建数据库失败: %v", err)
		return
	}

	log.Printf("[MySQL] 数据库 %s 已就绪", dbName)
}
