<template>
  <div class="feed-type-list-main">
    <div class="feed-type-list-center" :class="{ 'has-notifications': hasUnreadNotifications, 'no-notifications': !hasUnreadNotifications }">
      <!-- 用户个人信息头部 -->
      <UserProfileHeader 
        :active-tab="activeContentTab"
        @create-post="handleCreatePost"
        @edit-profile="handleEditProfile"
        @switch-tab="handleSwitchTab"
        @back-to-top="handleBackToTop"
      />
      
      <!-- 消息通知徽章 - 放在用户卡片和动态列表之间 -->
      <div class="message-notification-area" v-if="hasUnreadNotifications" :class="{ 'sticky-active': isNotificationSticky }">
        <div class="message-badge" @click.stop="goToNotifications">
          <div class="badge-content">
            <div class="badge-icon">
              <i class="icon icon-bell"></i>
            </div>
            <div class="badge-text">你有 {{ unreadCount }} 条新消息</div>
            <div class="badge-arrow">
              <i class="icon icon-chevron-right"></i>
            </div>
          </div>
          <div class="badge-indicator">
            <div class="red-dot">{{ unreadCount }}</div>
          </div>
        </div>
      </div>
      
      <div
        v-for="feed in displayFeeds"
        :key="feed.id"
        class="feed-type-feed-item"
        @click="selectFeed(feed)"
      >
        <div class="feed-header">
          <img 
            :src="feed.user.avatar" 
            class="avatar" 
            @mouseenter="(e) => showUserProfile(feed.user, e)"
            @mouseleave="hideUserProfile"
          />
          <div class="user-info">
            <div 
              class="name"
              @mouseenter="(e) => showUserProfile(feed.user, e)"
              @mouseleave="hideUserProfile"
            >{{ feed.user.name }}</div>
            <div class="time">{{ feed.time }}</div>
          </div>
        </div>
        <div class="feed-content" @click="selectFeed(feed)">
          <div class="feed-text">{{ feed.content }}</div>
          <!-- 兼容旧的单图片格式 -->
          <div class="feed-image" v-if="feed.image && (!feed.images || feed.images.length === 0)">
            <img :src="feed.image" alt="动态图片" @click.stop="viewImage(feed, 0)" />
          </div>
          <!-- 新的多图片格式 -->
          <div class="feed-images" v-if="feed.images && feed.images.length > 0">
            <div :class="['images-grid', getImageGridClass(feed.images.length)]">
              <div
                v-for="(image, index) in feed.images.slice(0, 9)"
                :key="index"
                class="image-item"
                @click.stop="viewImage(feed, index)"
              >
                <img :src="image" alt="动态图片" class="feed-image-item">
                <div v-if="feed.images.length > 9 && index === 8" class="more-images-count">
                  +{{ feed.images.length - 9 }}
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="feed-likes" v-if="feed.likes && feed.likes.length" @click.stop>
          <div class="likes-container">
            <div class="likes-avatars">
              <img 
                v-for="(user, index) in feed.likes.slice(0, 15)" 
                :key="user"
                :src="`https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(user)}`"
                :alt="user"
                class="like-avatar"
                @mouseenter="(e) => showUserProfile({ name: user, avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(user)}` }, e)"
                @mouseleave="hideUserProfile"
              />
              <span v-if="feed.likes.length > 15" class="like-more">+{{ feed.likes.length - 15 }}</span>
            </div>
            <div class="likes-text">
              <span v-for="(user, index) in feed.likes.slice(0, 15)" :key="user" class="like-user">
                <span 
                  class="like-user-highlight"
                  @mouseenter="(e) => showUserProfile({ name: user, avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(user)}` }, e)"
                  @mouseleave="hideUserProfile"
                >{{ user }}</span>{{ index < Math.min(14, feed.likes.length - 1) ? '、' : '' }}
              </span>
              <span v-if="feed.likes.length > 15">等{{ feed.likes.length }}人觉得很赞</span>
              <span v-else>觉得很赞</span>
            </div>
          </div>
        </div>
        
        <div class="feed-actions" @click.stop>
          <button class="action-btn like-btn" :class="{ active: feed.liked }" @click.stop="toggleLike(feed)">
            <i class="icon" :class="feed.liked ? 'icon-liked' : 'icon-like'"></i>
            <span>{{ feed.liked ? '已点赞' : '点赞' }}</span>
          </button>
          <button class="action-btn comment-btn" @click.stop="focusCommentInput(feed)">
            <i class="icon icon-comment"></i>
            <span>评论</span>
          </button>
        </div>
        
        <div class="feed-comment-input" @click.stop>
          <div class="comment-input-container">
            <img :src="currentUser.avatar" alt="头像" class="comment-avatar">
            <div class="comment-input-wrapper">
              <input
                ref="commentInput"
                type="text"
                v-model="commentText"
                placeholder="写评论..."
                @keyup.enter="submitComment(feed)"
                class="comment-input"
                @click.stop
              />
              <button @click.stop="submitComment(feed)" class="comment-submit-btn">发送</button>
            </div>
          </div>
        </div>
        
        <div class="feed-comments" v-if="feed.comments && feed.comments.length" @click.stop>
          <CommentThread 
            :comments="feed.comments"
            :format-comment="formatComment"
            :show-user-profile="showUserProfile"
            :hide-user-profile="hideUserProfile"
            @reply="(data) => handleCommentReply(feed, data)"
            @like-comment="(comment) => handleCommentLike(feed, comment)"
            @like-reply="(reply) => handleReplyLike(feed, reply)"
            @delete-comment="(comment) => deleteComment(feed, comment)"
            @delete-reply="(data) => deleteReply(feed, data.comment, data.reply)"
          />
        </div>
      </div>
    </div>
    
    <!-- 加载和完成状态也要居中 -->
    <div v-if="loading" class="feed-loading">加载中...</div>
    <div v-if="noMore" class="feed-nomore">没有更多了</div>
  </div>

  <!-- 消息覆盖层 -->
  <Transition name="message-overlay">
    <div v-if="showMessageOverlay" class="message-overlay" @click.self="closeMessageOverlay">
      <div class="message-panel">
        <!-- 头部 -->
        <div class="message-header">
          <div class="header-content">
            <div class="header-icon">
              <i class="icon icon-bell"></i>
            </div>
            <div class="header-text">
              <h2 class="header-title">动态消息</h2>
              <p class="header-subtitle">{{ notifications.filter(n => !n.read).length }} 条未读消息</p>
            </div>
          </div>
          <div class="header-actions">
            <button class="action-button clear-button" @click="clearAllNotifications">
              <i class="icon icon-trash"></i>
              <span>清空全部</span>
            </button>
            <button class="close-button" @click="closeMessageOverlay" title="关闭">
              <i class="icon icon-x"></i>
              <span class="close-text">关闭</span>
            </button>
          </div>
        </div>

        <!-- 消息列表 -->
        <div class="message-list">
          <div v-if="notifications.length === 0" class="empty-state">
            <div class="empty-icon">
              <i class="icon icon-bell-off"></i>
            </div>
            <div class="empty-text">暂无消息</div>
            <div class="empty-desc">点赞和评论会出现在这里</div>
          </div>

          <div 
            v-for="(notification, index) in notifications" 
            :key="notification.id"
            class="notification-item"
            :class="{ 'unread': !notification.read }"
            @click="handleNotificationClick(notification)"
            :style="{ animationDelay: `${index * 0.1}s` }"
          >
            <div class="notification-avatar">
              <img 
                :src="notification.avatar" 
                :alt="notification.user"
                @mouseenter="(e) => showUserProfile({ name: notification.user, avatar: notification.avatar }, e)"
                @mouseleave="hideUserProfile"
              />
              <div class="notification-type-badge" :class="notification.type">
                <i class="icon" :class="getTypeIcon(notification.type)"></i>
              </div>
            </div>
            
            <div class="notification-content">
              <div class="notification-text">
                <span 
                  class="notification-user"
                  @mouseenter="(e) => showUserProfile({ name: notification.user, avatar: notification.avatar }, e)"
                  @mouseleave="hideUserProfile"
                >{{ notification.user }}</span>
                <span class="notification-action">{{ notification.action }}</span>
                <span class="notification-target">{{ notification.target }}</span>
              </div>
              
              <div class="notification-preview" v-if="notification.preview">
                {{ notification.preview }}
              </div>
              
              <div class="notification-time">{{ notification.time }}</div>
            </div>
            
            <div class="notification-indicator" v-if="!notification.read">
              <div class="unread-dot"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>

  <!-- 发布动态弹窗 -->
  <CreatePostModal
    :visible="showCreatePostModal"
    @close="closeCreatePostModal"
    @published="handlePostPublished"
  />

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
</template>
<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFeedStore } from '../stores/feed'
import CommentThread from './CommentThread.vue'
import UserProfileHeader from './UserProfileHeader.vue'
import CreatePostModal from './CreatePostModal.vue'
import ImageViewer from './ImageViewer.vue'
import UserProfileCard from './UserProfileCard.vue'

