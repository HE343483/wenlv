package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"wenlv-backend/logger"
	"wenlv-backend/pkg"
)

// 请求 ID 当日序号(进程内自增,配合日期保证同日唯一)。
var reqSeq uint64

// RequestID 为每个请求生成带日期的请求 ID,写入上下文与响应头 X-Request-Id,
// 便于把同一次请求的访问日志与报错日志串起来。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := fmt.Sprintf("REQ-%s-%05d", time.Now().Format("20060102"), atomic.AddUint64(&reqSeq, 1))
		c.Set(pkg.ContextKeyRequestID, id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

// AccessLogger 记录 HTTP 访问日志(方法/路径/状态码/耗时/来源 IP),写入 logs/access-*.log。
func AccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// 图片代理请求量大且无排查价值,不记录,避免刷屏
		if strings.HasPrefix(path, "/api/poi/image") {
			return
		}
		target := path
		if query != "" {
			target += "?" + query
		}
		logger.With(
			"request_id", RequestIDOf(c),
			"ip", c.ClientIP(),
			"status", c.Writer.Status(),
			"cost", fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
			"size", c.Writer.Size(),
		).Accessf("%s %s", c.Request.Method, target)
	}
}

// Recovery 捕获 panic:写入带错误 ID 的 CRITICAL 报错(含调用栈),并返回 500,
// 响应体附带 error_id,便于直接拿 ID 去 logs/error-*.log 里定位。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			brokenPipe := isBrokenPipe(r)
			id := logger.With(
				"request_id", RequestIDOf(c),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			).ErrorSev(logger.SeverityCritical, nil, "HTTP 处理发生 panic: %v\n%s", r, string(debug.Stack()))

			if brokenPipe {
				// 连接已断开,无法再写响应
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, pkg.Response{
				Code:    http.StatusInternalServerError,
				Message: "服务内部错误",
				ErrorID: id,
			})
		}()
		c.Next()
	}
}

// RequestIDOf 读取当前请求的请求 ID,未启用中间件时返回空串。
func RequestIDOf(c *gin.Context) string {
	return c.GetString(pkg.ContextKeyRequestID)
}

// isBrokenPipe 判断 panic 是否由客户端断开连接引起(此类情况无法再写响应)。
func isBrokenPipe(r any) bool {
	ne, ok := r.(*net.OpError)
	if !ok {
		return false
	}
	se, ok := ne.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	msg := strings.ToLower(se.Error())
	return strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer")
}
