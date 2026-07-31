<script setup lang="ts">
/**
 * LoginView.vue — 登录页
 * 基本表单: 用户名/邮箱 + 密码 → 提交到 /home
 */
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'

const router = useRouter()
const langStore = useLanguageStore()

const username = ref('')
const password = ref('')

function handleSubmit() {
  // TODO: 对接后端登录API
  // POST /api/auth/login { username, password }
  router.push('/home')
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- Logo -->
      <div class="auth-card__brand">
        <span class="auth-card__logo">蜀韵·成都</span>
        <span class="auth-card__sub">Shu·Chengdu</span>
      </div>

      <h1 class="auth-card__title">{{ langStore.t('login.title') }}</h1>

      <form class="auth-form" @submit.prevent="handleSubmit">
        <div class="auth-form__field">
          <label class="auth-form__label">{{ langStore.t('login.username') }}</label>
          <input
            v-model="username"
            class="auth-form__input"
            :placeholder="langStore.t('login.usernamePlaceholder')"
            type="text"
            autocomplete="username"
            required
          />
        </div>

        <div class="auth-form__field">
          <label class="auth-form__label">{{ langStore.t('login.password') }}</label>
          <input
            v-model="password"
            class="auth-form__input"
            :placeholder="langStore.t('login.passwordPlaceholder')"
            type="password"
            autocomplete="current-password"
            required
          />
        </div>

        <button type="submit" class="auth-form__submit">
          {{ langStore.t('login.submit') }}
        </button>
      </form>

      <div class="auth-card__footer">
        <span class="auth-card__footer-text">{{ langStore.t('login.noAccount') }}</span>
        <router-link to="/register" class="auth-card__footer-link">{{ langStore.t('login.goRegister') }}</router-link>
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

.auth-card {
  width: 100%;
  max-width: 400px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: var(--space-10);
}

.auth-card__brand {
  text-align: center;
  margin-bottom: var(--space-8);
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

.auth-card__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  color: var(--color-text-primary);
  text-align: center;
  margin-bottom: var(--space-8);
  letter-spacing: var(--tracking-wide);
}

/* Form */
.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
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
  transition: border-color var(--transition-fast);
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
}

.auth-form__submit:hover {
  background: var(--color-gold-light);
  box-shadow: 0 0 20px var(--color-gold-glow);
}

.auth-card__footer {
  text-align: center;
  margin-top: var(--space-6);
  font-size: var(--text-sm);
  color: var(--color-text-muted);
}

.auth-card__footer-link {
  color: var(--color-gold);
  margin-left: var(--space-1);
  transition: color var(--transition-fast);
}

.auth-card__footer-link:hover {
  color: var(--color-gold-light);
}
</style>
