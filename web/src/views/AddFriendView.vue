<template>
  <div class="add-friend-view">
    <div class="container">
      <div class="header">
        <h2>添加好友</h2>
        <p class="subtitle">搜索用户并发送好友申请</p>
      </div>

      <!-- 搜索框 -->
      <div class="search-section">
        <div class="search-container">
          <i class="icon icon-search"></i>
          <input 
            v-model="searchKeyword" 
            type="text"
            placeholder="输入用户昵称、手机号或邮箱" 
            @keyup.enter="handleSearch"
            @input="handleInput"
          />
          <button 
            v-if="searching" 
            class="btn-loading"
            disabled
          >
            <i class="icon icon-spinner"></i>
          </button>
          <button 
            v-else
            class="btn-search"
            @click="handleSearch"
            :disabled="!searchKeyword.trim()"
          >
            搜索
          </button>
        </div>
      </div>

      <!-- 搜索结果 -->
        <div v-if="searchResults.length > 0" class="results-section">
          <div class="results-header">
            <h3>搜索结果</h3>
            <span class="count">找到 {{ searchResults.length }} 个用户</span>
          </div>
          <div class="results-list">
            <div 
              v-for="user in searchResults" 
              :key="user.id" 
              class="result-item"
              @click="viewUserProfile(user)"
            >
            <img :src="user.avatar" :alt="user.name" class="avatar" />
            <div class="user-info" @click.stop>
              <div class="name-row">
                <div class="name">{{ user.name }}</div>
                <span v-if="user.requestSent" class="status-badge sent">
                  <i class="icon icon-check"></i> 已发送
                </span>
                <span v-else-if="user.sending" class="status-badge sending">
                  <i class="icon icon-spinner"></i> 发送中...
                </span>
              </div>
              <div class="account">ID: {{ user.account }}</div>
              <div v-if="user.tags && user.tags.length > 0" class="tags">
                <span v-for="tag in user.tags" :key="tag" class="tag">{{ tag }}</span>
              </div>
              <div v-if="user.mobile" class="detail">
                <i class="icon icon-phone"></i> {{ user.mobile }}
              </div>
              <div v-if="user.email" class="detail">
                <i class="icon icon-email"></i> {{ user.email }}
              </div>
            </div>
            <div class="user-actions" @click.stop>
              <button 
                v-if="!user.requestSent && !user.sending"
                class="btn-add" 
                @click="showRequestDialog(user)"
                :disabled="sendingRequest"
              >
                <i class="icon icon-add"></i>
                添加好友
              </button>
              <button 
                v-else-if="user.sending"
                class="btn-sending" 
                disabled
              >
                <i class="icon icon-spinner"></i>
                发送中...
              </button>
              <button 
                v-else
                class="btn-sent" 
                disabled
              >
                <i class="icon icon-check"></i>
                已发送
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else-if="hasSearched && !searching" class="empty-state">
        <i class="icon icon-search"></i>
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

      <!-- 初始状态 -->
      <div v-else class="initial-state">
        <i class="icon icon-users"></i>
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
            <img :src="selectedUser?.avatar" alt="头像" class="avatar" />
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
import UserProfileCard from '../components/UserProfileCard.vue'

const router = useRouter()

const searchKeyword = ref('')
const searchResults = ref([])
const searching = ref(false)
const hasSearched = ref(false)
const showUserProfileCard = ref(false)
const selectedUserId = ref(null)

// 申请消息对话框
const showRequestMessageDialog = ref(false)
const selectedUser = ref(null)
const requestMessage = ref('')
const sendingRequest = ref(false)

// 提示消息
const message = ref('')
const messageType = ref('success')

// 判断搜索关键词类型
const getSearchParams = (keyword) => {
  const trimmed = keyword.trim()
  
  // 判断是否是手机号（11位数字，以1开头）
  if (/^1[3-9]\d{9}$/.test(trimmed)) {
    return { phone: trimmed }
  }
  
  // 判断是否是邮箱（包含@符号）
  if (trimmed.includes('@') && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed)) {
    return { email: trimmed }
  }
  
  // 否则作为昵称搜索
  return { name: trimmed }
}

