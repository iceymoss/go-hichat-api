<template>
  <div class="settings-overlay" @click="closeSettings">
    <div class="settings-modal" @click.stop>
      <!-- 头部 -->
      <div class="modal-header">
        <h3>资料设置</h3>
        <button class="btn-close" @click="closeSettings">
          <i class="icon icon-x"></i>
        </button>
      </div>

      <div class="modal-body">
        <!-- 基础信息设置 -->
        <div class="settings-group">
          <div class="setting-item input-item">
            <span class="label">备注名</span>
            <input 
              v-model="remark" 
              type="text" 
              placeholder="添加备注名"
              class="clean-input"
              @blur="updateRemark"
            >
          </div>
          
          <div class="setting-item tags-item">
            <div class="tags-header">
              <span class="label">标签</span>
              <button class="btn-add-tag-text" @click="showTagInput = !showTagInput" v-if="!showTagInput">
                <i class="icon icon-plus"></i> 添加
              </button>
            </div>
            
            <div class="tags-wrapper">
              <span v-for="tag in displayTags" :key="tag" class="tag-chip">
                {{ tag }}
                <i class="icon icon-x remove-icon" @click="removeTag(tag)"></i>
              </span>
              
              <input 
                v-if="showTagInput"
                v-model="newTag"
                ref="tagInputRef"
                type="text"
                class="tag-input-inline"
                placeholder="输入后回车"
                @keyup.enter="addTag"
                @blur="handleTagInputBlur"
              >
            </div>
          </div>
        </div>

        <!-- 权限与开关 -->
        <div class="settings-group">
          <div class="setting-item">
            <span class="label">朋友权限</span>
            <select v-model="momentsPermission" class="clean-select" @change="updatePermission">
              <option value="all">聊天、朋友圈</option>
              <option value="limited">仅聊天</option>
              <option value="none">屏蔽</option>
            </select>
          </div>
          
          <div class="setting-item switch-item">
            <span class="label">消息免打扰</span>
            <label class="switch">
              <input type="checkbox" v-model="notifications" @change="updateNotifications">
              <span class="slider"></span>
            </label>
          </div>
          
          <div class="setting-item switch-item">
            <span class="label">置顶聊天</span>
            <label class="switch">
              <input type="checkbox" v-model="pinned" @change="updatePinned">
              <span class="slider"></span>
            </label>
          </div>
          
          <div class="setting-item switch-item">
            <span class="label">加入黑名单</span>
            <label class="switch">
              <input type="checkbox" v-model="blacklisted" @change="updateBlacklist">
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <!-- 操作类 -->
        <div class="settings-group actions-group">
          <button class="action-row" @click="shareFriend">
            <span class="label">把该好友推荐给...</span>
            <i class="icon icon-chevron-right"></i>
          </button>
          <button class="action-row" @click="reportFriend">
            <span class="label">投诉</span>
            <i class="icon icon-chevron-right"></i>
          </button>
        </div>
        
        <div class="settings-group danger-group">
          <button class="action-row danger" @click="showDeleteConfirm = true">
            <span class="label">删除好友</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="confirm-overlay" @click="showDeleteConfirm = false">
      <div class="confirm-modal" @click.stop>
        <h3>删除好友</h3>
        <p>将联系人“{{ friend.remark || friend.name }}”删除，同时删除与该联系人的聊天记录。</p>
        <div class="confirm-actions">
          <button class="btn-cancel" @click="showDeleteConfirm = false">取消</button>
          <button class="btn-confirm-delete" @click="confirmDelete">删除</button>
        </div>
      </div>
    </div>
    
    <Toast ref="toastRef" />
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useContactsStore } from '../stores/contacts'
import Toast from './ui/Toast.vue'

const props = defineProps({
  friend: { type: Object, required: true }
})

const emit = defineEmits(['close', 'delete-friend', 'share-friend'])

const contactsStore = useContactsStore()
const toastRef = ref(null)