const props = defineProps({ feeds: Array })
const emit = defineEmits(['select', 'like', 'comment'])

const router = useRouter()
const authStore = useAuthStore()
const feedStore = useFeedStore()
const currentUser = computed(() => authStore.user || { 
  avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=You',
  name: '我'
})

// 消息通知数据
const unreadCount = ref(10)
const hasUnreadNotifications = computed(() => unreadCount.value > 0)
const isNotificationSticky = ref(false)
const showMessageOverlay = ref(false)
const showCreatePostModal = ref(false)
const showImageViewer = ref(false)
const viewerImages = ref([])
const viewerInitialIndex = ref(0)

const activeCommentFeed = ref(null)
const commentText = ref('')
const commentInput = ref(null)

// 用户资料卡片状态
const profileCard = ref({
  visible: false,
  user: null,
  position: { x: 0, y: 0 },
  arrowPosition: 'top'
})

let hoverTimeout = null

// 内容标签页状态
const activeContentTab = ref('all')

// 无限滚动相关
const displayFeeds = ref([])
const feedPageSize = 10
const feedPage = ref(1)
const loading = ref(false)
const noMore = ref(false)

function loadFeeds() {
  if (loading.value || noMore.value) return
  loading.value = true
  nextTick(() => {
    const start = 0
    const end = feedPage.value * feedPageSize
    displayFeeds.value = props.feeds.slice(0, end)
    if (displayFeeds.value.length >= props.feeds.length) {
      noMore.value = true
    }
    loading.value = false
    feedPage.value++
  })
}

