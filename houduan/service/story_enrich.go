// 巴蜀故事引擎:基于库内已有中文简介,LLM 批量生成多语种(EN/JA)故事版介绍与文化注解,
// 美食额外生成直译英文名(点菜神器)、中英食材/过敏原标注。素材唯一来源是库内中文 desc,
// LLM 只做故事化改写与翻译,不得引入素材之外的事实;生成字段一律登记到 estimated_fields(参考值)。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

const (
	// storyLLMTimeout 单条素材的 LLM 生成超时(秒)。
	storyLLMTimeout = 120
	// storyDescCap 注入提示词的中文简介截断长度(字符)。
	storyDescCap = 1500
	// storyEnrichInterval 批处理节流间隔。
	storyEnrichInterval = 500 * time.Millisecond
)

// storyEnrichOutput 单条素材的 LLM 生成结果。
type storyEnrichOutput struct {
	NameEN        string `json:"name_en"`
	NameJA        string `json:"name_ja"`
	StoryEN       string `json:"story_en"`
	StoryJA       string `json:"story_ja"`
	NoteEN        string `json:"note_en"`
	NoteJA        string `json:"note_ja"`
	LiteralEN     string `json:"literal_en"`
	IngredientsZH string `json:"ingredients_zh"`
	IngredientsEN string `json:"ingredients_en"`
}

// routeStopOutput 路线站点多语种简介(按 name_zh 对位写回 stops JSON)。
type routeStopOutput struct {
	NameZH  string `json:"name_zh"`
	DescEN  string `json:"desc_en"`
	DescJA  string `json:"desc_ja"`
}

// StoryEnrichResult 单条记录的生成结果统计。
type StoryEnrichResult struct {
	Name          string
	UpdatedFields []string
	Note          string
}

// StoryEnricher 多语种故事内容生成器(巴蜀故事引擎 + 外国人点菜神器素材)。
type StoryEnricher struct {
	scenicRepo *repository.ScenicRepo
	foodRepo   *repository.FoodRepo
	routeRepo  *repository.RouteRepo
	llm        *TripLLM
}

// NewStoryEnricher 构造故事内容生成器。
func NewStoryEnricher(scenicRepo *repository.ScenicRepo, foodRepo *repository.FoodRepo, routeRepo *repository.RouteRepo, llm *TripLLM) *StoryEnricher {
	return &StoryEnricher{scenicRepo: scenicRepo, foodRepo: foodRepo, routeRepo: routeRepo, llm: llm}
}

// capDesc 中文简介截断,避免超长拖慢/超上下文。
func capDesc(text string) string {
	text = strings.TrimSpace(text)
	if r := []rune(text); len(r) > storyDescCap {
		return string(r[:storyDescCap])
	}
	return text
}

// storyPrompt 组装单条素材的生成提示词。kind 为 scenic(景点)/food(美食)。
func storyPrompt(kind, name, desc, extra string) string {
	foodPart := ""
	if kind == "food" {
		foodPart = `,
"literal_en":"菜名的字面直译英文(制造反差幽默的'直译陷阱',如 夫妻肺片→Husband and Wife Lung Slices,钟水饺→Zhong's Dumplings 之类按字面硬译,保留趣味但不必侮辱性)",
"ingredients_zh":"主要食材与常见过敏原(中文,如'牛肉、牛杂、花椒、辣椒;含内脏',80字内,依据仅限素材,素材没提就写空)",
"ingredients_en":"English translation of ingredients/allergens (same info as ingredients_zh)"`
	}
	extraPart := ""
	if extra != "" {
		extraPart = "\n补充素材:\n" + extra
	}
	return fmt.Sprintf(`你是面向外国游客的成都文旅内容编辑(巴蜀故事引擎)。下面是"%s"的中文介绍素材,请据此生成多语种故事版内容。
严格规则:
1. 仅可依据素材改写,不得引入素材之外的事实、数字、年份、排名、人物、店名;素材没有的信息宁可不写;
2. 故事版不是直译:面向外国读者讲故事,语言地道自然,可解释"为什么/什么背景",但事实必须全部来自素材;
3. culture note 是 1-3 条给外国游客的文化注解(解释文化背景、习俗、典故、怎么体验更地道),每条一句,条间用换行分隔;
4. 严格只输出 JSON,不要输出解释、不要使用 Markdown:
{"name_en":"the name in English (use a common established translation, or pinyin transliteration; do NOT invent new names)","name_ja":"日本語名称(汉字表记が自然なものは汉字、それ以外は通例のカタカナ表记;新しい名前を作らない)","story_en":"English storytelling intro (120-200 words)","story_ja":"日本の旅行者向けのストーリー紹介(200-350字)","note_en":"1-3 English culture notes, one per line","note_ja":"1-3 日本語カルチャーノート、1行ずつ"%s}
素材正文:
%s%s`, name, capDesc(desc), extraPart, foodPart)
}

