/** 语言切换 Store — 全局国际化状态管理 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Language } from '@/types'
import zh from '@/locales/zh'
import en from '@/locales/en'

type DeepValue<T, K extends string> = K extends keyof T
  ? T[K]
  : K extends `${infer A}.${infer B}`
    ? A extends keyof T
      ? DeepValue<T[A], B>
      : string
    : string

const locales = { zh, en } as const

export const useLanguageStore = defineStore('language', () => {
  const lang = ref<Language>('zh')

  const currentLang = computed(() => lang.value)

  function setLang(l: Language) {
    lang.value = l
  }

  function toggle() {
    lang.value = lang.value === 'zh' ? 'en' : 'zh'
  }

  /** 获取多语言文本：t('nav.home') => '首页' / 'Home' */
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

  return { lang, currentLang, setLang, toggle, t }
})
