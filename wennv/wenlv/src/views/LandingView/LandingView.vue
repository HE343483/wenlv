<template>
  <div class="landing-page">
    <div class="lower-shade" :style="lowerShadeStyle"></div>
    <NavBar />

    <div class="wrapper">
      <div class="page-header section-dark landing-header" :style="pageHeaderStyle">
        <div class="filter"></div>
        <div class="content-center" :style="heroContentStyle">
          <div class="container">
            <!-- <p class="landing-hero-badge text-center">{{ t('home.heroBadge') }}</p> -->
            <div class="title-brand">
              <h1 class="presentation-title">
                {{ t('app.brand') }}
              </h1>
            </div>
            <h2 class="presentation-subtitle text-center">{{ t('home.titleLine') }}</h2>
          </div>
        </div>
        <div class="moving-clouds" :style="movingCloudsStyle"></div>
        <div class="fog-low" :style="fogLowStyle">
          <img src="https://demos.creative-tim.com/paper-kit-2/assets/img/clouds.png" alt="fog" />
        </div>
        <div class="fog-low right" :style="fogLowRightStyle">
          <img src="https://demos.creative-tim.com/paper-kit-2/assets/img/clouds.png" alt="fog" />
        </div>
        <div class="hero-bottom-shade" :style="heroBottomShadeStyle"></div>
      </div>
    </div>

    <section ref="formRef" class="form-section">
      <div class="form-panel" :style="[formRevealStyle, { minHeight: panelHeight === 'auto' ? 'auto' : panelHeight + 'px' }]" ref="panelRef">
        <a-form v-show="!tripTask.generating" :model="formData" layout="vertical" @finish="handleSubmit">
          <div class="step">
            <div class="step-head">
              <span>01</span>
              <h3>{{ t('home.step1') }}</h3>
            </div>

            <!-- 多城市动态列表 -->
            <div class="city-list">
              <div v-for="(cs, idx) in formData.cities" :key="idx" class="city-row">
                <a-form-item class="city-row-name" :rules="[{ required: true, message: t('home.cityRequired') }]">
                  <template #label>
                    <span class="field-label">{{ t('home.cityNLabel', { n: idx + 1 }) }}</span>
                  </template>
                  <a-input
                    v-model:value="cs.city"
                    :placeholder="t('home.cityPlaceholder')"
                    size="large"
                    class="field-input"
                  />
                </a-form-item>
                <a-form-item class="city-row-days">
                  <template #label>
                    <span class="field-label">{{ t('home.cityStayDays') }}</span>
                  </template>
                  <a-input-number
                    v-model:value="cs.days"
                    :min="1"
                    :max="15"
                    size="large"
                    class="field-input"
                    style="width: 100%"
                  />
                </a-form-item>
                <button
                  v-if="formData.cities.length > 1"
                  type="button"
                  class="city-remove-btn"
                  @click="removeCity(idx)"
                >×</button>
              </div>
              <button type="button" class="city-add-btn" @click="addCity">
                + {{ t('home.addCity') }}
              </button>
            </div>

            <!-- 日期与天数 -->
            <div class="grid grid-date">
              <a-form-item name="start_date" :rules="formRules.startDate">
                <template #label>
                  <span class="field-label">{{ t('home.startDateLabel') }}</span>
                </template>
                <a-date-picker
                  v-model:value="formData.start_date"
                  style="width: 100%"
                  size="large"
                  class="field-input"
                  :placeholder="t('home.startDatePlaceholder')"
                />
              </a-form-item>

              <a-form-item>
                <template #label>
                  <span class="field-label">{{ t('home.travelDaysLabel') }}</span>
                </template>
                <div class="days-chip">
                  <span class="days-number">{{ totalDays }}</span>
                  <span class="days-unit">{{ t('home.travelDaysUnit') }}</span>
                </div>
              </a-form-item>
            </div>
          </div>

          <div class="step">
            <div class="step-head">
              <span>02</span>
              <h3>{{ t('home.step2') }}</h3>
            </div>
            <div class="grid grid2">
              <a-form-item name="transportation">
                <template #label>
                  <span class="field-label">{{ t('home.transportationLabel') }}</span>
                </template>
                <a-select v-model:value="formData.transportation" size="large" class="field-select">
                  <a-select-option value="公共交通">{{ t('home.transportation.public') }}</a-select-option>
                  <a-select-option value="自驾">{{ t('home.transportation.drive') }}</a-select-option>
                  <a-select-option value="步行">{{ t('home.transportation.walk') }}</a-select-option>
                  <a-select-option value="混合">{{ t('home.transportation.mixed') }}</a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item name="accommodation">
                <template #label>
                  <span class="field-label">{{ t('home.accommodationLabel') }}</span>
                </template>
                <a-select v-model:value="formData.accommodation" size="large" class="field-select">
                  <a-select-option value="经济型酒店">{{ t('home.accommodation.budget') }}</a-select-option>
                  <a-select-option value="舒适型酒店">{{ t('home.accommodation.comfort') }}</a-select-option>
                  <a-select-option value="豪华酒店">{{ t('home.accommodation.luxury') }}</a-select-option>
                  <a-select-option value="民宿">{{ t('home.accommodation.homestay') }}</a-select-option>
                </a-select>
              </a-form-item>
            </div>

            <a-form-item name="preferences">
              <template #label>
                <span class="field-label">{{ t('home.interestsLabel') }}</span>
              </template>
              <div class="interest-grid">
                <a-checkbox-group v-model:value="formData.preferences" class="interest-group">
                  <label
                    v-for="item in interestOptions"
                    :key="item.value"
                    class="interest-pill"
                    :class="{ active: formData.preferences.includes(item.value) }"
                    @click.prevent="togglePreference(item.value)"
                  >
                    {{ t(item.labelKey) }}
                  </label>
                </a-checkbox-group>
              </div>
            </a-form-item>

            <a-form-item name="attraction_source">
              <template #label>
                <span class="field-label">{{ t('home.attractionSourceLabel') }}</span>
              </template>
              <a-radio-group
                :value="formData.attraction_source"
                size="large"
                class="source-radio"
                @change="onSourceChange"
              >
                <a-radio-button value="xhs">{{ t('home.attractionSource.xhs') }}</a-radio-button>
                <a-radio-button value="douyin">{{ t('home.attractionSource.douyin') }}</a-radio-button>
                <a-radio-button value="map">{{ t('home.attractionSource.map') }}</a-radio-button>
              </a-radio-group>
              <p class="source-hint">{{ t('home.attractionSourceHint') }}</p>
            </a-form-item>
          </div>

          <div class="step">
            <div class="step-head">
              <span>03</span>
              <h3>{{ t('home.step3') }}</h3>
            </div>
            <a-form-item name="free_text_input">
              <div class="field-textarea">
                <a-textarea
                  v-model:value="formData.free_text_input"
                  :placeholder="t('home.specialNeedsPlaceholder')"
                  :rows="4"
                  size="large"
                  class="special-textarea"
                />
              </div>
            </a-form-item>
          </div>

          <!-- 用户偏好记忆开关：由用户自主选择是否让 AI 参考历史偏好 -->
          <a-form-item class="memory-toggle-item">
            <div class="memory-toggle">
              <a-switch v-model:checked="memoryEnabled" size="large" />
              <div class="memory-toggle-copy">
                <span class="memory-toggle-label">{{ t('home.memoryToggle.label') }}</span>
                <span class="memory-toggle-desc">{{ t('home.memoryToggle.desc') }}</span>
              </div>
            </div>
          </a-form-item>

          <a-form-item>
            <button type="submit" class="btn btn-danger btn-round submit-btn" :class="{ loading: tripTask.generating }" :disabled="tripTask.generating">
              <span v-if="!tripTask.generating">{{ t('home.submit') }}</span>
              <span v-else class="loading-row">
                <i class="spinner"></i>
                {{ t('home.submitting') }}
              </span>
            </button>
          </a-form-item>
        </a-form>

        <!-- Node Loading Stepper -->
        <div v-show="tripTask.generating" class="stepper-wrapper">
          <div class="stepper-header">
            <h2 class="stepper-title">{{ t('home.loading.planCode', { code: tripTask.planCode }) }}</h2>
            <p class="stepper-subtitle">{{ t('home.loading.preparing') }}</p>
          </div>

          <div class="stepper-container">
            <!-- Step 1: Searching Attractions -->
            <div class="step-node" :class="{ active: tripTask.progress >= 0 && tripTask.progress <= 30, completed: tripTask.progress > 30 }">
              <div class="node-icon">
                <i v-if="tripTask.progress >= 0 && tripTask.progress <= 30" class="spinner-small"></i>
                <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
              </div>
              <p class="node-text">{{ tripTask.progress > 30 ? t('home.loading.searchedAttractions') : t('home.loading.searchingAttractions') }}</p>
            </div>
            <div class="step-divider" :class="{ completed: tripTask.progress > 30 }"></div>

            <!-- Step 2: Weather -->
            <div class="step-node" :class="{ active: tripTask.progress > 30 && tripTask.progress <= 50, completed: tripTask.progress > 50 }">
              <div class="node-icon">
                <i v-if="tripTask.progress > 30 && tripTask.progress <= 50" class="spinner-small"></i>
                <svg v-else width="20px" height="20px" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
                  <path d="M10.5 1.5V3.1M3.6 10H2M5.4512 4.95137L4.31982 3.82M15.5498 4.95137L16.6812 3.82M19 10H17.4M6.50007 10.0001C6.50007 7.79093 8.29093 6.00007 10.5001 6.00007C12.0061 6.00007 13.3177 6.83235 14.0001 8.06206M6 22C3.79086 22 2 20.2091 2 18C2 15.7909 3.79086 14 6 14C6.46419 14 6.90991 14.0791 7.32442 14.2245C8.04061 12.3396 9.86387 11 12 11C14.1361 11 15.9594 12.3396 16.6756 14.2245C17.0901 14.0791 17.5358 14 18 14C20.2091 14 22 15.7909 22 18C22 20.2091 20.2091 22 18 22C13.3597 22 9.87921 22 6 22Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <p class="node-text">{{ tripTask.progress > 50 ? t('home.loading.queriedWeather') : t('home.loading.queryingWeather') }}</p>
            </div>
            <div class="step-divider" :class="{ completed: tripTask.progress > 50 }"></div>

            <!-- Step 3: Hotels -->
            <div class="step-node" :class="{ active: tripTask.progress > 50 && tripTask.progress <= 70, completed: tripTask.progress > 70 }">
              <div class="node-icon">
                <i v-if="tripTask.progress > 50 && tripTask.progress <= 70" class="spinner-small"></i>
                <svg v-else fill="currentColor" width="25px" height="25px" viewBox="0 0 24 24" version="1.1" xml:space="preserve" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
                    <g id="Layer_Grid"/><g id="Layer_2">
                    <path d="M21,8c0-2.2-1.8-4-4-4H7C4.8,4,3,5.8,3,8v3.8c-0.6,0.5-1,1.3-1,2.2v2.7V17v2c0,0.6,0.4,1,1,1s1-0.4,1-1v-1h16v1   c0,0.6,0.4,1,1,1s1-0.4,1-1v-2v-0.3V14c0-0.9-0.4-1.7-1-2.2V8z M5,8c0-1.1,0.9-2,2-2h10c1.1,0,2,0.9,2,2v3h-1v-1c0-1.7-1.3-3-3-3   h-1c-0.8,0-1.5,0.3-2,0.8C11.5,7.3,10.8,7,10,7H9c-1.7,0-3,1.3-3,3v1H5V8z M16,10v1h-3v-1c0-0.6,0.4-1,1-1h1C15.6,9,16,9.4,16,10z    M11,10v1H8v-1c0-0.6,0.4-1,1-1h1C10.6,9,11,9.4,11,10z M20,16H4v-2c0-0.6,0.4-1,1-1h3h3h2h3h3c0.6,0,1,0.4,1,1V16z"/></g>
                </svg>
              </div>
              <p class="node-text">{{ tripTask.progress > 70 ? t('home.loading.recommendedHotels') : t('home.loading.recommendingHotels') }}</p>
            </div>
            <div class="step-divider" :class="{ completed: tripTask.progress > 70 }"></div>

            <!-- Step 4: Planning -->
            <div class="step-node" :class="{ active: tripTask.progress > 70 && tripTask.progress < 100, completed: tripTask.progress >= 100 }">
              <div class="node-icon">
                <i v-if="tripTask.progress > 70 && tripTask.progress < 100" class="spinner-small"></i>
                <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 11 12 14 22 4"></polyline><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path></svg>
              </div>
              <p class="node-text">{{ tripTask.progress >= 100 ? t('home.loading.done') : t('home.loading.generatingPlan') }}</p>
            </div>
          </div>

          <div class="stepper-footer">
            <h3>{{ tripTask.statusText }}</h3>
            <p v-if="tripTask.progress < 100">{{ t('home.loading.workingTogether') }}</p>
            <p v-else>{{ t('home.loading.donePrepare') }}</p>
          </div>
        </div>
      </div>
    </section>

    <section class="history-section">
      <div class="history-panel">
        <div class="history-head">
          <div>
            <p class="history-eyebrow">{{ t('home.history.eyebrow') }}</p>
            <h3 class="history-title">{{ t('home.history.title') }}</h3>
          </div>
          <a-button type="link" class="history-refresh" @click="loadHistoryPlans">
            {{ t('home.history.refresh') }}
          </a-button>
        </div>

        <div v-if="historyLoading" class="history-loading">
          {{ t('common.loading') }}
        </div>
        <a-empty v-else-if="historyPlans.length === 0" :description="t('home.history.empty')" />
        <div v-else class="history-list">
          <div
            v-for="item in historyPlans"
            :key="item.plan_id"
            class="history-item"
            :class="{ 'history-item-failed': item.status === 'failed' }"
            role="button"
            tabindex="0"
            @click="handleHistoryItemClick(item)"
            @keydown.enter.prevent="handleHistoryItemClick(item)"
          >
            <div class="history-item-main">
              <div class="history-route">
                <span class="history-city">{{ item.city }}</span>
                <a-tag v-if="item.status === 'failed'" color="error" class="history-failed-tag">
                  {{ t('home.history.failedTag') }}
                </a-tag>
                <a-tag v-if="planLangLabel(item.language)" color="blue" class="history-lang-tag">
                  {{ planLangLabel(item.language) }}
                </a-tag>
                <span class="history-date">{{ item.start_date }} {{ t('common.to') }} {{ item.end_date }}</span>
              </div>
              <p class="history-meta">
                <span>Plan ID: {{ item.plan_id }}</span>
                <span v-if="item.status !== 'failed'">{{ item.travel_days }}{{ t('home.travelDaysUnit') }}</span>
                <span>{{ t('home.history.updatedAt') }} {{ formatHistoryTime(item.updated_at) }}</span>
              </p>
              <p v-if="item.status === 'failed' && item.error_message" class="history-error">
                {{ item.error_message }}
              </p>
              <p v-else-if="item.overall_suggestions" class="history-summary">{{ item.overall_suggestions }}</p>
            </div>
            <div class="history-item-actions">
              <a-popconfirm
                :title="t('home.history.deleteConfirm')"
                :ok-text="t('home.history.delete')"
                :cancel-text="t('home.history.cancel')"
                @confirm="removeHistoryPlan(item.plan_id)"
              >
                <button type="button" class="history-item-delete" @click.stop>
                  {{ t('home.history.delete') }}
                </button>
              </a-popconfirm>
              <span v-if="item.status !== 'failed'" class="history-open">{{ t('home.history.open') }}</span>
              <span v-else class="history-open history-open-failed">{{ t('home.history.failedHint') }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 悬浮设置按钮:配置 LLM / 地图 Key / Cookie 等运行时参数 -->
    <button
      type="button"
      class="floating-settings-btn"
      :title="t('settings.open')"
      :aria-label="t('settings.open')"
      @click="settingsVisible = true"
    >
      <svg width="22px" height="22px" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" aria-hidden="true"><path fill="currentColor" d="M600.704 64a32 32 0 0 1 30.464 22.208l35.2 109.376c14.784 7.232 28.928 15.36 42.432 24.512l112.384-24.192a32 32 0 0 1 34.432 15.36L944.32 364.8a32 32 0 0 1-4.032 37.504l-77.12 85.12a357.12 357.12 0 0 1 0 49.024l77.12 85.248a32 32 0 0 1 4.032 37.504l-88.704 153.6a32 32 0 0 1-34.432 15.296L708.8 803.904c-13.44 9.088-27.648 17.28-42.368 24.512l-35.264 109.376A32 32 0 0 1 600.704 960H423.296a32 32 0 0 1-30.464-22.208L357.696 828.48a351.616 351.616 0 0 1-42.56-24.64l-112.32 24.256a32 32 0 0 1-34.432-15.36L79.68 659.2a32 32 0 0 1 4.032-37.504l77.12-85.248a357.12 357.12 0 0 1 0-48.896l-77.12-85.248A32 32 0 0 1 79.68 364.8l88.704-153.6a32 32 0 0 1 34.432-15.296l112.32 24.256c13.568-9.152 27.776-17.408 42.56-24.64l35.2-109.312A32 32 0 0 1 423.232 64H600.64zm-23.424 64H446.72l-36.352 113.088-24.512 11.968a294.113 294.113 0 0 0-34.816 20.096l-22.656 15.36-116.224-25.088-65.28 113.152 79.68 88.192-1.92 27.136a293.12 293.12 0 0 0 0 40.192l1.92 27.136-79.808 88.192 65.344 113.152 116.224-25.024 22.656 15.296a294.113 294.113 0 0 0 34.816 20.096l24.512 11.968L446.72 896h130.688l36.48-113.152 24.448-11.904a288.282 288.282 0 0 0 34.752-20.096l22.592-15.296 116.288 25.024 65.28-113.152-79.744-88.192 1.92-27.136a293.12 293.12 0 0 0 0-40.256l-1.92-27.136 79.808-88.128-65.344-113.152-116.288 24.96-22.592-15.232a287.616 287.616 0 0 0-34.752-20.096l-24.448-11.904L577.344 128zM512 320a192 192 0 1 1 0 384 192 192 0 0 1 0-384zm0 64a128 128 0 1 0 0 256 128 128 0 0 0 0-256z"/></svg>
    </button>
    <TripSettingsModal v-model:open="settingsVisible" />
  </div>
</template>

<script setup lang="ts">
// 行程模块自带的全局样式(Paper Kit 暗色玻璃风格),随路由懒加载注入
import '@/assets/trip.css'
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import { getTripHistory, deleteTripPlan, isMemoryEnabled, setMemoryEnabled } from '@/api/trip'
import { getCurrentLocale } from '@/i18n'
import { useTripTaskStore } from '@/stores/tripTask'
import { findCuratedRoute, stopName } from '@/data/curatedRoutes'
import { useLanguageStore } from '@/stores/language'
import NavBar from '@/components/NavBar.vue'
import type { TripHistoryItem, CityStay } from '@/types/trip'
import type { Dayjs } from 'dayjs'

type LandingFormData = {
  cities: Array<{ city: string; days: number }>
  start_date: Dayjs | null
  transportation: string
  accommodation: string
  preferences: string[]
  free_text_input: string
  attraction_source: 'xhs' | 'douyin' | 'map'
}

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const langStore = useLanguageStore()
const tripTask = useTripTaskStore()

const scrollY = ref(0)
const formRef = ref<HTMLElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const panelHeight = ref<number | string>('auto')
const fogEnabled = ref(true)
const historyLoading = ref(false)
const historyPlans = ref<TripHistoryItem[]>([])

/** 计划生成语言的展示名(语言自称,无需翻译);旧记录无 language 字段时返回空不显示 */
const PLAN_LANG_LABELS: Record<string, string> = {
  zh: '🇨🇳 中文',
  en: '🇬🇧 English',
  ja: '🇯🇵 日本語',
}
const planLangLabel = (lang?: string) => (lang ? PLAN_LANG_LABELS[lang] || '' : '')

/** 运行时配置弹窗(LLM / 地图 Key / Cookie) */
const settingsVisible = ref(false)

/** 用户偏好记忆开关:持久化到 localStorage,提交行程时随请求发送 */
const memoryEnabled = ref(isMemoryEnabled())
watch(memoryEnabled, (val) => {
  setMemoryEnabled(val)
})

const interestOptions = [
  { value: '历史文化', labelKey: 'home.interests.history' },
  { value: '自然风光', labelKey: 'home.interests.nature' },
  { value: '美食', labelKey: 'home.interests.food' },
  { value: '购物', labelKey: 'home.interests.shopping' },
  { value: '艺术', labelKey: 'home.interests.art' },
  { value: '休闲', labelKey: 'home.interests.leisure' },
]

const formRules = computed(() => ({
  startDate: [{ required: true, message: t('home.startDateRequired') }],
}))

const formData = reactive<LandingFormData>({
  cities: [{ city: '', days: 2 }],
  start_date: null,
  transportation: '公共交通',
  accommodation: '经济型酒店',
  preferences: [],
  free_text_input: '',
  attraction_source: 'xhs',
})

const totalDays = computed(() => formData.cities.reduce((sum, cs) => sum + (cs.days || 1), 0))

/**
 * 精选路线预填：路线页「用 AI 生成同款行程」跳转 /trip?prefill=<路线id> 时，
 * 按精选路线填充城市天数、旅行偏好与自由文本，用户只需选出发日期。
 */
const applyCuratedRoutePrefill = () => {
  const prefillId = route.query.prefill
  const curated = findCuratedRoute(typeof prefillId === 'string' ? prefillId : undefined)
  if (!curated) return
  const routeTitle = langStore.t(`routes.items.${curated.id}.title`)
  formData.cities = [{ city: '成都', days: curated.days }]
  formData.preferences = [...curated.interests]
  formData.free_text_input = `参考精选路线「${routeTitle}」：${curated.stops.map(s => stopName(s, 'zh')).join(' → ')}`
  message.info(t('home.prefillApplied', { name: routeTitle }))
}

/**
 * 景点来源切换:抖音真人分享暂未实现,点击时保持原选项并提示开发中
 */
function onSourceChange(e: any) {
  const val = e?.target?.value as 'xhs' | 'douyin' | 'map'
  if (val === 'douyin') {
    message.info(t('home.douyinComingSoon'))
    return
  }
  formData.attraction_source = val
}

const computedEndDate = computed(() => {
  if (!formData.start_date) return null
  return formData.start_date.add(totalDays.value - 1, 'day')
})

const addCity = () => {
  if (formData.cities.length >= 5) return
  formData.cities.push({ city: '', days: 2 })
}

const removeCity = (index: number) => {
  if (formData.cities.length <= 1) return
  formData.cities.splice(index, 1)
}

const heroProgress = computed(() => Math.min(scrollY.value / 320, 1))
const toneProgress = computed(() => Math.min(Math.max((scrollY.value - 20) / 360, 0), 1))
const pageHeaderStyle = computed(() => ({
  backgroundImage: "url('/images/culture-scroll/era.png')",
  backgroundPosition: `center ${Math.max(-scrollY.value * 0.08, -120)}px`,
  backgroundSize: 'cover',
  backgroundRepeat: 'no-repeat',
}))
const movingCloudsStyle = computed(() => ({
  backgroundImage: "url('https://demos.creative-tim.com/paper-kit-2/assets/img/clouds.png')",
  opacity: fogEnabled.value ? '0.55' : '0',
}))
const fogLowStyle = computed(() => ({
  opacity: fogEnabled.value ? '0.82' : '0',
}))
const fogLowRightStyle = computed(() => ({
  opacity: fogEnabled.value ? '0.72' : '0',
}))
const heroContentStyle = computed(() => ({
  opacity: `${1 - heroProgress.value * 0.95}`,
  transform: `translate3d(0, ${-heroProgress.value * 46}px, 0)`,
}))
const heroBottomShadeStyle = computed(() => ({
  opacity: `${(0.48 + toneProgress.value * 0.44) * (fogEnabled.value ? 1 : 0)}`,
}))
const lowerShadeStyle = computed(() => ({
  opacity: `${(0.34 + toneProgress.value * 0.52) * (fogEnabled.value ? 1 : 0)}`,
}))
const formRevealStyle = computed(() => {
  const progress = Math.min(Math.max((scrollY.value - 80) / 340, 0), 1)
  return {
    opacity: `${0.2 + progress * 0.8}`,
    transform: `translate3d(0, ${(1 - progress) * 56}px, 0)`,
  }
})

const togglePreference = (value: string) => {
  const index = formData.preferences.indexOf(value)
  if (index === -1) formData.preferences.push(value)
  else formData.preferences.splice(index, 1)
}

const onScroll = () => {
  scrollY.value = window.scrollY || 0
}

const formatHistoryTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

const openHistoryPlan = (planId: string) => {
  if (!planId) return
  sessionStorage.removeItem('tripPlan')
  sessionStorage.removeItem('graphData')
  sessionStorage.setItem('planId', planId)
  router.push({ path: '/trip/result', query: { plan_id: planId } })
}

// 失败的历史计划不可回看,点击给出提示
const handleHistoryItemClick = (item: TripHistoryItem) => {
  if (item.status === 'failed') {
    message.warning(item.error_message || t('home.history.failedHint'))
    return
  }
  openHistoryPlan(item.plan_id)
}

const loadHistoryPlans = async () => {
  historyLoading.value = true
  try {
    historyPlans.value = await getTripHistory(8)
  } catch (error: any) {
    historyPlans.value = []
    message.error(error.message || t('home.history.loadFailed'))
  } finally {
    historyLoading.value = false
  }
}

// 删除一条落库的历史计划
const removeHistoryPlan = async (planId: string) => {
  if (!planId) return
  try {
    await deleteTripPlan(planId)
    message.success(t('home.history.deleteSuccess'))
    historyPlans.value = historyPlans.value.filter(item => item.plan_id !== planId)
  } catch (error: any) {
    message.error(error.message || t('home.history.deleteFailed'))
  }
}

onMounted(() => {
  onScroll()
  applyCuratedRoutePrefill()
  window.addEventListener('scroll', onScroll, { passive: true })
  void loadHistoryPlans()
})
onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})

