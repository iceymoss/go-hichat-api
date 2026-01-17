<template>
  <div class="new-groups-wrapper">
    <div class="new-groups-container">
      <div class="header">
        <div class="title-group">
          <div class="title">群聊申请</div>
          <!-- Class 分类：我收到的申请 vs 我发起的申请 -->
          <div class="class-tabs">
            <span 
              :class="{ active: classFilter === 'received' }" 
              @click="classFilter = 'received'"
            >我收到的</span>
            <span 
              :class="{ active: classFilter === 'sent' }" 
              @click="classFilter = 'sent'"
            >我发起的</span>
          </div>
        </div>
        <!-- Type 分类：待处理、已通过、已拒绝、已忽略 -->
        <div class="type-tabs">
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
            <i class="icon icon-users"></i>
          </div>
          <p>暂无群聊申请</p>
        </div>

        <div 
          v-for="req in filteredRequests" 
          :key="req.id" 
          class="request-item"
          @click="showDetail(req)"
        >
          <!-- 我收到的申请：显示申请人的头像和群名称 -->
          <!-- 我发起的申请：显示群的头像和群名称 -->
          <img :src="classFilter === 'received' ? req.avatar : req.group_avatar" alt="头像" class="avatar">
          <div class="info">
            <div class="name-row">
              <span class="name">{{ classFilter === 'received' ? req.name : req.group_name }}</span>
              <span class="tag-label" v-if="req.status === 'pending' && req.requestType === 'received'">申请中</span>
              <span class="tag-label sent" v-if="req.requestType === 'sent'">我发起的</span>
            </div>
            <div class="message">
              <span v-if="req.requestType === 'sent'">我：{{ req.req_msg || '申请加入群聊' }}</span>
              <span v-else>{{ req.req_msg || '申请加入群聊' }}</span>
              <span v-if="req.requestType === 'received' && req.group_name" class="group-name"> · {{ req.group_name }}</span>
              <span v-if="req.requestType === 'sent' && req.name" class="user-name"> · {{ req.name }}</span>
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

    <!-- 处理申请弹窗（简化的，不需要备注和标签） -->
    <div v-if="showHandleModal" class="handle-modal-overlay" @click="closeHandleModal">
      <div class="handle-modal" @click.stop>
        <div class="modal-header">
          <h3>通过群聊申请</h3>
          <button class="btn-close" @click="closeHandleModal">
            <i class="icon icon-x"></i>
          </button>
        </div>
        
        <div class="user-preview">
          <img :src="currentReq.avatar" alt="头像" class="preview-avatar">
          <div class="preview-info">
            <div class="preview-name">{{ currentReq.name }}</div>
            <div class="preview-msg" v-if="currentReq.group_name">申请加入：{{ currentReq.group_name }}</div>
            <div class="preview-msg">验证消息：{{ currentReq.req_msg || '无' }}</div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-cancel" @click="closeHandleModal">取消</button>
          <button class="btn-confirm" @click="confirmAccept">同意</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useContactsStore } from '../stores/contacts'

const emit = defineEmits(['select-group'])

const contactsStore = useContactsStore()
// Class 分类：'received' 表示我收到的申请，'sent' 表示我发起的申请
const classFilter = ref('received')
// Type 分类：'0' 表示待处理，'1' 表示已通过，'2' 表示已拒绝，'3' 表示已忽略
const typeFilter = ref('0') // 默认显示"待处理"
const showHandleModal = ref(false)
const currentReq = ref(null)

// 记录已请求的数据，避免重复请求
const loadedData = ref({
  received: new Set(), // 我收到的申请已加载的type
  sent: new Set()      // 我发起的申请已加载的type
})

// 根据筛选器请求数据（只请求特定的type）
const loadDataByFilter = async () => {
  const classType = classFilter.value === 'received' ? 2 : 1
  const loadedSet = classFilter.value === 'received' ? loadedData.value.received : loadedData.value.sent
  
  // 只请求当前选中的type的数据（如果还未加载）
  // 注意：我收到的申请可以不传groupId查看所有群的申请
  if (!loadedSet.has(typeFilter.value)) {
    await contactsStore.fetchGroupRequestsByType(parseInt(typeFilter.value), classType, null)
    loadedSet.add(typeFilter.value)
  }
}

// 监听筛选器变化，按需请求数据
watch([classFilter, typeFilter], async () => {
  try {
    await loadDataByFilter()
  } catch (error) {
    console.error('加载数据失败:', error)
  }
}, { immediate: false })

// 组件挂载时加载初始数据
onMounted(async () => {
  try {
    // 加载初始筛选器对应的数据（默认：我收到的申请，待处理）
    await loadDataByFilter()
  } catch (error) {
    console.error('初始化失败:', error)
  }
})

const filteredRequests = computed(() => {
  // 根据 classFilter 选择对应的数据源
  let sourceData = {}
  if (classFilter.value === 'received') {
    // 我收到的申请
    sourceData = contactsStore.receivedGroupRequests || {}
  } else {
    // 我发起的申请
    sourceData = contactsStore.sentGroupRequests || {}
  }
  
  // 根据 typeFilter 获取对应的数据
  let sourceList = sourceData[typeFilter.value] || []
  
  // 添加 requestType 标识
  const requestType = classFilter.value === 'received' ? 'received' : 'sent'
  sourceList = sourceList.map(req => ({ ...req, requestType }))
  
  // 按时间倒序排序（最新的在前）
  sourceList.sort((a, b) => (b.req_time || 0) - (a.req_time || 0))
  
  return sourceList
})

const showDetail = (req) => {
  // 如果已同意，可以跳转到群详情
  if (req.status === 'accepted' && req.group_id) {
    emit('select-group', { id: req.group_id })
  }
}

const handleAccept = (req) => {
  currentReq.value = req
  showHandleModal.value = true
}

const closeHandleModal = () => {
  showHandleModal.value = false
  currentReq.value = null
}

const confirmAccept = async () => {
  if (!currentReq.value) return
  
  try {
    // 处理群聊申请
    // group_req_id 是群聊申请记录的ID，需要转换为number类型传递给API
    await contactsStore.handleGroupRequest(
      Number(currentReq.value.group_req_id),
      currentReq.value.group_id,
      true
    )
    
    // 刷新申请列表
    const classType = classFilter.value === 'received' ? 2 : 1
    await contactsStore.fetchGroupRequestsByType(parseInt(typeFilter.value), classType, null)
    
    closeHandleModal()
  } catch (error) {
    console.error('处理群聊申请失败:', error)
    alert('操作失败，请重试')
  }
}
</script>

<style scoped>
.new-groups-wrapper {
  height: 100%;
  width: 100%;
  display: flex;
  justify-content: center;
  background: #f5f5f5;
}

.new-groups-container {
  width: 100%;
  max-width: 1000px;
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

.group-name, .user-name {
  color: #666;
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
  margin-top: 2px;
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

