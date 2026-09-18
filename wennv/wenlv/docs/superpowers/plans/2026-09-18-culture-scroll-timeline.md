# Culture Scroll Embedded Timeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace CultureScroll clickable hotspots/dialog with a mid-layer left-to-right timeline whose nodes show period on the axis and full era cards alternating above/below (no click).

**Architecture:** Keep sticky scrub + mid-layer parallax. Mount a `.culture-scroll__timeline` inside the mid layer (same `layerStyle(1)`). Map each `CULTURE_SCROLL_HOTSPOTS` entry to a marker at `xPercent` with a non-interactive card (full `era`/`period`/`desc`). Remove hotspot buttons, scrim, and dialog state.

**Tech Stack:** Vue 3 `<script setup>` + TypeScript, Pinia language store, Vitest + Vue Test Utils, existing `@/data/cultureScroll`.

## Global Constraints

- Cards are **not clickable** (`pointer-events: none`); no dialog/scrim/Esc handlers.
- Card copy is **full** segment fields: `era`, `period`, `desc` (localized).
- Timeline is **embedded in mid layer**, not viewport-fixed; vertical position ≈ `top: 56%`.
- Alternation: index `0` even → card **above** axis; odd → card **below**.
- Reuse `CULTURE_SCROLL_SEGMENTS` + `CULTURE_SCROLL_HOTSPOTS` positions (rename optional; not required).
- Do not reintroduce near/road/walker layers.
- Prefer editing `CultureScroll.vue` + `CultureScroll.spec.ts` only unless rename is chosen.

## File map

| File | Responsibility |
|------|----------------|
| `src/components/CultureScroll.spec.ts` | Assert timeline markers + full cards; assert no hotspot/dialog |
| `src/components/CultureScroll.vue` | Remove hotspot/dialog; render timeline line + markers + cards; styles |
| `src/data/cultureScroll.ts` | Unchanged data (positions + copy); touch only if renaming hotspots |

---

### Task 1: Failing tests for always-visible timeline

**Files:**
- Modify: `src/components/CultureScroll.spec.ts`
- Test: `src/components/CultureScroll.spec.ts`

**Interfaces:**
- Consumes: existing `CultureScroll` mount helpers; `useLanguageStore`
- Produces: specs that expect `.culture-scroll__timeline`, `.culture-scroll__marker` × 6, full card text, no `.culture-scroll__hotspot` / `[role="dialog"]`

- [ ] **Step 1: Replace the hotspot/dialog test with timeline assertions**

In `CultureScroll.spec.ts`, delete the test named `places six mid-layer hotspots and opens era dialog` and add:

```ts
it('renders mid-layer timeline with six always-visible era cards', () => {
  const wrapper = mount(CultureScroll)
  expect(wrapper.find('.culture-scroll__timeline').exists()).toBe(true)
  expect(wrapper.findAll('.culture-scroll__hotspot')).toHaveLength(0)
  expect(wrapper.find('[role="dialog"]').exists()).toBe(false)

  const markers = wrapper.findAll('.culture-scroll__marker')
  expect(markers).toHaveLength(6)
  expect(markers.map((m) => m.attributes('data-marker'))).toEqual([
    'ancient-shu',
    'qin',
    'shu-han',
    'tang-song',
    'ming-qing',
    'modern',
  ])

  const first = markers[0]!
  expect(first.attributes('style')).toContain('%')
  expect(first.classes()).toContain('culture-scroll__marker--above')
  expect(markers[1]!.classes()).toContain('culture-scroll__marker--below')

  const card = first.find('.culture-scroll__era-card')
  expect(card.exists()).toBe(true)
  expect(card.text()).toContain('古蜀时期')
  expect(card.text()).toContain('约公元前1600年')
  expect(card.text()).toContain('三星堆')

  const axisTime = first.find('.culture-scroll__axis-time')
  expect(axisTime.exists()).toBe(true)
  expect(axisTime.text()).toContain('约公元前1600年')
})

it('timeline cards follow language store without dialog', async () => {
  const wrapper = mount(CultureScroll)
  useLanguageStore().setLang('en')
  await flushPromises()
  const first = wrapper.find('[data-marker="ancient-shu"]')
  expect(first.text()).toContain('Ancient Shu')
  expect(first.text()).toContain('1600 BC')
  expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npm run test:unit -- src/components/CultureScroll.spec.ts`

Expected: FAIL — missing `.culture-scroll__timeline` / `.culture-scroll__marker` (old hotspot test gone).

