package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"wenlv-backend/middleware"
	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// ChatSessionHandler AI 对话会话与角色长期记忆管理接口(仅登录用户)。
type ChatSessionHandler struct {
	sessions *service.ChatSessionService
	memory   *service.PersonaMemoryService
}

// NewChatSessionHandler 构造会话管理处理器。
func NewChatSessionHandler(sessions *service.ChatSessionService, memory *service.PersonaMemoryService) *ChatSessionHandler {
	return &ChatSessionHandler{sessions: sessions, memory: memory}
}

// List 当前用户某角色的会话列表:GET /api/chat/sessions?persona_id=du-fu
// 返回 data:[{id,title,updated_at}],按 updated_at 倒序,限 30 条。
func (h *ChatSessionHandler) List(c *gin.Context) {
	userID := formatUserID(middleware.GetUID(c))
	personaID := strings.TrimSpace(c.Query("persona_id"))
	if personaID == "" {
		pkg.BadRequest(c, "参数错误:persona_id 不能为空")
		return
	}
	items, err := h.sessions.ListSessions(userID, personaID)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "查询会话列表失败")
		return
	}
	pkg.OK(c, items)
}

// Create 创建会话:POST /api/chat/sessions
// 入参 {persona_id,language},返回 data:{id}(title 初始为空,首条消息时回填)。
func (h *ChatSessionHandler) Create(c *gin.Context) {
	var req struct {
		PersonaID string `json:"persona_id"`
		Language  string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PersonaID) == "" {
		pkg.BadRequest(c, "参数错误:persona_id 必填")
		return
	}
	userID := formatUserID(middleware.GetUID(c))
	id, err := h.sessions.CreateSession(userID, strings.TrimSpace(req.PersonaID), req.Language)
	if err != nil {
		if errors.Is(err, service.ErrInvalidChatPersona) {
			pkg.BadRequest(c, err.Error())
			return
		}
		pkg.ServerErrorWithErr(c, err, "创建会话失败")
		return
	}
	pkg.OK(c, gin.H{"id": id})
}

// Messages 会话历史消息:GET /api/chat/sessions/:id/messages
// 校验会话归属当前用户,返回 data:[{role,content,created_at}] 按时间正序(限最近 100 条)。
func (h *ChatSessionHandler) Messages(c *gin.Context) {
	userID := formatUserID(middleware.GetUID(c))
	sessionID, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "参数错误:会话ID不合法")
		return
	}
	msgs, err := h.sessions.SessionMessages(userID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			pkg.Fail(c, 404, 404, "会话不存在")
		case errors.Is(err, service.ErrSessionDenied):
			pkg.Fail(c, 403, 403, "无权访问该会话")
		default:
			pkg.ServerErrorWithErr(c, err, "查询会话消息失败")
		}
		return
	}
	pkg.OK(c, msgs)
}

// ClearMemories 清除当前用户某角色的全部长期记忆:DELETE /api/chat/memories?persona_id=du-fu
// 返回 data:{cleared:n}。
func (h *ChatSessionHandler) ClearMemories(c *gin.Context) {
	userID := formatUserID(middleware.GetUID(c))
	personaID := strings.TrimSpace(c.Query("persona_id"))
	if personaID == "" {
		pkg.BadRequest(c, "参数错误:persona_id 不能为空")
		return
	}
	cleared, err := h.memory.ClearMemories(userID, personaID)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "清除记忆失败")
		return
	}
	pkg.OK(c, gin.H{"cleared": cleared})
}
