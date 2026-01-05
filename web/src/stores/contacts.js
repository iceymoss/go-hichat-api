import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { socialApi } from '../utils/api'

export const useContactsStore = defineStore('contacts', () => {
  // 好友列表
  const friends = ref([])
  
  // 好友请求列表
  const friendRequests = ref([])
  
  // 在线状态映射
  const onlineStatus = ref({})
  
  // 加载状态
  const loading = ref(false)
  const requestsLoading = ref(false)
  
  // 搜索关键词
  const searchKeyword = ref('')
  
  // 分组好友列表（按在线状态）
  const groupedFriends = computed(() => {
    const groups = {
      online: [],
      away: [],
      offline: []
    }
    
    friends.value.forEach(friend => {
      const isOnline = onlineStatus.value[friend.friend_uid || friend.id]
      if (isOnline === true) {
        groups.online.push(friend)
      } else if (isOnline === false) {
        groups.offline.push(friend)
      } else {
        // 如果没有在线状态信息，根据 friend.status 判断
      if (friend.status === 'online') groups.online.push(friend)
      else if (friend.status === 'away') groups.away.push(friend)
      else groups.offline.push(friend)
      }
    })
    
    return groups
  })
  
  // 过滤搜索结果
  const searchResults = computed(() => {
    if (!searchKeyword.value.trim()) return []
    
    const keyword = searchKeyword.value.toLowerCase()
    return friends.value.filter(friend => {
      const name = (friend.nickname || friend.name || '').toLowerCase()
      const remark = (friend.remark || '').toLowerCase()
      return name.includes(keyword) || remark.includes(keyword)
    })
  })
  
  // 获取好友列表
  const fetchFriends = async () => {
    try {
      loading.value = true
      const response = await socialApi.friendList()
      
      if (response && response.list) {
        // 转换数据格式以匹配前端组件期望的格式
        friends.value = response.list.map(friend => {
          // 解析tags（如果是JSON字符串）
          let tags = []
          if (friend.tags) {
            try {
              tags = typeof friend.tags === 'string' ? JSON.parse(friend.tags) : friend.tags
            } catch (e) {
              tags = []
            }
          }
          
          return {
            id: friend.id || friend.friend_uid, // 用户ID
            friend_uid: friend.friend_uid || friend.id, // 好友用户ID
            name: friend.nickname || '未设置昵称',
            nickname: friend.nickname,
            avatar: friend.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${friend.friend_uid || friend.id}`,
            // 注意：备注应保持“真实值”，不要用昵称兜底，否则前端无法判断是否真的设置了备注
            remark: friend.remark || '',
            status: 'offline', // 默认离线，后续通过在线状态 API 更新
            tags: tags, // 个人标签数组
            sex: friend.sex, // 性别（0-未知 1-男 2-女）
            email: friend.email, // 邮箱
            phone: friend.phone, // 手机号
            signature: friend.introduction, // 个性签名
            introduction: friend.introduction, // 个性签名（兼容字段）
            location: friend.region, // 地区
            region: friend.region, // 地区（兼容字段）
            occupation: friend.occupation, // 职业
            gender: friend.sex === 1 ? 'male' : friend.sex === 2 ? 'female' : 'other', // 性别文本
            lastActive: '未知',
            last_login: friend.last_login, // 最后登录时间戳
            status_code: friend.status, // 用户状态（0-禁用 1-正常）
            type: friend.type // 用户类型（0-普通用户 1-管理员）
          }
        })
        
        // 获取在线状态
        await fetchFriendsOnline()
      } else {
        friends.value = []
      }
    } catch (error) {
      console.error('获取好友列表失败:', error)
      friends.value = []
    } finally {
      loading.value = false
    }
  }
  
  // 获取好友在线状态
  const fetchFriendsOnline = async () => {
    try {
      const response = await socialApi.friendsOnline()
      if (response && response.onLineList) {
        onlineStatus.value = response.onLineList
        
        // 更新好友列表中的在线状态
        friends.value = friends.value.map(friend => {
          const uid = friend.friend_uid || friend.id
          const isOnline = response.onLineList[uid]
          return {
            ...friend,
            status: isOnline ? 'online' : 'offline',
            lastActive: isOnline ? '在线' : '离线'
          }
        })
      }
    } catch (error) {
      console.error('获取好友在线状态失败:', error)
    }
  }
  
  // 获取好友申请列表
  const fetchFriendRequests = async (type = 0, classType = '1') => {
    // type: 0-待处理, 1-已通过, 2-已拒绝
    // class: 0-我发起的申请, 1-我收到的申请
    try {
      requestsLoading.value = true
      const response = await socialApi.friendPutInList(type, classType)
      
      if (response && response.list) {
        // 转换数据格式（返回原始数据，时间格式化在组件中处理）
        const formattedList = response.list.map(req => ({
          id: req.id,
          friend_req_id: req.id,
          user_id: req.user_id,
          req_uid: req.req_uid,
          req_msg: req.req_msg,
          req_time: req.req_time, // 保留原始时间戳（中国时区）
          handleResult: req.handle_result !== undefined && req.handle_result !== null ? req.handle_result : 0, // 处理结果，确保有默认值
          handle_result: req.handle_result !== undefined && req.handle_result !== null ? req.handle_result : 0, // 兼容字段
          handle_msg: req.handle_msg,
          // 直接使用后端返回的状态字段（后端已返回 status 和 status_text）
          status: req.status || (req.handle_result === 0 ? 'pending' : req.handle_result === 1 ? 'accepted' : req.handle_result === 2 ? 'rejected' : req.handle_result === 3 ? 'ignored' : 'pending'),
          statusText: req.status_text || (req.handle_result === 0 ? '待处理' : req.handle_result === 1 ? '已同意' : req.handle_result === 2 ? '已拒绝' : req.handle_result === 3 ? '已忽略' : '待处理'),
          status_text: req.status_text || (req.handle_result === 0 ? '待处理' : req.handle_result === 1 ? '已同意' : '已拒绝'), // 兼容字段
          message_status: req.message_status !== undefined && req.message_status !== null ? req.message_status : 1, // 消息状态（0:已删除 1:正常显示 2:忽略不显示）
          // 这些字段会在组件中通过用户信息填充
          name: req.req_uid, // 临时使用，组件中会替换
          avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${req.req_uid}`, // 临时使用
          message: req.req_msg || '申请添加你为好友'
        }))
        
        // 如果是"我收到的申请"，更新 friendRequests
        if (classType === '1') {
          friendRequests.value = formattedList
        }
        
        // 返回格式化后的列表
        return formattedList
      } else {
        if (classType === '1') {
          friendRequests.value = []
        }
        return []
      }
    } catch (error) {
      console.error('获取好友申请列表失败:', error)
      if (classType === '1') {
        friendRequests.value = []
      }
      return []
    } finally {
      requestsLoading.value = false
    }
  }
  
  // 发送好友申请
  const sendFriendRequest = async (userUid, reqMsg = '') => {
    try {
      await socialApi.friendPutIn(userUid, reqMsg)
      return {
          success: true,
          message: '好友请求已发送'
      }
    } catch (error) {
      console.error('发送好友请求失败:', error)
      throw error
    }
  }
  
  // 处理好友申请
  const handleFriendRequest = async (friendReqId, accept) => {
    try {
      // handleResult: 1-同意, 2-拒绝
      const handleResult = accept ? 1 : 2
      await socialApi.friendPutInHandle(friendReqId, handleResult)
      
      // 如果同意，刷新好友列表
      if (accept) {
        await fetchFriends()
      }
      
      // 刷新申请列表
      await fetchFriendRequests(0, '1')
      
      return {
        success: true,
        message: accept ? '已同意好友申请' : '已拒绝好友申请'
      }
    } catch (error) {
      console.error('处理好友申请失败:', error)
      throw error
    }
      }
      
  // 添加好友请求（用于本地添加，通常不需要）
  const addFriendRequest = (request) => {
    friendRequests.value.unshift(request)
  }
  
  // 格式化时间戳
  const formatTime = (timestamp) => {
    if (!timestamp) return '未知'
    
    const now = Date.now()
    const time = timestamp * 1000 // 后端返回的是秒级时间戳
    const diff = now - time
    
    if (diff < 60000) return '刚刚'
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
    if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
    
    const date = new Date(time)
    return `${date.getMonth() + 1}月${date.getDate()}日`
  }
  
  return {
    friends,
    friendRequests,
    onlineStatus,
    loading,
    requestsLoading,
    searchKeyword,
    groupedFriends,
    searchResults,
    fetchFriends,
    fetchFriendsOnline,
    fetchFriendRequests,
    sendFriendRequest,
    handleFriendRequest,
    addFriendRequest
  }
})
