package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"wenlv-backend/model"
	"wenlv-backend/service"
)

// TripToolHandler 行程模块的配套工具接口(POI / 地图 / 运行时配置 / 用户记忆)。
type TripToolHandler struct {
	settings *service.TripSettings
	amap     *service.AmapService
	xhs      *service.XHSService
	memory   *service.TripMemoryService
}

// NewTripToolHandler 构造工具处理器。
func NewTripToolHandler(settings *service.TripSettings, amap *service.AmapService,
	xhs *service.XHSService, memory *service.TripMemoryService) *TripToolHandler {
	return &TripToolHandler{settings: settings, amap: amap, xhs: xhs, memory: memory}
}

// googleService 按当前配置返回 Google 地图服务(未配置则为 nil)。
func (h *TripToolHandler) googleService() *service.GoogleMapService {
	if service.CurrentMapProvider(h.settings) != "google" {
		return nil
	}
	return service.NewGoogleMapService(h.settings)
}

// ============ POI ============

// POIDetail 获取 POI 详情。
func (h *TripToolHandler) POIDetail(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "POI ID 必填"})
		return
	}
	if svc := h.googleService(); svc != nil {
		detail, err := svc.GetPOIDetail(c.Request.Context(), poiID)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "获取POI详情成功", "data": detail})
			return
		}
	}
	detail, err := h.amap.GetPOIDetail(c.Request.Context(), poiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取POI详情失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "获取POI详情成功", "data": detail})
}

// POISearch 按关键词搜索 POI(高德/Google 自动选择)。
func (h *TripToolHandler) POISearch(c *gin.Context) {
	keywords := c.Query("keywords")
	city := c.DefaultQuery("city", "成都")
	if strings.TrimSpace(keywords) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "keywords 必填"})
		return
	}
	ctx := c.Request.Context()
	var pois []model.POIInfo
	if svc := h.googleService(); svc != nil {
		pois = svc.SearchPOI(ctx, keywords, city, true)
	} else {
		pois = h.amap.SearchPOI(ctx, keywords, city, true)
	}
	c.JSON(http.StatusOK, gin.H{"success": len(pois) > 0, "message": "搜索成功", "data": pois})
}

// POIImage 代理小红书图片:
//   - name: 按景点名取图(缓存 miss 时自动重搜新直链并立即下载)
//   - url:  代理白名单内的小红书稳定直链
func (h *TripToolHandler) POIImage(c *gin.Context) {
	name := c.Query("name")
	rawURL := c.Query("url")
	ctx := c.Request.Context()

	if name != "" {
		content, contentType, err := h.xhs.PhotoBytes(ctx, name, c.Query("city"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "未能获取 " + name + " 的景点图片"})
			return
		}
		writeImageResponse(c, content, contentType)
		return
	}
	if rawURL != "" {
		content, contentType, err := h.xhs.FetchImageBytes(ctx, rawURL)
		if err != nil {
			status := http.StatusBadGateway
			if strings.Contains(err.Error(), "仅允许") || strings.Contains(err.Error(), "协议") {
				status = http.StatusBadRequest
			}
			c.JSON(status, gin.H{"detail": err.Error()})
			return
		}
		writeImageResponse(c, content, contentType)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"detail": "必须提供 name 或 url 查询参数"})
}

// POIPhoto 按景点名返回小红书图片直链(前端兜底展示用)。
func (h *TripToolHandler) POIPhoto(c *gin.Context) {
	name := c.Query("name")
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "name 必填"})
		return
	}
	photoURL := h.xhs.PhotoURL(c.Request.Context(), name, c.Query("city"))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取图片成功",
		"data":    gin.H{"name": name, "photo_url": photoURL},
	})
}

func writeImageResponse(c *gin.Context, content []byte, contentType string) {
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, contentType, content)
}

// ============ 地图服务 ============

// MapPOI 搜索 POI。
func (h *TripToolHandler) MapPOI(c *gin.Context) {
	keywords := c.Query("keywords")
	city := c.Query("city")
	if strings.TrimSpace(keywords) == "" || strings.TrimSpace(city) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "keywords 与 city 必填"})
		return
	}
	cityLimit := c.DefaultQuery("citylimit", "true") != "false"
	ctx := c.Request.Context()
	var pois []model.POIInfo
	if svc := h.googleService(); svc != nil {
		pois = svc.SearchPOI(ctx, keywords, city, cityLimit)
	} else {
		pois = h.amap.SearchPOI(ctx, keywords, city, cityLimit)
	}
	message := "POI搜索成功"
	if len(pois) == 0 {
		message = "未获取到 POI 数据"
	}
	c.JSON(http.StatusOK, gin.H{"success": len(pois) > 0, "message": message, "data": pois})
}

