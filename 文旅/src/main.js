// GSAP巴蜀文化品牌应用 - 入口文件
import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'
import './stores/auth.js'

const app = createApp(App)
app.use(router)
app.mount('#app')
