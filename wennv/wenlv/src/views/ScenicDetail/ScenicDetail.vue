<script setup lang="ts">
/**
 * ScenicDetail.vue — 景点详情页
 * 说明：当前为纯布局骨架（图片占位 / 文案占位），
 *       后期由后端接口按路由参数 :id 拉取景点详情数据填充。
 */
import { computed, ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { getScenic, getScenicAround, getScenicTransport, getScenicPoems } from '@/api/content'
import type { ScenicItem, ScenicAroundItem, ScenicTransitStop, PoemItem } from '@/api/content'
import { parseSections, estimatedSet, splitList, displayFact } from '@/utils/scenicDetail'
import { pickDesc, pickCultureNote, pickName } from '@/utils/storyI18n'
import { listCheckIns } from '@/api/checkin'
import { hasToken } from '@/utils/token'
import { isSpeaking as voiceSpeaking, speak, stopSpeaking, hasVoiceFor } from '@/utils/speech'
import dictZh from '@/locales/zh'
import dictEn from '@/locales/en'
import dictJa from '@/locales/ja'
import { getRuntimeMapJsKey } from '@/api/trip'
import AMapLoader from '@amap/amap-jsapi-loader'
import AppIcon from '@/components/AppIcon.vue'
import CheckInModal from '@/components/CheckInModal.vue'
import StampReward from '@/components/StampReward.vue'
import QuizCard from '@/components/QuizCard.vue'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()

/* 路由参数：景点ID（后端数据接入后据此拉取详情） */
const scenicId = computed(() => String(route.params.id ?? ''))

function goBack() {
  router.back()
}

/* 占位标签 — 后期替换为接口返回的 tags */
const placeholderTags = ['文化', '地标', '美食']
const tagIndex = computed(() =>
  Math.abs([...scenicId.value].reduce((acc, ch) => acc + ch.charCodeAt(0), 0)) % placeholderTags.length
)

/* 占位图片区标题 — 后期替换为接口返回的名称 */
const placeholderName = computed(() =>
  langStore.lang === 'zh' ? `景点 · ${scenicId.value}` : `Scenic · ${scenicId.value}`
)

/* ── 接口数据:卡片 id 形如 scenic-3,取数字部分请求详情 ── */
const scenic = ref<ScenicItem | null>(null)
const imgFailed = ref(false)

const numericId = computed(() => {
  const n = Number(scenicId.value.replace(/^scenic-/, ''))
  return Number.isFinite(n) && n > 0 ? n : 0
})

onMounted(async () => {
  if (!numericId.value) return
  try {
    scenic.value = await getScenic(numericId.value)
  } catch {
    /* 加载失败时保持占位展示 */
  }
  refreshCheckInState()
  await initMap()
  /* 三个接口相互独立:任一失败仅清空自身字段,不影响其他 */
  const [aroundRes, transitRes, poemsRes] = await Promise.allSettled([
    getScenicAround(numericId.value, 6),
    getScenicTransport(numericId.value),
    getScenicPoems(numericId.value),
  ])
  around.value = aroundRes.status === 'fulfilled' ? (aroundRes.value || []) : []
  transit.value = transitRes.status === 'fulfilled' ? (transitRes.value || []) : []
  poems.value = poemsRes.status === 'fulfilled' ? (poemsRes.value || []) : []
})

const displayName = computed(() => {
  if (!scenic.value) return placeholderName.value
  return pickName(scenic.value, langStore.lang)
})

/* ── 打卡集章:上传现场照片落库 MySQL 后获得护照印章 ── */
const checkedIn = ref(false)
const checkinModalOpen = ref(false)
const rewardShow = ref(false)

async function refreshCheckInState() {
  if (!hasToken() || !numericId.value) return
  try {
    const items = await listCheckIns()
    checkedIn.value = items.some((it) => it.scenic_id === numericId.value)
  } catch {
    /* 网络失败按未打卡处理,不阻塞详情页主流程 */
  }
}

function onCheckinClick() {
  if (checkedIn.value) return
  if (!hasToken()) {
    router.push('/login')
    return
  }
  checkinModalOpen.value = true
}

function onCheckinSuccess() {
  checkedIn.value = true
  rewardShow.value = true
}

const subName = computed(() => {
  if (!scenic.value) return ''
  return langStore.lang === 'zh' ? (scenic.value.name_en || '') : scenic.value.name_zh
})
const spotTags = computed(() =>
  (scenic.value?.tags || '').split(',').map(t => t.trim()).filter(Boolean)
)
/* 简介正文:外语模式取故事版多语种介绍(未生成回落中文),中文模式维持原展示 */
const spotDesc = computed(() => pickDesc(scenic.value, langStore.lang))
/* 文化注解:仅非中文语言且后端已生成时展示 */
const cultureNotes = computed(() => pickCultureNote(scenic.value, langStore.lang))

/* ── 相关诗词(诗词地图):中文原文任何界面语言都展示;译文/赏析按语言展开 ── */
const poems = ref<PoemItem[]>([])
const openPoemId = ref(0)

function togglePoem(poem: PoemItem) {
  openPoemId.value = openPoemId.value === poem.id ? 0 : poem.id
}

/** 中文原文按行拆分(原文入库时以 \n 分行) */
function poemLines(contentZh: string): string[] {
  return contentZh.split('\n').map((line) => line.trim()).filter(Boolean)
}

/** 当前语言下的展开内容:中文界面显示白话赏析;外语界面显示对应译文 */
function poemExtras(poem: PoemItem): { key: string; label: string; text: string }[] {
  if (langStore.lang === 'zh') {
    return poem.plain_zh
      ? [{ key: 'analysis', label: langStore.t('poems.analysis'), text: poem.plain_zh }]
      : []
  }
  const text = langStore.lang === 'en' ? (poem.content_en || '') : (poem.content_ja || '')
  return text ? [{ key: 'translation', label: langStore.t('poems.translation'), text }] : []
}
const heroImage = computed(() => (scenic.value?.images && !imgFailed.value) ? scenic.value.images : '')

/* ── 多语语音导览:朗读当前语言的故事正文 ── */
const speakText = computed(() => spotDesc.value.trim())
const voiceAvailable = computed(() => !!speakText.value && hasVoiceFor(langStore.lang))
const voiceTitle = computed(() => {
  if (voiceSpeaking.value) return langStore.t('voice.stop')
  return voiceAvailable.value ? langStore.t('voice.play') : langStore.t('voice.unsupported')
})

function toggleVoice() {
  if (voiceSpeaking.value) {
    stopSpeaking()
    return
  }
  if (speakText.value) speak(speakText.value, langStore.lang)
}

/* 切换语言时停止朗读,避免跨语言续播 */
watch(() => langStore.lang, () => stopSpeaking())

/* ── 四川话一分钟:结构化方言数据(数组取值走语言字典,langStore.t 仅支持字符串) ── */
interface DialectWord { word: string; pinyin: string; meaning: string; scene: string }
const dialectDicts = { zh: dictZh, en: dictEn, ja: dictJa } as const
const dialectWords = computed<DialectWord[]>(
  () => (dialectDicts[langStore.lang] as unknown as { dialect: { words: DialectWord[] } }).dialect.words
)

/* ── 详情扩展数据:图文详情 / 精彩瞬间 / 参考值标注 ── */
const sections = computed(() => parseSections(scenic.value?.detail_sections))
const galleryImages = computed(() => splitList(scenic.value?.gallery_images))
const estimated = computed(() => estimatedSet(scenic.value?.estimated_fields))
const noData = computed(() => langStore.t('scenicDetail.noData'))
const isEstimated = (field: string) => estimated.value.has(field)

/* ── 周边推荐 / 交通站点:运行时查询高德,失败降级为空数组 ── */
const around = ref<ScenicAroundItem[]>([])
const transit = ref<ScenicTransitStop[]>([])

/**
 * 距离文案:后端返回单位为米,按量级切换 米/km 并走 i18n。
 * 空值 / 0 / 非正数返回空串,调用方据此隐藏该段文案。
 */
function formatDistance(meters?: number): string {
  if (!meters || meters <= 0) return ''
  if (meters < 1000) return langStore.t('scenicDetail.distanceM', { m: Math.round(meters) })
  return langStore.t('scenicDetail.distanceKm', { km: (meters / 1000).toFixed(1) })
}

const transportText = computed(() => {
  if (!transit.value.length) return ''
  return transit.value
    .slice(0, 2)
    .map((s) => `${s.name} ${formatDistance(s.distance)}`.trim())
    .join(' / ')
})

/* ── 景区位置:高德 JS 地图,Key 缺失或加载失败时回退占位 ── */
/* Key 与行程页保持一致:优先取运行时配置(设置页保存到 localStorage),缺失再回退构建期 env。
   @amap/amap-jsapi-loader 为单例,两处 key 不一致会导致后加载方 reject。 */
const mapKey = getRuntimeMapJsKey() || import.meta.env.VITE_AMAP_WEB_JS_KEY || ''
const mapReady = ref(false)
let amapInstance: { destroy: () => void } | null = null

async function initMap() {
  const spot = scenic.value
  if (!mapKey || !spot?.lng || !spot?.lat) return
  try {
    const AMap = await AMapLoader.load({ key: mapKey, version: '2.0', plugins: ['AMap.Marker'] })
    const map = new AMap.Map('scenic-map', {
      zoom: 15,
      center: [spot.lng, spot.lat],
      viewMode: '3D',
    })
    new AMap.Marker({ position: [spot.lng, spot.lat], title: displayName.value, map })
    amapInstance = map
    // 容器始终占位可见(占位块绝对定位覆盖其上),此处仅等待渲染完成并重算尺寸
    mapReady.value = true
    await nextTick()
    if (typeof map.resize === 'function') map.resize()
  } catch (err) {
    mapReady.value = false
    console.warn('高德地图初始化失败', err)
  }
}

onBeforeUnmount(() => {
  stopSpeaking()
  if (amapInstance) amapInstance.destroy()
})
</script>

<template>
  <div class="detail-page">
    <!-- ──── 顶栏 ──── -->
    <header class="detail-topbar">
      <button class="detail-topbar__back" @click="goBack">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>

      <div class="detail-topbar__brand">
        <span class="detail-topbar__title">蜀韵·成都</span>
        <span class="detail-topbar__crumb">{{ langStore.t('scenicDetail.breadcrumb') }}</span>
      </div>

      <div class="detail-topbar__actions">
        <!-- 语言切换：分段式 中/EN -->
        <div class="detail-topbar__lang" role="group" :aria-label="langStore.t('nav.ariaLang')">
          <button
            type="button"
            class="detail-topbar__lang-btn"
            :class="{ 'detail-topbar__lang-btn--active': langStore.lang === 'zh' }"
            @click="langStore.setLang('zh')"
          >
            中
          </button>
          <button
            type="button"
            class="detail-topbar__lang-btn"
            :class="{ 'detail-topbar__lang-btn--active': langStore.lang === 'en' }"
            @click="langStore.setLang('en')"
          >
            EN
          </button>
          <button
            type="button"
            class="detail-topbar__lang-btn"
            :class="{ 'detail-topbar__lang-btn--active': langStore.lang === 'ja' }"
            @click="langStore.setLang('ja')"
          >
            日
          </button>
        </div>
      </div>
    </header>

    <main class="detail-main">
      <!-- ──── HERO — 图片区（占位）──── -->
      <section class="detail-hero">
        <div class="detail-hero__media">
          <!-- 接口返回封面图,加载失败回退占位 -->
          <img
            v-if="heroImage"
            class="detail-hero__photo"
            :src="heroImage"
            :alt="displayName"
            referrerpolicy="no-referrer"
            @error="imgFailed = true"
          />
          <div v-else class="detail-hero__placeholder">
            <div class="detail-hero__shu" aria-hidden="true">景</div>
            <span class="detail-hero__api-badge">{{ langStore.t('scenicDetail.imagePlaceholder') }}</span>
          </div>

          <!-- 名称浮层 -->
          <div class="detail-hero__overlay">
            <h1 class="detail-hero__title">{{ displayName }}</h1>
            <p class="detail-hero__en-title">{{ subName }}</p>
            <div class="detail-hero__tags">
              <span
                v-for="(t, i) in (spotTags.length
                  ? spotTags
                  : (langStore.lang === 'zh' ? placeholderTags : ['Culture', 'Landmark', 'Food']))"
                :key="t"
                class="detail-hero__tag"
                :class="{ 'detail-hero__tag--accent': i === tagIndex }"
              >
                {{ t }}
              </span>
            </div>
            <!-- 打卡集章:需上传现场照片,通过后落库并盖护照印章 -->
            <button
              type="button"
              class="detail-hero__checkin"
              :class="{ 'detail-hero__checkin--done': checkedIn }"
              @click="onCheckinClick"
            >
              <span class="detail-hero__checkin-seal" aria-hidden="true">印</span>
              {{ checkedIn ? langStore.t('checkin.done') : langStore.t('checkin.button') }}
            </button>
          </div>
        </div>
      </section>

      <!-- ──── 概要：评价横条 + 简介 ──── -->
      <section class="detail-section container">
        <div class="detail-summary">
          <!-- 评价横条：分数 + 星级 + 关键信息，横向排列 -->
          <div class="detail-scorebar">
            <div class="detail-scorebar__rating">
              <span class="detail-scorebar__num">{{ scenic?.score ?? '—' }}</span>
              <span class="detail-scorebar__stars">
                <AppIcon v-for="i in 5" :key="i" name="star" :size="16" />
              </span>
              <span class="detail-scorebar__label">{{ langStore.t('scenic.rating') }}</span>
            </div>

            <span class="detail-scorebar__divider" aria-hidden="true" />

            <div class="detail-scorebar__fact">
              <span class="detail-scorebar__fact-key">{{ langStore.t('scenicDetail.district') }}</span>
              <span class="detail-scorebar__fact-value">{{ scenic?.district || '—' }}</span>
            </div>
            <div class="detail-scorebar__fact">
              <span class="detail-scorebar__fact-key">{{ langStore.t('scenicDetail.visits') }}</span>
              <span class="detail-scorebar__fact-value">
                {{ displayFact(scenic?.yearly_visitors, noData) }}
                <em v-if="isEstimated('yearly_visitors')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
              </span>
            </div>
            <div class="detail-scorebar__fact">
              <span class="detail-scorebar__fact-key">{{ langStore.t('scenicDetail.recommendTime') }}</span>
              <span class="detail-scorebar__fact-value">
                {{ displayFact(scenic?.recommend_hours, noData) }}
                <em v-if="isEstimated('recommend_hours')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
              </span>
            </div>
          </div>

          <!-- 简介 -->
          <div class="detail-summary__body">
            <div class="detail-summary__eyebrow">
              <span>◈</span>
              {{ langStore.t('scenicDetail.overview') }}
              <span>◈</span>
              <!-- 语音导览:朗读当前语言的故事正文 -->
              <button
                type="button"
                class="voice-btn"
                :class="{ 'voice-btn--active': voiceSpeaking }"
                :disabled="!voiceAvailable && !voiceSpeaking"
                :title="voiceTitle"
                @click="toggleVoice"
              >
                <span class="voice-btn__icon" aria-hidden="true">{{ voiceSpeaking ? '⏹' : '🔊' }}</span>
                {{ voiceSpeaking ? langStore.t('voice.stop') : langStore.t('voice.play') }}
              </button>
            </div>
            <p class="detail-summary__text">
              {{ spotDesc || langStore.t('scenicDetail.overviewPlaceholder') }}
            </p>

            <!-- 文化注解:外语模式下展示 LLM 生成的当地文化背景 -->
            <div v-if="cultureNotes.length" class="culture-note">
              <div class="culture-note__head">
                <span class="culture-note__icon" aria-hidden="true">📖</span>
                <div class="culture-note__titles">
                  <h3 class="culture-note__title">{{ langStore.t('cultureNote.title') }}</h3>
                  <p class="culture-note__hint">{{ langStore.t('cultureNote.hint') }}</p>
                </div>
              </div>
              <ul class="culture-note__list">
                <li v-for="(note, i) in cultureNotes" :key="i" class="culture-note__item">
                  <span class="culture-note__dot" aria-hidden="true" />
                  <span>{{ note }}</span>
                </li>
              </ul>
            </div>

            <!-- 相关诗词(诗词地图):原文人工权威录入,任何界面语言都显示中文原文 -->
            <div v-if="poems.length" class="poem-card">
              <div class="poem-card__head">
                <span class="poem-card__icon" aria-hidden="true">🖌️</span>
                <div class="poem-card__titles">
                  <h3 class="poem-card__title">{{ langStore.t('poems.title') }}</h3>
                  <p class="poem-card__hint">{{ langStore.t('poems.hint') }}</p>
                </div>
              </div>
              <div class="poem-card__list">
                <article
                  v-for="poem in poems"
                  :key="poem.id"
                  class="poem-item"
                  :class="{ 'poem-item--open': openPoemId === poem.id }"
                >
                  <header class="poem-item__head">
                    <div class="poem-item__meta">
                      <h4 class="poem-item__title">{{ poem.title }}</h4>
                      <span class="poem-item__author">{{ poem.dynasty }}·{{ poem.author }}</span>
                    </div>
                    <button
                      v-if="poemExtras(poem).length"
                      type="button"
                      class="poem-item__toggle"
                      @click="togglePoem(poem)"
                    >
                      {{ openPoemId === poem.id ? langStore.t('poems.collapse') : langStore.t('poems.expand') }}
                      <span aria-hidden="true">{{ openPoemId === poem.id ? '▴' : '▾' }}</span>
                    </button>
                  </header>
                  <div class="poem-item__original">
                    <p v-for="(line, i) in poemLines(poem.content_zh)" :key="i" class="poem-item__line">{{ line }}</p>
                  </div>
                  <div
                    v-if="openPoemId === poem.id && poemExtras(poem).length"
                    class="poem-item__extra"
                  >
                    <div v-for="block in poemExtras(poem)" :key="block.key" class="poem-item__extra-block">
                      <span class="poem-item__extra-label">{{ block.label }}</span>
                      <p class="poem-item__extra-text">{{ block.text }}</p>
                    </div>
                  </div>
                </article>
              </div>
            </div>

            <!-- 四川话一分钟:方言彩蛋,所有语言均展示 -->
            <div class="dialect-card">
              <div class="dialect-card__head">
                <span class="dialect-card__icon" aria-hidden="true">🌶</span>
                <div class="dialect-card__titles">
                  <h3 class="dialect-card__title">{{ langStore.t('dialect.title') }}</h3>
                  <p class="dialect-card__hint">{{ langStore.t('dialect.hint') }}</p>
                </div>
              </div>
              <ul class="dialect-card__list">
                <li v-for="w in dialectWords" :key="w.word" class="dialect-card__item">
                  <span class="dialect-card__word">{{ w.word }}</span>
                  <span class="dialect-card__pinyin">{{ w.pinyin }}</span>
                  <span class="dialect-card__meaning">{{ w.meaning }}</span>
                  <span class="dialect-card__scene">{{ w.scene }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 实用信息 ──── -->
      <section class="detail-section container">
        <div class="detail-info">
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="clock" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.openHours') }}</h3>
            <p class="detail-info__value">
              {{ displayFact(scenic?.open_hours, noData) }}
              <em v-if="isEstimated('open_hours')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="ticket" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.ticket') }}</h3>
            <p class="detail-info__value">
              {{ displayFact(scenic?.ticket_price, noData) }}
              <em v-if="isEstimated('ticket_price')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="bus" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.transport') }}</h3>
            <p class="detail-info__value">
              {{ transportText || langStore.t('scenicDetail.transportEmpty') }}
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="pin" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.address') }}</h3>
            <p class="detail-info__value">{{ displayFact(scenic?.address, noData) }}</p>
          </div>
        </div>
      </section>

      <!-- ──── 图文详情 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.detailTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.detailSubtitle') }}</p>
        </header>

        <div v-if="sections.length" class="detail-content">
          <div
            v-for="(sec, i) in sections"
            :key="sec.title"
            class="detail-content__row"
            :class="{ 'detail-content__row--reverse': i % 2 === 1 }"
          >
            <div class="detail-content__figure">
              <img
                v-if="sec.image"
                class="detail-content__img detail-content__img--photo"
                :src="sec.image"
                :alt="sec.title"
                referrerpolicy="no-referrer"
              />
              <div v-else class="detail-content__img detail-content__img--empty">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <path d="M21 15l-5-5L5 21"/>
                </svg>
              </div>
            </div>
            <div class="detail-content__text">
              <h3 class="detail-content__caption">{{ sec.title }}</h3>
              <p class="detail-content__para">{{ sec.text }}</p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('scenicDetail.noData') }}</p>
      </section>

      <!-- ──── 相册 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.galleryTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.gallerySubtitle') }}</p>
        </header>

        <div v-if="galleryImages.length" class="detail-gallery">
          <div v-for="(img, i) in galleryImages" :key="img" class="detail-gallery__item">
            <img class="detail-gallery__photo" :src="img" :alt="`${displayName} ${i + 1}`" referrerpolicy="no-referrer" />
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('scenicDetail.noData') }}</p>
      </section>

      <!-- ──── 地图 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.mapTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.mapSubtitle') }}</p>
        </header>

        <div class="detail-map">
          <div id="scenic-map" class="detail-map__canvas" />
          <div v-if="!mapReady" class="detail-map__placeholder">
            <span class="detail-map__pin"><AppIcon name="pin" :size="40" /></span>
            <span class="detail-map__hint">{{ langStore.t('scenicDetail.noData') }}</span>
          </div>
        </div>
      </section>

      <!-- ──── 周边推荐 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.aroundTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.aroundSubtitle') }}</p>
        </header>

        <div v-if="around.length" class="detail-around">
          <div v-for="item in around" :key="item.id" class="detail-around__card">
            <div class="detail-around__body">
              <h3 class="detail-around__name">{{ item.name }}</h3>
              <p class="detail-around__desc">
                {{ item.address || item.type || '' }}
                <span v-if="formatDistance(item.distance)">· {{ formatDistance(item.distance) }}</span>
              </p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('scenicDetail.aroundEmpty') }}</p>

        <p class="detail-source">
          {{ langStore.t('scenicDetail.dataSource') }}
          <template v-if="estimated.size">· {{ langStore.t('scenicDetail.estimatedTip') }}</template>
        </p>
      </section>

      <!-- ──── 蜀文化知识闯关:答对得徽章,与旅行护照集章打通;无题目时组件自隐藏 ──── -->
      <section v-if="numericId" class="detail-section container">
        <QuizCard :spot-id="numericId" />
      </section>
    </main>

    <!-- 打卡弹窗 + 集章奖励动画 -->
    <CheckInModal
      v-model:open="checkinModalOpen"
      :scenic-id="numericId"
      :spot-name="displayName"
      @success="onCheckinSuccess"
    />
    <StampReward :show="rewardShow" :spot-name="displayName" @close="rewardShow = false" />
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
}

