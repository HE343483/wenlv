# 美食详情页数据打通(高德门店 + LLM 参考值 + 多来源图片)Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `/food/:id`（美食详情页）显示真实数据——封面、评分、风味标签、辣度、人均消费、招牌推荐、推荐场景、风味简介、风味故事（图文段落）、美味瞬间（相册）、寻味地图、相关推荐——而不是本地 mock 与写死的占位。

**Architecture:** 沿用景点详情页已验证的三层取数：① 高德 POI 门店事实（门店名/地址/坐标/评分/人均/相册图）离线采集；② 无权威公开源的字段（风味标签、辣度、招牌推荐、推荐场景、风味故事、兜底人均/评分）由 LLM 生成并登记到 `estimated_fields`，前端加"参考值"徽标；③ 相关推荐不入库，前端按标签从 `/api/food` 列表实时匹配。所有外部图片下载后上传 OSS，库内只存 OSS URL。

**Tech Stack:** Go 1.2x + Gin + GORM + MySQL + Redis(go-redis/v9)；高德 Web 服务（place/text、place/detail）；OpenAI 兼容 LLM（已有 `TripLLM.ChatWithEffort`）；阿里云 OSS V1 Header 签名直传（已有 `pkg.OssSigner`）；Vue 3 + Vite 8 + TypeScript + Vitest。

**Spec:** 本文件「现状与数据来源」章节即规格说明。

## Global Constraints

- 分支：继续用现有分支 `feature/scenic-detail-data`（景点详情的改动也在这个未提交分支上，属于同一批交付）；不要改 `main`、不要合并。
- **本仓库当前约定：所有改动都不提交，只留在工作区**（用户明确选择"先不提交"）。计划里的 Commit 步骤一律跳过。
- 所有新增数据库字段必须有中文 `comment`（Navicat 可读）。
- `/api/food`、`/api/food/:id` 只能**新增**字段，不得改名或删除已有字段（`id/name_zh/name_en/district/tags/desc/images` 原样保留）。
- LLM 生成的字段**必须**写入 `estimated_fields`（逗号分隔字段名），前端展示时必须带"参考值"标注；高德提供的事实值（评分/人均/门店/地址）**不得**登记为参考值。
- LLM 只允许在给定简介内改写；"不确定就写空字符串"；禁止编造精确数字、年份、官方排名。
- 所有外部图片必须下载后上传 OSS（bucket `wenlv-tdx`，`food/` 目录，`Cache-Control: public, max-age=86400`），库内只存 OSS URL。
- 外部接口失败必须降级：采集跳过并记日志；接口返回空数据；前端显示"暂无数据"，**不得**出现加载文案、不得 500。
- 不新增第三方依赖。
- 三语文案（`zh/en/ja`）必须同步；本次生成的美食故事只做中文（与既有 `desc` 行为一致）。

---

## 现状与数据来源

现状（已核对代码与数据库）：

- `foods` 表 23 条，`name_zh/name_en/tags/desc/images` **全部有值**（`images` 已是 OSS 地址，由 `crawler -food` 抓取维基文本 + Commons 图片写入）。
- `FoodDetail.vue` 仍是"本地 mock"版：详情数据来自文件内的 `localDetails` 对象（按 `hotpot`/`chuanchuan` 这类**分类 key** 索引），后端 `getFood`/`fetchFoodDetail` 调用被注释掉未启用。而 `FoodPage.vue` 的数据库卡片跳转用的是 **数据库 id**（`String(f.id)`），于是 `localDetails['3']` 取不到、i18n `food.card.3.name` 也不存在 → 整页空白。
- `/api/food/:id` 已存在（返回 `model.Food`），但没有前端要用的字段。

数据来源与实测证据：

