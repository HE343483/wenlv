<template>
  <div class="map-page">
    <!-- Map Container -->
    <div ref="mapContainerRef" class="map-container"></div>

    <!-- Top bar -->
    <div class="map-topbar">
      <div class="map-title">
        <span>成都文旅 · 3D 交互地图</span>
        <span class="map-sub">CHENGDU CULTURAL TOURISM · 3D EXPLORER</span>
      </div>
      <div class="map-stats">
        <span>📍 标记景点 <em>{{ SPOTS.length }}</em></span>
      </div>
    </div>

    <!-- Side panel -->
    <div class="side-panel">
      <div class="panel-title">景点列表</div>
      <ul class="spot-list">
        <li v-for="s in SPOTS" :key="s.id" class="spot-item"
          :style="{ '--dot-color': s.color }"
          @click="openSpot3D(s.id)">
          <span class="spot-dot"></span>
          <span class="spot-label" :class="{ active: activeSpotId === s.id }">{{ s.name }}</span>
        </li>
      </ul>
    </div>

    <!-- Legend -->
    <div class="map-legend">
      <div class="legend-item"><span class="l-dot culture"></span>人文古迹</div>
      <div class="legend-item"><span class="l-dot ancient"></span>古蜀文化</div>
      <div class="legend-item"><span class="l-dot nature"></span>自然生态</div>
      <div class="legend-item"><span class="l-dot town"></span>古镇风情</div>
      <div class="legend-item"><span class="l-dot modern"></span>现代文化</div>
    </div>

    <!-- Explore overlay -->
    <div class="explore-overlay" :class="{ hidden: spotsRevealed }" ref="exploreRef">
      <button class="explore-btn" ref="exploreBtnRef" @click="startReveal">
        <span>点亮成都 · 开启探索</span>
        <span class="ex-sub">LIGHT UP CHENGDU <span class="ex-arrow">▼</span></span>
      </button>
    </div>

    <!-- 3D Modal -->
    <Teleport to="body">
      <div class="modal-overlay" :class="{ active: modalActive }" @click.self="closeModal">
        <div class="modal-box">
          <div class="modal-header">
            <div>
              <div class="modal-spotname">{{ currentSpot?.name || '' }}</div>
              <div class="modal-coords" v-if="currentSpot">{{ currentSpot.coords[1] }}°N  {{ currentSpot.coords[0] }}°E</div>
            </div>
            <button class="modal-close" @click="closeModal">✕</button>
          </div>
          <div class="modal-body">
            <div ref="threeContainerRef" class="three-container">
              <div class="modal-loading" v-if="modelLoading">
                <div class="spinner"></div>
                <span>生成 3D 模型中...</span>
              </div>
            </div>
            <div class="modal-info">
              <h3>景点介绍</h3>
              <p>{{ currentSpot?.desc || '' }}</p>
              <div class="modal-tags" v-if="currentSpot">
                <span v-for="t in currentSpot.tags" :key="t" :class="tagClass(t)">{{ t }}</span>
              </div>
              <div class="model-hint">
                💡 当前为程序生成示意模型。<br>
                替换真实 .glb：<br>
                <code>spots[i].model = 'assets/spot.glb'</code>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import * as THREE from 'three'
import { OrbitControls } from 'three/addons/controls/OrbitControls.js'
import * as echarts from 'echarts'

