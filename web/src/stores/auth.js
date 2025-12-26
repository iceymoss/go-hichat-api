import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  // 当前用户信息
  const user = ref(null)
  
  // 登录状态
  const isAuthenticated = ref(false)
  
  // 登录方法
  const login = async (username, password) => {
    // 暂时模拟登录，后续替换为真实API
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        if (username === 'demo' && password === 'demo123') {
          user.value = {
            id: 1001,
            username: 'demo',
            name: 'demo',
            avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=demo'
          }
          isAuthenticated.value = true
          resolve()
        } else {
          reject(new Error('用户名或密码不正确'))
        }
      }, 1000)
    })
  }
  
  // 注册方法
  const register = async (userData) => {
    // 暂时模拟注册
    return new Promise((resolve) => {
      setTimeout(() => {
        console.log('注册信息:', userData)
        resolve()
      }, 1000)
    })
  }
  
  // 登出方法
  const logout = () => {
    user.value = null
    isAuthenticated.value = false
  }
  
  return {
    user,
    isAuthenticated,
    login,
    register,
    logout
  }
})