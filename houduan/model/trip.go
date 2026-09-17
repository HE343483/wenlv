package model

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// ============ 通用容错数值类型 ============
// 大模型输出常见 "120分钟"、"¥324"、25°C 等脏数据,这里统一做容错解析。

// FlexInt 兼容 数字 / 数字字符串 / 带单位文本 的整数。
type FlexInt int

// UnmarshalJSON 容错解析:剥离非数字字符后取整。
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	if s[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		s = text
	}
	s = strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '-' || r == '.' {
			return r
		}
		return -1
	}, s)
	if s == "" || s == "-" || s == "." {
		*f = 0
		return nil
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		*f = FlexInt(int(v + sign(v)*0.5))
		return nil
	}
	*f = 0
	return nil
}

// MarshalJSON 始终输出纯整数。
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(f))), nil
}

// Int 转为 int。
func (f FlexInt) Int() int { return int(f) }

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

// ============ 请求模型 ============

// CityStay 单城市停留配置。
type CityStay struct {
	City string `json:"city"`
	Days int    `json:"days"`
}

// TripRequest 旅行规划请求。
type TripRequest struct {
	City           string     `json:"city"`
	Cities         []CityStay `json:"cities"`
	StartDate      string     `json:"start_date"`
	EndDate        string     `json:"end_date"`
	TravelDays     int        `json:"travel_days"`
	Transportation string     `json:"transportation"`
	Accommodation  string     `json:"accommodation"`
	Preferences    []string   `json:"preferences"`
	FreeTextInput  string     `json:"free_text_input"`
	Language       string     `json:"language"`
	UserID         string     `json:"user_id"`
	// AttractionSource 景点数据来源:xhs=小红书真人推荐(默认) / map=高德地图检索
	AttractionSource string `json:"attraction_source"`
}

// Normalize 兼容处理:只填 city 未填 cities 时自动转换,并统一语言代码。
func (r *TripRequest) Normalize() {
	if len(r.Cities) == 0 && r.City != "" {
		days := r.TravelDays
		if days <= 0 {
			days = 1
		}
		r.Cities = []CityStay{{City: r.City, Days: days}}
	}
	if len(r.Cities) > 0 && r.City == "" {
		r.City = r.Cities[0].City
	}
	if r.StartDate == "" || r.EndDate == "" {
		// 由调用方保证,这里只做兜底
	}
	if r.Language == "" {
		r.Language = "zh"
	}
	// 景点来源归一化:仅接受 map,其余一律按 xhs(默认)处理
	r.AttractionSource = strings.ToLower(strings.TrimSpace(r.AttractionSource))
	if r.AttractionSource != "map" {
		r.AttractionSource = "xhs"
	}
}

// Lang 返回归一化后的语言代码(zh/en/ja...)。
func (r *TripRequest) Lang() string {
	return NormalizeLang(r.Language)
}

// NormalizeLang 将 zh-CN / en-US 之类的语言代码归一化为主语言。
func NormalizeLang(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	if l == "" {
		return "zh"
	}
	if i := strings.IndexAny(l, "-_"); i > 0 {
		l = l[:i]
	}
	return l
}

// POISearchRequest POI 搜索请求。
type POISearchRequest struct {
	Keywords  string `json:"keywords"`
	City      string `json:"city"`
	CityLimit bool   `json:"citylimit"`
}

// RouteRequest 路线规划请求。
type RouteRequest struct {
	OriginAddress      string `json:"origin_address"`
	DestinationAddress string `json:"destination_address"`
	OriginCity         string `json:"origin_city"`
	DestinationCity    string `json:"destination_city"`
	RouteType          string `json:"route_type"`
}

// ============ 行程数据模型 ============

// Location 经纬度。
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Attraction 景点信息。
type Attraction struct {
	Name                string    `json:"name"`
	Address             string    `json:"address"`
	Location            *Location `json:"location"`
	VisitDuration       FlexInt   `json:"visit_duration"`
	Description         string    `json:"description"`
	Category            string    `json:"category,omitempty"`
	Rating              *float64  `json:"rating,omitempty"`
	Photos              []string  `json:"photos,omitempty"`
	POIID               string    `json:"poi_id,omitempty"`
	ImageURL            string    `json:"image_url,omitempty"`
	TicketPrice         FlexInt   `json:"ticket_price"`
	ReservationRequired bool      `json:"reservation_required"`
	ReservationTips     string    `json:"reservation_tips,omitempty"`
}