// ===================== DATA =====================
const SPOTS = [
  { id:'panda', name:'大熊猫基地', coords:[104.145,30.733], type:'panda', color:'#10B981',
    tags:['自然生态','国宝','科普教育'],
    desc:'世界著名大熊猫迁地保护基地，占地1000亩。这里生活着近百只大熊猫，是全球最大、最完整的大熊猫繁育研究机构。' },
  { id:'wuhou', name:'武侯祠', coords:[104.047,30.644], type:'temple', color:'#E85D3A',
    tags:['人文古迹','三国文化','全国重点文物'],
    desc:'中国唯一君臣合祀祠庙，纪念诸葛亮、刘备及蜀汉英雄。享有"三国圣地"之美誉，是研究三国历史的重要文化遗存。' },
  { id:'jinli', name:'锦里古街', coords:[104.048,30.645], type:'street', color:'#F5A623',
    tags:['人文古迹','古街文化','美食'],
    desc:'建于秦汉时期的商业古街，早在三国时期就闻名遐迩。目前为成都最具人气的文化旅游街区，浓缩了成都的市井生活。' },
  { id:'ducao', name:'杜甫草堂', coords:[104.031,30.662], type:'garden', color:'#34D399',
    tags:['人文古迹','诗歌文化','全国重点文物'],
    desc:'唐代大诗人杜甫流寓成都时的故居。在此居住近四年，创作诗篇240余首。被誉为"中国文学圣地"，是中国规模最大的杜甫纪念地。' },
  { id:'jinsha', name:'金沙遗址博物馆', coords:[104.012,30.682], type:'artifact', color:'#E8B84B',
    tags:['古蜀文化','世界级考古','太阳神鸟'],
    desc:'距今约3000年的古蜀文化遗址，出土金器、玉器、青铜器等珍贵文物上万件。镇馆之宝"太阳神鸟"金饰是中国文化遗产标志。' },
  { id:'kuanzhai', name:'宽窄巷子', coords:[104.054,30.668], type:'street', color:'#E85D3A',
    tags:['人文古迹','历史文化街区','成都名片'],
    desc:'由宽巷子、窄巷子、井巷子组成的清代古街巷。宽巷子代表休闲生活，窄巷子代表精致生活，是体验成都慢生活的最佳去处。' },
  { id:'dujiangyan', name:'都江堰', coords:[103.607,31.017], type:'water', color:'#0EA5A0',
    tags:['自然生态','古代工程','世界遗产'],
    desc:'李冰父子率众修建于战国时期的大型水利工程。两千多年来持续发挥作用，使成都平原成为"天府之国"。是世界水利文化的鼻祖。' },
  { id:'qingchengshan', name:'青城山', coords:[103.578,31.002], type:'mountain', color:'#34D399',
    tags:['自然生态','道教名山','世界遗产'],
    desc:'中国道教发祥地之一，全真道圣地。群峰环绕如城郭，享有"青城天下幽"之美誉。前山道教宫观集中，后山自然风光清幽。' },
  { id:'wenshu', name:'文殊院', coords:[104.073,30.675], type:'temple', color:'#E85D3A',
    tags:['人文古迹','佛教文化','禅茶'],
    desc:'始建于隋朝，是成都保存最完整的佛教寺院。院内珍藏有唐代玄奘法师顶骨舍利。文殊院茶馆是体验成都盖碗茶文化的名所。' },
  { id:'peoplepark', name:'人民公园', coords:[104.058,30.660], type:'garden', color:'#10B981',
    tags:['人文古迹','市井文化','鹤鸣茶社'],
    desc:'成都最具代表性的市民公园，始建于1911年。园内的鹤鸣茶社是成都最老的传统茶馆之一，是体验"巴适"成都生活的绝佳之地。' },
  { id:'wangjiang', name:'望江楼公园', coords:[104.083,30.637], type:'garden', color:'#34D399',
    tags:['人文古迹','园林','唐代女诗人'],
    desc:'为纪念唐代女诗人薛涛而建。园内望江楼是成都标志性古建筑，锦江边的竹林幽静雅致，是体验成都诗意氛围的绝佳去处。' },
  { id:'xiling', name:'西岭雪山', coords:[103.233,30.683], type:'snow', color:'#E0E7FF',
    tags:['自然生态','雪山风光','杜甫诗意'],
    desc:'因杜甫"窗含西岭千秋雪"诗句得名。海拔5364米，终年积雪。冬季滑雪，夏季避暑，是成都人最近的雪山度假胜地。' },
  { id:'luodai', name:'洛带古镇', coords:[104.283,30.637], type:'town', color:'#F5A623',
    tags:['古镇风情','客家文化','会馆建筑'],
    desc:'中国西部最大的客家移民古镇。广东会馆、江西会馆等清代建筑保存完好。客家文化浓郁，伤心凉粉、客家擂茶等美食不容错过。' },
  { id:'huanglongxi', name:'黄龙溪古镇', coords:[103.967,30.317], type:'town', color:'#F5A623',
    tags:['古镇风情','水乡古镇','天府第一名镇'],
    desc:'建镇1700余年，以古街、古树、古庙、古码头著称。府河与鹿溪河交汇处，被誉为"天府第一名镇"，是成都近郊最具水乡风情之地。' },
  { id:'anren', name:'安仁古镇', coords:[103.617,30.550], type:'town', color:'#F5A623',
    tags:['古镇风情','博物馆小镇','公馆文化'],
    desc:'中国唯一的"博物馆小镇"，现存27座民国时期老公馆。建川博物馆聚落是了解中国近现代史的重要窗口，被誉为"中国博物馆小镇"。' },
  { id:'chengdumuseum', name:'成都博物馆', coords:[104.064,30.656], type:'modern', color:'#8B7EC8',
    tags:['现代文化','综合博物馆','城市历史'],
    desc:'成都最大的综合型博物馆，藏品总数近30万件。从古蜀文明到近现代成都，全面展现成都4500年城市文明发展史。' },
  { id:'yongling', name:'永陵博物馆', coords:[104.041,30.682], type:'tomb', color:'#E8B84B',
    tags:['古蜀文化','前蜀帝王陵','石刻艺术'],
    desc:'前蜀皇帝王建的陵墓，是中国唯一一座地上皇陵。棺床周围的"二十四伎乐"石刻是研究唐代音乐舞蹈的珍贵实物资料。' },
  { id:'naturalmuseum', name:'成都自然博物馆', coords:[104.150,30.683], type:'modern', color:'#0EA5A0',
    tags:['现代文化','自然科学','恐龙化石'],
    desc:'西南地区最大的自然博物馆。合川马门溪龙化石、大竹重庆鱼化石等珍贵标本，带你探索地球生命演化的壮丽历程。' }
]

