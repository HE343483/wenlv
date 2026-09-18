# Culture Scroll Media Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the mid-layer shanshui JPG so only the yellow-blue stage wash + black lineart remain, and add an era photo card on the opposite side of each timeline text card.

**Architecture:** Extend `CultureScrollSegment` with `imageUrl`. Drop `SCENERY_SRC` / scenery track from `CultureScroll.vue`. Inside each marker, render `.culture-scroll__era-media` opposite `.culture-scroll__era-card` (above text → media below; below text → media above). Keep existing timeline scrub/parallax unchanged.

**Tech Stack:** Vue 3 `<script setup>` + TypeScript, Vitest + Vue Test Utils, `@/data/cultureScroll`.

## Global Constraints

- No `mid-scenery` / `SCENERY_SRC` rendering; stage keeps yellow-blue CSS gradient; far SVG stays.
- Black lineart only: `LINEART_SRC = '/images/culture-scroll/era-scroll-lineart-transparent-v1.png'`.
- Image card class: `.culture-scroll__era-media`; opposite the text card; `pointer-events: none`; `alt=""`.
- Exact image paths (verbatim):
  - `ancient-shu` → `/images/home/1.jpg`
  - `qin` → `/images/home/carousel-xiling.jpg`
  - `shu-han` → `/images/home/hero-chengdu.jpg`
  - `tang-song` → `/images/home/154b682806f848688077dbabd5bcac3f_720.jpg`
  - `ming-qing` → `/images/culture-scroll/era-scenery-shanshui-v1.jpg`
  - `modern` → `/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg`
- Do not reintroduce near/road/walker; no click/dialog on media cards.
- Prefer editing `cultureScroll.ts`, `CultureScroll.vue`, `CultureScroll.spec.ts` only.
- Skip git commits unless the user explicitly asks (environment may lack git).

## File map

| File | Responsibility |
|------|----------------|
| `src/data/cultureScroll.ts` | Add `imageUrl` to segment type + TIMELINE_FIELDS |
| `src/components/CultureScroll.vue` | Remove scenery; render media; CSS for opposition |
| `src/components/CultureScroll.spec.ts` | Assert no scenery; six media srcs; opposition |

---

### Task 1: Data `imageUrl` + failing tests

**Files:**
- Modify: `src/data/cultureScroll.ts`
- Modify: `src/components/CultureScroll.spec.ts`
- Test: `src/components/CultureScroll.spec.ts`

**Interfaces:**
- Consumes: existing `CultureScrollSegment` / `TIMELINE_FIELDS`
- Produces:
  - `CultureScrollSegment.imageUrl: string`
  - Each mapped segment carries the path from Global Constraints
  - Tests expect no scenery and six `.culture-scroll__era-media img` with exact `src`

- [ ] **Step 1: Add `imageUrl` to the data module**

In `src/data/cultureScroll.ts`:

1. Add to the interface:

```ts
export interface CultureScrollSegment {
  id: string
  era: string
  eraEn: string
  period: string
  periodEn: string
  desc: string
  descEn: string
  imageUrl: string
  progressStart: number
  progressEnd: number
}
```

2. Add `imageUrl` to each `TIMELINE_FIELDS` entry (exact strings):

```ts
// ancient-shu
imageUrl: '/images/home/1.jpg',
// qin
imageUrl: '/images/home/carousel-xiling.jpg',
// shu-han
imageUrl: '/images/home/hero-chengdu.jpg',
// tang-song
imageUrl: '/images/home/154b682806f848688077dbabd5bcac3f_720.jpg',
// ming-qing
imageUrl: '/images/culture-scroll/era-scenery-shanshui-v1.jpg',
// modern
imageUrl: '/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg',
```

`CULTURE_SCROLL_SEGMENTS` already spreads `TIMELINE_FIELDS`, so `imageUrl` flows through.

- [ ] **Step 2: Update / add failing UI tests**

In `CultureScroll.spec.ts`, replace the test named `renders far SVG with mid shanshui scenery + transparent lineart` with:

```ts
it('renders far SVG with transparent lineart only (no mid scenery)', () => {
  const wrapper = mount(CultureScroll)
  expect(wrapper.find('[data-scroll-art="far"]').exists()).toBe(true)
  expect(wrapper.find('[data-scroll-art="mid-scenery"]').exists()).toBe(false)
  expect(wrapper.findAll('.culture-scroll__scenery')).toHaveLength(0)
  expect(wrapper.find('[data-scroll-art="mid-lineart"]').exists()).toBe(true)
  expect(wrapper.findAll('.culture-scroll__lineart')).toHaveLength(4)
  expect(wrapper.find('.culture-scroll__lineart').attributes('src')).toContain(
    'era-scroll-lineart-transparent-v1.png',
  )
  expect(wrapper.find('canvas').exists()).toBe(false)
})
```

Add (or extend the timeline test with) media assertions:

```ts
it('places era media cards opposite text cards with confirmed srcs', () => {
  const wrapper = mount(CultureScroll)
  const markers = wrapper.findAll('.culture-scroll__marker')
  expect(markers).toHaveLength(6)

  const expectedSrcs = [
    '/images/home/1.jpg',
    '/images/home/carousel-xiling.jpg',
    '/images/home/hero-chengdu.jpg',
    '/images/home/154b682806f848688077dbabd5bcac3f_720.jpg',
    '/images/culture-scroll/era-scenery-shanshui-v1.jpg',
    '/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg',
  ]

  const medias = wrapper.findAll('.culture-scroll__era-media img')
  expect(medias).toHaveLength(6)
  expect(medias.map((img) => img.attributes('src'))).toEqual(expectedSrcs)

  // even index: text above → media below (no --media-above class)
  expect(markers[0]!.find('.culture-scroll__era-card').exists()).toBe(true)
  expect(markers[0]!.classes()).toContain('culture-scroll__marker--above')
  expect(markers[0]!.find('.culture-scroll__era-media').exists()).toBe(true)

  // odd: text below → media above
  expect(markers[1]!.classes()).toContain('culture-scroll__marker--below')
  expect(markers[1]!.find('.culture-scroll__era-media').exists()).toBe(true)
})
```

