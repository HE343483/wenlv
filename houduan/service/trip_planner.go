package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wenlv-backend/model"
)

// ============ 行程规划编排(多智能体协作流程的 Go 实现) ============
// 流程与原项目一致:
//  1. 按城市并发/串行搜集:小红书景点(LLM 提纯) / 天气 / 酒店
//  2. 汇总上下文交给规划大模型生成完整行程 JSON
//  3. 多层容错解析 + 知识图谱构建 + 结果持久化与推送

const tripPlannerPrompt = `你是行程规划专家。你的任务是根据景点信息和天气信息,生成详细的旅行计划。支持单城市和多城市行程。

请严格按照以下JSON格式返回旅行计划:
` + "```json" + `
{
  "city": "首个城市名称(兼容字段)",
  "cities": ["城市1", "城市2"],
  "start_date": "YYYY-MM-DD",
  "end_date": "YYYY-MM-DD",
  "days": [
    {
      "date": "YYYY-MM-DD",
      "day_index": 0,
      "city": "当天所在城市",
      "is_transfer_day": false,
      "transfer_info": "",
      "description": "第1天行程概述",
      "transportation": "交通方式",
      "accommodation": "住宿类型",
      "hotel": {
        "name": "酒店名称",
        "address": "酒店地址",
        "location": {"longitude": 116.397128, "latitude": 39.916527},
        "price_range": "300-500元",
        "rating": "4.5",
        "distance": "距离景点2公里",
        "type": "经济型酒店",
        "estimated_cost": 400
      },
      "attractions": [
        {
          "name": "景点名称",
          "address": "详细地址",
          "location": {"longitude": 116.397128, "latitude": 39.916527},
          "visit_duration": 120,
          "description": "景点详细描述",
          "category": "景点类别",
          "ticket_price": 60,
          "reservation_required": false,
          "reservation_tips": ""
        }
      ],
      "meals": [
        {"type": "breakfast", "name": "早餐推荐", "description": "早餐描述", "estimated_cost": 30},
        {"type": "lunch", "name": "午餐推荐", "description": "午餐描述", "estimated_cost": 50},
        {"type": "dinner", "name": "晚餐推荐", "description": "晚餐描述", "estimated_cost": 80}
      ]
    }
  ],
  "weather_info": [
    {
      "date": "YYYY-MM-DD",
      "city": "当天所在城市",
      "day_weather": "晴",
      "night_weather": "多云",
      "day_temp": 25,
      "night_temp": 15,
      "wind_direction": "南风",
      "wind_power": "1-3级"
    }
  ],
  "overall_suggestions": "总体建议",
  "budget": {
    "total_attractions": 180,
    "total_hotels": 1200,
    "total_meals": 480,
    "total_transportation": 200,
    "total_inter_city_transport": 0,
    "total": 2060
  }
}
` + "```" + `

**⚠️ JSON 格式关键约束(违反将导致系统崩溃):**
- budget 中所有费用字段(total_attractions、total_hotels、total_meals、total_transportation、total_inter_city_transport、total)必须是**纯数字**,绝对禁止出现算术表达式!
  - ✅ 正确: "total_attractions": 324
  - ❌ 错误: "total_attractions": 30+54+120+120=324
  -  错误: "total_attractions": "324元"
- ticket_price、estimated_cost 等所有价格字段也必须是纯数字,不带单位

**重要提示:**
1. weather_info数组必须包含每一天的天气信息,每条记录必须包含 city 字段标明该天所在城市
2. 温度必须是纯数字(不要带°C等单位)
3. 每天安排2-3个景点(城际移动日可减少为1-2个)
4. 考虑景点之间的距离和游览时间
5. 每天必须包含早中晚三餐
6. 提供实用的旅行建议
7. **必须包含预算信息**:
   - 景点门票价格(ticket_price)
   - 餐饮预估费用(estimated_cost)
   - 酒店预估费用(estimated_cost)
   - 预算汇总(budget)包含各项总费用
8. **预约信息透传**: 如果景点搜索数据中包含 reservation_required 和 reservation_tips 字段,请务必将它们完整保留在对应景点的JSON中。需要预约的景点请在 description 中也提醒游客提前预约
9. **景点图片**: 不需要在JSON中填写 image_url 字段,图片由前端根据景点名称自动从所选景点来源获取。
10. **多城市行程要求**:
    - 每个 day 对象中必须包含 "city" 字段标明当天所在城市
    - 城市切换当天设置 "is_transfer_day": true,并在 "transfer_info" 中**仅给出交通方式建议和大致时长**(如"建议乘坐高铁,约2-3小时"),**禁止编造具体车次、班次号、出发时间、到达时间等不可验证的信息**
    - 城际移动日的景点数量可适当减少为1-2个
    - budget 中的 "total_inter_city_transport" 统计城际交通费用(单城市时为0)
    - "cities" 数组列出所有途经城市(单城市时只有一个元素)
`