// ===================== REACTIVE STATE =====================
const mapContainerRef = ref(null)
const threeContainerRef = ref(null)
const exploreRef = ref(null)
const exploreBtnRef = ref(null)

const modalActive = ref(false)
const modelLoading = ref(false)
const activeSpotId = ref(null)
const currentSpot = ref(null)
const spotsRevealed = ref(false)

// Three.js state
let threeInit = false
let threeScene, threeCamera, threeRenderer, threeControls
let currentModel = null
let animId = null
// ECharts state
let myChart = null

// ===================== TAG CLASS =====================
function tagClass(t) {
  if (t.includes('人文') || t.includes('三国') || t.includes('诗歌') || t.includes('佛教') || t.includes('市井')) return 'tag tag-culture'
  if (t.includes('古蜀') || t.includes('世界级考古')) return 'tag tag-ancient'
  if (t.includes('自然') || t.includes('国宝') || t.includes('雪山') || t.includes('古代工程')) return 'tag tag-nature'
  if (t.includes('古镇')) return 'tag tag-town'
  return 'tag tag-modern'
}

// ===================== 3D MODEL FACTORY =====================
function createProceduralModel(spot) {
  const group = new THREE.Group()
  switch(spot.type){
    case 'panda': {
      const pb=new THREE.Mesh(new THREE.SphereGeometry(0.9,12,10),new THREE.MeshPhysicalMaterial({color:0xFFFFFF,roughness:0.8}))
      pb.position.y=0.7;pb.scale.set(1,0.85,1);group.add(pb)
      const ph=new THREE.Mesh(new THREE.SphereGeometry(0.55,10,8),new THREE.MeshPhysicalMaterial({color:0xFFFFFF,roughness:0.8}))
      ph.position.set(0,1.4,0);group.add(ph)
      ;[-0.35,0.35].forEach(x=>{
        const e=new THREE.Mesh(new THREE.SphereGeometry(0.15,8,6),new THREE.MeshPhysicalMaterial({color:0x212121,roughness:0.7}))
        e.position.set(x,1.7,0);group.add(e)
      })
      ;[-0.55,0.55].forEach(x=>{
        const a=new THREE.Mesh(new THREE.SphereGeometry(0.2,8,6),new THREE.MeshPhysicalMaterial({color:0x212121,roughness:0.7}))
        a.position.set(x,0.7,0.5);group.add(a)
      })
      break
    }
    case 'temple': {
      const body=new THREE.Mesh(new THREE.CylinderGeometry(0.9,1.1,1.5,8),new THREE.MeshPhysicalMaterial({color:0x8D6E63,roughness:0.6,metalness:0.1}))
      body.position.y=0.75;group.add(body)
      const roof=new THREE.Mesh(new THREE.ConeGeometry(1.2,0.5,8),new THREE.MeshPhysicalMaterial({color:spot.id==='wenshu'?0x8B6914:0xE85D3A,roughness:0.5}))
      roof.position.y=1.5;group.add(roof)
      const spire=new THREE.Mesh(new THREE.CylinderGeometry(0.06,0.06,0.4,4),new THREE.MeshPhysicalMaterial({color:0xD4AF37,metalness:0.5,roughness:0.3}))
      spire.position.y=1.8;group.add(spire)
      break
    }
    case 'street': {
      for(let i=0;i<6;i++){
        const a=i*Math.PI/3
        const building=new THREE.Mesh(new THREE.BoxGeometry(0.25,0.3+Math.random()*0.4,0.25),
          new THREE.MeshPhysicalMaterial({color:0x8D6E63,roughness:0.7}))
        building.position.set(0.8*Math.cos(a),0.15+Math.random()*0.2,0.8*Math.sin(a))
        group.add(building)
        const roof=new THREE.Mesh(new THREE.BoxGeometry(0.28,0.06,0.28),
          new THREE.MeshPhysicalMaterial({color:0xE85D3A,roughness:0.6}))
        roof.position.set(0.8*Math.cos(a),0.4+Math.random()*0.3,0.8*Math.sin(a))
        group.add(roof)
      }
      const lantern=new THREE.Mesh(new THREE.SphereGeometry(0.12,8,6),
        new THREE.MeshPhysicalMaterial({color:0xE85D3A,emissive:0xE85D3A,emissiveIntensity:0.2}))
      lantern.position.y=0.8;group.add(lantern)
      break
    }
    case 'garden': {
      for(let i=0;i<5;i++){
        const a=Math.random()*Math.PI*2,r=0.3+Math.random()*0.8
        const tree=new THREE.Mesh(new THREE.ConeGeometry(0.15,0.3+Math.random()*0.3,6),
          new THREE.MeshPhysicalMaterial({color:0x34D399,roughness:0.7}))
        tree.position.set(r*Math.cos(a),0.15,r*Math.sin(a));group.add(tree)
      }
      const pavilion=new THREE.Mesh(new THREE.CylinderGeometry(0.3,0.3,0.4,8),
        new THREE.MeshPhysicalMaterial({color:0x8D6E63,roughness:0.6}))
      pavilion.position.y=0.2;group.add(pavilion)
      const proof=new THREE.Mesh(new THREE.ConeGeometry(0.45,0.25,8),
        new THREE.MeshPhysicalMaterial({color:0xE85D3A,roughness:0.5}))
      proof.position.y=0.5;group.add(proof)
      break
    }
    case 'water': {
      const colors=[0x0EA5A0,0x14B8A6,0x0D9488,0x0F766E]
      for(let i=0;i<4;i++){
        const r=0.7+i*0.3
        const t=new THREE.Mesh(new THREE.TorusGeometry(r,0.06,10,24),
          new THREE.MeshPhysicalMaterial({color:colors[i],transparent:true,opacity:0.35+i*0.07,metalness:0.1,roughness:0.2}))
        t.position.y=i*0.18;t.rotation.x=i%2===0?0.08:-0.08;t.rotation.z=i*0.12
        group.add(t)
      }
      break
    }
    case 'mountain': {
      const peak=new THREE.Mesh(new THREE.ConeGeometry(1.4,2.6,6),
        new THREE.MeshPhysicalMaterial({color:0x34D399,roughness:0.6,metalness:0.1,emissive:0x34D399,emissiveIntensity:0.03}))
      peak.position.y=1.3;group.add(peak)
      for(let i=4;i>=0;i--){
        const ring=new THREE.Mesh(new THREE.TorusGeometry(1.0+i*0.15,0.03,4,16),
          new THREE.MeshPhysicalMaterial({color:0x6EE7B7,transparent:true,opacity:0.15-i*0.03}))
        ring.position.y=0.1+i*0.12;ring.rotation.x=Math.PI/2;group.add(ring)
      }
      break
    }
    case 'artifact': {
      const base=new THREE.Mesh(new THREE.CylinderGeometry(0.8,0.9,0.15,8),new THREE.MeshPhysicalMaterial({color:0x546C67,metalness:0.4,roughness:0.3}))
      base.position.y=0.08;group.add(base)
      const ring=new THREE.Mesh(new THREE.TorusGeometry(0.6,0.1,8,20),new THREE.MeshPhysicalMaterial({color:0xE8B84B,metalness:0.6,roughness:0.2,emissive:0xE8B84B,emissiveIntensity:0.1}))
      ring.position.y=0.5;ring.rotation.x=Math.PI/3;group.add(ring)
      const disc=new THREE.Mesh(new THREE.RingGeometry(0.15,0.5,16),new THREE.MeshPhysicalMaterial({color:0xF5A623,metalness:0.7,roughness:0.2,side:THREE.DoubleSide,emissive:0xF5A623,emissiveIntensity:0.05}))
      disc.position.y=0.7;disc.rotation.x=-Math.PI/2;group.add(disc)
      break
    }
    case 'snow': {
      const sp=new THREE.Mesh(new THREE.ConeGeometry(1.6,2.8,6),new THREE.MeshPhysicalMaterial({color:0x6B7280,roughness:0.5}))
      sp.position.y=1.4;group.add(sp)
      const sc=new THREE.Mesh(new THREE.ConeGeometry(0.6,0.5,6),new THREE.MeshPhysicalMaterial({color:0xFFFFFF,roughness:0.2}))
      sc.position.y=2.8;group.add(sc)
      const lake=new THREE.Mesh(new THREE.CylinderGeometry(1.1,1.2,0.06,16),new THREE.MeshPhysicalMaterial({color:0x818CF8,transparent:true,opacity:0.4,metalness:0.2,roughness:0.1}))
      lake.position.y=0.03;group.add(lake)
      break
    }
    case 'town': {
      for(let i=0;i<8;i++){
        const a=Math.random()*Math.PI*2,r=0.2+Math.random()*0.9
        const h=0.2+Math.random()*0.35
        const b=new THREE.Mesh(new THREE.BoxGeometry(0.18,h,0.18),
          new THREE.MeshPhysicalMaterial({color:0x9CA3AF,roughness:0.7}))
        b.position.set(r*Math.cos(a),h/2,r*Math.sin(a));group.add(b)
      }
      break
    }
    case 'tomb': {
      const mound=new THREE.Mesh(new THREE.SphereGeometry(0.8,12,10),new THREE.MeshPhysicalMaterial({color:0x6B7280,roughness:0.7}))
      mound.position.y=0.4;mound.scale.set(1,0.3,1);group.add(mound)
      const stele=new THREE.Mesh(new THREE.BoxGeometry(0.15,0.3,0.05),new THREE.MeshPhysicalMaterial({color:0xE8B84B,metalness:0.3,roughness:0.4}))
      stele.position.y=0.55;group.add(stele)
      break
    }
    case 'modern': {
      const body=new THREE.Mesh(new THREE.BoxGeometry(1.2,1.6,1.2),
        new THREE.MeshPhysicalMaterial({color:spot.id==='naturalmuseum'?0x0EA5A0:0x8B7EC8,metalness:0.2,roughness:0.4}))
      body.position.y=0.8;group.add(body)
      const glass=new THREE.Mesh(new THREE.BoxGeometry(1.0,0.5,0.05),
        new THREE.MeshPhysicalMaterial({color:0x93C5FD,transparent:true,opacity:0.5,metalness:0.8,roughness:0.1}))
      glass.position.set(0,1.2,0.6);group.add(glass)
      break
    }
    default: {
      const fb=new THREE.Mesh(new THREE.BoxGeometry(1.0,1.0,1.0),new THREE.MeshPhysicalMaterial({color:0xE85D3A,emissive:0xE85D3A,emissiveIntensity:0.05,metalness:0.3,roughness:0.4}))
      fb.position.y=0.5;group.add(fb)
    }
  }
  const sd=new THREE.Mesh(new THREE.CircleGeometry(1.6,16),new THREE.MeshBasicMaterial({color:0x000000,transparent:true,opacity:0.08,side:THREE.DoubleSide}))
  sd.rotation.x=-Math.PI/2;sd.position.y=0.001;group.add(sd)
  group.userData.rotateSpeed=0.005
  return group
}

