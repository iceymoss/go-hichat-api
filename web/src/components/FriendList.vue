<template>
  <div class="friend-list-container">
    <!-- 顶部操作区 -->
    <div class="friend-requests-section">
      <button class="btn-friend-requests" @click="showFriendRequestsModal = true">
        <div class="icon-wrapper">
          <i class="icon icon-user-plus"></i>
        </div>
        <span class="text">新的朋友</span>
        <span v-if="pendingRequestsCount > 0" class="badge-count">{{ pendingRequestsCount }}</span>
      </button>
      <button class="btn-friend-requests" @click="emit('add-friend')">
        <div class="icon-wrapper add">
          <i class="icon icon-plus"></i>
        </div>
        <span class="text">添加朋友</span>
      </button>
    </div>

    <!-- 搜索框 -->
    <div class="search-section">
      <div class="search-container">
        <i class="icon icon-search"></i>
        <input 
          type="text" 
          v-model="searchKeyword" 
          placeholder="搜索" 
          @input="handleSearch"
        >
        <button v-if="searchKeyword" class="btn-clear" @click="clearSearch">
          <i class="icon icon-x"></i>
        </button>
      </div>
    </div>

    <!-- 好友列表 -->
    <div class="friends-content">
      <div class="friends-section">
        <!-- 无搜索结果/无好友时 -->
        <div v-if="sortedFriends.length === 0" class="no-friends">
          <div class="empty-state">
            <p v-if="searchKeyword">无搜索结果</p>
            <p v-else>暂无好友</p>
          </div>
        </div>

        <!-- 好友列表 -->
        <div v-else class="friends-list">
          <div 
            v-for="group in groupedFriends" 
            :key="group.letter" 
            class="friend-group"
          >
            <div class="group-header">{{ group.letter }}</div>
            <div 
              v-for="friend in group.friends" 
              :key="friend.id" 
              class="friend-item"
              :class="{ active: selectedFriendId === friend.id || selectedFriendId === friend.friend_uid }"
              @click="selectFriend(friend)"
            >
              <div class="friend-avatar">
                <img :src="friend.avatar" alt="头像" class="avatar" :data-friend-uid="friend.friend_uid || friend.id" @error="handleAvatarError">
              </div>
              <div class="friend-info">
                <div class="name">
                  {{ (friend.remark && friend.remark.trim()) ? friend.remark : (friend.nickname || '未知用户') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 好友申请弹窗 -->
    <FriendRequestsModal 
      v-if="showFriendRequestsModal"
      @close="showFriendRequestsModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { sortFriendsByPinyin, groupFriendsByPinyin } from '../utils/sortUtils'
import { socialApi } from '../utils/api'
import FriendRequestsModal from './FriendRequestsModal.vue'

const props = defineProps({
  selectedFriendId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['select-friend', 'add-friend'])

const contactsStore = useContactsStore()
const searchKeyword = ref('')
const showFriendRequestsModal = ref(false)
const messageCount = ref(0) // 消息数量
let pollTimer = null // 轮询定时器

// 处理头像加载错误
const handleAvatarError = (event) => {
  const friendUid = event.target.getAttribute('data-friend-uid')
  event.target.src = `https://api.dicebear.com/7.x/personas/svg?seed=${friendUid || 'default'}`
}

// 计算属性
const friendRequests = computed(() => contactsStore.friendRequests)

// 待处理的好友申请消息数量
const pendingRequestsCount = computed(() => {
  return messageCount.value
})

// 获取消息数量
const fetchMessageCount = async () => {
  try {
    const response = await socialApi.friendPutInMessageCount()
    if (response && response.count !== undefined) {
      messageCount.value = response.count
    }
  } catch (error) {
    console.error('获取消息数量失败:', error)
  }
}

// 启动轮询
const startPolling = () => {
  fetchMessageCount()
  pollTimer = setInterval(() => {
    fetchMessageCount()
  }, 20000)
}

// 停止轮询
const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(() => {
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})

watch(() => showFriendRequestsModal.value, (newVal) => {
  if (!newVal) {
    fetchMessageCount()
  }
})

// 按字母排序的好友列表
const sortedFriends = computed(() => {
  let friends = contactsStore.friends
  
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    friends = friends.filter(friend => 
      (friend.nickname && friend.nickname.toLowerCase().includes(keyword)) || 
      (friend.remark && friend.remark.toLowerCase().includes(keyword)) ||
      (friend.tags && friend.tags.some(tag => tag.toLowerCase().includes(keyword)))
    )
  }
  
  return sortFriendsByPinyin(friends)
})

const groupedFriends = computed(() => {
  return groupFriendsByPinyin(sortedFriends.value)
})

const handleSearch = () => {}

const clearSearch = () => {
  searchKeyword.value = ''
}

const selectFriend = (friend) => {
  emit('select-friend', friend)
}
</script>

<style scoped>
.friend-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f7f7f7; /* 微信风格底色 */
  border-right: 1px solid #e7e7e7;
}

/* 搜索框区域 */
.search-section {
  padding: 12px 12px;
  background: #f7f7f7;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
  background-color: #e2e2e2;
  border-radius: 6px;
  padding: 4px 8px;
  height: 28px;
  transition: all 0.2s;
}

.search-container:focus-within {
  background-color: #fff;
  border: 1px solid #d1d1d1;
}

.search-container .icon {
  color: #999;
  font-size: 14px;
  margin-right: 6px;
}

.search-container input {
  flex: 1;
  border: none;
  background: none;
  font-size: 12px;
  color: #333;
  outline: none;
  padding: 0;
}

.btn-clear {
  background: none;
  border: none;
  padding: 2px;
  cursor: pointer;
  color: #999;
  display: flex;
  align-items: center;
}

/* 顶部功能入口 */
.friend-requests-section {
  padding: 0;
  background: #f7f7f7;
}

.btn-friend-requests {
  width: 100%;
  padding: 12px 18px;
  background: transparent;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
  transition: background-color 0.2s;
  border-bottom: 1px solid #e7e7e7;
}

.btn-friend-requests:hover {
  background-color: #e2e2e2;
}

.icon-wrapper {
  width: 36px;
  height: 36px;
  background-color: #fa9d3b; /* 橙色背景 */
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-wrapper.add {
  background-color: #07c160; /* 微信绿 */
}

.icon-wrapper .icon {
  color: white;
  font-size: 20px;
}

.btn-friend-requests .text {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

/* 列表内容区 */
.friends-content {
  flex: 1;
  overflow-y: auto;
}

.friends-content::-webkit-scrollbar {
  width: 6px;
}

.friends-content::-webkit-scrollbar-thumb {
  background-color: #c1c1c1;
  border-radius: 3px;
}

.friends-list {
  padding-bottom: 20px;
}

.group-header {
  padding: 4px 18px;
  font-size: 12px;
  color: #999;
  background-color: #f7f7f7;
  font-weight: normal;
}

.friend-item {
  display: flex;
  align-items: center;
  padding: 10px 18px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #e7e7e7;
}

.friend-item:last-child {
  border-bottom: none;
}

.friend-item:hover {
  background-color: #e2e2e2;
}

.friend-item.active {
  background-color: #c6c6c6; /* 微信选中态 */
}

.friend-avatar {
  margin-right: 12px;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 4px; /* 微信风格是小圆角 */
  object-fit: cover;
}

.friend-info {
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 14px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 空状态 */
.no-friends {
  padding: 40px 0;
  text-align: center;
  color: #999;
  font-size: 13px;
}

/* 徽标 */
.badge-count {
  margin-left: auto;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: #fa5151;
  color: white;
  border-radius: 9px;
  font-size: 11px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>