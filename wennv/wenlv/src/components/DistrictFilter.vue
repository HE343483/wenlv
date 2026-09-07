<script setup lang="ts">
/**
 * DistrictFilter.vue — 区域单选下拉筛选器
 * 功能：选择「全部」或具体区县 → 单选 → 选中项金色高亮 + 对勾
 * 交互：点击触发下拉 → 点击选中即收起；支持外部点击 / Esc 关闭、键盘上下键导航
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { districts } from '@/data/chengdu'

const props = defineProps<{
  /** 当前选中的区域 id（'all' 表示全部） */
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [id: string]
}>()

const langStore = useLanguageStore()

/* ── 下拉状态 ── */
const open = ref(false)
const triggerRef = ref<HTMLButtonElement | null>(null)
const listRef = ref<HTMLDivElement | null>(null)
const focusIndex = ref(-1)

/* 选项：全部 + 各区县 */
const options = computed(() => [
  { id: 'all', name: langStore.t('district.allDistricts') },
  ...districts.map(d => ({
    id: d.id,
    name: langStore.lang === 'zh' ? d.nameZh : d.nameEn,
  })),
])

/* 触发按钮显示的当前名称 */
const currentName = computed(
  () => options.value.find(o => o.id === props.modelValue)?.name ?? langStore.t('district.allDistricts'),
)

function toggle() {
  open.value = !open.value
  if (open.value) focusIndex.value = -1
}

function select(id: string) {
  emit('update:modelValue', id)
  open.value = false
  focusIndex.value = -1
  triggerRef.value?.focus()
}

/* ── 键盘导航 ── */
function onKeydown(e: KeyboardEvent) {
  if (!open.value) {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
      e.preventDefault()
      open.value = true
      focusIndex.value = -1
    }
    return
  }

  switch (e.key) {
    case 'Escape':
      e.preventDefault()
      open.value = false
      triggerRef.value?.focus()
      break
    case 'ArrowDown':
      e.preventDefault()
      focusIndex.value = Math.min(focusIndex.value + 1, options.value.length - 1)
      scrollIntoView()
      break
    case 'ArrowUp':
      e.preventDefault()
      focusIndex.value = Math.max(focusIndex.value - 1, 0)
      scrollIntoView()
      break
    case 'Enter':
    case ' ':
      if (focusIndex.value >= 0 && focusIndex.value < options.value.length) {
        e.preventDefault()
        const opt = options.value[focusIndex.value]
        if (opt) select(opt.id)
      }
      break
  }
}

function scrollIntoView() {
  const idx = focusIndex.value
  if (idx < 0) return
  const el = listRef.value?.children[idx] as HTMLElement | undefined
  el?.scrollIntoView({ block: 'nearest' })
}

/* ── 点击外部关闭 ── */
function onOutsideClick(e: MouseEvent) {
  if (!open.value) return
  const t = e.target as HTMLElement
  if (triggerRef.value?.contains(t) || listRef.value?.contains(t)) return
  open.value = false
}

onMounted(() => document.addEventListener('mousedown', onOutsideClick))
onBeforeUnmount(() => document.removeEventListener('mousedown', onOutsideClick))
</script>

<template>
  <div class="district-filter">
    <!-- 触发器 -->
    <button
      ref="triggerRef"
      class="district-filter__trigger"
      @click="toggle"
      @keydown="onKeydown"
      :aria-expanded="open"
      aria-haspopup="listbox"
      :aria-label="langStore.t('district.title')"
    >
      <svg class="district-filter__pin" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 2C8 2 5 5 5 9c0 5 7 13 7 13s7-8 7-13c0-4-3-7-7-7z"/>
        <circle cx="12" cy="9" r="3"/>
      </svg>
      <span class="district-filter__label">{{ currentName }}</span>
      <svg
        class="district-filter__chevron"
        :class="{ 'district-filter__chevron--flip': open }"
        width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"
      >
        <path d="M6 9l6 6 6-6"/>
      </svg>
    </button>

    <!-- 下拉面板 -->
    <Transition name="district-filter-drop">
      <div v-if="open" ref="listRef" class="district-filter__dropdown" role="listbox">
        <div
          v-for="(opt, i) in options"
          :key="opt.id"
          class="district-filter__option"
          :class="{
            'district-filter__option--active': opt.id === modelValue,
            'district-filter__option--focused': focusIndex === i,
          }"
          role="option"
          :aria-selected="opt.id === modelValue"
          @click="select(opt.id)"
          @mouseenter="focusIndex = i"
        >
          <span class="district-filter__option-label">{{ opt.name }}</span>
          <svg
            v-if="opt.id === modelValue"
            class="district-filter__check"
            width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
          >
            <path d="M20 6L9 17l-5-5"/>
          </svg>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.district-filter {
  position: relative;
}

/* ========================================
   触发器
   ======================================== */
.district-filter__trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-family: var(--font-body);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
  outline: none;
  min-width: 150px;
}

.district-filter__trigger:hover {
  border-color: var(--color-gold-dark);
  color: var(--color-text-primary);
}

.district-filter__trigger:focus-visible {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.district-filter__pin {
  flex-shrink: 0;
  opacity: 0.7;
}

.district-filter__label {
  flex: 1;
  text-align: left;
}

.district-filter__chevron {
  flex-shrink: 0;
  opacity: 0.6;
  transition: transform var(--transition-fast);
}

.district-filter__chevron--flip {
  transform: rotate(180deg);
}

/* ========================================
   下拉面板
   ======================================== */
.district-filter__dropdown {
  position: absolute;
  top: calc(100% + var(--space-2));
  left: 0;
  z-index: 50;
  min-width: 200px;
  max-height: 240px;
  overflow-y: auto;
  padding: var(--space-1);
  background: var(--color-surface-elevated);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg), 0 0 0 1px var(--color-border);
}

.district-filter__option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
  user-select: none;
}

.district-filter__option:hover,
.district-filter__option--focused {
  background: var(--color-surface-hover);
}

.district-filter__option--active {
  color: var(--color-gold);
}

.district-filter__option-label {
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
}

.district-filter__option--active .district-filter__option-label {
  color: var(--color-gold);
}

.district-filter__check {
  color: var(--color-gold);
  flex-shrink: 0;
}

/* ========================================
   过渡动画
   ======================================== */
.district-filter-drop-enter-active {
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.district-filter-drop-leave-active {
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.district-filter-drop-enter-from {
  opacity: 0;
  transform: translateY(-8px) scale(0.96);
}

.district-filter-drop-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.97);
}
</style>