function handleScroll() {
  const container = document.querySelector('.feed-type-list-main')
  if (!container) return
  
  // 检查是否需要加载更多动态
  const threshold = 40
  if (container.scrollTop + container.clientHeight >= container.scrollHeight - threshold) {
    loadFeeds()
  }
  
  // 检查消息气泡的sticky状态
  const scrollTop = container.scrollTop
  const userHeader = document.querySelector('.user-profile-header')
  const headerHeight = userHeader ? userHeader.offsetHeight : 288
  
  // 当滚动超过用户卡片高度时激活sticky效果
  isNotificationSticky.value = scrollTop > headerHeight / 2
  
  // 动态更新消息气泡的top位置
  updateNotificationPosition(headerHeight)
}

function updateNotificationPosition(headerHeight) {
  const notificationArea = document.querySelector('.message-notification-area')
  if (notificationArea) {
    // 确保消息气泡始终贴在用户卡片下方
    notificationArea.style.top = `${headerHeight - 1}px`
  }
}

onMounted(() => {
  loadFeeds()
  nextTick(() => {
    const container = document.querySelector('.feed-type-list-main')
    if (container) {
      container.addEventListener('scroll', handleScroll)
    }
    
    // 初始化消息气泡位置
    const userHeader = document.querySelector('.user-profile-header')
    if (userHeader) {
      updateNotificationPosition(userHeader.offsetHeight)
    }
  })
})
onBeforeUnmount(() => {
  const container = document.querySelector('.feed-type-list-main')
  if (container) {
    container.removeEventListener('scroll', handleScroll)
  }
})
watch(() => props.feeds, (newFeeds, oldFeeds) => {
  // 检查是否是新动态添加（第一个动态发生变化）
  const isNewPost = newFeeds && oldFeeds && newFeeds.length > oldFeeds.length && 
                   newFeeds[0]?.id !== oldFeeds[0]?.id
  
  if (isNewPost) {
    // 新动态添加，保持当前页面位置，只更新显示的动态
    const currentDisplayCount = displayFeeds.value.length
    displayFeeds.value = newFeeds.slice(0, Math.max(currentDisplayCount, feedPageSize))
  } else {
    // 切换类型或其他情况，重置所有状态
    feedPage.value = 1
    noMore.value = false
    displayFeeds.value = []
    loadFeeds()
  }
}, { immediate: true })

// 监听消息通知状态变化，更新布局
watch(hasUnreadNotifications, () => {
  nextTick(() => {
    const userHeader = document.querySelector('.user-profile-header')
    if (userHeader) {
      updateNotificationPosition(userHeader.offsetHeight)
    }
  })
})

function selectFeed(feed) {
  emit('select', { ...feed, id: String(feed.id) })
}

function formatComment(content) {
  // 高亮@用户
  return content.replace(/@([\u4e00-\u9fa5\w\-]+)/g, '<span class="at-user">@$1</span>')
}

function toggleLike(feed) {
  emit('like', feed.id)
}

function focusCommentInput(feed) {
  activeCommentFeed.value = feed.id
  commentText.value = ''
  // 延迟聚焦，确保DOM更新完成
  setTimeout(() => {
    if (commentInput.value) {
      commentInput.value.focus()
    }
  }, 100)
}

function submitComment(feed) {
  if (!commentText.value.trim()) return
  emit('comment', {
    feedId: feed.id,
    content: commentText.value,
    user: currentUser.value.name,
    avatar: currentUser.value.avatar
  })
  commentText.value = ''
  activeCommentFeed.value = null
}

function handleCommentReply(feed, { comment, reply }) {
  const type = feed.type || 'friend'
  feedStore.addReply(feed.id, comment.id, reply.content, type)
}

function handleCommentLike(feed, comment) {
  const type = feed.type || 'friend'
  feedStore.toggleCommentLike(feed.id, comment.id, type)
}

function handleReplyLike(feed, reply) {
  const comment = feed.comments.find(c => 
    c.replies && c.replies.some(r => r.id === reply.id)
  )
  if (!comment) return
  
  const type = feed.type || 'friend'
  feedStore.toggleReplyLike(feed.id, comment.id, reply.id, type)
}

