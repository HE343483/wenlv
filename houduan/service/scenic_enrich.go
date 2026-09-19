// 景点详情数据采集与补全:高德 POI 事实字段 + LLM 参考值 + 图文详情分段。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
)

// EnrichOptions 采集开关。
type EnrichOptions struct {
	WithImages bool // 抓取 POI 相册并上传 OSS
	WithLLM    bool // 调用 LLM 生成图文详情与参考值字段
	Force      bool // 已有值时也覆盖(默认只补空字段)
}

// EnrichResult 单个景点的采集结果,用于批处理日志。
type EnrichResult struct {
	ScenicID      uint
	Name          string
	POIMatched    bool
	Uploaded      int
	LLMUsed       bool
	UpdatedFields []string
	Note          string
}

// ScenicEnricher 景点数据采集器。
type ScenicEnricher struct {
	repo          *repository.ScenicRepo
	amap          *AmapService
	llm           *TripLLM
	signer        *pkg.OssSigner
	xhs           *XHSService // 小红书图源,未配置时为 nil
	xhsState      xhsShortCircuit
	commonsClient *http.Client // Wikimedia Commons 专用客户端(可选代理)
}

// NewScenicEnricher 构造采集器。signer 未配置时自动跳过图片上传;xhs 为 nil 时跳过小红书图源。
func NewScenicEnricher(repo *repository.ScenicRepo, amap *AmapService, llm *TripLLM, signer *pkg.OssSigner, xhs *XHSService) *ScenicEnricher {
	return &ScenicEnricher{
		repo:          repo,
		amap:          amap,
		llm:           llm,
		signer:        signer,
		xhs:           xhs,
		commonsClient: newCommonsHTTPClient(),
	}
}

// Enrich 采集单个景点并落库。任一步失败只记录 Note 并继续,不返回错误中断批处理。
func (e *ScenicEnricher) Enrich(ctx context.Context, s model.ScenicSpot, opts EnrichOptions) (*EnrichResult, error) {
	res := &EnrichResult{ScenicID: s.ID, Name: s.NameZH}
	updates := map[string]any{}

	// 1) 匹配高德 POI 并取详情(地址/电话/开放时间)
	poiID, poiName, ok := e.MatchPOI(ctx, s)
	var detail *POIDetailResult
	if ok {
		res.POIMatched = true
		d, err := e.amap.GetPOIDetail(ctx, poiID)
		if err != nil {
			res.Note = "POI 详情获取失败: " + err.Error()
		} else {
			detail = d
			updates["amap_poi_id"] = poiID
			if v := strings.TrimSpace(d.Address); v != "" {
				updates["address"] = v
			}
			if v := strings.TrimSpace(d.Tel); v != "" {
				updates["tel"] = v
			}
			if v := strings.TrimSpace(d.OpenHours); v != "" {
				updates["open_hours"] = v
			}
			res.Note = "匹配 POI: " + poiName
		}
	} else {
		res.Note = "未匹配到高德 POI"
	}

	// 2) 相册图片:多来源候选(高德/小红书/Commons)下载后上传 OSS,数据库只存 OSS URL
	images := splitComma(s.GalleryImages)
	if opts.WithImages && (opts.Force || len(images) == 0) && e.signer != nil && e.signer.Configured() {
		var amapPhotos []string
		if detail != nil {
			amapPhotos = detail.Photos
		}
		candidates := e.collectImageCandidates(ctx, s, amapPhotos, scenicGalleryWant)
		if len(candidates) > 0 {
			counts := map[string]int{}
			for _, c := range candidates {
				counts[c.Source]++
			}
			log.Printf("    图片: 高德%d 小红书%d Commons%d",
				counts[scenicImageSourceAmap], counts[scenicImageSourceXHS], counts[scenicImageSourceCommons])
			uploaded, err := e.uploadPhotos(ctx, s, candidates)
			if err != nil {
				res.Note += "; 图片上传失败: " + err.Error()
			} else if len(uploaded) > 0 {
				images = uploaded
				res.Uploaded = len(uploaded)
				updates["gallery_images"] = strings.Join(uploaded, ",")
			}
		}
	}
	if isPlaceholderCover(s.Images) && len(images) > 0 {
		updates["images"] = images[0]
	}

	// 3) LLM:图文详情 + 参考值字段(段落数按可用图片数对齐,保证每段有图)
	maxSections := clampMaxSections(len(images))
	var llmFields []string
	if opts.WithLLM && e.llm != nil && e.llm.Available() {
		sections, err := e.buildLLMSections(ctx, s, images, maxSections)
		if err != nil {
			res.Note += "; 图文详情生成失败: " + err.Error()
		} else if len(sections) > 0 {
			if raw, err := json.Marshal(sections); err == nil {
				updates["detail_sections"] = string(raw)
				res.LLMUsed = true
			}
		}
		est, err := e.fillEstimated(ctx, s)
		if err != nil {
			res.Note += "; 参考值生成失败: " + err.Error()
		} else {
			filled := 0
			for _, kv := range filterEstimated(est, updates) {
				if v := strings.TrimSpace(kv.Value); v != "" {
					updates[kv.Field] = v
					llmFields = append(llmFields, kv.Field)
					filled++
				}
			}
			if filled > 0 {
				res.LLMUsed = true
			}
		}
	}

	// 4) 非空字段保护:非 Force 模式下不覆盖已有事实值
	if !opts.Force {
		for key := range updates {
			if !isEmptyField(s, key) {
				delete(updates, key)
			}
		}
	}

	// 4.1) 参考值登记:只记本次真正写进库的 LLM 字段,
	// 被非 Force 保护删掉的字段不再被前端打上"参考值"徽标。
	estimated := survivingEstimated(s.EstimatedFields, llmFields, updates)

	// 5) 空跑保护:本次没有任何业务字段变更时不写库,
	// 避免把已有的 data_source / data_updated_at 等溯源信息降级。
	if !hasBusinessUpdate(updates) {
		res.Note += "; 无字段需要更新"
		return res, nil
	}

	updates["estimated_fields"] = estimated
	updates["data_source"] = buildDataSource(res.POIMatched, res.LLMUsed)
	now := time.Now()
	updates["data_updated_at"] = now
	for k := range updates {
		res.UpdatedFields = append(res.UpdatedFields, k)
	}
	if err := e.repo.UpdateFields(s.ID, updates); err != nil {
		return res, err
	}
	return res, nil
}