// ===================== OPEN 3D =====================
function openSpot3D(id) {
  const spot = SPOTS.find(s => s.id === id)
  if (!spot) return
  currentSpot.value = spot
  activeSpotId.value = id
  modalActive.value = true
  modelLoading.value = true

  nextTick(() => {
    if (!threeInit) initThreeScene()

    if (currentModel) {
      threeScene.remove(currentModel)
      currentModel.traverse(c => {
        if(c.isMesh){c.geometry?.dispose();if(Array.isArray(c.material))c.material.forEach(m=>m.dispose());else c.material?.dispose()}
      })
      currentModel = null
    }

    setTimeout(() => {
      if (activeSpotId.value !== id) return
      currentModel = createProceduralModel(spot)
      const box = new THREE.Box3().setFromObject(currentModel)
      const size = box.getSize(new THREE.Vector3()).length()
      const scale = size > 0 ? 4.5 / size : 1
      currentModel.scale.set(scale, scale, scale)
      currentModel.position.y = 0.5
      threeScene.add(currentModel)
      modelLoading.value = false

      threeCamera.position.set(4, 2.5, 4)
      threeControls.target.set(0, 0.8, 0)
      threeControls.update()
    }, 300)
  })
}

function closeModal() {
  modalActive.value = false
  activeSpotId.value = null
  if (currentModel) {
    threeScene.remove(currentModel)
    currentModel.traverse(c => {
      if(c.isMesh){c.geometry?.dispose();if(Array.isArray(c.material))c.material.forEach(m=>m.dispose());else c.material?.dispose()}
    })
    currentModel = null
  }
}

