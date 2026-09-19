// Package router 注册路由与中间件。
package router

import (
	"context"

	"github.com/gin-gonic/gin"

	"wenlv-backend/handler"
	"wenlv-backend/middleware"
)

// Setup 构建并返回配置好的 Gin 引擎。
// validate 用于 access Token 校验;rl 为用户级限流器(可为 nil,便于测试)。
func Setup(h *handler.Bootstrap, validate func(ctx context.Context, token string) (uint, error), rl *middleware.RateLimiter) *gin.Engine {
	r := gin.New()
	// 顺序:请求 ID → 跨域 → 访问日志 → panic 恢复(最内层,能兜住所有路由 panic)
	r.Use(middleware.RequestID(), middleware.CORS(), middleware.AccessLogger(), middleware.Recovery())

	api := r.Group("/api")

	// 内容(公开)
	api.GET("/scenic", h.Scenic.List)
	api.GET("/scenic/:id", h.Scenic.Get)
	api.GET("/scenic/:id/around", h.ScenicExtra.Around)
	api.GET("/scenic/:id/transport", h.ScenicExtra.Transport)
	api.GET("/food", h.Food.List)
	api.GET("/food/:id", h.Food.Get)
	api.GET("/food-cards", h.FoodCard.List)
	api.GET("/food-category/:key", h.FoodCategory.Get)
	api.GET("/routes", h.Route.List)
	api.GET("/routes/:id", h.Route.Get)

	// 文旅热点(公开,定时抓取官方文旅新闻源)
	api.GET("/news/hotspots", h.HotTopic.List)

	// 评价/图墙/游记(公开读取)
	api.GET("/reviews", h.Review.List)
	api.GET("/reviews/summary/:targetType/:targetId", h.Review.Summary)
	api.GET("/photos", h.Photo.ListByTarget)
	api.GET("/photos/wall", h.Photo.Wall)
	api.GET("/articles", h.Article.List)
	api.GET("/articles/:id", h.Article.Get)

	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	// 鉴权区
	authed := api.Group("", middleware.Auth(validate))
	{
		authAuthed := authed.Group("/auth")
		{
			authAuthed.POST("/logout", h.Auth.Logout)
			authAuthed.GET("/me", h.Auth.Me)
		}

		// 收藏
		authed.GET("/favorites", h.Favorite.List)
		authed.POST("/favorites", h.Favorite.Add)
		authed.DELETE("/favorites/:targetType/:targetId", h.Favorite.Remove)

		// 打卡
		authed.POST("/check-ins", h.CheckIn.CheckIn)
		authed.GET("/check-ins", h.CheckIn.List)
		authed.DELETE("/check-ins/:id", h.CheckIn.Remove)

		// 评价
		authed.POST("/reviews", h.Review.Create)
		authed.GET("/reviews/mine", h.Review.Mine)
		authed.DELETE("/reviews/:id", h.Review.Delete)

		// 图墙上传
		authed.POST("/photos", h.Photo.Add)

		// 游记
		authed.POST("/articles", h.Article.Create)
		authed.PUT("/articles/:id", h.Article.Update)
		authed.DELETE("/articles/:id", h.Article.Delete)

		// 图片直传
		authed.POST("/upload/policy", h.Upload.Policy)
	}

	// ===== AI 行程规划模块(公开访问,与原 TripStar 一致使用匿名 user_id) =====
	// 各类接口按用户限流:防止共享 LLM/高德额度被打爆、小红书搜图触发风控
	if h.Trip != nil {
		trip := api.Group("/trip")
		{
			if rl != nil {
				trip.POST("/plan", rl.PlanLimit(), h.Trip.Plan)
				trip.POST("/chat/ask", rl.ChatLimit(), h.Trip.Ask)
			} else {
				trip.POST("/plan", h.Trip.Plan)
				trip.POST("/chat/ask", h.Trip.Ask)
			}
			trip.POST("/story-card", h.Trip.StoryCard) // 旅行故事卡片文案(出海分享)
			trip.GET("/status/:taskId", h.Trip.Status)
			trip.GET("/history", h.Trip.History)
			trip.GET("/plans/:planId", h.Trip.PlanDetail)
			trip.DELETE("/history/:planId", h.Trip.DeleteHistory)
			trip.GET("/health", h.Trip.Health)
			trip.GET("/ws/:taskId", h.Trip.WS)
		}
	}
	if h.TripTool != nil {
		poi := api.Group("/poi")
		{
			if rl != nil {
				poi.GET("/detail/:poiId", rl.MapLimit(), h.TripTool.POIDetail)
				poi.GET("/search", rl.MapLimit(), h.TripTool.POISearch)
				poi.GET("/image", rl.ImageLimit(), h.TripTool.POIImage)
				poi.GET("/photo", rl.ImageLimit(), h.TripTool.POIPhoto)
			} else {
				poi.GET("/detail/:poiId", h.TripTool.POIDetail)
				poi.GET("/search", h.TripTool.POISearch)
				poi.GET("/image", h.TripTool.POIImage)
				poi.GET("/photo", h.TripTool.POIPhoto)
			}
		}
		mapGroup := api.Group("/map")
		{
			if rl != nil {
				mapGroup.GET("/poi", rl.MapLimit(), h.TripTool.MapPOI)
				mapGroup.GET("/weather", rl.MapLimit(), h.TripTool.MapWeather)
				mapGroup.GET("/districts", rl.MapLimit(), h.TripTool.MapDistricts)
				mapGroup.POST("/route", rl.MapLimit(), h.TripTool.MapRoute)
			} else {
				mapGroup.GET("/poi", h.TripTool.MapPOI)
				mapGroup.GET("/weather", h.TripTool.MapWeather)
				mapGroup.GET("/districts", h.TripTool.MapDistricts)
				mapGroup.POST("/route", h.TripTool.MapRoute)
			}
			mapGroup.GET("/health", h.TripTool.MapHealth)
		}
		chat := api.Group("/chat")
		{
			if rl != nil {
				chat.POST("/ask", rl.ChatLimit(), h.Trip.Ask)
				chat.POST("/ask/stream", rl.ChatLimit(), h.Trip.AskStream) // SSE 流式问答
			} else {
				chat.POST("/ask", h.Trip.Ask)
				chat.POST("/ask/stream", h.Trip.AskStream) // SSE 流式问答
			}
		}
		settings := api.Group("/settings")
		{
			settings.GET("", h.TripTool.GetSettings)
			settings.PUT("", h.TripTool.SaveSettings)
		}
		memory := api.Group("/memory")
		{
			memory.GET("/list", h.TripTool.MemoryList)
			memory.POST("/add-explicit", h.TripTool.MemoryAddExplicit)
			memory.DELETE("/item", h.TripTool.MemoryDeleteItem)
			memory.DELETE("/clear", h.TripTool.MemoryClear)
		}
	}

	return r
}
