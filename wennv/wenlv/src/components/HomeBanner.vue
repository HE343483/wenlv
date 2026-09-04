<script setup lang="ts">
/**
 * HomeBanner.vue — 统一欢迎横幅
 * 适用：首页、探索、美食、路线等 `/home` 子页面
 */

defineProps<{
  eyebrow: string
  title: string
  subtitle?: string
  enTitle?: string
  watermark: string
}>()
</script>

<template>
  <section class="home-banner shu-pattern">
    <div class="home-banner__bg" aria-hidden="true">
      <div class="home-banner__gradient" />
    </div>
    <span class="home-banner__watermark" aria-hidden="true">{{ watermark }}</span>
    <div class="home-banner__content">
      <p class="home-banner__eyebrow">
        <span>◈</span>
        {{ eyebrow }}
        <span>◈</span>
      </p>
      <h1 class="home-banner__title">{{ title }}</h1>
      <p v-if="enTitle" class="home-banner__en-title">{{ enTitle }}</p>
      <p v-if="subtitle" class="home-banner__subtitle">{{ subtitle }}</p>
      <div v-if="$slots.default" class="home-banner__extra">
        <slot />
      </div>
    </div>
  </section>
</template>

<style scoped>
.home-banner {
  position: relative;
  overflow: hidden;
  text-align: center;
  padding: var(--space-16) 0 var(--space-10);
  min-height: 340px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.home-banner__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.home-banner__gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 72% 58% at 50% 38%, var(--color-gold-glow) 0%, transparent 68%),
    radial-gradient(ellipse 44% 36% at 18% 62%, var(--color-cinnabar-dim) 0%, transparent 60%);
}

.home-banner__watermark {
  position: absolute;
  font-family: var(--font-display);
  font-size: clamp(180px, 32vw, 420px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1.4px var(--banner-watermark-stroke, rgba(201, 169, 110, 0.16));
  text-shadow: 0 0 18px var(--banner-watermark-glow, rgba(201, 169, 110, 0.08));
  user-select: none;
  pointer-events: none;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  opacity: 0.95;
  z-index: 0;
}

.home-banner__content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-6) var(--space-4);
}

.home-banner__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.home-banner__title {
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3.5rem);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.08;
}

.home-banner__en-title {
  font-family: var(--font-en-display);
  font-style: italic;
  font-size: var(--text-lg);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
  font-weight: 400;
  margin-top: calc(-1 * var(--space-2));
}

.home-banner__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
  max-width: 42rem;
}

/* 额外内容（如探索页搜索栏、路线页统计） */
.home-banner__extra {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  width: 100%;
  margin-top: var(--space-2);
}

/* 浅色主题：金色描边加深，保证水印可见 */
:global(.theme-light) .home-banner__watermark {
  --banner-watermark-stroke: rgba(133, 106, 46, 0.34);
  --banner-watermark-glow: rgba(133, 106, 46, 0.12);
}

@media (max-width: 640px) {
  .home-banner {
    min-height: 260px;
    padding: var(--space-10) 0 var(--space-8);
  }

  .home-banner__content {
    padding: var(--space-4);
  }

  .home-banner__subtitle {
    font-size: var(--text-sm);
  }
}
</style>