function deleteComment(feed, comment) {
  // 使用feed store删除评论
  const type = feed.type || 'friend'
  feedStore.deleteComment(feed.id, comment.id, type)
}

function deleteReply(feed, comment, reply) {
  // 使用feed store删除回复
  const type = feed.type || 'friend'
  feedStore.deleteReply(feed.id, comment.id, reply.id, type)
}

// 用户个人信息头部事件处理
function handleCreatePost() {
  showCreatePostModal.value = true
}

// 关闭发布动态弹窗
function closeCreatePostModal() {
  showCreatePostModal.value = false
}

// 处理发布动态成功
function handlePostPublished(newPost) {
  // 重新加载feeds以确保新动态显示
  feedPage.value = 1
  noMore.value = false
  loadFeeds()
  
  // 滚动到顶部查看新动态
  handleBackToTop()
}

// 获取图片网格布局类名
function getImageGridClass(count) {
  if (count === 1) return 'grid-single'
  if (count === 2) return 'grid-double'
  if (count === 3) return 'grid-triple'
  if (count === 4) return 'grid-quad'
  return 'grid-multi'
}

// 查看图片
function viewImage(feed, index) {
  // 收集当前动态的所有图片
  const images = []
  
  // 如果有新的多图片格式
  if (feed.images && feed.images.length > 0) {
    images.push(...feed.images)
  } 
  // 如果只有旧的单图片格式
  else if (feed.image) {
    images.push(feed.image)
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

function handleEditProfile() {
  // TODO: 打开编辑资料对话框
  console.log('编辑个人资料')
}

function handleSwitchTab(tabKey) {
  activeContentTab.value = tabKey
  console.log('切换到标签页:', tabKey)
  // TODO: 根据标签页筛选动态内容
}

function handleBackToTop() {
  const container = document.querySelector('.feed-type-list-main')
  if (container) {
    container.scrollTo({
      top: 0,
      behavior: 'smooth'
    })
  }
}

// 模拟消息数据
const notifications = ref([
  {
    id: 1,
    type: 'like',
    user: '小明',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaoming',
    action: '赞了你的',
    target: '动态',
    preview: '今天天气真不错，适合出去走走，看看外面的风景，感受大自然的美好...',
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
    preview: '我也觉得！一起出去吧，我知道一个很不错的地方，风景特别美',
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
    preview: '回复: 好的，约个时间，我这周末有空，咱们可以一起去看看',
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
    preview: '评论: 这个想法很棒！我觉得这样的活动很有意义，大家一起参与',
    time: '2小时前',
    read: false,
    feedId: 4
  },
  {
    id: 5,
    type: 'comment',
    user: '王五',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=wangwu',
    action: '评论了你的',
    target: '动态',
    preview: '太美了！在哪里拍的？看起来像是某个景区，色彩很丰富',
    time: '3小时前',
    read: false,
    feedId: 5
  },
  {
    id: 6,
    type: 'reply',
    user: '小刘',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaoliu',
    action: '回复了你的',
    target: '评论',
    preview: '回复: 确实很棒，我上次也去过类似的地方，体验很不错',
    time: '4小时前',
    read: false,
    feedId: 6
  },
  {
    id: 7,
    type: 'like',
    user: '小陈',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaochen',
    action: '赞了你的',
    target: '动态',
    preview: '分享的这篇文章很有深度，让我学到了很多新的知识和观点...',
    time: '5小时前',
    read: false,
    feedId: 7
  },
  {
    id: 8,
    type: 'comment',
    user: '小李',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaoli',
    action: '评论了你的',
    target: '动态',
    preview: '这个地方我也想去，看起来很有趣，可以分享一下具体位置吗？',
    time: '6小时前',
    read: false,
    feedId: 8
  },
  {
    id: 9,
    type: 'reply',
    user: '小张',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaozhang',
    action: '回复了你的',
    target: '评论',
    preview: '回复: 我也有同样的想法，咱们可以组个团一起去，人多比较热闹',
    time: '8小时前',
    read: false,
    feedId: 9
  },
  {
    id: 10,
    type: 'like',
    user: '小赵',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaozhao',
    action: '赞了你的',
    target: '评论',
    preview: '评论: 说得很有道理，这个观点我很赞同，值得大家深入思考',
    time: '10小时前',
    read: false,
    feedId: 10
  },
  {
    id: 11,
    type: 'comment',
    user: '小王',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaowang',
    action: '评论了你的',
    target: '动态',
    preview: '照片拍得真好，构图和色彩都很棒，用的什么设备拍的？',
    time: '昨天',
    read: true,
    feedId: 11
  },
  {
    id: 12,
    type: 'reply',
    user: '小孙',
    avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=xiaosun',
    action: '回复了你的',
    target: '评论',
    preview: '回复: 谢谢夸奖，就是用手机拍的，主要是光线和角度比较好',
    time: '昨天',
    read: true,
    feedId: 12
  }
])

// 消息通知相关函数
function goToNotifications() {
  console.log('点击消息徽章，显示覆盖层消息列表')
  // 显示消息覆盖层
  showMessageOverlay.value = true
  // 清除未读消息计数
  unreadCount.value = 0
}

function closeMessageOverlay() {
  showMessageOverlay.value = false
}

function handleNotificationClick(notification) {
  // 标记为已读
  notification.read = true
  // 跳转到对应的动态详情或其他操作
  console.log('点击消息:', notification)
  // 可以在这里添加跳转逻辑
  closeMessageOverlay()
}

function clearAllNotifications() {
  // 添加确认提示
  if (notifications.value.length === 0) {
    console.log('没有消息需要清空')
    return
  }
  
  // 清空所有消息
  notifications.value = []
  // 重置未读消息数量
  unreadCount.value = 0
  // 添加反馈提示
  console.log('已清空所有消息')
  // 不关闭覆盖层，让用户看到清空结果
  // closeMessageOverlay()
}

// 获取消息类型图标
function getTypeIcon(type) {
  const icons = {
    like: 'icon-heart',
    comment: 'icon-message-circle',
    reply: 'icon-corner-up-left'
  }
  return icons[type] || 'icon-bell'
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
.feed-type-list-main {
  padding: 0;
  overflow-y: auto;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  background: var(--bg);
  position: relative;
  width: 100%;
  box-sizing: border-box;
  font-family: var(--font-ui);
}
.feed-type-list-center {
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 0 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-sizing: border-box;
}
.feed-type-feed-item {
  background: var(--card);
  border-radius: var(--radius-md);
  margin: 0 0 24px 0;
  padding: 20px 24px 16px 24px;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--border);
  position: relative;
}
.feed-type-feed-item:hover {
  background: var(--card-hover);
  border-color: var(--accent);
  box-shadow: 0 10px 28px rgba(0,0,0,0.3);
  transform: translateY(-3px);
}
.feed-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}
.avatar {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  margin-right: 14px;
  border: 2px solid var(--border);
  cursor: pointer;
  transition: all 0.3s ease;
}

.avatar:hover {
  transform: scale(1.05);
  border-color: var(--accent);
  box-shadow: 0 4px 16px rgba(200,149,108,0.2);
}
.user-info .name {
  color: var(--fg);
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 4px;
  padding: 2px 4px;
  margin: -2px -4px;
}

.user-info .name:hover {
  color: var(--accent);
  background: var(--accent-muted);
  transform: scale(1.02);
}
.user-info .time {
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
}
.feed-content {
  color: var(--fg);
  font-size: 16px;
  line-height: 1.6;
  margin-bottom: 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  padding: 8px;
  border-radius: var(--radius-sm);
}
.feed-content:hover {
  background-color: var(--accent-muted);
}
.feed-text {
  margin-bottom: 8px;
}
.feed-image {
  margin-top: 8px;
  text-align: center;
}
.feed-image img {
  max-width: 100%;
  max-height: 220px;
  border-radius: var(--radius-sm);
  object-fit: cover;
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all 0.3s ease;
}

.feed-image img:hover {
  transform: scale(1.02);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
  border-color: var(--accent);
}

/* 多图片网格布局 */
.feed-images {
  margin-top: 12px;
}

.images-grid {
  display: grid;
  gap: 4px;
  border-radius: var(--radius-md);
  overflow: hidden;
  max-width: 100%;
}

.images-grid.grid-single {
  grid-template-columns: 1fr;
  max-width: 400px;
}

.images-grid.grid-double {
  grid-template-columns: 1fr 1fr;
  max-width: 400px;
}

.images-grid.grid-triple {
  grid-template-columns: 2fr 1fr;
  grid-template-rows: 1fr 1fr;
  max-width: 400px;
}

.images-grid.grid-triple .image-item:first-child {
  grid-row: 1 / 3;
}

.images-grid.grid-quad {
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  max-width: 400px;
}

.images-grid.grid-multi {
  grid-template-columns: repeat(3, 1fr);
  max-width: 400px;
}

.image-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  background: var(--input-bg);
}