// callStoryLLM 调 LLM 并解析 JSON 输出;解析失败自动重试一次(模型偶发引号未转义)。
func (e *StoryEnricher) callStoryLLM(ctx context.Context, kind, name, desc, extra string) (*storyEnrichOutput, error) {
	if e.llm == nil || !e.llm.Available() {
		return nil, fmt.Errorf("LLM 未配置")
	}
	messages := []llmMessage{
		{Role: "system", Content: "你是严谨的文旅内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: storyPrompt(kind, name, desc, extra)},
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(storyEnrichInterval)
		}
		reply, err := e.llm.ChatWithEffort(ctx, storyLLMTimeout, messages, 0.4, 2200, llmThinkingLevel())
		if err != nil {
			lastErr = err
			continue
		}
		var out storyEnrichOutput
		if err := json.Unmarshal([]byte(extractJSONObject(reply)), &out); err != nil {
			lastErr = fmt.Errorf("LLM 输出解析失败: %w", err)
			continue
		}
		return &out, nil
	}
	return nil, lastErr
}

// extractJSONObject 从回复中提取最外层 JSON 对象文本(先走围栏提取,再兜底首尾大括号截取)。
func extractJSONObject(reply string) string {
	cleaned := extractJSONFromResponse(reply)
	if start := strings.Index(cleaned, "{"); start >= 0 {
		if end := strings.LastIndex(cleaned, "}"); end > start {
			return cleaned[start : end+1]
		}
	}
	return cleaned
}

// appendNameFields 把名称翻译落库:只补空字段,-force 时才覆盖已有值。
func appendNameFields(updates map[string]any, out *storyEnrichOutput, curEN, curJA string, force bool) {
	if v := strings.TrimSpace(out.NameEN); v != "" && (force || strings.TrimSpace(curEN) == "") {
		updates["name_en"] = v
	}
	if v := strings.TrimSpace(out.NameJA); v != "" && (force || strings.TrimSpace(curJA) == "") {
		updates["name_ja"] = v
	}
}

// appendStoryFields 把生成结果转为落库字段表(空字符串跳过,不覆盖)。
func appendStoryFields(updates map[string]any, out *storyEnrichOutput, withFood bool) {
	if v := strings.TrimSpace(out.StoryEN); v != "" {
		updates["desc_en"] = v
	}
	if v := strings.TrimSpace(out.StoryJA); v != "" {
		updates["desc_ja"] = v
	}
	if v := strings.TrimSpace(out.NoteEN); v != "" {
		updates["culture_note_en"] = v
	}
	if v := strings.TrimSpace(out.NoteJA); v != "" {
		updates["culture_note_ja"] = v
	}
	if withFood {
		if v := strings.TrimSpace(out.LiteralEN); v != "" {
			updates["name_literal_en"] = v
		}
		if v := strings.TrimSpace(out.IngredientsZH); v != "" {
			updates["ingredients_zh"] = v
		}
		if v := strings.TrimSpace(out.IngredientsEN); v != "" {
			updates["ingredients_en"] = v
		}
	}
}

