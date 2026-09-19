# 美食大类详情页(川菜 / 名小吃 / 夜宵)Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 「美食名片」六个卡片中，`川菜`、`名小吃`、`夜宵` 是**大类**（不是单道菜），点开后应进入"介绍这个类别"的详情页：类别概述、若干图文段落（起源/特点/代表/去哪吃）、类别图集、以及该类别下的**真实菜品**（可点进菜品详情）。而 `火锅`、`串串香`、`盖碗茶` 是具体品类，仍直达菜品详情页。

**Architecture:** 新增 `food_categories` 表存三大类的介绍内容。内容采集分两步：① **文本**以中文维基百科条目正文为唯一事实素材（`variant=zh-cn` 保证简体），交给 LLM **只做压缩改写**（不得引入素材外的事实），逐段带配图；② **图片**三个来源按精确度优先级聚合：该类别下**已采集菜品的 OSS 图**（100% 属于该类）→ Wikimedia Commons 类别关键词搜图（CC 授权）→ 现有菜品图兜底。全部图片下载后上传 OSS，库内只存 OSS URL。前端详情页按 `key` 判断走"类别页"还是"菜品页"。

**Tech Stack:** Go 1.2x + Gin + GORM + MySQL；中文维基 Action API（需代理，复用 `COMMONS_PROXY`）；OpenAI 兼容 LLM（`TripLLM.ChatWithEffort`）；阿里云 OSS（`pkg.OssSigner`，复用 `service` 包已有的 `searchCommonsImagesFor` / `uploadImageCandidates` / `fetchBytes`）；Vue 3 + Vite + TypeScript + Vitest。

**Spec:** 本文件「类别内容设计」与「数据来源与精确性」两节即规格说明。

## Global Constraints

- 分支：继续用 `feature/scenic-detail-data`；**不要提交**（用户明确"先不提交"），改动只留工作区。
- 所有新增数据库字段/表必须有中文 `comment`（Navicat 可读）。
- 只能**新增**接口与字段：`/api/food`、`/api/food/:id` 现有行为不得改变。
- 类别页的文本由 LLM 依据维基素材改写，**必须**在 `estimated_fields` 中登记被 LLM 生成/改写的字段（`intro`、`sections`），前端展示带"参考值"与来源说明；**不得编造素材之外的事实、数字、排名**。
- 图片必须下载后上传 OSS（bucket `wenlv-tdx`，`foodcat/` 目录，`Cache-Control: public, max-age=86400`），库内只存 OSS URL；外部图片外链不得直接落库。
- 外部接口失败必须降级：采集跳过并记日志；接口返回空数据；前端显示"暂无数据"，不得 500、不得白屏。
- 不新增第三方依赖。
- 三语文案（`zh/en/ja`）同步：类别名与页面 UI 文案三语；**类别正文只生成中文**（与景点/美食详情一致），页面在 EN/JA 下显示中文正文并保留"参考值"标注。

---

## 类别内容设计

三个大类的内容结构与来源：

| 字段 | 说明 | 来源 |
|---|---|---|
| `key` | `cuisine`（川菜）/ `snacks`（名小吃）/ `nightfood`（夜宵） | 固定 |
| `name_zh` / `name_en` | 川菜/名小吃/夜宵 · Sichuan Cuisine / Street Snacks / Night Food | 沿用前端现有 i18n 文案（保持卡片与详情页一致） |
| `intro` | 150–250 字概述 | LLM 依据维基素材改写（参考值） |
| `sections` | 3–4 段 JSON `[{"title","text","image"}]`，建议结构：起源与特点 / 味型与技法（或品种与做法）/ 代表菜品 / 去哪里吃 | 同上（参考值），配图见下 |
| `gallery_images` | 4–6 张类别图（OSS URL，逗号分隔） | 见「数据来源与精确性」 |
| `source_url` | 素材来源链接（维基条目 URL） | 维基 API 返回 |
| `estimated_fields` | 固定为 `intro,sections` | — |
| `data_source` | `wiki+llm+oss` | — |
| `data_updated_at` | 采集时间 | — |

## 数据来源与精确性

