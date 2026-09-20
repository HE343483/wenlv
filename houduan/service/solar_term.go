// 节气服务:应季美食查询(前端"节气蜀俗"条幅)+ 美食节气 LLM 打标(cmd/solar-term-tag 批处理)。
// solar_terms 以逗号分隔存储(如"冬至,大雪"),查询端 FIND_IN_SET 精确匹配单个节气。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/repository"
)

// SolarTermList 二十四节气全集(与前端 src/utils/solar-term.ts 的节气表保持一致)。
const SolarTermList = "立春,雨水,惊蛰,春分,清明,谷雨,立夏,小满,芒种,夏至,小暑,大暑,立秋,处暑,白露,秋分,寒露,霜降,立冬,小雪,大雪,冬至,小寒,大寒"

// SolarFoodItem 应季美食精简字段(条幅横滑小卡所需)。
type SolarFoodItem struct {
	ID       uint   `json:"id"`
	NameZH   string `json:"name_zh"`
	NameEN   string `json:"name_en"`
	NameJA   string `json:"name_ja"`
	Desc     string `json:"desc"`
	Images   string `json:"images"`
	District string `json:"district"`
}

// SolarFoodService 应季美食查询服务。
type SolarFoodService struct {
	foodRepo *repository.FoodRepo
}

// NewSolarFoodService 构造应季美食服务。
func NewSolarFoodService(foodRepo *repository.FoodRepo) *SolarFoodService {
	return &SolarFoodService{foodRepo: foodRepo}
}

// List 查询适用节气为 term 的美食;节气名非法(不在二十四节气内)或无匹配时返回空数组,
// 前端据此整体隐藏条幅。
func (s *SolarFoodService) List(term string) ([]SolarFoodItem, error) {
	term = strings.TrimSpace(term)
	if !isValidSolarTerm(term) {
		return []SolarFoodItem{}, nil
	}
	foods, err := s.foodRepo.FindSolarFoods(term)
	if err != nil {
		return nil, err
	}
	items := make([]SolarFoodItem, 0, len(foods))
	for _, f := range foods {
		items = append(items, SolarFoodItem{
			ID:       f.ID,
			NameZH:   f.NameZH,
			NameEN:   f.NameEN,
			NameJA:   f.NameJA,
			Desc:     f.Desc,
			Images:   f.Images,
			District: f.District,
		})
	}
	return items, nil
}

// isValidSolarTerm 校验节气名是否在二十四节气全集内(拒绝逗号/空白等注入式输入)。
func isValidSolarTerm(term string) bool {
	if term == "" || strings.ContainsAny(term, ",， \t") {
		return false
	}
	for _, t := range strings.Split(SolarTermList, ",") {
		if t == term {
			return true
		}
	}
	return false
}

// ===== LLM 节气打标(cmd/solar-term-tag 批处理) =====

const (
	// solarTagTimeout 单条打标的 LLM 超时(秒)。
	solarTagTimeout = 60
	// solarTagInterval 批处理节流间隔。
	solarTagInterval = 500 * time.Millisecond
	// solarTagMaxTerms 单条美食最多关联的节气数。
	solarTagMaxTerms = 3
)

// solarTermTagOutput 单条美食的打标结果。
type solarTermTagOutput struct {
	SolarTerms string `json:"solar_terms"`
}

// SolarTermTagger 美食节气打标器(按食材与蜀地食俗判断,宁缺勿滥)。
type SolarTermTagger struct {
	foodRepo *repository.FoodRepo
	llm      *TripLLM
}

// NewSolarTermTagger 构造节气打标器。
func NewSolarTermTagger(foodRepo *repository.FoodRepo, llm *TripLLM) *SolarTermTagger {
	return &SolarTermTagger{foodRepo: foodRepo, llm: llm}
}

