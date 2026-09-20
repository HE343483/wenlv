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
	// Poems 景点诗词(诗词地图)
	Poems *PoemsHandler

	// CultureDaily 每日蜀签(诗句/方言/冷知识三语日签)
	CultureDaily *CultureDailyHandler

	// SolarFood 节气蜀俗(按当前节气查应季美食)
	SolarFood *SolarFoodHandler

	// Quiz 蜀文化知识闯关(景点抽题与判分)
	Quiz *QuizHandler

	// AI 行程规划模块(移植自 TripStar)
	Trip     *TripHandler
	TripTool *TripToolHandler

	// PersonaChat 历史人物 AI 角色对话(杜甫/诸葛亮)
	PersonaChat *PersonaChatHandler

	// ChatSessions AI 对话会话与角色长期记忆管理(仅登录用户)
	ChatSessions *ChatSessionHandler
}
