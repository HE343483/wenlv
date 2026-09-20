<template>
  <nav class="navbar navbar-toggleable-md fixed-top navbar-transparent landing-navbar">
    <div class="container">
      <div class="navbar-translate">
        <button
          class="navbar-toggler navbar-toggler-right navbar-burger landing-burger"
          type="button"
          aria-label="Toggle navigation"
        >
          <span class="navbar-toggler-bar"></span>
          <span class="navbar-toggler-bar"></span>
          <span class="navbar-toggler-bar"></span>
        </button>
        <button class="navbar-brand landing-brand" type="button" @click="handleBrandClick">
          {{ t('app.brand') }}
        </button>
        <button class="landing-back-home" type="button" :title="t('home.nav.backHome')" @click="goHome">
          <svg width="14px" height="14px" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" aria-hidden="true"><path fill="currentColor" d="M641.024 89.344l-402.688 380.032a63.616 63.616 0 0 0-20.224 46.848c0.128 17.92 7.168 34.944 20.224 46.784l402.688 371.648a21.312 21.312 0 0 0 36.224-15.296V104.64a21.312 21.312 0 0 0-36.224-15.296z"/></svg>
          <span>{{ t('home.nav.backHome') }}</span>
        </button>
      </div>
      <div class="navbar-collapse landing-navbar-collapse" id="navbarToggler">
        <ul class="navbar-nav ml-auto landing-nav">
          <li class="nav-item">
            <!-- <button class="nav-link landing-nav-btn fog-toggle" type="button" :aria-pressed="fogEnabled" @click="toggleFog">
              {{ fogEnabled ? t('home.nav.fogOn') : t('home.nav.fogOff') }}
            </button> -->
          </li>
          <li class="nav-item landing-lang-item">
            <a-select v-model:value="locale" class="lang-select-nav" size="small" :aria-label="t('app.language.label')">
              <a-select-option value="zh-CN">{{ t('app.language.zh') }}</a-select-option>
              <a-select-option value="ja-JP">{{ t('app.language.ja') }}</a-select-option>
              <a-select-option value="en-US">{{ t('app.language.en') }}</a-select-option>
            </a-select>
          </li>
          <li class="nav-item">
            <button
              type="button"
              class="nav-link landing-nav-btn settings-btn"
              :title="t('settings.open')"
              :aria-label="t('settings.open')"
              @click="settingsVisible = true"
            >
              <svg width="25px" height="25px" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M600.704 64a32 32 0 0 1 30.464 22.208l35.2 109.376c14.784 7.232 28.928 15.36 42.432 24.512l112.384-24.192a32 32 0 0 1 34.432 15.36L944.32 364.8a32 32 0 0 1-4.032 37.504l-77.12 85.12a357.12 357.12 0 0 1 0 49.024l77.12 85.248a32 32 0 0 1 4.032 37.504l-88.704 153.6a32 32 0 0 1-34.432 15.296L708.8 803.904c-13.44 9.088-27.648 17.28-42.368 24.512l-35.264 109.376A32 32 0 0 1 600.704 960H423.296a32 32 0 0 1-30.464-22.208L357.696 828.48a351.616 351.616 0 0 1-42.56-24.64l-112.32 24.256a32 32 0 0 1-34.432-15.36L79.68 659.2a32 32 0 0 1 4.032-37.504l77.12-85.248a357.12 357.12 0 0 1 0-48.896l-77.12-85.248A32 32 0 0 1 79.68 364.8l88.704-153.6a32 32 0 0 1 34.432-15.296l112.32 24.256c13.568-9.152 27.776-17.408 42.56-24.64l35.2-109.312A32 32 0 0 1 423.232 64H600.64zm-23.424 64H446.72l-36.352 113.088-24.512 11.968a294.113 294.113 0 0 0-34.816 20.096l-22.656 15.36-116.224-25.088-65.28 113.152 79.68 88.192-1.92 27.136a293.12 293.12 0 0 0 0 40.192l1.92 27.136-79.808 88.192 65.344 113.152 116.224-25.024 22.656 15.296a294.113 294.113 0 0 0 34.816 20.096l24.512 11.968L446.72 896h130.688l36.48-113.152 24.448-11.904a288.282 288.282 0 0 0 34.752-20.096l22.592-15.296 116.288 25.024 65.28-113.152-79.744-88.192 1.92-27.136a293.12 293.12 0 0 0 0-40.256l-1.92-27.136 79.808-88.128-65.344-113.152-116.288 24.96-22.592-15.232a287.616 287.616 0 0 0-34.752-20.096l-24.448-11.904L577.344 128zM512 320a192 192 0 1 1 0 384 192 192 0 0 1 0-384zm0 64a128 128 0 1 0 0 256 128 128 0 0 0 0-256z"/></svg>
            </button>
          </li>
          <li class="nav-item">
            <button type="button" class="btn btn-danger btn-round landing-cta" @click="handleCtaClick">
              {{ t('home.nav.cta') }}
            </button>
          </li>
        </ul>
      </div>
    </div>
    <TripSettingsModal v-model:open="settingsVisible" />
  </nav>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import type { RuntimeSettings } from '@/types/trip'
