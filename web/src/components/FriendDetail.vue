<template>
  <div class="friend-detail-container">
    <!-- 空状态 -->
    <div v-if="!friend" class="empty-state">
      <div class="empty-content">
        <i class="icon icon-users"></i>
        <h3>选择好友</h3>
        <p>从左侧列表中选择一个好友查看详情</p>
      </div>
    </div>

    <!-- 好友详情 -->
    <div v-else class="friend-detail">
      <!-- 头部信息 -->
      <div class="detail-header">
        <div class="header-content">
          <div class="avatar-section">
            <img :src="friend.avatar" alt="头像" class="avatar" @error="handleAvatarError">
            <div class="status-indicator" :class="friend.status || 'offline'"></div>
          </div>
          <div class="info-section">
            <div class="name-row">
              <h2 class="display-name">{{ friend.remark || friend.nickname || '未知用户' }}</h2>
              <div class="gender-icon" v-if="friend.sex">
                <i class="icon" :class="friend.sex === 1 ? 'icon-male' : friend.sex === 2 ? 'icon-female' : ''"></i>
              </div>
            </div>
            <p class="sub-name" v-if="friend.remark && friend.remark !== friend.nickname">昵称: {{ friend.nickname }}</p>
            <p class="account-id">微信号: {{ friend.friend_uid || friend.id }}</p>
            <p class="location" v-if="friend.location">地区: {{ friend.location }}</p>
          </div>
        </div>
        <!-- 设置备注按钮 -->
        <button class="btn-icon" @click="handleSetRemark" title="设置备注">
          <i class="icon icon-edit-3"></i>
        </button>
      </div>

      <!-- 分隔线 -->
      <div class="divider"></div>

      <!-- 详细信息列表 -->
      <div class="info-list">
        <div class="info-item" v-if="friend.phone">
          <span class="label">电话</span>
          <span class="value">{{ friend.phone }}</span>
        </div>
        <div class="info-item" v-if="friend.email">
          <span class="label">邮箱</span>
          <span class="value">{{ friend.email }}</span>
        </div>
        <div class="info-item" v-if="friend.occupation">
          <span class="label">职业</span>
          <span class="value">{{ friend.occupation }}</span>
        </div>
        <div class="info-item" v-if="friend.signature">
          <span class="label">个性签名</span>
          <span class="value">{{ friend.signature }}</span>
        </div>
        <div class="info-item">
          <span class="label">标签</span>
          <div class="tags-wrapper" v-if="friend.tags && friend.tags.length > 0">
            <span v-for="tag in friend.tags" :key="tag" class="tag">{{ tag }}</span>
          </div>
          <span v-else class="value placeholder">无标签</span>
        </div>
        <div class="info-item clickable" @click="viewMoments">
          <span class="label">朋友圈</span>
          <div class="moments-preview">
            <template v-if="friendMoments.length > 0">
              <div v-for="moment in friendMoments.slice(0, 3)" :key="moment.id" class="moment-thumb" :style="{ backgroundImage: `url(${moment.image})` }"></div>
            </template>
            <i class="icon icon-chevron-right arrow"></i>
          </div>
        </div>
      </div>

      <!-- 分隔线 -->
      <div class="divider"></div>

      <!-- 操作按钮 -->
      <div class="action-buttons">
        <button class="btn-action primary" @click="sendMessage">
          <i class="icon icon-message-circle"></i>
          <span>发消息</span>
        </button>
        <button class="btn-action" @click="audioCall">
          <i class="icon icon-phone"></i>
          <span>语音通话</span>
        </button>
        <button class="btn-action" @click="videoCall">
          <i class="icon icon-video"></i>
          <span>视频通话</span>
        </button>
      </div>
    </div>

    <!-- 好友设置弹窗 -->
    <FriendSettings 
      v-if="showSettings" 
      :friend="friend"
      @close="showSettings = false"
      @update-friend="updateFriend"
      @delete-friend="deleteFriend"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import FriendSettings from './FriendSettings.vue'

