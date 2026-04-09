<template>
  <div class="group-list-container">
    <!-- 顶部操作区 -->
    <div class="action-section">
      <button class="list-entry-btn" @click="emit('select-new-groups')">
        <div class="icon-wrapper orange">
          <i class="icon icon-users"></i>
        </div>
        <span class="text">新的群聊</span>
      </button>
    </div>
    
    <div class="search-section">
      <input v-model="keyword" placeholder="搜索群聊/群号" />
      <button class="add-group-btn" title="Join by token" @click="showJoinModal = true">
        <i class="icon icon-link"></i>
      </button>
      <button class="add-group-btn" @click="showCreateModal = true">
        <i class="icon icon-plus"></i>
      </button>
    </div>
    <div v-if="filteredGroups.length === 0" class="empty">暂无群聊</div>
    <div v-for="group in filteredGroups" :key="group.id" class="group-item" :class="{active: selectedGroupId === group.id}" @click="selectGroup(group)">
      <img :src="group.avatar" class="avatar" />
      <div class="info">
        <div class="name">{{ group.name }}</div>
        <div class="desc">{{ group.desc }}</div>
      </div>
      <div class="unread" v-if="group.unread">{{ group.unread }}</div>
    </div>

    <CreateGroupModal 
      v-if="showCreateModal" 
      @close="showCreateModal = false" 
      @success="handleCreateSuccess" 
    />

    <JoinGroupByTokenModal
      v-if="showJoinModal"
      @close="showJoinModal = false"
      @success="handleJoinSuccess"
    />
    <Toast ref="toastRef" />
  </div>
</template>
<script setup>
import { ref, computed } from 'vue'
import { useContactsStore } from '../stores/contacts'
import CreateGroupModal from './CreateGroupModal.vue'
import JoinGroupByTokenModal from './JoinGroupByTokenModal.vue'
import Toast from './ui/Toast.vue'

const props = defineProps({
  selectedGroupId: {
    type: [Number, String],
    default: null
  }
})
const emit = defineEmits(['select-group', 'select-new-groups'])
const contactsStore = useContactsStore()
const keyword = ref('')
const showCreateModal = ref(false)
const showJoinModal = ref(false)
const toastRef = ref(null)

const filteredGroups = computed(() => {
  if (!keyword.value) return contactsStore.groups
  return contactsStore.groups.filter(g => 
    g.name.includes(keyword.value) || (g.desc && g.desc.includes(keyword.value))
  )
})

const selectGroup = (group) => {
  emit('select-group', group)
}

const handleCreateSuccess = () => {
  toastRef.value?.show('群聊创建成功')
}

const handleJoinSuccess = (resp) => {
  const isPass = resp?.is_pass ?? resp?.isPass
  if (isPass === 1) toastRef.value?.show('已加入群聊')
  else toastRef.value?.show('已提交入群申请')
}
</script>
<style scoped>
.group-list-container { height: 100%; display: flex; flex-direction: column; background: #fff; }

/* 操作入口 */
.action-section {
  border-bottom: 1px solid #f1f5f9;
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
  transition: background 0.15s;
}
.list-entry-btn:hover { background: #f8fafc; }

.icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon-wrapper.orange { background: #f59e0b; }
.icon-wrapper .icon { color: white; font-size: 16px; }

.text { font-size: 13px; color: #334155; font-weight: 500; }

.search-section {
  padding: 8px 12px;
  display: flex;
  gap: 6px;
}
.search-section input {
  flex: 1;
  padding: 7px 10px;
  border-radius: 8px;
  border: 1px solid transparent;
  background: #f1f5f9;
  font-size: 13px;
  color: #334155;
  outline: none;
  transition: all 0.15s;
}
.search-section input::placeholder { color: #94a3b8; }
.search-section input:focus { background: #fff; border-color: #cbd5e1; }
.add-group-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  color: #64748b;
  transition: all 0.15s;
  flex-shrink: 0;
}
.add-group-btn:hover { background: #e2e8f0; color: #2563eb; }
.empty { color: #94a3b8; text-align: center; margin: 40px 0; font-size: 13px; }
.group-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background 0.1s;
}
.group-item:hover { background: #f8fafc; }
.group-item.active { background: #eff6ff; }
.avatar { width: 36px; height: 36px; border-radius: 50%; object-fit: cover; }
.info { flex: 1; min-width: 0; }
.name { font-size: 14px; font-weight: 500; color: #1e293b; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.desc { font-size: 12px; color: #94a3b8; margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.unread { background: #ef4444; color: #fff; border-radius: 10px; padding: 1px 6px; font-size: 11px; min-width: 18px; text-align: center; }
</style> 