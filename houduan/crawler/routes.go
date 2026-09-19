package main

// 精选路线爬取:路线页(/home/routes)数据采集。
// 对每条路线的途经站点抓取维基百科简介与配图(需代理),图片上传 OSS,
// 组装站点 JSON 后按 route_key 写入 routes 表(存在则更新,不存在则插入)。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./crawler -routes -proxy http://127.0.0.1:7897 -upload

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"wenlv-backend/model"
	"wenlv-backend/pkg"
)

// routeStopSeed 路线站点种子:Key 用于 OSS 对象名与站点排序。
// Wiki 为维基百科搜索词(缺省用 NameZH);Desc 为无词条时的回退简介。
type routeStopSeed struct {
	Key    string
	NameZH string
	NameEN string
	NameJA string
	Wiki   string
	Desc   string
}

// routeSeed 精选路线种子:Interests 取值需与前端 AI 行程兴趣项一致。
type routeSeed struct {
	Key         string
	Sort        int
	TitleZH     string
	TitleEN     string
	TitleJA     string
	Theme       string
	Description string
	Days        int
	Interests   []string
	Stops       []routeStopSeed
}

// routeSeeds 四条精选路线种子数据(与前端 RoutesPage 主题一致)。
func routeSeeds() []routeSeed {
	return []routeSeed{
		{
			Key: "classic", Sort: 1,
			TitleZH: "经典成都一日游", TitleEN: "Classic Chengdu Day Tour", TitleJA: "成都クラシック一日観光",
			Theme: "城市经典", Days: 1,
			Description: "祠堂庙宇与街巷烟火一次打包：上午拜访武侯祠与锦里，下午穿行宽窄巷子，在人民公园喝盖碗茶，晚上到春熙路看霓虹不夜城。",
			Interests:   []string{"历史文化", "美食"},
			Stops: []routeStopSeed{
				{Key: "wuhouci", NameZH: "武侯祠", NameEN: "Wuhou Shrine", NameJA: "武侯祠", Wiki: "武侯祠"},
				{Key: "jinli", NameZH: "锦里古街", NameEN: "Jinli Ancient Street", NameJA: "錦里古街", Wiki: "锦里"},
				{Key: "kuanzhai", NameZH: "宽窄巷子", NameEN: "Kuanzhai Alley", NameJA: "寛窄巷子", Wiki: "宽窄巷子"},
				{Key: "renmin-park", NameZH: "人民公园", NameEN: "People's Park", NameJA: "人民公園", Wiki: "人民公园 (成都)"},
				{Key: "chunxi", NameZH: "春熙路·太古里", NameEN: "Chunxi Road & Taikoo Li", NameJA: "春熙路・太古里", Wiki: "春熙路"},
			},
		},
		{
			Key: "panda", Sort: 2,
			TitleZH: "熊猫亲子二日游", TitleEN: "Panda Family 2-Day Trip", TitleJA: "パンダファミリー二日間",
			Theme: "亲子自然", Days: 2,
			Description: "第一天蹲守大熊猫基地的干饭名场面，逛文殊院、东郊记忆；第二天前往都江堰感受千年水利智慧，夜游南桥看灯光水景。",
			Interests:   []string{"自然风光", "休闲"},
			Stops: []routeStopSeed{
				{Key: "panda-base", NameZH: "大熊猫繁育研究基地", NameEN: "Chengdu Panda Base", NameJA: "パンダ繁殖研究基地", Wiki: "成都大熊猫繁育研究基地"},
				{Key: "wenshu", NameZH: "文殊院", NameEN: "Wenshu Monastery", NameJA: "文殊院", Wiki: "文殊院"},
				{Key: "dongjiao", NameZH: "东郊记忆", NameEN: "Eastern Suburb Memory", NameJA: "東郊記憶", Wiki: "东郊记忆"},
				{Key: "dujiangyan", NameZH: "都江堰", NameEN: "Dujiangyan Irrigation System", NameJA: "都江堰", Wiki: "都江堰"},
				{Key: "nanqiao", NameZH: "南桥夜景", NameEN: "Nanqiao Bridge Night View", NameJA: "南橋の夜景", Wiki: "南桥 (都江堰)"},
			},
		},
		{
			Key: "food", Sort: 3,
			TitleZH: "美食夜宵一日线", TitleEN: "Foodie Night Crawl", TitleJA: "グルメ夜食一行",
			Theme: "美食市井", Days: 1,
			Description: "从建设路小吃街的蛋烘糕开始，到玉林路的小酒馆坐坐，钻进香香巷来一顿麻辣盛宴，最后在九眼桥的夜色里收尾。",
			Interests:   []string{"美食", "休闲"},
			Stops: []routeStopSeed{
				{Key: "jianshe-road", NameZH: "建设路小吃街", NameEN: "Chengdu street food", NameJA: "建設路グルメ街",
					Desc: "成华区人气最旺的小吃街，蛋烘糕、烤脑花、铁板鱿鱼一路吃过去，是本地学生的深夜食堂。"},
				{Key: "yulin", NameZH: "玉林路小酒馆", NameEN: "Yulin Road bar Chengdu", NameJA: "玉林路の酒場",
					Desc: "民谣《成都》唱到的玉林路，小酒馆与苍蝇馆子林立，是最有烟火气的老成都街区。"},
				{Key: "xiangxiang", NameZH: "望平街·香香巷", NameEN: "Chengdu food alley", NameJA: "望平街・香香巷",
					Desc: "滨河小巷藏着数十家苍蝇馆子，麻辣烫、把把烧、泰式火锅一巷吃遍全城。"},
				{Key: "jiuyanqiao", NameZH: "九眼桥酒吧街", NameEN: "Jiuyanqiao bridge Chengdu", NameJA: "九眼橋バーストリート", Wiki: "安顺廊桥",
					Desc: "锦江边的酒吧一条街，夜景璀璨，是成都夜生活的地标。"},
			},
		},
		{
			Key: "culture", Sort: 4,
			TitleZH: "文化深度三日游", TitleEN: "Culture Deep-Dive 3 Days", TitleJA: "文化深掘り三日間",
			Theme: "人文历史", Days: 3,
			Description: "探秘三星堆的古蜀文明，走访杜甫草堂与四川博物院，登青城山问道，夜里泛舟锦江，慢读这座城两千年的人文底蕴。",
			Interests:   []string{"历史文化", "艺术"},
			Stops: []routeStopSeed{
				{Key: "sanxingdui", NameZH: "三星堆博物馆", NameEN: "Sanxingdui Museum", NameJA: "三星堆博物館", Wiki: "三星堆博物馆"},
				{Key: "dufu", NameZH: "杜甫草堂", NameEN: "Du Fu Thatched Cottage", NameJA: "杜甫草堂", Wiki: "杜甫草堂"},
				{Key: "sichuan-museum", NameZH: "四川博物院", NameEN: "Sichuan Museum", NameJA: "四川博物院", Wiki: "四川博物院"},
				{Key: "qingcheng", NameZH: "青城山", NameEN: "Mount Qingcheng", NameJA: "青城山", Wiki: "青城山"},
				{Key: "jinjiang", NameZH: "夜游锦江", NameEN: "Jinjiang river Chengdu", NameJA: "錦江ナイトクルーズ", Wiki: "锦江 (岷江支流)",
					Desc: "夜幕下的锦江灯光水秀，泛舟东门码头，两岸光影重现千年锦官城的繁华。"},
			},
		},
	}
}

