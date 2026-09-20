package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// ============ 角色长期记忆服务 ============
// 对话结束后异步(goroutine)由 LLM 判断本轮是否产生值得长期记住的访客事实,
// 命中则按 用户+角色 落库 persona_memories(≤20条,超限删最旧,重复跳过)。
// 任何失败只打日志,绝不影响对话响应。

const (
	// personaMemoryMax 每位访客对每个角色的记忆条数上限。
	personaMemoryMax = 20
	// personaMemoryContentMax 单条记忆内容长度上限(与表字段 varchar(500) 对齐)。
	personaMemoryContentMax = 500
)

// PersonaMemoryService 角色长期记忆服务。
type PersonaMemoryService struct {
	llm  *TripLLM
	repo *repository.ChatRepo
}

// NewPersonaMemoryService 构造角色长期记忆服务。
func NewPersonaMemoryService(llm *TripLLM, repo *repository.ChatRepo) *PersonaMemoryService {
	return &PersonaMemoryService{llm: llm, repo: repo}
}

// ExtractMemoryAsync 异步提取本轮对话中的新访客事实并落库。
// 内部起 goroutine,立即返回;LLM 失败/解析失败/无新事实均静默(仅日志)。
func (s *PersonaMemoryService) ExtractMemoryAsync(userID, personaID, userMsg, assistantMsg string) {
	if userID == "" || personaID == "" {
		return
	}
	userMsg = strings.TrimSpace(userMsg)
	assistantMsg = strings.TrimSpace(assistantMsg)
	if userMsg == "" && assistantMsg == "" {
		return
	}
	go func() {
		// 脱离请求上下文:SSE 响应结束后提取仍需继续,超时由 LLM 超时兜底
		ctx, cancel := context.WithTimeout(context.Background(),
			time.Duration(s.llm.settings.LLMTimeoutSeconds())*time.Second)
		defer cancel()
		s.extractMemory(ctx, userID, personaID, userMsg, assistantMsg)
	}()
}

// extractMemory 同步执行记忆提取:查现有记忆 → LLM 判断 → 去重 → 落库(超限删最旧)。
func (s *PersonaMemoryService) extractMemory(ctx context.Context, userID, personaID, userMsg, assistantMsg string) {
	existing, err := s.repo.ListMemories(userID, personaID, personaMemoryMax)
	if err != nil {
		logger.Warnf("[PersonaMemory] 记忆查询失败 user_id=%s persona=%s: %v", userID, personaID, err)
		return
	}

	systemPrompt := `你是记忆管理助手。你将得到一份"与同一位访客的历史记忆清单"和"本轮对话"。
请判断本轮对话是否包含值得长期记住的访客事实(偏好、背景、关注点、同行人情况等,
例如"带小孩出行""对杜甫诗歌特别感兴趣""正在筹备成都三日亲子游")。
规则:
1. 只提取稳定、可复用的事实;一次性问候、闲聊、寒暄一律不记。
2. 与历史记忆清单中内容重复或语义相同的,不要再输出。
3. 只输出一个 JSON 对象,不要任何解释或代码块标记:
   有新事实时 {"memory":"访客事实(50字内)"}
   无新事实时 {"memory":""}`

	userPrompt := "【历史记忆清单】\n"
	if len(existing) == 0 {
		userPrompt += "（暂无历史记忆）"
	}
	for _, m := range existing {
		userPrompt += "- " + m.Content + "\n"
	}
	userPrompt += "\n【本轮对话】\n访客：" + userMsg + "\n角色：" + assistantMsg

	reply, err := s.llm.ChatWithEffort(ctx, s.llm.settings.LLMTimeoutSeconds(),
		[]llmMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: userPrompt}},
		0.2, 256, llmThinkingLevel())
	if err != nil {
		logger.Warnf("[PersonaMemory] 记忆提取 LLM 调用失败 user_id=%s persona=%s: %v", userID, personaID, err)
		return
	}

	// 解析失败静默:不重试、不报错,只留一条日志便于排查
	var out struct {
		Memory string `json:"memory"`
	}
	if err := json.Unmarshal([]byte(extractJSONObject(reply)), &out); err != nil {
		logger.Warnf("[PersonaMemory] 记忆提取结果解析失败 user_id=%s persona=%s reply=%q", userID, personaID, reply)
		return
	}
	memory := strings.TrimSpace(out.Memory)
	if memory == "" {
		return // 无新事实
	}
	if len([]rune(memory)) > personaMemoryContentMax {
		memory = string([]rune(memory)[:personaMemoryContentMax])
	}
	// 去重:与现有记忆完全相同则跳过
	for _, m := range existing {
		if strings.TrimSpace(m.Content) == memory {
			return
		}
	}
	if err := s.repo.InsertMemory(&model.PersonaMemory{
		UserID:    userID,
		PersonaID: personaID,
		Content:   memory,
	}); err != nil {
		logger.Warnf("[PersonaMemory] 记忆落库失败 user_id=%s persona=%s: %v", userID, personaID, err)
		return
	}
	if err := s.repo.TrimMemories(userID, personaID, personaMemoryMax); err != nil {
		logger.Warnf("[PersonaMemory] 记忆条数修剪失败 user_id=%s persona=%s: %v", userID, personaID, err)
	}
	logger.Infof("[PersonaMemory] 已记录新访客事实 user_id=%s persona=%s: %s", userID, personaID, memory)
}

// MemoryPrompt 拼装"关于这位访客的记忆"提示段(每条一行,前缀"-")。
// userID 为空/查询失败/无记忆时返回空串,调用方跳过该段注入。
func (s *PersonaMemoryService) MemoryPrompt(userID, personaID string) string {
	if userID == "" {
		return ""
	}
	items, err := s.repo.ListMemories(userID, personaID, personaMemoryMax)
	if err != nil || len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for i, m := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- ")
		b.WriteString(m.Content)
	}
	return b.String()
}

// ClearMemories 清空当前用户某角色的全部记忆,返回清除条数。
func (s *PersonaMemoryService) ClearMemories(userID, personaID string) (int64, error) {
	return s.repo.ClearMemories(userID, personaID)
}
