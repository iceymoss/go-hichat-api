<template>
  <div class="modal-overlay" @click.self="close">
    <div class="modal-container user-profile-card">
      <div class="modal-header">
        <h2>用户资料</h2>
        <button class="btn-close" @click="close">
          <i class="icon icon-x"></i>
        </button>
      </div>
      
      <div class="modal-body" v-if="loading">
        <div class="loading-state">
          <i class="icon icon-spinner"></i>
          <span>加载中...</span>
        </div>
      </div>
      
      <div class="modal-body" v-else-if="user">
        <!-- 用户头像和基本信息 -->
        <div class="user-header">
          <div class="avatar-container">
            <img :src="user.avatar" alt="头像" class="avatar" />
          </div>
          <div class="user-info">
            <h3 class="name">{{ user.name }}</h3>
            <div class="account">ID: {{ user.account }}</div>
            <div class="signature" v-if="user.signature">{{ user.signature }}</div>
            <div v-else class="signature placeholder">这个人很懒，什么都没写</div>
          </div>
        </div>
        
        <!-- 用户详细信息 -->
        <div class="user-details">
          <div class="detail-section">
            <h4>基础资料</h4>
            <div class="detail-item">
              <span class="label">性别</span>
              <span class="value">{{ user.gender === 'male' ? '男' : user.gender === 'female' ? '女' : '未知' }}</span>
            </div>
            <div class="detail-item" v-if="user.region">
              <span class="label">地区</span>
              <span class="value">{{ user.region }}</span>
            </div>
            <div class="detail-item" v-if="user.occupation">
              <span class="label">职业</span>
              <span class="value">{{ user.occupation }}</span>
            </div>
          </div>
          
          <div class="detail-section" v-if="user.tags && user.tags.length > 0">
            <h4>个人标签</h4>
            <div class="tags">
              <span v-for="tag in user.tags" :key="tag" class="tag">{{ tag }}</span>
            </div>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="user-actions">
          <button 
            v-if="user.isFriend"
            class="btn-chat" 
            @click="startChat"
          >
            <i class="icon icon-chat"></i>
            发消息
          </button>
          <!-- 如果有待处理的好友申请（我收到的），显示接受申请按钮 -->
          <button 
            v-else-if="pendingFriendRequest"
            class="btn-accept-request" 
            @click="handleAcceptRequest"
          >
            <i class="icon icon-check"></i>
            接受申请
          </button>
          <button 
            v-else
            class="btn-add-friend" 
            @click="showAddFriendDialog = true"
            :disabled="sendingFriendRequest"
          >
            <i v-if="sendingFriendRequest" class="icon icon-spinner"></i>
            <i v-else class="icon icon-add"></i>
            {{ sendingFriendRequest ? '发送中...' : '添加好友' }}
          </button>
        </div>
      </div>
      
      <div class="modal-body" v-else>
        <div class="error-state">
          <i class="icon icon-error"></i>
          <p>用户不存在或无法访问</p>
        </div>
      </div>
    </div>
    
    <!-- 申请好友对话框 -->
    <div v-if="showAddFriendDialog" class="modal-overlay add-friend-overlay" @click.self="showAddFriendDialog = false">
      <div class="modal-container add-friend-dialog">
        <div class="modal-header">
          <h3>发送好友申请</h3>
          <button type="button" class="modal-close" @click="showAddFriendDialog = false">×</button>
        </div>
        
        <div class="modal-body">
          <div class="user-preview">
            <img :src="user?.avatar" alt="头像" class="avatar" />
            <div class="user-name">{{ user?.name }}</div>
          </div>
          <div class="message-input">
            <label>申请消息（可选）</label>
            <textarea 
              v-model="requestMessage" 
              placeholder="请输入申请消息..."
              rows="3"
              maxlength="100"
            ></textarea>
            <div class="char-count">{{ requestMessage.length }}/100</div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button type="button" class="btn-cancel-modal" @click="showAddFriendDialog = false">取消</button>
          <button 
            type="button"
            class="btn-confirm-send"
            @click="sendFriendRequest"
            :disabled="sendingFriendRequest"
          >
            <span v-if="sendingFriendRequest" class="loading-spinner"></span>
            <span v-else>发送</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useContactsStore } from '../stores/contacts'
import { userApi, socialApi } from '../utils/api'

const props = defineProps({
  userId: {
    type: String,
    required: true
  },
  // 可选：好友申请信息（用于显示接受申请按钮）
  friendRequest: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'accepted', 'open-accept-dialog'])

const router = useRouter()
const authStore = useAuthStore()
const contactsStore = useContactsStore()

const loading = ref(false)
const user = ref(null)
const showAddFriendDialog = ref(false)
const requestMessage = ref('')
const sendingFriendRequest = ref(false)

// 计算是否有待处理的好友申请（我收到的申请）
const pendingFriendRequest = computed(() => {
  if (!props.friendRequest) return false
  return props.friendRequest.status === 'pending' && 
         props.friendRequest.requestType === 'received' &&
         props.friendRequest.handleResult === 0
})

