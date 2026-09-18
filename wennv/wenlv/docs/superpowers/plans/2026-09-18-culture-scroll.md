# 文明脉络 · 平面透视建筑长卷 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将首页「文明脉络」从竖向时间线卡片替换为可滚动/拖动展卷的 2D 平面透视建筑长卷：三层视差、小人跟镜走步、画中热点展示时代文案。

**Architecture:** 新建 `CultureScroll.vue` 自包含舞台；外层用加高滚动轨 + `position: sticky` 将竖滚映射为横移进度 `progress∈[0,1]`；远/中/近三层以不同系数 `translateX`；小人固定在视口中下，按进度方向播放 CSS 走步。热点挂在中景，点击弹出文案（数据来自现有 timeline）。**不引入 GSAP/Three.js**；首期 SVG 用几何简笔拼出可读天际，可后续换精细稿。

**Tech Stack:** Vue 3 + TypeScript；纯 CSS/SVG；Vitest + `@vue/test-utils`（挂载与 reduced-motion 冒烟）；无新运行时依赖。

**Spec:** `docs/superpowers/specs/2026-09-18-culture-scroll-design.md`

## Global Constraints

- 2D 平面插画，**允许透视**；禁止 3D 模型 / WebGL 场景引擎。
- 操控：**滚动/拖动展卷，小人自动跟镜头**（非 WASD）。
- 主视觉是连续长卷，禁止抽象点线时间表作主角。
- `prefers-reduced-motion: reduce` → 取消 sticky 强绑定，改为横向 `overflow-x: auto`；小人静态。
- 热点文案对齐现有 timeline 六段（古蜀→秦并巴蜀→蜀汉→唐宋→明清→现代成都）。
- 未经用户明确要求不要 `git commit`（计划内 Commit 步骤默认跳过）。

---

## File Map

| 文件 | 职责 |
|------|------|
| `src/components/CultureScroll.vue` | 长卷舞台：进度、三层、小人、拖拽、热点浮层 |
| `src/components/CultureScroll.spec.ts` | 挂载冒烟 + reduced-motion 类名/模式 |
| `src/components/culture-scroll/eras.svg` 或内联 SVG 片段 | 远/中/近景简笔内容（可拆多个） |
| `src/data/cultureScroll.ts` | 区段与热点数据（从 HomeView timeline 抽出） |
| `src/views/HomeView/HomeView.vue` | 替换文明脉络区为 `<CultureScroll />`，删旧 track 样式依赖 |

---

### Task 1: 数据抽出 + `CultureScroll` 滚动壳（三层占位）

**Files:**
- Create: `src/data/cultureScroll.ts`
- Create: `src/components/CultureScroll.vue`
- Create: `src/components/CultureScroll.spec.ts`

**Interfaces:**
- Consumes: 无
- Produces:
  - `CultureScrollSegment`：`{ id, era, eraEn, period, periodEn, desc, descEn, progressStart, progressEnd }`
  - `CULTURE_SCROLL_SEGMENTS: CultureScrollSegment[]`（六段，progress 区间均分约 `i/6 … (i+1)/6`）
  - `CultureScroll.vue` props：无必填；暴露根 class `culture-scroll`
  - 内部 `progress` ref `0…1`；`reducedMotion` 时根带 `culture-scroll--static`

- [ ] **Step 1: 写失败单测**

```ts
// src/components/CultureScroll.spec.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import CultureScroll from './CultureScroll.vue'

describe('CultureScroll', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn().mockImplementation((query: string) => ({
        matches: query.includes('prefers-reduced-motion'),
        media: query,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn(),
        onchange: null,
      })),
    )
  })
  afterEach(() => vi.unstubAllGlobals())

  it('reduced-motion: root has culture-scroll--static', () => {
    const wrapper = mount(CultureScroll)
    expect(wrapper.find('.culture-scroll').classes()).toContain('culture-scroll--static')
  })
})
```

- [ ] **Step 2: 跑测确认 RED**

```powershell
cd e:\wenlv\wennv\wenlv
npm run test:unit -- src/components/CultureScroll.spec.ts
```

