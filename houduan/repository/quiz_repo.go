package repository

import (
	"gorm.io/gorm"

	"wenlv-backend/model"
)

// QuizRepo 蜀文化知识闯关题目数据访问。
type QuizRepo struct {
	db *gorm.DB
}

// NewQuizRepo 构造题目仓储。
func NewQuizRepo(db *gorm.DB) *QuizRepo {
	return &QuizRepo{db: db}
}

// CountBySpot 单景点题目数量(生成器幂等判断用)。
func (r *QuizRepo) CountBySpot(spotID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.QuizQuestion{}).
		Where("spot_id = ?", spotID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountGroupBySpot 全部景点的题目数量映射(生成器幂等判断用)。
func (r *QuizRepo) CountGroupBySpot() (map[uint]int64, error) {
	rows := make([]struct {
		SpotID uint  `gorm:"column:spot_id"`
		Total  int64 `gorm:"column:total"`
	}, 0)
	if err := r.db.Model(&model.QuizQuestion{}).
		Select("spot_id, COUNT(*) AS total").
		Group("spot_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[uint]int64, len(rows))
	for _, row := range rows {
		m[row.SpotID] = row.Total
	}
	return m, nil
}

// FindRandomBySpot 按景点随机抽题(MySQL ORDER BY RAND(),题量小无性能压力)。
// count 传 0 时抽全部。
func (r *QuizRepo) FindRandomBySpot(spotID uint, count int) ([]model.QuizQuestion, error) {
	items := make([]model.QuizQuestion, 0)
	q := r.db.Where("spot_id = ?", spotID).Order("RAND()")
	if count > 0 {
		q = q.Limit(count)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID 按 ID 取一条(答案校验用)。
func (r *QuizRepo) FindByID(id uint) (*model.QuizQuestion, error) {
	var item model.QuizQuestion
	if err := r.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// Create 落库一条题目。
func (r *QuizRepo) Create(item *model.QuizQuestion) error {
	return r.db.Create(item).Error
}
