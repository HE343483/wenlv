// 美食详情数据采集与补全:高德门店事实字段 + LLM 参考值 + 风味故事分段 + 多来源图片。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
)

// foodGalleryWant 单个美食最多落库的相册张数(第 1 张作封面)。
const foodGalleryWant = 6

// FoodEnrichOptions 采集开关。
type FoodEnrichOptions struct {
	WithImages bool
	WithLLM    bool
	Force      bool
}

// FoodEnrichResult 单个美食的采集结果。
type FoodEnrichResult struct {
	FoodID        uint
	Name          string
	POIMatched    bool
	Uploaded      int
	LLMUsed       bool
	UpdatedFields []string
	Note          string
}

// FoodEnricher 美食数据采集器。
type FoodEnricher struct {
	repo   *repository.FoodRepo
	amap   *AmapService
	llm    *TripLLM
	signer *pkg.OssSigner
}

// NewFoodEnricher 构造采集器。
func NewFoodEnricher(repo *repository.FoodRepo, amap *AmapService, llm *TripLLM, signer *pkg.OssSigner) *FoodEnricher {
	return &FoodEnricher{repo: repo, amap: amap, llm: llm, signer: signer}
}

// foodIsUsablePOI 只接受餐饮服务类 POI(实测搜「麻婆豆腐」会返回超市等无关类型)。
func foodIsUsablePOI(poiType string) bool {
	return strings.Contains(poiType, "餐饮服务")
}

// matchFoodPOI 在候选门店中挑与菜名同名的门店:完全相同(3 分) > 名称包含菜名(2 分),
// 同分取评分高者。一个都不满足时返回未匹配 —— 宁可不写门店事实,也不写无关餐厅。
func matchFoodPOI(candidates []model.POIInfo, nameZH string) (model.POIInfo, bool) {
	best := model.POIInfo{}
	bestScore, bestRating := 0, -1.0
	for _, c := range candidates {
		if !foodIsUsablePOI(c.Type) {
			continue
		}
		score := 0
		if c.Name == nameZH {
			score = 3
		} else if strings.Contains(c.Name, nameZH) {
			score = 2
		}
		if score == 0 {
			continue
		}
		rating := 0.0
		if score > bestScore || (score == bestScore && rating > bestRating) {
			best, bestScore, bestRating = c, score, rating
		}
	}
	if bestScore == 0 {
		return model.POIInfo{}, false
	}
	return best, true
}

// isEmptyFoodField 判断库中该字段是否为空(占位封面同样视为空,便于用真实图片替换)。
func isEmptyFoodField(f model.Food, field string) bool {
	switch field {
	case "rating":
		return strings.TrimSpace(f.Rating) == ""
	case "flavor":
		return strings.TrimSpace(f.Flavor) == ""
	case "spice_level":
		return strings.TrimSpace(f.SpiceLevel) == ""
	case "avg_price":
		return strings.TrimSpace(f.AvgPrice) == ""
	case "signature":
		return strings.TrimSpace(f.Signature) == ""
	case "recommend_scene":
		return strings.TrimSpace(f.RecommendScene) == ""
	case "story_sections":
		return strings.TrimSpace(f.StorySections) == ""
	case "gallery_images":
		return strings.TrimSpace(f.GalleryImages) == ""
	case "images":
		return isPlaceholderCover(f.Images)
	case "poi_name":
		return strings.TrimSpace(f.POIName) == ""
	case "address":
		return strings.TrimSpace(f.Address) == ""
	default:
		return true
	}
}

// MatchPOI 用与 Enrich 完全相同的判定逻辑匹配门店,供 dry-run 预演复用,
// 保证「未匹配清单」与正式跑结论一致。
func (e *FoodEnricher) MatchPOI(ctx context.Context, f model.Food) (model.POIInfo, bool) {
	candidates := e.amap.SearchPOI(ctx, f.NameZH, "成都", true)
	return matchFoodPOI(candidates, f.NameZH)
}

