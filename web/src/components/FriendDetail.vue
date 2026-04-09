<template>
  <div class="friend-detail-wrapper">
    <!-- 空状态 -->
    <div v-if="!friend" class="empty-state">
      <div class="empty-icon-circle">
        <i class="icon icon-message-square"></i>
      </div>
      <p class="empty-text">选择联系人查看详情</p>
    </div>

    <!-- 个人资料主内容 -->
    <div v-else class="profile-container">
      <div class="profile-inner">
        <!-- 1. 头部区域：头像 | 信息 | 设置 -->
        <div class="profile-header">
          <div class="avatar-box">
            <img :src="friend.avatar" alt="头像" class="avatar" @error="handleAvatarError">
          </div>
          <div class="header-info">
            <div class="primary-name">
              <span class="remark">{{ friend.remark || friend.nickname || '未设置昵称' }}</span>
              <i v-if="friend.sex" class="gender-icon" :class="friend.sex === 1 ? 'icon-male' : 'icon-female'"></i>
            </div>
            <div class="sub-info" v-if="friend.remark && friend.remark.trim()">昵称：{{ friend.nickname }}</div>
            <div class="sub-info">微信号：{{ friend.friend_uid || friend.id }}</div>
            <div class="sub-info" v-if="friend.phone">手机号：{{ friend.phone }}</div>
          </div>
          <button class="btn-settings" @click="showSettings = true" title="资料设置">
            <i class="icon icon-settings"></i>
          </button>
        </div>

        <div class="divider"></div>

        <!-- 2. 信息列表区域 -->
        <div class="info-list">
          <!-- 备注 -->
          <div class="info-item clickable" @click="handleSetRemark">
            <span class="label">备注</span>
            <div class="content">{{ friend.remark || '点击设置备注' }}</div>
            <i class="icon icon-chevron-right arrow"></i>
          </div>

          <!-- 标签 & 地区 -->
          <div class="info-item">
            <span class="label">标签</span>
            <div class="content tags-wrapper">
              <template v-if="friend.friend_tags && friend.friend_tags.length">
                <span v-for="tag in friend.friend_tags" :key="tag" class="tag-badge">{{ tag }}</span>
              </template>
              <span v-else class="placeholder">无标签</span>
            </div>
          </div>
          
          <div class="info-item" v-if="friend.location">
            <span class="label">地区</span>
            <div class="content">{{ friend.location }}</div>
          </div>

          <div class="info-item" v-if="friend.signature">
            <span class="label">个性签名</span>
            <div class="content signature">{{ friend.signature }}</div>
          </div>

          <!-- 朋友圈 -->
          <div class="info-item clickable moments-item" @click="viewMoments">
            <span class="label">朋友圈</span>
            <div class="content moments-preview">
              <div class="thumbs">
                <div v-for="i in 3" :key="i" class="thumb-img"></div>
              </div>
            </div>
            <i class="icon icon-chevron-right arrow"></i>
          </div>
        </div>

        <div class="divider"></div>

        <!-- 3. 底部操作按钮区域 -->
        <div class="action-footer">
          <button class="btn-block btn-primary" @click="sendMessage">
            <i class="icon icon-message-circle"></i>
            <span>发消息</span>
          </button>
          
          <button class="btn-block btn-default" @click="audioCall">
            <i class="icon icon-phone"></i>
            <span>语音通话</span>
          </button>
          
          <button class="btn-block btn-default" @click="videoCall">
            <i class="icon icon-video"></i>
            <span>视频通话</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 弹窗组件 -->
    <FriendSettings 
      v-if="showSettings" 
      :friend="friend"
      @close="showSettings = false"
      @update-friend="handleUpdateFriend"
      @delete-friend="handleDeleteFriend"
      @share-friend="handleShareFriend"
    />
    
    <Toast ref="toastRef" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useContactsStore } from '../stores/contacts'
import FriendSettings from './FriendSettings.vue'
import Toast from './ui/Toast.vue'

const props = defineProps({
  friend: { type: Object, default: null }
})

const emit = defineEmits(['send-message', 'audio-call', 'video-call', 'view-moments'])

const contactsStore = useContactsStore()
const showSettings = ref(false)
const toastRef = ref(null)

const handleAvatarError = (event) => {
  const seed = props.friend?.friend_uid || props.friend?.id || 'default'
  event.target.src = `https://api.dicebear.com/7.x/personas/svg?seed=${seed}`
}

