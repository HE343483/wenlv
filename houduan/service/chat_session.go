package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// ============ AI 对话会话持久化服务 ============
// 普通智能助手(persona_id='assistant')与杜甫/诸葛亮同一套表统一存储;
// 仅登录用户持久化:userID 为 users.id 的字符串形式,空串视为未登录。

const (
	// AssistantPersonaID 普通智能助手在会话存储中的角色 ID。
	AssistantPersonaID = "assistant"
	// sessionListLimit 会话列表条数上限。
	sessionListLimit = 30
	// chatMessageListLimit 单会话历史消息返回条数上限(最近 N 条)。
	chatMessageListLimit = 100
	// sessionTitleMaxRunes 会话标题长度(首条用户消息前 30 字)。
	sessionTitleMaxRunes = 30
)

// 会话操作错误(handler 据此映射 HTTP 状态码)。
var (
	// ErrSessionNotFound 会话不存在。
	ErrSessionNotFound = errors.New("会话不存在")
	// ErrSessionDenied 会话不属于当前用户。
	ErrSessionDenied = errors.New("无权访问该会话")
	// ErrInvalidChatPersona 角色ID不合法(assistant 或已注册历史人物之外)。
	ErrInvalidChatPersona = errors.New("未知角色")
)

// ChatSessionService 对话会话持久化服务。
type ChatSessionService struct {
	repo *repository.ChatRepo
}

// NewChatSessionService 构造会话持久化服务。
func NewChatSessionService(repo *repository.ChatRepo) *ChatSessionService {
	return &ChatSessionService{repo: repo}
}

// ValidateChatPersonaID 校验会话可用的角色 ID:'assistant' 或已注册历史人物。
func ValidateChatPersonaID(personaID string) error {
	if personaID == AssistantPersonaID {
		return nil
	}
	if personaByID(personaID) == nil {
		return ErrInvalidChatPersona
	}
	return nil
}

// CreateSession 创建会话,返回会话 ID(title 初始为空,首轮对话时回填)。
func (s *ChatSessionService) CreateSession(userID, personaID, lang string) (uint, error) {
	if err := ValidateChatPersonaID(personaID); err != nil {
		return 0, err
	}
	session := &model.ChatSession{
		UserID:    userID,
		PersonaID: personaID,
		Lang:      model.NormalizeLang(lang),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return 0, err
	}
	return session.ID, nil
}

// SessionSummary 会话列表项(响应字段严格为 id/title/updated_at)。
type SessionSummary struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
}

// ListSessions 当前用户某角色的会话列表(按 updated_at 倒序,限 30 条)。
func (s *ChatSessionService) ListSessions(userID, personaID string) ([]SessionSummary, error) {
	items, err := s.repo.ListSessions(userID, personaID, sessionListLimit)
	if err != nil {
		return nil, err
	}
	out := make([]SessionSummary, 0, len(items))
	for _, it := range items {
		out = append(out, SessionSummary{
			ID:        it.ID,
			Title:     it.Title,
			UpdatedAt: it.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}

// MessageItem 会话历史消息项(响应字段严格为 role/content/created_at)。
type MessageItem struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// SessionMessages 校验会话归属后返回历史消息(时间正序,限最近 100 条)。
func (s *ChatSessionService) SessionMessages(userID string, sessionID uint) ([]MessageItem, error) {
	session, err := s.repo.FindSession(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if session.UserID != userID {
		return nil, ErrSessionDenied
	}
	msgs, err := s.repo.ListMessages(sessionID, chatMessageListLimit)
	if err != nil {
		return nil, err
	}
	out := make([]MessageItem, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, MessageItem{
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}

// AppendTurn 会话追加一轮对话:校验会话归属当前用户 → 落库 user/assistant 消息;
// 会话标题为空时用本轮用户消息前 30 字回填。任何失败由调用方决定仅记日志。
func (s *ChatSessionService) AppendTurn(userID string, sessionID uint, personaID, userMsg, assistantMsg string) error {
	if userID == "" || sessionID == 0 {
		return nil // 未登录或未带会话:静默跳过
	}
	session, err := s.repo.FindSession(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	if session.UserID != userID {
		return ErrSessionDenied
	}
	return s.repo.AppendTurn(session, userMsg, assistantMsg, sessionTitle(userMsg))
}

// sessionTitle 取首条用户消息前 30 字作为会话标题。
func sessionTitle(userMsg string) string {
	runes := []rune(strings.TrimSpace(userMsg))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) > sessionTitleMaxRunes {
		runes = runes[:sessionTitleMaxRunes]
	}
	return string(runes)
}