// TripPlanner 行程规划服务。
type TripPlanner struct {
	settings *TripSettings
	llm      *TripLLM
	amap     *AmapService
	xhs      *XHSService
	douyin   *DouyinService
	memory   *TripMemoryService
	tasks    *TripTaskStore
}

// NewTripPlanner 构造行程规划服务。
func NewTripPlanner(settings *TripSettings, llm *TripLLM, amap *AmapService, xhs *XHSService,
	douyin *DouyinService, memory *TripMemoryService, tasks *TripTaskStore) *TripPlanner {
	return &TripPlanner{settings: settings, llm: llm, amap: amap, xhs: xhs, douyin: douyin, memory: memory, tasks: tasks}
}

// RunPlanning 后台执行旅行规划并推送进度(由 handler 以 goroutine 方式调用)。
func (p *TripPlanner) RunPlanning(ctx context.Context, taskID string, req *model.TripRequest) {
	defer func() {
		if r := recover(); r != nil {
			p.failTask(taskID, fmt.Sprintf("旅行计划生成失败: %v", r))
		}
	}()

	req.Normalize()
	cities := req.Cities
	totalCities := len(cities)
	cityNames := make([]string, 0, totalCities)
	for _, cs := range cities {
		cityNames = append(cityNames, cs.City)
	}

	fmt.Printf("\n%s\n🚀 开始多智能体协作规划旅行...\n途经城市: %s\n日期: %s 至 %s\n天数: %d天\n偏好: %s\n%s\n",
		strings.Repeat("=", 60), strings.Join(cityNames, " → "), req.StartDate, req.EndDate, req.TravelDays,
		firstNonEmpty(strings.Join(req.Preferences, ", "), "无"), strings.Repeat("=", 60))

	p.updateTask(taskID, func(t *TripTask) {
		t.Status = TripTaskProcessing
		t.Stage = "attraction_search"
		t.Progress = 10
		t.Message = "正在准备数据..."
	})

	keywords := "景点"
	if len(req.Preferences) > 0 {
		keywords = req.Preferences[0]
	}
	lang := req.Lang()
	langHint := ""
	if lang != "zh" {
		name := langNames[lang]
		if name == "" {
			name = lang
		}
		langHint = " Please respond in " + name + "."
	}

	provider := CurrentMapProvider(p.settings)
	attractions := map[string]string{}
	weather := map[string]string{}
	hotels := map[string]string{}

	for idx, cityStay := range cities {
		city := cityStay.City
		progressBase := 10 + (idx*65)/max(totalCities, 1)
		progressStep := (65 / max(totalCities, 1)) / 3
		if progressStep < 3 {
			progressStep = 3
		}
		cityLabel := ""
		if totalCities > 1 {
			cityLabel = fmt.Sprintf(" (%d/%d)", idx+1, totalCities)
		}

		// [1] 景点搜索(用户选择小红书/抖音真人分享或高德地图检索)
		p.updateTask(taskID, func(t *TripTask) {
			t.Stage = "attraction_search"
			t.Progress = progressBase
			t.Message = fmt.Sprintf("正在搜索 %s 的景点...%s", city, cityLabel)
		})
		fmt.Printf("  [%d/%d] 正在搜索 %s 的景点...\n", idx+1, totalCities, city)
		var attractionText string
		switch req.AttractionSource {
		case "map":
			// 用户明确选择地图检索,直接走高德/Google POI
			attractionText = p.attractionsFromMap(ctx, city, keywords)
		case "douyin":
			// 抖音真人分享(LLM 提纯)
			text, err := p.douyin.SearchAttractionsText(ctx, city, keywords, lang)
			if err != nil {
				if isDouyinCookieExpired(err) {
					// 风控/Cookie 失效属于致命问题,直接失败并提示更换 Cookie
					p.failTask(taskID, "【认证失败】"+err.Error())
					return
				}
				// 抖音不可用(未配置 Cookie / 抓取失败)时降级为地图 POI 兜底,保证行程仍能生成
				fmt.Printf("  ⚠️ %s 抖音景点获取失败,降级使用地图 POI 兜底: %v\n", city, err)
				text = p.attractionsFromMap(ctx, city, keywords)
			}
			attractionText = text
		default:
			// 小红书真人推荐(LLM 提纯)
			text, err := p.xhs.SearchAttractionsText(ctx, city, keywords, lang)
			if err != nil {
				var cookieErr *XHSCookieExpiredError
				if asCookieExpired(err, &cookieErr) {
					// 风控/Cookie 失效属于致命问题,直接失败并提示更换 Cookie
					p.failTask(taskID, "【认证失败】"+err.Error())
					return
				}
				// 小红书不可用(未配置 Cookie / 抓取失败)时降级为地图 POI 兜底,保证行程仍能生成
				fmt.Printf("  ⚠️ %s 小红书景点获取失败,降级使用地图 POI 兜底: %v\n", city, err)
				text = p.attractionsFromMap(ctx, city, keywords)
			}
			attractionText = text
		}
		attractions[city] = attractionText

		// [2] 天气查询
		p.updateTask(taskID, func(t *TripTask) {
			t.Stage = "weather_search"
			t.Progress = progressBase + progressStep
			t.Message = fmt.Sprintf("正在查询 %s 的天气...%s", city, cityLabel)
		})
		fmt.Printf("  [%d/%d] 正在查询 %s 的天气...\n", idx+1, totalCities, city)
		weather[city] = p.queryWeatherText(ctx, city, provider, langHint)

		// [3] 酒店搜索
		p.updateTask(taskID, func(t *TripTask) {
			t.Stage = "hotel_search"
			t.Progress = progressBase + progressStep*2
			t.Message = fmt.Sprintf("正在搜索 %s 的酒店...%s", city, cityLabel)
		})
		fmt.Printf("  [%d/%d] 正在搜索 %s 的酒店...\n", idx+1, totalCities, city)
		hotels[city] = p.queryHotelText(ctx, city, req.Accommodation, provider, langHint)
	}

	fmt.Printf("\n✅ 全部 %d 个城市基础信息搜集完成\n\n", totalCities)

	// ========== 统一规划阶段 ==========
	planningLabel := "正在生成旅行计划..."
	if totalCities > 1 {
		planningLabel = "正在生成多城市行程计划..."
	}
	p.updateTask(taskID, func(t *TripTask) {
		t.Stage = "planning"
		t.Progress = 85
		t.Message = planningLabel
	})

	memorySnippet := ""
	if p.memory.Enabled() && req.MemoryEnabledFor() && strings.TrimSpace(req.UserID) != "" {
		memorySnippet = p.memory.BuildPromptSnippet(ctx, req.UserID)
	}

	plannerResponse, err := p.runPlannerWithRetry(ctx, req, attractions, weather, hotels, memorySnippet)
	if err != nil {
		p.failTask(taskID, fmt.Sprintf("旅行计划生成失败: %v", err))
		return
	}
	fmt.Printf("行程规划结果: %s...\n\n", truncateText(plannerResponse, 300))

	plan, err := ParseTripPlan(ctx, p.llm, plannerResponse, req)
	if err != nil {
		p.failTask(taskID, err.Error())
		return
	}

	// 补全 cities / 每日 city 字段(LLM 可能遗漏)
	if len(plan.Cities) == 0 {
		plan.Cities = cityNames
	}
	if plan.City == "" {
		plan.City = cityNames[0]
	}
	// 补全预算(LLM 偶尔遗漏 budget 字段,按每日明细汇总兜底)
	if plan.Budget == nil {
		plan.Budget = budgetFromDays(plan.Days)
	}
	// 补全整体建议(LLM 偶尔遗漏 overall_suggestions 字段,按行程要素本地兜底)
	if strings.TrimSpace(plan.OverallSuggestions) == "" {
		plan.OverallSuggestions = FallbackSuggestions(plan, cityNames, lang)
	}
	// 补全天气(LLM 偶尔遗漏 weather_info 字段,按日期匹配高德/Google 实时预报兜底;预报仅覆盖未来数日,超出范围不虚构)
	if len(plan.WeatherInfo) == 0 {
		plan.WeatherInfo = p.weatherFromForecast(ctx, cityNames, plan, provider)
	}
	// 记录景点数据来源,供前端展示图片来源标注
	plan.AttractionSource = req.AttractionSource
	if totalCities == 1 {
		for i := range plan.Days {
			if plan.Days[i].City == "" {
				plan.Days[i].City = cityNames[0]
			}
		}
	}

	// 异步提取用户偏好到记忆库(不阻塞主流程;用户关闭记忆开关时跳过)
	if p.memory.Enabled() && req.MemoryEnabledFor() && strings.TrimSpace(req.UserID) != "" {
		userID := req.UserID
		go p.memory.SavePreferencesAfterTrip(context.Background(), userID, req, plan)
	}

	p.updateTask(taskID, func(t *TripTask) {
		t.Stage = "graph_building"
		t.Progress = 95
		t.Message = "正在构建知识图谱..."
	})
	graph := BuildKnowledgeGraph(plan, lang)

	result := &model.TripPlanResponse{
		Success:   true,
		Message:   "旅行计划生成成功",
		PlanID:    taskID,
		Data:      plan,
		GraphData: graph,
	}
	fmt.Printf("✅ 任务 %s 完成\n", taskID)
	p.updateTask(taskID, func(t *TripTask) {
		t.Status = TripTaskCompleted
		t.Stage = "completed"
		t.Progress = 100
		t.Message = "旅行计划生成成功"
		t.Result = result
		t.Error = ""
	})
}

