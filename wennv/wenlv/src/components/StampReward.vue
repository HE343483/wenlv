<script setup lang="ts">
/**
 * StampReward.vue — 打卡成功奖励动画（纯 CSS，无依赖）
 * 全屏遮罩：红章从高空"砸"下落定（scale+rotate+回弹），墨圈扩散，
 * 环绕金色纸屑粒子迸发；约 2.6s 后自动关闭。
 */
import { watch, onBeforeUnmount } from 'vue'
import { useLanguageStore } from '@/stores/language'

const props = defineProps<{
  show: boolean
  spotName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const langStore = useLanguageStore()

/* 14 枚纸屑粒子的方向角（CSS 自定义属性驱动 transform） */
const PARTICLES = Array.from({ length: 14 }, (_, i) => ({
  angle: (360 / 14) * i + (i % 3) * 8,
  dist: 130 + (i % 4) * 34,
  delay: (i % 5) * 60,
  color: ['#D9A441', '#B93A2B', '#2D6A4F', '#E8C97E'][i % 4],
}))

let timer: number | undefined

watch(
  () => props.show,
  (show) => {
    if (timer) {
      window.clearTimeout(timer)
      timer = undefined
    }
    if (show) {
      timer = window.setTimeout(() => emit('close'), 2600)
    }
  }
)

onBeforeUnmount(() => {
  if (timer) window.clearTimeout(timer)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="stamp-fade">
      <div v-if="show" class="stamp-reward" @click="emit('close')">
        <div class="stamp-reward__burst" aria-hidden="true">
          <span
            v-for="(p, i) in PARTICLES"
            :key="i"
            class="stamp-reward__confetti"
            :style="{
              '--angle': `${p.angle}deg`,
              '--dist': `${p.dist}px`,
              '--delay': `${p.delay}ms`,
              background: p.color,
            }"
          />
        </div>

        <div class="stamp-reward__ring" aria-hidden="true" />

        <div class="stamp-reward__seal">
          <span class="stamp-reward__char">印</span>
        </div>

        <p class="stamp-reward__title">{{ langStore.t('checkin.rewardTitle') }}</p>
        <p class="stamp-reward__name">{{ spotName }}</p>
        <p class="stamp-reward__desc">{{ langStore.t('checkin.rewardDesc') }}</p>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.stamp-reward {
  position: fixed;
  inset: 0;
  z-index: 1300;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  background: rgba(20, 16, 12, 0.72);
  cursor: pointer;
}

.stamp-fade-enter-active,
.stamp-fade-leave-active {
  transition: opacity 0.28s ease;
}
.stamp-fade-enter-from,
.stamp-fade-leave-to {
  opacity: 0;
}

/* 红章：高空砸落 + 回弹 + 微旋转 */
.stamp-reward__seal {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 132px;
  height: 132px;
  border-radius: var(--radius-full);
  border: 4px solid var(--color-cinnabar);
  box-shadow:
    inset 0 0 0 6px var(--color-surface),
    inset 0 0 0 8px color-mix(in srgb, var(--color-cinnabar) 60%, transparent),
    0 0 42px color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 6%, transparent);
  animation: seal-slam 0.62s cubic-bezier(0.22, 1.4, 0.36, 1) both;
}

.stamp-reward__char {
  font-family: var(--font-display);
  font-size: 64px;
  font-weight: 900;
  color: var(--color-cinnabar);
}

@keyframes seal-slam {
  0% {
    transform: scale(3.2) rotate(-32deg);
    opacity: 0;
  }
  55% {
    transform: scale(0.92) rotate(-6deg);
    opacity: 1;
  }
  72% {
    transform: scale(1.08) rotate(-10deg);
  }
  100% {
    transform: scale(1) rotate(-8deg);
    opacity: 1;
  }
}

/* 墨圈：印章落定瞬间向外扩散 */
.stamp-reward__ring {
  position: absolute;
  width: 132px;
  height: 132px;
  border-radius: var(--radius-full);
  border: 2px solid var(--color-cinnabar);
  animation: ring-spread 0.9s ease-out 0.4s both;
  pointer-events: none;
}

@keyframes ring-spread {
  0% {
    transform: scale(1);
    opacity: 0.8;
  }
  100% {
    transform: scale(2.6);
    opacity: 0;
  }
}

/* 纸屑：从中心沿各自方向迸发后坠落淡出 */
.stamp-reward__burst {
  position: absolute;
  pointer-events: none;
}

.stamp-reward__confetti {
  position: absolute;
  width: 10px;
  height: 14px;
  border-radius: 2px;
  opacity: 0;
  animation: confetti-pop 1.1s ease-out var(--delay) forwards;
}

@keyframes confetti-pop {
  0% {
    transform: translate(0, 0) rotate(0deg) scale(0.6);
    opacity: 0;
  }
  12% {
    opacity: 1;
  }
  100% {
    transform: translate(
      calc(cos(var(--angle)) * var(--dist)),
      calc(sin(var(--angle)) * var(--dist) + 60px)
    ) rotate(220deg) scale(1);
    opacity: 0;
  }
}

.stamp-reward__title {
  margin: var(--space-5) 0 0;
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 800;
  color: #f5e6c8;
  letter-spacing: var(--tracking-wide);
  animation: text-rise 0.5s ease-out 0.55s both;
}

.stamp-reward__name {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: #e8b96a;
  animation: text-rise 0.5s ease-out 0.7s both;
}

.stamp-reward__desc {
  margin: 0;
  font-size: var(--text-sm);
  color: rgba(245, 230, 200, 0.75);
  animation: text-rise 0.5s ease-out 0.85s both;
}

@keyframes text-rise {
  from {
    transform: translateY(14px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* cos/sin 不支持时的兜底：至少保证粒子向外淡出 */
@supports not (transform: translate(cos(1deg) * 1px, 0)) {
  .stamp-reward__confetti {
    animation-name: confetti-fallback;
  }
  @keyframes confetti-fallback {
    0% {
      transform: scale(0.6);
      opacity: 0;
    }
    12% {
      opacity: 1;
    }
    100% {
      transform: scale(1.4) translateY(80px) rotate(200deg);
      opacity: 0;
    }
  }
}
</style>
