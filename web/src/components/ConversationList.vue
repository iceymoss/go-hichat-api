<template>
  <div class="conversation-list">
    <div class="conversation-scroll">
      <div 
        v-for="conversation in sortedConversations" 
        :key="conversation.id"
        class="conversation-item"
        :class="{ 'active': conversation.id === activeConversationId }"
        @click="selectConversation(conversation.id)"
      >
        <div class="conversation-avatar">
          <img :src="conversation.avatar" :alt="conversation.name" />
          <div 
            v-if="conversation.unreadCount > 0" 
            class="unread-badge"
          >{{ conversation.unreadCount > 99 ? '99+' : conversation.unreadCount }}</div>
          <div 
            v-if="conversation.isOnline" 
            class="online-indicator"
          ></div>
        </div>
        
        <div class="conversation-content">
          <div class="conversation-header">
            <h4 class="conversation-name">{{ conversation.name }}</h4>
            <span class="conversation-time">{{ formatTime(conversation.lastMessageTime) }}</span>
          </div>
          
          <div class="conversation-preview">
            <span class="last-message" :class="{ 'unread': conversation.unreadCount > 0 }">
              {{ formatLastMessage(conversation.lastMessage, conversation.lastMessageType) }}
            </span>
            <div class="conversation-status">
              <i v-if="conversation.isPinned" class="icon icon-pin status-icon"></i>
              <i v-if="conversation.isMuted" class="icon icon-volume-off status-icon"></i>
            </div>
          </div>
        </div>
        
        <div class="conversation-actions">
          <button 
            class="btn-action" 
            @click.stop="togglePin(conversation.id)"
            :title="conversation.isPinned ? '取消置顶' : '置顶'"
          >
            <i class="icon" :class="conversation.isPinned ? 'icon-pin-filled' : 'icon-pin'"></i>
          </button>
          <button 
            class="btn-action" 
            @click.stop="toggleMute(conversation.id)"
            :title="conversation.isMuted ? '取消静音' : '静音'"
          >
            <i class="icon" :class="conversation.isMuted ? 'icon-volume-off' : 'icon-volume'"></i>
          </button>
          <button 
            class="btn-action btn-delete" 
            @click.stop="deleteConversation(conversation.id)"
            title="删除会话"
          >
            <i class="icon icon-trash"></i>
          </button>
        </div>
      </div>
      
      <!-- 空状态 -->
      <div v-if="conversations.length === 0" class="empty-conversations">
        <div class="empty-icon">
          <i class="icon icon-message-circle"></i>
        </div>
        <div class="empty-text">暂无会话</div>
        <div class="empty-desc">开始一段新的对话吧</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  conversations: {
    type: Array,
    default: () => []
  },
  activeConversationId: {
    type: [String, Number],
    default: null
  }
})

const emit = defineEmits(['select-conversation', 'delete-conversation', 'toggle-pin', 'toggle-mute'])

// 按置顶和时间排序的会话列表
const sortedConversations = computed(() => {
  return [...props.conversations].sort((a, b) => {
    // 置顶的会话排在前面
    if (a.isPinned && !b.isPinned) return -1
    if (!a.isPinned && b.isPinned) return 1
    
    // 按最后消息时间排序
    return new Date(b.lastMessageTime) - new Date(a.lastMessageTime)
  })
})

function selectConversation(conversationId) {
  emit('select-conversation', conversationId)
}

function deleteConversation(conversationId) {
  if (confirm('确定要删除这个会话吗？')) {
    emit('delete-conversation', conversationId)
  }
}

function togglePin(conversationId) {
  emit('toggle-pin', conversationId)
}

function toggleMute(conversationId) {
  emit('toggle-mute', conversationId)
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  
  const now = new Date()
  const messageTime = new Date(timestamp)
  const diff = now - messageTime
  
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 0) {
    if (days < 7) {
      return `${days}天前`
    } else {
      return messageTime.toLocaleDateString()
    }
  } else if (hours > 0) {
    return `${hours}小时前`
  } else if (minutes > 0) {
    return `${minutes}分钟前`
  } else {
    return '刚刚'
  }
}

