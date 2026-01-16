// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

// 更新路由配置
const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { guest: true }
  },
  {
    path: '/app',
    name: 'Main',
    component: () => import('../views/MainView.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'chat',
        name: 'Chat',
        component: () => import('../views/ChatView.vue')
      },
      {
        path: 'contacts',
        name: 'Contacts',
        component: () => import('../views/ContactsView.vue')
      },
      {
        path: 'group-requests',
        name: 'GroupRequests',
        component: () => import('../views/GroupRequestsView.vue')
      },
      {
        path: 'discover',
        name: 'Discover',
        component: () => import('../views/DiscoverView.vue')
      },
      {
        path: 'profile/:userId?',
        name: 'Profile',
        component: () => import('../views/ProfileView.vue'),
        props: (route) => ({ 
          userId: route.params.userId || 'current' 
        })
      },
      {
        path: '',
        redirect: '/app/chat'
      }
    ]
  },
  {
    path: '/app/notifications',
    name: 'Notifications',
    component: () => import('../views/NotificationsView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 添加路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // 如果还没有初始化且有 token，先初始化认证状态
  if (!authStore.isInitialized && authStore.token) {
    try {
      await authStore.init()
    } catch (error) {
      console.error('初始化认证状态失败:', error)
    }
  }
  
  // 等待初始化完成（最多等待1秒，避免无限等待）
  let waitCount = 0
  while (!authStore.isInitialized && authStore.token && waitCount < 10) {
    await new Promise(resolve => setTimeout(resolve, 100))
    waitCount++
  }
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.meta.guest && authStore.isAuthenticated) {
    next('/app')
  } else {
    next()
  }
})

export default router