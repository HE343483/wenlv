# Hero 背景水波纹 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在公开首页 Hero 照片层加入轻柔 Canvas 水面折射波纹：桌面鼠标跟手 + 两端偶尔自动涟漪；手机仅自动；标题与遮罩不被扭曲。

**Architecture:** 将波纹算法抽成纯函数模块 `waterRipple.ts`（可单测）；`HeroRipple.vue` 负责加载图片、canvas 生命周期、pointer/自动扰动、IntersectionObserver / visibility / reduced-motion 降级。`HomeView` 在 `hero__bg` 中于静态 `hero__photo` 之上挂载组件，其上仍叠渐变与文案。

**Tech Stack:** Vue 3 + TypeScript；Canvas 2D `ImageData`；Vitest（算法单测 + 组件挂载冒烟）；无第三方涟漪库。

**Spec:** `docs/superpowers/specs/2026-09-18-hero-water-ripple-design.md`

## Global Constraints

- 范围仅公开首页 `/` 的 Hero 背景；不改登录页 / 内部页。
- 强度：轻柔克制；桌面 pointer 跟手 + 自动；手机仅自动。
- `prefers-reduced-motion: reduce` → 不启动动画，静态图兜底。
- Hero 不可见或 `document.hidden` → 暂停 rAF / 自动定时器。
- 波纹只作用于照片层；渐变、文案、「蜀」水印不扭曲。
- 未经用户明确要求不要 `git commit`（本计划勾选的 Commit 步骤默认跳过）。

---

## File Map

| 文件 | 职责 |
|------|------|
| `src/utils/waterRipple.ts` | 振幅双缓冲、disturb、step（写出扭曲后的 ImageData） |
| `src/utils/waterRipple.spec.ts` | 算法单测：扰动后像素变化、衰减、边界 |
| `src/components/HeroRipple.vue` | Canvas 挂载、图片加载、交互与生命周期 |
| `src/components/HeroRipple.spec.ts` | 挂载冒烟 + reduced-motion 不启 canvas 动画标记 |
| `src/views/HomeView/HomeView.vue` | 接入 HeroRipple，叠层样式 |

---

### Task 1: 纯算法模块 `waterRipple.ts`

**Files:**
- Create: `src/utils/waterRipple.ts`
- Create: `src/utils/waterRipple.spec.ts`

**Interfaces:**
- Consumes: 无
- Produces:
  - `createRippleState(width: number, height: number): RippleState`
  - `disturb(state: RippleState, x: number, y: number, radius?: number, strength?: number): void`
  - `stepRipple(state: RippleState, texture: Uint8ClampedArray, output: Uint8ClampedArray): boolean`
    - 返回本帧是否仍有明显扰动（便于空闲时降频；实现可简化为始终 `true` 若不想优化）
  - 轻柔默认：`DEFAULT_RADIUS = 2`，`DEFAULT_STRENGTH = 256`（远低于常见 demo 的 512–4048）

- [ ] **Step 1: 写失败单测**

```ts
// src/utils/waterRipple.spec.ts
import { describe, it, expect } from 'vitest'
import { createRippleState, disturb, stepRipple } from './waterRipple'

function solidTexture(w: number, h: number, rgba: [number, number, number, number]) {
  const data = new Uint8ClampedArray(w * h * 4)
  for (let i = 0; i < w * h; i++) {
    data[i * 4] = rgba[0]
    data[i * 4 + 1] = rgba[1]
    data[i * 4 + 2] = rgba[2]
    data[i * 4 + 3] = rgba[3]
  }
  return data
}

describe('waterRipple', () => {
  it('disturb + step changes center-ish pixels vs flat texture', () => {
    const w = 32
    const h = 32
    const state = createRippleState(w, h)
    const texture = solidTexture(w, h, [10, 20, 30, 255])
    // 给纹理一个水平色带，便于检测偏移
    for (let x = 0; x < w; x++) {
      const i = ((h / 2 | 0) * w + x) * 4
      texture[i] = 200
    }
    const out = new Uint8ClampedArray(texture)
    disturb(state, 16, 16, 3, 512)
    stepRipple(state, texture, out)
    stepRipple(state, texture, out)
    let changed = 0
    for (let i = 0; i < out.length; i++) if (out[i] !== texture[i]) changed++
    expect(changed).toBeGreaterThan(0)
  })

  it('without disturb, step keeps output equal to texture', () => {
    const w = 16
    const h = 16
    const state = createRippleState(w, h)
    const texture = solidTexture(w, h, [1, 2, 3, 255])
    const out = new Uint8ClampedArray(w * h * 4)
    stepRipple(state, texture, out)
    expect(Array.from(out)).toEqual(Array.from(texture))
  })
})
```