// attractionsFromMap 小红书不可用时的兜底:用地图 POI 检索景点并组装为相同的上下文格式。
func (p *TripPlanner) attractionsFromMap(ctx context.Context, city, keywords string) string {
	query := keywords
	if query == "" || query == "景点" {
		query = "景点"
	}
	var pois []model.POIInfo
	if CurrentMapProvider(p.settings) == "google" {
		if svc := NewGoogleMapService(p.settings); svc != nil {
			pois = svc.SearchPOI(ctx, query, city, true)
		}
	}
	if len(pois) == 0 {
		pois = p.amap.SearchPOI(ctx, query, city, true)
	}
	if len(pois) == 0 {
		return fmt.Sprintf("未能获取 %s 的景点数据,请基于该城市最著名的经典景点规划。", city)
	}

	var out strings.Builder
	out.WriteString("这是地图服务检索到的景点候选(真人内容源未启用,数据来源为地图 POI):\n")
	for _, poi := range pois {
		item := map[string]any{
			"name":                 poi.Name,
			"name_zh":              poi.Name,
			"name_en":              "",
			"reason":               firstNonEmpty(poi.Address, "地图热门景点"),
			"duration":             120,
			"reservation_required": false,
			"reservation_tips":     "",
		}
		if poi.Location.Longitude != 0 || poi.Location.Latitude != 0 {
			item["location"] = map[string]float64{
				"longitude": poi.Location.Longitude,
				"latitude":  poi.Location.Latitude,
			}
		}
		if line, err := json.Marshal(item); err == nil {
			out.Write(line)
			out.WriteString("\n")
		}
	}
	return out.String()
}