import { getRuntimeSettings, saveRuntimeSettings } from '@/api/trip'

const { t, locale } = useI18n()
const router = useRouter()

/** 返回主站首页 */
const goHome = () => {
  router.push('/')
}
const settingsVisible = ref(false)

const emit = defineEmits<{
  (e: 'brand-click'): void
  (e: 'cta-click'): void
}>()

const handleBrandClick = () => {
  emit('brand-click')
}

const handleCtaClick = () => {
  emit('cta-click')
}
</script>

<style scoped>
.landing-navbar {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  width: 100% !important;
  z-index: 1030 !important;
  min-height: 70px;
  padding: 0 !important;
  background: transparent !important;
  background-color: transparent !important;
  background-image: none !important;
  box-shadow: none !important;
  border: none !important;
  border-bottom: 1px solid transparent !important;
  backdrop-filter: none !important;
  -webkit-backdrop-filter: none !important;
  transition: background 0.3s;
}

.landing-navbar:not(.navbar-transparent) {
  background: transparent !important;
  background-color: transparent !important;
  background-image: none !important;
  border-color: transparent !important;
  backdrop-filter: none !important;
  -webkit-backdrop-filter: none !important;
}

.landing-navbar.navbar-transparent {
  padding-top: 0 !important;
  background: transparent !important;
  background-color: transparent !important;
  background-image: none !important;
  box-shadow: none !important;
}

.landing-navbar *,
.landing-navbar::before,
.landing-navbar::after {
  box-sizing: border-box;
}

.landing-navbar .container {
  width: 100% !important;
  max-width: 100vw !important;
  min-height: 70px;
  display: flex !important;
  align-items: center !important;
  justify-content: space-between !important;
  padding-left: 20px;
  padding-right: 20px;
  box-sizing: border-box !important;
}

.landing-navbar .navbar-translate {
  display: flex !important;
  align-items: center !important;
  min-height: 70px;
  flex: 0 0 auto;
}

.landing-navbar .navbar-brand {
  margin: 0 !important;
  padding: 0 !important;
  line-height: 1 !important;
}

/* 返回首页按钮 */
.landing-back-home {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: 18px;
  padding: 6px 16px;
  border: 1px solid rgba(255, 255, 255, 0.35);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(6px);
  color: rgba(255, 255, 255, 0.92);
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.landing-back-home:hover {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.6);
  transform: translateY(-1px);
}

.landing-back-home svg {
  flex: none;
}

.landing-burger {
  display: none !important;
}

