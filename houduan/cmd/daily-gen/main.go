// daily-gen 每日蜀签批处理:LLM 逐条生成 60 条蜀文化日签(诗句引用/四川方言/蜀文化冷知识,三语)。
// 红线:诗句引用必须是真实存在的经典诗文;中文内容 ≤60 字;按 content_zh 幂等去重。
// DisplayDate 不在本程序设置(留空),展示端按 天序+偏移 对总数取模滚动轮换。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/daily-gen
//	go run ./cmd/daily-gen -count 10   # 仅生成 10 条试跑
package main

import (
	"context"
	"flag"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/logger"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

func main() {
	count := flag.Int("count", 60, "生成条数(逐条调 LLM,三类轮换)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("daily-gen")
	defer logger.Close()

	db := mustDB(cfg)
	dailyRepo := repository.NewCultureDailyRepo(db)

	// 与 main.go 一致:LLM Key 先取 .env,再被 data/runtime_settings.json 覆盖(设置页热更新过的值)
	tripSettings := service.NewTripSettings(service.TripSettingsOptions{
		Defaults: service.TripRuntimeSettings{
			ViteAmapWebKey:   cfg.Trip.AmapWebKey,
			GoogleMapsAPIKey: cfg.Trip.GoogleMapsAPIKey,
			GoogleMapsProxy:  cfg.Trip.GoogleMapsProxy,
			OpenAIAPIKey:     cfg.Trip.LLMAPIKey,
			OpenAIBaseURL:    cfg.Trip.LLMBaseURL,
			OpenAIModel:      cfg.Trip.LLMModel,
		},
		DataDir:            cfg.Trip.DataDir,
		PlannerTimeout:     cfg.Trip.PlannerTimeout,
		LLMTimeout:         cfg.Trip.LLMTimeout,
		EnableUserMemory:   cfg.Trip.EnableUserMemory,
		ImageCacheTTLHours: cfg.Trip.ImageCacheTTLHours,
		ImageCacheMaxMB:    cfg.Trip.ImageCacheMaxMB,
	})
	tripSettings.AttachDB(db) // 设置页写入的值优先取自 trip_settings 表,与主服务保持一致
	llm := service.NewTripLLM(tripSettings)
	if !llm.Available() {
		logger.Fatalf("LLM 未配置:请在 .env 配置 OPENAI_API_KEY 或在设置页配置后重试")
	}

	generator := service.NewCultureDailyGenerator(dailyRepo, db, llm)
	ok, skip, fail := generator.Generate(context.Background(), *count)
	logger.Infof("蜀签生成完成: 成功 %d / 跳过 %d / 失败 %d", ok, skip, fail)
}

// mustDB 连接数据库并确保表结构最新(新增列需要迁移)。
func mustDB(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	// 工具日志落库 log_entries(module=daily-gen)
	if err := logger.SetDB(db); err != nil {
		logger.Fatalf("日志表 log_entries 初始化失败: %v", err)
	}
	database.MustAutoMigrate(db)
	return db
}
