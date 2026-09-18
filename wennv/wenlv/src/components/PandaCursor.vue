<script setup lang="ts">
/**
 * PandaCursor.vue — 自定义小箭头 + 小熊猫伴游（触控本友好）
 *
 * 关键修复：
 *  - 用 (any-pointer: fine) 判断，触控本只要有触控板/鼠标即可启用
 *  - 仅在自定义光标可见时才 cursor:none；隐藏时立刻恢复系统箭头
 *  - 容器不用 0×0（部分浏览器会裁切 overflow 子元素）
 */
import { onBeforeUnmount, onMounted, watch, ref } from 'vue'

const rootRef = ref<HTMLDivElement | null>(null)
const visible = ref(false)

let rafId = 0
let targetX = -100
let targetY = -100

function hasFinePointer(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  // 触控本：主指针可能是 coarse，但 any-pointer: fine 仍为 true
  return window.matchMedia('(any-pointer: fine)').matches
}

const enabled = hasFinePointer()

function applyPosition() {
  rafId = 0
  rootRef.value?.style.setProperty('transform', `translate3d(${targetX}px, ${targetY}px, 0)`)
}

function schedulePosition() {
  if (!rafId) rafId = requestAnimationFrame(applyPosition)
}

function setSystemCursorHidden(hide: boolean) {
  document.body.classList.toggle('panda-cursor-active', hide)
}

function onPointerMove(e: PointerEvent) {
  if (e.pointerType === 'touch') return
  targetX = e.clientX
  targetY = e.clientY
  visible.value = true
  schedulePosition()
}

function onPointerDown(e: PointerEvent) {
  if (e.pointerType === 'touch') return
  targetX = e.clientX
  targetY = e.clientY
  visible.value = true
  applyPosition()

  const el = rootRef.value
  if (!el) return
  el.classList.remove('is-clicking')
  void el.offsetWidth
  el.classList.add('is-clicking')
}

function onAnimationEnd() {
  rootRef.value?.classList.remove('is-clicking')
}

function onPointerLeave() {
  visible.value = false
}

function onVisibilityChange() {
  if (document.hidden) visible.value = false
}

watch(visible, (v) => {
  // 只有画得出来时才藏系统指针，避免「全灭」
  setSystemCursorHidden(v)
})

