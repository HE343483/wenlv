<script setup lang="ts">
/**
 * LoginView.vue — 双面卡片登录页
 * 正面 = 登录（左表单 + 右风景A + 「去注册」）
 * 反面 = 注册（左表单 + 右风景B + 「去登录」）
 * 切换：0ms 旧面板交叉滑出 → 400ms 换内容并就位起始侧 → 强制回流 → 新面板滑入 → 800ms 结束
 */
import { nextTick, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'

type Mode = 'login' | 'register'

const router = useRouter()
const langStore = useLanguageStore()

const mode = ref<Mode>('login')
const phase = ref<'idle' | 'exit' | 'enter'>('idle')
const busy = ref(false)

const username = ref('')
const password = ref('')
const regUsername = ref('')
const regEmail = ref('')
const regPassword = ref('')
const regConfirm = ref('')

let timers: number[] = []
function delay(ms: number) {
  return new Promise<void>((resolve) => {
    timers.push(window.setTimeout(resolve, ms))
  })
}

async function switchTo(target: Mode) {
  if (busy.value || target === mode.value) return
  busy.value = true

  phase.value = 'exit' // 0ms：旧面板贴上退场标签，开始滑出
  await delay(400)

  mode.value = target // 400ms：退场完毕，换上新内容
  phase.value = 'enter' // 新面板贴上「起始位置」标签（透明待命）
  await nextTick()
  void document.body.offsetHeight // 强制回流，锁定起始位置
  phase.value = 'idle' // 撕掉起始标签 → 0.4s 平滑滑入

  await delay(400)
  busy.value = false // 800ms：换幕结束
}

function handleLogin() {
  // TODO: 对接后端登录API
  // POST /api/auth/login { username, password }
  router.push('/home')
}

function handleRegister() {
  // TODO: 对接后端注册API
  // POST /api/auth/register { username, email, password }
  router.push('/home')
}

onBeforeUnmount(() => {
  timers.forEach((id) => clearTimeout(id))
  timers = []
})
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-card__brand">
        <span class="auth-card__logo">蜀韵·成都</span>
        <span class="auth-card__sub">Shu·Chengdu</span>
      </div>

      <!-- 登录面 · 左：表单 -->
      <div
        v-show="mode === 'login'"
        class="auth-panel auth-panel--left login-panel"
        :class="{ 'panel-exit-left': phase === 'exit', 'panel-enter-from-left': phase === 'enter' }"
      >
        <h1 class="auth-panel__title">{{ langStore.t('login.title') }}</h1>
        <form class="auth-form" @submit.prevent="handleLogin">
          <div class="auth-form__field">
            <label class="auth-form__label" for="login-username">{{ langStore.t('login.username') }}</label>
            <input
              id="login-username"
              v-model="username"
              class="auth-form__input"
              :placeholder="langStore.t('login.usernamePlaceholder')"
              type="text"
              autocomplete="username"
              required
            />
          </div>
          <div class="auth-form__field">
            <label class="auth-form__label" for="login-password">{{ langStore.t('login.password') }}</label>
            <input
              id="login-password"
              v-model="password"
              class="auth-form__input"
              :placeholder="langStore.t('login.passwordPlaceholder')"
              type="password"
              autocomplete="current-password"
              required
            />
          </div>
          <button type="submit" class="auth-form__submit">{{ langStore.t('login.submit') }}</button>
        </form>
      </div>

      <!-- 登录面 · 右：风景A + 去注册 -->
      <div
        v-show="mode === 'login'"
        class="auth-panel auth-panel--right login-panel"
        :class="{ 'panel-exit-right': phase === 'exit', 'panel-enter-from-right': phase === 'enter' }"
      >
        <div class="auth-scene auth-scene--a">
          <span class="auth-scene__sun" aria-hidden="true"></span>
          <span class="auth-scene__layer auth-scene__layer--far" aria-hidden="true"></span>
          <span class="auth-scene__layer auth-scene__layer--near" aria-hidden="true"></span>
          <div class="auth-scene__overlay">
            <p class="auth-scene__title">{{ langStore.t('login.sceneTitle') }}</p>
            <button type="button" class="auth-scene__switch" @click="switchTo('register')">
              {{ langStore.t('login.goRegister') }} →
            </button>
          </div>
        </div>
      </div>

      <!-- 注册面 · 左：表单 -->
      <div
        v-show="mode === 'register'"
        class="auth-panel auth-panel--left login-panel"
        :class="{ 'panel-exit-left': phase === 'exit', 'panel-enter-from-left': phase === 'enter' }"
      >
        <h1 class="auth-panel__title">{{ langStore.t('register.title') }}</h1>
        <form class="auth-form" @submit.prevent="handleRegister">
          <div class="auth-form__field">
            <label class="auth-form__label" for="reg-username">{{ langStore.t('register.username') }}</label>
            <input
              id="reg-username"
              v-model="regUsername"
              class="auth-form__input"
              :placeholder="langStore.t('register.usernamePlaceholder')"
              type="text"
              autocomplete="username"
              required
            />
          </div>
          <div class="auth-form__field">
            <label class="auth-form__label" for="reg-email">{{ langStore.t('register.email') }}</label>
            <input
              id="reg-email"
              v-model="regEmail"
              class="auth-form__input"
              :placeholder="langStore.t('register.emailPlaceholder')"
              type="email"
              autocomplete="email"
              required
            />
          </div>
          <div class="auth-form__field">
            <label class="auth-form__label" for="reg-password">{{ langStore.t('register.password') }}</label>
            <input
              id="reg-password"
              v-model="regPassword"
              class="auth-form__input"
              :placeholder="langStore.t('register.passwordPlaceholder')"
              type="password"
              autocomplete="new-password"
              required
            />
          </div>
          <div class="auth-form__field">
            <label class="auth-form__label" for="reg-confirm">{{ langStore.t('register.confirm') }}</label>
            <input
              id="reg-confirm"
              v-model="regConfirm"
              class="auth-form__input"
              :placeholder="langStore.t('register.confirmPlaceholder')"
              type="password"
              autocomplete="new-password"
              required
            />
          </div>
          <button type="submit" class="auth-form__submit">{{ langStore.t('register.submit') }}</button>
        </form>
      </div>

      <!-- 注册面 · 右：风景B + 去登录 -->
      <div
        v-show="mode === 'register'"
        class="auth-panel auth-panel--right login-panel"
        :class="{ 'panel-exit-right': phase === 'exit', 'panel-enter-from-right': phase === 'enter' }"
      >
        <div class="auth-scene auth-scene--b">
          <span class="auth-scene__sun" aria-hidden="true"></span>
          <span class="auth-scene__layer auth-scene__layer--far" aria-hidden="true"></span>
          <span class="auth-scene__layer auth-scene__layer--near" aria-hidden="true"></span>
          <div class="auth-scene__overlay">
            <p class="auth-scene__title">{{ langStore.t('register.sceneTitle') }}</p>
            <button type="button" class="auth-scene__switch" @click="switchTo('login')">
              ← {{ langStore.t('register.goLogin') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background: var(--color-bg);
}

/* 2×2 舞台：上行品牌带，下行左右两块内容 */
.auth-card {
  width: 100%;
  max-width: 920px;
  min-height: 620px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: auto 1fr;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: 0 24px 60px rgba(46, 58, 61, 0.08);
}

.auth-card__brand {
  grid-column: 1 / -1;
  grid-row: 1;
  text-align: center;
  padding: var(--space-6) var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
}

.auth-card__logo {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
  display: block;
}

.auth-card__sub {
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wider);
  font-style: italic;
}

/* 面板定位：同一时刻左侧只显示一块，右侧只显示一块 */
.auth-panel {
  grid-row: 2;
  min-width: 0;
}

.auth-panel--left {
  grid-column: 1;
  padding: var(--space-10) var(--space-12);
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.auth-panel--right {
  grid-column: 2;
  padding: 0;
}

/* ① CSS 负责怎么变：贴上 .login-panel 后位移/透明度以 0.4s 平滑过渡 */
.login-panel {
  transition:
    transform 0.4s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: transform, opacity;
}

.panel-exit-left {
  transform: translateX(-64px);
  opacity: 0;
}

.panel-exit-right {
  transform: translateX(64px);
  opacity: 0;
}

.panel-enter-from-left {
  transform: translateX(-64px);
  opacity: 0;
}

.panel-enter-from-right {
  transform: translateX(64px);
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .login-panel {
    transition: none;
  }
}

.auth-panel__title {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--space-6);
  padding-bottom: var(--space-3);
  position: relative;
}

.auth-panel__title::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: 0;
  width: 2.5rem;
  height: 2px;
  background: var(--color-gold);
}