const handleSetRemark = async () => {
  if (!props.friend) return
  const currentRemark = props.friend.remark || ''
  const newRemark = prompt('设置备注名', currentRemark)
  if (newRemark !== null && newRemark !== currentRemark) {
    try {
      const uid = props.friend.friend_uid || props.friend.id
      await contactsStore.updateFriendRemark(uid, newRemark)
      toastRef.value?.show('备注已更新')
    } catch (e) {
      toastRef.value?.show('更新失败', 'error')
    }
  }
}

const sendMessage = () => emit('send-message', props.friend)
const audioCall = () => emit('audio-call', props.friend)
const videoCall = () => emit('video-call', props.friend)
const viewMoments = () => emit('view-moments', props.friend)

const handleUpdateFriend = () => {}
const handleDeleteFriend = async (friendId) => {
  const uid = friendId || props.friend?.friend_uid || props.friend?.id
  if (!uid) return
  await contactsStore.deleteFriend(uid)
  showSettings.value = false
}
const handleShareFriend = async (f) => {
  const uid = f?.friend_uid || f?.id
  if (!uid) return
  try {
    await navigator.clipboard.writeText(`[好友推荐] ${f.nickname} (${uid})`)
    toastRef.value?.show('已复制推荐信息')
  } catch (e) {}
}
</script>

<style scoped>
.friend-detail-wrapper {
  height: 100%;
  width: 100%;
  background: #f8fafc;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 空状态 */
.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}
.empty-icon-circle {
  width: 72px;
  height: 72px;
  background: #f1f5f9;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}
.empty-icon-circle .icon { font-size: 28px; color: #94a3b8; }
.empty-text { font-size: 14px; }

/* 资料容器 */
.profile-container {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 24px;
}

.profile-inner {
  width: 100%;
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 1. 头部区域 */
.profile-header {
  display: flex;
  align-items: center;
  position: relative;
}

.avatar-box {
  width: 80px;
  height: 80px;
  flex-shrink: 0;
  margin-right: 20px;
}

.avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.primary-name {
  display: flex;
  align-items: center;
}

.remark {
  font-size: 22px;
  font-weight: 600;
  color: #0f172a;
  margin-right: 8px;
}

.gender-icon { font-size: 14px; }
.icon-male { color: #3b82f6; }
.icon-female { color: #ec4899; }

.sub-info {
  font-size: 13px;
  color: #64748b;
  line-height: 1.6;
}

.btn-settings {
  position: absolute;
  top: 0;
  right: 0;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 6px;
  color: #94a3b8;
  border-radius: 6px;
  transition: all 0.15s;
}
.btn-settings:hover { color: #64748b; background: #f1f5f9; }
.btn-settings .icon { font-size: 18px; }

.divider {
  height: 1px;
  background: #e2e8f0;
}

/* 2. 信息列表区域 */
.info-list {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

.info-item {
  display: flex;
  padding: 12px 16px;
  align-items: flex-start;
  border-bottom: 1px solid #f1f5f9;
}
.info-item:last-child { border-bottom: none; }

.info-item.clickable {
  cursor: pointer;
  transition: background 0.15s;
}
.info-item.clickable:hover {
  background: #f8fafc;
}

.label {
  width: 72px;
  color: #64748b;
  font-size: 13px;
  flex-shrink: 0;
  padding-top: 1px;
}

.content {
  flex: 1;
  font-size: 13px;
  color: #1e293b;
  line-height: 1.5;
}

.placeholder { color: #cbd5e1; }

.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-badge {
  background: #eff6ff;
  color: #2563eb;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.signature {
  color: #475569;
}

.arrow {
  color: #cbd5e1;
  font-size: 14px;
  margin-left: 8px;
}

.moments-item {
  align-items: center;
}

.moments-preview {
  display: flex;
}

.thumbs {
  display: flex;
  gap: 6px;
}

.thumb-img {
  width: 36px;
  height: 36px;
  background: #f1f5f9;
  border-radius: 6px;
}

/* 3. 底部操作按钮区域 */
.action-footer {
  display: flex;
  gap: 10px;
}

.btn-block {
  flex: 1;
  height: 44px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.15s;
}

.btn-primary {
  background: #2563eb;
  color: white;
  border: none;
}
.btn-primary:hover {
  background: #1d4ed8;
}

.btn-default {
  background: #fff;
  color: #334155;
  border: 1px solid #e2e8f0;
}
.btn-default:hover {
  background: #f8fafc;
}

.btn-block .icon {
  font-size: 16px;
}
</style>