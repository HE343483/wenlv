<template>
  <div class="home-page" ref="pageRef">
    <!-- ===== 欢迎区 ===== -->
    <div class="welcome-section" ref="welcomeRef">
      <div class="welcome-deco">
        <svg viewBox="0 0 120 120" class="welcome-svg">
          <circle cx="60" cy="60" r="58" fill="none" stroke="rgba(232,93,58,.06)" stroke-width="2"/>
          <circle cx="60" cy="60" r="44" fill="none" stroke="rgba(232,93,58,.04)" stroke-width="1.5"/>
          <g v-for="i in 12" :key="i">
            <line :x1="60 + 26 * Math.cos(i * Math.PI / 6)" :y1="60 + 26 * Math.sin(i * Math.PI / 6)"
                  :x2="60 + 42 * Math.cos(i * Math.PI / 6)" :y2="60 + 42 * Math.sin(i * Math.PI / 6)"
                  stroke="rgba(232,93,58,.06)" stroke-width="2" stroke-linecap="round"/>
          </g>
        </svg>
      </div>
      <div class="welcome-content">
        <div class="welcome-badge" ref="badgeRef">CHENGDU</div>
        <h2 ref="titleRef">安逸<span class="hl">成都</span></h2>
        <p ref="descRef">今日 {{ weatherDesc }} · 适合出游</p>
      </div>
      <div class="welcome-weather" ref="weatherRef">
        <span class="w-temp">32°</span>
        <span class="w-icon">☀️</span>
      </div>
    </div>

    <!-- ===== 快捷入口（Bento 2+2布局） ===== -->
    <div class="quick-bento" ref="bentoRef">
      <div class="bento-card bento-map" @click="$router.push('/main/map')">
        <div class="bc-icon">🗺️</div>
        <div class="bc-text">
          <div class="bc-title">成都地图</div>
          <div class="bc-sub">探索18个景点</div>
        </div>
      </div>
      <div class="bento-card bento-itin" @click="$router.push('/main/itinerary')">
        <div class="bc-icon">📋</div>
        <div class="bc-text">
          <div class="bc-title">我的行程</div>
          <div class="bc-sub">规划蓉城之旅</div>
        </div>
      </div>
      <div class="bento-card bento-scan" @click="openScanner">
        <div class="bc-icon">📷</div>
        <div class="bc-text">
          <div class="bc-title">扫一扫</div>
          <div class="bc-sub">AR识别导览</div>
        </div>
      </div>
      <div class="bento-card bento-ticket">
        <div class="bc-icon">🎫</div>
        <div class="bc-text">
          <div class="bc-title">票务预订</div>
          <div class="bc-sub">优惠抢先购</div>
        </div>
      </div>
    </div>

    <!-- ===== 今日推荐 ===== -->
    <section class="section" ref="featuredRef">
      <div class="sec-header">
        <h3>🔥 今日推荐</h3>
      </div>
      <div class="featured-card" @click="$router.push('/main/map')">
        <div class="fc-bg">
          <span class="fc-emoji">🐼</span>
        </div>
        <div class="fc-body">
          <div class="fc-title">大熊猫繁育研究基地</div>
          <div class="fc-tags">
            <span class="fc-tag" style="--tag-clr: #10B981">自然生态</span>
            <span class="fc-tag" style="--tag-clr: #E85D3A">热门TOP1</span>
          </div>
          <div class="fc-desc">与国宝近距离接触，看呆萌熊猫嬉戏玩耍</div>
          <div class="fc-meta">
            <span>📍 成华区 · 距市中心30min</span>
            <span class="fc-arrow">→</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 推荐路线 ===== -->
    <section class="section" ref="routesRef">
      <div class="sec-header">
        <h3>🏛️ 推荐路线</h3>
        <span class="sec-more" @click="$router.push('/main/map')">查看全部 →</span>
      </div>
      <div class="route-list">
        <div class="route-card" v-for="(r, i) in routes" :key="r.name"
             :style="'--idx: ' + i" @click="$router.push('/main/map')">
          <div class="rc-stripe" :style="'background: ' + r.color"></div>
          <div class="rc-body">
            <div class="rc-top">
              <span class="rc-icon">{{ r.icon }}</span>
              <div>
                <div class="rc-title">{{ r.name }}</div>
                <div class="rc-meta">{{ r.days }} · {{ r.spots }} 个景点</div>
              </div>
            </div>
            <div class="rc-tags">
              <span v-for="t in r.tags" :key="t">{{ t }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 热门景点 ===== -->
    <section class="section" ref="spotsRef">
      <div class="sec-header">
        <h3>⭐ 热门景点</h3>
        <span class="sec-more" @click="$router.push('/main/map')">查看全部 →</span>
      </div>
      <div class="spot-grid">
        <div class="spot-card" v-for="s in hotSpots" :key="s.name" @click="$router.push('/main/map')">
          <div class="sc-dot" :style="'background: ' + s.color"></div>
          <span class="sc-name">{{ s.name }}</span>
          <span class="sc-count">{{ s.count }}人收藏</span>
        </div>
      </div>
    </section>

    <!-- ===== 底部 ===== -->
    <div class="home-footer" ref="footerRef">
      <div class="footer-brand">蓉城漫游</div>
      <div class="footer-slogan">让世界看见成都</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

const router = useRouter()

const pageRef = ref(null)
const welcomeRef = ref(null)
const badgeRef = ref(null)
const titleRef = ref(null)
const descRef = ref(null)
const weatherRef = ref(null)
const bentoRef = ref(null)
const featuredRef = ref(null)
const routesRef = ref(null)
const spotsRef = ref(null)
const footerRef = ref(null)

const weatherDesc = ref('晴 22°C')

const routes = [
  { name: '蓉城市井文化之旅', days: '2日', spots: 5, icon: '🏛️', color: '#E85D3A',
    tags: ['宽窄巷子','武侯祠','锦里','人民公园'] },
  { name: '古蜀文明探秘之旅', days: '3日', spots: 6, icon: '🌟', color: '#E8B84B',
    tags: ['金沙遗址','杜甫草堂','永陵','成都博物馆'] },
  { name: '自然生态休闲之旅', days: '2日', spots: 4, icon: '🌿', color: '#10B981',
    tags: ['大熊猫基地','青城山','都江堰'] },
  { name: '古镇风情深度之旅', days: '4日', spots: 7, icon: '🏘️', color: '#F5A623',
    tags: ['洛带古镇','黄龙溪','安仁古镇','平乐古镇'] }
]

const hotSpots = [
  { name: '大熊猫基地', color: '#10B981', count: '2.3k' },
  { name: '宽窄巷子', color: '#E85D3A', count: '1.8k' },
  { name: '武侯祠', color: '#E85D3A', count: '1.5k' },
  { name: '锦里', color: '#F5A623', count: '1.4k' },
  { name: '青城山', color: '#34D399', count: '1.2k' },
  { name: '金沙遗址', color: '#E8B84B', count: '980' }
]

function openScanner() { alert('📷 扫码功能已对接景区AR导览') }

onMounted(() => {
  // ===== 入场动画序列 =====
  const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })

  tl.fromTo(badgeRef.value,    { y: -10, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5 }, 0)
  tl.fromTo(titleRef.value,    { y: 20, opacity: 0 },  { y: 0, opacity: 1, duration: 0.6 }, 0.1)
  tl.fromTo(descRef.value,     { y: 12, opacity: 0 },  { y: 0, opacity: 1, duration: 0.4 }, 0.35)
  tl.fromTo(weatherRef.value,  { scale: 0.8, opacity: 0 }, { scale: 1, opacity: 1, duration: 0.5, ease: 'back.out(1.7)' }, 0.5)
  tl.fromTo(bentoRef.value?.children, { y: 20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.4, stagger: 0.06 }, 0.7)

  // ===== Scroll-triggered 动画 =====
  const sections = [featuredRef.value, routesRef.value, spotsRef.value, footerRef.value]
  sections.forEach(el => {
    if (!el) return
    gsap.fromTo(el,
      { y: 30, opacity: 0 },
      {
        y: 0, opacity: 1, duration: 0.6, ease: 'power3.out',
        scrollTrigger: { trigger: el, start: 'top 88%', toggleActions: 'play none none none' }
      }
    )
  })
})
</script>

