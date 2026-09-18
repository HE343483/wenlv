package repository

import (
	"gorm.io/gorm"

	"wenlv-backend/model"
)

// UserMemoryRepo 用户偏好记忆数据访问(行程规划模块)。
type UserMemoryRepo struct {
	db *gorm.DB
}

// NewUserMemoryRepo 构造用户记忆仓储。
func NewUserMemoryRepo(db *gorm.DB) *UserMemoryRepo {
	return &UserMemoryRepo{db: db}
}

// ListAll 查询某用户的全部记忆。
func (r *UserMemoryRepo) ListAll(userID string) ([]model.UserMemory, error) {
	var items []model.UserMemory
	if err := r.db.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Save 新增或更新单条记忆。
func (r *UserMemoryRepo) Save(item *model.UserMemory) error {
	return r.db.Save(item).Error
}

// Delete 删除单条记忆,返回是否命中。
func (r *UserMemoryRepo) Delete(userID, memoryID string) (bool, error) {
	res := r.db.Where("user_id = ? AND memory_id = ?", userID, memoryID).Delete(&model.UserMemory{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// Clear 清空某用户全部记忆。
func (r *UserMemoryRepo) Clear(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserMemory{}).Error
}