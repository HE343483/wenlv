<script setup lang="ts">
/**
 * CulturePage.vue — 巴蜀文化页面
 * 内容：英雄区 → 数据统计 → 文化名片 → 文明脉络 → 非遗传承
 */
import { ref, onMounted } from 'vue'
import { useLanguageStore } from '@/stores/language'

const langStore = useLanguageStore()

/* 滚动动画 */
const animated = ref(false)
onMounted(() => {
  setTimeout(() => { animated.value = true }, 200)
})

/* 文化名片 */
const cultureCards = [
  { key: 'brocade', icon: '🧵' },
  { key: 'opera', icon: '🎭' },
  { key: 'tea', icon: '🍵' },
  { key: 'cuisine', icon: '🌶' },
  { key: 'history', icon: '🏺' },
  { key: 'embroidery', icon: '🪡' },
]

/* 文明脉络 — 时间线 */
const timeline = [
  {
    era: '古蜀时期',
    eraEn: 'Ancient Shu',
    period: '约公元前1600年 — 公元前316年',
    periodEn: 'c. 1600 BC — 316 BC',
    desc: '三星堆与金沙遗址代表了古蜀文明的辉煌成就，青铜神树、黄金面具、太阳神鸟等文物震惊世界。',
    descEn: 'Sanxingdui and Jinsha sites represent the brilliance of ancient Shu — bronze trees, gold masks, and the Sun Bird.',
  },
  {
    era: '秦并巴蜀',
    eraEn: 'Qin Annexation',
    period: '公元前316年 — 公元221年',
    periodEn: '316 BC — 221 AD',
    desc: '秦灭巴蜀后，李冰父子修建都江堰水利工程，使成都平原成为"天府之国"。',
    descEn: 'After Qin conquered Shu, Li Bing and his son built the Dujiangyan irrigation system, turning Chengdu into the "Land of Abundance."',
  },
  {
    era: '蜀汉风云',
    eraEn: 'Shu Han',
    period: '公元221年 — 263年',
    periodEn: '221 — 263 AD',
    desc: '刘备在成都称帝建立蜀汉政权，诸葛亮六出祁山，三国文化自此深深烙印在成都的血脉中。',
    descEn: 'Liu Bei founded Shu Han in Chengdu. Zhuge Liang\'s northern campaigns etched Three Kingdoms legacy into the city\'s soul.',
  },
  {
    era: '唐宋锦绣',
    eraEn: 'Tang & Song',
    period: '公元618年 — 1279年',
    periodEn: '618 — 1279 AD',
    desc: '"锦官城"之名响彻天下，杜甫、陆游等诗人留居成都，留下无数传世诗篇。蜀锦、蜀绣、雕版印刷空前繁荣。',
    descEn: 'Chengdu flourished as the "City of Brocade." Poets Du Fu and Lu You lived here, leaving timeless verses.',
  },
  {
    era: '明清延续',
    eraEn: 'Ming & Qing',
    period: '公元1368年 — 1911年',
    periodEn: '1368 — 1911 AD',
    desc: '宽窄巷子、锦里等明清古街格局形成，川剧、川菜、茶馆文化日趋成熟，成都慢生活文化源远流长。',
    descEn: 'Kuanzhai Alley and Jinli took shape. Sichuan opera, cuisine, and teahouse culture matured into the city\'s signature slow pace.',
  },
  {
    era: '现代成都',
    eraEn: 'Modern Chengdu',
    period: '1911年 — 至今',
    periodEn: '1911 — Present',
    desc: '千年古都焕发新生，以公园城市理念建设践行新发展理念，天府新区、大运会场馆见证城市蝶变。',
    descEn: 'The ancient capital reinvents itself as a "Park City." Tianfu New Area and global events showcase Chengdu\'s transformation.',
  },
]

