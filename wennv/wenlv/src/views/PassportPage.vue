<script setup lang="ts">
/**
 * PassportPage.vue — 数字足迹护照(/passport)
 * 浏览景点详情即"盖章"（详见 src/utils/passport.ts），本页展示集章进度：
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
import { getStamps } from '@/utils/passport'

const router = useRouter()
const langStore = useLanguageStore()

const scenics = ref<ScenicItem[]>([])
const loading = ref(true)
const loadFailed = ref(false)
/** 集章数据读取版本号：onMounted 时自增以触发重算（返回本页时刷新集章状态） */
const stampsVersion = ref(0)

onMounted(async () => {
  try {
    const page = await listScenics({ page: 1, page_size: 100 })
    scenics.value = page.items
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
  stampsVersion.value++
})

/** 已盖章景点 id 集合（随版本号重读 localStorage） */
const stampIds = computed(() => {
  void stampsVersion.value
  return new Set(getStamps().map((s) => s.id))
})

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
          <p v-if="collected === 0" class="passport-progress__empty">
            {{ langStore.t('passport.stampEmpty') }}
          </p>
        </div>

        <!-- 集章网格 -->
        <div class="stamp-grid">
          <button
            v-for="item in scenics"
            :key="item.id"
            type="button"
            class="stamp-cell"
            :class="{ 'stamp-cell--got': stampIds.has(item.id) }"
            @click="goScenic(item)"
          >
            <span class="stamp-cell__mark" aria-hidden="true">
              <span v-if="stampIds.has(item.id)" class="stamp-cell__seal">印</span>
              <span v-else class="stamp-cell__unknown">？</span>
            </span>
            <span class="stamp-cell__name">{{ nameOf(item) }}</span>
          </button>
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

    <PassportModal v-model:open="modalOpen" :scenics="scenics" />
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

/* ── 集章网格 ── */
.stamp-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: var(--space-4);
}

.stamp-cell {
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
