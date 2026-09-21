package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/model"
)

// ============ 高德地图 Web 服务(REST)封装 ============
// 原 Python 项目通过 amap-mcp-server(MCP)调用,这里改为直连官方 REST 接口,
// 行为等价且无需额外进程依赖。

const amapGeocodeRateLimit = 3 // 高德地理编码上限 3 次/秒

// AmapService 高德地图服务。
type AmapService struct {
	settings *TripSettings
	client   *http.Client

	rateMu    sync.Mutex
	rateTimes []time.Time

	warnMu   sync.Mutex
	warnKeys map[string]bool
}

// NewAmapService 构造高德服务。
func NewAmapService(settings *TripSettings) *AmapService {
	return &AmapService{
		settings: settings,
		client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: &http.Transport{Proxy: nil},
		},
		warnKeys: map[string]bool{},
	}
}

func (s *AmapService) apiKey() string {
	return s.settings.Snapshot().ViteAmapWebKey
}

// Available 是否已配置高德 Web 服务 Key。
func (s *AmapService) Available() bool { return s.apiKey() != "" }

// waitGeocodeSlot 限制地理编码请求启动速率,避免超过官方 3 次/秒上限。
func (s *AmapService) waitGeocodeSlot() {
	for {
		wait := time.Duration(0)
		now := time.Now()
		s.rateMu.Lock()
		for len(s.rateTimes) > 0 && now.Sub(s.rateTimes[0]) >= time.Second {
			s.rateTimes = s.rateTimes[1:]
		}
		if len(s.rateTimes) < amapGeocodeRateLimit {
			s.rateTimes = append(s.rateTimes, now)
			s.rateMu.Unlock()
			return
		}
		wait = time.Second - now.Sub(s.rateTimes[0])
		s.rateMu.Unlock()
		if wait > 0 {
			time.Sleep(wait)
		}
	}
}

func (s *AmapService) getJSON(ctx context.Context, endpoint string, params url.Values, out any) error {
	key := s.apiKey()
	if key == "" {
		return errors.New("高德地图 API Key 未配置,请先在前端设置页完成配置")
	}
	params.Set("key", key)
	params.Set("output", "JSON")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("高德响应解析失败: %s", truncateText(string(raw), 200))
	}
	return nil
}

// IPLocateResult 高德 IP 定位结果。
type IPLocateResult struct {
	Province string
	City     string
	Adcode   string
}

// IPLocate 按客户端 IP 定位省市(两级链路)。
// 1) 高德 v3/ip(Key 需在控制台开通"IP 定位"服务,未开通时静默返回空);
// 2) 兜底 ipwho.is 免费库取经纬度,再用高德逆地理编码转中文省市 + adcode。
// ip 为空或内网地址时由服务按出口 IP 定位;全部失败返回 nil,调用方降级默认城市。
func (s *AmapService) IPLocate(ctx context.Context, ip string) *IPLocateResult {
	if isPrivateIP(ip) {
		ip = ""
	}
	if r := s.ipLocateAmap(ctx, ip); r != nil {
		return r
	}
	return s.ipLocateRegeoFallback(ctx, ip)
}

// isPrivateIP 判断是否内网/回环地址:外部 IP 库不识别这类地址,需去掉参数走出口 IP 定位。
func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() || parsed.IsUnspecified()
}

// ipLocateAmap 高德 v3/ip 定位(境内 IP;Key 未开通该服务时返回空数据,视为失败)。
func (s *AmapService) ipLocateAmap(ctx context.Context, ip string) *IPLocateResult {
	params := url.Values{}
	if ip != "" {
		params.Set("ip", ip)
	}
	var result struct {
		Status   string `json:"status"`
		Province any    `json:"province"` // 直辖市/境外异常时高德可能返回 []，用 toStringValue 兜底
		City     any    `json:"city"`
		Adcode   any    `json:"adcode"` // Key 未开通 IP 定位服务时高德返回 [],string 解析会报错
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/ip", params, &result); err != nil {
		s.warnOnce("ip-locate", "[Amap] IP 定位请求失败: "+err.Error())
		return nil
	}
	adcode := toStringValue(result.Adcode)
	if result.Status != "1" || adcode == "" || adcode == "[]" {
		return nil
	}
	return &IPLocateResult{
		Province: toStringValue(result.Province),
		City:     toStringValue(result.City),
		Adcode:   adcode,
	}
}

