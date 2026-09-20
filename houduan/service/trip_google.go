package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"wenlv-backend/logger"
	"wenlv-backend/model"
)

// ============ Google Maps Platform 服务封装(与高德服务对等) ============

// GoogleMapService Google 地图服务。
type GoogleMapService struct {
	settings *TripSettings
	client   *http.Client
	apiKey   string
	proxyURL string
}

// NewGoogleMapService 根据当前配置构造 Google 地图服务;未配置 Key 时返回 nil。
func NewGoogleMapService(settings *TripSettings) *GoogleMapService {
	cur := settings.Snapshot()
	if cur.GoogleMapsAPIKey == "" {
		return nil
	}
	transport := &http.Transport{}
	if p := strings.TrimSpace(cur.GoogleMapsProxy); p != "" {
		if u, err := url.Parse(p); err == nil {
			switch strings.ToLower(u.Scheme) {
			case "socks5", "socks5h":
				// SOCKS5 代理:仅用于 Google Maps 请求,不影响其他服务
				if dialer, derr := proxy.SOCKS5("tcp", u.Host, nil, proxy.Direct); derr == nil {
					if ctxDialer, ok := dialer.(proxy.ContextDialer); ok {
						transport.DialContext = ctxDialer.DialContext
					}
				}
			default:
				transport.Proxy = http.ProxyURL(u)
			}
		}
	}
	return &GoogleMapService{
		settings: settings,
		client:   &http.Client{Timeout: 20 * time.Second, Transport: transport},
		apiKey:   cur.GoogleMapsAPIKey,
		proxyURL: cur.GoogleMapsProxy,
	}
}