/* ========================================
   顶栏
   ======================================== */
.detail-topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--nav-height);
  padding: 0 var(--space-6);
  background: var(--color-nav-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--color-border);
}

.detail-topbar__back {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}

.detail-topbar__back:hover {
  background: var(--color-surface-hover);
  color: var(--color-gold);
}

.detail-topbar__brand {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.detail-topbar__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
}

.detail-topbar__crumb {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  padding-left: var(--space-3);
  border-left: 1px solid var(--color-border-light);
}

.detail-topbar__actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

/* 语言切换：胶囊分段器 中|EN */
.detail-topbar__lang {
  display: inline-flex;
  align-items: center;
  padding: 2px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface) 60%, transparent);
}

.detail-topbar__lang-btn {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.detail-topbar__lang-btn:hover {
  color: var(--color-gold-dark);
}

.detail-topbar__lang-btn--active {
  background: var(--color-gold);
  color: var(--color-bg);
}

/* ========================================
   主体
   ======================================== */
.detail-main {
  flex: 1;
}

.detail-section {
  padding: var(--space-12) 0;
}

/* ========================================
   HERO — 图片占位
   ======================================== */
.detail-hero__media {
  position: relative;
  width: 100%;
  height: clamp(320px, 46vh, 480px);
  overflow: hidden;
}

.detail-hero__photo {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-hero__placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  background:
    radial-gradient(ellipse 70% 60% at 50% 35%, color-mix(in srgb, var(--color-gold) 10%, transparent) 0%, transparent 70%),
    linear-gradient(180deg, var(--color-surface-elevated), var(--color-bg-alt));
  border-bottom: 1px solid var(--color-border);
}

.detail-hero__shu {
  font-family: var(--font-display);
  font-size: clamp(120px, 22vw, 240px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 40%, transparent);
  user-select: none;
  pointer-events: none;
}

.detail-hero__api-badge {
  font-size: var(--text-xs);
  padding: 4px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  background: color-mix(in srgb, var(--color-bg) 60%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
}

.detail-hero__overlay {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: var(--space-8) var(--space-6);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  text-align: center;
  background: linear-gradient(180deg, transparent 0%, color-mix(in srgb, var(--color-bg) 75%, transparent) 100%);
}

.detail-hero__title {
  font-family: var(--font-display);
  font-size: clamp(var(--text-3xl), 5vw, var(--text-5xl));
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.15;
  text-shadow: 0 2px 16px color-mix(in srgb, var(--color-bg) 60%, transparent);
}

.detail-hero__en-title {
  font-family: var(--font-en-display);
  font-style: italic;
  font-size: var(--text-base);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
}

.detail-hero__tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--space-2);
  margin-top: var(--space-1);
}

