// route-theme-seed 蜀文化叙事路线种子程序:写入 3 条以蜀文化故事串联景点的
// 文化主题路线(theme=culture),区别于现有 4 条玩法型路线(theme 为中文分类词)。
// 站点取自 scenic_spots 库内景点(按 name_zh 精确匹配,候选中不存在的自动剔除),
// 站点 desc 摘自景点中文介绍(截断 200 字,供 story-enrich 生成多语种故事),
// 站点 image 取景点库内第一张 OSS 图。
//
// 幂等:按 route_key(其次 title_zh)查找,已存在则更新 theme/简介/站点等字段,
// 不存在则插入;重复执行结果一致。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/route-theme-seed
package main

import (
	"encoding/json"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/logger"
	"wenlv-backend/model"
)

// cultureRouteSeed 一条文化主题路线的种子定义。
// Stops 为有序候选站点中文名(必须能在 scenic_spots 表中按 name_zh 匹配到,否则剔除)。
type cultureRouteSeed struct {
	RouteKey    string
	Sort        int
	TitleZH     string
	TitleEN     string // 标题翻译人工定稿(专名不走 LLM,保证稳定)
	TitleJA     string
	Description string
	Interests   []string
	Stops       []string
}

// cultureRouteSeeds 三条蜀文化叙事路线种子(站点均经库内核实存在)。
func cultureRouteSeeds() []cultureRouteSeed {
	return []cultureRouteSeed{
		{
			RouteKey: "sanguo", Sort: 5,
			TitleZH: "三国寻踪",
			TitleEN: "In the Footsteps of the Three Kingdoms",
			TitleJA: "三国志の足跡をたどる",
			Description: "一千八百年前，蜀汉丞相诸葛亮在这里鞠躬尽瘁；如今，君臣合祀的武侯祠仍静卧于红墙竹影之间。这条路线带你从一个祠堂读懂三国：细品“攻心联”的千年治蜀智慧，出了朱门便是锦里的市井烟火——飞檐斗拱下，三大炮的脆响与皮影戏的锣鼓声交织。历史与烟火气只隔一道门，这正是成都最奇妙的地方。跟随丞相的足迹，在红墙翠竹间赴一场穿越千年的君臣之约。",
			Interests: []string{"历史文化"},
			Stops:     []string{"武侯祠", "锦里古街"},
		},
		{
			RouteKey: "poetry", Sort: 6,
			TitleZH: "诗歌之路",
			TitleEN: "The Poetry Trail: Following Du Fu",
			TitleJA: "詩歌の道 — 杜甫と歩く成都",
			Description: "“窗含西岭千秋雪，门泊东吴万里船。”公元七六〇年，杜甫在浣花溪畔筑起一间茅屋，成都的温柔从此有了注脚。这条路线带你跟随杜甫的足迹：清晨在草堂的竹影里读一句“好雨知时节”，午后向北远眺，西岭雪山的万古积雪正应了诗中的千古名句。一日之间，从一座茅屋读懂成都的温柔，再登临雪山之巅，亲眼验证诗人笔下的春秋——诗里的风景，至今仍在。",
			Interests: []string{"历史文化", "自然风光"},
			Stops:     []string{"杜甫草堂", "西岭雪山"},
		},
		{
			RouteKey: "gushu", Sort: 7,
			TitleZH: "古蜀探源",
			TitleEN: "Tracing Ancient Shu",
			TitleJA: "古蜀の源をたずねて",
			Description: "成都的古，比三国更久远。传说老子骑青牛西来，在青羊肆为关令尹喜讲经，青羊宫因此得名，两千年香火不绝；而城西的永陵，是中国唯一建于地面之上的帝王陵，前蜀皇帝王建的二十四伎乐石刻，凝固了晚唐五代的宫廷笙歌。这条路线从神话走进史实，从道观飞檐走到皇家陵阙，触摸这座城市被时光层层包裹的源头——古蜀的密码，就藏在这些青瓦石刻之间。",
			Interests: []string{"历史文化", "艺术"},
			Stops:     []string{"青羊宫", "永陵博物馆"},
		},
	}
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("route-theme-seed")
	defer logger.Close()

	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	if err := logger.SetDB(db); err != nil {
		logger.Fatalf("日志表 log_entries 初始化失败: %v", err)
	}
	// 自动迁移保证 routes.theme 列存在(model 新增/调整后第一次跑本工具即生效)
	database.MustAutoMigrate(db)

	// 库内景点索引:name_zh → 景点(站点信息全部来自库内,不引入外部素材)
	spots := map[string]model.ScenicSpot{}
	var all []model.ScenicSpot
	if err := db.Find(&all).Error; err != nil {
		logger.Fatalf("读取 scenic_spots 失败: %v", err)
	}
	for _, s := range all {
		spots[s.NameZH] = s
	}

	insCnt, updCnt, skipStop := 0, 0, 0
	for _, seed := range cultureRouteSeeds() {
		stops := buildStops(seed, spots, &skipStop)
		if len(stops) == 0 {
			logger.Errorf("路线 [%s] 无可用站点,跳过", seed.TitleZH)
			continue
		}
		stopsJSON, err := json.Marshal(stops)
		if err != nil {
			logger.Errorf("路线 [%s] 站点 JSON 序列化失败: %v", seed.TitleZH, err)
			continue
		}
		cover := ""
		for _, st := range stops {
			if st["image"].(string) != "" {
				cover = st["image"].(string)
				break
			}
		}

		var existing model.Route
		err = db.Where("route_key = ?", seed.RouteKey).First(&existing).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			err = db.Where("title_zh = ?", seed.TitleZH).First(&existing).Error
		}
		if err == nil {
			// 更新前把已入库站点的多语种简介(desc_en/desc_ja)合并回来,避免重跑种子丢失翻译
			mergeStopTranslations(existing.Stops, stops)
			stopsJSON, err = json.Marshal(stops)
			if err != nil {
				logger.Errorf("路线 [%s] 站点 JSON 序列化失败: %v", seed.TitleZH, err)
				continue
			}
			updates := map[string]any{
				"route_key":   seed.RouteKey,
				"theme":       "culture",
				"title_en":    seed.TitleEN,
				"title_ja":    seed.TitleJA,
				"description": seed.Description,
				"days":        1,
				"interests":   strings.Join(seed.Interests, ","),
				"cover_image": cover,
				"stops":       string(stopsJSON),
				"sort":        seed.Sort,
			}
			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				logger.Errorf("路线 [%s] 更新失败: %v", seed.TitleZH, err)
				continue
			}
			logger.Infof("路线 [%s] 已更新(id=%d,%d 个站点,封面 %s)", seed.TitleZH, existing.ID, len(stops), truncate(cover, 50))
			updCnt++
			continue
		}
		if err != gorm.ErrRecordNotFound {
			logger.Fatalf("查询路线失败: %v", err)
		}
		row := model.Route{
			RouteKey:    seed.RouteKey,
			TitleZH:     seed.TitleZH,
			TitleEN:     seed.TitleEN,
			TitleJA:     seed.TitleJA,
			Theme:       "culture",
			Description: seed.Description,
			Days:        1,
			Interests:   strings.Join(seed.Interests, ","),
			CoverImage:  cover,
			Stops:       string(stopsJSON),
			Sort:        seed.Sort,
		}
		if err := db.Create(&row).Error; err != nil {
			logger.Errorf("路线 [%s] 插入失败: %v", seed.TitleZH, err)
			continue
		}
		logger.Infof("路线 [%s] 已插入(id=%d,%d 个站点,封面 %s)", seed.TitleZH, row.ID, len(stops), truncate(cover, 50))
		insCnt++
	}
	logger.Infof("完成: 插入 %d / 更新 %d / 剔除站点 %d 个", insCnt, updCnt, skipStop)
}