| 页面字段 | 数据来源 | 实测证据 | 落库 | 标注 |
|---|---|---|---|---|
| 封面 / 美味瞬间相册 | 高德 `place/detail` 的 `photos` → 不足时 Wikimedia Commons | 麻婆豆腐/龙抄手/钟水饺 均返回 3 张相册图 | 是 | 事实 |
| 寻味地图（门店名/地址/坐标） | 高德 POI 门店（名称包含菜名的餐饮门店） | `陈麻婆豆腐(旗舰店)` 地址=青华路10号附10-12号；`龙抄手(总店)` 地址=春熙路中山广场东侧城守街63号 | 是 | 事实 |
| 评分 | 高德门店 `biz_ext.rating`；无门店时 LLM 兜底 | 麻婆豆腐 4.7、龙抄手 4.6、钟水饺 4.5 | 是 | 高德=事实；LLM=参考值 |
| 人均消费 | 高德门店 `biz_ext.cost`；无门店时 LLM 兜底 | 麻婆豆腐 71 元、龙抄手 33 元、钟水饺 23 元 | 是 | 同上 |
| 风味标签 / 辣度 / 招牌推荐 / 推荐场景 / 风味故事 | 仅 LLM（基于已有 `desc`） | — | 是 | 参考值 |
| 相关推荐 | 不入库：前端按 `tags` 交集从 `/api/food` 列表取 3 条 | — | 否 | 事实 |

「高德门店匹配」规则（**必须照此实现**，否则会把无关餐厅写进来）：
- 关键词 = `name_zh`，`city=成都`，`citylimit=true`，取 `offset=10`；
- 候选筛选：`type` 含 `餐饮服务`；`name` 含 `name_zh`（如「陈麻婆豆腐」含「麻婆豆腐」、「龙抄手(总店)」含「龙抄手」）；
- 打分：`name == name_zh` 得 3 分，`name` 包含 `name_zh` 得 2 分；同分时取 `biz_ext.rating` 更高者，其次取 `photos` 更多者；
- 一个候选都不满足时**视为未匹配**（宁可不写事实值，也不写无关餐厅）——实测「担担面」就是这种情况（返回的是「肖家河美味家常面」等无同名门店），此时 `address/lat/lng/poi_name` 留空，评分与人均改由 LLM 出参考值。

## 新增字段设计（`foods`）

| 字段 | 类型 | 说明 |
|---|---|---|
| `rating` | varchar(16) | 评分（高德事实或 LLM 参考值） |
| `flavor` | varchar(64) | 风味标签（参考值），如"麻辣鲜香" |
| `spice_level` | varchar(32) | 辣度（参考值），如"重辣" |
| `avg_price` | varchar(32) | 人均消费，如"人均 ¥71" |
| `signature` | varchar(255) | 招牌推荐（参考值） |
| `recommend_scene` | varchar(64) | 推荐场景（参考值） |
| `story_sections` | text | 风味故事段落 JSON：`[{"title","text","image"}]` |
| `gallery_images` | text | 相册图片 OSS URL，逗号分隔 |
| `poi_name` | varchar(128) | 高德门店名（事实） |
| `address` | varchar(255) | 门店地址（事实） |
| `lat` | double | 门店纬度（事实） |
| `lng` | double | 门店经度（事实） |
| `estimated_fields` | varchar(255) | 参考值字段名，逗号分隔 |
| `data_source` | varchar(32) | `amap` / `amap+llm` / `llm` / `wiki` |
| `data_updated_at` | datetime | 本次采集时间 |

## File Map

| 文件 | 职责 |
|---|---|
| `houduan/model/food.go` | 新增字段 + `FoodStorySection` 结构 |
| `houduan/repository/content_repo.go` | `FoodRepo` 新增 `ListAll`、`UpdateFields` |
| `houduan/service/scenic_images.go` | 把 Commons 搜索与图片管线中性别化（供美食复用，行为不变） |
| `houduan/service/food_enrich.go` | 美食采集编排：门店匹配、图片、故事分段、参考值 |
| `houduan/service/food_enrich_test.go` | 纯函数单测（门店匹配打分、字段登记、故事分段） |
| `houduan/cmd/food-enrich/main.go` | 批处理入口（`-limit/-only/-dry-run/-force`） |
| `wennv/wenlv/src/api/content.ts` | `FoodItem` 类型扩展 |
| `wennv/wenlv/src/views/FoodDetail/FoodDetail.vue` | 改为数据驱动（十二个区块 + 地图 + 相关推荐 + 参考值徽标） |
| `wennv/wenlv/src/locales/{zh,en,ja}.ts` | `foodDetail` 新增键与占位文案修正 |

