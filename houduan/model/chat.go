package model

import "time"

// ============ AI 对话上下文持久化(普通智能助手与历史人物角色统一存储) ============
// 仅登录用户落库:未登录照旧可聊,但不落库、无历史会话。
// persona_id='assistant' 为普通智能问答,du-fu/zhuge-liang 为历史人物角色对话。

// ChatSession 对话会话:一个用户与一个角色的一条会话线程。
type ChatSession struct {
	ID uint `gorm:"primaryKey;comment:会话ID" json:"id"`
	// UserID 登录用户ID(users.id 的字符串形式),与 PersonaID 联合决定会话归属
	UserID    string `gorm:"type:varchar(64);index;comment:用户ID(登录用户,users.id字符串)" json:"user_id"`
	PersonaID string `gorm:"type:varchar(32);index;comment:角色ID(assistant/du-fu/zhuge-liang)" json:"persona_id"`
	// Title 会话标题:首条用户消息前30字,创建时为空,首轮对话时回填
	Title     string    `gorm:"type:varchar(120);comment:会话标题(首条用户消息前30字,首轮对话时回填)" json:"title"`
	Lang      string    `gorm:"type:varchar(16);comment:会话界面语言(zh/en/ja)" json:"lang"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间(每次追加消息刷新,会话列表按此倒序)" json:"updated_at"`
}

// TableName 固定表名,防止 GORM 复数化歧义。
func (ChatSession) TableName() string { return "chat_sessions" }

// ChatMessageRecord 会话内的单条消息(user/assistant)。
// 命名区别于 trip.go 中表示请求载荷的 ChatMessage。
type ChatMessageRecord struct {
	ID uint `gorm:"primaryKey;comment:消息ID" json:"id"`
	// SessionID 所属会话,索引加速按会话取消息
	SessionID uint   `gorm:"index;comment:所属会话ID(chat_sessions.id)" json:"session_id"`
	Role      string `gorm:"type:varchar(16);comment:消息角色(user/assistant)" json:"role"`
	Content   string `gorm:"type:text;comment:消息内容" json:"content"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}

// TableName 固定表名,防止 GORM 复数化歧义。
func (ChatMessageRecord) TableName() string { return "chat_messages" }

// PersonaMemory 角色长期记忆:LLM 从对话中异步提取的访客事实(偏好/背景/关注点)。
type PersonaMemory struct {
	ID uint `gorm:"primaryKey;comment:记忆ID" json:"id"`
	// UserID 登录用户ID,与 PersonaID 决定记忆归属(每位访客对每个角色独立记忆)
	UserID    string `gorm:"type:varchar(64);index;comment:用户ID(登录用户,users.id字符串)" json:"user_id"`
	PersonaID string `gorm:"type:varchar(32);index;comment:角色ID(assistant/du-fu/zhuge-liang)" json:"persona_id"`
	// Content 记忆内容(单条不超过500字),每人每角色最多保留20条,超限删最旧
	Content   string    `gorm:"type:varchar(500);comment:记忆内容(访客事实,如带小孩出行/对杜甫诗歌感兴趣)" json:"content"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}

// TableName 固定表名,防止 GORM 复数化歧义。
func (PersonaMemory) TableName() string { return "persona_memories" }
