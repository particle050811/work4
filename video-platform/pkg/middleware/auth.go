package middleware

import (
	"context"
	"strings"

	"video-platform/pkg/auth"
	"video-platform/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 从 Authorization 头获取 Token
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"base": response.Unauthorized("请提供认证令牌"),
			})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"base": response.Unauthorized("认证令牌格式错误"),
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 Token
		jwtMgr := auth.GetJWTManager()
		claims, err := jwtMgr.ValidateAccessToken(tokenString)
		if err != nil {
			if err == auth.ErrTokenExpired {
				c.JSON(consts.StatusUnauthorized, map[string]interface{}{
					"base": response.Error(response.CodeTokenExpired),
				})
			} else {
				c.JSON(consts.StatusUnauthorized, map[string]interface{}{
					"base": response.Error(response.CodeTokenInvalid),
				})
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next(ctx)
	}
}

