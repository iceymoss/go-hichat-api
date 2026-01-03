<template>
  <div class="feed-detail-container">
    <div class="feed-detail" v-if="selectedReactive">
      <div class="feed-detail-header">
        <button class="back-btn" @click="$emit('back')">
          <div class="back-btn-icon">
            <i class="icon icon-arrow-left"></i>
          </div>
          <span class="back-btn-text">返回动态列表</span>
          <div class="back-btn-ripple"></div>
        </button>
        <img 
          :src="selectedReactive.user.avatar" 
          class="avatar" 
          @mouseenter="(e) => showUserProfile(selectedReactive.user, e)"
          @mouseleave="hideUserProfile"
        />
        <div class="user-info">
          <div 
            class="name"
            @mouseenter="(e) => showUserProfile(selectedReactive.user, e)"
            @mouseleave="hideUserProfile"
          >{{ selectedReactive.user.name }}</div>
          <div class="time">{{ selectedReactive.time }}</div>
        </div>
      </div>
      
      <div class="feed-detail-content">
        <div class="feed-detail-text">{{ selectedReactive.content }}</div>
        
        <!-- 新的多图片格式 -->
        <div class="feed-detail-images" v-if="selectedReactive.images && selectedReactive.images.length > 0">
          <div :class="['images-grid', getImageGridClass(selectedReactive.images.length)]">
            <div
              v-for="(image, index) in selectedReactive.images.slice(0, 9)"
              :key="index"
              class="image-item"
              @click="viewImage(index)"
            >
              <img :src="image" alt="动态图片" class="feed-image-item">
              <div v-if="selectedReactive.images.length > 9 && index === 8" class="more-images-count">
                +{{ selectedReactive.images.length - 9 }}
              </div>
            </div>
          </div>
        </div>
        
        <!-- 兼容旧的单图片格式 -->
        <div class="feed-detail-image" v-if="selectedReactive.image && (!selectedReactive.images || selectedReactive.images.length === 0)">
          <img :src="selectedReactive.image" alt="动态图片" @click="viewImage(0)" />
        </div>
      </div>
      
      <div class="feed-detail-stats">
        <div class="stat-item">
          <i class="icon icon-like"></i>
          <span>{{ selectedReactive.likes ? selectedReactive.likes.length : 0 }} 个赞</span>
        </div>
        <div class="stat-item">
          <i class="icon icon-comment"></i>
          <span>{{ selectedReactive.comments ? selectedReactive.comments.length : 0 }} 条评论</span>
        </div>
        <div class="stat-item">
          <i class="icon icon-share"></i>
          <span>分享</span>
        </div>
      </div>
      
      <div class="feed-detail-likes" v-if="selectedReactive.likes && selectedReactive.likes.length">
        <div class="likes-avatars">
          <img 
            v-for="(user, index) in Array.from(new Set(selectedReactive.likes)).slice(0, 15)" 
            :key="user"
            :src="`https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(user)}`"
            :alt="user"
            class="like-avatar"
            @mouseenter="(e) => showUserProfile({ name: user, avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(user)}` }, e)"
            @mouseleave="hideUserProfile"
          />
          <span v-if="selectedReactive.likes.length > 15" class="more-likes">+{{ selectedReactive.likes.length - 15 }}</span>
        </div>
        <div class="likes-text">
          <span v-for="(user, index) in Array.from(new Set(selectedReactive.likes)).slice(0, 15)" :key="user" class="like-user">
            <span 
              class="like-user-name"
              @mouseenter="(e) => showUserProfile({ name: user, avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(user)}` }, e)"
              @mouseleave="hideUserProfile"
            >{{ user }}</span>{{ index < Math.min(14, selectedReactive.likes.length - 1) ? '、' : '' }}
          </span>
          <span v-if="selectedReactive.likes.length > 15">等{{ selectedReactive.likes.length }}位好友点赞</span>
        </div>
      </div>
      
      <div class="feed-detail-actions">
        <button class="action-btn like-btn" :class="{ active: selectedReactive.liked }" @click="toggleLike">
          <i class="icon" :class="selectedReactive.liked ? 'icon-liked' : 'icon-like'"></i>
          <span>{{ selectedReactive.liked ? '已点赞' : '点赞' }}</span>
        </button>
        <button class="action-btn comment-btn" @click="focusCommentInput">
          <i class="icon icon-comment"></i>
          <span>评论</span>
        </button>
      </div>
      

      
      <div class="feed-detail-comments">
        <div class="comments-title">全部评论</div>
        <div class="add-comment">
          <input v-model="newComment" placeholder="写评论，支持@用户" @keyup.enter="addComment" />
          <button @click="addComment">发送</button>
        </div>
        
        <!-- 评论列表 -->
        <CommentThread 
          :comments="selectedReactive.comments"
          :format-comment="formatComment"
          :show-user-profile="showUserProfile"
          :hide-user-profile="hideUserProfile"
          @reply="handleCommentReply"
          @like-comment="handleCommentLike"
          @like-reply="handleReplyLike"
          @delete-comment="handleDeleteComment"
          @delete-reply="handleDeleteReply"
        />
      </div>
    </div>
    <div v-else class="feed-detail-empty">请选择一条动态</div>
    
    <!-- 图片查看器 -->
    <ImageViewer
      :visible="showImageViewer"
      :images="viewerImages"
      :initial-index="viewerInitialIndex"
      @close="closeImageViewer"
    />
    
    <!-- 用户资料卡片 -->
    <UserProfileCard
      :visible="profileCard.visible"
      :user="profileCard.user"
      :position="profileCard.position"
      :arrow-position="profileCard.arrowPosition"
      @close="closeProfileCard"
      @view-profile="handleViewProfile"
      @view-space="handleViewSpace"
      @mouse-enter-card="handleMouseEnterCard"
      @mouse-leave-card="handleMouseLeaveCard"
    />
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import CommentThread from './CommentThread.vue'
import ImageViewer from './ImageViewer.vue'
import UserProfileCard from './UserProfileCard.vue'
import { useFeedStore } from '../stores/feed'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const props = defineProps({ selected: Object })
const emit = defineEmits(['back', 'like', 'comment'])