| 来源 | 用途 | 实测结论 | 是否启用 |
|---|---|---|---|
| 中文维基百科（`variant=zh-cn`） | **文本唯一素材** | 实测条目存在：`川菜`、`成都小吃`、`夜宵` 均 OK 且返回简体；`四川小吃`/`成都夜市` 无条目（用 `redirects=1` 自动跟随重定向） | ✅ |
| 该类下已采集菜品的 OSS 图 | **图片主源** | 精确：川菜类含麻婆豆腐/宫保鸡丁/回锅肉，名小吃含担担面/龙抄手/钟水饺等，它们的 `images`/`gallery_images` 都已是 OSS 图 | ✅ |
| Wikimedia Commons | 图片补充 | 复用已实现的 `searchCommonsImagesFor`（要求文件名含关键词、宽度≥800、优先 `thumburl` 缩略图规避 429） | ✅ |
| 小红书 | 图片补充 | **不可用**：Cookie 失效（接口返回 `code=-100 登录已过期`） | ❌ 待 Cookie 更新 |
| 百度百科 | 文本/图片补充 | 页面可直连，但图片版权不清晰、HTML 结构易变 | ❌ 本次不启用（如需可后续单开任务） |

**图片精确性规则（必须照此实现）**：
1. 先取该类别下菜品的图：查询 `/api/food` 对应的 DB 行（复用 `FoodRepo.ListAll()`），按「类别标签映射」筛出菜品，收集其 `gallery_images`/`images` 的 OSS URL，按顺序去重取前 6 张；
2. 不足 6 张时用 Commons 按类别关键词补：川菜 → `川菜`/`Sichuan cuisine`；名小吃 → `成都小吃`/`Chengdu snacks`/`四川小吃`；夜宵 → `成都夜市`/`Chengdu night market`/`烧烤`（Commons 图必须文件名包含关键词之一）；
3. 仍不足则保持现有张数（不硬凑）。

**类别标签映射（与既有分类页一致，已用数据库核对）**：
- `cuisine` → 标签包含 `川菜`：麻婆豆腐、宫保鸡丁、回锅肉
- `snacks` → 标签包含 `小吃`：担担面、龙抄手、钟水饺、三大炮、糖油果子、蛋烘糕、叶儿粑、韩包子、赖汤圆
- `nightfood` → 标签包含 `夜宵`：兔头（另可包含 `串串`/`凉菜` 作为夜间食物补充，但**必须**在报告里说明并保证图上菜名一致）

## File Map

| 文件 | 职责 |
|---|---|
| `houduan/model/food_category.go` | `FoodCategory` 模型 + `CategorySection` 结构 |
| `houduan/database/mysql.go` | AutoMigrate 列表 + 表注释 |
| `houduan/repository/content_repo.go` | `FoodCategoryRepo`（`ListAll`/`GetByKey`/`UpdateFields`） |
| `houduan/service/food_category_enrich.go` | 维基素材抓取 + LLM 改写 + 图片聚合 + 落库 |
| `houduan/service/food_category.go` | 只读服务（按 key 取类别） |
| `houduan/service/food_category_enrich_test.go` | 纯函数单测（类别标签映射、图片去重、段落构建） |
| `houduan/cmd/food-category-enrich/main.go` | 批处理入口（`-only/-dry-run/-force`） |
| `houduan/handler/content_handler.go` | 新增 `FoodCategoryHandler.Get` |
| `houduan/handler/bootstrap.go`、`houduan/router/router.go`、`houduan/main.go` | 装配与路由注册 |
| `wennv/wenlv/src/api/content.ts` | `FoodCategoryItem` 类型 + `getFoodCategory(key)` |
| `wennv/wenlv/src/views/FoodDetail/FoodDetail.vue` | 类别分支改为拉类别数据渲染（保留原"本类美食"列表作为兜底） |
| `wennv/wenlv/src/views/FoodPage/FoodPage.vue` | 卡片映射调整：`cuisine/snacks/nightfood` 走类别页，`hotpot/chuanchuan/tea` 仍走菜品详情 |
| `wennv/wenlv/src/locales/{zh,en,ja}.ts` | 类别页 UI 文案（`overview`/`typicalDishes`/`dataSource` 等，已有部分可复用） |

---

### Task 0: 分支确认

- [ ] **Step 1**

```powershell
cd "d:\桌面\文旅\wenlv"; git branch --show-current
```

Expected: `feature/scenic-detail-data`（不要切分支、不要提交）。

---

### Task 1: `FoodCategory` 模型与迁移

**Files:** Create `houduan/model/food_category.go`；Modify `houduan/database/mysql.go`

