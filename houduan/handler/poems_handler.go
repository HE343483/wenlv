package handler

import (
	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// PoemsHandler 景点诗词接口(诗词地图)。
type PoemsHandler struct {
	svc *service.PoemService
}

// NewPoemsHandler 构造处理器。
func NewPoemsHandler(svc *service.PoemService) *PoemsHandler {
	return &PoemsHandler{svc: svc}
}

// ListBySpot 景点关联诗词列表:GET /api/scenic/:id/poems
// 无关联诗词时返回空数组,前端据此整体隐藏诗词卡。
func (h *PoemsHandler) ListBySpot(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的景点 ID")
		return
	}
	items, err := h.svc.ListBySpot(id)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "查询诗词失败")
		return
	}
	pkg.OK(c, items)
}
