<template>
  <div class="modal-overlay" @click.self="close">
    <div class="modal-container">
      <div class="modal-header">
        <h2>好友申请</h2>
        <button class="btn-close" @click="close">
          <i class="icon icon-x"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <!-- Tab切换 -->
        <div class="tabs">
          <button 
            :class="{ active: activeTab === 'received' }" 
            @click="activeTab = 'received'"
          >
            我收到的
            <span v-if="receivedPendingCount > 0" class="badge">{{ receivedPendingCount }}</span>
          </button>
          <button 
            :class="{ active: activeTab === 'sent' }" 
            @click="activeTab = 'sent'"
          >
            我发起的
            <span v-if="sentPendingCount > 0" class="badge">{{ sentPendingCount }}</span>
          </button>
        </div>
        
        <!-- 加载状态 -->
        <div v-if="loading" class="loading">
          <i class="icon icon-spinner"></i>
          <span>加载中...</span>
        </div>
        
        <!-- 我收到的申请列表 -->
        <div v-else-if="activeTab === 'received'" class="requests-content">
          <div v-if="receivedRequests.length === 0" class="empty">
            <i class="icon icon-inbox"></i>
            <p>暂无收到的好友申请</p>
          </div>
          <div v-else class="requests-list">
            <div 
              v-for="req in receivedRequests" 
              :key="req.id" 
              class="request-item"
              :class="{ 
                'pending': req.status === 'pending', 
                'accepted': req.status === 'accepted', 
                'rejected': req.status === 'rejected',
                'ignored': req.status === 'ignored'
              }"
            >
              <img :src="req.avatar" class="avatar" :alt="req.name" @click="viewUserProfile(req)" />
              <div class="info">
                <div class="name-row">
                  <div class="name">{{ req.name }}</div>
                  <span v-if="req.status === 'pending'" class="status-badge pending">{{ req.statusText || '待处理' }}</span>
                  <span v-else-if="req.status === 'accepted'" class="status-badge accepted">{{ req.statusText || '已同意' }}</span>
                  <span v-else-if="req.status === 'rejected'" class="status-badge rejected">{{ req.statusText || '已拒绝' }}</span>
                  <span v-else-if="req.status === 'ignored'" class="status-badge ignored">{{ req.statusText || '已忽略' }}</span>
                </div>
                <!-- 消息气泡显示规则：我收到的申请，status=1（正常显示）且 handle_result=0（待处理）时显示 -->
                <div v-if="Number(req.messageStatus) === 1 && Number(req.handleResult) === 0 && req.message" class="message">{{ req.message }}</div>
                <div class="time">{{ req.time }}</div>
              </div>
              <div v-if="req.status === 'pending'" class="actions">
                <button 
                  class="btn-reject" 
                  @click="handleRequest(req, false)"
                  :disabled="processing"
                >
                  拒绝
                </button>
                <button 
                  class="btn-accept" 
                  @click="handleRequest(req, true)"
                  :disabled="processing"
                >
                  同意
                </button>
              </div>
              <div v-else class="actions">
                <span class="status-text-final">
                  {{ req.statusText || (req.status === 'accepted' ? '已同意' : req.status === 'rejected' ? '已拒绝' : req.status === 'ignored' ? '已忽略' : '待处理') }}
                </span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 我发起的申请列表 -->
        <div v-else class="requests-content sent-requests">
          <div v-if="sentRequests.length === 0" class="empty">
            <i class="icon icon-send"></i>
            <p>暂无发起的好友申请</p>
          </div>
          <div v-else class="requests-list">
            <div 
              v-for="req in sentRequests" 
              :key="req.id" 
              class="request-item"
              :class="{ 
                'pending': req.status === 'pending', 
                'accepted': req.status === 'accepted', 
                'rejected': req.status === 'rejected',
                'ignored': req.status === 'ignored'
              }"
            >
              <img :src="req.avatar" class="avatar" :alt="req.name" @click="viewUserProfile(req)" />
              <div class="info">
                <div class="name-row">
                  <div class="name">{{ req.name }}</div>
                  <!-- 我发起的申请列表不显示状态标签 -->
                </div>
                <!-- 我发起的申请列表不显示消息气泡 -->
                <div class="time">{{ req.time }}</div>
              </div>
              <!-- 我发起的申请列表不显示右侧状态文本 -->
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 用户资料卡片弹窗 -->
    <UserProfileCard 
      v-if="showUserProfileCard"
      :userId="selectedUserId"
      @close="showUserProfileCard = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { userApi } from '../utils/api'
