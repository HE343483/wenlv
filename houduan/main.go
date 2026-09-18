// wenlv-backend 成都文旅后端服务入口。
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/handler"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/router"
	"wenlv-backend/service"
)

func main() {
	cfg := config.Load()

	// MySQL 连接与自动迁移
	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatalf("MySQL 连接失败: %v", err)
	}
	database.MustAutoMigrate(db)

	// Redis 连接
	rdb := database.MustRedis(cfg)

	// 组装仓储
	userRepo := repository.NewUserRepo(db)
	scenicRepo := repository.NewScenicRepo(db)
	foodRepo := repository.NewFoodRepo(db)
	routeRepo := repository.NewRouteRepo(db)
	favoriteRepo := repository.NewFavoriteRepo(db)
	checkInRepo := repository.NewCheckInRepo(db)
	reviewRepo := repository.NewReviewRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	articleRepo := repository.NewArticleRepo(db)
	userMemoryRepo := repository.NewUserMemoryRepo(db)
	hotTopicRepo := repository.NewHotTopicRepo(db)

	// 鉴权工具
	jwtMgr := pkg.NewJWTManager(pkg.JWTConfig{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessTTL:     cfg.JWT.AccessTTL,
		RefreshTTL:    cfg.JWT.RefreshTTL,
		Issuer:        cfg.JWT.Issuer,
	})

	// 业务服务
	authSvc := service.NewAuthService(
		userRepo, rdb, jwtMgr,
		time.Duration(cfg.JWT.AccessTTL)*time.Second,
		time.Duration(cfg.JWT.RefreshTTL)*time.Second,
	)
	scenicSvc := service.NewScenicService(scenicRepo)
	foodSvc := service.NewFoodService(foodRepo)
	routeSvc := service.NewRouteService(routeRepo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo)
	checkInSvc := service.NewCheckInService(checkInRepo)
	reviewSvc := service.NewReviewService(reviewRepo, userRepo)
	photoSvc := service.NewPhotoService(photoRepo)
	articleSvc := service.NewArticleService(articleRepo, userRepo)

	// OSS 签名器(未配置时仅上传接口不可用,其余正常)
	signer := pkg.NewOssSigner(&pkg.OssConfig{
		Endpoint:     cfg.OSS.Endpoint,
		AccessKey:    cfg.OSS.AccessKey,
		SecretKey:    cfg.OSS.SecretKey,
		Bucket:       cfg.OSS.Bucket,
		PublicDomain: "",
	})
	uploadSvc := service.NewUploadService(signer, 15)

	// ===== AI 行程规划模块(移植自 TripStar) =====
	tripSettings := service.NewTripSettings(service.TripSettingsOptions{
		Defaults: service.TripRuntimeSettings{
			ViteAmapWebKey:   cfg.Trip.AmapWebKey,
			ViteAmapWebJSKey: cfg.Trip.AmapWebJSKey,
			GoogleMapsAPIKey: cfg.Trip.GoogleMapsAPIKey,
			GoogleMapsProxy:  cfg.Trip.GoogleMapsProxy,
			XHSCookie:        cfg.Trip.XHSCookie,
			DouyinCookie:     cfg.Trip.DouyinCookie,
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
		Memory: service.TripMemoryParams{
			DecayFactor:      cfg.Trip.MemoryDecayFactor,
			WeightThreshold:  cfg.Trip.MemoryWeightThreshold,
			MaxRecall:        cfg.Trip.MemoryMaxRecall,
			MinInitWeight:    cfg.Trip.MemoryMinInitWeight,
			MaxSingleContent: cfg.Trip.MemoryMaxSingleContent,
		},
	})
	tripLLM := service.NewTripLLM(tripSettings)
	tripAmap := service.NewAmapService(tripSettings)
	tripSigner := service.NewXHSSigner(getEnvOr("TRIP_XHS_SIGN_DIR", "xhs_sign"))
	tripXHS := service.NewXHSService(tripSettings, tripLLM, tripAmap, tripSigner)
	// 抖音 a_bogus 签名与数据源(签名 JS 需放置到 douyin_sign/douyin.js)
	tripDouyinSigner := service.NewDouyinSigner(getEnvOr("TRIP_DOUYIN_SIGN_DIR", "douyin_sign"))
	tripDouyin := service.NewDouyinService(tripSettings, tripLLM, tripAmap, tripDouyinSigner)
	tripMemory := service.NewTripMemoryService(tripSettings, userMemoryRepo, tripLLM)
	tripTasks := service.NewTripTaskStore(cfg.Trip.DataDir)
	tripTasks.AttachDB(db)    // 行程完成后落库 trip_plans,历史/回看走数据库
	tripSettings.AttachDB(db) // 设置页配置落库 trip_settings,重启不丢失
	tripPlanner := service.NewTripPlanner(tripSettings, tripLLM, tripAmap, tripXHS, tripDouyin, tripMemory, tripTasks)
	tripChat := service.NewTripChatService(tripLLM)
	// 小红书 Cookie 保活:定时访问首页续期并持久化
	service.NewXHSKeepalive(tripSettings).Start()

	// 文旅热点:定时抓取四川文旅厅"行业动态",LLM 翻译多语言,落库 hot_topics(列表走 Redis 缓存 24h)
	hotTopicSvc := service.NewHotTopicService(hotTopicRepo, tripLLM, rdb)
	hotTopicSvc.Start()

	// 处理器
	h := &handler.Bootstrap{
		Auth:     handler.NewAuthHandler(authSvc),
		Scenic:   handler.NewScenicHandler(scenicSvc),
		Food:     handler.NewFoodHandler(foodSvc),
		Route:    handler.NewRouteHandler(routeSvc),
		Favorite: handler.NewFavoriteHandler(favoriteSvc),
		CheckIn:  handler.NewCheckInHandler(checkInSvc),
		Review:   handler.NewReviewHandler(reviewSvc),
		Photo:    handler.NewPhotoHandler(photoSvc),
		Article:  handler.NewArticleHandler(articleSvc),
		Upload:   handler.NewUploadHandler(uploadSvc),
		HotTopic: handler.NewHotTopicHandler(hotTopicSvc),
		Trip:     handler.NewTripHandler(tripPlanner, tripChat, tripTasks, tripSettings),
		TripTool: handler.NewTripToolHandler(tripSettings, tripAmap, tripXHS, tripDouyin, tripMemory),
	}

	engine := router.Setup(h, authSvc.ValidateAccess)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("wenlv-backend 启动于 %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// getEnvOr 读取环境变量,未配置时返回默认值。
func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
