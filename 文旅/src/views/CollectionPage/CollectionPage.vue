<template>
  <div class="collection-page">
    <!-- ===== 顶部工具栏 ===== -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="toolbar-title">收藏馆</h2>
        <span class="toolbar-count">{{ filteredPhotos.length }} 张照片</span>
      </div>
      <button class="upload-btn" @click="showUploader = true">
        <span>+</span> 上传照片
      </button>
    </div>

    <!-- ===== 景点分类标签页 ===== -->
    <div class="spot-tabs">
      <button
        v-for="spot in spotTabs"
        :key="spot.id"
        class="spot-tab"
        :class="{ active: activeSpot === spot.id }"
        @click="activeSpot = spot.id"
      >
        <span class="st-emoji">{{ spot.emoji }}</span>
        {{ spot.name }}
      </button>
    </div>

    <!-- ===== 照片网格 ===== -->
    <div class="photo-grid">
      <div
        v-for="photo in filteredPhotos"
        :key="photo.id"
        class="photo-card"
        @click="previewPhoto = photo"
      >
        <!-- 视觉区 -->
        <div class="pc-visual" :style="{ background: photo.gradient }">
          <span class="pc-emoji">{{ photo.emoji }}</span>
          <!-- 可见性徽标 -->
          <span class="pc-badge" :class="`badge--${photo.visibility}`">
            {{ visibilityLabel(photo.visibility) }}
          </span>
        </div>

        <!-- 信息区 -->
        <div class="pc-body">
          <div class="pc-spot">{{ photo.spotName }}</div>
          <div class="pc-actions">
            <button class="pc-like" :class="{ liked: photo.liked }" @click.stop="toggleLike(photo)">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
              </svg>
              <span>{{ photo.likeCount }}</span>
            </button>
            <div class="pc-vis" @click.stop>
              <button class="vis-btn" @click="toggleVisMenu(photo)">🔒</button>
              <div class="vis-menu" v-if="photo.showVisMenu">
                <button :class="{ selected: photo.visibility === 'public' }" @click="setVisibility(photo, 'public')">🌐 公开可见</button>
                <button :class="{ selected: photo.visibility === 'friends' }" @click="setVisibility(photo, 'friends')">👥 仅朋友可见</button>
                <button :class="{ selected: photo.visibility === 'private' }" @click="setVisibility(photo, 'private')">🔒 仅自己可见</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 空状态 ===== -->
    <div class="empty-state" v-if="filteredPhotos.length === 0">
      <span class="empty-emoji">📸</span>
      <p>该分类暂无照片</p>
      <button class="empty-upload" @click="showUploader = true">去上传一张</button>
    </div>

    <!-- ===== Lightbox 预览 ===== -->
    <transition name="lightbox">
      <div class="lightbox-overlay" v-if="previewPhoto" @click.self="previewPhoto = null">
        <div class="lightbox-content">
          <button class="lb-close" @click="previewPhoto = null">✕</button>
          <div class="lb-visual" :style="{ background: previewPhoto.gradient }">
            <span class="lb-emoji">{{ previewPhoto.emoji }}</span>
          </div>
          <div class="lb-meta">
            <span class="lb-spot">{{ previewPhoto.spotName }}</span>
            <span class="lb-vis">{{ visibilityLabel(previewPhoto.visibility) }}</span>
          </div>
        </div>
      </div>
    </transition>

    <!-- ===== 上传弹窗 ===== -->
    <transition name="drawer-fade">
      <div class="lightbox-overlay" v-if="showUploader" @click.self="showUploader = false">
        <div class="upload-dialog" @click.stop>
          <div class="ud-header">
            <h3>添加照片到景点</h3>
            <button class="dh-close" @click="showUploader = false">✕</button>
          </div>

          <div class="ud-body">
            <!-- 预览 -->
            <div class="ud-preview" v-if="uploadPreview" :style="{ background: uploadGradient }">
              <span class="ud-preview-emoji">{{ uploadEmoji }}</span>
            </div>
            <div class="ud-preview ud-preview--empty" v-else @click="triggerFileInput">
              <span>+</span>
              <span class="ud-preview-hint">点击选择图片</span>
            </div>
            <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onFileSelect" />

            <!-- 表单 -->
            <div class="ud-form">
              <div class="ud-field">
                <label>选择景点</label>
                <select v-model="uploadSpot" class="ud-select">
                  <option v-for="s in spotTabs.slice(1)" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
              </div>
              <div class="ud-field">
                <label>标题</label>
                <input v-model="uploadTitle" class="ud-input" placeholder="给这张照片起个名字" />
              </div>
              <div class="ud-field">
                <label>可见性</label>
                <div class="ud-vis-options">
                  <label class="ud-radio">
                    <input type="radio" v-model="uploadVisibility" value="public" />
                    <span>🌐 公开</span>
                  </label>
                  <label class="ud-radio">
                    <input type="radio" v-model="uploadVisibility" value="friends" />
                    <span>👥 朋友</span>
                  </label>
                  <label class="ud-radio">
                    <input type="radio" v-model="uploadVisibility" value="private" />
                    <span>🔒 私密</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <div class="ud-footer">
            <button class="ud-cancel" @click="showUploader = false">取消</button>
            <button class="ud-confirm" @click="confirmUpload" :disabled="!uploadSpot">确认上传</button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'