// storyEstimatedFields 生成字段对应的 estimated_fields 登记(供追加)。
var storyEstimatedFields = []string{"desc_en", "desc_ja", "culture_note_en", "culture_note_ja"}
var storyFoodEstimatedFields = []string{"name_literal_en", "ingredients_zh", "ingredients_en"}

// mergeEstimatedFields 把生成的参考值字段追加进已有 estimated_fields(去重)。
func mergeEstimatedFields(existing string, extra []string) string {
	seen := map[string]bool{}
	parts := make([]string, 0, 16)
	for _, p := range strings.Split(existing, ",") {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			seen[p] = true
			parts = append(parts, p)
		}
	}
	for _, p := range extra {
		if !seen[p] {
			seen[p] = true
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ",")
}

// enrichScenic 生成单个景点的多语种故事并落库。
func (e *StoryEnricher) enrichScenic(ctx context.Context, s model.ScenicSpot, force bool) (*StoryEnrichResult, error) {
	res := &StoryEnrichResult{Name: s.NameZH}
	if strings.TrimSpace(s.Desc) == "" {
		res.Note = "中文介绍为空,跳过(无素材不编造)"
		return res, nil
	}
	if !force && s.DescEN != "" && s.DescJA != "" && s.CultureNoteEN != "" && s.CultureNoteJA != "" && s.NameEN != "" && s.NameJA != "" {
		res.Note = "多语种字段已齐全,跳过(如需重生成使用 -force)"
		return res, nil
	}
	out, err := e.callStoryLLM(ctx, "scenic", s.NameZH, s.Desc, "")
	if err != nil {
		res.Note = "LLM 生成失败: " + err.Error()
		return res, nil
	}
	updates := map[string]any{}
	appendStoryFields(updates, out, false)
	appendNameFields(updates, out, s.NameEN, s.NameJA, force)
	if len(updates) == 0 {
		res.Note = "LLM 未产出有效字段"
		return res, nil
	}
	updates["estimated_fields"] = mergeEstimatedFields(s.EstimatedFields, storyEstimatedFields)
	updates["data_updated_at"] = time.Now()
	for k := range updates {
		if k != "estimated_fields" && k != "data_updated_at" {
			res.UpdatedFields = append(res.UpdatedFields, k)
		}
	}
	if err := e.scenicRepo.UpdateFields(s.ID, updates); err != nil {
		return res, err
	}
	return res, nil
}

// enrichFood 生成单个美食的多语种故事与点菜神器素材并落库。
func (e *StoryEnricher) enrichFood(ctx context.Context, f model.Food, force bool) (*StoryEnrichResult, error) {
	res := &StoryEnrichResult{Name: f.NameZH}
	if strings.TrimSpace(f.Desc) == "" {
		res.Note = "中文介绍为空,跳过(无素材不编造)"
		return res, nil
	}
	if !force && f.DescEN != "" && f.DescJA != "" && f.NameLiteralEN != "" && f.NameJA != "" {
		res.Note = "多语种字段已齐全,跳过(如需重生成使用 -force)"
		return res, nil
	}
	extra := ""
	if v := strings.TrimSpace(f.NameEN); v != "" {
		extra = "菜名英文(意译,与 literal_en 区分):" + v
	}
	out, err := e.callStoryLLM(ctx, "food", f.NameZH, f.Desc, extra)
	if err != nil {
		res.Note = "LLM 生成失败: " + err.Error()
		return res, nil
	}
	updates := map[string]any{}
	appendStoryFields(updates, out, true)
	appendNameFields(updates, out, f.NameEN, f.NameJA, force)
	if len(updates) == 0 {
		res.Note = "LLM 未产出有效字段"
		return res, nil
	}
	updates["estimated_fields"] = mergeEstimatedFields(f.EstimatedFields, append(append([]string{}, storyEstimatedFields...), storyFoodEstimatedFields...))
	updates["data_updated_at"] = time.Now()
	for k := range updates {
		if k != "estimated_fields" && k != "data_updated_at" {
			res.UpdatedFields = append(res.UpdatedFields, k)
		}
	}
	if err := e.foodRepo.UpdateFields(f.ID, updates); err != nil {
		return res, err
	}
	return res, nil
}

