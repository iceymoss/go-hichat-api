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
  // 只统计待处理（type=0）且未读的申请
  const pendingRequests = contactsStore.friendRequests['0'] || []
  return pendingRequests.filter(req => req.read_state === 0).length
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
  background: var(--card);
  user-select: none;
}

/* 操作入口 */
.action-section {
  border-bottom: 1px solid var(--border);
}
.list-entry-btn {
  width: 100%;
  padding: 10px 16px;
  background: transparent;
  border: none;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: background 0.2s;
}
.list-entry-btn:hover { background: var(--card-hover); }

.icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-wrapper.orange { background: var(--accent); }
.icon-wrapper.green { background: #a07850; }
.icon-wrapper .icon { color: #fff; font-size: 16px; }

.text { font-size: 13px; color: var(--fg); font-family: var(--font-ui); font-weight: 500; }

.badge {
  margin-left: auto;
  background: var(--accent);
  color: #fff;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: var(--radius-full);
  min-width: 18px;
  text-align: center;
  font-weight: 600;
}

/* 搜索框 */
.search-section {
  padding: 8px 12px;
}
.search-box {
  background: var(--input-bg);
  border-radius: var(--radius-md);
  padding: 7px 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
  border: 1px solid transparent;
}
.search-box:focus-within { background: var(--bg); border-color: var(--border); }
.search-box input {
  border: none;
  background: transparent;
  width: 100%;
  font-size: 13px;
  font-family: var(--font-ui);
  outline: none;
  color: var(--fg);
}
.search-box input::placeholder { color: var(--muted); }
.search-box .icon { color: var(--muted); font-size: 14px; }
.clear-btn { cursor: pointer; color: var(--muted); }
.clear-btn:hover { color: var(--fg); }

/* 列表区域 */
.scroll-area {
  flex: 1;
  overflow-y: auto;
}
.empty-list {
  text-align: center;
  color: var(--muted);
  padding-top: 40px;
  font-size: 13px;
}

.group-header {
  padding: 6px 16px;
  font-size: 11px;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: var(--bg);
}

.friend-row {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  gap: 12px;
  cursor: pointer;
  transition: background 0.15s;
}
.friend-row:hover { background: var(--card-hover); }
.friend-row.active { background: var(--accent-muted); }

.friend-avatar {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  object-fit: cover;
}

.friend-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.friend-name {
  font-size: 14px;
  color: var(--fg);
  font-family: var(--font-ui);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.friend-nickname {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>