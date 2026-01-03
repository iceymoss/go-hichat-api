import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import './assets/css/icons.css'

const app = createApp(App)

// 使用状态管理
const pinia = createPinia()
app.use(pinia)

// 使用路由
app.use(router)

// 挂载应用
app.mount('#app')

// 初始化认证状态（在应用挂载后，避免阻塞）
const authStore = useAuthStore()
authStore.init().catch(err => {
  console.error('初始化认证状态失败:', err)
})