const feedStore = useFeedStore()
const authStore = useAuthStore()
const router = useRouter()
const selectedReactive = computed(() => {
  if (!props.selected) return null
  const allFeeds = [...feedStore.getFriendFeeds, ...feedStore.getCommunityFeeds]
  const foundFeed = allFeeds.find(f => String(f.id) === String(props.selected.id))
  if (foundFeed) {
    return foundFeed
  }
  return props.selected
})

const newComment = ref('')
const showImageViewer = ref(false)
const viewerImages = ref([])
const viewerInitialIndex = ref(0)

// 用户资料卡片状态
const profileCard = ref({
  visible: false,
  user: null,
  position: { x: 0, y: 0 },
  arrowPosition: 'top'
})

let hoverTimeout = null

function handleDeleteComment(comment) {
  if (!selectedReactive.value) return
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  feedStore.deleteComment(postId, comment.id, type)
}

function addComment() {
  if (!newComment.value.trim() || !selectedReactive.value) return
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  feedStore.addComment(postId, newComment.value, type)
  newComment.value = ''
}

function handleCommentReply({ comment, reply }) {
  if (!selectedReactive.value) return
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  feedStore.addReply(postId, comment.id, reply.content, type)
}

function handleCommentLike(comment) {
  if (!selectedReactive.value) return
  
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  
  feedStore.toggleCommentLike(postId, comment.id, type)
}

function handleReplyLike(reply) {
  if (!selectedReactive.value) return
  
  const comment = selectedReactive.value.comments.find(c => 
    c.replies && c.replies.some(r => r.id === reply.id)
  )
  
  if (!comment) return
  
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  
  feedStore.toggleReplyLike(postId, comment.id, reply.id, type)
}

function formatComment(content) {
  return content.replace(/@([\u4e00-\u9fa5\w\-]+)/g, '<span class="at-user">@$1</span>')
}

function toggleLike() {
  if (!selectedReactive.value) return
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  feedStore.toggleLike(postId, type)
}

function focusCommentInput() {
  const commentInput = document.querySelector('.add-comment input')
  if (commentInput) {
    commentInput.focus()
  }
}

