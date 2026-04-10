<template>
  <div class="modal-overlay" @click.self="close">
    <div class="add-friend-modal">
      <div class="modal-header">
        <h3>添加好友</h3>
        <button class="btn-close" @click="close">
          <i class="icon icon-close"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="search-container">
          <i class="icon icon-search"></i>
          <input 
            type="text" 
            v-model="searchKeyword" 
            placeholder="输入用户昵称、手机号或邮箱" 
            @input="handleSearch"
            @keyup.enter="handleSearch"
          >
          <button v-if="searching" class="btn-loading">
            <i class="icon icon-spinner"></i>
          </button>
        </div>
        
        <div v-if="searchResults.length > 0" class="search-results">
          <div 
            class="result-item" 
            v-for="user in searchResults" 
            :key="user.id"
            @click="viewUserProfile(user)"
          >
            <img :src="user.avatar" alt="头像" class="avatar" @click.stop="viewUserProfile(user)">
            <div class="user-info" @click.stop>
              <div class="name">{{ user.name }} 
                <span class="account">@{{ user.account }}</span>
              </div>
              <div class="tags">
                <span v-for="tag in user.tags" :key="tag" class="tag">{{ tag }}</span>
              </div>
            </div>
            <div class="user-actions" @click.stop>
              <button 
                v-if="user.isSelf"
                class="btn-sending"
                disabled
              >
                <span>自己</span>
              </button>
              <button
                v-else-if="user.isFriend"
                class="btn-add"
                @click="startChat(user)"
              >
                <i class="icon icon-chat"></i>
                <span>发消息</span>
              </button>
              <button 
                v-else
                class="btn-add" 
                @click="showRequestDialog(user)"
                :disabled="sendingRequest"
              >
              <i class="icon icon-add"></i>
                <span>添加</span>
              </button>
            </div>
          </div>
        </div>
        
        <!-- 申请消息对话框 -->
        <div v-if="showRequestMessageDialog" class="request-message-dialog">
          <div class="dialog-overlay" @click="showRequestMessageDialog = false"></div>
          <div class="dialog-content">
            <div class="dialog-header">
              <h4>发送好友申请</h4>
              <button class="btn-close" @click="showRequestMessageDialog = false">
                <i class="icon icon-close"></i>
              </button>
            </div>
            <div class="dialog-body">
              <div class="user-preview">
                <img :src="selectedUser?.avatar" alt="头像" class="avatar">
                <div class="user-name">{{ selectedUser?.name }}</div>
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
            <div class="dialog-footer">
              <button class="btn-cancel" @click="showRequestMessageDialog = false">取消</button>
              <button 
                class="btn-confirm" 
                @click="confirmSendRequest"
                :disabled="sendingRequest"
              >
                <span v-if="sendingRequest">
                  <i class="icon icon-spinner"></i> 发送中...
                </span>
                <span v-else>发送</span>
            </button>
            </div>
          </div>
        </div>
        
        <!-- 提示消息 -->
        <div v-if="message" :class="['message-toast', messageType]">
          <i :class="messageType === 'success' ? 'icon icon-check' : 'icon icon-error'"></i>
          <span>{{ message }}</span>
        </div>
        
        <div v-else-if="hasSearched && !searching" class="no-results">
          <p>没有找到用户</p>
          <p class="tip">请检查搜索条件是否正确</p>
          <div class="search-tips">
            <p class="tips-title">搜索提示：</p>
            <ul>
              <li>昵称支持模糊匹配</li>
              <li>手机号和邮箱需要完整输入</li>
            </ul>
          </div>
        </div>
        
        <div v-else class="instruction">
          <p>请输入用户昵称、手机号或邮箱进行搜索</p>
          <div class="tips">
            <div class="tip-item">
              <i class="icon icon-tip"></i>
              <span>昵称支持模糊匹配</span>
            </div>
            <div class="tip-item">
              <i class="icon icon-tip"></i>
              <span>手机号和邮箱需要完整输入</span>
            </div>
            <div class="tip-item">
              <i class="icon icon-tip"></i>
              <span>对方接受后你们将成为好友</span>
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
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { userApi, socialApi } from '../utils/api'
import { useContactsStore } from '../stores/contacts'
import { useAuthStore } from '../stores/auth'
import UserProfileCard from './UserProfileCard.vue'

const emit = defineEmits(['close', 'search', 'send-request'])

const router = useRouter()
const contactsStore = useContactsStore()
const authStore = useAuthStore()

const showUserProfileCard = ref(false)
const selectedUserId = ref(null)

const searchKeyword = ref('')
const searchResults = ref([])
const searching = ref(false)
const hasSearched = ref(false)

// 申请消息对话框
const showRequestMessageDialog = ref(false)
const selectedUser = ref(null)
const requestMessage = ref('')
const sendingRequest = ref(false)

