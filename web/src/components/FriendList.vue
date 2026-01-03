<template>
  <div class="friend-list-container">
    <!-- 搜索框 -->
    <div class="search-section">
      <div class="search-container">
        <i class="icon icon-search"></i>
        <input 
          type="text" 
          v-model="searchKeyword" 
          placeholder="搜索好友昵称、备注或标签" 
          @input="handleSearch"
        >
        <button v-if="searchKeyword" class="btn-clear" @click="clearSearch">
          <i class="icon icon-x"></i>
        </button>
      </div>
    </div>

    <!-- 好友列表 -->
    <div class="friends-content">
      <div v-if="!searchMode" class="friends-section">
        <div class="section-header">
          <h3>好友请求</h3>
          <span v-if="friendRequests.length > 0" class="badge">{{ friendRequests.length }}</span>
        </div>
        
        <div v-if="friendRequests.length > 0" class="request-list">
          <div v-for="request in friendRequests" :key="request.id" class="request-item">
            <img :src="request.avatar" alt="头像" class="avatar">
            <div class="request-info">
              <div class="name">{{ request.name }}</div>
              <div class="message">{{ request.message }}</div>
              <div class="time">{{ request.time }}</div>
            </div>
            <div class="request-actions">
              <button class="btn-accept" @click="handleRequest(request.id, true)">接受</button>
              <button class="btn-reject" @click="handleRequest(request.id, false)">拒绝</button>
            </div>
          </div>
        </div>
        <div v-else class="no-requests">
          暂无好友请求
        </div>
      </div>

      <!-- 按字母排序的好友列表 -->
      <div class="friends-section">
        <div class="section-header">
          <h3>我的好友 <span class="count">({{ sortedFriends.length }})</span></h3>
        </div>
        
        <div v-if="sortedFriends.length > 0" class="friends-list">
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
              :class="{ active: selectedFriendId === friend.id }"
              @click="selectFriend(friend)"
            >
              <div class="friend-avatar">
                <img :src="friend.avatar" alt="头像" class="avatar">
                <span class="status-badge" :class="friend.status"></span>
              </div>
              <div class="friend-info">
                <div class="name">{{ friend.remark || friend.name }}</div>
                <div class="status">
                  <span v-if="friend.status === 'online'" class="status-text online">在线</span>
                  <span v-else-if="friend.status === 'away'" class="status-text away">离开</span>
                  <span v-else class="status-text offline">{{ friend.lastActive }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div v-else class="no-friends">
          <div class="empty-state">
            <i class="icon icon-users"></i>
            <p>暂无好友</p>
            <button class="btn-add-friend" @click="$emit('add-friend')">
              <i class="icon icon-plus"></i>
              添加好友
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { sortFriendsByPinyin, groupFriendsByPinyin } from '../utils/sortUtils'

const props = defineProps({
  selectedFriendId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['select-friend', 'add-friend'])

const contactsStore = useContactsStore()
const searchKeyword = ref('')

// 计算属性
const searchMode = computed(() => searchKeyword.value.trim() !== '')
const friendRequests = computed(() => contactsStore.friendRequests)

// 按字母排序的好友列表
const sortedFriends = computed(() => {
  let friends = contactsStore.friends
  
  // 如果有搜索关键词，过滤好友
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    friends = friends.filter(friend => 
      friend.name.toLowerCase().includes(keyword) || 
      friend.remark.toLowerCase().includes(keyword) ||
      friend.tags.some(tag => tag.toLowerCase().includes(keyword))
    )
  }
  
  // 按备注或名称拼音排序
  return sortFriendsByPinyin(friends)
})

// 按字母分组的好友
const groupedFriends = computed(() => {
  return groupFriendsByPinyin(sortedFriends.value)
})

// 方法
const handleSearch = () => {
  // 搜索逻辑已在computed中处理
}

const clearSearch = () => {
  searchKeyword.value = ''
}

const selectFriend = (friend) => {
  emit('select-friend', friend)
}

const handleRequest = (requestId, accept) => {
  contactsStore.handleFriendRequest(requestId, accept)
}
</script>

<style scoped>
.friend-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.search-section {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f2f5;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
  background-color: #f8f9fa;
  border-radius: 20px;
  padding: 8px 15px;
  border: 1px solid #e9ecef;
}

.search-container .icon {
  color: #6c757d;
  font-size: 16px;
  margin-right: 8px;
}

.search-container input {
  flex: 1;
  border: none;
  background: none;
  font-size: 14px;
  outline: none;
  color: #495057;
}

.search-container input::placeholder {
  color: #adb5bd;
}

.btn-clear {
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  color: #6c757d;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-clear:hover {
  background-color: #e9ecef;
}

.friends-content {
  flex: 1;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: #4a8cff #e9ecef;
}

.friends-content::-webkit-scrollbar {
  width: 6px;
}

.friends-content::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  border-radius: 3px;
}

