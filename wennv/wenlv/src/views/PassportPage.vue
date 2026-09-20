<script setup lang="ts">
/**
 * PassportPage.vue — 数字足迹护照(/passport)
 * 集章数据来自后端 check_ins 表（MySQL 永久保存）：用户在景点详情页
 * 上传现场照片打卡后获得印章；本页展示集章进度：
 * 已集章显示红章 + 景点名，未集章显示灰色"？"剪影；
 * 任意进度均可生成「成都旅行护照」证书海报（PassportModal）。
 */
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import NavBar from '@/components/NavBar.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import PassportModal from '@/components/PassportModal.vue'
import { listScenics, type ScenicItem } from '@/api/content'
import { pickName } from '@/utils/storyI18n'
import { fetchStamps } from '@/utils/passport'
import { loadQuizBadges, quizRankOf } from '@/utils/quizBadges'
import { hasToken } from '@/utils/token'

const router = useRouter()
const langStore = useLanguageStore()

const scenics = ref<ScenicItem[]>([])
const loading = ref(true)
const loadFailed = ref(false)
/** 集章数据来自后端,登录后按用户账号永久保存 */
const stamps = ref<{ id: number }[]>([])
const loggedIn = ref(hasToken())

onMounted(async () => {
  try {
    const page = await listScenics({ page: 1, page_size: 100 })
    scenics.value = page.items
  } catch {
    loadFailed.value = true
  }
  if (loggedIn.value) {
    try {
      stamps.value = await fetchStamps()
    } catch {
      /* 拉取失败按空集章展示 */
    }
  }
  loading.value = false
})

/** 已盖章景点 id 集合 */
const stampIds = computed(() => new Set(stamps.value.map((s) => s.id)))

const collected = computed(() =>
  scenics.value.filter((s) => stampIds.value.has(s.id)).length
)
const total = computed(() => scenics.value.length)
const progressPercent = computed(() =>
  total.value ? Math.round((collected.value / total.value) * 100) : 0
)

const nameOf = (item: ScenicItem) => pickName(item, langStore.lang)

function goScenic(item: ScenicItem) {
  router.push(`/scenic/${item.id}`)
}

/* ── 徽章墙:蜀文化知识闯关答题得徽章(localStorage 本地存档,不绑账号) ── */
const quizBadges = ref(loadQuizBadges())
/** 已获徽章数 = 答对过题目的景点数 */
const badgeCount = computed(() => quizBadges.value.spots.length)
/** 文化段位:按累计答对数(0-1 青铜 / 2-3 白银 / 4-5 黄金) */
const quizRankKey = computed(() => quizRankOf(quizBadges.value.correctTotal).key)

/* ── 证书弹窗 ── */
const modalOpen = ref(false)
</script>