// ===== 景点分类 =====
const spotTabs = [
  { id: 'all', name: '全部照片', emoji: '🖼️' },
  { id: 'panda', name: '大熊猫基地', emoji: '🐼' },
  { id: 'kuanzhai', name: '宽窄巷子', emoji: '🏛️' },
  { id: 'qingcheng', name: '青城山', emoji: '🌿' },
  { id: 'wuhou', name: '武侯祠', emoji: '⚔️' },
  { id: 'jinli', name: '锦里古街', emoji: '🏮' },
  { id: 'dufu', name: '杜甫草堂', emoji: '📜' },
  { id: 'dujiangyan', name: '都江堰', emoji: '🌊' },
  { id: 'jinsha', name: '金沙遗址', emoji: '🌟' },
]

const spotGradients = {
  panda: 'linear-gradient(135deg, #10B981, #34D399)',
  kuanzhai: 'linear-gradient(135deg, #E85D3A, #F5A623)',
  qingcheng: 'linear-gradient(135deg, #0EA5A0, #34D399)',
  wuhou: 'linear-gradient(135deg, #C43E1C, #E85D3A)',
  jinli: 'linear-gradient(135deg, #D4A854, #F5A623)',
  dufu: 'linear-gradient(135deg, #6B7280, #9CA3AF)',
  dujiangyan: 'linear-gradient(135deg, #0EA5A0, #06B6D4)',
  jinsha: 'linear-gradient(135deg, #D4A854, #E8B84B)',
}

const spotIds = ['panda', 'kuanzhai', 'qingcheng', 'wuhou', 'jinli', 'dufu', 'dujiangyan', 'jinsha']

// ===== 生成模拟照片数据 =====
function generatePhotos() {
  const photos = []
  let id = 1
  for (let i = 0; i < 24; i++) {
    const idx = i % spotIds.length
    const spotId = spotIds[idx]
    const tab = spotTabs.find(t => t.id === spotId)
    const batch = Math.floor(i / spotIds.length) + 1
    photos.push({
      id: id++,
      spotId,
      spotName: tab.name,
      emoji: tab.emoji,
      gradient: spotGradients[spotId] || 'linear-gradient(135deg, #E85D3A, #F5A623)',
      title: `${tab.name} · 第${batch}张`,
      visibility: ['public', 'friends', 'private'][Math.floor(Math.random() * 3)],
      liked: Math.random() > 0.65,
      likeCount: Math.floor(Math.random() * 80) + 5,
      showVisMenu: false
    })
  }
  return photos
}

const photos = reactive(generatePhotos())
const activeSpot = ref('all')
const previewPhoto = ref(null)
const showUploader = ref(false)
const fileInput = ref(null)