.friends-content::-webkit-scrollbar-track {
  background: #f8f9fa;
}

.friends-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  padding: 16px 20px 12px 20px;
  background-color: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #495057;
}

.badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background-color: #dc3545;
  color: white;
  font-size: 12px;
  margin-left: 8px;
  font-weight: 600;
}

.count {
  color: #6c757d;
  font-weight: normal;
  font-size: 14px;
}

.request-list {
  padding: 8px 0;
}

.request-item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #f8f9fa;
}

.request-item:last-child {
  border-bottom: none;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
  margin-right: 12px;
}

.request-info {
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 15px;
  font-weight: 500;
  color: #495057;
  margin-bottom: 2px;
}

.message {
  font-size: 13px;
  color: #6c757d;
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  font-size: 12px;
  color: #adb5bd;
}

.request-actions {
  display: flex;
  gap: 8px;
}

.btn-accept, .btn-reject {
  padding: 6px 12px;
  font-size: 12px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-accept {
  background-color: #28a745;
  color: white;
}

.btn-accept:hover {
  background-color: #218838;
}

.btn-reject {
  background-color: #f8f9fa;
  color: #6c757d;
  border: 1px solid #dee2e6;
}

.btn-reject:hover {
  background-color: #e9ecef;
}

.no-requests {
  padding: 20px;
  text-align: center;
  color: #6c757d;
  font-size: 14px;
}

.friends-list {
  padding: 0;
}

.friend-group {
  margin-bottom: 8px;
}

.group-header {
  padding: 8px 20px;
  background-color: #f8f9fa;
  font-size: 14px;
  font-weight: 600;
  color: #495057;
  border-bottom: 1px solid #e9ecef;
}

.friend-item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  cursor: pointer;
  transition: all 0.2s;
  border-bottom: 1px solid #f8f9fa;
}

.friend-item:hover {
  background-color: #f8f9fa;
}

.friend-item.active {
  background-color: #e3f2fd;
  border-left: 3px solid #4a8cff;
}

.friend-item:last-child {
  border-bottom: none;
}

.friend-avatar {
  position: relative;
  margin-right: 12px;
}

.status-badge {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid white;
}

.status-badge.online {
  background-color: #28a745;
}

.status-badge.away {
  background-color: #ffc107;
}

.status-badge.offline {
  background-color: #6c757d;
}

.friend-info {
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 15px;
  font-weight: 500;
  color: #495057;
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status {
  display: flex;
  align-items: center;
}

.status-text {
  font-size: 12px;
  font-weight: 500;
}

.status-text.online {
  color: #28a745;
}

.status-text.away {
  color: #ffc107;
}

.status-text.offline {
  color: #6c757d;
}

.no-friends {
  padding: 40px 20px;
  text-align: center;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.empty-state .icon {
  font-size: 48px;
  color: #adb5bd;
}

.empty-state p {
  color: #6c757d;
  font-size: 16px;
  margin: 0;
}

.btn-add-friend {
  padding: 10px 20px;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  color: white;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}

.btn-add-friend:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.3);
}
</style> 