func (p *TripPlanner) updateTask(taskID string, mutate func(*TripTask)) {
	p.tasks.Update(taskID, mutate)
}

// budgetFromDays 按每日明细(门票/酒店/餐费)汇总生成预算,用于 LLM 遗漏 budget 字段时的兜底。
func budgetFromDays(days []model.DayPlan) *model.Budget {
	var attractions, hotels, meals int
	for _, day := range days {
		for _, a := range day.Attractions {
			attractions += a.TicketPrice.Int()
		}
		if day.Hotel != nil {
			hotels += day.Hotel.EstimatedCost.Int()
		}
		for _, m := range day.Meals {
			meals += m.EstimatedCost.Int()
		}
	}
	total := attractions + hotels + meals
	return &model.Budget{
		TotalAttractions: model.FlexInt(attractions),
		TotalHotels:      model.FlexInt(hotels),
		TotalMeals:       model.FlexInt(meals),
		Total:            model.FlexInt(total),
	}
}

// FallbackSuggestions overall_suggestions 缺失时按行程要素本地生成兜底建议。
// LLM 偶尔会遗漏该字段(尤其多城市任务),这里用确定性文案兜底,保证结果页与导出不为空。
func FallbackSuggestions(plan *model.TripPlan, cities []string, lang string) string {
	if plan == nil {
		return ""
	}
	if len(cities) == 0 && len(plan.Cities) > 0 {
		cities = plan.Cities
	}
	if len(cities) == 0 && plan.City != "" {
		cities = []string{plan.City}
	}

	// 统计行程要素:需预约景点 / 城际移动日 / 雨天
	reservations := make([]string, 0, 4)
	transferDays := 0
	rainDays := 0
	for _, day := range plan.Days {
		if day.IsTransferDay && strings.TrimSpace(day.TransferInfo) != "" {
			transferDays++
		}
		for _, a := range day.Attractions {
			if a.ReservationRequired {
				reservations = append(reservations, a.Name)
			}
		}
	}
	for _, w := range plan.WeatherInfo {
		if strings.Contains(w.DayWeather, "雨") || strings.Contains(w.NightWeather, "雨") {
			rainDays++
		}
	}

	// 预算概览(无 budget 时按明细汇总)
	budget := plan.Budget
	if budget == nil {
		budget = budgetFromDays(plan.Days)
	}

	// 按语言生成文案片段(条件项可能缺失,最后统一连续编号)
	var parts []string
	switch lang {
	case "en":
		parts = append(parts,
			fmt.Sprintf("This %d-day itinerary covers %s. Follow the daily schedule and allow extra time for popular attractions.", len(plan.Days), strings.Join(cities, " → ")))
		if len(reservations) > 0 {
			parts = append(parts, fmt.Sprintf("Reservation required: %s. Book tickets in advance, especially on holidays.", strings.Join(reservations, ", ")))
		} else {
			parts = append(parts, "Check opening hours and ticket policies of each attraction before departure.")
		}
		if transferDays > 0 {
			parts = append(parts, fmt.Sprintf("The itinerary includes %d inter-city transfer day(s). Refer to the daily transfer notes and avoid rush hours.", transferDays))
		}
		if rainDays > 0 {
			parts = append(parts, fmt.Sprintf("Rain is expected on %d day(s). Bring an umbrella and consider indoor alternatives.", rainDays))
		}
		parts = append(parts, fmt.Sprintf("Estimated total budget: CNY %d (attractions %d / hotels %d / meals %d / transport %d). Adjust by your own spending habits.",
			budget.Total.Int(), budget.TotalAttractions.Int(), budget.TotalHotels.Int(), budget.TotalMeals.Int(), budget.TotalTransportation.Int()))
	case "ja":
		parts = append(parts,
			fmt.Sprintf("この%d日間の旅程は%sをカバーします。人気スポットは時間に余裕を持って訪れてください。", len(plan.Days), strings.Join(cities, " → ")))
		if len(reservations) > 0 {
			parts = append(parts, fmt.Sprintf("事前予約が必要なスポット: %s。連休・祝日は早めの予約をおすすめします。", strings.Join(reservations, "、")))
		} else {
			parts = append(parts, "出発前に各スポットの営業時間とチケットポリシーをご確認ください。")
		}
		if transferDays > 0 {
			parts = append(parts, fmt.Sprintf("都市間移動日が%d日含まれます。各日の移動メモを参照し、ラッシュ時を避けてください。", transferDays))
		}
		if rainDays > 0 {
			parts = append(parts, fmt.Sprintf("%d日間は雨の予報です。傘をお持ちください。", rainDays))
		}
		parts = append(parts, fmt.Sprintf("概算総予算: %d元(観光%d/ホテル%d/食事%d/交通%d)。ご自身の支出に合わせて調整してください。",
			budget.Total.Int(), budget.TotalAttractions.Int(), budget.TotalHotels.Int(), budget.TotalMeals.Int(), budget.TotalTransportation.Int()))
	default: // 中文
		parts = append(parts,
			fmt.Sprintf("本次行程共%d天,途经 %s,建议按每日行程顺序游览,热门景点请预留充足排队与游览时间。", len(plan.Days), strings.Join(cities, " → ")))
		if len(reservations) > 0 {
			parts = append(parts, fmt.Sprintf("以下景点需要提前预约: %s。建议出行前在官方渠道预约门票,节假尤其要提早。", strings.Join(reservations, "、")))
		} else {
			parts = append(parts, "出行前请确认各景点开放时间与门票政策,避免现场闭馆或限流。")
		}
		if transferDays > 0 {
			parts = append(parts, fmt.Sprintf("行程包含%d个城际移动日,请参考当日交通建议,错峰出行并预留候车时间。", transferDays))
		}
		if rainDays > 0 {
			parts = append(parts, fmt.Sprintf("行程期间有%d天预报降雨,请随身携带雨具,并可为当天准备室内备选景点。", rainDays))
		}
		parts = append(parts, fmt.Sprintf("预估总预算约 %d 元(门票 %d / 住宿 %d / 餐饮 %d / 交通 %d),请结合个人消费习惯适当调整。",
			budget.Total.Int(), budget.TotalAttractions.Int(), budget.TotalHotels.Int(), budget.TotalMeals.Int(), budget.TotalTransportation.Int()))
	}
	// 统一连续编号
	lines := make([]string, len(parts))
	for i, text := range parts {
		lines[i] = fmt.Sprintf("%d. %s", i+1, text)
	}
	return strings.Join(lines, "\n")
}

