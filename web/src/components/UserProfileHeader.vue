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
  border-radius: 12px 12px 0 0;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  border-bottom: none;
  margin: 0;
  position: relative;
}

.cover-background {
  position: relative;
  height: 180px;
  overflow: hidden;
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.45));
}

.cover-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20px 28px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  z-index: 2;
}

.user-main-info {
  display: flex;
  align-items: flex-end;
  gap: 16px;
}

.user-avatar-container {
  position: relative;
}

.user-avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  border: 3px solid #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  cursor: pointer;
  transition: opacity 0.15s;
}

.user-avatar:hover {
  opacity: 0.9;
}

.avatar-glow {
  display: none;
}

.user-details {
  color: white;
  margin-bottom: 6px;
}

.user-name {
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 2px 0;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.user-signature {
  font-size: 13px;
  opacity: 0.9;
  margin: 0;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.user-stats {
  display: flex;
  gap: 20px;
  margin-bottom: 6px;
}

.stat-item {
  text-align: center;
  color: white;
}

.stat-number {
  display: block;
  font-size: 18px;
  font-weight: 700;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.stat-label {
  display: block;
  font-size: 11px;
  opacity: 0.85;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  margin-top: 2px;
}

.action-buttons {
  position: absolute;
  top: 16px;
  right: 20px;
  display: flex;
  gap: 8px;
  z-index: 10;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 7px 14px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  pointer-events: auto;
}

.action-btn.primary {
  background: rgba(255, 255, 255, 0.95);
  color: #374151;
}

.action-btn.primary:hover {
  background: #fff;
}

.action-btn.secondary {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.action-btn.secondary:hover {
  background: rgba(255, 255, 255, 0.3);
}

.action-btn .icon {
  font-size: 14px;
}

.quick-nav {
  background: #fafbfc;
  border-top: 1px solid #f3f4f6;
}

.nav-tabs {
  display: flex;
  padding: 0 28px;
}

.nav-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px 8px;
  cursor: pointer;
  transition: color 0.15s;
  color: #6b7280;
  border-bottom: 2px solid transparent;
}

.nav-tab:hover {
  color: #111827;
}

.nav-tab.active {
  color: #2563eb;
  border-bottom-color: #2563eb;
}

.tab-icon {
  font-size: 16px;
}

.tab-label {
  font-size: 13px;
  font-weight: 500;
}

.tab-count {
  background-color: #e5e7eb;
  color: #374151;
  font-size: 11px;
  font-weight: 600;
  padding: 1px 7px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

.nav-tab.active .tab-count {
  background-color: #eff6ff;
  color: #2563eb;
}

/* Back to top */
.back-to-top-btn {
  position: absolute;
  top: 16px;
  left: 16px;
  background: rgba(255, 255, 255, 0.95);
  color: #374151;
  border: none;
  border-radius: 8px;
  padding: 8px 14px;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  transition: background-color 0.15s;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
}

.back-to-top-btn:hover {
  background: #fff;
}

.back-to-top-btn .icon {
  font-size: 14px;
}

/* Responsive */
@media (max-width: 768px) {
  .cover-content {
    padding: 14px 16px;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .user-main-info {
    width: 100%;
    justify-content: space-between;
  }

  .user-stats {
    gap: 14px;
  }

  .action-buttons {
    position: static;
    margin-top: 8px;
  }

  .nav-tabs {
    padding: 0 12px;
  }

  .nav-tab {
    padding: 10px 6px;
    gap: 4px;
  }

  .tab-label {
    font-size: 12px;
  }

  .back-to-top-btn {
    top: 12px;
    left: 12px;
    padding: 6px 10px;
    font-size: 11px;
  }
}
</style> 