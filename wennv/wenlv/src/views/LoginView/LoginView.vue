<script setup lang="ts">
/**
 * LoginView.vue — 双面卡片登录页
 * 正面 = 登录（左表单 + 右风景A + 「去注册」）
 * 反面 = 注册（左表单 + 右风景B + 「去登录」）
 * 切换：0ms 旧面板交叉滑出 → 400ms 换内容并就位起始侧 → 强制回流 → 新面板滑入 → 800ms 结束
 */
import { nextTick, onBeforeUnmount, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useUserStore } from '@/stores/user'
import { login as apiLogin, register as apiRegister } from '@/api/auth'
import AuthBamboo from '@/components/AuthBamboo.vue'
import PandaCursor from '@/components/PandaCursor.vue'

type Mode = 'login' | 'register'

const router = useRouter()
const route = useRoute()
const langStore = useLanguageStore()
const userStore = useUserStore()

const mode = ref<Mode>('login')
const phase = ref<'idle' | 'exit' | 'enter'>('idle')
const busy = ref(false)
const submitting = ref(false)
const errorMsg = ref('')

/* 熊猫提示框：登录/注册成功与报错共用 */
const toastVisible = ref(false)
const toastKind = ref<'success' | 'error'>('success')
const toastTitle = ref('')
const toastMsg = ref('')
let toastSeq = 0

function hideToast() {
  toastVisible.value = false
}

async function showToast(kind: 'success' | 'error', title: string, msg: string, autoHideMs = 3600) {
  const seq = ++toastSeq
  toastKind.value = kind
  toastTitle.value = title
  toastMsg.value = msg
  toastVisible.value = true
  await delay(autoHideMs)
  if (seq === toastSeq) toastVisible.value = false
}

function authToastText(kind: 'login-success' | 'login-error' | 'register-success' | 'register-error', detail = '') {
  const lang = langStore.lang
  if (kind === 'login-success') {
    if (lang === 'en') return { title: 'Welcome back!', msg: `Hi ${detail || 'friend'}, Panpan is rolling over to greet you!` }
    if (lang === 'ja') return { title: 'おかえりなさい！', msg: `${detail || 'ともだち'}さん、パンダが転がってお出迎えします！` }
    return { title: '登录成功', msg: `${detail ? `欢迎回来，${detail}！` : '欢迎回来！'}熊猫滚滚来迎接你啦` }
  }
  if (kind === 'register-success') {
    if (lang === 'en') return { title: 'Account created!', msg: 'Panpan saved you a bamboo shoot. Logging you in…' }
    if (lang === 'ja') return { title: '登録成功！', msg: 'パンダが笹を用意しました。ログインします…' }
    return { title: '注册成功', msg: '熊猫给你留了一根竹笋，正在带你进入主页…' }
  }
  if (kind === 'register-error') {
    if (lang === 'en') return { title: 'Registration failed', msg: detail || 'Please try another username.' }
    if (lang === 'ja') return { title: '登録に失敗しました', msg: detail || '別のユーザー名をお試しください。' }
    return { title: '注册失败', msg: detail || '换个用户名再试一次吧' }
  }
  if (lang === 'en') return { title: 'Login failed', msg: detail || 'Please check your username and password.' }
  if (lang === 'ja') return { title: 'ログインに失敗しました', msg: detail || 'ユーザー名とパスワードを確認してください。' }
  return { title: '登录失败', msg: detail || '检查一下账号密码，熊猫陪你再试一次' }
}

