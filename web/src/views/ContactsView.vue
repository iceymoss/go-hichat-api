<template>
  <div class="contacts-view">
    <div class="sidebar">
      <div class="tabs">
        <button :class="{active: tab==='friends'}" @click="tab='friends'">好友</button>
        <button :class="{active: tab==='groups'}" @click="tab='groups'">群聊</button>
      </div>
      <div class="list-scroll">
        <FriendList v-if="tab==='friends'" :selectedFriendId="selectedFriendId" @select-friend="selectFriend" @add-friend="showAddFriendModal = true" />
        <GroupList v-if="tab==='groups'" :selectedGroupId="selectedGroupId" @select-group="selectGroup" />
      </div>
    </div>
    <div class="detail-panel">
      <FriendDetail v-if="tab==='friends'" :friend="selectedFriend" />
      <div v-if="tab==='groups' && selectedGroup" class="group-detail-center">
        <GroupDetail :group="selectedGroup" />
      </div>
    </div>
    <AddFriendModal v-if="showAddFriendModal" @close="showAddFriendModal = false" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useContactsStore } from '../stores/contacts'
import FriendList from '../components/FriendList.vue'
import FriendDetail from '../components/FriendDetail.vue'
import AddFriendModal from '../components/AddFriendModal.vue'
import GroupDetail from '../components/GroupDetail.vue'
import GroupList from '../components/GroupList.vue'

const tab = ref('friends')
const selectedFriendId = ref(null)
const selectedGroupId = ref(null)
const showAddFriendModal = ref(false)

const contactsStore = useContactsStore()

const selectedFriend = computed(() => {
  return contactsStore.friends.find(f => f.id === selectedFriendId.value) || null
})

// mock更多群聊
const groups = ref([
  { id: 1, name: '前端交流群', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=fe', desc: '前端技术交流', unread: 2, notice: '欢迎新成员！', members: [
    { id: 1, name: '小明', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ming', role: 'owner' },
    { id: 2, name: '小芳', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Fang', role: 'admin' },
    { id: 3, name: '小军', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Jun', role: 'member' }
  ] },
  { id: 2, name: 'Vue3实战', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=vue', desc: 'Vue3项目实战讨论', unread: 0, notice: '群公告：欢迎讨论Vue3', members: [
    { id: 4, name: '小李', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Li', role: 'owner' },
    { id: 5, name: '小王', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Wang', role: 'member' }
  ] },
  { id: 3, name: 'AI极客群', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=ai', desc: 'AI极客技术交流', unread: 5, notice: 'AI极客欢迎你', members: [
    { id: 6, name: '极客A', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=GeekA', role: 'owner' },
    { id: 7, name: '极客B', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=GeekB', role: 'admin' },
    { id: 8, name: '极客C', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=GeekC', role: 'member' }
  ] },
  { id: 4, name: '产品经理群', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=pm', desc: '产品经理交流', unread: 0, notice: '产品经理群公告', members: [
    { id: 9, name: 'PM小张', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=PMZhang', role: 'owner' },
    { id: 10, name: 'PM小李', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=PMLi', role: 'member' }
  ] },
  { id: 5, name: '设计师群', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=design', desc: '设计师灵感交流', unread: 1, notice: '设计师群公告', members: [
    { id: 11, name: '设计师A', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=DesignA', role: 'owner' },
    { id: 12, name: '设计师B', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=DesignB', role: 'member' }
  ] },
  { id: 6, name: 'Go开发者', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=go', desc: 'Go开发技术交流', unread: 0, notice: 'Go开发者群公告', members: [
    { id: 13, name: 'Go小明', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=GoMing', role: 'owner' },
    { id: 14, name: 'Go小军', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=GoJun', role: 'member' }
  ] }
])

const selectedGroup = computed(() => {
  return groups.value.find(g => g.id === selectedGroupId.value) || null
})

const selectFriend = (friend) => {
  selectedFriendId.value = friend.id
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
  background: #f8fafc;
  display: flex;
  flex-direction: column;
  overflow: auto;
  align-items: center;
  justify-content: center;
}
.group-detail-center {
  width: 30%;
  min-width: 340px;
  max-width: 480px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
</style>