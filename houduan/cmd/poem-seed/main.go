// poem-seed 诗词地图种子程序:内置 12 首与库内景点真实相关的经典诗词。
// 红线:原文均为人工权威录入的通行文本(依据《全唐诗》/通行选本),绝不由 LLM 生成;
// 启动时建表(AutoMigrate) → 按 标题+作者 幂等 upsert 中文原文 →
// 调用 TripLLM 逐首生成英/日翻译与白话赏析(失败重试一次后跳过并记录日志,不中断整批)。
//
// 挂靠原则:一首诗只挂一个景点,只挂内容真实相关的库内景点;不确定相关性的不收录。
// 库内无 望江楼/浣花溪公园/金沙遗址,故薛涛诗等暂不录入。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/poem-seed
//	go run ./cmd/poem-seed -force        # 已有译文/赏析也重新生成
//	go run ./cmd/poem-seed -only 蜀相
package main

import (
	"context"
	"errors"
	"flag"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/service"
)

// seedPoem 一首待录入诗词:原文为权威通行文本,SpotName 按 scenic_spots.name_zh 匹配。
type seedPoem struct {
	Title    string
	Dynasty  string
	Author   string
	Content  string
	SpotName string
}

// seedPoems 诗词地图首批数据(12 首)。
var seedPoems = []seedPoem{
	// ===== 杜甫草堂:杜甫寓居成都草堂时期的作品 =====
	{
		Title:   "春夜喜雨",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "好雨知时节，当春乃发生。\n随风潜入夜，润物细无声。\n野径云俱黑，江船火独明。\n晓看红湿处，花重锦官城。",
		SpotName: "杜甫草堂",
	},
	{
		Title:   "客至",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "舍南舍北皆春水，但见群鸥日日来。\n花径不曾缘客扫，蓬门今始为君开。\n盘飧市远无兼味，樽酒家贫只旧醅。\n肯与邻翁相对饮，隔篱呼取尽余杯。",
		SpotName: "杜甫草堂",
	},
	{
		Title:   "江畔独步寻花·其六",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "黄四娘家花满蹊，千朵万朵压枝低。\n留连戏蝶时时舞，自在娇莺恰恰啼。",
		SpotName: "杜甫草堂",
	},
	{
		Title:   "水槛遣心二首·其一",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "去郭轩楹敞，无村眺望赊。\n澄江平少岸，幽树晚多花。\n细雨鱼儿出，微风燕子斜。\n城中十万户，此地两三家。",
		SpotName: "杜甫草堂",
	},
	{
		// 节选开篇完整段落(至"归来倚杖自叹息"),全诗较长故取其起。
		Title:   "茅屋为秋风所破歌（节选）",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "八月秋高风怒号，卷我屋上三重茅。\n茅飞渡江洒江郊，高者挂罥长林梢，下者飘转沉塘坳。\n南村群童欺我老无力，忍能对面为盗贼。\n公然抱茅入竹去，唇焦口燥呼不得，归来倚杖自叹息。",
		SpotName: "杜甫草堂",
	},
	// ===== 武侯祠:怀诸葛亮之作,《蜀相》即咏成都武侯祠 =====
	{
		Title:   "蜀相",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "丞相祠堂何处寻？锦官城外柏森森。\n映阶碧草自春色，隔叶黄鹂空好音。\n三顾频烦天下计，两朝开济老臣心。\n出师未捷身先死，长使英雄泪满襟。",
		SpotName: "武侯祠",
	},
	{
		Title:   "八阵图",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "功盖三分国，名成八阵图。\n江流石不转，遗恨失吞吴。",
		SpotName: "武侯祠",
	},
	{
		// 节选开篇一段,"孔明庙前有老柏"咏祠前古柏。
		Title:   "古柏行（节选）",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "孔明庙前有老柏，柯如青铜根如石。\n霜皮溜雨四十围，黛色参天二千尺。\n云来气接巫峡长，月出寒通雪山白。\n君臣已与时际会，树木犹为人爱惜。",
		SpotName: "武侯祠",
	},
	// ===== 锦里古街:张籍笔下的锦江、万里桥畔酒家市井(万里桥毗邻锦里) =====
	{
		Title:   "成都曲",
		Dynasty: "唐",
		Author:  "张籍",
		Content: "锦江近西烟水绿，新雨山头荔枝熟。\n万里桥边多酒家，游人爱向谁家宿。",
		SpotName: "锦里古街",
	},
	// ===== 西岭雪山:"窗含西岭千秋雪"即咏西岭 =====
	{
		Title:   "绝句",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "两个黄鹂鸣翠柳，一行白鹭上青天。\n窗含西岭千秋雪，门泊东吴万里船。",
		SpotName: "西岭雪山",
	},
	// ===== 青羊宫:陆游诗直接点名"青羊宫到浣花溪" =====
	{
		Title:   "梅花绝句",
		Dynasty: "宋",
		Author:  "陆游",
		Content: "当年走马锦城西，曾为梅花醉似泥。\n二十里中香不断，青羊宫到浣花溪。",
		SpotName: "青羊宫",
	},
	// ===== 青城山:丈人山即青城山主峰,杜甫游山之作 =====
	{
		Title:   "丈人山",
		Dynasty: "唐",
		Author:  "杜甫",
		Content: "自为青城客，不唾青城地。\n为爱丈人山，丹梯近幽意。\n丈人祠西佳气浓，缘云拟住最高峰。\n扫除白发黄精在，君看他时冰雪容。",
		SpotName: "青城山",
	},
}

