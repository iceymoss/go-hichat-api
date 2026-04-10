<template>
  <div class="notifications-view">
    <!-- 头部标题栏 -->
    <div class="notifications-header">
      <button class="back-btn" @click="goBackToFeed">
        <i class="icon icon-arrow-left"></i>
      </button>
      <h1 class="page-title">动态消息</h1>
      <button class="clear-all-btn" @click="clearAllNotifications">
        <i class="icon icon-trash"></i>
        清空
      </button>
    </div>

    <!-- 消息列表 -->
    <div class="notifications-content">
      <div class="notifications-list">
        <div 
          v-for="notification in notifications" 
          :key="notification.id"
          class="notification-item"
          :class="{ 'unread': !notification.read }"
          @click="handleNotificationClick(notification)"
        >
          <div class="notification-avatar">
            <img :src="notification.avatar" :alt="notification.user" />
            <div class="notification-type-badge" :class="notification.type">
              <i class="icon" :class="getTypeIcon(notification.type)"></i>
            </div>
          </div>
          
          <div class="notification-content">
            <div class="notification-text">
              <span class="notification-user">{{ notification.user }}</span>
              <span class="notification-action">{{ notification.action }}</span>
              <span class="notification-target">{{ notification.target }}</span>
            </div>
            
            <div class="notification-preview" v-if="notification.preview">
              {{ notification.preview }}
            </div>
            
            <div class="notification-time">{{ notification.time }}</div>
          </div>
          
          <div class="notification-indicator" v-if="!notification.read">
            <div class="red-dot"></div>
          </div>
        </div>
      </div>
      
      <!-- 空状态 -->
      <div v-if="!notifications.length" class="empty-state">
        <div class="empty-icon">
          <i class="icon icon-bell"></i>
        </div>
        <div class="empty-text">暂无新消息</div>
        <div class="empty-desc">点赞和评论会出现在这里</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// 模拟消息数据
const notifications = ref([
  {
    id: 1,
    type: 'like',
    user: '小明',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaoming',
    action: '赞了你的',
    target: '动态',
    preview: '今天天气真不错，适合出去走走...',
    time: '3分钟前',
    read: false,
    feedId: 1
  },
  {
    id: 2,
    type: 'comment',
    user: '小红',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaohong',
    action: '评论了你的',
    target: '动态',
    preview: '我也觉得！一起出去吧',
    time: '10分钟前',
    read: false,
    feedId: 2
  },
  {
    id: 3,
    type: 'reply',
    user: '李华',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=lihua',
    action: '回复了你的',
    target: '评论',
    preview: '回复: 好的，约个时间',
    time: '1小时前',
    read: false,
    feedId: 3
  },
  {
    id: 4,
    type: 'like',
    user: '张三',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=zhangsan',
    action: '赞了你的',
    target: '评论',
    preview: '评论: 这个想法很棒！',
    time: '2小时前',
    read: true,
    feedId: 4
  },
  {
    id: 5,
    type: 'comment',
    user: '王五',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=wangwu',
    action: '评论了你的',
    target: '动态',
    preview: '太美了！在哪里拍的？',
    time: '昨天',
    read: true,
    feedId: 5
  }
])

// 获取消息类型图标
function getTypeIcon(type) {
  const icons = {
    like: 'icon-heart',
    comment: 'icon-message-circle',
    reply: 'icon-corner-up-left'
  }
  return icons[type] || 'icon-bell'
}

// 处理消息点击
function handleNotificationClick(notification) {
  // 标记为已读
  notification.read = true
  
  // 跳转到对应的动态详情页
  // TODO: 实现跳转逻辑
  console.log('跳转到动态:', notification.feedId)
  
  // 可以跳转到主页的动态列表
  router.push('/app/discover')
}

// 清空所有消息
function clearAllNotifications() {
  notifications.value = []
}

// 返回动态列表
function goBackToFeed() {
  router.push('/app/discover')
}

onMounted(() => {
  // 页面加载时标记所有消息为已读
  // notifications.value.forEach(notification => {
  //   notification.read = true
  // })
})
</script>

<style scoped>
.notifications-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg);
}

.notifications-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--card);
  border-bottom: 1px solid var(--border);
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  background: var(--accent-muted);
  border-radius: var(--radius-full);
  color: var(--accent);
  cursor: pointer;
  transition: all 0.2s ease;
}

.back-btn:hover {
  background: var(--accent);
  color: var(--bg);
}

.page-title {
  font-size: 18px;
  font-family: var(--font-display);
  font-weight: 700;
  color: var(--fg);
  margin: 0;
}

.clear-all-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: none;
  background: rgba(192, 57, 43, 0.12);
  border-radius: var(--radius-full);
  color: var(--danger);
  font-size: 14px;
  font-family: var(--font-ui);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.clear-all-btn:hover {
  background: rgba(192, 57, 43, 0.22);
}

.notifications-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.notifications-list {
  max-width: 600px;
  margin: 0 auto;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 16px 20px;
  margin-bottom: 12px;
  background: var(--card);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid var(--border);
  position: relative;
}

.notification-item:hover {
  background: var(--card-hover);
  border-color: var(--accent);
}

.notification-item.unread {
  border-left: 3px solid var(--accent);
}

.notification-avatar {
  position: relative;
  flex-shrink: 0;
}

.notification-avatar img {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  border: 2px solid var(--border);
}

.notification-type-badge {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--card);
}

.notification-type-badge.like {
  background: var(--danger);
  color: white;
}

.notification-type-badge.comment {
  background: var(--accent);
  color: var(--bg);
}

.notification-type-badge.reply {
  background: var(--success);
  color: white;
}

.notification-type-badge .icon {
  font-size: 10px;
}

.notification-content {
  flex: 1;
}

.notification-text {
  font-size: 15px;
  font-family: var(--font-ui);
  line-height: 1.5;
  color: var(--muted);
  margin-bottom: 8px;
}

.notification-user {
  font-weight: 700;
  color: var(--fg);
}

.notification-action {
  color: var(--muted);
  margin: 0 4px;
}

.notification-target {
  color: var(--fg);
  font-weight: 500;
}

.notification-preview {
  font-size: 14px;
  color: var(--muted);
  line-height: 1.4;
  margin-bottom: 8px;
  padding: 8px 12px;
  background: var(--input-bg);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--border);
}

.notification-time {
  font-size: 12px;
  color: var(--muted);
  font-family: var(--font-ui);
  font-weight: 500;
}

.notification-indicator {
  position: absolute;
  top: 16px;
  right: 16px;
}

.red-dot {
  width: 8px;
  height: 8px;
  background: var(--accent);
  border-radius: 50%;
  animation: pulse-dot 2s infinite;
}

@keyframes pulse-dot {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.2);
  }
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  width: 80px;
  height: 80px;
  background: var(--accent-muted);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.empty-icon .icon {
  font-size: 32px;
  color: var(--muted);
}

.empty-text {
  font-size: 18px;
  font-family: var(--font-display);
  font-weight: 600;
  color: var(--fg);
  margin-bottom: 8px;
}

.empty-desc {
  font-size: 14px;
  font-family: var(--font-ui);
  color: var(--muted);
}

/* Responsive */
@media (max-width: 768px) {
  .notifications-header {
    padding: 12px 16px;
  }

  .notifications-content {
    padding: 12px 16px;
  }

  .notification-item {
    padding: 12px 16px;
    gap: 12px;
  }

  .notification-avatar img {
    width: 40px;
    height: 40px;
  }

  .notification-type-badge {
    width: 16px;
    height: 16px;
  }

  .notification-type-badge .icon {
    font-size: 8px;
  }
}
</style> 