// TagAll 全量打标:逐条调 LLM 判断适用节气并落库。
// 幂等:solar_terms 已有值时跳过,force 为 true 时才覆盖;置信度低(空结果)落库为空串。
// 返回:成功打标数 / 留空数 / 跳过数 / 失败数。
func (t *SolarTermTagger) TagAll(ctx context.Context, only string, force bool) (tagged, empty, skip, fail int) {
	foods, err := t.foodRepo.ListAll()
	if err != nil {
		logger.Fatalf("美食列表查询失败: %v", err)
	}
	total := len(foods)
	for i, f := range foods {
		if only != "" && !strings.Contains(f.NameZH, only) {
			continue
		}
		if !force && strings.TrimSpace(f.SolarTerms) != "" {
			logger.Infof("[%d/%d] %s 已有节气标签(%s),跳过", i+1, total, f.NameZH, f.SolarTerms)
			skip++
			continue
		}
		terms, err := t.tagOne(ctx, f.NameZH, f.Desc)
		if err != nil {
			logger.Errorf("[%d/%d] %s 打标失败: %v", i+1, total, f.NameZH, err)
			fail++
			continue
		}
		if err := t.foodRepo.UpdateFields(f.ID, map[string]any{"solar_terms": terms}); err != nil {
			logger.Errorf("[%d/%d] %s 落库失败: %v", i+1, total, f.NameZH, err)
			fail++
			continue
		}
		if terms == "" {
			logger.Infof("[%d/%d] %s → 留空(无明确节气关联,宁缺勿滥)", i+1, total, f.NameZH)
			empty++
			continue
		}
		logger.Infof("[%d/%d] %s → %s", i+1, total, f.NameZH, terms)
		tagged++
		time.Sleep(solarTagInterval)
	}
	return tagged, empty, skip, fail
}

// tagOne 调 LLM 打标单条美食并清洗输出(仅保留合法节气名、去重、限 3 个)。
func (t *SolarTermTagger) tagOne(ctx context.Context, name, desc string) (string, error) {
	if t.llm == nil || !t.llm.Available() {
		return "", fmt.Errorf("LLM 未配置")
	}
	messages := []llmMessage{
		{Role: "system", Content: "你是严谨的成都美食文化编辑,只输出合法 JSON。"},
		{Role: "user", Content: solarTermPrompt(name, desc)},
	}
	reply, err := t.llm.ChatWithTimeout(ctx, solarTagTimeout, messages, 0.3, 600)
	if err != nil {
		return "", err
	}
	var out solarTermTagOutput
	if err := json.Unmarshal([]byte(extractJSONObject(reply)), &out); err != nil {
		return "", fmt.Errorf("LLM 输出解析失败: %w", err)
	}
	return normalizeSolarTerms(out.SolarTerms), nil
}

// solarTermPrompt 组装单条美食的节气判断提示词(含蜀地食俗映射示例,宁缺勿滥)。
func solarTermPrompt(name, desc string) string {
	return fmt.Sprintf(`你是成都美食文化编辑。请根据下面美食的食材属性与蜀地食俗,判断它适合在哪些二十四节气向游客推荐展示。
严格规则:
1. 只能从以下二十四节气中选择:立春,雨水,惊蛰,春分,清明,谷雨,立夏,小满,芒种,夏至,小暑,大暑,立秋,处暑,白露,秋分,寒露,霜降,立冬,小雪,大雪,冬至,小寒,大寒;
2. 参考蜀地食俗映射:简阳羊肉汤/羊肉类→冬至,大雪;汤圆→冬至(元宵在立春,雨水前后);粽子→芒种(端午前后);月饼→秋分(中秋前后);腊味/香肠→小雪,大雪,冬至,小寒;冰粉/凉糕等冷食→夏至,小暑,大暑;春卷→立春;青团/艾粑→清明;菊花茶/桂花酒酿→白露,秋分,寒露;老鸭汤→立秋,处暑;
3. 宁缺勿滥:没有明确节气食俗或食材时令关联的,输出空字符串;最多选 3 个节气,按关联度从高到低排序;
4. 严格只输出 JSON,不要输出解释、不要使用 Markdown:
{"solar_terms":"节气1,节气2"}
美食名称:%s
介绍素材:
%s`, name, capDesc(desc))
}

// normalizeSolarTerms 清洗 LLM 输出:按逗号/顿号/空格拆分,仅保留二十四节气内的合法值,
// 去重后按输入顺序取前 3 个,逗号拼接(无空格,保证 FIND_IN_SET 精确匹配)。
func normalizeSolarTerms(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ',' || r == '、' || r == ' ' || r == ';' || r == ';'
	})
	seen := make(map[string]bool, solarTagMaxTerms)
	terms := make([]string, 0, solarTagMaxTerms)
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" || seen[f] || !isValidSolarTerm(f) {
			continue
		}
		seen[f] = true
		terms = append(terms, f)
		if len(terms) >= solarTagMaxTerms {
			break
		}
	}
	return strings.Join(terms, ",")
}
