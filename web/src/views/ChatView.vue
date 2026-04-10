<template>
  <div class="chat-view">
    <!-- 左侧会话列表 -->
    <div class="conversation-sidebar">
      <div class="sidebar-header">
        <h2 class="sidebar-title">消息</h2>
        <div class="sidebar-actions">
          <button class="btn-new-chat" @click="showSearchUserModal = true" title="搜索用户">
            <i class="icon icon-plus"></i>
          </button>
          <button class="btn-search" @click="toggleSearch" title="搜索会话">
            <i class="icon icon-search"></i>
          </button>
        </div>
      </div>
      
      <!-- 搜索框 -->
      <div class="search-container" v-if="showSearch">
        <input 
          type="text" 
          class="search-input" 
          placeholder="搜索会话..." 
          v-model="searchText"
          @input="handleSearch"
        >
      </div>
      
      <!-- 会话列表 -->
      <ConversationList 
        :conversations="filteredConversations"
        :active-conversation-id="activeConversationId"
        @select-conversation="selectConversation"
        @delete-conversation="deleteConversation"
        @toggle-pin="togglePin"
        @toggle-mute="toggleMute"
      />
    </div>
    
    <!-- 右侧聊天区域 -->
    <div class="chat-main">
      <!-- 空状态 -->
      <ChatEmptyState v-if="!activeConversationId" @start-chat="handleStartChat" />
      
      <!-- 聊天面板 -->
      <ChatPanel 
        v-else
        :conversation="activeConversation"
        @close-chat="closeChat"
        @send-message="handleSendMessage"
      />
    </div>
    
    <!-- 搜索用户模态框 -->
    <SearchUserModal 
      v-if="showSearchUserModal"
      @close="showSearchUserModal = false"
      @start-chat="handleStartChatFromSearch"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useConversationStore } from '../stores/conversation'
import ConversationList from '../components/ConversationList.vue'
import ChatPanel from '../components/ChatPanel.vue'
import ChatEmptyState from '../components/ChatEmptyState.vue'
import SearchUserModal from '../components/SearchUserModal.vue'

const conversationStore = useConversationStore()

// 响应式数据
const showSearch = ref(false)
const searchText = ref('')
const activeConversationId = ref(null)
const showSearchUserModal = ref(false)

// 计算属性
const activeConversation = computed(() => {
  return conversationStore.conversations.find(conv => conv.id === activeConversationId.value)
})

const filteredConversations = computed(() => {
  if (!searchText.value) {
    return conversationStore.conversations
  }
  return conversationStore.conversations.filter(conv => 
    conv.name.toLowerCase().includes(searchText.value.toLowerCase()) ||
    conv.lastMessage.toLowerCase().includes(searchText.value.toLowerCase())
  )
})

// 方法
function selectConversation(conversationId) {
  activeConversationId.value = conversationId
  conversationStore.setActiveConversation(conversationId)
}

function closeChat() {
  activeConversationId.value = null
  conversationStore.setActiveConversation(null)
}

function handleStartChatFromSearch(userId) {
  // 从搜索结果开始聊天
  showSearchUserModal.value = false
  const conversation = conversationStore.getOrCreateConversation(userId)
  selectConversation(conversation.id)
}

function handleStartChat(userId) {
  // 从空状态开始聊天
  const conversation = conversationStore.getOrCreateConversation(userId)
  selectConversation(conversation.id)
}

function toggleSearch() {
  showSearch.value = !showSearch.value
  if (!showSearch.value) {
    searchText.value = ''
  }
}

function handleSearch() {
  // 搜索逻辑已在computed中处理
}

function deleteConversation(conversationId) {
  conversationStore.deleteConversation(conversationId)
  if (activeConversationId.value === conversationId) {
    closeChat()
  }
}

function handleSendMessage(messageData) {
  conversationStore.sendMessage(activeConversationId.value, messageData)
}

function togglePin(conversationId) {
  conversationStore.togglePinConversation(conversationId)
}

function toggleMute(conversationId) {
  conversationStore.toggleMuteConversation(conversationId)
}

// 生命周期
onMounted(() => {
  // 初始化会话数据
  conversationStore.initializeConversations()
  
  // 检查是否有预设的活跃会话（从好友详情页跳转过来）
  const activeConversation = conversationStore.activeConversation
  if (activeConversation) {
    activeConversationId.value = activeConversation.id
  }
})

// 监听会话状态变化
watch(() => conversationStore.activeConversation, (newConversation) => {
  if (newConversation) {
    activeConversationId.value = newConversation.id
  }
})
</script>

<style scoped>
.chat-view {
  display: flex;
  height: 100vh;
  background: var(--bg);
  overflow: hidden;
}

.conversation-sidebar {
  width: 360px;
  min-width: 300px;
  max-width: 400px;
  background: var(--card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 10;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 20px 16px 20px;
  border-bottom: 1px solid var(--border);
  background: var(--card);
}

.sidebar-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  font-family: var(--font-display);
  color: var(--fg);
}

.sidebar-actions {
  display: flex;
  gap: 8px;
}

.btn-new-chat,
.btn-search {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent-muted);
  color: var(--accent);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  position: relative;
  overflow: hidden;
}

.btn-new-chat:hover,
.btn-search:hover {
  background: var(--accent);
  color: var(--bg);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(200, 149, 108, 0.25);
}

.search-container {
  padding: 0 20px 16px 20px;
  background: var(--card);
  border-bottom: 1px solid var(--border);
}

.search-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: var(--font-ui);
  background: var(--input-bg);
  color: var(--fg);
  transition: all 0.3s ease;
  outline: none;
}

.search-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(200, 149, 108, 0.15);
}

.search-input::placeholder {
  color: var(--muted);
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg);
  position: relative;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chat-view {
    flex-direction: column;
  }

  .conversation-sidebar {
    width: 100%;
    height: 50vh;
    min-width: unset;
    max-width: unset;
    border-right: none;
    border-bottom: 1px solid var(--border);
  }

  .chat-main {
    height: 50vh;
  }

  .sidebar-header {
    padding: 16px;
  }

  .sidebar-title {
    font-size: 20px;
  }
}

@media (max-width: 480px) {
  .conversation-sidebar {
    height: 40vh;
  }

  .chat-main {
    height: 60vh;
  }
}
</style>