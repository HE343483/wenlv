import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // 以空前缀加载全部变量,兼容 Docker/CI 直接注入的 VITE_* 环境变量
  const env = loadEnv(mode, process.cwd(), '')
  const amapSecurityJsCode = env.VITE_AMAP_SECURITY_JS_CODE || ''

  return {
    plugins: [
      vue(),
      vueJsx(),
      vueDevTools(),
      {
        // 注入高德地图 JS API 安全密钥(行程规划页地图使用)
        name: 'wenlv-inject-amap-security-code',
        transformIndexHtml(html: string) {
          return html.replaceAll('__AMAP_SECURITY_JS_CODE__', amapSecurityJsCode)
        },
      },
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL || 'http://localhost:8081', // 本地联调(8080 被系统程序占用);可用 VITE_API_BASE_URL 覆盖为 natapp/远程地址
          changeOrigin: true,
          ws: true, // AI 行程规划通过 WebSocket 推送任务进度
        },
      },
    },
  }
})