// stopJSON 单个站点,字段结构与现有路线 stops 一致(story-enrich 会补 desc_en/desc_ja)。
type stopJSON = map[string]any

// mergeStopTranslations 把已有 stops JSON 中各站点的 desc_en/desc_ja 按 name_zh
// 对位合并进新站点列表(story-enrich 生成后重跑本种子不丢翻译)。
func mergeStopTranslations(existingJSON string, stops []stopJSON) {
	var existing []stopJSON
	if err := json.Unmarshal([]byte(existingJSON), &existing); err != nil {
		return
	}
	byName := map[string]stopJSON{}
	for _, st := range existing {
		if name, _ := st["name_zh"].(string); name != "" {
			byName[name] = st
		}
	}
	for _, st := range stops {
		old, ok := byName[st["name_zh"].(string)]
		if !ok {
			continue
		}
		for _, key := range []string{"desc_en", "desc_ja"} {
			if v, _ := old[key].(string); strings.TrimSpace(v) != "" {
				st[key] = v
			}
		}
	}
}

// buildStops 按候选顺序从库内景点组装站点 JSON;未收录的候选剔除并计数。
func buildStops(seed cultureRouteSeed, spots map[string]model.ScenicSpot, skipCnt *int) []stopJSON {
	stops := make([]stopJSON, 0, len(seed.Stops))
	for _, name := range seed.Stops {
		s, ok := spots[name]
		if !ok {
			logger.Warnf("路线 [%s] 候选站点 %q 库内不存在,已剔除", seed.TitleZH, name)
			*skipCnt++
			continue
		}
		stops = append(stops, stopJSON{
			"name_zh": s.NameZH,
			"name_en": s.NameEN,
			"name_ja": s.NameJA,
			"desc":    capRunes(s.Desc, 200),
			"image":   firstOSS(s.Images),
		})
	}
	return stops
}

// capRunes 按 rune 截断中文文本。
func capRunes(text string, max int) string {
	text = strings.TrimSpace(text)
	if r := []rune(text); len(r) > max {
		return string(r[:max])
	}
	return text
}

// firstOSS 取逗号分隔图片列表中第一张 aliyuncs 地址(与入库图片规范一致,绝不落外链)。
func firstOSS(images string) string {
	for _, u := range strings.Split(images, ",") {
		u = strings.TrimSpace(u)
		if strings.Contains(u, "aliyuncs.com") {
			return u
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