// 本地状态
const remark = ref('')
const newTag = ref('')
const localTags = ref([]) // 新增：本地标签状态
const showTagInput = ref(false)
const tagInputRef = ref(null)
const notifications = ref(true)
const pinned = ref(false)
const blacklisted = ref(false)
const momentsPermission = ref('all')
const showDeleteConfirm = ref(false)

const friendUid = computed(() => props.friend?.friend_uid || props.friend?.id)

// 辅助函数：权限转换
const getPermissionValue = (p) => {
  if (p === 0 || p === 'all') return 'all'
  if (p === 1 || p === 'limited') return 'limited'
  if (p === 2 || p === 'none') return 'none'
  return 'all'
}
const getPermissionBackend = (p) => {
  if (p === 'all') return 0
  if (p === 'limited') return 1
  if (p === 'none') return 2
  return 0
}

const displayTags = computed(() => localTags.value) // 修改：使用本地状态

// 初始化数据
watch(() => props.friend, (f) => {
  if (f) {
    remark.value = f.remark || ''
    localTags.value = [...(f.friend_tags || [])] // 使用好友标签（关系维度）
    // "消息免打扰"开关: true=免打扰(通知关闭), false=正常(通知开启)
    notifications.value = f.notify_enabled !== undefined ? !f.notify_enabled : false
    pinned.value = f.pinned || false
    blacklisted.value = f.blacklisted || false
    momentsPermission.value = getPermissionValue(f.moments_permission)
  }
}, { immediate: true })

// 聚焦标签输入框
watch(showTagInput, (val) => {
  if (val) {
    nextTick(() => tagInputRef.value?.focus())
  }
})

const closeSettings = () => emit('close')

// --- Actions ---

const updateRemark = async () => {
  if (remark.value !== props.friend.remark) {
    try {
      await contactsStore.updateFriendRemark(friendUid.value, remark.value)
      toastRef.value?.show('备注已更新')
    } catch (e) {
      toastRef.value?.show('更新失败', 'error')
    }
  }
}

const addTag = async () => {
  if (!newTag.value.trim()) {
    showTagInput.value = false
    return
  }
  // 使用本地 localTags
  const tags = [...localTags.value, newTag.value.trim()]
  try {
    await contactsStore.updateFriendTags(friendUid.value, tags)
    localTags.value = tags // 立即更新本地状态
    newTag.value = ''
    // showTagInput.value = false 
  } catch (e) {
    toastRef.value?.show('添加标签失败', 'error')
  }
}

const handleTagInputBlur = () => {
  if (newTag.value) addTag()
  else showTagInput.value = false
}

const removeTag = async (tag) => {
  // 使用本地 localTags
  const tags = localTags.value.filter(t => t !== tag)
  try {
    await contactsStore.updateFriendTags(friendUid.value, tags)
    localTags.value = tags // 立即更新本地状态
  } catch (e) {
    toastRef.value?.show('删除标签失败', 'error')
  }
}

const updatePermission = async () => {
  try {
    await contactsStore.updateMomentsPermission(friendUid.value, getPermissionBackend(momentsPermission.value))
    toastRef.value?.show('权限已更新')
  } catch (e) {
    toastRef.value?.show('操作失败', 'error')
  }
}

const updateNotifications = async () => {
  try {
    // notifications 是"免打扰"开关，取反后得到 API 需要的 enabled 值
    await contactsStore.updateFriendNotification(friendUid.value, !notifications.value)
  } catch (e) {
    notifications.value = !notifications.value
    toastRef.value?.show('操作失败', 'error')
  }
}

const updatePinned = async () => {
  try {
    await contactsStore.updateFriendPin(friendUid.value, pinned.value)
  } catch (e) {
    pinned.value = !pinned.value
    toastRef.value?.show('操作失败', 'error')
  }
}

const updateBlacklist = async () => {
  try {
    await contactsStore.blockFriend(friendUid.value, blacklisted.value)
    toastRef.value?.show(blacklisted.value ? '已加入黑名单' : '已移出黑名单')
  } catch (e) {
    blacklisted.value = !blacklisted.value
    toastRef.value?.show('操作失败', 'error')
  }
}