.image-item:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.feed-image-item {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s ease;
}

.image-item:hover .feed-image-item {
  transform: scale(1.05);
}

.more-images-count {
  position: absolute;
  inset: 0;
  background: var(--overlay);
  color: var(--fg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
}
.feed-likes {
  margin-bottom: 8px;
  padding: 10px 16px;
  background: var(--accent-muted);
  border-radius: var(--radius-md);
  border-left: 3px solid var(--accent);
  display: flex;
  align-items: center;
  gap: 10px;
}
.likes-container {
  display: flex;
  align-items: center;
  gap: 12px;
}
.likes-avatars {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 3px;
}
.like-avatar {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  border: 2px solid var(--border);
  background: var(--input-bg);
  flex-shrink: 0;
  margin-right: -6px;
  z-index: 1;
  cursor: pointer;
  transition: all 0.3s ease;
}

.like-avatar:hover {
  transform: scale(1.2);
  box-shadow: 0 4px 12px rgba(200,149,108,0.3);
  border-color: var(--accent);
  z-index: 10;
}
.like-user-highlight {
  color: var(--accent);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 2px 4px;
  border-radius: 4px;
  margin: -2px -4px;
}

.like-user-highlight:hover {
  color: var(--accent-hover);
  background: var(--accent-muted);
  transform: scale(1.05);
}
.like-more {
  color: var(--accent);
  font-size: 11px;
  margin-left: 4px;
  font-weight: 600;
  background: var(--accent-muted);
  padding: 3px 8px;
  border-radius: var(--radius-full);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
}
.likes-text {
  color: var(--fg);
  font-size: 13px;
  font-weight: 500;
  flex: 1;
}
.like-user {
  color: var(--accent);
  font-weight: 600;
  margin-right: 4px;
}
/* 评论容器样式 */
.feed-comments {
  margin-bottom: 8px;
  padding: 8px 12px;
  background: var(--input-bg);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--muted);
}
/* @用户样式 */
:deep(.at-user) {
  color: var(--accent);
  font-weight: 600;
}
.feed-actions {
  display: flex;
  gap: 0;
  color: var(--muted);
  font-size: 15px;
  margin-top: 8px;
  border-top: 1px solid var(--border-light);
  padding-top: 12px;
}
.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  background: none;
  border: none;
  color: var(--muted);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: var(--radius-sm);
  position: relative;
  overflow: hidden;
  font-family: var(--font-ui);
}
.action-btn:hover {
  color: var(--accent);
  background: var(--accent-muted);
  transform: translateY(-1px);
}
.action-btn.active {
  color: var(--accent);
  background: var(--accent-muted);
}
.action-btn .icon {
  font-size: 18px;
  position: relative;
  z-index: 1;
}
.action-btn span {
  position: relative;
  z-index: 1;
  font-weight: 600;
}
.feed-comment-input {
  margin-top: 12px;
  padding: 12px;
  background: var(--input-bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}
.comment-input-container {
  display: flex;
  align-items: center;
  gap: 10px;
}
.comment-input-container .comment-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  flex-shrink: 0;
}
.comment-input-wrapper {
  flex: 1;
  display: flex;
  gap: 8px;
}
.comment-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s ease;
  background: var(--card);
  color: var(--fg);
  font-family: var(--font-ui);
}
.comment-input::placeholder {
  color: var(--muted);
}
.comment-input:focus {
  border-color: var(--accent);
}
.comment-submit-btn {
  padding: 8px 16px;
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-full);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
  font-family: var(--font-ui);
}
.comment-submit-btn:hover {
  background: var(--accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(200,149,108,0.3);
}
.feed-loading {
  text-align: center;
  color: var(--muted);
  font-size: 14px;
  padding: 12px 0;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
}
.feed-nomore {
  text-align: center;
  color: var(--muted);
  font-size: 13px;
  padding: 10px 0 20px 0;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
}

/* 消息通知区域 */
.message-notification-area {
  width: 100%;
  margin: 0;
  position: sticky;
  z-index: 9;
  padding: 16px 20px;
  margin-top: -1px;
  transition: all 0.3s ease;
  background: var(--bg);
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  border: 1px solid var(--border);
  border-top: none;
  box-sizing: border-box;
}

.message-notification-area.sticky-active {
  background: var(--bg);
  border-color: var(--border);
  border-top: none;
}

.message-notification-area.sticky-active .message-badge {
  border-color: var(--danger);
}

.message-badge {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  width: 100%;
  box-sizing: border-box;
  border-left: 4px solid var(--danger);
  margin: 0;
}

.message-badge:hover {
  background: var(--card-hover);
  transform: translateY(-1px);
  border-color: var(--danger);
  border-left-color: var(--danger);
}

.message-badge:active {
  transform: translateY(0);
  transition: all 0.1s ease;
}

.message-badge:focus-visible {
  outline: 2px solid var(--danger);
  outline-offset: 2px;
}

.badge-content {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  gap: 12px;
  position: relative;
  z-index: 1;
}

.badge-icon {
  width: 32px;
  height: 32px;
  background: var(--danger);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--fg);
  flex-shrink: 0;
  position: relative;
  transition: all 0.3s ease;
}

