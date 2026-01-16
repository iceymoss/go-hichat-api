<template>
  <div class="new-friends-wrapper">
    <div class="new-friends-container">
      <div class="header">
        <div class="title-group">
          <div class="title">新的朋友</div>
          <div class="filter-tabs">
            <span 
              :class="{ active: filter === 'all' }" 
              @click="filter = 'all'"
            >全部</span>
            <span 
              :class="{ active: filter === 'pending' }" 
              @click="filter = 'pending'"
            >待处理</span>
          </div>
        </div>
      </div>

      <div class="request-list">
        <div v-if="filteredRequests.length === 0" class="empty-state">
          <div class="empty-icon-circle">
            <i class="icon icon-user-plus"></i>
          </div>
          <p>暂无新的朋友</p>
        </div>

        <div 
          v-for="req in filteredRequests" 
          :key="req.id" 
          class="request-item"
          @click="showDetail(req)"
        >
          <img :src="req.avatar" alt="头像" class="avatar">
          <div class="info">
            <div class="name-row">
              <span class="name">{{ req.name }}</span>
              <span class="tag-label" v-if="req.status === 'pending'">申请中</span>
            </div>
            <div class="message">{{ req.req_msg || '请求添加你为好友' }}</div>
          </div>
          <div class="action-area">
            <button v-if="req.status === 'pending'" class="btn-accept" @click.stop="handleAccept(req)">
              接受
            </button>
            <span v-else class="status-text">{{ req.statusText }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 处理申请弹窗 -->
    <div v-if="showHandleModal" class="handle-modal-overlay" @click="closeHandleModal">
      <div class="handle-modal" @click.stop>
        <div class="modal-header">
          <h3>通过好友申请</h3>
          <button class="btn-close" @click="closeHandleModal">
            <i class="icon icon-x"></i>
          </button>
        </div>
        
        <div class="user-preview">
          <img :src="currentReq.avatar" alt="头像" class="preview-avatar">
          <div class="preview-info">
            <div class="preview-name">{{ currentReq.name }}</div>
            <div class="preview-msg">验证消息：{{ currentReq.req_msg }}</div>
          </div>
        </div>

        <div class="form-group">
          <label>设置备注</label>
          <input 
            v-model="handleForm.remark" 
            type="text" 
            placeholder="为朋友设置备注"
            class="input-field"
          >
        </div>

        <div class="form-group">
          <label>添加标签</label>
          <div class="tags-input-wrapper">
            <div class="tags-list">
              <span v-for="(tag, index) in handleForm.tags" :key="index" class="tag-item">
                {{ tag }}
                <i class="icon icon-x remove-tag" @click="removeTag(index)"></i>
              </span>
            </div>
            <input 
              v-model="newTag" 
              type="text" 
              placeholder="输入标签按回车添加"
              class="tag-input"
              @keyup.enter="addTag"
            >
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-cancel" @click="closeHandleModal">取消</button>
          <button class="btn-confirm" @click="confirmAccept">完成并添加</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { socialApi } from '../utils/api'

const contactsStore = useContactsStore()
const filter = ref('all')
const showHandleModal = ref(false)
const currentReq = ref(null)
const newTag = ref('')
const handleForm = ref({
  remark: '',
  tags: []
})

// 组件挂载时，将所有申请标记为已读
onMounted(async () => {
  try {
    // 0 表示标记所有为已读
    await socialApi.friendPutInRead(0)
    // 刷新列表（可选，如果后端返回的数据能反映已读状态，但目前我们前端只看 pending）
    // 其实我们主要目的是告诉后端“我已读了”，后端可能会更新 read_state。
    // 前端的红点（ContactsView）是基于 friendRequests 的 pending 数量。
    // 如果“已读”只是为了消红点，但这些请求依然是 pending 状态（待处理），
    // 那么这里就存在歧义：是“未读的申请”还是“待处理的申请”产生红点？
    // 微信逻辑：
    // 1. 红点 = 有新的好友申请（未读）。
    // 2. 点进去后，红点消失。
    // 3. 列表里依然显示“接受”按钮。
    // 所以，红点应该基于 read_state，而不是 status。
    // 目前 ContactsView 计算 unreadCount 是基于 req.status === 'pending'。
    // 我们需要调整 unreadCount 的计算逻辑，或者确保后端返回的 list 中包含 read_state。
    // 刚才我们给 friend_requests 加了 read_state。
    // 让我们看看 fetchFriendRequests 的返回是否包含 read_state。
  } catch (error) {
    console.error('标记已读失败:', error)
  }
})

const filteredRequests = computed(() => {
  const list = contactsStore.friendRequests
  if (filter.value === 'pending') {
    return list.filter(req => req.status === 'pending')
  }
  return list
})

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  const now = new Date()
  
  if (date.toDateString() === now.toDateString()) {
    return date.getHours().toString().padStart(2, '0') + ':' + date.getMinutes().toString().padStart(2, '0')
  }
  return (date.getMonth() + 1) + '/' + date.getDate()
}

