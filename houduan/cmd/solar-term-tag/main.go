// solar-term-tag 美食节气打标批处理:逐条调 LLM 按食材属性与蜀地食俗,
// 给 foods 表打"适用节气"标签(逗号分隔,如"冬至,大雪"),置信度低留空(宁缺勿滥)。
// 幂等:solar_terms 已有值时跳过,-force 时才重新打标覆盖。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/solar-term-tag
//	go run ./cmd/solar-term-tag -only 羊肉   # 仅处理名称含关键字的记录
//	go run ./cmd/solar-term-tag -force       # 已有值也重新打标
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
	only := flag.String("only", "", "仅处理名称包含该关键字的美食")
	force := flag.Bool("force", false, "已有节气值时也重新打标(默认跳过)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("solar-term-tag")
	defer logger.Close()

	db := mustDB(cfg)
	foodRepo := repository.NewFoodRepo(db)

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

	tagger := service.NewSolarTermTagger(foodRepo, llm)
	tagged, empty, skip, fail := tagger.TagAll(context.Background(), *only, *force)
	logger.Infof("节气打标完成: 打标 %d / 留空 %d / 跳过 %d / 失败 %d", tagged, empty, skip, fail)
}

// mustDB 连接数据库并确保表结构最新(foods.solar_terms 新列由 AutoMigrate 自动补齐)。
func mustDB(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	// 工具日志落库 log_entries(module=solar-term-tag)
	if err := logger.SetDB(db); err != nil {
		logger.Fatalf("日志表 log_entries 初始化失败: %v", err)
	}
	database.MustAutoMigrate(db)
	return db
}