.message-badge:hover .badge-icon {
  transform: scale(1.05);
}

.badge-icon .icon {
  font-size: 16px;
  transition: transform 0.3s ease;
}

.message-badge:hover .badge-icon .icon {
  transform: scale(1.1);
}

.badge-text {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  letter-spacing: 0.1px;
  line-height: 1.4;
}

.badge-arrow {
  color: var(--muted);
  font-size: 14px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 0.7;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background: var(--accent-muted);
}

.message-badge:hover .badge-arrow {
  color: var(--danger);
  transform: translateX(2px);
  opacity: 1;
}

.badge-indicator {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
}

.red-dot {
  min-width: 20px;
  height: 20px;
  background: var(--danger);
  border-radius: var(--radius-full);
  border: 1.5px solid var(--card);
  animation: pulse-red 2s infinite;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--fg);
  font-size: 10px;
  font-weight: 700;
  padding: 0 4px;
  letter-spacing: -0.3px;
  transition: all 0.3s ease;
}

.message-badge:hover .red-dot {
  transform: scale(1.1);
}

@keyframes pulse-red {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.9;
    transform: scale(1.05);
  }
}

/* 用户个人信息头部容器样式调整 */
.feed-type-list-center :deep(.user-profile-header) {
  position: sticky;
  top: 0;
  z-index: 10;
  margin: 0;
  transition: all 0.3s ease;
  overflow: hidden;
}