---

### Task 0: 分支确认（无需改动）

- [ ] **Step 1: 确认分支**

```powershell
cd "d:\桌面\文旅\wenlv"
git branch --show-current
```

Expected: `feature/scenic-detail-data`（若不在该分支，先 `git checkout feature/scenic-detail-data`；**不要**切到 main、不要新建分支）。

---

### Task 1: 扩展 `Food` 模型与迁移

**Files:**
- Modify: `houduan/model/food.go`

**Interfaces:**
- Consumes: 无
- Produces: `model.Food` 新增 `Rating/Flavor/SpiceLevel/AvgPrice/Signature/RecommendScene/StorySections/GalleryImages/POIName/Address string`、`Lat/Lng float64`、`EstimatedFields/DataSource string`、`DataUpdatedAt *time.Time`；`model.FoodStorySection{Title, Text, Image string}`

- [ ] **Step 1: 追加字段与结构**

把 `houduan/model/food.go` 改为（保留原有 8 个字段与注释不动，在 `Images` 之后插入）：

```go
	Images    string    `gorm:"type:text;comment:图片URL列表(逗号分隔)" json:"images"`
	// ===== 详情页扩展字段(高德门店采集 + LLM 参考值) =====
	Rating          string     `gorm:"size:16;comment:评分(高德事实或LLM参考值)" json:"rating"`
	Flavor          string     `gorm:"size:64;comment:风味标签(参考值)" json:"flavor"`
	SpiceLevel      string     `gorm:"size:32;comment:辣度(参考值)" json:"spice_level"`
	AvgPrice        string     `gorm:"size:32;comment:人均消费(如 人均 ¥71)" json:"avg_price"`
	Signature       string     `gorm:"size:255;comment:招牌推荐(参考值)" json:"signature"`
	RecommendScene  string     `gorm:"size:64;comment:推荐场景(参考值)" json:"recommend_scene"`
	StorySections   string     `gorm:"type:text;comment:风味故事段落JSON" json:"story_sections"`
	GalleryImages   string     `gorm:"type:text;comment:相册图片URL列表(逗号分隔)" json:"gallery_images"`
	POIName         string     `gorm:"size:128;comment:高德门店名(事实)" json:"poi_name"`
	Address         string     `gorm:"size:255;comment:门店地址(事实)" json:"address"`
	Lat             float64    `gorm:"comment:门店纬度(事实)" json:"lat"`
	Lng             float64    `gorm:"comment:门店经度(事实)" json:"lng"`
	EstimatedFields string     `gorm:"size:255;comment:参考值字段(逗号分隔,前端加标注)" json:"estimated_fields"`
	DataSource      string     `gorm:"size:32;comment:数据来源(amap/amap+llm/llm/wiki)" json:"data_source"`
	DataUpdatedAt   *time.Time `gorm:"comment:数据采集时间" json:"data_updated_at"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// FoodStorySection 风味故事的一段(标题 + 正文 + 配图)。
// 以 JSON 数组序列化后存入 Food.StorySections。
type FoodStorySection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}
```

- [ ] **Step 2: 迁移**

`houduan/database/mysql.go` 的 `AutoMigrate` 已包含 `&model.Food{}`，无需修改。

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
gofmt -w model/food.go
go build ./...
```

Expected: 无输出。

---

### Task 2: 仓储方法 + 图片管线中性别化

**Files:**
- Modify: `houduan/repository/content_repo.go`
- Modify: `houduan/service/scenic_images.go`

**Interfaces:**
- Produces:
  - `(*FoodRepo) ListAll() ([]model.Food, error)`、`(*FoodRepo) UpdateFields(id uint, updates map[string]any) error`
  - 包级函数 `searchCommonsImagesFor(ctx context.Context, client *http.Client, keyword string, names []string, limit int) ([]commonsImageResult, error)`（由现有方法 `(*ScenicEnricher).searchCommonsImages` 抽出，行为不变，方法改为调用它）
  - 包级函数 `uploadImageCandidates(ctx context.Context, signer *pkg.OssSigner, keyPrefix string, candidates []ScenicImageCandidate, galleryCap int) ([]string, int)`（把景点采集里"候选 → 下载 → OSS"的循环抽成包级函数，`keyPrefix` 形如 `scenic/12` 或 `food/3`；返回 OSS URL 列表与上传成功数）

