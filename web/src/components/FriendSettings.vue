<template>
  <div class="settings-overlay" @click="closeSettings">
    <div class="settings-modal" @click.stop>
      <!-- 头部 -->
      <div class="modal-header">
        <h3>好友设置</h3>
        <button class="btn-close" @click="closeSettings">
          <i class="icon icon-x"></i>
        </button>
      </div>

      <!-- 好友信息 -->
      <div class="friend-info">
        <img :src="friend.avatar" alt="头像" class="avatar">
        <div class="info">
          <h4>{{ friend.remark || friend.name }}</h4>
          <p>账号ID: {{ friend.id }}</p>
        </div>
      </div>

      <!-- 设置选项 -->
      <div class="settings-options">
        <!-- 备注设置 -->
        <div class="setting-item">
          <div class="setting-label">
            <i class="icon icon-edit"></i>
            <span>好友备注</span>
          </div>
          <div class="setting-content">
            <input 
              v-model="remark" 
              type="text" 
              placeholder="设置好友备注"
              class="remark-input"
              @blur="updateRemark"
            >
          </div>
        </div>

        <!-- 标签设置 -->
        <div class="setting-item">
          <div class="setting-label">
            <i class="icon icon-tag"></i>
            <span>好友标签</span>
          </div>
          <div class="setting-content">
            <div class="tags-display">
              <span v-for="tag in friend.tags" :key="tag" class="tag">
                {{ tag }}
                <button class="tag-remove" @click="removeTag(tag)">
                  <i class="icon icon-x"></i>
                </button>
              </span>
            </div>
            <div class="add-tag">
              <input 
                v-model="newTag" 
                type="text" 
                placeholder="添加标签"
                class="tag-input"
                @keyup.enter="addTag"
              >
              <button class="btn-add-tag" @click="addTag">
                <i class="icon icon-plus"></i>
              </button>
            </div>
          </div>
        </div>

        <!-- 消息设置 -->
        <div class="setting-item">
          <div class="setting-label">
            <i class="icon icon-bell"></i>
            <span>消息通知</span>
          </div>
          <div class="setting-content">
            <label class="switch">
              <input type="checkbox" v-model="notifications" @change="updateNotifications">
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <!-- 置顶聊天 -->
        <div class="setting-item">
          <div class="setting-label">
            <i class="icon icon-pin"></i>
            <span>置顶聊天</span>
          </div>
          <div class="setting-content">
            <label class="switch">
              <input type="checkbox" v-model="pinned" @change="updatePinned">
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <!-- 静音设置 -->
        <div class="setting-item">
          <div class="setting-label">
            <i class="icon icon-volume-off"></i>
            <span>静音消息</span>
          </div>
          <div class="setting-content">
            <label class="switch">
              <input type="checkbox" v-model="muted" @change="updateMuted">
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <!-- 黑名单 -->
        <div class="setting-item">
          <div class="setting-label">
            <i class="icon icon-shield"></i>
            <span>加入黑名单</span>
          </div>
          <div class="setting-content">
            <label class="switch">
              <input type="checkbox" v-model="blacklisted" @change="updateBlacklist">
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <!-- 分享好友 -->
        <div class="setting-item clickable" @click="shareFriend">
          <div class="setting-label">
            <i class="icon icon-share"></i>
            <span>分享好友</span>
          </div>
          <div class="setting-content">
            <i class="icon icon-chevron-right"></i>
          </div>
        </div>

        <!-- 举报 -->
        <div class="setting-item clickable" @click="reportFriend">
          <div class="setting-label">
            <i class="icon icon-warning"></i>
            <span>举报好友</span>
          </div>
          <div class="setting-content">
            <i class="icon icon-chevron-right"></i>
          </div>
        </div>
      </div>

      <!-- 危险操作 -->
      <div class="danger-zone">
        <h4>危险操作</h4>
        
        <!-- 删除好友 -->
        <div class="danger-item" @click="showDeleteConfirm = true">
          <div class="danger-label">
            <i class="icon icon-trash"></i>
            <span>删除好友</span>
          </div>
          <div class="danger-content">
            <i class="icon icon-chevron-right"></i>
          </div>
        </div>
      </div>

      <!-- 删除确认弹窗 -->
      <div v-if="showDeleteConfirm" class="delete-overlay" @click="showDeleteConfirm = false">
        <div class="delete-modal" @click.stop>
          <div class="delete-header">
            <i class="icon icon-warning"></i>
            <h4>删除好友</h4>
          </div>
          <div class="delete-content">
            <p>确定要删除好友 "{{ friend.remark || friend.name }}" 吗？</p>
            <p class="warning-text">删除后将无法恢复，聊天记录也会被删除。</p>
          </div>
          <div class="delete-actions">
            <button class="btn-cancel" @click="showDeleteConfirm = false">
              取消
            </button>
            <button class="btn-delete" @click="confirmDelete">
              删除
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'

