<template>
  <div class="user-profile-header" ref="headerElement">
    <!-- 封面背景 -->
    <div class="cover-background">
      <img :src="coverImage" alt="封面" class="cover-image" />
      <div class="cover-overlay"></div>
      <div class="cover-content">
        <!-- 用户主要信息 -->
        <div class="user-main-info">
          <div class="user-avatar-container">
            <img :src="currentUser.avatar" alt="头像" class="user-avatar" />
            <div class="avatar-glow"></div>
          </div>
          <div class="user-details">
            <h2 class="user-name">{{ currentUser.name }}</h2>
            <p class="user-signature">{{ userSignature }}</p>
          </div>
        </div>
        
        <!-- 统计信息 -->
        <div class="user-stats">
          <div class="stat-item">
            <span class="stat-number">{{ feedCount }}</span>
            <span class="stat-label">动态</span>
          </div>
          <div class="stat-item">
            <span class="stat-number">{{ friendCount }}</span>
            <span class="stat-label">好友</span>
          </div>
          <div class="stat-item">
            <span class="stat-number">{{ likeCount }}</span>
            <span class="stat-label">获赞</span>
          </div>
        </div>
      </div>
      
      <!-- 操作按钮 -->
      <div class="action-buttons">
        <button class="action-btn primary" @click="handleCreatePostClick">
          <i class="icon icon-plus"></i>
          <span>发动态</span>
        </button>
        <button class="action-btn secondary" @click="$emit('edit-profile')">
          <i class="icon icon-edit"></i>
          <span>编辑资料</span>
        </button>
      </div>
    </div>
    
    <!-- 快捷导航 -->
    <div class="quick-nav">
      <div class="nav-tabs">
        <div 
          v-for="tab in tabs" 
          :key="tab.key"
          :class="['nav-tab', { active: activeTab === tab.key }]"
          @click="switchTab(tab.key)"
        >
          <i :class="['tab-icon', tab.icon]"></i>
          <span class="tab-label">{{ tab.label }}</span>
          <span v-if="tab.count" class="tab-count">{{ tab.count }}</span>
        </div>
      </div>
    </div>
    
    <!-- 返回顶部按钮 -->
    <div 
      v-show="showBackToTop" 
      class="back-to-top-btn"
      @click="$emit('back-to-top')"
    >
      <i class="icon icon-arrow-up"></i>
      <span>最新</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFeedStore } from '../stores/feed'

const props = defineProps({
  activeTab: {
    type: String,
    default: 'all'
  }
})

const emit = defineEmits(['create-post', 'edit-profile', 'switch-tab', 'back-to-top'])

const authStore = useAuthStore()
const feedStore = useFeedStore()

const currentUser = computed(() => authStore.user || { 
  avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=You',
  name: '我'
})

// 用户信息
const userSignature = ref('✨ 记录生活中的美好瞬间 ✨')
const coverImage = ref('https://picsum.photos/seed/cover/800/200')

// 统计数据
const feedCount = computed(() => feedStore.getFriendFeeds.length + feedStore.getCommunityFeeds.length)
const friendCount = ref(42)
const likeCount = computed(() => {
  const allFeeds = [...feedStore.getFriendFeeds, ...feedStore.getCommunityFeeds]
  return allFeeds.reduce((total, feed) => total + (feed.likes?.length || 0), 0)
})

// 导航标签
const tabs = computed(() => [
  { key: 'all', label: '全部', icon: 'icon-grid', count: feedCount.value },
  { key: 'photo', label: '相册', icon: 'icon-image', count: Math.floor(feedCount.value * 0.6) },
  { key: 'video', label: '视频', icon: 'icon-video', count: Math.floor(feedCount.value * 0.3) },
  { key: 'article', label: '文章', icon: 'icon-document', count: Math.floor(feedCount.value * 0.1) }
])

// 返回顶部按钮状态
const showBackToTop = ref(false)

// sticky状态
const isSticky = ref(false)
const headerElement = ref(null)



// 滚动监听
function handleScroll() {
  const container = document.querySelector('.feed-type-list-main')
  if (container) {
    const scrollTop = container.scrollTop
    showBackToTop.value = scrollTop > 300
    
    // 根据滚动位置添加sticky类
    if (headerElement.value) {
      isSticky.value = scrollTop > 50
      if (isSticky.value) {
        headerElement.value.classList.add('sticky')
      } else {
        headerElement.value.classList.remove('sticky')
      }
    }
  }
}

function switchTab(tabKey) {
  emit('switch-tab', tabKey)
}

function handleCreatePostClick() {
  emit('create-post')
}



onMounted(() => {
  const container = document.querySelector('.feed-type-list-main')
  if (container) {
    container.addEventListener('scroll', handleScroll)
  }
})

onUnmounted(() => {
  const container = document.querySelector('.feed-type-list-main')
  if (container) {
    container.removeEventListener('scroll', handleScroll)
  }
})
</script>

<style scoped>
.user-profile-header {
  width: 100%;
  box-sizing: border-box;
  background: #fff;
  border-radius: 20px 20px 0 0; /* 只保留上边圆角，底部与消息气泡连接 */
  overflow: hidden;
  box-shadow: 
    0 20px 40px rgba(74, 140, 255, 0.08),
    0 8px 16px rgba(138, 105, 255, 0.05);
  border: 1px solid rgba(229, 231, 235, 0.2);
  backdrop-filter: blur(20px);
  margin: 0; /* 去除所有边距 */
  position: relative; /* 为绝对定位的按钮提供参考 */
}

.cover-background {
  position: relative;
  height: 200px;
  overflow: hidden;
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.cover-background:hover .cover-image {
  transform: scale(1.05);
}

.cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    135deg, 
    rgba(74, 140, 255, 0.7) 0%, 
    rgba(138, 105, 255, 0.6) 50%,
    rgba(74, 140, 255, 0.8) 100%
  );
  backdrop-filter: blur(2px);
}

