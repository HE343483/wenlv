import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import App from './App.vue'
import router from './router'
import { tripI18n } from './trip/i18n'
import './assets/main.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
// AI 行程规划模块(移植自 TripStar)使用的组件库与多语言实例
app.use(Antd)
app.use(tripI18n)

app.mount('#app')