const handleSubmit = async () => {
  // 校验：至少一个城市名非空
  const validCities = formData.cities.filter(cs => cs.city.trim())
  if (validCities.length === 0) {
    message.error(t('home.atLeastOneCity'))
    return
  }
  if (!formData.start_date) {
    message.error(t('home.messages.selectDate'))
    return
  }
  if (totalDays.value > 30) {
    message.warning(t('home.messages.travelDaysTooLong'))
    return
  }

  if (panelRef.value) {
    panelHeight.value = panelRef.value.offsetHeight
  }

  const citiesPayload: CityStay[] = validCities.map(cs => ({ city: cs.city.trim(), days: cs.days || 1 }))
  const endDate = computedEndDate.value!

  // 生成任务挂在全局 store 上：切换页面不会中断，完成后以 notification 提醒
  void tripTask.start({
    city: citiesPayload[0]!.city,
    cities: citiesPayload,
    start_date: formData.start_date.format('YYYY-MM-DD'),
    end_date: endDate.format('YYYY-MM-DD'),
    travel_days: totalDays.value,
    transportation: formData.transportation,
    accommodation: formData.accommodation,
    preferences: formData.preferences,
    free_text_input: formData.free_text_input,
    language: getCurrentLocale(),
    attraction_source: formData.attraction_source,
  })
}
</script>