// 提示消息
const message = ref('')
const messageType = ref('success') // 'success' | 'error'

// 判断搜索关键词类型
const getSearchParams = (keyword) => {
  const trimmed = keyword.trim()
  
  // 判断是否是手机号（11位数字，以1开头）
  if (/^1[3-9]\d{9}$/.test(trimmed)) {
    return { phone: trimmed }
  }

  // 纯数字（非手机号格式）：按用户ID精准匹配
  if (/^\d+$/.test(trimmed)) {
    return { ids: [trimmed] }
  }
  
  // 判断是否是邮箱（包含@符号）
  if (trimmed.includes('@') && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed)) {
    return { email: trimmed }
  }
  
  // 否则作为昵称搜索
  return { name: trimmed }
}

const close = () => {
  emit('close')
}

const currentUserId = computed(() => {
  if (authStore.user?.id) return String(authStore.user.id)
  try {
    const saved = localStorage.getItem('user')
    if (saved) return String(JSON.parse(saved)?.id || '')
  } catch {}
  return ''
})

const isFriendById = (uid) => {
  const id = String(uid)
  return contactsStore.friends.some(f => String(f.friend_uid) === id || String(f.id) === id)
}

const startChat = (user) => {
  router.push(`/app/chat?userId=${user.id}`)
  close()
}

