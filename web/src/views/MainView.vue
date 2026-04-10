<template>
  <div class="main-view main-bg">
    <AppHeader />
    <div class="main-nav">
      <router-link 
        v-for="tab in tabs" 
        :key="tab.key" 
        :to="tab.route"
        :class="['nav-item', { active: $route.name === tab.routeName }]"
      >
        <i :class="['icon', tab.icon]"></i>
        <span>{{ tab.label }}</span>
      </router-link>
    </div>
    <div class="main-body">
      <!-- 使用router-view来渲染子路由 -->
      <router-view />
    </div>
  </div>
</template>

<script setup>
import AppHeader from '../components/AppHeader.vue'

// 导航配置
const tabs = [
  { 
    key: 'chat', 
    label: '会话', 
    icon: 'icon-chat', 
    route: '/app/chat',
    routeName: 'Chat'
  },
  { 
    key: 'contacts', 
    label: '好友', 
    icon: 'icon-contacts', 
    route: '/app/contacts',
    routeName: 'Contacts'
  },
  { 
    key: 'discover', 
    label: '动态', 
    icon: 'icon-feed', 
    route: '/app/discover',
    routeName: 'Discover'
  },
  { 
    key: 'profile', 
    label: '我的', 
    icon: 'icon-user', 
    route: '/app/profile',
    routeName: 'Profile'
  }
]
</script>

<style scoped>
.main-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  background: var(--bg);
}
.main-nav {
  height: 56px;
  background: var(--card);
  display: flex;
  align-items: center;
  justify-content: flex-start;
  border-bottom: 1px solid var(--border);
  padding-left: 32px;
}
.nav-item {
  margin-right: 36px;
  color: var(--muted);
  font-size: 16px;
  padding: 8px 0 2px 0;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
  display: flex;
  align-items: center;
  font-weight: 500;
  font-family: var(--font-ui);
  text-decoration: none;
  border-bottom: 2.5px solid transparent;
}

.nav-item:hover {
  color: var(--accent-hover);
}

.nav-item.active {
  color: var(--accent);
  font-weight: bold;
  border-bottom: 2.5px solid var(--accent);
}

.nav-item.router-link-active {
  color: var(--accent);
  font-weight: bold;
  border-bottom: 2.5px solid var(--accent);
}
.nav-item .icon {
  font-size: 22px;
  margin-right: 6px;
}
.main-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  background: var(--bg);
}
</style>