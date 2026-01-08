<template>
  <div class="group-list-container">
    <div class="search-section">
      <input v-model="keyword" placeholder="搜索群聊/群号" />
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
    <Toast ref="toastRef" />
  </div>
</template>
<script setup>
import { ref, computed } from 'vue'
import { useContactsStore } from '../stores/contacts'
import CreateGroupModal from './CreateGroupModal.vue'
import Toast from './ui/Toast.vue'

const props = defineProps({
  selectedGroupId: {
    type: [Number, String],
    default: null
  }
})
const emit = defineEmits(['select-group'])
const contactsStore = useContactsStore()
const keyword = ref('')
const showCreateModal = ref(false)
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
</script>
<style scoped>
.group-list-container { height: 100%; display: flex; flex-direction: column; background: #fff; }
.search-section { padding: 16px 20px; border-bottom: 1px solid #f0f2f5; display: flex; gap: 10px; }
.search-section input { flex: 1; padding: 8px; border-radius: 6px; border: 1px solid #ddd; }
.add-group-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  color: #64748b;
  transition: all 0.2s;
}
.add-group-btn:hover { background: #e2e8f0; color: #4a8cff; }
.empty { color: #aaa; text-align: center; margin: 40px 0; }
.group-item { display: flex; align-items: center; gap: 16px; background: #f8fafc; border-radius: 12px; padding: 12px 16px; margin: 12px 12px 0 12px; cursor: pointer; transition: background 0.2s; }
.group-item.active { background: #e0e7ff; }
.avatar { width: 44px; height: 44px; border-radius: 10px; }
.info { flex: 1; }
.name { font-size: 16px; font-weight: 600; }
.desc { font-size: 13px; color: #666; margin-top: 2px; }
.unread { background: #e53e3e; color: #fff; border-radius: 10px; padding: 2px 10px; font-size: 13px; }
</style> 