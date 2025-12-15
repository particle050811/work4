package swagger

import (
	"context"
	_ "embed"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
)

//go:embed user/openapi.yaml
var userYAML []byte

//go:embed video/openapi.yaml
var videoYAML []byte

//go:embed interaction/openapi.yaml
var interactionYAML []byte

//go:embed relation/openapi.yaml
var relationYAML []byte

// BindSwagger 绑定所有模块的 Swagger UI
func BindSwagger(h *server.Hertz) {
	// 各模块独立的 Swagger UI
	// 用户模块: /swagger/user/index.html
	h.GET("/swagger/user/*any", swagger.WrapHandler(swaggerFiles.Handler, swagger.URL("/openapi/user.yaml")))
	// 视频模块: /swagger/video/index.html
	h.GET("/swagger/video/*any", swagger.WrapHandler(swaggerFiles.Handler, swagger.URL("/openapi/video.yaml")))
	// 互动模块: /swagger/interaction/index.html
	h.GET("/swagger/interaction/*any", swagger.WrapHandler(swaggerFiles.Handler, swagger.URL("/openapi/interaction.yaml")))
	// 社交模块: /swagger/relation/index.html
	h.GET("/swagger/relation/*any", swagger.WrapHandler(swaggerFiles.Handler, swagger.URL("/openapi/relation.yaml")))

	// OpenAPI YAML 文件
	h.GET("/openapi/user.yaml", func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Content-Type", "application/x-yaml")
		ctx.Write(userYAML)
	})
	h.GET("/openapi/video.yaml", func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Content-Type", "application/x-yaml")
		ctx.Write(videoYAML)
	})
	h.GET("/openapi/interaction.yaml", func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Content-Type", "application/x-yaml")
		ctx.Write(interactionYAML)
	})
	h.GET("/openapi/relation.yaml", func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Content-Type", "application/x-yaml")
		ctx.Write(relationYAML)
	})
}
