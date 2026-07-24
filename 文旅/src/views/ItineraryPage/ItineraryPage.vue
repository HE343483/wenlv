<template>
  <div class="itinerary-page">
    <!-- ===== 左侧面板 ===== -->
    <aside class="left-panel">
      <!-- 顶部栏 -->
      <div class="lp-header">
        <button class="lp-back" @click="$router.push('/main/home')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
        </button>
        <div class="lp-title">规划行程</div>
        <button class="lp-clear" v-if="routeActive" @click="resetAll">重置</button>
      </div>

      <!-- 搜索框 -->
      <div class="lp-search" :class="{ focus: searchFocused || searchQuery }">
        <span class="lps-icon">🔍</span>
        <input
          v-model="searchQuery"
          class="lps-input"
          placeholder="搜索景点..."
          @focus="searchFocused = true"
          @blur="setTimeout(() => searchFocused = false, 200)"
          @input="onSearchInput"
        />
        <button class="lps-clear" v-if="searchQuery" @click="searchQuery = ''; searchResults = []">✕</button>
        <transition name="drop">
          <div class="lps-suggest" v-if="searchResults.length && searchFocused">
            <div class="lps-item" v-for="s in searchResults" :key="s.id" @mousedown="selectSpot(s)">
              <span class="lps-emoji">{{ s.emoji }}</span>
              <div class="lps-body">
                <div class="lps-name">{{ s.name }}</div>
                <div class="lps-tag">{{ s.tag }}</div>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- 出行方式切换 -->
      <transition name="slide">
        <div class="lp-transport" v-if="routeActive">
          <button
            v-for="m in modes"
            :key="m.key"
            class="lpt-btn"
            :class="{ active: transportMode === m.key }"
            @click="switchMode(m.key)"
          >
            <span class="lpt-icon">{{ m.icon }}</span>
            <span class="lpt-label">{{ m.label }}</span>
          </button>
        </div>
      </transition>

      <!-- 路线信息 -->
      <transition name="slide">
        <div class="lp-route" v-if="routeActive">
          <div class="lpr-head">
            <span class="lpr-icon">{{ currentMode.icon }}</span>
            <div class="lpr-info">
              <span class="lpr-dist">{{ routeInfo.distance }}</span>
              <span class="lpr-time">{{ routeInfo.duration }}</span>
            </div>
          </div>
          <div class="lpr-path">
            <span class="lpr-dot lpr-dot-start"></span>
            <span class="lpr-line"></span>
            <span class="lpr-dot lpr-dot-end" :style="{ background: selectedColor }"></span>
          </div>
          <div class="lpr-dest">
            <div class="lpr-from">{{ routeInfo.from }}</div>
            <div class="lpr-to">{{ routeInfo.to }}</div>
          </div>
        </div>
      </transition>

      <!-- 景点列表 -->
      <div class="lp-spots" ref="spotListRef">
        <div class="lps-label">成都景点</div>
        <div
          v-for="spot in filteredSpots"
          :key="spot.id"
          class="lp-spot"
          :class="{ active: selectedSpot?.id === spot.id }"
          @click="selectSpot(spot)"
        >
          <div class="lps-avatar" :style="{ background: spot.color + '18' }">
            <span>{{ spot.emoji }}</span>
          </div>
          <div class="lps-body">
            <div class="lps-title">{{ spot.name }}</div>
            <div class="lps-desc">{{ spot.tag }}</div>
          </div>
          <div class="lps-rating">⭐ {{ spot.rating }}</div>
        </div>
      </div>

      <!-- AI 点评卡片 -->
      <transition name="card">
        <div class="lp-ai" v-if="selectedSpot" :key="selectedSpot.id">
          <div class="lpa-header">
            <div class="lpa-avatar" :style="{ background: selectedSpot.color + '18' }">
              <span>{{ selectedSpot.emoji }}</span>
            </div>
            <div class="lpa-head-text">
              <div class="lpa-name">{{ selectedSpot.name }}</div>
              <div class="lpa-tag">{{ selectedSpot.tag }}</div>
            </div>
            <button class="lpa-close" @click="closeAICard">✕</button>
          </div>
          <div class="lpa-quote">
            <span class="lpa-quote-text" ref="quoteTextRef"></span>
            <span class="lpa-quote-cursor" v-if="typing">|</span>
          </div>
          <div class="lpa-divider"></div>
          <div class="lpa-info">
            <div class="lpa-info-item"><span class="lpa-info-l">建议游玩</span><span class="lpa-info-v">{{ selectedSpot.duration }}</span></div>
            <div class="lpa-info-item"><span class="lpa-info-l">门票</span><span class="lpa-info-v">{{ selectedSpot.ticket }}</span></div>
            <div class="lpa-info-item"><span class="lpa-info-l">评分</span><span class="lpa-info-v star">⭐ {{ selectedSpot.rating }}</span></div>
          </div>
          <button class="lpa-action" @click="setAsDestination">
            <span>✨ 设为行程下一站</span>
            <span class="lpa-arrow">→</span>
          </button>
        </div>
      </transition>

      <!-- 底部推荐行程 -->
      <div class="lp-recommend" v-if="!selectedSpot && !routeActive">
        <div class="lpr-title">推荐行程</div>
        <div class="lpr-cards">
          <div class="lpr-card" v-for="itin in itineraries" :key="itin.id" @click="selectItinerary(itin)">
            <div class="lprc-top">
              <span class="lprc-name">{{ itin.name }}</span>
              <span class="lprc-badge">{{ itin.badge }}</span>
            </div>
            <div class="lprc-meta">⭐ {{ itin.rating }} · {{ itin.spotCount }} 景点 · {{ itin.days }}</div>
          </div>
        </div>
      </div>

      <!-- 加载 / 空状态 -->
      <div class="lp-empty" v-if="!selectedSpot && !routeActive && filteredSpots.length === 0 && !searchQuery">
        <div class="lpe-icon">🗺️</div>
        <div class="lpe-text">在成都地图上选择景点<br/>或在上方搜索目的地</div>
      </div>
    </aside>

    <!-- ===== 右侧高德地图面板 ===== -->
    <main class="right-panel">
      <div ref="mapContainerRef" class="amap-wrapper"></div>

      <!-- 地图加载中 -->
      <div class="map-overlay" v-if="mapStatus === 'loading'">
        <div class="mo-spinner"></div>
        <span class="mo-text">地图加载中...</span>
      </div>

      <!-- 地图加载失败 -->
      <div class="map-overlay" v-if="mapStatus === 'error'">
        <div class="mo-icon">🗺️</div>
        <div class="mo-title">高德地图加载失败</div>
        <div class="mo-desc">请确保已在 <code>src/config.js</code> 中配置有效的 AMap API Key</div>
        <div class="mo-desc">获取地址：<a href="https://console.amap.com/" target="_blank">高德开放平台</a></div>
      </div>

      <!-- 地图空状态 Air badge -->
      <div class="map-badge" v-if="mapStatus === 'ready' && !routeActive && !selectedSpot">
        <span>点击地图上的📍标记或左侧景点开始规划</span>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import gsap from 'gsap'
