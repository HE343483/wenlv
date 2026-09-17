package model

import "time"

// Article 游记/攻略。
type Article struct {
	ID        uint      `gorm:"primaryKey;comment:游记ID" json:"id"`
	UserID    uint      `gorm:"index;comment:作者用户ID" json:"user_id"`
	Title     string    `gorm:"size:128;not null;comment:游记标题" json:"title"`
	Content   string    `gorm:"type:text;comment:游记正文内容" json:"content"`
	CoverURL  string    `gorm:"size:512;comment:封面图URL" json:"cover_url"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
