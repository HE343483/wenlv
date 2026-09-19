package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// scenicClientCacheSeconds 浏览器端缓存秒数:5 分钟内同 URL 不再请求后端,
// 由后端 Redis(TTL 见 SCENIC_CACHE_TTL_HOURS)兜住；用户强刷可立即更新。
const scenicClientCacheSeconds = 300

// ScenicExtraHandler 景点详情页实时数据(周边/交通)接口。
type ScenicExtraHandler struct {
	svc *service.ScenicExtraService
}

// NewScenicExtraHandler 构造处理器。
func NewScenicExtraHandler(svc *service.ScenicExtraService) *ScenicExtraHandler {
	return &ScenicExtraHandler{svc: svc}
}

// Around 周边推荐:GET /api/scenic/:id/around?limit=6
// 查询失败或景点无坐标时返回空数组,前端显示"暂无数据"。
func (h *ScenicExtraHandler) Around(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的景点 ID")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	items, err := h.svc.Around(c.Request.Context(), id, limit)
	if handleNotFound(c, err) {
		return
	}
	if err != nil || len(items) == 0 {
		// 高德不可用或无数据时降级为空列表;空结果不加浏览器缓存头,便于恢复后立刻可见
		pkg.OK(c, []service.AroundItem{})
		return
	}
	setScenicClientCache(c)
	pkg.OK(c, items)
}

// Transport 邻近交通站点:GET /api/scenic/:id/transport
func (h *ScenicExtraHandler) Transport(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的景点 ID")
		return
	}
	items, err := h.svc.Transport(c.Request.Context(), id)
	if handleNotFound(c, err) {
		return
	}
	if err != nil || len(items) == 0 {
		pkg.OK(c, []service.TransitStop{})
		return
	}
	setScenicClientCache(c)
	pkg.OK(c, items)
}

// setScenicClientCache 设置浏览器缓存头(仅在有数据时调用)。
func setScenicClientCache(c *gin.Context) {
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", scenicClientCacheSeconds))
}