/* 非遗项目 */
const heritageItems = [
  {
    name: '蜀锦织造技艺',
    nameEn: 'Shu Brocade Weaving',
    category: '国家级非遗',
    categoryEn: 'National ICH',
    desc: '蜀锦是中国四大名锦之一，以经线彩条和纬线彩条起花为特色，被誉为"东方瑰宝"。',
    descEn: 'Shu brocade is one of China\'s four great brocades, renowned for its warp and weft color techniques.',
  },
  {
    name: '蜀绣',
    nameEn: 'Shu Embroidery',
    category: '国家级非遗',
    categoryEn: 'National ICH',
    desc: '蜀绣以针法严谨、色彩明快、形象生动著称，与苏绣、湘绣、粤绣并称中国四大名绣。',
    descEn: 'Known for precise stitching, vivid colors and lifelike images — one of China\'s four great embroidery styles.',
  },
  {
    name: '川剧',
    nameEn: 'Sichuan Opera',
    category: '国家级非遗',
    categoryEn: 'National ICH',
    desc: '川剧以高腔、变脸、吐火等绝技闻名，是巴蜀文化最具代表性的表演艺术形式之一。',
    descEn: 'Famous for high-pitched singing, face-changing, and fire-spitting — the most iconic Bashu performing art.',
  },
  {
    name: '川菜烹饪技艺',
    nameEn: 'Sichuan Cuisine',
    category: '国家级非遗',
    categoryEn: 'National ICH',
    desc: '川菜以"一菜一格，百菜百味"著称，麻、辣、鲜、香层次丰富，享誉全球。',
    descEn: 'Known as "one dish, one style; a hundred dishes, a hundred flavors" — numbing, spicy, fresh and aromatic.',
  },
  {
    name: '成都漆艺',
    nameEn: 'Chengdu Lacquer Art',
    category: '国家级非遗',
    categoryEn: 'National ICH',
    desc: '成都漆艺历史悠久，以雕花填彩、银片镶嵌等工艺著称，被誉为"漆器之乡"。',
    descEn: 'Chengdu lacquerware has a long history, famed for carved fill and silver inlay techniques.',
  },
  {
    name: '青城山道教音乐',
    nameEn: 'Mt. Qingcheng Taoist Music',
    category: '省级非遗',
    categoryEn: 'Provincial ICH',
    desc: '青城山道教音乐融合了道教仪轨与川西民间音乐元素，古朴悠远，意境空灵。',
    descEn: 'Blending Taoist ritual with western Sichuan folk music — ancient, ethereal and meditative.',
  },
]
</script>