// estimatedKV LLM 参考值字段。
type estimatedKV struct {
	Field string
	Value string
}

// unusablePOITypes 不可用的高德 POI 类型关键词。
// 「地名地址信息」(热点地名/道路名) 只有区县级地址、无电话与开放时间;
// 「交通设施服务」(地铁站/公交站/停车场) 是交通设施而非景点,都不能作为景点数据来源。
var unusablePOITypes = []string{"地名地址信息", "交通设施服务", "地铁站", "公交站", "停车场"}

// isUsablePOI 判断高德返回的 POI type 是否可作为景点数据来源。
func isUsablePOI(poiType string) bool {
	for _, bad := range unusablePOITypes {
		if strings.Contains(poiType, bad) {
			return false
		}
	}
	return true
}

// unusablePOINameKeywords 名字里带这些词的 POI 是景区的附属设施而非景点本身
// (实测:景点「凤凰湖湿地公园」的同名候选只有「凤凰湖湿地公园-公共厕所」,会被误当景点)。
var unusablePOINameKeywords = []string{"公共厕所", "厕所", "卫生间", "停车场", "出入口", "送车点", "售票", "收费处", "岗亭"}

// isUsablePOIName 判断候选 POI 的名字是否为景点本身,而不是厕所/停车场等附属设施。
func isUsablePOIName(name string) bool {
	for _, bad := range unusablePOINameKeywords {
		if strings.Contains(name, bad) {
			return false
		}
	}
	return true
}

// MatchPOI 在候选 POI 中挑最可信的一个:名称完全相同 > 名称互相包含,
// 并在库内坐标有效时要求直线距离 < 5km,避免同名景点跨城市错配。
func (e *ScenicEnricher) MatchPOI(ctx context.Context, s model.ScenicSpot) (string, string, bool) {
	candidates := e.amap.SearchPOI(ctx, s.NameZH, "成都", true)
	if len(candidates) == 0 {
		candidates = e.amap.SearchPOI(ctx, s.NameZH, "", false)
	}
	best, bestName, bestScore := "", "", -1
	for _, c := range candidates {
		if !isUsablePOI(c.Type) || !isUsablePOIName(c.Name) {
			continue
		}
		score := 0
		if c.Name == s.NameZH {
			score = 3
		} else if strings.Contains(c.Name, s.NameZH) || strings.Contains(s.NameZH, c.Name) {
			score = 2
		}
		if score == 0 {
			continue
		}
		if s.Lat != 0 && s.Lng != 0 && c.Location.Latitude != 0 && c.Location.Longitude != 0 {
			if distanceMeters(s.Lng, s.Lat, c.Location.Longitude, c.Location.Latitude) > 5000 {
				continue
			}
			score++
		}
		if score > bestScore {
			best, bestName, bestScore = c.ID, c.Name, score
		}
	}
	return best, bestName, best != ""
}

