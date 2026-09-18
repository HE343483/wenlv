package repository

import (
	"wenlv-backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HotTopicRepo 文旅热点数据访问。
type HotTopicRepo struct {
	db *gorm.DB
}

// NewHotTopicRepo 构造文旅热点仓储。
func NewHotTopicRepo(db *gorm.DB) *HotTopicRepo {
	return &HotTopicRepo{db: db}
}

// ExistsByURL 按原文链接判断是否已抓取过。
func (r *HotTopicRepo) ExistsByURL(url string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.HotTopic{}).Where("url = ?", url).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Upsert 按原文链接去重写入(存在则更新标题/摘要/翻译字段)。
func (r *HotTopicRepo) Upsert(topic *model.HotTopic) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "url"}},
		DoUpdates: clause.AssignmentColumns([]string{"title_zh", "title_en", "title_ja", "summary_zh", "summary_en", "summary_ja", "hot", "published_at"}),
	}).Create(topic).Error
}

// ListPage 分页查询热点,按发布时间倒序,返回条目与总数。
func (r *HotTopicRepo) ListPage(page, pageSize int) ([]model.HotTopic, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 6
	}
	var total int64
	if err := r.db.Model(&model.HotTopic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.HotTopic
	if err := r.db.Order("published_at desc, id desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