**Interfaces:**
- Produces: `model.FoodCategory{ID, Key, NameZH, NameEN, Intro, Sections, GalleryImages, SourceURL, EstimatedFields, DataSource string; DataUpdatedAt *time.Time; CreatedAt, UpdatedAt time.Time}`；`model.CategorySection{Title, Text, Image string}`

- [ ] **Step 1: 模型文件**

```go
package model

import "time"

// FoodCategory 美食大类(川菜/名小吃/夜宵),对应前端美食名片里的大类卡片。
// 与 Food(具体菜品)区分:大类详情页介绍"这一类",不指向某道菜。
type FoodCategory struct {
	ID              uint       `gorm:"primaryKey;comment:美食大类ID" json:"id"`
	Key             string     `gorm:"size:32;uniqueIndex;comment:类别键(cuisine/snacks/nightfood)" json:"key"`
	NameZH          string     `gorm:"size:64;comment:类别中文名" json:"name_zh"`
	NameEN          string     `gorm:"size:64;comment:类别英文名" json:"name_en"`
	Intro           string     `gorm:"type:text;comment:类别概述(LLM依据维基素材改写,参考值)" json:"intro"`
	Sections        string     `gorm:"type:text;comment:类别图文段落JSON" json:"sections"`
	GalleryImages   string     `gorm:"type:text;comment:类别图集URL列表(逗号分隔,OSS)" json:"gallery_images"`
	SourceURL       string     `gorm:"size:512;comment:素材来源链接(维基条目)" json:"source_url"`
	EstimatedFields string     `gorm:"size:255;comment:参考值字段(逗号分隔)" json:"estimated_fields"`
	DataSource      string     `gorm:"size:64;comment:数据来源(如 wiki+llm+oss)" json:"data_source"`
	DataUpdatedAt   *time.Time `gorm:"comment:数据采集时间" json:"data_updated_at"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// CategorySection 类别介绍的一段(标题 + 正文 + 配图)。
type CategorySection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}
```

- [ ] **Step 2: 迁移与表注释**

`houduan/database/mysql.go`：`AutoMigrate(...)` 列表里加 `&model.FoodCategory{}`；`tableComments` 加 `"food_categories": "美食大类表(川菜/名小吃/夜宵)"`；`idComments` 加 `"food_categories": "美食大类ID"`。

- [ ] **Step 3: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"; gofmt -l model/food_category.go; go build ./...
```

Expected: 无输出。

---

### Task 2: 仓储 + 只读服务 + 接口

**Files:** Modify `houduan/repository/content_repo.go`、Create `houduan/service/food_category.go`、Modify `houduan/handler/content_handler.go`、`houduan/handler/bootstrap.go`、`houduan/router/router.go`、`houduan/main.go`

**Interfaces:**
- Produces:
  - `(*repository.FoodCategoryRepo) ListAll() ([]model.FoodCategory, error)`、`GetByKey(key string) (*model.FoodCategory, error)`、`UpdateFields(id uint, updates map[string]any) error`
  - `service.NewFoodCategoryService(repo *repository.FoodCategoryRepo) *FoodCategoryService`、`(s *FoodCategoryService) Get(key string) (*model.FoodCategory, error)`
  - `handler.NewFoodCategoryHandler(svc *service.FoodCategoryService) *FoodCategoryHandler`、方法 `Get(c *gin.Context)`
  - 路由：`api.GET("/food-category/:key", h.FoodCategory.Get)`（**放在 `/food/:id` 之前或使用不同前缀，避免与 `/food/:id` 冲突**）

- [ ] **Step 1: 仓储三个方法**（实现风格与 `FoodRepo.ListAll/UpdateFields` 完全一致）

- [ ] **Step 2: 只读服务**

```go
// FoodCategoryService 美食大类只读查询。
type FoodCategoryService struct{ repo *repository.FoodCategoryRepo }

func NewFoodCategoryService(repo *repository.FoodCategoryRepo) *FoodCategoryService {
	return &FoodCategoryService{repo: repo}
}

// Get 按类别键查询;不存在时返回 ErrNotFound。
func (s *FoodCategoryService) Get(key string) (*model.FoodCategory, error) {
	v, err := s.repo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrNotFound
	}
	return v, nil
}
```

