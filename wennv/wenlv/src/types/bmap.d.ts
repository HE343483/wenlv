/** 百度地图 JS API (v3.0) 全局类型声明 — BMap 由 script 动态加载 */

declare global {
  const BMap: any
  const BMAP_NAVIGATION_CONTROL_ZOOM: number
  const BMAP_NAVIGATION_CONTROL_PAN: number
  const BMAP_STATUS_SUCCESS: number
  const BMAP_ANCHOR_TOP_LEFT: number

  interface Window {
    BMap: any
  }
}

export {}
