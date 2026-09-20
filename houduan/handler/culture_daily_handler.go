package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// CultureDailyHandler 每日蜀签接口(公开,无需登录)。
type CultureDailyHandler struct {
	svc *service.CultureDailyService
}

// NewCultureDailyHandler 构造处理器。
func NewCultureDailyHandler(svc *service.CultureDailyService) *CultureDailyHandler {
	return &CultureDailyHandler{svc: svc}
}

// Get 今日蜀签:GET /api/culture/daily?offset=N
// 返回第 (dayOfYear + offset) % 总数 条;"换一条"按钮 offset+1。
// 库内无数据时返回 data:null,前端据此整体隐藏蜀签卡。
func (h *CultureDailyHandler) Get(c *gin.Context) {
	offset := 0
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			pkg.BadRequest(c, "无效的偏移量")
			return
		}
		offset = n
	}
	item, err := h.svc.Daily(offset)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "查询今日蜀签失败")
		return
	}
	pkg.OK(c, item)
}
