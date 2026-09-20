package model

import "time"

// Poem 景点关联诗词(诗词地图)。
// 红线:诗词原文必须人工权威录入,LLM 仅负责生成译文与白话赏析,禁止生成原文。
type Poem struct {
	ID        uint   `gorm:"primaryKey;comment:诗词ID" json:"id"`
	Title     string `gorm:"size:128;index;comment:诗词标题" json:"title"`
	Dynasty   string `gorm:"size:16;comment:朝代(如 唐/宋)" json:"dynasty"`
	Author    string `gorm:"size:64;comment:作者" json:"author"`
	ContentZH string `gorm:"type:text;comment:诗词原文(人工权威录入,禁止LLM生成)" json:"content_zh"`
	ContentEN string `gorm:"type:text;comment:英文翻译(LLM忠实翻译,参考值)" json:"content_en"`
	ContentJA string `gorm:"type:text;comment:日语翻译(LLM忠实翻译,参考值)" json:"content_ja"`
	PlainZH   string `gorm:"type:text;comment:白话赏析(LLM生成,100字内,参考值)" json:"plain_zh"`
	// RelatedSpotID 挂靠景点:一首诗只挂一个与内容真实相关的景点。
	RelatedSpotID uint      `gorm:"index;comment:关联景点ID(scenic_spots.id)" json:"related_spot_id"`
	CreatedAt     time.Time `gorm:"comment:创建时间" json:"created_at"`
}
