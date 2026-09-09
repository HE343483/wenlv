<script setup lang="ts">
/**
 * PandaCursor.vue — 小熊猫自定义光标
 * 用一只 Q 版小熊猫替换系统指针；点击时小熊猫"啃一口"地动一下（压扁→回弹 + 眯眼笑）
 * 说明：
 *  - 仅在精确指针设备（鼠标/触控板）启用；触屏设备自动跳过
 *  - 挂载期间通过 body.panda-cursor-active 隐藏原生指针，卸载时自动恢复
 *  - 想全站生效：把 <PandaCursor /> 移到 App.vue 即可
 */
import { onBeforeUnmount, onMounted, ref } from 'vue'

const pandaRef = ref<HTMLDivElement | null>(null)
const visible = ref(false)

let rafId = 0
let targetX = -100
let targetY = -100

const finePointer =
  typeof window.matchMedia === 'function' && window.matchMedia('(pointer: fine)').matches

/** rAF 合帧更新位置，避免高频 pointermove 直接触发重排 */
function applyPosition() {
  rafId = 0
  pandaRef.value?.style.setProperty('transform', `translate3d(${targetX}px, ${targetY}px, 0)`)
}

function schedulePosition() {
  if (!rafId) rafId = requestAnimationFrame(applyPosition)
}

function onPointerMove(e: PointerEvent) {
  targetX = e.clientX
  targetY = e.clientY
  visible.value = true
  schedulePosition()
}

function onPointerDown(e: PointerEvent) {
  targetX = e.clientX
  targetY = e.clientY
  applyPosition()
  visible.value = true

  const el = pandaRef.value
  if (!el) return
  el.classList.remove('is-clicking')
  void el.offsetWidth // 强制回流，保证连续点击也能重启动画
  el.classList.add('is-clicking')
}

function onAnimationEnd() {
  pandaRef.value?.classList.remove('is-clicking')
}

/** 鼠标移出窗口时隐藏小熊猫 */
function onPointerLeave() {
  visible.value = false
}

onMounted(() => {
  if (!finePointer) return
  document.body.classList.add('panda-cursor-active')
  window.addEventListener('pointermove', onPointerMove, { passive: true })
  window.addEventListener('pointerdown', onPointerDown, { passive: true })
  document.documentElement.addEventListener('pointerleave', onPointerLeave)
})

onBeforeUnmount(() => {
  document.body.classList.remove('panda-cursor-active')
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerdown', onPointerDown)
  document.documentElement.removeEventListener('pointerleave', onPointerLeave)
  if (rafId) cancelAnimationFrame(rafId)
})
</script>

<template>
  <div
    ref="pandaRef"
    class="panda-cursor"
    :class="{ 'panda-cursor--hidden': !visible }"
    aria-hidden="true"
  >
    <svg
      class="panda-cursor__svg"
      width="36"
      height="36"
      viewBox="0 0 64 64"
      fill="none"
      @animationend="onAnimationEnd"
    >
      <!-- 耳朵 -->
      <circle cx="15" cy="13" r="9.5" fill="#2E3A3D" />
      <circle cx="49" cy="13" r="9.5" fill="#2E3A3D" />
      <!-- 脸 -->
      <ellipse cx="32" cy="35" rx="24" ry="21.5" fill="#FFFDF8" />
      <!-- 腮红 -->
      <ellipse cx="17.5" cy="39" rx="4.2" ry="2.6" fill="#B8453E" opacity="0.22" />
      <ellipse cx="46.5" cy="39" rx="4.2" ry="2.6" fill="#B8453E" opacity="0.22" />
      <!-- 眼斑 -->
      <ellipse cx="21.5" cy="30" rx="6.8" ry="9" transform="rotate(-16 21.5 30)" fill="#2E3A3D" />
      <ellipse cx="42.5" cy="30" rx="6.8" ry="9" transform="rotate(16 42.5 30)" fill="#2E3A3D" />
      <!-- 眼睛（常态：圆眼） -->
      <g class="panda-cursor__eyes-open">
        <circle cx="23" cy="29.5" r="2.4" fill="#FFFDF8" />
        <circle cx="41" cy="29.5" r="2.4" fill="#FFFDF8" />
        <circle cx="23.8" cy="28.8" r="0.9" fill="#2E3A3D" />
        <circle cx="41.8" cy="28.8" r="0.9" fill="#2E3A3D" />
      </g>
      <!-- 眼睛（点击：眯眼笑 ^ ^） -->
      <g class="panda-cursor__eyes-happy">
        <path d="M19.8 30 Q23 26.2 26.2 30" stroke="#FFFDF8" stroke-width="2" stroke-linecap="round" fill="none" />
        <path d="M37.8 30 Q41 26.2 44.2 30" stroke="#FFFDF8" stroke-width="2" stroke-linecap="round" fill="none" />
      </g>
      <!-- 鼻子 -->
      <path d="M29.2 37.6 H34.8 L32 41 Z" fill="#2E3A3D" stroke="#2E3A3D" stroke-width="1.6" stroke-linejoin="round" />
      <!-- 嘴 -->
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
  z-index: 9999;
  pointer-events: none;
  will-change: transform;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.panda-cursor--hidden {
  opacity: 0;
}

.panda-cursor__svg {
  display: block;
  transform: translate(-50%, -56%) rotate(-8deg);
  transform-origin: center;
  filter: drop-shadow(0 2px 5px rgba(46, 58, 61, 0.35));
}

/* 点击态：眯眼笑替换圆眼 */
.panda-cursor__eyes-happy {
  opacity: 0;
}

.panda-cursor.is-clicking .panda-cursor__eyes-open {
  opacity: 0;
}

.panda-cursor.is-clicking .panda-cursor__eyes-happy {
  opacity: 1;
}

/* 点击动画：压扁 → 拉伸回弹，像小熊猫啃了一口竹子 */
@keyframes panda-nibble {
  0%   { transform: translate(-50%, -56%) rotate(-8deg) scale(1); }
  30%  { transform: translate(-50%, -48%) rotate(-22deg) scale(1.08, 0.84); }
  55%  { transform: translate(-50%, -63%) rotate(6deg) scale(0.94, 1.1); }
  75%  { transform: translate(-50%, -54%) rotate(-14deg) scale(1.02, 0.96); }
  100% { transform: translate(-50%, -56%) rotate(-8deg) scale(1); }
}

.panda-cursor.is-clicking .panda-cursor__svg {
  animation: panda-nibble 0.45s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@media (prefers-reduced-motion: reduce) {
  .panda-cursor {
    transition: none;
  }
  .panda-cursor.is-clicking .panda-cursor__svg {
    animation: none;
  }
}

/* 触屏设备兜底隐藏 */
@media (pointer: coarse) {
  .panda-cursor {
    display: none;
  }
}
</style>

<style>
/* 挂载期间隐藏原生指针（非 scoped：需要作用到 body 及所有元素） */
body.panda-cursor-active,
body.panda-cursor-active * {
  cursor: none !important;
}
</style>