onMounted(() => {
  setSystemCursorHidden(false)
  if (!enabled) return
  window.addEventListener('pointermove', onPointerMove, { passive: true })
  window.addEventListener('pointerdown', onPointerDown, { passive: true })
  document.documentElement.addEventListener('pointerleave', onPointerLeave)
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onBeforeUnmount(() => {
  setSystemCursorHidden(false)
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerdown', onPointerDown)
  document.documentElement.removeEventListener('pointerleave', onPointerLeave)
  document.removeEventListener('visibilitychange', onVisibilityChange)
  if (rafId) cancelAnimationFrame(rafId)
})
</script>

<template>
  <div
    v-if="enabled"
    ref="rootRef"
    class="panda-cursor"
    :class="{ 'panda-cursor--hidden': !visible }"
    aria-hidden="true"
  >
    <!-- 小三角形：尖端 = 指针热点 -->
    <svg
      class="panda-cursor__arrow"
      width="14"
      height="16"
      viewBox="0 0 14 16"
      fill="none"
    >
      <path
        d="M1 1 L1 15 L12 8 Z"
        fill="#FFFDF8"
        stroke="#2E3A3D"
        stroke-width="1.2"
        stroke-linejoin="round"
        stroke-linecap="round"
      />
    </svg>

    <svg
      class="panda-cursor__panda"
      width="32"
      height="32"
      viewBox="0 0 64 64"
      fill="none"
      @animationend="onAnimationEnd"
    >
      <circle cx="15" cy="13" r="9.5" fill="#2E3A3D" />
      <circle cx="49" cy="13" r="9.5" fill="#2E3A3D" />
      <ellipse cx="32" cy="35" rx="24" ry="21.5" fill="#FFFDF8" />
      <ellipse cx="17.5" cy="39" rx="4.2" ry="2.6" fill="#B8453E" opacity="0.22" />
      <ellipse cx="46.5" cy="39" rx="4.2" ry="2.6" fill="#B8453E" opacity="0.22" />
      <ellipse cx="21.5" cy="30" rx="6.8" ry="9" transform="rotate(-16 21.5 30)" fill="#2E3A3D" />
      <ellipse cx="42.5" cy="30" rx="6.8" ry="9" transform="rotate(16 42.5 30)" fill="#2E3A3D" />
      <g class="panda-cursor__eyes-open">
        <circle cx="23" cy="29.5" r="2.4" fill="#FFFDF8" />
        <circle cx="41" cy="29.5" r="2.4" fill="#FFFDF8" />
        <circle cx="23.8" cy="28.8" r="0.9" fill="#2E3A3D" />
        <circle cx="41.8" cy="28.8" r="0.9" fill="#2E3A3D" />
      </g>
      <g class="panda-cursor__eyes-happy">
        <path d="M19.8 30 Q23 26.2 26.2 30" stroke="#FFFDF8" stroke-width="2" stroke-linecap="round" fill="none" />
        <path d="M37.8 30 Q41 26.2 44.2 30" stroke="#FFFDF8" stroke-width="2" stroke-linecap="round" fill="none" />
      </g>
      <path d="M29.2 37.6 H34.8 L32 41 Z" fill="#2E3A3D" stroke="#2E3A3D" stroke-width="1.6" stroke-linejoin="round" />
      <path
        d="M32 41.6 V43.2 M32 43.2 Q29.4 45.8 27 43.8 M32 43.2 Q34.6 45.8 37 43.8"
        stroke="#2E3A3D"
        stroke-width="1.5"
        stroke-linecap="round"
        fill="none"
      />
    </svg>
  </div>
</template>

<style scoped>
.panda-cursor {
  position: fixed;
  left: 0;
  top: 0;
  /* 给足绘制盒，避免 0×0 在部分浏览器被裁切 */
  width: 56px;
  height: 56px;
  z-index: 2147483646;
  pointer-events: none;
  will-change: transform;
  opacity: 1;
  transition: opacity 0.12s ease;
  overflow: visible;
}

.panda-cursor--hidden {
  opacity: 0;
}

.panda-cursor__arrow {
  position: absolute;
  left: 0;
  top: 0;
  display: block;
  filter: drop-shadow(0 1px 2px rgba(46, 58, 61, 0.45));
}

.panda-cursor__panda {
  position: absolute;
  left: 10px;
  top: 12px;
  display: block;
  transform: rotate(-8deg);
  transform-origin: center;
  filter: drop-shadow(0 2px 4px rgba(46, 58, 61, 0.35));
}

.panda-cursor__eyes-happy {
  opacity: 0;
}

.panda-cursor.is-clicking .panda-cursor__eyes-open {
  opacity: 0;
}

.panda-cursor.is-clicking .panda-cursor__eyes-happy {
  opacity: 1;
}

@keyframes panda-nibble {
  0%   { transform: rotate(-8deg) scale(1); }
  30%  { transform: rotate(-22deg) scale(1.08, 0.84); }
  55%  { transform: rotate(6deg) scale(0.94, 1.1); }
  75%  { transform: rotate(-14deg) scale(1.02, 0.96); }
  100% { transform: rotate(-8deg) scale(1); }
}

.panda-cursor.is-clicking .panda-cursor__panda {
  animation: panda-nibble 0.45s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@media (prefers-reduced-motion: reduce) {
  .panda-cursor {
    transition: none;
  }
  .panda-cursor.is-clicking .panda-cursor__panda {
    animation: none;
  }
}
</style>

<style>
body.panda-cursor-active,
body.panda-cursor-active * {
  cursor: none !important;
}
</style>
