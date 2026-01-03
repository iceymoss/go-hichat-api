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
.group-detail-view { max-width: 700px; margin: 40px auto; }
.group-header { display: flex; align-items: center; gap: 24px; background: #f8fafc; border-radius: 16px; padding: 24px; margin-bottom: 32px; }
.avatar { width: 72px; height: 72px; border-radius: 16px; }
.info { flex: 1; }
.name { font-size: 22px; font-weight: 700; }
.id, .desc, .notice { font-size: 14px; color: #666; margin-top: 4px; }
.members-section { margin-bottom: 32px; }
.members-list { display: flex; gap: 20px; flex-wrap: wrap; }
.member-item { display: flex; flex-direction: column; align-items: center; }
.member-avatar { width: 48px; height: 48px; border-radius: 12px; }
.member-name { font-size: 14px; margin-top: 4px; }
.role { font-size: 12px; margin-top: 2px; padding: 2px 8px; border-radius: 8px; }
.owner { background: #4a8cff; color: #fff; }
.admin { background: #10b981; color: #fff; }
.group-actions { display: flex; gap: 16px; justify-content: flex-end; }
.group-actions button { padding: 8px 20px; border-radius: 8px; border: none; background: #4a8cff; color: #fff; font-weight: 600; }
</style> 