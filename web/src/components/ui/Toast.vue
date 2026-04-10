<template>
  <Transition name="toast-fade">
    <div v-if="visible" class="toast-container" :class="type">
      <i class="icon" :class="iconClass"></i>
      <span>{{ message }}</span>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  duration: {
    type: Number,
    default: 2000
  }
})

const visible = ref(false)
const message = ref('')
const type = ref('success') // success, error, warning

const iconClass = computed(() => {
  switch (type.value) {
    case 'success': return 'icon-check'
    case 'error': return 'icon-alert-circle'
    case 'warning': return 'icon-alert-triangle'
    default: return 'icon-info'
  }
})

let timer = null

const show = (msg, msgType = 'success') => {
  message.value = msg
  type.value = msgType
  visible.value = true
  
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => {
    visible.value = false
  }, props.duration)
}

defineExpose({ show })
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 40px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-family: var(--font-ui);
  font-weight: 500;
  z-index: 9999;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  background: var(--card);
  border: 1px solid var(--border);
  color: var(--fg);
  min-width: 120px;
  justify-content: center;
}

.toast-container.success { color: var(--success); }
.toast-container.error { color: var(--danger); }
.toast-container.warning { color: var(--accent); }

.icon { font-size: 18px; }

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.3s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}
</style>