// MapWeather 查询天气。
func (h *TripToolHandler) MapWeather(c *gin.Context) {
	city := c.Query("city")
	if strings.TrimSpace(city) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "city 必填"})
		return
	}
	ctx := c.Request.Context()
	var list []model.WeatherInfo
	if svc := h.googleService(); svc != nil {
		list = svc.GetWeather(ctx, city)
	}
	if len(list) == 0 {
		list = h.amap.GetWeather(ctx, city)
	}
	message := "天气查询成功"
	if len(list) == 0 {
		message = "未获取到天气数据"
	}
	c.JSON(http.StatusOK, gin.H{"success": len(list) > 0, "message": message, "data": list})
}

// MapRoute 规划路线。
func (h *TripToolHandler) MapRoute(c *gin.Context) {
	var req model.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误:起点/终点必填"})
		return
	}
	if req.RouteType == "" {
		req.RouteType = "walking"
	}
	ctx := c.Request.Context()
	var info map[string]any
	var err error
	if svc := h.googleService(); svc != nil {
		info, err = svc.PlanRoute(ctx, req.OriginAddress, req.DestinationAddress,
			req.OriginCity, req.DestinationCity, req.RouteType)
	} else {
		info, err = h.amap.PlanRoute(ctx, req.OriginAddress, req.DestinationAddress,
			req.OriginCity, req.DestinationCity, req.RouteType)
	}
	if err != nil || info == nil {
		msg := "未能获取路线数据"
		if err != nil {
			msg = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg, "data": nil})
		return
	}
	distance, _ := info["distance"].(float64)
	duration := 0
	switch v := info["duration"].(type) {
	case int:
		duration = v
	case float64:
		duration = int(v)
	case int64:
		duration = int(v)
	}
	description, _ := info["distance_text"].(string)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "路线规划成功",
		"data": model.RouteInfo{
			Distance:    distance,
			Duration:    duration,
			RouteType:   req.RouteType,
			Description: description,
		},
	})
}

// MapHealth 地图服务健康检查。
func (h *TripToolHandler) MapHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "map-service",
		"provider": service.CurrentMapProvider(h.settings),
	})
}

// ============ 运行时配置 ============

// GetSettings 获取当前运行时配置(机密字段掩码)。
func (h *TripToolHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ok", "data": h.settings.Masked()})
}

// SaveSettings 保存运行时配置并立即生效。
func (h *TripToolHandler) SaveSettings(c *gin.Context) {
	var payload map[string]string
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}
	updated := h.settings.Update(payload)
	// 配置更新后重置 Google 地理编码失败标记,允许重新尝试
	service.ResetGoogleGeoFailure()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "配置已保存并立即生效", "data": updated})
}

// ============ 用户偏好记忆 ============

func (h *TripToolHandler) memoryDisabled() bool { return !h.memory.Enabled() }

// MemoryList 查询用户有效记忆。
func (h *TripToolHandler) MemoryList(c *gin.Context) {
	userID := c.Query("user_id")
	if h.memoryDisabled() {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "记忆模块未开启", "data": []any{}})
		return
	}
	items, err := h.memory.RecallUserMemory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "查询记忆失败: " + err.Error(), "data": []any{}})
		return
	}
	if items == nil {
		items = []model.UserMemory{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": items})
}

// MemoryAddExplicit 手动添加偏好(权重高于自动提取)。
func (h *TripToolHandler) MemoryAddExplicit(c *gin.Context) {
	if h.memoryDisabled() {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "记忆模块未开启"})
		return
	}
	userID := c.Query("user_id")
	content := c.Query("content")
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(content) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "user_id 与 content 必填"})
		return
	}
	if err := h.memory.AddMemory(c.Request.Context(), userID, content, "explicit", service.TripExplicitInitWeight); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "添加失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "手动偏好添加完成"})
}

// MemoryDeleteItem 删除单条记忆。
func (h *TripToolHandler) MemoryDeleteItem(c *gin.Context) {
	if h.memoryDisabled() {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "记忆模块未开启"})
		return
	}
	userID := c.Query("user_id")
	memoryID := c.Query("memory_id")
	ok, err := h.memory.DeleteMemory(c.Request.Context(), userID, memoryID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "未找到该记忆条目"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// MemoryClear 清空用户全部记忆。
func (h *TripToolHandler) MemoryClear(c *gin.Context) {
	if h.memoryDisabled() {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "记忆模块未开启"})
		return
	}
	if err := h.memory.ClearMemory(c.Request.Context(), c.Query("user_id")); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "清空失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已清空该用户全部记忆"})
}