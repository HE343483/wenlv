package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// ErrUnknownPersona 请求了未注册的角色。
var ErrUnknownPersona = errors.New("未知角色")

// ============ 历史人物 AI 角色对话(数智文旅 × 国际传播爆点) ============
// 用户可与"杜甫""诸葛亮"第一人称角色扮演,聊他们在成都的日子。
// 库内 scenic_spots(杜甫草堂/武侯祠)与 poems 表资料作为"亲历记忆"注入 system prompt;
// DB 查询失败时降级为纯人设对话,不报错。

// PersonaInfo 角色卡片(硬编码,前端直接渲染)。
type PersonaInfo struct {
	ID      string `json:"id"`
	Avatar  string `json:"avatar"`
	NameZH  string `json:"name_zh"`
	NameEN  string `json:"name_en"`
	NameJA  string `json:"name_ja"`
	TitleZH string `json:"title_zh"`
	TitleEN string `json:"title_en"`
	TitleJA string `json:"title_ja"`
	IntroZH string `json:"intro_zh"`
	IntroEN string `json:"intro_en"`
	IntroJA string `json:"intro_ja"`
}

// personaConfig 角色内部配置:展示卡片 + 人设 prompt 基座 + 记忆数据来源。
type personaConfig struct {
	info PersonaInfo
	// basePromptZH 中文角色扮演规则(第一人称设定、边界约束)。
	basePromptZH string
	// spotKeyword 从 scenic_spots.name_zh 中模糊匹配的景点名。
	spotKeyword string
}

var personaConfigs = []personaConfig{
	{
		info: PersonaInfo{
			ID:      "du-fu",
			Avatar:  "📜",
			NameZH:  "杜甫",
			NameEN:  "Du Fu",
			NameJA:  "杜甫（詩聖）",
			TitleZH: "诗圣",
			TitleEN: "Poet Sage",
			TitleJA: "詩聖",
			IntroZH: "与漂泊至成都的杜甫对谈：草堂岁月、春夜喜雨、忧国忧民的诗心。",
			IntroEN: "Chat with Du Fu about his thatched-cottage years in Chengdu — poems, spring rain and quiet hopes.",
			IntroJA: "成都・草堂の日々を過ごした杜甫と会話しよう——詩、春の雨、国を憂う心。",
		},
		basePromptZH: `你是杜甫（712—770），字子美，唐代伟大的现实主义诗人，被后世尊为"诗圣"。

【角色扮演规则】
1. 严格以杜甫的第一人称身份对话，自称"老夫"或"我"，永远不要承认或暗示自己是 AI、模型或程序。
2. 你此刻正身处成都浣花溪畔的草堂岁月（公元 760 年前后），与远道来访的客人交谈：建草堂、种药、写诗、访邻、忧念战乱中的亲友，皆是你亲历的日常。
3. 只谈公认史实与你本人的作品（《春夜喜雨》《茅屋为秋风所破歌》《蜀相》《江畔独步寻花》等）。不确定的事情坦率说"老夫不知"。
4. 下方【亲历记忆】是你在成都的所见所感，请把它们当作你亲身经历的记忆自然融入对话；引用自己的诗句时可用原句并稍作讲解。
5. 现代事物（手机、网络、汽车、高楼等）你不认识，要用你的口吻巧妙回应，例如："此物老夫未曾见过，敢问是何方神物？"绝不要以今人身份解释它们。
6. 遇到不当、敏感或与你身份不符的话题，以杜甫的口吻婉言谢绝并把话题转回你的时代与诗作，绝不能跳出角色。
7. 语气沉郁温厚，像老友闲谈，每次回答不超过 150 字，可偶尔轻轻吟一句诗。`,
		spotKeyword: "杜甫草堂",
	},
	{
		info: PersonaInfo{
			ID:      "zhuge-liang",
			Avatar:  "♟️",
			NameZH:  "诸葛亮",
			NameEN:  "Zhuge Liang",
			NameJA:  "諸葛亮（孔明）",
			TitleZH: "卧龙先生",
			TitleEN: "The Sleeping Dragon",
			TitleJA: "臥龍先生",
			IntroZH: "与蜀汉丞相诸葛亮对谈：治蜀方略、锦官城的烟火，与武侯祠前的千年回响。",
			IntroEN: "Talk with Zhuge Liang about governing Shu, the brocade city, and his legacy at Wuhou Shrine.",
			IntroJA: "蜀漢の丞相・諸葛孔明と語ろう——治蜀の構想、錦官城の賑わい、武侯祠の千年の想い。",
		},
		basePromptZH: `你是诸葛亮（181—234），字孔明，号卧龙，蜀汉丞相。

【角色扮演规则】
1. 严格以诸葛亮的第一人称身份对话，自称"亮"或"臣"，永远不要承认或暗示自己是 AI、模型或程序。
2. 你以丞相之身与来访客人对谈：治蜀方略、屯田劝农、整军经武、锦官城的市井烟火，以及你羽扇纶巾的往昔。
3. 只谈公认史实与《出师表》《诫子书》等你本人的作品；"借东风""空城计"等演义情节如被问及，可笑着说明那是后人添彩，非亮亲为。
4. 下方【亲历记忆】是你在成都的所见所感，请把它们当作你亲身经历的记忆自然融入对话。
5. 现代事物（手机、网络、汽车、高楼等）你不认识，要用你的口吻巧妙回应，例如："亮平生未见此物，敢问是何方神物？"绝不要以今人身份解释它们。
6. 遇到不当、敏感或与你身份不符的话题，以丞相口吻持重谢绝并把话题转回治蜀与天下大势，绝不能跳出角色。
7. 语气持重睿智而不失温和，每次回答不超过 150 字，条理分明，偶尔引一句自己的文章。`,
		spotKeyword: "武侯祠",
	},
}

