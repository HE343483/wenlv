# 主站公开页 Hero / 轮播配图 + 后端联调地址

**日期：** 2026-09-18  
**范围：** 仅主站公开首页（`/`），不含登录后内部页、不含 AI 行程视觉统一  
**状态：** 设计已口头确认，待用户审阅本 spec 后进入实现计划

---

## 1. 背景与目标

公开首页 Hero 与景点轮播目前无真实摄影（轮播 `imageUrl` 为空，仅靠渐变占位），观感偏空壳。同时本地 Vite 代理仍指向 `localhost:8081`，需改为内网穿透后端以便联调。

**成功标准：**
1. `/` Hero 为清晰成都风景全幅图，标题与 CTA 可读。
2. 轮播 5 张图与景点一一对应，清晰、无明显拉伸模糊。
3. 开发环境下 `/api/*` 代理到 `http://ra9a3c94.natappfree.cc`。
4. 桌面与手机宽度下 Hero / 轮播不破版。

---

## 2. 方案选型

采用 **方案 1：本地静态图 + 轻改公开首页**。

| 方案 | 结论 |
|------|------|
| 1. 本地 `public/images/home/` + 轻改 Hero/Carousel + 改代理 | **采用** |
| 2. 外链图床 | 不采用（稳定性/外网依赖） |
| 3. 方案 1 + 登录页配图 | 本期不做，避免扩 scope |

**Hero 布局：** 全幅背景图 + 半透明遮罩，保留现有文案与 CTA（用户选定 A）。

**图片来源：** 公开可商用高清图，下载压缩后放入仓库（用户选定 A）。

---

## 3. 资源规划

目录：`public/images/home/`

| 文件 | 用途 |
|------|------|
| `hero-chengdu.jpg` | Hero 全幅成都风景 |
| `carousel-panda.jpg` | 大熊猫繁育研究基地 |
| `carousel-kuanzhai.jpg` | 宽窄巷子 |
| `carousel-dujiangyan.jpg` | 都江堰 · 青城山 |
| `carousel-jinli.jpg` | 锦里 · 武侯祠 |
| `carousel-xiling.jpg` | 西岭雪山 |

- 引用路径：`/images/home/<filename>`
- 体积目标：单张约 200–500KB，兼顾清晰度与加载

---

## 4. UI 行为

### 4.1 Hero（`HomeView.vue`）

- 在 `hero__bg` 增加全幅背景：`cover`、居中。
- 保留「蜀」水印、badge、中英标题、副标题、主 CTA。
- 遮罩：深色半透明 + 底部淡入 `--color-bg`，保证文字对比度，并与下方页面衔接。
- `min-height` 保持全屏；实现时可改为 `100dvh` 改善移动端视口跳动。
- 桌面/手机共用同一张 Hero 图。

### 4.2 轮播（`Carousel.vue` + `HomeView` 数据）

- `carouselItems` 五条写入对应 `imageUrl`。
- 有 `imageUrl` 时渲染真实图片；无图时保留现有渐变占位（兜底）。
- 标题/副标题叠在图上；遮罩改为下深上浅的深色渐变，避免浅米白盖住风景。
- 自动播放、箭头、菱形指示器逻辑不变。
- 图片提供有意义的 `alt`（景点名）。

### 4.3 明确不做

- 文化名片 / 时间线 / 非遗区配图
- 登录页、内部页、行程模块视觉重构
- 换字体、换主色、重构导航、挂载 BottomNav

---

## 5. 后端联调

| 位置 | 改动 |
|------|------|
| `vite.config.ts` | `server.proxy['/api'].target` → `http://ra9a3c94.natappfree.cc` |
| `.env.example`（及本地 `.env` 若存在） | `VITE_API_BASE_URL=http://ra9a3c94.natappfree.cc` |

主站 `src/api/request.ts` 继续使用 `VITE_API_BASE || '/api'`，开发时走 Vite 代理，避免浏览器直连跨域。Trip 模块可读 `VITE_API_BASE_URL` 直连同一穿透地址。

**说明：** natapp 免费域名可能变更；地址变更时只改代理 / env，不必改页面组件。

---

## 6. 改动文件清单

| 文件 | 改动 |
|------|------|
| `public/images/home/*` | 新增 6 张图 |
| `src/views/HomeView/HomeView.vue` | Hero 背景图与遮罩；轮播 `imageUrl` |
| `src/components/Carousel.vue` | 渲染图片 + 深色渐变遮罩 |
| `vite.config.ts` | 代理 target |
| `.env.example`（+ 可选 `.env`） | API 基址 |

---

## 7. 错误处理与降级

- 单张图 404：该槽位显示现有占位渐变，不白屏。
- 后端穿透失效：前端页面与配图仍可用；仅 API 调用失败（与本期视觉目标分离）。

---

## 8. 测试计划

1. `npm run dev` 打开 `/`，目视 Hero 与轮播清晰度、文字对比度。
2. 切换中/英，确认标题叠字正常。
3. 缩放到手机宽度，检查 Hero 全幅与轮播 `aspect-ratio` 不破版。
4. 浏览器 Network：任意 `/api/...` 请求目标主机为 `ra9a3c94.natappfree.cc`（或经代理转发成功）。

---

## 9. 风险

| 风险 | 缓解 |
|------|------|
| 图片版权 | 选用明确可商用/许可清晰的公开图源，必要时在目录旁注来源 |
| 仓库体积增大 | 压缩至目标体积；不引入超大原图 |
| natapp 地址失效 | 仅配置层变更，组件不硬编码穿透域名 |
