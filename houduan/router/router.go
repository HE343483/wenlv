// Package router 注册路由与中间件。
package router

import (
	"context"

	"github.com/gin-gonic/gin"

	"wenlv-backend/handler"
	"wenlv-backend/middleware"
)

// Setup 构建并返回配置好的 Gin 引擎。
// validate 用于 access Token 校验。
func Setup(h *handler.Bootstrap, validate func(ctx context.Context, token string) (uint, error)) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	api := r.Group("/api")

	// 内容(公开)
	api.GET("/scenic", h.Scenic.List)
	api.GET("/scenic/:id", h.Scenic.Get)
	api.GET("/food", h.Food.List)
	api.GET("/food/:id", h.Food.Get)
	api.GET("/routes", h.Route.List)
	api.GET("/routes/:id", h.Route.Get)

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
	if h.Trip != nil {
		trip := api.Group("/trip")
		{
			trip.POST("/plan", h.Trip.Plan)
			trip.GET("/status/:taskId", h.Trip.Status)
			trip.GET("/history", h.Trip.History)
			trip.GET("/plans/:planId", h.Trip.PlanDetail)
			trip.DELETE("/history/:planId", h.Trip.DeleteHistory)
			trip.GET("/health", h.Trip.Health)
			trip.GET("/ws/:taskId", h.Trip.WS)
			trip.POST("/chat/ask", h.Trip.Ask)
		}
	}
	if h.TripTool != nil {
		poi := api.Group("/poi")
		{
			poi.GET("/detail/:poiId", h.TripTool.POIDetail)
			poi.GET("/search", h.TripTool.POISearch)
			poi.GET("/image", h.TripTool.POIImage)
			poi.GET("/photo", h.TripTool.POIPhoto)
		}
		mapGroup := api.Group("/map")
		{
			mapGroup.GET("/poi", h.TripTool.MapPOI)
			mapGroup.GET("/weather", h.TripTool.MapWeather)
			mapGroup.POST("/route", h.TripTool.MapRoute)
			mapGroup.GET("/health", h.TripTool.MapHealth)
		}
		chat := api.Group("/chat")
		{
			chat.POST("/ask", h.Trip.Ask)
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