Expected: FAIL（组件不存在）

- [ ] **Step 3: 实现 `cultureScroll.ts` 数据**

从 `HomeView` 现有 `timeline` 数组迁移字段，并加：

```ts
progressStart: number  // 0..1
progressEnd: number
```

六段均分即可（后续可手工调）。

- [ ] **Step 4: 实现 `CultureScroll.vue` 壳**

结构要点：

```vue
<section
  ref="railRef"
  class="culture-scroll"
  :class="{ 'culture-scroll--static': reducedMotion }"
  :style="reducedMotion ? undefined : { height: railHeightCss }"
>
  <div ref="stageRef" class="culture-scroll__stage">
    <div class="culture-scroll__world" :style="worldStyle">
      <div class="culture-scroll__layer culture-scroll__layer--far" :style="layerStyle(0.35)" />
      <div class="culture-scroll__layer culture-scroll__layer--mid" :style="layerStyle(1)">
        <!-- Task 3 内容；Task 1 用色块占位 width ≈ 600vw -->
      </div>
      <div class="culture-scroll__layer culture-scroll__layer--near" :style="layerStyle(1.25)" />
    </div>
    <div class="culture-scroll__walker" aria-hidden="true" />
  </div>
</section>
```

行为：
- `reducedMotion === false`：`stage` 为 `position: sticky; top: 0; height: 100vh`；`rail` 高度约为 `worldWidth - viewport + 100vh`（或 `progress` 用 `scrollY` 在 rail 内归一化）。
- `onScroll` / `ResizeObserver`：计算 `progress = clamp((scrollY - railTop) / (railHeight - viewport), 0, 1)`。
- `layerStyle(factor)`：`transform: translate3d(${-progress * maxTravel * factor}px, 0, 0)`；`maxTravel = worldWidth - viewportWidth`。
- **拖拽：** 在 stage 上 `pointerdown/move/up`：把 `deltaX` 换成对 `window.scrollBy` 或直接改目标 scroll（优先 `window.scrollTo` 映射到 rail 进度），避免与滚轮打架——拖拽时 `preventDefault` 仅在水平主导手势。
- Task 1 层内容可用渐变色带 + 区段标签文字占位，证明横移与视差。

- [ ] **Step 5: 跑测 GREEN**

```powershell
npm run test:unit -- src/components/CultureScroll.spec.ts
```

Expected: PASS

- [ ] **Step 6: Commit**（默认跳过）

---

### Task 2: 小人跟镜走步

**Files:**
- Modify: `src/components/CultureScroll.vue`
- Modify: `src/components/CultureScroll.spec.ts`（可选：断言 walker 存在）

**Interfaces:**
- Consumes: `progress` from Task 1
- Produces: `.culture-scroll__walker` 带 `data-walking` / `data-dir="left|right"`

- [ ] **Step 1: 扩展单测（可选冒烟）**

```ts
it('renders walker', () => {
  vi.stubGlobal('matchMedia', vi.fn().mockImplementation(() => ({
    matches: false,
    media: '',
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
    onchange: null,
  })))
  const wrapper = mount(CultureScroll)
  expect(wrapper.find('.culture-scroll__walker').exists()).toBe(true)
})
```

- [ ] **Step 2: RED 后实现 walker**

- 固定于 stage：`position: absolute; left: 42%; bottom: 18%;`（勿随 world 横移）。
- 用纯 CSS 剪影（圆头 + 躯干矩形即可）或内联 SVG；`transform: scaleX(-1)` 按滚动方向。
- 检测 `progress` 变化率：`|Δp| > ε` 时 `data-walking="true"`，否则 false；方向取 `Δp` 符号（progress↑ = 向右展卷 = 人朝右）。
- `@keyframes walk-bob` 轻微上下/腿交换（两条细线腿 rotate）。
- `reducedMotion`：无 animation，无 `data-walking`。

- [ ] **Step 3: 跑测 PASS + 手动：滚长卷时小人原地踏步、方向对**

- [ ] **Step 4: Commit**（默认跳过）

---

### Task 3: 平面透视简笔建筑长卷（SVG）