import UserProfileCard from './UserProfileCard.vue'

const emit = defineEmits(['close'])

const contactsStore = useContactsStore()
const activeTab = ref('received') // 'received' | 'sent'
const loading = ref(false)
const processing = ref(false)
const showUserProfileCard = ref(false)
const selectedUserId = ref(null)

// 我收到的申请列表
const receivedRequests = ref([])
// 我发起的申请列表
const sentRequests = ref([])

// 待处理数量统计
// 我收到的申请：需要显示消息气泡的数量（messageStatus === 1 && handleResult === 0）
const receivedPendingCount = computed(() => {
  return receivedRequests.value.filter(req => 
    Number(req.messageStatus) === 1 && Number(req.handleResult) === 0
  ).length
})

// 我发起的申请：待处理的数量（handleResult === 0）
const sentPendingCount = computed(() => {
  return sentRequests.value.filter(req => 
    Number(req.messageStatus) === 1 && Number(req.handleResult) === 0
  ).length
})

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return '未知'
  
  // 后端返回的是秒级时间戳，需要转换为毫秒
  const timeMs = typeof timestamp === 'number' 
    ? (timestamp < 10000000000 ? timestamp * 1000 : timestamp) 
    : new Date(timestamp).getTime()
  
  if (isNaN(timeMs)) return '未知'
  
  const now = Date.now()
  const diff = now - timeMs
  
  if (diff < 0) return '刚刚' // 未来时间
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
  
  const date = new Date(timeMs)
  const month = date.getMonth() + 1
  const day = date.getDate()
  return `${month}月${day}日`
}

