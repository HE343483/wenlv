import { fileURLToPath } from 'node:url'
import { mergeConfig, defineConfig, configDefaults } from 'vitest/config'
import viteConfig from './vite.config'

export default mergeConfig(
  // vite.config.ts 使用函数式配置(需要按 mode 注入高德安全密钥),这里先解析为对象
  typeof viteConfig === 'function'
    ? viteConfig({ mode: 'test', command: 'serve', isSsrBuild: false })
    : viteConfig,
  defineConfig({
    test: {
      environment: 'jsdom',
      exclude: [...configDefaults.exclude, 'e2e/**'],
      root: fileURLToPath(new URL('./', import.meta.url)),
    },
  }),
)
