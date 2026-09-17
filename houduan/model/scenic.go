package model

import "time"

// ScenicSpot 景点,对应前端 chengdu.ts 的 scenicSpots。
type ScenicSpot struct {
	ID        uint      `gorm:"primaryKey;comment:景点ID" json:"id"`
	NameZH    string    `gorm:"size:128;index;comment:景点中文名称" json:"name_zh"`
	NameEN    string    `gorm:"size:128;comment:景点英文名称" json:"name_en"`
	District  string    `gorm:"size:64;index;comment:所属区县" json:"district"`
	Tags      string    `gorm:"size:255;comment:特色标签(逗号分隔)" json:"tags"` // 逗号分隔
	Score     float64   `gorm:"type:decimal(3,1);comment:评分" json:"score"`
	Lat       float64   `gorm:"comment:纬度" json:"lat"`
	Lng       float64   `gorm:"comment:经度" json:"lng"`
	Desc      string    `gorm:"type:text;comment:景点介绍" json:"desc"`
	Images    string    `gorm:"type:text;comment:图片URL列表(逗号分隔)" json:"images"` // 逗号分隔 URL
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