<style scoped>
.landing-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #F8F3E9 0%, #F1EADB 58%, #F8F3E9 100%);
  color: #2E3A3D;
  position: relative;
  isolation: isolate;
  overflow-x: hidden; /* 防止水平溢出导致的出界感 */
}

/* 使用主站 NavBar：压过行程页 Paper Kit 全局 .navbar 样式干扰 */
.landing-page :deep(header.navbar) {
  z-index: 1100;
}

.lower-shade {
  position: fixed;
  inset: 0% 0 -1px 0;
  z-index: 0;
  pointer-events: none;
  background: transparent;
  transition: opacity 0.18s linear;
}

.lower-shade::before {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  top: -28px;
  height: 28px;
  background: transparent;
}

.landing-header {
  /* 确保 hero 区域占满全屏高度，背景图不重复 */
  height: 100vh;
  min-height: 100vh;
  position: relative;
  display: block;
  background-size: cover !important;
  background-repeat: no-repeat !important;
  background-position: center center !important;
  overflow: hidden;
  z-index: 1;
}

.history-section {
  position: relative;
  z-index: 1;
  padding: 0 24px 72px;
}

.history-panel {
  max-width: 1120px;
  margin: 0 auto;
  background: rgba(255, 253, 248, 0.9);
  border: 1px solid #E8EFED;
  border-radius: 28px;
  padding: 24px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  backdrop-filter: blur(14px);
}