<template>
  <div class="culture-page">
    <!-- ──── HERO ──── -->
    <section class="culture-hero">
      <div class="culture-hero__bg">
        <div class="culture-hero__gradient" />
        <div class="culture-hero__pattern" />
      </div>
      <div class="culture-hero__shu" aria-hidden="true">蜀</div>

      <div class="culture-hero__content">
        <div class="culture-hero__badge">
          <span class="culture-hero__badge-diamond">◈</span>
          <span class="culture-hero__badge-text">{{ langStore.t('culture.badge') }}</span>
          <span class="culture-hero__badge-diamond">◈</span>
        </div>

        <h1 class="culture-hero__title">
          <span class="culture-hero__title-zh">{{ langStore.t('culture.title') }}</span>
          <span class="culture-hero__title-en-row">
            <span class="culture-hero__title-en">{{ langStore.t('culture.titleEn') }}</span>
            <span class="culture-hero__title-sign" aria-hidden="true">蜀</span>
          </span>
        </h1>

        <p class="culture-hero__subtitle">{{ langStore.t('culture.subtitle') }}</p>
      </div>
    </section>

    <!-- ──── 引言 ──── -->
    <section class="culture-intro">
      <div class="container">
        <p class="culture-intro__text">{{ langStore.t('culture.intro') }}</p>
      </div>
    </section>

    <!-- ──── 数据统计 ──── -->
    <section class="culture-stats">
      <div class="container">
        <div class="culture-stats__grid">
          <div class="culture-stat">
            <span class="culture-stat__number">5000+</span>
            <span class="culture-stat__label">{{ langStore.t('culture.stats.history') }}</span>
          </div>
          <div class="culture-stat">
            <span class="culture-stat__number">89</span>
            <span class="culture-stat__label">{{ langStore.t('culture.stats.heritage') }}</span>
          </div>
          <div class="culture-stat">
            <span class="culture-stat__number">262</span>
            <span class="culture-stat__label">{{ langStore.t('culture.stats.landmark') }}</span>
          </div>
          <div class="culture-stat">
            <span class="culture-stat__number">3亿+</span>
            <span class="culture-stat__label">{{ langStore.t('culture.stats.visitors') }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ──── 文化名片 ──── -->
    <section class="culture-cards">
      <div class="container">
        <div class="culture-cards__header">
          <h2 class="section-title">{{ langStore.t('culture.cardsTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('culture.cardsSubtitle') }}</p>
        </div>

        <div class="culture-cards__grid">
          <article
            v-for="card in cultureCards"
            :key="card.key"
            class="culture-card"
            :class="{ 'culture-card--visible': animated }"
            :style="{ '--delay': `${cultureCards.indexOf(card) * 0.08}s` }"
          >
            <div class="culture-card__icon">{{ card.icon }}</div>
            <h3 class="culture-card__name">
              {{ langStore.t(`culture.card.${card.key}.name`) }}
            </h3>
            <p class="culture-card__en">{{ langStore.t(`culture.card.${card.key}.en`) }}</p>
            <p class="culture-card__desc">{{ langStore.t(`culture.card.${card.key}.desc`) }}</p>
            <button class="culture-card__action">
              {{ langStore.t('culture.exploreBtn') }}
              <span class="culture-card__arrow">→</span>
            </button>
          </article>
        </div>
      </div>
    </section>

    <!-- ──── 文明脉络 ──── -->
    <section class="culture-timeline">
      <div class="container">
        <div class="culture-timeline__header">
          <h2 class="section-title">{{ langStore.t('culture.timelineTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('culture.timelineSubtitle') }}</p>
        </div>

        <div class="culture-timeline__track">
          <div class="culture-timeline__line" />
          <div
            v-for="(item, i) in timeline"
            :key="item.era"
            class="culture-timeline__item"
            :class="{ 'culture-timeline__item--visible': animated }"
            :style="{ '--delay': `${i * 0.1}s` }"
          >
            <div class="culture-timeline__dot">
              <span class="culture-timeline__dot-inner" />
            </div>
            <div class="culture-timeline__content">
              <div class="culture-timeline__meta">
                <span class="culture-timeline__era">
                  {{ langStore.lang === 'zh' ? item.era : item.eraEn }}
                </span>
                <span class="culture-timeline__period">
                  {{ langStore.lang === 'zh' ? item.period : item.periodEn }}
                </span>
              </div>
              <p class="culture-timeline__desc">
                {{ langStore.lang === 'zh' ? item.desc : item.descEn }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ──── 非遗传承 ──── -->
    <section class="culture-heritage">
      <div class="container">
        <div class="culture-heritage__header">
          <h2 class="section-title">{{ langStore.t('culture.heritageTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('culture.heritageSubtitle') }}</p>
        </div>

        <div class="culture-heritage__grid">
          <div
            v-for="(item, i) in heritageItems"
            :key="item.name"
            class="heritage-card"
            :class="{ 'heritage-card--visible': animated }"
            :style="{ '--delay': `${i * 0.08}s` }"
          >
            <div class="heritage-card__badge">
              {{ langStore.lang === 'zh' ? item.category : item.categoryEn }}
            </div>
            <h3 class="heritage-card__name">
              {{ langStore.lang === 'zh' ? item.name : item.nameEn }}
            </h3>
            <p class="heritage-card__desc">
              {{ langStore.lang === 'zh' ? item.desc : item.descEn }}
            </p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   HERO
   ======================================== */
.culture-hero {
  position: relative;
  min-height: 50vh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: var(--space-8) 0;
}

.culture-hero__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.culture-hero__gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 60% at 50% 40%, rgba(201, 169, 110, 0.08) 0%, transparent 70%),
    radial-gradient(ellipse 60% 50% at 30% 80%, rgba(162, 59, 59, 0.04) 0%, transparent 60%),
    radial-gradient(ellipse 50% 40% at 70% 20%, rgba(122, 138, 122, 0.04) 0%, transparent 50%),
    linear-gradient(180deg, rgba(15, 13, 11, 0.5) 0%, var(--color-bg) 100%);
}

.culture-hero__pattern {
  position: absolute;
  inset: 0;
  background-image:
    repeating-conic-gradient(
      transparent 0deg 89deg,
      rgba(201, 169, 110, 0.02) 90deg 91deg,
      transparent 91deg 179deg,
      rgba(201, 169, 110, 0.02) 180deg 181deg
    );
  background-size: 80px 80px;
  opacity: 0.5;
}

.culture-hero__content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--space-6);
  padding: var(--space-16) var(--space-8);
}

.culture-hero__shu {
  position: absolute;
  font-family: var(--font-display);
  font-size: clamp(200px, 32vw, 400px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px rgba(201, 169, 110, 0.08);
  user-select: none;
  pointer-events: none;
  z-index: 0;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  opacity: 0.9;
}

.culture-hero__badge {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-2);
}

.culture-hero__badge-diamond {
  font-size: var(--text-xs);
  color: var(--color-gold);
  opacity: 0.7;
}

.culture-hero__badge-text {
  font-family: var(--font-body);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.culture-hero__title {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

.culture-hero__title-zh {
  font-family: var(--font-display);
  font-size: var(--text-5xl);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
}

.culture-hero__title-en {
  font-family: var(--font-en-display);
  font-size: var(--text-2xl);
  font-weight: 400;
  font-style: italic;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
}

.culture-hero__title-en-row {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
}

.culture-hero__title-sign {
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-gold);
  border: 1px solid color-mix(in srgb, var(--color-gold) 42%, transparent);
  width: 34px;
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transform: rotate(-2deg);
}

.culture-hero__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-lg);
  color: var(--color-text-secondary);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
  max-width: 480px;
  line-height: var(--leading-relaxed);
}

/* ========================================
   INTRO
   ======================================== */
.culture-intro {
  padding: var(--space-8) 0 var(--space-16);
}

.culture-intro__text {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  max-width: 720px;
  margin: 0 auto;
  font-weight: 400;
}

/* ========================================
   STATS
   ======================================== */
.culture-stats {
  padding: 0 0 var(--space-16);
}

.culture-stats__grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-6);
}

.culture-stat {
  text-align: center;
  padding: var(--space-8) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.culture-stat:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
}

.culture-stat__number {
  display: block;
  font-family: var(--font-en-display);
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
  margin-bottom: var(--space-2);
}

.culture-stat__label {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   CULTURE CARDS
   ======================================== */
.culture-cards {
  padding: var(--space-16) 0;
}

.culture-cards__header {
  text-align: center;
  margin-bottom: var(--space-10);
}

.culture-cards__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.culture-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  transition: all var(--transition-base);
  opacity: 0;
  transform: translateY(20px);
}

.culture-card--visible {
  animation: card-fade-in 0.6s ease forwards;
  animation-delay: var(--delay);
}

@keyframes card-fade-in {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.culture-card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg);
}

.culture-card__icon {
  font-size: 2.5rem;
  line-height: 1;
  margin-bottom: var(--space-1);
}

.culture-card__name {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.culture-card__en {
  font-family: var(--font-en-display);
  font-size: var(--text-sm);
  font-style: italic;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
}

.culture-card__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  flex: 1;
}

.culture-card__action {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  padding: var(--space-2) 0;
  margin-top: var(--space-2);
  transition: color var(--transition-fast);
  border-top: 1px solid var(--color-border);
  padding-top: var(--space-4);
}

.culture-card__action:hover {
  color: var(--color-gold);
}

.culture-card__arrow {
  transition: transform var(--transition-fast);
}

.culture-card__action:hover .culture-card__arrow {
  transform: translateX(3px);
}

/* ========================================
   TIMELINE
   ======================================== */
.culture-timeline {
  padding: var(--space-16) 0;
  background: var(--color-bg-alt);
}

.culture-timeline__header {
  text-align: center;
  margin-bottom: var(--space-10);
}

.culture-timeline__track {
  position: relative;
  max-width: 720px;
  margin: 0 auto;
}

.culture-timeline__line {
  position: absolute;
  left: 20px;
  top: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, var(--color-gold), var(--color-border), var(--color-gold));
}

.culture-timeline__item {
  position: relative;
  padding-left: 56px;
  padding-bottom: var(--space-8);
  opacity: 0;
  transform: translateX(-10px);
}

.culture-timeline__item:last-child {
  padding-bottom: 0;
}

.culture-timeline__item--visible {
  animation: timeline-fade-in 0.6s ease forwards;
  animation-delay: var(--delay);
}

@keyframes timeline-fade-in {
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.culture-timeline__dot {
  position: absolute;
  left: 12px;
  top: 4px;
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.culture-timeline__dot-inner {
  width: 10px;
  height: 10px;
  border-radius: var(--radius-full);
  background: var(--color-gold);
  box-shadow: 0 0 12px var(--color-gold-glow);
}

.culture-timeline__content {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-5);
  transition: border-color var(--transition-base);
}

.culture-timeline__content:hover {
  border-color: var(--color-gold-dark);
}

.culture-timeline__meta {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
  flex-wrap: wrap;
}

.culture-timeline__era {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.culture-timeline__period {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
}

.culture-timeline__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

/* ========================================
   HERITAGE
   ======================================== */
.culture-heritage {
  padding: var(--space-16) 0;
}

.culture-heritage__header {
  text-align: center;
  margin-bottom: var(--space-10);
}

.culture-heritage__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.heritage-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  transition: all var(--transition-base);
  opacity: 0;
  transform: translateY(16px);
}

.heritage-card--visible {
  animation: card-fade-in 0.6s ease forwards;
  animation-delay: var(--delay);
}

.heritage-card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px var(--color-gold-glow);
}

.heritage-card__badge {
  display: inline-block;
  font-size: var(--text-xs);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  background: var(--color-gold-glow);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
  align-self: flex-start;
  border: 1px solid color-mix(in srgb, var(--color-gold) 30%, transparent);
}

.heritage-card__name {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.heritage-card__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  flex: 1;
}

/* ========================================
   RESPONSIVE
   ======================================== */
@media (max-width: 1024px) {
  .culture-cards__grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .culture-heritage__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .culture-hero__title-zh {
    font-size: var(--text-4xl);
  }
  .culture-hero__title-en {
    font-size: var(--text-xl);
  }
  .culture-hero__content {
    padding: var(--space-12) var(--space-4);
  }
  .culture-stats__grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .culture-cards__grid {
    grid-template-columns: 1fr;
  }
  .culture-heritage__grid {
    grid-template-columns: 1fr;
  }
  .culture-intro__text {
    font-size: var(--text-base);
  }
}

@media (max-width: 480px) {
  .culture-hero__title-zh {
    font-size: var(--text-3xl);
  }
  .culture-stats__grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-4);
  }
  .culture-stat {
    padding: var(--space-6) var(--space-3);
  }
  .culture-stat__number {
    font-size: var(--text-3xl);
  }
}
</style>