// 上传表单
const uploadPreview = ref('')
const uploadEmoji = ref('🖼️')
const uploadGradient = ref('linear-gradient(135deg, #E85D3A, #F5A623)')
const uploadSpot = ref('panda')
const uploadTitle = ref('')
const uploadVisibility = ref('public')

// ===== 计算过滤后照片 =====
const filteredPhotos = computed(() => {
  if (activeSpot.value === 'all') return photos
  return photos.filter(p => p.spotId === activeSpot.value)
})

// ===== 可见性方法 =====
function visibilityLabel(level) {
  const map = { public: '公开', friends: '朋友', private: '私密' }
  return map[level] || level
}

function toggleVisMenu(photo) {
  // 关闭其他菜单
  photos.forEach(p => { if (p !== photo) p.showVisMenu = false })
  photo.showVisMenu = !photo.showVisMenu
}

function setVisibility(photo, level) {
  photo.visibility = level
  photo.showVisMenu = false
}

// ===== 点赞 =====
function toggleLike(photo) {
  photo.liked = !photo.liked
  photo.likeCount += photo.liked ? 1 : -1
}

// ===== 上传相关 =====
function triggerFileInput() {
  fileInput.value?.click()
}

function onFileSelect(e) {
  const file = e.target.files?.[0]
  if (file) {
    uploadPreview.value = URL.createObjectURL(file)
    // 根据选择的景点更新预览表情
    const tab = spotTabs.find(t => t.id === uploadSpot.value)
    uploadEmoji.value = tab?.emoji || '🖼️'
    uploadGradient.value = spotGradients[uploadSpot.value] || 'linear-gradient(135deg, #E85D3A, #F5A623)'
  }
}

function confirmUpload() {
  if (!uploadSpot.value || !uploadPreview.value) return

  const tab = spotTabs.find(t => t.id === uploadSpot.value)
  photos.unshift({
    id: Date.now(),
    spotId: uploadSpot.value,
    spotName: tab?.name || '未知',
    emoji: tab?.emoji || '🖼️',
    gradient: spotGradients[uploadSpot.value] || 'linear-gradient(135deg, #E85D3A, #F5A623)',
    title: uploadTitle.value || `${tab?.name || ''} · 新照片`,
    visibility: uploadVisibility.value,
    liked: false,
    likeCount: 0,
    showVisMenu: false
  })

  // 重置
  showUploader.value = false
  uploadPreview.value = ''
  uploadTitle.value = ''
  uploadVisibility.value = 'public'
}
</script>

<style scoped>
.collection-page {
  height: 100%;
  overflow-y: auto;
  padding: 0 20px 24px;
  background: #FCF7F2;
}
.collection-page::-webkit-scrollbar { width: 3px; }
.collection-page::-webkit-scrollbar-thumb { background: rgba(232,93,58,.12); border-radius: 2px; }

/* ===== 工具栏 ===== */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 0 16px;
}
.toolbar-left { display: flex; align-items: baseline; gap: 10px; }
.toolbar-title {
  font-family: "Noto Serif SC", "Source Han Serif SC", "Songti SC", serif;
  font-size: 1.2rem;
  font-weight: 700;
  color: #2D2D3A;
  letter-spacing: 2px;
}
.toolbar-count { font-size: .6rem; color: #9CA3AF; }

.upload-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 10px;
  border: none;
  background: linear-gradient(135deg, #E85D3A, #D4A854);
  color: #fff;
  font-size: .72rem;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(232,93,58,.18);
  transition: all .3s;
}
.upload-btn:hover { transform: translateY(-1px); box-shadow: 0 6px 20px rgba(232,93,58,.25); }
.upload-btn span { font-size: 1rem; line-height: 1; }

/* ===== 分类标签 ===== */
.spot-tabs {
  display: flex;
  gap: 6px;
  margin-bottom: 16px;
  overflow-x: auto;
  padding-bottom: 2px;
  -webkit-overflow-scrolling: touch;
}
.spot-tabs::-webkit-scrollbar { display: none; }

.spot-tab {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid rgba(0,0,0,.04);
  background: rgba(255,255,255,.6);
  color: #6B7280;
  font-size: .66rem;
  font-family: inherit;
  white-space: nowrap;
  cursor: pointer;
  transition: all .3s;
}
.spot-tab:hover { background: rgba(255,255,255,.9); border-color: rgba(232,93,58,.1); color: #2D2D3A; }
.spot-tab.active {
  background: rgba(232,93,58,.1);
  border-color: rgba(232,93,58,.2);
  color: #E85D3A;
  font-weight: 600;
}
.st-emoji { font-size: .8rem; line-height: 1; }

/* ===== 照片网格 ===== */
.photo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  gap: 14px;
  padding-bottom: 20px;
}

.photo-card {
  border-radius: 12px;
  overflow: hidden;
  background: #FFFFFF;
  border: 1px solid rgba(0,0,0,.04);
  cursor: pointer;
  transition: transform .35s cubic-bezier(0.16,1,0.3,1), box-shadow .35s;
}
.photo-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 32px rgba(0,0,0,.05);
}

