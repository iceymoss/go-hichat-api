<template>
  <div class="group-list-container">
    <div class="search-section">
      <input v-model="keyword" placeholder="搜索群聊/群号" />
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
  </div>
</template>
<script setup>
import { ref, computed } from 'vue'
const props = defineProps({
  selectedGroupId: {
    type: [Number, String],
    default: null
  }
})
const emit = defineEmits(['select-group'])
const keyword = ref('')
const groups = ref([
  { id: 1, name: '前端交流群', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=fe', desc: '前端技术交流', unread: 2 },
  { id: 2, name: 'Vue3实战', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=vue', desc: 'Vue3项目实战讨论', unread: 0 }
])
const filteredGroups = computed(() => groups.value.filter(g => g.name.includes(keyword.value) || g.desc.includes(keyword.value)))
const selectGroup = (group) => {
  emit('select-group', group)
}
</script>
<style scoped>
.group-list-container { height: 100%; display: flex; flex-direction: column; background: #fff; }
.search-section { padding: 16px 20px; border-bottom: 1px solid #f0f2f5; }
.search-section input { width: 100%; padding: 8px; border-radius: 6px; border: 1px solid #ddd; }
.empty { color: #aaa; text-align: center; margin: 40px 0; }
.group-item { display: flex; align-items: center; gap: 16px; background: #f8fafc; border-radius: 12px; padding: 12px 16px; margin: 12px 12px 0 12px; cursor: pointer; transition: background 0.2s; }
.group-item.active { background: #e0e7ff; }
.avatar { width: 44px; height: 44px; border-radius: 10px; }
.info { flex: 1; }
.name { font-size: 16px; font-weight: 600; }
.desc { font-size: 13px; color: #666; margin-top: 2px; }
.unread { background: #e53e3e; color: #fff; border-radius: 10px; padding: 2px 10px; font-size: 13px; }
</style> 