<template>
  <div class="contacts-view">
    <div class="sidebar">
      <div class="tabs">
        <button :class="{active: tab==='friends'}" @click="tab='friends'">好友</button>
        <button :class="{active: tab==='groups'}" @click="tab='groups'">群聊</button>
      </div>
      <div class="list-scroll">
        <FriendList 
          v-if="tab==='friends'" 
          :selectedFriendId="selectedFriendId" 
          @select-friend="selectFriend" 
          @add-friend="showAddFriendModal = true"
          @select-new-friends="selectNewFriends"
        />
        <GroupList v-if="tab==='groups'" :selectedGroupId="selectedGroupId" @select-group="selectGroup" />
      </div>
    </div>
    <div class="detail-panel">
      <NewFriends 
        v-if="selectedFriendId === 'new-friends'" 
        @select-friend="handleSelectFriendFromNewFriends"
      />
      <FriendDetail v-else-if="tab==='friends' && selectedFriend" :friend="selectedFriend" />
      <div v-else-if="tab==='groups' && selectedGroup" class="group-detail-center">
        <GroupDetail :group="selectedGroup" />
      </div>
    </div>
    <AddFriendModal 
      v-if="showAddFriendModal" 
      @close="showAddFriendModal = false" 
      @send-request="handleSendRequest" 
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useContactsStore } from '../stores/contacts'
import FriendList from '../components/FriendList.vue'
import FriendDetail from '../components/FriendDetail.vue'
import AddFriendModal from '../components/AddFriendModal.vue'
import GroupDetail from '../components/GroupDetail.vue'
import GroupList from '../components/GroupList.vue'
import NewFriends from '../components/NewFriends.vue'

const tab = ref('friends')
const selectedFriendId = ref(null)
const selectedGroupId = ref(null)
const showAddFriendModal = ref(false)

const contactsStore = useContactsStore()

const selectedFriend = computed(() => {
  if (selectedFriendId.value === 'new-friends') return null
  return contactsStore.friends.find(f => (f.id === selectedFriendId.value || f.friend_uid === selectedFriendId.value)) || null
})

// 组件挂载时加载数据
onMounted(async () => {
  await contactsStore.fetchFriends()
  await contactsStore.fetchGroups()
  // 同时请求两类申请的所有 type 数据
  const requests = []
  // 我收到的申请（class=1）：请求所有 type
  for (let type = 0; type <= 3; type++) {
    requests.push(contactsStore.fetchFriendRequests(type, '1'))
  }
  // 我发起的申请（class=0）：请求所有 type
  for (let type = 0; type <= 3; type++) {
    requests.push(contactsStore.fetchFriendRequests(type, '0'))
  }
  await Promise.all(requests)
})

const selectNewFriends = () => {
  selectedFriendId.value = 'new-friends'
  selectedGroupId.value = null
}

// 处理添加好友请求（用于刷新列表）
const handleSendRequest = async (userId) => {
  // AddFriendModal 已经自己处理了发送逻辑
  // 这里只需要刷新好友申请列表（可选）
  try {
    await contactsStore.fetchFriendRequests(0, '0') // 刷新我发起的申请列表
  } catch (error) {
    console.error('刷新申请列表失败:', error)
  }
}

const selectedGroup = computed(() => {
  return contactsStore.groups.find(g => g.id === selectedGroupId.value) || null
})

const selectFriend = (friend) => {
  selectedFriendId.value = friend.friend_uid || friend.id
}

// 处理从 NewFriends 组件选择好友
const handleSelectFriendFromNewFriends = (friend) => {
  selectedFriendId.value = friend.friend_uid || friend.id
}

const selectGroup = (group) => {
  selectedGroupId.value = group.id
}
</script>

<style scoped>
.contacts-view {
  display: flex;
  height: 100vh;
  background: #f8fafc;
}
.sidebar {
  width: 340px;
  min-width: 260px;
  max-width: 400px;
  background: #fff;
  border-right: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
}
.tabs {
  display: flex;
  gap: 8px;
  padding: 16px 12px 0 12px;
}
.tabs button {
  flex: 1;
  padding: 8px 0;
  border: none;
  border-radius: 8px 8px 0 0;
  background: #f1f5f9;
  color: #333;
  font-weight: 500;
  cursor: pointer;
}
.tabs button.active {
  background: #4a8cff;
  color: #fff;
}
.list-scroll {
  flex: 1;
  overflow-y: auto;
}
.detail-panel {
  flex: 1;
  background: #fff; /* 改为白色 */
  display: flex;
  flex-direction: column;
  overflow: hidden; /* 防止出现双重滚动条 */
  /* 移除 align-items: center 和 justify-content: center，让子元素自动撑满 */
}
.group-detail-center {
  width: 100%; /* 群组详情也改为撑满，内部自己控制布局 */
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>