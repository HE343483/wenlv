// 每日蜀签生成器:LLM 批量生成蜀文化日签(诗句引用/四川方言/蜀文化冷知识,三语)。
// 红线:诗句引用必须是真实存在的经典诗文(附作者与篇名),严禁编造或张冠李戴;
// 冷知识/方言基于库内成都景点与蜀文化常识,不得凭空捏造事实。
// 每条中文不超过 60 字;按 content_zh 幂等去重,解析失败重试一次后跳过并记录日志。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

const (
	// dailyGenTimeout 单条蜀签的 LLM 生成超时(秒)。
	dailyGenTimeout = 90
	// dailyGenInterval 批处理节流间隔。
	dailyGenInterval = 500 * time.Millisecond
	// dailyContentCap 中文内容字数上限(提示词约束)。
	dailyContentCap = 60
)

// dailyCategories 三类轮换(落库值,展示端按语言映射标签文案)。
var dailyCategories = []string{"诗句", "方言", "冷知识"}

// dailyGenOutput 单条蜀签的 LLM 生成结果。
type dailyGenOutput struct {
	ContentZH       string `json:"content_zh"`
	ContentEN       string `json:"content_en"`
	ContentJA       string `json:"content_ja"`
	Category        string `json:"category"`
	RelatedSpotName string `json:"related_spot_name"`
}

// CultureDailyGenerator 每日蜀签批量生成器。
type CultureDailyGenerator struct {
	repo *repository.CultureDailyRepo
	db   *gorm.DB
	llm  *TripLLM
}

// NewCultureDailyGenerator 构造蜀签生成器。
func NewCultureDailyGenerator(repo *repository.CultureDailyRepo, db *gorm.DB, llm *TripLLM) *CultureDailyGenerator {
	return &CultureDailyGenerator{repo: repo, db: db, llm: llm}
}

// Generate 批量生成 count 条蜀签(三类轮换,逐条调 LLM)。
// 返回 成功/跳过(重复)/失败 数。
func (g *CultureDailyGenerator) Generate(ctx context.Context, count int) (ok, skip, fail int) {
	// 库内景点名 → id 映射:related_spot_name 精确匹配才挂景点
	spotIDs, err := g.loadSpotIDs()
	if err != nil {
		logger.Errorf("加载库内景点失败: %v", err)
		return
	}
	// 预加载已有中文内容集合:除落库判重外避免同批次内重复
	existing, err := g.loadExisting()
	if err != nil {
		logger.Errorf("加载已有蜀签失败: %v", err)
		return
	}

	for i := 0; i < count; i++ {
		category := dailyCategories[i%len(dailyCategories)]
		out, err := g.callDailyLLM(ctx, category, spotIDs.names())
		if err != nil {
			fail++
			logger.Errorf("蜀签生成失败(第 %d 条,类别 %s): %v", i+1, category, err)
			continue
		}
		content := strings.TrimSpace(out.ContentZH)
		// 幂等:content_zh 重复则跳过(同批已生成 or 库内已有)
		if content == "" || existing[content] {
			skip++
			logger.Infof("蜀签跳过(第 %d 条,类别 %s): 中文内容为空或已存在", i+1, category)
			continue
		}
		item := &model.CultureDaily{
			ContentZH: content,
			ContentEN: strings.TrimSpace(out.ContentEN),
			ContentJA: strings.TrimSpace(out.ContentJA),
			Category:  category,
		}
		if id, hit := spotIDs[strings.TrimSpace(out.RelatedSpotName)]; hit {
			item.RelatedSpotID = &id
		}
		if err := g.repo.Create(item); err != nil {
			fail++
			logger.Errorf("蜀签落库失败(第 %d 条): %v", i+1, err)
			continue
		}
		existing[content] = true
		ok++
		logger.Infof("蜀签生成成功(第 %d 条,类别 %s): %s", i+1, category, content)
		time.Sleep(dailyGenInterval)
	}
	return
}

// dailySpotIDMap 库内景点名 → ID 映射,附带 names() 便于注入提示词。
type dailySpotIDMap map[string]uint

// names 返回景点名列表(逗号分隔)。
func (m dailySpotIDMap) names() string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	return strings.Join(names, "、")
}

// loadSpotIDs 加载库内全部景点(名称精确匹配挂靠)。
func (g *CultureDailyGenerator) loadSpotIDs() (dailySpotIDMap, error) {
	spots := make([]model.ScenicSpot, 0)
	if err := g.db.Select("id", "name_zh").Find(&spots).Error; err != nil {
		return nil, err
	}
	m := make(dailySpotIDMap, len(spots))
	for _, s := range spots {
		m[s.NameZH] = s.ID
	}
	return m, nil
}

// loadExisting 已有蜀签中文内容集合(跨批次幂等)。
func (g *CultureDailyGenerator) loadExisting() (map[string]bool, error) {
	items := make([]model.CultureDaily, 0)
	if err := g.db.Select("content_zh").Find(&items).Error; err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it.ContentZH] = true
	}
	return m, nil
}

// dailyPrompt 组装单条蜀签的生成提示词。
func dailyPrompt(category, spotNames string) string {
	rule := ""
	switch category {
	case "诗句":
		rule = "引用一句真实存在的经典诗文中与蜀地/成都相关的名句,并注明作者与篇名(如「晓看红湿处,花重锦官城」——杜甫《春夜喜雨》);诗句必须真实出自该篇目,严禁编造或张冠李戴"
	case "方言":
		rule = "选一个四川方言词并给出简洁释义(如「巴适:舒服、安逸,川人对美好的最高称赞」),可带一个日常用例"
	case "冷知识":
		rule = "讲一条真实可靠的蜀文化/成都冷知识(古蜀文明、川剧、茶馆、蜀锦、熊猫、市井民俗等),有据可查,不得编造"
	}
	return fmt.Sprintf(`你是成都文旅项目的蜀文化内容编辑,为「每日蜀签」栏目生成一条日签,类别: %s。
严格规则:
1. %s;
2. 内容基于库内成都景点与蜀文化常识,不引入无法核实的事实、数字、排名;
3. 中文 content_zh 不超过 %d 字(诗句引用含出处);content_en/content_ja 为忠实翻译,英文地道、日文自然;
4. related_spot_name:若内容与下列库内景点之一直接相关,填该景点中文名称(必须与列表完全一致),否则填空字符串;
5. 严格只输出 JSON,不要解释、不要 Markdown:
{"content_zh":"...","content_en":"...","content_ja":"...","category":"%s","related_spot_name":"..."}
库内景点列表: %s`, category, rule, dailyContentCap, category, spotNames)
}

// callDailyLLM 调 LLM 并解析 JSON 输出;解析失败自动重试一次(模型偶发引号未转义)。
func (g *CultureDailyGenerator) callDailyLLM(ctx context.Context, category, spotNames string) (*dailyGenOutput, error) {
	if g.llm == nil || !g.llm.Available() {
		return nil, fmt.Errorf("LLM 未配置")
	}
	messages := []llmMessage{
		{Role: "system", Content: "你是严谨的蜀文化内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: dailyPrompt(category, spotNames)},
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(dailyGenInterval)
		}
		reply, err := g.llm.ChatWithEffort(ctx, dailyGenTimeout, messages, 0.7, 1500, llmThinkingLevel())
		if err != nil {
			lastErr = err
			continue
		}
		var out dailyGenOutput
		if err := json.Unmarshal([]byte(extractJSONObject(reply)), &out); err != nil {
			lastErr = fmt.Errorf("LLM 输出解析失败: %w", err)
			continue
		}
		return &out, nil
	}
	return nil, lastErr
}
