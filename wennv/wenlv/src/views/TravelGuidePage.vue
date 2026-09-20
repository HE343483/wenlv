<script setup lang="ts">
/**
 * TravelGuidePage.vue — 入境游实用工具箱(/guide)
 * 面向首次来华国际游客的三语"落地生存指南":支付 / 交通 / 住宿 / 通讯 / 应急 / 礼仪
 * 六大板块,每条为"标题 + 一两句说明",信息密度低、可扫读。
 * 内容全部走 i18n(guide.*);条目为数组,经语言字典读取(RoutesPage 同款模式)。
 */
import { computed } from 'vue'
import { useLanguageStore } from '@/stores/language'
import NavBar from '@/components/NavBar.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import dictZh from '@/locales/zh'
import dictEn from '@/locales/en'
import dictJa from '@/locales/ja'

const langStore = useLanguageStore()

interface GuideItem {
  title: string
  desc: string
}

/** 语言字典(读取 guide.* 下的条目数组等非字符串值) */
const guideDicts = { zh: dictZh, en: dictEn, ja: dictJa } as const

/** 六大板块:图标 + i18n 键(key 对应 guide.<key>.title / guide.<key>.items) */
const cardMeta = [
  { key: 'payment', icon: '💳' },
  { key: 'transport', icon: '🚄' },
  { key: 'hotel', icon: '🏨' },
  { key: 'telecom', icon: '📱' },
  { key: 'emergency', icon: '🚨' },
  { key: 'etiquette', icon: '🙏' },
] as const

/** 按当前语言组装板块视图模型(标题走 t(),条目数组走字典) */
const sections = computed(() =>
  cardMeta.map(({ key, icon }) => {
    const dict = guideDicts[langStore.lang] as unknown as {
      guide: Record<string, { title: string; items: GuideItem[] }>
    }
    const section = dict.guide[key]
    return { key, icon, title: section?.title ?? '', items: section?.items ?? [] }
  })
)
</script>

<template>
  <div class="guide-page shu-pattern">
    <NavBar />

    <!-- ──── HERO ──── -->
    <HomeBanner
      :eyebrow="langStore.t('guide.eyebrow')"
      :title="langStore.t('guide.title')"
      :subtitle="langStore.t('guide.hero')"
      watermark="行"
    />

    <!-- ──── 六大板块卡片 ──── -->
    <section class="guide-page__grid container">
      <article
        v-for="s in sections"
        :key="s.key"
        class="guide-card"
        :class="{ 'guide-card--alert': s.key === 'emergency' }"
      >
        <header class="guide-card__head">
          <span class="guide-card__icon" aria-hidden="true">{{ s.icon }}</span>
          <h2 class="guide-card__title">{{ langStore.t(`guide.${s.key}.title`) }}</h2>
        </header>

        <ul class="guide-card__list">
          <li v-for="item in s.items" :key="item.title" class="guide-card__item">
            <h3 class="guide-card__item-title">{{ item.title }}</h3>
            <p class="guide-card__item-desc">{{ item.desc }}</p>
          </li>
        </ul>
      </article>
    </section>
  </div>
</template>

<style scoped>
/* 顶部固定 NavBar 的占位补偿 */
.guide-page {
  min-height: 100vh;
  padding-top: var(--nav-height);
  background: var(--color-bg);
}

/* ── 板块网格 ── */
.guide-page__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
  padding-top: var(--space-10);
  padding-bottom: var(--space-16);
}

.guide-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.guide-card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg);
}

/* 应急板块:朱砂色强调,提示"这一卡最要紧" */
.guide-card--alert {
  border-color: color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
}

.guide-card--alert:hover {
  border-color: var(--color-cinnabar);
  box-shadow: 0 8px 24px var(--color-cinnabar-dim), var(--shadow-lg);
}

.guide-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--color-border-light);
}

.guide-card__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-2xl);
  line-height: 1;
}

.guide-card__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.guide-card--alert .guide-card__title {
  color: var(--color-cinnabar);
}

.guide-card__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.guide-card__item-title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--space-1);
}

.guide-card__item-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .guide-page__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .guide-page__grid {
    grid-template-columns: 1fr;
    padding-top: var(--space-6);
    padding-bottom: var(--space-10);
  }
}
</style>
