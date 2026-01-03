<template>
  <aside :class="['app-sidebar', { open: open }]">
    <nav class="nav-links">
      <router-link to="/app/chat" class="nav-item" active-class="active">
        <i class="icon icon-chat"></i>
        <span>会话</span>
      </router-link>
      <router-link to="/app/contacts" class="nav-item" active-class="active">
        <i class="icon icon-contacts"></i>
        <span>好友</span>
      </router-link>
      <router-link to="/app/discover" class="nav-item" active-class="active">
        <i class="icon icon-discover"></i>
        <span>动态</span>
      </router-link>
      <router-link to="/app/profile" class="nav-item" active-class="active">
        <i class="icon icon-profile"></i>
        <span>我</span>
      </router-link>
    </nav>
    
    <div class="sidebar-conversations">
      <h3 class="sidebar-title">聊天列表</h3>
      <ConversationList />
    </div>
  </aside>
</template>

<script setup>
import ConversationList from './ConversationList.vue'

defineProps({
  open: {
    type: Boolean,
    default: true
  }
})
</script>

<style scoped>
.app-sidebar {
  width: 300px;
  height: 100%;
  background: linear-gradient(135deg, #4a8cffcc, #8a69ffcc);
  color: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: transform 0.3s ease;
}

.app-sidebar:not(.open) {
  transform: translateX(-100%);
}

.nav-links {
  display: flex;
  padding: 10px 15px;
  border-bottom: 1px solid #2a2f38;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 15px;
  text-decoration: none;
  color: #ccc;
  font-size: 14px;
  border-radius: 8px;
}

.nav-item.active {
  background-color: #2e323b;
  color: #4a8cff;
}

.nav-item i {
  font-size: 22px;
  margin-bottom: 5px;
}

.sidebar-conversations {
  flex: 1;
  overflow-y: auto;
}

.sidebar-title {
  padding: 15px 20px;
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: #888;
}
</style>