// routeStopResult 单个站点的抓取结果。
type routeStopResult struct {
	NameZH string `json:"name_zh"`
	NameEN string `json:"name_en"`
	NameJA string `json:"name_ja"`
	Desc   string `json:"desc"`
	Image  string `json:"image"`
}

// runRoutesCrawl 爬取各路线站点简介与配图,上传 OSS 并写入 routes 表。
func runRoutesCrawl(client *http.Client, db *gorm.DB, signer *pkg.OssSigner, outDir string, delayMs int) {
	seeds := routeSeeds()
	if signer == nil {
		log.Printf("开始爬取精选路线(%d 条),图片保存目录: %s(未配置 -upload,仅本地留档)", len(seeds), outDir)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			log.Fatalf("创建图片目录失败: %v", err)
		}
	} else {
		log.Printf("开始爬取精选路线(%d 条),站点图片直接上传 OSS", len(seeds))
	}

	okCnt, failCnt := 0, 0
	for _, r := range seeds {
		log.Printf("── 路线 [%s] %s(%d 个站点)", r.Key, r.TitleZH, len(r.Stops))

		// 已入库站点:OSS 图片跳过重传,失败站点仅补图(避免反复重爬触发限流)
		existingStops := loadRouteStops(db, r.Key)
		existingByName := map[string]routeStopResult{}
		for _, es := range existingStops {
			existingByName[es.NameZH] = es
		}

		var stops []routeStopResult
		cover := ""
		for i, st := range r.Stops {
			log.Printf("  [%d/%d] %s", i+1, len(r.Stops), st.NameZH)
			res := routeStopResult{NameZH: st.NameZH, NameEN: st.NameEN, NameJA: st.NameJA, Desc: st.Desc}
			prev, hasPrev := existingByName[st.NameZH]

			// 0. 已有 OSS 配图的历史站点:整站复用,不再访问维基百科/OSS
			if hasPrev && strings.Contains(prev.Image, "aliyuncs.com") {
				log.Printf("      已有 OSS 配图,跳过")
				if prev.Desc != "" {
					res.Desc = prev.Desc
				}
				res.Image = prev.Image
				stops = append(stops, res)
				if cover == "" {
					cover = res.Image
				}
				continue
			}

			// 1. 维基百科简介(无 Wiki 词或未匹配时保留种子简介)
			searchName := st.Wiki
			if searchName == "" {
				searchName = st.NameZH
			}
			titles, err := searchWiki(client, searchName)
			if err != nil {
				log.Printf("      搜索失败: %v", err)
			}
			for _, t := range titles {
				extract, err := fetchIntro(client, t)
				if err != nil {
					log.Printf("      正文获取失败: %v", err)
					break
				}
				if extract == "" {
					sum, err := fetchSummary(client, t)
					if err != nil || sum.Type == "disambiguation" {
						continue
					}
					extract = strings.TrimSpace(sum.Extract)
				}
				if !mentionsChengdu(extract) {
					log.Printf("      候选 [%s] 与成都/四川无关,跳过", t)
					continue
				}
				res.Desc = extract
				// 2. 配图:词条原图 → 缩略图 → Commons 图库
				sum, _ := fetchSummary(client, t)
				if sum.OriginalImage != nil && sum.OriginalImage.Source != "" {
					res.Image = sum.OriginalImage.Source
				} else if sum.Thumbnail != nil {
					res.Image = sum.Thumbnail.Source
				}
				break
			}
			if res.Image == "" {
				if img := commonsImage(client, st.NameZH, st.NameEN); img != "" {
					res.Image = img
					log.Printf("      Commons 图库补图")
				}
			}

			// 3. 图片处理:配置 OSS 时先探测对象是否已存在(此前轮次可能已传成功),
			//    存在直接复用 URL,不存在才上传(带重试);失败时保留原 OSS 图或置空,
			//    绝不把维基外链写库(大陆无法访问)
			if res.Image != "" {
				ext := imageExt(res.Image)
				key := "routes/" + r.Key + "/" + st.Key + ext
				if signer != nil {
					ossURL := signer.ResolveURL(key)
					if ossObjectExists(ossURL) {
						log.Printf("      OSS 已有同名对象,直接复用: %s", ossURL)
						res.Image = ossURL
					} else {
						uploaded, err := uploadImageToOSSWithRetry(client, signer, res.Image, key)
						if err != nil {
							log.Printf("      OSS 上传失败(已重试): %v", err)
							if hasPrev && strings.Contains(prev.Image, "aliyuncs.com") {
								res.Image = prev.Image
								log.Printf("      保留原有 OSS 配图")
							} else {
								res.Image = ""
							}
						} else {
							log.Printf("      已上传 OSS: %s", uploaded)
							res.Image = uploaded
						}
					}
				} else {
					local := filepath.Join(outDir, r.Key+"-"+st.Key+ext)
					if err := downloadImage(client, res.Image, local); err != nil {
						log.Printf("      图片下载失败: %v", err)
					} else {
						log.Printf("      图片已保存: %s", local)
					}
				}
			} else {
				log.Printf("      未获取到配图,保留空")
			}
			if cover == "" && strings.Contains(res.Image, "aliyuncs.com") {
				cover = res.Image // 封面取第一张已上传 OSS 的站点图
			}
			stops = append(stops, res)
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// 4. 组装站点 JSON 并入库
		stopsJSON, err := json.Marshal(stops)
		if err != nil {
			log.Printf("  站点 JSON 序列化失败: %v", err)
			failCnt++
			continue
		}
		if err := upsertRoute(db, r, string(stopsJSON), cover); err != nil {
			log.Printf("  入库失败: %v", err)
			failCnt++
			continue
		}
		log.Printf("  已入库(封面: %s)", truncate(cover, 60))
		okCnt++
	}
	log.Printf("路线爬取完成: 成功 %d / 失败 %d", okCnt, failCnt)
}

