<template>
  <div class="group-requests-view">
    <div class="header">
      <button class="back" @click="goBack">返回</button>
      <h2>加群申请</h2>
      <button class="refresh" @click="refresh" :disabled="loading">{{ loading ? '刷新中' : '刷新' }}</button>
    </div>

    <div class="sub">群ID：{{ groupId || '未选择' }}</div>

    <div v-if="!groupId" class="empty">请从群聊详情进入本页面</div>
    <div v-else-if="loading && requests.length === 0" class="empty">加载中...</div>
    <div v-else-if="requests.length === 0" class="empty">暂无加群申请</div>

    <div v-for="req in requests" :key="req.group_req_id" class="request-item">
      <img :src="req.avatar" class="avatar" />
      <div class="info">
        <div class="name">{{ req.name }}</div>
        <div class="message">{{ req.message }}</div>
        <div class="time">{{ formatTime(req.req_time) }}</div>
      </div>
      <div class="actions">
        <button @click="accept(req)" :disabled="loading">同意</button>
        <button class="danger" @click="reject(req)" :disabled="loading">拒绝</button>
      </div>
    </div>

    <Toast ref="toastRef" />
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useContactsStore } from '../stores/contacts'
import Toast from '../components/ui/Toast.vue'

const route = useRoute()
const router = useRouter()
const contactsStore = useContactsStore()
const toastRef = ref(null)

const groupId = computed(() => route.query.group_id || route.query.groupId || '')
const loading = computed(() => contactsStore.requestsLoading)
const requests = computed(() => contactsStore.groupRequests)

const refresh = async () => {
  if (!groupId.value) return
  try {
    await contactsStore.fetchGroupRequests(String(groupId.value), [0], 2)
  } catch (e) {
    toastRef.value?.show('加载失败', 'error')
  }
}

const accept = async (req) => {
  if (!groupId.value) return
  try {
    await contactsStore.handleGroupRequest(req.group_req_id, String(groupId.value), true)
    toastRef.value?.show('已同意')
  } catch (e) {
    toastRef.value?.show('操作失败', 'error')
  }
}

const reject = async (req) => {
  if (!groupId.value) return
  try {
    await contactsStore.handleGroupRequest(req.group_req_id, String(groupId.value), false)
    toastRef.value?.show('已拒绝')
  } catch (e) {
    toastRef.value?.show('操作失败', 'error')
  }
}

const goBack = () => {
  router.back()
}

const formatTime = (ts) => {
  if (!ts) return '未知'
  return contactsStore.formatTime ? contactsStore.formatTime(ts) : new Date(ts * 1000).toLocaleString()
}

onMounted(async () => {
  await refresh()
})
</script>
<style scoped>
.group-requests-view { max-width: 680px; margin: 40px auto; padding: 0 16px; }
.header { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
h2 { margin: 0; }
.sub { color: #64748b; font-size: 13px; margin-bottom: 16px; }
.back, .refresh { padding: 6px 12px; border-radius: 10px; border: 1px solid #e2e8f0; background: #fff; cursor: pointer; }
.refresh:disabled { opacity: .6; cursor: not-allowed; }
.empty { color: #aaa; text-align: center; margin: 40px 0; }
.request-item { display: flex; align-items: center; gap: 16px; background: #f8fafc; border-radius: 12px; padding: 16px; margin-bottom: 16px; }
.avatar { width: 48px; height: 48px; border-radius: 10px; }
.info { flex: 1; }
.name { font-size: 16px; font-weight: 600; }
.message, .time { font-size: 13px; color: #666; margin-top: 2px; }
.actions { display: flex; flex-direction: column; gap: 8px; }
.actions button { padding: 6px 16px; border-radius: 8px; border: none; background: #2563eb; color: #fff; cursor: pointer; }
.actions button:disabled { opacity: .6; cursor: not-allowed; }
.actions button.danger { background: #e53e3e; }
</style> 