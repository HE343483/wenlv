package model

import "time"

// FoodCard 美食名片,对应前端 FoodPage 六大风味卡片(火锅/串串香/川菜/名小吃/盖碗茶/夜宵)。
// CardKey 与前端 i18n 键 food.card.{key} 一一对应,图片由爬虫抓取后存 OSS 地址。
type FoodCard struct {
	ID        uint      `gorm:"primaryKey;comment:名片ID" json:"id"`
	CardKey   string    `gorm:"size:32;uniqueIndex;comment:名片标识(对应前端卡片key)" json:"card_key"`
	NameZH    string    `gorm:"size:64;comment:名片中文名称" json:"name_zh"`
	NameEN    string    `gorm:"size:64;comment:名片英文名称" json:"name_en"`
	Image     string    `gorm:"type:text;comment:名片配图URL(OSS地址)" json:"image"`
	Sort      int       `gorm:"default:0;comment:展示排序" json:"sort"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