// Enrich 采集单个美食并落库。任一步失败只记录 Note 并继续,不中断批处理。
func (e *FoodEnricher) Enrich(ctx context.Context, f model.Food, opts FoodEnrichOptions) (*FoodEnrichResult, error) {
	res := &FoodEnrichResult{FoodID: f.ID, Name: f.NameZH}
	updates := map[string]any{}
	estimated := f.EstimatedFields

	// 1) 高德门店匹配(名称包含菜名 + 餐饮服务类型)
	var photos []string
	if candidates := e.amap.SearchPOI(ctx, f.NameZH, "成都", true); len(candidates) > 0 {
		if poi, ok := matchFoodPOI(candidates, f.NameZH); ok {
			if detail, err := e.amap.GetPOIDetail(ctx, poi.ID); err != nil {
				res.Note = "门店详情获取失败: " + err.Error()
			} else {
				res.POIMatched = true
				updates["poi_name"] = detail.Name
				if v := strings.TrimSpace(detail.Address); v != "" {
					updates["address"] = v
				}
				if detail.Location.Latitude != 0 && detail.Location.Longitude != 0 {
					updates["lat"] = detail.Location.Latitude
					updates["lng"] = detail.Location.Longitude
				}
				// 高德评分/人均是事实值,不进 estimated_fields
				if v := strings.TrimSpace(detail.Rating); v != "" {
					updates["rating"] = v
				}
				if v := strings.TrimSpace(detail.Cost); v != "" && v != "0" {
					updates["avg_price"] = "人均 ¥" + v
				}
				photos = detail.Photos
				res.Note = "匹配门店: " + detail.Name
			}
		} else {
			res.Note = "无同名门店(该菜名无对应老字号),跳过门店事实"
		}
	} else {
		res.Note = "未搜到候选门店"
	}

	// 2) 图片:高德门店相册 → Wikimedia Commons 兜底,全部落 OSS
	images := splitComma(f.GalleryImages)
	if opts.WithImages && (opts.Force || len(images) == 0) && e.signer != nil && e.signer.Configured() {
		candidates := make([]ScenicImageCandidate, 0, foodGalleryWant)
		for _, u := range photos {
			candidates = append(candidates, ScenicImageCandidate{URL: u, Source: "amap"})
		}
		if len(candidates) < foodGalleryWant {
			names := []string{f.NameZH}
			if strings.TrimSpace(f.NameEN) != "" {
				names = append(names, f.NameEN)
			}
			if found, err := searchCommonsImagesFor(ctx, newCommonsHTTPClient(), f.NameZH, names, foodGalleryWant-len(candidates)); err == nil {
				for _, it := range found {
					candidates = append(candidates, ScenicImageCandidate{URL: it.URL, Source: "commons"})
				}
			} else {
				logger.Warnf("Commons 取图失败: %v", err)
			}
		}
		if urls, n := uploadImageCandidates(ctx, newCommonsHTTPClient(), e.signer, fmt.Sprintf("food/%d", f.ID), candidates, foodGalleryWant); n > 0 {
			images = urls
			res.Uploaded = n
			updates["gallery_images"] = strings.Join(urls, ",")
			logger.Infof("图片: 上传 %d 张", n)
		}
	}
	if isPlaceholderCover(f.Images) && len(images) > 0 {
		updates["images"] = images[0]
	}

	// 3) LLM:风味故事 + 参考值字段
	if opts.WithLLM && e.llm != nil && e.llm.Available() {
		maxSections := clampMaxSections(len(images))
		if sections, err := e.buildStory(ctx, f, images, maxSections); err != nil {
			res.Note += "; 风味故事生成失败: " + err.Error()
		} else if len(sections) > 0 {
			if raw, err := json.Marshal(sections); err == nil {
				updates["story_sections"] = string(raw)
				res.LLMUsed = true
			}
		}
		est, err := e.fillFoodEstimated(ctx, f)
		if err != nil {
			res.Note += "; 参考值生成失败: " + err.Error()
		} else {
			filled := 0
			for _, kv := range est {
				if v := strings.TrimSpace(kv.Value); v != "" {
					updates[kv.Field] = v
					filled++
				}
			}
			if filled > 0 {
				res.LLMUsed = true
			}
		}
	}

	// 4) 非 Force 模式保护已有值
	if !opts.Force {
		for key := range updates {
			if !isEmptyFoodField(f, key) {
				delete(updates, key)
			}
		}
	}
	// 5) 参考值登记:只登记本次真正写入的 LLM 字段
	llmFields := []string{}
	for _, kv := range []string{"flavor", "spice_level", "signature", "recommend_scene"} {
		if _, ok := updates[kv]; ok {
			llmFields = append(llmFields, kv)
		}
	}
	// rating / avg_price 仅在高德未提供(即这两个键由 LLM 写入)时才登记
	for _, kv := range []string{"rating", "avg_price"} {
		if _, ok := updates[kv]; ok && !res.POIMatched {
			llmFields = append(llmFields, kv)
		} else if _, ok := updates[kv]; ok && res.POIMatched && isEmptyFoodField(f, kv) {
			// 门店匹配成功但仍补了值:说明是高德给的,不登记
			continue
		}
	}
	estimated = survivingEstimated(estimated, llmFields, updates)
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
	if err := e.repo.UpdateFields(f.ID, updates); err != nil {
		return res, err
	}
	return res, nil
}