function formatLastMessage(message, type) {
  if (!message) return '暂无消息'
  
  switch (type) {
    case 'image':
      return '[图片]'
    case 'voice':
      return '[语音]'
    case 'file':
      return '[文件]'
    case 'emoji':
      return '[表情]'
    default:
      return message.length > 30 ? message.substring(0, 30) + '...' : message
  }
}
</script>

<style scoped>
.conversation-list {
  flex: 1;
  overflow: hidden;
}

.conversation-scroll {
  height: 100%;
  overflow-y: auto;
  padding: 8px 0;
  
  /* 自定义滚动条 */
  scrollbar-width: thin;
  scrollbar-color: rgba(74, 140, 255, 0.3) transparent;
}

.conversation-scroll::-webkit-scrollbar {
  width: 6px;
}

.conversation-scroll::-webkit-scrollbar-track {
  background: transparent;
}

.conversation-scroll::-webkit-scrollbar-thumb {
  background: rgba(74, 140, 255, 0.3);
  border-radius: 3px;
}

.conversation-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(74, 140, 255, 0.5);
}

.conversation-item {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  border-radius: 0 24px 24px 0;
  margin: 2px 0 2px 8px;
  background: transparent;
}

.conversation-item:hover {
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.06) 0%, rgba(138, 105, 255, 0.03) 100%);
  transform: translateX(8px);
}

.conversation-item.active {
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.12) 0%, rgba(138, 105, 255, 0.06) 100%);
  border-left: 4px solid #4a8cff;
  transform: translateX(4px);
}

.conversation-item.active::before {
  content: '';
  position: absolute;
  left: -8px;
  top: 0;
  bottom: 0;
  width: 4px;
  background: linear-gradient(to bottom, #4a8cff, #8a69ff);
  border-radius: 2px;
}

.conversation-avatar {
  position: relative;
  margin-right: 12px;
  flex-shrink: 0;
}

.conversation-avatar img {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.conversation-item:hover .conversation-avatar img {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(74, 140, 255, 0.2);
}

.unread-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
  border-radius: 12px;
  padding: 2px 6px;
  font-size: 10px;
  font-weight: 700;
  min-width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
  animation: badgePulse 2s infinite;
}

@keyframes badgePulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.online-indicator {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 14px;
  height: 14px;
  background: linear-gradient(135deg, #10b981, #059669);
  border: 2px solid white;
  border-radius: 50%;
  box-shadow: 0 2px 6px rgba(16, 185, 129, 0.3);
}

.conversation-content {
  flex: 1;
  min-width: 0;
  margin-right: 8px;
}

.conversation-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.conversation-name {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  margin-right: 8px;
}

.conversation-time {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
  white-space: nowrap;
}

.conversation-preview {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.last-message {
  font-size: 14px;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  margin-right: 8px;
}

.last-message.unread {
  color: #374151;
  font-weight: 600;
}

.conversation-status {
  display: flex;
  gap: 4px;
}

.status-icon {
  font-size: 12px;
  color: #9ca3af;
}

.conversation-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.conversation-item:hover .conversation-actions {
  opacity: 1;
}

.btn-action {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 6px;
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.btn-action:hover {
  background: rgba(74, 140, 255, 0.1);
  color: #4a8cff;
  transform: scale(1.1);
}

.btn-delete:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.empty-conversations {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  height: calc(100% - 120px);
}

.empty-icon {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, rgba(107, 114, 128, 0.1), rgba(156, 163, 175, 0.05));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.empty-icon .icon {
  font-size: 36px;
  color: #9ca3af;
}

.empty-text {
  font-size: 18px;
  font-weight: 600;
  color: #6b7280;
  margin-bottom: 8px;
}

.empty-desc {
  font-size: 14px;
  color: #9ca3af;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .conversation-item {
    padding: 12px 16px;
    margin: 1px 0 1px 4px;
  }
  
  .conversation-avatar img {
    width: 40px;
    height: 40px;
  }
  
  .conversation-name {
    font-size: 15px;
  }
  
  .last-message {
    font-size: 13px;
  }
  
  .conversation-actions {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    opacity: 0.7;
  }
}
</style>