func main() {
	force := flag.Bool("force", false, "已有译文/赏析时也重新生成")
	only := flag.String("only", "", "仅处理标题包含该关键字的诗词")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	logger.ConfigureTool("poem-seed")
	defer logger.Close()

	db := mustDB(cfg)

	// 与 main.go 一致:LLM Key 先取 .env,再被 data/runtime_settings.json 覆盖(设置页热更新过的值)
	tripSettings := service.NewTripSettings(service.TripSettingsOptions{
		Defaults: service.TripRuntimeSettings{
			ViteAmapWebKey:   cfg.Trip.AmapWebKey,
			GoogleMapsAPIKey: cfg.Trip.GoogleMapsAPIKey,
			GoogleMapsProxy:  cfg.Trip.GoogleMapsProxy,
			OpenAIAPIKey:     cfg.Trip.LLMAPIKey,
			OpenAIBaseURL:    cfg.Trip.LLMBaseURL,
			OpenAIModel:      cfg.Trip.LLMModel,
		},
		DataDir:        cfg.Trip.DataDir,
		PlannerTimeout: cfg.Trip.PlannerTimeout,
		LLMTimeout:     cfg.Trip.LLMTimeout,
	})
	tripSettings.AttachDB(db) // 设置页写入的值优先取自 trip_settings 表,与主服务保持一致
	llm := service.NewTripLLM(tripSettings)
	if !llm.Available() {
		logger.Fatalf("LLM 未配置:请在 .env 配置 OPENAI_API_KEY 或在设置页配置后重试")
	}
	enricher := service.NewPoemEnricher(llm)
	ctx := context.Background()

	// 1) 按名称匹配库内景点,未收录的景点宁可不挂靠
	spotIDs, err := matchSpots(db)
	if err != nil {
		logger.Fatalf("查询景点失败: %v", err)
	}

	// 2) 幂等 upsert 中文原文(按 标题+作者 去重)
	enrichTargets := make([]*model.Poem, 0, len(seedPoems))
	ok, skip := 0, 0
	for _, sp := range seedPoems {
		if *only != "" && !strings.Contains(sp.Title, *only) {
			continue
		}
		spotID, found := spotIDs[sp.SpotName]
		if !found {
			logger.Warnf("跳过《%s》:景点 %q 不在库内,宁可不挂靠", sp.Title, sp.SpotName)
			skip++
			continue
		}
		p, err := upsertPoem(db, sp, spotID)
		if err != nil {
			logger.Errorf("《%s》原文入库失败: %v", sp.Title, err)
			skip++
			continue
		}
		enrichTargets = append(enrichTargets, p)
		ok++
	}
	logger.Infof("原文录入完成: 成功 %d / 跳过 %d", ok, skip)

	// 3) 逐首生成英/日翻译与白话赏析(译文属于参考值,失败跳过不中断)
	done, fail := 0, 0
	need := 0
	for _, p := range enrichTargets {
		if *force || p.ContentEN == "" || p.ContentJA == "" || p.PlainZH == "" {
			need++
		}
	}
	index := 0
	for _, p := range enrichTargets {
		if !*force && p.ContentEN != "" && p.ContentJA != "" && p.PlainZH != "" {
			continue
		}
		index++
		logger.Infof("[%d/%d] 生成《%s》(%s·%s) 译文与赏析...", index, need, p.Title, p.Dynasty, p.Author)
		changed, err := enricher.EnrichPoem(ctx, p)
		if err != nil {
			logger.Errorf("《%s》生成失败(已重试),跳过: %v", p.Title, err)
			fail++
			continue
		}
		if !changed {
			logger.Errorf("《%s》生成结果为空,跳过写库", p.Title)
			fail++
			continue
		}
		if err := db.Model(&model.Poem{}).Where("id = ?", p.ID).Updates(map[string]any{
			"content_en": p.ContentEN,
			"content_ja": p.ContentJA,
			"plain_zh":   p.PlainZH,
		}).Error; err != nil {
			logger.Errorf("《%s》译文写库失败: %v", p.Title, err)
			fail++
			continue
		}
		done++
		logger.Infof("《%s》完成", p.Title)
		if index < need {
			time.Sleep(500 * time.Millisecond) // 逐首节流,避免打爆 LLM 额度
		}
	}
	logger.Infof("译文赏析完成: 成功 %d / 失败 %d", done, fail)
	logger.Infof("全部完成")
}