- [ ] **Step 3: 处理器 + 路由 + 装配**（照抄 `FoodHandler.Get` 的写法：`parseUintParam` 不适用，这里取路径参数 `c.Param("key")`，空则 `pkg.BadRequest`；`handleNotFound(c, err)` 处理 404；成功 `pkg.OK(c, v)`）

- [ ] **Step 4: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"; gofmt -l ; go build ./...; go vet ./handler/ ./service/ ./repository/
```

Expected: 无输出。（接口实机验证放在 Task 5，控制器会重启后端后 curl。）

---

### Task 3: 类别内容采集（维基素材 → LLM 改写 → 多来源图片 → OSS）

**Files:** Create `houduan/service/food_category_enrich.go`、`houduan/service/food_category_enrich_test.go`、`houduan/cmd/food-category-enrich/main.go`

**Interfaces:**
- Consumes: `TripLLM.ChatWithEffort`、`fetchBytes`、`searchCommonsImagesFor`、`uploadImageCandidates`、`pkg.OssSigner`、`FoodRepo.ListAll`（取该类菜品图）、`FoodCategoryRepo`
- Produces:
  - `type CategorySeed struct{ Key, NameZH, NameEN, WikiTitle string; WikiTitles []string; CommonsKeywords []string }`
  - `func categorySeeds() []CategorySeed`（三个类别；`WikiTitle` 为主条目，`WikiTitles` 为备选：川菜 → `["川菜","川菜系"]`；名小吃 → `["成都小吃","四川小吃"]`；夜宵 → `["夜宵","消夜","成都夜市"]`）
  - `func fetchWikiArticle(ctx context.Context, client *http.Client, titles []string) (title, text, pageURL string, err error)`：Action API `action=query&prop=extracts&explaintext=1&variant=zh-cn&redirects=1&titles=<标题>`（**不带 `exintro`**，取全文），命中第一个存在的条目；`pageURL = "https://zh.wikipedia.org/wiki/" + url.PathEscape(title)`
  - `func newWikiHTTPClient() *http.Client`：与 `newCommonsHTTPClient` 同样支持 `COMMONS_PROXY`，为空时 `http.ProxyFromEnvironment`
  - `type FoodCategoryEnricher struct{...}`；`func NewFoodCategoryEnricher(catRepo *repository.FoodCategoryRepo, foodRepo *repository.FoodRepo, llm *TripLLM, signer *pkg.OssSigner) *FoodCategoryEnricher`
  - `func (e *FoodCategoryEnricher) Enrich(ctx context.Context, seed CategorySeed, opts FoodCategoryEnrichOptions) (*FoodCategoryEnrichResult, error)`
  - 纯函数：`func dishesForCategory(foods []model.Food, keywords []string) []model.Food`、`func collectCategoryImages(foods []model.Food, keywords []string, limit int) []string`、`func buildCategorySections(llmSections []model.CategorySection, images []string) []model.CategorySection`

- [ ] **Step 1: 写失败测试** `houduan/service/food_category_enrich_test.go`

覆盖：
1. `TestDishesForCategory`：构造 4 个 `model.Food`（标签分别含 `川菜`/`小吃`/`夜宵`/`面食`），断言 `dishesForCategory(foods, []string{"川菜"})` 只返回标签含"川菜"的那条（包含匹配，如标签 `一个人的火锅` 应能被 `火锅` 命中）。
2. `TestCollectCategoryImages`：菜品的 `gallery_images="a.jpg,b.jpg"`、`images="c.jpg"`，断言收集顺序为 `a.jpg,b.jpg,c.jpg` 且去重、受 `limit` 截断。
3. `TestBuildCategorySections`：3 段且 images 只有 2 张时，**每段都要有图**（不足则复用第 2 张）；段落缺 title/text 时被丢弃。

- [ ] **Step 2: 运行确认失败**：`go test ./service/ -count=1 -run 'TestDishesForCategory|TestCollectCategoryImages|TestBuildCategorySections' -v` → 编译失败。

- [ ] **Step 3: 实现 `food_category_enrich.go`**（关键要求）

```
Enrich 流程:
1) 抓素材: fetchWikiArticle(seed.WikiTitles) → (标题, 正文, 链接)。正文为空 → 返回结果并把 Note 记为"未取到维基素材",不写库(避免编造)。
2) 抽图: dishesForCategory(foodRepo.ListAll(), 标签关键词) → collectCategoryImages(...) 取该类菜品 OSS 图(前 6 张);
   不足 6 张时用 searchCommonsImagesFor(按 seed.CommonsKeywords 逐个搜) 补,过滤规则同景点侧;
   合并去重后 uploadImageCandidates(..., "foodcat/"+seed.Key, 候选, 6) 上传 OSS,得到最终 URL 列表。
   —— 注意:菜品图已是 OSS URL,不要重复上传,只上传 Commons 新图。