function getLoginErrorMessage(error: any): string {
  const status = error?.response?.status
  const hasResponse = Boolean(error?.response)
  const lang = langStore.lang

  if (!hasResponse) {
    if (lang === 'en') return 'The network is unavailable. Please check your connection and try again.'
    if (lang === 'ja') return 'ネットワークに接続できません。接続を確認してもう一度お試しください。'
    return '网络连接失败，请检查网络后重试'
  }

  if (status === 401) {
    if (lang === 'en') return 'Incorrect username or password.'
    if (lang === 'ja') return 'アカウントまたはパスワードが正しくありません。'
    return '账号或密码错误'
  }

  if (status === 404) {
    if (lang === 'en') return 'The login service is unavailable. Please check your network or try again later.'
    if (lang === 'ja') return 'ログインサービスに接続できません。ネットワークを確認して再試行してください。'
    return '登录服务无法访问，请检查网络后重试'
  }

  if (status >= 500) {
    if (lang === 'en') return 'The server is busy. Please contact staff if the problem persists.'
    if (lang === 'ja') return 'サーバーエラーが発生しました。解決しない場合はスタッフにご連絡ください。'
    return '服务器出现异常，如持续无法登录请联系工作人员'
  }

  if (status === 400 || status === 422) {
    if (lang === 'en') return 'The login information is invalid. Please check and try again.'
    if (lang === 'ja') return 'ログイン情報が正しくありません。確認してもう一度お試しください。'
    return '登录信息格式不正确，请检查后重试'
  }

  if (lang === 'en') return 'Login failed. Please try again later.'
  if (lang === 'ja') return 'ログインに失敗しました。しばらくしてから再試行してください。'
  return '登录失败，请稍后重试'
}

const username = ref('')
const password = ref('')
const regUsername = ref('')
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
  if (submitting.value) return
  errorMsg.value = ''
  hideToast()
  submitting.value = true
  apiLogin(username.value.trim(), password.value)
    .then(async (res) => {
      const nickname = res.user?.username || ''
      if (nickname) {
        userStore.updateProfile({
          nickname,
          email: '',
          phone: '',
          avatar: res.user.avatar_url || '',
        })
      }
      const copy = authToastText('login-success', nickname)
      await showToast('success', copy.title, copy.msg, 1600)
      // 登录后跳回守卫拦截前的目标页(无则进 /home)，与其他受保护路由行为一致
      const back = route.query.redirect
      const target = typeof back === 'string' && back.startsWith('/') ? back : '/home'
      router.push(target)
    })
    .catch((err) => {
      const detail = getLoginErrorMessage(err)
      errorMsg.value = detail
      const copy = authToastText('login-error', detail)
      void showToast('error', copy.title, copy.msg, 4200)
    })
    .finally(() => {
      submitting.value = false
    })
}

function handleRegister() {
  if (submitting.value) return
  if (regPassword.value !== regConfirm.value) {
    const copy = langStore.lang === 'en'
      ? { title: 'Passwords do not match', msg: 'Panpan tilted its head: the two passwords are different.' }
      : langStore.lang === 'ja'
        ? { title: 'パスワードが一致しません', msg: 'パンダが首をかしげています：2つのパスワードが違います。' }
        : { title: '两次密码不一致', msg: '熊猫歪头：两遍输入的密码不一样哦' }
    errorMsg.value = copy.msg
    void showToast('error', copy.title, copy.msg, 4200)
    return
  }
  errorMsg.value = ''
  hideToast()
  submitting.value = true
  register2(regUsername.value.trim(), regPassword.value)
    .then(() => {
      const copy = authToastText('register-success')
      return showToast('success', copy.title, copy.msg, 1600).then(() => {
        // 注册成功后自动登录并进入主页
        return apiLogin(regUsername.value.trim(), regPassword.value)
      })
    })
    .then((res) => {
      if (res.user?.username) {
        userStore.updateProfile({
          nickname: res.user.username,
          email: '',
          phone: '',
          avatar: res.user.avatar_url || '',
        })
      }
      // 注册后自动登录：同样跳回守卫拦截前的目标页
      const back = route.query.redirect
      const target = typeof back === 'string' && back.startsWith('/') ? back : '/home'
      router.push(target)
    })
    .catch((err) => {
      const detail = err?.message || ''
      errorMsg.value = detail || authToastText('register-error').msg
      const copy = authToastText('register-error', detail)
      void showToast('error', copy.title, copy.msg, 4200)
    })
    .finally(() => {
      submitting.value = false
    })
}

/** 注册并规范化错误信息 */
async function register2(u: string, p: string) {
  try {
    await apiRegister(u, p)
  } catch (err) {
    const msg = (err as Error)?.message || '注册失败'
    throw new Error(msg.includes('已存在') ? '用户名已存在' : msg)
  }
}

/** 返回公开首页（HomeView） */
function goBackHome() {
  router.push({ name: 'home' })
}

