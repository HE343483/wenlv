package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"wenlv-backend/logger"
	"wenlv-backend/model"
)

// ============ 景点文本提纯(内容平台无关) ============
// 小红书/抖音等真人内容平台的搜索结果都是"杂文",统一在这里经由 LLM 提纯为结构化景点,
// 并补齐经纬度,供行程规划大模型消费。platform 仅用于提示词与兜底文案中的平台名。

// extractAttractionsFromNotes 调用 LLM 从平台游记/短视频文案中提纯景点,并并发补齐经纬度。
func extractAttractionsFromNotes(ctx context.Context, llm *TripLLM, settings *TripSettings, amap *AmapService,
	city, content, language, platform string) string {
	if platform == "" {
		platform = "内容平台"
	}
	lang := model.NormalizeLang(language)
	translation := ""
	if name, ok := langNames[lang]; ok && lang != "zh" {
		translation = fmt.Sprintf(`
**极其重要的翻译要求:**
目标语言为 %s。你必须将提取结果中的 "name", "reason", "reservation_tips" 字段的内容翻译为 %s。
- "name" 字段使用目标语言 %s 的景点名称(例如中文"故宫博物院" → English "The Palace Museum")。
- "reason" 和 "reservation_tips" 也必须翻译为 %s。
- "duration" 和 "reservation_required" 保持原始数值/布尔值不变。
- **注意**: "name_zh" 必须始终保持简体中文名称,"name_en" 必须始终保持英文名称,不受目标语言影响!
- 严格保持 JSON schema 格式不变!
`, name, name, name, name)
	}

	prompt := fmt.Sprintf(`请从以下真实的素人%s打卡游记中,提纯出真实存在的【游玩景点】。
要求返回严格的 JSON 数组格式(哪怕只提取到了1个),切勿返回除了JSON以外的任何冗余 markdown 文字!
%s
数组中每个对象必须包含以下字段:
"name": 景点官方名称(用于前端展示,按目标语言填写;若目标语言为中文则与 name_zh 相同)
"name_zh": 景点的中文简体名称(必须是简体中文,例如 "故宫博物院"。此字段始终为中文,不受目标语言影响)
"name_en": 景点的英文名称(必须是英文,使用景点在国际上通用的官方英文名。此字段始终为英文,不受目标语言影响)
"reason": %s用户的真实评价/避坑指南
"duration": 游玩时长(数字, 分钟)
"reservation_required": 是否需要提前预约(布尔值 true/false)。请根据游记中提到的"需要预约"、"提前预约"、"抢票"、"约满"、"官方预约"等关键词判断,如果游记未提及则默认为 false
"reservation_tips": 预约相关提示(字符串)。如果需要预约,请提取预约渠道、提前天数等具体信息;如果不需要预约则填空字符串

游记杂文内容如下:
%s

JSON 返回示例:
[{"name": "故宫博物院", "name_zh": "故宫博物院", "name_en": "The Palace Museum", "reason": "必去打卡,建议走中轴线。", "duration": 240, "reservation_required": true, "reservation_tips": "需要提前7天在故宫官网或微信小程序预约"},
 {"name": "老君山金顶", "name_zh": "老君山金顶", "name_en": "Laojun Mountain Golden Summit", "reason": "网红打卡点,夜景绝美。", "duration": 180, "reservation_required": false, "reservation_tips": ""}]
`, platform, translation, platform, content)

	reply, err := llm.Chat(ctx, UserMessage(prompt), 0.1, 4000)
	if err != nil {
		logger.Warnf("大模型提纯%s数据异常: %v", platform, err)
		return fmt.Sprintf("尝试提取%s结构化数据失败,降级回常规处理。", platform)
	}
	jsonText := extractJSONArray(reply)
	if jsonText == "" {
		return fmt.Sprintf("尝试提取%s结构化数据失败,降级回常规处理。", platform)
	}
	var extracted []map[string]any
	if err := json.Unmarshal([]byte(jsonText), &extracted); err != nil {
		logger.Warnf("%s提纯 JSON 解析失败: %v", platform, err)
		return fmt.Sprintf("尝试提取%s结构化数据失败,降级回常规处理。", platform)
	}

	valid := make([]map[string]any, 0, len(extracted))
	for _, item := range extracted {
		if stringValue(item["name"]) != "" {
			valid = append(valid, item)
		}
	}
	if len(valid) == 0 {
		return fmt.Sprintf("未在%s检索到关于 %s 的有效景点信息。", platform, city)
	}

	// 并发补齐经纬度(最多 3 个并发)
	locations := make([]*model.Location, len(valid))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)
	for i, item := range valid {
		wg.Add(1)
		go func(idx int, it map[string]any) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			name := stringValue(it["name"])
			locations[idx] = GeocodeUnified(ctx, settings, amap, name, city,
				firstNonEmpty(stringValue(it["name_zh"]), name),
				firstNonEmpty(stringValue(it["name_en"]), name))
		}(i, item)
	}
	wg.Wait()

	var out strings.Builder
	fmt.Fprintf(&out, "这是%s热门精选游记的提取结果,附带确切坐标(图片由前端单独搜索获取):\n", platform)
	for i, item := range valid {
		if loc := locations[i]; loc != nil {
			item["location"] = map[string]float64{"longitude": loc.Longitude, "latitude": loc.Latitude}
		}
		line, err := json.Marshal(item)
		if err == nil {
			out.Write(line)
			out.WriteString("\n")
		}
	}
	logger.Infof("%s数据挖掘完毕,已装载进上下文。", platform)
	return out.String()
}
