// wenlv-backend 成都文旅后端服务入口。
package main

import (
	"fmt"
	"os"
	"time"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/handler"
	"wenlv-backend/logger"
	"wenlv-backend/middleware"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/router"
	"wenlv-backend/service"
)

func main() {
	cfg := config.Load()

	// 日志模块:按级别分文件落盘到 logs/(info/warn/error/access/debug/fatal),
	// 报错自动生成带日期与严重程度的错误 ID,便于精确定位
	logger.Configure(logger.Options{
		Dir:            cfg.Log.Dir,
		Level:          cfg.Log.Level,
		DisableConsole: !cfg.Log.Console,
		RetentionDays:  cfg.Log.RetentionDays,
	})
	defer logger.Close()

	// MySQL 连接与自动迁移
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	logger.Infof("MySQL 连接成功: %s:%s/%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.DBName)
	database.MustAutoMigrate(db)
	logger.Infof("数据库自动迁移完成")

	// Redis 连接
	rdb := database.MustRedis(cfg)
	logger.Infof("Redis 连接成功: %s (db=%d)", cfg.Redis.Addr, cfg.Redis.DB)

	// 组装仓储
	userRepo := repository.NewUserRepo(db)
	scenicRepo := repository.NewScenicRepo(db)
	foodRepo := repository.NewFoodRepo(db)
	foodCardRepo := repository.NewFoodCardRepo(db)
	foodCategoryRepo := repository.NewFoodCategoryRepo(db)
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
	foodCardSvc := service.NewFoodCardService(foodCardRepo)
	foodCategorySvc := service.NewFoodCategoryService(foodCategoryRepo)
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
	// 用户级限流:防止共享 LLM/高德额度被打爆、小红书搜图触发风控
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		PlanPer10Min:      cfg.RateLimit.PlanPer10Min,
		PlanPerDay:        cfg.RateLimit.PlanPerDay,
		PlanMaxConcurrent: cfg.RateLimit.PlanMaxConcurrent,
		ChatPerMin:        cfg.RateLimit.ChatPerMin,
		MapPerMin:         cfg.RateLimit.MapPerMin,
		ImagePerMin:       cfg.RateLimit.ImagePerMin,
	}, authSvc.ValidateAccess)
	// 小红书 Cookie 保活:定时访问首页续期并持久化
	service.NewXHSKeepalive(tripSettings).Start()

	// 文旅热点:定时抓取四川文旅厅"行业动态",LLM 翻译多语言,落库 hot_topics(列表走 Redis 缓存 24h)
	hotTopicSvc := service.NewHotTopicService(hotTopicRepo, tripLLM, rdb)
	hotTopicSvc.Start()
	logger.Infof("后台任务已启动: 小红书 Cookie 保活、文旅热点定时抓取")

	// 景点详情页实时数据:Redis 缓存时长由 SCENIC_CACHE_TTL_HOURS 配置(默认 24h)
	scenicExtraSvc := service.NewScenicExtraService(
		scenicRepo, tripAmap, rdb,
		time.Duration(cfg.Trip.ScenicCacheTTLHours)*time.Hour,
	)

	// 处理器
	h := &handler.Bootstrap{
		Auth:     handler.NewAuthHandler(authSvc),
		Scenic:   handler.NewScenicHandler(scenicSvc),
		Food:     handler.NewFoodHandler(foodSvc),
		FoodCard: handler.NewFoodCardHandler(foodCardSvc),
		// FoodCategory 美食大类(川菜/名小吃/夜宵)详情
		FoodCategory: handler.NewFoodCategoryHandler(foodCategorySvc),
		Route:        handler.NewRouteHandler(routeSvc),
		Favorite:     handler.NewFavoriteHandler(favoriteSvc),
		CheckIn:      handler.NewCheckInHandler(checkInSvc),
		Review:       handler.NewReviewHandler(reviewSvc),
		Photo:        handler.NewPhotoHandler(photoSvc),
		Article:      handler.NewArticleHandler(articleSvc),
		Upload:       handler.NewUploadHandler(uploadSvc),
		HotTopic:     handler.NewHotTopicHandler(hotTopicSvc),

		// ScenicExtra 景点详情页实时数据(周边推荐/交通站点)
		ScenicExtra: handler.NewScenicExtraHandler(scenicExtraSvc),
		Trip:        handler.NewTripHandler(tripPlanner, tripChat, tripTasks, tripSettings, rateLimiter, service.NewTripStoryCardService(tripLLM)),
		TripTool:    handler.NewTripToolHandler(tripSettings, tripAmap, tripXHS, tripDouyin, tripMemory),
	}

	// 行程图片 OSS 落库:完成后后台抓图上传,历史详情直连 OSS(未配置 OSS 时自动跳过)
	tripTasks.AttachImageEnricher(service.NewTripImageEnricher(
		cfg.OSS.Endpoint, cfg.OSS.AccessKey, cfg.OSS.SecretKey, cfg.OSS.Bucket,
		h.TripTool.PhotoBytesWithFallback,
	))

	engine := router.Setup(h, authSvc.ValidateAccess, rateLimiter)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Infof("wenlv-backend 服务已就绪,监听 %s (日志目录: %s, 日志级别: %s)",
		addr, logger.LogDir(), cfg.Log.Level)
	if err := engine.Run(addr); err != nil {
		logger.Fatalf("服务启动失败: %v", err)
	}
}

// getEnvOr 读取环境变量,未配置时返回默认值。
func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
