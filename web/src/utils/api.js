import axios from 'axios'

// 创建 axios 实例
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8887',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理错误和 token 过期
api.interceptors.response.use(
  (response) => {
    // 如果响应数据有 code 字段，检查是否成功
    if (response.data && response.data.code !== undefined) {
      if (response.data.code === 0 || response.data.code === 200) {
        return response.data
      } else {
        return Promise.reject(new Error(response.data.msg || response.data.message || '请求失败'))
      }
    }
    return response.data
  },
  (error) => {
    // 处理 HTTP 错误
    if (error.response) {
      const { status, data } = error.response
      
      // 401 未授权，清除 token 并跳转到登录页
      if (status === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        window.location.href = '/login'
        return Promise.reject(new Error('登录已过期，请重新登录'))
      }
      
      // 返回服务器错误信息
      const errorMessage = data?.msg || data?.message || `请求失败 (${status})`
      return Promise.reject(new Error(errorMessage))
    }
    
    // 网络错误
    if (error.request) {
      return Promise.reject(new Error('网络错误，请检查网络连接'))
    }
    
    return Promise.reject(error)
  }
)

// User API
export const userApi = {
  // 登录
  login: (phone, password) => {
    return api.post('/api/v1/user/login', { phone, password })
  },
  
  // 注册
  register: (data) => {
    return api.post('/api/v1/user/register', data)
  },
  
  // 获取用户信息
  getUserInfo: () => {
    return api.get('/api/v1/user/detail')
  },
  
  // 更新用户信息
  updateUser: (data) => {
    return api.put('/api/v1/user/update', data)
  },
  
  // 登出
  logout: () => {
    return api.delete('/api/v1/user/logout')
  },
  
  // 发送邮箱验证码
  sendEmailCode: (email) => {
    return api.post('/api/v1/user/email/code', { email })
  },
  
  // 验证邮箱
  verifyEmail: (email, code) => {
    return api.post('/api/v1/user/email/verify', { email, code })
  },
  
  // 重置密码
  resetPassword: (email, code, password) => {
    return api.put('/api/v1/user/reset_pwd', { email, code, password })
  }
}

export default api