// ipLocateRegeoFallback 兜底定位:ipwho.is 查 IP 的经纬度,再用高德逆地理编码换取中文省市与 adcode。
func (s *AmapService) ipLocateRegeoFallback(ctx context.Context, ip string) *IPLocateResult {
	endpoint := "https://ipwho.is/"
	if ip != "" {
		endpoint += ip
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.warnOnce("ip-locate-fallback", "[Amap] ipwho.is 请求失败: "+err.Error())
		return nil
	}
	defer resp.Body.Close()
	var who struct {
		Success   bool     `json:"success"`
		Country   string   `json:"country"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8192)).Decode(&who); err != nil ||
		!who.Success || who.Latitude == nil || who.Longitude == nil {
		return nil
	}

	// 非中国 IP 不再反查(高德逆地理对境外支持有限,且本站面向境内访客)
	if !strings.Contains(who.Country, "China") {
		return nil
	}

	params := url.Values{}
	params.Set("location", fmt.Sprintf("%.6f,%.6f", *who.Longitude, *who.Latitude))
	var regeo struct {
		Status    string `json:"status"`
		Regeocode struct {
			AddressComponent struct {
				Province any    `json:"province"`
				City     any    `json:"city"`
				Adcode   string `json:"adcode"`
			} `json:"addressComponent"`
		} `json:"regeocode"`
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/geocode/regeo", params, &regeo); err != nil {
		s.warnOnce("ip-locate-fallback", "[Amap] IP 兜底逆地理失败: "+err.Error())
		return nil
	}
	if regeo.Status != "1" || regeo.Regeocode.AddressComponent.Adcode == "" {
		return nil
	}
	ac := regeo.Regeocode.AddressComponent
	province := toStringValue(ac.Province)
	// 直辖市高德返回 city 为 [],此时 province 即城市名
	return &IPLocateResult{
		Province: province,
		City:     firstNonEmpty(toStringValue(ac.City), province),
		Adcode:   ac.Adcode,
	}
}

// GeocodeRaw 地理编码(地址 → 坐标),失败返回 nil。
func (s *AmapService) GeocodeRaw(ctx context.Context, address, city string) *model.Location {
	if s.apiKey() == "" || strings.TrimSpace(address) == "" {
		return nil
	}
	params := url.Values{}
	params.Set("address", address)
	if city != "" {
		params.Set("city", city)
	}
	var result struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		Geocodes []struct {
			Location string `json:"location"`
		} `json:"geocodes"`
	}
	s.waitGeocodeSlot()
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/geocode/geo", params, &result); err != nil {
		return nil
	}
	if result.Status != "1" || len(result.Geocodes) == 0 {
		s.warnOnce(address+"|"+result.Info, fmt.Sprintf("高德地理编码无结果 (%s): status=%s info=%s", address, result.Status, result.Info))
		return nil
	}
	parts := strings.Split(result.Geocodes[0].Location, ",")
	if len(parts) != 2 {
		return nil
	}
	lon, err1 := strconv.ParseFloat(parts[0], 64)
	lat, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	return &model.Location{Longitude: lon, Latitude: lat}
}

// GeocodeAdcode 地理编码并返回行政区域 adcode(高新区/天府新区等非标准区名
// 直接查天气会失败,先转 adcode 再查),失败返回空串。
func (s *AmapService) GeocodeAdcode(ctx context.Context, address, city string) string {
	if s.apiKey() == "" || strings.TrimSpace(address) == "" {
		return ""
	}
	params := url.Values{}
	params.Set("address", address)
	if city != "" {
		params.Set("city", city)
	}
	var result struct {
		Status   string `json:"status"`
		Geocodes []struct {
			Adcode string `json:"adcode"`
		} `json:"geocodes"`
	}
	s.waitGeocodeSlot()
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/geocode/geo", params, &result); err != nil {
		return ""
	}
	if result.Status != "1" || len(result.Geocodes) == 0 {
		return ""
	}
	return strings.TrimSpace(result.Geocodes[0].Adcode)
}

// GetDistricts 高德行政区划查询(逐级获取 市→区县→镇/街道),返回原始 JSON。
// keywords 支持名称或 adcode,subdistrict 为向下级数(1-3)。
func (s *AmapService) GetDistricts(ctx context.Context, keywords, subdistrict string) (map[string]any, error) {
	if s.apiKey() == "" {
		return nil, fmt.Errorf("高德 Key 未配置")
	}
	if strings.TrimSpace(keywords) == "" {
		keywords = "四川省"
	}
	if subdistrict == "" {
		subdistrict = "1"
	}
	params := url.Values{}
	params.Set("keywords", keywords)
	params.Set("subdistrict", subdistrict)

	var out map[string]any
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/config/district", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AmapService) warnOnce(key, msg string) {
	s.warnMu.Lock()
	defer s.warnMu.Unlock()
	if s.warnKeys[key] {
		return
	}
	s.warnKeys[key] = true
	logger.Warnf("%s", msg)
}

// SearchPOI POI 关键词搜索。
func (s *AmapService) SearchPOI(ctx context.Context, keywords, city string, cityLimit bool) []model.POIInfo {
	if s.apiKey() == "" || strings.TrimSpace(keywords) == "" {
		return nil
	}
	params := url.Values{}
	params.Set("keywords", keywords)
	if city != "" {
		params.Set("city", city)
	}
	params.Set("citylimit", strconv.FormatBool(cityLimit))
	params.Set("offset", "20")
	params.Set("page", "1")
	params.Set("extensions", "base")

	var result struct {
		Status string `json:"status"`
		Pois   []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Address  any    `json:"address"`
			Location string `json:"location"`
			Tel      any    `json:"tel"`
		} `json:"pois"`
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/place/text", params, &result); err != nil {
		logger.Warnf("高德 POI 搜索失败: %v", err)
		return nil
	}
	out := make([]model.POIInfo, 0, len(result.Pois))
	for _, p := range result.Pois {
		loc := parseAmapLocation(p.Location)
		out = append(out, model.POIInfo{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type,
			Address:  toStringValue(p.Address),
			Location: loc,
			Tel:      toStringValue(p.Tel),
		})
	}
	return out
}

// POIDetailResult POI 详情。
type POIDetailResult struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Address   string         `json:"address"`
	Location  Location       `json:"location"`
	Tel       string         `json:"tel,omitempty"`
	Photos    []string       `json:"photos,omitempty"`
	Rating    string         `json:"rating,omitempty"`
	Cost      string         `json:"cost,omitempty"`
	OpenHours string         `json:"open_hours,omitempty"`
	BizExt    map[string]any `json:"biz_ext,omitempty"`
}

// Location 为 POI 详情使用的坐标结构(与 model.Location 字段一致)。
type Location = model.Location

// GetPOIDetail 获取 POI 详情(含图片)。
func (s *AmapService) GetPOIDetail(ctx context.Context, poiID string) (*POIDetailResult, error) {
	if s.apiKey() == "" {
		return nil, errors.New("高德地图 API Key 未配置")
	}
	params := url.Values{}
	params.Set("id", poiID)
	params.Set("extensions", "all")

	var result struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Pois   []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Address  any    `json:"address"`
			Location string `json:"location"`
			Tel      any    `json:"tel"`
			Photos   []struct {
				URL string `json:"url"`
			} `json:"photos"`
			BizExt map[string]any `json:"biz_ext"`
		} `json:"pois"`
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/place/detail", params, &result); err != nil {
		return nil, err
	}
	if result.Status != "1" || len(result.Pois) == 0 {
		return nil, fmt.Errorf("未查询到 POI 详情: %s", result.Info)
	}
	p := result.Pois[0]
	detail := &POIDetailResult{
		ID:       p.ID,
		Name:     p.Name,
		Type:     p.Type,
		Address:  toStringValue(p.Address),
		Location: parseAmapLocation(p.Location),
		Tel:      toStringValue(p.Tel),
	}
	for _, ph := range p.Photos {
		if ph.URL != "" {
			detail.Photos = append(detail.Photos, ph.URL)
		}
	}
	detail.BizExt = p.BizExt
	detail.Rating = pickBizExt(p.BizExt, "rating")
	detail.Cost = pickBizExt(p.BizExt, "cost")
	detail.OpenHours = pickOpenHours(p.BizExt)
	return detail, nil
}

// GetWeather 查询城市天气预报(未来 3~4 天)。
func (s *AmapService) GetWeather(ctx context.Context, city string) []model.WeatherInfo {
	if s.apiKey() == "" || strings.TrimSpace(city) == "" {
		return nil
	}
	params := url.Values{}
	params.Set("city", city)
	params.Set("extensions", "all")

	var result struct {
		Status    string `json:"status"`
		Forecasts []struct {
			Casts []struct {
				Date         string `json:"date"`
				DayWeather   string `json:"dayweather"`
				NightWeather string `json:"nightweather"`
				DayTemp      string `json:"daytemp"`
				NightTemp    string `json:"nighttemp"`
				DayWind      string `json:"daywind"`
				DayPower     string `json:"daypower"`
			} `json:"casts"`
		} `json:"forecasts"`
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/weather/weatherInfo", params, &result); err != nil {
		logger.Warnf("高德天气查询失败: %v", err)
		return nil
	}
	if len(result.Forecasts) == 0 {
		return nil
	}
	out := make([]model.WeatherInfo, 0, len(result.Forecasts[0].Casts))
	for _, c := range result.Forecasts[0].Casts {
		out = append(out, model.WeatherInfo{
			Date:          c.Date,
			City:          city,
			DayWeather:    c.DayWeather,
			NightWeather:  c.NightWeather,
			DayTemp:       parseFlexInt(c.DayTemp),
			NightTemp:     parseFlexInt(c.NightTemp),
			WindDirection: c.DayWind,
			WindPower:     c.DayPower,
		})
	}
	return out
}

// WeatherText 以文本形式返回天气,供行程规划 Prompt 使用。
func (s *AmapService) WeatherText(ctx context.Context, city string) string {
	// 高德天气严格要求城市名或 adcode,"中国-北京" 这类前缀需要裁掉
	cleanCity := city
	if idx := strings.LastIndex(cleanCity, "-"); idx >= 0 {
		cleanCity = strings.TrimSpace(cleanCity[idx+1:])
	}
	list := s.GetWeather(ctx, cleanCity)
	if len(list) == 0 {
		return fmt.Sprintf("暂时无法获取 %s 的天气数据", city)
	}
	lines := make([]string, 0, len(list))
	for _, w := range list {
		lines = append(lines, fmt.Sprintf("%s: 白天%s %d°C, 夜间%s %d°C, %s风 %s",
			w.Date, w.DayWeather, w.DayTemp.Int(), w.NightWeather, w.NightTemp.Int(), w.WindDirection, w.WindPower))
	}
	return strings.Join(lines, "\n")
}

// PlanRoute 规划两点之间的路线(内部先地理编码再调用对应方向接口)。
func (s *AmapService) PlanRoute(ctx context.Context, originAddr, destAddr, originCity, destCity, routeType string) (map[string]any, error) {
	if s.apiKey() == "" {
		return nil, errors.New("高德地图 API Key 未配置")
	}
	origin := s.GeocodeRaw(ctx, originAddr, originCity)
	dest := s.GeocodeRaw(ctx, destAddr, destCity)
	if origin == nil || dest == nil {
		return nil, errors.New("起点或终点地理编码失败")
	}
	originStr := fmt.Sprintf("%.6f,%.6f", origin.Longitude, origin.Latitude)
	destStr := fmt.Sprintf("%.6f,%.6f", dest.Longitude, dest.Latitude)

	params := url.Values{}
	params.Set("origin", originStr)
	params.Set("destination", destStr)

	endpoint := "https://restapi.amap.com/v3/direction/walking"
	switch routeType {
	case "driving":
		endpoint = "https://restapi.amap.com/v3/direction/driving"
	case "transit":
		endpoint = "https://restapi.amap.com/v3/direction/transit/integrated"
		params.Set("city", firstNonEmpty(originCity, destCity))
		params.Set("cityd", firstNonEmpty(destCity, originCity))
		params.Set("extensions", "base")
	}

	var result struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Route  struct {
			Paths []struct {
				Distance string `json:"distance"`
				Duration string `json:"duration"`
			} `json:"paths"`
		} `json:"route"`
	}
	if err := s.getJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}
	if result.Status != "1" || len(result.Route.Paths) == 0 {
		return nil, fmt.Errorf("未获取到路线数据: %s", result.Info)
	}
	path := result.Route.Paths[0]
	distance, _ := strconv.ParseFloat(path.Distance, 64)
	duration, _ := strconv.Atoi(path.Duration)
	return map[string]any{
		"distance":      distance,
		"duration":      duration,
		"distance_text": formatDistanceZH(distance),
	}, nil
}

func parseAmapLocation(raw string) model.Location {
	parts := strings.Split(raw, ",")
	if len(parts) != 2 {
		return model.Location{}
	}
	lon, _ := strconv.ParseFloat(parts[0], 64)
	lat, _ := strconv.ParseFloat(parts[1], 64)
	return model.Location{Longitude: lon, Latitude: lat}
}

func parseFlexInt(raw string) model.FlexInt {
	var v model.FlexInt
	_ = json.Unmarshal([]byte(strconv.Quote(raw)), &v)
	return v
}

func toStringValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprintf("%v", t)
	}
}

func formatDistanceZH(meters float64) string {
	if meters >= 1000 {
		return fmt.Sprintf("约%.1f公里", meters/1000)
	}
	return fmt.Sprintf("约%.0f米", meters)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// pickOpenHours 从高德 biz_ext 中提取开放时间。
// 高德不同 POI 回传的键名不一致,按优先级取第一个非空值;全部为空返回空串,
// 由人工或 LLM 参考值兜底(高德该字段覆盖率不高是已知情况)。
func pickOpenHours(bizExt map[string]any) string {
	for _, k := range []string{"open_time", "opentime_today", "opentime2", "opentime_week"} {
		v := strings.TrimSpace(toStringValue(bizExt[k]))
		if v == "" || v == "[]" {
			continue
		}
		return v
	}
	return ""
}

// pickBizExt 读取 biz_ext 中的单个字符串字段。
func pickBizExt(bizExt map[string]any, key string) string {
	return strings.TrimSpace(toStringValue(bizExt[key]))
}

// AroundPOI 周边 POI(含直线距离,单位米)。
type AroundPOI struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
	Distance int      `json:"distance"`
	Tel      string   `json:"tel,omitempty"`
}

// SearchAround 以坐标为中心做周边搜索(高德 place/around)。
// keywords 与 types 至少给一个;两者都空时用 types 兜底为风景名胜,
// 避免高德返回参数错误。失败时返回 nil 由调用方降级为空数组。
func (s *AmapService) SearchAround(ctx context.Context, lng, lat float64, keywords, types string, radius, offset int) []AroundPOI {
	if s.apiKey() == "" {
		return nil
	}
	if radius <= 0 || radius > 50000 {
		radius = 3000
	}
	if offset <= 0 || offset > 25 {
		offset = 10
	}
	if keywords == "" && types == "" {
		types = "060000"
	}
	params := url.Values{}
	params.Set("location", fmt.Sprintf("%.6f,%.6f", lng, lat))
	params.Set("keywords", keywords)
	params.Set("types", types)
	params.Set("radius", strconv.Itoa(radius))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("page", "1")
	params.Set("extensions", "base")

	var result struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Pois   []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Address  any    `json:"address"`
			Location string `json:"location"`
			Distance any    `json:"distance"`
			Tel      any    `json:"tel"`
		} `json:"pois"`
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/place/around", params, &result); err != nil {
		logger.Warnf("高德周边搜索失败: %v", err)
		return nil
	}
	if result.Status != "1" {
		logger.Warnf("高德周边搜索返回异常: %s", result.Info)
		return nil
	}
	out := make([]AroundPOI, 0, len(result.Pois))
	for _, p := range result.Pois {
		out = append(out, AroundPOI{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type,
			Address:  toStringValue(p.Address),
			Location: parseAmapLocation(p.Location),
			Distance: int(parseFlexInt(toStringValue(p.Distance))),
			Tel:      toStringValue(p.Tel),
		})
	}
	return out
}
