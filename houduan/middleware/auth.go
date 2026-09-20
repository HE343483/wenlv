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

// OptionalAuth 可选认证中间件:请求带合法 Bearer Token 时把 uid 注入上下文,
// 缺失或无效时放行(uid=0),不拦截请求。用于"未登录可匿名使用、登录后享有持久化"的接口
// (如 AI 角色对话 SSE:登录用户落库对话与长期记忆,游客仅聊天)。
func OptionalAuth(validate func(ctx context.Context, token string) (uint, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			token := strings.TrimPrefix(header, "Bearer ")
			if uid, err := validate(c.Request.Context(), token); err == nil && uid > 0 {
				c.Set(ContextKeyUID, uid)
			}
		}
		c.Next()
	}
}

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