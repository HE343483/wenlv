package model

import "time"

// Photo 用户上传照片,用于图墙聚合展示。
type Photo struct {
	ID         uint      `gorm:"primaryKey;comment:照片ID" json:"id"`
	UserID     uint      `gorm:"index;comment:上传用户ID" json:"user_id"`
	TargetType string    `gorm:"size:32;index;comment:关联目标类型(scenic景点-food美食-route路线)" json:"target_type"`
	TargetID   uint      `gorm:"index;comment:关联目标ID" json:"target_id"`
	URL        string    `gorm:"size:512;not null;comment:图片URL" json:"url"`
	CreatedAt  time.Time `gorm:"comment:上传时间" json:"created_at"`
}
