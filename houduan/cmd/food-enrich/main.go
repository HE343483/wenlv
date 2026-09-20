// food-enrich 美食详情数据补全批处理。
// 数据源:高德门店(名称含菜名的餐饮 POI:地址/坐标/评分/人均/相册) + LLM(风味标签/辣度/招牌/场景/风味故事)。
// 产出:字段写入 MySQL foods,相册图上传 OSS。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/food-enrich -limit 3 -dry-run
//	go run ./cmd/food-enrich -only 麻婆豆腐 -force
//	go run ./cmd/food-enrich
package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

func main() {
	limit := flag.Int("limit", 0, "仅处理前 N 个美食(0 表示全部)")
	only := flag.String("only", "", "仅处理名称包含该关键字的美食")
	dryRun := flag.Bool("dry-run", false, "只打印门店匹配结果,不写库、不上传、不调 LLM")
	withImages := flag.Bool("with-images", true, "抓取相册图并上传 OSS")
	withLLM := flag.Bool("with-llm", true, "调用 LLM 生成风味故事与参考值")
	force := flag.Bool("force", false, "已有值时也覆盖(默认只补空字段)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("food-enrich")
	defer logger.Close()

	db := mustDB(cfg, !*dryRun)
	repo := repository.NewFoodRepo(db)

	// 与 main.go 一致:高德/LLM Key 先取 .env,再被 data/runtime_settings.json 覆盖(设置页热更新过的值)
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
	amap := service.NewAmapService(tripSettings)
	llm := service.NewTripLLM(tripSettings)
	signer := pkg.NewOssSigner(&pkg.OssConfig{
		Endpoint:  os.Getenv("OSS_ENDPOINT"),
		AccessKey: os.Getenv("OSS_ACCESS_KEY"),
		SecretKey: os.Getenv("OSS_SECRET_KEY"),
		Bucket:    os.Getenv("OSS_BUCKET"),
	})
	if !amap.Available() {
		logger.Fatalf("高德 Web 服务 Key 未配置:请在 .env 的 AMAP_WEB_KEY 或前端设置页配置后重试")
	}
	if !signer.Configured() {
		logger.Infof("提示:OSS 未配置,将跳过相册图片采集")
	}
	if !llm.Available() {
		logger.Infof("提示:LLM 未配置,将跳过风味故事与参考值生成")
	}

	foods, err := repo.ListAll()
	if err != nil {
		logger.Fatalf("读取美食失败: %v", err)
	}
	if *only != "" {
		filtered := make([]model.Food, 0, 1)
		for _, f := range foods {
			if strings.Contains(f.NameZH, *only) {
				filtered = append(filtered, f)
			}
		}
		foods = filtered
	}
	if *limit > 0 && *limit < len(foods) {
		foods = foods[:*limit]
	}
	logger.Infof("待处理美食 %d 个(dry-run=%v, with-llm=%v, with-images=%v)", len(foods), *dryRun, *withLLM, *withImages)
	if *dryRun {
		logger.Infof("dry-run 模式:复用正式跑的 POI 判定逻辑打印匹配结果,不写库、不上传、不调 LLM")
	}

	enricher := service.NewFoodEnricher(repo, amap, llm, signer)
	ctx := context.Background()
	okCnt, failCnt, skipCnt := 0, 0, 0
	for i, f := range foods {
		time.Sleep(300 * time.Millisecond) // 控制高德 QPS:dry-run 与正式跑都节流,每轮只 sleep 一次
		logger.Infof("[%d/%d] %s", i+1, len(foods), f.NameZH)
		if *dryRun {
			printFoodDryRun(ctx, enricher, amap, f)
			skipCnt++
			continue
		}
		res, err := enricher.Enrich(ctx, f, service.FoodEnrichOptions{
			WithImages: *withImages,
			WithLLM:    *withLLM,
			Force:      *force,
		})
		if err != nil {
			logger.Errorf("    落库失败: %v", err)
			failCnt++
			continue
		}
		logger.Infof("    POI匹配=%v 上传图片=%d LLM=%v 更新字段=%v 备注=%s",
			res.POIMatched, res.Uploaded, res.LLMUsed, res.UpdatedFields, res.Note)
		okCnt++
	}
	logger.Infof("完成: 成功 %d / 失败 %d / 试跑跳过 %d", okCnt, failCnt, skipCnt)
}

// printFoodDryRun 打印门店匹配结果(门店名/地址/评分/人均/相册张数),供人工核对后再正式跑。
// 复用 enricher.MatchPOI(与 Enrich 同一判定:餐饮服务类型 + 名称打分),
// 保证「未匹配清单」与正式跑结论一致。
func printFoodDryRun(ctx context.Context, e *service.FoodEnricher, amap *service.AmapService, f model.Food) {
	poi, ok := e.MatchPOI(ctx, f)
	if !ok {
		logger.Infof("    未匹配到同名门店(门店事实留空,评分/人均改由 LLM 出参考值)")
		return
	}
	logger.Infof("    匹配门店: %s", poi.Name)
	d, err := amap.GetPOIDetail(ctx, poi.ID)
	if err != nil {
		logger.Warnf("    门店详情获取失败: %v", err)
		return
	}
	logger.Infof("    门店名=%s 地址=%s 评分=%s 人均=%s 相册=%d 张",
		d.Name, d.Address, d.Rating, d.Cost, len(d.Photos))
}

// mustDB 连接数据库;migrate 为真时确保表结构最新(新增列需要迁移)。
// dry-run 时跳过迁移,避免 DDL 与「不写库」的 flag 承诺矛盾。
func mustDB(cfg *config.Config, migrate bool) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	if !migrate {
		logger.Infof("dry-run 模式:跳过数据库迁移(不写库)")
		return db
	}
	database.MustAutoMigrate(db)
	return db
}
