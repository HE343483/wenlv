<script setup lang="ts">
/**
 * AuthBamboo.vue — 认证页墨竹装饰
 * 内联 SVG 线描竹，鼠尾草灰绿、低透明度、顶部淡化，
 * 挂在认证页左右两侧作留白画框，不上重色、不抢内容。
 */
defineProps<{ side?: 'left' | 'right' }>()
</script>

<template>
  <div class="auth-bamboo" :class="`auth-bamboo--${side ?? 'left'}`" aria-hidden="true">
    <svg
      class="auth-bamboo__svg"
      viewBox="0 0 200 720"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <!-- 竹竿 -->
      <g stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" fill="none">
        <path d="M120 720 C 112 620, 122 540, 112 460 C 104 380, 114 300, 106 220 C 102 170, 106 120, 98 60" stroke-width="4.5" />
        <path d="M84 720 C 78 630, 88 560, 82 480 C 76 410, 82 350, 76 300" stroke-width="3.5" />
        <path d="M108 224 C 132 214, 152 206, 172 194" stroke-width="2.5" />
        <path d="M80 322 C 62 312, 46 300, 30 286" stroke-width="2.5" />
      </g>

      <!-- 竹节 -->
      <g stroke="currentColor" stroke-width="2" stroke-linecap="round" fill="none" opacity="0.9">
        <path d="M108 645 h 22" />
        <path d="M106 545 h 22" />
        <path d="M104 472 h 22" />
        <path d="M102 385 h 22" />
        <path d="M99 302 h 22" />
        <path d="M96 218 h 22" />
        <path d="M94 132 h 20" />
        <path d="M76 622 h 20" />
        <path d="M74 508 h 20" />
        <path d="M72 420 h 20" />
        <path d="M70 340 h 20" />
      </g>

      <!-- 竹叶 -->
      <g fill="currentColor">
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(98 60) rotate(-38)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(96 74) rotate(-12)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(94 88) rotate(12)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(92 102) rotate(46)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(94 84) rotate(-62)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(172 196) rotate(-24)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(176 190) rotate(-52)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(150 204) rotate(6)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(30 288) scale(-1 1) rotate(-24)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(26 282) scale(-1 1) rotate(-52)" />
        <path d="M0 0 C 14 -8, 42 -10, 60 -2 C 44 8, 18 7, 0 0 Z" transform="translate(52 298) scale(-1 1) rotate(10)" />
      </g>
    </svg>
  </div>
</template>

<style scoped>
.auth-bamboo {
  position: absolute;
  bottom: 0;
  height: min(76vh, 660px);
  width: 180px;
  pointer-events: none;
  z-index: 0;
  color: var(--color-sage);
  opacity: 0.3;
}

.auth-bamboo--left {
  left: 1.5%;
}

.auth-bamboo--right {
  right: 1.5%;
}

.auth-bamboo__svg {
  height: 100%;
  width: auto;
  /* 顶部淡出，融入留白 */
  -webkit-mask-image: linear-gradient(to bottom, transparent 0%, #000 16%, #000 100%);
  mask-image: linear-gradient(to bottom, transparent 0%, #000 16%, #000 100%);
}

.auth-bamboo--right .auth-bamboo__svg {
  transform: scaleX(-1);
}

@media (max-width: 900px) {
  .auth-bamboo {
    display: none;
  }
}
</style>
