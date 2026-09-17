package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

// parseQuery 从查询参数解析分页与过滤条件。
func parseQuery(c *gin.Context) repository.QueryOptions {
	return repository.QueryOptions{
		District: c.Query("district"),
		Tag:      c.Query("tag"),
		Keyword:  c.Query("keyword"),
		Page:     atoiDefault(c.Query("page"), 1),
		PageSize: atoiDefault(c.Query("page_size"), repository.DefaultPageSize),
	}
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(v), true
}

// ScenicHandler 景点查询接口。
type ScenicHandler struct {
	svc *service.ScenicService
}

// NewScenicHandler 构造景点处理器。
func NewScenicHandler(svc *service.ScenicService) *ScenicHandler {
	return &ScenicHandler{svc: svc}
}

// List 景点列表。
func (h *ScenicHandler) List(c *gin.Context) {
	opts := parseQuery(c)
	items, total, err := h.svc.List(opts)
	if err != nil {
		pkg.ServerError(c, "查询景点失败")
		return
	}
	listOK(c, items, total, opts)
}

// Get 景点详情。
func (h *ScenicHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的景点 ID")
		return
	}
	v, err := h.svc.Get(id)
	if handleNotFound(c, err) {
		return
	}
	pkg.OK(c, v)
}

// FoodHandler 美食查询接口。
type FoodHandler struct {
	svc *service.FoodService
}

// NewFoodHandler 构造美食处理器。
func NewFoodHandler(svc *service.FoodService) *FoodHandler {
	return &FoodHandler{svc: svc}
}

// List 美食列表。
func (h *FoodHandler) List(c *gin.Context) {
	opts := parseQuery(c)
	items, total, err := h.svc.List(opts)
	if err != nil {
		pkg.ServerError(c, "查询美食失败")
		return
	}
	listOK(c, items, total, opts)
}

// Get 美食详情。
func (h *FoodHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的美食 ID")
		return
	}
	v, err := h.svc.Get(id)
	if handleNotFound(c, err) {
		return
	}
	pkg.OK(c, v)
}

// RouteHandler 路线查询接口。
type RouteHandler struct {
	svc *service.RouteService
}

// NewRouteHandler 构造路线处理器。
func NewRouteHandler(svc *service.RouteService) *RouteHandler {
	return &RouteHandler{svc: svc}
}

// List 路线列表。
func (h *RouteHandler) List(c *gin.Context) {
	opts := parseQuery(c)
	items, total, err := h.svc.List(opts)
	if err != nil {
		pkg.ServerError(c, "查询路线失败")
		return
	}
	listOK(c, items, total, opts)
}

// Get 路线详情。
func (h *RouteHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的路线 ID")
		return
	}
	v, err := h.svc.Get(id)
	if handleNotFound(c, err) {
		return
	}
	pkg.OK(c, v)
}

// listOK 统一的分页响应结构。
func listOK(c *gin.Context, items interface{}, total int64, opts repository.QueryOptions) {
	pkg.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      pageNum(opts.Page),
		"page_size": pageSizeNum(opts.PageSize),
	})
}

func pageNum(p int) int {
	if p <= 0 {
		return 1
	}
	return p
}

func pageSizeNum(ps int) int {
	if ps <= 0 || ps > repository.MaxPageSize {
		return repository.DefaultPageSize
	}
	return ps
}

// handleNotFound 将"资源不存在"映射为 404。
func handleNotFound(c *gin.Context, err error) bool {
	switch err {
	case service.ErrNotFound, repository.ErrNotFound:
		pkg.Fail(c, 404, 40404, "资源不存在")
		return true
	case nil:
		return false
	default:
		return false
	}
}