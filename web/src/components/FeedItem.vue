<template>
  <div class="feed-item">
    <div class="feed-header">
      <img :src="post.avatar" alt="头像" class="avatar">
      <div class="user-info">
        <div class="user-name">{{ post.userName }}</div>
        <div class="post-time">{{ formatTime(post.createdAt) }}</div>
      </div>
    </div>
    
    <div class="feed-content">
      <p>{{ post.content }}</p>
    </div>
    
    <div v-if="post.images && post.images.length > 0" class="feed-images">
      <div :class="['image-grid', gridClass]">
        <div 
          v-for="(image, index) in post.images" 
          :key="index" 
          class="image-item"
          @click="openImageViewer(index)"
        >
          <img :src="image" alt="动态图片" class="feed-image">
          <div v-if="post.images.length > 1 && index === 8" class="image-count">+{{ post.images.length - 9 }}</div>
        </div>
      </div>
    </div>
    
    <div class="feed-stats">
      <span>{{ post.likes.length }} 点赞</span>
      <span>{{ post.comments.length }} 评论</span>
    </div>
    
    <div class="feed-actions">
      <button 
        class="action-btn" 
        :class="{ active: isLiked }"
        @click="toggleLike"
      >
        <i class="icon" :class="isLiked ? 'icon-liked' : 'icon-like'"></i>
        <span>{{ isLiked ? '已赞' : '赞' }}</span>
      </button>
      <button class="action-btn" @click="focusCommentInput">
        <i class="icon icon-comment"></i>
        <span>评论</span>
      </button>
    </div>
    
    <div class="feed-comments">
      <div v-if="post.comments.length > 0" class="comment-list">
        <div 
          v-for="comment in visibleComments" 
          :key="comment.id" 
          class="comment-item"
        >
          <img :src="getUserAvatar(comment.userId)" alt="头像" class="comment-avatar">
          <div class="comment-content">
            <div class="comment-user">{{ comment.userName }}</div>
            <div class="comment-text">{{ comment.content }}</div>
            <div class="comment-time">{{ formatTime(comment.time) }}</div>
          </div>
        </div>
        <div v-if="post.comments.length > 2" class="view-more" @click="showAllComments">
          查看更多{{ post.comments.length - 2 }}条评论
        </div>
      </div>
      
      <div class="comment-input-box">
        <img :src="user.avatar" alt="头像" class="comment-avatar">
        <input 
          ref="commentInput"
          type="text" 
          placeholder="评论..." 
          v-model="commentText"
          @keyup.enter="addComment"
        >
        <button class="send-btn" @click="addComment">
          <i class="icon icon-send"></i>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const props = defineProps({
  post: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['like', 'comment', 'open-image'])

const authStore = useAuthStore()
const commentInput = ref(null)
const commentText = ref('')
const showAll = ref(false)
const user = computed(() => authStore.user)

// 是否已点赞
const isLiked = computed(() => {
  if (!user.value) return false
  return props.post.likes.includes(user.value.id)
})

// 动态图片布局
const gridClass = computed(() => {
  const count = props.post.images.length
  if (count === 1) return 'single'
  if (count === 2) return 'double'
  if (count === 4) return 'quad'
  return 'multi'
})

// 可见评论（默认显示2条）
const visibleComments = computed(() => {
  if (showAll.value) return props.post.comments
  return props.post.comments.slice(0, 2)
})

// 格式化时间显示
const formatTime = (timeStr) => {
  const now = new Date()
  const postDate = new Date(timeStr)
  const diffHours = Math.floor((now - postDate) / (1000 * 60 * 60))
  
  if (diffHours < 1) {
    const diffMins = Math.floor((now - postDate) / (1000 * 60))
    return `${diffMins}分钟前`
  }
  
  if (diffHours < 24) {
    return `${diffHours}小时前`
  }
  
  if (diffHours < 168) { // 一周内
    return `${Math.floor(diffHours / 24)}天前`
  }
  
  return postDate.toLocaleDateString()
}

// 获取用户头像（实际应用中应从用户数据获取）
const getUserAvatar = (userId) => {
  // 这里简化实现，实际应用中应从用户数据获取
  return 'https://api.dicebear.com/7.x/personas/svg?seed=User' + userId
}

// 打开图片查看器
const openImageViewer = (index) => {
  emit('open-image', index)
}

// 点赞/取消点赞
const toggleLike = () => {
  emit('like', props.post.id)
}

// 聚焦评论输入框
const focusCommentInput = () => {
  commentInput.value.focus()
}

// 添加评论
const addComment = () => {
  if (!commentText.value.trim()) return
  
  emit('comment', {
    postId: props.post.id,
    content: commentText.value
  })
  
  commentText.value = ''
  showAll.value = true
}

// 显示全部评论
const showAllComments = () => {
  showAll.value = true
}
</script>

<style scoped>
.feed-item {
  background-color: white;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.feed-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  margin-right: 12px;
  object-fit: cover;
}

.user-info {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-weight: 500;
  margin-bottom: 2px;
}

.post-time {
  font-size: 12px;
  color: #999;
}

.feed-content {
  margin-bottom: 16px;
  font-size: 15px;
  line-height: 1.5;
  color: #333;
}

.feed-images {
  margin-bottom: 12px;
}

.image-grid {
  display: grid;
  gap: 4px;
  border-radius: 8px;
  overflow: hidden;
}

.image-grid.single {
  grid-template-columns: 1fr;
  max-height: 500px;
}

.image-grid.double {
  grid-template-columns: 1fr 1fr;
}

.image-grid.quad {
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
}

.image-grid.multi {
  grid-template-columns: repeat(3, 1fr);
}

.image-item {
  aspect-ratio: 1;
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.image-item:nth-child(10) {
  display: none;
}

.feed-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.image-item:hover .feed-image {
  transform: scale(1.05);
}

.image-count {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  color: white;
  font-size: 20px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
}

.feed-stats {
  display: flex;
  padding: 8px 0;
  border-top: 1px solid #f0f2f5;
  border-bottom: 1px solid #f0f2f5;
  color: #666;
  font-size: 13px;
}

.feed-stats span {
  margin-right: 15px;
}

.feed-actions {
  display: flex;
  padding: 8px 0;
  border-bottom: 1px solid #f0f2f5;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: none;
  border: none;
  padding: 8px;
  font-size: 14px;
  color: #666;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.action-btn:hover, .action-btn.active {
  background-color: #f0f2f5;
  color: #4a8cff;
}

.feed-comments {
  padding-top: 12px;
}

.comment-list {
  margin-bottom: 12px;
}

.comment-item {
  display: flex;
  margin-bottom: 10px;
}

.comment-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  margin-right: 8px;
  flex-shrink: 0;
  object-fit: cover;
}

.comment-content {
  flex: 1;
  background-color: #f0f2f5;
  border-radius: 16px;
  padding: 8px 12px;
}

.comment-user {
  font-weight: 500;
  font-size: 13px;
  margin-bottom: 2px;
}

.comment-text {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
}

.comment-time {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.view-more {
  font-size: 13px;
  color: #999;
  padding: 6px 0;
  cursor: pointer;
}

.view-more:hover {
  color: #4a8cff;
}

.comment-input-box {
  display: flex;
  align-items: center;
}

.comment-input-box .comment-avatar {
  margin-right: 8px;
}

.comment-input-box input {
  flex: 1;
  background-color: #f0f2f5;
  border: none;
  border-radius: 16px;
  padding: 10px 15px;
  font-size: 14px;
}

.send-btn {
  margin-left: 8px;
  background: none;
  border: none;
  color: #4a8cff;
  font-size: 20px;
  cursor: pointer;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.send-btn:hover {
  background-color: #eef5ff;
}
</style>