func (p *TripPlanner) failTask(taskID, msg string) {
	fmt.Printf("❌ 任务 %s 失败: %s\n", taskID, msg)
	p.updateTask(taskID, func(t *TripTask) {
		t.Status = TripTaskFailed
		t.Stage = "failed"
		t.Progress = 100
		t.Message = msg
		t.Error = msg
	})
}

// WeatherFallback 供 handler 在返回历史/快照数据时补全缺失的天气(旧版生成遗漏 weather_info 的记录)。
func (p *TripPlanner) WeatherFallback(ctx context.Context, plan *model.TripPlan) {
	if plan == nil || len(plan.WeatherInfo) > 0 {
		return
	}
	cities := plan.Cities
	if len(cities) == 0 && plan.City != "" {
		cities = []string{plan.City}
	}
	plan.WeatherInfo = p.weatherFromForecast(ctx, cities, plan, CurrentMapProvider(p.settings))
}

// weatherFromForecast LLM 遗漏 weather_info 时的天气兜底:逐城市查询实时预报,
// 按日期匹配填充到对应行程日(预报仅覆盖未来 3~4 天,更远的日期不做虚构)。
func (p *TripPlanner) weatherFromForecast(ctx context.Context, cities []string, plan *model.TripPlan, provider string) []model.WeatherInfo {
	fc := map[string]map[string]model.WeatherInfo{}
	for _, city := range cities {
		var forecasts []model.WeatherInfo
		if provider == "google" {
			if svc := NewGoogleMapService(p.settings); svc != nil {
				forecasts = svc.GetWeather(ctx, city)
			}
		}
		if len(forecasts) == 0 {
			forecasts = p.amap.GetWeather(ctx, city)
		}
		if len(forecasts) == 0 {
			continue
		}
		byDate := map[string]model.WeatherInfo{}
		for _, f := range forecasts {
			byDate[f.Date] = f
		}
		fc[city] = byDate
	}
	if len(fc) == 0 {
		return nil
	}

	out := make([]model.WeatherInfo, 0, len(plan.Days))
	for _, day := range plan.Days {
		for _, city := range cities {
			if day.City != "" && day.City != city {
				continue
			}
			if f, ok := fc[city][day.Date]; ok {
				item := f
				item.City = firstNonEmpty(day.City, city)
				out = append(out, item)
				break
			}
		}
	}
	return out
}

