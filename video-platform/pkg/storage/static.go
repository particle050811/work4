package storage

import (
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// BindStatic 挂载静态文件目录
func BindStatic(h *server.Hertz) {
	// 头像目录：/storage/avatars/<filename>
	if err := os.MkdirAll("./storage/avatars", 0o755); err != nil {
		log.Printf("创建头像目录失败: %v", err)
	}
	h.StaticFS("/storage/avatars", &app.FS{
		Root:        "./storage/avatars",
		PathRewrite: app.NewPathSlashesStripper(2),
	})

	// 视频目录：/storage/videos/<filename>
	if err := os.MkdirAll("./storage/videos", 0o755); err != nil {
		log.Printf("创建视频目录失败: %v", err)
	}
	h.StaticFS("/storage/videos", &app.FS{
		Root:        "./storage/videos",
		PathRewrite: app.NewPathSlashesStripper(2),
	})
}
