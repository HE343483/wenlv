// story-enrich 巴蜀故事引擎批处理:基于库内中文简介,LLM 批量生成
// 景点/美食/路线的多语种(EN/JA)故事版介绍 + 文化注解,美食额外生成
// 直译英文名与中英食材标注(外国人点菜神器),路线站点 desc 同步翻译。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/story-enrich -kind scenic -limit 2 -dry-run
//	go run ./cmd/story-enrich -kind food -only 麻婆豆腐 -force
//	go run ./cmd/story-enrich -kind all
package main

import (
	"context"
	"flag"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/logger"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

func main() {
	kind := flag.String("kind", "all", "处理对象: scenic / food / route / all")
	only := flag.String("only", "", "仅处理名称包含该关键字的记录")
	limit := flag.Int("limit", 0, "仅处理前 N 个记录(0 表示全部)")
	force := flag.Bool("force", false, "已有值时也覆盖(默认只补空字段)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("story-enrich")
	defer logger.Close()

	db := mustDB(cfg)
	scenicRepo := repository.NewScenicRepo(db)
	foodRepo := repository.NewFoodRepo(db)
	routeRepo := repository.NewRouteRepo(db)

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

	enricher := service.NewStoryEnricher(scenicRepo, foodRepo, routeRepo, llm)
	ctx := context.Background()
	kinds := []string{*kind}
	if *kind == "all" {
		kinds = []string{"scenic", "food", "route"}
	}
	for _, k := range kinds {
		logger.Infof("===== 开始处理 %s (only=%q limit=%d force=%v) =====", k, *only, *limit, *force)
		switch strings.TrimSpace(k) {
		case "scenic":
			ok, fail, skip := enricher.EnrichScenics(ctx, *only, *limit, *force)
			logger.Infof("景点完成: 成功 %d / 跳过 %d / 失败 %d", ok, skip, fail)
		case "food":
			ok, fail, skip := enricher.EnrichFoods(ctx, *only, *limit, *force)
			logger.Infof("美食完成: 成功 %d / 跳过 %d / 失败 %d", ok, skip, fail)
		case "route":
			ok, fail, skip := enricher.EnrichRoutes(ctx, *only, *limit, *force)
			logger.Infof("路线完成: 成功 %d / 跳过 %d / 失败 %d", ok, skip, fail)
		default:
			logger.Fatalf("未知 kind: %s(可选 scenic/food/route/all)", k)
		}
	}
	logger.Infof("全部完成")
}

// mustDB 连接数据库并确保表结构最新(新增列需要迁移)。
func mustDB(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	database.MustAutoMigrate(db)
	return db
}