Optionally add a tiny data unit check in the same file:

```ts
import { CULTURE_SCROLL_SEGMENTS } from '@/data/cultureScroll'

it('segments expose imageUrl for every era', () => {
  expect(CULTURE_SCROLL_SEGMENTS.map((s) => s.imageUrl)).toEqual([
    '/images/home/1.jpg',
    '/images/home/carousel-xiling.jpg',
    '/images/home/hero-chengdu.jpg',
    '/images/home/154b682806f848688077dbabd5bcac3f_720.jpg',
    '/images/culture-scroll/era-scenery-shanshui-v1.jpg',
    '/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg',
  ])
})
```

- [ ] **Step 3: Run tests — expect UI media tests to FAIL**

Run: `npm run test:unit -- src/components/CultureScroll.spec.ts`

Expected:
- `segments expose imageUrl…` PASS (after Step 1)
- scenery-only / media UI tests FAIL until Task 2 (missing media markup / scenery still present)

- [ ] **Step 4: Commit (skip if no git / user did not ask)**

```bash
git add src/data/cultureScroll.ts src/components/CultureScroll.spec.ts
git commit -m "test: expect culture scroll era media cards without scenery"
```

---

### Task 2: Remove scenery; render opposite media markup

**Files:**
- Modify: `src/components/CultureScroll.vue` (script + template; minimal CSS ok)
- Test: `src/components/CultureScroll.spec.ts`

**Interfaces:**
- Consumes: `segmentForHotspot(...).imageUrl`
- Produces: DOM `.culture-scroll__era-media > img`; no `[data-scroll-art="mid-scenery"]`

- [ ] **Step 1: Strip scenery from script + template**

In `CultureScroll.vue`:

1. Delete `const SCENERY_SRC = ...`
2. Keep `LINEART_SRC` and `LINEART_TILES`
3. Delete the entire `.culture-scroll__scenery-track` block from the template

- [ ] **Step 2: Add media inside each marker**

Inside the `v-for` marker `article`, after the era-card block, add:

```vue
<div class="culture-scroll__era-media">
  <img
    :src="segmentForHotspot(hotspot.segmentId)!.imageUrl"
    alt=""
    draggable="false"
  />
</div>
```

Prefer resolving the segment once (e.g. local const via a small helper in the loop, or `v-for` over a computed list of `{ hotspot, segment, index }`) to avoid repeated `!` — either is fine if class names and `src` match tests.

- [ ] **Step 3: Run unit tests**

Run: `npm run test:unit -- src/components/CultureScroll.spec.ts`

Expected: all CultureScroll tests PASS (layout CSS can wait for Task 3 if opposition is structural via existing `--above`/`--below` parent classes).

- [ ] **Step 4: Commit (skip unless asked)**

```bash
git add src/components/CultureScroll.vue
git commit -m "feat: drop mid scenery and add era media on timeline"
```

---

### Task 3: Opposite-side media CSS + cleanup

**Files:**
- Modify: `src/components/CultureScroll.vue` (`<style scoped>`)

**Interfaces:**
- Consumes: `.culture-scroll__marker--above|below`, `.culture-scroll__era-media`
- Produces: media visually opposite text; width ~ text card; height ~140px; cover crop

- [ ] **Step 1: Delete scenery CSS**

Remove rules for `.culture-scroll__scenery-track`, `.culture-scroll__scenery`, and any shared selectors that only existed for scenery (keep lineart-track / lineart rules).

- [ ] **Step 2: Add media CSS**

```css
.culture-scroll__era-media {
  position: absolute;
  left: 50%;
  width: 100%;
  height: 140px;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-border) 80%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-surface) 70%, transparent);
  transform: translateX(-50%);
  pointer-events: none;
}

.culture-scroll__era-media img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  user-select: none;
  -webkit-user-drag: none;
}

/* text above → media below axis */
.culture-scroll__marker--above .culture-scroll__era-media {
  top: 36px;
}

/* text below → media above axis */
.culture-scroll__marker--below .culture-scroll__era-media {
  bottom: calc(100% + 28px);
}

@media (max-width: 720px) {
  .culture-scroll__era-media {
    height: 112px;
  }
}
```

If media collides with `axis-time` on `--above` markers, nudge `top` (e.g. `48px`) so axis label stays readable — keep opposition intact.

- [ ] **Step 3: Verify**

Run:

```bash
npm run test:unit -- src/components/CultureScroll.spec.ts
npm run type-check
```

Expected: PASS / clean.

Manual: `npm run dev` → 文明脉络 — yellow-blue wash, black lines only, six text/image pairs on opposite sides of the axis.

- [ ] **Step 4: Commit (skip unless asked)**

```bash
git add src/components/CultureScroll.vue
git commit -m "style: place era media opposite timeline text cards"
```

---

## Spec coverage checklist

| Spec requirement | Task |
|------------------|------|
| No mid scenery JPG | 1–2 |
| Yellow-blue stage + black lineart | 2–3 (stage CSS unchanged) |
| Opposite media cards | 2–3 |
| Exact six image paths | 1–2 |
| Non-clickable | 3 (`pointer-events: none`) |
| Tests updated | 1–2 |

## Self-review notes

- No TBD placeholders; paths match spec table verbatim.
- Class name `.culture-scroll__era-media` consistent across tasks.
- Scenery JPG may still exist on disk for `ming-qing` card use — do not delete the file.
