// 蜀文化知识闯关题目生成器:LLM 逐景点批量生成三语选择题。
// 红线:题目必须严格基于库内 scenic_spots 的 desc/desc_en/desc_ja 与
// culture_note_en/culture_note_ja 出题,解析须引用库内资料原文要点(可溯源),
// 严禁引入库外事实、数字与排名;每景点 2 道题(3 选项 1 正确)。
// 幂等:按景点已落库题量判断,凑够每景点目标题数即跳过。
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
	// quizGenTimeout 单景点题目的 LLM 生成超时(秒)。
	quizGenTimeout = 120
	// quizGenInterval 批处理节流间隔。
	quizGenInterval = 500 * time.Millisecond
	// quizPerSpot 每景点目标题数。
	quizPerSpot = 2
	// quizOptionCount 每题选项数(1 正确 + 2 干扰)。
	quizOptionCount = 3
)

// quizGenOutput 单道题的 LLM 生成结果。
type quizGenOutput struct {
	QuestionZH string   `json:"question_zh"`
	QuestionEN string   `json:"question_en"`
	QuestionJA string   `json:"question_ja"`
	OptionsZH  []string `json:"options_zh"`
	OptionsEN  []string `json:"options_en"`
	OptionsJA  []string `json:"options_ja"`
	AnswerIdx  int      `json:"answer_idx"`
	ExplainZH  string   `json:"explain_zh"`
	ExplainEN  string   `json:"explain_en"`
	ExplainJA  string   `json:"explain_ja"`
}

// QuizGenerator 蜀文化知识闯关题目批量生成器。
type QuizGenerator struct {
	repo *repository.QuizRepo
	db   *gorm.DB
	llm  *TripLLM
}

// NewQuizGenerator 构造题目生成器。
func NewQuizGenerator(repo *repository.QuizRepo, db *gorm.DB, llm *TripLLM) *QuizGenerator {
	return &QuizGenerator{repo: repo, db: db, llm: llm}
}

// Generate 逐景点批量生成题目:每景点目标 quizPerSpot 道,已够则跳过(幂等)。
// 返回 成功/跳过/失败 的题目数与景点数。
func (g *QuizGenerator) Generate(ctx context.Context) (okCount, skipCount, failCount, spotOK, spotSkip, spotFail int) {
	spots, err := g.loadSpots()
	if err != nil {
		logger.Errorf("加载库内景点失败: %v", err)
		return
	}
	counts, err := g.repo.CountGroupBySpot()
	if err != nil {
		logger.Errorf("加载已有题目统计失败: %v", err)
		return
	}

	for i, spot := range spots {
		need := quizPerSpot - int(counts[spot.ID])
		if need <= 0 {
			spotSkip++
			skipCount += int(counts[spot.ID])
			logger.Infof("题目生成跳过(景点 %s#%d): 已有 %d 道,凑够 %d 道", spot.NameZH, spot.ID, counts[spot.ID], quizPerSpot)
			continue
		}
		out, err := g.callQuizLLM(ctx, spot, need)
		if err != nil {
			spotFail++
			failCount += need
			logger.Errorf("题目生成失败(景点 %s#%d): %v", spot.NameZH, spot.ID, err)
			continue
		}
		// 校验并落库:单道题不合规直接丢弃,不影响其余题目
		spotGot := 0
		for _, q := range out {
			item, verr := toQuizQuestion(spot.ID, q)
			if verr != nil {
				logger.Warnf("题目丢弃(景点 %s#%d): %v", spot.NameZH, spot.ID, verr)
				continue
			}
			if err := g.repo.Create(item); err != nil {
				logger.Errorf("题目落库失败(景点 %s#%d): %v", spot.NameZH, spot.ID, err)
				continue
			}
			okCount++
			spotGot++
		}
		if spotGot > 0 {
			spotOK++
			logger.Infof("题目生成成功(景点 %s#%d,第 %d/%d 个): 新增 %d 道", spot.NameZH, spot.ID, i+1, len(spots), spotGot)
		} else {
			spotFail++
			failCount += need
			logger.Errorf("题目生成失败(景点 %s#%d): 无一道合规", spot.NameZH, spot.ID)
		}
		time.Sleep(quizGenInterval)
	}
	return
}

// quizSpotInfo 出题素材:库内景点多语资料原文(严格出题红线的事实来源)。
type quizSpotInfo struct {
	ID       uint
	NameZH   string
	DescZH   string
	DescEN   string
	DescJA   string
	NoteEN   string
	NoteJA   string
}

// loadSpots 加载库内全部景点的出题素材。
func (g *QuizGenerator) loadSpots() ([]quizSpotInfo, error) {
	spots := make([]model.ScenicSpot, 0)
	if err := g.db.
		Select("id", "name_zh", "desc", "desc_en", "desc_ja", "culture_note_en", "culture_note_ja").
		Order("id asc").Find(&spots).Error; err != nil {
		return nil, err
	}
	infos := make([]quizSpotInfo, 0, len(spots))
	for _, s := range spots {
		infos = append(infos, quizSpotInfo{
			ID: s.ID, NameZH: s.NameZH,
			DescZH: s.Desc, DescEN: s.DescEN, DescJA: s.DescJA,
			NoteEN: s.CultureNoteEN, NoteJA: s.CultureNoteJA,
		})
	}
	return infos, nil
}