- [ ] **Step 2: 跑测确认失败**

```powershell
cd e:\wenlv\wennv\wenlv
npm run test:unit -- src/utils/waterRipple.spec.ts
```

Expected: FAIL（模块不存在或导出缺失）

- [ ] **Step 3: 实现算法**

```ts
// src/utils/waterRipple.ts
export const DEFAULT_RADIUS = 2
export const DEFAULT_STRENGTH = 256

export type RippleState = {
  width: number
  height: number
  rippleMap: Int16Array
  lastMap: Int16Array
  oldIndex: number
  newIndex: number
}

export function createRippleState(width: number, height: number): RippleState {
  const size = width * (height + 2) * 2
  return {
    width,
    height,
    rippleMap: new Int16Array(size),
    lastMap: new Int16Array(width * height),
    oldIndex: width,
    newIndex: width * (height + 3),
  }
}

export function disturb(
  state: RippleState,
  x: number,
  y: number,
  radius = DEFAULT_RADIUS,
  strength = DEFAULT_STRENGTH,
): void {
  const { width, height, rippleMap, oldIndex } = state
  const xi = x | 0
  const yi = y | 0
  for (let j = yi - radius; j < yi + radius; j++) {
    for (let k = xi - radius; k < xi + radius; k++) {
      if (j >= 0 && j < height && k >= 0 && k < width) {
        rippleMap[oldIndex + j * width + k] += strength
      }
    }
  }
}

/** 经典双缓冲振幅 + 按振幅差折射采样；返回是否仍有非零振幅差 */
export function stepRipple(
  state: RippleState,
  texture: Uint8ClampedArray,
  output: Uint8ClampedArray,
): boolean {
  const { width, height, rippleMap, lastMap } = state
  let i = state.oldIndex
  state.oldIndex = state.newIndex
  state.newIndex = i

  i = 0
  let mapIndex = state.oldIndex
  let active = false
  const halfW = width >> 1
  const halfH = height >> 1

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const data =
        ((rippleMap[mapIndex - width] +
          rippleMap[mapIndex + width] +
          rippleMap[mapIndex - 1] +
          rippleMap[mapIndex + 1]) >>
          1) - rippleMap[state.newIndex + i]

      rippleMap[state.newIndex + i] = data - (data >> 5)

      let amplitude = 1024 - data
      const oldAmp = lastMap[i]
      lastMap[i] = amplitude
      mapIndex++

      if (oldAmp !== amplitude) {
        active = true
        let dx = (((x - halfW) * amplitude) / 1024 + halfW) | 0
        let dy = (((y - halfH) * amplitude) / 1024 + halfH) | 0
        if (dx >= width) dx = width - 1
        if (dx < 0) dx = 0
        if (dy >= height) dy = height - 1
        if (dy < 0) dy = 0
        const src = (dx + dy * width) * 4
        const dst = i * 4
        output[dst] = texture[src]
        output[dst + 1] = texture[src + 1]
        output[dst + 2] = texture[src + 2]
        output[dst + 3] = texture[src + 3]
      } else {
        const dst = i * 4
        output[dst] = texture[dst]
        output[dst + 1] = texture[dst + 1]
        output[dst + 2] = texture[dst + 2]
        output[dst + 3] = texture[dst + 3]
      }
      i++
    }
  }
  return active
}
```

> 实现时若边界/索引与经典 demo 差一像素，以单测「有扰动会变、无扰动不变」为准微调，不必逐字节对齐某篇博文。

