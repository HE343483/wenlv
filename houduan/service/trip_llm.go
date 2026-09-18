package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TripLLM 行程模块的 LLM 客户端(OpenAI 兼容 Chat Completions)。
// 每次调用都从运行时配置读取参数,支持前端设置页热更新。
type TripLLM struct {
	settings *TripSettings
	client   *http.Client
}

// NewTripLLM 构造 LLM 客户端。禁用系统代理继承,避免国内环境误走代理。
// 注意:不设置 client.Timeout,统一由每次请求的 context 控制超时,
// 避免规划阶段(PlannerTimeout,默认180s)被客户端固定超时截断。
func NewTripLLM(settings *TripSettings) *TripLLM {
	return &TripLLM{
		settings: settings,
		client: &http.Client{
			Transport: &http.Transport{Proxy: nil},
		},
	}
}

// Available 是否已配置 API Key。
func (l *TripLLM) Available() bool {
	return l.settings.Snapshot().OpenAIAPIKey != ""
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmRequest struct {
	Model       string       `json:"model"`
	Messages    []llmMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream"`
}

type llmResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat 调用对话补全接口,返回纯文本内容(常规调用,超时取 LLM_TIMEOUT)。
func (l *TripLLM) Chat(ctx context.Context, messages []llmMessage, temperature float64, maxTokens int) (string, error) {
	return l.ChatWithTimeout(ctx, l.settings.LLMTimeoutSeconds(), messages, temperature, maxTokens)
}

// ChatWithTimeout 以指定超时(秒)调用对话补全接口;行程规划等长任务使用 PlannerTimeout。
func (l *TripLLM) ChatWithTimeout(ctx context.Context, timeoutSeconds int, messages []llmMessage, temperature float64, maxTokens int) (string, error) {
	cfg := l.settings.Snapshot()
	if cfg.OpenAIAPIKey == "" {
		return "", errors.New("LLM API Key 未配置,请先在设置页完成配置")
	}
	base := cfg.OpenAIBaseURL
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := cfg.OpenAIModel
	if model == "" {
		model = "gpt-4"
	}

	body, err := json.Marshal(llmRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return "", err
	}

	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM 请求失败: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var parsed llmResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("LLM 响应解析失败(HTTP %d): %s", resp.StatusCode, truncateText(string(raw), 200))
	}
	if resp.StatusCode != http.StatusOK {
		msg := ""
		if parsed.Error != nil {
			msg = parsed.Error.Message
		}
		return "", fmt.Errorf("LLM 返回错误(HTTP %d): %s", resp.StatusCode, truncateText(msg, 300))
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM 未返回任何内容")
	}
	return parsed.Choices[0].Message.Content, nil
}

// NewLLMMessages 构造一条 system + 一条 user 的消息列表(便捷方法)。
func NewLLMMessages(system, user string) []llmMessage {
	msgs := make([]llmMessage, 0, 2)
	if system != "" {
		msgs = append(msgs, llmMessage{Role: "system", Content: system})
	}
	msgs = append(msgs, llmMessage{Role: "user", Content: user})
	return msgs
}

// UserMessage 构造单条 user 消息。
func UserMessage(content string) []llmMessage {
	return []llmMessage{{Role: "user", Content: content}}
}

func truncateText(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}