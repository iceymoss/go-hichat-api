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
  background: var(--bg);
}

.new-groups-container {
  width: 100%;
  max-width: 960px;
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

.header {
  padding: 0 24px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  height: 52px;
  flex-shrink: 0;
}

.title-group {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title {
  font-size: 15px;
  font-weight: 600;
  color: var(--fg);
  font-family: var(--font-display);
}

.class-tabs {
  display: flex;
  background: var(--card);
  padding: 2px;
  border-radius: var(--radius-sm);
}

.class-tabs span {
  font-size: 12px;
  color: var(--muted);
  cursor: pointer;
  padding: 4px 14px;
  border-radius: 4px;
  transition: all 0.15s;
  font-weight: 500;
  font-family: var(--font-ui);
}

.class-tabs span.active {
  color: #fff;
  background: var(--card-hover);
  box-shadow: 0 1px 2px rgba(0,0,0,0.2);
}

.type-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 24px;
  border-bottom: 1px solid var(--border-light);
}

.type-tabs span {
  font-size: 12px;
  color: var(--muted);
  cursor: pointer;
  padding: 5px 12px;
  border-radius: var(--radius-sm);
  transition: all 0.15s;
  font-weight: 500;
  font-family: var(--font-ui);
}

.type-tabs span:hover {
  background: var(--card);
}

.type-tabs span.active {
  color: var(--accent);
  background: var(--accent-muted);
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
  padding-top: 80px;
  color: var(--muted);
}

.empty-icon-circle {
  width: 56px;
  height: 56px;
  background: var(--card);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}

.empty-icon-circle .icon {
  font-size: 24px;
  color: var(--muted);
}

.empty-state p {
  font-size: 13px;
  font-family: var(--font-ui);
}

.request-item {
  display: flex;
  padding: 12px 24px;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  transition: background 0.15s;
  align-items: center;
}

.request-item:hover {
  background: var(--card);
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  margin-right: 12px;
  object-fit: cover;
}

.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name {
  font-size: 14px;
  font-weight: 500;
  color: var(--fg);
  font-family: var(--font-ui);
}

.tag-label {
  font-size: 10px;
  background: var(--accent-muted);
  color: var(--accent);
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 500;
}

.tag-label.sent {
  background: var(--accent-muted);
  color: var(--accent);
}

.message {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-name, .user-name {
  color: var(--muted);
}

.action-area {
  margin-left: 16px;
  flex-shrink: 0;
}

.btn-accept {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 6px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.15s;
  font-family: var(--font-ui);
}

.btn-accept:hover {
  background: var(--accent-hover);
}

.status-text {
  font-size: 12px;
  color: var(--muted);
}

/* Modal Styles */
.handle-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.handle-modal {
  background: var(--card);
  border-radius: var(--radius-lg);
  width: 400px;
  padding: 24px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
  border: 1px solid var(--border);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--fg);
  font-family: var(--font-display);
}

.btn-close {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--muted);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: all 0.15s;
}
.btn-close:hover { background: var(--card-hover); color: var(--fg); }

.user-preview {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  padding: 12px;
  background: var(--input-bg);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
}

.preview-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  margin-right: 12px;
}

.preview-name {
  font-weight: 500;
  color: var(--fg);
  font-size: 14px;
  font-family: var(--font-ui);
}

.preview-msg {
  font-size: 12px;
  color: var(--muted);
  margin-top: 2px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 24px;
}

.btn-cancel, .btn-confirm {
  flex: 1;
  padding: 9px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
  font-family: var(--font-ui);
}

.btn-cancel {
  background: var(--border);
  color: var(--fg);
}
.btn-cancel:hover { background: var(--card-hover); }

.btn-confirm {
  background: var(--accent);
  color: #fff;
}
.btn-confirm:hover { background: var(--accent-hover); }
</style>

