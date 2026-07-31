/**
 * Theme Store — 深色/浅色模式切换
 * 作用于 `<html>` 元素的 `.theme-light` / `.theme-dark` 类
 */
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type Theme = 'dark' | 'light'

export const useThemeStore = defineStore('theme', () => {
  const theme = ref<Theme>('dark')

  function toggle() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function setTheme(t: Theme) {
    theme.value = t
  }

  // 同步到 <html> class
  watch(theme, (val) => {
    document.documentElement.classList.remove('theme-dark', 'theme-light')
    document.documentElement.classList.add(`theme-${val}`)
  }, { immediate: true })

  return { theme, toggle, setTheme }
})