3) LLM 改写:提示词必须包含"仅可依据下方素材,不得引入素材之外的事实/数字/年份/排名" + 素材全文(截断到 6000 字以内) + 输出 JSON:
   {"intro":"概述(150-250字)","sections":[{"title":"小标题(4-10字)","text":"段落(80-150字)"}]}
   要求:概述之外写 3-4 段;使用简体中文;不要 Markdown。maxTokens 建议 2000,effort 用 llmThinkingLevel()。
4) 段落配图: buildCategorySections(parsed.Sections, images)。
5) 落库(仅当本轮真的生成成功): updates = {name_zh,name_en,intro,sections(JSON),gallery_images,source_url,estimated_fields:"intro,sections",data_source:"wiki+llm+oss",data_updated_at}
   —— 非 Force 模式下不覆盖已有非空值;无实际更新时不写库(复用 hasBusinessUpdate)。
6) 结果结构 FoodCategoryEnrichResult{Key, WikiTitle, ImageCount int, LLMUsed bool, UpdatedFields []string, Note string}
```

- [ ] **Step 4: 实现 CLI `cmd/food-category-enrich/main.go`**（骨架照抄 `cmd/food-enrich/main.go`：`godotenv.Load()` → `config.Load()` → `mustDB(cfg, !*dryRun)` → `TripSettings` + `AttachDB` → `llm`/`signer` → 仓储 → 遍历三个 seed → 打印汇总；flags：`-only`（按 key 过滤）、`-dry-run`（只抓维基素材并打印标题/正文长度/该类图片张数，不调 LLM、不写库、不上传）、`-force`）

- [ ] **Step 5: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
gofmt -l service/food_category_enrich.go service/food_category_enrich_test.go cmd/food-category-enrich/main.go
go build ./...; go vet ./service/ ./cmd/food-category-enrich/
go test ./service/ -count=1 -run 'TestDishesForCategory|TestCollectCategoryImages|TestBuildCategorySections' -v
go run ./cmd/food-category-enrich -dry-run
```

Expected: gofmt/构建/vet 无输出；3 个测试 PASS；dry-run 打印三个类别的**维基标题 + 正文长度 + 该类图片张数**（例如 `cuisine 川菜 正文字数=xxxxx 图片=6`），且不写库、不调 LLM。

---

### Task 4: 前端类别页

**Files:** Modify `wennv/wenlv/src/views/FoodDetail/FoodDetail.vue`、`wennv/wenlv/src/views/FoodPage/FoodPage.vue`、`wennv/wenlv/src/api/content.ts`、`wennv/wenlv/src/locales/{zh,en,ja}.ts`

**Interfaces:**
- `content.ts`：新增
```ts
export interface CategorySection { title: string; text: string; image?: string }
export interface FoodCategoryItem {
  id: number; key: string; name_zh: string; name_en?: string; intro?: string
  sections?: string; gallery_images?: string; source_url?: string
  estimated_fields?: string; data_source?: string; data_updated_at?: string
}
export function getFoodCategory(key: string): Promise<FoodCategoryItem> {
  return get<FoodCategoryItem>(`/food-category/${key}`)
}
```
- `FoodPage.vue`：卡片映射改为——`hotpot`/`chuanchuan`/`tea` 仍解析菜品 id 直达 `/food/{id}`；`cuisine`/`snacks`/`nightfood` 直接 `router.push({ name: 'food-detail', params: { id: card.key } })`（走类别页）。保留"解析不到就回退 key"的兜底与防连点逻辑。

- [ ] **Step 1: 类别分支改为数据驱动**

