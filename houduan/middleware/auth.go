// Package middleware 提供 Gin 中间件。
package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
)

// ContextKeyUID 注入到上下文中的用户 ID 键。
const ContextKeyUID = "uid"

// Auth 认证中间件:校验 Authorization 头中的 Bearer access Token,
// 并通过 Redis 会话校验其有效性。validate 由上层注入,避免依赖反向。
func Auth(validate func(ctx context.Context, token string) (uint, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			pkg.Unauthorized(c, "未提供访问令牌")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		uid, err := validate(c.Request.Context(), token)
		if err != nil {
			pkg.Unauthorized(c, "访问令牌无效或已过期")
			c.Abort()
			return
		}
		c.Set(ContextKeyUID, uid)
		c.Next()
	}
}

// GetUID 从上下文取当前登录用户 ID。
func GetUID(c *gin.Context) uint {
	if v, ok := c.Get(ContextKeyUID); ok {
		if uid, ok := v.(uint); ok {
			return uid
		}
	}
	return 0
}