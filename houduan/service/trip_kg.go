package service

import (
	"fmt"
	"strings"

	"wenlv-backend/logger"
	"wenlv-backend/model"
)

// ============ 知识图谱构建服务 ============
// 将 TripPlan 转换为 ECharts 力导向图所需的 nodes / edges / categories。

var kgNodeColors = map[string]string{
	"city":       "#4A90D9",
	"day":        "#5B8FF9",
	"attraction": "#5AD8A6",
	"hotel":      "#F6BD16",
	"meal":       "#E8684A",
	"weather":    "#6DC8EC",
	"budget":     "#FF9845",
	"preference": "#B37FEB",
}

var kgNodeSizes = map[string]int{
	"city":       70,
	"day":        45,
	"attraction": 35,
	"hotel":      35,
	"meal":       25,
	"weather":    28,
	"budget":     40,
	"preference": 30,
}

var kgI18N = map[string]map[string]string{
	"zh": {
		"cat_city": "城市", "cat_day": "日程", "cat_attraction": "景点",
		"cat_hotel": "酒店", "cat_meal": "餐饮", "cat_weather": "天气",
		"cat_budget": "预算", "cat_preference": "偏好/建议",
		"edge_itinerary": "行程", "edge_visit": "游览", "edge_next": "下一站",
		"edge_checkin": "入住", "edge_weather": "天气", "edge_budget": "预算",
		"edge_suggestion": "建议",
		"day_n":           "第{n}天",
		"visit_duration":  "游览{min}分钟", "ticket_price": "门票¥{price}",
		"hotel_cost":   "{range} | ¥{cost}/晚",
		"total_budget": "总预算 ¥{total}",
		"breakfast":    "早餐", "lunch": "午餐", "dinner": "晚餐", "snack": "小吃",
		"budget_attraction": "景点", "budget_hotel": "酒店",
		"budget_meal": "餐饮", "budget_transport": "交通", "budget_inter_city": "城际交通",
	},
	"en": {
		"cat_city": "City", "cat_day": "Schedule", "cat_attraction": "Attraction",
		"cat_hotel": "Hotel", "cat_meal": "Dining", "cat_weather": "Weather",
		"cat_budget": "Budget", "cat_preference": "Tips",
		"edge_itinerary": "Itinerary", "edge_visit": "Visit", "edge_next": "Next",
		"edge_checkin": "Check-in", "edge_weather": "Weather", "edge_budget": "Budget",
		"edge_suggestion": "Tips",
		"day_n":           "Day {n}",
		"visit_duration":  "Visit {min} min", "ticket_price": "Ticket ¥{price}",
		"hotel_cost":   "{range} | ¥{cost}/night",
		"total_budget": "Total Budget ¥{total}",
		"breakfast":    "Breakfast", "lunch": "Lunch", "dinner": "Dinner", "snack": "Snack",
		"budget_attraction": "Attractions", "budget_hotel": "Hotels",
		"budget_meal": "Dining", "budget_transport": "Transport", "budget_inter_city": "Inter-city",
	},
	"ja": {
		"cat_city": "都市", "cat_day": "スケジュール", "cat_attraction": "観光地",
		"cat_hotel": "ホテル", "cat_meal": "グルメ", "cat_weather": "天気",
		"cat_budget": "予算", "cat_preference": "おすすめ",
		"edge_itinerary": "旅程", "edge_visit": "観光", "edge_next": "次へ",
		"edge_checkin": "宿泊", "edge_weather": "天気", "edge_budget": "予算",
		"edge_suggestion": "提案",
		"day_n":           "{n}日目",
		"visit_duration":  "観光{min}分", "ticket_price": "入場料¥{price}",
		"hotel_cost":   "{range} | ¥{cost}/泊",
		"total_budget": "総予算 ¥{total}",
		"breakfast":    "朝食", "lunch": "昼食", "dinner": "夕食", "snack": "軽食",
		"budget_attraction": "観光地", "budget_hotel": "ホテル",
		"budget_meal": "グルメ", "budget_transport": "交通", "budget_inter_city": "都市間交通",
	},
}