.detail-hero__tag {
  font-size: var(--text-xs);
  padding: 4px 14px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border-light);
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-bg) 45%, transparent);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  letter-spacing: var(--tracking-wide);
}

.detail-hero__tag--accent {
  border-color: color-mix(in srgb, var(--color-gold) 55%, transparent);
  color: var(--color-gold);
}

/* 打卡集章按钮（hero 浮层内） */
.detail-hero__checkin {
  margin-top: var(--space-3);
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: 8px 20px 8px 10px;
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 55%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-cinnabar) 82%, black);
  color: #fff;
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast), opacity var(--transition-fast);
}

.detail-hero__checkin:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
}

.detail-hero__checkin--done {
  border-color: color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 16%, rgba(0, 0, 0, 0.35));
  color: color-mix(in srgb, var(--color-cinnabar) 35%, #fff);
  cursor: default;
  opacity: 0.9;
}

.detail-hero__checkin-seal {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  border: 1.5px solid currentColor;
  font-size: 12px;
  font-weight: 800;
  transform: rotate(-8deg);
}

/* ========================================
   概要 — 评价横条 + 简介
   ======================================== */
.detail-summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* 评价横条：分数 + 星级 + 关键信息，一行排开 */
.detail-scorebar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-5) var(--space-8);
  padding: var(--space-5) var(--space-8);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.detail-scorebar__rating {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.detail-scorebar__num {
  font-family: var(--font-en-display);
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-gold);
  line-height: 1;
}

