<script setup lang="ts">
/**
 * HomeView.vue — 成都文旅推广首页
 * 布局：Hero → 轮播图 → 文化数据统计 → 文化名片 → 文明脉络 → 非遗传承 → Footer
 */
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useLanguageStore } from '@/stores/language'
import NavBar from '@/components/NavBar.vue'
import PandaCursor from '@/components/PandaCursor.vue'
import Carousel from '@/components/Carousel.vue'
import AppIcon from '@/components/AppIcon.vue'
import HeroRipple from '@/components/HeroRipple.vue'
import CultureScroll from '@/components/CultureScroll.vue'
import type { CarouselItem } from '@/components/Carousel.vue'

const langStore = useLanguageStore()
const heroRippleRef = ref<InstanceType<typeof HeroRipple> | null>(null)

function scrollToExplore() {
  document.getElementById('explore')?.scrollIntoView({ behavior: 'smooth' })
}

function onCtaPointer(e: PointerEvent) {
  if (e.pointerType === 'touch') return
  heroRippleRef.value?.rippleAtClient?.(e.clientX, e.clientY)
}

const carouselItems: CarouselItem[] = [
  {
    id: 'panda',
    imageUrl: '/images/home/carousel-panda.jpg',
    titleZh: '大熊猫繁育研究基地',
    titleEn: 'Giant Panda Base',
    subtitleZh: '近距离观察国宝大熊猫，感受自然之美',
    subtitleEn: 'See giant pandas up close in their natural habitat',
  },
  {
    id: 'kuanzhai',
    imageUrl: '/images/home/carousel-kuanzhai.jpg',
    titleZh: '宽窄巷子',
    titleEn: 'Kuanzhai Alleys',
    subtitleZh: '漫步清朝古街，品茗听戏，感受成都慢生活',
    subtitleEn: "Stroll Qing Dynasty alleys, sip tea, and feel Chengdu's slow pace",
  },
  {
    id: 'dujiangyan',
    imageUrl: '/images/home/carousel-dujiangyan.jpg',
    titleZh: '都江堰 · 青城山',
    titleEn: 'Dujiangyan & Mt. Qingcheng',
    subtitleZh: '千年水利工程与道教发源地的完美融合',
    subtitleEn: 'Ancient irrigation wonder meets the birthplace of Taoism',
  },
  {
    id: 'jinli',
    imageUrl: '/images/home/carousel-jinli.jpg',
    titleZh: '锦里 · 武侯祠',
    titleEn: 'Jinli & Wuhou Shrine',
    subtitleZh: '三国文化圣地，西蜀最古老的商业街',
    subtitleEn: "Three Kingdoms heritage on western Sichuan's oldest street",
  },
  {
    id: 'xiling',
    imageUrl: '/images/home/carousel-xiling.jpg',
    titleZh: '西岭雪山',
    titleEn: 'Xiling Snow Mountain',
    subtitleZh: '"窗含西岭千秋雪"——诗圣杜甫笔下的雪山胜景',
    subtitleEn: 'The snow-capped peak immortalized by poet Du Fu',
  },
]

/* ──── 巴蜀文化内容 ──── */

/* 文化名片 */
const cultureCards = [
  { key: 'brocade', icon: 'brocade' },
  { key: 'opera', icon: 'opera' },
  { key: 'tea', icon: 'tea' },
  { key: 'cuisine', icon: 'cuisine' },
  { key: 'history', icon: 'history' },
  { key: 'embroidery', icon: 'embroidery' },
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

/* 滚动进入视口动画 — 文化区块在滚动到视口时才显现 */
const cultureSections = ref<HTMLElement[]>([])
let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.remove('is-leaving')
          entry.target.classList.add('is-visible')
        } else if (entry.target.classList.contains('is-visible')) {
          entry.target.classList.remove('is-visible')
          entry.target.classList.add('is-leaving')
        }
      })
    },
    { threshold: 0.1, rootMargin: '0px 0px -40px 0px' }
  )
  cultureSections.value.forEach((el) => observer?.observe(el))
})

onBeforeUnmount(() => {
  observer?.disconnect()
})

function setSectionRef(el: unknown, index: number) {
  if (el instanceof HTMLElement) {
    cultureSections.value[index] = el
  }
}
</script>