- [ ] **Step 1: `FoodRepo` 追加两个方法**

在 `FoodRepo.GetByID` 之后追加（与 `ScenicRepo` 的同名方法保持一致的实现风格）：

```go
// ListAll 返回全部美食(采集批处理使用,不分页)。
func (r *FoodRepo) ListAll() ([]model.Food, error) {
	var items []model.Food
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateFields 按主键更新指定字段(字段名 → 新值),用于采集结果落库。
func (r *FoodRepo) UpdateFields(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.Food{}).Where("id = ?", id).Updates(updates).Error
}
```

- [ ] **Step 2: Commons 搜索抽成包级函数**

`houduan/service/scenic_images.go` 里把 `func (e *ScenicEnricher) searchCommonsImages(...)` 的实现整体搬到包级 `func searchCommonsImagesFor(ctx, client, keyword, names, limit)`（函数体不变，只去掉接收者；`e.commonsClient` 改为入参 `client`），再保留一个薄方法：

```go
// searchCommonsImages 保留原方法签名,内部委托包级函数,供景点采集继续使用。
func (e *ScenicEnricher) searchCommonsImages(ctx context.Context, keyword string, names []string, limit int) ([]commonsImageResult, error) {
	return searchCommonsImagesFor(ctx, e.commonsClient, keyword, names, limit)
}
```

- [ ] **Step 3: 图片上传循环抽成包级函数**

在 `houduan/service/scenic_images.go` 追加（实现从 `ScenicEnricher.uploadPhotos` 抽出，逻辑与失败处理逐字保留：单张失败跳过、Commons 来源带 429 退避、key = `<keyPrefix>/g<n><ext>`）：

```go
// uploadImageCandidates 下载候选图片并上传 OSS,返回 OSS URL 列表与成功张数。
// keyPrefix 形如 "scenic/12" 或 "food/3";单张失败只记日志并跳过;
// 达到 galleryCap 张即停止。
func uploadImageCandidates(ctx context.Context, signer *pkg.OssSigner, keyPrefix string, candidates []ScenicImageCandidate, galleryCap int) ([]string, int)
```

然后把 `(*ScenicEnricher).uploadPhotos` 改为调用它（保持景点行为不变）：

```go
func (e *ScenicEnricher) uploadPhotos(ctx context.Context, s model.ScenicSpot, candidates []ScenicImageCandidate) ([]string, error) {
	if e.signer == nil || !e.signer.Configured() {
		return nil, fmt.Errorf("OSS 未配置")
	}
	urls, _ := uploadImageCandidates(ctx, e.signer, fmt.Sprintf("scenic/%d", s.ID), candidates, scenicGalleryWant)
	return urls, nil
}
```

- [ ] **Step 4: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
gofmt -l repository/content_repo.go service/scenic_images.go service/scenic_enrich.go
go build ./...
go vet ./service/ ./repository/
go test ./service/ -count=1 -run 'Test' | Select-Object -Last 5
```

Expected: gofmt 无输出；构建/vet 无输出；既有测试全 PASS（这次重构**不得改变景点采集行为**）。

---

### Task 3: 美食采集服务 `service/food_enrich.go`

**Files:**
- Create: `houduan/service/food_enrich.go`
- Create: `houduan/service/food_enrich_test.go`

**Interfaces:**
- Consumes: `AmapService.SearchPOI/GetPOIDetail`（后者已返回 `Rating/Cost/Photos/BizExt`）、`TripLLM.ChatWithEffort`、`pkg.OssSigner`、`searchCommonsImagesFor`/`uploadImageCandidates`（Task 2）、`FoodRepo`（Task 2）、`extractJSONFromResponse`、`llmThinkingLevel()`
- Produces:
  - `type FoodEnrichOptions struct{ WithImages, WithLLM, Force bool }`
  - `type FoodEnrichResult struct{ FoodID uint; Name string; POIMatched bool; Uploaded int; LLMUsed bool; UpdatedFields []string; Note string }`
  - `func NewFoodEnricher(repo *repository.FoodRepo, amap *AmapService, llm *TripLLM, signer *pkg.OssSigner) *FoodEnricher`
  - `func (e *FoodEnricher) Enrich(ctx context.Context, f model.Food, opts FoodEnrichOptions) (*FoodEnrichResult, error)`
  - 纯函数：`matchFoodPOI(candidates []model.POIInfo, nameZH string) (model.POIInfo, bool)`、`foodIsUsablePOI(poiType string) bool`、`appendEstimatedField(existing, field string) string`（**已存在于 `scenic_enrich.go` 的同名函数，直接复用，不要重复定义**）、`isEmptyFoodField(f model.Food, field string) bool`

- [ ] **Step 1: 写失败测试**

创建 `houduan/service/food_enrich_test.go`：

```go
package service