.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.history-eyebrow {
  margin: 0 0 6px;
  color: #8A9A9E;
  font-size: 12px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.history-title {
  margin: 0;
  font-family: var(--font-display);
  color: var(--color-text-primary, #2E3A3D);
  font-size: 24px;
  font-weight: 700;
}

.history-refresh {
  padding-inline: 0;
}

.history-loading {
  color: #5E6E72;
  padding: 12px 4px;
}

.history-list {
  display: grid;
  gap: 14px;
}

.history-item {
  width: 100%;
  border: 1px solid #E8EFED;
  border-radius: 20px;
  background: #FFFDF8;
  color: inherit;
  padding: 18px 20px;
  text-align: left;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease;
}

.history-item:hover {
  transform: translateY(-1px);
  border-color: rgba(93, 164, 177, 0.4);
  background: #F6F0E5;
}

/* 失败的历史计划:置灰不可回看 */
.history-item-failed {
  opacity: 0.75;
  cursor: not-allowed;
}
.history-item-failed:hover {
  transform: none;
  border-color: rgba(184, 69, 62, 0.35);
  background: rgba(184, 69, 62, 0.04);
}
.history-failed-tag {
  margin-left: 8px;
  margin-right: 0;
  flex-shrink: 0;
}
.history-error {
  margin: 6px 0 0;
  color: #B8453E;
  font-size: 12px;
  line-height: 1.5;
}
.history-open-failed {
  color: #8A9A9E;
  cursor: not-allowed;
}

.history-item-main {
  min-width: 0;
  flex: 1;
}

.history-route {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 10px;
}

.history-city {
  color: #2E3A3D;
  font-size: 20px;
  font-weight: 700;
}

.history-date {
  color: #5E6E72;
  font-size: 14px;
}

.history-meta {
  margin: 8px 0 0;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: #8A9A9E;
  font-size: 13px;
}

.history-summary {
  margin: 10px 0 0;
  color: #5E6E72;
  font-size: 14px;
  line-height: 1.6;
}

.history-open {
  flex: none;
  color: #3E7D8A;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
}

.history-item-actions {
  flex: none;
  display: flex;
  align-items: center;
  gap: 12px;
}

.history-item-delete {
  flex: none;
  border: 1px solid rgba(184, 69, 62, 0.35);
  border-radius: 999px;
  background: transparent;
  color: rgba(184, 69, 62, 0.9);
  font-size: 13px;
  padding: 5px 14px;
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease, border-color 0.18s ease;
}

.history-item-delete:hover {
  background: rgba(184, 69, 62, 0.08);
  border-color: rgba(184, 69, 62, 0.6);
  color: #B8453E;
}

.landing-header .content-center {
  margin-top: 0 !important;
  height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  /* 内容垂直居中，确保在 .filter::after 霁罩和 hero-bottom-shade 之上 */
  position: relative;
  z-index: 3;
}

.landing-header .content-center .container {
  transform: translate3d(0, 45px, 0);
}

/* moving-clouds: 依赖 global.css 的定位 (bottom:0, width:250em, cloudLoop 80s) */
/* .landing-header .moving-clouds {
  transition: opacity 0.2s ease;
  pointer-events: none;
  z-index: 2;
} */

/* fog-low: 依赖 global.css 的定位 (margin-left:-35%, width:110%, bottom:0) */
.fog-low {
  pointer-events: none;
  z-index: 2;
  transition: opacity 0.2s ease;
  /* margin-bottom: -35px; */
}

/* fog-low.right: 依赖 global.css 的 margin-left:30%; opacity:1 */

.landing-hero-badge {
  margin: 0 0 18px;
  font-size: 12px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: rgba(236, 243, 250, 0.78);
}

.landing-header .presentation-title {
  font-family: var(--font-display);
  font-size: clamp(44px, 7vw, 90px);
  font-weight: 800;
  letter-spacing: var(--tracking-wide, 0.04em);
  color: #fffdf8;
  background: none;
  -webkit-background-clip: unset;
  background-clip: unset;
  -webkit-text-fill-color: #fffdf8;
  text-shadow: 0 2px 24px rgba(15, 13, 11, 0.55);
}

.landing-header .presentation-subtitle {
  max-width: 620px;
  font-family: var(--font-display);
  color: color-mix(in srgb, var(--color-gold-light) 72%, #fffdf8);
  font-size: clamp(15px, 1.8vw, 19px);
  line-height: 1.75;
  letter-spacing: var(--tracking-wide, 0.04em);
  text-shadow: 0 1px 14px rgba(15, 13, 11, 0.4);
  justify-self: center;
}

.hero-bottom-shade {
  position: absolute;
  inset: auto 0 0 0;
  height: 56%;
  z-index: 1;
  pointer-events: none;
  background: linear-gradient(
    to top,
    rgba(248, 243, 233, 0.95) 0%,
    rgba(248, 243, 233, 0.6) 46%,
    rgba(248, 243, 233, 0) 100%
  );
  transition: opacity 0.18s linear;
}


.form-section {
  margin-top: -112px;
  padding: 0 20px 86px;
  position: relative;
  z-index: 3;
}

.form-panel {
  max-width: 1000px;
  margin: 0 auto;
  border: 1.2px solid #E8EFED;
  border-radius: 22px;
  background: rgba(255, 253, 248, 0.92);
  backdrop-filter: blur(18px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  padding: 20px;
  transition: 0.25s;
}

.step {
  margin-bottom: 8px;
}

.step-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.step-head span {
  width: 26px;
  height: 23px;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--color-gold-glow, rgba(93, 164, 177, 0.15));
  border: 1.2px solid color-mix(in srgb, var(--color-gold) 42%, transparent);
  color: var(--color-gold-dark, #3E7D8A);
  font-size: 12px;
  font-weight: 700;
}

.step-head h3 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary, #2E3A3D);
}

.grid {
  display: grid;
  gap: 12px;
}

.grid4 {
  grid-template-columns: 1.5fr 1fr 1fr 0.8fr;
}

.grid-date {
  grid-template-columns: 1fr 0.6fr;
  margin-top: 12px;
}

.grid2 {
  grid-template-columns: 1fr 1fr;
}

.city-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 4px;
}

.city-row {
  display: flex;
  align-items: flex-end;
  gap: 10px;
}

.city-row-name {
  flex: 2;
  margin-bottom: 0;
}

.city-row-days {
  flex: 0.8;
  margin-bottom: 0;
}

.city-remove-btn {
  flex-shrink: 0;
  width: 36px;
  height: 40px;
  margin-bottom: 0;
  border: 1.2px solid #D9E2E0;
  border-radius: 10px;
  background: #FFFDF8;
  color: #8A9A9E;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.city-remove-btn:hover {
  border-color: rgba(184, 69, 62, 0.6);
  color: #B8453E;
  background: rgba(184, 69, 62, 0.1);
}

.city-add-btn {
  align-self: flex-start;
  padding: 6px 16px;
  border: 1.2px dashed rgba(184, 69, 62, 0.5);
  border-radius: 10px;
  background: transparent;
  color: rgba(184, 69, 62, 0.85);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.city-add-btn:hover {
  border-color: rgba(184, 69, 62, 0.9);
  background: rgba(184, 69, 62, 0.1);
  color: #B8453E;
}

.field-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #5E6E72;
}

.field-input.ant-input,
.field-input.ant-input-lg,
.field-input.ant-input-number,
.field-input.ant-input-number-lg,
.field-input.ant-picker,
.field-select :deep(.ant-select-selector),
.field-textarea :deep(textarea),
.field-textarea :deep(.ant-input),
.field-textarea.ant-input,
.special-textarea.ant-input {
  border: 1.2px solid #D9E2E0 !important;
  border-radius: 12px !important;
  background: #FFFFFF !important;
  background-color: #FFFFFF !important;
  background-image: none !important;
  color: #2E3A3D !important;
}

/* 浏览器自动填充（Autofill）背景色修复 */
:deep(.field-input.ant-input:-webkit-autofill),
:deep(.field-input.ant-input:-webkit-autofill:hover),
:deep(.field-input.ant-input:-webkit-autofill:focus),
:deep(.field-input.ant-input:-webkit-autofill:active),
:deep(.field-input .ant-picker-input > input:-webkit-autofill),
:deep(.field-textarea textarea:-webkit-autofill),
:deep(.special-textarea:-webkit-autofill) {
  -webkit-box-shadow: 0 0 0 1000px #FFFFFF inset !important;
  -webkit-text-fill-color: #2E3A3D !important;
  transition: background-color 5000s ease-in-out 0s !important;
}

.field-input.ant-input-number :deep(.ant-input-number-input),
.field-input.ant-input-number :deep(.ant-input-number-handler-wrap) {
  color: #2E3A3D !important;
}

.field-input.ant-input::placeholder,
:deep(.field-input .ant-picker-input > input::placeholder),
.field-textarea :deep(textarea::placeholder),
.field-textarea.ant-input::placeholder {
  color: rgba(138, 154, 158, 0.7) !important;
}

.field-input.ant-input:hover,
.field-input.ant-picker:hover,
.field-select:hover :deep(.ant-select-selector),
.field-textarea :deep(textarea:hover),
.field-textarea.ant-input:hover {
  border-color: rgba(93, 164, 177, 0.5) !important;
}

.field-input.ant-input:focus,
.field-input.ant-picker-focused,
.field-textarea :deep(textarea:focus),
.field-textarea.ant-input:focus {
  border-color: rgba(184, 69, 62, 0.55) !important;
  box-shadow: 0 0 0 3px rgba(184, 69, 62, 0.15) !important;
  background: #FFFFFF !important;
  outline: none !important;
}

:deep(.field-input .ant-picker-input > input),
.field-select :deep(.ant-select-selection-item),
:deep(.field-input .ant-picker-suffix),
:deep(.field-input .ant-picker-clear),
.field-select :deep(.ant-select-arrow) {
  color: #2E3A3D !important;
}

.days-chip {
  min-height: 40px;
  border-radius: 12px;
  border: 1.2px solid rgba(184, 69, 62, 0.35);
  background: #FFFDF8;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.days-number {
  color: #2E3A3D;
  font-size: 18px;
  line-height: 1;
  font-weight: 700;
}

.days-unit {
  font-size: 16px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #5E6E72;
  font-weight: 700;
}

.interest-grid {
  width: 100%;
}

.interest-group {
  display: grid !important;
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
  width: 100%;
}

.interest-group :deep(.ant-checkbox-wrapper) {
  display: none !important;
}

.interest-pill {
  min-height: 38px;
  border-radius: 10px;
  border: 1.2px solid #E8EFED;
  background: #FFFDF8;
  color: #5E6E72;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  user-select: none;
  transition: all 0.55s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.interest-pill:hover {
  background: #F6F0E5;
  border-color: #D9E2E0;
  /* transform: translateY(-2px); */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.interest-pill:active {
  transform: translateY(1px) scale(0.96);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.interest-pill.active {
  border-color: rgba(184, 69, 62, 0.8);
  background: rgba(184, 69, 62, 0.2);
  color: #B8453E;
}

.interest-pill.active:hover {
  background: rgba(184, 69, 62, 0.28);
  border-color: rgba(184, 69, 62, 1);
}

/* 景点来源选择按钮组 */
.source-radio {
  display: flex;
  width: 100%;
}

.source-radio :deep(.ant-radio-button-wrapper) {
  flex: 1;
  text-align: center;
  background: #FFFDF8;
  border-color: #E8EFED !important;
  color: #5E6E72;
}

.source-radio :deep(.ant-radio-button-wrapper:first-child) {
  border-radius: 10px 0 0 10px;
}

.source-radio :deep(.ant-radio-button-wrapper:last-child) {
  border-radius: 0 10px 10px 0;
}

.source-radio :deep(.ant-radio-button-wrapper:not(:first-child))::before {
  background: #E8EFED;
}

.source-radio :deep(.ant-radio-button-wrapper-checked) {
  border-color: rgba(184, 69, 62, 0.55) !important;
  background: rgba(184, 69, 62, 0.08);
  color: #B8453E;
}

.source-radio :deep(.ant-radio-button-wrapper-checked::before) {
  background: rgba(184, 69, 62, 0.5) !important;
}

.source-hint {
  margin: 8px 2px 0;
  font-size: 12px;
  line-height: 1.6;
  color: #8A9A9E;
}

/* 用户偏好记忆开关 */
.memory-toggle {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 12px;
  background: rgba(93, 164, 177, 0.06);
  border: 1px solid rgba(93, 164, 177, 0.18);
}

.memory-toggle-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.memory-toggle-label {
  font-size: 13px;
  font-weight: 600;
  color: #2E3A3D;
}

.memory-toggle-desc {
  font-size: 12px;
  line-height: 1.6;
  color: #8A9A9E;
}

.submit-btn {
  width: 100%;
  min-height: 48px;
  border-radius: 12px;
  /* border: 1px solid rgba(236, 243, 250, 0.28);
  background: linear-gradient(135deg, #d76e42, #a14625);
  color: #fff; */
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  cursor: pointer;
}

.submit-btn.loading {
  background: rgba(184, 69, 62, 0.7);
  cursor: wait;
}

.loading-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.spinner {
  width: 15px;
  height: 15px;
  border-radius: 50%;
  border: 2px solid rgba(236, 243, 250, 0.24);
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}

.spinner-small {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2.5px solid rgba(184, 69, 62, 0.24);
  border-top-color: #B8453E;
  animation: spin 0.8s linear infinite;
}

/* 节点动画相关样式 */
.stepper-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 480px;
  animation: fadeIn 0.4s ease;
  padding: 30px 20px;
  box-sizing: border-box;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.stepper-header {
  text-align: center;
  margin-bottom: 50px;
}

.stepper-title {
  font-size: 28px;
  font-weight: 700;
  color: #2E3A3D;
  margin-bottom: 8px;
  letter-spacing: 0.05em;
}

.stepper-subtitle {
  font-size: 15px;
  color: #8A9A9E;
}

.stepper-container {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  width: 100%;
  max-width: 680px;
  margin: 0 auto 50px auto;
}

.step-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100px;
  z-index: 2;
}

.node-icon {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: #FFFFFF;
  border: 1.5px solid #E8EFED;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
  color: #8A9A9E;
  transition: all 0.35s ease;
}

.step-node.active .node-icon {
  border-color: #B8453E;
  background: rgba(184, 69, 62, 0.08);
  color: #B8453E;
  box-shadow: 0 0 16px rgba(184, 69, 62, 0.2);
}

.step-node.completed .node-icon {
  background: #B8453E;
  border-color: #B8453E;
  color: #fff;
  box-shadow: 0 0 12px rgba(184, 69, 62, 0.2);
}

.node-text {
  font-size: 12px;
  font-weight: 600;
  color: #8A9A9E;
  text-align: center;
  transition: color 0.35s ease;
  line-height: 1.3;
}

.step-node.active .node-text {
  color: #B8453E;
}

.step-node.completed .node-text {
  color: #2E3A3D;
}

.step-divider {
  flex: 1;
  height: 3px;
  background: #E8EFED;
  margin-top: 25px; /* (52px / 2) - 1.5px */
  border-radius: 2px;
  position: relative;
  overflow: hidden;
}

.step-divider::after {
  content: '';
  position: absolute;
  top: 0; left: 0; bottom: 0; width: 0%;
  background: #B8453E;
  transition: width 0.45s ease;
}

.step-divider.completed::after {
  width: 100%;
}

.stepper-footer {
  text-align: center;
  margin-top: 10px;
}

.stepper-footer h3 {
  font-size: 20px;
  font-weight: 600;
  color: #B8453E;
  margin-bottom: 8px;
}

.stepper-footer p {
  font-size: 14px;
  color: #8A9A9E;
}

:deep(.ant-form-item-label > label) {
  color: transparent !important;
}

:deep(.ant-form-item-explain-error) {
  color: #B8453E !important;
}

/* @keyframes cloudLoop {
  from {
    transform: translate3d(0, 0, 0);
  }
  to {
    transform: translate3d(-50%, 0, 0);
  }
} */

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1080px) {
  .grid4 {
    grid-template-columns: 1fr 1fr;
  }

  .interest-group {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 991px) {
  .form-section {
    padding: 0 14px 72px;
  }

  .form-panel {
    padding: 22px 18px;
  }

  .grid4,
  .grid2,
  .grid-date {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 520px) {
  .landing-header .presentation-title {
    font-size: clamp(34px, 10vw, 52px);
  }

  .landing-header .presentation-subtitle {
    font-size: 14px;
    padding: 0 10px;
  }

  .landing-header .content-center .container {
    transform: translate3d(0, 16px, 0);
  }

  .interest-group {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* 悬浮设置按钮:固定在右下角,不随页面滚动 */
.floating-settings-btn {
  position: fixed;
  right: 22px;
  bottom: 22px;
  z-index: 1040;
  width: 46px;
  height: 46px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.35);
  border-radius: 50%;
  background: rgba(20, 32, 38, 0.72);
  color: rgba(255, 255, 255, 0.92);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.28);
  cursor: pointer;
  transition: background 0.2s ease, transform 0.2s ease;
}

.floating-settings-btn:hover {
  background: rgba(34, 52, 60, 0.9);
  transform: translateY(-2px);
}

@media (max-width: 520px) {
  .floating-settings-btn {
    right: 14px;
    bottom: 14px;
    width: 40px;
    height: 40px;
  }
}
</style>
