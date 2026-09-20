<script setup lang="ts">
/**
 * CheckInModal.vue — 景点打卡弹窗（护照集章凭证）
 * 流程：选择 1~9 张现场照片 → 压缩后经 /upload/policy 直传 OSS →
 *       POST /check-ins 落库 MySQL → emit success（父组件播放奖励动画）。
 * 照片压缩复用 ProfilePage 的 canvas 方案（≤1600px JPEG），避免手机原图直传过慢。
 */
import { computed, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useLanguageStore } from '@/stores/language'
import { createCheckIn } from '@/api/checkin'
import { uploadToOss } from '@/api/upload'

const props = defineProps<{
  open: boolean
  scenicId: number
  spotName: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const langStore = useLanguageStore()

const MAX_PHOTOS = 9

/** 待上传预览列表（dataURL，仅本地展示） */
const previews = ref<string[]>([])
/** 压缩后的待上传文件（与 previews 一一对应） */
const pendingFiles = ref<File[]>([])
/** 已上传成功的 OSS URL */
const uploadedUrls = ref<string[]>([])
const comment = ref('')
const submitting = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const canSubmit = computed(
  () => uploadedUrls.value.length > 0 && !submitting.value
)
const progressText = computed(() => {
  const total = previews.value.length
  if (!total) return ''
  return `${langStore.t('checkin.uploadProgress', { done: uploadedUrls.value.length, total })}`
})

const handleClose = (value: boolean) => {
  if (submitting.value) return
  emit('update:open', value)
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      previews.value = []
      pendingFiles.value = []
      uploadedUrls.value = []
      comment.value = ''
      submitting.value = false
    }
  }
)

function triggerPick() {
  fileInputRef.value?.click()
}

async function onFilesPicked(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files ? Array.from(input.files) : []
  input.value = ''
  if (!files.length) return
  const remain = MAX_PHOTOS - previews.value.length
  if (remain <= 0) {
    message.warning(langStore.t('checkin.photoLimit'))
    return
  }
  const picked = files.filter((f) => f.type.startsWith('image/')).slice(0, remain)
  for (const file of picked) {
    try {
      const compressed = await compressImage(file)
      previews.value.push(await fileToDataUrl(compressed))
      pendingFiles.value.push(compressed)
    } catch {
      /* 单张失败不阻塞其余 */
    }
  }
}

function removePreview(index: number) {
  if (submitting.value) return
  previews.value.splice(index, 1)
  pendingFiles.value.splice(index, 1)
  uploadedUrls.value.splice(index, 1)
}

/** canvas 压缩：最长边 ≤1600px、JPEG 质量 0.8（不足时原样返回 File） */
function compressImage(file: File, maxSide = 1600, quality = 0.8): Promise<File> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const img = new Image()
      img.onload = () => {
        let { width: w, height: h } = img
        if (w <= maxSide && h <= maxSide) {
          resolve(file)
          return
        }
        if (w > h) {
          h = Math.round((h * maxSide) / w)
          w = maxSide
        } else {
          w = Math.round((w * maxSide) / h)
          h = maxSide
        }
        const canvas = document.createElement('canvas')
        canvas.width = w
        canvas.height = h
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          resolve(file)
          return
        }
        ctx.drawImage(img, 0, 0, w, h)
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              resolve(file)
              return
            }
            resolve(new File([blob], file.name.replace(/\.\w+$/, '') + '.jpg', { type: 'image/jpeg' }))
          },
          'image/jpeg',
          quality
        )
      }
      img.onerror = () => reject(new Error('decode failed'))
      img.src = String(reader.result ?? '')
    }
    reader.onerror = () => reject(new Error('read failed'))
    reader.readAsDataURL(file)
  })
}

function fileToDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(new Error('read failed'))
    reader.readAsDataURL(file)
  })
}