.pc-visual {
  position: relative;
  aspect-ratio: 4 / 3;
  display: flex;
  align-items: center;
  justify-content: center;
}
.pc-emoji { font-size: 2.4rem; filter: drop-shadow(0 2px 8px rgba(0,0,0,.08)); }

.pc-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: .48rem;
  font-weight: 600;
  letter-spacing: 1px;
  backdrop-filter: blur(4px);
}
.badge--public { background: rgba(16,185,129,.8); color: #fff; }
.badge--friends { background: rgba(59,130,246,.8); color: #fff; }
.badge--private { background: rgba(107,114,128,.8); color: #fff; }

.pc-body {
  padding: 10px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.pc-spot { font-size: .64rem; color: #6B7280; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pc-actions { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }

/* 点赞 */
.pc-like {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 3px 8px;
  border-radius: 6px;
  border: none;
  background: rgba(232,93,58,.04);
  color: #9CA3AF;
  font-size: .6rem;
  font-family: inherit;
  cursor: pointer;
  transition: all .25s;
}
.pc-like:hover { background: rgba(232,93,58,.08); color: #E85D3A; }
.pc-like.liked { background: rgba(232,93,58,.1); color: #E85D3A; }
.pc-like.liked svg { fill: #E85D3A; }

/* 可见性菜单 */
.pc-vis { position: relative; }
.vis-btn {
  width: 26px; height: 26px;
  border-radius: 6px;
  border: none;
  background: rgba(0,0,0,.03);
  cursor: pointer;
  font-size: .7rem;
  transition: background .2s;
}
.vis-btn:hover { background: rgba(0,0,0,.06); }

.vis-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  min-width: 130px;
  background: #FFF;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,.06);
  box-shadow: 0 8px 24px rgba(0,0,0,.08);
  z-index: 50;
  overflow: hidden;
}
.vis-menu button {
  display: block;
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: transparent;
  font-size: .65rem;
  text-align: left;
  cursor: pointer;
  font-family: inherit;
  color: #4B5563;
  transition: background .15s;
}
.vis-menu button:hover { background: rgba(232,93,58,.06); color: #E85D3A; }
.vis-menu button.selected { color: #E85D3A; font-weight: 600; }

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 60px 0;
  color: #9CA3AF;
  font-size: .72rem;
}
.empty-emoji { font-size: 2.4rem; }
.empty-upload {
  padding: 8px 20px;
  border-radius: 8px;
  border: 1px solid rgba(232,93,58,.15);
  background: transparent;
  color: #E85D3A;
  font-size: .68rem;
  font-family: inherit;
  cursor: pointer;
  transition: all .25s;
}
.empty-upload:hover { background: rgba(232,93,58,.06); }

/* ===== Lightbox ===== */
.lightbox-overlay {
  position: fixed;
  inset: 0;
  z-index: 500;
  background: rgba(0,0,0,.55);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30px;
}

.lightbox-content {
  max-width: 500px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  position: relative;
}

.lb-close {
  position: absolute;
  top: -40px;
  right: 0;
  width: 32px; height: 32px;
  border-radius: 8px;
  border: none;
  background: rgba(255,255,255,.15);
  color: #fff;
  font-size: .9rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background .2s;
}
.lb-close:hover { background: rgba(255,255,255,.25); }

.lb-visual {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.lb-emoji { font-size: 4rem; filter: drop-shadow(0 4px 12px rgba(0,0,0,.1)); }

.lb-meta {
  display: flex;
  gap: 14px;
  color: rgba(255,255,255,.7);
  font-size: .7rem;
}

/* Lightbox 过渡 */
.lightbox-enter-active,
.lightbox-leave-active { transition: opacity .3s ease; }
.lightbox-enter-from,
.lightbox-leave-to { opacity: 0; }

/* ===== 上传弹窗 ===== */
.upload-dialog {
  width: 100%;
  max-width: 420px;
  background: #FFF;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 24px 64px rgba(0,0,0,.15);
}

.ud-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 20px;
  border-bottom: 1px solid rgba(0,0,0,.04);
}
.ud-header h3 {
  font-size: .88rem;
  font-weight: 700;
  color: #2D2D3A;
}
.dh-close {
  width: 28px; height: 28px;
  border-radius: 8px;
  border: none; background: transparent;
  color: #9CA3AF; cursor: pointer;
  font-size: .8rem;
  display: flex; align-items: center; justify-content: center;
  transition: all .25s;
}
.dh-close:hover { color: #E85D3A; background: rgba(232,93,58,.06); }

.ud-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.ud-preview {
  width: 100%;
  height: 160px;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all .3s;
}
.ud-preview--empty {
  background: rgba(0,0,0,.03);
  border: 2px dashed rgba(0,0,0,.08);
  color: #9CA3AF;
  font-size: 1.6rem;
  gap: 6px;
}
.ud-preview--empty:hover { border-color: rgba(232,93,58,.2); background: rgba(232,93,58,.03); }
.ud-preview-hint { font-size: .62rem; }
.ud-preview-emoji { font-size: 3rem; }

.ud-form { display: flex; flex-direction: column; gap: 12px; }
.ud-field { display: flex; flex-direction: column; gap: 4px; }
.ud-field label { font-size: .58rem; color: #6B7280; letter-spacing: 1px; }

.ud-select, .ud-input {
  padding: 9px 12px;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,.06);
  background: rgba(255,255,255,.8);
  color: #2D2D3A;
  font-size: .78rem;
  font-family: inherit;
  outline: none;
  transition: border-color .3s;
}
.ud-select:focus, .ud-input:focus { border-color: rgba(232,93,58,.25); }

.ud-vis-options {
  display: flex;
  gap: 12px;
  padding-top: 4px;
}
.ud-radio {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: .68rem;
  color: #6B7280;
  cursor: pointer;
}
.ud-radio input:checked + span { color: #E85D3A; }

.ud-footer {
  display: flex;
  gap: 8px;
  padding: 12px 20px 18px;
}
.ud-cancel, .ud-confirm {
  flex: 1;
  padding: 10px;
  border-radius: 8px;
  font-size: .72rem;
  font-family: inherit;
  cursor: pointer;
  transition: all .25s;
}
.ud-cancel {
  border: 1px solid rgba(0,0,0,.06);
  background: transparent;
  color: #6B7280;
}
.ud-cancel:hover { background: rgba(0,0,0,.03); }
.ud-confirm {
  border: none;
  background: linear-gradient(135deg, #E85D3A, #D4A854);
  color: #fff;
  font-weight: 600;
}
.ud-confirm:disabled { opacity: .4; cursor: not-allowed; }
.ud-confirm:hover:not(:disabled) { box-shadow: 0 4px 16px rgba(232,93,58,.2); }

.drawer-fade-enter-active,
.drawer-fade-leave-active { transition: opacity .25s ease; }
.drawer-fade-enter-from,
.drawer-fade-leave-to { opacity: 0; }

/* ===== 响应式 ===== */
@media (max-width: 640px) {
  .photo-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .toolbar { flex-direction: column; align-items: flex-start; gap: 10px; }
}
</style>
