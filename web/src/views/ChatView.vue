<template>
  <div class="chat-view">
    <!-- 左侧会话列表 -->
    <div class="conversation-sidebar">
      <div class="sidebar-header">
        <h2 class="sidebar-title">消息</h2>
        <div class="sidebar-actions">
          <button class="btn-new-chat" @click="startNewChat" title="新建会话">
            <i class="icon icon-plus"></i>
          </button>
          <button class="btn-search" @click="toggleSearch" title="搜索">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useConversationStore } from '../stores/conversation'
import ConversationList from '../components/ConversationList.vue'
import ChatPanel from '../components/ChatPanel.vue'
import ChatEmptyState from '../components/ChatEmptyState.vue'

const conversationStore = useConversationStore()

// 响应式数据
const showSearch = ref(false)
const searchText = ref('')
const activeConversationId = ref(null)

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

function startNewChat() {
  // TODO: 实现新建会话功能
  console.log('开始新会话')
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
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  overflow: hidden;
}

.conversation-sidebar {
  width: 360px;
  min-width: 300px;
  max-width: 400px;
  background: rgba(255, 255, 255, 0.95);
  border-right: 1px solid rgba(229, 231, 235, 0.8);
  backdrop-filter: blur(20px);
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
  border-bottom: 1px solid rgba(229, 231, 235, 0.5);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 250, 252, 0.95) 100%);
}

.sidebar-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(135deg, #1e293b 0%, #4a8cff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
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
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.1) 0%, rgba(138, 105, 255, 0.05) 100%);
  color: #4a8cff;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  position: relative;
  overflow: hidden;
}

.btn-new-chat::before,
.btn-search::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.15) 0%, rgba(138, 105, 255, 0.1) 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.btn-new-chat:hover,
.btn-search:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(74, 140, 255, 0.25);
}

.btn-new-chat:hover::before,
.btn-search:hover::before {
  opacity: 1;
}

.search-container {
  padding: 0 20px 16px 20px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 250, 252, 0.95) 100%);
  border-bottom: 1px solid rgba(229, 231, 235, 0.5);
}

.search-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid rgba(229, 231, 235, 0.6);
  border-radius: 12px;
  font-size: 14px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  outline: none;
}

.search-input:focus {
  border-color: #4a8cff;
  box-shadow: 0 0 0 3px rgba(74, 140, 255, 0.1);
  background: rgba(255, 255, 255, 1);
}

.search-input::placeholder {
  color: #9ca3af;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
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
    border-bottom: 1px solid rgba(229, 231, 235, 0.8);
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