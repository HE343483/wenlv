2<template>
  <div class="login-page">
    <!-- 背景粒子 -->
    <canvas ref="bgCanvas" class="bg-canvas"></canvas>

    <div class="login-card" ref="cardRef">
      <!-- 返回按钮 -->
      <button class="back-btn" @click="$router.push('/')">← <span>返回首页</span></button>

      <!-- 成都图标 — 太阳神鸟简化 -->
      <div class="login-icon" ref="iconRef">
        <svg viewBox="0 0 80 80" class="icon-svg">
          <circle cx="40" cy="40" r="36" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity="0.3"/>
          <circle cx="40" cy="40" r="28" fill="none" stroke="#E85D3A" stroke-width="1" opacity="0.5"/>
          <!-- 12道光芒 -->
          <g v-for="i in 12" :key="i">
            <line :x1="40 + 16 * Math.cos(i * Math.PI / 6)" :y1="40 + 16 * Math.sin(i * Math.PI / 6)"
                  :x2="40 + 28 * Math.cos(i * Math.PI / 6)" :y2="40 + 28 * Math.sin(i * Math.PI / 6)"
                  stroke="#E85D3A" stroke-width="2" stroke-linecap="round" opacity="0.6"/>
          </g>
          <circle cx="40" cy="40" r="12" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity="0.4"/>
          <!-- 神鸟简影 -->
          <g v-for="(_, i) in 4" :key="i">
            <path :d="`M${40 + 20 * Math.cos(i * Math.PI / 2 + 0.3)} ${40 + 20 * Math.sin(i * Math.PI / 2 + 0.3)}
              Q${40 + 24 * Math.cos(i * Math.PI / 2 + 0.7)} ${40 + 24 * Math.sin(i * Math.PI / 2 + 0.7)}
              ${40 + 20 * Math.cos(i * Math.PI / 2 + 1.1)} ${40 + 20 * Math.sin(i * Math.PI / 2 + 1.1)}`"
                  fill="none" stroke="#E85D3A" stroke-width="1.8" stroke-linecap="round" opacity="0.7"/>
          </g>
        </svg>
      </div>

      <h2 ref="titleRef">蓉城漫游</h2>
      <p class="login-sub" ref="subRef">CHENGDU CULTURAL TOURISM · 请登录</p>

      <!-- 错误提示 -->
      <div class="error-msg" v-if="errorMsg" ref="errorRef">{{ errorMsg }}</div>

      <form @submit.prevent="handleLogin" ref="formRef">
        <div class="input-group" ref="userGroupRef">
          <label>用户名 / USERNAME</label>
          <input
            v-model="username"
            type="text"
            placeholder="输入用户名"
            class="input-field"
            ref="userInput"
            @focus="clearError"
          />
        </div>
        <div class="input-group" ref="passGroupRef">
          <label>密码 / PASSWORD</label>
          <input
            v-model="password"
            type="password"
            placeholder="输入密码"
            class="input-field"
            @focus="clearError"
          />
        </div>

        <button type="submit" class="login-btn" ref="loginBtnRef" :disabled="isLoading">
          <span v-if="!isLoading">进入蓉城 →</span>
          <span v-else class="loading-text">验证中...</span>
        </button>
      </form>

      <div class="login-hint" ref="hintRef">
        <span class="hint-label">演示账号</span>
        <span class="hint-value">chengdu / 2024</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { login } from '../../stores/auth.js'

const router = useRouter()

const bgCanvas = ref(null)
const cardRef = ref(null)
const iconRef = ref(null)
const titleRef = ref(null)
const subRef = ref(null)
const formRef = ref(null)
const userGroupRef = ref(null)
const passGroupRef = ref(null)
const loginBtnRef = ref(null)
const hintRef = ref(null)
const errorRef = ref(null)
const userInput = ref(null)

const username = ref('')
const password = ref('')
const errorMsg = ref('')
const isLoading = ref(false)
let animCtx = null

function clearError() { errorMsg.value = '' }

function handleLogin() {
  if (!username.value || !password.value) {
    errorMsg.value = '请输入用户名和密码'
    gsap.fromTo(cardRef.value, { x: -8 }, { x: 8, duration: 0.08, repeat: 3, yoyo: true, ease: 'none' })
    return
  }
  isLoading.value = true
  setTimeout(() => {
    if (login(username.value, password.value)) {
      gsap.to(cardRef.value, {
        scale: 0.9, opacity: 0, y: -30, duration: 0.5, ease: 'power2.in',
        onComplete: () => router.push('/main')
      })
    } else {
      isLoading.value = false
      errorMsg.value = '用户名或密码错误 (提示: chengdu / 2024)'
      gsap.fromTo(cardRef.value, { x: -6 }, { x: 6, duration: 0.06, repeat: 4, yoyo: true, ease: 'none' })
    }
  }, 800)
}

onMounted(() => {
  // 粒子
  const canvas = bgCanvas.value
  if (canvas) {
    animCtx = canvas.getContext('2d')
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
    const pts = Array.from({ length: 40 }, () => ({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      r: Math.random() * 1.5 + 0.3,
      dx: (Math.random() - 0.5) * 0.2,
      dy: (Math.random() - 0.5) * 0.2,
      a: Math.random() * 0.3 + 0.05
    }))
    function draw() {
      animCtx.clearRect(0, 0, canvas.width, canvas.height)
      pts.forEach(p => {
        p.x += p.dx; p.y += p.dy
        if (p.x < 0) p.x = canvas.width; if (p.x > canvas.width) p.x = 0
        if (p.y < 0) p.y = canvas.height; if (p.y > canvas.height) p.y = 0
        animCtx.beginPath()
        animCtx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
        animCtx.fillStyle = `rgba(232,93,58,${p.a})`
        animCtx.fill()
      })
      requestAnimationFrame(draw)
    }
    draw()
  }

  // 入场动画
  const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })
  tl.fromTo(cardRef.value, { y: 40, opacity: 0, scale: 0.95 }, { y: 0, opacity: 1, scale: 1, duration: 1 }, 0.2)
  tl.fromTo(iconRef.value, { y: -20, opacity: 0, rotation: -10 }, { y: 0, opacity: 1, rotation: 0, duration: 0.6 }, 0.3)
  tl.fromTo(titleRef.value, { y: 15, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5 }, 0.7)
  tl.fromTo(subRef.value, { y: 10, opacity: 0 }, { y: 0, opacity: 1, duration: 0.4 }, 0.9)
  tl.fromTo(formRef.value, { y: 20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.6 }, 1.1)
  tl.fromTo(hintRef.value, { opacity: 0 }, { opacity: 1, duration: 0.4 }, 1.6)
})

onUnmounted(() => {
  if (animCtx?.canvas) {
    // cleanup
  }
})
</script>

<style scoped>
.login-page {
  position: relative;
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #FFF5ED 0%, #FCF7F2 50%, #F0F7F4 100%);
  overflow: hidden;
}
.bg-canvas {
  position: absolute;
  inset: 0;
  pointer-events: none;
}
.login-card {
  position: relative;
  z-index: 5;
  width: 380px;
  padding: 40px 36px;
  border-radius: 16px;
  background: rgba(255,255,255,.85);
  border: 1px solid rgba(232,93,58,.1);
  backdrop-filter: blur(20px);
  box-shadow: 0 20px 60px rgba(232,93,58,.08);
}
.back-btn {
  background: none;
  border: none;
  color: #9CA3AF;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0;
  margin-bottom: 20px;
  font-family: inherit;
  transition: color 0.3s;
}
.back-btn:hover { color: #E85D3A; }
.login-icon {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}
.icon-svg { width: 60px; height: 60px; }
h2 {
  text-align: center;
  font-size: 1.4rem;
  color: #E85D3A;
  letter-spacing: 4px;
  margin-bottom: 4px;
}
.login-sub {
  text-align: center;
  font-size: 0.6rem;
  color: #9CA3AF;
  letter-spacing: 3px;
  margin-bottom: 24px;
}
.error-msg {
  background: rgba(239,68,68,.06);
  border: 1px solid rgba(239,68,68,.15);
  border-radius: 8px;
  padding: 8px 14px;
  font-size: 0.75rem;
  color: #EF4444;
  margin-bottom: 16px;
  text-align: center;
}
.input-group {
  margin-bottom: 18px;
}
.input-group label {
  display: block;
  font-size: 0.6rem;
  color: #6B7280;
  letter-spacing: 2px;
  margin-bottom: 6px;
}
.input-field {
  width: 100%;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,.06);
  background: rgba(255,255,255,.8);
  color: #2D2D3A;
  font-size: 0.85rem;
  font-family: inherit;
  outline: none;
  transition: border-color 0.3s, box-shadow 0.3s;
}
.input-field:focus {
  border-color: rgba(232,93,58,.35);
  box-shadow: 0 0 20px rgba(232,93,58,.04);
}
.input-field::placeholder { color: #D1D5DB; }
.login-btn {
  width: 100%;
  padding: 13px;
  border-radius: 10px;
  border: none;
  background: linear-gradient(135deg, #E85D3A, #F5A623);
  color: #fff;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
  font-family: inherit;
  letter-spacing: 2px;
  transition: all 0.3s;
  margin-top: 8px;
  box-shadow: 0 4px 15px rgba(232,93,58,.2);
}
.login-btn:hover:not(:disabled) {
  box-shadow: 0 6px 25px rgba(232,93,58,.3);
  transform: translateY(-1px);
}
.login-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.loading-text { font-size: 0.8rem; }
.login-hint {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 20px;
  font-size: 0.65rem;
}
.hint-label {
  color: #9CA3AF;
  letter-spacing: 2px;
}
.hint-value {
  color: rgba(232,93,58,.55);
  letter-spacing: 2px;
  padding: 2px 10px;
  border-radius: 4px;
  background: rgba(232,93,58,.06);
}
</style>