<template>
  <div class="homepage shu-pattern">
    <NavBar />
    <PandaCursor />

    <!-- ──── HERO ──── -->
    <section class="hero">
      <div class="hero__bg">
        <div
          class="hero__photo"
          role="img"
          aria-label="成都风景"
        />
        <HeroRipple ref="heroRippleRef" />
        <div class="hero__gradient" />
        <div class="hero__pattern" />
      </div>

      <!-- 蜀字水印 — 签名元素 -->
      <div class="hero__shu" aria-hidden="true">蜀</div>

      <div class="hero__content container-wide">
        <div class="hero__badge">
          <span class="hero__badge-diamond">◈</span>
          <span class="hero__badge-text">{{ langStore.lang === 'zh' ? '四川省 · 成都' : 'Chengdu · Sichuan' }}</span>
          <span class="hero__badge-diamond">◈</span>
        </div>

        <h1 class="hero__title">
          <span class="hero__title-zh-wrap">
            <span class="hero__title-zh">{{ langStore.lang === 'zh' ? '巴蜀文化' : 'Bashu Culture' }}</span>
            <span class="hero__title-sign" aria-hidden="true">蜀</span>
          </span>
          <span class="hero__title-en">{{ langStore.lang === 'zh' ? '锦绣天府' : 'Splendid Tianfu' }}</span>
        </h1>

        <p class="hero__subtitle">
          {{ langStore.t('hero.subtitle') }}
        </p>

        <div class="hero__cta-group">
          <button
            class="hero__cta hero__cta--primary"
            @click="scrollToExplore"
            @pointerenter="onCtaPointer"
            @pointermove="onCtaPointer"
          >
            {{ langStore.t('hero.cta') }}
            <span class="hero__cta-arrow">→</span>
          </button>
        </div>
      </div>
    </section>
    <!-- ──── 轮播图 ──── -->
    <section id="explore" class="carousel-section">
      <div class="container">
        <Carousel :items="carouselItems" :interval="4000" />
      </div>
    </section>
    <!-- ──── 引言 + 数据统计 ──── -->
    <section
      :ref="(el) => setSectionRef(el, 0)"
      class="culture-stats reveal"
    >
      <div class="container">
        <p class="culture-intro__text">{{ langStore.t('culture.intro') }}</p>
        <div class="culture-stats__grid">
          <div
            v-for="(stat, i) in [
              { number: '5000+', label: langStore.t('culture.stats.history') },
              { number: '89', label: langStore.t('culture.stats.heritage') },
              { number: '262', label: langStore.t('culture.stats.landmark') },
              { number: '3亿+', label: langStore.t('culture.stats.visitors') },
            ]"
            :key="stat.label"
            class="culture-stat"
            :style="{ '--delay': `${i * 0.1}s` }"
          >
            <span class="culture-stat__number">{{ stat.number }}</span>
            <span class="culture-stat__label">{{ stat.label }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ──── 文化名片 ──── -->
    <section
      id="culture-cards"
      :ref="(el) => setSectionRef(el, 1)"
      class="culture-cards reveal"
    >
      <div class="container">
        <div class="culture-cards__header">
          <h2 class="section-title">{{ langStore.t('culture.cardsTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('culture.cardsSubtitle') }}</p>
        </div>

        <div class="culture-cards__grid">
          <article
            v-for="(card, index) in cultureCards"
            :key="card.key"
            class="culture-card"
            :style="{ '--delay': `${index * 0.08}s` }"
          >
            <div class="culture-card__icon">
              <AppIcon :name="card.icon" :size="32" />
            </div>
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
    <section
      :ref="(el) => setSectionRef(el, 2)"
      class="culture-scroll-section reveal"
    >
      <div class="culture-scroll-section__header container">
        <h2 class="section-title">{{ langStore.t('culture.timelineTitle') }}</h2>
        <p class="section-subtitle">{{ langStore.t('culture.timelineSubtitle') }}</p>
      </div>
      <CultureScroll />
    </section>

    <!-- ──── 非遗传承 ──── -->
    <section
      id="culture-heritage"
      :ref="(el) => setSectionRef(el, 3)"
      class="culture-heritage reveal"
    >
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

    <!-- ──── Footer ──── -->
    <footer class="footer">
      <div class="container">
        <div class="footer__inner">
          <div class="footer__brand">
            <span class="footer__logo">蜀韵·成都</span>
            <span class="footer__logo-en">Shu·Chengdu</span>
          </div>
          <div class="footer__info">
            <p>{{ langStore.t('footer.copyright') }}</p>
            <p>{{ langStore.t('footer.rights') }}</p>
          </div>
        </div>
        <div class="footer__divider" />
        <div class="footer__bottom">
          <span>© 2026</span>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* ========================================
   HERO
   ======================================== */
.hero {
  position: relative;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.hero__photo {
  position: absolute;
  inset: 0;
  z-index: 0;
  background-image: url('/images/home/hero-chengdu.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  transform: scale(1.02);
}

.hero__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.hero__gradient {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  background:
    linear-gradient(
      180deg,
      rgba(15, 13, 11, 0.55) 0%,
      rgba(15, 13, 11, 0.35) 45%,
      var(--color-bg) 100%
    );
}

.hero__pattern {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  background-image:
    repeating-conic-gradient(
      transparent 0deg 89deg,
      rgba(201, 169, 110, 0.02) 90deg 91deg,
      transparent 91deg 179deg,
      rgba(201, 169, 110, 0.02) 180deg 181deg
    );
  background-size: 80px 80px;
  opacity: 0.25;
}

.hero__content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--space-6);
  padding: var(--space-20) var(--space-8);
}

/* 蜀字水印 — 签名元素 */
.hero__shu {
  position: absolute;
  font-family: var(--font-display);
  font-size: clamp(280px, 42vw, 560px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px rgba(201, 169, 110, 0.10);
  user-select: none;
  pointer-events: none;
  z-index: 0;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  opacity: 0.9;
}

.hero__badge {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.hero__badge-diamond {
  font-size: var(--text-xs);
  color: var(--color-gold);
  opacity: 0.7;
}

.hero__badge-text {
  font-family: var(--font-body);
  font-size: var(--text-xs);
  color: #fffdf8;
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
  text-shadow: 0 1px 12px rgba(15, 13, 11, 0.55);
}

.hero__title {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

.hero__title-zh-wrap {
  position: relative;
  display: inline-block;
}

.hero__title-zh {
  font-family: var(--font-display);
  font-size: var(--text-6xl);
  font-weight: 900;
  color: #fffdf8;
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
  text-shadow: 0 2px 24px rgba(15, 13, 11, 0.65);
}

.hero__title-en {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 400;
  font-style: italic;
  color: #fffdf8;
  letter-spacing: var(--tracking-wider);
  text-shadow: 0 2px 20px rgba(15, 13, 11, 0.6);
}

/* 落款印章 — 挂在「巴蜀文化」右上角，不占文档流 */
.hero__title-sign {
  position: absolute;
  top: 0;
  right: 0;
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
  transform: translate(55%, -45%) rotate(-2deg);
  pointer-events: none;
}

.hero__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-lg);
  color: rgba(255, 253, 248, 0.92);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
  max-width: 520px;
  line-height: var(--leading-relaxed);
  text-shadow: 0 1px 16px rgba(15, 13, 11, 0.55);
}

.hero__cta-group {
  margin-top: var(--space-4);
}

.hero__cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-10);
  border-radius: var(--radius-sm);
  font-family: 'SimSun', '宋体', 'Songti SC', 'Noto Serif SC', serif;
  font-size: var(--text-2xl);
  font-weight: 700;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-base);
}

.hero__cta--primary {
  background: transparent;
  color: #fff;
  font-weight: 700;
}

.hero__cta--primary:hover {
  background: transparent;
  color: #fff;
  transform: translateY(-4px);
}

.hero__cta-arrow {
  transition: transform var(--transition-fast);
}

.hero__cta--primary:hover .hero__cta-arrow {
  transform: translateX(4px);
}

@keyframes float-down {
  0%, 100% { transform: translateY(0); opacity: 1; }
  50% { transform: translateY(8px); opacity: 0.5; }
}

/* ========================================
   轮播图区
   ======================================== */
.carousel-section {
  padding: var(--space-16) 0;
  background: var(--color-bg-alt);
}

/* ========================================
   文化探索 — 过渡分隔区
   衔接"景点轮播"与"文化内容"，让页面过渡自然
   ======================================== */
.culture-bridge {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-6);
  padding: var(--space-4) var(--space-8);
  background: var(--color-bg-alt);
}

.culture-bridge__line {
  flex: none;
  width: 200px;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--color-border-light));
}