.cover-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 24px 32px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  z-index: 2;
}

.user-main-info {
  display: flex;
  align-items: flex-end;
  gap: 20px;
}

.user-avatar-container {
  position: relative;
}

.user-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 4px solid rgba(255, 255, 255, 0.9);
  box-shadow: 
    0 8px 24px rgba(0, 0, 0, 0.15),
    0 0 0 2px rgba(74, 140, 255, 0.3);
  transition: all 0.3s ease;
  cursor: pointer;
}

.user-avatar:hover {
  transform: scale(1.1);
  border-color: rgba(255, 255, 255, 1);
  box-shadow: 
    0 12px 32px rgba(0, 0, 0, 0.2),
    0 0 0 3px rgba(74, 140, 255, 0.5);
}

.avatar-glow {
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: -1;
  filter: blur(8px);
}

.user-avatar:hover + .avatar-glow {
  opacity: 0.6;
}

.user-details {
  color: white;
  margin-bottom: 8px;
}

.user-name {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 4px 0;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  letter-spacing: 0.5px;
}

.user-signature {
  font-size: 14px;
  opacity: 0.9;
  margin: 0;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  font-weight: 500;
}

.user-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 8px;
}

.stat-item {
  text-align: center;
  color: white;
}

.stat-number {
  display: block;
  font-size: 20px;
  font-weight: 700;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.stat-label {
  display: block;
  font-size: 12px;
  opacity: 0.9;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  font-weight: 500;
  margin-top: 2px;
}

.action-buttons {
  position: absolute;
  top: 20px;
  right: 24px;
  display: flex;
  gap: 12px;
  z-index: 10; /* 确保按钮在其他元素之上 */
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
  position: relative;
  z-index: 5; /* 确保按钮本身可以被点击 */
  pointer-events: auto; /* 确保指针事件正常工作 */
}

.action-btn.primary {
  background: rgba(255, 255, 255, 0.95);
  color: #4a8cff;
  box-shadow: 0 4px 12px rgba(255, 255, 255, 0.2);
}

.action-btn.primary:hover {
  background: rgba(255, 255, 255, 1);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 255, 255, 0.3);
}

.action-btn.secondary {
  background: rgba(255, 255, 255, 0.15);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.action-btn.secondary:hover {
  background: rgba(255, 255, 255, 0.25);
  border-color: rgba(255, 255, 255, 0.5);
  transform: translateY(-2px);
}

.action-btn .icon {
  font-size: 16px;
}

.quick-nav {
  background: rgba(248, 250, 252, 0.8);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(229, 231, 235, 0.3);
  border-radius: 0; /* 去除圆角，与消息气泡连接 */
  border-bottom: none; /* 移除底部边框 */
}



.nav-tabs {
  display: flex;
  padding: 0 32px;
  position: relative;
}

.nav-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  color: #64748b;
  border-bottom: 3px solid transparent;
}

.nav-tab:hover {
  color: #4a8cff;
  background: rgba(74, 140, 255, 0.05);
}

.nav-tab.active {
  color: #4a8cff;
  border-bottom-color: #4a8cff;
  background: rgba(74, 140, 255, 0.08);
}

.tab-icon {
  font-size: 18px;
}

.tab-label {
  font-size: 14px;
  font-weight: 600;
}

.tab-count {
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  color: white;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 12px;
  min-width: 20px;
  text-align: center;
  box-shadow: 0 2px 6px rgba(74, 140, 255, 0.3);
}

.nav-tab.active .tab-count {
  background: linear-gradient(135deg, #5a9cff, #9a79ff);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}



/* 返回顶部按钮 */
.back-to-top-btn {
  position: absolute;
  top: 20px;
  left: 20px; /* 移到用户卡片左上角 */
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.9), rgba(138, 105, 255, 0.9));
  color: white;
  border: none;
  border-radius: 50px;
  padding: 10px 16px;
  cursor: pointer;
  box-shadow: 
    0 4px 16px rgba(74, 140, 255, 0.25),
    0 0 0 1px rgba(255, 255, 255, 0.3);
  backdrop-filter: blur(20px);
  transition: all 0.3s ease;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  opacity: 0.95; /* 稍微透明，避免过于突出 */
}

.back-to-top-btn:hover {
  transform: translateY(-2px) scale(1.05);
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  opacity: 1; /* 悬浮时完全不透明 */
  box-shadow: 
    0 6px 20px rgba(74, 140, 255, 0.4),
    0 0 0 2px rgba(255, 255, 255, 0.4);
}

.back-to-top-btn:active {
  transform: translateY(0) scale(1.02);
  box-shadow: 
    0 4px 12px rgba(74, 140, 255, 0.3),
    0 0 0 1px rgba(255, 255, 255, 0.15);
}

.back-to-top-btn .icon {
  font-size: 16px;
}



/* 响应式设计 */
@media (max-width: 768px) {
  .cover-content {
    padding: 16px 20px;
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .user-main-info {
    width: 100%;
    justify-content: space-between;
  }
  
  .user-stats {
    gap: 16px;
  }
  
  .action-buttons {
    position: static;
    margin-top: 12px;
  }
  
  .nav-tabs {
    padding: 0 16px;
  }
  
  .nav-tab {
    padding: 12px 8px;
    gap: 4px;
  }
  
  .tab-label {
    font-size: 12px;
  }
  

  
  .back-to-top-btn {
    top: 16px;
    left: 16px; /* 移动端保持在左上角 */
    padding: 8px 12px;
    font-size: 11px;
    gap: 4px;
  }
}
</style> 