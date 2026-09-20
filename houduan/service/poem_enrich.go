// 诗词多语种内容生成:poems 表原文已人工权威录入,LLM 仅负责英/日翻译与白话赏析,
// 绝不参与原文生成,避免经典文本被模型改写。prompt 严格约束忠实原意、不强行押韵。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wenlv-backend/model"
)

const (
	// poemLLMTimeout 单首诗词的 LLM 生成超时(秒)。
	poemLLMTimeout = 120
	// poemLLMInterval 逐首生成节流间隔(重试等待共用)。
	poemLLMInterval = 500 * time.Millisecond
)

// poemLLMOutput 单首诗词的 LLM 生成结果。
type poemLLMOutput struct {
	ContentEN string `json:"content_en"`
	ContentJA string `json:"content_ja"`
	PlainZH   string `json:"plain_zh"`
}

// PoemEnricher 诗词翻译与赏析生成器(不触碰原文)。
type PoemEnricher struct {
	llm *TripLLM
}

// NewPoemEnricher 构造诗词内容生成器。
func NewPoemEnricher(llm *TripLLM) *PoemEnricher {
	return &PoemEnricher{llm: llm}
}

// poemPrompt 组装单首诗词的翻译与赏析提示词。
func poemPrompt(title, dynasty, author, content string) string {
	return fmt.Sprintf(`你是面向外国游客的文旅内容编辑。下面是一首中国古典诗词的权威原文(人工录入,不可改动),请生成翻译与白话赏析。
严格规则:
1. content_en: 英文翻译,忠实传达原意与意象,语言自然流畅,不逐字硬译、不自由发挥、不强求押韵;
2. content_ja: 日本語訳,原意とイメージを忠実に伝える自然な現代日本語(韻を踏む必要はない);
3. plain_zh: 中文白话赏析,100字以内,通俗讲解诗意与意境,仅可使用通行的文学常识,不得编造史实;
4. 严格只输出 JSON,不要输出解释、不要使用 Markdown:
{"content_en":"...","content_ja":"...","plain_zh":"..."}
诗词原文:
《%s》 %s·%s
%s`, title, dynasty, author, content)
}

// EnrichPoem 生成单首诗词的英/日翻译与白话赏析,直接填充到记录字段。
// 解析失败自动重试一次(模型偶发引号未转义);生成结果为空视为失败。
// 返回是否有字段被填充。
func (e *PoemEnricher) EnrichPoem(ctx context.Context, p *model.Poem) (bool, error) {
	if e.llm == nil || !e.llm.Available() {
		return false, fmt.Errorf("LLM 未配置")
	}
	messages := []llmMessage{
		{Role: "system", Content: "你是严谨的古典诗词编辑,只输出合法 JSON。"},
		{Role: "user", Content: poemPrompt(p.Title, p.Dynasty, p.Author, p.ContentZH)},
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(poemLLMInterval)
		}
		reply, err := e.llm.ChatWithEffort(ctx, poemLLMTimeout, messages, 0.3, 1500, llmThinkingLevel())
		if err != nil {
			lastErr = err
			continue
		}
		var out poemLLMOutput
		if err := json.Unmarshal([]byte(extractJSONObject(reply)), &out); err != nil {
			lastErr = err
			continue
		}
		changed := false
		if v := strings.TrimSpace(out.ContentEN); v != "" {
			p.ContentEN = v
			changed = true
		}
		if v := strings.TrimSpace(out.ContentJA); v != "" {
			p.ContentJA = v
			changed = true
		}
		if v := strings.TrimSpace(out.PlainZH); v != "" {
			p.PlainZH = v
			changed = true
		}
		if !changed {
			return false, fmt.Errorf("生成内容为空")
		}
		return true, nil
	}
	return false, lastErr
}