.detail-scorebar__stars {
  display: flex;
  align-items: center;
  gap: 2px;
  color: var(--color-gold);
}

.detail-scorebar__label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* 评分与关键信息之间的竖分隔线 */
.detail-scorebar__divider {
  width: 1px;
  height: 32px;
  background: var(--color-border);
  flex-shrink: 0;
}

.detail-scorebar__fact {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.detail-scorebar__fact-key {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.detail-scorebar__fact-value {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
}

.detail-summary__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.detail-summary__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.detail-summary__text {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
  font-weight: 400;
}

/* ========================================
   文化注解卡(外语模式:琥珀色渐变)
   ======================================== */
.culture-note {
  padding: var(--space-5) var(--space-6);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(201, 138, 43, 0.35);
  background: linear-gradient(135deg, rgba(243, 224, 178, 0.6) 0%, rgba(252, 246, 230, 0.9) 60%, rgba(250, 236, 205, 0.55) 100%);
}

.culture-note__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.culture-note__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-xl);
}

.culture-note__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #8a5a12;
  letter-spacing: var(--tracking-wide);
}

.culture-note__hint {
  margin-top: 2px;
  font-size: var(--text-xs);
  color: rgba(138, 90, 18, 0.72);
  letter-spacing: var(--tracking-wide);
}

