<template>
  <div class="friend-requests-view">
    <h2>好友请求</h2>
    <div v-if="requests.length === 0" class="empty">暂无好友请求</div>
    <div v-for="req in requests" :key="req.id" class="request-item">
      <img :src="req.avatar" class="avatar" />
      <div class="info">
        <div class="name">{{ req.name }}</div>
        <div class="remark">{{ req.remark }}</div>
        <div class="message">{{ req.message }}</div>
        <div class="time">{{ req.time }}</div>
      </div>
      <div class="actions">
        <button @click="accept(req)">同意</button>
        <button @click="reject(req)">拒绝</button>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref } from 'vue'
const requests = ref([
  { id: 1, name: '小红', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Hong', remark: '产品经理', message: '你好，想加你为好友', time: '1分钟前' },
  { id: 2, name: '小军', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Jun', remark: '后端开发', message: '一起学习Go', time: '5分钟前' }
])
const accept = req => { requests.value = requests.value.filter(r => r.id !== req.id) }
const reject = req => { requests.value = requests.value.filter(r => r.id !== req.id) }
</script>
<style scoped>
.friend-requests-view { max-width: 520px; margin: 40px auto; }
h2 { margin-bottom: 24px; }
.empty { color: #aaa; text-align: center; margin: 40px 0; }
.request-item { display: flex; align-items: center; gap: 16px; background: #f8fafc; border-radius: 12px; padding: 16px; margin-bottom: 16px; }
.avatar { width: 48px; height: 48px; border-radius: 10px; }
.info { flex: 1; }
.name { font-size: 16px; font-weight: 600; }
.remark, .message, .time { font-size: 13px; color: #666; margin-top: 2px; }
.actions { display: flex; flex-direction: column; gap: 8px; }
.actions button { padding: 6px 16px; border-radius: 8px; border: none; background: #4a8cff; color: #fff; }
.actions button:last-child { background: #e53e3e; }
</style> 