import AMapLoader from '@amap/amap-jsapi-loader'

// =================== TYPES ===================
/** @type {any} - AMap loaded dynamically from script */

// =================== CONFIG ===================
// 🔑 高德地图 API Key — 已经配好了
// 如果地图不显示，去 https://console.amap.com/ 检查：
//   1. Key 的服务类型选了「Web端(JS API)」
//   2. 「白名单」添加了 http://localhost:5173
//   3. 「安全密钥 JSAPI 安全设置」→ 复制 securityJsCode 填到下面
const AMAP_KEY = '956e1be8f2d7092b6b1f12dba62bd857d'
const AMAP_SECURITY_CODE = '' // ⬅ 试一下清空后能不能加载，如果不能再去控制台重新生成一个

// =================== DATA ===================
const modes = [
  { key: 'walk',    icon: '🚶', label: '步行',   color: '#0EA5A0' },
  { key: 'drive',   icon: '🚗', label: '驾车',   color: '#E85D3A' },
  { key: 'transit', icon: '🚄', label: '公交',   color: '#8B7EC8' },
  { key: 'bike',    icon: '🚲', label: '骑行',   color: '#F5A623' }
]

const SPOTS = [
  { id:'panda',   name:'大熊猫基地', coords:[104.145,30.733], emoji:'🐼', color:'#10B981', tag:'自然生态 · 国宝', rating:'4.8', duration:'3-4小时', ticket:'¥55',
    quote:'🐼 黑白团子的快乐星球，治愈一切不开心' },
  { id:'kuanzhai',name:'宽窄巷子',   coords:[104.054,30.666], emoji:'🏮', color:'#E85D3A', tag:'历史文化 · 成都名片', rating:'4.6', duration:'2-3小时', ticket:'免费',
    quote:'🏮 宽的是历史，窄的是生活，井里是成都' },
  { id:'wuhou',   name:'武侯祠',     coords:[104.047,30.644], emoji:'⚔️', color:'#E85D3A', tag:'三国文化 · 全国重点文物', rating:'4.7', duration:'2小时', ticket:'¥50',
    quote:'⚔️ 三国迷的精神故乡，出师表里的热血犹在' },
  { id:'jinli',   name:'锦里古街',   coords:[104.048,30.645], emoji:'🏮', color:'#F5A623', tag:'古街 · 美食 · 夜景', rating:'4.5', duration:'1.5-2小时', ticket:'免费',
    quote:'🌙 入夜后的锦里，才是成都最柔软的梦境' },
  { id:'ducao',   name:'杜甫草堂',   coords:[104.031,30.662], emoji:'📜', color:'#34D399', tag:'诗歌文化 · 文学圣地', rating:'4.6', duration:'1.5-2小时', ticket:'¥50',
    quote:'📜 安得广厦千万间，诗圣千年前的叹息' },
  { id:'jinsha',  name:'金沙遗址',   coords:[104.012,30.682], emoji:'🌟', color:'#E8B84B', tag:'古蜀文明 · 太阳神鸟', rating:'4.7', duration:'2-3小时', ticket:'¥70',
    quote:'🌟 三千年前的古蜀先民，把太阳藏进了金箔' },
  { id:'qingcheng',name:'青城山',   coords:[103.578,31.002], emoji:'🌿', color:'#34D399', tag:'道教名山 · 世界遗产', rating:'4.8', duration:'4-6小时', ticket:'¥80',
    quote:'🌿 青城天下幽，修仙问道的绝佳去处' },
  { id:'dujiang', name:'都江堰',     coords:[103.607,31.017], emoji:'🌊', color:'#0EA5A0', tag:'古代工程 · 世界遗产', rating:'4.7', duration:'3-4小时', ticket:'¥80',
    quote:'🌊 两千年的水利奇迹，天府之国的起点' },
  { id:'wenshu',  name:'文殊院',     coords:[104.073,30.675], emoji:'☕', color:'#E85D3A', tag:'佛教文化 · 盖碗茶', rating:'4.5', duration:'1-2小时', ticket:'免费',
    quote:'☕ 一壶盖碗茶，偷得浮生半日闲' },
  { id:'people',  name:'人民公园',   coords:[104.058,30.660], emoji:'🍵', color:'#10B981', tag:'市井文化 · 鹤鸣茶社', rating:'4.4', duration:'1-2小时', ticket:'免费',
    quote:'🍵 成都慢生活的灵魂，都在鹤鸣茶社的竹椅里' }
]

