<script setup lang="ts">
/**
 * App.vue — 根组件
 * 数智文旅 + 国际传播：巴蜀文化出海
 * 初始化主题 store + 天气色调计算器
 */
import { watch } from 'vue'
import { RouterView } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useThemeStore } from '@/stores/theme'
import { useWeatherStore } from '@/stores/weather'
import { applyToneToRoot } from '@/utils/weatherTone'

// 初始化主题 — 将 dark/light 类同步到 <html>
useThemeStore()

// 依据平均温度 ({{min_temp}}+{{max_temp}})/2 动态调整页面色调
const weatherStore = useWeatherStore()
const { avgTemp } = storeToRefs(weatherStore)

watch(avgTemp, (t) => {
  applyToneToRoot(t)
}, { immediate: true })
</script>

<template>
  <RouterView />
</template>

<style>
html {
  scroll-padding-top: calc(var(--nav-height) + var(--space-4));
}
</style>