<template>
  <div class="passport-page">
    <NavBar />

    <!-- ──── HERO ──── -->
    <HomeBanner
      :eyebrow="langStore.t('passport.eyebrow')"
      :title="langStore.t('passport.title')"
      :subtitle="langStore.t('passport.hero')"
      watermark="章"
    />

    <!-- ──── 集章面板 ──── -->
    <section class="passport-page__panel container">
      <div v-if="loading" class="passport-page__state">
        {{ langStore.t('scenic.loading') }}
      </div>

      <template v-else>
        <div v-if="loadFailed" class="passport-page__state passport-page__state--error">
          {{ langStore.t('passport.loadFailed') }}
        </div>

        <!-- 进度 -->
        <div class="passport-progress">
          <div class="passport-progress__row">
            <span class="passport-progress__count">
              {{ langStore.t('passport.progress', { count: collected, total }) }}
            </span>
          </div>
          <div class="passport-progress__bar">
            <div
              class="passport-progress__fill"
              :style="{ width: `${progressPercent}%` }"
            />
          </div>
          <p v-if="collected === 0 && loggedIn" class="passport-progress__empty">
            {{ langStore.t('passport.stampEmpty') }}
          </p>

          <!-- 未登录:集章需登录后按账号永久保存 -->
          <div v-if="!loggedIn" class="passport-login">
            <p class="passport-login__text">{{ langStore.t('passport.loginRequired') }}</p>
            <button type="button" class="passport-login__btn" @click="router.push('/login')">
              {{ langStore.t('passport.loginCta') }}
            </button>
          </div>
        </div>

        <!-- 集章网格:已集章景点格显示"答题"角标,点击进入景点详情答题闯关 -->
        <div class="stamp-grid">
          <button
            v-for="item in scenics"
            :key="item.id"
            type="button"
            class="stamp-cell"
            :class="{ 'stamp-cell--got': stampIds.has(item.id) }"
            @click="goScenic(item)"
          >
            <span
              v-if="stampIds.has(item.id)"
              class="stamp-cell__quiz"
              :title="langStore.t('quiz.title')"
            >
              {{ langStore.t('quiz.goQuiz') }}
            </span>
            <span class="stamp-cell__mark" aria-hidden="true">
              <span v-if="stampIds.has(item.id)" class="stamp-cell__seal">印</span>
              <span v-else class="stamp-cell__unknown">？</span>
            </span>
            <span class="stamp-cell__name">{{ nameOf(item) }}</span>
          </button>
        </div>

        <!-- 徽章墙:答题得徽章 + 文化段位 -->
        <div class="badge-wall">
          <div class="badge-wall__head">
            <span class="badge-wall__icon" aria-hidden="true">🎖️</span>
            <h3 class="badge-wall__title">{{ langStore.t('quiz.badgeWall') }}</h3>
          </div>
          <div class="badge-wall__stats">
            <span class="badge-wall__count">
              {{ langStore.t('quiz.badgeCount', { count: badgeCount }) }}
            </span>
            <span class="badge-wall__rank">
              {{ langStore.t(quizRankKey) }}
            </span>
          </div>
          <p class="badge-wall__hint">
            {{ langStore.t('quiz.badgeTotalCorrect', { count: quizBadges.correctTotal }) }}
          </p>
        </div>

        <!-- 生成证书 -->
        <div class="passport-page__actions">
          <button
            type="button"
            class="passport-page__generate"
            :disabled="total === 0"
            @click="modalOpen = true"
          >
            {{ langStore.t('passport.generate') }}
          </button>
        </div>
      </template>
    </section>

    <PassportModal v-model:open="modalOpen" :scenics="scenics" :stamp-ids="stampIds" />
  </div>
</template>

<style scoped>
/* 顶部固定 NavBar 的占位补偿 */
.passport-page {
  min-height: 100vh;
  padding-top: var(--nav-height);
  background: var(--color-bg);
}

.passport-page__panel {
  padding-top: var(--space-10);
  padding-bottom: var(--space-16);
}

.passport-page__state {
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  padding: var(--space-10) 0;
}

.passport-page__state--error {
  color: var(--color-cinnabar);
}

/* ── 进度 ── */
.passport-progress {
  max-width: 560px;
  margin: 0 auto var(--space-8);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.passport-progress__row {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: var(--space-3);
}

.passport-progress__count {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wide);
}

.passport-progress__bar {
  height: 8px;
  border-radius: var(--radius-full);
  background: var(--color-border-light);
  overflow: hidden;
}

.passport-progress__fill {
  height: 100%;
  border-radius: var(--radius-full);
  background: linear-gradient(90deg, var(--color-gold), #2d6a4f);
  transition: width var(--transition-slow);
}

.passport-progress__empty {
  margin: 0;
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* 未登录提示 */
.passport-login {
  margin-top: var(--space-4);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-5);
  border: 1px dashed color-mix(in srgb, var(--color-gold) 50%, transparent);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--color-gold) 5%, var(--color-surface));
}

.passport-login__text {
  margin: 0;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  text-align: center;
}