const ORIGIN = [104.071, 30.665] // 天府广场（默认出发点）

const itineraries = [
  { id:1, name:'蓉城市井文化之旅',  badge:'🔥 热门', rating:4.8, spotCount:5, days:'2日', spots:['kuanzhai','wuhou','jinli','people','wenshu'] },
  { id:2, name:'古蜀文明探秘之旅',  badge:'🌟 精选', rating:4.7, spotCount:4, days:'2日', spots:['jinsha','ducao','wuhou','wenshu'] },
  { id:3, name:'熊猫自然生态之旅',  badge:'🐼 亲子', rating:4.9, spotCount:3, days:'1日', spots:['panda','qingcheng','dujiang'] }
]

// =================== STATE ===================
const quoteTextRef = ref(null)
const mapContainerRef = ref(null)

const searchQuery = ref('')
const searchFocused = ref(false)
const searchResults = ref([])
const selectedSpot = ref(null)
const transportMode = ref('walk')
const routeActive = ref(false)
const typing = ref(false)
const mapStatus = ref('loading') // 'loading' | 'ready' | 'error'

const routeInfo = reactive({
  distance: '', duration: '', from: '天府广场', to: ''
})

// =================== COMPUTED ===================
const currentMode = computed(() => modes.find(m => m.key === transportMode.value))
const selectedColor = computed(() => selectedSpot.value?.color || '#E85D3A')

