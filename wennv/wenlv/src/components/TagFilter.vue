<script setup lang="ts">
/**
 * TagFilter.vue — 多选标签下拉筛选器
 * 功能：搜索标签 → 勾选/取消 → 选中标签以 Chip 展示
 * 交互：点击触发下拉 → 搜索过滤 → 点击勾选 → 点击 Chip × 移除
 */
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useLanguageStore } from '@/stores/language'

const props = defineProps<{
  tags: string[]
  selected: string[]
}>()

const emit = defineEmits<{
  'update:selected': [tags: string[]]
  clear: []
}>()

const langStore = useLanguageStore()

/* ── 下拉状态 ── */
const open = ref(false)
const search = ref('')
const listRef = ref<HTMLDivElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)

/* 聚焦索引（键盘导航） */
const focusIndex = ref(-1)

const filteredTags = computed(() => {
  if (!search.value.trim()) return props.tags
  const q = search.value.trim().toLowerCase()
  return props.tags.filter(t => t.toLowerCase().includes(q))
})

function toggleTag(tag: string) {
  const idx = props.selected.indexOf(tag)
  if (idx >= 0) {
    emit('update:selected', props.selected.filter(t => t !== tag))
  } else {
    emit('update:selected', [...props.selected, tag])
  }
}

function removeTag(tag: string, event?: MouseEvent) {
  event?.stopPropagation()
  emit('update:selected', props.selected.filter(t => t !== tag))
}

function clearAll() {
  emit('clear')
}

function openDropdown() {
  open.value = true
  search.value = ''
  focusIndex.value = -1
  // 等下一帧让 input 聚焦
  requestAnimationFrame(() => {
    searchInputRef.value?.focus()
  })
}

function closeDropdown() {
  open.value = false
  search.value = ''
  focusIndex.value = -1
}

/* ── 键盘导航 ── */
function onKeydown(e: KeyboardEvent) {
  if (!open.value) {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
      e.preventDefault()
      openDropdown()
    }
    return
  }

  switch (e.key) {
    case 'Escape':
      e.preventDefault()
      closeDropdown()
      triggerRef.value?.focus()
      break
    case 'ArrowDown':
      e.preventDefault()
      focusIndex.value = Math.min(focusIndex.value + 1, filteredTags.value.length - 1)
      scrollIntoView()
      break
    case 'ArrowUp':
      e.preventDefault()
      focusIndex.value = Math.max(focusIndex.value - 1, 0)
      scrollIntoView()
      break
    case 'Enter':
    case ' ':
      if (focusIndex.value >= 0 && focusIndex.value < filteredTags.value.length) {
        e.preventDefault()
        const tag = filteredTags.value[focusIndex.value]
        if (tag) toggleTag(tag)
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
function onDocumentClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!open.value) return
  if (triggerRef.value?.contains(target) || listRef.value?.contains(target)) return
  closeDropdown()
}

onMounted(() => document.addEventListener('mousedown', onDocumentClick))
onBeforeUnmount(() => document.removeEventListener('mousedown', onDocumentClick))

/* 选中变化时更新聚焦索引 */
watch(filteredTags, () => {
  focusIndex.value = -1
})
</script>

