// 景点详情页的实时数据:周边推荐与邻近交通站点。
// 不落 MySQL,只缓存在 Redis(cacheTTL 由 SCENIC_CACHE_TTL_HOURS 配置,默认 24h);
// Redis 不可用时直接回源高德,不影响页面。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"

	"wenlv-backend/repository"
)

// AroundItem 周边推荐条目。
type AroundItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Address  string  `json:"address"`
	Distance int     `json:"distance"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

// TransitStop 邻近交通站点。
type TransitStop struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Distance int    `json:"distance"`
}

// ScenicExtraService 景点实时信息查询。
type ScenicExtraService struct {
	repo     *repository.ScenicRepo
	amap     *AmapService
	rdb      *redis.Client
	cacheTTL time.Duration
}

// NewScenicExtraService 构造服务。cacheTTL <= 0 时回退 24 小时。
func NewScenicExtraService(repo *repository.ScenicRepo, amap *AmapService, rdb *redis.Client, cacheTTL time.Duration) *ScenicExtraService {
	if cacheTTL <= 0 {
		cacheTTL = 24 * time.Hour
	}
	return &ScenicExtraService{repo: repo, amap: amap, rdb: rdb, cacheTTL: cacheTTL}
}

// Around 返回景点周边 3km 内的景点/餐饮 POI,按距离升序;失败返回空切片而不是错误。
// 只取"风景名胜(110000)"与"餐饮服务(050000)":实测若带上"购物服务(060000)",
// 郊区景点会把五金店/建材市场/超市当成周边推荐(它们的分类确实属于购物服务)。
func (s *ScenicExtraService) Around(ctx context.Context, id uint, limit int) ([]AroundItem, error) {
	if limit <= 0 || limit > 12 {
		limit = 6
	}
	// key 带 v2 版本号:推荐类型收窄后,旧缓存(24h TTL)里的购物类结果需要绕过
	key := fmt.Sprintf("scenic:around:v2:%d:%d", id, limit)
	if cached := s.getCache(ctx, key); cached != nil {
		return cached, nil
	}
	spot, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	out := []AroundItem{}
	if spot.Lng == 0 || spot.Lat == 0 {
		return out, nil
	}
	pois := s.amap.SearchAround(ctx, spot.Lng, spot.Lat, "", "050000|110000", 3000, 20)
	sort.SliceStable(pois, func(i, j int) bool { return pois[i].Distance < pois[j].Distance })
	for _, p := range pois {
		if p.Name == spot.NameZH {
			continue
		}
		out = append(out, AroundItem{
			ID: p.ID, Name: p.Name, Type: p.Type, Address: p.Address,
			Distance: p.Distance, Lat: p.Location.Latitude, Lng: p.Location.Longitude,
		})
		if len(out) >= limit {
			break
		}
	}
	s.setCache(ctx, key, out)
	return out, nil
}

// Transport 返回景点附近 1.5km 内的地铁站与公交站,按距离升序。
func (s *ScenicExtraService) Transport(ctx context.Context, id uint) ([]TransitStop, error) {
	key := fmt.Sprintf("scenic:transit:%d", id)
	if cached := s.getTransitCache(ctx, key); cached != nil {
		return cached, nil
	}
	spot, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	out := []TransitStop{}
	if spot.Lng == 0 || spot.Lat == 0 {
		return out, nil
	}
	// 用高德 typecode 精确筛选交通设施:150700=地铁站,150500=公交车站。
	// 实测(春熙路坐标):keywords="地铁站|公交站" 会混入停车场/酒店等无关 POI,
	// 而空 keywords + types="150700|150500" 返回的全是公交站/地铁站。
	pois := s.amap.SearchAround(ctx, spot.Lng, spot.Lat, "", "150700|150500", 1500, 10)
	sort.SliceStable(pois, func(i, j int) bool { return pois[i].Distance < pois[j].Distance })
	for _, p := range pois {
		st := TransitStop{Name: p.Name, Type: p.Type, Distance: p.Distance}
		out = append(out, st)
		if len(out) >= 4 {
			break
		}
	}
	s.setTransitCache(ctx, key, out)
	return out, nil
}

// getCache 读取 JSON 缓存,未命中或 Redis 不可用时返回 nil。
func (s *ScenicExtraService) getCache(ctx context.Context, key string) []AroundItem {
	if s.rdb == nil {
		return nil
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil || len(raw) == 0 {
		return nil
	}
	var items []AroundItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

// setCache 写入 JSON 缓存。空结果不写缓存:避免把"高德暂时失败/暂无数据"
// 固化一整个 TTL,宁可下次请求再回源。
func (s *ScenicExtraService) setCache(ctx context.Context, key string, items []AroundItem) {
	if s.rdb == nil || len(items) == 0 {
		return
	}
	if raw, err := json.Marshal(items); err == nil {
		_ = s.rdb.Set(ctx, key, raw, s.cacheTTL).Err()
	}
}

func (s *ScenicExtraService) getTransitCache(ctx context.Context, key string) []TransitStop {
	if s.rdb == nil {
		return nil
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil || len(raw) == 0 {
		return nil
	}
	var items []TransitStop
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

// setTransitCache 写入交通站点缓存,规则同 setCache(空结果不缓存)。
func (s *ScenicExtraService) setTransitCache(ctx context.Context, key string, items []TransitStop) {
	if s.rdb == nil || len(items) == 0 {
		return
	}
	if raw, err := json.Marshal(items); err == nil {
		_ = s.rdb.Set(ctx, key, raw, s.cacheTTL).Err()
	}
}
