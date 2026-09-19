// 旅行故事卡片文案生成:基于已生成的行程计划,LLM 产出一段多语种"旅行故事"
// (标题 + 正文),供前端绘制可分享的故事海报。素材只来自行程本身(城市/天数/
// 每日景点与美食),禁止编造未到访的地点或事实;LLM 失败时前端用本地模板兜底。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wenlv-backend/model"
)

const (
	// storyCardLLMTimeout 故事文案生成超时(秒)。
	storyCardLLMTimeout = 90
)

// TripStoryCardContent 故事卡片文案(多语种)。
type TripStoryCardContent struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Language string `json:"language"`
}

// TripStoryCardService 故事卡片文案生成服务。
type TripStoryCardService struct {
	llm *TripLLM
}

// NewTripStoryCardService 构造故事卡片服务。
func NewTripStoryCardService(llm *TripLLM) *TripStoryCardService {
	return &TripStoryCardService{llm: llm}
}

// NormalizeStoryLang 归一化语言参数(仅支持 zh/en/ja,默认 zh)。
func NormalizeStoryLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "en":
		return "en"
	case "ja", "jp":
		return "ja"
	default:
		return "zh"
	}
}

// storyCardLangDesc 语言对应的输出要求。
func storyCardLangDesc(lang string) string {
	switch lang {
	case "en":
		return "English (natural, warm, social-media friendly)"
	case "ja":
		return "日本語(自然で温かみのある表現)"
	default:
		return "简体中文"
	}
}

// summarizePlan 从行程计划提取文案素材:城市、天数、每日亮点(景点 + 一餐)。
// 素材注入上限:每天最多 3 个景点名,避免提示词过长。
func summarizePlan(plan *model.TripPlan) string {
	city := plan.City
	if len(plan.Cities) > 0 {
		city = strings.Join(plan.Cities, "→")
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("目的地:%s;天数:%d", city, len(plan.Days)))
	for _, d := range plan.Days {
		dayLabel := d.Date
		if dayLabel == "" {
			dayLabel = fmt.Sprintf("第%d天", d.DayIndex)
		}
		spots := make([]string, 0, 3)
		for _, a := range d.Attractions {
			if strings.TrimSpace(a.Name) == "" {
				continue
			}
			spots = append(spots, a.Name)
			if len(spots) >= 3 {
				break
			}
		}
		meal := ""
		for _, m := range d.Meals {
			if strings.TrimSpace(m.Name) != "" {
				meal = m.Name
				break
			}
		}
		line := fmt.Sprintf("\n- %s: 景点 %s", dayLabel, strings.Join(spots, "、"))
		if meal != "" {
			line += "; 美食 " + meal
		}
		b.WriteString(line)
	}
	return b.String()
}

// Generate 生成故事卡片文案。
func (s *TripStoryCardService) Generate(ctx context.Context, lang string, plan *model.TripPlan) (*TripStoryCardContent, error) {
	lang = NormalizeStoryLang(lang)
	if s.llm == nil || !s.llm.Available() {
		return nil, fmt.Errorf("LLM 未配置")
	}
	prompt := fmt.Sprintf(`你是旅行内容创作者。一位旅行者刚完成以下成都文旅行程,请为他写一段"旅行故事"文案,将印在可分享的旅行故事海报上。
要求:
1. title: 一句有画面感的标题(15字以内);
2. body: 80-120 字的旅行故事短文,第一人称,口语化、有情绪,自然融入 2-3 个行程中的亮点(只用下方素材中出现的地点);
3. 不得编造素材之外的地点、店名、数字;输出语言: %s;
4. 严格只输出 JSON,不要解释、不要 Markdown:
{"title":"...","body":"..."}
行程素材:
%s`, storyCardLangDesc(lang), summarizePlan(plan))
	reply, err := s.llm.ChatWithEffort(ctx, storyCardLLMTimeout, []llmMessage{
		{Role: "system", Content: "你是旅行内容创作者,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.7, 600, llmThinkingLevel())
	if err != nil {
		return nil, err
	}
	var out struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	// 与故事引擎同一套提取逻辑:围栏提取 + 首尾大括号兜底,降低模型偶发转义导致的解析失败
	if err := json.Unmarshal([]byte(extractJSONObject(reply)), &out); err != nil {
		return nil, fmt.Errorf("LLM 输出解析失败: %w", err)
	}
	if strings.TrimSpace(out.Title) == "" || strings.TrimSpace(out.Body) == "" {
		return nil, fmt.Errorf("LLM 未产出有效文案")
	}
	return &TripStoryCardContent{Title: strings.TrimSpace(out.Title), Body: strings.TrimSpace(out.Body), Language: lang}, nil
}