/** 提交：逐张直传 OSS → 落库打卡 */
async function submit() {
  if (!canSubmit.value) return
  if (uploadedUrls.value.length < pendingFiles.value.length) {
    submitting.value = true
    try {
      const urls: string[] = []
      for (const file of pendingFiles.value) {
        urls.push(await uploadToOss(file))
      }
      uploadedUrls.value = urls
    } catch {
      message.error(langStore.t('checkin.uploadFailed'))
      submitting.value = false
      return
    }
    submitting.value = false
  }
  submitting.value = true
  try {
    await createCheckIn({
      scenic_id: props.scenicId,
      photos: uploadedUrls.value,
      comment: comment.value.trim() || undefined,
    })
    emit('update:open', false)
    emit('success')
  } catch (err) {
    const msg = err instanceof Error ? err.message : ''
    message.error(msg || langStore.t('checkin.submitFailed'))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="checkin-mask" @click.self="handleClose(false)">
      <div class="checkin-modal" role="dialog" aria-modal="true">
        <header class="checkin-modal__head">
          <div class="checkin-modal__titles">
            <h3 class="checkin-modal__title">{{ langStore.t('checkin.modalTitle') }}</h3>
            <p class="checkin-modal__sub">{{ spotName }}</p>
          </div>
          <button type="button" class="checkin-modal__close" @click="handleClose(false)">×</button>
        </header>

        <p class="checkin-modal__desc">{{ langStore.t('checkin.modalDesc') }}</p>

        <!-- 照片选择与预览 -->
        <div class="checkin-photos">
          <div v-for="(src, idx) in previews" :key="src" class="checkin-photos__cell">
            <img :src="src" class="checkin-photos__img" :alt="`photo-${idx + 1}`" />
            <button
              type="button"
              class="checkin-photos__del"
              :disabled="submitting"
              @click="removePreview(idx)"
            >×</button>
            <span class="checkin-photos__idx">{{ idx + 1 }}</span>
          </div>

          <button
            v-if="previews.length < MAX_PHOTOS"
            type="button"
            class="checkin-photos__add"
            :disabled="submitting"
            @click="triggerPick"
          >
            <span class="checkin-photos__plus">＋</span>
            <span class="checkin-photos__hint">{{ langStore.t('checkin.pickPhotos') }}</span>
          </button>
        </div>
        <input
          ref="fileInputRef"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          class="checkin-modal__hidden-input"
          @change="onFilesPicked"
        />
        <p v-if="previews.length" class="checkin-modal__progress">{{ progressText }}</p>

        <!-- 感言（可选） -->
        <textarea
          v-model="comment"
          class="checkin-modal__comment"
          rows="2"
          maxlength="100"
          :placeholder="langStore.t('checkin.commentPlaceholder')"
        />

        <footer class="checkin-modal__foot">
          <button type="button" class="checkin-modal__cancel" :disabled="submitting" @click="handleClose(false)">
            {{ langStore.t('checkin.cancel') }}
          </button>
          <button type="button" class="checkin-modal__submit" :disabled="!canSubmit" @click="submit">
            {{ submitting ? langStore.t('checkin.submitting') : langStore.t('checkin.submit') }}
          </button>
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.checkin-mask {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  background: rgba(20, 16, 12, 0.62);
  backdrop-filter: blur(4px);
}

.checkin-modal {
  width: min(520px, 100%);
  max-height: 86vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-6);
  background: var(--color-surface);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
}

.checkin-modal__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.checkin-modal__titles {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.checkin-modal__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-cinnabar);
}

.checkin-modal__sub {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-primary);
}

.checkin-modal__close {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: var(--color-bg-alt);
  color: var(--color-text-secondary);
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
}

.checkin-modal__desc {
  margin: 0;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  line-height: var(--leading-normal);
}

.checkin-photos {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-2);
}

.checkin-photos__cell {
  position: relative;
  aspect-ratio: 1;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.checkin-photos__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.checkin-photos__del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.62);
  color: #fff;
  font-size: 13px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.checkin-photos__idx {
  position: absolute;
  left: 4px;
  bottom: 4px;
  min-width: 18px;
  padding: 0 4px;
  border-radius: var(--radius-sm);
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  font-size: 10px;
  line-height: 16px;
  text-align: center;
}

.checkin-photos__add {
  aspect-ratio: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  border: 1px dashed color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-cinnabar) 3%, var(--color-surface));
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.checkin-photos__add:hover:not(:disabled) {
  border-color: var(--color-cinnabar);
  background: color-mix(in srgb, var(--color-cinnabar) 8%, var(--color-surface));
}

.checkin-photos__add:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.checkin-photos__plus {
  font-size: var(--text-2xl);
  color: var(--color-cinnabar);
  line-height: 1;
}

.checkin-photos__hint {
  font-size: 10px;
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.checkin-modal__hidden-input {
  display: none;
}

.checkin-modal__progress {
  margin: 0;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  text-align: right;
}

.checkin-modal__comment {
  padding: var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-alt);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-sm);
  resize: none;
  outline: none;
}

.checkin-modal__comment:focus {
  border-color: var(--color-cinnabar);
}

.checkin-modal__foot {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

.checkin-modal__cancel {
  padding: var(--space-3) var(--space-5);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
}

.checkin-modal__submit {
  padding: var(--space-3) var(--space-6);
  border: none;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--color-cinnabar), #8f2a1e);
  color: #fff;
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.checkin-modal__submit:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px color-mix(in srgb, var(--color-cinnabar) 35%, transparent);
}

.checkin-modal__submit:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

@media (max-width: 640px) {
  .checkin-photos {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
