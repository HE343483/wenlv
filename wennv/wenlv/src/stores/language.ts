/** 语言切换 Store — 全局国际化状态管理(与 AI 行程模块语言双向同步) */

import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { Language } from '@/types'
import zh from '@/locales/zh'
import en from '@/locales/en'
import ja from '@/locales/ja'
import { tripI18n, setAppLocale, getCurrentLocale } from '@/trip/i18n'

const locales = { zh, en, ja } as const

/** 主站语言 -> AI 行程 locale(zh -> zh-CN, en -> en-US, ja -> ja-JP) */
const toAppLocale = (l: Language): 'zh-CN' | 'en-US' | 'ja-JP' =>
  l === 'zh' ? 'zh-CN' : l === 'en' ? 'en-US' : 'ja-JP'

/** AI 行程 locale -> 主站语言 */
const fromAppLocale = (locale: string): Language | undefined => {
  if (locale.startsWith('zh')) return 'zh'
  if (locale.startsWith('en')) return 'en'
  if (locale.startsWith('ja')) return 'ja'
  return undefined
}

export const useLanguageStore = defineStore('language', () => {
  // 初始值与 AI 行程模块保持一致(其内部已读取 localStorage / 浏览器语言)
  const lang = ref<Language>(fromAppLocale(getCurrentLocale()) ?? 'zh')

  const currentLang = computed(() => lang.value)

  function setLang(l: Language) {
    lang.value = l
    // 同步 AI 行程模块的语言(其内部会写入 localStorage 并更新 <html lang>)
    setAppLocale(toAppLocale(l))
  }

  /** 循环切换:中 → EN → 日 → 中 */
  function toggle() {
    const order: Language[] = ['zh', 'en', 'ja']
    const next = order[(order.indexOf(lang.value) + 1) % order.length] as Language
    setLang(next)
  }

  /** 获取多语言文本：t('nav.home') => '首页' / 'Home' / 'ホーム' */
  function t(key: string): string {
    const locale = locales[lang.value] as Record<string, unknown>
    const parts = key.split('.')
    let result: unknown = locale
    for (const part of parts) {
      if (result && typeof result === 'object' && part in result) {
        result = (result as Record<string, unknown>)[part]
      } else {
        return key
      }
    }
    return typeof result === 'string' ? result : key
  }

  // AI 行程模块内切换语言时,反向同步主站语言
  watch(tripI18n.global.locale, (locale) => {
    const normalized = fromAppLocale(locale)
    if (normalized && normalized !== lang.value) {
      lang.value = normalized
    }
  })

  return { lang, currentLang, setLang, toggle, t }
})
