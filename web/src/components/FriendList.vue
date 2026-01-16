<template>
  <div class="friend-list-container">
    <!-- 顶部操作区 -->
    <div class="action-section">
      <button class="list-entry-btn" @click="emit('select-new-friends')">
        <div class="icon-wrapper orange">
          <i class="icon icon-user-plus"></i>
        </div>
        <span class="text">新的朋友</span>
        <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
      </button>
      
      <button class="list-entry-btn" @click="emit('add-friend')">
        <div class="icon-wrapper green">
          <i class="icon icon-plus"></i>
        </div>
        <span class="text">添加朋友</span>
      </button>
    </div>

    <!-- 搜索框 -->
    <div class="search-section">
      <div class="search-box">
        <i class="icon icon-search"></i>
        <input 
          type="text" 
          v-model="searchKeyword" 
          placeholder="搜索" 
        >
        <i v-if="searchKeyword" class="icon icon-x clear-btn" @click="searchKeyword = ''"></i>
      </div>
    </div>

    <!-- 好友列表 -->
    <div class="scroll-area">
      <!-- 无搜索结果/无好友时 -->
      <div v-if="sortedFriends.length === 0" class="empty-list">
        {{ searchKeyword ? '无搜索结果' : '暂无好友' }}
      </div>

      <!-- 分组列表 -->
      <div v-else class="friends-group-list">
        <div 
          v-for="group in groupedFriends" 
          :key="group.letter" 
          class="group-block"
        >
          <div class="group-header">{{ group.letter }}</div>
          <div 
            v-for="friend in group.friends" 
            :key="friend.id" 
            class="friend-row"
            :class="{ active: selectedFriendId === friend.id || selectedFriendId === friend.friend_uid }"
            @click="selectFriend(friend)"
          >
            <img :src="friend.avatar" alt="头像" class="friend-avatar" @error="handleAvatarError">
            <div class="friend-info">
              <span class="friend-name" :class="{ 'is-remark': friend.remark && friend.remark.trim() }">
                {{ (friend.remark && friend.remark.trim()) ? friend.remark : (friend.nickname || '未知用户') }}
              </span>
            </div>
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

const emit = defineEmits(['select-friend', 'add-friend', 'select-new-friends'])

const contactsStore = useContactsStore()
const searchKeyword = ref('')

// 直接从 Store 计算未读数，移除冗余的轮询
const unreadCount = computed(() => {
  return contactsStore.friendRequests.filter(req => req.read_state === 0).length
})

const handleAvatarError = (event) => {
  event.target.src = `https://api.dicebear.com/7.x/personas/svg?seed=default`
}

const sortedFriends = computed(() => {
  let friends = contactsStore.friends
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    friends = friends.filter(friend => 
      (friend.nickname && friend.nickname.toLowerCase().includes(keyword)) || 
      (friend.remark && friend.remark.toLowerCase().includes(keyword))
    )
  }
  return sortFriendsByPinyin(friends)
})

const groupedFriends = computed(() => {
  return groupFriendsByPinyin(sortedFriends.value)
})

const selectFriend = (friend) => {
  emit('select-friend', friend)
}
</script>

<style scoped>
.friend-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f7f7f7;
  border-right: 1px solid #e5e5e5;
  user-select: none;
}

/* 操作入口 */
.action-section {
  background: white;
}
.list-entry-btn {
  width: 100%;
  padding: 12px 16px;
  background: transparent;
  border: none;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: background 0.2s;
}
.list-entry-btn:hover { background: #f3f3f3; }

.icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-wrapper.orange { background: #fa9d3b; }
.icon-wrapper.green { background: #07c160; }
.icon-wrapper .icon { color: white; font-size: 20px; }

.text { font-size: 14px; color: #111827; font-weight: 500; }

.badge {
  margin-left: auto;
  background: #ef4444;
  color: white;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

/* 搜索框 */
.search-section {
  padding: 10px;
  background: #f7f7f7;
}
.search-box {
  background: #e2e2e2;
  border-radius: 6px;
  padding: 6px 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}
.search-box:focus-within { background: white; border: 1px solid #d1d5db; }
.search-box input {
  border: none;
  background: transparent;
  width: 100%;
  font-size: 13px;
  outline: none;
}
.search-box .icon { color: #9ca3af; font-size: 14px; }
.clear-btn { cursor: pointer; }

/* 列表区域 */
.scroll-area {
  flex: 1;
  overflow-y: auto;
}
.empty-list {
  text-align: center;
  color: #9ca3af;
  padding-top: 40px;
  font-size: 13px;
}

.group-header {
  padding: 4px 16px;
  font-size: 12px;
  color: #9ca3af;
  background: #f7f7f7;
}

.friend-row {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  gap: 12px;
  cursor: pointer;
}
.friend-row:hover { background: #eaeaeb; }
.friend-row.active { background: #c6c6c6; }

.friend-avatar {
  width: 36px;
  height: 36px;
  border-radius: 4px;
}

.friend-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.friend-name {
  font-size: 14px;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.friend-name.is-remark {
  /* color: #111827; */ /* 移除颜色区分，保持一致 */
  /* font-weight: 500; */ /* 移除字重区分，保持一致 */
}

.friend-nickname {
  font-size: 12px;
  color: #9ca3af;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>