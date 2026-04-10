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
      <div class="user-menu-container">
        <button class="btn-user" @click="toggleUserMenu">
          <img :src="user.avatar" alt="头像" class="avatar">
        </button>
        <div v-if="showUserMenu" class="user-menu">
          <div class="user-menu-item" @click="goToProfile">
            <i class="icon icon-user"></i>
            <span>个人资料</span>
          </div>
          <div class="user-menu-divider"></div>
          <div class="user-menu-item logout-item" @click="handleLogout">
            <i class="icon icon-logout"></i>
            <span>退出登录</span>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const user = computed(() => authStore.user || { avatar: '' })
const showUserMenu = ref(false)

const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
}

const goToProfile = () => {
  showUserMenu.value = false
  router.push('/app/profile')
}

const handleLogout = async () => {
  showUserMenu.value = false
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('登出失败:', error)
    // 即使后端登出失败，也清除本地数据并跳转
    router.push('/login')
  }
}

// 点击外部关闭菜单
const handleClickOutside = (event) => {
  const menuContainer = event.target.closest('.user-menu-container')
  if (!menuContainer) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.app-header {
  background: var(--card);
  color: var(--fg);
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 32px;
  border-bottom: 1px solid var(--border);
}

.header-left, .header-right {
  display: flex;
  align-items: center;
}

.menu-toggle {
  background: none;
  border: none;
  color: var(--fg);
  font-size: 24px;
  cursor: pointer;
  margin-right: 15px;
  transition: color 0.2s;
}

.menu-toggle:hover {
  color: var(--accent);
}

.app-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  font-family: var(--font-display);
  color: var(--fg);
  letter-spacing: 0.02em;
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
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  background-color: var(--input-bg);
  color: var(--fg);
  font-family: var(--font-ui);
  font-size: 14px;
  transition: border-color 0.2s;
}

.search-input::placeholder {
  color: var(--muted);
}

.search-input:focus {
  outline: none;
  border-color: var(--accent);
}

.icon-search {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--muted);
}

.btn-notification {
  background: none;
  border: none;
  color: var(--muted);
  font-size: 20px;
  cursor: pointer;
  margin-right: 20px;
  transition: color 0.2s;
}

.btn-notification:hover {
  color: var(--accent);
}

.btn-user {
  background: none;
  border: none;
  cursor: pointer;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  object-fit: cover;
  border: 2px solid var(--border);
  transition: border-color 0.2s;
}

.avatar:hover {
  border-color: var(--accent);
}

.user-menu-container {
  position: relative;
}

.user-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  min-width: 160px;
  z-index: 1000;
  overflow: hidden;
}

.user-menu-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  color: var(--fg);
  font-size: 14px;
  font-family: var(--font-ui);
}

.user-menu-item:hover {
  background-color: var(--card-hover);
}

.user-menu-item i {
  margin-right: 8px;
  font-size: 16px;
  color: var(--muted);
}

.logout-item {
  color: var(--danger);
}

.logout-item i {
  color: var(--danger);
}

.logout-item:hover {
  background-color: rgba(192, 57, 43, 0.1);
}

.user-menu-divider {
  height: 1px;
  background-color: var(--border);
  margin: 4px 0;
}
</style>