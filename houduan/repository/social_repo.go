package repository

import (
	"gorm.io/gorm"

	"wenlv-backend/model"
)

// ReviewRepo 评价数据访问。
type ReviewRepo struct {
	db *gorm.DB
}

// NewReviewRepo 构造评价仓储。
func NewReviewRepo(db *gorm.DB) *ReviewRepo {
	return &ReviewRepo{db: db}
}

// Create 新增评价。
func (r *ReviewRepo) Create(rv *model.Review) error {
	return r.db.Create(rv).Error
}

// ListByTarget 查询某目标下评价列表,并附带用户名。
func (r *ReviewRepo) ListByTarget(targetType string, targetID, page, size int) ([]model.Review, int64, error) {
	var total int64
	q := r.db.Model(&model.Review{}).Where("target_type = ? AND target_id = ?", targetType, targetID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Review
	if err := q.Order("created_at desc").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListByUser 查询用户全部评价。
func (r *ReviewRepo) ListByUser(userID uint) ([]model.Review, error) {
	var items []model.Review
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// AvgRating 目标平均评分。
func (r *ReviewRepo) AvgRating(targetType string, targetID uint) (float64, error) {
	var avg *float64
	err := r.db.Model(&model.Review{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Select("COALESCE(AVG(rating),0)").
		Scan(&avg).Error
	if avg == nil {
		return 0, err
	}
	return *avg, err
}

// DeleteOwn 删除自己的评价。
func (r *ReviewRepo) DeleteOwn(userID, reviewID uint) error {
	return r.db.Where("id = ? AND user_id = ?", reviewID, userID).Delete(&model.Review{}).Error
}

// PhotoRepo 照片数据访问。
type PhotoRepo struct {
	db *gorm.DB
}

// NewPhotoRepo 构造照片仓储。
func NewPhotoRepo(db *gorm.DB) *PhotoRepo {
	return &PhotoRepo{db: db}
}

// Create 新增照片。
func (r *PhotoRepo) Create(p *model.Photo) error {
	return r.db.Create(p).Error
}

// ListByTarget 查询目标的照片(图墙)。
func (r *PhotoRepo) ListByTarget(targetType string, targetID uint) ([]model.Photo, error) {
	var items []model.Photo
	if err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ListAll 全量照片流。
func (r *PhotoRepo) ListAll(page, size int) ([]model.Photo, int64, error) {
	var total int64
	if err := r.db.Model(&model.Photo{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Photo
	if err := r.db.Order("created_at desc").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ArticleRepo 游记数据访问。
type ArticleRepo struct {
	db *gorm.DB
}

// NewArticleRepo 构造游记仓储。
func NewArticleRepo(db *gorm.DB) *ArticleRepo {
	return &ArticleRepo{db: db}
}

// Create 新增游记。
func (r *ArticleRepo) Create(a *model.Article) error {
	return r.db.Create(a).Error
}

// List 分页查询全部游记。
func (r *ArticleRepo) List(page, size int) ([]model.Article, int64, error) {
	var total int64
	if err := r.db.Model(&model.Article{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Article
	if err := r.db.Order("created_at desc").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetByID 查询单篇游记。
func (r *ArticleRepo) GetByID(id uint) (*model.Article, error) {
	var a model.Article
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// Update 更新游记(校验作者)。
func (r *ArticleRepo) Update(a *model.Article) error {
	return r.db.Model(&model.Article{}).
		Where("id = ? AND user_id = ?", a.ID, a.UserID).
		Updates(map[string]any{"title": a.Title, "content": a.Content, "cover_url": a.CoverURL}).Error
}

// DeleteOwn 删除自己的游记。
func (r *ArticleRepo) DeleteOwn(userID, articleID uint) error {
	return r.db.Where("id = ? AND user_id = ?", articleID, userID).Delete(&model.Article{}).Error
}