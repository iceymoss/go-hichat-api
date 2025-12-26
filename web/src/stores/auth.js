import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userApi } from '../utils/api'

export const useAuthStore = defineStore('auth', () => {
  // 当前用户信息
  const user = ref(null)
  
  // Token 信息
  const token = ref(localStorage.getItem('token') || null)
  const tokenExpire = ref(localStorage.getItem('tokenExpire') || null)
  
  // 登录状态
  const isAuthenticated = computed(() => {
    if (!token.value) return false
    // 检查 token 是否过期
    if (tokenExpire.value) {
      const expireTime = parseInt(tokenExpire.value)
      if (expireTime && Date.now() > expireTime) {
        // Token 已过期，清除
        clearAuth()
        return false
      }
    }
    return !!token.value && !!user.value
  })
  
  // 清除认证信息
  const clearAuth = () => {
    token.value = null
    tokenExpire.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('tokenExpire')
    localStorage.removeItem('user')
  }
  
  // 登录方法
  const login = async (phone, password) => {
    try {
      const response = await userApi.login(phone, password)
      
      // 保存 token
      token.value = response.token
      tokenExpire.value = response.expire ? Date.now() + response.expire * 1000 : null
      
      localStorage.setItem('token', response.token)
      if (tokenExpire.value) {
        localStorage.setItem('tokenExpire', tokenExpire.value.toString())
      }
      
      // 获取用户信息
      await fetchUserInfo()
      
      return response
    } catch (error) {
      throw error
    }
  }
  
  // 注册方法
  const register = async (userData) => {
    try {
      const response = await userApi.register({
        phone: userData.phone,
        password: userData.password,
        nickname: userData.nickname || userData.phone,
        email: userData.email || '',
        avatar: userData.avatar || '',
        sex: userData.sex || 0
      })
      
      // 注册成功后自动登录
      if (response.token) {
        token.value = response.token
        tokenExpire.value = response.expire ? Date.now() + response.expire * 1000 : null
        
        localStorage.setItem('token', response.token)
        if (tokenExpire.value) {
          localStorage.setItem('tokenExpire', tokenExpire.value.toString())
        }
        
        // 获取用户信息
        await fetchUserInfo()
      }
      
      return response
    } catch (error) {
      throw error
    }
  }
  
  // 获取用户信息
  const fetchUserInfo = async () => {
    try {
      const response = await userApi.getUserInfo()
      if (response && response.info) {
        user.value = {
          id: response.info.id,
          phone: response.info.mobile,
          nickname: response.info.nickname,
          name: response.info.nickname,
          avatar: response.info.avatar || 'https://api.dicebear.com/7.x/personas/svg?seed=' + response.info.id,
          sex: response.info.sex,
          email: response.info.email,
          introduction: response.info.introduction,
          lastLogin: response.info.lastLogin
        }
        localStorage.setItem('user', JSON.stringify(user.value))
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
      // 如果获取用户信息失败，可能是 token 无效，清除认证信息
      if (error.message && error.message.includes('401')) {
        clearAuth()
      }
    }
  }
  
  // 更新用户信息
  const updateUser = async (userData) => {
    try {
      await userApi.updateUser(userData)
      // 更新成功后重新获取用户信息
      await fetchUserInfo()
    } catch (error) {
      throw error
    }
  }
  
  // 登出方法
  const logout = async () => {
    try {
      // 调用后端登出接口
      await userApi.logout()
    } catch (error) {
      console.error('登出失败:', error)
    } finally {
      // 无论成功与否都清除本地认证信息
      clearAuth()
    }
  }
  
  // 初始化：如果本地有 token，尝试获取用户信息
  const init = async () => {
    if (token.value && !user.value) {
      await fetchUserInfo()
    } else if (token.value && user.value) {
      // 从 localStorage 恢复用户信息
      const savedUser = localStorage.getItem('user')
      if (savedUser) {
        try {
          user.value = JSON.parse(savedUser)
        } catch (e) {
          // 如果解析失败，重新获取
          await fetchUserInfo()
        }
      }
    }
  }
  
  return {
    user,
    token,
    tokenExpire,
    isAuthenticated,
    login,
    register,
    logout,
    updateUser,
    fetchUserInfo,
    init
  }
})