<template>
  <div class="tag-filter" :class="{ 'tag-filter--open': open }">
    <!-- 触发器 -->
    <button
      ref="triggerRef"
      class="tag-filter__trigger"
      :class="{ 'tag-filter__trigger--active': selected.length > 0 }"
      @click="open ? closeDropdown() : openDropdown()"
      @keydown="onKeydown"
      :aria-expanded="open"
      aria-haspopup="listbox"
    >
      <svg class="tag-filter__trigger-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
        <line x1="7" y1="7" x2="7.01" y2="7"/>
      </svg>
      <span class="tag-filter__trigger-label">
        {{ langStore.lang === 'zh' ? '标签筛选' : 'Tags' }}
      </span>
      <span v-if="selected.length > 0" class="tag-filter__trigger-badge">{{ selected.length }}</span>
      <svg
        class="tag-filter__trigger-chevron"
        :class="{ 'tag-filter__trigger-chevron--flip': open }"
        width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"
      >
        <path d="M6 9l6 6 6-6"/>
      </svg>
    </button>

    <!-- 下拉面板 -->
    <Transition name="tag-drop">
      <div v-if="open" ref="listRef" class="tag-filter__dropdown" role="listbox" :aria-multiselectable="true">
        <!-- 搜索框 -->
        <div class="tag-filter__search">
          <svg class="tag-filter__search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
          </svg>
          <input
            ref="searchInputRef"
            v-model="search"
            class="tag-filter__search-input"
            :placeholder="langStore.lang === 'zh' ? '搜索标签...' : 'Search tags...'"
            @keydown.stop="onKeydown"
          />
        </div>

        <!-- 标签列表 -->
        <div class="tag-filter__list">
          <div
            v-for="(tag, i) in filteredTags"
            :key="tag"
            class="tag-filter__option"
            :class="{
              'tag-filter__option--selected': selected.includes(tag),
              'tag-filter__option--focused': focusIndex === i,
            }"
            role="option"
            :aria-selected="selected.includes(tag)"
            @click="toggleTag(tag)"
            @mouseenter="focusIndex = i"
          >
            <span class="tag-filter__check">
              <svg v-if="selected.includes(tag)" width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
              </svg>
            </span>
            <span class="tag-filter__option-label">{{ tag }}</span>
          </div>
          <div v-if="filteredTags.length === 0" class="tag-filter__empty">
            {{ langStore.lang === 'zh' ? '无匹配标签' : 'No matching tags' }}
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.tag-filter {
  position: relative;
}

/* ========================================
   触发器
   ======================================== */
.tag-filter__trigger {
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

.tag-filter__trigger:hover {
  border-color: var(--color-gold-dark);
  color: var(--color-text-primary);
}

.tag-filter__trigger:focus-visible {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.tag-filter__trigger--active {
  border-color: var(--color-gold-dark);
  color: var(--color-gold);
}

.tag-filter__trigger-icon {
  flex-shrink: 0;
  opacity: 0.7;
}

.tag-filter__trigger-label {
  flex: 1;
  text-align: left;
}

.tag-filter__trigger-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: var(--radius-full);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
}

.tag-filter__trigger-chevron {
  flex-shrink: 0;
  transition: transform var(--transition-fast);
  opacity: 0.6;
}

.tag-filter__trigger-chevron--flip {
  transform: rotate(180deg);
}

/* ========================================
   下拉面板
   ======================================== */
.tag-filter__dropdown {
  position: absolute;
  top: calc(100% + var(--space-2));
  left: 0;
  z-index: 50;
  width: 260px;
  background: var(--color-surface-elevated);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg), 0 0 0 1px var(--color-border);
  overflow: hidden;
}

/* 搜索框 */
.tag-filter__search {
  position: relative;
  padding: var(--space-3);
  border-bottom: 1px solid var(--color-border);
}

.tag-filter__search-icon {
  position: absolute;
  left: var(--space-5);
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-muted);
  pointer-events: none;
}

.tag-filter__search-input {
  width: 100%;
  padding: var(--space-2) var(--space-3) var(--space-2) var(--space-8);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-sm);
  outline: none;
  transition: border-color var(--transition-fast);
}

.tag-filter__search-input:focus {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 1px var(--color-gold-glow);
}

.tag-filter__search-input::placeholder {
  color: var(--color-text-muted);
}

/* 标签列表 */
.tag-filter__list {
  max-height: 240px;
  overflow-y: auto;
  padding: var(--space-1);
}

.tag-filter__option {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--transition-fast);
  user-select: none;
}

.tag-filter__option:hover,
.tag-filter__option--focused {
  background: var(--color-surface-hover);
}

.tag-filter__option--selected {
  color: var(--color-gold);
}

.tag-filter__option--selected.tag-filter__option--focused {
  background: color-mix(in srgb, var(--color-gold) 8%, var(--color-surface-hover));
}

.tag-filter__check {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 3px;
  border: 1.5px solid var(--color-border);
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.tag-filter__option--selected .tag-filter__check {
  background: var(--color-gold);
  border-color: var(--color-gold);
  color: var(--color-text-inverse);
}

.tag-filter__option-label {
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
}

.tag-filter__option--selected .tag-filter__option-label {
  color: var(--color-gold);
}

.tag-filter__empty {
  padding: var(--space-6) var(--space-3);
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   过渡动画
   ======================================== */
.tag-drop-enter-active {
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.tag-drop-leave-active {
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.tag-drop-enter-from {
  opacity: 0;
  transform: translateY(-8px) scale(0.96);
}

.tag-drop-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.97);
}
</style>