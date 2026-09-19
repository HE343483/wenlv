package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	Model           string       `json:"model"`
	Messages        []llmMessage `json:"messages"`
	Temperature     float64      `json:"temperature"`
	MaxTokens       int          `json:"max_tokens,omitempty"`
	Stream          bool         `json:"stream"`
	ReasoningEffort string       `json:"reasoning_effort,omitempty"`
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
	return l.ChatWithEffort(ctx, timeoutSeconds, messages, temperature, maxTokens, "")
}

// ChatWithEffort 与 ChatWithTimeout 相同,额外附加 reasoning_effort。
// GLM 等"始终思考"的推理模型(如 glm-5.3-flash)在长文本生成任务上会把 max_tokens
// 全部消耗在思考上,导致 content 为空、finish_reason=length;显式指定 reasoning_effort
// (low/high/max) 才能拿到正文。effort 为空时不发送该字段。
func (l *TripLLM) ChatWithEffort(ctx context.Context, timeoutSeconds int, messages []llmMessage, temperature float64, maxTokens int, effort string) (string, error) {
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
		Model:           model,
		Messages:        messages,
		Temperature:     temperature,
		MaxTokens:       maxTokens,
		ReasoningEffort: effort,
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
	content := parsed.Choices[0].Message.Content
	// 仅在显式指定 effort 时校验空内容:ChatWithTimeout(effort 为空)保持"空串照常返回"的既有行为。
	if content == "" && effort != "" {
		return "", errors.New("LLM 返回内容为空(可能是思考 token 耗尽 max_tokens,可调大 max_tokens 或设置 LLM_THINKING_LEVEL)")
	}
	return content, nil
}

// llmThinkingLevel 读取 LLM_THINKING_LEVEL(取值 low/high/max,其它值视为未配置)。
// 用于"始终思考"的推理模型:不设置时长文本生成的 content 会是空的。
func llmThinkingLevel() string {
	switch v := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_THINKING_LEVEL"))); v {
	case "low", "high", "max":
		return v
	default:
		return ""
	}
}

// llmStreamChunk OpenAI 兼容流式响应的单个 SSE 分片。
type llmStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatStream 以 stream 模式调用对话补全接口,每收到一段增量文本回调一次 onDelta。
// 用于 AI 问答等需要低首字延迟的场景。
func (l *TripLLM) ChatStream(ctx context.Context, messages []llmMessage, temperature float64, maxTokens int, onDelta func(string)) error {
	cfg := l.settings.Snapshot()
	if cfg.OpenAIAPIKey == "" {
		return errors.New("LLM API Key 未配置,请先在设置页完成配置")
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
		Stream:      true,
	})
	if err != nil {
		return err
	}

	timeout := time.Duration(l.settings.LLMTimeoutSeconds()) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("LLM 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("LLM 返回错误(HTTP %d): %s", resp.StatusCode, truncateText(string(raw), 300))
	}

	// 逐行解析 SSE:data: {...} / data: [DONE]
	reader := bufio.NewReader(resp.Body)
	for {
		line, readErr := reader.ReadString('\n')
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if data == "[DONE]" {
				return nil
			}
			var chunk llmStreamChunk
			if json.Unmarshal([]byte(data), &chunk) == nil && len(chunk.Choices) > 0 {
				if delta := chunk.Choices[0].Delta.Content; delta != "" {
					onDelta(delta)
				}
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return fmt.Errorf("读取 LLM 流式响应失败: %w", readErr)
		}
	}
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