const showDetail = (req) => {
  // 如果是待处理状态，可以直接打开处理弹窗
  if (req.status === 'pending') {
    handleAccept(req)
  }
}

const handleAccept = (req) => {
  currentReq.value = req
  handleForm.value.remark = req.name // 默认填入昵称
  handleForm.value.tags = []
  showHandleModal.value = true
}

const closeHandleModal = () => {
  showHandleModal.value = false
  currentReq.value = null
  newTag.value = ''
}

const addTag = () => {
  if (newTag.value.trim()) {
    if (!handleForm.value.tags.includes(newTag.value.trim())) {
      handleForm.value.tags.push(newTag.value.trim())
    }
    newTag.value = ''
  }
}

const removeTag = (index) => {
  handleForm.value.tags.splice(index, 1)
}

const confirmAccept = async () => {
  if (!currentReq.value) return
  
  try {
    // 调用更新后的 handleFriendRequest 方法
    await contactsStore.handleFriendRequest(
      currentReq.value.id, 
      true, 
      handleForm.value.remark,
      handleForm.value.tags
    )
    closeHandleModal()
  } catch (error) {
    console.error('处理好友申请失败:', error)
    alert('操作失败，请重试')
  }
}
</script>

<style scoped>
.new-friends-wrapper {
  height: 100%;
  width: 100%;
  display: flex;
  justify-content: center;
  background: #f5f5f5;
}

.new-friends-container {
  width: 100%;
  max-width: 1000px; /* 增加最大宽度 */
  background: #fff;
  display: flex;
  flex-direction: column;
  box-shadow: 0 0 2px rgba(0,0,0,0.05);
}

.header {
  padding: 0 30px;
  height: 60px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
}

.title-group {
  display: flex;
  align-items: center;
  gap: 20px;
}

.title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.filter-tabs {
  display: flex;
  background: #f5f5f5;
  padding: 2px;
  border-radius: 4px;
}

.filter-tabs span {
  font-size: 13px;
  color: #666;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 4px;
  transition: all 0.2s;
}

.filter-tabs span.active {
  color: #333;
  background: #fff;
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.request-list {
  flex: 1;
  overflow-y: auto;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 100px;
  color: #999;
}

.empty-icon-circle {
  width: 64px;
  height: 64px;
  background: #f5f5f5;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}

.empty-icon-circle .icon {
  font-size: 28px;
  color: #ccc;
}

.request-item {
  display: flex;
  padding: 16px 30px;
  border-bottom: 1px solid #f9f9f9;
  cursor: pointer;
  transition: background 0.2s;
  align-items: center;
}

.request-item:hover {
  background: #fbfbfb;
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 6px;
  margin-right: 16px;
  object-fit: cover;
}

.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name {
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.tag-label {
  font-size: 11px;
  background: #fffbe6;
  color: #faad14;
  padding: 1px 4px;
  border-radius: 2px;
}

.message {
  font-size: 13px;
  color: #999;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.action-area {
  margin-left: 20px;
}

.btn-accept {
  background: #07c160;
  color: white;
  border: none;
  padding: 6px 20px;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
}

.btn-accept:hover {
  background: #06ad56;
}

.status-text {
  font-size: 13px;
  color: #999;
}

/* Modal Styles */
.handle-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.handle-modal {
  background: white;
  border-radius: 16px;
  width: 400px;
  padding: 24px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1e293b;
}

.btn-close {
  background: none;
  border: none;
  font-size: 20px;
  color: #94a3b8;
  cursor: pointer;
}

.user-preview {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
}

.preview-avatar {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  margin-right: 12px;
}

.preview-name {
  font-weight: 500;
  color: #1e293b;
  margin-bottom: 2px;
}

.preview-msg {
  font-size: 12px;
  color: #64748b;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  color: #475569;
  margin-bottom: 8px;
}

.input-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  outline: none;
}

.input-field:focus {
  border-color: #4a8cff;
}

.tags-input-wrapper {
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 4px;
  min-height: 38px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tags-list {
  display: contents;
}

.tag-item {
  background: #eff6ff;
  color: #4a8cff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.remove-tag {
  cursor: pointer;
  font-size: 12px;
}

.tag-input {
  border: none;
  outline: none;
  flex: 1;
  min-width: 100px;
  padding: 4px 8px;
  font-size: 14px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}

.btn-cancel, .btn-confirm {
  flex: 1;
  padding: 10px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}

.btn-cancel {
  background: #f1f5f9;
  color: #64748b;
}

.btn-confirm {
  background: #4a8cff;
  color: white;
}
</style>