.culture-note__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.culture-note__item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  font-size: var(--text-sm);
  line-height: var(--leading-relaxed);
  color: #6b4f26;
}

.culture-note__dot {
  flex-shrink: 0;
  width: 6px;
  height: 6px;
  margin-top: 8px;
  border-radius: var(--radius-full);
  background: #c98a2b;
}

/* ========================================
   相关诗词卡(诗词地图:淡雅宣纸色系,与文化注解卡呼应)
   ======================================== */
.poem-card {
  margin-top: var(--space-4);
  padding: var(--space-5) var(--space-6);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(107, 79, 38, 0.26);
  background: linear-gradient(150deg, rgba(252, 250, 244, 0.92) 0%, rgba(247, 242, 230, 0.88) 55%, rgba(252, 248, 238, 0.92) 100%);
}

.poem-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.poem-card__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-xl);
}

.poem-card__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #5b4a2f;
  letter-spacing: var(--tracking-wide);
}

.poem-card__hint {
  margin-top: 2px;
  font-size: var(--text-xs);
  color: rgba(91, 74, 47, 0.66);
  letter-spacing: var(--tracking-wide);
}

.poem-card__list {
  display: flex;
  flex-direction: column;
}

.poem-item {
  padding: var(--space-3) 0;
  border-top: 1px dashed rgba(107, 79, 38, 0.24);
}