import (
	"testing"

	"wenlv-backend/model"
)

func TestMatchFoodPOI(t *testing.T) {
	candidates := []model.POIInfo{
		{Name: "肖家河美味家常面(肖家河北街店)", Type: "餐饮服务;中餐厅;中餐厅"},
		{Name: "担担面(总店)", Type: "餐饮服务;中餐厅;特色/地方风味餐厅"},
		{Name: "担担面(文殊坊店)", Type: "餐饮服务;中餐厅;中餐厅"},
	}
	got, ok := matchFoodPOI(candidates, "担担面")
	if !ok {
		t.Fatal("应当匹配到同名门店")
	}
	if got.Name != "担担面(总店)" {
		t.Errorf("应优先取名称包含菜名的门店, got %q", got.Name)
	}
	if _, ok := matchFoodPOI([]model.POIInfo{{Name: "肖家河美味家常面", Type: "餐饮服务;中餐厅;中餐厅"}}, "担担面"); ok {
		t.Error("无同名门店时必须返回未匹配,不能随便取一个餐厅")
	}
	if _, ok := matchFoodPOI([]model.POIInfo{{Name: "陈麻婆豆腐(旗舰店)", Type: "购物服务;超市"}}, "麻婆豆腐"); ok {
		t.Error("非餐饮服务类型应被排除")
	}
	if _, ok := matchFoodPOI(nil, "火锅"); ok {
		t.Error("空候选应返回未匹配")
	}
}

func TestFoodIsUsablePOI(t *testing.T) {
	cases := map[string]bool{
		"餐饮服务;中餐厅;四川菜(川菜)": true,
		"购物服务;超级市场;超市":      false,
		"餐饮服务;快餐厅;快餐厅":      true,
		"":               false,
	}
	for in, want := range cases {
		if got := foodIsUsablePOI(in); got != want {
			t.Errorf("foodIsUsablePOI(%q)=%v, want %v", in, got, want)
		}
	}
}

func TestIsEmptyFoodField(t *testing.T) {
	f := model.Food{}
	if !isEmptyFoodField(f, "flavor") {
		t.Error("空记录的风味字段应视为空")
	}
	f2 := model.Food{Flavor: "麻辣鲜香", Images: "/images/placeholder-food.jpg"}
	if isEmptyFoodField(f2, "flavor") {
		t.Error("已有值不应视为空")
	}
	if !isEmptyFoodField(f2, "images") {
		t.Error("占位封面应视为空(可被真实封面替换)")
	}
	if !isEmptyFoodField(f2, "unknown_field") {
		t.Error("未知字段默认视为空")
	}
}
```

- [ ] **Step 2: 运行确认失败**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go test ./service/ -count=1 -run 'TestMatchFoodPOI|TestFoodIsUsablePOI|TestIsEmptyFoodField' -v
```

Expected: 编译失败 `undefined: matchFoodPOI`。

- [ ] **Step 3: 实现 `service/food_enrich.go`**