// queryWeatherText 查询天气文本:优先 Google(配置时),失败自动降级高德。
func (p *TripPlanner) queryWeatherText(ctx context.Context, city, provider, langHint string) string {
	if provider == "google" {
		if svc := NewGoogleMapService(p.settings); svc != nil {
			list := svc.GetWeather(ctx, city)
			if len(list) == 0 {
				// Google 无数据,降级高德天气 REST
				fmt.Printf("  ⚠️ %s Google 天气查询失败,降级到高德天气 API...\n", city)
				return p.amap.WeatherText(ctx, city)
			}
			lines := make([]string, 0, len(list))
			for _, w := range list {
				lines = append(lines, fmt.Sprintf("%s: 白天%s %d°C, 夜间%s %d°C, %s风 %s",
					w.Date, w.DayWeather, w.DayTemp.Int(), w.NightWeather, w.NightTemp.Int(), w.WindDirection, w.WindPower))
			}
			return strings.Join(lines, "\n")
		}
	}
	text := p.amap.WeatherText(ctx, city)
	if langHint != "" {
		text += "\n" + strings.TrimSpace(langHint)
	}
	return text
}

// queryHotelText 搜索酒店文本(Google 优先,高德兜底)。
func (p *TripPlanner) queryHotelText(ctx context.Context, city, accommodation, provider, langHint string) string {
	keyword := "酒店"
	if strings.Contains(accommodation, "民宿") {
		keyword = "民宿"
	}

	var pois []model.POIInfo
	if provider == "google" {
		if svc := NewGoogleMapService(p.settings); svc != nil {
			pois = svc.SearchPOI(ctx, keyword, city, true)
		}
	}
	if len(pois) == 0 {
		pois = p.amap.SearchPOI(ctx, keyword, city, true)
	}
	if len(pois) == 0 {
		return fmt.Sprintf("暂时无法获取 %s 的%s信息,请基于常见%s给出合理建议。", city, keyword, accommodation)
	}
	lines := make([]string, 0, len(pois))
	for _, poi := range pois {
		line := fmt.Sprintf("- %s | 地址: %s | 类型: %s", poi.Name, poi.Address, poi.Type)
		if poi.Tel != "" {
			line += " | 电话: " + poi.Tel
		}
		if poi.Location.Longitude != 0 || poi.Location.Latitude != 0 {
			line += fmt.Sprintf(" | 坐标: %.6f,%.6f", poi.Location.Longitude, poi.Location.Latitude)
		}
		lines = append(lines, line)
	}
	text := fmt.Sprintf("以下是 %s 的候选%s列表:\n%s", city, keyword, strings.Join(lines, "\n"))
	if langHint != "" {
		text += "\n" + strings.TrimSpace(langHint)
	}
	return text
}

