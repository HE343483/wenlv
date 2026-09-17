package model

import "time"

// Review 评价:针对景点/美食/路线的评论与星级评分。
type Review struct {
	ID         uint      `gorm:"primaryKey;comment:评价ID" json:"id"`
	UserID     uint      `gorm:"index:idx_target;comment:用户ID" json:"user_id"`
	TargetType string    `gorm:"size:32;index:idx_target;comment:目标类型(scenic景点-food美食-route路线)" json:"target_type"`
	TargetID   uint      `gorm:"index:idx_target;comment:目标ID" json:"target_id"`
	Content    string    `gorm:"size:1000;comment:评论内容" json:"content"`
	Rating     int       `gorm:"type:tinyint;comment:星级评分(1-5)" json:"rating"` // 1-5
	CreatedAt  time.Time `gorm:"comment:评价时间" json:"created_at"`
}
