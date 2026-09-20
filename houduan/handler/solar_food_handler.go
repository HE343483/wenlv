package handler

import (
	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// SolarFoodHandler 节气蜀俗接口(公开,无需登录)。
type SolarFoodHandler struct {
	svc *service.SolarFoodService
}

// NewSolarFoodHandler 构造处理器。
func NewSolarFoodHandler(svc *service.SolarFoodService) *SolarFoodHandler {
	return &SolarFoodHandler{svc: svc}
}

// List 应季美食:GET /api/culture/solar-food?term=冬至
// 返回 solar_terms 包含该节气的美食精简列表;节气名非法/无匹配时返回空数组,
// 前端据此整体隐藏条幅(不留空块)。
func (h *SolarFoodHandler) List(c *gin.Context) {
	term := c.Query("term")
	items, err := h.svc.List(term)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "查询应季美食失败")
		return
	}
	pkg.OK(c, items)
}
