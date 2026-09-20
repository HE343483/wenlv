// quiz-gen 蜀文化知识闯关批处理:LLM 逐景点基于库内 desc/culture_note 出题,
// 每景点 2 道三语选择题(3 选项 1 正确),解析引用库内资料原文要点(可溯源)。
// 幂等:按景点已落库题量判断,凑够每景点目标题数即跳过。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/quiz-gen
package main

import (
	"context"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/logger"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("quiz-gen")
	defer logger.Close()

	db := mustDB(cfg)
	quizRepo := repository.NewQuizRepo(db)

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

	generator := service.NewQuizGenerator(quizRepo, db, llm)
	ok, skip, fail, spotOK, spotSkip, spotFail := generator.Generate(context.Background())
	logger.Infof("闯关题目生成完成: 题目 成功 %d / 跳过 %d / 失败 %d;景点 成功 %d / 跳过 %d / 失败 %d",
		ok, skip, fail, spotOK, spotSkip, spotFail)
}

// mustDB 连接数据库并确保表结构最新(quiz_questions 新表由 AutoMigrate 自动建齐)。
func mustDB(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	// 工具日志落库 log_entries(module=quiz-gen)
	if err := logger.SetDB(db); err != nil {
		logger.Fatalf("日志表 log_entries 初始化失败: %v", err)
	}
	database.MustAutoMigrate(db)
	return db
}