.passport-login__btn {
  padding: var(--space-2) var(--space-8);
  border: none;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-gold), var(--color-gold-dark));
  color: var(--color-bg);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.passport-login__btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px var(--color-gold-glow);
}

/* ── 集章网格 ── */
.stamp-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: var(--space-4);
}

.stamp-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-5) var(--space-3);
  background: var(--color-surface);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base), transform var(--transition-base),
    box-shadow var(--transition-base);
}

/* "答题"角标:已集章景点可进入知识闯关赢徽章 */
.stamp-cell__quiz {
  position: absolute;
  top: -8px;
  right: -6px;
  padding: 2px var(--space-2);
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-gold), var(--color-gold-dark));
  color: var(--color-text-inverse);
  font-size: var(--text-xs);
  font-weight: 700;
  letter-spacing: var(--tracking-wide);
  box-shadow: 0 2px 8px var(--color-gold-glow);
  pointer-events: none;
}

.stamp-cell:hover {
  transform: translateY(-3px);
}

.stamp-cell--got {
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 40%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 4%, var(--color-surface));
}

.stamp-cell--got:hover {
  border-color: var(--color-cinnabar);
  box-shadow: 0 8px 24px var(--color-cinnabar-dim), var(--shadow-lg);
}

.stamp-cell__mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
}

/* 红章：圆形双环 + 印字 */
.stamp-cell__seal {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: var(--radius-full);
  border: 2px solid var(--color-cinnabar);
  box-shadow: inset 0 0 0 3px var(--color-surface), inset 0 0 0 4px color-mix(in srgb, var(--color-cinnabar) 60%, transparent);
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 800;
  color: var(--color-cinnabar);
  transform: rotate(-8deg);
}

/* 未集章：灰色"？"剪影 */
.stamp-cell__unknown {
  font-family: var(--font-display);
  font-size: var(--text-3xl);
  font-weight: 900;
  color: color-mix(in srgb, var(--color-text-muted) 40%, transparent);
  user-select: none;
}

.stamp-cell__name {
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  line-height: var(--leading-normal);
}

.stamp-cell--got .stamp-cell__name {
  color: var(--color-text-primary);
}

/* ── 徽章墙 ── */
.badge-wall {
  max-width: 560px;
  margin: var(--space-10) auto 0;
  padding: var(--space-6);
  border: 1px solid color-mix(in srgb, var(--color-gold) 40%, transparent);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--color-gold) 5%, var(--color-surface));
}

.badge-wall__head {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.badge-wall__icon {
  font-size: var(--text-xl);
}

.badge-wall__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wide);
}

.badge-wall__stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
}

.badge-wall__count {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.badge-wall__rank {
  padding: var(--space-1) var(--space-4);
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-gold), var(--color-gold-dark));
  color: var(--color-text-inverse);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 700;
  letter-spacing: var(--tracking-wider);
}

.badge-wall__hint {
  margin: var(--space-3) 0 0;
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ── 生成证书 ── */
.passport-page__actions {
  display: flex;
  justify-content: center;
  margin-top: var(--space-10);
}

.passport-page__generate {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-10);
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  color: var(--color-bg);
  background: linear-gradient(135deg, var(--color-gold), var(--color-gold-dark));
  border: none;
  border-radius: var(--radius-full);
  box-shadow: 0 8px 24px var(--color-gold-glow);
  cursor: pointer;
  transition: transform var(--transition-base), box-shadow var(--transition-base);
}

.passport-page__generate:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 12px 32px var(--color-gold-glow), var(--shadow-lg);
}

.passport-page__generate:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .stamp-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 640px) {
  .passport-page__panel {
    padding-top: var(--space-6);
    padding-bottom: var(--space-10);
  }
  .stamp-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-3);
  }
  .stamp-cell {
    padding: var(--space-4) var(--space-2);
  }
  .stamp-cell__mark,
  .stamp-cell__seal {
    width: 52px;
    height: 52px;
  }
}
</style>