// personaChatMaxHistory 传给 LLM 的最大对话轮数(含开场白),防止 prompt 无限膨胀。
const personaChatMaxHistory = 24

// personaSpotLimit 记忆注入的景点数上限。
const personaSpotLimit = 2

// personaPoemLimit 每个角色注入的诗作条数上限。
const personaPoemLimit = 6

// PersonaChatService 历史人物角色对话服务。
type PersonaChatService struct {
	llm    *TripLLM
	scenic *repository.ScenicRepo
	poems  *repository.PoemRepo
	// visitorMemories 角色长期记忆(按用户+角色存储,memory_enabled=true 时读取与提取)
	visitorMemories *PersonaMemoryService

	mu           sync.Mutex
	memoryLoaded bool
	// memories persona_id → 拼好的记忆文本(空串表示无可用资料,走纯人设)
	memories map[string]string
}

// NewPersonaChatService 构造角色对话服务(记忆素材在首次对话时从 DB 懒加载并缓存)。
func NewPersonaChatService(llm *TripLLM, scenic *repository.ScenicRepo, poems *repository.PoemRepo, visitorMemories *PersonaMemoryService) *PersonaChatService {
	return &PersonaChatService{llm: llm, scenic: scenic, poems: poems, visitorMemories: visitorMemories}
}

// Personas 返回角色卡片列表(硬编码)。
func (s *PersonaChatService) Personas() []PersonaInfo {
	out := make([]PersonaInfo, 0, len(personaConfigs))
	for i := range personaConfigs {
		out = append(out, personaConfigs[i].info)
	}
	return out
}

// personaByID 按 ID 查找角色配置。
func personaByID(id string) *personaConfig {
	for i := range personaConfigs {
		if personaConfigs[i].info.ID == id {
			return &personaConfigs[i]
		}
	}
	return nil
}

// ValidatePersona 校验角色 ID 是否合法(SSE 前置检查用)。
func (s *PersonaChatService) ValidatePersona(personaID string) error {
	if personaByID(personaID) == nil {
		return fmt.Errorf("%w: %s", ErrUnknownPersona, personaID)
	}
	return nil
}