- [ ] **Step 3: Commit**

```bash
git add src/components/CultureScroll.spec.ts
git commit -m "test: expect culture scroll embedded timeline cards"
```

Skip commit if the repo is not a git workspace or the user has not asked to commit.

---

### Task 2: Remove hotspot/dialog; render timeline markup

**Files:**
- Modify: `src/components/CultureScroll.vue` (script + template)
- Test: `src/components/CultureScroll.spec.ts`

**Interfaces:**
- Consumes: `CULTURE_SCROLL_HOTSPOTS`, `CULTURE_SCROLL_SEGMENTS`, `localizedField`, `activeSegment`
- Produces:
  - Template classes: `culture-scroll__timeline`, `culture-scroll__marker`, `culture-scroll__marker--above|below`, `culture-scroll__marker--active`, `culture-scroll__dot`, `culture-scroll__axis-time`, `culture-scroll__era-card`
  - `data-marker="<segmentId>"`
  - Helper: `segmentForHotspot(segmentId: string): CultureScrollSegment | undefined`
  - Helper: `isActiveMarker(segmentId: string): boolean` → `activeSegment?.id === segmentId`

- [ ] **Step 1: Strip hotspot/dialog script state**

In `CultureScroll.vue` `<script setup>`:

1. Remove `openHotspotId`, `activeHotspot`, `openHotspot`, `closeHotspot`, `onHotspotPointerDown`, `onDialogKeydown`, `hotspotAriaLabel`.
2. Remove `window.addEventListener('keydown', onDialogKeydown)` and matching unmount removal.
3. Keep `localizedField` and `hotspots` / `segments`.
4. Add:

```ts
function segmentForHotspot(segmentId: string) {
  return segments.find((seg) => seg.id === segmentId)
}

function isActiveMarker(segmentId: string) {
  return activeSegment.value?.id === segmentId
}
```

- [ ] **Step 2: Replace hotspot buttons + dialog with timeline in mid layer**

Inside `.culture-scroll__layer--mid`, **after** the lineart track, replace the `v-for="hotspot in hotspots"` buttons and remove the scrim/dialog block entirely. Insert:

```vue
<div class="culture-scroll__timeline" aria-hidden="false">
  <div class="culture-scroll__timeline-line" aria-hidden="true" />
  <article
    v-for="(hotspot, index) in hotspots"
    :key="hotspot.segmentId"
    class="culture-scroll__marker"
    :class="{
      'culture-scroll__marker--above': index % 2 === 0,
      'culture-scroll__marker--below': index % 2 === 1,
      'culture-scroll__marker--active': isActiveMarker(hotspot.segmentId),
    }"
    :data-marker="hotspot.segmentId"
    :style="{ left: `${hotspot.xPercent}%` }"
  >
    <span class="culture-scroll__dot" aria-hidden="true" />
    <p class="culture-scroll__axis-time">
      {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'period') }}
    </p>
    <div class="culture-scroll__era-card">
      <h3 class="culture-scroll__era-card-title">
        {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'era') }}
      </h3>
      <p class="culture-scroll__era-card-period">
        {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'period') }}
      </p>
      <p class="culture-scroll__era-card-desc">
        {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'desc') }}
      </p>
    </div>
  </article>
</div>
```

Prefer resolving the segment once in a small computed map if preferred, but matching the above class names is required for tests.

Keep the colophon:

```vue
<p class="culture-scroll__colophon">
  {{ activeSegment?.era }} · {{ activeSegment?.eraEn }}
</p>
```

- [ ] **Step 3: Run unit tests**

Run: `npm run test:unit -- src/components/CultureScroll.spec.ts`

Expected: markup tests PASS; styling not required for pass. If `segmentForHotspot(...)!` throws in edge cases, guard with `v-if` on segment.

- [ ] **Step 4: Commit**

```bash
git add src/components/CultureScroll.vue
git commit -m "feat: embed always-visible era timeline on culture scroll"
```

Skip commit unless the user asked.

---

### Task 3: Timeline / card visual styles

**Files:**
- Modify: `src/components/CultureScroll.vue` (`<style scoped>`)

**Interfaces:**
- Consumes: Task 2 class names
- Produces: axis at ~56% height; cards alternate above/below; `pointer-events: none` on timeline; active marker subtle emphasis

- [ ] **Step 1: Remove obsolete hotspot/dialog CSS**

Delete rules for:

- `.culture-scroll__hotspot` (+ focus + pulse keyframes)
- `.culture-scroll__scrim`
- `.culture-scroll__dialog*`

- [ ] **Step 2: Add timeline CSS**

Append (tune tokens to existing CSS variables):

```css
.culture-scroll__timeline {
  position: absolute;
  top: 56%;
  left: 0;
  width: 100%;
  height: 0;
  z-index: 4;
  pointer-events: none;
}

.culture-scroll__timeline-line {
  position: absolute;
  left: 4%;
  right: 4%;
  top: 0;
  height: 1px;
  background: color-mix(in srgb, var(--color-text-primary) 42%, transparent);
}

.culture-scroll__marker {
  position: absolute;
  top: 0;
  width: min(280px, 22vw);
  transform: translate(-50%, 0);
}

.culture-scroll__dot {
  position: absolute;
  left: 50%;
  top: 0;
  width: 10px;
  height: 10px;
  border-radius: var(--radius-full);
  border: 1.5px solid color-mix(in srgb, var(--color-gold) 75%, #fff);
  background: color-mix(in srgb, var(--color-gold) 85%, #f4ead0);
  transform: translate(-50%, -50%);
}

.culture-scroll__marker--active .culture-scroll__dot {
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--color-gold) 22%, transparent);
}

.culture-scroll__axis-time {
  position: absolute;
  left: 50%;
  top: 10px;
  margin: 0;
  transform: translateX(-50%);
  white-space: nowrap;
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wider);
  color: color-mix(in srgb, var(--color-text-primary) 70%, transparent);
}

.culture-scroll__marker--above .culture-scroll__axis-time {
  top: auto;
  bottom: 14px;
}

.culture-scroll__era-card {
  position: absolute;
  left: 50%;
  width: 100%;
  padding: var(--space-3) var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-border) 80%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-surface) 88%, transparent);
  box-shadow: 0 1px 0 color-mix(in srgb, var(--color-text-primary) 6%, transparent);
  transform: translateX(-50%);
}

.culture-scroll__marker--above .culture-scroll__era-card {
  bottom: calc(100% + 28px);
}

.culture-scroll__marker--below .culture-scroll__era-card {
  top: 36px;
}

.culture-scroll__era-card-title {
  margin: 0 0 var(--space-1);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 700;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
}

.culture-scroll__era-card-period {
  margin: 0 0 var(--space-2);
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  color: var(--color-gold);
}

.culture-scroll__era-card-desc {
  margin: 0;
  font-size: var(--text-xs);
  line-height: var(--leading-relaxed);
  color: var(--color-text-secondary);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 5;
  overflow: hidden;
}

@media (max-width: 720px) {
  .culture-scroll__marker {
    width: min(220px, 42vw);
  }
  .culture-scroll__era-card-desc {
    -webkit-line-clamp: 4;
  }
}
```

Note: for `--above` markers, axis-time sits just above the line (between card and dot); for `--below`, axis-time sits just under the dot. Adjust if visual QA shows collision.

- [ ] **Step 3: Re-run unit tests + type-check**

Run:

```bash
npm run test:unit -- src/components/CultureScroll.spec.ts
npm run type-check
```

Expected: all CultureScroll tests PASS; `vue-tsc` clean for touched files.

- [ ] **Step 4: Manual visual check**

Run: `npm run dev` → open homepage 文明脉络 → scrub/drag:

1. Thin axis across mid illustration (~56%).
2. Six cards alternate above/below with full copy.
3. Period visible on axis at each node.
4. No clickable gold hotspot pulse; no dialog.
5. Switch EN: card text English.

- [ ] **Step 5: Commit**

```bash
git add src/components/CultureScroll.vue
git commit -m "style: polish culture scroll timeline cards"
```

Skip commit unless the user asked.

---

## Spec coverage checklist

| Spec requirement | Task |
|------------------|------|
| Mid embedded axis @ ~56% | Task 3 |
| Node shows period on axis | Task 2–3 |
| Full era/period/desc cards | Task 1–2 |
| Even above / odd below | Task 1–2 |
| No click / no dialog | Task 1–2 |
| Moves with mid layer | Task 2 (inside mid) |
| Keep colophon; active highlight | Task 2–3 |
| Update unit tests | Task 1–2 |
| No near/walker | unchanged; Task 1 keeps existing assert |

## Self-review notes

- No TBD placeholders.
- Class names consistent across tests and implementation tasks.
- Hotspot data kept as position source; rename deferred (YAGNI).
