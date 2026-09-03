/**
 * 百度地图个性化底图样式 (setMapStyleV2 styleJson)
 * 暗色 — 墨韵金：深暖墨底 #0F0D0B + 蜀锦金道路，匹配 dark 主题
 * 亮色 — 宣纸金：暖白纸底 #F4EFE6 + 金箔道路，匹配 light 主题
 * 说明：样式元素较多，按 featureType 分组；值均为 CSS 颜色 + 透明度十六进制
 */

export const darkMapStyle = [
  /* 地表 */
  { featureType: 'land', elementType: 'geometry', stylers: { color: '#151210ff', visibility: 'on' } },
  { featureType: 'manmade', elementType: 'geometry', stylers: { color: '#1a1714ff' } },
  { featureType: 'green', elementType: 'geometry', stylers: { color: '#232a20ff', visibility: 'on' } },
  { featureType: 'water', elementType: 'geometry', stylers: { color: '#1c2529ff', visibility: 'on' } },
  { featureType: 'water', elementType: 'labels', stylers: { visibility: 'off' } },

  /* 建筑物 */
  { featureType: 'building', elementType: 'geometry', stylers: { color: '#201c18ff', visibility: 'on' } },
  { featureType: 'building', elementType: 'geometry.fill', stylers: { color: '#221e19ff' } },
  { featureType: 'building', elementType: 'geometry.stroke', stylers: { color: '#2f2923ff' } },
  { featureType: 'building', elementType: 'labels', stylers: { visibility: 'off' } },

  /* 高速路 — 蜀锦金 */
  { featureType: 'highway', elementType: 'geometry', stylers: { color: '#c9a96eff', visibility: 'on', weight: 4 } },
  { featureType: 'highway', elementType: 'geometry.fill', stylers: { color: '#c9a96eff' } },
  { featureType: 'highway', elementType: 'geometry.stroke', stylers: { color: '#a8883eff' } },
  { featureType: 'highway', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'highway', elementType: 'labels.text.fill', stylers: { color: '#e8d5a8ff' } },
  { featureType: 'highway', elementType: 'labels.text.stroke', stylers: { color: '#0f0d0bff' } },

  /* 主干道 */
  { featureType: 'arterial', elementType: 'geometry', stylers: { color: '#8a7a5cff', visibility: 'on', weight: 2 } },
  { featureType: 'arterial', elementType: 'geometry.fill', stylers: { color: '#6d5f47ff' } },
  { featureType: 'arterial', elementType: 'geometry.stroke', stylers: { color: '#4a4030ff' } },
  { featureType: 'arterial', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'arterial', elementType: 'labels.text.fill', stylers: { color: '#a09786ff' } },
  { featureType: 'arterial', elementType: 'labels.text.stroke', stylers: { color: '#151210ff' } },

  /* 次干道/小路 */
  { featureType: 'local', elementType: 'geometry', stylers: { color: '#3d362eff', visibility: 'on', weight: 1 } },
  { featureType: 'local', elementType: 'geometry.fill', stylers: { color: '#3a332bff' } },
  { featureType: 'local', elementType: 'geometry.stroke', stylers: { color: '#2f2923ff' } },
  { featureType: 'local', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'local', elementType: 'labels.text.fill', stylers: { color: '#756a5cff' } },
  { featureType: 'local', elementType: 'labels.text.stroke', stylers: { color: '#151210ff' } },

  /* 道路 labels 统一弱化 */
  { featureType: 'road', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'road', elementType: 'labels.text.fill', stylers: { color: '#756a5cff' } },

  /* 铁路/地铁 */
  { featureType: 'railway', elementType: 'geometry', stylers: { color: '#4a3a2eff', visibility: 'on' } },
  { featureType: 'railway', elementType: 'geometry.stroke', stylers: { color: '#2f2923ff' } },
  { featureType: 'subway', elementType: 'geometry', stylers: { color: '#5a7a5aff', visibility: 'on' } },
  { featureType: 'subway', elementType: 'geometry.stroke', stylers: { color: '#2f2923ff' } },
  { featureType: 'subway', elementType: 'labels.text.fill', stylers: { color: '#7a8a7aff' } },

  /* 区划名称 */
  { featureType: 'districtlabel', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'districtlabel', elementType: 'labels.text.fill', stylers: { color: '#c9a96eff' } },
  { featureType: 'districtlabel', elementType: 'labels.text.stroke', stylers: { color: '#151210ff' } },

  /* 城市/乡镇名 */
  { featureType: 'city', elementType: 'labels.text.fill', stylers: { color: '#e8d5a8ff' } },
  { featureType: 'city', elementType: 'labels.text.stroke', stylers: { color: '#151210ff' } },
  { featureType: 'town', elementType: 'labels.text.fill', stylers: { color: '#a09786ff' } },
  { featureType: 'town', elementType: 'labels.text.stroke', stylers: { color: '#151210ff' } },
  { featureType: 'continent', elementType: 'labels.text.fill', stylers: { color: '#756a5cff' } },

  /* POI 文字弱化（保留名称，隐藏图标） */
  { featureType: 'poilabel', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'poilabel', elementType: 'labels.text.fill', stylers: { color: '#8a7a5cff' } },
  { featureType: 'poilabel', elementType: 'labels.text.stroke', stylers: { color: '#151210ff' } },
  { featureType: 'poilabel', elementType: 'labels.icon', stylers: { visibility: 'off' } },

  /* 景区点 — 亮金点缀 */
  { featureType: 'scenicspots', elementType: 'geometry', stylers: { color: '#3d5a3dff', visibility: 'on' } },
  { featureType: 'scenicspots', elementType: 'labels.text.fill', stylers: { color: '#c9a96eff' } },
  { featureType: 'entertainment', elementType: 'geometry', stylers: { color: '#2f2923ff' } },
  { featureType: 'shopping', elementType: 'geometry', stylers: { color: '#2f2923ff' } },
  { featureType: 'education', elementType: 'geometry', stylers: { color: '#2f2923ff' } },
]

