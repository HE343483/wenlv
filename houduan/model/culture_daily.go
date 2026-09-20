package model

import "time"

// TableName 显式指定表名:GORM 默认复数化会把 CultureDaily 映射为 culture_dailies,
// 这里固定为 culture_daily,与表注释维护的表名保持一致。
func (CultureDaily) TableName() string { return "culture_daily" }

// CultureDaily 每日蜀签(蜀文化日签:诗句引用/四川方言/蜀文化冷知识,三语)。
// 内容由 cmd/daily-gen 批量生成落库;展示端按 天序+偏移 对总数取模轮换,DisplayDate 留作排期扩展。
type CultureDaily struct {
	ID        uint   `gorm:"primaryKey;comment:蜀签ID" json:"id"`
	ContentZH string `gorm:"type:varchar(200);comment:蜀签中文内容(不超过60字)" json:"content_zh"`
	ContentEN string `gorm:"type:text;comment:蜀签英文内容(LLM翻译,参考值)" json:"content_en"`
	ContentJA string `gorm:"type:text;comment:蜀签日文内容(LLM翻译,参考值)" json:"content_ja"`
	Category  string `gorm:"size:20;index;comment:分类(诗句/方言/冷知识)" json:"category"`
	// RelatedSpotID 关联景点:内容与该景点直接相关时可跳转详情;不相关则留空(可空索引)。
	RelatedSpotID *uint      `gorm:"index;comment:关联景点ID(scenic_spots.id,可空)" json:"related_spot_id"`
	DisplayDate   *time.Time `gorm:"type:date;uniqueIndex;comment:展示日期(空表示未排期,按序轮换)" json:"display_date"`
	CreatedAt     time.Time  `gorm:"comment:创建时间" json:"created_at"`
}