```go
// 美食详情数据采集与补全:高德门店事实字段 + LLM 参考值 + 风味故事分段 + 多来源图片。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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
				log.Printf("    Commons 取图失败: %v", err)
			}
		}
		if urls, n := uploadImageCandidates(ctx, e.signer, fmt.Sprintf("food/%d", f.ID), candidates, foodGalleryWant); n > 0 {
			images = urls
			res.Uploaded = n
			updates["gallery_images"] = strings.Join(urls, ",")
			log.Printf("    图片: 上传 %d 张", n)
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
```

**注意**：`splitComma`、`isPlaceholderCover`、`survivingEstimated`、`hasBusinessUpdate`、`buildDataSource`、`clampMaxSections`、`estimatedKV`、`llmMessage`、`extractJSONFromResponse`、`llmThinkingLevel`、`ScenicImageCandidate`、`newCommonsHTTPClient`、`searchCommonsImagesFor`、`uploadImageCandidates` 都是同包已存在的符号，**直接复用，不要重复定义**（若签名有细微差异，以现有代码为准并在报告里说明）。

- [ ] **Step 4: 运行测试确认通过**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
gofmt -l service/food_enrich.go service/food_enrich_test.go
go build ./...
go vet ./service/
go test ./service/ -count=1 -run 'TestMatchFoodPOI|TestFoodIsUsablePOI|TestIsEmptyFoodField' -v
```

Expected: gofmt 无输出；构建/vet 无输出；3 个测试 PASS。

---

### Task 4: 批处理入口 `cmd/food-enrich`

**Files:**
- Create: `houduan/cmd/food-enrich/main.go`

**Interfaces:**
- Consumes: `service.NewFoodEnricher`、`service.FoodEnrichOptions`（Task 3）、`service.NewTripSettings`/`NewAmapService`/`NewTripLLM`、`config.Load()`、`database.InitMySQL`/`MustAutoMigrate`、`repository.NewFoodRepo`、`pkg.NewOssSigner`
- Produces: CLI，flags `-limit -only -dry-run -force`

- [ ] **Step 1: 新建入口**（结构与 `cmd/scenic-enrich/main.go` 保持一致：`godotenv.Load()` → `config.Load()` → `mustDB(cfg, !*dryRun)` → `TripSettings`（含 `AttachDB`）→ `amap`/`llm`/`signer` → `repo.ListAll()` → `-only`/`-limit` 过滤 → 逐个 `Enrich` → 打印汇总）

要求：
- 包注释写明用法：
```go
// food-enrich 美食详情数据补全批处理。
// 数据源:高德门店(名称含菜名的餐饮 POI:地址/坐标/评分/人均/相册) + LLM(风味标签/辣度/招牌/场景/风味故事)。
//
//	go run ./cmd/food-enrich -limit 3 -dry-run
//	go run ./cmd/food-enrich -only 麻婆豆腐 -force
//	go run ./cmd/food-enrich
```
- `-dry-run` 只做只读查询（该模式下**不调用** `Enrich`，改为打印 `MatchFoodPOI` 的匹配结果与门店地址/评分/人均/相册张数），且 `mustDB(cfg, !*dryRun)` 与景点 CLI 一致（dry-run 不跑迁移）；
- 每轮迭代开头 `time.Sleep(300 * time.Millisecond)`（与景点 CLI 一致，避免高德限流）；
- 结束时打印 `完成: 成功 X / 失败 Y / 试跑跳过 Z`。

- [ ] **Step 2: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
gofmt -l cmd/food-enrich/main.go
go build ./...
go vet ./cmd/food-enrich/
go run ./cmd/food-enrich -only 麻婆豆腐 -dry-run
go run ./cmd/food-enrich -only 担担面 -dry-run
```

Expected：
- `麻婆豆腐` → 匹配到 `陈麻婆豆腐(旗舰店)`，打印地址/评分（约 4.7）/人均（约 71）/相册 3 张；
- `担担面` → 打印「未匹配到同名门店」（实测高德只有「肖家河美味家常面」等无关餐厅，属预期）；
- 无迁移 DDL 日志、无写库。

---

### Task 5: 前端数据驱动 + 三语文案

**Files:**
- Modify: `wennv/wenlv/src/api/content.ts`
- Modify: `wennv/wenlv/src/views/FoodDetail/FoodDetail.vue`
- Modify: `wennv/wenlv/src/locales/{zh,en,ja}.ts`

