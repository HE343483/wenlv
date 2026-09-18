package model

import "time"

// HotTopic 文旅热点资讯,由后端定时抓取官方文旅新闻源(四川省文旅厅)自动入库。
type HotTopic struct {
	ID          uint      `gorm:"primaryKey;comment:热点ID" json:"id"`
	TitleZH     string    `gorm:"size:512;comment:标题(中文)" json:"title_zh"`
	TitleEN     string    `gorm:"size:512;comment:标题(英文,LLM翻译)" json:"title_en"`
	TitleJA     string    `gorm:"size:512;comment:标题(日文,LLM翻译)" json:"title_ja"`
	SummaryZH   string    `gorm:"type:text;comment:摘要(中文,取自正文首段)" json:"summary_zh"`
	SummaryEN   string    `gorm:"type:text;comment:摘要(英文,LLM翻译)" json:"summary_en"`
	SummaryJA   string    `gorm:"type:text;comment:摘要(日文,LLM翻译)" json:"summary_ja"`
	SourceZH    string    `gorm:"size:128;comment:来源(中文)" json:"source_zh"`
	SourceEN    string    `gorm:"size:128;comment:来源(英文)" json:"source_en"`
	URL         string    `gorm:"size:512;uniqueIndex;comment:原文链接" json:"url"`
	Hot         bool      `gorm:"comment:是否热点(标题命中成都相关关键词)" json:"hot"`
	PublishedAt time.Time `gorm:"index;comment:发布时间" json:"published_at"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// TableName 显式指定表名。
func (HotTopic) TableName() string { return "hot_topics" }
