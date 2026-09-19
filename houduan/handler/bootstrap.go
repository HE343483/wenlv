package handler

// Bootstrap 汇总全部 handler,便于 router 统一注册。
type Bootstrap struct {
	Auth     *AuthHandler
	Scenic   *ScenicHandler
	Food     *FoodHandler
	FoodCard *FoodCardHandler
	// FoodCategory 美食大类(川菜/名小吃/夜宵)详情
	FoodCategory *FoodCategoryHandler
	Route        *RouteHandler
	Favorite     *FavoriteHandler
	CheckIn      *CheckInHandler
	Review       *ReviewHandler
	Photo        *PhotoHandler
	Article      *ArticleHandler
	Upload       *UploadHandler
	HotTopic     *HotTopicHandler
	// ScenicExtra 景点详情页实时数据(周边推荐/交通站点)
	ScenicExtra *ScenicExtraHandler

	// AI 行程规划模块(移植自 TripStar)
	Trip     *TripHandler
	TripTool *TripToolHandler
}