const filteredSpots = computed(() => {
  if (!searchQuery.value) return SPOTS
  const q = searchQuery.value.toLowerCase()
  return SPOTS.filter(s => s.name.includes(q) || s.tag.includes(q))
})

// =================== AMap REFERENCES ===================
let mapInstance = null
let AMapInstance = null
let markers = []
let currentRouting = null

// =================== SEARCH ===================
function onSearchInput() {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) { searchResults.value = []; return }
  searchResults.value = SPOTS.filter(s => s.name.includes(q) || s.tag.includes(q)).slice(0, 5)
}

// =================== SPOT SELECTION ===================
function selectSpot(spot) {
  selectedSpot.value = spot
  searchQuery.value = ''
  searchResults.value = []
  searchFocused.value = false
  routeActive.value = false
  typing.value = true

  // 清除旧路线
  if (currentRouting) { currentRouting.clear(); currentRouting = null }

  // 高亮地图标记
  highlightMarker(spot.id)

  // 打字机效果
  nextTick(() => {
    typewrite(spot.quote)
    // 分隔线动画
    const divider = document.querySelector('.lpa-divider')
    if (divider) gsap.fromTo(divider, { scaleX: 0 }, { scaleX: 1, duration: 0.5, ease: 'power2.out' })

    // 地图移动到该景点
    if (mapInstance) {
      mapInstance.setCenter(spot.coords, true)
      mapInstance.setZoom(14)
    }
  })
}

function closeAICard() {
  selectedSpot.value = null
  clearMarkerHighlight()
}

function typewrite(text) {
  const el = quoteTextRef.value
  if (!el) return
  el.textContent = ''
  let i = 0
  const timer = setInterval(() => {
    el.textContent += text[i] || ''
    i++
    if (i >= text.length) {
      clearInterval(timer)
      typing.value = false
    }
  }, 45)
}

// =================== ROUTE ===================
function setAsDestination() {
  const spot = selectedSpot.value
  if (!spot) return

  routeInfo.to = spot.name
  routeActive.value = true

  const prev = selectedSpot.value
  selectedSpot.value = null

  nextTick(() => calculateRoute(prev))
}

function calculateRoute(spot) {
  if (!mapInstance || !AMapInstance) return

  // 清除旧路线
  if (currentRouting) { currentRouting.clear(); currentRouting = null }

  const AMap = AMapInstance
  const origin = ORIGIN
  const dest = spot.coords

  const RoutingClass = {
    walk:    AMap.Walking,
    drive:   AMap.Driving,
    transit: AMap.Transfer,
    bike:    AMap.Riding
  }

  const Klass = RoutingClass[transportMode.value]
  if (!Klass) return

  const routing = new Klass({
    map: mapInstance,
    hideMarkers: true,
    panel: null,
    policy: AMap.DrivingPolicy.LEAST_TIME
  })

  currentRouting = routing

  routing.search(origin, dest, (status, result) => {
    if (status === 'complete') {
      const route = result.routes?.[0] || result.plans?.[0]
      if (route) {
        const dist = route.distance > 1000 ? (route.distance / 1000).toFixed(1) + ' km' : route.distance + ' m'
        const time = Math.round((route.time || route.duration) / 60) + ' 分钟'

        routeInfo.distance = dist
        routeInfo.duration = transportMode.value === 'walk' ? `步行约 ${time}` :
                           transportMode.value === 'bike' ? `骑行约 ${time}` :
                           transportMode.value === 'transit' ? `约 ${time}` :
                           `驾车约 ${time}`
      }
      // 缩放至路线范围
      if (result.routes?.[0]) {
        mapInstance.setFitView(null, false, [60, 60, 60, 60])
      }
    } else {
      routeInfo.distance = '—'
      routeInfo.duration = '无法规划路线'
    }
  })
}

function switchMode(mode) {
  if (mode === transportMode.value) return
  transportMode.value = mode
  if (routeActive.value && selectedSpot.value) {
    calculateRoute(selectedSpot.value)
  } else if (routeActive.value) {
    // 查找最近选过的景点重新计算
    const spot = SPOTS.find(s => s.name === routeInfo.to)
    if (spot) calculateRoute(spot)
  }
}