// runPlannerWithRetry 规划阶段超时后重试一次。
func (p *TripPlanner) runPlannerWithRetry(ctx context.Context, req *model.TripRequest,
	attractions, weather, hotels map[string]string, memorySnippet string) (string, error) {
	query := p.buildPlannerQuery(req, attractions, weather, hotels, memorySnippet)
	reply, err := p.llm.ChatWithTimeout(ctx, p.settings.PlannerTimeoutSeconds(), NewLLMMessages(tripPlannerPrompt, query), 0.2, 16000)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if !strings.Contains(lower, "timeout") && !strings.Contains(lower, "超时") && !strings.Contains(lower, "deadline") {
			return "", err
		}
		fmt.Println("⚠️  首次行程规划超时,正在重试一次...")
		retryQuery := query + "\n\n**补充要求:** 如果部分辅助信息不足,请使用保守、常见、可执行的建议补齐,但必须输出完整合法的 JSON,不要输出解释性文字。"
		return p.llm.ChatWithTimeout(ctx, p.settings.PlannerTimeoutSeconds(), NewLLMMessages(tripPlannerPrompt, retryQuery), 0.2, 16000)
	}
	return reply, nil
}

// buildPlannerQuery 构建行程规划查询(支持多城市)。
func (p *TripPlanner) buildPlannerQuery(req *model.TripRequest,
	attractions, weather, hotels map[string]string, memorySnippet string) string {
	cities := req.Cities
	isMultiCity := len(cities) > 1

	title := fmt.Sprintf("%s的%d天旅行计划", cities[0].City, req.TravelDays)
	citiesDesc := fmt.Sprintf("- %s: %d 天", cities[0].City, cities[0].Days)
	if isMultiCity {
		lines := make([]string, 0, len(cities))
		dayOffset := 0
		for _, cs := range cities {
			lines = append(lines, fmt.Sprintf("- %s: 停留 %d 天 (第%d天 ~ 第%d天)", cs.City, cs.Days, dayOffset+1, dayOffset+cs.Days))
			dayOffset += cs.Days
		}
		citiesDesc = strings.Join(lines, "\n")
		names := make([]string, 0, len(cities))
		for _, cs := range cities {
			names = append(names, cs.City)
		}
		title = fmt.Sprintf("跨城市旅行计划(%s)", strings.Join(names, " → "))
	}

	var b strings.Builder
	fmt.Fprintf(&b, `请根据以下信息生成%s:

**基本信息:**
- 途经城市及天数分配:
%s
- 总天数: %d天
- 日期: %s 至 %s
- 交通方式: %s
- 住宿: %s
- 偏好: %s
`, title, citiesDesc, req.TravelDays, req.StartDate, req.EndDate,
		req.Transportation, req.Accommodation, firstNonEmpty(strings.Join(req.Preferences, ", "), "无"))

	if memorySnippet != "" {
		fmt.Fprintf(&b, "\n%s\n请在规划时充分参考以上用户历史偏好,使其在景点选择、餐饮、住宿、行程节奏中体现。\n", memorySnippet)
	}

	for _, cs := range cities {
		city := cs.City
		if isMultiCity {
			fmt.Fprintf(&b, "\n--- %s (%d天) ---\n**%s 景点信息:**\n%s\n**%s 天气信息:**\n%s\n**%s 酒店信息:**\n%s\n",
				city, cs.Days, city, firstNonEmpty(attractions[city], "无"),
				city, firstNonEmpty(weather[city], "无"), city, firstNonEmpty(hotels[city], "无"))
		} else {
			fmt.Fprintf(&b, "\n**景点信息:**\n%s\n\n**天气信息:**\n%s\n\n**酒店信息:**\n%s\n",
				firstNonEmpty(attractions[city], "无"), firstNonEmpty(weather[city], "无"), firstNonEmpty(hotels[city], "无"))
		}
	}

	b.WriteString(`
**要求:**
1. 每天安排2-3个景点(城际移动日可减少为1-2个)
2. 每天必须包含早中晚三餐
3. 每天推荐一个具体的酒店(从酒店信息中选择)
4. 考虑景点之间的距离和交通方式
5. 返回完整的JSON格式数据
6. 景点的经纬度坐标要真实准确
7. 如果天气或酒店信息不足,请基于保守、通用的旅行建议补齐,但不要输出"无法查询"之类的解释文字
`)
	if isMultiCity {
		b.WriteString(`
**多城市特殊要求:**
1. 每个 day 对象中必须包含 "city" 字段标明当天所在城市
2. 城市切换当天标记 "is_transfer_day": true, 并在 "transfer_info" 中说明城际交通方式和预计时长
3. 城际移动日的景点数量可适当减少为 1-2 个
4. budget 中增加 "total_inter_city_transport" 字段统计城际交通费用
5. 景点顺序要考虑同城市内的地理位置关系
6. "cities" 数组列出所有途经城市名称
`)
	}
	if req.FreeTextInput != "" {
		fmt.Fprintf(&b, "\n**额外要求:** %s\n", req.FreeTextInput)
	}
	if lang := req.Lang(); lang != "zh" {
		name := langNames[lang]
		if name == "" {
			name = lang
		}
		fmt.Fprintf(&b, `
**语言要求 (Language Requirement):**
请用 %s 语言输出所有文字内容(包括 description, overall_suggestions, meals 中的 name/description, hotel 中的 name/address, attractions 中的 name/address/description 等)。
JSON 的 key 名称保持英文不变,只翻译 value 中的文字。`, name)
	}
	return b.String()
}

// asCookieExpired 判断错误链中是否包含小红书 Cookie 过期异常。
func asCookieExpired(err error, target **XHSCookieExpiredError) bool {
	for err != nil {
		if e, ok := err.(*XHSCookieExpiredError); ok {
			*target = e
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

// isDouyinCookieExpired 判断错误链中是否包含抖音 Cookie 过期异常。
func isDouyinCookieExpired(err error) bool {
	for err != nil {
		if _, ok := err.(*DouyinCookieExpiredError); ok {
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}
