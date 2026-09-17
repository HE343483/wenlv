package handler

import (
	"github.com/gin-gonic/gin"

	"wenlv-backend/middleware"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

// ReviewHandler 评价接口。
type ReviewHandler struct {
	svc *service.ReviewService
}

// NewReviewHandler 构造评价处理器。
func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

type reviewReq struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Rating     int    `json:"rating" binding:"required"`
}

// Create 发表评价。
func (h *ReviewHandler) Create(c *gin.Context) {
	var req reviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误:target/评分/内容必填")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Create(uid, req.TargetType, req.TargetID, req.Content, req.Rating); err != nil {
		switch err {
		case service.ErrInvalidTarget:
			pkg.BadRequest(c, "目标类型不合法")
		case service.ErrInvalidInput:
			pkg.BadRequest(c, "评分需在 1-5 且内容不能为空")
		default:
			pkg.ServerError(c, "发表评价失败")
		}
		return
	}
	pkg.OK(c, nil)
}

// List 目标的评价列表。
func (h *ReviewHandler) List(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := atoiQueryUint(c, "target_id")
	page := atoiDefault(c.Query("page"), 1)
	size := atoiDefault(c.Query("page_size"), repository.DefaultPageSize)
	if targetType == "" || targetID == 0 {
		pkg.BadRequest(c, "target_type 与 target_id 必填")
		return
	}
	items, total, err := h.svc.ListByTarget(targetType, int(targetID), page, size)
	if err != nil {
		pkg.ServerError(c, "查询评价失败")
		return
	}
	listOK(c, items, total, repository.QueryOptions{Page: page, PageSize: size})
}

// Mine 我的评价。
func (h *ReviewHandler) Mine(c *gin.Context) {
	uid := middleware.GetUID(c)
	items, err := h.svc.ListByUser(uid)
	if err != nil {
		pkg.ServerError(c, "查询评价失败")
		return
	}
	pkg.OK(c, items)
}

// Summary 目标平均评分。
func (h *ReviewHandler) Summary(c *gin.Context) {
	targetType, ok1 := c.Params.Get("targetType")
	targetID, ok2 := parseUintParam(c, "targetId")
	if !ok1 || !ok2 {
		pkg.BadRequest(c, "参数错误")
		return
	}
	avg, err := h.svc.AvgRating(targetType, targetID)
	if err != nil {
		pkg.ServerError(c, "查询评分失败")
		return
	}
	pkg.OK(c, gin.H{"target_type": targetType, "target_id": targetID, "avg_rating": round1(avg)})
}

// Delete 删除我的评价。
func (h *ReviewHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的评价 ID")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Delete(uid, id); err != nil {
		pkg.Fail(c, 404, 40404, "评价不存在")
		return
	}
	pkg.OK(c, nil)
}

// PhotoHandler 图墙照片接口。
type PhotoHandler struct {
	svc *service.PhotoService
}

// NewPhotoHandler 构造照片处理器。
func NewPhotoHandler(svc *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{svc: svc}
}

type photoReq struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	URL        string `json:"url" binding:"required"`
}

// Add 上传照片记录。
func (h *PhotoHandler) Add(c *gin.Context) {
	var req photoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Add(uid, req.TargetType, req.TargetID, req.URL); err != nil {
		pkg.BadRequest(c, "目标类型或链接不合法")
		return
	}
	pkg.OK(c, nil)
}

// ListByTarget 目标的照片。
func (h *PhotoHandler) ListByTarget(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := atoiQueryUint(c, "target_id")
	if targetType == "" || targetID == 0 {
		pkg.BadRequest(c, "target_type 与 target_id 必填")
		return
	}
	items, err := h.svc.ListByTarget(targetType, targetID)
	if err != nil {
		pkg.BadRequest(c, "目标类型不合法")
		return
	}
	pkg.OK(c, items)
}

// Wall 全量照片流。
func (h *PhotoHandler) Wall(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	size := atoiDefault(c.Query("page_size"), repository.DefaultPageSize)
	items, total, err := h.svc.ListAll(page, size)
	if err != nil {
		pkg.ServerError(c, "查询照片失败")
		return
	}
	listOK(c, items, total, repository.QueryOptions{Page: page, PageSize: size})
}

// ArticleHandler 游记接口。
type ArticleHandler struct {
	svc *service.ArticleService
}

// NewArticleHandler 构造游记处理器。
func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

type articleReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	CoverURL string `json:"cover_url"`
}

// Create 发布游记。
func (h *ArticleHandler) Create(c *gin.Context) {
	var req articleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误:标题与内容必填")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Create(uid, req.Title, req.Content, req.CoverURL); err != nil {
		pkg.BadRequest(c, "标题与内容不能为空")
		return
	}
	pkg.OK(c, nil)
}

// List 游记列表。
func (h *ArticleHandler) List(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	size := atoiDefault(c.Query("page_size"), repository.DefaultPageSize)
	items, total, err := h.svc.List(page, size)
	if err != nil {
		pkg.ServerError(c, "查询游记失败")
		return
	}
	listOK(c, items, total, repository.QueryOptions{Page: page, PageSize: size})
}

// Get 游记详情。
func (h *ArticleHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的游记 ID")
		return
	}
	v, err := h.svc.Get(id)
	if handleNotFound(c, err) {
		return
	}
	pkg.OK(c, v)
}

// Update 更新自己的游记。
func (h *ArticleHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的游记 ID")
		return
	}
	var req articleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Update(uid, id, req.Title, req.Content, req.CoverURL); err != nil {
		pkg.BadRequest(c, "标题与内容不能为空或无权修改")
		return
	}
	pkg.OK(c, nil)
}

// Delete 删除自己的游记。
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的游记 ID")
		return
	}
	uid := middleware.GetUID(c)
	if err := h.svc.Delete(uid, id); err != nil {
		pkg.Fail(c, 404, 40404, "游记不存在")
		return
	}
	pkg.OK(c, nil)
}

func atoiQueryUint(c *gin.Context, key string) uint {
	return uint(atoiDefault(c.Query(key), 0))
}

func round1(f float64) float64 {
	return float64(int64(f*10+0.5)) / 10
}