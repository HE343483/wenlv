package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"wenlv-backend/logger"
	"wenlv-backend/middleware"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// PersonaChatHandler 历史人物 AI 角色对话接口(杜甫/诸葛亮)。
type PersonaChatHandler struct {
	svc      *service.PersonaChatService
	sessions *service.ChatSessionService
	memory   *service.PersonaMemoryService
}

// NewPersonaChatHandler 构造处理器。
func NewPersonaChatHandler(svc *service.PersonaChatService, sessions *service.ChatSessionService, memory *service.PersonaMemoryService) *PersonaChatHandler {
	return &PersonaChatHandler{svc: svc, sessions: sessions, memory: memory}
}

// maxPersonaMessages 单次请求携带的历史消息上限(超出截尾,防大 payload)。
const maxPersonaMessages = 40

// maxPersonaContentLen 单条消息内容长度上限(字符)。
const maxPersonaContentLen = 2000

// Personas 角色卡片列表:GET /api/trip/personas
func (h *PersonaChatHandler) Personas(c *gin.Context) {
	pkg.OK(c, h.svc.Personas())
}

// Chat 与历史人物对话(SSE 流式):POST /api/trip/persona-chat
// 入参 {persona_id, messages:[{role,content}], language, session_id, memory_enabled};校验阶段返回 JSON,通过后切换 SSE:
// 事件格式 data: {"delta":"..."} / data: {"error":"..."} / data: [DONE],与 /api/chat/ask/stream 一致。
// 仅登录用户持久化:携带合法 session_id 且会话归属当前用户时,流结束前落库本轮 user+assistant 消息;
// memory_enabled=true 时异步提取角色长期记忆,失败只打日志不影响响应。
func (h *PersonaChatHandler) Chat(c *gin.Context) {
	var req model.PersonaChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误:persona_id 与 messages 必填")
		return
	}
	if personaID := strings.TrimSpace(req.PersonaID); personaID == "" {
		pkg.BadRequest(c, "参数错误:persona_id 不能为空")
		return
	} else {
		req.PersonaID = personaID
	}
	if len(req.Messages) == 0 {
		pkg.BadRequest(c, "参数错误:messages 不能为空")
		return
	}
	if len(req.Messages) > maxPersonaMessages {
		req.Messages = req.Messages[len(req.Messages)-maxPersonaMessages:]
	}
	for i := range req.Messages {
		req.Messages[i].Content = truncateRuneText(req.Messages[i].Content, maxPersonaContentLen)
	}
	// 角色合法性提前校验(避免 SSE 已开始后才报错)
	if err := h.svc.ValidatePersona(req.PersonaID); err != nil {
		pkg.BadRequest(c, err.Error())
		return
	}

	// 登录用户标识(0=未登录,不落库、不读记忆)
	uid := middleware.GetUID(c)
	userID := formatUserID(uid)

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 避免 Nginx 反向代理缓冲 SSE

	writeEvent := func(payload string) {
		fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
		c.Writer.Flush()
	}

	var full strings.Builder
	err := h.svc.ChatStream(c.Request.Context(), req.PersonaID, req.Language, req.Messages, userID, req.MemoryEnabled, func(delta string) {
		full.WriteString(delta)
		data, _ := json.Marshal(map[string]string{"delta": delta})
		writeEvent(string(data))
	})
	if err != nil {
		if errors.Is(err, service.ErrUnknownPersona) {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			writeEvent(string(data))
			return
		}
		data, _ := json.Marshal(map[string]string{"error": "角色对话服务异常: " + err.Error()})
		writeEvent(string(data))
		return
	}

	// SSE 结束前持久化本轮对话(仅登录用户且带 session_id);失败只打日志,不影响 SSE 结果
	if userID != "" && req.SessionID > 0 {
		userMsg := lastUserChatMessage(req.Messages)
		if err := h.sessions.AppendTurn(userID, req.SessionID, req.PersonaID, userMsg, full.String()); err != nil {
			logger.Warnf("[ChatSession] 角色对话落库失败 user_id=%s session_id=%d persona=%s: %v", userID, req.SessionID, req.PersonaID, err)
		} else if req.MemoryEnabled {
			// 记忆提取异步执行,绝不阻塞响应
			h.memory.ExtractMemoryAsync(userID, req.PersonaID, userMsg, full.String())
		}
	}
	writeEvent("[DONE]")
}

// lastUserChatMessage 取消息列表中最后一条用户消息(即本次提问)。
func lastUserChatMessage(msgs []model.ChatMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != "assistant" && strings.TrimSpace(msgs[i].Content) != "" {
			return msgs[i].Content
		}
	}
	return ""
}

// formatUserID 登录用户 ID 转字符串(存储层 user_id 为 varchar);未登录返回空串。
func formatUserID(uid uint) string {
	if uid == 0 {
		return ""
	}
	return fmt.Sprintf("%d", uid)
}

// truncateRuneText 按 rune 截断文本,超长时省略。
func truncateRuneText(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