// uploadPhotos 下载候选图片并上传 OSS,返回 OSS URL 列表(顺序与候选一致)。
// 单张失败跳过,不影响其余图片。实际下载与上传由包级 uploadImageCandidates 完成。
func (e *ScenicEnricher) uploadPhotos(ctx context.Context, s model.ScenicSpot, candidates []ScenicImageCandidate) ([]string, error) {
	if e.signer == nil || !e.signer.Configured() {
		return nil, fmt.Errorf("OSS 未配置")
	}
	urls, _ := uploadImageCandidates(ctx, e.commonsClient, e.signer, fmt.Sprintf("scenic/%d", s.ID), candidates, scenicGalleryWant)
	return urls, nil
}

// clampMaxSections 计算图文详情的段落数上限:第 1 张图作封面,其余每段配 1 张,夹在 [2,4] 区间。
func clampMaxSections(imageCount int) int {
	n := imageCount - 1
	if n < 2 {
		return 2
	}
	if n > 4 {
		return 4
	}
	return n
}

// buildLLMSections 调 LLM 把已有简介改写成 2-maxSections 段图文详情,并按顺序配相册图。
func (e *ScenicEnricher) buildLLMSections(ctx context.Context, s model.ScenicSpot, images []string, maxSections int) ([]model.ScenicDetailSection, error) {
	if strings.TrimSpace(s.Desc) == "" {
		return nil, nil
	}
	sectionRange := fmt.Sprintf("2-%d 段", maxSections)
	if maxSections <= 2 {
		sectionRange = "2 段"
	}
	prompt := fmt.Sprintf(`你是成都文旅内容编辑。请为景点"%s"撰写图文详情,严格只输出 JSON,不要输出解释:
{"sections":[{"title":"小标题(6-12字)","text":"段落(80-150字)"}]}
参考简介:%s
要求:写成 %s;只使用参考简介里能确认的信息;不确定的内容不要写;不要编造数字、价格、年份;不要使用 Markdown 语法;使用简体中文。`,
		s.NameZH, s.Desc, sectionRange)
	reply, err := e.llm.ChatWithEffort(ctx, 90, []llmMessage{
		{Role: "system", Content: "你是严谨的文旅内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.3, 1200, llmThinkingLevel())
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Sections []model.ScenicDetailSection `json:"sections"`
	}
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return buildDetailSections(parsed.Sections, images), nil
}

// fillEstimated 让 LLM 给出无权威公开源的字段参考值,未知填空串。
func (e *ScenicEnricher) fillEstimated(ctx context.Context, s model.ScenicSpot) ([]estimatedKV, error) {
	prompt := fmt.Sprintf(`请根据下列已知信息,给出成都景点"%s"的参考信息。严格只输出 JSON,不要输出解释:
{"ticket_price":"","open_hours":"","recommend_hours":"","yearly_visitors":""}
已知信息:
- 所属区县:%s
- 简介:%s
- 高德开放时间:%s
规则:
1. 这些值会作为"参考值"展示,必须保守,宁缺勿错;
2. ticket_price 写具体票制,如"免费""成人60元/人";不确定写空字符串;
3. open_hours 仅在已知信息里没有开放时间时才填;有则留空字符串;
4. recommend_hours 写"2-3小时"这类区间;
5. yearly_visitors 写量级概数,如"约数十万人次""约百万人次""约500万人次",必须带"约"字且不要写到个位;
   只能判断数量级,无法判断量级时再写空字符串;
6. 不要编造精确数字、具体年份、官方排名;ticket_price 仍保持保守,不确定就留空。`,
		s.NameZH, s.District, s.Desc, s.OpenHours)
	reply, err := e.llm.ChatWithEffort(ctx, 60, []llmMessage{
		{Role: "system", Content: "你是谨慎的文旅信息编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.2, 600, llmThinkingLevel())
	if err != nil {
		return nil, err
	}
	var parsed struct {
		TicketPrice    string `json:"ticket_price"`
		OpenHours      string `json:"open_hours"`
		RecommendHours string `json:"recommend_hours"`
		YearlyVisitors string `json:"yearly_visitors"`
	}
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return []estimatedKV{
		{Field: "ticket_price", Value: parsed.TicketPrice},
		{Field: "open_hours", Value: parsed.OpenHours},
		{Field: "recommend_hours", Value: parsed.RecommendHours},
		{Field: "yearly_visitors", Value: parsed.YearlyVisitors},
	}, nil
}

// buildDetailSections 把 LLM 段落与相册图配对:第 1 张图是封面,段落从第 2 张开始配;
// 图片不足的段落复用第 2 张图兜底(只有 1 张时用第 1 张),保证每段都有图;
// images 为空时维持空图,交由前端占位。
func buildDetailSections(llmSections []model.ScenicDetailSection, images []string) []model.ScenicDetailSection {
	out := make([]model.ScenicDetailSection, 0, len(llmSections))
	for i, sec := range llmSections {
		title := strings.TrimSpace(sec.Title)
		text := strings.TrimSpace(sec.Text)
		if title == "" || text == "" {
			continue
		}
		img := ""
		if len(images) > 0 {
			img = images[0]
			if len(images) > 1 {
				img = images[1]
			}
			if idx := i + 1; idx < len(images) {
				img = images[idx]
			}
		}
		out = append(out, model.ScenicDetailSection{Title: title, Text: text, Image: img})
	}
	return out
}

// appendEstimatedField 以逗号分隔追加参考值字段名,已存在则不重复。
func appendEstimatedField(existing, field string) string {
	if field == "" {
		return existing
	}
	for _, f := range splitComma(existing) {
		if f == field {
			return existing
		}
	}
	if strings.TrimSpace(existing) == "" {
		return field
	}
	return existing + "," + field
}

// survivingEstimated 只把本次真正写入库的 LLM 字段登记为参考值,
// 避免被非 Force 保护删除的字段仍被前端打上"参考值"徽标。
func survivingEstimated(existing string, llmFields []string, updates map[string]any) string {
	out := existing
	for _, f := range llmFields {
		if _, ok := updates[f]; !ok {
			continue
		}
		out = appendEstimatedField(out, f)
	}
	return out
}

// filterEstimated 过滤掉本次已由高德等事实来源写入的字段,保证事实值优先、参考值不覆盖事实。
// 返回新切片,不修改入参。
func filterEstimated(est []estimatedKV, taken map[string]any) []estimatedKV {
	out := make([]estimatedKV, 0, len(est))
	for _, kv := range est {
		if _, ok := taken[kv.Field]; ok {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// hasBusinessUpdate 判断本次是否真的有业务字段变更(排除 estimated_fields/data_source/data_updated_at 三个溯源键)。
func hasBusinessUpdate(updates map[string]any) bool {
	for key := range updates {
		switch key {
		case "estimated_fields", "data_source", "data_updated_at":
			// 溯源键,不算业务变更
		default:
			return true
		}
	}
	return false
}

// buildDataSource 拼装数据来源标记。
func buildDataSource(poiMatched, llmUsed bool) string {
	switch {
	case poiMatched && llmUsed:
		return "amap+llm"
	case poiMatched:
		return "amap"
	case llmUsed:
		return "llm"
	default:
		return "wiki"
	}
}

// isPlaceholderCover 判断库中封面是否为前端占位路径(如 /images/placeholder-xxx.jpg)或空值,
// 这类值应视为"没有封面",可用相册首图替代。
func isPlaceholderCover(cover string) bool {
	v := strings.TrimSpace(cover)
	return v == "" || strings.HasPrefix(v, "/images/placeholder")
}

// isEmptyField 判断库中该字段是否为空(用于非 Force 模式跳过覆盖)。
func isEmptyField(s model.ScenicSpot, field string) bool {
	switch field {
	case "address":
		return strings.TrimSpace(s.Address) == ""
	case "tel":
		return strings.TrimSpace(s.Tel) == ""
	case "open_hours":
		return strings.TrimSpace(s.OpenHours) == ""
	case "ticket_price":
		return strings.TrimSpace(s.TicketPrice) == ""
	case "recommend_hours":
		return strings.TrimSpace(s.RecommendHours) == ""
	case "yearly_visitors":
		return strings.TrimSpace(s.YearlyVisitors) == ""
	case "gallery_images":
		return strings.TrimSpace(s.GalleryImages) == ""
	case "detail_sections":
		return strings.TrimSpace(s.DetailSections) == ""
	case "images":
		// 占位封面与空值一样视为"没有封面",否则非 Force 保护会把 step2 写入的相册首图删掉。
		return isPlaceholderCover(s.Images)
	case "amap_poi_id":
		return strings.TrimSpace(s.AmapPOIID) == ""
	default:
		return true
	}
}

// distanceMeters 用 Haversine 公式计算两点直线距离(米)。
// 经度/lat 顺序与高德一致:先经度后纬度。
func distanceMeters(lng1, lat1, lng2, lat2 float64) float64 {
	if lng1 == lng2 && lat1 == lat2 {
		return 0
	}
	const earthRadius = 6371000.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadius * math.Asin(math.Sqrt(a))
}

// splitComma 拆分逗号分隔串并去空。
func splitComma(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// extFromContentType 按响应类型推断图片后缀,推断不出时回退 URL 后缀,再回退 .jpg。
func extFromContentType(contentType, rawURL string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	}
	if ext := strings.ToLower(filepath.Ext(rawURL)); ext != "" && len(ext) <= 6 {
		return ext
	}
	return ".jpg"
}
