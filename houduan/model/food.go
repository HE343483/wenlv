package model

import "time"

// Food 美食,对应前端 FoodPage 数据。
type Food struct {
	ID        uint      `gorm:"primaryKey;comment:美食ID" json:"id"`
	NameZH    string    `gorm:"size:128;comment:美食中文名称" json:"name_zh"`
	NameEN    string    `gorm:"size:128;comment:美食英文名称" json:"name_en"`
	District  string    `gorm:"size:64;index;comment:所属区县" json:"district"`
	Tags      string    `gorm:"size:255;comment:特色标签(逗号分隔)" json:"tags"`
	Desc      string    `gorm:"type:text;comment:美食介绍" json:"desc"`
	Images    string    `gorm:"type:text;comment:图片URL列表(逗号分隔)" json:"images"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
