// Package pkg 提供跨层复用的通用工具。
package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wenlv-backend/logger"
)

// ContextKeyRequestID 请求上下文中存放请求 ID 的键(由 middleware.RequestID 写入)。
const ContextKeyRequestID = "request_id"

// Response 统一响应体。code=0 表示成功,非 0 为业务错误码。
// ErrorID 仅在服务端报错时返回,与 logs/error-*.log 中的错误 ID 一致,便于按 ID 定位。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	ErrorID string      `json:"error_id,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Fail 业务错误响应。
func Fail(c *gin.Context, httpStatus, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Message: msg})
}

// ServerError 服务端内部错误:落一条带错误 ID 的报错日志,并把 ID 回传给前端。
func ServerError(c *gin.Context, msg string) {
	replyServerError(c, nil, msg)
}

// ServerErrorWithErr 服务端内部错误(带原始 error),日志中会记录 err 详情。
func ServerErrorWithErr(c *gin.Context, err error, msg string) {
	replyServerError(c, err, msg)
}

// replyServerError 统一记录并返回 500。
func replyServerError(c *gin.Context, err error, msg string) {
	id := logger.With(
		"request_id", c.GetString(ContextKeyRequestID),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	).ErrorSev(logger.SeverityMedium, err, "%s", msg)

	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: msg,
		ErrorID: id,
	})
}

// BadRequest 参数/请求错误。
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 400, msg)
}

// Unauthorized 未认证/凭据失效。
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 401, msg)
}