<style scoped>
.home-page {
  height: 100%; overflow-y: auto; padding: 0 16px 20px;
  background: #FCF7F2;
}
.home-page::-webkit-scrollbar { width: 3px; }
.home-page::-webkit-scrollbar-thumb { background: rgba(232,93,58,.15); border-radius: 2px; }

/* ================================
   欢迎区
   ================================ */
.welcome-section {
  position: relative;
  display: flex; align-items: flex-end; justify-content: space-between;
  padding: 24px 0 18px; margin-bottom: 4px;
  overflow: hidden;
}
.welcome-deco {
  position: absolute; right: 10px; top: -10px;
  width: 100px; height: 100px; pointer-events: none;
}
.welcome-svg { width: 100%; height: 100%; }
.welcome-content { position: relative; z-index: 2; }
.welcome-badge {
  display: inline-block;
  padding: 2px 10px; margin-bottom: 8px;
  border-radius: 4px;
  background: rgba(232,93,58,.06);
  color: rgba(232,93,58,.55);
  font-size: .55rem; letter-spacing: 3px;
}
.welcome-content h2 {
  font-size: 1.6rem; font-weight: 900;
  color: #2D2D3A; line-height: 1.1; margin-bottom: 6px;
}
.welcome-content h2 .hl {
  background: linear-gradient(135deg, #E85D3A, #F5A623);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.welcome-content p {
  font-size: .7rem;
  color: #9CA3AF; letter-spacing: 0.5px;
}
.welcome-weather {
  display: flex; align-items: center; gap: 4px;
  padding: 6px 14px 6px 12px;
  border-radius: 20px;
  background: rgba(255,255,255,.7);
  border: 1px solid rgba(232,93,58,.08);
  backdrop-filter: blur(8px);
  position: relative; z-index: 2;
}
.w-temp { font-size: 1.1rem; font-weight: 700; color: #E85D3A; line-height: 1; }
.w-icon { font-size: 1.1rem; }

/* ================================
   快捷入口 Bento
   ================================ */
.quick-bento {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px; margin-bottom: 20px;
}
.bento-card {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 16px;
  border-radius: 14px;
  background: #FFF;
  border: 1px solid rgba(0,0,0,.04);
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.16,1,0.3,1);
  box-shadow: 0 2px 8px rgba(0,0,0,.03);
}
.bento-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(232,93,58,.08);
}
.bento-map { border-left: 3px solid #E85D3A; }
.bento-itin { border-left: 3px solid #0EA5A0; }
.bento-scan { border-left: 3px solid #E8B84B; }
.bento-ticket { border-left: 3px solid #8B7EC8; }
.bc-icon { font-size: 1.5rem; width: 36px; text-align: center; }
.bc-title { font-size: .78rem; font-weight: 600; color: #2D2D3A; margin-bottom: 2px; }
.bc-sub { font-size: .58rem; color: #9CA3AF; }

/* ================================
   区块通用
   ================================ */
.section { margin-bottom: 20px; }
.sec-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 10px;
}
.sec-header h3 {
  font-size: .82rem; font-weight: 700; color: #2D2D3A;
  letter-spacing: 0.5px;
}
.sec-more {
  font-size: .6rem; color: rgba(232,93,58,.5);
  cursor: pointer; transition: color .3s;
}
.sec-more:hover { color: #E85D3A; }

/* ================================
   今日推荐
   ================================ */
.featured-card {
  display: flex; gap: 14px;
  padding: 14px;
  border-radius: 14px;
  background: linear-gradient(135deg, #FFF, rgba(232,93,58,.02));
  border: 1px solid rgba(232,93,58,.08);
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.16,1,0.3,1);
}
.featured-card:hover {
  border-color: rgba(232,93,58,.2);
  transform: translateY(-1px);
  box-shadow: 0 8px 24px rgba(232,93,58,.06);
}
.fc-bg {
  width: 72px; height: 72px; border-radius: 12px; flex-shrink: 0;
  background: linear-gradient(135deg, #10B981, #34D399);
  display: flex; align-items: center; justify-content: center;
}
.fc-emoji { font-size: 2.2rem; filter: drop-shadow(0 2px 4px rgba(0,0,0,.1)); }
.fc-body { flex: 1; display: flex; flex-direction: column; gap: 5px; }
.fc-title { font-size: .82rem; font-weight: 700; color: #2D2D3A; }
.fc-tags { display: flex; gap: 4px; }
.fc-tag {
  padding: 1px 8px; border-radius: 4px;
  font-size: .55rem;
  color: var(--tag-clr);
  background: color-mix(in srgb, var(--tag-clr) 10%, transparent);
}
.fc-desc { font-size: .65rem; color: #6B7280; line-height: 1.4; }
.fc-meta {
  display: flex; justify-content: space-between; align-items: center;
  font-size: .58rem; color: #9CA3AF; margin-top: 2px;
}
.fc-arrow { color: rgba(232,93,58,.4); font-size: .8rem; }

/* ================================
   推荐路线
   ================================ */
.route-list { display: flex; flex-direction: column; gap: 8px; }
.route-card {
  display: flex; gap: 0;
  border-radius: 12px;
  background: #FFF;
  border: 1px solid rgba(0,0,0,.04);
  cursor: pointer;
  overflow: hidden;
  transition: all 0.35s cubic-bezier(0.16,1,0.3,1);
  box-shadow: 0 2px 8px rgba(0,0,0,.02);
}
.route-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(0,0,0,.04);
}
.rc-stripe {
  width: 4px; flex-shrink: 0;
}
.rc-body {
  flex: 1; padding: 12px 14px;
  display: flex; flex-direction: column; gap: 6px;
}
.rc-top { display: flex; align-items: center; gap: 10px; }
.rc-icon { font-size: 1.3rem; width: 30px; text-align: center; }
.rc-title { font-size: .76rem; font-weight: 600; color: #2D2D3A; margin-bottom: 1px; }
.rc-meta { font-size: .58rem; color: #9CA3AF; }
.rc-tags { display: flex; gap: 4px; flex-wrap: wrap; }
.rc-tags span {
  padding: 1px 7px; border-radius: 4px;
  font-size: .52rem;
  background: rgba(232,93,58,.04);
  color: rgba(232,93,58,.55);
}

/* ================================
   热门景点
   ================================ */
.spot-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.spot-card {
  display: flex; flex-direction: column; align-items: center; gap: 5px;
  padding: 16px 8px 12px;
  border-radius: 12px;
  background: #FFF;
  border: 1px solid rgba(0,0,0,.04);
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.16,1,0.3,1);
  box-shadow: 0 2px 8px rgba(0,0,0,.02);
  position: relative;
  overflow: hidden;
}
.spot-card::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 3px;
  background: var(--sc-clr, #E85D3A);
  opacity: 0.4;
}
.spot-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(232,93,58,.06);
}
.sc-dot {
  width: 10px; height: 10px; border-radius: 50%;
  box-shadow: 0 0 12px var(--dot-glow, rgba(232,93,58,.15));
}
.sc-name { font-size: .7rem; font-weight: 600; color: #2D2D3A; }
.sc-count { font-size: .5rem; color: #9CA3AF; }

/* ================================
   底部
   ================================ */
.home-footer {
  text-align: center; padding: 24px 0 8px;
}
.footer-brand {
  font-size: .7rem; font-weight: 700;
  color: rgba(232,93,58,.3);
  letter-spacing: 3px; margin-bottom: 4px;
}
.footer-slogan {
  font-size: .55rem;
  color: rgba(0,0,0,.1);
  letter-spacing: 4px;
}
</style>
