package handler

import (
	"github.com/gin-gonic/gin"

	"wenlv-backend/middleware"
	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// FavoriteHandler 收藏接口。
type FavoriteHandler struct {
	svc *service.FavoriteService
}

// NewFavoriteHandler 构造收藏处理器。
func NewFavoriteHandler(svc *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

type favoriteReq struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
}

// Add 添加收藏。
func (h *FavoriteHandler) Add(c *gin.Context) {
	var req favoriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Add(uid, req.TargetType, req.TargetID); err != nil {
		pkg.BadRequest(c, "目标类型不合法")
		return
	}
	pkg.OK(c, nil)
}

// Remove 取消收藏。
func (h *FavoriteHandler) Remove(c *gin.Context) {
	targetType, ok1 := c.Params.Get("targetType")
	targetID, ok2 := parseUintParam(c, "targetId")
	if !ok1 || !ok2 {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Remove(uid, targetType, targetID); err != nil {
		pkg.BadRequest(c, "目标类型不合法")
		return
	}
	pkg.OK(c, nil)
}

// List 我的收藏。
func (h *FavoriteHandler) List(c *gin.Context) {
	uid := middleware.GetUID(c)
	targetType := c.Query("target_type")
	items, err := h.svc.List(uid, targetType)
	if err != nil {
		pkg.BadRequest(c, "目标类型不合法")
		return
	}
	pkg.OK(c, items)
}

// CheckInHandler 打卡接口。
type CheckInHandler struct {
	svc *service.CheckInService
}

// NewCheckInHandler 构造打卡处理器。
func NewCheckInHandler(svc *service.CheckInService) *CheckInHandler {
	return &CheckInHandler{svc: svc}
}

type checkInReq struct {
	ScenicID  uint   `json:"scenic_id"`
	PhotoURL  string `json:"photo_url"`
	Comment   string `json:"comment"`
	VisitedAt string `json:"visited_at"`
}

// CheckIn 打卡。
func (h *CheckInHandler) CheckIn(c *gin.Context) {
	var req checkInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.CheckIn(uid, req.ScenicID, req.PhotoURL, req.Comment, req.VisitedAt); err != nil {
		switch err {
		case service.ErrDuplicatedCheckIn:
			pkg.Fail(c, 409, 20001, "已经打过卡了")
		case service.ErrInvalidInput:
			pkg.BadRequest(c, "参数错误:缺少景点")
		default:
			pkg.ServerError(c, "打卡失败")
		}
		return
	}
	pkg.OK(c, nil)
}

// List 我的打卡列表。
func (h *CheckInHandler) List(c *gin.Context) {
	uid := middleware.GetUID(c)
	items, err := h.svc.List(uid)
	if err != nil {
		pkg.ServerError(c, "查询打卡失败")
		return
	}
	pkg.OK(c, items)
}

// Remove 删除打卡记录。
func (h *CheckInHandler) Remove(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的打卡 ID")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Remove(uid, id); err != nil {
		pkg.Fail(c, 404, 40404, "打卡记录不存在")
		return
	}
	pkg.OK(c, nil)
}