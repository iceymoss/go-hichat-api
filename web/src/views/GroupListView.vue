<template>
  <div class="group-list-view">
    <div class="search-bar">
      <input v-model="keyword" placeholder="搜索群聊/群号" />
    </div>
    <div v-if="groups.length === 0" class="empty">暂无群聊</div>
    <div v-for="group in filteredGroups" :key="group.id" class="group-item">
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
const keyword = ref('')
const groups = ref([
  { id: 1, name: '前端交流群', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=fe', desc: '前端技术交流', unread: 2 },
  { id: 2, name: 'Vue3实战', avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=vue', desc: 'Vue3项目实战讨论', unread: 0 }
])
const filteredGroups = computed(() => groups.value.filter(g => g.name.includes(keyword.value) || g.desc.includes(keyword.value)))
</script>
<style scoped>
.group-list-view {
  max-width: 600px;
  margin: 40px auto;
  padding: 0 20px;
  font-family: var(--font-ui);
}
.search-bar {
  margin-bottom: 24px;
}
.search-bar input {
  width: 100%;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--input-bg);
  color: var(--fg);
  font-family: var(--font-ui);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s ease;
}
.search-bar input::placeholder {
  color: var(--muted);
}
.search-bar input:focus {
  border-color: var(--accent);
}
.empty {
  color: var(--muted);
  text-align: center;
  margin: 40px 0;
}
.group-item {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--card);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all 0.2s ease;
}
.group-item:hover {
  background: var(--card-hover);
  border-color: var(--accent);
}
.avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}
.info {
  flex: 1;
}
.name {
  font-size: 16px;
  font-weight: 600;
  color: var(--fg);
}
.desc {
  font-size: 13px;
  color: var(--muted);
  margin-top: 2px;
}
.unread {
  background: var(--danger);
  color: #fff;
  border-radius: var(--radius-full);
  padding: 2px 10px;
  font-size: 13px;
  font-weight: 500;
}
</style> 