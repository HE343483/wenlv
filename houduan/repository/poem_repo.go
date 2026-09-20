package repository

import (
	"gorm.io/gorm"

	"wenlv-backend/model"
)

// PoemRepo 景点诗词数据访问。
type PoemRepo struct {
	db *gorm.DB
}

// NewPoemRepo 构造诗词仓储。
func NewPoemRepo(db *gorm.DB) *PoemRepo {
	return &PoemRepo{db: db}
}

// ListBySpot 查询景点关联诗词(按入库顺序,无关联时返回空切片)。
func (r *PoemRepo) ListBySpot(spotID uint) ([]model.Poem, error) {
	items := make([]model.Poem, 0)
	if err := r.db.Where("related_spot_id = ?", spotID).Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