- [ ] **Step 4: 跑测确认通过**

```powershell
npm run test:unit -- src/utils/waterRipple.spec.ts
```

Expected: PASS

- [ ] **Step 5: Commit**（默认跳过，除非用户要求）

---

### Task 2: `HeroRipple.vue` 组件

**Files:**
- Create: `src/components/HeroRipple.vue`
- Create: `src/components/HeroRipple.spec.ts`

**Interfaces:**
- Consumes: `createRippleState` / `disturb` / `stepRipple` / `DEFAULT_*` from `@/utils/waterRipple`
- Produces: SFC props
  - `src?: string` 默认 `'/images/home/hero-chengdu.jpg'`
  - 根节点 class `hero-ripple`；成功运行时 `data-active="true"`；reduced-motion 或失败时无 `data-active` 或不挂载 canvas 动画

行为摘要：
1. `matchMedia('(prefers-reduced-motion: reduce)')` 为真 → 不创建动画循环，模板可只渲染空容器或隐藏 canvas。
2. 加载 `src` 图片 → 按容器尺寸（最长边上限 1280、`devicePixelRatio` cap 1.5）绘制到离屏/本 canvas，取出 `texture` ImageData。
3. `requestAnimationFrame` 调 `stepRipple` + `putImageData`。
4. 桌面且 `matchMedia('(pointer: fine)')`：在根上 `pointermove` → `disturb`（坐标按 canvas 内部分辨率缩放）；触摸不绑跟手。
5. `setInterval` 2.5–4s 随机间隔：在 `[0.15,0.85]` 宽高范围内随机 `disturb(..., DEFAULT_RADIUS, DEFAULT_STRENGTH * 0.75)`。
6. `IntersectionObserver` rootMargin `0px`：不可见暂停；`visibilitychange` 同步暂停。
7. `ResizeObserver` 防抖 150ms 重建 state / 重绘纹理。
8. canvas：`aria-hidden="true"`，`pointer-events: none`（跟手用父级监听也可；若父级监听则 canvas 可不接收事件——推荐在组件根 `div` 上听 pointer，canvas `pointer-events: none`，避免挡住下方点击？注意：Hero 文案在上层，背景本就不需点击；跟手区域应是整个 hero 背景。为简单：组件 `position:absolute; inset:0`，根 div 接收 pointer，但 `pointer-events: none` 会让跟手失效——改为 **canvas `pointer-events: auto` 仅当需要跟手**，文案在更高 z-index 仍可点。组件不要盖住 content：只放在 `hero__bg` 内即可。）

- [ ] **Step 1: 写组件冒烟单测**

```ts
// src/components/HeroRipple.spec.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import HeroRipple from './HeroRipple.vue'

describe('HeroRipple', () => {
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
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('reduced-motion: root is not data-active', () => {
    const wrapper = mount(HeroRipple)
    expect(wrapper.find('.hero-ripple').attributes('data-active')).toBeUndefined()
  })
})
```

（`matchMedia` mock 让 `prefers-reduced-motion` 为 true；实现里 reduced-motion 时不要设 `data-active`。）

- [ ] **Step 2: 跑测确认失败**

```powershell
npm run test:unit -- src/components/HeroRipple.spec.ts
```

Expected: FAIL（组件不存在）

- [ ] **Step 3: 实现 `HeroRipple.vue`**

关键实现骨架（完整细节按 Spec §3–4 补全）：

```vue
<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef } from 'vue'
import {
  createRippleState,
  disturb,
  stepRipple,
  DEFAULT_RADIUS,
  DEFAULT_STRENGTH,
  type RippleState,
} from '@/utils/waterRipple'

const props = withDefaults(defineProps<{ src?: string }>(), {
  src: '/images/home/hero-chengdu.jpg',
})

const rootRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const active = ref(false)

// reducedMotion / finePointer / loadImage / rebuildBuffers /
// rAF loop / auto disturb / IO / visibility / resize — 按 Spec 实现
</script>

<template>
  <div
    ref="rootRef"
    class="hero-ripple"
    :data-active="active ? 'true' : undefined"
    aria-hidden="true"
  >
    <canvas ref="canvasRef" class="hero-ripple__canvas" />
  </div>
</template>

<style scoped>
.hero-ripple {
  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
}
.hero-ripple__canvas {
  width: 100%;
  height: 100%;
  display: block;
}
.hero-ripple:not([data-active]) .hero-ripple__canvas {
  opacity: 0;
}
</style>
```