/* Form */
.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.auth-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.auth-form__label {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
}

.auth-form__input {
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-base);
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
  outline: none;
}

.auth-form__input:focus {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.auth-form__input::placeholder {
  color: var(--color-text-muted);
}

.auth-form__submit {
  padding: var(--space-3);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  margin-top: var(--space-2);
  cursor: pointer;
}

.auth-form__submit:hover {
  background: var(--color-gold-light);
  box-shadow: 0 0 20px var(--color-gold-glow);
}

/* 风景卡占位：渐变天空 + 山影剪层，后续可直接替换为真实图片 */
.auth-scene {
  position: relative;
  height: 100%;
  overflow: hidden;
}

.auth-scene--a {
  background: linear-gradient(180deg, #9ecdd6 0%, #5da4b1 52%, #3a7683 100%);
}

.auth-scene--b {
  background: linear-gradient(180deg, #2e3a3d 0%, #5c4448 55%, #b8453e 110%);
}

.auth-scene__sun {
  position: absolute;
  top: 14%;
  right: 18%;
  width: 4.5rem;
  height: 4.5rem;
  border-radius: 50%;
  background: rgba(248, 243, 233, 0.92);
  box-shadow: 0 0 40px 12px rgba(248, 243, 233, 0.45);
}

.auth-scene--b .auth-scene__sun {
  top: 12%;
  right: 22%;
  width: 3.4rem;
  height: 3.4rem;
  box-shadow: 0 0 32px 10px rgba(248, 243, 233, 0.3);
}

.auth-scene__layer {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
}

.auth-scene__layer--far {
  height: 52%;
  clip-path: polygon(0 70%, 12% 44%, 25% 62%, 40% 30%, 55% 60%, 70% 38%, 85% 64%, 100% 46%, 100% 100%, 0 100%);
}

.auth-scene__layer--near {
  height: 38%;
  clip-path: polygon(0 58%, 14% 80%, 30% 50%, 47% 82%, 64% 52%, 80% 76%, 100% 54%, 100% 100%, 0 100%);
}

.auth-scene--a .auth-scene__layer--far {
  background: #2f616c;
  opacity: 0.5;
}

.auth-scene--a .auth-scene__layer--near {
  background: #244e58;
}

.auth-scene--b .auth-scene__layer--far {
  background: #3a2f33;
  opacity: 0.7;
}

.auth-scene--b .auth-scene__layer--near {
  background: #261e22;
}

.auth-scene__overlay {
  position: absolute;
  inset: auto 0 0 0;
  padding: var(--space-10) var(--space-6) var(--space-8);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-5);
  background: linear-gradient(180deg, transparent 0%, rgba(38, 30, 34, 0.55) 100%);
}

.auth-scene__title {
  font-family: var(--font-display);
  color: var(--color-text-inverse);
  font-size: var(--text-xl);
  letter-spacing: var(--tracking-widest);
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.35);
}

.auth-scene__switch {
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-full);
  background: rgba(248, 243, 233, 0.92);
  color: var(--color-text-primary);
  border: 1px solid rgba(248, 243, 233, 0.6);
  font-size: var(--text-sm);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  cursor: pointer;
}

.auth-scene__switch:hover {
  background: var(--color-text-inverse);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.25);
}

/* 窄屏：风景在上、表单在下，纵向排列 */
@media (max-width: 860px) {
  .auth-card {
    max-width: 460px;
    min-height: 0;
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto;
  }

  .auth-card__brand {
    grid-row: 1;
  }

  .auth-panel--right {
    grid-column: 1;
    grid-row: 2;
  }

  .auth-panel--left {
    grid-column: 1;
    grid-row: 3;
    padding: var(--space-8) var(--space-6);
  }

  .auth-scene {
    height: 230px;
  }

  .auth-panel__title {
    font-size: var(--text-xl);
  }
}
</style>