.culture-bridge__line:last-child {
  background: linear-gradient(90deg, var(--color-border-light), transparent);
}

.culture-bridge__ornament {
  font-size: var(--text-xs);
  color: var(--color-gold);
  opacity: 0.8;
  animation: bridge-glow 3s ease-in-out infinite;
}

@keyframes bridge-glow {
  0%, 100% { opacity: 0.4; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.15); }
}

/* ========================================
   文化 — 滚动进出动画
   区块进入视口后浮现；滚动离开视口时平滑退场
   ======================================== */
.reveal {
  opacity: 0;
  transform: translateY(28px);
  transition: opacity 0.7s ease, transform 0.7s ease;
}

.reveal.is-visible {
  opacity: 1;
  transform: translateY(0);
}

/* 离开视口：整体向上轻移并淡出 */
.reveal.is-leaving {
  opacity: 0;
  transform: translateY(-18px) scale(0.985);
  transition-duration: 0.55s;
}

/* 子元素离场时同步消散，避免内容瞬间消失 */
.reveal.is-leaving .culture-stat,
.reveal.is-leaving .culture-card,
.reveal.is-leaving .heritage-card {
  transition: opacity 0.35s ease, transform 0.35s ease;
  opacity: 0;
  transform: translateY(14px);
}

/* ========================================
   引言（并入数据统计区块）
   ======================================== */