function handleDeleteReply({ comment, reply }) {
  if (!selectedReactive.value) return
  const postId = selectedReactive.value.id
  const type = selectedReactive.value.type || 'friend'
  feedStore.deleteReply(postId, comment.id, reply.id, type)
}

// 查看图片
function viewImage(index) {
  if (!selectedReactive.value) return
  
  // 收集当前动态的所有图片
  const images = []
  
  // 如果有新的多图片格式
  if (selectedReactive.value.images && selectedReactive.value.images.length > 0) {
    images.push(...selectedReactive.value.images)
  } 
  // 如果只有旧的单图片格式
  else if (selectedReactive.value.image) {
    images.push(selectedReactive.value.image)
  }
  
  if (images.length > 0) {
    viewerImages.value = images
    viewerInitialIndex.value = Math.max(0, Math.min(index, images.length - 1))
    showImageViewer.value = true
  }
}

// 关闭图片查看器
function closeImageViewer() {
  showImageViewer.value = false
  viewerImages.value = []
  viewerInitialIndex.value = 0
}

// 获取图片网格布局类名
function getImageGridClass(count) {
  if (count === 1) return 'grid-single'
  if (count === 2) return 'grid-double'
  if (count === 3) return 'grid-triple'
  if (count === 4) return 'grid-quad'
  return 'grid-multi'
}

// 用户资料卡片相关函数
function showUserProfile(user, event) {
  clearTimeout(hoverTimeout)
  
  // 计算位置
  const rect = event.target.getBoundingClientRect()
  const x = rect.left + rect.width / 2
  const y = rect.bottom + window.scrollY
  
  // 检查是否靠近页面底部
  const windowHeight = window.innerHeight
  const cardHeight = 400 // 估算卡片高度
  const arrowPosition = rect.bottom + cardHeight > windowHeight ? 'bottom' : 'top'
  
  // 构建用户资料数据
  const profileUser = {
    ...user,
    description: user.description || '这个人很懒，什么都没写...',
    isOnline: Math.random() > 0.5, // 随机在线状态
    feedCount: Math.floor(Math.random() * 100) + 10,
    friendCount: Math.floor(Math.random() * 500) + 50,
    likeCount: Math.floor(Math.random() * 1000) + 100
  }
  
  profileCard.value = {
    visible: true,
    user: profileUser,
    position: { x, y },
    arrowPosition
  }
}

function hideUserProfile() {
  hoverTimeout = setTimeout(() => {
    profileCard.value.visible = false
  }, 300)
}

function handleMouseEnterCard() {
  // 清除关闭定时器
  clearTimeout(hoverTimeout)
}

function handleMouseLeaveCard() {
  // 延迟关闭卡片
  hoverTimeout = setTimeout(() => {
    profileCard.value.visible = false
  }, 300)
}

function handleViewProfile(user) {
  router.push(`/profile/${user.id || user.name}`)
}

function handleViewSpace(user) {
  router.push(`/space/${user.id || user.name}`)
}

function closeProfileCard() {
  profileCard.value.visible = false
}


</script>

<style scoped>
.feed-detail-container {
  height: 100%;
  overflow-y: auto;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 50%, #e0e7ff 100%);
  padding: 24px;
  position: relative;
  z-index: 1; /* 确保正确的层级 */
}

.feed-detail-container::before {
  content: '';
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(circle at 20% 20%, rgba(74,140,255,0.03) 0%, transparent 50%),
    radial-gradient(circle at 80% 80%, rgba(138,105,255,0.02) 0%, transparent 50%);
  pointer-events: none;
  z-index: 0;
}

.feed-detail {
  max-width: 800px;
  margin: 0 auto;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  padding: 36px;
  box-shadow: 
    0 20px 40px rgba(74, 140, 255, 0.08),
    0 8px 16px rgba(138, 105, 255, 0.05),
    0 0 0 1px rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(229, 231, 235, 0.2);
  backdrop-filter: blur(20px);
  position: relative;
  z-index: 2;
}
.feed-detail-header {
  display: flex;
  align-items: center;
  margin-bottom: 32px;
  gap: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(229, 231, 235, 0.3);
}
.back-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  box-shadow: 
    0 3px 10px rgba(74, 140, 255, 0.25),
    0 1px 4px rgba(138, 105, 255, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  min-width: 120px;
  justify-content: center;
}

