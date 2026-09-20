package service

import (
	"context"
	"encoding/json"
	"fmt"

	"wenlv-backend/model"
)

// ============ AI 行程问答服务 ============

const tripChatSystemPromptZH = `你是一个专业且贴心的私人旅行管家「天府智能AI」。

你当前正在为用户提供关于一份 **已生成的旅行计划** 的答疑服务。
用户可能会问你关于行程中的景点、酒店、餐饮、天气、交通、门票、费用等任何细节问题。

请根据下方提供的【当前旅行计划】JSON 上下文来回答用户的问题。
回答规则:
1. 如果行程数据中包含相关信息,请精确引用并给出详细回答。
2. 如果行程数据中没有明确信息,可以基于常识进行合理推断,但需说明"行程中未提供该信息,以下是建议"。
3. 回答要有温度、条理清晰,适当使用 emoji 增加亲切感 🌟。
4. 回答尽量简洁,控制在200字以内,除非用户要求详细展开。
5. 使用中文回答。`

// buildChatSystemPrompt 按界面语言生成问答 system prompt:
// 非中文时要求用目标语言回答(菜名/地名等专有名词可括注中文或英文原文),若用户改用其他语言提问则跟随用户。
func buildChatSystemPrompt(lang string) string {
	if lang == "zh" || lang == "" {
		return tripChatSystemPromptZH
	}
	name := langNames[lang]
	if name == "" {
		name = lang
	}
	return `You are a professional and caring personal travel concierge "Tianfu AI".

You are answering questions about a **previously generated trip plan**.
The user may ask about any detail of the itinerary: attractions, hotels, dining, weather, transport, tickets, costs, etc.

Answer based on the [Current Trip Plan] JSON context provided below.
Rules:
1. If the plan contains relevant information, quote it precisely and answer in detail.
2. If the plan lacks the information, you may reasonably infer from common knowledge, but state that "the plan does not include this; here is a suggestion".
3. Keep the answer warm and well-structured, with a few emojis for friendliness 🌟.
4. Keep the answer concise (under 200 words) unless the user asks for details.
5. Respond in ` + name + `. Keep proper nouns (dish names, place names) recognizable, optionally with the Chinese original in parentheses, e.g. "Kuanzhai Alley (宽窄巷子)".
6. If the user writes in another language, follow the user's language instead.`
}

// TripChatService 行程问答服务。
type TripChatService struct {
	llm *TripLLM
}

// NewTripChatService 构造问答服务。
func NewTripChatService(llm *TripLLM) *TripChatService {
	return &TripChatService{llm: llm}
}

// buildTripChatMessages 组装问答消息列表:系统人设(按界面语言) + 行程计划上下文 + 历史对话 + 本次提问。
func (s *TripChatService) buildTripChatMessages(message string, lang string, tripPlan map[string]any, history []model.ChatMessage) ([]llmMessage, error) {
	// 紧凑序列化(无缩进):相比 MarshalIndent 可减少大量 token,显著降低首字延迟
	planJSON, err := json.Marshal(tripPlan)
	if err != nil {
		return nil, fmt.Errorf("行程上下文序列化失败: %w", err)
	}
	contextMsg := fmt.Sprintf("【当前旅行计划】\n```json\n%s\n```", string(planJSON))

	messages := []llmMessage{
		{Role: "system", Content: buildChatSystemPrompt(model.NormalizeLang(lang))},
		{Role: "user", Content: contextMsg},
	}
	for _, h := range history {
		role := h.Role
		if role != "assistant" {
			role = "user"
		}
		messages = append(messages, llmMessage{Role: role, Content: h.Content})
	}
	messages = append(messages, llmMessage{Role: "user", Content: message})
	return messages, nil
}

// ChatWithTripContext 结合行程上下文回答用户提问。lang 为界面语言(zh/en/ja)。
func (s *TripChatService) ChatWithTripContext(ctx context.Context, message, lang string, tripPlan map[string]any, history []model.ChatMessage) (string, error) {
	if !s.llm.Available() {
		return "抱歉,AI 服务尚未配置 API Key,请先在设置页面中完成配置。", nil
	}
	messages, err := s.buildTripChatMessages(message, lang, tripPlan, history)
	if err != nil {
		return "", err
	}
	reply, err := s.llm.Chat(ctx, messages, 0.7, 1024)
	if err != nil {
		return "", err
	}
	return reply, nil
}

// ChatWithTripContextStream 流式结合行程上下文回答用户提问,每收到增量文本回调 onDelta。
func (s *TripChatService) ChatWithTripContextStream(ctx context.Context, message, lang string, tripPlan map[string]any, history []model.ChatMessage, onDelta func(string)) error {
	if !s.llm.Available() {
		onDelta("抱歉,AI 服务尚未配置 API Key,请先在设置页面中完成配置。")
		return nil
	}
	messages, err := s.buildTripChatMessages(message, lang, tripPlan, history)
	if err != nil {
		return err
	}
	return s.llm.ChatStream(ctx, messages, 0.7, 1024, onDelta)
}