.poem-item:first-child {
  padding-top: 0;
  border-top: 0;
}

.poem-item__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}

.poem-item__meta {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  min-width: 0;
}

.poem-item__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #3f3222;
  letter-spacing: var(--tracking-wide);
}

.poem-item__author {
  flex-shrink: 0;
  font-size: var(--text-xs);
  color: rgba(91, 74, 47, 0.72);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

.poem-item__toggle {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 10px;
  font-size: var(--text-xs);
  color: #6b4f26;
  letter-spacing: var(--tracking-wide);
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(107, 79, 38, 0.28);
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: background 0.2s ease;
}

.poem-item__toggle:hover {
  background: rgba(255, 255, 255, 0.85);
}

.poem-item__original {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.poem-item__line {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  line-height: var(--leading-relaxed);
  letter-spacing: 0.08em;
  color: #4a3b26;
}

.poem-item__extra {
  margin-top: var(--space-2);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.55);
}

.poem-item__extra-label {
  display: inline-block;
  margin-bottom: 4px;
  font-size: var(--text-xs);
  font-weight: 700;
  color: #8a5a12;
  letter-spacing: var(--tracking-wide);
}

.poem-item__extra-text {
  font-size: var(--text-sm);
  line-height: var(--leading-relaxed);
  color: #6b5c42;
}

/* ── 语音导览按钮(故事标题旁) ── */
.voice-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: var(--space-3);
  padding: 3px 12px;
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
  color: var(--color-gold-dark);
  border: 1px solid color-mix(in srgb, var(--color-gold) 55%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
  transition: color var(--transition-fast), background var(--transition-fast),
    border-color var(--transition-fast), transform var(--transition-fast);
}

.voice-btn:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-gold) 16%, transparent);
  border-color: var(--color-gold);
  transform: translateY(-1px);
}

