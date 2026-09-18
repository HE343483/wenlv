package service

import (
	"errors"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// ReviewItem 评价 + 作者名。
type ReviewItem struct {
	model.Review
	UserName string `json:"user_name"`
}

// ReviewService 评价业务。
type ReviewService struct {
	review *repository.ReviewRepo
	user   *repository.UserRepo
}

// NewReviewService 构造评价服务。
func NewReviewService(review *repository.ReviewRepo, user *repository.UserRepo) *ReviewService {
	return &ReviewService{review: review, user: user}
}

// Create 发表评价。
func (s *ReviewService) Create(uid uint, targetType string, targetID uint, content string, rating int) error {
	if !CheckTargetType(targetType) || targetID == 0 {
		return ErrInvalidTarget
	}
	if content == "" {
		return ErrInvalidInput
	}
	if rating < 1 || rating > 5 {
		return ErrInvalidInput
	}
	return s.review.Create(&model.Review{
		UserID: uid, TargetType: targetType, TargetID: targetID,
		Content: content, Rating: rating,
	})
}

// ListByTarget 查询目标的评价列表(附作者名)。
func (s *ReviewService) ListByTarget(targetType string, targetID, page, size int) ([]ReviewItem, int64, error) {
	items, total, err := s.review.ListByTarget(targetType, targetID, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]ReviewItem, 0, len(items))
	for _, it := range items {
		res = append(res, ReviewItem{Review: it, UserName: s.userName(it.UserID)})
	}
	return res, total, nil
}

// ListByUser 查询用户全部评价。
func (s *ReviewService) ListByUser(uid uint) ([]model.Review, error) {
	return s.review.ListByUser(uid)
}

// AvgRating 目标平均评分。
func (s *ReviewService) AvgRating(targetType string, targetID uint) (float64, error) {
	return s.review.AvgRating(targetType, targetID)
}

// Delete 删除自己的评价。
func (s *ReviewService) Delete(uid, reviewID uint) error {
	err := s.review.DeleteOwn(uid, reviewID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *ReviewService) userName(id uint) string {
	u, err := s.user.FindByID(id)
	if err != nil {
		return ""
	}
	return u.Username
}

// PhotoService 图墙照片业务。
type PhotoService struct {
	repo *repository.PhotoRepo
}

// NewPhotoService 构造照片服务。
func NewPhotoService(repo *repository.PhotoRepo) *PhotoService {
	return &PhotoService{repo: repo}
}

// Add 上传记录。
func (s *PhotoService) Add(uid uint, targetType string, targetID uint, url string) error {
	if !CheckTargetType(targetType) || targetID == 0 {
		return ErrInvalidTarget
	}
	if url == "" {
		return ErrInvalidInput
	}
	return s.repo.Create(&model.Photo{UserID: uid, TargetType: targetType, TargetID: targetID, URL: url})
}

// ListByTarget 查询目标照片。
func (s *PhotoService) ListByTarget(targetType string, targetID uint) ([]model.Photo, error) {
	if !CheckTargetType(targetType) {
		return nil, ErrInvalidTarget
	}
	return s.repo.ListByTarget(targetType, targetID)
}

// ListAll 全量照片流。
func (s *PhotoService) ListAll(page, size int) ([]model.Photo, int64, error) {
	return s.repo.ListAll(page, size)
}

// ArticleItem 游记 + 作者名。
type ArticleItem struct {
	model.Article
	UserName string `json:"user_name"`
}

// ArticleService 游记业务。
type ArticleService struct {
	repo *repository.ArticleRepo
	user *repository.UserRepo
}

// NewArticleService 构造游记服务。
func NewArticleService(repo *repository.ArticleRepo, user *repository.UserRepo) *ArticleService {
	return &ArticleService{repo: repo, user: user}
}

// Create 发布游记。
func (s *ArticleService) Create(uid uint, title, content, cover string) error {
	if title == "" || content == "" {
		return ErrInvalidInput
	}
	return s.repo.Create(&model.Article{UserID: uid, Title: title, Content: content, CoverURL: cover})
}

// List 分页查询游记(附作者名)。
func (s *ArticleService) List(page, size int) ([]ArticleItem, int64, error) {
	items, total, err := s.repo.List(page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]ArticleItem, 0, len(items))
	for _, it := range items {
		res = append(res, ArticleItem{Article: it, UserName: s.userName(it.UserID)})
	}
	return res, total, nil
}

// Get 查询游记详情。
func (s *ArticleService) Get(id uint) (*ArticleItem, error) {
	a, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ArticleItem{Article: *a, UserName: s.userName(a.UserID)}, nil
}

// Update 更新自己的游记。
func (s *ArticleService) Update(uid, id uint, title, content, cover string) error {
	if title == "" || content == "" {
		return ErrInvalidInput
	}
	return s.repo.Update(&model.Article{ID: id, UserID: uid, Title: title, Content: content, CoverURL: cover})
}

// Delete 删除自己的游记。
func (s *ArticleService) Delete(uid, articleID uint) error {
	return s.repo.DeleteOwn(uid, articleID)
}

func (s *ArticleService) userName(id uint) string {
	u, err := s.user.FindByID(id)
	if err != nil {
		return ""
	}
	return u.Username
}