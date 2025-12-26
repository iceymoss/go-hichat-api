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
            <img :src="friend.avatar" alt="头像" class="avatar">
            <div class="status-indicator" :class="friend.status">
              <span class="status-dot"></span>
              <span class="status-text">{{ getStatusText(friend.status) }}</span>
            </div>
          </div>
          <div class="info-section">
            <h2 class="display-name">{{ friend.remark || friend.name }}</h2>
            <p class="nickname" v-if="friend.remark">昵称：{{ friend.name }}</p>
            <p class="account-id">账号ID：{{ friend.id }}</p>
            <p class="gender" v-if="friend.gender">性别：{{ getGenderText(friend.gender) }}</p>
            <p class="email" v-if="friend.email">邮箱：{{ friend.email }}</p>
            <p class="phone" v-if="friend.phone">手机号：{{ friend.phone }}</p>
            <p class="signature" v-if="friend.signature">个性签名：{{ friend.signature }}</p>
            <p class="location" v-if="friend.location">地区：{{ friend.location }}</p>
          </div>
        </div>
        <!-- 设置按钮 -->
        <button class="btn-settings" @click="showSettings = true" title="设置">
          <i class="icon icon-settings"></i>
        </button>
      </div>

      <!-- 标签 -->
      <div class="tags-section" v-if="friend.tags && friend.tags.length > 0">
        <h4>标签</h4>
        <div class="tags-list">
          <span v-for="tag in friend.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>
      </div>

      <!-- 朋友圈区块 -->
      <div class="moments-section">
        <div class="moments-header">
          <div class="moments-title">
            <i class="icon icon-feed"></i>
            <span>朋友圈</span>
          </div>
          <button class="btn-view-moments" @click="viewMoments">
            <span>查看</span>
            <i class="icon icon-chevron-right"></i>
          </button>
        </div>
        <div class="moments-preview">
          <div class="moments-gallery" v-if="friendMoments.length > 0">
            <div v-for="(moment, index) in friendMoments.slice(0, 3)" :key="moment.id" class="moment-thumbnail" :style="{ backgroundImage: `url(${moment.image})` }">
              <div class="moment-overlay">
                <span class="moment-time">{{ moment.time }}</span>
              </div>
            </div>
          </div>
          <div v-else class="no-moments">
            <i class="icon icon-feed"></i>
            <p>暂无动态</p>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="action-buttons">
        <button class="btn-action btn-message" @click="sendMessage">
          <i class="icon icon-message"></i>
          <span>发消息</span>
        </button>
        <button class="btn-action btn-call" @click="audioCall">
          <i class="icon icon-phone"></i>
          <span>语音通话</span>
        </button>
        <button class="btn-action btn-video" @click="videoCall">
          <i class="icon icon-video"></i>
          <span>视频通话</span>
        </button>
      </div>

      <!-- 最近活跃 -->
      <div class="activity-section">
        <h4>最近活跃</h4>
        <p class="last-active">{{ friend.lastActive }}</p>
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
const emit = defineEmits(['send-message', 'audio-call', 'video-call', 'view-moments'])
const showSettings = ref(false)
// 模拟朋友圈数据
const friendMoments = computed(() => {
  if (!props.friend) return []
  return [
    { id: 1, content: '周末愉快！', time: '2小时前', image: 'https://picsum.photos/seed/friend1/200/200' },
    { id: 2, content: '美食分享', time: '1天前', image: 'https://picsum.photos/seed/friend2/200/200' },
    { id: 3, content: '旅行记录', time: '3天前', image: 'https://picsum.photos/seed/friend3/200/200' }
  ]
})
const getStatusText = (status) => {
  const statusMap = { online: '在线', away: '离开', offline: '离线' }
  return statusMap[status] || '离线'
}
const getGenderText = (gender) => {
  const genderMap = { male: '男', female: '女', other: '其他' }
  return genderMap[gender] || gender
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
  width: 80%;
  min-width: 600px;
  max-width: 900px;
  margin: 48px auto 0 auto;
  padding: 24px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
  height: 100%;
  display: flex;
  flex-direction: column;
}
.empty-state { flex: 1; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%); height: 100%; }
.empty-content { text-align: center; color: #64748b; }
.empty-content .icon { font-size: 80px; margin-bottom: 20px; opacity: 0.5; }
.empty-content h3 { font-size: 24px; font-weight: 600; margin: 0 0 12px 0; color: #475569; }
.empty-content p { font-size: 16px; margin: 0; color: #94a3b8; }
.friend-detail { flex: 1; overflow-y: auto; padding: 0; }
.detail-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 32px; padding-bottom: 24px; border-bottom: 1px solid #e2e8f0; }
.header-content { display: flex; align-items: flex-start; gap: 24px; }
.avatar-section { position: relative; }
.avatar { width: 100px; height: 100px; border-radius: 20px; object-fit: cover; border: 4px solid #f1f5f9; }
.status-indicator { position: absolute; bottom: -6px; right: -6px; background: #fff; border-radius: 16px; padding: 6px 12px; display: flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 500; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15); }
.status-dot { width: 10px; height: 10px; border-radius: 50%; }
.status-indicator.online .status-dot { background-color: #10b981; }
.status-indicator.away .status-dot { background-color: #f59e0b; }
.status-indicator.offline .status-dot { background-color: #6b7280; }
.status-indicator.online .status-text { color: #10b981; }
.status-indicator.away .status-text { color: #f59e0b; }
.status-indicator.offline .status-text { color: #6b7280; }
.info-section { flex: 1; }
.display-name { font-size: 28px; font-weight: 700; color: #1e293b; margin: 0 0 12px 0; }
.nickname, .account-id, .gender, .email, .phone, .location { font-size: 16px; color: #64748b; margin: 6px 0; }
.signature { font-size: 16px; color: #475569; margin: 12px 0 0 0; font-style: italic; line-height: 1.5; }
.btn-settings { width: 48px; height: 48px; border: none; border-radius: 12px; background: #f8fafc; color: #64748b; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.btn-settings:hover { background: #e2e8f0; color: #475569; }
.tags-section { margin-bottom: 32px; }
.tags-section h4 { font-size: 18px; font-weight: 600; color: #1e293b; margin: 0 0 16px 0; }
.tags-list { display: flex; flex-wrap: wrap; gap: 12px; }
.tag { padding: 8px 16px; background: linear-gradient(135deg, #4a8cff, #8a69ff); color: white; border-radius: 20px; font-size: 14px; font-weight: 500; }
.moments-section {
  margin-bottom: 32px;
  background: #f8fafc;
  border-radius: 16px;
  padding: 20px;
  max-width: 864px;
  min-height: 110px;
  margin-left: auto;
  margin-right: auto;
}
.moments-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.moments-title { display: flex; align-items: center; gap: 10px; font-size: 18px; font-weight: 600; color: #1e293b; }
.moments-title .icon { font-size: 20px; color: #4a8cff; }
.btn-view-moments { background: none; border: none; color: #4a8cff; cursor: pointer; padding: 10px 16px; border-radius: 8px; font-size: 16px; font-weight: 500; display: flex; align-items: center; gap: 6px; transition: all 0.2s; }
.btn-view-moments:hover { background: #e0f2fe; }
.moments-preview { display: flex; flex-direction: column; gap: 20px; }
.moments-gallery { display: flex; gap: 12px; }
.moment-thumbnail { border-radius: 12px; background-size: cover; background-position: center; position: relative; overflow: hidden; width: 80px; height: 80px; }
.moment-overlay { position: absolute; bottom: 0; left: 0; right: 0; background: linear-gradient(transparent, rgba(0, 0, 0, 0.6)); padding: 10px; }
.moment-time { color: white; font-size: 12px; font-weight: 500; }
.no-moments { text-align: center; color: #94a3b8; padding: 24px 0; }
.no-moments .icon { font-size: 40px; margin-bottom: 12px; opacity: 0.5; }
.no-moments p { font-size: 16px; margin: 0; }
.action-buttons { display: flex; flex-direction: column; gap: 16px; margin-bottom: 32px; }
.btn-action { width: 100%; padding: 20px 24px; border: none; border-radius: 16px; font-size: 18px; font-weight: 500; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 16px; transition: all 0.2s; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); }
.btn-message { background: linear-gradient(135deg, #4a8cff, #8a69ff); color: white; }
.btn-message:hover { transform: translateY(-3px); box-shadow: 0 12px 24px rgba(74, 140, 255, 0.3); }
.btn-call { background: linear-gradient(135deg, #10b981, #059669); color: white; }
.btn-call:hover { transform: translateY(-3px); box-shadow: 0 12px 24px rgba(16, 185, 129, 0.3); }
.btn-video { background: linear-gradient(135deg, #f59e0b, #d97706); color: white; }
.btn-video:hover { transform: translateY(-3px); box-shadow: 0 12px 24px rgba(245, 158, 11, 0.3); }
.btn-action .icon { font-size: 24px; }
.activity-section { padding-top: 24px; border-top: 1px solid #e2e8f0; }
.activity-section h4 { font-size: 18px; font-weight: 600; color: #1e293b; margin: 0 0 12px 0; }
.last-active { font-size: 16px; color: #64748b; margin: 0; }
</style> 