// upsertRoute 按 route_key 写入 routes 表;已有记录仅在拿到新内容时覆盖。
func upsertRoute(db *gorm.DB, r routeSeed, stopsJSON, cover string) error {
	var existing model.Route
	err := db.Where("route_key = ?", r.Key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(&model.Route{
			RouteKey:    r.Key,
			TitleZH:     r.TitleZH,
			TitleEN:     r.TitleEN,
			TitleJA:     r.TitleJA,
			Theme:       r.Theme,
			Description: r.Description,
			Days:        r.Days,
			Interests:   strings.Join(r.Interests, ","),
			CoverImage:  cover,
			Stops:       stopsJSON,
			Sort:        r.Sort,
		}).Error
	}
	if err != nil {
		return err
	}
	updates := map[string]any{
		"title_zh":    r.TitleZH,
		"title_en":    r.TitleEN,
		"title_ja":    r.TitleJA,
		"theme":       r.Theme,
		"description": r.Description,
		"days":        r.Days,
		"interests":   strings.Join(r.Interests, ","),
		"sort":        r.Sort,
		"stops":       stopsJSON,
	}
	// 仅在拿到 OSS 封面,或原封面为空时覆盖
	if cover != "" {
		updates["cover_image"] = cover
	}
	return db.Model(&existing).Updates(updates).Error
}

