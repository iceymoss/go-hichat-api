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
        <GroupList 
          v-if="tab==='groups'" 
          :selectedGroupId="selectedGroupId" 
          @select-group="selectGroup"
          @select-new-groups="selectNewGroups"
        />
      </div>
    </div>
    <div class="detail-panel">
      <NewFriends 
        v-if="selectedFriendId === 'new-friends'" 
        @select-friend="handleSelectFriendFromNewFriends"
      />
      <NewGroups
        v-else-if="selectedGroupId === 'new-groups'"
        @select-group="handleSelectGroupFromNewGroups"
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
import NewGroups from '../components/NewGroups.vue'

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
  // 不再一次性请求所有数据，改为按需请求（由NewFriends组件控制）
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
  if (selectedGroupId.value === 'new-groups') return null
  return contactsStore.groups.find(g => g.id === selectedGroupId.value) || null
})

const selectFriend = (friend) => {
  selectedFriendId.value = friend.friend_uid || friend.id
}

// 处理从 NewFriends 组件选择好友
const handleSelectFriendFromNewFriends = (friend) => {
  selectedFriendId.value = friend.friend_uid || friend.id
}

const selectNewGroups = () => {
  selectedGroupId.value = 'new-groups'
  selectedFriendId.value = null
}

const selectGroup = (group) => {
  selectedGroupId.value = group.id
}

// 处理从 NewGroups 组件选择群组
const handleSelectGroupFromNewGroups = (group) => {
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
  width: 300px;
  min-width: 260px;
  max-width: 360px;
  background: #fff;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
}
.tabs {
  display: flex;
  padding: 0 16px;
  border-bottom: 1px solid #e2e8f0;
}
.tabs button {
  flex: 1;
  padding: 12px 0;
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: #64748b;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}
.tabs button:hover {
  color: #334155;
}
.tabs button.active {
  color: #2563eb;
  border-bottom-color: #2563eb;
}
.list-scroll {
  flex: 1;
  overflow-y: auto;
}
.detail-panel {
  flex: 1;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.group-detail-center {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>