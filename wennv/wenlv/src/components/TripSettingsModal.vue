<template>
  <a-modal
    :open="open"
    class="settings-modal"
    :title="t('settings.title')"
    :width="820"
    :confirm-loading="saving"
    :ok-text="t('settings.saveApply')"
    :cancel-text="t('settings.cancel')"
    @update:open="emit('update:open', $event)"
    @ok="handleSave"
  >
    <a-spin :spinning="loading">
      <section class="runtime-settings-panel">
        <a-alert class="runtime-settings-notice" type="info" show-icon :message="t('settings.notice')" />
        <a-form layout="vertical" class="runtime-settings-form">
          <div class="runtime-settings-grid">
            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.apiBaseUrl') }}</span>
              </template>
              <a-input
                v-model:value="settingsForm.api_base_url"
                :placeholder="t('settings.placeholders.apiBaseUrl')"
                allow-clear
              />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.amapJsKey') }}</span>
              </template>
              <a-input-password v-model:value="settingsForm.vite_amap_web_js_key" allow-clear />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.amapWebKey') }}</span>
              </template>
              <a-input-password v-model:value="settingsForm.vite_amap_web_key" allow-clear />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.googleMapsApiKey') }}</span>
              </template>
              <a-input-password
                v-model:value="settingsForm.google_maps_api_key"
                :placeholder="t('settings.placeholders.googleMapsApiKey')"
                allow-clear
              />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.googleMapsProxy') }}</span>
              </template>
              <a-input
                v-model:value="settingsForm.google_maps_proxy"
                :placeholder="t('settings.placeholders.googleMapsProxy')"
                allow-clear
              />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.openaiBaseUrl') }}</span>
              </template>
              <a-input
                v-model:value="settingsForm.openai_base_url"
                :placeholder="t('settings.placeholders.openaiBaseUrl')"
                allow-clear
              />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.openaiModel') }}</span>
              </template>
              <a-input
                v-model:value="settingsForm.openai_model"
                :placeholder="t('settings.placeholders.openaiModel')"
                allow-clear
              />
            </a-form-item>

            <a-form-item>
              <template #label>
                <span class="field-label">{{ t('settings.labels.openaiApiKey') }}</span>
              </template>
              <a-input-password v-model:value="settingsForm.openai_api_key" allow-clear />
            </a-form-item>
          </div>

          <a-form-item class="runtime-settings-full">
            <template #label>
              <span class="field-label">{{ t('settings.labels.xhsCookie') }}</span>
            </template>
            <a-textarea
              v-model:value="settingsForm.xhs_cookie"
              :rows="4"
              :placeholder="t('settings.placeholders.xhsCookie')"
              allow-clear
            />
          </a-form-item>

          <a-form-item class="runtime-settings-full">
            <template #label>
              <span class="field-label">{{ t('settings.labels.douyinCookie') }}</span>
            </template>
            <a-textarea
              v-model:value="settingsForm.douyin_cookie"
              :rows="4"
              :placeholder="t('settings.placeholders.douyinCookie')"
              allow-clear
            />
          </a-form-item>
        </a-form>
      </section>
    </a-spin>
  </a-modal>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import type { RuntimeSettings } from '@/trip/types'
import { getRuntimeSettings, saveRuntimeSettings } from '@/trip/services/api'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved', settings: RuntimeSettings): void
}>()

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const settingsForm = reactive<RuntimeSettings>({
  api_base_url: '',
  vite_amap_web_key: '',
  vite_amap_web_js_key: '',
  google_maps_api_key: '',
  google_maps_proxy: '',
  xhs_cookie: '',
  douyin_cookie: '',
  openai_api_key: '',
  openai_base_url: '',
  openai_model: '',
})

const applyRuntimeSettings = (settings: RuntimeSettings) => {
  settingsForm.api_base_url = settings.api_base_url || ''
  settingsForm.vite_amap_web_key = settings.vite_amap_web_key || ''
  settingsForm.vite_amap_web_js_key = settings.vite_amap_web_js_key || ''
  settingsForm.google_maps_api_key = settings.google_maps_api_key || ''
  settingsForm.google_maps_proxy = settings.google_maps_proxy || ''
  settingsForm.xhs_cookie = settings.xhs_cookie || ''
  settingsForm.douyin_cookie = settings.douyin_cookie || ''
  settingsForm.openai_api_key = settings.openai_api_key || ''
  settingsForm.openai_base_url = settings.openai_base_url || ''
  settingsForm.openai_model = settings.openai_model || ''
}

const loadSettings = async () => {
  loading.value = true
  try {
    const settings = await getRuntimeSettings()
    applyRuntimeSettings(settings)
  } catch (error: any) {
    message.error(error?.message || t('settings.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      loadSettings()
    }
  }
)

const handleSave = async () => {
  saving.value = true
  try {
    const saved = await saveRuntimeSettings({ ...settingsForm })
    applyRuntimeSettings(saved)
    message.success(t('settings.messages.saved'))
    emit('update:open', false)
    emit('saved', saved)
  } catch (error: any) {
    message.error(error?.message || t('settings.messages.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.runtime-settings-notice {
  margin-bottom: 14px;
}

.runtime-settings-form :deep(.ant-form-item) {
  margin-bottom: 12px;
}
</style>