const props = defineProps({
  friend: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['send-message', 'audio-call', 'video-call', 'view-moments', 'update-remark'])
const showSettings = ref(false)

// 模拟朋友圈数据
const friendMoments = computed(() => {
  if (!props.friend) return []
  return [
    { id: 1, image: 'https://picsum.photos/seed/friend1/200/200' },
    { id: 2, image: 'https://picsum.photos/seed/friend2/200/200' },
    { id: 3, image: 'https://picsum.photos/seed/friend3/200/200' }
  ]
})

const handleAvatarError = (event) => {
  const friendUid = props.friend.friend_uid || props.friend.id
  event.target.src = `https://api.dicebear.com/7.x/personas/svg?seed=${friendUid || 'default'}`
}

const handleSetRemark = () => {
  const currentRemark = props.friend.remark || props.friend.nickname
  const newRemark = prompt('设置备注名', currentRemark)
  if (newRemark !== null && newRemark !== currentRemark) {
    // 这里应该调用 API 更新备注
    console.log('更新备注为:', newRemark)
    // 模拟更新
    // emit('update-remark', { ...props.friend, remark: newRemark })
  }
}

const sendMessage = () => { emit('send-message', props.friend) }
const audioCall = () => { emit('audio-call', props.friend) }
const videoCall = () => { emit('video-call', props.friend) }
const viewMoments = () => { emit('view-moments', props.friend) }
const updateFriend = (updatedFriend) => { /* 可补充更新逻辑 */ }
const deleteFriend = (friendId) => { showSettings.value = false }
</script>

<style scoped>
.friend-detail-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow: hidden;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  color: #999;
}

.empty-content .icon {
  font-size: 64px;
  margin-bottom: 16px;
  color: #e0e0e0;
}

.friend-detail {
  flex: 1;
  padding: 60px 80px;
  overflow-y: auto;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 40px;
}

.header-content {
  display: flex;
  gap: 24px;
}

.avatar-section {
  position: relative;
}

.avatar {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid #e0e0e0;
}

.status-indicator {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid #fff;
}

.status-indicator.online { background-color: #28a745; }
.status-indicator.away { background-color: #ffc107; }
.status-indicator.offline { background-color: #999; }

.info-section {
  padding-top: 4px;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.display-name {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.gender-icon .icon {
  font-size: 16px;
}

.icon-male { color: #4a8cff; }
.icon-female { color: #ff6b6b; }

.sub-name {
  font-size: 14px;
  color: #999;
  margin: 0 0 8px 0;
}

.account-id, .location {
  font-size: 14px;
  color: #999;
  margin: 4px 0;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  color: #666;
  transition: background-color 0.2s;
}

.btn-icon:hover {
  background-color: #f0f0f0;
}

.btn-icon .icon {
  font-size: 20px;
}

.divider {
  height: 1px;
  background-color: #e7e7e7;
  margin: 0 0 30px 0;
}

.info-list {
  margin-bottom: 40px;
}

.info-item {
  display: flex;
  margin-bottom: 20px;
  font-size: 14px;
}

.info-item.clickable {
  cursor: pointer;
  align-items: center;
}

.info-item .label {
  width: 80px;
  color: #999;
}

.info-item .value {
  color: #333;
  flex: 1;
}

.value.placeholder {
  color: #ccc;
}

.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  background-color: #f0f0f0;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.moments-preview {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.moment-thumb {
  width: 40px;
  height: 40px;
  background-size: cover;
  background-position: center;
  border-radius: 4px;
}

.arrow {
  margin-left: auto;
  color: #ccc;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 24px;
}

.btn-action {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 12px 24px;
  border-radius: 8px;
  color: #666;
  transition: all 0.2s;
  width: 100px;
}

.btn-action:hover {
  background-color: #f5f5f5;
}

.btn-action.primary {
  color: #4a8cff;
}

.btn-action .icon {
  font-size: 24px;
}

.btn-action span {
  font-size: 13px;
  font-weight: 500;
}
</style>