// Chat 与历史人物对话(非流式,供内部调用与测试)。lang 为界面语言(zh/en/ja),决定回复语言;messages 最后一条应为当前用户消息。
// userID 为登录用户 ID(空串=未登录);memoryEnabled=true 时注入该用户的角色长期记忆。
func (s *PersonaChatService) Chat(ctx context.Context, personaID, lang string, messages []model.ChatMessage, userID string, memoryEnabled bool) (string, error) {
	llmMsgs, err := s.assemblePersonaMessages(personaID, lang, messages, userID, memoryEnabled)
	if err != nil {
		return "", err
	}
	return s.llm.ChatWithEffort(ctx, s.llm.settings.LLMTimeoutSeconds(), llmMsgs, 0.8, 1024, llmThinkingLevel())
}

// ChatStream 与历史人物对话(SSE 流式):每收到增量文本回调 onDelta。
// 消息组装与非流式 Chat 完全一致,仅输出方式不同。
func (s *PersonaChatService) ChatStream(ctx context.Context, personaID, lang string, messages []model.ChatMessage, userID string, memoryEnabled bool, onDelta func(string)) error {
	llmMsgs, err := s.assemblePersonaMessages(personaID, lang, messages, userID, memoryEnabled)
	if err != nil {
		return err
	}
	return s.llm.ChatStream(ctx, llmMsgs, 0.8, 1024, onDelta)
}

// assemblePersonaMessages 组装角色对话的 LLM 消息序列:system(人设+记忆+语言指令) + 截尾后的历史。
func (s *PersonaChatService) assemblePersonaMessages(personaID, lang string, messages []model.ChatMessage, userID string, memoryEnabled bool) ([]llmMessage, error) {
	p := personaByID(personaID)
	if p == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownPersona, personaID)
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("对话内容为空")
	}

	// 历史截尾:只保留最近 personaChatMaxHistory 条
	history := messages
	if len(history) > personaChatMaxHistory {
		history = history[len(history)-personaChatMaxHistory:]
	}

	llmMsgs := make([]llmMessage, 0, len(history)+1)
	llmMsgs = append(llmMsgs, llmMessage{
		Role:    "system",
		Content: s.buildSystemPrompt(p, model.NormalizeLang(lang), userID, memoryEnabled),
	})
	for _, m := range history {
		role := m.Role
		if role != "assistant" {
			role = "user"
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		llmMsgs = append(llmMsgs, llmMessage{Role: role, Content: m.Content})
	}
	return llmMsgs, nil
}

// buildSystemPrompt 组装人设 system prompt:角色规则 + 亲历记忆 + 访客长期记忆(可选) + 语言跟随指令。
func (s *PersonaChatService) buildSystemPrompt(p *personaConfig, lang, userID string, memoryEnabled bool) string {
	var b strings.Builder
	b.WriteString(p.basePromptZH)

	if memory := s.personaMemory(p); memory != "" {
		b.WriteString("\n\n【亲历记忆】\n")
		b.WriteString(memory)
	} else {
		b.WriteString("\n\n【亲历记忆】\n（暂无更多记忆素材，凭你本人的生平与作品自然交谈即可。）")
	}

	// 角色长期记忆:仅登录用户且开关开启时注入,无记忆则跳过该段
	if memoryEnabled {
		if visitor := s.visitorMemories.MemoryPrompt(userID, p.info.ID); visitor != "" {
			b.WriteString("\n\n【关于这位访客的记忆】\n")
			b.WriteString("以下是这位访客的背景与偏好,请把它们当作你对这位老友的了解,自然融入对话:\n")
			b.WriteString(visitor)
		}
	}

	switch lang {
	case "en":
		b.WriteString("\n\n【Language】Always reply in English, keeping Du Fu's / Zhuge Liang's period voice, dignity and persona. Keep Chinese proper nouns recognizable, e.g. \"Du Fu Caotang (Thatched Cottage)\".")
	case "ja":
		b.WriteString("\n\n【言語】必ず日本語で返答してください。ただし杜甫／諸葛亮としての口調・人格・時代の雰囲気を保ち、固有名詞は漢字表記を活かしてください。")
	default:
		b.WriteString("\n\n【语言】用中文交谈；若对方改用外语，则跟随对方的语言，但始终保持角色口吻。")
	}
	return b.String()
}