// 执行搜索
const handleSearch = async () => {
  if (!searchKeyword.value.trim()) {
    searchResults.value = []
    hasSearched.value = false
    return
  }
  
  searching.value = true
  hasSearched.value = true
  
  try {
    // 用于判断按钮状态：自己 / 已是好友
    await contactsStore.fetchFriends()

    const params = getSearchParams(searchKeyword.value)
    const response = await userApi.searchUser(params)
    
    if (response && response.users) {
      // 转换数据格式以匹配组件期望的格式
      searchResults.value = response.users.map(user => ({
        id: user.id,
        name: user.nickname || user.name || '未设置昵称',
        account: user.id,
        avatar: user.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${user.id}`,
        tags: parseTags(user.tags),
        mobile: user.mobile,
        email: user.email,
        isSelf: currentUserId.value && String(user.id) === String(currentUserId.value),
        isFriend: isFriendById(user.id)
      }))
    } else {
      searchResults.value = []
    }
  } catch (err) {
    console.error('搜索用户失败:', err)
    searchResults.value = []
  } finally {
    searching.value = false
  }
}

// 解析标签（可能是JSON字符串或数组）
const parseTags = (tags) => {
  if (!tags) return []
  if (Array.isArray(tags)) return tags
  try {
    const parsed = JSON.parse(tags)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

// 显示申请消息对话框
const showRequestDialog = (user) => {
  if (user.isSelf) {
    showMessage('不能添加自己', 'error')
    return
  }
  if (user.isFriend) {
    startChat(user)
    return
  }
  selectedUser.value = user
  requestMessage.value = ''
  showRequestMessageDialog.value = true
}

// 确认发送申请
const confirmSendRequest = async () => {
  if (!selectedUser.value) return
  
  sendingRequest.value = true
  
  try {
    if (selectedUser.value.isSelf) {
      showMessage('不能添加自己', 'error')
      return
    }
    if (selectedUser.value.isFriend) {
      showMessage('你们已经是好友了', 'success')
      showRequestMessageDialog.value = false
      return
    }
    // 调用 API 发送好友申请
    await socialApi.friendPutIn(selectedUser.value.id, requestMessage.value.trim())
    
    // 显示成功提示
    showMessage('好友申请已发送', 'success')
    
    // 关闭对话框
    showRequestMessageDialog.value = false
    
    // 可选：触发父组件事件（用于刷新好友申请列表）
    emit('send-request', selectedUser.value.id)
    
    // 3秒后自动关闭提示
    setTimeout(() => {
      message.value = ''
    }, 3000)
  } catch (error) {
    console.error('发送好友申请失败:', error)
    
    // 更新用户状态
    const userIndex = searchResults.value.findIndex(u => u.id === selectedUser.value.id)
    if (userIndex !== -1) {
      searchResults.value[userIndex].sending = false
    }
    
    // 显示错误提示
    const errorMsg = error.message || '发送失败，请稍后重试'
    showMessage(errorMsg, 'error')
    
    // 5秒后自动关闭错误提示
    setTimeout(() => {
      message.value = ''
    }, 5000)
  } finally {
    sendingRequest.value = false
  }
}

// 显示提示消息
const showMessage = (msg, type = 'success') => {
  message.value = msg
  messageType.value = type
}

// 查看用户信息主页
const viewUserProfile = (user) => {
  selectedUserId.value = user.id
  showUserProfileCard.value = true
}
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
}

.add-friend-modal {
  background-color: var(--card);
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--border);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid var(--border);
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-family: var(--font-display);
  color: var(--fg);
}

.btn-close {
  background: none;
  border: none;
  font-size: 20px;
  color: var(--muted);
  cursor: pointer;
  transition: color 0.2s;
}

.btn-close:hover {
  color: var(--fg);
}

.modal-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
  background-color: var(--input-bg);
  border-radius: var(--radius-md);
  padding: 10px 15px;
  margin-bottom: 20px;
  border: 1px solid var(--border);
}

.search-container .icon {
  color: var(--muted);
  margin-right: 10px;
}

.search-container input {
  flex: 1;
  border: none;
  background: none;
  font-size: 14px;
  font-family: var(--font-ui);
  outline: none;
  color: var(--fg);
}

.search-container input::placeholder {
  color: var(--muted);
}

.btn-loading {
  background: none;
  border: none;
  padding: 5px;
  cursor: pointer;
}

.icon-spinner {
  animation: spin 1s infinite linear;
  color: var(--muted);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.search-results {
  margin-top: 10px;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  transition: background-color 0.2s;
}

.result-item:hover {
  background-color: var(--card-hover);
}

.result-item .avatar {
  width: 50px;
  height: 50px;
  border-radius: var(--radius-sm);
  margin-right: 15px;
}

.user-info {
  flex: 1;
}

.name {
  font-weight: 500;
  margin-bottom: 5px;
  color: var(--fg);
}

.account {
  font-weight: normal;
  font-size: 13px;
  color: var(--muted);
}

.tags {
  display: flex;
  gap: 5px;
}

.tag {
  font-size: 12px;
  padding: 3px 8px;
  background-color: var(--accent-muted);
  border-radius: 4px;
  color: var(--accent);
}

.btn-add {
  background: var(--accent);
  border: none;
  color: var(--bg);
  font-size: 14px;
  font-weight: 500;
  font-family: var(--font-ui);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  white-space: nowrap;
}

.btn-add:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-add:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-add .icon {
  font-size: 16px;
}

.no-results, .instruction {
  text-align: center;
  padding: 30px 20px;
  color: var(--muted);
}

.tip {
  margin-top: 10px;
  font-size: 13px;
}

.search-tips {
  margin-top: 16px;
  text-align: left;
  background: var(--input-bg);
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-light);
}

.tips-title {
  font-weight: 600;
  color: var(--fg);
  margin-bottom: 8px;
  font-size: 13px;
}

.search-tips ul {
  margin: 0;
  padding-left: 20px;
  color: var(--muted);
  font-size: 12px;
}

.search-tips li {
  margin-bottom: 4px;
}

.tips {
  margin-top: 20px;
  text-align: left;
}

.tip-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  font-size: 13px;
  color: var(--muted);
}

.tip-item i {
  margin-right: 8px;
  color: var(--accent);
}

.user-actions {
  display: flex;
  align-items: center;
}

.btn-sending {
  background: var(--border);
  border: none;
  color: var(--muted);
  font-size: 14px;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: not-allowed;
  white-space: nowrap;
}

.btn-sending .icon {
  font-size: 16px;
}

.btn-sent {
  color: var(--success);
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn-sent .icon {
  font-size: 16px;
}

/* 申请消息对话框 */
.request-message-dialog {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 2000;
}

.dialog-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--overlay);
}

.dialog-content {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: var(--card);
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 400px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--border);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid var(--border);
}

.dialog-header h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  font-family: var(--font-display);
  color: var(--fg);
}

.dialog-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

.user-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 12px;
  background: var(--input-bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-light);
}

.user-preview .avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-sm);
}

.user-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--fg);
}

.message-input {
  margin-top: 16px;
}

.message-input label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--muted);
  font-weight: 500;
}

.message-input textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-family: var(--font-ui);
  background: var(--input-bg);
  color: var(--fg);
  resize: vertical;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.message-input textarea::placeholder {
  color: var(--muted);
}

.message-input textarea:focus {
  border-color: var(--accent);
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border);
}

.btn-cancel,
.btn-confirm {
  padding: 8px 20px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  font-family: var(--font-ui);
  cursor: pointer;
  border: none;
  transition: background-color 0.2s;
}

.btn-cancel {
  background: var(--border);
  color: var(--muted);
}

.btn-cancel:hover {
  background: var(--card-hover);
  color: var(--fg);
}

.btn-confirm {
  background: var(--accent);
  color: var(--bg);
}

.btn-confirm:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 提示消息 */
.message-toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 20px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-family: var(--font-ui);
  z-index: 3000;
  animation: slideDown 0.3s ease-out;
}

.message-toast.success {
  background: var(--success);
  color: var(--fg);
}

.message-toast.error {
  background: var(--danger);
  color: var(--fg);
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}
</style>