// Meal 餐饮信息。
type Meal struct {
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	Address       string    `json:"address,omitempty"`
	Location      *Location `json:"location,omitempty"`
	Description   string    `json:"description,omitempty"`
	EstimatedCost FlexInt   `json:"estimated_cost"`
}

// Hotel 酒店信息。
type Hotel struct {
	Name          string    `json:"name"`
	Address       string    `json:"address,omitempty"`
	Location      *Location `json:"location,omitempty"`
	PriceRange    string    `json:"price_range,omitempty"`
	Rating        string    `json:"rating,omitempty"`
	Distance      string    `json:"distance,omitempty"`
	Type          string    `json:"type,omitempty"`
	EstimatedCost FlexInt   `json:"estimated_cost"`
}

// DayPlan 单日行程。
type DayPlan struct {
	Date           string       `json:"date"`
	DayIndex       int          `json:"day_index"`
	City           string       `json:"city,omitempty"`
	IsTransferDay  bool         `json:"is_transfer_day"`
	TransferInfo   string       `json:"transfer_info,omitempty"`
	Description    string       `json:"description"`
	Transportation string       `json:"transportation"`
	Accommodation  string       `json:"accommodation"`
	Hotel          *Hotel       `json:"hotel,omitempty"`
	Attractions    []Attraction `json:"attractions"`
	Meals          []Meal       `json:"meals"`
}

// WeatherInfo 天气信息。
type WeatherInfo struct {
	Date          string  `json:"date"`
	City          string  `json:"city,omitempty"`
	DayWeather    string  `json:"day_weather"`
	NightWeather  string  `json:"night_weather"`
	DayTemp       FlexInt `json:"day_temp"`
	NightTemp     FlexInt `json:"night_temp"`
	WindDirection string  `json:"wind_direction"`
	WindPower     string  `json:"wind_power"`
}

// Budget 预算信息。
type Budget struct {
	TotalAttractions      FlexInt `json:"total_attractions"`
	TotalHotels           FlexInt `json:"total_hotels"`
	TotalMeals            FlexInt `json:"total_meals"`
	TotalTransportation   FlexInt `json:"total_transportation"`
	TotalInterCityTransit FlexInt `json:"total_inter_city_transport"`
	Total                 FlexInt `json:"total"`
}

// TripPlan 旅行计划。
type TripPlan struct {
	City               string        `json:"city"`
	Cities             []string      `json:"cities"`
	StartDate          string        `json:"start_date"`
	EndDate            string        `json:"end_date"`
	Days               []DayPlan     `json:"days"`
	WeatherInfo        []WeatherInfo `json:"weather_info"`
	OverallSuggestions string        `json:"overall_suggestions"`
	Budget             *Budget       `json:"budget,omitempty"`
}

// ============ 知识图谱模型 ============

// GraphNode 图谱节点。
type GraphNode struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Category   int            `json:"category"`
	SymbolSize int            `json:"symbolSize"`
	ItemStyle  map[string]any `json:"itemStyle,omitempty"`
	Value      string         `json:"value"`
}

// GraphEdge 图谱边。
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

// GraphCategory 图谱分类。
type GraphCategory struct {
	Name string `json:"name"`
}

// KnowledgeGraphData 知识图谱数据。
type KnowledgeGraphData struct {
	Nodes      []GraphNode     `json:"nodes"`
	Edges      []GraphEdge     `json:"edges"`
	Categories []GraphCategory `json:"categories"`
}

// TripPlanResponse 旅行计划响应。
type TripPlanResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	PlanID    string              `json:"plan_id,omitempty"`
	Data      *TripPlan           `json:"data,omitempty"`
	GraphData *KnowledgeGraphData `json:"graph_data,omitempty"`
}

// ============ POI / 路线 / 天气响应 ============

// POIInfo POI 信息。
type POIInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
	Tel      string   `json:"tel,omitempty"`
}

// RouteInfo 路线信息。
type RouteInfo struct {
	Distance    float64 `json:"distance"`
	Duration    int     `json:"duration"`
	RouteType   string  `json:"route_type"`
	Description string  `json:"description"`
}

// ============ AI 行程问答 ============

