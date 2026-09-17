package model

import "time"

// CheckIn 打卡记录,可携带用户上传的游玩照片 URL 与评论。
type CheckIn struct {
	ID        uint      `gorm:"primaryKey;comment:打卡ID" json:"id"`
	UserID    uint      `gorm:"index:idx_uid_scenic,unique;comment:用户ID" json:"user_id"`
	ScenicID  uint      `gorm:"index:idx_uid_scenic,unique;comment:景点ID" json:"scenic_id"`
	PhotoURL  string    `gorm:"size:512;comment:打卡照片URL" json:"photo_url"`
	Comment   string    `gorm:"size:255;comment:打卡感言" json:"comment"`
	VisitedAt time.Time `gorm:"comment:到访时间" json:"visited_at"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
