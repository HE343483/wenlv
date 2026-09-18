# 文明脉络 · 黄蓝底 + 黑线 + 对立图片卡

**日期：** 2026-09-18  
**范围：** `CultureScroll.vue` 中景视觉与时间轴 marker 布局；`cultureScroll` 数据增补 `imageUrl`  
**状态：** 口头确认（方案 1 + 配图表）；待用户审阅本 spec 后进入实现计划  
**前置：**
- [2026-09-18-culture-scroll-design.md](./2026-09-18-culture-scroll-design.md)
- [2026-09-18-culture-scroll-timeline-design.md](./2026-09-18-culture-scroll-timeline-design.md)

---

## 1. 背景与目标

当前中景叠了黄蓝山水 JPG + 透明黑线；时间轴仅有文案卡。用户要求：

1. **背景回到黄蓝色块**（舞台 CSS 渐变），去掉山水实底图。
2. **简笔插画只留黑线**（透明 PNG）。
3. 在时间轴文案卡的 **对立面** 增加对应时代的 **图片卡**（文上则图下，文下则图上）。

**成功标准：**
1. 无 `mid-scenery` / `SCENERY_SRC` 渲染；舞台为黄蓝纸色渐变。
2. 黑线 `LINEART_SRC` 仍铺在中景并随卷移动。
3. 每个区段：一侧完整文案卡，对侧图片卡；均不可点。
4. 六段 `imageUrl` 使用下方已确认路径；中英切换不影响图片。
5. 现有时间轴单测更新后仍通过。

---

## 2. 方案选型（已确认）

| 项 | 选择 |
|----|------|
| 背景 | 去掉山水 JPG；保留 stage 黄蓝渐变 + 远景淡山 SVG |
| 插画 | 仅透明黑线 PNG |
| 图片卡位置 | 与文案卡 **对立**（`--above` 文案 → 图片在下；`--below` 文案 → 图片在上） |
| 图源 | 项目内现有资源，按下表（按画面内容选用，不拘文件名） |

不采用：图嵌文案卡内、本期另生成六张专图。

---

## 3. 时代配图（已确认）

| segmentId | 路径 |
|-----------|------|
| `ancient-shu` | `/images/home/1.jpg` |
| `qin` | `/images/home/carousel-xiling.jpg` |
| `shu-han` | `/images/home/hero-chengdu.jpg` |
| `tang-song` | `/images/home/154b682806f848688077dbabd5bcac3f_720.jpg` |
| `ming-qing` | `/images/culture-scroll/era-scenery-shanshui-v1.jpg` |
| `modern` | `/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg` |

说明：部分 `carousel-*` 文件名与实图不符，以本表路径为准。日后可替换为更贴题图而不改交互。

---

## 4. 视觉与布局

### 4.1 背景与黑线

- 删除 `.culture-scroll__scenery-track` 及 `SCENERY_SRC` / `LINEART_TILES` 中仅服务底图的重复逻辑（黑线仍可 4 段平铺）。
- `.culture-scroll__stage` 保持既有黄蓝/纸色径向+线性渐变。
- 黑线保持现有略下移（`translateY` / `object-position`），叠在色块之上、时间轴之下。

### 4.2 Marker 结构

每个 `culture-scroll__marker`：

- 轴：圆点 + `axis-time`（不变）
- **文案卡** `.culture-scroll__era-card`：位置规则不变（even 上 / odd 下）
- **图片卡** `.culture-scroll__era-media`：与文案卡对立
  - `--above`（文在上）：media 在轴下方
  - `--below`（文在下）：media 在轴上方
- 图片：`<img>`，`object-fit: cover`；宽约与文案卡同（`min(280px, 22vw)`）；高约 `140px`（窄屏略减）
- `pointer-events: none`；装饰性 `alt=""`（文案已在旁卡）

### 4.3 数据

在 `CultureScrollSegment`（或 hotspot 关联字段）增加：

```ts
imageUrl: string
```

写入 `TIMELINE_FIELDS` / segments，与上表一致。组件用 `segment.imageUrl` 渲染，不硬编码路径在模板里散落。

---

## 5. 技术改动面

| 文件 | 改动 |
|------|------|
| `src/data/cultureScroll.ts` | 各段加 `imageUrl` |
| `src/components/CultureScroll.vue` | 去 scenery；marker 加 media；样式 |
| `src/components/CultureScroll.spec.ts` | 断言无 scenery；有 6 张 media 且 src 正确；对立 class 关系 |

---

## 6. 明确不做

- 本期不新绘/生成时代专图
- 图片卡不点击、不弹层
- 不恢复 near/road/walker
- 不改 segment 历史文案正文

---

## 7. 验收清单

- [ ] 中景无山水 JPG 底，仅黄蓝色块 + 黑线
- [ ] 六段文案卡与图片卡对立展示
- [ ] 图片 src 与上表一致
- [ ] 展卷时轴/文/图随中景移动
- [ ] 单测更新并通过

---

## 8. 实现顺序（供后续 plan）

1. 数据层加 `imageUrl`
2. 去掉 scenery；模板加 era-media
3. 对立布局 CSS
4. 更新单测与目视验收