**Interfaces:**
- Consumes: `getFood(id)`、`listFoods({ page_size: 50 })`、`parseSections`/`estimatedSet`/`splitList`/`displayFact`（`@/utils/scenicDetail`，其 JSON 形状与美食故事一致，直接复用）、`getRuntimeMapJsKey`（`@/trip/services/api`）
- Produces: `FoodItem` 新增可选字段 `rating/flavor/spice_level/avg_price/signature/recommend_scene/story_sections/gallery_images/poi_name/address/lat/lng/estimated_fields/data_source/data_updated_at`；FoodDetail 渲染真实数据

- [ ] **Step 1: 三语文案**

`foodDetail` 段新增（三份文件同步）：
```ts
    estimated: '参考值',
    estimatedTip: '该信息由 AI 依据公开资料整理，仅供参考，请以商家实际为准',
    noData: '暂无数据',
    dataSource: '数据来源：高德地图 / 维基百科 / Wikimedia Commons',
    relatedEmpty: '暂无相关推荐',
```
en：
```ts
    estimated: 'Reference',
    estimatedTip: 'AI-generated from public sources for reference only. Please check with the merchant.',
    noData: 'No data',
    dataSource: 'Sources: AMap / Wikipedia / Wikimedia Commons',
    relatedEmpty: 'No related recommendations',
```
ja：
```ts
    estimated: '参考値',
    estimatedTip: 'AIが公開情報をもとに整理した参考値です。実際の店舗情報をご確認ください。',
    noData: 'データなし',
    dataSource: '出典：高徳地図 / ウィキペディア / Wikimedia Commons',
    relatedEmpty: '関連のおすすめはありません',
```
同时把仍是占位口吻的 `gallerySubtitle`、`mapSubtitle` 改为已接入的说法（三语文案自拟，语义与景点详情页一致，例如 zh：`gallerySubtitle: '更多美食影像'`、`mapSubtitle: '推荐品尝门店位置'`）。

- [ ] **Step 2: `content.ts` 扩展 `FoodItem`**

在 `FoodItem` 内追加 15 个可选字段（字段名与后端 json tag 一致）；`listFoods`/`getFood` 不变。

- [ ] **Step 3: `FoodDetail.vue` 改造**

要点（写法照抄 `ScenicDetail.vue` 的既有实现，只换数据字段与文案 key）：
- 删除文件内的 `localDetails` mock 与 `placeholder*` 占位逻辑；当路由 id **不是数字**时（老分类卡片 `hotpot` 等）保留现有 i18n 文案渲染作为降级分支，不要白屏；
- 数字 id：`const numericId = Number(foodId.value)`，`getFood(numericId)` 拉详情；
- `storySections = computed(() => parseSections(food.story_sections))`、`galleryImages = computed(() => splitList(food.gallery_images))`、`estimated = computed(() => estimatedSet(food.estimated_fields))`、`displayFact` 负责空值显示 `noData`；
- 顶部 hero 用 `food.images`（加载失败回退占位）；评分用 `food.rating`，风味标签/辣度/人均用 `food.flavor`/`food.spice_level`/`food.avg_price`，招牌/场景用 `food.signature`/`food.recommend_scene`，命中 `estimated` 的加 `detail-est` 徽标；
- 风味简介用 `food.desc`；风味故事用 `storySections` 做左右交替的图文行，无数据时显示 `noData`；
- 美味瞬间用 `galleryImages` 网格；无数据时 `noData`；
- 寻味地图：`food.lat`/`food.lng` 有效且拿到地图 Key 时用 `AMapLoader` 渲染标记（容器常驻可见、占位块 `v-if="!mapReady"` 绝对定位覆盖、卸载 `destroy()`、失败 `console.warn` 回退占位），否则显示占位 + `noData`；
- 相关推荐：`listFoods({ page_size: 50 })` 后按 `tags` 交集排序（排除自身）取 3 条，渲染名称 + 简介前 40 字；点击跳转到对应 `/food/{id}`；
- 页脚加 `dataSource` 文案，`estimated.size` 非空时追加 `estimatedTip`。
- 三语切换沿用 `langStore`；不得出现 `common.loading` 文案。