/* 有消息气泡时的样式 */
.feed-type-list-center.has-notifications :deep(.user-profile-header) {
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  border-bottom: none;
}

.feed-type-list-center.has-notifications :deep(.user-profile-header .quick-nav) {
  border-radius: 0;
  border-bottom: none;
}

/* 没有消息气泡时的样式 */
.feed-type-list-center.no-notifications :deep(.user-profile-header) {
  border-radius: var(--radius-lg);
  margin-bottom: 24px;
}

.feed-type-list-center.no-notifications :deep(.user-profile-header .quick-nav) {
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
}

/* 调整用户卡片内部的快速导航样式 */
.feed-type-list-center :deep(.user-profile-header .quick-nav) {
  border-radius: 0;
  border-bottom: none;
}

/* 滚动时的置顶效果 */
.feed-type-list-center :deep(.user-profile-header.sticky) {
  top: 0;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
}

/* 动态项目间距调整 - 有消息气泡时 */
.feed-type-list-center.has-notifications .feed-type-feed-item:first-of-type {
  margin-top: 16px;
  border-radius: var(--radius-md);
  border-top: 1px solid var(--border);
}

/* 动态项目间距调整 - 没有消息气泡时 */
.feed-type-list-center.no-notifications .feed-type-feed-item:first-of-type {
  margin-top: 0;
  border-radius: var(--radius-md);
  border-top: 1px solid var(--border);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .feed-type-list-center {
    max-width: 840px;
    padding: 0 30px;
  }
}

@media (max-width: 768px) {
  .feed-type-list-center {
    max-width: 100%;
    padding: 0 20px;
  }

  .feed-type-list-main {
    padding: 0;
  }

  .feed-type-list-center :deep(.user-profile-header) {
    margin: 0;
  }

  .feed-type-list-center.has-notifications .feed-type-feed-item:first-of-type {
    margin-top: 16px;
  }

  .feed-type-list-center.no-notifications .feed-type-feed-item:first-of-type {
    margin-top: 0;
  }

  .message-notification-area {
    margin: 0;
    padding: 16px 16px;
    margin-top: -2px;
  }

  .message-badge {
    margin: 0;
  }

  .badge-content {
    padding: 10px 12px;
    gap: 10px;
  }

  .badge-icon {
    width: 28px;
    height: 28px;
  }

  .badge-icon .icon {
    font-size: 14px;
  }

  .badge-text {
    font-size: 13px;
  }

  .badge-arrow {
    font-size: 12px;
    width: 20px;
    height: 20px;
  }

  .red-dot {
    min-width: 16px;
    height: 16px;
    font-size: 9px;
    padding: 0 3px;
  }

  .badge-indicator {
    top: 6px;
    right: 6px;
  }
}

/* 消息覆盖层样式 */
.message-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--overlay);
  backdrop-filter: blur(8px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.message-panel {
  background: var(--card);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 700px;
  max-height: 85vh;
  border: 1px solid var(--border);
  overflow: hidden;
  position: relative;
}

.message-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 28px;
  background: var(--input-bg);
  border-bottom: 1px solid var(--border);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: var(--accent);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--bg);
  font-size: 20px;
}

.header-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.header-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: var(--accent);
  font-family: var(--font-display);
}

.header-subtitle {
  font-size: 13px;
  color: var(--muted);
  margin: 0;
  font-weight: 500;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: var(--font-ui);
}

.clear-button {
  background: var(--danger);
  color: var(--fg);
  font-weight: 600;
  font-size: 13px;
  padding: 8px 16px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.clear-button:hover {
  opacity: 0.85;
  transform: translateY(-1px);
}

.clear-button:active {
  transform: translateY(0);
}

.close-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: var(--radius-md);
  background: var(--accent-muted);
  color: var(--muted);
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid var(--border);
  min-width: 70px;
  font-family: var(--font-ui);
}

.close-button:hover {
  background: var(--card-hover);
  color: var(--fg);
  transform: translateY(-1px);
  border-color: var(--accent);
}

.close-button:active {
  transform: translateY(0);
}

.close-button .icon {
  font-size: 16px;
  font-weight: bold;
  color: inherit;
  transition: all 0.3s ease;
}

.close-button:hover .icon {
  transform: rotate(90deg);
}

