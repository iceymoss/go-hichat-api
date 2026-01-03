<template>
  <div class="modal-overlay" @click.self="close">
    <div class="search-user-modal">
      <div class="modal-header">
        <h3>搜索用户</h3>
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
            @input="handleSearchInput"
            @keyup.enter="performSearch"
          >
          <button 
            v-if="searching" 
            class="btn-loading"
            disabled
          >
            <i class="icon icon-spinner"></i>
          </button>
          <button 
            v-else-if="searchKeyword.trim()"
            class="btn-search-action"
            @click="performSearch"
          >
            搜索
          </button>
        </div>
        
        <div v-if="searchResults.length > 0" class="search-results">
          <div class="result-item" v-for="user in searchResults" :key="user.id">
            <img :src="user.avatar || getDefaultAvatar(user.id)" alt="头像" class="avatar">
            <div class="user-info">
              <div class="name">{{ user.nickname || user.name || '未设置昵称' }}</div>
              <div class="meta">
                <span v-if="user.mobile" class="meta-item">
                  <i class="icon icon-phone"></i>
                  {{ user.mobile }}
                </span>
                <span v-if="user.email" class="meta-item">
                  <i class="icon icon-email"></i>
                  {{ user.email }}
                </span>
              </div>
              <div v-if="user.tags && user.tags.length > 0" class="tags">
                <span v-for="tag in parseTags(user.tags)" :key="tag" class="tag">{{ tag }}</span>
              </div>
            </div>
            <button class="btn-chat" @click="startChat(user)" title="开始聊天">
              <i class="icon icon-message"></i>
            </button>
          </div>
        </div>
        
        <div v-else-if="hasSearched && !searching" class="no-results">
          <i class="icon icon-search-empty"></i>
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
          <i class="icon icon-search-large"></i>
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
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { userApi } from '../utils/api'

const emit = defineEmits(['close', 'start-chat'])

const searchKeyword = ref('')
const searchResults = ref([])
const searching = ref(false)
const hasSearched = ref(false)

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

// 执行搜索
const performSearch = async () => {
  if (!searchKeyword.value.trim()) {
    return
  }
  
  searching.value = true
  hasSearched.value = true
  searchResults.value = []
  
  try {
    const params = getSearchParams(searchKeyword.value)
    const response = await userApi.searchUser(params)
    
    if (response && response.users) {
      searchResults.value = response.users
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

// 输入时延迟搜索（防抖）
let searchTimer = null
const handleSearchInput = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  
  // 如果输入为空，清空结果
  if (!searchKeyword.value.trim()) {
    searchResults.value = []
    hasSearched.value = false
    return
  }
  
  // 延迟500ms执行搜索
  searchTimer = setTimeout(() => {
    performSearch()
  }, 500)
}

// 开始聊天
const startChat = (user) => {
  emit('start-chat', user.id)
}

// 关闭模态框
const close = () => {
  emit('close')
}

// 获取默认头像
const getDefaultAvatar = (userId) => {
  return `https://api.dicebear.com/7.x/personas/svg?seed=${userId || 'default'}`
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
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.search-user-modal {
  background-color: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  color: #999;
  cursor: pointer;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-close:hover {
  background-color: #f5f5f5;
  color: #333;
}

.modal-body {
  padding: 24px;
  flex: 1;
  overflow-y: auto;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
  background-color: #f0f2f5;
  border-radius: 10px;
  padding: 12px 16px;
  margin-bottom: 20px;
  gap: 10px;
}

.search-container .icon {
  color: #999;
  font-size: 18px;
}

.search-container input {
  flex: 1;
  border: none;
  background: none;
  font-size: 14px;
  outline: none;
  color: #333;
}

.search-container input::placeholder {
  color: #999;
}

.btn-search-action {
  padding: 6px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-search-action:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-loading {
  background: none;
  border: none;
  padding: 6px;
  cursor: default;
}

.icon-spinner {
  animation: spin 1s infinite linear;
  color: #999;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.search-results {
  margin-top: 10px;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border: 1px solid #f0f2f5;
  border-radius: 12px;
  margin-bottom: 12px;
  transition: all 0.2s;
  background: #fff;
}

.result-item:hover {
  border-color: #4a8cff;
  box-shadow: 0 2px 8px rgba(74, 140, 255, 0.1);
  transform: translateY(-1px);
}

.result-item .avatar {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  margin-right: 16px;
  object-fit: cover;
  border: 2px solid #f0f2f5;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.name {
  font-weight: 600;
  font-size: 16px;
  color: #333;
  margin-bottom: 6px;
  word-break: break-word;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #666;
}

.meta-item .icon {
  font-size: 14px;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.tag {
  font-size: 12px;
  padding: 4px 10px;
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.1) 0%, rgba(138, 105, 255, 0.1) 100%);
  border-radius: 12px;
  color: #4a8cff;
  font-weight: 500;
}

.btn-chat {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 18px;
  margin-left: 12px;
}

.btn-chat:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.no-results,
.instruction {
  text-align: center;
  padding: 40px 20px;
  color: #999;
}

.no-results .icon,
.instruction .icon {
  font-size: 48px;
  color: #ddd;
  margin-bottom: 16px;
}

.no-results p,
.instruction p {
  margin: 8px 0;
  font-size: 14px;
}

.tip {
  margin-top: 8px;
  font-size: 13px;
  color: #bbb;
}

.search-tips {
  margin-top: 24px;
  text-align: left;
  background: #f9f9f9;
  padding: 16px;
  border-radius: 8px;
}

.tips-title {
  font-weight: 600;
  color: #666;
  margin-bottom: 8px;
}

.search-tips ul {
  margin: 0;
  padding-left: 20px;
  color: #999;
  font-size: 13px;
}

.search-tips li {
  margin-bottom: 4px;
}

.tips {
  margin-top: 24px;
  text-align: left;
}

.tip-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  color: #666;
  font-size: 14px;
}

.tip-item .icon {
  margin-right: 8px;
  color: #4a8cff;
  font-size: 16px;
}

@media (max-width: 480px) {
  .search-user-modal {
    width: 95%;
    max-height: 85vh;
  }
  
  .result-item {
    padding: 12px;
  }
  
  .result-item .avatar {
    width: 48px;
    height: 48px;
    margin-right: 12px;
  }
}
</style>

