package model

import "time"

// ScenicSpot 景点,对应前端 chengdu.ts 的 scenicSpots。
type ScenicSpot struct {
	ID       uint    `gorm:"primaryKey;comment:景点ID" json:"id"`
	NameZH   string  `gorm:"size:128;index;comment:景点中文名称" json:"name_zh"`
	NameEN   string  `gorm:"size:128;comment:景点英文名称" json:"name_en"`
	District string  `gorm:"size:64;index;comment:所属区县" json:"district"`
	Tags     string  `gorm:"size:255;comment:特色标签(逗号分隔)" json:"tags"` // 逗号分隔
	Score    float64 `gorm:"type:decimal(3,1);comment:评分" json:"score"`
	Lat      float64 `gorm:"comment:纬度" json:"lat"`
	Lng      float64 `gorm:"comment:经度" json:"lng"`
	Desc     string  `gorm:"type:text;comment:景点介绍" json:"desc"`
	Images   string  `gorm:"type:text;comment:图片URL列表(逗号分隔)" json:"images"` // 逗号分隔 URL
	// ===== 详情页扩展字段(高德 POI 采集 + LLM 参考值) =====
	Address         string     `gorm:"size:255;comment:详细地址" json:"address"`
	Tel             string     `gorm:"size:64;comment:咨询电话" json:"tel"`
	AmapPOIID       string     `gorm:"size:64;index;comment:高德POI ID(去重与刷新用)" json:"amap_poi_id"`
	OpenHours       string     `gorm:"size:255;comment:开放时间" json:"open_hours"`
	TicketPrice     string     `gorm:"size:255;comment:门票价格(参考值)" json:"ticket_price"`
	RecommendHours  string     `gorm:"size:64;comment:建议游玩时长(参考值)" json:"recommend_hours"`
	YearlyVisitors  string     `gorm:"size:64;comment:年接待游客(参考值)" json:"yearly_visitors"`
	GalleryImages   string     `gorm:"type:text;comment:相册图片URL列表(逗号分隔)" json:"gallery_images"`
	DetailSections  string     `gorm:"type:text;comment:图文详情段落JSON" json:"detail_sections"`
	EstimatedFields string     `gorm:"size:255;comment:参考值字段(逗号分隔,前端加标注)" json:"estimated_fields"`
	DataSource      string     `gorm:"size:32;comment:数据来源(amap/amap+llm/wiki/llm/manual)" json:"data_source"`
	DataUpdatedAt   *time.Time `gorm:"comment:数据采集时间" json:"data_updated_at"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// ScenicDetailSection 图文详情的一段(标题 + 正文 + 配图)。
// 以 JSON 数组序列化后存入 ScenicSpot.DetailSections。
type ScenicDetailSection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}
