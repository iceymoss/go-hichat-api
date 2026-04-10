<template>
  <div id="app">
    <router-view />
    <div v-if="globalLoading" class="global-loading">
      <div class="loading-spinner"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const globalLoading = ref(true)

onMounted(() => {
  setTimeout(() => {
    globalLoading.value = false
  }, 1500)
})
</script>

<style>
.global-loading {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.loading-spinner {
  width: 48px;
  height: 48px;
  border: 3px solid var(--border);
  border-radius: 50%;
  border-top-color: var(--accent);
  animation: spin 0.8s infinite linear;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
