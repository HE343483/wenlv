// Package database 维护 MySQL 与 Redis 连接。
package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"wenlv-backend/config"
	"wenlv-backend/model"
)

// InitMySQL 建立 GORM 连接并自动迁移表结构。
func InitMySQL(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		// 开发期保留 SQL 日志,方便排查
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// AutoMigrate 注册所有业务实体并自动建表。
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.ScenicSpot{},
		&model.Food{},
		&model.Route{},
		&model.Favorite{},
		&model.CheckIn{},
		&model.Review{},
		&model.Photo{},
		&model.Article{},
		&model.UserMemory{},
		&model.TripPlanRecord{},
		&model.TripSetting{},
	); err != nil {
		return err
	}
	return applyTableComments(db)
}

// tableComments 表级中文注释(字段注释由模型 gorm tag 维护)。
var tableComments = map[string]string{
	"users":         "用户账号表",
	"scenic_spots":  "景点表",
	"foods":         "美食表",
	"routes":        "精品旅游路线表",
	"favorites":     "用户收藏表(景点/美食/路线)",
	"check_ins":     "景点打卡记录表",
	"reviews":       "评价表(景点/美食/路线)",
	"photos":        "用户照片表(图墙聚合)",
	"articles":      "游记攻略表",
	"user_memories": "AI行程用户偏好记忆表",
	"trip_plans":    "AI行程计划历史表",
	"trip_settings": "AI行程模块设置表",
}

// idComments 自增主键 id 的注释(GORM AutoMigrate 不修改主键定义,需单独补)。
var idComments = map[string]string{
	"users":        "用户ID",
	"scenic_spots": "景点ID",
	"foods":        "美食ID",
	"routes":       "路线ID",
	"favorites":    "收藏ID",
	"check_ins":    "打卡ID",
	"reviews":      "评价ID",
	"photos":       "照片ID",
	"articles":     "游记ID",
	"trip_plans":   "行程记录ID",
}

// applyTableComments 给所有表补充表级注释,便于 Navicat 等工具阅读。
// 注意:MySQL DDL 中的 COMMENT 不支持占位符参数,这里直接拼接(注释均为内置常量,无注入风险)。
func applyTableComments(db *gorm.DB) error {
	for table, comment := range tableComments {
		if err := db.Exec("ALTER TABLE `" + table + "` COMMENT '" + comment + "'").Error; err != nil {
			return err
		}
	}
	return applyIDComments(db)
}

// applyIDComments 补充自增主键 id 列的注释。
func applyIDComments(db *gorm.DB) error {
	for table, comment := range idComments {
		var colType string
		if err := db.Raw(
			"SELECT COLUMN_TYPE FROM information_schema.COLUMNS "+
				"WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = 'id'",
			table).Scan(&colType).Error; err != nil || colType == "" {
			continue // 表或 id 列不存在时跳过
		}
		sql := "ALTER TABLE `" + table + "` MODIFY COLUMN `id` " + colType + " NOT NULL AUTO_INCREMENT COMMENT '" + comment + "'"
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	// user_memories 主键为字符串 memory_id,单独补充
	if err := db.Exec("ALTER TABLE `user_memories` MODIFY COLUMN `memory_id` varchar(64) NOT NULL COMMENT '记忆ID'").Error; err != nil {
		return err
	}
	// trip_settings 主键为字符串 key,单独补充
	if err := db.Exec("ALTER TABLE `trip_settings` MODIFY COLUMN `key` varchar(64) NOT NULL COMMENT '配置键'").Error; err != nil {
		return err
	}
	return nil
}

// MustAutoMigrate 迁移失败则直接退出,保证上线前结构一致。
func MustAutoMigrate(db *gorm.DB) {
	if err := AutoMigrate(db); err != nil {
		log.Fatalf("数据库自动迁移失败: %v", err)
	}
}
