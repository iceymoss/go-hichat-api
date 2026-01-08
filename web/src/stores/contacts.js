import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { socialApi } from '../utils/api'

export const useContactsStore = defineStore('contacts', () => {
  // 好友列表
  const friends = ref([])
  
  // 好友请求列表
  const friendRequests = ref([])

  // 群组列表
  const groups = ref([])
  const currentGroup = ref(null)
  
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
          
          // 解析好友标签（从friend_tags数组或friendTags数组）
          let friendTags = []
          if (friend.friend_tags && Array.isArray(friend.friend_tags)) {
            friendTags = friend.friend_tags
          } else if (friend.friendTags && Array.isArray(friend.friendTags)) {
            friendTags = friend.friendTags
          }
          
          return {
            id: friend.id || friend.friend_uid, // 用户ID
            friend_uid: friend.friend_uid || friend.id, // 好友用户ID
            name: friend.nickname || '未设置昵称',
            nickname: friend.nickname,
            avatar: friend.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${friend.friend_uid || friend.id}`,
            // 注意：备注应保持"真实值"，不要用昵称兜底，否则前端无法判断是否真的设置了备注
            remark: friend.remark || '',
            status: 'offline', // 默认离线，后续通过在线状态 API 更新
            tags: tags, // 个人标签数组
            friend_tags: friendTags, // 好友标签数组（关系维度）
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
            type: friend.type, // 用户类型（0-普通用户 1-管理员）
            // 好友设置相关字段
            blacklisted: friend.blacklisted || false,
            moments_permission: friend.moments_permission !== undefined ? friend.moments_permission : (friend.momentsPermission !== undefined ? friend.momentsPermission : 0),
            momentsPermission: friend.moments_permission !== undefined ? friend.moments_permission : (friend.momentsPermission !== undefined ? friend.momentsPermission : 0),
            notify_enabled: friend.notify_enabled !== undefined ? friend.notify_enabled : (friend.notifyEnabled !== undefined ? friend.notifyEnabled : true),
            notifyEnabled: friend.notify_enabled !== undefined ? friend.notify_enabled : (friend.notifyEnabled !== undefined ? friend.notifyEnabled : true),
            pinned: friend.pinned || false,
            muted: friend.muted || false
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
  
  // 更新好友备注（同时更新本地 store）
  const updateFriendRemark = async (friendUid, remark) => {
    try {
      await socialApi.friendUpdateRemark(friendUid, remark)
    } catch (error) {
      console.error('更新好友备注失败:', error)
    } finally {
      patchFriendLocal(friendUid, { remark })
    }
  }

  // 删除好友
  const deleteFriend = async (friendUid) => {
    try {
      await socialApi.friendDelete(friendUid)
    } catch (error) {
      console.error('删除好友失败:', error)
    } finally {
      friends.value = friends.value.filter(friend => (friend.friend_uid || friend.id) !== friendUid)
    }
  }

  // 拉黑 / 取消拉黑好友
  const blockFriend = async (friendUid, blocked = true) => {
    try {
      await socialApi.friendBlock(friendUid, blocked)
    } catch (error) {
      console.error('更新拉黑状态失败:', error)
    } finally {
      patchFriendLocal(friendUid, { blacklisted: blocked })
    }
  }

  // 更新朋友圈权限
  const updateMomentsPermission = async (friendUid, permission) => {
    try {
      await socialApi.friendMomentsPermission(friendUid, permission)
    } catch (error) {
      console.error('更新朋友圈权限失败:', error)
    } finally {
      patchFriendLocal(friendUid, { momentsPermission: permission })
    }
  }

  // 更新好友标签
  const updateFriendTags = async (friendUid, tags) => {
    try {
      await socialApi.friendTags(friendUid, tags)
      patchFriendLocal(friendUid, { tags })
    } catch (error) {
      console.error('更新好友标签失败:', error)
      throw error
    }
  }

  // 更新消息通知开关
  const updateFriendNotification = async (friendUid, enabled) => {
    try {
      await socialApi.friendNotification(friendUid, enabled)
      patchFriendLocal(friendUid, { notify_enabled: enabled, notifyEnabled: enabled })
    } catch (error) {
      console.error('更新消息通知失败:', error)
      throw error
    }
  }

  // 更新置顶开关
  const updateFriendPin = async (friendUid, pinned) => {
    try {
      await socialApi.friendPin(friendUid, pinned)
      patchFriendLocal(friendUid, { pinned })
    } catch (error) {
      console.error('更新置顶状态失败:', error)
      throw error
    }
  }

  // 更新静音开关
  const updateFriendMute = async (friendUid, muted) => {
    try {
      await socialApi.friendMute(friendUid, muted)
      patchFriendLocal(friendUid, { muted })
    } catch (error) {
      console.error('更新静音状态失败:', error)
      throw error
    }
  }

  // 举报好友
  const reportFriend = async (friendUid, reason) => {
    try {
      await socialApi.friendReport(friendUid, reason)
    } catch (error) {
      console.error('举报好友失败:', error)
      throw error
    }
  }

  // 分享好友（预留，若需要可补 targetUid）
  const shareFriend = async (friendUid, targetUid) => {
    try {
      await socialApi.friendShare(friendUid, targetUid)
    } catch (error) {
      console.error('分享好友接口调用失败:', error)
    }
  }

  // 本地更新好友数据
  const patchFriendLocal = (friendUid, payload) => {
    friends.value = friends.value.map(friend => {
      const uid = friend.friend_uid || friend.id
      if (uid === friendUid) {
        return { ...friend, ...payload }
      }
      return friend
    })
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
          read_state: req.read_state !== undefined ? req.read_state : 0, // 读取状态 0未读 1已读
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
  const handleFriendRequest = async (friendReqId, accept, remark, tags) => {
    try {
      // handleResult: 1-同意, 2-拒绝
      const handleResult = accept ? 1 : 2
      await socialApi.friendPutInHandle(friendReqId, handleResult, remark, tags)
      
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
  
  // 获取群组列表
  const fetchGroups = async () => {
    try {
      const response = await socialApi.groupList()
      if (response && response.list) {
        groups.value = response.list.map(g => ({
          ...g,
          avatar: g.icon || `https://api.dicebear.com/7.x/identicon/svg?seed=${g.id}`,
          desc: g.notification || '暂无公告',
          unread: 0 // TODO: fetch unread count
        }))
      } else {
        groups.value = []
      }
    } catch (error) {
      console.error('获取群组列表失败:', error)
      groups.value = []
    }
  }

  // 获取群详情
  const fetchGroupDetail = async (groupId) => {
    try {
      const response = await socialApi.groupDetail(groupId)
      if (response && response.group) {
        currentGroup.value = {
          ...response.group,
          members: response.members || []
        }
        return currentGroup.value
      }
    } catch (error) {
      console.error('获取群详情失败:', error)
      throw error
    }
  }

  // 退出群组
  const quitGroup = async (groupId) => {
    try {
      await socialApi.groupQuit(groupId)
      // Remove from list
      groups.value = groups.value.filter(g => g.id !== groupId)
      if (currentGroup.value && currentGroup.value.id === groupId) {
        currentGroup.value = null
      }
    } catch (error) {
      console.error('退出群组失败:', error)
      throw error
    }
  }

  // 邀请成员
  const inviteGroupMembers = async (groupId, friendIds) => {
    try {
      await socialApi.groupInvite(groupId, friendIds)
      // Refresh group detail to see new members
      if (currentGroup.value && currentGroup.value.id === groupId) {
        await fetchGroupDetail(groupId)
      }
    } catch (error) {
      console.error('邀请成员失败:', error)
      throw error
    }
  }

  // 踢出成员
  const kickGroupMember = async (groupId, memberIds) => {
    try {
      await socialApi.groupKick(groupId, memberIds)
      // Update current group members locally
      if (currentGroup.value && currentGroup.value.id === groupId) {
        currentGroup.value.members = currentGroup.value.members.filter(m => !memberIds.includes(m.user_id))
      }
    } catch (error) {
      console.error('踢出成员失败:', error)
      throw error
    }
  }

  // 更新群信息
  const updateGroupInfo = async (groupId, data) => {
    try {
      await socialApi.groupUpdate(groupId, data)
      // Update local data
      const group = groups.value.find(g => g.id === groupId)
      if (group) {
        Object.assign(group, data)
      }
      if (currentGroup.value && currentGroup.value.id === groupId) {
        Object.assign(currentGroup.value, data)
      }
    } catch (error) {
      console.error('更新群信息失败:', error)
      throw error
    }
  }

  // 创建群组
  const createGroup = async (name, icon) => {
    try {
      await socialApi.createGroup(name, icon)
      // Refresh group list
      await fetchGroups()
    } catch (error) {
      console.error('创建群组失败:', error)
      throw error
    }
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
    groups,
    currentGroup,
    onlineStatus,
    loading,
    requestsLoading,
    searchKeyword,
    groupedFriends,
    searchResults,
    fetchFriends,
    fetchGroups,
    fetchGroupDetail,
    quitGroup,
    inviteGroupMembers,
    kickGroupMember,
    updateGroupInfo,
    createGroup,
    fetchFriendsOnline,
    fetchFriendRequests,
    sendFriendRequest,
    handleFriendRequest,
    addFriendRequest,
    updateFriendRemark,
    deleteFriend,
    blockFriend,
    updateMomentsPermission,
    updateFriendTags,
    updateFriendNotification,
    updateFriendPin,
    updateFriendMute,
    reportFriend,
    shareFriend,
    patchFriendLocal
  }
})
