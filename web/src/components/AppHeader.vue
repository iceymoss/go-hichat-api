<template>
  <header class="app-header">
    <div class="header-left">
      <button class="menu-toggle" @click="$emit('toggle-sidebar')">
        <i class="icon icon-menu"></i>
      </button>
      <h1 class="app-title">HiChat2</h1>
    </div>
    
    <div class="header-center">
      <div class="search-container">
        <input type="text" placeholder="搜索..." class="search-input">
        <i class="icon icon-search"></i>
      </div>
    </div>
    
    <div class="header-right">
      <button class="btn-notification">
        <i class="icon icon-bell"></i>
      </button>
      <button class="btn-user">
        <img :src="user.avatar" alt="头像" class="avatar">
      </button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const user = computed(() => authStore.user || { avatar: '' })
</script>

<style scoped>
.app-header {
  background: linear-gradient(90deg, #4a8cff, #8a69ff);
  color: #fff;
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 32px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.header-left, .header-right {
  display: flex;
  align-items: center;
}

.menu-toggle {
  background: none;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  margin-right: 15px;
}

.app-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.header-center {
  flex: 1;
  max-width: 500px;
}

.search-container {
  position: relative;
  max-width: 400px;
  margin: 0 auto;
}

.search-input {
  width: 100%;
  padding: 8px 15px 8px 35px;
  border-radius: 18px;
  border: none;
  background-color: #444a55;
  color: white;
}

.icon-search {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #bbb;
}

.btn-notification {
  background: none;
  border: none;
  color: white;
  font-size: 20px;
  cursor: pointer;
  margin-right: 20px;
}

.btn-user {
  background: none;
  border: none;
  cursor: pointer;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
}
</style>