// enrichRoute 生成单条路线的多语种简介与站点多语种简介并落库。
func (e *StoryEnricher) enrichRoute(ctx context.Context, rt model.Route, force bool) (*StoryEnrichResult, error) {
	res := &StoryEnrichResult{Name: rt.TitleZH}
	if !force && rt.DescriptionEN != "" && rt.DescriptionJA != "" {
		res.Note = "多语种简介已齐全,跳过"
		return res, nil
	}
	// 1) 路线简介:素材 = 中文 description + 站点名列表
	stopNames := routeStopNames(rt.Stops)
	extra := "途经站点:" + strings.Join(stopNames, "、")
	if strings.TrimSpace(rt.Description) == "" {
		res.Note = "路线中文简介为空,仅尝试站点翻译"
	} else {
		out, err := e.callStoryLLM(ctx, "scenic", rt.TitleZH, rt.Description, extra)
		if err != nil {
			res.Note = "简介 LLM 生成失败: " + err.Error()
		} else {
			updates := map[string]any{}
			if v := strings.TrimSpace(out.StoryEN); v != "" {
				updates["description_en"] = v
			}
			if v := strings.TrimSpace(out.StoryJA); v != "" {
				updates["description_ja"] = v
			}
			if len(updates) > 0 {
				for k := range updates {
					res.UpdatedFields = append(res.UpdatedFields, k)
				}
				if err := e.routeRepo.UpdateFields(rt.ID, updates); err != nil {
					return res, err
				}
			}
		}
	}
	// 2) 站点多语种简介:逐条生成,失败站点跳过,不整体失败
	if err := e.enrichRouteStops(ctx, rt); err != nil {
		res.Note = res.Note + "; 站点翻译部分失败: " + err.Error()
	}
	return res, nil
}

// routeStopNames 解析 stops JSON,返回站点中文名列表。
func routeStopNames(stopsJSON string) []string {
	var stops []map[string]any
	if err := json.Unmarshal([]byte(stopsJSON), &stops); err != nil {
		return nil
	}
	names := make([]string, 0, len(stops))
	for _, st := range stops {
		if v, ok := st["name_zh"].(string); ok && strings.TrimSpace(v) != "" {
			names = append(names, v)
		}
	}
	return names
}

// enrichRouteStops 为路线站点 desc 生成 desc_en/desc_ja 并写回 stops JSON。
// 单个站点失败仅跳过;所有站点都无中文 desc 时不动 JSON。
func (e *StoryEnricher) enrichRouteStops(ctx context.Context, rt model.Route) error {
	var stops []map[string]any
	if err := json.Unmarshal([]byte(rt.Stops), &stops); err != nil {
		return fmt.Errorf("stops JSON 解析失败: %w", err)
	}
	changed := false
	for i, st := range stops {
		name, _ := st["name_zh"].(string)
		desc, _ := st["desc"].(string)
		if strings.TrimSpace(name) == "" || strings.TrimSpace(desc) == "" {
			continue
		}
		if _, hasEN := st["desc_en"].(string); hasEN && strings.TrimSpace(st["desc_en"].(string)) != "" {
			continue // 已有翻译,不重做
		}
		time.Sleep(storyEnrichInterval)
		out, err := e.callStoryLLM(ctx, "scenic", name, desc, "")
		if err != nil {
			logger.Warnf("站点 %s 生成失败: %v", name, err)
			continue
		}
		if v := strings.TrimSpace(out.StoryEN); v != "" {
			st["desc_en"] = v
			changed = true
		}
		if v := strings.TrimSpace(out.StoryJA); v != "" {
			st["desc_ja"] = v
			changed = true
		}
		stops[i] = st
	}
	if !changed {
		return nil
	}
	raw, err := json.Marshal(stops)
	if err != nil {
		return err
	}
	return e.routeRepo.UpdateFields(rt.ID, map[string]any{"stops": string(raw)})
}

