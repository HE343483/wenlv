package model

import "time"

// DistinguishTargetType 收藏/打卡的目标类型。
const (
	TargetScenic = "scenic"
	TargetFood   = "food"
	TargetRoute  = "route"
)

// Favorite 用户收藏(景点/美食/路线)。
type Favorite struct {
	ID         uint      `gorm:"primaryKey;comment:收藏ID" json:"id"`
	UserID     uint      `gorm:"index:idx_uid_target,unique;comment:用户ID" json:"user_id"`
	TargetType string    `gorm:"size:32;index:idx_uid_target,unique;comment:目标类型(scenic景点-food美食-route路线)" json:"target_type"`
	TargetID   uint      `gorm:"index:idx_uid_target,unique;comment:目标ID" json:"target_id"`
	CreatedAt  time.Time `gorm:"comment:收藏时间" json:"created_at"`
}
