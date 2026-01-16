<template>
  <div class="new-friends-wrapper">
    <div class="new-friends-container">
      <div class="header">
        <div class="title-group">
          <div class="title">新的朋友</div>
          <!-- Class 分类：我收到的申请 vs 我发起的申请 -->
          <div class="class-tabs">
            <span 
              :class="{ active: classFilter === 'received' }" 
              @click="classFilter = 'received'"
            >申请添加</span>
            <span 
              :class="{ active: classFilter === 'sent' }" 
              @click="classFilter = 'sent'"
            >我发起的</span>
          </div>
        </div>
        <!-- Type 分类：全部、待处理、已通过、已拒绝、已忽略 -->
        <div class="type-tabs">
          <span 
            :class="{ active: typeFilter === 'all' }" 
            @click="typeFilter = 'all'"
          >全部</span>
          <span 
            :class="{ active: typeFilter === '0' }" 
            @click="typeFilter = '0'"
          >待处理</span>
          <span 
            :class="{ active: typeFilter === '1' }" 
            @click="typeFilter = '1'"
          >已通过</span>
          <span 
            :class="{ active: typeFilter === '2' }" 
            @click="typeFilter = '2'"
          >已拒绝</span>
          <span 
            :class="{ active: typeFilter === '3' }" 
            @click="typeFilter = '3'"
          >已忽略</span>
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
              <span class="tag-label" v-if="req.status === 'pending' && req.requestType === 'received'">申请中</span>
              <span class="tag-label sent" v-if="req.requestType === 'sent'">我发起的</span>
            </div>
            <div class="message">
              <span v-if="req.requestType === 'sent'">我：{{ req.req_msg || '请求添加为好友' }}</span>
              <span v-else>{{ req.req_msg || '请求添加你为好友' }}</span>
            </div>
          </div>
          <div class="action-area">
            <button v-if="req.status === 'pending' && req.requestType === 'received'" class="btn-accept" @click.stop="handleAccept(req)">
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

    <!-- 用户资料卡片 -->
    <UserProfileCard
      v-if="showProfileCard && profileCardUserId"
      :userId="profileCardUserId"
      :friendRequest="currentFriendRequest"
      @close="closeProfileCard"
      @open-accept-dialog="handleOpenAcceptDialog"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { socialApi } from '../utils/api'
import UserProfileCard from './UserProfileCard.vue'

const emit = defineEmits(['select-friend'])

const contactsStore = useContactsStore()
// Class 分类：'received' 表示我收到的申请，'sent' 表示我发起的申请
const classFilter = ref('received')
// Type 分类：'all' 表示全部，'0' 表示待处理，'1' 表示已通过，'2' 表示已拒绝，'3' 表示已忽略
const typeFilter = ref('all')
const showHandleModal = ref(false)
const currentReq = ref(null)
const newTag = ref('')
const handleForm = ref({
  remark: '',
  tags: []
})

// 用户资料卡片相关
const showProfileCard = ref(false)
const profileCardUserId = ref(null)
const currentFriendRequest = ref(null)

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
  // 根据 classFilter 选择对应的数据源
  let sourceData = {}
  if (classFilter.value === 'received') {
    // 我收到的申请
    sourceData = contactsStore.friendRequests || {}
  } else {
    // 我发起的申请
    sourceData = contactsStore.sentFriendRequests || {}
  }
  
  // 根据 typeFilter 获取对应的数据
  let sourceList = []
  if (typeFilter.value === 'all') {
    // 全部：合并所有 type 的数据
    sourceList = [
      ...(sourceData['0'] || []),
      ...(sourceData['1'] || []),
      ...(sourceData['2'] || []),
      ...(sourceData['3'] || [])
    ]
  } else {
    // 特定 type：只获取对应 type 的数据
    sourceList = sourceData[typeFilter.value] || []
  }
  
  // 添加 requestType 标识
  const requestType = classFilter.value === 'received' ? 'received' : 'sent'
  sourceList = sourceList.map(req => ({ ...req, requestType }))
  
  // 按时间倒序排序（最新的在前）
  sourceList.sort((a, b) => (b.req_time || 0) - (a.req_time || 0))
  
  return sourceList
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
  // 如果是待处理状态，显示用户资料卡片
  if (req.status === 'pending') {
    showUserProfileCard(req)
  } 
  // 如果是已同意状态，跳转到好友详情页
  else if (req.status === 'accepted') {
    // 确定用户ID：
    // 根据数据库schema：user_id是申请人，req_uid是被申请人
    // - 如果是"我收到的申请"（requestType='received'）：req_uid是我，user_id是对方，应该使用user_id
    // - 如果是"我发起的申请"（requestType='sent'）：user_id是我，req_uid是对方，应该使用req_uid
    const friendId = req.requestType === 'received' ? req.user_id : req.req_uid
    if (friendId) {
      // 通知父组件选择该好友
      emit('select-friend', { friend_uid: friendId, id: friendId })
    }
  }
}

// 显示用户资料卡片
const showUserProfileCard = (req) => {
  // 确定用户ID：
  // 根据数据库schema：user_id是申请人，req_uid是被申请人
  // - 如果是"我收到的申请"（requestType='received'）：req_uid是我，user_id是对方，应该使用user_id
  // - 如果是"我发起的申请"（requestType='sent'）：user_id是我，req_uid是对方，应该使用req_uid
  const userId = req.requestType === 'received' ? req.user_id : req.req_uid
  
  if (userId) {
    profileCardUserId.value = userId
    currentFriendRequest.value = req // 保存申请信息，用于显示接受申请按钮
    showProfileCard.value = true
  }
}

// 关闭用户资料卡片
const closeProfileCard = () => {
  showProfileCard.value = false
  profileCardUserId.value = null
  currentFriendRequest.value = null
}

// 处理打开接受申请弹窗事件（从用户资料卡片触发）
const handleOpenAcceptDialog = (friendRequest) => {
  // 关闭用户资料卡片
  closeProfileCard()
  
  // 打开处理申请的弹窗
  handleAccept(friendRequest)
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

.class-tabs {
  display: flex;
  background: #f5f5f5;
  padding: 2px;
  border-radius: 4px;
  margin-left: 20px;
}

.class-tabs span {
  font-size: 13px;
  color: #666;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 4px;
  transition: all 0.2s;
}

.class-tabs span.active {
  color: #333;
  background: #fff;
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.type-tabs {
  display: flex;
  gap: 8px;
  padding: 12px 30px;
  border-bottom: 1px solid #f0f0f0;
  background: #fff;
}

.type-tabs span {
  font-size: 13px;
  color: #666;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 4px;
  transition: all 0.2s;
}

.type-tabs span.active {
  color: #4a8cff;
  background: #e6f7ff;
  font-weight: 500;
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

.tag-label.sent {
  background: #e6f7ff;
  color: #1890ff;
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
