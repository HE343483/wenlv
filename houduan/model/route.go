package model

import "time"

// Route 精选旅游路线,对应前端路线页(/home/routes)卡片与详情弹窗。
// 数据由爬虫 -routes 模式从维基百科抓取站点简介与配图后写入,
// 图片存 OSS(公开读),Stops 为 JSON 数组便于整存整取。
type Route struct {
	ID          uint      `gorm:"primaryKey;comment:路线ID" json:"id"`
	RouteKey    string    `gorm:"size:32;uniqueIndex;comment:路线标识(如 classic/panda/food/culture)" json:"route_key"`
	TitleZH     string    `gorm:"size:128;comment:路线中文名称" json:"title_zh"`
	TitleEN     string    `gorm:"size:128;comment:路线英文名称" json:"title_en"`
	TitleJA     string    `gorm:"size:128;comment:路线日语名称" json:"title_ja"`
	Theme       string    `gorm:"size:20;comment:路线主题(玩法路线为中文分类词,culture为蜀文化叙事路线)" json:"theme"`
	Description string    `gorm:"type:text;comment:路线简介" json:"description"`
	DescriptionEN string  `gorm:"type:text;comment:路线故事版英文简介(LLM基于中文description改写,参考值)" json:"description_en"`
	DescriptionJA string  `gorm:"type:text;comment:路线故事版日语简介(LLM基于中文description改写,参考值)" json:"description_ja"`
	Days        int       `gorm:"default:1;comment:建议游玩天数" json:"days"`
	Interests   string    `gorm:"size:128;comment:偏好标签(逗号分隔,对应AI行程兴趣项)" json:"interests"`
	CoverImage  string    `gorm:"type:text;comment:封面图片URL(OSS地址)" json:"cover_image"`
	Stops       string    `gorm:"type:longtext;comment:途经景点列表(JSON数组:name_zh/name_en/name_ja/desc/image)" json:"stops"`
	Sort        int       `gorm:"default:0;comment:展示排序" json:"sort"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