func kgT(key, lang string, kv ...string) string {
	table, ok := kgI18N[lang]
	if !ok {
		table = kgI18N["zh"]
	}
	tpl, ok := table[key]
	if !ok {
		tpl = kgI18N["zh"][key]
		if tpl == "" {
			tpl = key
		}
	}
	for i := 0; i+1 < len(kv); i += 2 {
		tpl = strings.ReplaceAll(tpl, "{"+kv[i]+"}", kv[i+1])
	}
	return tpl
}

// BuildKnowledgeGraph 由行程计划构建知识图谱。
func BuildKnowledgeGraph(plan *model.TripPlan, language string) *model.KnowledgeGraphData {
	lang := model.NormalizeLang(language)
	logger.Infof("[KG] build_knowledge_graph language='%s' -> resolved='%s'", language, lang)

	nodes := make([]model.GraphNode, 0, 32)
	edges := make([]model.GraphEdge, 0, 32)
	nodeIDs := map[string]bool{}

	catKeys := []string{"cat_city", "cat_day", "cat_attraction", "cat_hotel",
		"cat_meal", "cat_weather", "cat_budget", "cat_preference"}
	categories := make([]model.GraphCategory, 0, len(catKeys))
	catIndex := map[string]int{}
	for i, k := range catKeys {
		name := kgT(k, lang)
		categories = append(categories, model.GraphCategory{Name: name})
		catIndex[name] = i
	}
	styleKey := map[string]string{
		kgT("cat_city", lang): "city", kgT("cat_day", lang): "day",
		kgT("cat_attraction", lang): "attraction", kgT("cat_hotel", lang): "hotel",
		kgT("cat_meal", lang): "meal", kgT("cat_weather", lang): "weather",
		kgT("cat_budget", lang): "budget", kgT("cat_preference", lang): "preference",
	}

	addNode := func(id, name, categoryName, extra string) {
		if nodeIDs[id] {
			return
		}
		nodeIDs[id] = true
		key := styleKey[categoryName]
		nodes = append(nodes, model.GraphNode{
			ID:         id,
			Name:       name,
			Category:   catIndex[categoryName],
			SymbolSize: kgNodeSizes[key],
			ItemStyle:  map[string]any{"color": kgNodeColors[key]},
			Value:      extra,
		})
	}
	addEdge := func(source, target, label string) {
		edges = append(edges, model.GraphEdge{Source: source, Target: target, Label: label})
	}

	// ---- 1. 城市节点(支持多城市) ----
	cities := plan.Cities
	if len(cities) == 0 {
		cities = []string{plan.City}
	}
	cityNodeIDs := map[string]string{}
	rootID := ""
	if len(cities) > 1 {
		rootID = "trip_root"
		addNode(rootID, strings.Join(cities, " → "), kgT("cat_city", lang),
			plan.StartDate+" ~ "+plan.EndDate)
		for _, cityName := range cities {
			cid := "city_" + cityName
			addNode(cid, cityName, kgT("cat_city", lang), "")
			addEdge(rootID, cid, kgT("edge_itinerary", lang))
			cityNodeIDs[cityName] = cid
		}
	} else {
		rootID = "city_" + plan.City
		addNode(rootID, plan.City, kgT("cat_city", lang), plan.StartDate+" ~ "+plan.EndDate)
		cityNodeIDs[plan.City] = rootID
	}

	// ---- 2. 每日节点 ----
	for _, day := range plan.Days {
		dayID := fmt.Sprintf("day_%d", day.DayIndex)
		dayCity := day.City
		if dayCity == "" {
			dayCity = plan.City
		}
		parentID, ok := cityNodeIDs[dayCity]
		if !ok {
			parentID = rootID
		}
		addNode(dayID, kgT("day_n", lang, "n", fmt.Sprintf("%d", day.DayIndex+1)), kgT("cat_day", lang), day.Date)
		addEdge(parentID, dayID, kgT("edge_itinerary", lang))

		// 景点
		for i, attr := range day.Attractions {
			attrID := fmt.Sprintf("attr_%d_%d_%s", day.DayIndex, i, attr.Name)
			parts := make([]string, 0, 3)
			if attr.Address != "" {
				parts = append(parts, attr.Address)
			}
			if attr.VisitDuration > 0 {
				parts = append(parts, kgT("visit_duration", lang, "min", fmt.Sprintf("%d", attr.VisitDuration.Int())))
			}
			if attr.TicketPrice > 0 {
				parts = append(parts, kgT("ticket_price", lang, "price", fmt.Sprintf("%d", attr.TicketPrice.Int())))
			}
			addNode(attrID, attr.Name, kgT("cat_attraction", lang), strings.Join(parts, " | "))
			addEdge(dayID, attrID, kgT("edge_visit", lang))
			if i > 0 {
				prev := day.Attractions[i-1]
				prevID := fmt.Sprintf("attr_%d_%d_%s", day.DayIndex, i-1, prev.Name)
				addEdge(prevID, attrID, kgT("edge_next", lang))
			}
		}

		// 酒店
		if day.Hotel != nil {
			hotelID := fmt.Sprintf("hotel_%d_%s", day.DayIndex, day.Hotel.Name)
			value := day.Hotel.PriceRange
			if day.Hotel.EstimatedCost > 0 {
				value = kgT("hotel_cost", lang, "range", day.Hotel.PriceRange,
					"cost", fmt.Sprintf("%d", day.Hotel.EstimatedCost.Int()))
			}
			addNode(hotelID, day.Hotel.Name, kgT("cat_hotel", lang), value)
			addEdge(dayID, hotelID, kgT("edge_checkin", lang))
		}

		// 餐饮
		for j, meal := range day.Meals {
			mealTypeLabel := meal.Type
			switch meal.Type {
			case "breakfast", "lunch", "dinner", "snack":
				mealTypeLabel = kgT(meal.Type, lang)
			}
			mealID := fmt.Sprintf("meal_%d_%d_%s", day.DayIndex, j, meal.Name)
			value := ""
			if meal.EstimatedCost > 0 {
				value = fmt.Sprintf("¥%d", meal.EstimatedCost.Int())
			}
			addNode(mealID, mealTypeLabel+": "+meal.Name, kgT("cat_meal", lang), value)
			addEdge(dayID, mealID, mealTypeLabel)
		}
	}

	// ---- 3. 天气节点 ----
	for _, w := range plan.WeatherInfo {
		wID := "weather_" + w.Date
		addNode(wID, fmt.Sprintf("%s %d°C", w.DayWeather, w.DayTemp.Int()), kgT("cat_weather", lang), w.Date)
		for _, day := range plan.Days {
			if day.Date == w.Date {
				addEdge(fmt.Sprintf("day_%d", day.DayIndex), wID, kgT("edge_weather", lang))
				break
			}
		}
	}

	// ---- 4. 预算节点 ----
	if plan.Budget != nil {
		b := plan.Budget
		budgetID := "budget_total"
		addNode(budgetID, kgT("total_budget", lang, "total", fmt.Sprintf("%d", b.Total.Int())), kgT("cat_budget", lang), "")
		addEdge(rootID, budgetID, kgT("edge_budget", lang))

		items := []struct {
			key   string
			value model.FlexInt
		}{
			{"budget_attraction", b.TotalAttractions},
			{"budget_hotel", b.TotalHotels},
			{"budget_meal", b.TotalMeals},
			{"budget_transport", b.TotalTransportation},
			{"budget_inter_city", b.TotalInterCityTransit},
		}
		for _, it := range items {
			if it.value <= 0 {
				continue
			}
			label := kgT(it.key, lang)
			subID := "budget_" + it.key
			addNode(subID, fmt.Sprintf("%s ¥%d", label, it.value.Int()), kgT("cat_budget", lang), "")
			addEdge(budgetID, subID, label)
		}
	}

	// ---- 5. 总体建议节点 ----
	if plan.OverallSuggestions != "" {
		text := plan.OverallSuggestions
		runes := []rune(text)
		short := text
		if len(runes) > 30 {
			short = string(runes[:30]) + "..."
		}
		addNode("suggestion_overall", short, kgT("cat_preference", lang), text)
		addEdge(rootID, "suggestion_overall", kgT("edge_suggestion", lang))
	}

	return &model.KnowledgeGraphData{Nodes: nodes, Edges: edges, Categories: categories}
}