export const lightMapStyle = [
  /* 地表 */
  { featureType: 'land', elementType: 'geometry', stylers: { color: '#f4efe6ff', visibility: 'on' } },
  { featureType: 'manmade', elementType: 'geometry', stylers: { color: '#ece6daff' } },
  { featureType: 'green', elementType: 'geometry', stylers: { color: '#e3ead9ff', visibility: 'on' } },
  { featureType: 'water', elementType: 'geometry', stylers: { color: '#d3e2e6ff', visibility: 'on' } },
  { featureType: 'water', elementType: 'labels', stylers: { visibility: 'off' } },

  /* 建筑物 */
  { featureType: 'building', elementType: 'geometry', stylers: { color: '#fbf8f2ff', visibility: 'on' } },
  { featureType: 'building', elementType: 'geometry.fill', stylers: { color: '#f6f0e7ff' } },
  { featureType: 'building', elementType: 'geometry.stroke', stylers: { color: '#d7cdbeff' } },
  { featureType: 'building', elementType: 'labels', stylers: { visibility: 'off' } },

  /* 高速路 — 金箔 */
  { featureType: 'highway', elementType: 'geometry', stylers: { color: '#cfb274ff', visibility: 'on', weight: 4 } },
  { featureType: 'highway', elementType: 'geometry.fill', stylers: { color: '#cfb274ff' } },
  { featureType: 'highway', elementType: 'geometry.stroke', stylers: { color: '#ae8b48ff' } },
  { featureType: 'highway', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'highway', elementType: 'labels.text.fill', stylers: { color: '#856a2eff' } },
  { featureType: 'highway', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },

  /* 主干道 */
  { featureType: 'arterial', elementType: 'geometry', stylers: { color: '#d8cdbbff', visibility: 'on', weight: 2 } },
  { featureType: 'arterial', elementType: 'geometry.fill', stylers: { color: '#d7cdbeff' } },
  { featureType: 'arterial', elementType: 'geometry.stroke', stylers: { color: '#e6ded2ff' } },
  { featureType: 'arterial', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'arterial', elementType: 'labels.text.fill', stylers: { color: '#6e6151ff' } },
  { featureType: 'arterial', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },

  /* 次干道/小路 */
  { featureType: 'local', elementType: 'geometry', stylers: { color: '#ddd3c2ff', visibility: 'on', weight: 1 } },
  { featureType: 'local', elementType: 'geometry.fill', stylers: { color: '#e6ded2ff' } },
  { featureType: 'local', elementType: 'geometry.stroke', stylers: { color: '#fbf8f2ff' } },
  { featureType: 'local', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'local', elementType: 'labels.text.fill', stylers: { color: '#9c8e7aff' } },
  { featureType: 'local', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },

  /* 道路 labels */
  { featureType: 'road', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'road', elementType: 'labels.text.fill', stylers: { color: '#9c8e7aff' } },

  /* 铁路/地铁 */
  { featureType: 'railway', elementType: 'geometry', stylers: { color: '#c4b59aff', visibility: 'on' } },
  { featureType: 'railway', elementType: 'geometry.stroke', stylers: { color: '#e6ded2ff' } },
  { featureType: 'subway', elementType: 'geometry', stylers: { color: '#8fa98fff', visibility: 'on' } },
  { featureType: 'subway', elementType: 'geometry.stroke', stylers: { color: '#fbf8f2ff' } },
  { featureType: 'subway', elementType: 'labels.text.fill', stylers: { color: '#6a7a6aff' } },

  /* 区划名称 */
  { featureType: 'districtlabel', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'districtlabel', elementType: 'labels.text.fill', stylers: { color: '#ae8b48ff' } },
  { featureType: 'districtlabel', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },

  /* 城市/乡镇名 */
  { featureType: 'city', elementType: 'labels.text.fill', stylers: { color: '#856a2eff' } },
  { featureType: 'city', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },
  { featureType: 'town', elementType: 'labels.text.fill', stylers: { color: '#6e6151ff' } },
  { featureType: 'town', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },
  { featureType: 'continent', elementType: 'labels.text.fill', stylers: { color: '#9c8e7aff' } },

  /* POI */
  { featureType: 'poilabel', elementType: 'labels', stylers: { visibility: 'on' } },
  { featureType: 'poilabel', elementType: 'labels.text.fill', stylers: { color: '#a08a66ff' } },
  { featureType: 'poilabel', elementType: 'labels.text.stroke', stylers: { color: '#fbf8f2ff' } },
  { featureType: 'poilabel', elementType: 'labels.icon', stylers: { visibility: 'off' } },

  /* 景区点 */
  { featureType: 'scenicspots', elementType: 'geometry', stylers: { color: '#cfe0cfff', visibility: 'on' } },
  { featureType: 'scenicspots', elementType: 'labels.text.fill', stylers: { color: '#6a7a6aff' } },
  { featureType: 'entertainment', elementType: 'geometry', stylers: { color: '#f6f0e7ff' } },
  { featureType: 'shopping', elementType: 'geometry', stylers: { color: '#f6f0e7ff' } },
  { featureType: 'education', elementType: 'geometry', stylers: { color: '#f6f0e7ff' } },
]
