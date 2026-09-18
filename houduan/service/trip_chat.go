package service

import (
	"context"
	"encoding/json"
	"fmt"

	"wenlv-backend/model"
)

// ============ AI 行程问答服务 ============

const tripChatSystemPrompt = `你是一个专业且贴心的私人旅行管家「天府智能AI」。

你当前正在为用户提供关于一份 **已生成的旅行计划** 的答疑服务。
用户可能会问你关于行程中的景点、酒店、餐饮、天气、交通、门票、费用等任何细节问题。

请根据下方提供的【当前旅行计划】JSON 上下文来回答用户的问题。
回答规则:
1. 如果行程数据中包含相关信息,请精确引用并给出详细回答。
2. 如果行程数据中没有明确信息,可以基于常识进行合理推断,但需说明"行程中未提供该信息,以下是建议"。
3. 回答要有温度、条理清晰,适当使用 emoji 增加亲切感 🌟。
4. 回答尽量简洁,控制在200字以内,除非用户要求详细展开。
5. 使用中文回答。`

// TripChatService 行程问答服务。
type TripChatService struct {
	llm *TripLLM
}

// NewTripChatService 构造问答服务。
func NewTripChatService(llm *TripLLM) *TripChatService {
	return &TripChatService{llm: llm}
}

// ChatWithTripContext 结合行程上下文回答用户提问。
func (s *TripChatService) ChatWithTripContext(ctx context.Context, message string, tripPlan map[string]any, history []model.ChatMessage) (string, error) {
	if !s.llm.Available() {
		return "抱歉,AI 服务尚未配置 API Key,请先在设置页面中完成配置。", nil
	}
	planJSON, err := json.MarshalIndent(tripPlan, "", "  ")
	if err != nil {
		return "", fmt.Errorf("行程上下文序列化失败: %w", err)
	}
	contextMsg := fmt.Sprintf("【当前旅行计划】\n```json\n%s\n```", string(planJSON))

	messages := []llmMessage{
		{Role: "system", Content: tripChatSystemPrompt},
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

	reply, err := s.llm.Chat(ctx, messages, 0.7, 1024)
	if err != nil {
		return "", err
	}
	return reply, nil
}