`FoodDetail.vue` 中 `isFallback`（非数字 id）分支：
1. 用 `getFoodCategory(foodId)` 拉类别数据（失败时回退到现有 i18n 文案 + 本类美食列表，保证不白屏）；
2. 渲染：hero（类别名 + 英文名 + `intro` 作为副标题下方概述，`estimated` 命中 `intro` 时加"参考值"徽标）→「类别概述」区块（`foodDetail.overview` 文案 + `intro`）→ 图文段落（`parseSections(category.sections)`，左右交替，每段标题+正文+配图）→「美味瞬间」图集（`splitList(category.gallery_images)`）→「本类美食」列表（沿用现有 `categoryDishes`，点击进真实菜品）→ 页脚数据来源（`foodDetail.dataSource` + `source_url` 标注"内容素材来源：维基百科"）；
3. 类别正文为空时该区块显示 `noData`；
4. 不要渲染菜品专属区块（评分/辣度/人均/招牌/场景/风味故事/寻味地图/相关推荐）——这条已有，保持。

- [ ] **Step 2: 三语文案补齐**：`foodDetail` 段新增/复用 `categoryOverview: '类别概述'`、`categoryStories: '类别故事'`、`typicalDishes: '本类美食'`、`contentSource: '内容素材来源：维基百科（CC BY-SA）'`（en/ja 同义）。

- [ ] **Step 3: 验证**

```powershell
cd "d:\桌面\文旅\wenlv\wennv\wenlv"; npm run type-check; npx vitest run
```

Expected: 类型检查无错误；既有测试全部通过。

---

### Task 5: 采集 + 端到端验收

- [ ] **Step 1: 全量采集三个类别**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"; go run ./cmd/food-category-enrich 2>&1 | Tee-Object -FilePath ../.superpowers/sdd/food-category-enrich.log
```

Expected: 三个类别都成功（`完成: 成功 3 / 失败 0`），日志显示各自维基标题与图片张数。

- [ ] **Step 2: 数据校验**

```sql
SELECT key, name_zh, LENGTH(intro) AS intro_len, JSON_LENGTH(sections) AS secs,
 (LENGTH(gallery_images)-LENGTH(REPLACE(gallery_images,',',''))+1) AS imgs,
 estimated_fields, data_source, source_url FROM food_categories;
SELECT COUNT(*) AS bad_img FROM food_categories WHERE gallery_images<>'' AND gallery_images NOT LIKE '%aliyuncs.com%';
SELECT COUNT(*) AS bad_json FROM food_categories WHERE sections<>'' AND JSON_VALID(sections)=0;
SELECT COUNT(*) AS bad_label FROM food_categories WHERE (intro<>'' AND estimated_fields NOT LIKE '%intro%') OR (sections<>'' AND estimated_fields NOT LIKE '%sections%');
```

Expected: 3 行，`intro_len` 150–300、`secs` 3–4、`imgs` 4–6；三个 bad_* 均为 0。

- [ ] **Step 3: 浏览器验收**（控制器执行）

打开 `/food/cuisine`、`/food/snacks`、`/food/nightfood`：
1. hero 显示类别名（川菜/名小吃/夜宵）+ 英文名 + 概述（带"参考值"）；
2. 类别概述与 3–4 段图文（每段有配图）正常显示；
3. 美味瞬间 4–6 张图（OSS 域名）正常显示；
4. 本类美食列表正确（川菜 3 / 名小吃 9 / 夜宵 ≥1），点击进入真实菜品详情；
5. 页脚显示数据来源与"内容素材来源：维基百科"；
6. 控制台无 error、无未翻译 key、无"数据加载中"残留。

同时确认 `火锅`/`串串香`/`盖碗茶` 三张卡片**仍然直达菜品详情**（不回归）。

- [ ] **Step 4: 汇总**

```powershell
git status --short
```

Expected: 只有源码改动；**不提交**。

---

## 风险与合规

- **版权**：维基百科正文为 CC BY-SA，页面已标注"内容素材来源：维基百科"并给出来源链接；LLM 仅做压缩改写，不逐字复制长段落。
- **精确性**：图片优先取该类下已采集菜品的 OSS 图（必然属于该类）；Commons 图要求文件名含关键词，且在报告中打印实际选中的文件名以便人工核对。
- **夜宵类**：维基条目为通用"夜宵/消夜"定义，成都本地特色主要靠该类菜品（兔头等）与 Commons 图体现；若素材过泛，宁可段落少而准，不编造成都的具体店铺与数字。
- **小红书 / 百度百科**：本次不启用（Cookie 失效 / 版权与结构风险），已在「数据来源与精确性」中记录，后续可单开任务。