// storyTask 单个批处理任务(名称用于 only 过滤)。
type storyTask struct {
	name string
	run  func() (*StoryEnrichResult, error)
}

// EnrichScenics 批量生成景点多语种故事。
func (e *StoryEnricher) EnrichScenics(ctx context.Context, only string, limit int, force bool) (okCnt, failCnt, skipCnt int) {
	items, err := e.scenicRepo.ListAll()
	if err != nil {
		logger.Fatalf("读取景点失败: %v", err)
	}
	tasks := make([]storyTask, 0, len(items))
	for i := range items {
		s := items[i]
		tasks = append(tasks, storyTask{name: s.NameZH, run: func() (*StoryEnrichResult, error) {
			return e.enrichScenic(ctx, s, force)
		}})
	}
	return e.runBatch(ctx, tasks, only, limit)
}

// EnrichFoods 批量生成美食多语种故事与点菜素材。
func (e *StoryEnricher) EnrichFoods(ctx context.Context, only string, limit int, force bool) (okCnt, failCnt, skipCnt int) {
	items, err := e.foodRepo.ListAll()
	if err != nil {
		logger.Fatalf("读取美食失败: %v", err)
	}
	tasks := make([]storyTask, 0, len(items))
	for i := range items {
		f := items[i]
		tasks = append(tasks, storyTask{name: f.NameZH, run: func() (*StoryEnrichResult, error) {
			return e.enrichFood(ctx, f, force)
		}})
	}
	return e.runBatch(ctx, tasks, only, limit)
}

// EnrichRoutes 批量生成路线多语种简介与站点翻译。
func (e *StoryEnricher) EnrichRoutes(ctx context.Context, only string, limit int, force bool) (okCnt, failCnt, skipCnt int) {
	items, err := e.routeRepo.ListAll()
	if err != nil {
		logger.Fatalf("读取路线失败: %v", err)
	}
	tasks := make([]storyTask, 0, len(items))
	for i := range items {
		rt := items[i]
		tasks = append(tasks, storyTask{name: rt.TitleZH, run: func() (*StoryEnrichResult, error) {
			return e.enrichRoute(ctx, rt, force)
		}})
	}
	return e.runBatch(ctx, tasks, only, limit)
}

// runBatch 批处理公共流程:过滤、节流、日志、统计。
func (e *StoryEnricher) runBatch(ctx context.Context, tasks []storyTask, only string, limit int) (okCnt, failCnt, skipCnt int) {
	if only != "" {
		filtered := make([]storyTask, 0, 4)
		for _, t := range tasks {
			if strings.Contains(t.name, only) {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}
	if limit > 0 && limit < len(tasks) {
		tasks = tasks[:limit]
	}
	for n, t := range tasks {
		time.Sleep(storyEnrichInterval)
		logger.Infof("[%d/%d] %s", n+1, len(tasks), t.name)
		res, err := t.run()
		if err != nil {
			logger.Warnf("落库失败: %v", err)
			failCnt++
			continue
		}
		if res != nil {
			if len(res.UpdatedFields) == 0 {
				logger.Infof("跳过: %s", res.Note)
				skipCnt++
			} else {
				logger.Infof("更新字段=%v 备注=%s", res.UpdatedFields, res.Note)
				okCnt++
			}
		}
	}
	return
}