// 输入处理（防抖）
let searchTimer = null
const handleInput = () => {
  // 可以在这里实现防抖搜索
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
    const params = getSearchParams(searchKeyword.value)
    const response = await userApi.searchUser(params)
    
    if (response && response.users) {
      // 转换数据格式
      searchResults.value = response.users.map(user => ({
        id: user.id,
        name: user.nickname || user.name || '未设置昵称',
        account: user.id,
        avatar: user.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${user.id}`,
        tags: parseTags(user.tags),
        mobile: user.mobile,
        email: user.email,
        requestSent: false,
        sending: false
      }))
    } else {
      searchResults.value = []
    }
  } catch (err) {
    console.error('搜索用户失败:', err)
    searchResults.value = []
    showMessage(err.message || '搜索失败，请稍后重试', 'error')
  } finally {
    searching.value = false
  }
}

// 解析标签
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
  selectedUser.value = user
  requestMessage.value = ''
  showRequestMessageDialog.value = true
}

// 确认发送申请
const confirmSendRequest = async () => {
  if (!selectedUser.value) return
  
  sendingRequest.value = true
  
  // 更新用户状态为发送中
  const userIndex = searchResults.value.findIndex(u => u.id === selectedUser.value.id)
  if (userIndex !== -1) {
    searchResults.value[userIndex].sending = true
  }
  
  try {
    // 调用 API 发送好友申请
    await socialApi.friendPutIn(selectedUser.value.id, requestMessage.value.trim())
    
    // 更新用户状态
    if (userIndex !== -1) {
      searchResults.value[userIndex].requestSent = true
      searchResults.value[userIndex].sending = false
    }
    
    // 显示成功提示
    showMessage('好友申请已发送', 'success')
    
    // 关闭对话框
    showRequestMessageDialog.value = false
    
    // 3秒后自动关闭提示
    setTimeout(() => {
      message.value = ''
    }, 3000)
  } catch (error) {
    console.error('发送好友申请失败:', error)
    
    // 更新用户状态
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
.add-friend-view {
  min-height: 100vh;
  background: #f8fafc;
  padding: 20px;
}

.container {
  max-width: 800px;
  margin: 0 auto;
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.header {
  margin-bottom: 24px;
  text-align: center;
}

.header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.subtitle {
  margin: 0;
  font-size: 14px;
  color: #666;
}

.search-section {
  margin-bottom: 24px;
}

.search-container {
  display: flex;
  align-items: center;
  background: #f8f9fa;
  border-radius: 10px;
  padding: 12px 16px;
  border: 1px solid #e9ecef;
  transition: border-color 0.2s;
}

.search-container:focus-within {
  border-color: #2563eb;
}

.search-container .icon {
  color: #999;
  margin-right: 12px;
  font-size: 18px;
}

.search-container input {
  flex: 1;
  border: none;
  background: none;
  font-size: 15px;
  outline: none;
  color: #333;
}

.search-container input::placeholder {
  color: #999;
}

.btn-search,
.btn-loading {
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  margin-left: 12px;
}

.btn-search {
  background: #2563eb;
  color: white;
}

.btn-search:hover:not(:disabled) {
  background: #357abd;
}

.btn-search:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-loading {
  background: #e9ecef;
  color: #999;
  cursor: not-allowed;
}

.icon-spinner {
  animation: spin 1s infinite linear;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.results-section {
  margin-top: 24px;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.results-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.count {
  font-size: 14px;
  color: #666;
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 12px;
  border: 1px solid #e9ecef;
  transition: all 0.2s;
  cursor: pointer;
}

.result-item:hover {
  background: #f0f2f5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transform: translateY(-1px);
}

.avatar {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  object-fit: cover;
  flex-shrink: 0;
}

.user-info {
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
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.status-badge.sent {
  background: #d4edda;
  color: #155724;
}

.status-badge.sending {
  background: #fff3cd;
  color: #856404;
}

.account {
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}

.tags {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.tag {
  font-size: 12px;
  padding: 4px 8px;
  background: #e9ecef;
  border-radius: 4px;
  color: #666;
}

.detail {
  font-size: 13px;
  color: #666;
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.user-actions {
  flex-shrink: 0;
}

.btn-add,
.btn-sending,
.btn-sent {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.btn-add {
  background: #2563eb;
  color: white;
}

.btn-add:hover:not(:disabled) {
  background: #357abd;
}

.btn-sending {
  background: #e9ecef;
  color: #666;
  cursor: not-allowed;
}

.btn-sent {
  background: #d4edda;
  color: #155724;
  cursor: default;
}

.empty-state,
.initial-state {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-state .icon,
.initial-state .icon {
  font-size: 64px;
  color: #ccc;
  margin-bottom: 16px;
}

.empty-state p,
.initial-state p {
  margin: 8px 0;
  font-size: 16px;
}

.tip {
  font-size: 14px;
  color: #999;
  margin-top: 8px;
}

.search-tips {
  margin-top: 24px;
  text-align: left;
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}

.tips-title {
  font-weight: 600;
  color: #666;
  margin-bottom: 8px;
  font-size: 14px;
}

.search-tips ul {
  margin: 0;
  padding-left: 20px;
  color: #999;
  font-size: 13px;
}

.tips {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}

.tip-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #666;
}

.tip-item .icon {
  font-size: 16px;
  color: #2563eb;
}

/* 申请消息对话框样式（与 AddFriendModal 相同） */
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
  background-color: rgba(0, 0, 0, 0.5);
}

.dialog-content {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #f0f2f5;
}

.dialog-header h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.dialog-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

.user-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.user-preview .avatar {
  width: 64px;
  height: 64px;
  border-radius: 12px;
}

.user-name {
  font-size: 16px;
  font-weight: 600;
}

.message-input {
  margin-top: 16px;
}

.message-input label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.message-input textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  outline: none;
  transition: border-color 0.2s;
}

.message-input textarea:focus {
  border-color: #2563eb;
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #f0f2f5;
}

.btn-cancel,
.btn-confirm {
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-cancel {
  background: #f0f2f5;
  color: #666;
}

.btn-cancel:hover {
  background: #e0e0e0;
}

.btn-confirm {
  background: #2563eb;
  color: white;
}

.btn-confirm:hover:not(:disabled) {
  background: #357abd;
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 提示消息 */
.message-toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 20px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  z-index: 3000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  animation: slideDown 0.3s ease-out;
}

.message-toast.success {
  background: #28a745;
  color: white;
}

.message-toast.error {
  background: #dc3545;
  color: white;
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