func (s *GoogleMapService) get(ctx context.Context, endpoint string, params url.Values, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *GoogleMapService) postJSON(ctx context.Context, endpoint string, body any, headers map[string]string) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Geocode 地理编码(地址 → 坐标)。
func (s *GoogleMapService) Geocode(ctx context.Context, address, city string) *model.Location {
	if s.apiKey == "" {
		return nil
	}
	query := address
	if city != "" {
		query = address + ", " + city
	}
	params := url.Values{}
	params.Set("address", query)
	params.Set("key", s.apiKey)
	params.Set("language", "zh-CN")
	raw, err := s.get(ctx, "https://maps.googleapis.com/maps/api/geocode/json", params, nil)
	if err != nil {
		logger.Warnf("[Google] 地理编码失败 (%s): %v", address, err)
		return nil
	}
	var result struct {
		Results []struct {
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	if len(result.Results) == 0 {
		return nil
	}
	loc := result.Results[0].Geometry.Location
	return &model.Location{Longitude: loc.Lng, Latitude: loc.Lat}
}

// SearchPOI 使用 Places API (New) Text Search 搜索 POI。
func (s *GoogleMapService) SearchPOI(ctx context.Context, keywords, city string, cityLimit bool) []model.POIInfo {
	if s.apiKey == "" {
		return nil
	}
	textQuery := keywords
	if cityLimit && city != "" {
		textQuery = city + " " + keywords
	}
	headers := map[string]string{
		"X-Goog-Api-Key":   s.apiKey,
		"X-Goog-FieldMask": "places.id,places.displayName,places.formattedAddress,places.location,places.types,places.internationalPhoneNumber",
	}
	raw, err := s.postJSON(ctx, "https://places.googleapis.com/v1/places:searchText",
		map[string]any{"textQuery": textQuery, "languageCode": "zh-CN"}, headers)
	if err != nil {
		logger.Warnf("[Google] POI 搜索失败: %v", err)
		return nil
	}
	var result struct {
		Places []struct {
			ID          string `json:"id"`
			DisplayName struct {
				Text string `json:"text"`
			} `json:"displayName"`
			FormattedAddress string   `json:"formattedAddress"`
			Types            []string `json:"types"`
			Location         struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			InternationalPhoneNumber string `json:"internationalPhoneNumber"`
		} `json:"places"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	out := make([]model.POIInfo, 0, len(result.Places))
	for _, p := range result.Places {
		types := p.Types
		if len(types) > 3 {
			types = types[:3]
		}
		out = append(out, model.POIInfo{
			ID:      p.ID,
			Name:    p.DisplayName.Text,
			Type:    strings.Join(types, ","),
			Address: p.FormattedAddress,
			Location: model.Location{
				Longitude: p.Location.Longitude,
				Latitude:  p.Location.Latitude,
			},
			Tel: p.InternationalPhoneNumber,
		})
	}
	return out
}

// PlanRoute 使用 Directions API 规划路线。
func (s *GoogleMapService) PlanRoute(ctx context.Context, originAddr, destAddr, originCity, destCity, routeType string) (map[string]any, error) {
	if s.apiKey == "" {
		return nil, errors.New("Google Maps API Key 未配置")
	}
	mode := routeType
	switch routeType {
	case "walking", "driving", "transit":
	default:
		mode = "walking"
	}
	origin := originAddr
	if originCity != "" {
		origin = originAddr + ", " + originCity
	}
	destination := destAddr
	if destCity != "" {
		destination = destAddr + ", " + destCity
	}
	params := url.Values{}
	params.Set("origin", origin)
	params.Set("destination", destination)
	params.Set("mode", mode)
	params.Set("key", s.apiKey)
	params.Set("language", "zh-CN")
	raw, err := s.get(ctx, "https://maps.googleapis.com/maps/api/directions/json", params, nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Routes []struct {
			Legs []struct {
				Distance struct {
					Value int    `json:"value"`
					Text  string `json:"text"`
				} `json:"distance"`
				Duration struct {
					Value int    `json:"value"`
					Text  string `json:"text"`
				} `json:"duration"`
			} `json:"legs"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if len(result.Routes) == 0 || len(result.Routes[0].Legs) == 0 {
		return nil, errors.New("Google 未返回路线数据")
	}
	leg := result.Routes[0].Legs[0]
	return map[string]any{
		"distance":      float64(leg.Distance.Value),
		"duration":      leg.Duration.Value,
		"distance_text": leg.Distance.Text,
		"duration_text": leg.Duration.Text,
	}, nil
}

// GetPOIDetail 获取 Place 详情。
func (s *GoogleMapService) GetPOIDetail(ctx context.Context, poiID string) (map[string]any, error) {
	if s.apiKey == "" {
		return nil, errors.New("Google Maps API Key 未配置")
	}
	headers := map[string]string{
		"X-Goog-Api-Key":   s.apiKey,
		"X-Goog-FieldMask": "id,displayName,formattedAddress,location,types,photos,editorialSummary,rating,userRatingCount",
	}
	raw, err := s.get(ctx, "https://places.googleapis.com/v1/places/"+url.PathEscape(poiID), nil, headers)
	if err != nil {
		return nil, err
	}
	var detail map[string]any
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, err
	}
	return detail, nil
}

var googleConditionMap = map[string]string{
	"CLEAR": "晴", "MOSTLY_CLEAR": "晴",
	"PARTLY_CLOUDY": "多云", "MOSTLY_CLOUDY": "多云",
	"CLOUDY": "阴", "OVERCAST": "阴",
	"LIGHT_RAIN": "小雨", "RAIN": "中雨", "MODERATE_RAIN": "中雨", "HEAVY_RAIN": "大雨",
	"LIGHT_SNOW": "小雪", "SNOW": "中雪", "HEAVY_SNOW": "大雪",
	"THUNDERSTORM": "雷阵雨", "DRIZZLE": "毛毛雨", "FOG": "雾", "HAZE": "霾", "WIND": "大风",
}

// GetWeather 使用 Google Weather API 查询未来多日天气。
func (s *GoogleMapService) GetWeather(ctx context.Context, city string) []model.WeatherInfo {
	if s.apiKey == "" {
		return nil
	}
	loc := s.Geocode(ctx, city, "")
	if loc == nil {
		logger.Warnf("[Google] 天气查询: 无法解析城市 '%s' 的坐标", city)
		return nil
	}
	params := url.Values{}
	params.Set("key", s.apiKey)
	params.Set("location.latitude", fmt.Sprintf("%f", loc.Latitude))
	params.Set("location.longitude", fmt.Sprintf("%f", loc.Longitude))
	params.Set("days", "7")
	params.Set("languageCode", "zh-CN")
	params.Set("unitsSystem", "METRIC")

	raw, err := s.get(ctx, "https://weather.googleapis.com/v1/forecast/days:lookup", params, nil)
	if err != nil {
		logger.Warnf("[Google] 天气查询失败: %v", err)
		return nil
	}
	var result struct {
		ForecastDays []struct {
			DisplayDate struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"displayDate"`
			DaytimeForecast struct {
				WeatherCondition string `json:"weatherCondition"`
				Wind             struct {
					Direction struct {
						Cardinal string `json:"cardinal"`
					} `json:"direction"`
					Speed struct {
						Value float64 `json:"value"`
					} `json:"speed"`
				} `json:"wind"`
			} `json:"daytimeForecast"`
			NighttimeForecast struct {
				WeatherCondition string `json:"weatherCondition"`
			} `json:"nighttimeForecast"`
			MaxTemperature struct {
				Degrees float64 `json:"degrees"`
			} `json:"maxTemperature"`
			MinTemperature struct {
				Degrees float64 `json:"degrees"`
			} `json:"minTemperature"`
		} `json:"forecastDays"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	out := make([]model.WeatherInfo, 0, len(result.ForecastDays))
	for _, d := range result.ForecastDays {
		desc := d.DisplayDate
		windPower := "微风"
		switch speed := d.DaytimeForecast.Wind.Speed.Value; {
		case speed < 6:
			windPower = "微风"
		case speed < 12:
			windPower = "1-2级"
		case speed < 20:
			windPower = "3级"
		case speed < 29:
			windPower = "4级"
		case speed < 39:
			windPower = "5级"
		default:
			windPower = "6级以上"
		}
		out = append(out, model.WeatherInfo{
			Date:          fmt.Sprintf("%04d-%02d-%02d", desc.Year, desc.Month, desc.Day),
			City:          city,
			DayWeather:    mapGoogleCondition(d.DaytimeForecast.WeatherCondition),
			NightWeather:  mapGoogleCondition(d.NighttimeForecast.WeatherCondition),
			DayTemp:       model.FlexInt(int(d.MaxTemperature.Degrees + 0.5)),
			NightTemp:     model.FlexInt(int(d.MinTemperature.Degrees + 0.5)),
			WindDirection: d.DaytimeForecast.Wind.Direction.Cardinal,
			WindPower:     windPower,
		})
	}
	logger.Infof("[Google] 天气查询成功: %s, %d 天预报", city, len(out))
	return out
}

func mapGoogleCondition(code string) string {
	if v, ok := googleConditionMap[code]; ok {
		return v
	}
	return code
}

// ============ 双引擎调度层 ============

var (
	googleGeoFailedMu  sync.Mutex
	googleGeoFailedFlg bool
)

// CurrentMapProvider 返回当前使用的地图供应商:配置了 Google Key 则用 google,否则高德。
func CurrentMapProvider(settings *TripSettings) string {
	if settings.Snapshot().GoogleMapsAPIKey != "" {
		return "google"
	}
	return "amap"
}

// ResetGoogleGeoFailure 清除 Google 地理编码失败标记(配置更新后允许重新尝试)。
func ResetGoogleGeoFailure() {
	googleGeoFailedMu.Lock()
	defer googleGeoFailedMu.Unlock()
	googleGeoFailedFlg = false
}

// GeocodeUnified 统一地理编码:Google 优先(可用时),失败自动降级高德。
func GeocodeUnified(ctx context.Context, settings *TripSettings, amap *AmapService, address, city, addressZh, addressEn string) *model.Location {
	if CurrentMapProvider(settings) == "google" {
		googleGeoFailedMu.Lock()
		shouldTry := !googleGeoFailedFlg
		googleGeoFailedMu.Unlock()

		if shouldTry {
			if svc := NewGoogleMapService(settings); svc != nil {
				target := firstNonEmpty(addressEn, address)
				if loc := svc.Geocode(ctx, target, city); loc != nil {
					return loc
				}
			}
			googleGeoFailedMu.Lock()
			if !googleGeoFailedFlg {
				googleGeoFailedFlg = true
				logger.Warnf("[Dispatcher] Google 地理编码失败(后续景点采用高德): %s", firstNonEmpty(addressEn, address))
			}
			googleGeoFailedMu.Unlock()
		}
	}
	return amap.GeocodeRaw(ctx, firstNonEmpty(addressZh, address), city)
}