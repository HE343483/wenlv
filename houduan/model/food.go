package model

import "time"

// Food 美食,对应前端 FoodPage 数据。
type Food struct {
	ID       uint   `gorm:"primaryKey;comment:美食ID" json:"id"`
	NameZH   string `gorm:"size:128;comment:美食中文名称" json:"name_zh"`
	NameEN   string `gorm:"size:128;comment:美食英文名称" json:"name_en"`
	NameJA   string `gorm:"size:128;comment:美食日文名称(LLM翻译,参考值)" json:"name_ja"`
	District string `gorm:"size:64;index;comment:所属区县" json:"district"`
	Tags     string `gorm:"size:255;comment:特色标签(逗号分隔)" json:"tags"`
	Desc     string `gorm:"type:text;comment:美食介绍" json:"desc"`
	DescEN   string `gorm:"type:text;comment:美食故事版英文介绍(LLM基于中文desc改写,参考值)" json:"desc_en"`
	DescJA   string `gorm:"type:text;comment:美食故事版日语介绍(LLM基于中文desc改写,参考值)" json:"desc_ja"`
	CultureNoteEN string `gorm:"type:text;comment:饮食文化注解英文(向外国游客解释饮食文化,参考值)" json:"culture_note_en"`
	CultureNoteJA string `gorm:"type:text;comment:饮食文化注解日语(向外国游客解释饮食文化,参考值)" json:"culture_note_ja"`
	NameLiteralEN string `gorm:"size:128;comment:菜名字面直译英文(直译陷阱提示,参考值)" json:"name_literal_en"`
	IngredientsZH string `gorm:"size:255;comment:主要食材(中文,含过敏原提示,参考值)" json:"ingredients_zh"`
	IngredientsEN string `gorm:"size:255;comment:主要食材英文(含过敏原提示,参考值)" json:"ingredients_en"`
	Images   string `gorm:"type:text;comment:图片URL列表(逗号分隔)" json:"images"`
	// ===== 详情页扩展字段(高德门店采集 + LLM 参考值) =====
	Rating          string     `gorm:"size:16;comment:评分(高德事实或LLM参考值)" json:"rating"`
	Flavor          string     `gorm:"size:64;comment:风味标签(参考值)" json:"flavor"`
	SpiceLevel      string     `gorm:"size:32;comment:辣度(参考值)" json:"spice_level"`
	AvgPrice        string     `gorm:"size:32;comment:人均消费(如 人均 ¥71)" json:"avg_price"`
	Signature       string     `gorm:"size:255;comment:招牌推荐(参考值)" json:"signature"`
	RecommendScene  string     `gorm:"size:64;comment:推荐场景(参考值)" json:"recommend_scene"`
	StorySections   string     `gorm:"type:text;comment:风味故事段落JSON" json:"story_sections"`
	GalleryImages   string     `gorm:"type:text;comment:相册图片URL列表(逗号分隔)" json:"gallery_images"`
	POIName         string     `gorm:"size:128;comment:高德门店名(事实)" json:"poi_name"`
	Address         string     `gorm:"size:255;comment:门店地址(事实)" json:"address"`
	Lat             float64    `gorm:"comment:门店纬度(事实)" json:"lat"`
	Lng             float64    `gorm:"comment:门店经度(事实)" json:"lng"`
	EstimatedFields string     `gorm:"size:255;comment:参考值字段(逗号分隔,前端加标注)" json:"estimated_fields"`
	DataSource      string     `gorm:"size:32;comment:数据来源(amap/amap+llm/llm/wiki)" json:"data_source"`
	DataUpdatedAt   *time.Time `gorm:"comment:数据采集时间" json:"data_updated_at"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// FoodStorySection 风味故事的一段(标题 + 正文 + 配图)。
// 以 JSON 数组序列化后存入 Food.StorySections。
type FoodStorySection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}