// loadRouteStops 读取路线已入库的站点列表(反序列化失败视为无历史数据)。
func loadRouteStops(db *gorm.DB, routeKey string) []routeStopResult {
	var row model.Route
	if err := db.Select("stops").Where("route_key = ?", routeKey).First(&row).Error; err != nil {
		return nil
	}
	var stops []routeStopResult
	if err := json.Unmarshal([]byte(row.Stops), &stops); err != nil {
		return nil
	}
	return stops
}

// ossObjectExists 匿名 HEAD 探测 OSS 公读对象是否已存在(公开读桶无需签名)。
func ossObjectExists(ossURL string) bool {
	resp, err := http.Head(ossURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// uploadImageToOSSWithRetry 带指数退避的 OSS 上传:2s/5s/10s 三次重试,
// 缓解连续上传触发的桶级 429 限流。
func uploadImageToOSSWithRetry(client *http.Client, signer *pkg.OssSigner, imageURL, key string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := time.Duration(2<<uint(attempt-1)) * time.Second // 2s/4s→5s 近似退避
			log.Printf("      OSS 限流,第 %d 次重试(等待 %v)", attempt, wait)
			time.Sleep(wait)
		}
		ossURL, err := uploadImageToOSS(client, signer, imageURL, key)
		if err == nil {
			return ossURL, nil
		}
		lastErr = err
		if !strings.Contains(err.Error(), "429") {
			return "", err // 非限流错误不重试
		}
	}
	return "", lastErr
}