注意：图片 `drawImage` 使用 `object-fit: cover` 等效逻辑（计算 sx/sy/sw/sh 或先画到 canvas 再 `getImageData`），与现有 CSS `background-size: cover` 视觉一致。

- [ ] **Step 4: 跑测确认通过**

```powershell
npm run test:unit -- src/components/HeroRipple.spec.ts src/utils/waterRipple.spec.ts
```

Expected: PASS

- [ ] **Step 5: Commit**（默认跳过）

---

### Task 3: 接入 `HomeView` Hero

**Files:**
- Modify: `src/views/HomeView/HomeView.vue`

**Interfaces:**
- Consumes: `HeroRipple` default export
- Produces: Hero DOM 中静态图 + 波纹叠层

- [ ] **Step 1: 在 script 增加 import**

```ts
import HeroRipple from '@/components/HeroRipple.vue'
```

- [ ] **Step 2: 调整 Hero 模板**

将：

```vue
<div class="hero__bg">
  <div class="hero__photo" role="img" aria-label="成都风景" />
  <div class="hero__gradient" />
  <div class="hero__pattern" />
</div>
```

改为：

```vue
<div class="hero__bg">
  <div class="hero__photo" role="img" aria-label="成都风景" />
  <HeroRipple />
  <div class="hero__gradient" />
  <div class="hero__pattern" />
</div>
```

- [ ] **Step 3: 确认样式叠层**

- `.hero__bg` 保持 `position: absolute; inset: 0; z-index: 0`
- `.hero__photo` / `.hero-ripple` 均绝对铺满；canvas 在静态图之上
- `.hero__gradient`、`.hero__pattern` 的 `z-index` 高于 ripple（若当前无 z-index，给 gradient/pattern `z-index: 1`，content 已有更高层）
- `.hero__content` / `.hero__shu` 保持在前景，不改文案

示例补丁：

```css
.hero__photo {
  z-index: 0;
}
.hero__gradient,
.hero__pattern {
  z-index: 1;
}
```

（`HeroRipple` 根默认 z-index: 0，与 photo 同级但 DOM 在后，故盖住静态图；active 时可见。）

- [ ] **Step 4: 手动验收（dev server）**

```powershell
cd e:\wenlv\wennv\wenlv
npm run dev
```

按 Spec §6：
1. 桌面移入 Hero → 轻柔跟手涟漪，标题清晰
2. 静置 → 偶发自动涟漪
3. 窄屏 / 手机模式 → 仅自动，滑动不跟手
4. DevTools 模拟 `prefers-reduced-motion: reduce` → 无波纹，静态图在
5. 滚离 Hero → 动画停

- [ ] **Step 5: Commit**（默认跳过）

---

## Self-Review

| Spec 要求 | 对应 Task |
|-----------|-----------|
| Canvas 2D 振幅折射 | Task 1–2 |
| 桌面跟手 + 自动 | Task 2 |
| 手机仅自动 | Task 2 `pointer: fine` |
| 轻柔强度 | Task 1 `DEFAULT_*` + 自动 0.75× |
| 只扭曲照片层 | Task 3 叠层 |
| reduced-motion 静态兜底 | Task 2 + 静态 `hero__photo` |
| 不可见 / hidden 暂停 | Task 2 |
| 不引入第三方库 | File Map 无新依赖 |

无 TBD / 占位步骤；类型名 `RippleState` / `createRippleState` / `disturb` / `stepRipple` 前后一致。

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-09-18-hero-water-ripple.md`.

**两种执行方式：**

1. **Subagent-Driven（推荐）** — 每个 Task 派一个新子代理，Task 间复核  
2. **Inline Execution** — 本会话按 executing-plans 连续执行并设检查点  

你选哪种？