**Files:**
- Create: `src/components/culture-scroll/ScrollArtwork.vue`（或同等）
- Modify: `CultureScroll.vue` 中景/远景/近景挂载 artwork

**Interfaces:**
- Consumes: `CULTURE_SCROLL_SEGMENTS`
- Produces: 连续 SVG 宽约 `min 5000px` 或 `600vw`；六段视觉可辨；**平面+透视线条**，黑色描边为主

- [ ] **Step 1: 绘制/拼装简笔**

每段至少一个可识别剪影（不必写实）：
1. 古蜀：高台 + 神树简化折线  
2. 秦蜀：堰渠折线 + 水纹  
3. 蜀汉：阙楼  
4. 唐宋：连片坡屋顶  
5. 明清：街巷进深（一点透视）  
6. 现代：竖向矩形楼群剪影  

远景：淡山连续带；近景：路面水平线 + 偶发树/桥栏。段与段用树丛或江岸衔接，避免硬切。

- [ ] **Step 2: 接入三层**；确认无 Three.js / canvas 3D。

- [ ] **Step 3: 目视验收时代演进可读**

- [ ] **Step 4: Commit**（默认跳过）

---

### Task 4: 热点 + 接入 HomeView

**Files:**
- Modify: `CultureScroll.vue`（热点按钮 + 浮层）
- Modify: `src/views/HomeView/HomeView.vue`
- 删除或停用 HomeView 内联 `timeline` 数组与 `.culture-timeline__*` 大段样式（改由组件负责）

**Interfaces:**
- Hotspot：`{ segmentId, xPercent, label? }` 相对中景宽度定位
- 点击：`dialog`/`role="dialog"` 展示 era + period + desc；Esc / 遮罩关闭

- [ ] **Step 1: 热点至少 4 个**（建议六段各一），`position: absolute; left: …%` 叠在 mid 层。

- [ ] **Step 2: HomeView 替换**

```vue
<section
  :ref="(el) => setSectionRef(el, 2)"
  class="culture-scroll-section reveal"
>
  <div class="culture-scroll-section__header container">
    <h2 class="section-title">{{ langStore.t('culture.timelineTitle') }}</h2>
    <p class="section-subtitle">{{ langStore.t('culture.timelineSubtitle') }}</p>
  </div>
  <CultureScroll />
</section>
```

`import CultureScroll from '@/components/CultureScroll.vue'`  
移除旧 `timeline` const 与旧 track 模板。

- [ ] **Step 3: 清理无用 CSS**（`.culture-timeline` 相关）

- [ ] **Step 4: 手动验收 Spec §7**

- [ ] **Step 5: Commit**（默认跳过）

---

### Task 5: 降级打磨与回归测试

**Files:**
- Modify: `CultureScroll.vue`（static 模式横向滑动）
- Modify: specs if needed

- [ ] **Step 1: static 模式**

`culture-scroll--static`：`rail` 高度 auto；`stage` 非 sticky；`world` 包在 `overflow-x: auto; -webkit-overflow-scrolling: touch`；禁用竖滚映射与拖拽改 scroll。

- [ ] **Step 2: 跑全相关单测**

```powershell
npm run test:unit -- src/components/CultureScroll.spec.ts
npm run type-check
```

- [ ] **Step 3: 窄屏拖滑 / reduced-motion 系统开关抽查**

- [ ] **Step 4: Commit**（默认跳过）

---

## Self-Review

| Spec 要求 | Task |
|-----------|------|
| DOM/SVG 分层 + 滚动 | 1 |
| 小人跟镜 | 2 |
| 平面+透视建筑长卷 | 3 |
| 热点文案 / 替换 HomeView | 4 |
| reduced-motion / 窄屏 | 1+5 |
| 无 3D / 无抽象主轴 | 全局约束 |

无 TBD；组件名 `CultureScroll` 前后一致。

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-09-18-culture-scroll.md`.

**两种执行方式：**

1. **Subagent-Driven（推荐）** — 每 Task 独立子代理 + 复核  
2. **Inline Execution** — 本会话连续执行  

你选哪种？