function initThreeScene() {
  const container = threeContainerRef.value
  if (!container) return

  threeScene = new THREE.Scene()
  threeScene.background = new THREE.Color(0xF5F0EB)

  threeCamera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 50)
  threeCamera.position.set(4, 2.5, 4)

  threeRenderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' })
  threeRenderer.setSize(container.clientWidth, container.clientHeight)
  threeRenderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  threeRenderer.toneMapping = THREE.ACESFilmicToneMapping
  threeRenderer.toneMappingExposure = 1.2
  container.appendChild(threeRenderer.domElement)

  threeControls = new OrbitControls(threeCamera, threeRenderer.domElement)
  threeControls.enableDamping = true
  threeControls.dampingFactor = 0.08
  threeControls.minDistance = 1.5
  threeControls.maxDistance = 12
  threeControls.target.set(0, 0.8, 0)
  threeControls.autoRotate = true
  threeControls.autoRotateSpeed = 1.5

  const ambient = new THREE.AmbientLight(0xFFE8D6, 0.8); threeScene.add(ambient)
  const key = new THREE.DirectionalLight(0xE85D3A, 0.8); key.position.set(5, 8, 5); threeScene.add(key)
  const fill = new THREE.DirectionalLight(0x0EA5A0, 0.35); fill.position.set(-4, 3, -3); threeScene.add(fill)
  const rim = new THREE.DirectionalLight(0xFFF5ED, 0.5); rim.position.set(-2, 5, -5); threeScene.add(rim)

  const gr = new THREE.Mesh(new THREE.RingGeometry(0.8, 2.2, 32), new THREE.MeshBasicMaterial({ color: 0xE85D3A, transparent: true, opacity: 0.05, side: THREE.DoubleSide }))
  gr.rotation.x = -Math.PI / 2; gr.position.y = -0.01; threeScene.add(gr)

  threeInit = true

  function animate() {
    animId = requestAnimationFrame(animate)
    if (currentModel) currentModel.rotation.y += currentModel.userData?.rotateSpeed || 0.005
    threeControls.update()
    threeRenderer.render(threeScene, threeCamera)
  }
  animate()

  const ro = new ResizeObserver(() => {
    if (!threeRenderer) return
    const w = container.clientWidth, h = container.clientHeight
    threeRenderer.setSize(w, h)
    threeCamera.aspect = w / h
    threeCamera.updateProjectionMatrix()
  })
  ro.observe(container)
}

