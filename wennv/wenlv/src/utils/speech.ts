/**
 * speech.ts — 多语语音导览工具(浏览器原生 SpeechSynthesis,零 API 费用)
 * - speak(text, lang):按句子拆分长文本,逐个 SpeechSynthesisUtterance 入队朗读
 * - stopSpeaking():取消朗读并复位状态
 * - isSpeaking / speechSupported:响应式状态,供详情页朗读按钮绑定
 * - hasVoiceFor(lang):判断当前浏览器是否有指定语言的语音(无则按钮置灰)
 */
import { ref } from 'vue'
import type { Language } from '@/types'

/** 主站语言 → BCP47 语音标签 */
const BCP47: Record<Language, string> = { zh: 'zh-CN', en: 'en-US', ja: 'ja-JP' }

/** 语言 → 语音 lang 前缀(匹配 voice.lang 用) */
const LANG_PREFIX: Record<Language, string> = { zh: 'zh', en: 'en', ja: 'ja' }

/** 是否正在朗读(朗读开始 true,队列结束或 stopSpeaking 后 false) */
export const isSpeaking = ref(false)

/** 浏览器是否有可用语音(getVoices 异步加载完成后刷新) */
export const speechSupported = ref(false)

/** 已加载的语音列表(voiceschanged 后刷新;computed 依赖它获得响应式) */
const voices = ref<SpeechSynthesisVoice[]>([])

/** voiceschanged 是否已监听(getVoices 在部分浏览器异步加载,需监听重取) */
let voicesHooked = false

/** 朗读代次:stopSpeaking 后旧 Utterance 的回调不再改状态,避免竞态 */
let generation = 0

function speechApiSupported(): boolean {
  return typeof window !== 'undefined' && 'speechSynthesis' in window
}

/** 挂载 voiceschanged 监听并立即刷新一次语音列表 */
function hookVoicesChanged() {
  if (voicesHooked || !speechApiSupported()) return
  voicesHooked = true
  window.speechSynthesis.addEventListener('voiceschanged', refreshVoices)
  refreshVoices()
}

function refreshVoices() {
  if (!speechApiSupported()) {
    voices.value = []
    speechSupported.value = false
    return
  }
  voices.value = window.speechSynthesis.getVoices()
  speechSupported.value = voices.value.length > 0
}

/** 按语言前缀匹配最优 voice:精确 → 主子标签 → 任意前缀 */
function pickVoice(lang: Language): SpeechSynthesisVoice | undefined {
  const prefix = LANG_PREFIX[lang]
  const full = BCP47[lang].toLowerCase()
  const normalized = (v: SpeechSynthesisVoice) => v.lang.replace('_', '-').toLowerCase()
  return (
    voices.value.find((v) => normalized(v) === full) ??
    voices.value.find((v) => normalized(v).startsWith(`${prefix}-`)) ??
    voices.value.find((v) => normalized(v).startsWith(prefix))
  )
}

/** 是否有该语言的可用语音(无则详情页朗读按钮置灰;voiceschanged 后响应式刷新) */
export function hasVoiceFor(lang: Language): boolean {
  hookVoicesChanged()
  return pickVoice(lang) !== undefined
}

/** 按中/英/日句读符号拆句,标点跟随前句,避免长文本在部分浏览器被截断 */
function splitSentences(text: string): string[] {
  return (text.match(/[^。！？.!?]+[。！？.!?]*/g) ?? [text])
    .map((s) => s.trim())
    .filter(Boolean)
}

/**
 * 朗读文本:逐句创建 SpeechSynthesisUtterance 入队。
 * 朗读开始 isSpeaking=true,全部句子结束(或中途 stopSpeaking)后复位 false。
 */
export function speak(text: string, lang: Language): void {
  const trimmed = text.trim()
  if (!speechApiSupported() || !trimmed) return
  stopSpeaking()
  hookVoicesChanged()
  refreshVoices()

  const gen = generation
  const sentences = splitSentences(trimmed)
  if (!sentences.length) return
  isSpeaking.value = true

  let remaining = sentences.length
  const settle = () => {
    remaining -= 1
    if (remaining <= 0 && gen === generation) isSpeaking.value = false
  }

  for (const sentence of sentences) {
    const utterance = new SpeechSynthesisUtterance(sentence)
    utterance.lang = BCP47[lang]
    utterance.rate = 0.95
    const voice = pickVoice(lang)
    if (voice) utterance.voice = voice
    utterance.onend = settle
    utterance.onerror = settle
    window.speechSynthesis.speak(utterance)
  }
}

/** 停止朗读:取消队列并复位状态 */
export function stopSpeaking(): void {
  if (!speechApiSupported()) return
  generation += 1
  window.speechSynthesis.cancel()
  isSpeaking.value = false
}
