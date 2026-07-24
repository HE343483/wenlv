<template>
  <div class="landing" ref="landingRef">
    <!-- 背景粒子 -->
    <canvas ref="particleCanvas" class="particle-bg"></canvas>

    <!-- 浮动文化符号 -->
    <div class="float-symbols" ref="symbolsRef">
      <div class="symbol panda" ref="pandaRef">🐼</div>
      <div class="symbol teahouse" ref="teahouseRef">🍵</div>
      <div class="symbol bamboo" ref="bambooRef">🎋</div>
      <div class="symbol lantern" ref="lanternRef">🏮</div>
      <div class="symbol opera" ref="operaRef">🎭</div>
      <div class="symbol bridge" ref="bridgeRef">🌉</div>
    </div>

    <!-- 英雄区域 -->
    <div class="hero" ref="heroRef">
      <!-- 成都标志 — 太阳神鸟 -->
      <svg class="bird-svg" viewBox="0 0 200 200" ref="birdSvgRef">
        <defs>
          <linearGradient id="warmGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#E85D3A" />
            <stop offset="100%" stop-color="#F5A623" />
          </linearGradient>
          <linearGradient id="tealGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#0EA5A0" />
            <stop offset="100%" stop-color="#34D399" />
          </linearGradient>
        </defs>
        <!-- 外圈太阳 -->
        <circle cx="100" cy="100" r="85" fill="none" stroke="url(#warmGrad)" stroke-width="2" opacity="0.3"/>
        <circle cx="100" cy="100" r="75" fill="none" stroke="url(#warmGrad)" stroke-width="1.5" opacity="0.5"/>
        <!-- 太阳光芒 -->
        <g v-for="i in 12" :key="i">
          <line :x1="100 + 48 * Math.cos(i * Math.PI / 6)" :y1="100 + 48 * Math.sin(i * Math.PI / 6)"
                :x2="100 + 70 * Math.cos(i * Math.PI / 6)" :y2="100 + 70 * Math.sin(i * Math.PI / 6)"
                stroke="url(#warmGrad)" stroke-width="2.5" stroke-linecap="round" opacity="0.7"/>
        </g>
        <!-- 中间太阳 -->
        <circle cx="100" cy="100" r="30" fill="none" stroke="url(#warmGrad)" stroke-width="2" opacity="0.6"/>
        <!-- 神鸟环绕 -->
        <g v-for="(_, i) in 4" :key="i">
          <path :d="`M${100 + 52 * Math.cos(i * Math.PI / 2 + 0.4)} ${100 + 52 * Math.sin(i * Math.PI / 2 + 0.4)}
            Q${100 + 62 * Math.cos(i * Math.PI / 2 + 0.8)} ${100 + 62 * Math.sin(i * Math.PI / 2 + 0.8)}
            ${100 + 52 * Math.cos(i * Math.PI / 2 + 1.2)} ${100 + 52 * Math.sin(i * Math.PI / 2 + 1.2)}`"
            fill="none" stroke="url(#warmGrad)" stroke-width="2" stroke-linecap="round" opacity="0.8"/>
        </g>
      </svg>

      <div class="hero-text" ref="heroTextRef">
        <div class="subtitle-top" ref="subtitleTopRef">CHENGDU · 天府</div>
        <h1 ref="titleRef">
          <span class="title-line" ref="titleLine1">烟火成都</span>
          <span class="title-line accent" ref="titleLine2">诗意蓉城</span>
        </h1>
        <p class="description" ref="descRef">
          在宽窄巷子的茶香里品读千年<br>
          在熊猫的怀抱中感受成都的温暖与活力
        </p>
        <div class="cta-row" ref="ctaRef">
          <button class="cta-btn" @click="goToLogin">
            <span>开始探索</span>
            <span class="btn-arrow">→</span>
            <div class="btn-glow"></div>
          </button>
          <div class="cta-sub" ref="ctaSubRef">
            <span class="dot-pulse"></span>
            即刻开启蓉城之旅
          </div>
        </div>
      </div>
    </div>

    <!-- 底部滚动提示 -->
    <div class="scroll-hint" ref="scrollHintRef">
      <span>SCROLL</span>
      <div class="scroll-line"><div class="scroll-dot"></div></div>
    </div>

    <!-- 侧边标签 -->
    <div class="side-tags" ref="sideTagsRef">
      <span>宽窄巷子</span><span>·</span><span>锦里</span><span>·</span><span>熊猫基地</span><span>·</span><span>都江堰</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

const router = useRouter()

// Refs
const landingRef = ref(null)
const particleCanvas = ref(null)
const symbolsRef = ref(null)
const pandaRef = ref(null)
const teahouseRef = ref(null)
const bambooRef = ref(null)
const lanternRef = ref(null)
const operaRef = ref(null)
const bridgeRef = ref(null)
const heroRef = ref(null)
const birdSvgRef = ref(null)
const heroTextRef = ref(null)
const subtitleTopRef = ref(null)
const titleRef = ref(null)
const titleLine1 = ref(null)
const titleLine2 = ref(null)
const descRef = ref(null)
const ctaRef = ref(null)
const ctaSubRef = ref(null)
const scrollHintRef = ref(null)
const sideTagsRef = ref(null)

let particleAnimId = null
let ctx = null

function goToLogin() {
  router.push('/login')
}

onMounted(() => {
  // ===== 1. 粒子背景 =====
  const canvas = particleCanvas.value
  if (canvas) {
    ctx = canvas.getContext('2d')
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight

    const particles = Array.from({ length: 60 }, () => ({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      r: Math.random() * 2 + 0.5,
      dx: (Math.random() - 0.5) * 0.2,
      dy: (Math.random() - 0.5) * 0.2,
      alpha: Math.random() * 0.3 + 0.05
    }))

    function drawParticles() {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      particles.forEach(p => {
        p.x += p.dx
        p.y += p.dy
        if (p.x < 0) p.x = canvas.width
        if (p.x > canvas.width) p.x = 0
        if (p.y < 0) p.y = canvas.height
        if (p.y > canvas.height) p.y = 0
        ctx.beginPath()
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
        ctx.fillStyle = `rgba(232, 93, 58, ${p.alpha})`
        ctx.fill()
      })
      particleAnimId = requestAnimationFrame(drawParticles)
    }
    drawParticles()
  }

  // ===== 2. 主时间线: GSAP 动画序列 =====
  const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })

  // 2a. SVG太阳神鸟旋转淡入
  tl.fromTo(birdSvgRef.value,
    { scale: 0.6, opacity: 0, rotation: -30 },
    { scale: 1, opacity: 1, rotation: 0, duration: 1.8, ease: 'back.out(1.4)' },
    0.2
  )

  // 2b. 副标题
  tl.fromTo(subtitleTopRef.value,
    { y: -20, opacity: 0 },
    { y: 0, opacity: 1, duration: 0.8 },
    0.6
  )

  // 2c. 标题行
  tl.fromTo(titleLine1.value,
    { x: -60, opacity: 0 },
    { x: 0, opacity: 1, duration: 1 },
    1.0
  )
  tl.fromTo(titleLine2.value,
    { x: 60, opacity: 0 },
    { x: 0, opacity: 1, duration: 1 },
    1.2
  )

  // 2d. 描述文字
  tl.fromTo(descRef.value,
    { y: 30, opacity: 0 },
    { y: 0, opacity: 1, duration: 0.8 },
    1.6
  )

  // 2e. CTA按钮
  tl.fromTo(ctaRef.value,
    { y: 25, opacity: 0, scale: 0.95 },
    { y: 0, opacity: 1, scale: 1, duration: 0.7, ease: 'back.out(1.7)' },
    2.0
  )

  // 2f. 副按钮文字
  tl.fromTo(ctaSubRef.value,
    { opacity: 0 },
    { opacity: 1, duration: 0.5 },
    2.2
  )

  // 2g. 滚动提示
  tl.fromTo(scrollHintRef.value,
    { opacity: 0, y: 15 },
    { opacity: 1, y: 0, duration: 0.6 },
    2.5
  )

  // 2h. 侧标签
  tl.fromTo(sideTagsRef.value,
    { opacity: 0, x: 20 },
    { opacity: 1, x: 0, duration: 0.6 },
    2.3
  )

  // ===== 3. 浮动符号动画 =====
  const floatSym = [pandaRef, teahouseRef, bambooRef, lanternRef, operaRef, bridgeRef]
  floatSym.forEach((sym, i) => {
    if (!sym.value) return
    gsap.to(sym.value, {
      y: -15 - Math.random() * 25,
      x: (Math.random() - 0.5) * 15,
      rotation: (Math.random() - 0.5) * 12,
      duration: 2 + Math.random() * 2,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut',
      delay: i * 0.3
    })
    gsap.fromTo(sym.value,
      { opacity: 0, scale: 0 },
      { opacity: 0.5, scale: 1, duration: 1.5, delay: 1 + i * 0.2, ease: 'back.out(2)' }
    )
  })

  // ===== 4. 鼠标跟随视差 =====
  function handleMouse(e) {
    const x = (e.clientX / window.innerWidth - 0.5) * 2
    const y = (e.clientY / window.innerHeight - 0.5) * 2
    gsap.to(symbolsRef.value, {
      x: x * 12,
      y: y * 8,
      duration: 1.5,
      ease: 'power2.out'
    })
  }
  window.addEventListener('mousemove', handleMouse)

  // ===== 5. CTA按钮呼吸光效 =====
  const ctaBtn = document.querySelector('.cta-btn')
  if (ctaBtn) {
    gsap.to('.btn-glow', {
      scale: 1.2,
      opacity: 0.25,
      duration: 1.5,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut'
    })
  }
})

onUnmounted(() => {
  if (particleAnimId) cancelAnimationFrame(particleAnimId)
  ScrollTrigger.getAll().forEach(t => t.kill())
})
</script>

<style scoped>
.landing {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: radial-gradient(ellipse at 50% 30%, #FFF5ED, #FCF7F2 70%);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ===== 粒子背景 ===== */
.particle-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

/* ===== 浮动符号 ===== */
.float-symbols {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}
.symbol {
  position: absolute;
  font-size: 2rem;
  opacity: 0;
  filter: blur(0.5px);
}
.panda    { top: 8%; left: 6%; }
.teahouse { top: 12%; right: 10%; }
.bamboo   { bottom: 20%; left: 4%; }
.lantern  { top: 35%; right: 4%; }
.opera    { bottom: 12%; right: 6%; }
.bridge   { bottom: 28%; left: 10%; }

/* ===== 英雄区域 ===== */
.hero {
  position: relative;
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 50px;
  padding: 0 60px;
  max-width: 1100px;
  width: 100%;
}

/* SVG 太阳神鸟 */
.bird-svg {
  width: 180px;
  height: 180px;
  flex-shrink: 0;
  filter: drop-shadow(0 0 30px rgba(232,93,58,.12));
}

/* 英雄文字 */
.hero-text {
  flex: 1;
}
.subtitle-top {
  font-size: 0.85rem;
  letter-spacing: 8px;
  color: rgba(232,93,58,.55);
  margin-bottom: 10px;
  font-weight: 400;
}
h1 {
  margin-bottom: 18px;
}
.title-line {
  display: block;
  font-size: 3.4rem;
  font-weight: 900;
  letter-spacing: 4px;
  line-height: 1.15;
  color: #2D2D3A;
}
.title-line.accent {
  background: linear-gradient(90deg, #E85D3A, #F5A623, #E85D3A);
  background-size: 200% auto;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.description {
  font-size: 0.95rem;
  line-height: 1.9;
  color: #6B7280;
  margin-bottom: 28px;
  max-width: 500px;
}

/* CTA按钮 */
.cta-row {
  display: flex;
  align-items: center;
  gap: 20px;
}
.cta-btn {
  position: relative;
  padding: 14px 40px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #E85D3A, #F5A623);
  color: #fff;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  font-family: inherit;
  letter-spacing: 2px;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all 0.4s cubic-bezier(0.16,1,0.3,1);
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(232,93,58,.25);
}
.cta-btn:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 8px 35px rgba(232,93,58,.35);
}
.btn-arrow {
  display: inline-block;
  transition: transform 0.3s;
}
.cta-btn:hover .btn-arrow {
  transform: translateX(4px);
}
.btn-glow {
  position: absolute;
  inset: 0;
  border-radius: 12px;
  background: radial-gradient(ellipse at 50% 100%, rgba(255,255,255,.2), transparent 70%);
  pointer-events: none;
}
.cta-sub {
  font-size: 0.7rem;
  color: #6B7280;
  letter-spacing: 2px;
}
.dot-pulse {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #E85D3A;
  margin-right: 6px;
  animation: dotPulse 2s ease-in-out infinite;
  vertical-align: middle;
}
@keyframes dotPulse {
  0%, 100% { opacity: .3; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.3); }
}

/* 滚动提示 */
.scroll-hint {
  position: absolute;
  bottom: 40px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  z-index: 5;
}
.scroll-hint span {
  font-size: 0.55rem;
  letter-spacing: 4px;
  color: #9CA3AF;
}
.scroll-line {
  width: 1px;
  height: 30px;
  background: rgba(0,0,0,.06);
  position: relative;
}
.scroll-dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #E85D3A;
  position: absolute;
  top: 0;
  left: -1px;
  animation: scrollDot 2s ease-in-out infinite;
}
@keyframes scrollDot {
  0% { top: 0; opacity: 1; }
  100% { top: 100%; opacity: 0; }
}

/* 侧边标签 */
.side-tags {
  position: absolute;
  bottom: 40px;
  right: 40px;
  display: flex;
  gap: 8px;
  font-size: 0.65rem;
  color: #9CA3AF;
  letter-spacing: 2px;
  z-index: 5;
}
.side-tags span:nth-child(odd) {
  color: rgba(232,93,58,.4);
}

/* 响应式 */
@media (max-width: 1024px) {
  .hero { flex-direction: column; text-align: center; gap: 24px; }
  .bird-svg { width: 130px; height: 130px; }
  .title-line { font-size: 2.2rem; }
  .description { max-width: 100%; }
  .cta-row { justify-content: center; flex-direction: column; }
  .side-tags { display: none; }
}
</style>