// ===================== EXPLORE: SEQUENTIAL REVEAL =====================
function startReveal() {
  if (spotsRevealed.value || !myChart) return
  spotsRevealed.value = true

  const btn = exploreBtnRef.value
  if (btn) {
    btn.classList.add('loading')
    btn.innerHTML = '<span>✨ 景点逐一呈现中...</span><span class="ex-sub">REVEALING SCENIC SPOTS</span>'
  }

  const allData = SPOTS.map(s => ({ value: [...s.coords, 1], name: s.name, id: s.id, color: s.color }))
  let idx = 0

  function addNext() {
    if (idx >= allData.length) {
      if (btn) {
        btn.classList.remove('loading'); btn.classList.add('done')
        btn.innerHTML = '<span>✅ 18个景点已全部点亮</span><span class="ex-sub">ALL 18 SPOTS REVEALED</span>'
      }
      if (exploreRef.value) setTimeout(() => exploreRef.value.classList.add('hidden'), 2000)
      return
    }
    const vis = allData.slice(0, idx + 1)
    myChart.setOption({
      series: [{}, { data: vis, symbolSize: 22 }, { data: vis, symbolSize: 12 }]
    }, { notMerge: false })
    idx++
    setTimeout(addNext, 150)
  }
  addNext()
}

// ===================== MAP INIT =====================
onMounted(() => {
  const dom = mapContainerRef.value
  if (!dom) return
  myChart = echarts.init(dom, null, { renderer: 'canvas' })

  const geoUrl = 'https://geo.datav.aliyun.com/areas_v3/bound/510100_full.json'

  fetch(geoUrl)
    .then(r => r.json())
    .then(data => {
      echarts.registerMap('chengdu', data)

      const scatterData = SPOTS.map(s => ({
        value: [...s.coords, 1], name: s.name, id: s.id, color: s.color
      }))

      myChart.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(255,255,255,.95)',
          borderColor: 'rgba(232,93,58,.2)',
          borderWidth: 1,
          textStyle: { color: '#2D2D3A', fontSize: 12 },
          formatter: p => {
            if (p.componentType === 'series') {
              return `<div style="color:#E85D3A;font-weight:700;font-size:14px;margin-bottom:4px">${p.name}</div>
                <div style="color:#6B7280;font-size:11px">经度: ${p.value[0]} · 纬度: ${p.value[1]}</div>
                <div style="color:#9CA3AF;font-size:11px;margin-top:4px">点击查看 3D 景点</div>`
            }
            return ''
          }
        },
        visualMap: { show: false },
        series: [
          {
            type: 'map', map: 'chengdu', roam: true, selectedMode: false,
            zoom: 1.3, center: [104.07, 30.62],
            label: { show: true, color: '#6B7280', fontSize: 9 },
            itemStyle: { areaColor: '#F5F0EB', borderColor: 'rgba(232,93,58,.15)', borderWidth: 1 },
            emphasis: { label: { color: '#E85D3A' }, itemStyle: { areaColor: '#FFEDE5' } }
          },
          {
            type: 'effectScatter', coordinateSystem: 'geo', geoIndex: 0,
            data: [],
            symbolSize: 18,
            rippleEffect: { brushType: 'stroke', scale: 3.5, period: 3 },
            itemStyle: {
              color: p => p.data.color || '#E85D3A',
              shadowBlur: 10, shadowColor: 'rgba(232,93,58,.25)'
            },
            zlevel: 2
          },
          {
            type: 'scatter', coordinateSystem: 'geo', geoIndex: 0,
            data: [],
            symbolSize: 10,
            itemStyle: {
              color: p => p.data.color || '#E85D3A',
              borderColor: 'rgba(255,255,255,.8)', borderWidth: 1.5,
              shadowBlur: 8,
              shadowColor: p => `rgba(${new THREE.Color(p.data.color||'#E85D3A').r*255|0},${new THREE.Color(p.data.color||'#E85D3A').g*255|0},${new THREE.Color(p.data.color||'#E85D3A').b*255|0},.3)`
            },
            label: {
              show: true, formatter: p => p.data.name,
              color: '#4B5563', fontSize: 10, offset: [0, -14]
            },
            emphasis: {
              itemStyle: { borderColor: '#E85D3A', borderWidth: 2, shadowBlur: 20, shadowColor: 'rgba(232,93,58,.4)' },
              label: { color: '#E85D3A', fontWeight: 'bold' }
            },
            zlevel: 3
          }
        ]
      })

      myChart.on('click', p => {
        if (p.componentType === 'series' && p.data?.id) openSpot3D(p.data.id)
      })
    })
    .catch(() => {
      fetch('https://geo.datav.aliyun.com/areas_v3/bound/510100.json')
        .then(r => r.json())
        .then(data => {
          echarts.registerMap('chengdu', data)
          myChart.setOption({
            backgroundColor: 'transparent',
            series: [{ type: 'map', map: 'chengdu', roam: true, zoom: 1.3, center: [104.07, 30.62] }]
          })
        })
    })

  window.addEventListener('resize', () => myChart?.resize())
})