// quizPrompt 组装单景点出题提示词:素材为库内资料原文,要求解析引用原文要点。
func quizPrompt(spot quizSpotInfo, need int) string {
	return fmt.Sprintf(`你是成都文旅项目的蜀文化内容编辑,为「蜀文化知识闯关」给景点「%s」出 %d 道三语选择题。
出题素材(库内资料原文,题目只能基于这些事实,严禁引入素材之外的事实、数字、排名或编造):
【中文介绍】%s
【英文介绍】%s
【日文介绍】%s
【文化注解(英)】%s
【文化注解(日)】%s
严格规则:
1. 每题 3 个选项(options_zh/options_en/options_ja 各 3 项,一一对应),answer_idx 为正确项下标(0-2),正确项在三个语言版本中必须是同一项;
2. 题目考察该景点的核心知识(历史文化背景/建筑与遗存/民俗趣闻),3 个选项均为可信的同类表述,错误项为相近但错误的说法,不设明显送分项;
3. 解析(explain_zh/en/ja)必须点明正确答案,并引用上述素材中的原文要点作为依据,保证可溯源;
4. 素材为空的部分不要引用;事实要点在素材中找不到支撑时,宁可选更宽泛的问法;
5. question_zh 简洁明确不超过 40 字;question_en/question_ja 与 options_zh 语义完全一致;
6. 严格只输出 JSON 数组,不要解释、不要 Markdown,数组共 %d 个元素:
[{"question_zh":"...","question_en":"...","question_ja":"...","options_zh":["","",""],"options_en":["","",""],"options_ja":["","",""],"answer_idx":0,"explain_zh":"...","explain_en":"...","explain_ja":"..."}]`,
		spot.NameZH, need,
		spot.DescZH, spot.DescEN, spot.DescJA, spot.NoteEN, spot.NoteJA,
		need)
}

// callQuizLLM 调 LLM 并解析 JSON 数组;解析失败自动重试一次(模型偶发引号未转义)。
func (g *QuizGenerator) callQuizLLM(ctx context.Context, spot quizSpotInfo, need int) ([]quizGenOutput, error) {
	if g.llm == nil || !g.llm.Available() {
		return nil, fmt.Errorf("LLM 未配置")
	}
	messages := []llmMessage{
		{Role: "system", Content: "你是严谨的蜀文化内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: quizPrompt(spot, need)},
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(quizGenInterval)
		}
		reply, err := g.llm.ChatWithEffort(ctx, quizGenTimeout, messages, 0.5, 3000, llmThinkingLevel())
		if err != nil {
			lastErr = err
			continue
		}
		var out []quizGenOutput
		// extractJSONArray 复用 trip_xhs.go 的数组提取(首 "[" 到尾 "]" 贪婪截取,
		// 兼容代码围栏场景);不能用 extractJSONObject——它按大括号截取,
		// 会把数组元素的多个对象拼成非法 JSON。
		if err := json.Unmarshal([]byte(extractJSONArray(reply)), &out); err != nil {
			lastErr = fmt.Errorf("LLM 输出解析失败: %w", err)
			continue
		}
		return out, nil
	}
	return nil, lastErr
}

// toQuizQuestion 校验单道题并转换为落库模型(不合规返回错误)。
func toQuizQuestion(spotID uint, q quizGenOutput) (*model.QuizQuestion, error) {
	q.QuestionZH = strings.TrimSpace(q.QuestionZH)
	if q.QuestionZH == "" {
		return nil, fmt.Errorf("中文题干为空")
	}
	if len(q.OptionsZH) != quizOptionCount || len(q.OptionsEN) != quizOptionCount || len(q.OptionsJA) != quizOptionCount {
		return nil, fmt.Errorf("选项数量不为 %d", quizOptionCount)
	}
	if q.AnswerIdx < 0 || q.AnswerIdx >= quizOptionCount {
		return nil, fmt.Errorf("正确项下标越界: %d", q.AnswerIdx)
	}
	for i, opt := range q.OptionsZH {
		if strings.TrimSpace(opt) == "" || strings.TrimSpace(q.OptionsEN[i]) == "" || strings.TrimSpace(q.OptionsJA[i]) == "" {
			return nil, fmt.Errorf("第 %d 个选项存在空值", i)
		}
	}
	if strings.TrimSpace(q.ExplainZH) == "" {
		return nil, fmt.Errorf("中文解析为空")
	}
	return &model.QuizQuestion{
		SpotID:     spotID,
		QuestionZH: q.QuestionZH,
		QuestionEN: strings.TrimSpace(q.QuestionEN),
		QuestionJA: strings.TrimSpace(q.QuestionJA),
		OptionsZH:  marshalQuizOptions(q.OptionsZH),
		OptionsEN:  marshalQuizOptions(q.OptionsEN),
		OptionsJA:  marshalQuizOptions(q.OptionsJA),
		AnswerIdx:  q.AnswerIdx,
		ExplainZH:  strings.TrimSpace(q.ExplainZH),
		ExplainEN:  strings.TrimSpace(q.ExplainEN),
		ExplainJA:  strings.TrimSpace(q.ExplainJA),
	}, nil
}

// marshalQuizOptions 选项数组序列化为 JSON 字符串(展示端再反序列化)。
func marshalQuizOptions(opts []string) string {
	b, err := json.Marshal(opts)
	if err != nil {
		return "[]"
	}
	return string(b)
}
