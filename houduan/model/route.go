package model

import "time"

// Route 精品旅游路线。
type Route struct {
	ID         uint      `gorm:"primaryKey;comment:路线ID" json:"id"`
	TitleZH    string    `gorm:"size:128;comment:路线中文名称" json:"title_zh"`
	TitleEN    string    `gorm:"size:128;comment:路线英文名称" json:"title_en"`
	Theme      string    `gorm:"size:64;comment:路线主题" json:"theme"`
	Stops      string    `gorm:"type:text;comment:途经站点列表(JSON数组)" json:"stops"` // JSON 数组:各站点
	Duration   string    `gorm:"size:64;comment:建议游玩时长" json:"duration"`
	Difficulty string    `gorm:"size:32;comment:路线难度" json:"difficulty"`
	CreatedAt  time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