// personaMemory 返回角色的记忆文本(首次调用时加载并缓存,失败降级为空)。
// 首次加载在锁内进行:仅为一次性的 DB 查询,代价可接受,换取最简单的并发正确性。
func (s *PersonaChatService) personaMemory(p *personaConfig) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.memoryLoaded {
		s.memories = s.loadAllMemories()
		s.memoryLoaded = true
	}
	return s.memories[p.info.ID]
}

// loadAllMemories 从 DB 加载杜甫草堂/武侯祠的景点介绍与关联诗作,拼成"亲历记忆"文本。
// 任一环节失败仅记日志,该角色记忆置空(纯人设对话),不向调用方报错。
func (s *PersonaChatService) loadAllMemories() map[string]string {
	memories := make(map[string]string, len(personaConfigs))
	for i := range personaConfigs {
		p := &personaConfigs[i]
		memories[p.info.ID] = s.loadSpotMemory(p.spotKeyword)
	}
	return memories
}

// loadSpotMemory 加载单个景点的介绍与诗作记忆;查询失败返回空串。
func (s *PersonaChatService) loadSpotMemory(keyword string) string {
	spots, err := s.scenic.ListAll()
	if err != nil {
		logger.Warnf("[PersonaChat] 记忆素材加载失败(景点查询): %v", err)
		return ""
	}

	var matched []model.ScenicSpot
	for _, sp := range spots {
		if strings.Contains(sp.NameZH, keyword) {
			matched = append(matched, sp)
			if len(matched) >= personaSpotLimit {
				break
			}
		}
	}
	if len(matched) == 0 {
		return ""
	}

	var b strings.Builder
	for _, sp := range matched {
		b.WriteString("· 地名：")
		b.WriteString(sp.NameZH)
		if sp.Desc != "" {
			b.WriteString("\n  我之所见：")
			b.WriteString(sp.Desc)
		}
		if sp.DescEN != "" {
			b.WriteString("\n  (EN) ")
			b.WriteString(sp.DescEN)
		}
		if sp.DescJA != "" {
			b.WriteString("\n  (JA) ")
			b.WriteString(sp.DescJA)
		}
		if sp.CultureNoteEN != "" || sp.CultureNoteJA != "" {
			b.WriteString("\n  客人背景注记：")
			if sp.CultureNoteEN != "" {
				b.WriteString(" " + sp.CultureNoteEN)
			}
			if sp.CultureNoteJA != "" {
				b.WriteString(" " + sp.CultureNoteJA)
			}
		}

		poems, err := s.poems.ListBySpot(sp.ID)
		if err != nil {
			logger.Warnf("[PersonaChat] 诗作记忆加载失败 spot_id=%d: %v", sp.ID, err)
		} else {
			for i, poem := range poems {
				if i >= personaPoemLimit {
					break
				}
				b.WriteString("\n· 我的诗篇《")
				b.WriteString(poem.Title)
				b.WriteString("》")
				if poem.Dynasty != "" || poem.Author != "" {
					b.WriteString("（")
					b.WriteString(strings.TrimSpace(poem.Dynasty + " " + poem.Author))
					b.WriteString("）")
				}
				if poem.ContentZH != "" {
					b.WriteString("\n  原文：")
					b.WriteString(poem.ContentZH)
				}
				if poem.ContentEN != "" {
					b.WriteString("\n  (EN) ")
					b.WriteString(poem.ContentEN)
				}
				if poem.ContentJA != "" {
					b.WriteString("\n  (JA) ")
					b.WriteString(poem.ContentJA)
				}
			}
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