// 加载我收到的申请
const loadReceivedRequests = async () => {
  try {
    // 加载所有状态的申请（待处理、已通过、已拒绝）
    const [pending, accepted, rejected] = await Promise.all([
      contactsStore.fetchFriendRequests(0, '1'), // 待处理
      contactsStore.fetchFriendRequests(1, '1'), // 已通过
      contactsStore.fetchFriendRequests(2, '1')  // 已拒绝
    ])
    
    // 合并并获取用户信息
    const allRequests = [...(pending || []), ...(accepted || []), ...(rejected || [])]
    
    // 获取用户信息
    // 我收到的申请：user_id 是发起申请的用户（需要显示的用户）
    const userIds = [...new Set(allRequests.map(req => req.user_id).filter(Boolean))]
    if (userIds.length > 0) {
      const userDetails = await userApi.searchUser({ ids: userIds })
      const userMap = userDetails?.users?.reduce((acc, user) => {
        acc[user.id] = user
        return acc
      }, {}) || {}
      
      receivedRequests.value = allRequests.map(req => {
        const userId = req.user_id // 发起申请的用户ID
        const user = userMap[userId]
        return {
          id: req.id,
          friend_req_id: req.id,
          user_id: req.user_id,
          req_uid: req.req_uid,
          name: user?.nickname || user?.name || userId || '未知用户',
          avatar: user?.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${userId}`,
          message: req.message || req.req_msg || '申请添加你为好友',
          // 直接使用后端返回的 message_status 字段，如果没有则默认为1
          messageStatus: req.message_status !== undefined && req.message_status !== null ? req.message_status : 1, // 消息状态（0:已删除 1:正常显示 2:忽略不显示）
          // 直接使用后端返回的状态字段（优先使用后端返回的字段）
          // store 返回的字段：status, statusText, handleResult, message_status
          status: req.status || (req.handleResult === 0 ? 'pending' : req.handleResult === 1 ? 'accepted' : req.handleResult === 2 ? 'rejected' : req.handleResult === 3 ? 'ignored' : 'pending'),
          statusText: req.statusText || req.status_text || (req.handleResult === 0 ? '待处理' : req.handleResult === 1 ? '已同意' : req.handleResult === 2 ? '已拒绝' : req.handleResult === 3 ? '已忽略' : '待处理'),
          time: formatTime(req.req_time), // 使用原始时间戳（中国时区）
          // 确保 handleResult 正确映射，优先使用 handleResult，其次 handle_result，最后默认为0
          handleResult: (req.handleResult !== undefined && req.handleResult !== null) ? Number(req.handleResult) : ((req.handle_result !== undefined && req.handle_result !== null) ? Number(req.handle_result) : 0) // 处理结果（0:待处理 1:已同意 2:已拒绝 3:已忽略）
        }
      })
    } else {
      receivedRequests.value = []
    }
  } catch (error) {
    console.error('加载我收到的申请失败:', error)
    receivedRequests.value = []
  }
}

// 加载我发起的申请
const loadSentRequests = async () => {
  try {
    // 加载所有状态的申请（待处理、已通过、已拒绝）
    const [pending, accepted, rejected] = await Promise.all([
      contactsStore.fetchFriendRequests(0, '0'), // 待处理
      contactsStore.fetchFriendRequests(1, '0'), // 已通过
      contactsStore.fetchFriendRequests(2, '0')  // 已拒绝
    ])
    
    // 合并并获取用户信息
    const allRequests = [...(pending || []), ...(accepted || []), ...(rejected || [])]
    
    // 获取用户信息（我发起的申请，req_uid是对方，需要显示的用户）
    const userIds = [...new Set(allRequests.map(req => req.req_uid).filter(Boolean))]
    if (userIds.length > 0) {
      const userDetails = await userApi.searchUser({ ids: userIds })
      const userMap = userDetails?.users?.reduce((acc, user) => {
        acc[user.id] = user
        return acc
      }, {}) || {}
      
      sentRequests.value = allRequests.map(req => {
        const userId = req.req_uid // 被申请的用户ID（对方）
        const user = userMap[userId]
        
        // 直接使用 store 返回的数据，确保类型正确
        const handleResult = Number(req.handleResult ?? req.handle_result ?? 0)
        
        return {
          id: req.id,
          friend_req_id: req.id,
          user_id: req.user_id,
          req_uid: req.req_uid,
          name: user?.nickname || user?.name || userId || '未知用户',
          avatar: user?.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${userId}`,
          // 我发起的申请不需要 message 和 messageStatus 字段
          status: req.status || (handleResult === 0 ? 'pending' : handleResult === 1 ? 'accepted' : handleResult === 2 ? 'rejected' : handleResult === 3 ? 'ignored' : 'pending'),
          statusText: req.statusText || req.status_text || (handleResult === 0 ? '待处理' : handleResult === 1 ? '已同意' : handleResult === 2 ? '已拒绝' : handleResult === 3 ? '已忽略' : '待处理'),
          time: formatTime(req.req_time), // 使用原始时间戳（中国时区）
          handleResult: handleResult // 处理结果（0:待处理 1:已同意 2:已拒绝 3:已忽略）
        }
      })
    } else {
      sentRequests.value = []
    }
  } catch (error) {
    console.error('加载我发起的申请失败:', error)
    sentRequests.value = []
  }
}

// 加载数据
const loadRequests = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'received') {
      await loadReceivedRequests()
    } else {
      await loadSentRequests()
    }
  } finally {
    loading.value = false
  }
}

// 处理好友申请
const handleRequest = async (req, accept) => {
  if (processing.value) return
  
  processing.value = true
  try {
    await contactsStore.handleFriendRequest(req.friend_req_id || req.id, accept)
    alert(accept ? '已同意好友申请' : '已拒绝好友申请')
    // 刷新列表
    await loadReceivedRequests()
    // 刷新好友列表
    await contactsStore.fetchFriends()
  } catch (error) {
    console.error('处理好友请求失败:', error)
    alert(error.message || '处理失败，请稍后重试')
  } finally {
    processing.value = false
  }
}

// 查看用户资料
const viewUserProfile = (req) => {
  const userId = activeTab.value === 'received' ? req.user_id : req.req_uid
  if (userId) {
    selectedUserId.value = userId
    showUserProfileCard.value = true
  }
}

// 关闭弹窗
const close = () => {
  emit('close')
}

// 监听tab切换
watch(activeTab, () => {
  loadRequests()
})

// 组件挂载时加载数据
onMounted(() => {
  loadRequests()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-container {
  background: var(--card);
  border-radius: var(--radius-lg);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
  border: 1px solid var(--border);
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  animation: slideUp 0.3s;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border);
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--fg);
  font-family: var(--font-display);
}