// 加载用户信息
const loadUserInfo = async () => {
  if (!props.userId) return
  
  loading.value = true
  try {
    const userData = await userApi.getUserById(props.userId)
    
    if (!userData) {
      user.value = null
      return
    }
    
    // 解析标签
    let tags = []
    if (userData.tags) {
      try {
        tags = typeof userData.tags === 'string' ? JSON.parse(userData.tags) : userData.tags
      } catch (e) {
        tags = []
      }
    }
    
    user.value = {
      id: userData.id,
      name: userData.nickname || userData.name || '未设置昵称',
      account: userData.id,
      avatar: userData.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${userData.id}`,
      signature: userData.introduction || '',
      gender: userData.sex === 1 ? 'male' : userData.sex === 2 ? 'female' : 'unknown',
      region: userData.region || '',
      occupation: userData.occupation || '',
      tags: tags,
      isFriend: false
    }
    
    // 检查是否是好友
    await checkFriendStatus()
    
    // 允许重复发送好友申请（不检查是否已发送）
  } catch (error) {
    console.error('加载用户信息失败:', error)
    user.value = null
  } finally {
    loading.value = false
  }
}

// 检查好友关系
const checkFriendStatus = async () => {
  try {
    await contactsStore.fetchFriends()
    const isFriend = contactsStore.friends.some(
      friend => friend.friend_uid === props.userId || friend.id === props.userId
    )
    if (user.value) {
      user.value.isFriend = isFriend
    }
  } catch (error) {
    console.error('检查好友关系失败:', error)
  }
}

// 接受好友申请 - 打开处理申请的弹窗
const handleAcceptRequest = () => {
  if (!props.friendRequest) return
  
  // 关闭用户资料卡片
  emit('close')
  
  // 通知父组件打开处理申请的弹窗
  emit('open-accept-dialog', props.friendRequest)
}

// 发送好友申请
const sendFriendRequest = async () => {
  if (!user.value || sendingFriendRequest.value) return
  
  sendingFriendRequest.value = true
  
  try {
    await contactsStore.sendFriendRequest(props.userId, requestMessage.value.trim())
    
    showAddFriendDialog.value = false
    requestMessage.value = ''
    
    alert('好友申请已发送')
  } catch (error) {
    console.error('发送好友申请失败:', error)
    alert(error.message || '发送失败，请稍后重试')
  } finally {
    sendingFriendRequest.value = false
  }
}

// 发消息
const startChat = () => {
  close()
  router.push(`/app/chat?userId=${user.value.id}`)
}

// 关闭弹窗
const close = () => {
  emit('close')
}

// 监听 userId 变化
watch(() => props.userId, (newUserId) => {
  if (newUserId) {
    loadUserInfo()
  }
}, { immediate: true })

onMounted(() => {
  loadUserInfo()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12);
  width: 90%;
  max-width: 460px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  border-bottom: 1px solid #f3f4f6;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.modal-header h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: #111827;
}

.btn-close, .modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 18px;
  color: #9ca3af;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.15s;
}

.btn-close:hover, .modal-close:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
}

.loading-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  color: #9ca3af;
  gap: 10px;
}

.loading-state .icon {
  font-size: 28px;
  animation: spin 0.8s infinite linear;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-state .icon {
  font-size: 40px;
  color: #ef4444;
}

.user-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f3f4f6;
}

.avatar-container {
  flex-shrink: 0;
}

.avatar {
  width: 72px;
  height: 72px;
  border-radius: 12px;
  object-fit: cover;
  border: 1px solid #e5e7eb;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.name {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.account {
  font-size: 13px;
  color: #9ca3af;
  margin-bottom: 6px;
}

.signature {
  font-size: 14px;
  color: #6b7280;
  line-height: 1.5;
}

.signature.placeholder {
  color: #9ca3af;
  font-style: italic;
}

.user-details {
  margin-bottom: 20px;
}

.detail-section {
  margin-bottom: 16px;
}

.detail-section h4 {
  margin: 0 0 10px 0;
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-item {
  display: flex;
  margin-bottom: 8px;
  align-items: center;
}

.detail-item .label {
  flex: 0 0 72px;
  font-size: 13px;
  color: #9ca3af;
}

.detail-item .value {
  font-size: 14px;
  color: #374151;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  background-color: #f3f4f6;
  color: #374151;
  border-radius: 6px;
  font-size: 13px;
}

.user-actions {
  display: flex;
  gap: 10px;
  padding-top: 16px;
  border-top: 1px solid #f3f4f6;
}

.btn-chat,
.btn-add-friend,
.btn-accept-request,
.btn-request-sent {
  flex: 1;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: background-color 0.15s;
}

.btn-chat {
  background-color: #2563eb;
  color: white;
}

.btn-chat:hover {
  background-color: #1d4ed8;
}

.btn-accept-request {
  background-color: #22c55e;
  color: white;
}

.btn-accept-request:hover:not(:disabled) {
  background-color: #16a34a;
}

.btn-accept-request:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-add-friend {
  background-color: #2563eb;
  color: white;
}

.btn-add-friend:hover:not(:disabled) {
  background-color: #1d4ed8;
}

.btn-add-friend:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-request-sent {
  background-color: #f3f4f6;
  color: #6b7280;
  cursor: default;
}

.add-friend-overlay {
  z-index: 1001;
}

.add-friend-dialog {
  max-width: 380px;
}

.user-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding: 14px;
  background: #f9fafb;
  border-radius: 8px;
}

.user-preview .avatar {
  width: 56px;
  height: 56px;
  border-radius: 10px;
}

.user-preview .user-name {
  font-size: 15px;
  font-weight: 600;
  color: #111827;
}

.message-input {
  margin-top: 12px;
}

.message-input label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.message-input textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  outline: none;
  transition: border-color 0.15s;
}

.message-input textarea:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: #9ca3af;
  margin-top: 4px;
}

.modal-footer {
  display: flex;
  gap: 8px;
  padding: 16px 24px;
  border-top: 1px solid #f3f4f6;
  justify-content: flex-end;
}

.btn-cancel-modal {
  padding: 8px 16px;
  background: #fff;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-cancel-modal:hover {
  background-color: #f9fafb;
}

.btn-confirm-send {
  padding: 8px 20px;
  background-color: #2563eb;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-confirm-send:hover:not(:disabled) {
  background-color: #1d4ed8;
}

.btn-confirm-send:disabled {
  background-color: #d1d5db;
  cursor: not-allowed;
}

.loading-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}
</style>