.voice-btn--active {
  color: var(--color-cinnabar);
  border-color: color-mix(in srgb, var(--color-cinnabar) 55%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 8%, transparent);
}

.voice-btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.voice-btn__icon {
  font-size: var(--text-sm);
  line-height: 1;
}

/* ========================================
   四川话一分钟卡(方言彩蛋:辣椒红渐变,与琥珀色 Culture Note 区分)
   ======================================== */
.dialect-card {
  padding: var(--space-5) var(--space-6);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(190, 47, 30, 0.32);
  background: linear-gradient(135deg, rgba(248, 218, 208, 0.6) 0%, rgba(253, 244, 239, 0.9) 60%, rgba(250, 226, 216, 0.55) 100%);
}

.dialect-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.dialect-card__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-xl);
}

.dialect-card__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #a02c1d;
  letter-spacing: var(--tracking-wide);
}

.dialect-card__hint {
  margin-top: 2px;
  font-size: var(--text-xs);
  color: rgba(160, 44, 29, 0.72);
  letter-spacing: var(--tracking-wide);
}

.dialect-card__list {
  display: flex;
  flex-direction: column;
}

.dialect-card__item {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--space-1) var(--space-3);
  padding: var(--space-2) 0;
  border-bottom: 1px dashed rgba(190, 47, 30, 0.22);
}

.dialect-card__item:last-child {
  border-bottom: none;
}

.dialect-card__word {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 800;
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-wide);
  line-height: 1.2;
}

.dialect-card__pinyin {
  font-family: var(--font-en-body);
  font-style: italic;
  font-size: var(--text-xs);
  color: rgba(160, 44, 29, 0.78);
  letter-spacing: var(--tracking-wide);
}