- [ ] **Step 4: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\wennv\wenlv"
npm run type-check
npx vitest run
```

Expected: 类型检查无错误；既有测试全部通过。

---

### Task 6: 全量采集与端到端验收

- [ ] **Step 1: 后端构建与单测**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go build ./...
go test ./... 2>&1 | Select-Object -Last 15
```

Expected: 构建通过；`pkg` 与 `service` 的测试全 PASS。

- [ ] **Step 2: 全量采集**

```powershell
go run ./cmd/food-enrich 2>&1 | Tee-Object -FilePath ../.superpowers/sdd/food-enrich.log
```

Expected: 23 条美食全部处理完（`完成: 成功 23 / 失败 0`）；日志里能看到「匹配门店: 陈麻婆豆腐(旗舰店)」这类行，以及无同名门店的那些（如担担面）。

- [ ] **Step 3: 数据质量抽查**

```sql
SELECT COUNT(*) AS total, SUM(rating<>'') AS has_rating, SUM(flavor<>'') AS has_flavor,
 SUM(avg_price<>'') AS has_price, SUM(signature<>'') AS has_signature,
 SUM(story_sections<>'') AS has_story, SUM(gallery_images<>'') AS has_gallery,
 SUM(poi_name<>'') AS has_poi, SUM(estimated_fields<>'') AS has_est FROM foods;
SELECT COUNT(*) AS bad_label FROM foods
 WHERE (flavor<>'' AND estimated_fields NOT LIKE '%flavor%')
    OR (spice_level<>'' AND estimated_fields NOT LIKE '%spice_level%')
    OR (signature<>'' AND estimated_fields NOT LIKE '%signature%')
    OR (recommend_scene<>'' AND estimated_fields NOT LIKE '%recommend_scene%');
SELECT COUNT(*) AS bad_img FROM foods WHERE gallery_images<>'' AND gallery_images NOT LIKE '%aliyuncs.com%';
SELECT COUNT(*) AS bad_story FROM foods WHERE story_sections<>'' AND JSON_VALID(story_sections)=0;
```

Expected: `bad_label` = 0（LLM 字段必须登记参考值）、`bad_img` = 0（只存 OSS）、`bad_story` = 0（合法 JSON）；`has_story`/`has_flavor` 覆盖率明显高于 0。

- [ ] **Step 4: 浏览器验收**

前后端同时运行（后端 8081、前端 5173），在浏览器打开 `/food/{id}`（对有同名门店的如「麻婆豆腐」「龙抄手」与无同名门店的「担担面」各一个）：
1. 封面与 hero 名称/英文名正常；
2. 评分/风味/辣度/人均/招牌/场景有值或"暂无数据"，LLM 字段带"参考值"徽标；
3. 风味简介、风味故事（每段都有配图）、美味瞬间（3-6 张 OSS 图）；
4. 寻味地图：有门店坐标的显示真实地图与标记；无门店的显示占位不报错；
5. 相关推荐 3 条且不包含自己，点击能跳转；
6. 页脚数据来源文案；
7. 中/EN/日 切换正常，无未翻译 key；
8. 控制台无 error。

- [ ] **Step 5: 汇总（不提交）**

```powershell
git status --short
```

Expected: 只有源码改动与之相关文件；**不要执行 commit**（用户明确"先不提交"）。

---

## 风险与合规

- **门店匹配**：只接受"名称包含菜名 + 餐饮服务"的 POI。实测「担担面」无同名门店 → 该条不写门店事实（地址/坐标/评分/人均），评分与人均改由 LLM 参考值兜底并标注，避免把无关餐厅当成该菜的"推荐门店"。
- **LLM 参考值**：风味/辣度/招牌/场景/故事一律标注"参考值"并附免责说明；人均与评分在无门店时同样标注。
- **图片**：高德门店相册通常 3 张，不足 6 张时用 Wikimedia Commons 兜底（CC 授权，页脚已标注来源）；小红书管线已接在景点侧，本次美食图片**不启用**小红书，避免 Cookie 失效时的额外耗时。
- **三语**：风味故事正文只生成中文，与景点详情页保持一致。
