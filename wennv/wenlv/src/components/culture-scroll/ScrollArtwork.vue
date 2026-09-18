<script setup lang="ts">
import { computed } from 'vue'
import { CULTURE_SCROLL_SEGMENTS } from '@/data/cultureScroll'
import {
  ALLEY_VP,
  SCROLL_ART,
  alleyLanterns,
  alleyLines,
  brickFacade,
  cityGate,
  crenellatedWall,
  cypressPath,
  eraMarks,
  farBirds,
  farRidges,
  housePath,
  lamps,
  marketStalls,
  midGrass,
  midTrees,
  modernTowers,
  nearRailings,
  nearRoad,
  nearTrees,
  pagodaPath,
  pavilion,
  qinWater,
  queTower,
  republicanBlocks,
  ritualPoles,
  sacredOrnaments,
  sacredTreePath,
  stall,
  stoneBridge,
  sunBirdPath,
  tangHouses,
  templeHall,
  terracePaths,
  towerPath,
  treePath,
} from './scrollGeometry'

defineProps<{
  layer: 'far' | 'mid' | 'near'
}>()

const viewBox = `${0} ${0} ${SCROLL_ART.width} ${SCROLL_ART.height}`
const ground = SCROLL_ART.ground
const segments = CULTURE_SCROLL_SEGMENTS

const terrace = terracePaths()
const tree = sacredTreePath()
const sunBird = sunBirdPath(910, ground - 168, 46)
const leftQue = queTower(2588, 'left', 1)
const rightQue = queTower(3284, 'right', 1)
const wall = crenellatedWall(2508, 3488, ground - 96)
const gate = cityGate(2988, ground - 96)
const hallFront = templeHall(2788, 140, 48, 28, 10)
const hallMain = templeHall(2868, 168, 62, 36, 18)
const pagoda = pagodaPath(4324, 7)
const bridgeQinSide = stoneBridge(1148, 160)
const bridgeAlley = stoneBridge(4740, 200)
const midTreePaths = midTrees.map((item) => ({ ...item, d: treePath(item) }))
const nearTreePaths = nearTrees.map((item) => ({ ...item, d: treePath(item) }))
const towers = modernTowers.map((item) => ({ ...item, ...towerPath(item) }))
const brick = republicanBlocks.map((item) => brickFacade(item))
const houses = tangHouses.map((item) => housePath(item))

const colophons = computed(() =>
  eraMarks.map((mark) => {
    const seg = segments.find((s) => s.id === mark.id)
    return { ...mark, label: seg?.era ?? mark.id }
  }),
)

const poleFlags = ritualPoles.map((pole) => {
  const top = ground - pole.h
  return {
    pole: `M${pole.x} ${ground}V${top}`,
    flag: `M${pole.x} ${top + 8}L${pole.x + 22} ${top + 18}L${pole.x} ${top + 28}`,
  }
})

const lotus = [
  `M3888 ${ground - 8}q12 -18 24 0q-12 10 -24 0`,
  `M3924 ${ground - 4}q10 -14 20 0`,
  `M4412 ${ground - 6}q14 -16 26 2`,
]

const overpass = `M6188 ${ground - 22}C6360 ${ground - 88} 6588 ${ground - 96} 6840 ${ground - 40}C6960 ${ground - 18} 7060 ${ground - 8} 7188 ${ground - 12}`
</script>