.landing-brand {
  background: transparent !important;
  border: 0;
  color: #f4f8fc !important;
  font-weight: 700 !important;
  letter-spacing: 0.12em !important;
  text-transform: uppercase;
  font-size: 13px !important;
  cursor: pointer;
  min-height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.landing-navbar-collapse {
  display: flex !important;
  justify-content: flex-end;
  align-items: center;
  flex: 1;
  /* 允许收缩但不撑破容器:配合导航项 flex:0 0 auto,保证窄屏下导航项保持自然宽度且不被推出可视区 */
  min-width: 0;
  position: static !important;
  transform: none !important;
  width: auto !important;
  height: auto !important;
  background: transparent !important;
  border: 0 !important;
  padding: 0 !important;
  overflow: visible !important;
}

.landing-navbar-collapse::before,
.landing-navbar-collapse::after {
  display: none !important;
  content: none !important;
  background: transparent !important;
}

.landing-nav {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 0 auto !important;
  padding: 0;
  list-style: none;
}

.landing-nav .nav-item {
  margin: 0;
  padding: 0;
  display: inline-flex;
  align-items: center;
  /* 多语言文案长短不一,导航项一律不参与压缩,避免被挤成窄条导致文字换行(会把导航栏撑高并遮挡页面顶部操作栏) */
  flex: 0 0 auto;
}

.landing-nav .nav-item .nav-link {
  margin: 0 !important;
  padding: 0 !important;
  line-height: 1 !important;
  opacity: 1 !important;
}

.landing-nav-btn {
  border: 1.2px solid #e8efed;
  background: rgba(255, 253, 248, 0.92);
  color: #2e3a3d;
  border-radius: 999px;
  padding: 0 12px;
  min-height: 34px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  text-transform: uppercase;
  /* 文案永不换行:导航栏高度不随语言变化 */
  white-space: nowrap;
}

.landing-nav-btn:hover {
  border-color: #5da4b1;
  color: #3e7d8a;
}

.settings-btn {
  text-transform: none;
  border: none !important;
  background: none !important;
}

.fog-toggle[aria-pressed='true'] {
  border-color: rgba(184, 69, 62, 0.55);
  background: rgba(184, 69, 62, 0.2);
  color: #ffffff;
}

.landing-lang-item {
  display: flex;
  align-items: center;
}

.lang-select-nav {
  width: 110px;
}

.lang-select-nav :deep(.ant-select-selector) {
  height: 34px !important;
  padding: 0 12px !important;
  border: 1.2px solid #e8efed !important;
  background: rgba(255, 253, 248, 0.92) !important;
  border-radius: 999px !important;
  display: flex !important;
  align-items: center !important;
}

.lang-select-nav :deep(.ant-select-selection-item) {
  line-height: 32px !important;
  font-size: 12px !important;
}

.lang-select-nav :deep(.ant-select-selection-item),
.lang-select-nav :deep(.ant-select-arrow) {
  color: #2e3a3d !important;
}

.landing-cta {
  min-height: 32px;
  padding: 0 14px !important;
  font-size: 12px !important;
  border: none !important;
  letter-spacing: 0.06em;
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  margin: 0 !important;
  /* 文案永不换行,避免按钮被压缩后文字竖排撑高导航栏 */
  white-space: nowrap;
  flex-shrink: 0;
}

.landing-navbar .btn {
  margin: 0 !important;
}

@media (max-width: 991px) {
  .landing-navbar {
    min-height: 64px;
  }

  .landing-navbar .container {
    padding-left: 14px;
    padding-right: 14px;
    min-height: 64px;
  }

  .landing-navbar .navbar-translate {
    min-height: 64px;
  }

  .lang-select-nav {
    width: 92px;
  }

  .landing-cta {
    min-height: 34px;
    padding: 0 10px !important;
  }

  .landing-nav {
    gap: 6px;
  }
}

@media (max-width: 520px) {
  .landing-navbar .container {
    padding-left: 8px;
    padding-right: 8px;
  }

  .landing-brand {
    font-size: 10px !important;
    letter-spacing: 0.05em !important;
    max-width: 70px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .landing-nav {
    gap: 3px !important;
  }

  .lang-select-nav {
    width: 68px !important;
  }

  .lang-select-nav :deep(.ant-select-selector) {
    padding: 0 4px !important;
  }

  .landing-cta {
    padding: 0 8px !important;
    font-size: 10px !important;
    min-height: 30px !important;
    margin-right: 0 !important;
  }
}

@media (max-width: 400px) {
  .lang-select-nav {
    width: 62px !important;
  }
}
</style>
