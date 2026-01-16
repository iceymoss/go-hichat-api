<template>
  <div class="friend-requests-view">
    <div class="header">
    <h2>好友请求</h2>
      <div class="tabs">
        <button 
          :class="{ active: activeTab === 'received' }" 
          @click="activeTab = 'received'"
        >
          我收到的
          <span v-if="receivedCount > 0" class="badge">{{ receivedCount }}</span>
        </button>
        <button 
          :class="{ active: activeTab === 'sent' }" 
          @click="activeTab = 'sent'"
        >
          我发起的
          <span v-if="sentCount > 0" class="badge">{{ sentCount }}</span>
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <i class="icon icon-spinner"></i>
      <span>加载中...</span>
    </div>

    <div v-else-if="requests.length === 0" class="empty">
      <i class="icon icon-inbox"></i>
      <p>暂无好友请求</p>
    </div>

    <div v-else class="requests-list">
      <div 
        v-for="req in requests" 
        :key="req.id" 
        class="request-item"
        :class="{ 'pending': req.status === 'pending', 'accepted': req.status === 'accepted', 'rejected': req.status === 'rejected' }"
      >
        <img :src="req.avatar" class="avatar" :alt="req.name" />
      <div class="info">
          <div class="name-row">
        <div class="name">{{ req.name }}</div>
            <span v-if="req.status === 'pending'" class="status-badge pending">待处理</span>
            <span v-else-if="req.status === 'accepted'" class="status-badge accepted">已同意</span>
            <span v-else-if="req.status === 'rejected'" class="status-badge rejected">已拒绝</span>
          </div>
          <div v-if="activeTab === 'received' && req.message" class="message">{{ req.message }}</div>
        <div class="time">{{ req.time }}</div>
      </div>
        <div v-if="activeTab === 'received' && req.status === 'pending'" class="actions">
          <button 
            class="btn-accept" 
            @click="handleRequest(req, true)"
            :disabled="processing"
          >
            同意
          </button>
          <button 
            class="btn-reject" 
            @click="handleRequest(req, false)"
            :disabled="processing"
          >
            拒绝
          </button>
        </div>
        <div v-else-if="activeTab === 'sent'" class="actions">
          <span class="status-text">
            {{ req.status === 'pending' ? '等待对方处理' : req.status === 'accepted' ? '已同意' : '已拒绝' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { socialApi } from '../utils/api'

const contactsStore = useContactsStore()

const activeTab = ref('received') // 'received' | 'sent'
const loading = ref(false)
const processing = ref(false)

// 获取当前标签页的请求列表
const requests = computed(() => {
  if (activeTab.value === 'received') {
    return contactsStore.friendRequests.filter(req => req.status === 'pending' || req.status === 'accepted' || req.status === 'rejected')
  } else {
    // 我发起的申请需要单独获取
    return sentRequests.value
  }
})

// 我发起的申请列表
const sentRequests = ref([])

// 统计数量
const receivedCount = computed(() => {
  return contactsStore.friendRequests.filter(req => req.status === 'pending').length
})

const sentCount = computed(() => {
  return sentRequests.value.filter(req => req.status === 'pending').length
})

// 加载数据
const loadRequests = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'received') {
      // 加载我收到的申请（待处理、已通过、已拒绝）
      await contactsStore.fetchFriendRequests(0, '1') // 待处理的我收到的申请
      // 可以加载已处理的申请
      // await contactsStore.fetchFriendRequests(1, '1') // 已通过
      // await contactsStore.fetchFriendRequests(2, '1') // 已拒绝
    } else {
      // 加载我发起的申请
      await loadSentRequests()
    }
  } catch (error) {
    console.error('加载好友请求失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载我发起的申请
const loadSentRequests = async () => {
  try {
    // 使用 store 的方法获取数据
    const pendingList = await contactsStore.fetchFriendRequests(0, '0')
    const acceptedList = await contactsStore.fetchFriendRequests(1, '0')
    const rejectedList = await contactsStore.fetchFriendRequests(2, '0')
    
    // 合并所有申请
    sentRequests.value = [
      ...(pendingList || []),
      ...(acceptedList || []),
      ...(rejectedList || [])
    ]
  } catch (error) {
    console.error('加载我发起的申请失败:', error)
    sentRequests.value = []
  }
}

// 处理好友申请
const handleRequest = async (req, accept) => {
  if (processing.value) return
  
  processing.value = true
  try {
    await contactsStore.handleFriendRequest(req.friend_req_id || req.id, accept)
    // 刷新列表
    await loadRequests()
  } catch (error) {
    console.error('处理好友请求失败:', error)
    alert(error.message || '处理失败，请稍后重试')
  } finally {
    processing.value = false
  }
}

// 格式化时间
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

// 监听标签切换
watch(activeTab, () => {
  loadRequests()
})

// 组件挂载时加载数据
onMounted(() => {
  loadRequests()
})
</script>

<style scoped>
.friend-requests-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  background: #fff;
  min-height: 100vh;
}

.header {
  margin-bottom: 24px;
}

.header h2 {
  margin: 0 0 16px 0;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.tabs {
  display: flex;
  gap: 12px;
  border-bottom: 2px solid #f0f2f5;
}

.tabs button {
  position: relative;
  padding: 12px 20px;
  border: none;
  background: none;
  font-size: 15px;
  font-weight: 500;
  color: #666;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tabs button:hover {
  color: #4a8cff;
}

.tabs button.active {
  color: #4a8cff;
  border-bottom-color: #4a8cff;
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: #dc3545;
  color: white;
  font-size: 12px;
  font-weight: 600;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #999;
  gap: 12px;
}

.loading .icon {
  font-size: 32px;
  animation: spin 1s infinite linear;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #999;
  gap: 12px;
}

.empty .icon {
  font-size: 48px;
  color: #ccc;
}

.empty p {
  margin: 0;
  font-size: 16px;
}

.requests-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.request-item {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #f8fafc;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #e9ecef;
  transition: all 0.2s;
}

.request-item:hover {
  background: #f0f2f5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.request-item.pending {
  border-left: 4px solid #ffc107;
}

.request-item.accepted {
  border-left: 4px solid #28a745;
}

.request-item.rejected {
  border-left: 4px solid #dc3545;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  object-fit: cover;
  flex-shrink: 0;
}

.info {
  flex: 1;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.name {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.pending {
  background: #fff3cd;
  color: #856404;
}

.status-badge.accepted {
  background: #d4edda;
  color: #155724;
}

.status-badge.rejected {
  background: #f8d7da;
  color: #721c24;
}

.message {
  font-size: 14px;
  color: #666;
  margin: 4px 0;
  line-height: 1.5;
}

.time {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}

.btn-accept,
.btn-reject {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-accept {
  background: #28a745;
  color: white;
}

.btn-accept:hover:not(:disabled) {
  background: #218838;
}

.btn-reject {
  background: #f8f9fa;
  color: #6c757d;
  border: 1px solid #dee2e6;
}

.btn-reject:hover:not(:disabled) {
  background: #e9ecef;
}

.btn-accept:disabled,
.btn-reject:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.status-text {
  font-size: 14px;
  color: #666;
  padding: 8px 16px;
}
</style> 
