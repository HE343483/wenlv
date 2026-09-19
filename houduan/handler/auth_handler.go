// Package handler 提供 HTTP 处理函数。
package handler

import (
	"github.com/gin-gonic/gin"

	"wenlv-backend/middleware"
	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// AuthHandler 认证相关接口。
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler 构造认证处理器。
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register 注册。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误:用户名必填,密码至少 6 位")
		return
	}
	if err := h.svc.Register(c.Request.Context(), req.Username, req.Password); err != nil {
		if err == service.ErrUsernameTaken {
			pkg.Fail(c, 422, 1001, err.Error())
			return
		}
		pkg.ServerErrorWithErr(c, err, "注册失败")
		return
	}
	pkg.OK(c, nil)
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResp struct {
	TokenType string            `json:"token_type"`
	User      map[string]any    `json:"user"`
	Tokens    service.TokenPair `json:"tokens"`
}

// Login 登录。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "用户名和密码必填")
		return
	}
	u, pair, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err == service.ErrCredential {
			pkg.Fail(c, 401, 1002, err.Error())
			return
		}
		pkg.ServerErrorWithErr(c, err, "登录失败")
		return
	}
	pkg.OK(c, loginResp{
		TokenType: "Bearer",
		User: map[string]any{
			"id":         u.ID,
			"username":   u.Username,
			"avatar_url": u.AvatarURL,
		},
		Tokens: *pair,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 刷新 Token。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "refresh_token 必填")
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrRefreshInvalid {
			pkg.Fail(c, 401, 1003, err.Error())
			return
		}
		pkg.ServerErrorWithErr(c, err, "刷新失败")
		return
	}
	pkg.OK(c, pair)
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout 登出,注销 access 与 refresh 会话。
func (h *AuthHandler) Logout(c *gin.Context) {
	accessJTI := ""
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		accessJTI, _ = h.svc.ParseAccessJTI(token[7:])
	}
	var req logoutReq
	_ = c.ShouldBindJSON(&req)
	refreshJTI := ""
	if req.RefreshToken != "" {
		refreshJTI, _ = h.svc.ParseRefreshJTI(req.RefreshToken)
	}
	if err := h.svc.Logout(c.Request.Context(), accessJTI, refreshJTI); err != nil {
		pkg.ServerErrorWithErr(c, err, "登出失败")
		return
	}
	pkg.OK(c, nil)
}

// Me 返回当前用户信息。
func (h *AuthHandler) Me(c *gin.Context) {
	uid := middleware.GetUID(c)
	u, err := h.svc.GetUser(c.Request.Context(), uid)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "获取用户失败")
		return
	}
	pkg.OK(c, map[string]any{
		"id":         u.ID,
		"username":   u.Username,
		"avatar_url": u.AvatarURL,
	})
}
