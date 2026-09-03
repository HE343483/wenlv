/** 百度地图 JS API 动态加载工具 */

let loadPromise: Promise<void> | null = null

/**
 * 动态加载百度地图 JS API（v3.0）
 * 幂等：已加载则直接返回；未配置 AK 或加载失败时抛出错误
 */
export function loadBaiduMap(ak: string): Promise<void> {
  const win = window as unknown as { BMap?: unknown }
  if (win.BMap) return Promise.resolve()
  if (loadPromise) return loadPromise

  loadPromise = new Promise((resolve, reject) => {
    const callbackName = `__bmap_init_${Date.now()}`
    const win2 = window as unknown as Record<string, unknown>

    win2[callbackName] = () => {
      delete win2[callbackName]
      resolve()
    }

    const script = document.createElement('script')
    script.src = `https://api.map.baidu.com/api?v=3.0&ak=${encodeURIComponent(ak)}&callback=${callbackName}&s=1`
    script.async = true
    script.onerror = () => {
      delete win2[callbackName]
      loadPromise = null
      reject(new Error('Baidu Map API failed to load'))
    }
    document.head.appendChild(script)
  })

  return loadPromise
}