.close-text {
  font-size: 13px;
  font-weight: 600;
  color: inherit;
  line-height: 1;
}

/* 移动端隐藏关闭按钮文字 */
@media (max-width: 768px) {
  .close-button {
    min-width: 36px;
    padding: 8px;
  }

  .close-text {
    display: none;
  }

  .clear-button {
    font-size: 12px;
    padding: 8px 12px;
  }
}

.message-list {
  overflow-y: auto;
  max-height: calc(85vh - 120px);
  min-height: calc(85vh - 120px);
  padding: 0;
  scrollbar-width: thin;
  scrollbar-color: var(--accent-muted) transparent;
}

.message-list::-webkit-scrollbar {
  width: 6px;
}

.message-list::-webkit-scrollbar-track {
  background: transparent;
}

.message-list::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 3px;
}

.message-list::-webkit-scrollbar-thumb:hover {
  background: var(--accent);
}

.notification-item {
  display: flex;
  align-items: center;
  padding: 18px 28px;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  animation: slideInUp 0.6s ease forwards;
  opacity: 0;
  transform: translateY(20px);
}

.notification-item:hover {
  background: var(--card-hover);
  transform: translateX(8px);
}

.notification-item.unread {
  background: var(--accent-muted);
  border-left: 4px solid var(--accent);
}

.notification-item:last-child {
  border-bottom: none;
}

.notification-avatar {
  position: relative;
  margin-right: 16px;
}

.notification-avatar img {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  border: 2px solid var(--border);
  cursor: pointer;
  transition: all 0.3s ease;
}

.notification-avatar img:hover {
  transform: scale(1.1);
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(200,149,108,0.2);
}

.notification-type-badge {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 20px;
  height: 20px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--card);
  font-size: 10px;
  color: var(--fg);
}

.notification-type-badge.like {
  background: var(--danger);
}

.notification-type-badge.comment {
  background: var(--accent);
}

.notification-type-badge.reply {
  background: var(--success);
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-text {
  font-size: 15px;
  line-height: 1.5;
  color: var(--fg);
  margin-bottom: 6px;
}

.notification-user {
  font-weight: 700;
  color: var(--fg);
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 2px 4px;
  border-radius: 4px;
  margin: -2px -4px;
}

.notification-user:hover {
  color: var(--accent);
  background: var(--accent-muted);
  transform: scale(1.05);
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
  font-size: 13px;
  color: var(--muted);
  line-height: 1.4;
  margin-bottom: 6px;
  padding: 6px 10px;
  background: var(--input-bg);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--border);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.notification-time {
  font-size: 12px;
  color: var(--muted);
  font-weight: 500;
}

.notification-indicator {
  margin-left: 12px;
}

.unread-dot {
  width: 8px;
  height: 8px;
  background: var(--accent);
  border-radius: var(--radius-full);
  animation: pulse 2s infinite;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  min-height: calc(85vh - 180px);
  height: 100%;
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
  font-weight: 600;
  color: var(--fg);
  margin-bottom: 8px;
  font-family: var(--font-display);
}

.empty-desc {
  font-size: 14px;
  color: var(--muted);
}

/* 动画效果 */
.message-overlay-enter-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.message-overlay-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.message-overlay-enter-from {
  opacity: 0;
  backdrop-filter: blur(0px);
}

.message-overlay-enter-to {
  opacity: 1;
  backdrop-filter: blur(8px);
}

.message-overlay-leave-from {
  opacity: 1;
  backdrop-filter: blur(8px);
}

.message-overlay-leave-to {
  opacity: 0;
  backdrop-filter: blur(0px);
}

.message-overlay-enter-active .message-panel {
  animation: slideInScale 0.5s cubic-bezier(0.4, 0, 0.2, 1) forwards;
}

.message-overlay-leave-active .message-panel {
  animation: slideOutScale 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards;
}

@keyframes slideInScale {
  0% {
    opacity: 0;
    transform: scale(0.9) translateY(20px);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

@keyframes slideOutScale {
  0% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
  100% {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
}

@keyframes slideInUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.2);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .message-overlay {
    padding: 8px;
  }

  .message-panel {
    max-width: 100%;
    max-height: 90vh;
    border-radius: var(--radius-lg);
  }

  .message-header {
    padding: 20px 20px;
  }

  .header-icon {
    width: 40px;
    height: 40px;
    font-size: 18px;
  }

  .header-title {
    font-size: 18px;
  }

  .message-list {
    max-height: calc(90vh - 120px);
    min-height: calc(90vh - 120px);
  }

  .notification-item {
    padding: 16px 20px;
  }

  .notification-avatar img {
    width: 40px;
    height: 40px;
  }

  .notification-avatar img:hover {
    transform: scale(1.15);
  }

  .notification-type-badge {
    width: 16px;
    height: 16px;
    font-size: 9px;
  }

  .empty-state {
    min-height: calc(90vh - 180px);
    padding: 40px 20px;
  }
}
</style> 