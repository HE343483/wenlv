// food-category-enrich 美食大类(川菜/名小吃/夜宵)详情数据采集批处理。
// 数据源:中文维基百科条目全文(文本唯一素材,需代理,复用 COMMONS_PROXY) + LLM 压缩改写
// + 该类下已采集菜品的 OSS 图 / Wikimedia Commons 关键词补图。
// 产出:字段写入 MySQL food_categories,Commons 补图上传 OSS(foodcat/<key>/)。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/food-category-enrich -dry-run
//	go run ./cmd/food-category-enrich -only cuisine -force
//	go run ./cmd/food-category-enrich
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

func main() {
	only := flag.String("only", "", "仅处理指定类别键(cuisine/snacks/nightfood)")
	dryRun := flag.Bool("dry-run", false, "只抓维基素材并打印标题/正文字数/该类图片张数;不调 LLM、不上传 OSS、不写库")
	force := flag.Bool("force", false, "已有值时也覆盖(默认只补空字段)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	db := mustDB(cfg, !*dryRun)
	catRepo := repository.NewFoodCategoryRepo(db)
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
	signer := pkg.NewOssSigner(&pkg.OssConfig{
		Endpoint:  os.Getenv("OSS_ENDPOINT"),
		AccessKey: os.Getenv("OSS_ACCESS_KEY"),
		SecretKey: os.Getenv("OSS_SECRET_KEY"),
		Bucket:    os.Getenv("OSS_BUCKET"),
	})
	if !signer.Configured() {
		log.Println("提示:OSS 未配置,将跳过 Commons 补图")
	}
	if !llm.Available() {
		log.Println("提示:LLM 未配置,将跳过类别文本改写(图片与来源仍会落库)")
	}

	seeds := service.CategorySeeds()
	if *only != "" {
		filtered := make([]service.CategorySeed, 0, len(seeds))
		for _, s := range seeds {
			if s.Key == *only {
				filtered = append(filtered, s)
			}
		}
		seeds = filtered
	}
	log.Printf("待处理美食大类 %d 个(dry-run=%v, force=%v)", len(seeds), *dryRun, *force)
	if *dryRun {
		log.Println("dry-run 模式:只抓维基素材与统计图片,不调 LLM、不上传 OSS、不写库")
	}

	enricher := service.NewFoodCategoryEnricher(catRepo, foodRepo, llm, signer)
	ctx := context.Background()
	okCnt, failCnt := 0, 0
	for i, seed := range seeds {
		time.Sleep(300 * time.Millisecond) // 控制外部接口 QPS:dry-run 与正式跑都节流
		log.Printf("[%d/%d] %s(%s)", i+1, len(seeds), seed.NameZH, seed.Key)
		res, err := enricher.Enrich(ctx, seed, service.FoodCategoryEnrichOptions{
			WithImages: !*dryRun,
			WithLLM:    !*dryRun,
			Force:      *force,
			DryRun:     *dryRun,
		})
		if err != nil {
			log.Printf("    采集失败: %v", err)
			failCnt++
			continue
		}
		if *dryRun {
			log.Printf("    维基标题=%s 正文字数=%d 该类图片=%d 张 备注=%s",
				res.WikiTitle, res.WikiTextLen, res.DishImageCount, res.Note)
			okCnt++
			continue
		}
		log.Printf("    维基标题=%s 正文=%d字 菜品图=%d 补图=%d 图集合计=%d LLM=%v 更新字段=%v 备注=%s",
			res.WikiTitle, res.WikiTextLen, res.DishImageCount, res.CommonsImageCount, res.ImageCount,
			res.LLMUsed, res.UpdatedFields, res.Note)
		okCnt++
	}
	log.Printf("完成: 成功 %d / 失败 %d", okCnt, failCnt)
	if failCnt > 0 {
		os.Exit(1)
	}
}

// mustDB 连接数据库;migrate 为真时确保表结构最新(新增表需要迁移)。
// dry-run 时跳过迁移,避免 DDL 与「不写库」的 flag 承诺矛盾。
func mustDB(cfg *config.Config, migrate bool) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatalf("MySQL 连接失败: %v", err)
	}
	if !migrate {
		log.Println("dry-run 模式:跳过数据库迁移(不写库)")
		return db
	}
	database.MustAutoMigrate(db)
	return db
}