const shareFriend = () => {
  emit('share-friend', props.friend)
  closeSettings()
}

const reportFriend = () => {
  toastRef.value?.show('举报功能开发中...', 'warning')
}

const confirmDelete = () => {
  // 使用 friend_uid 而不是 id，因为 id 是好友关系ID，friend_uid 才是用户ID
  const friendUid = props.friend.friend_uid || props.friend.id
  emit('delete-friend', friendUid)
  showDeleteConfirm.value = false
}
</script>

<style scoped>
.settings-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(15, 23, 42, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.settings-modal {
  width: 380px;
  background: #f8fafc;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.12);
}

.modal-header {
  background: white;
  padding: 14px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e2e8f0;
}
.modal-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}
.btn-close {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: #94a3b8;
  border-radius: 6px;
  transition: all 0.15s;
}
.btn-close:hover { background: #f1f5f9; color: #64748b; }

.modal-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 70vh;
  overflow-y: auto;
}

.settings-group {
  background: white;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  font-size: 13px;
}
.setting-item:last-child { border-bottom: none; }

.label { color: #1e293b; font-weight: 500; }

.clean-input {
  border: none;
  text-align: right;
  outline: none;
  font-size: 13px;
  color: #64748b;
  width: 150px;
}
.clean-select {
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: #64748b;
  text-align: right;
  direction: rtl;
}

/* 标签样式 */
.tags-item {
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}
.tags-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.btn-add-tag-text {
  border: none;
  background: none;
  color: #2563eb;
  font-size: 12px;
  cursor: pointer;
  font-weight: 500;
}
.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.tag-chip {
  background: #eff6ff;
  color: #2563eb;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.remove-icon {
  font-size: 12px;
  cursor: pointer;
  color: #93c5fd;
}
.remove-icon:hover { color: #2563eb; }
.tag-input-inline {
  border: 1px solid #2563eb;
  border-radius: 4px;
  padding: 3px 6px;
  font-size: 12px;
  outline: none;
  width: 80px;
  color: #1e293b;
}

/* 开关样式 */
.switch {
  position: relative;
  width: 36px;
  height: 20px;
}
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #cbd5e1;
  transition: .2s;
  border-radius: 20px;
}
.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: .2s;
  border-radius: 50%;
}
input:checked + .slider { background-color: #2563eb; }
input:checked + .slider:before { transform: translateX(16px); }

/* 操作按钮组 */
.action-row {
  width: 100%;
  background: white;
  border: none;
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.15s;
  border-bottom: 1px solid #f1f5f9;
}
.action-row:last-child { border-bottom: none; }
.action-row:hover { background: #f8fafc; }
.action-row .icon { color: #cbd5e1; font-size: 14px; }
.action-row.danger {
  justify-content: center;
  color: #dc2626;
  font-weight: 500;
}
.action-row.danger:hover { background: #fef2f2; }

/* 确认弹窗 */
.confirm-overlay {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(15, 23, 42, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2100;
}
.confirm-modal {
  background: white;
  width: 280px;
  padding: 24px;
  border-radius: 12px;
  text-align: center;
  box-shadow: 0 20px 40px rgba(0,0,0,0.12);
}
.confirm-modal h3 { margin: 0 0 8px 0; font-size: 15px; font-weight: 600; color: #0f172a; }
.confirm-modal p { margin: 0 0 20px 0; font-size: 13px; color: #64748b; line-height: 1.5; }
.confirm-actions { display: flex; gap: 10px; }
.confirm-actions button {
  flex: 1;
  padding: 8px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s;
}
.btn-cancel { background: #f1f5f9; color: #475569; }
.btn-cancel:hover { background: #e2e8f0; }
.btn-confirm-delete { background: #dc2626; color: white; }
.btn-confirm-delete:hover { background: #b91c1c; }
</style>