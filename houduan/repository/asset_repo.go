package repository

import (
	"errors"

	"gorm.io/gorm"

	"wenlv-backend/model"
)

// FavoriteRepo 收藏数据访问。
type FavoriteRepo struct {
	db *gorm.DB
}

// NewFavoriteRepo 构造收藏仓储。
func NewFavoriteRepo(db *gorm.DB) *FavoriteRepo {
	return &FavoriteRepo{db: db}
}

// Add 添加收藏,同一目标不可重复。
func (r *FavoriteRepo) Add(userID uint, targetType string, targetID uint) error {
	if t, err := r.IsFavorited(userID, targetType, targetID); err == nil && t {
		return ErrDuplicate
	}
	return r.db.Create(&model.Favorite{
		UserID: userID, TargetType: targetType, TargetID: targetID,
	}).Error
}

// Remove 取消收藏。
func (r *FavoriteRepo) Remove(userID uint, targetType string, targetID uint) error {
	return r.db.Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).Delete(&model.Favorite{}).Error
}

// IsFavorited 判断是否已收藏。
func (r *FavoriteRepo) IsFavorited(userID uint, targetType string, targetID uint) (bool, error) {
	var cnt int64
	err := r.db.Model(&model.Favorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&cnt).Error
	return cnt > 0, err
}

// ListByUser 查询用户某类型收藏列表。
func (r *FavoriteRepo) ListByUser(userID uint, targetType string) ([]model.Favorite, error) {
	var items []model.Favorite
	q := r.db.Where("user_id = ?", userID)
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if err := q.Order("id desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CheckInRepo 打卡数据访问。
type CheckInRepo struct {
	db *gorm.DB
}

// NewCheckInRepo 构造打卡仓储。
func NewCheckInRepo(db *gorm.DB) *CheckInRepo {
	return &CheckInRepo{db: db}
}

// Add 打卡,同一用户同一景点不可重复。
func (r *CheckInRepo) Add(c *model.CheckIn) error {
	var cnt int64
	if err := r.db.Model(&model.CheckIn{}).
		Where("user_id = ? AND scenic_id = ?", c.UserID, c.ScenicID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return ErrDuplicate
	}
	return r.db.Create(c).Error
}

// ListByUser 查询用户打卡记录。
func (r *CheckInRepo) ListByUser(userID uint) ([]model.CheckIn, error) {
	var items []model.CheckIn
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Remove 删除打卡记录。
func (r *CheckInRepo) Remove(userID, checkinID uint) error {
	res := r.db.Where("id = ? AND user_id = ?", checkinID, userID).Delete(&model.CheckIn{})
	if res.RowsAffected == 0 {
		return errors.Join(ErrNotFound)
	}
	return res.Error
}