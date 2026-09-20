package model

import "time"

// TableName 显式指定表名:GORM 默认复数化会把 QuizQuestion 映射为 quiz_questions
// 恰好一致,这里仍显式固定,与表注释维护的表名保持一致,避免依赖默认行为。
func (QuizQuestion) TableName() string { return "quiz_questions" }

// QuizQuestion 蜀文化知识闯关题目(景点答题得徽章,三语)。
// 题目由 cmd/quiz-gen 批量生成,严格基于 scenic_spots 库内 desc/culture_note 出题,
// 解析引用库内资料原文要点,保证可溯源;答题记录存前端 localStorage,不做账号绑定。
type QuizQuestion struct {
	ID uint `gorm:"primaryKey;comment:题目ID" json:"id"`
	// SpotID 所属景点:答题入口挂在景点详情页,按景点抽题。
	SpotID     uint   `gorm:"index;comment:景点ID(scenic_spots.id)" json:"spot_id"`
	QuestionZH string `gorm:"type:varchar(255);comment:题干中文" json:"question_zh"`
	QuestionEN string `gorm:"type:text;comment:题干英文(LLM翻译,参考值)" json:"question_en"`
	QuestionJA string `gorm:"type:text;comment:题干日文(LLM翻译,参考值)" json:"question_ja"`
	// Options* 三个选项的 JSON 数组字符串,如 ["武侯祠","望江楼","青羊宫"],正确项下标为 AnswerIdx。
	OptionsZH string `gorm:"type:text;comment:中文选项JSON数组(3项)" json:"options_zh"`
	OptionsEN string `gorm:"type:text;comment:英文选项JSON数组(3项)" json:"options_en"`
	OptionsJA string `gorm:"type:text;comment:日文选项JSON数组(3项)" json:"options_ja"`
	// AnswerIdx 正确选项下标(0-2),校验答案在后端完成,前端不接触明文正确项。
	AnswerIdx int `gorm:"comment:正确选项下标(0-2)" json:"answer_idx"`
	// Explain* 答案解析(三语),内容引用库内资料原文要点,保证可溯源。
	ExplainZH string    `gorm:"type:text;comment:中文解析(引用库内资料要点)" json:"explain_zh"`
	ExplainEN string    `gorm:"type:text;comment:英文解析(LLM翻译,参考值)" json:"explain_en"`
	ExplainJA string    `gorm:"type:text;comment:日文解析(LLM翻译,参考值)" json:"explain_ja"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
