package service

import (
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// QuizService 蜀文化知识闯关展示与判分服务(公开接口,无需登录)。
// 答题记录由前端 localStorage 保存(徽章),后端只负责出题与判分。
type QuizService struct {
	repo *repository.QuizRepo
}

// NewQuizService 构造闯关服务。
func NewQuizService(repo *repository.QuizRepo) *QuizService {
	return &QuizService{repo: repo}
}

// QuizCheckResult 答案校验结果:correct 判对错,answer_idx 回传正确下标,
// 解析三语返回(前端按当前语言取,缺失回落中文)。
type QuizCheckResult struct {
	QuestionID uint   `json:"question_id"`
	Correct    bool   `json:"correct"`
	AnswerIdx  int    `json:"answer_idx"`
	ExplainZH  string `json:"explain_zh"`
	ExplainEN  string `json:"explain_en"`
	ExplainJA  string `json:"explain_ja"`
}

// ListBySpot 按景点随机抽 count 道题;题目不足时返回全部,无题返回空数组
// (前端据此整体隐藏答题卡)。
func (s *QuizService) ListBySpot(spotID uint, count int) ([]model.QuizQuestion, error) {
	if count <= 0 {
		count = 5
	}
	return s.repo.FindRandomBySpot(spotID, count)
}

// Check 校验答案:answer 为用户所选选项下标,判分在后端完成(简版防作弊)。
// 题目不存在返回 nil(前端提示重试)。
func (s *QuizService) Check(questionID uint, answer int) (*QuizCheckResult, error) {
	q, err := s.repo.FindByID(questionID)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, nil
	}
	return &QuizCheckResult{
		QuestionID: q.ID,
		Correct:    answer == q.AnswerIdx,
		AnswerIdx:  q.AnswerIdx,
		ExplainZH:  q.ExplainZH,
		ExplainEN:  q.ExplainEN,
		ExplainJA:  q.ExplainJA,
	}, nil
}
