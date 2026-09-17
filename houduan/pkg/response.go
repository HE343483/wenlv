// Package pkg 提供跨层复用的通用工具。
package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应体。code=0 表示成功,非 0 为业务错误码。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Fail 业务错误响应。
func Fail(c *gin.Context, httpStatus, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Message: msg})
}

// ServerError 服务端内部错误。
func ServerError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 500, msg)
}

// BadRequest 参数/请求错误。
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 400, msg)
}

// Unauthorized 未认证/凭据失效。
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 401, msg)
}