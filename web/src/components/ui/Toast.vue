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
  padding: 10px 20px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  z-index: 9999;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  background: white;
  min-width: 120px;
  justify-content: center;
}

.toast-container.success { color: #10b981; }
.toast-container.error { color: #ef4444; }
.toast-container.warning { color: #f59e0b; }

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