.dialect-card__meaning {
  font-size: var(--text-sm);
  color: #6f2a20;
}

.dialect-card__scene {
  margin-left: auto;
  font-size: var(--text-xs);
  color: rgba(111, 42, 32, 0.6);
  text-align: right;
}

@media (max-width: 640px) {
  .dialect-card__scene {
    margin-left: 0;
    width: 100%;
    text-align: left;
  }
}

/* ========================================
   实用信息
   ======================================== */
.detail-info {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

.detail-info__card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-2);
  padding: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.detail-info__card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
}

.detail-info__icon {
  display: flex;
  align-items: center;
  color: var(--color-gold);
  margin-bottom: var(--space-1);
}

.detail-info__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-info__value {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   区块标题
   ======================================== */
.detail-block-head {
  text-align: center;
  margin-bottom: var(--space-8);
}

/* ========================================
   图文详情
   ======================================== */
.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-12);
}

.detail-content__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-10);
  align-items: center;
}

.detail-content__figure {
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.detail-content__img {
  aspect-ratio: 4 / 3;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-gold) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.detail-content__img--empty {
  opacity: 0.7;
}

/* 非 4:3 的配图裁切填充,避免拉伸变形(占位 div 不受 object-fit 影响) */
.detail-content__img--photo {
  display: block;
  object-fit: cover;
}

.detail-content__text {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.detail-content__caption {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-content__para {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  white-space: pre-line;
}

/* ========================================
   相册
   ======================================== */
.detail-gallery {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

.detail-gallery__item {
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--color-border);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.detail-gallery__item:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
}

.detail-gallery__img {
  aspect-ratio: 4 / 3;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-gold) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.detail-gallery__num {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 45%, transparent);
  user-select: none;
}

/* ========================================
   地图占位
   ======================================== */
.detail-map {
  position: relative;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

/* 占位块绝对定位覆盖在地图容器之上:未就绪时看到占位,就绪后占位消失露出地图 */
.detail-map__placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  background:
    radial-gradient(ellipse 50% 70% at 50% 50%, color-mix(in srgb, var(--color-gold) 6%, transparent) 0%, transparent 70%),
    repeating-linear-gradient(0deg, transparent 0 39px, color-mix(in srgb, var(--color-border) 45%, transparent) 39px 40px),
    repeating-linear-gradient(90deg, transparent 0 39px, color-mix(in srgb, var(--color-border) 45%, transparent) 39px 40px),
    var(--color-bg-alt);
}

.detail-map__pin {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
  filter: drop-shadow(0 0 16px var(--color-gold-glow));
  animation: pin-float 2.6s ease-in-out infinite;
}

@keyframes pin-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.detail-map__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  background: color-mix(in srgb, var(--color-bg) 60%, transparent);
  padding: 4px 12px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
}

/* ========================================
   周边推荐
   ======================================== */
.detail-around {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.detail-around__card {
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-base);
}

.detail-around__card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg);
}

.detail-around__img {
  aspect-ratio: 16 / 9;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-gold) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
  opacity: 0.8;
}

.detail-around__body {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.detail-around__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-around__desc {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  line-height: var(--leading-relaxed);
}

/* ========================================
   数据来源标注 / 参考值徽标 / 空数据
   ======================================== */
.detail-est {
  display: inline-block;
  margin-left: var(--space-1);
  padding: 1px 6px;
  font-size: var(--text-xs);
  font-style: normal;
  color: var(--color-gold);
  border: 1px solid color-mix(in srgb, var(--color-gold) 45%, transparent);
  border-radius: var(--radius-full);
  vertical-align: middle;
}

.detail-empty {
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.detail-gallery__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  aspect-ratio: 4 / 3;
}

.detail-map__canvas {
  width: 100%;
  aspect-ratio: 16 / 6;
}

.detail-source {
  margin-top: var(--space-8);
  text-align: center;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .detail-info {
    grid-template-columns: repeat(2, 1fr);
  }
  .detail-gallery {
    grid-template-columns: repeat(2, 1fr);
  }
  .detail-around {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .detail-topbar__crumb {
    display: none;
  }
  .detail-topbar {
    padding: 0 var(--space-4);
  }
  .detail-section {
    padding: var(--space-8) 0;
  }
  .detail-summary {
    gap: var(--space-5);
  }
  .detail-scorebar {
    padding: var(--space-4) var(--space-5);
    gap: var(--space-4) var(--space-6);
  }
  .detail-scorebar__divider {
    display: none;
  }
  .detail-summary__text {
    font-size: var(--text-lg);
  }
  .detail-content__row {
    grid-template-columns: 1fr;
    gap: var(--space-6);
  }
}

@media (max-width: 640px) {
  .detail-info,
  .detail-gallery,
  .detail-around {
    grid-template-columns: 1fr;
  }
  .detail-hero__overlay {
    padding: var(--space-6) var(--space-4);
  }
}
</style>