.culture-intro__text {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  max-width: 720px;
  margin: 0 auto var(--space-10);
  font-weight: 400;
}

/* ========================================
   数据统计
   ======================================== */
.culture-stats {
  padding: var(--space-16) 0;
  background: var(--color-bg-alt);
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
  opacity: 0;
  transform: translateY(20px);
}

.reveal.is-visible .culture-stat {
  animation: card-fade-in 0.6s ease forwards;
  animation-delay: var(--delay);
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
   文化名片
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

.reveal.is-visible .culture-card {
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
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
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
   文明脉络 — 长卷
   ======================================== */
.culture-scroll-section {
  padding-top: var(--space-16);
  background: var(--color-bg-alt);
}

/* Sticky pin must not sit under a transformed ancestor. */
.culture-scroll-section.reveal,
.culture-scroll-section.reveal.is-visible,
.culture-scroll-section.reveal.is-leaving {
  opacity: 1;
  transform: none;
}

.culture-scroll-section__header {
  text-align: center;
  margin-bottom: var(--space-8);
}

/* ========================================
   非遗传承
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

.reveal.is-visible .heritage-card {
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
   FOOTER
   ======================================== */
.footer {
  padding: var(--space-12) 0 var(--space-8);
  border-top: 1px solid var(--color-border);
  margin-top: var(--space-8);
  background: var(--color-bg-alt);
}

.footer__inner {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-8);
}

.footer__brand {
  display: flex;
  flex-direction: column;
}

.footer__logo {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
}

.footer__logo-en {
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wider);
}

.footer__info {
  text-align: right;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  line-height: var(--leading-relaxed);
}

.footer__divider {
  height: 1px;
  background: var(--color-border);
  margin: var(--space-6) 0;
}

.footer__bottom {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   Responsive
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
  .hero__title-zh {
    font-size: var(--text-4xl);
  }
  .hero__title-en {
    font-size: var(--text-2xl);
  }
  .hero__subtitle {
    font-size: var(--text-base);
  }
  .hero__content {
    padding: var(--space-16) var(--space-4);
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
  .footer__inner {
    flex-direction: column;
    text-align: center;
  }
  .footer__info {
    text-align: center;
  }
}

@media (max-width: 480px) {
  .hero__title-zh {
    font-size: var(--text-3xl);
  }
  .hero__title-en {
    font-size: var(--text-2xl);
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
