package handler

import (
	"encoding/json"
	"errors"

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
	ScenicID  uint     `json:"scenic_id"`
	PhotoURL  string   `json:"photo_url"`
	Photos    []string `json:"photos"`
	Comment   string   `json:"comment"`
	VisitedAt string   `json:"visited_at"`
}

// CheckIn 打卡(至少携带一张现场照片)。
func (h *CheckInHandler) CheckIn(c *gin.Context) {
	var req checkInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.CheckIn(uid, req.ScenicID, req.PhotoURL, req.Comment, req.VisitedAt, req.Photos); err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicatedCheckIn):
			pkg.Fail(c, 409, 20001, "已经打过卡了")
		case errors.Is(err, service.ErrPhotoRequired):
			pkg.BadRequest(c, "请至少上传一张打卡照片")
		case errors.Is(err, service.ErrInvalidInput):
			pkg.BadRequest(c, "参数错误:缺少景点或照片不合法")
		default:
			pkg.ServerErrorWithErr(c, err, "打卡失败")
		}
		return
	}
	pkg.OK(c, nil)
}

// checkInItem 打卡响应:photos 由库内 JSON 字符串解析为数组,便于前端直接使用。
type checkInItem struct {
	ID        uint     `json:"id"`
	ScenicID  uint     `json:"scenic_id"`
	PhotoURL  string   `json:"photo_url"`
	Photos    []string `json:"photos"`
	Comment   string   `json:"comment"`
	VisitedAt string   `json:"visited_at"`
	CreatedAt string   `json:"created_at"`
}

// List 我的打卡列表。
func (h *CheckInHandler) List(c *gin.Context) {
	uid := middleware.GetUID(c)
	items, err := h.svc.List(uid)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "查询打卡失败")
		return
	}
	res := make([]checkInItem, 0, len(items))
	for _, it := range items {
		photos := []string{}
		if it.Photos != "" {
			_ = json.Unmarshal([]byte(it.Photos), &photos)
		}
		if len(photos) == 0 && it.PhotoURL != "" {
			photos = []string{it.PhotoURL}
		}
		res = append(res, checkInItem{
			ID:        it.ID,
			ScenicID:  it.ScenicID,
			PhotoURL:  it.PhotoURL,
			Photos:    photos,
			Comment:   it.Comment,
			VisitedAt: it.VisitedAt.Format("2006-01-02 15:04:05"),
			CreatedAt: it.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	pkg.OK(c, res)
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