// ChatMessage 单条对话消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TripChatRequest 行程问答请求。
type TripChatRequest struct {
	Message  string         `json:"message"`
	TripPlan map[string]any `json:"trip_plan"`
	History  []ChatMessage  `json:"history"`
}

// TripChatResponse 行程问答响应。
type TripChatResponse struct {
	Success bool   `json:"success"`
	Reply   string `json:"reply"`
}

// ============ 任务事件 ============

// TripTaskEvent WebSocket / 轮询返回的任务事件。
type TripTaskEvent struct {
	TaskID         string            `json:"task_id"`
	PlanID         string            `json:"plan_id"`
	Status         string            `json:"status"`
	Stage          string            `json:"stage"`
	Progress       int               `json:"progress"`
	Message        string            `json:"message"`
	Error          string            `json:"error,omitempty"`
	RequestPayload *TripRequest      `json:"request_payload,omitempty"`
	Result         *TripPlanResponse `json:"result,omitempty"`
}

// TripHistoryItem 历史计划摘要。
type TripHistoryItem struct {
	PlanID             string   `json:"plan_id"`
	TaskID             string   `json:"task_id"`
	City               string   `json:"city"`
	Cities             []string `json:"cities,omitempty"`
	StartDate          string   `json:"start_date"`
	EndDate            string   `json:"end_date"`
	TravelDays         int      `json:"travel_days"`
	UpdatedAt          string   `json:"updated_at"`
	OverallSuggestions string   `json:"overall_suggestions,omitempty"`
}

// UserMemory 用户偏好记忆(持久化到 MySQL)。
type UserMemory struct {
	MemoryID       string  `gorm:"primaryKey;size:64;comment:记忆ID" json:"memory_id"`
	UserID         string  `gorm:"index;size:64;comment:前端匿名用户ID" json:"user_id"`
	Content        string  `gorm:"size:512;comment:偏好内容" json:"content"`
	Source         string  `gorm:"size:32;comment:来源(explicit手动输入-implicit自动提取)" json:"source"`
	Weight         float64 `gorm:"comment:权重(随时间与访问衰减)" json:"weight"`
	CreateTime     float64 `gorm:"comment:创建时间(Unix秒)" json:"create_time"`
	LastAccessTime float64 `gorm:"comment:最近访问时间(Unix秒)" json:"last_access_time"`
}

// TableName 指定表名。
func (UserMemory) TableName() string { return "user_memories" }

// TripPlanRecord 已生成行程的持久化记录(落库后历史查询更快,支持回看与删除)。
type TripPlanRecord struct {
	ID                 uint      `gorm:"primaryKey;comment:行程记录ID" json:"id"`
	PlanID             string    `gorm:"size:64;uniqueIndex;comment:计划ID(与任务ID对应)" json:"plan_id"`
	TaskID             string    `gorm:"size:64;index;comment:生成该计划的任务ID" json:"task_id"`
	UserID             string    `gorm:"index;size:64;comment:用户ID(匿名或登录用户)" json:"user_id"`
	City               string    `gorm:"size:128;comment:主城市" json:"city"`
	Cities             string    `gorm:"size:512;comment:多城市路线(逗号分隔)" json:"cities"`
	StartDate          string    `gorm:"size:32;comment:出发日期" json:"start_date"`
	EndDate            string    `gorm:"size:32;comment:结束日期" json:"end_date"`
	TravelDays         int       `gorm:"comment:行程天数" json:"travel_days"`
	OverallSuggestions string    `gorm:"type:text;comment:AI总体建议" json:"overall_suggestions"`
	PlanJSON           string    `gorm:"type:longtext;comment:行程计划完整JSON" json:"-"`
	GraphJSON          string    `gorm:"type:longtext;comment:知识图谱JSON" json:"-"`
	RequestJSON        string    `gorm:"type:longtext;comment:原始规划请求JSON" json:"-"`
	CreatedAt          time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt          time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// TableName 指定表名。
func (TripPlanRecord) TableName() string { return "trip_plans" }

// TripSetting 行程模块运行时配置(设置页持久化到数据库,重启不丢失)。
type TripSetting struct {
	Key       string    `gorm:"primaryKey;size:64;comment:配置键" json:"key"`
	Value     string    `gorm:"type:text;comment:配置值" json:"value"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// TableName 指定表名。
func (TripSetting) TableName() string { return "trip_settings" }