.back-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, transparent 50%);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.back-btn:hover {
  transform: translateY(-1px) scale(1.02);
  box-shadow: 
    0 4px 16px rgba(74, 140, 255, 0.35),
    0 2px 8px rgba(138, 105, 255, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  background: linear-gradient(135deg, #5a9cff 0%, #9a79ff 100%);
}

.back-btn:hover::before {
  opacity: 1;
}

.back-btn:active {
  transform: translateY(-1px) scale(0.98);
  transition: all 0.1s ease;
}

.back-btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.15);
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}

.back-btn:hover .back-btn-icon {
  background: rgba(255, 255, 255, 0.25);
  transform: translateX(-2px) rotate(-5deg);
}

.back-btn-icon .icon {
  font-size: 11px;
  color: #ffffff;
  font-weight: bold;
}

.back-btn-text {
  white-space: nowrap;
  letter-spacing: 0.3px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.back-btn-ripple {
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  border-radius: 12px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  opacity: 0;
  transform: scale(0);
  transition: all 0.6s ease;
  pointer-events: none;
}

.back-btn:active .back-btn-ripple {
  opacity: 1;
  transform: scale(1);
  transition: all 0.3s ease;
}

/* 添加聚焦状态的可访问性支持 */
.back-btn:focus {
  outline: none;
  box-shadow: 
    0 3px 10px rgba(74, 140, 255, 0.25),
    0 1px 4px rgba(138, 105, 255, 0.15),
    0 0 0 2px rgba(74, 140, 255, 0.3);
}
.avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  border: 3px solid #fff;
  box-shadow: 
    0 4px 12px rgba(0,0,0,0.08),
    0 0 0 1px rgba(74, 140, 255, 0.1);
  transition: all 0.3s ease;
  cursor: pointer;
}

.avatar:hover {
  transform: scale(1.05);
  box-shadow: 
    0 6px 16px rgba(0,0,0,0.12),
    0 0 0 2px rgba(74, 140, 255, 0.2);
}

.user-info {
  flex: 1;
}

.user-info .name {
  color: #1e293b;
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 2px;
  background: linear-gradient(135deg, #1e293b 0%, #4a8cff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 6px;
  padding: 4px 6px;
  margin: -4px -6px 2px -6px;
}

.user-info .name:hover {
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  transform: scale(1.02);
  box-shadow: 0 2px 8px rgba(74, 140, 255, 0.2);
}

.user-info .time {
  color: #64748b;
  font-size: 14px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.user-info .time::before {
  content: '•';
  color: #94a3b8;
  font-weight: bold;
}
.feed-detail-content {
  margin-bottom: 24px;
}
.feed-detail-text {
  color: #222;
  font-size: 16px;
  line-height: 1.7;
  margin-bottom: 16px;
}
/* 单图片样式 */
.feed-detail-image img {
  max-width: 100%;
  max-height: 500px;
  border-radius: 12px;
  object-fit: cover;
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.3s ease;
}

.feed-detail-image img:hover {
  transform: scale(1.02);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
  border-color: #cbd5e1;
}

/* 多图片网格布局 */
.feed-detail-images {
  margin: 16px 0;
}

.images-grid {
  display: grid;
  gap: 8px;
  border-radius: 12px;
  overflow: hidden;
  max-width: 100%;
}

.images-grid.grid-single {
  grid-template-columns: 1fr;
  max-width: 500px;
}

.images-grid.grid-double {
  grid-template-columns: 1fr 1fr;
  max-width: 500px;
}

.images-grid.grid-triple {
  grid-template-columns: 2fr 1fr;
  grid-template-rows: 1fr 1fr;
  max-width: 500px;
}

.images-grid.grid-triple .image-item:first-child {
  grid-row: 1 / 3;
}

.images-grid.grid-quad {
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  max-width: 500px;
}

.images-grid.grid-multi {
  grid-template-columns: repeat(3, 1fr);
  max-width: 500px;
}

.image-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #f8fafc;
}

.image-item:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 3;
}