// buildStory 让 LLM 基于已有简介写出 2..maxSections 段风味故事,并按顺序配相册图。
func (e *FoodEnricher) buildStory(ctx context.Context, f model.Food, images []string, maxSections int) ([]model.FoodStorySection, error) {
	if strings.TrimSpace(f.Desc) == "" {
		return nil, nil
	}
	prompt := fmt.Sprintf(`你是成都美食内容编辑。请为"%s"撰写风味故事,严格只输出 JSON,不要输出解释:
{"sections":[{"title":"小标题(4-10字)","text":"段落(60-120字)"}]}
参考简介:%s
要求:写成 2-%d 段;只使用参考简介里能确认的信息;不确定的内容不要写;不要编造数字、价格、年份;不要使用 Markdown 语法;使用简体中文。`,
		f.NameZH, f.Desc, maxSections)
	reply, err := e.llm.ChatWithEffort(ctx, 90, []llmMessage{
		{Role: "system", Content: "你是严谨的美食内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.3, 1200, llmThinkingLevel())
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Sections []model.FoodStorySection `json:"sections"`
	}
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return buildFoodStorySections(parsed.Sections, images), nil
}

// fillFoodEstimated 让 LLM 给出无权威源的字段参考值,未知填空串。
func (e *FoodEnricher) fillFoodEstimated(ctx context.Context, f model.Food) ([]estimatedKV, error) {
	prompt := fmt.Sprintf(`请为成都美食"%s"给出下列参考信息。严格只输出 JSON,不要输出解释:
{"flavor":"","spice_level":"","signature":"","recommend_scene":"","avg_price":"","rating":""}
已知信息:
- 标签:%s
- 简介:%s
规则:
1. 这些值会作为"参考值"展示,必须保守,宁缺勿错;
2. flavor 写 4-6 字风味概括,如"麻辣鲜香";
3. spice_level 从"不辣/微辣/中辣/重辣"中选一个;
4. signature 写 2-3 个代表性吃法或搭配,用"·"分隔;
5. recommend_scene 写 2 个场景,用"/"分隔,如"朋友聚餐 / 夜宵";
6. avg_price 写"人均 ¥数字"区间,不确定写空字符串;rating 写 4.0-5.0 的一位小数,不确定写空字符串;
7. 不要编造品牌名、具体门店、年份。`, f.NameZH, f.Tags, f.Desc)
	reply, err := e.llm.ChatWithEffort(ctx, 60, []llmMessage{
		{Role: "system", Content: "你是谨慎的美食信息编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.2, 800, llmThinkingLevel())
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Flavor         string `json:"flavor"`
		SpiceLevel     string `json:"spice_level"`
		Signature      string `json:"signature"`
		RecommendScene string `json:"recommend_scene"`
		AvgPrice       string `json:"avg_price"`
		Rating         string `json:"rating"`
	}
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return []estimatedKV{
		{Field: "flavor", Value: parsed.Flavor},
		{Field: "spice_level", Value: parsed.SpiceLevel},
		{Field: "signature", Value: parsed.Signature},
		{Field: "recommend_scene", Value: parsed.RecommendScene},
		{Field: "avg_price", Value: parsed.AvgPrice},
		{Field: "rating", Value: parsed.Rating},
	}, nil
}

// buildFoodStorySections 段落与图片配对:第 1 张图作封面,段落从第 2 张开始;
// 图片不足时复用第 2 张(没有则第 1 张),保证每段都有图。
func buildFoodStorySections(llmSections []model.FoodStorySection, images []string) []model.FoodStorySection {
	out := make([]model.FoodStorySection, 0, len(llmSections))
	for i, sec := range llmSections {
		title := strings.TrimSpace(sec.Title)
		text := strings.TrimSpace(sec.Text)
		if title == "" || text == "" {
			continue
		}
		img := ""
		if idx := i + 1; idx < len(images) {
			img = images[idx]
		} else if len(images) > 1 {
			img = images[1]
		} else if len(images) == 1 {
			img = images[0]
		}
		out = append(out, model.FoodStorySection{Title: title, Text: text, Image: img})
	}
	return out
}
