// Package database 维护 MySQL 与 Redis 连接。
package database

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"wenlv-backend/config"
	"wenlv-backend/logger"
	"wenlv-backend/model"
)

// InitMySQL 建立 GORM 连接并自动迁移表结构。
func InitMySQL(cfg *config.Config) (*gorm.DB, error) {
	// SQL 日志统一走日志模块(见 logger.GormWriter):
	// 默认只记录慢 SQL(>=1s)与 SQL 错误,避免远程库常规查询把 WARN/ERROR 级别日志刷满;
	// 需要逐条 SQL 明细时把 LOG_LEVEL 调成 debug,明细会以 DEBUG 级别落库 log_entries。
	gormLevel := gormlogger.Warn
	if logger.ParseLevel(cfg.Log.Level) == logger.LevelDebug {
		gormLevel = gormlogger.Info
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.New(logger.GormWriter{}, gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
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
		&model.FoodCard{},
		&model.FoodCategory{},
		&model.Route{},
		&model.Favorite{},
		&model.CheckIn{},
		&model.Review{},
		&model.Photo{},
		&model.Article{},
		&model.UserMemory{},
		&model.TripPlanRecord{},
		&model.TripSetting{},
		&model.HotTopic{},
		&model.Poem{},
		&model.CultureDaily{},
		&model.QuizQuestion{},
		&model.ChatSession{},
		&model.ChatMessageRecord{},
		&model.PersonaMemory{},
	); err != nil {
		return err
	}
	return applyTableComments(db)
}

// tableComments 表级中文注释(字段注释由模型 gorm tag 维护)。
var tableComments = map[string]string{
	"users":           "用户账号表",
	"scenic_spots":    "景点表",
	"foods":           "美食表",
	"food_cards":      "美食名片表(美食页六大风味卡片)",
	"food_categories": "美食大类表(川菜/名小吃/夜宵)",
	"routes":          "精品旅游路线表",
	"favorites":       "用户收藏表(景点/美食/路线)",
	"check_ins":       "景点打卡记录表",
	"reviews":         "评价表(景点/美食/路线)",
	"photos":          "用户照片表(图墙聚合)",
	"articles":        "游记攻略表",
	"user_memories":   "AI行程用户偏好记忆表",
	"trip_plans":      "AI行程计划历史表",
	"trip_settings":   "AI行程模块设置表",
	"hot_topics":      "文旅热点资讯表(定时抓取官方文旅新闻源)",
	"poems":           "景点关联诗词表(原文人工权威录入,LLM仅生成译文与赏析)",
	"culture_daily":   "每日蜀签表(诗句/方言/冷知识三语日签)",
	"quiz_questions":  "蜀文化知识闯关题目表(景点答题得徽章,三语)",
	"chat_sessions":   "AI对话会话表(普通助手与历史人物角色统一存储,仅登录用户)",
	"chat_messages":   "AI对话消息表(按会话存储用户与助手消息)",
	"persona_memories": "角色长期记忆表(LLM异步提取的访客事实,按用户+角色存储)",
}

// idComments 自增主键 id 的注释(GORM AutoMigrate 不修改主键定义,需单独补)。
var idComments = map[string]string{
	"users":           "用户ID",
	"scenic_spots":    "景点ID",
	"foods":           "美食ID",
	"food_cards":      "名片ID",
	"food_categories": "美食大类ID",
	"routes":          "路线ID",
	"favorites":       "收藏ID",
	"check_ins":       "打卡ID",
	"reviews":         "评价ID",
	"photos":          "照片ID",
	"articles":        "游记ID",
	"trip_plans":      "行程记录ID",
	"hot_topics":      "文旅热点ID",
	"poems":           "诗词ID",
	"culture_daily":   "蜀签ID",
	"quiz_questions":  "题目ID",
	"chat_sessions":   "会话ID",
	"chat_messages":   "消息ID",
	"persona_memories": "记忆ID",
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
		logger.Fatalf("数据库自动迁移失败: %v", err)
	}
}