.feed-image-item {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.image-item:hover .feed-image-item {
  transform: scale(1.05);
}

.more-images-count {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  backdrop-filter: blur(2px);
}
.feed-detail-stats {
  display: flex;
  gap: 32px;
  margin-bottom: 20px;
  padding: 16px 0 0 0;
  border-top: 1px solid #f1f5f9;
}
.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #64748b;
  font-size: 14px;
}
.feed-detail-likes {
  margin-bottom: 20px;
  padding: 12px 0 0 0;
}
.likes-avatars {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.like-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid #fff;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.like-avatar:hover {
  transform: scale(1.15);
  box-shadow: 0 4px 16px rgba(74, 140, 255, 0.3);
  border-color: rgba(74, 140, 255, 0.5);
}
.more-likes {
  color: #64748b;
  font-size: 12px;
  margin-left: 8px;
}
.likes-text {
  color: #374151;
  font-size: 14px;
}
.like-user {
  font-weight: 600;
  color: #1f2937;
}

.like-user-name {
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 2px 4px;
  border-radius: 4px;
  margin: -2px -4px;
  color: #2563eb;
  font-weight: 600;
}

.like-user-name:hover {
  color: #8a69ff;
  background: rgba(138, 105, 255, 0.1);
  transform: scale(1.05);
}
.feed-detail-actions {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px 0 0 0;
  border-bottom: 1px solid #f1f5f9;
}
.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  background: #f8fafc;
  color: #4a8cff;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}
.action-btn:hover {
  background: #e0e7ef;
  color: #2563eb;
}
.action-btn.active {
  background: #dbeafe;
  color: #2563eb;
}
.comments-title {
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 16px;
}
.add-comment {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  align-items: center;
}
.add-comment input {
  flex: 1;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 15px;
  outline: none;
  transition: border-color 0.2s;
  background: #fff;
}
.add-comment input:focus {
  border-color: #4a8cff;
}
.add-comment button {
  padding: 12px 20px;
  background: #4a8cff;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: background-color 0.2s;
  font-weight: 500;
}
.add-comment button:hover {
  background: #2563eb;
}
.feed-detail-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  height: 100%;
  color: #64748b;
  font-size: 16px;
  gap: 12px;
  text-align: center;
  padding: 40px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  backdrop-filter: blur(20px);
  box-shadow: 
    0 20px 40px rgba(74, 140, 255, 0.08),
    0 8px 16px rgba(138, 105, 255, 0.05);
  border: 1px solid rgba(229, 231, 235, 0.2);
  position: relative;
  z-index: 2;
  max-width: 400px;
  margin: 0 auto;
}

.feed-detail-empty::before {
  content: '📝';
  font-size: 48px;
  opacity: 0.6;
}
.more-comments {
  color: #4a8cff;
  font-size: 14px;
  cursor: pointer;
  margin: 8px 0 0 0;
  user-select: none;
  transition: color 0.2s;
  display: inline-block;
}
.more-comments:hover {
  text-decoration: underline;
  color: #2563eb;
}
/* @用户样式 */
:deep(.at-user) {
  color: #4a8cff;
  font-weight: 600;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .feed-detail-container {
    padding: 16px;
  }
  
  .feed-detail {
    padding: 24px;
  }
  
  .feed-detail-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .back-btn {
    align-self: stretch;
    justify-content: center;
    min-width: 100px;
    padding: 10px 14px;
    font-size: 12px;
  }
  
  .back-btn-icon {
    width: 16px;
    height: 16px;
  }
  
  .back-btn-icon .icon {
    font-size: 10px;
  }
  
  .images-grid.grid-single {
    max-width: 100%;
  }
  
  .images-grid.grid-double {
    max-width: 100%;
  }
  
  .images-grid.grid-triple {
    max-width: 100%;
  }
  
  .images-grid.grid-quad {
    max-width: 100%;
  }
  
  .images-grid.grid-multi {
    max-width: 100%;
  }
  
  .feed-detail-stats {
    gap: 16px;
    flex-wrap: wrap;
  }
  
  .feed-detail-actions {
    flex-direction: column;
    gap: 12px;
  }
  
  .action-btn {
    justify-content: center;
  }
  
  .add-comment {
    flex-direction: column;
    gap: 12px;
  }
  
  .add-comment input {
    order: 1;
  }
  
  .add-comment button {
    order: 2;
    align-self: stretch;
  }
}
</style> 