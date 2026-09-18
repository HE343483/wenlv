package handler

// Bootstrap 汇总全部 handler,便于 router 统一注册。
type Bootstrap struct {
	Auth     *AuthHandler
	Scenic   *ScenicHandler
	Food     *FoodHandler
	Route    *RouteHandler
	Favorite *FavoriteHandler
	CheckIn  *CheckInHandler
	Review   *ReviewHandler
	Photo    *PhotoHandler
	Article  *ArticleHandler
	Upload   *UploadHandler

	// AI 行程规划模块(移植自 TripStar)
	Trip     *TripHandler
	TripTool *TripToolHandler
}