onBeforeUnmount(() => {
  timers.forEach((id) => clearTimeout(id))
  timers = []
})
</script>

<template>
  <div class="auth-page">

    <AuthBamboo side="left" />
    <PandaCursor />
    <div class="auth-card">
      <button type="button" class="auth-back" @click="goBackHome">
      <span class="auth-back__arrow" aria-hidden="true">←</span>
      {{ langStore.t('login.backHome') }}
    </button>
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
          <button type="submit" class="auth-form__submit" :disabled="submitting">
            {{ submitting ? langStore.t('login.loading') || '登录中…' : langStore.t('login.submit') }}
          </button>
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
          <p v-if="errorMsg" class="auth-form__error" role="alert">{{ errorMsg }}</p>
          <button type="submit" class="auth-form__submit" :disabled="submitting">
            {{ submitting ? langStore.t('register.loading') || '提交中…' : langStore.t('register.submit') }}
          </button>
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
    <AuthBamboo side="right" />

    <!-- 熊猫提示框：成功 / 报错 -->
    <Transition name="panda-toast">
      <div v-if="toastVisible" class="panda-toast" :class="`panda-toast--${toastKind}`" role="status" aria-live="polite">
        <button type="button" class="panda-toast__close" aria-label="关闭提示" @click="hideToast">×</button>
        <div class="panda-toast__panda" aria-hidden="true">
          <span class="panda-toast__bamboo"></span>
          <span class="panda-toast__head">
            <span class="panda-toast__ear panda-toast__ear--left"></span>
            <span class="panda-toast__ear panda-toast__ear--right"></span>
            <span class="panda-toast__face">
              <span class="panda-toast__eye panda-toast__eye--left"><i></i></span>
              <span class="panda-toast__eye panda-toast__eye--right"><i></i></span>
              <span class="panda-toast__nose"></span>
              <span class="panda-toast__mouth" :class="{ 'panda-toast__mouth--sad': toastKind === 'error' }"></span>
              <span v-if="toastKind === 'error'" class="panda-toast__tear panda-toast__tear--left"></span>
              <span v-if="toastKind === 'error'" class="panda-toast__tear panda-toast__tear--right"></span>
              <span v-if="toastKind === 'success'" class="panda-toast__blush panda-toast__blush--left"></span>
              <span v-if="toastKind === 'success'" class="panda-toast__blush panda-toast__blush--right"></span>
            </span>
          </span>
          <span class="panda-toast__body">
            <span class="panda-toast__arm panda-toast__arm--left"></span>
            <span class="panda-toast__belly"></span>
            <span class="panda-toast__arm panda-toast__arm--right"></span>
          </span>
          <span class="panda-toast__shadow"></span>
        </div>
        <div class="panda-toast__text">
          <p class="panda-toast__title">{{ toastTitle }}</p>
          <p class="panda-toast__msg">{{ toastMsg }}</p>
        </div>
        <span class="panda-toast__progress" :key="`${toastKind}-${toastTitle}-${toastMsg}`"></span>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.auth-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background: var(--color-bg);
}

/* 返回首页：左上角浮动按钮 */
.auth-back {
  position: absolute;
  top: var(--space-5);
  left: var(--space-6);
  z-index: 2;
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  font-family: var(--font-body);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(46, 58, 61, 0.08);
  transition:
    color var(--transition-fast),
    border-color var(--transition-fast),
    box-shadow var(--transition-fast),
    transform var(--transition-fast);
}

.auth-back:hover {
  color: var(--color-gold);
  border-color: var(--color-gold);
  box-shadow: 0 0 16px var(--color-gold-glow);
}

.auth-back__arrow {
  font-size: var(--text-base);
  line-height: 1;
  transition: transform var(--transition-fast);
}

.auth-back:hover .auth-back__arrow {
  transform: translateX(-3px);
}

