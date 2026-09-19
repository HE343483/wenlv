package model

import "time"

// FoodCategory 美食大类(川菜/名小吃/夜宵),对应前端美食名片里的大类卡片。
// 与 Food(具体菜品)区分:大类详情页介绍"这一类",不指向某道菜。
type FoodCategory struct {
	ID              uint       `gorm:"primaryKey;comment:美食大类ID" json:"id"`
	Key             string     `gorm:"size:32;uniqueIndex;comment:类别键(cuisine/snacks/nightfood)" json:"key"`
	NameZH          string     `gorm:"size:64;comment:类别中文名" json:"name_zh"`
	NameEN          string     `gorm:"size:64;comment:类别英文名" json:"name_en"`
	Intro           string     `gorm:"type:text;comment:类别概述(LLM依据维基素材改写,参考值)" json:"intro"`
	Sections        string     `gorm:"type:text;comment:类别图文段落JSON" json:"sections"`
	GalleryImages   string     `gorm:"type:text;comment:类别图集URL列表(逗号分隔,OSS)" json:"gallery_images"`
	SourceURL       string     `gorm:"size:512;comment:素材来源链接(维基条目)" json:"source_url"`
	EstimatedFields string     `gorm:"size:255;comment:参考值字段(逗号分隔)" json:"estimated_fields"`
	DataSource      string     `gorm:"size:64;comment:数据来源(如 wiki+llm+oss)" json:"data_source"`
	DataUpdatedAt   *time.Time `gorm:"comment:数据采集时间" json:"data_updated_at"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// CategorySection 类别介绍的一段(标题 + 正文 + 配图)。
type CategorySection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}
