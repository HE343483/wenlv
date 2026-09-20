<script setup lang="ts">
/**
 * App.vue — 根组件
 * 数智文旅 + 国际传播：巴蜀文化出海
 * 初始化天气色调计算器 + 全局 AI 行程生成浮动提示
 */
import { computed, watch } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'
import { applyToneToRoot } from '@/utils/weatherTone'
import { useTripTaskStore } from '@/stores/tripTask'
import { i18n } from '@/i18n'

// 依据平均温度 ({{min_temp}}+{{max_temp}})/2 动态调整页面色调
const weatherStore = useWeatherStore()
const { avgTemp } = storeToRefs(weatherStore)

watch(avgTemp, (t) => {
  applyToneToRoot(t)
}, { immediate: true })

// AI 行程生成中：浮动进度提示（用户切换页面后仍可见，点击回到行程页）
const route = useRoute()
const router = useRouter()
const tripTask = useTripTaskStore()
const showTaskFloat = computed(() => tripTask.generating && route.path !== '/trip')
const floatText = computed(() => i18n.global.t('home.messages.backgroundBadge'))
const gotoTrip = () => router.push('/trip')
</script>

<template>
  <RouterView />
  <Transition name="task-float">
    <div
      v-if="showTaskFloat"
      class="trip-task-float"
      role="status"
      @click="gotoTrip"
    >
      <i class="trip-task-float-spinner"></i>
      <span class="trip-task-float-text">{{ floatText }}</span>
      <span class="trip-task-float-percent">{{ Math.round(tripTask.progress) }}%</span>
    </div>
  </Transition>
</template>

<style>
html {
  scroll-padding-top: calc(var(--nav-height) + var(--space-4));
}
</style>

<style scoped>
.trip-task-float {
  position: fixed;
  right: 20px;
  bottom: 24px;
  z-index: 1200;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 999px;
  background: rgba(12, 23, 32, 0.88);
  border: 1px solid rgba(215, 110, 66, 0.55);
  color: #ffe3d6;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 10px 30px rgba(4, 11, 18, 0.45);
  cursor: pointer;
  user-select: none;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.trip-task-float:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 34px rgba(4, 11, 18, 0.55);
}

.trip-task-float-spinner {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid rgba(215, 110, 66, 0.3);
  border-top-color: #d76e42;
  animation: trip-task-spin 0.8s linear infinite;
}

.trip-task-float-percent {
  color: #f0a078;
}

@keyframes trip-task-spin {
  to {
    transform: rotate(360deg);
  }
}

.task-float-enter-active,
.task-float-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.task-float-enter-from,
.task-float-leave-to {
  opacity: 0;
  transform: translateY(12px);
}
</style>