/* 2×2 舞台：上行品牌带，下行左右两块内容 */
.auth-card {
  position: relative;
  z-index: 1;
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

.auth-form__submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.auth-form__error {
  font-size: var(--text-sm);
  color: var(--color-danger, #b8453e);
  letter-spacing: var(--tracking-wide);
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

/* 熊猫提示框 */
.panda-toast {
  --toast-accent: var(--color-gold);
  --toast-accent-soft: var(--color-gold-glow);
  position: fixed;
  top: calc(var(--nav-height) + var(--space-1));
  right: var(--space-6);
  bottom: auto;
  z-index: 60;
  display: flex;
  align-items: center;
  gap: var(--space-4);
  width: min(360px, calc(100vw - 48px));
  padding: var(--space-4) var(--space-5);
  border-radius: var(--radius-xl);
  background: color-mix(in srgb, var(--color-surface) 94%, transparent);
  border: 1px solid var(--color-border);
  box-shadow:
    0 18px 44px rgba(46, 58, 61, 0.16),
    0 0 0 4px var(--toast-accent-soft);
  backdrop-filter: blur(14px);
  overflow: hidden;
}

.panda-toast--success {
  --toast-accent: var(--color-gold);
  --toast-accent-soft: var(--color-gold-glow);
}

.panda-toast--error {
  --toast-accent: var(--color-danger, #b8453e);
  --toast-accent-soft: rgba(184, 69, 62, 0.12);
}

.panda-toast::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  background: linear-gradient(180deg, var(--toast-accent), transparent);
}

.panda-toast__close {
  position: absolute;
  top: 6px;
  right: 10px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  color: var(--color-text-muted);
  font-size: 16px;
  line-height: 1;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.panda-toast__close:hover {
  color: var(--toast-accent);
  background: var(--toast-accent-soft);
}

.panda-toast__text {
  min-width: 0;
}

.panda-toast__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  margin-bottom: 2px;
}

.panda-toast__msg {
  font-size: var(--text-sm);
  line-height: 1.6;
  color: var(--color-text-secondary);
  word-break: break-word;
}

.panda-toast__progress {
  position: absolute;
  left: 0;
  bottom: 0;
  height: 3px;
  width: 100%;
  transform-origin: left center;
  background: var(--toast-accent);
  animation: panda-toast-progress 3.6s linear forwards;
}

.panda-toast--success .panda-toast__progress {
  animation-duration: 1.6s;
}

@keyframes panda-toast-progress {
  from { transform: scaleX(1); }
  to { transform: scaleX(0); }
}

/* 熊猫本体：CSS 拼出的坐姿小熊猫 */
.panda-toast__panda {
  position: relative;
  flex: none;
  width: 76px;
  height: 84px;
  animation: panda-toast-bob 2.6s ease-in-out infinite;
}

@keyframes panda-toast-bob {
  0%, 100% { transform: translateY(0) rotate(-1deg); }
  50% { transform: translateY(-4px) rotate(1deg); }
}

.panda-toast__head {
  position: absolute;
  left: 8px;
  top: 0;
  width: 60px;
  height: 52px;
  z-index: 2;
}

.panda-toast__ear {
  position: absolute;
  top: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #2e3a3d;
}

.panda-toast__ear--left { left: 0; transform: rotate(-12deg); }
.panda-toast__ear--right { right: 0; transform: rotate(12deg); }

.panda-toast__face {
  position: absolute;
  left: 4px;
  right: 4px;
  top: 6px;
  bottom: 0;
  border-radius: 46% 46% 48% 48%;
  background: #fffdf8;
  border: 1px solid rgba(46, 58, 61, 0.12);
  box-shadow: inset 0 -6px 12px rgba(46, 58, 61, 0.06);
}

.panda-toast__eye {
  position: absolute;
  top: 12px;
  width: 16px;
  height: 20px;
  border-radius: 50%;
  background: #2e3a3d;
  transform: rotate(-14deg);
  animation: panda-toast-blink 4.2s infinite;
}

.panda-toast__eye--left { left: 8px; }
.panda-toast__eye--right { right: 8px; transform: rotate(14deg); }

.panda-toast__eye i {
  position: absolute;
  left: 4px;
  top: 4px;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #fff;
}

@keyframes panda-toast-blink {
  0%, 93%, 100% { transform: rotate(-14deg) scaleY(1); }
  95% { transform: rotate(-14deg) scaleY(0.12); }
}

.panda-toast__nose {
  position: absolute;
  left: 50%;
  top: 30px;
  width: 8px;
  height: 6px;
  border-radius: 50%;
  background: #2e3a3d;
  transform: translateX(-50%);
}

.panda-toast__mouth {
  position: absolute;
  left: 50%;
  top: 36px;
  width: 14px;
  height: 7px;
  border: 2px solid #2e3a3d;
  border-top: none;
  border-left-color: transparent;
  border-right-color: transparent;
  border-radius: 0 0 14px 14px;
  transform: translateX(-50%);
}

.panda-toast__mouth--sad {
  top: 38px;
  border-radius: 14px 14px 0 0;
  border: 2px solid #2e3a3d;
  border-bottom: none;
  border-left-color: transparent;
  border-right-color: transparent;
}

.panda-toast__blush {
  position: absolute;
  top: 30px;
  width: 10px;
  height: 6px;
  border-radius: 50%;
  background: rgba(184, 69, 62, 0.28);
}

.panda-toast__blush--left { left: 4px; }
.panda-toast__blush--right { right: 4px; }

.panda-toast__tear {
  position: absolute;
  top: 30px;
  width: 5px;
  height: 8px;
  border-radius: 50%;
  background: rgba(93, 164, 177, 0.85);
  animation: panda-toast-tear 1.4s ease-in infinite;
}

.panda-toast__tear--left { left: 10px; }
.panda-toast__tear--right { right: 10px; animation-delay: 0.35s; }

@keyframes panda-toast-tear {
  0% { transform: translateY(0); opacity: 0; }
  25% { opacity: 1; }
  100% { transform: translateY(8px); opacity: 0; }
}

.panda-toast__body {
  position: absolute;
  left: 14px;
  bottom: 10px;
  width: 48px;
  height: 34px;
  z-index: 1;
}

.panda-toast__belly {
  position: absolute;
  inset: 0;
  border-radius: 48% 48% 46% 46%;
  background: #fffdf8;
  border: 1px solid rgba(46, 58, 61, 0.12);
}

.panda-toast__arm {
  position: absolute;
  top: 2px;
  width: 14px;
  height: 24px;
  border-radius: 999px;
  background: #2e3a3d;
  animation: panda-toast-wave 2.6s ease-in-out infinite;
  transform-origin: top center;
}

.panda-toast__arm--left { left: -6px; }
.panda-toast__arm--right { right: -6px; animation-delay: 0.5s; }

@keyframes panda-toast-wave {
  0%, 100% { transform: rotate(8deg); }
  50% { transform: rotate(-16deg); }
}

.panda-toast__bamboo {
  position: absolute;
  right: 2px;
  bottom: 12px;
  width: 8px;
  height: 44px;
  border-radius: 999px;
  background: linear-gradient(180deg, #6a7a6a, #4c5b4c);
  transform: rotate(14deg);
  z-index: 0;
}

.panda-toast__bamboo::before,
.panda-toast__bamboo::after {
  content: '';
  position: absolute;
  left: -3px;
  width: 14px;
  height: 8px;
  border-radius: 999px;
  background: #7d927d;
}

.panda-toast__bamboo::before { top: 8px; transform: rotate(-18deg); }
.panda-toast__bamboo::after { top: 22px; transform: rotate(18deg); }

.panda-toast__shadow {
  position: absolute;
  left: 10px;
  right: 10px;
  bottom: 2px;
  height: 8px;
  border-radius: 50%;
  background: rgba(46, 58, 61, 0.14);
  filter: blur(2px);
}

.panda-toast-enter-active,
.panda-toast-leave-active {
  transition: transform 0.38s cubic-bezier(0.34, 1.4, 0.64, 1), opacity 0.3s ease;
}

.panda-toast-enter-from {
  transform: translateY(-28px) scale(0.94);
  opacity: 0;
}

.panda-toast-leave-to {
  transform: translateY(-12px) scale(0.96);
  opacity: 0;
}

/* 窄屏：风景在上、表单在下，纵向排列 */
@media (max-width: 860px) {
  .panda-toast {
    top: var(--space-1);
    right: var(--space-4);
    left: var(--space-4);
    bottom: auto;
    width: auto;
  }

  .auth-back {
    top: var(--space-3);
    left: var(--space-3);
    padding: var(--space-2) var(--space-3);
    font-size: var(--text-xs);
  }

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
