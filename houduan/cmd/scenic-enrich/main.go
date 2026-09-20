// scenic-enrich 景点详情数据补全批处理。
// 数据源:高德 Web 服务(地址/电话/开放时间/相册图) + LLM(图文详情与参考值)。
// 产出:字段写入 MySQL scenic_spots,相册图上传 OSS。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/scenic-enrich -limit 3 -dry-run                      # 试跑前 3 个,不落库
//	go run ./cmd/scenic-enrich -only 春熙路 -with-llm -with-images    # 重跑单个景点
//	go run ./cmd/scenic-enrich                                        # 全量
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
	limit := flag.Int("limit", 0, "仅处理前 N 个景点(0 表示全部)")
	only := flag.String("only", "", "仅处理名称包含该关键字的景点")
	dryRun := flag.Bool("dry-run", false, "只打印将要写入的内容,不写库、不上传")
	withImages := flag.Bool("with-images", true, "抓取相册图并上传 OSS")
	withLLM := flag.Bool("with-llm", true, "调用 LLM 生成图文详情与参考值")
	force := flag.Bool("force", false, "已有值时也覆盖(默认只补空字段)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("scenic-enrich")
	defer logger.Close()

	db := mustDB(cfg, !*dryRun)
	repo := repository.NewScenicRepo(db)

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
	tripXHSSigner := service.NewXHSSigner(getEnvOr("TRIP_XHS_SIGN_DIR", "xhs_sign"))
	tripXHS := service.NewXHSService(tripSettings, llm, amap, tripXHSSigner)
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
		logger.Infof("提示:LLM 未配置,将跳过图文详情与参考值生成")
	}

	spots, err := repo.ListAll()
	if err != nil {
		logger.Fatalf("读取景点失败: %v", err)
	}
	if *only != "" {
		filtered := make([]model.ScenicSpot, 0, 1)
		for _, s := range spots {
			if strings.Contains(s.NameZH, *only) {
				filtered = append(filtered, s)
			}
		}
		spots = filtered
	}
	if *limit > 0 && *limit < len(spots) {
		spots = spots[:*limit]
	}
	logger.Infof("待处理景点 %d 个(dry-run=%v, with-llm=%v, with-images=%v)", len(spots), *dryRun, *withLLM, *withImages)

	// dry-run 时不落库、不上传:把 enricher 的写库动作短路
	if *dryRun {
		logger.Infof("dry-run 模式:复用正式跑的 POI 判定逻辑打印匹配结果,不写库、不上传")
	}

	enricher := service.NewScenicEnricher(repo, amap, llm, signer, tripXHS)
	ctx := context.Background()
	okCnt, failCnt, skipCnt := 0, 0, 0
	for i, s := range spots {
		time.Sleep(300 * time.Millisecond) // 控制高德 QPS:dry-run 与正式跑都节流,每轮只 sleep 一次
		logger.Infof("[%d/%d] %s", i+1, len(spots), s.NameZH)
		if *dryRun {
			// dry-run 复用 enricher 的真实判定,只打印匹配与字段取值,不落库
			printDryRun(ctx, enricher, amap, s)
			skipCnt++
			continue
		}
		res, err := enricher.Enrich(ctx, s, service.EnrichOptions{
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

// printDryRun 打印高德匹配与字段取值,供人工核对后再正式跑。
// 复用 enricher.MatchPOI(与 Enrich 同一判定:名称打分 + 5km 距离过滤 + 全国搜索回退),
// 保证「未匹配清单」与正式跑结论一致。
func printDryRun(ctx context.Context, e *service.ScenicEnricher, amap *service.AmapService, s model.ScenicSpot) {
	poiID, poiName, ok := e.MatchPOI(ctx, s)
	if !ok {
		logger.Infof("    未匹配到高德 POI(与正式跑同一判定逻辑)")
		return
	}
	logger.Infof("    匹配 POI: %s / %s", poiID, poiName)
	d, err := amap.GetPOIDetail(ctx, poiID)
	if err != nil {
		logger.Warnf("    详情获取失败: %v", err)
		return
	}
	logger.Infof("    地址=%s 电话=%s 开放时间=%q 相册=%d 张", d.Address, d.Tel, d.OpenHours, len(d.Photos))
}

// getEnvOr 读取环境变量,为空时返回默认值(与 houduan/main.go 的同名工具保持一致)。
func getEnvOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
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
