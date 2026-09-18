package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// HotTopicHandler 文旅热点接口。
type HotTopicHandler struct {
	svc *service.HotTopicService
}

// NewHotTopicHandler 构造文旅热点处理器。
func NewHotTopicHandler(svc *service.HotTopicService) *HotTopicHandler {
	return &HotTopicHandler{svc: svc}
}

// List GET /api/news/hotspots?page=1&page_size=6 分页返回文旅热点。
func (h *HotTopicHandler) List(c *gin.Context) {
	page := 1
	pageSize := 6
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			pageSize = n
		}
	}
	result, err := h.svc.ListPage(page, pageSize)
	if err != nil {
		pkg.ServerError(c, "查询文旅热点失败")
		return
	}
	pkg.OK(c, result)
}
