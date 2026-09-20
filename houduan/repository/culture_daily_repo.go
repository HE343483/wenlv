package repository

import (
	"gorm.io/gorm"

	"wenlv-backend/model"
)

// CultureDailyRepo 每日蜀签数据访问。
type CultureDailyRepo struct {
	db *gorm.DB
}

// NewCultureDailyRepo 构造蜀签仓储。
func NewCultureDailyRepo(db *gorm.DB) *CultureDailyRepo {
	return &CultureDailyRepo{db: db}
}

// Count 蜀签总数。
func (r *CultureDailyRepo) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&model.CultureDaily{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindByIndex 按序号取一条(按 id 升序为固定顺序,序号由展示端按 天序+偏移 取模计算)。
// idx 越界(理论上不会,调用方先取模)时返回 nil, nil。
func (r *CultureDailyRepo) FindByIndex(idx int) (*model.CultureDaily, error) {
	if idx < 0 {
		return nil, nil
	}
	var item model.CultureDaily
	err := r.db.Order("id asc").Offset(idx).Limit(1).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// ExistsByContentZH 按中文内容判重(幂等生成用)。
func (r *CultureDailyRepo) ExistsByContentZH(content string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.CultureDaily{}).
		Where("content_zh = ?", content).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create 落库一条蜀签。
func (r *CultureDailyRepo) Create(item *model.CultureDaily) error {
	return r.db.Create(item).Error
}
