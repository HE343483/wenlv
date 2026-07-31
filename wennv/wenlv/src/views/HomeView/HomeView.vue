<script setup lang="ts">
/**
 * HomeView.vue — 成都文旅推广首页
 * 布局：Hero → 轮播图 → Footer
 */
import { useLanguageStore } from '@/stores/language'
import NavBar from '@/components/NavBar.vue'
import Carousel from '@/components/Carousel.vue'
import type { CarouselItem } from '@/components/Carousel.vue'

const langStore = useLanguageStore()

function scrollToExplore() {
  document.getElementById('explore')?.scrollIntoView({ behavior: 'smooth' })
}

const carouselItems: CarouselItem[] = [
  {
    id: 'panda',
    imageUrl: '',
    titleZh: '大熊猫繁育研究基地',
    titleEn: 'Giant Panda Base',
    subtitleZh: '近距离观察国宝大熊猫，感受自然之美',
    subtitleEn: 'See giant pandas up close in their natural habitat',
  },
  {
    id: 'kuanzhai',
    imageUrl: '',
    titleZh: '宽窄巷子',
    titleEn: 'Kuanzhai Alleys',
    subtitleZh: '漫步清朝古街，品茗听戏，感受成都慢生活',
    subtitleEn: 'Stroll Qing Dynasty alleys, sip tea, and feel Chengdu\'s slow pace',
  },
  {
    id: 'dujiangyan',
    imageUrl: '',
    titleZh: '都江堰 · 青城山',
    titleEn: 'Dujiangyan & Mt. Qingcheng',
    subtitleZh: '千年水利工程与道教发源地的完美融合',
    subtitleEn: 'Ancient irrigation wonder meets the birthplace of Taoism',
  },
  {
    id: 'jinli',
    imageUrl: '',
    titleZh: '锦里 · 武侯祠',
    titleEn: 'Jinli & Wuhou Shrine',
    subtitleZh: '三国文化圣地，西蜀最古老的商业街',
    subtitleEn: 'Three Kingdoms heritage on western Sichuan\'s oldest street',
  },
  {
    id: 'xiling',
    imageUrl: '',
    titleZh: '西岭雪山',
    titleEn: 'Xiling Snow Mountain',
    subtitleZh: '"窗含西岭千秋雪"——诗圣杜甫笔下的雪山胜景',
    subtitleEn: 'The snow-capped peak immortalized by poet Du Fu',
  },
]
</script>

<template>
  <div class="homepage shu-pattern">
    <NavBar />

    <!-- ──── HERO ──── -->
    <section class="hero">
      <div class="hero__bg">
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
          <span class="hero__title-zh">{{ langStore.lang === 'zh' ? '巴蜀文化' : 'Bashu Culture' }}</span>
          <span class="hero__title-en-row">
            <span class="hero__title-en">{{ langStore.lang === 'zh' ? '锦绣天府' : 'Splendid Tianfu' }}</span>
            <!-- 落款印章 — 呼应全站"决策点即印章"的语言 -->
            <span class="hero__title-sign" aria-hidden="true">蜀</span>
          </span>
        </h1>

        <p class="hero__subtitle">
          {{ langStore.t('hero.subtitle') }}
        </p>

        <div class="hero__cta-group">
          <button class="hero__cta hero__cta--primary" @click="scrollToExplore">
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
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.hero__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.hero__gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 60% at 50% 40%, rgba(201, 169, 110, 0.06) 0%, transparent 70%),
    radial-gradient(ellipse 60% 50% at 30% 80%, rgba(162, 59, 59, 0.03) 0%, transparent 60%),
    radial-gradient(ellipse 50% 40% at 70% 20%, rgba(122, 138, 122, 0.03) 0%, transparent 50%),
    linear-gradient(180deg, rgba(15, 13, 11, 0.7) 0%, var(--color-bg) 100%);
}

.hero__pattern {
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
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.hero__title {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

.hero__title-zh {
  font-family: var(--font-display);
  font-size: var(--text-6xl);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
}

.hero__title-en {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 400;
  font-style: italic;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
}

.hero__title-en-row {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
}

/* 落款印章 — 金线描边，安静平和 */
.hero__title-sign {
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

.hero__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-lg);
  color: var(--color-text-secondary);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
  max-width: 520px;
  line-height: var(--leading-relaxed);
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
  font-size: var(--text-base);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-base);
}

.hero__cta--primary {
  background: var(--color-gold);
  color: var(--color-bg);
  font-weight: 500;
}

.hero__cta--primary:hover {
  background: var(--color-gold-light);
  box-shadow: 0 0 32px var(--color-gold-glow);
  transform: translateY(-1px);
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

/* 轮播图区*/
.carousel-section {
  padding: var(--space-16) 0;
  margin-bottom: 100px;

}

/* ========================================
   FOOTER
   ======================================== */
.footer {
  padding: var(--space-12) 0 var(--space-8);
  border-top: 1px solid var(--color-border);
  margin-top: var(--space-8);
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
}
</style>
