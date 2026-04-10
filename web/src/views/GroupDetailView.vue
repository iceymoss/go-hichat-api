<template>
  <div class="group-detail-view">
    <div class="group-header">
      <img :src="group.avatar" class="avatar" />
      <div class="info">
        <div class="name">{{ group.name }}</div>
        <div class="id">群号：{{ group.id }}</div>
        <div class="desc">{{ group.desc }}</div>
        <div class="notice">公告：{{ group.notice }}</div>
      </div>
    </div>
    <div class="members-section">
      <h3>群成员 ({{ group.members.length }})</h3>
      <div class="members-list">
        <div v-for="m in group.members" :key="m.id" class="member-item">
          <img :src="m.avatar" class="member-avatar" />
          <div class="member-name">{{ m.name }}</div>
          <span v-if="m.role==='owner'" class="role owner">群主</span>
          <span v-else-if="m.role==='admin'" class="role admin">管理员</span>
        </div>
      </div>
    </div>
    <div class="group-actions">
      <button>邀请成员</button>
      <button>退群</button>
      <button>群设置</button>
    </div>
  </div>
</template>
<script setup>
import { ref } from 'vue'
const group = ref({
  id: 10001,
  name: '前端交流群',
  avatar: 'https://api.dicebear.com/7.x/icons/svg?seed=fe',
  desc: '前端技术交流与分享',
  notice: '欢迎新成员！',
  members: [
    { id: 1, name: '小明', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ming', role: 'owner' },
    { id: 2, name: '小芳', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Fang', role: 'admin' },
    { id: 3, name: '小军', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Jun', role: 'member' }
  ]
})
</script>
<style scoped>
.group-detail-view {
  max-width: 700px;
  margin: 40px auto;
  padding: 0 20px;
  color: var(--fg);
  font-family: var(--font-ui);
}
.group-header {
  display: flex;
  align-items: center;
  gap: 24px;
  background: var(--card);
  border-radius: var(--radius-lg);
  padding: 24px;
  margin-bottom: 32px;
  border: 1px solid var(--border);
}
.avatar {
  width: 72px;
  height: 72px;
  border-radius: var(--radius-lg);
  border: 2px solid var(--border);
}
.info {
  flex: 1;
}
.name {
  font-size: 22px;
  font-family: var(--font-display);
  font-weight: 700;
  color: var(--fg);
}
.id, .desc, .notice {
  font-size: 14px;
  color: var(--muted);
  margin-top: 4px;
}
.members-section {
  margin-bottom: 32px;
}
.members-section h3 {
  font-family: var(--font-display);
  color: var(--fg);
  margin-bottom: 16px;
}
.members-list {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}
.member-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.member-avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}
.member-name {
  font-size: 14px;
  margin-top: 4px;
  color: var(--fg);
}
.role {
  font-size: 12px;
  margin-top: 2px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-weight: 500;
}
.owner {
  background: var(--accent);
  color: var(--bg);
}
.admin {
  background: var(--success);
  color: #fff;
}
.group-actions {
  display: flex;
  gap: 16px;
  justify-content: flex-end;
}
.group-actions button {
  padding: 8px 20px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--fg);
  font-family: var(--font-ui);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}
.group-actions button:hover {
  background: var(--accent-muted);
  border-color: var(--accent);
  color: var(--accent);
}
</style> 