const props = defineProps({
  friend: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'update-friend', 'delete-friend'])

// 响应式数据
const remark = ref(props.friend.remark || '')
const newTag = ref('')
const notifications = ref(true)
const pinned = ref(false)
const muted = ref(false)
const blacklisted = ref(false)
const showDeleteConfirm = ref(false)

// 方法
const closeSettings = () => {
  emit('close')
}

const updateRemark = () => {
  const updatedFriend = { ...props.friend, remark: remark.value }
  emit('update-friend', updatedFriend)
}

const addTag = () => {
  if (newTag.value.trim()) {
    const updatedFriend = { 
      ...props.friend, 
      tags: [...(props.friend.tags || []), newTag.value.trim()]
    }
    emit('update-friend', updatedFriend)
    newTag.value = ''
  }
}

const removeTag = (tagToRemove) => {
  const updatedFriend = {
    ...props.friend,
    tags: props.friend.tags.filter(tag => tag !== tagToRemove)
  }
  emit('update-friend', updatedFriend)
}

const updateNotifications = () => {
  const updatedFriend = { ...props.friend, notifications: notifications.value }
  emit('update-friend', updatedFriend)
}

const updatePinned = () => {
  const updatedFriend = { ...props.friend, pinned: pinned.value }
  emit('update-friend', updatedFriend)
}

const updateMuted = () => {
  const updatedFriend = { ...props.friend, muted: muted.value }
  emit('update-friend', updatedFriend)
}

const updateBlacklist = () => {
  const updatedFriend = { ...props.friend, blacklisted: blacklisted.value }
  emit('update-friend', updatedFriend)
}

const shareFriend = () => {
  // 实现分享好友功能
  console.log('分享好友:', props.friend)
}

const reportFriend = () => {
  // 实现举报好友功能
  console.log('举报好友:', props.friend)
}

const confirmDelete = () => {
  emit('delete-friend', props.friend.id)
  showDeleteConfirm.value = false
}
</script>

<style scoped>
.settings-overlay {
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
  backdrop-filter: blur(4px);
}

.settings-modal {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 480px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
}

.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: #f8fafc;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-close:hover {
  background: #e2e8f0;
  color: #475569;
}

.friend-info {
  display: flex;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.friend-info .avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  object-fit: cover;
  margin-right: 16px;
}

.friend-info .info h4 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}

.friend-info .info p {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.settings-options {
  padding: 0;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid #f1f5f9;
  transition: background-color 0.2s;
}

.setting-item:hover {
  background-color: #f8fafc;
}

.setting-item.clickable {
  cursor: pointer;
}

.setting-label {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  color: #1e293b;
}

.setting-label .icon {
  font-size: 18px;
  color: #64748b;
}

.setting-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.remark-input, .tag-input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.remark-input:focus, .tag-input:focus {
  border-color: #4a8cff;
}

.tags-display {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
}

.tag {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  color: white;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.tag-remove {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 2px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
}

.tag-remove:hover {
  background: rgba(255, 255, 255, 0.2);
}

.add-tag {
  display: flex;
  gap: 8px;
}

.tag-input {
  flex: 1;
}

.btn-add-tag {
  padding: 8px 12px;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-add-tag:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.3);
}

/* 开关样式 */
.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #cbd5e1;
  transition: 0.3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.danger-zone {
  padding: 20px 24px;
  border-top: 1px solid #e2e8f0;
  background: #fef2f2;
}

.danger-zone h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #dc2626;
}

.danger-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  cursor: pointer;
  transition: background-color 0.2s;
}

.danger-item:hover {
  background-color: #fee2e2;
  margin: 0 -12px;
  padding: 12px;
  border-radius: 8px;
}

.danger-label {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  color: #dc2626;
}

.danger-label .icon {
  font-size: 18px;
}

.danger-content .icon {
  color: #dc2626;
}

/* 删除确认弹窗 */
.delete-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.delete-modal {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 400px;
  width: 90%;
  text-align: center;
}

.delete-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 16px;
}

.delete-header .icon {
  font-size: 24px;
  color: #dc2626;
}

.delete-header h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #dc2626;
}

.delete-content {
  margin-bottom: 24px;
}

.delete-content p {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #475569;
}

.warning-text {
  color: #dc2626 !important;
  font-size: 12px !important;
}

.delete-actions {
  display: flex;
  gap: 12px;
}

.btn-cancel, .btn-delete {
  flex: 1;
  padding: 10px 16px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel {
  background: #f1f5f9;
  color: #64748b;
}

.btn-cancel:hover {
  background: #e2e8f0;
}

.btn-delete {
  background: #dc2626;
  color: white;
}

.btn-delete:hover {
  background: #b91c1c;
}
</style> 