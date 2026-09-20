package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// QuizHandler 蜀文化知识闯关接口(公开,无需登录)。
type QuizHandler struct {
	svc *service.QuizService
}

// NewQuizHandler 构造处理器。
func NewQuizHandler(svc *service.QuizService) *QuizHandler {
	return &QuizHandler{svc: svc}
}

// ListBySpot 景点抽题:GET /api/quiz?spot_id=7&count=5
// 随机抽该景点题目,题目不足时返回全部;spot_id 必传,无题返回空数组
// (前端据此整体隐藏答题卡)。
func (h *QuizHandler) ListBySpot(c *gin.Context) {
	spotID, err := strconv.ParseUint(c.Query("spot_id"), 10, 64)
	if err != nil || spotID <= 0 {
		pkg.BadRequest(c, "spot_id 必传且为正整数")
		return
	}
	count := 5
	if v := c.Query("count"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			pkg.BadRequest(c, "无效的题目数量")
			return
		}
		count = n
	}
	items, err := h.svc.ListBySpot(uint(spotID), count)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "查询闯关题目失败")
		return
	}
	pkg.OK(c, items)
}

// Check 校验答案:GET /api/quiz/check?question_id=&answer=
// 判分在后端完成(简版防作弊),返回 {correct, answer_idx, explain_zh/en/ja}。
func (h *QuizHandler) Check(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Query("question_id"), 10, 64)
	if err != nil || questionID <= 0 {
		pkg.BadRequest(c, "question_id 必传且为正整数")
		return
	}
	answer, err := strconv.Atoi(c.Query("answer"))
	if err != nil || answer < 0 {
		pkg.BadRequest(c, "answer 必传且为非负整数")
		return
	}
	result, err := h.svc.Check(uint(questionID), answer)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "答案校验失败")
		return
	}
	if result == nil {
		pkg.Fail(c, 404, 404, "题目不存在")
		return
	}
	pkg.OK(c, result)
}