// mustDB 连接数据库并确保表结构最新(新增列需要迁移)。
func mustDB(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		logger.Fatalf("MySQL 连接失败: %v", err)
	}
	// 工具日志落库 log_entries(module=poem-seed)
	if err := logger.SetDB(db); err != nil {
		logger.Fatalf("日志表 log_entries 初始化失败: %v", err)
	}
	database.MustAutoMigrate(db)
	return db
}

// matchSpots 按 name_zh 建立景点名到 ID 的映射。
func matchSpots(db *gorm.DB) (map[string]uint, error) {
	spots := make([]model.ScenicSpot, 0)
	if err := db.Select("id, name_zh").Find(&spots).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uint, len(spots))
	for _, s := range spots {
		m[s.NameZH] = s.ID
	}
	return m, nil
}

// upsertPoem 按 标题+作者 幂等写入:不存在则新增;已存在时以代码内置权威原文校正
// 原文/朝代/挂靠景点,译文与赏析字段不在此处触碰(由 LLM 生成流程负责)。
func upsertPoem(db *gorm.DB, sp seedPoem, spotID uint) (*model.Poem, error) {
	var p model.Poem
	err := db.Where("title = ? AND author = ?", sp.Title, sp.Author).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p = model.Poem{
			Title:         sp.Title,
			Dynasty:       sp.Dynasty,
			Author:        sp.Author,
			ContentZH:     sp.Content,
			RelatedSpotID: spotID,
		}
		if err := db.Create(&p).Error; err != nil {
			return nil, err
		}
		logger.Infof("新增《%s》(%s·%s) → %s", sp.Title, sp.Dynasty, sp.Author, sp.SpotName)
		return &p, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if p.ContentZH != sp.Content {
		updates["content_zh"] = sp.Content
	}
	if p.Dynasty != sp.Dynasty {
		updates["dynasty"] = sp.Dynasty
	}
	if p.RelatedSpotID != spotID {
		updates["related_spot_id"] = spotID
	}
	if len(updates) > 0 {
		if err := db.Model(&p).Updates(updates).Error; err != nil {
			return nil, err
		}
		logger.Infof("校正《%s》: %d 个字段以权威原文为准", sp.Title, len(updates))
	} else {
		logger.Infof("《%s》已存在,原文一致,跳过", sp.Title)
	}
	return &p, nil
}