onUnmounted(() => {
  if (animId) cancelAnimationFrame(animId)
  if (currentModel) {
    currentModel.traverse(c => {
      if(c.isMesh){c.geometry?.dispose();if(Array.isArray(c.material))c.material.forEach(m=>m.dispose());else c.material?.dispose()}
    })
  }
  myChart?.dispose()
})
</script>

<style scoped>
.map-page {
  position: relative;
  width: 100%;
  height: 100%;
  background: #FCF7F2;
  overflow: hidden;
}
.map-container { width: 100%; height: 100%; }

/* Top bar */
.map-topbar {
  position: absolute; top: 0; left: 0; right: 0; z-index: 20;
  padding: 14px 24px;
  display: flex; justify-content: space-between; align-items: center;
  background: linear-gradient(180deg,rgba(255,255,255,.9),transparent);
  pointer-events: none;
}
.map-topbar > * { pointer-events: auto; }
.map-title { font-size: 1rem; font-weight: 700; color: #E85D3A; letter-spacing: 1px; }
.map-sub { display: block; font-size: .55rem; color: #9CA3AF; letter-spacing: 3px; margin-top: 1px; font-weight: 400; }
.map-stats { font-size: .7rem; color: #6B7280; }
.map-stats em { color: #E85D3A; font-style: normal; font-weight: 700; }

/* Side panel */
.side-panel {
  position: absolute; right: 14px; top: 64px; z-index: 15;
  width: 170px; padding: 12px 14px;
  background: rgba(255,255,255,.8); border: 1px solid rgba(232,93,58,.08);
  border-radius: 10px; backdrop-filter: blur(12px);
  pointer-events: none;
  box-shadow: 0 4px 12px rgba(0,0,0,.03);
}
.side-panel > * { pointer-events: auto; }
.panel-title { font-size: .6rem; color: #9CA3AF; letter-spacing: 2px; margin-bottom: 8px; text-transform: uppercase; }
.spot-list { list-style: none; }
.spot-item {
  display: flex; align-items: center; gap: 6px;
  padding: 4px 0; font-size: .68rem; color: #6B7280;
  cursor: pointer; transition: color .3s; border-bottom: 1px solid rgba(0,0,0,.02);
}
.spot-item:hover { color: #E85D3A; }
.spot-dot { width: 5px; height: 5px; border-radius: 50%; flex-shrink: 0; background: var(--dot-color); }
.spot-label { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.spot-label.active { color: #E85D3A; font-weight: 600; }

/* Legend */
.map-legend {
  position: absolute; bottom: 60px; left: 14px; z-index: 15;
  padding: 10px 14px; border-radius: 8px;
  background: rgba(255,255,255,.7); border: 1px solid rgba(0,0,0,.04);
  backdrop-filter: blur(6px); font-size: .6rem; color: #6B7280;
  pointer-events: none;
}
.legend-item { display: flex; align-items: center; gap: 6px; margin-bottom: 3px; }
.legend-item:last-child { margin-bottom: 0; }
.l-dot { width: 6px; height: 6px; border-radius: 50%; }
.l-dot.culture { background: #E85D3A; }
.l-dot.ancient { background: #E8B84B; }
.l-dot.nature { background: #10B981; }
.l-dot.town { background: #F5A623; }
.l-dot.modern { background: #8B7EC8; }

/* Explore overlay */
.explore-overlay {
  position: absolute; inset: 0; z-index: 25;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  background: radial-gradient(ellipse at center,rgba(255,255,255,.3),transparent 70%);
  transition: opacity .8s ease,visibility .8s ease;
  pointer-events: none;
}
.explore-overlay.hidden { opacity: 0; visibility: hidden; }
.explore-overlay > * { pointer-events: auto; }
.explore-btn {
  padding: 14px 40px; border: none;
  border-radius: 14px;
  background: linear-gradient(135deg, #E85D3A, #F5A623);
  color: #fff; font-size: 1rem; font-weight: 700;
  cursor: pointer; font-family: inherit;
  backdrop-filter: blur(10px);
  transition: all .4s cubic-bezier(.16,1,.3,1);
  letter-spacing: 3px; text-align: center;
  box-shadow: 0 4px 20px rgba(232,93,58,.2);
}
.explore-btn:hover { transform: translateY(-2px); box-shadow: 0 8px 35px rgba(232,93,58,.3); }
.explore-btn.loading { opacity: .6; pointer-events: none; }
.explore-btn.done { background: linear-gradient(135deg, #10B981, #34D399); box-shadow: 0 4px 20px rgba(16,185,129,.2); }
.ex-sub { display: block; font-size: .55rem; font-weight: 400; letter-spacing: 4px; color: rgba(255,255,255,.6); margin-top: 4px; }
.ex-arrow { display: inline-block; animation: arrowBounce 1.5s ease-in-out infinite; }
@keyframes arrowBounce { 0%,100%{transform:translateY(0);opacity:.5} 50%{transform:translateY(3px);opacity:1} }

/* Modal */
.modal-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(0,0,0,.4); backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center;
  opacity: 0; visibility: hidden;
  transition: all .5s cubic-bezier(.16,1,.3,1);
}
.modal-overlay.active { opacity: 1; visibility: visible; }
.modal-box {
  width: 75vw; height: 80vh; max-width: 1100px;
  border-radius: 18px;
  background: #FFF;
  border: 1px solid rgba(232,93,58,.1);
  box-shadow: 0 40px 100px rgba(0,0,0,.12);
  display: flex; flex-direction: column;
  transform: scale(.92) translateY(20px);
  transition: all .5s cubic-bezier(.16,1,.3,1);
  overflow: hidden;
}
.modal-overlay.active .modal-box { transform: scale(1) translateY(0); }
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px 24px;
  background: linear-gradient(90deg,rgba(232,93,58,.04),transparent);
  border-bottom: 1px solid rgba(232,93,58,.06);
  flex-shrink: 0;
}
.modal-spotname { font-size: 1.05rem; font-weight: 700; color: #E85D3A; }
.modal-coords { font-size: .65rem; color: #9CA3AF; margin-top: 2px; }
.modal-close {
  width: 30px; height: 30px; border-radius: 8px;
  border: 1px solid rgba(0,0,0,.06); background: rgba(0,0,0,.02);
  color: #6B7280; font-size: .9rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all .3s;
}
.modal-close:hover { border-color: #E85D3A; color: #E85D3A; background: rgba(232,93,58,.06); }
.modal-body { flex: 1; display: flex; min-height: 0; }
.three-container { flex: 1; position: relative; background: #F5F0EB; }
.three-container canvas { display: block; }
.modal-info {
  width: 240px; padding: 18px; flex-shrink: 0;
  background: rgba(255,255,255,.6); border-left: 1px solid rgba(232,93,58,.04);
  display: flex; flex-direction: column; gap: 10px; overflow-y: auto;
}
.modal-info h3 { font-size: .8rem; color: #E85D3A; font-weight: 700; padding-bottom: 6px; border-bottom: 1px solid rgba(232,93,58,.06); }
.modal-info p { font-size: .74rem; line-height: 1.8; color: #6B7280; }
.modal-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: .6rem; }
.tag-culture { background: rgba(232,93,58,.08); color: #E85D3A; border: 1px solid rgba(232,93,58,.12); }
.tag-ancient { background: rgba(232,184,75,.1); color: #B8860B; border: 1px solid rgba(232,184,75,.15); }
.tag-nature { background: rgba(16,185,129,.08); color: #10B981; border: 1px solid rgba(16,185,129,.12); }
.tag-town { background: rgba(245,166,35,.08); color: #D97706; border: 1px solid rgba(245,166,35,.12); }
.tag-modern { background: rgba(139,126,200,.08); color: #7C3AED; border: 1px solid rgba(139,126,200,.12); }
.model-hint { margin-top: auto; padding: 8px; border-radius: 6px; background: rgba(0,0,0,.02); border: 1px dashed rgba(0,0,0,.06); font-size: .6rem; color: #9CA3AF; line-height: 1.5; }
.model-hint code { font-size: .55rem; color: #6B7280; }
.modal-loading {
  position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%);
  text-align: center; z-index: 5;
}
.spinner {
  width: 28px; height: 28px; border: 2px solid rgba(232,93,58,.1);
  border-top-color: #E85D3A; border-radius: 50%;
  animation: spin .8s linear infinite; margin: 0 auto 10px;
}
@keyframes spin { to { transform: rotate(360deg); } }
.modal-loading span { font-size: .7rem; color: #9CA3AF; }

@media (max-width: 1200px) {
  .modal-box { width: 85vw; height: 75vh; }
  .modal-info { width: 180px; }
  .side-panel { display: none; }
}
</style>
