package handler

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"wenlv-backend/model"
	"wenlv-backend/service"
)

// distEntry 行政区划缓存条目。
type distEntry struct {
	data gin.H
	exp  time.Time
}

// poiImgEntry 高德图片兜底的内存缓存条目(空 content 表示负缓存:该景点无图)。
type poiImgEntry struct {
	content     []byte
	contentType string
	exp         time.Time
}

// TripToolHandler 行程模块的配套工具接口(POI / 地图 / 运行时配置 / 用户记忆)。
type TripToolHandler struct {
	settings *service.TripSettings
	amap     *service.AmapService
	xhs      *service.XHSService
	douyin   *service.DouyinService
	memory   *service.TripMemoryService
	// 区名→adcode 缓存(非标准区名天气兜底用)
	admu    sync.Mutex
	adcache map[string]string
	// 行政区划查询缓存(keywords|subdistrict → 响应,24h)
	distMu    sync.Mutex
	distCache map[string]distEntry
	// 高德图片兜底缓存(name|city → 图片字节,24h;负缓存 10min)
	amapImgMu    sync.Mutex
	amapImgCache map[string]poiImgEntry
}

// NewTripToolHandler 构造工具处理器。
func NewTripToolHandler(settings *service.TripSettings, amap *service.AmapService,
	xhs *service.XHSService, douyin *service.DouyinService, memory *service.TripMemoryService) *TripToolHandler {
	return &TripToolHandler{
		settings:     settings,
		amap:         amap,
		xhs:          xhs,
		douyin:       douyin,
		memory:       memory,
		distCache:    map[string]distEntry{},
		amapImgCache: map[string]poiImgEntry{},
	}
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

// POIImage 代理景点图片:
//   - name: 按景点名取图(缓存 miss 时自动重搜新直链并立即下载)
//   - url:  代理白名单内的小红书/抖音稳定直链
//   - source: 图片来源(xhs=小红书 / douyin=抖音 / map=高德);缺省按 xhs 处理,保持旧客户端兼容
func (h *TripToolHandler) POIImage(c *gin.Context) {
	name := c.Query("name")
	rawURL := c.Query("url")
	source := normalizeImageSource(c.Query("source"))
	ctx := c.Request.Context()

	if name != "" {
		content, contentType, err := h.photoBytesBySource(ctx, source, name, c.Query("city"))
		if err != nil {
			// 所选内容源不可用(风控/Cookie 失效/无结果)时回退高德 POI 图片
			if fbContent, fbType, fbErr := h.amapPhotoBytes(ctx, name, c.Query("city")); fbErr == nil {
				writeImageResponse(c, fbContent, fbType)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"detail": "未能获取 " + name + " 的景点图片"})
			return
		}
		writeImageResponse(c, content, contentType)
		return
	}
	if rawURL != "" {
		var content []byte
		var contentType string
		var err error
		if source == "douyin" {
			content, contentType, err = h.douyin.FetchImageBytes(ctx, rawURL)
		} else {
			content, contentType, err = h.xhs.FetchImageBytes(ctx, rawURL)
		}
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

// normalizeImageSource 归一化图片来源参数:仅 douyin/map 例外,其余(含空值)按 xhs 处理。
func normalizeImageSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "douyin", "map":
		return strings.ToLower(strings.TrimSpace(source))
	default:
		return "xhs"
	}
}

// photoBytesBySource 按图片来源取景点图片字节。
func (h *TripToolHandler) photoBytesBySource(ctx context.Context, source, name, city string) ([]byte, string, error) {
	if source == "douyin" {
		return h.douyin.PhotoBytes(ctx, name, city)
	}
	return h.xhs.PhotoBytes(ctx, name, city)
}

// PhotoBytesWithFallback 取景点图片字节,内容源失败时回退高德 POI 图片。
// 供行程图片 OSS 上传器复用(与 /api/poi/image 代理同一套缓存与兜底逻辑)。
func (h *TripToolHandler) PhotoBytesWithFallback(ctx context.Context, source, name, city string) ([]byte, string, error) {
	content, contentType, err := h.photoBytesBySource(ctx, normalizeImageSource(source), name, city)
	if err != nil {
		return h.amapPhotoBytes(ctx, name, city)
	}
	return content, contentType, nil
}

// amapPhotoBytes 高德图片兜底:按名称+城市检索 POI,取详情首图并下载字节。
// 结果带内存缓存(命中 24h,负缓存 10min),避免同一景点反复调用高德接口。
func (h *TripToolHandler) amapPhotoBytes(ctx context.Context, name, city string) ([]byte, string, error) {
	key := name + "|" + city
	h.amapImgMu.Lock()
	if e, ok := h.amapImgCache[key]; ok && time.Now().Before(e.exp) {
		h.amapImgMu.Unlock()
		if len(e.content) == 0 {
			return nil, "", fmt.Errorf("负缓存:该景点无高德图片")
		}
		return e.content, e.contentType, nil
	}
	h.amapImgMu.Unlock()

	content, contentType, err := h.fetchAmapPhotoBytes(ctx, name, city)

	h.amapImgMu.Lock()
	if err != nil {
		// 负缓存 10 分钟,避免前端并发重试放大高德调用量
		h.amapImgCache[key] = poiImgEntry{exp: time.Now().Add(10 * time.Minute)}
	} else {
		h.amapImgCache[key] = poiImgEntry{content: content, contentType: contentType, exp: time.Now().Add(24 * time.Hour)}
	}
	// 控制缓存条目数
	if len(h.amapImgCache) > 1024 {
		for k, v := range h.amapImgCache {
			if time.Now().After(v.exp) {
				delete(h.amapImgCache, k)
			}
		}
	}
	h.amapImgMu.Unlock()
	return content, contentType, err
}