.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 20px;
  color: var(--muted);
  cursor: pointer;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-close:hover {
  background-color: var(--card-hover);
  color: var(--fg);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.tabs {
  display: flex;
  border-bottom: 2px solid var(--border);
  padding: 0 24px;
}

.tabs button {
  position: relative;
  padding: 16px 24px;
  border: none;
  background: none;
  font-size: 15px;
  font-weight: 500;
  color: var(--muted);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--font-ui);
}

.tabs button:hover {
  color: var(--accent);
}

.tabs button.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: var(--danger);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}

.requests-content {
  padding: 20px 24px;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--muted);
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
  color: var(--muted);
  gap: 12px;
}

.empty .icon {
  font-size: 48px;
  color: var(--border);
}

.empty p {
  margin: 0;
  font-size: 16px;
  font-family: var(--font-ui);
}

.requests-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.request-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--card);
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-light);
  transition: background-color 0.2s;
}

.request-item:last-child {
  border-bottom: none;
}

.request-item:hover {
  background: var(--card-hover);
}

.request-item.pending {
  background: var(--accent-muted);
}

.request-item.accepted {
  background: var(--card);
}

.request-item.rejected {
  background: var(--card);
}

.request-item.ignored {
  background: var(--card);
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-sm);
  object-fit: cover;
  flex-shrink: 0;
  cursor: pointer;
  transition: opacity 0.2s;
}

.avatar:hover {
  opacity: 0.8;
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
  font-size: 15px;
  font-weight: 500;
  color: var(--fg);
  font-family: var(--font-ui);
}

.status-badge {
  padding: 2px 6px;
  border-radius: 2px;
  font-size: 11px;
  font-weight: 500;
}

.status-badge.pending {
  background: var(--accent);
  color: #fff;
}

.status-badge.accepted {
  background: var(--success);
  color: #fff;
}

.status-badge.rejected {
  background: var(--danger);
  color: #fff;
}

.status-badge.ignored {
  background: var(--muted);
  color: #fff;
}

.message {
  font-size: 13px;
  color: var(--muted);
  margin: 2px 0;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

/* 我发起的申请列表中隐藏所有消息气泡 - 使用更严格的选择器 */
.modal-container .requests-content.sent-requests .message,
.modal-container .sent-requests .requests-list .message,
.modal-container .sent-requests .request-item .message,
.modal-container .sent-requests .info .message,
.modal-container .sent-requests .request-item .info .message,
.sent-requests .message {
  display: none !important;
  visibility: hidden !important;
  height: 0 !important;
  margin: 0 !important;
  padding: 0 !important;
  overflow: hidden !important;
  opacity: 0 !important;
}

.time {
  font-size: 11px;
  color: var(--muted);
  margin-top: 2px;
}

.actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-items: center;
}

.btn-accept,
.btn-reject {
  padding: 6px 20px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  white-space: nowrap;
  min-width: 60px;
  font-family: var(--font-ui);
}

.btn-accept {
  background: var(--accent);
  color: #fff;
}

.btn-accept:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-reject {
  background: var(--border);
  color: var(--fg);
  border: 1px solid var(--border);
}

.btn-reject:hover:not(:disabled) {
  background: var(--card-hover);
  border-color: var(--muted);
}

.btn-accept:disabled,
.btn-reject:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.status-text-final {
  font-size: 14px;
  color: var(--muted);
  padding: 6px 12px;
}

/* 我发起的申请列表中隐藏右侧状态文本 - 使用最严格的选择器 */
.modal-container .sent-requests .request-item .actions,
.modal-container .sent-requests .request-item .status-text-final,
.modal-container .requests-content.sent-requests .actions,
.modal-container .requests-content.sent-requests .status-text-final,
.sent-requests .request-item .actions,
.sent-requests .request-item .status-text-final,
.requests-content.sent-requests .actions,
.requests-content.sent-requests .status-text-final,
.sent-requests .actions,
.sent-requests .status-text-final {
  display: none !important;
  visibility: hidden !important;
  width: 0 !important;
  height: 0 !important;
  margin: 0 !important;
  padding: 0 !important;
  overflow: hidden !important;
  opacity: 0 !important;
  position: absolute !important;
  left: -9999px !important;
}
</style>