// =================== MAP MARKERS ===================
function highlightMarker(id) {
  markers.forEach(m => {
    const size = m.getContent()?.dataset?.id === id ? 38 : 30
    m.setContent(createMarkerContent(m._spotData, size))
  })
}

function clearMarkerHighlight() {
  markers.forEach(m => {
    m.setContent(createMarkerContent(m._spotData, 30))
  })
}

function createMarkerContent(spot, size = 30) {
  const div = document.createElement('div')
  div.className = 'amap-marker-custom'
  div.dataset.id = spot.id
  div.innerHTML = spot.emoji
  div.style.cssText = `
    width: ${size}px; height: ${size}px;
    background: ${spot.color}22;
    border: 2px solid ${spot.color};
    border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    font-size: ${size * 0.5}px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(.16,1,.3,1);
    box-shadow: 0 2px 8px ${spot.color}33;
  `
  div.onmouseenter = () => {
    div.style.transform = 'scale(1.15)'
    div.style.boxShadow = `0 4px 16px ${spot.color}55`
  }
  div.onmouseleave = () => {
    div.style.transform = ''
    div.style.boxShadow = `0 2px 8px ${spot.color}33`
  }
  return div
}

// =================== ITINERARY ===================
function selectItinerary(itin) {
  if (itin.spots.length > 0) {
    const spot = SPOTS.find(s => s.id === itin.spots[0])
    if (spot) selectSpot(spot)
  }
}

// =================== RESET ===================
function resetAll() {
  selectedSpot.value = null
  routeActive.value = false
  searchQuery.value = ''
  searchResults.value = []
  if (currentRouting) { currentRouting.clear(); currentRouting = null }
  clearMarkerHighlight()
  if (mapInstance) {
    mapInstance.setZoom(12)
    mapInstance.setCenter([104.07, 30.62])
  }
}

// =================== MAP INIT ===================
async function initMap() {
  try {
    mapStatus.value = 'loading'

    if (!mapContainerRef.value) return

    // 安全认证 — 必须在加载前设置
    if (AMAP_SECURITY_CODE) {
      window._AMapSecurityConfig = { securityJsCode: AMAP_SECURITY_CODE }
    }
    const AMap = await AMapLoader.load({
      key: AMAP_KEY,
      version: '2.0',
      plugins: ['AMap.Walking', 'AMap.Driving', 'AMap.Transfer', 'AMap.Riding']
    })

    if (!mapContainerRef.value) return

    mapInstance = new AMap.Map(mapContainerRef.value, {
      center: [104.07, 30.62],
      zoom: 12,
      layers: [
        new AMap.TileLayer.Satellite(),
        new AMap.TileLayer.RoadNet()
      ],
      mapStyle: 'amap://styles/light',
      features: ['bg', 'road', 'building', 'point'],
      showIndoorMap: false,
      expandZoomRange: true,
      zooms: [10, 18],
      resizeEnable: true
    })

    // 添加控件
    mapInstance.addControl(new AMap.Scale())
    mapInstance.addControl(new AMap.MapType({
      defaultType: 0,
      showTraffic: false,
      showRoad: true
    }))

    // 创建标记
    markers = SPOTS.map(spot => {
      const marker = new AMap.Marker({
        position: spot.coords,
        content: createMarkerContent(spot),
        offset: new AMap.Pixel(-15, -15),
        zIndex: 100
      })
      marker._spotData = spot

      marker.on('click', () => selectSpot(spot))
      mapInstance.add(marker)
      return marker
    })

    mapStatus.value = 'ready'
    AMapInstance = AMap
  } catch (e) {
    console.error('AMap init error:', e)
    mapStatus.value = 'error'
  }
}

onMounted(() => {
  initMap()
})

onUnmounted(() => {
  if (mapInstance) {
    mapInstance.destroy()
    mapInstance = null
  }
  AMapInstance = null
  markers = []
  currentRouting = null
})
</script>

<style scoped>
.itinerary-page {
  display: flex;
  width: 100%; height: 100%;
  background: #FCF7F2;
  overflow: hidden;
}

/* ================================
   左侧面板
   ================================ */
.left-panel {
  width: 380px; min-width: 380px;
  height: 100%;
  display: flex; flex-direction: column;
  background: #FFF;
  border-right: 1px solid rgba(0,0,0,.04);
  z-index: 2;
  overflow: hidden;
}