func (h *TripToolHandler) fetchAmapPhotoBytes(ctx context.Context, name, city string) ([]byte, string, error) {
	pois := h.amap.SearchPOI(ctx, name, city, true)
	if len(pois) == 0 {
		return nil, "", fmt.Errorf("高德未检索到该景点")
	}
	detail, err := h.amap.GetPOIDetail(ctx, pois[0].ID)
	if err != nil {
		return nil, "", err
	}
	if len(detail.Photos) == 0 {
		return nil, "", fmt.Errorf("高德 POI 无图片")
	}
	client := &http.Client{Timeout: 15 * time.Second}
	for _, photoURL := range detail.Photos {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, photoURL, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK || len(data) == 0 {
			continue
		}
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/") || contentType == "" {
			contentType = "image/jpeg"
		}
		return data, contentType, nil
	}
	return nil, "", fmt.Errorf("高德图片下载失败")
}

// POIPhoto 按景点名返回景点图片直链(前端兜底展示用)。
func (h *TripToolHandler) POIPhoto(c *gin.Context) {
	name := c.Query("name")
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "name 必填"})
		return
	}
	ctx := c.Request.Context()
	city := c.Query("city")
	var photoURL string
	if normalizeImageSource(c.Query("source")) == "douyin" {
		photoURL = h.douyin.PhotoURL(ctx, name, city)
	} else {
		photoURL = h.xhs.PhotoURL(ctx, name, city)
	}
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

// MapIPLocate 按客户端 IP 定位省市(高德 IP 定位代理)。
// 用于天气按钮自动定位访客所在城市;境外/内网 IP 或失败时返回 success=false,前端降级默认城市。
func (h *TripToolHandler) MapIPLocate(c *gin.Context) {
	loc := h.amap.IPLocate(c.Request.Context(), clientPublicIP(c))
	if loc == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "IP 定位失败", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP 定位成功",
		"data":    gin.H{"province": loc.Province, "city": loc.City, "adcode": loc.Adcode},
	})
}

// clientPublicIP 提取客户端公网 IP:优先反向代理头,取不到再回退 RemoteAddr。
func clientPublicIP(c *gin.Context) string {
	if ip := strings.TrimSpace(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
			return first
		}
	}
	ip := strings.TrimSpace(c.Request.RemoteAddr)
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}

// MapWeather 查询天气。
func (h *TripToolHandler) MapWeather(c *gin.Context) {
	city := c.Query("city")
	adcode := strings.TrimSpace(c.Query("adcode"))
	if strings.TrimSpace(city) == "" && adcode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "city 或 adcode 必填"})
		return
	}
	ctx := c.Request.Context()
	var list []model.WeatherInfo
	// 前端传 adcode 时直接按 adcode 查(区县级最准,行政区划选择器走此路径)
	if adcode != "" {
		list = h.amap.GetWeather(ctx, adcode)
	}
	if svc := h.googleService(); svc != nil && len(list) == 0 {
		list = svc.GetWeather(ctx, city)
	}
	if len(list) == 0 {
		list = h.amap.GetWeather(ctx, city)
	}
	// 高新区/天府新区等非标准区名直查无结果,地理编码转 adcode 后重试(结果缓存)
	if len(list) == 0 {
		adcode := ""
		h.admu.Lock()
		if h.adcache != nil {
			adcode = h.adcache[city]
		}
		h.admu.Unlock()
		if adcode == "" {
			adcode = h.amap.GeocodeAdcode(ctx, city, "成都市")
			if adcode != "" {
				h.admu.Lock()
				if h.adcache == nil {
					h.adcache = map[string]string{}
				}
				h.adcache[city] = adcode
				h.admu.Unlock()
			}
		}
		if adcode != "" {
			list = h.amap.GetWeather(ctx, adcode)
		}
	}
	message := "天气查询成功"
	if len(list) == 0 {
		message = "未获取到天气数据"
	}
	c.JSON(http.StatusOK, gin.H{"success": len(list) > 0, "message": message, "data": list})
}

// MapDistricts 行政区划逐级查询(市→区县→镇/街道),代理高德行政区划接口,结果缓存 24h。
func (h *TripToolHandler) MapDistricts(c *gin.Context) {
	keywords := strings.TrimSpace(c.DefaultQuery("keywords", "四川省"))
	sub := strings.TrimSpace(c.DefaultQuery("subdistrict", "1"))
	if sub != "1" && sub != "2" && sub != "3" {
		sub = "1"
	}
	cacheKey := keywords + "|" + sub

	h.distMu.Lock()
	if h.distCache != nil {
		if e, ok := h.distCache[cacheKey]; ok && time.Now().Before(e.exp) {
			h.distMu.Unlock()
			c.JSON(http.StatusOK, e.data)
			return
		}
	}
	h.distMu.Unlock()

	data, err := h.amap.GetDistricts(c.Request.Context(), keywords, sub)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}
	resp := gin.H{"success": true, "message": "行政区划查询成功", "data": data}
	h.distMu.Lock()
	if h.distCache == nil {
		h.distCache = map[string]distEntry{}
	}
	h.distCache[cacheKey] = distEntry{data: resp, exp: time.Now().Add(24 * time.Hour)}
	h.distMu.Unlock()
	c.JSON(http.StatusOK, resp)
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
		"status":   "healthy",
		"service":  "map-service",
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
