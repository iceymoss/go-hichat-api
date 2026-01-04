import axios from 'axios'

// 创建 axios 实例（用于 user 等服务）
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8887',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 创建 social 服务的 axios 实例
// 开发环境使用相对路径（通过 Vite 代理），生产环境使用完整 URL
const getSocialApiBaseURL = () => {
  if (import.meta.env.VITE_SOCIAL_API_BASE_URL) {
    return import.meta.env.VITE_SOCIAL_API_BASE_URL
  }
  // 开发环境使用相对路径，通过 Vite 代理转发
  if (import.meta.env.DEV) {
    return '' // 使用相对路径，Vite 代理会处理
  }
  // 生产环境默认值
  return 'http://localhost:8889'
}

const socialApiInstance = axios.create({
  baseURL: getSocialApiBaseURL(),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加 token（应用到所有实例）
const requestInterceptor = (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
}

const requestErrorInterceptor = (error) => {
    return Promise.reject(error)
  }

api.interceptors.request.use(requestInterceptor, requestErrorInterceptor)
socialApiInstance.interceptors.request.use(requestInterceptor, requestErrorInterceptor)

// 响应拦截器 - 处理错误和 token 过期（应用到所有实例）
const responseInterceptor = (response) => {
    // 如果响应数据有 code 字段，检查是否成功（go-zero 标准格式）
    if (response.data && response.data.code !== undefined) {
      if (response.data.code === 0 || response.data.code === 200) {
        // 返回 data 字段的内容，而不是整个响应对象
        return response.data.data !== undefined ? response.data.data : response.data
      } else {
        return Promise.reject(new Error(response.data.msg || response.data.message || '请求失败'))
      }
    }
    return response.data
}

const responseErrorInterceptor = (error) => {
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

api.interceptors.response.use(responseInterceptor, responseErrorInterceptor)
socialApiInstance.interceptors.response.use(responseInterceptor, responseErrorInterceptor)

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
  
  // 绑定/更新邮箱
  bindEmail: (email, code) => {
    return api.post('/api/v1/user/email/bind', { email, code })
  },
  
  // 发送手机验证码
  sendPhoneCode: (phone) => {
    return api.post('/api/v1/user/phone/code', { phone })
  },
  
  // 验证手机验证码
  verifyPhoneCode: (phone, code) => {
    return api.post('/api/v1/user/phone/verify', { phone, code })
  },
  
  // 重置密码（支持手机号或邮箱）
  resetPassword: (params) => {
    // params 可以是 { phone, code, password } 或 { email, code, password }
    return api.put('/api/v1/user/reset_pwd', params)
  },
  
  // 上传头像
  uploadAvatar: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/api/v1/user/avatar/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },
  
  // 搜索用户（支持名称模糊匹配，手机号、邮箱和用户ID精准匹配）
  searchUser: (params) => {
    // params 可以是字符串（名称）或对象 {name, phone, email, ids}
    const queryParams = typeof params === 'string' 
      ? { name: params } 
      : params
    return api.get('/api/v1/user/search', { params: queryParams })
  },
  
  // 通过用户ID获取用户信息
  getUserById: (userId) => {
    // 使用搜索 API 的 Ids 参数
    return api.get('/api/v1/user/search', { 
      params: { ids: [userId] } 
    }).then(response => {
      // 返回第一个用户（应该只有一个）
      if (response && response.users && response.users.length > 0) {
        return response.users[0]
      }
      throw new Error('用户不存在')
    })
  }
}

// Social API - 好友相关
export const socialApi = {
  // ========== 好友相关 ==========
  
  // 好友申请
  friendPutIn: (userUid, reqMsg) => {
    return socialApiInstance.post('/v1/social/friend/putIn', {
      user_uid: userUid,
      req_msg: reqMsg || '',
      req_time: Math.floor(Date.now() / 1000)
    })
  },
  
  // 好友申请处理
  friendPutInHandle: (friendReqId, handleResult) => {
    // handleResult: 1-同意, 2-拒绝
    return socialApiInstance.put('/v1/social/friend/putIn', {
      friend_req_id: friendReqId,
      handle_result: handleResult
    })
  },
  
  // 好友申请列表
  friendPutInList: (type, classType) => {
    // type: 1-已通过, 2-已拒绝, 0-待处理
    // class: 0-我发起的申请, 1-我收到的申请
    const params = {}
    if (type !== undefined) params.type = type
    if (classType !== undefined) params.class = classType
    return socialApiInstance.get('/v1/social/friend/putIns', { params })
  },
  
  // 删除好友申请（将 status 设为 0）
  friendPutInDelete: (friendReqId) => {
    return socialApiInstance.post('/v1/social/friend/putIn/delete', { friend_req_id: friendReqId })
  },
  
  // 获取好友申请消息数量
  friendPutInMessageCount: () => {
    return socialApiInstance.get('/v1/social/friend/putIn/messageCount')
  },
  
  // 好友列表
  friendList: () => {
    return socialApiInstance.get('/v1/social/friends')
  },
  
  // 好友在线情况
  friendsOnline: () => {
    return socialApiInstance.get('/v1/social/friends/online')
  },
  
  // ========== 群组相关 ==========
  
  // 创建群组
  createGroup: (name, icon) => {
    return socialApiInstance.post('/v1/social/group', {
      name: name || '',
      icon: icon || ''
    })
  },
  
  // 申请进群
  groupPutIn: (groupId, reqMsg, joinSource, inviterUid) => {
    return socialApiInstance.post('/v1/social/group/putIn', {
      group_id: groupId,
      req_msg: reqMsg || '',
      req_time: Math.floor(Date.now() / 1000),
      join_source: joinSource || 0,
      inviter_uid: inviterUid || '' // 邀请人 UID，如果是主动申请则为空
    })
  },
  
  // 申请进群处理
  groupPutInHandle: (groupReqId, groupId, handleResult) => {
    // handleResult: 1-同意, 2-拒绝
    return socialApiInstance.put('/v1/social/group/putIn', {
      group_req_id: groupReqId,
      group_id: groupId,
      handle_result: handleResult
    })
  },
  
  // 申请进群列表
  groupPutInList: (groupId) => {
    const params = {}
    if (groupId) params.group_id = groupId
    return socialApiInstance.get('/v1/social/group/putIns', { params })
  },
  
  // 用户群列表
  groupList: () => {
    return socialApiInstance.get('/v1/social/groups')
  },
  
  // 群成员列表
  groupUserList: (groupId) => {
    return socialApiInstance.get('/v1/social/group/users', {
      params: { group_id: groupId }
    })
  },
  
  // 群在线用户
  groupUserOnline: (groupId) => {
    return socialApiInstance.get('/v1/social/group/users/online', {
      params: { group_id: groupId }
    })
  }
}

export default api