<template>
  <svg
    class="scroll-art"
    :class="`scroll-art--${layer}`"
    :viewBox="viewBox"
    preserveAspectRatio="none"
    aria-hidden="true"
    focusable="false"
    :data-scroll-art="layer"
  >
    <g v-if="layer === 'far'" class="scroll-art__far">
      <path
        v-for="(ridge, i) in farRidges"
        :key="i"
        class="scroll-art__ridge"
        :d="ridge.d"
        :style="{ opacity: ridge.opacity }"
      />
      <path v-for="(d, i) in farBirds" :key="`b-${i}`" class="scroll-art__hair" :d="d" />
    </g>

    <g v-else-if="layer === 'mid'" class="scroll-art__mid">
      <g data-era="ancient-shu">
        <path v-for="(d, i) in terrace" :key="`t-${i}`" class="scroll-art__ink scroll-art__soft" :d="d" />
        <path class="scroll-art__ink" :d="tree" />
        <circle
          v-for="(o, i) in sacredOrnaments"
          :key="`o-${i}`"
          class="scroll-art__ink"
          :cx="o.cx"
          :cy="o.cy"
          :r="o.r"
        />
        <path class="scroll-art__ink" :d="sunBird" />
        <g v-for="(p, i) in poleFlags" :key="`p-${i}`">
          <path class="scroll-art__ink" :d="p.pole" />
          <path class="scroll-art__ink" :d="p.flag" />
        </g>
      </g>

      <g data-era="qin">
        <path class="scroll-art__water" :d="qinWater.outer" />
        <path class="scroll-art__water" :d="qinWater.inner" />
        <path class="scroll-art__ink scroll-art__soft" :d="qinWater.fishMouth" />
        <path v-for="(d, i) in qinWater.weir" :key="`w-${i}`" class="scroll-art__ink" :d="d" />
        <path v-for="(d, i) in qinWater.bottleMouth" :key="`bm-${i}`" class="scroll-art__ink" :d="d" />
        <path v-for="(d, i) in qinWater.ripples" :key="`r-${i}`" class="scroll-art__hair" :d="d" />
        <path class="scroll-art__ink" :d="pavilion(1468, 602, 58)" />
        <path class="scroll-art__ink" :d="bridgeQinSide" />
      </g>

      <g data-era="shu-han">
        <path class="scroll-art__ink scroll-art__soft" :d="wall" />
        <path class="scroll-art__ink" :d="gate" />
        <path class="scroll-art__ink" :d="leftQue" />
        <path class="scroll-art__ink" :d="rightQue" />
        <path class="scroll-art__ink" :d="hallFront" />
        <path class="scroll-art__ink" :d="hallMain" />
        <path class="scroll-art__ink" :d="cypressPath(2704, 122)" />
        <path class="scroll-art__ink" :d="cypressPath(3188, 128)" />
      </g>

      <g data-era="tang-song">
        <path v-for="(d, i) in houses" :key="`h-${i}`" class="scroll-art__ink scroll-art__soft" :d="d" />
        <path class="scroll-art__ink" :d="pagoda" />
        <path v-for="(d, i) in marketStalls" :key="`s-${i}`" class="scroll-art__ink" :d="d" />
        <path class="scroll-art__ink" :d="stall(4260, 48)" />
        <path v-for="(d, i) in lotus" :key="`l-${i}`" class="scroll-art__hair" :d="d" />
        <path class="scroll-art__water" :d="`M3612 ${ground - 4}C3780 ${ground + 10} 3960 ${ground - 18} 4140 ${ground + 2}C4280 ${ground + 12} 4388 ${ground - 8} 4520 ${ground + 6}`" />
      </g>

      <g data-era="ming-qing">
        <path class="scroll-art__ink" :d="bridgeAlley" />
        <path v-for="(d, i) in alleyLines" :key="`a-${i}`" class="scroll-art__ink" :d="d" />
        <ellipse
          v-for="(lantern, i) in alleyLanterns"
          :key="`lan-${i}`"
          class="scroll-art__ink"
          :cx="lantern.cx"
          :cy="lantern.cy"
          :rx="lantern.rx"
          :ry="lantern.ry"
        />
        <circle class="scroll-art__vanishing" :cx="ALLEY_VP.x" :cy="ALLEY_VP.y" r="2.2" />
        <path class="scroll-art__ink" :d="pavilion(5748, ground, 72)" />
      </g>

      <g data-era="modern">
        <path v-for="(d, i) in brick" :key="`br-${i}`" class="scroll-art__ink scroll-art__soft" :d="d" />
        <g v-for="(tw, i) in towers" :key="`tw-${i}`">
          <path class="scroll-art__ink scroll-art__soft" :d="tw.body" />
          <path class="scroll-art__hair" :d="tw.windows" />
          <path class="scroll-art__ink" :d="tw.crown" />
        </g>
        <path class="scroll-art__hair" :d="overpass" />
      </g>

      <path v-for="(item, i) in midTreePaths" :key="`mt-${i}`" class="scroll-art__ink" :d="item.d" />
      <path v-for="(d, i) in midGrass" :key="`g-${i}`" class="scroll-art__hair" :d="d" />

      <text
        v-for="mark in colophons"
        :key="mark.id"
        class="scroll-art__colophon"
        :x="mark.x"
        :y="mark.y"
        text-anchor="middle"
      >
        {{ mark.label }}
      </text>
    </g>

    <g v-else class="scroll-art__near">
      <path class="scroll-art__road" :d="nearRoad.fill" />
      <path v-for="(d, i) in nearRoad.edges" :key="`e-${i}`" class="scroll-art__ink" :d="d" />
      <path v-for="(d, i) in nearRoad.dashes" :key="`d-${i}`" class="scroll-art__dash" :d="d" />
      <path v-for="(d, i) in nearRailings" :key="`rl-${i}`" class="scroll-art__ink" :d="d" />
      <path v-for="(item, i) in nearTreePaths" :key="`nt-${i}`" class="scroll-art__ink" :d="item.d" />
      <path v-for="(d, i) in lamps" :key="`lp-${i}`" class="scroll-art__ink" :d="d" />
    </g>
  </svg>
</template>

<style scoped>
.scroll-art {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  height: 100%;
  display: block;
  overflow: visible;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.65;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.scroll-art__ridge {
  fill: currentColor;
  stroke: none;
}

.scroll-art__ink {
  fill: none;
  stroke: currentColor;
}

.scroll-art__soft {
  fill: currentColor;
  fill-opacity: 0.045;
}

.scroll-art__water {
  fill: currentColor;
  fill-opacity: 0.06;
  stroke: currentColor;
  stroke-width: 1.35;
}

.scroll-art__hair {
  fill: none;
  stroke: currentColor;
  stroke-width: 1.15;
}

.scroll-art__road {
  fill: currentColor;
  fill-opacity: 0.1;
  stroke: none;
}

.scroll-art__dash {
  fill: none;
  stroke: currentColor;
  stroke-width: 1.2;
  stroke-dasharray: 0;
  opacity: 0.45;
}

.scroll-art__vanishing {
  fill: currentColor;
  fill-opacity: 0.35;
  stroke: none;
}

.scroll-art__colophon {
  fill: currentColor;
  stroke: none;
  font-size: 22px;
  letter-spacing: 0.28em;
  opacity: 0.28;
  font-family: var(--font-display), serif;
}
</style>
