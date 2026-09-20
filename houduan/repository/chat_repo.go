package repository

import (
	"time"

	"gorm.io/gorm"

	"wenlv-backend/model"
)

// ChatRepo AI 对话上下文持久化数据访问(chat_sessions / chat_messages / persona_memories)。
type ChatRepo struct {
	db *gorm.DB
}

// NewChatRepo 构造对话持久化仓储。
func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

// CreateSession 新建会话(title 初始为空,首轮对话时回填)。
func (r *ChatRepo) CreateSession(s *model.ChatSession) error {
	return r.db.Create(s).Error
}

// FindSession 按 ID 查会话(不存在返回 gorm.ErrRecordNotFound)。
func (r *ChatRepo) FindSession(id uint) (*model.ChatSession, error) {
	var s model.ChatSession
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ListSessions 某用户某角色的会话列表,按 updated_at 倒序。
func (r *ChatRepo) ListSessions(userID, personaID string, limit int) ([]model.ChatSession, error) {
	var items []model.ChatSession
	if err := r.db.Where("user_id = ? AND persona_id = ?", userID, personaID).
		Order("updated_at DESC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// AppendTurn 会话追加一轮对话(事务):写入 user/assistant 两条消息;
// 会话标题为空且 newTitle 非空时回填;每次都刷新 updated_at 供会话列表排序。
func (r *ChatRepo) AppendTurn(session *model.ChatSession, userMsg, assistantMsg, newTitle string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		msgs := []model.ChatMessageRecord{
			{SessionID: session.ID, Role: "user", Content: userMsg},
			{SessionID: session.ID, Role: "assistant", Content: assistantMsg},
		}
		if err := tx.Create(&msgs).Error; err != nil {
			return err
		}
		updates := map[string]any{"updated_at": time.Now()}
		if session.Title == "" && newTitle != "" {
			updates["title"] = newTitle
		}
		return tx.Model(&model.ChatSession{}).Where("id = ?", session.ID).Updates(updates).Error
	})
}

// ListMessages 取会话内最近 limit 条消息,按时间正序返回(倒序取 limit 后反转)。
func (r *ChatRepo) ListMessages(sessionID uint, limit int) ([]model.ChatMessageRecord, error) {
	var items []model.ChatMessageRecord
	if err := r.db.Where("session_id = ?", sessionID).
		Order("id DESC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items, nil
}

// ListMemories 某用户某角色的长期记忆,按写入先后正序,最多 limit 条。
func (r *ChatRepo) ListMemories(userID, personaID string, limit int) ([]model.PersonaMemory, error) {
	var items []model.PersonaMemory
	if err := r.db.Where("user_id = ? AND persona_id = ?", userID, personaID).
		Order("id ASC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// InsertMemory 新增一条记忆。
func (r *ChatRepo) InsertMemory(m *model.PersonaMemory) error {
	return r.db.Create(m).Error
}

// TrimMemories 记忆超限时删除最旧(保留最新 max 条)。
func (r *ChatRepo) TrimMemories(userID, personaID string, max int) error {
	var ids []uint
	if err := r.db.Model(&model.PersonaMemory{}).
		Where("user_id = ? AND persona_id = ?", userID, personaID).
		Order("id ASC").Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) <= max {
		return nil
	}
	stale := ids[:len(ids)-max]
	return r.db.Where("id IN ?", stale).Delete(&model.PersonaMemory{}).Error
}

// ClearMemories 清空某用户某角色全部记忆,返回删除条数。
func (r *ChatRepo) ClearMemories(userID, personaID string) (int64, error) {
	res := r.db.Where("user_id = ? AND persona_id = ?", userID, personaID).Delete(&model.PersonaMemory{})
	return res.RowsAffected, res.Error
}