/* 顶部栏 */
.lp-header {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 16px; flex-shrink: 0;
  border-bottom: 1px solid rgba(0,0,0,.03);
}
.lp-back {
  width: 32px; height: 32px; border-radius: 8px;
  border: none; background: rgba(0,0,0,.03); cursor: pointer;
  color: #6B7280; display: flex; align-items: center; justify-content: center;
  transition: all .2s;
}
.lp-back:hover { background: rgba(232,93,58,.06); color: #E85D3A; }
.lp-title { flex: 1; font-size: .88rem; font-weight: 700; color: #2D2D3A; }
.lp-clear {
  padding: 4px 12px; border-radius: 6px;
  border: none; background: rgba(232,93,58,.06); cursor: pointer;
  color: #E85D3A; font-size: .68rem; font-family: inherit;
  transition: all .2s;
}
.lp-clear:hover { background: rgba(232,93,58,.12); }

/* 搜索框 */
.lp-search {
  position: relative;
  display: flex; align-items: center; gap: 6px;
  margin: 12px 16px 8px; padding: 8px 12px;
  border-radius: 10px;
  background: rgba(0,0,0,.03);
  border: 1px solid transparent;
  transition: all .3s;
  flex-shrink: 0;
}
.lp-search.focus {
  border-color: rgba(232,93,58,.2);
  background: #FFF;
  box-shadow: 0 2px 12px rgba(232,93,58,.06);
}
.lps-icon { font-size: .85rem; flex-shrink: 0; }
.lps-input {
  flex: 1; border: none; background: none;
  font-size: .78rem; color: #2D2D3A; font-family: inherit;
  outline: none;
}
.lps-input::placeholder { color: #9CA3AF; }
.lps-clear { border: none; background: none; cursor: pointer; color: #9CA3AF; font-size: .7rem; padding: 2px; }
.lps-suggest {
  position: absolute; top: calc(100% + 4px); left: 0; right: 0;
  border-radius: 10px;
  background: #FFF; border: 1px solid rgba(0,0,0,.04);
  box-shadow: 0 8px 24px rgba(0,0,0,.06);
  overflow: hidden; z-index: 20;
}
.lps-item {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; cursor: pointer; transition: background .15s;
}
.lps-item:hover { background: rgba(232,93,58,.04); }
.lps-emoji { font-size: 1rem; }
.lps-name { font-size: .72rem; font-weight: 600; color: #2D2D3A; }
.lps-tag { font-size: .55rem; color: #9CA3AF; }

/* 出行方式 */
.lp-transport {
  display: flex; gap: 4px;
  margin: 4px 16px 8px; padding: 3px;
  border-radius: 10px;
  background: rgba(0,0,0,.03);
  flex-shrink: 0;
}
.lpt-btn {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 3px;
  padding: 6px 0; border: none; background: none;
  border-radius: 7px; font-size: .65rem; color: #6B7280;
  cursor: pointer; font-family: inherit;
  transition: all .3s;
}
.lpt-btn.active {
  background: #FFF;
  color: #E85D3A; font-weight: 600;
  box-shadow: 0 1px 4px rgba(0,0,0,.06);
}
.lpt-icon { font-size: .85rem; }
.lpt-label { font-size: .6rem; }

/* 路线信息 */
.lp-route {
  margin: 0 16px 10px; padding: 12px 14px;
  border-radius: 10px;
  background: rgba(232,93,58,.03);
  border: 1px solid rgba(232,93,58,.08);
  flex-shrink: 0;
}
.lpr-head { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.lpr-icon { font-size: 1.1rem; }
.lpr-info { display: flex; gap: 8px; align-items: baseline; }
.lpr-dist { font-size: .88rem; font-weight: 700; color: #2D2D3A; }
.lpr-time { font-size: .62rem; color: #6B7280; }
.lpr-path { display: flex; align-items: center; gap: 4px; margin-bottom: 6px; }
.lpr-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.lpr-dot-start { background: #10B981; }
.lpr-dot-end { background: #E85D3A; }
.lpr-line { flex: 1; height: 2px; background: linear-gradient(90deg, #10B981, #E85D3A); border-radius: 1px; }
.lpr-dest { display: flex; gap: 8px; font-size: .65rem; }
.lpr-from { color: #6B7280; }
.lpr-to { color: #E85D3A; font-weight: 600; }

/* 景点列表 */
.lp-spots {
  flex: 1; overflow-y: auto; padding: 4px 16px 8px;
}
.lp-spots::-webkit-scrollbar { width: 3px; }
.lp-spots::-webkit-scrollbar-thumb { background: rgba(0,0,0,.08); border-radius: 2px; }
.lps-label { font-size: .6rem; color: #9CA3AF; letter-spacing: 2px; margin-bottom: 6px; padding: 0 2px; }
.lp-spot {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 10px; margin-bottom: 4px;
  border-radius: 10px; cursor: pointer;
  transition: all .25s;
}
.lp-spot:hover { background: rgba(232,93,58,.03); }
.lp-spot.active {
  background: rgba(232,93,58,.06);
  border: 1px solid rgba(232,93,58,.1);
  margin: -1px 0 3px;
}
.lps-avatar {
  width: 36px; height: 36px; border-radius: 10px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.1rem;
}
.lps-body { flex: 1; min-width: 0; }
.lps-title { font-size: .74rem; font-weight: 600; color: #2D2D3A; margin-bottom: 1px; }
.lps-desc { font-size: .55rem; color: #9CA3AF; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.lps-rating { font-size: .6rem; color: #F5A623; white-space: nowrap; }

/* AI 点评卡片 */
.lp-ai {
  flex-shrink: 0;
  margin: 0 16px 12px; padding: 16px;
  border-radius: 12px;
  background: linear-gradient(135deg, #FFF8F5, #FFF);
  border: 1px solid rgba(232,93,58,.08);
  box-shadow: 0 4px 16px rgba(232,93,58,.04);
}
.lpa-header { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.lpa-avatar {
  width: 38px; height: 38px; border-radius: 10px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.2rem;
}
.lpa-head-text { flex: 1; }
.lpa-name { font-size: .82rem; font-weight: 700; color: #2D2D3A; }
.lpa-tag { font-size: .55rem; color: #9CA3AF; }
.lpa-close {
  width: 24px; height: 24px; border-radius: 6px; border: none;
  background: rgba(0,0,0,.03); cursor: pointer;
  color: #9CA3AF; font-size: .7rem;
  display: flex; align-items: center; justify-content: center;
}
.lpa-close:hover { background: rgba(232,93,58,.08); color: #E85D3A; }

.lpa-quote { margin-bottom: 10px; min-height: 22px; }
.lpa-quote-text {
  font-size: .8rem; font-weight: 500; line-height: 1.5;
  background: linear-gradient(90deg, #E85D3A, #F5A623, #0EA5A0);
  background-size: 300% auto;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: shiftGrad 4s ease infinite;
}
@keyframes shiftGrad {
  0%, 100% { background-position: 0% center; }
  50% { background-position: 100% center; }
}
.lpa-quote-cursor { color: #E85D3A; font-weight: 300; animation: blink .8s step-end infinite; }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }

.lpa-divider {
  height: 1px; margin-bottom: 10px;
  background: linear-gradient(90deg, rgba(232,93,58,.15), rgba(232,93,58,.03));
  transform-origin: left center;
}
.lpa-info { display: flex; gap: 8px; margin-bottom: 10px; }
.lpa-info-item {
  flex: 1; text-align: center; padding: 6px 2px;
  border-radius: 6px; background: rgba(0,0,0,.015);
}
.lpa-info-l { display: block; font-size: .52rem; color: #9CA3AF; margin-bottom: 2px; }
.lpa-info-v { font-size: .68rem; font-weight: 600; color: #2D2D3A; }
.lpa-info-v.star { color: #F5A623; }

.lpa-action {
  width: 100%; padding: 10px;
  border: none; border-radius: 8px;
  background: linear-gradient(135deg, #E85D3A, #F5A623);
  color: #fff; font-size: .75rem; font-weight: 600;
  cursor: pointer; font-family: inherit;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  transition: all .3s;
  box-shadow: 0 4px 12px rgba(232,93,58,.2);
}
.lpa-action:hover { transform: translateY(-1px); box-shadow: 0 6px 20px rgba(232,93,58,.3); }
.lpa-arrow { transition: transform .3s; }
.lpa-action:hover .lpa-arrow { transform: translateX(3px); }

/* 推荐行程 */
.lp-recommend {
  padding: 4px 16px 16px;
  flex-shrink: 0;
}
.lpr-title { font-size: .6rem; color: #9CA3AF; letter-spacing: 2px; margin-bottom: 6px; }
.lpr-cards { display: flex; flex-direction: column; gap: 6px; }
.lpr-card {
  padding: 10px 12px; border-radius: 8px;
  background: rgba(0,0,0,.015);
  border: 1px solid rgba(0,0,0,.03);
  cursor: pointer; transition: all .25s;
}
.lpr-card:hover { background: rgba(232,93,58,.03); border-color: rgba(232,93,58,.08); }
.lprc-top { display: flex; align-items: center; gap: 6px; margin-bottom: 3px; }
.lprc-name { font-size: .7rem; font-weight: 600; color: #2D2D3A; }
.lprc-badge { font-size: .5rem; padding: 1px 6px; border-radius: 3px; background: rgba(232,93,58,.08); color: #E85D3A; }
.lprc-meta { font-size: .55rem; color: #9CA3AF; }

/* 空状态 */
.lp-empty {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  color: #9CA3AF;
}
.lpe-icon { font-size: 2rem; margin-bottom: 8px; opacity: .5; }
.lpe-text { font-size: .7rem; text-align: center; line-height: 1.6; }

/* ================================
   右侧地图面板
   ================================ */
.right-panel {
  flex: 1; position: relative;
  overflow: hidden;
  background: #F0F0F0;
}
.amap-wrapper {
  width: 100%; height: 100%;
}

/* 地图覆盖层 */
.map-overlay {
  position: absolute; inset: 0; z-index: 10;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  background: rgba(252,247,242,.8);
  backdrop-filter: blur(4px);
}
.mo-spinner {
  width: 32px; height: 32px;
  border: 3px solid rgba(232,93,58,.1);
  border-top-color: #E85D3A;
  border-radius: 50%;
  animation: spin .8s linear infinite;
  margin-bottom: 12px;
}
@keyframes spin { to { transform: rotate(360deg); } }
.mo-text { font-size: .78rem; color: #6B7280; }
.mo-icon { font-size: 2.5rem; margin-bottom: 8px; }
.mo-title { font-size: .9rem; font-weight: 700; color: #2D2D3A; margin-bottom: 6px; }
.mo-desc { font-size: .68rem; color: #6B7280; margin-bottom: 2px; }
.mo-desc code { font-family: monospace; color: #E85D3A; background: rgba(232,93,58,.06); padding: 1px 6px; border-radius: 3px; }
.mo-desc a { color: #0EA5A0; }

/* 地图浮标 */
.map-badge {
  position: absolute; top: 16px; left: 50%; transform: translateX(-50%);
  z-index: 10;
  padding: 8px 18px; border-radius: 10px;
  background: rgba(255,255,255,.85);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(232,93,58,.08);
  font-size: .7rem; color: #6B7280;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(0,0,0,.03);
  pointer-events: none;
}

/* ================================
   过渡动画
   ================================ */
.slide-enter-active, .slide-leave-active {
  transition: all .3s cubic-bezier(.16,1,.3,1);
}
.slide-enter-from { opacity: 0; transform: translateY(-8px); max-height: 0; }
.slide-leave-to { opacity: 0; transform: translateY(-8px); max-height: 0; }

.card-enter-active {
  transition: all .4s cubic-bezier(.16,1,.3,1);
}
.card-leave-active {
  transition: all .25s cubic-bezier(.55,0,.84,.2);
}
.card-enter-from { opacity: 0; transform: translateY(16px) scale(.96); }
.card-leave-to { opacity: 0; transform: translateY(10px) scale(.97); }

.drop-enter-active, .drop-leave-active {
  transition: all .2s ease;
}
.drop-enter-from, .drop-leave-to { opacity: 0; transform: translateY(-4px); }

/* ================================
   响应式
   ================================ */
@media (max-width: 768px) {
  .left-panel { width: 100%; min-width: 0; }
  .right-panel { display: none; }
}

/* 自定义高德标记样式覆盖 */
:deep(.amap-marker-custom) {
  transition: all 0.3s cubic-bezier(.16,1,.3,1);
}
:deep(.amap-info-content) {
  border-radius: 8px !important;
  box-shadow: 0 4px 16px rgba(0,0,0,.08) !important;
}
:deep(.amap-info-sharp) { display: none !important; }
</style>
