package repository

import (
	"errors"

	"gorm.io/gorm"

	"wenlv-backend/model"
)

// 通用仓储错误。
var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("record already exists")
)

// QueryOptions 分页 + 过滤参数。
type QueryOptions struct {
	District string
	Tag      string
	Keyword  string
	Page     int
	PageSize int
}

const (
	// DefaultPageSize 默认每页条数。
	DefaultPageSize = 20
	// MaxPageSize 每页条数上限。
	MaxPageSize = 100
)

// ScenicRepo 景点数据访问。
type ScenicRepo struct {
	db *gorm.DB
}

// NewScenicRepo 构造景点仓储。
func NewScenicRepo(db *gorm.DB) *ScenicRepo {
	return &ScenicRepo{db: db}
}

// List 分页查询景点。
func (r *ScenicRepo) List(opts QueryOptions) ([]model.ScenicSpot, int64, error) {
	q := r.db.Model(&model.ScenicSpot{})
	q = applyQuery(q, opts)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.ScenicSpot
	if err := q.Order("id asc").Limit(pageSize(opts.PageSize)).Offset(offset(opts)).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetByID 按 ID 查询景点。
func (r *ScenicRepo) GetByID(id uint) (*model.ScenicSpot, error) {
	var s model.ScenicSpot
	if err := r.db.First(&s, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// ListAll 返回全部景点(采集批处理使用,不分页)。
func (r *ScenicRepo) ListAll() ([]model.ScenicSpot, error) {
	var items []model.ScenicSpot
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateFields 按主键更新指定字段(字段名 → 新值),用于采集结果落库。
func (r *ScenicRepo) UpdateFields(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.ScenicSpot{}).Where("id = ?", id).Updates(updates).Error
}

// FoodRepo 美食数据访问。
type FoodRepo struct {
	db *gorm.DB
}

// NewFoodRepo 构造美食仓储。
func NewFoodRepo(db *gorm.DB) *FoodRepo {
	return &FoodRepo{db: db}
}

// List 分页查询美食。
func (r *FoodRepo) List(opts QueryOptions) ([]model.Food, int64, error) {
	q := r.db.Model(&model.Food{})
	q = applyQuery(q, opts)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Food
	if err := q.Order("id asc").Limit(pageSize(opts.PageSize)).Offset(offset(opts)).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetByID 按 ID 查询美食。
func (r *FoodRepo) GetByID(id uint) (*model.Food, error) {
	var f model.Food
	if err := r.db.First(&f, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// ListAll 返回全部美食(采集批处理使用,不分页)。
func (r *FoodRepo) ListAll() ([]model.Food, error) {
	var items []model.Food
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateFields 按主键更新指定字段(字段名 → 新值),用于采集结果落库。
func (r *FoodRepo) UpdateFields(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.Food{}).Where("id = ?", id).Updates(updates).Error
}

// FoodCardRepo 美食名片数据访问。
type FoodCardRepo struct {
	db *gorm.DB
}

// NewFoodCardRepo 构造美食名片仓储。
func NewFoodCardRepo(db *gorm.DB) *FoodCardRepo {
	return &FoodCardRepo{db: db}
}

// ListAll 按排序返回全部美食名片(仅 6 条,无需分页)。
func (r *FoodCardRepo) ListAll() ([]model.FoodCard, error) {
	var items []model.FoodCard
	if err := r.db.Order("sort asc, id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FoodCategoryRepo 美食大类数据访问。
type FoodCategoryRepo struct {
	db *gorm.DB
}

// NewFoodCategoryRepo 构造美食大类仓储。
func NewFoodCategoryRepo(db *gorm.DB) *FoodCategoryRepo {
	return &FoodCategoryRepo{db: db}
}

// ListAll 返回全部美食大类(仅三个,采集批处理使用)。
func (r *FoodCategoryRepo) ListAll() ([]model.FoodCategory, error) {
	var items []model.FoodCategory
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetByKey 按类别键查询;不存在时返回 (nil, nil),由上层转换为 ErrNotFound。
func (r *FoodCategoryRepo) GetByKey(key string) (*model.FoodCategory, error) {
	var v model.FoodCategory
	if err := r.db.Where("`key` = ?", key).First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// Create 新建美食大类记录(首次采集时建空壳行,后续按主键更新)。
func (r *FoodCategoryRepo) Create(v *model.FoodCategory) error {
	return r.db.Create(v).Error
}

// UpdateFields 按主键更新指定字段(字段名 → 新值),用于采集结果落库。
func (r *FoodCategoryRepo) UpdateFields(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.FoodCategory{}).Where("id = ?", id).Updates(updates).Error
}

// RouteRepo 路线数据访问。
type RouteRepo struct {
	db *gorm.DB
}

// NewRouteRepo 构造路线仓储。
func NewRouteRepo(db *gorm.DB) *RouteRepo {
	return &RouteRepo{db: db}
}

// List 分页查询路线。
func (r *RouteRepo) List(opts QueryOptions) ([]model.Route, int64, error) {
	q := r.db.Model(&model.Route{})
	q = applyQuery(q, opts)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Route
	if err := q.Order("id asc").Limit(pageSize(opts.PageSize)).Offset(offset(opts)).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetByID 按 ID 查询路线。
func (r *RouteRepo) GetByID(id uint) (*model.Route, error) {
	var rt model.Route
	if err := r.db.First(&rt, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

// applyQuery 复用通用的过滤条件:区县精确、标签包含、关键词模糊匹配中英文名。
func applyQuery(q *gorm.DB, opts QueryOptions) *gorm.DB {
	if opts.District != "" {
		q = q.Where("district = ?", opts.District)
	}
	if opts.Tag != "" {
		q = q.Where("tags LIKE ?", "%"+opts.Tag+"%")
	}
	if opts.Keyword != "" {
		kw := "%" + opts.Keyword + "%"
		q = q.Where("name_zh LIKE ? OR name_en LIKE ?", kw, kw)
	}
	return q
}

func pageSize(ps int) int {
	if ps <= 0 || ps > MaxPageSize {
		return DefaultPageSize
	}
	return ps
}

func offset(opts QueryOptions) int {
	p := opts.Page
	if p <= 0 {
		p = 1
	}
	return (p - 1) * pageSize(opts.PageSize)
}
