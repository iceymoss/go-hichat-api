<template>
  <div class="chat-panel" v-if="conversation">
    <!-- 聊天头部 -->
    <div class="chat-header">
      <div class="header-left">
        <img 
          :src="conversation.avatar" 
          :alt="conversation.name" 
          class="avatar"
          @mouseenter="showHeaderUserProfile($event, conversation)"
          @mouseleave="hideUserProfile"
        >
        <div class="user-info">
          <h3 class="name">{{ conversation.name }}</h3>
          <div class="status" :class="{ 'online': conversation.isOnline }">
            {{ conversation.isOnline ? '在线' : '离线' }}
          </div>
        </div>
      </div>
      
      <div class="header-actions">
        <button class="action-btn" @click="handleVoiceCall" title="语音通话">
          <i class="icon icon-phone"></i>
        </button>
        <button class="action-btn" @click="handleVideoCall" title="视频通话">
          <i class="icon icon-video"></i>
        </button>
        <button class="action-btn" @click="handleMoreActions" title="更多">
          <i class="icon icon-more"></i>
        </button>
        <button class="action-btn close-btn" @click="handleClose" title="关闭">
          <i class="icon icon-x"></i>
        </button>
      </div>
    </div>
    
    <!-- 消息区域 -->
    <div class="chat-messages" ref="messagesContainer" @scroll="handleScroll">
      <!-- 加载历史消息指示器 -->
      <div v-if="isLoadingHistory" class="loading-history">
        <div class="loading-spinner"></div>
        <span>加载历史消息...</span>
      </div>
      
      <!-- 消息列表 -->
      <div class="messages-list">
        <div 
          v-for="(message, index) in displayMessages" 
          :key="message.id"
          class="message-wrapper"
          :class="{ 
            'own-message': message.sender === 'me',
            'other-message': message.sender !== 'me',
            'group-message': conversation.type === 'group' && message.sender !== 'me'
          }"
        >
          <!-- 消息时间分隔线 -->
          <div v-if="shouldShowTimeDivider(message, index)" class="time-divider">
            {{ formatMessageDate(message.timestamp) }}
          </div>
          
          <!-- 消息内容 -->
          <div class="message-item">
            <!-- 其他人的头像（群聊时显示） -->
            <img 
              v-if="conversation.type === 'group' && message.sender !== 'me'" 
              :src="message.senderAvatar" 
              :alt="message.senderName"
              class="message-avatar"
              @mouseenter="showMessageUserProfile($event, message)"
              @mouseleave="hideUserProfile"
            />
            
            <div class="message-content">
              <!-- 发送者名称（群聊中其他人的消息） -->
              <div 
                v-if="conversation.type === 'group' && message.sender !== 'me'" 
                class="sender-name"
              >
                {{ message.senderName }}
              </div>
              
              <!-- 消息主体 -->
              <div class="message-body" :class="`message-${message.type}`">
                <!-- 文字消息 -->
                <div v-if="message.type === 'text'" class="text-message">
                  {{ message.content }}
                </div>
                
                <!-- 图片消息 -->
                <div v-else-if="message.type === 'image'" class="image-message">
                  <img :src="message.content" alt="图片" @click="viewImage(message.content)" />
                </div>
                
                <!-- 语音消息 -->
                <div v-else-if="message.type === 'voice'" class="voice-message">
                  <button class="voice-play-btn" @click="playVoice(message)">
                    <i class="icon" :class="message.isPlaying ? 'icon-pause' : 'icon-play'"></i>
                  </button>
                  <div class="voice-wave">
                    <div class="wave-bar" v-for="i in 12" :key="i"></div>
                  </div>
                  <span class="voice-duration">{{ message.duration || '0:15' }}</span>
                </div>
                
                <!-- 文件消息 -->
                <div v-else-if="message.type === 'file'" class="file-message" @click="downloadFile(message)">
                  <div class="file-icon" :class="getFileIconClass(message)">
                    <i :class="['icon', getFileIcon(message)]"></i>
                  </div>
                  <div class="file-info">
                    <div class="file-name">{{ message.fileName || '文件' }}</div>
                    <div class="file-size">{{ message.fileSize || '1.2MB' }}</div>
                    <div class="file-type">{{ getFileTypeLabel(message) }}</div>
                  </div>
                  <button class="file-download" @click.stop="downloadFile(message)">
                    <i class="icon icon-eye"></i>
                  </button>
                </div>
                
                <!-- 表情消息 -->
                <div v-else-if="message.type === 'emoji'" class="emoji-message">
                  <span class="emoji-large">{{ message.content }}</span>
                </div>
              </div>
              
              <!-- 消息状态和时间 -->
              <div class="message-meta">
                <span class="message-time">{{ formatMessageTime(message.timestamp) }}</span>
                <span v-if="message.sender === 'me'" class="message-status" :class="message.status">
                  <i class="icon" :class="getStatusIcon(message.status)"></i>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 滚动到底部按钮 -->
      <Transition name="scroll-btn">
        <button 
          v-if="showScrollToBottom" 
          class="scroll-to-bottom"
          @click="scrollToBottom"
        >
          <i class="icon icon-chevron-down"></i>
          <span v-if="unreadCount > 0" class="unread-count">{{ unreadCount }}</span>
        </button>
      </Transition>
    </div>
    
    <!-- 输入区域 -->
    <div class="chat-input-area">
      <!-- 工具栏 -->
      <div class="input-toolbar">
        <button class="tool-btn" @click="toggleEmojiPicker" title="表情">
          <i class="icon icon-emoji"></i>
        </button>
        <button class="tool-btn" @click="selectImage" title="图片">
          <i class="icon icon-image"></i>
          <input ref="imageInput" type="file" accept="image/*" @change="handleImageSelect" style="display: none;" />
        </button>
        <button class="tool-btn" @click="selectFile" title="文件">
          <i class="icon icon-file"></i>
          <input ref="fileInput" type="file" @change="handleFileSelect" style="display: none;" />
        </button>
        <button 
          class="tool-btn voice-btn" 
          :class="{ 'recording': isRecording }"
          @mousedown="startVoiceRecord"
          @mouseup="stopVoiceRecord"
          @mouseleave="stopVoiceRecord"
          title="按住录音"
        >
          <i class="icon icon-mic"></i>
        </button>
      </div>
      
      <!-- 输入框 -->
      <div class="input-container">
        <div 
          class="message-input"
          ref="messageInput"
          :contenteditable="!isRecording"
          :data-placeholder="isRecording ? '正在录音...' : '输入消息...'"
          @input="handleInput"
          @keydown="handleKeydown"
          @paste="handlePaste"
        ></div>
        
        <button 
          class="send-btn" 
          :class="{ 'has-content': hasInputContent }"
          @click="sendMessage"
          :disabled="!hasInputContent && !isRecording"
        >
          <i class="icon icon-send"></i>
        </button>
      </div>
      
      <!-- 录音状态 -->
      <Transition name="recording">
        <div v-if="isRecording" class="recording-status">
          <div class="recording-wave">
            <div class="wave" v-for="i in 5" :key="i"></div>
          </div>
          <span class="recording-text">正在录音... 松开发送</span>
          <span class="recording-time">{{ recordingDuration }}s</span>
        </div>
      </Transition>
      
      <!-- 表情选择器 -->
      <Transition name="emoji-picker">
        <div v-if="showEmojiPicker" class="emoji-picker">
          <div class="emoji-grid">
            <button 
              v-for="emoji in emojiList" 
              :key="emoji"
              class="emoji-item"
              @click="insertEmoji(emoji)"
            >
              {{ emoji }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
    
    <!-- 用户资料卡片 -->
    <UserProfileCard
      :visible="showUserProfileCard"
      :user="currentProfileUser"
      :position="userProfilePosition"
      :arrow-position="userProfileArrowDirection"
      @close="closeUserProfileCard"
      @view-profile="handleViewProfile"
      @view-space="handleViewSpace"
      @mouse-enter-card="clearProfileCardTimer"
      @mouse-leave-card="hideUserProfileDelayed"
    />
    
    <!-- 图片预览 -->
    <ImagePreview
      v-if="showImagePreview"
      :visible="showImagePreview"
      :images="previewImages"
      :initial-index="currentImageIndex"
      @close="closeImagePreview"
      @change="onImageChange"
    />
    
    <!-- 文件预览 -->
    <FilePreview
      v-if="showFilePreview"
      :visible="showFilePreview"
      :file="currentPreviewFile"
      @close="closeFilePreview"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useConversationStore } from '../stores/conversation'
import UserProfileCard from './UserProfileCard.vue'
import ImagePreview from './ImagePreview.vue'
import FilePreview from './FilePreview.vue'

const props = defineProps({
  conversation: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close-chat', 'send-message'])

const conversationStore = useConversationStore()

// 响应式数据
const messagesContainer = ref(null)
const messageInput = ref(null)
const imageInput = ref(null)
const fileInput = ref(null)

const isLoadingHistory = ref(false)
const showScrollToBottom = ref(false)
const unreadCount = ref(0)
const hasInputContent = ref(false)
const inputContent = ref('')

// 表情相关
const showEmojiPicker = ref(false)
const emojiList = ['😀', '😃', '😄', '😁', '😆', '😅', '😂', '🤣', '😊', '😇', '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘', '😗', '😙', '😚', '😋', '😛', '😝', '😜', '🤪', '🤨', '🧐', '🤓', '😎', '🤩', '🥳', '😏', '😒', '😞', '😔', '😟', '😕', '🙁', '☹️', '😣', '😖', '😫', '😩', '🥺', '😢', '😭', '😤', '😠', '😡', '🤬', '🤯', '😳', '🥵', '🥶', '😱', '😨', '😰', '😥', '😓', '🤗', '🤔', '🤭', '🤫', '🤥', '😶', '😐', '😑', '😬', '🙄', '😯', '😦', '😧', '😮', '😲', '🥱', '😴', '🤤', '😪', '😵', '🤐', '🥴', '🤢', '🤮', '🤧', '😷', '🤒', '🤕']

// 录音相关
const isRecording = ref(false)
const recordingDuration = ref(0)
let recordingTimer = null

// 用户资料卡片相关
const showUserProfileCard = ref(false)
const currentProfileUser = ref(null)
const userProfilePosition = ref({ top: 0, left: 0 })
const userProfileArrowDirection = ref('bottom')
let profileCardTimer = null

// 图片预览相关
const showImagePreview = ref(false)
const previewImages = ref([])
const currentImageIndex = ref(0)

// 文件预览相关
const showFilePreview = ref(false)
const currentPreviewFile = ref(null)

// 计算属性
const displayMessages = computed(() => {
  if (!props.conversation) return []
  return conversationStore.getConversationMessages(props.conversation.id)
})

// 方法
function handleScroll() {
  if (!messagesContainer.value) return
  
  const { scrollTop, scrollHeight, clientHeight } = messagesContainer.value
  
  // 检查是否需要加载历史消息
  if (scrollTop === 0 && !isLoadingHistory.value) {
    loadMoreHistory()
  }
  
  // 检查是否显示回到底部按钮
  const isNearBottom = scrollHeight - scrollTop - clientHeight < 100
  showScrollToBottom.value = !isNearBottom
}

async function loadMoreHistory() {
  if (!props.conversation || isLoadingHistory.value) return
  
  isLoadingHistory.value = true
  const previousScrollHeight = messagesContainer.value.scrollHeight
  
  try {
    await conversationStore.loadMoreMessages(props.conversation.id)
    
    // 保持滚动位置
    await nextTick()
    const newScrollHeight = messagesContainer.value.scrollHeight
    messagesContainer.value.scrollTop = newScrollHeight - previousScrollHeight
  } finally {
    isLoadingHistory.value = false
  }
}

function scrollToBottom(smooth = true) {
  if (!messagesContainer.value) return
  
  const { scrollHeight, clientHeight } = messagesContainer.value
  messagesContainer.value.scrollTo({
    top: scrollHeight - clientHeight,
    behavior: smooth ? 'smooth' : 'instant'
  })
}

function handleInput(event) {
  const content = event.target.textContent || ''
  inputContent.value = content
  hasInputContent.value = content.trim().length > 0
}

function handleKeydown(event) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

function handlePaste(event) {
  event.preventDefault()
  const text = event.clipboardData.getData('text/plain')
  document.execCommand('insertText', false, text)
}

function sendMessage() {
  if (!hasInputContent.value || !props.conversation) return
  
  const messageData = {
    content: inputContent.value.trim(),
    type: 'text'
  }
  
  emit('send-message', messageData)
  
  // 清空输入框
  if (messageInput.value) {
    messageInput.value.textContent = ''
    inputContent.value = ''
    hasInputContent.value = false
  }
  
  // 滚动到底部
  nextTick(() => scrollToBottom())
}

// 表情功能
function toggleEmojiPicker() {
  showEmojiPicker.value = !showEmojiPicker.value
}

function insertEmoji(emoji) {
  if (messageInput.value) {
    const selection = window.getSelection()
    const range = selection.getRangeAt(0)
    
    // 插入表情
    const emojiNode = document.createTextNode(emoji)
    range.insertNode(emojiNode)
    
    // 移动光标到表情后
    range.setStartAfter(emojiNode)
    range.setEndAfter(emojiNode)
    selection.removeAllRanges()
    selection.addRange(range)
    
    // 更新内容
    handleInput({ target: messageInput.value })
  }
  
  showEmojiPicker.value = false
}

// 图片功能
function selectImage() {
  imageInput.value?.click()
}

function handleImageSelect(event) {
  const file = event.target.files[0]
  if (file) {
    // 模拟上传图片
    const reader = new FileReader()
    reader.onload = (e) => {
      const messageData = {
        content: e.target.result,
        type: 'image'
      }
      emit('send-message', messageData)
      nextTick(() => scrollToBottom())
    }
    reader.readAsDataURL(file)
    
    // 清空文件输入
    event.target.value = ''
  }
}

// 文件功能
function selectFile() {
  fileInput.value?.click()
}

function handleFileSelect(event) {
  const file = event.target.files[0]
  if (file) {
    const fileUrl = URL.createObjectURL(file)
    const fileExtension = file.name.split('.').pop()?.toLowerCase() || ''
    
    const messageData = {
      content: fileUrl,
      type: 'file',
      fileName: file.name,
      fileSize: formatFileSize(file.size),
      fileExtension: fileExtension,
      mimeType: file.type,
      originalFile: file  // 保存原始文件对象以便下载
    }
    
    emit('send-message', messageData)
    nextTick(() => scrollToBottom())
    
    // 清空文件输入
    event.target.value = ''
  }
}

// 录音功能
function startVoiceRecord() {
  if (isRecording.value) return
  
  isRecording.value = true
  recordingDuration.value = 0
  
  recordingTimer = setInterval(() => {
    recordingDuration.value++
  }, 1000)
}

function stopVoiceRecord() {
  if (!isRecording.value) return
  
  isRecording.value = false
  clearInterval(recordingTimer)
  
  if (recordingDuration.value > 0) {
    const messageData = {
      content: `voice_${Date.now()}.mp3`,
      type: 'voice',
      duration: `0:${recordingDuration.value.toString().padStart(2, '0')}`
    }
    emit('send-message', messageData)
    nextTick(() => scrollToBottom())
  }
  
  recordingDuration.value = 0
}

// 多媒体功能
function viewImage(imageSrc) {
  // 获取对话中的所有图片
  const allImages = displayMessages.value
    .filter(msg => msg.type === 'image')
    .map(msg => msg.content)
  
  openImagePreview(imageSrc, allImages)
}

function playVoice(message) {
  // TODO: 实现语音播放
  message.isPlaying = !message.isPlaying
  console.log('播放语音:', message.content)
}

function downloadFile(message) {
  // 打开文件预览
  const file = {
    src: message.content,
    name: message.fileName || '未知文件',
    size: parseFileSize(message.fileSize || '0B'),
    extension: message.fileExtension || getFileExtension(message.fileName),
    mimeType: message.mimeType || '',
    originalFile: message.originalFile
  }
  
  openFilePreview(file)
}

// 辅助函数：获取文件扩展名
function getFileExtension(fileName) {
  if (!fileName) return ''
  return fileName.split('.').pop()?.toLowerCase() || ''
}

// 通话功能
function handleVoiceCall() {
  console.log('发起语音通话')
}

function handleVideoCall() {
  console.log('发起视频通话')
}

function handleMoreActions() {
  console.log('更多操作')
}

function handleClose() {
  emit('close-chat')
}

// 用户资料卡片功能
function showHeaderUserProfile(event, user) {
  clearProfileCardTimer()
  
  const rect = event.target.getBoundingClientRect()
  currentProfileUser.value = {
    id: user.id || 'current',
    name: user.name,
    avatar: user.avatar,
    description: user.bio || '这个人很懒，什么都没留下...',
    isOnline: user.isOnline || false,
    feedCount: Math.floor(Math.random() * 100) + 20,
    friendCount: Math.floor(Math.random() * 1000) + 100,
    likeCount: Math.floor(Math.random() * 5000) + 500
  }
  
  // 计算位置
  const cardWidth = 320
  const cardHeight = 200
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  
  let left = rect.left + rect.width / 2 - cardWidth / 2
  let top = rect.bottom + 10
  let arrowDirection = 'top'
  
  // 边界检测
  if (left < 10) {
    left = 10
  } else if (left + cardWidth > viewportWidth - 10) {
    left = viewportWidth - cardWidth - 10
  }
  
  if (top + cardHeight > viewportHeight - 10) {
    top = rect.top - cardHeight - 10
    arrowDirection = 'bottom'
  }
  
  userProfilePosition.value = { x: left + cardWidth / 2, y: top }
  userProfileArrowDirection.value = arrowDirection
  showUserProfileCard.value = true
}

function showMessageUserProfile(event, message) {
  clearProfileCardTimer()
  
  const rect = event.target.getBoundingClientRect()
  currentProfileUser.value = {
    id: message.senderId || message.senderName,
    name: message.senderName,
    avatar: message.senderAvatar,
    description: '这个人很懒，什么都没留下...',
    isOnline: Math.random() > 0.5,
    feedCount: Math.floor(Math.random() * 100) + 20,
    friendCount: Math.floor(Math.random() * 1000) + 100,
    likeCount: Math.floor(Math.random() * 5000) + 500
  }
  
  // 计算位置
  const cardWidth = 320
  const cardHeight = 200
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  
  let left = rect.right + 10
  let top = rect.top + rect.height / 2 - cardHeight / 2
  let arrowDirection = 'left'
  
  // 边界检测
  if (left + cardWidth > viewportWidth - 10) {
    left = rect.left - cardWidth - 10
    arrowDirection = 'right'
  }
  
  if (top < 10) {
    top = 10
  } else if (top + cardHeight > viewportHeight - 10) {
    top = viewportHeight - cardHeight - 10
  }
  
  userProfilePosition.value = { x: left + (arrowDirection === 'left' ? 0 : cardWidth), y: top + cardHeight / 2 }
  userProfileArrowDirection.value = arrowDirection
  showUserProfileCard.value = true
}

function hideUserProfile() {
  profileCardTimer = setTimeout(() => {
    showUserProfileCard.value = false
    currentProfileUser.value = null
  }, 300)
}

function hideUserProfileDelayed() {
  profileCardTimer = setTimeout(() => {
    showUserProfileCard.value = false
    currentProfileUser.value = null
  }, 300)
}

function clearProfileCardTimer() {
  if (profileCardTimer) {
    clearTimeout(profileCardTimer)
    profileCardTimer = null
  }
}

function closeUserProfileCard() {
  showUserProfileCard.value = false
  currentProfileUser.value = null
  clearProfileCardTimer()
}

function handleViewProfile(user) {
  console.log('查看用户资料:', user.name)
  closeUserProfileCard()
}

function handleViewSpace(user) {
  console.log('查看用户空间:', user.name)
  closeUserProfileCard()
}

// 图片预览功能
function openImagePreview(imageUrl, allImages = []) {
  if (allImages.length === 0) {
    allImages = [imageUrl]
  }
  
  previewImages.value = allImages.map(img => ({
    src: img,
    name: `图片_${Date.now()}.jpg`
  }))
  
  currentImageIndex.value = allImages.findIndex(img => img === imageUrl)
  if (currentImageIndex.value === -1) {
    currentImageIndex.value = 0
  }
  
  showImagePreview.value = true
}

function closeImagePreview() {
  showImagePreview.value = false
  previewImages.value = []
  currentImageIndex.value = 0
}

function onImageChange(index) {
  currentImageIndex.value = index
}

// 文件预览功能
function openFilePreview(file) {
  currentPreviewFile.value = file
  showFilePreview.value = true
}

function closeFilePreview() {
  showFilePreview.value = false
  currentPreviewFile.value = null
}

// 辅助函数
function parseFileSize(sizeStr) {
  if (!sizeStr) return 0
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const match = sizeStr.match(/^(\d+(?:\.\d+)?)\s*([A-Z]+)$/i)
  
  if (!match) return 0
  
  const value = parseFloat(match[1])
  const unit = match[2].toUpperCase()
  const index = units.indexOf(unit)
  
  if (index === -1) return 0
  
  return value * Math.pow(1024, index)
}

function getAllImagesInConversation() {
  return displayMessages.value
    .filter(msg => msg.type === 'image')
    .map(msg => msg.content)
}

// 文件类型检测和图标获取
function getFileIcon(message) {
  if (!message || !message.fileName) return 'icon-file'
  
  const extension = (message.fileExtension || getFileExtension(message.fileName)).toLowerCase()
  const iconMap = {
    // 文档类型
    'pdf': 'icon-file-pdf',
    'doc': 'icon-file-word',
    'docx': 'icon-file-word',
    'xls': 'icon-file-excel',
    'xlsx': 'icon-file-excel',
    'ppt': 'icon-file-powerpoint',
    'pptx': 'icon-file-powerpoint',
    'odt': 'icon-file-text',
    'ods': 'icon-file-excel',
    'odp': 'icon-file-powerpoint',
    'rtf': 'icon-file-text',
    
    // 代码类型
    'js': 'icon-file-code',
    'ts': 'icon-file-code',
    'vue': 'icon-file-code',
    'jsx': 'icon-file-code',
    'tsx': 'icon-file-code',
    'py': 'icon-file-code',
    'java': 'icon-file-code',
    'cpp': 'icon-file-code',
    'c': 'icon-file-code',
    'h': 'icon-file-code',
    'php': 'icon-file-code',
    'rb': 'icon-file-code',
    'go': 'icon-file-code',
    'rs': 'icon-file-code',
    'swift': 'icon-file-code',
    'kt': 'icon-file-code',
    'html': 'icon-file-code',
    'css': 'icon-file-code',
    'scss': 'icon-file-code',
    'less': 'icon-file-code',
    
    // 文本类型
    'txt': 'icon-file-text',
    'md': 'icon-file-text',
    'markdown': 'icon-file-text',
    'json': 'icon-file-code',
    'xml': 'icon-file-code',
    'yaml': 'icon-file-code',
    'yml': 'icon-file-code',
    'log': 'icon-file-text',
    'conf': 'icon-file-text',
    'config': 'icon-file-text',
    'ini': 'icon-file-text',
    'csv': 'icon-file-text',
    'sql': 'icon-file-code',
    
    // 多媒体类型
    'mp4': 'icon-file-video',
    'avi': 'icon-file-video',
    'mov': 'icon-file-video',
    'wmv': 'icon-file-video',
    'flv': 'icon-file-video',
    'webm': 'icon-file-video',
    'mkv': 'icon-file-video',
    'm4v': 'icon-file-video',
    '3gp': 'icon-file-video',
    'mpg': 'icon-file-video',
    'mpeg': 'icon-file-video',
    
    'mp3': 'icon-file-audio',
    'wav': 'icon-file-audio',
    'ogg': 'icon-file-audio',
    'aac': 'icon-file-audio',
    'flac': 'icon-file-audio',
    'm4a': 'icon-file-audio',
    'wma': 'icon-file-audio',
    'opus': 'icon-file-audio',
    
    'jpg': 'icon-image',
    'jpeg': 'icon-image',
    'png': 'icon-image',
    'gif': 'icon-image',
    'bmp': 'icon-image',
    'svg': 'icon-image',
    'webp': 'icon-image',
    'ico': 'icon-image',
    'tiff': 'icon-image',
    'tif': 'icon-image',
    
    // 压缩类型
    'zip': 'icon-file-archive',
    'rar': 'icon-file-archive',
    '7z': 'icon-file-archive',
    'tar': 'icon-file-archive',
    'gz': 'icon-file-archive',
    'bz2': 'icon-file-archive',
    'xz': 'icon-file-archive',
    'cab': 'icon-file-archive',
    'iso': 'icon-file-archive',
    
    // 可执行文件
    'exe': 'icon-zap',
    'msi': 'icon-zap',
    'dmg': 'icon-zap',
    'app': 'icon-zap',
    'deb': 'icon-zap',
    'rpm': 'icon-zap',
    'pkg': 'icon-zap',
    'appx': 'icon-zap',
    
    // 设计文件
    'psd': 'icon-image',
    'ai': 'icon-image',
    'sketch': 'icon-image',
    'fig': 'icon-image',
    'xd': 'icon-image',
    'indd': 'icon-image',
    'eps': 'icon-image'
  }
  
  return iconMap[extension] || 'icon-file'
}

function getFileIconClass(message) {
  if (!message || !message.fileName) return ''
  
  const extension = (message.fileExtension || getFileExtension(message.fileName)).toLowerCase()
  
  // 根据文件类型返回不同的样式类
  if (['pdf'].includes(extension)) return 'file-icon-pdf'
  if (['doc', 'docx', 'odt', 'rtf'].includes(extension)) return 'file-icon-word'
  if (['xls', 'xlsx', 'ods', 'csv'].includes(extension)) return 'file-icon-excel'
  if (['ppt', 'pptx', 'odp'].includes(extension)) return 'file-icon-powerpoint'
  if (['js', 'ts', 'vue', 'jsx', 'tsx', 'py', 'java', 'cpp', 'c', 'h', 'php', 'rb', 'go', 'rs', 'swift', 'kt', 'html', 'css', 'scss', 'less', 'json', 'xml', 'yaml', 'yml', 'sql'].includes(extension)) return 'file-icon-code'
  if (['txt', 'md', 'markdown', 'log', 'conf', 'config', 'ini'].includes(extension)) return 'file-icon-text'
  if (['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv', 'm4v', '3gp', 'mpg', 'mpeg'].includes(extension)) return 'file-icon-video'
  if (['mp3', 'wav', 'ogg', 'aac', 'flac', 'm4a', 'wma', 'opus'].includes(extension)) return 'file-icon-audio'
  if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp', 'ico', 'tiff', 'tif'].includes(extension)) return 'file-icon-image'
  if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz', 'cab', 'iso'].includes(extension)) return 'file-icon-archive'
  if (['exe', 'msi', 'dmg', 'app', 'deb', 'rpm', 'pkg', 'appx'].includes(extension)) return 'file-icon-executable'
  if (['psd', 'ai', 'sketch', 'fig', 'xd', 'indd', 'eps'].includes(extension)) return 'file-icon-design'
  
  return 'file-icon-default'
}

function getFileTypeLabel(message) {
  if (!message || !message.fileName) return '未知文件'
  
  const extension = (message.fileExtension || getFileExtension(message.fileName)).toLowerCase()
  const typeMap = {
    'pdf': 'PDF文档',
    'doc': 'Word文档',
    'docx': 'Word文档',
    'xls': 'Excel表格',
    'xlsx': 'Excel表格',
    'ppt': 'PowerPoint',
    'pptx': 'PowerPoint',
    'odt': 'OpenDocument',
    'ods': 'OpenDocument',
    'odp': 'OpenDocument',
    'rtf': 'RTF文档',
    'txt': '文本文件',
    'md': 'Markdown',
    'markdown': 'Markdown',
    'json': 'JSON数据',
    'js': 'JavaScript',
    'ts': 'TypeScript',
    'vue': 'Vue组件',
    'jsx': 'React JSX',
    'tsx': 'React TSX',
    'py': 'Python脚本',
    'java': 'Java代码',
    'cpp': 'C++代码',
    'c': 'C代码',
    'h': 'C/C++头文件',
    'php': 'PHP脚本',
    'rb': 'Ruby脚本',
    'go': 'Go代码',
    'rs': 'Rust代码',
    'swift': 'Swift代码',
    'kt': 'Kotlin代码',
    'html': 'HTML网页',
    'css': 'CSS样式',
    'scss': 'Sass样式',
    'less': 'Less样式',
    'xml': 'XML文档',
    'yaml': 'YAML配置',
    'yml': 'YAML配置',
    'log': '日志文件',
    'conf': '配置文件',
    'config': '配置文件',
    'ini': '配置文件',
    'csv': 'CSV数据',
    'sql': 'SQL脚本',
    'mp4': 'MP4视频',
    'avi': 'AVI视频',
    'mov': 'MOV视频',
    'wmv': 'WMV视频',
    'flv': 'FLV视频',
    'webm': 'WebM视频',
    'mkv': 'MKV视频',
    'm4v': 'M4V视频',
    '3gp': '3GP视频',
    'mpg': 'MPG视频',
    'mpeg': 'MPEG视频',
    'mp3': 'MP3音频',
    'wav': 'WAV音频',
    'ogg': 'OGG音频',
    'aac': 'AAC音频',
    'flac': 'FLAC音频',
    'm4a': 'M4A音频',
    'wma': 'WMA音频',
    'opus': 'Opus音频',
    'jpg': 'JPEG图片',
    'jpeg': 'JPEG图片',
    'png': 'PNG图片',
    'gif': 'GIF图片',
    'bmp': 'BMP图片',
    'svg': 'SVG图片',
    'webp': 'WebP图片',
    'ico': 'ICO图标',
    'tiff': 'TIFF图片',
    'tif': 'TIFF图片',
    'zip': 'ZIP压缩包',
    'rar': 'RAR压缩包',
    '7z': '7Z压缩包',
    'tar': 'TAR压缩包',
    'gz': 'GZ压缩包',
    'bz2': 'BZ2压缩包',
    'xz': 'XZ压缩包',
    'cab': 'CAB压缩包',
    'iso': 'ISO镜像',
    'exe': 'Windows程序',
    'msi': 'Windows安装包',
    'dmg': 'macOS安装包',
    'app': 'macOS应用',
    'deb': 'Debian安装包',
    'rpm': 'RPM安装包',
    'pkg': '安装包',
    'appx': 'Windows应用包',
    'psd': 'Photoshop文档',
    'ai': 'Illustrator文档',
    'sketch': 'Sketch设计',
    'fig': 'Figma设计',
    'xd': 'Adobe XD设计',
    'indd': 'InDesign文档',
    'eps': 'EPS矢量图'
  }
  
  return typeMap[extension] || `${extension.toUpperCase()}文件`
}
function shouldShowTimeDivider(message, index) {
  if (index === 0) return true
  
  const currentTime = new Date(message.timestamp)
  const previousTime = new Date(displayMessages.value[index - 1].timestamp)
  
  // 如果时间间隔超过30分钟则显示时间分隔线
  return currentTime - previousTime > 30 * 60 * 1000
}

function formatMessageDate(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return '今天'
  } else if (diffDays === 1) {
    return '昨天'
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    return date.toLocaleDateString()
  }
}

function formatMessageTime(timestamp) {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function getStatusIcon(status) {
  const iconMap = {
    sending: 'icon-clock',
    sent: 'icon-check',
    delivered: 'icon-check-double',
    read: 'icon-check-double-read'
  }
  return iconMap[status] || 'icon-clock'
}

function formatFileSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 生命周期
onMounted(() => {
  scrollToBottom(false)
  
  // 监听点击事件，关闭表情选择器
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.emoji-picker') && !e.target.closest('.tool-btn')) {
      showEmojiPicker.value = false
    }
  })
})

onUnmounted(() => {
  if (recordingTimer) {
    clearInterval(recordingTimer)
  }
  if (profileCardTimer) {
    clearTimeout(profileCardTimer)
  }
})

// 监听会话变化
watch(() => props.conversation, () => {
  if (props.conversation) {
    nextTick(() => scrollToBottom(false))
  }
}, { immediate: true })
</script>

<style scoped>
.chat-panel {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  position: relative;
}

/* 聊天头部 */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: rgba(255, 255, 255, 0.95);
  border-bottom: 1px solid rgba(229, 231, 235, 0.8);
  backdrop-filter: blur(20px);
  position: relative;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.avatar:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(74, 140, 255, 0.2);
  cursor: pointer;
}

.message-avatar {
  cursor: pointer;
  transition: all 0.3s ease;
}

.message-avatar:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.3);
}

.user-info {
  flex: 1;
}

.name {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 4px 0;
  color: #1e293b;
  background: linear-gradient(135deg, #1e293b 0%, #4a8cff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.status {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 500;
  transition: color 0.3s ease;
}

.status.online {
  color: #10b981;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 12px;
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.action-btn:hover {
  background: rgba(74, 140, 255, 0.1);
  color: #4a8cff;
  transform: scale(1.05);
}

.close-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

/* 消息区域 */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  position: relative;
  
  /* 自定义滚动条 */
  scrollbar-width: thin;
  scrollbar-color: rgba(74, 140, 255, 0.3) transparent;
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(74, 140, 255, 0.3);
  border-radius: 3px;
}

.chat-messages::-webkit-scrollbar-thumb:hover {
  background: rgba(74, 140, 255, 0.5);
}

.loading-history {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 16px;
  color: #6b7280;
  font-size: 14px;
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #e5e7eb;
  border-top: 2px solid #4a8cff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.time-divider {
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
  margin: 16px 0;
  position: relative;
}

.time-divider::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(to right, transparent, #e5e7eb, transparent);
  z-index: 1;
}

.time-divider span {
  background: #f8fafc;
  padding: 0 12px;
  position: relative;
  z-index: 2;
}

.message-wrapper {
  display: flex;
  flex-direction: column;
}

.message-item {
  display: flex;
  gap: 12px;
  max-width: 70%;
}

.own-message .message-item {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.other-message .message-item {
  align-self: flex-start;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  align-self: flex-end;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
}

.sender-name {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.message-body {
  border-radius: 18px;
  padding: 12px 16px;
  position: relative;
  word-wrap: break-word;
  max-width: 100%;
}

.own-message .message-body {
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  color: white;
  border-bottom-right-radius: 6px;
}

.other-message .message-body {
  background: rgba(255, 255, 255, 0.9);
  color: #1e293b;
  border: 1px solid rgba(229, 231, 235, 0.5);
  border-bottom-left-radius: 6px;
}

.group-message .message-body {
  border-bottom-left-radius: 6px;
}

/* 不同类型的消息样式 */
.text-message {
  line-height: 1.5;
  font-size: 15px;
}

.image-message img {
  max-width: 200px;
  max-height: 200px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.image-message img:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.voice-message {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 160px;
}

.voice-play-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.2);
  color: inherit;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.voice-play-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}

.voice-wave {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 2px;
  height: 20px;
}

.wave-bar {
  width: 3px;
  background: currentColor;
  border-radius: 2px;
  opacity: 0.6;
  animation: waveAnimation 1.5s ease-in-out infinite;
}

.wave-bar:nth-child(odd) {
  animation-delay: 0.1s;
}

.wave-bar:nth-child(even) {
  animation-delay: 0.3s;
}

@keyframes waveAnimation {
  0%, 100% { height: 4px; }
  50% { height: 16px; }
}

.voice-duration {
  font-size: 12px;
  opacity: 0.8;
}

.file-message {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 200px;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 12px;
  padding: 8px;
  margin: -8px;
}

.file-message:hover {
  background: rgba(74, 140, 255, 0.1);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.2);
}

.file-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  transition: all 0.3s ease;
}

/* 不同文件类型的图标颜色 */
.file-icon-pdf {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.file-icon-word {
  background: rgba(37, 99, 235, 0.1);
  color: #2563eb;
}

.file-icon-excel {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.file-icon-powerpoint {
  background: rgba(249, 115, 22, 0.1);
  color: #ea580c;
}

.file-icon-code {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
}

.file-icon-text {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}

.file-icon-video {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.file-icon-audio {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.file-icon-image {
  background: rgba(168, 85, 247, 0.1);
  color: #a855f7;
}

.file-icon-archive {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.file-icon-executable {
  background: rgba(249, 115, 22, 0.1);
  color: #ea580c;
}

.file-icon-design {
  background: rgba(236, 72, 153, 0.1);
  color: #ec4899;
}

.file-icon-default {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}

.file-info {
  flex: 1;
}

.file-name {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 2px;
}

.file-size {
  font-size: 12px;
  opacity: 0.8;
}

.file-type {
  font-size: 11px;
  color: #6b7280;
  font-weight: 500;
  margin-top: 2px;
}

.file-download {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: rgba(74, 140, 255, 0.1);
  color: #4a8cff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  position: relative;
}

.file-download:hover {
  background: rgba(74, 140, 255, 0.2);
  transform: scale(1.05);
}

.file-download:after {
  content: "预览";
  position: absolute;
  top: -30px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;
}

.file-download:hover:after {
  opacity: 1;
}

.emoji-message .emoji-large {
  font-size: 48px;
  line-height: 1;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

.own-message .message-meta {
  justify-content: flex-end;
}

.message-status.sent {
  color: #10b981;
}

.message-status.delivered {
  color: #3b82f6;
}

.message-status.read {
  color: #8b5cf6;
}

/* 滚动到底部按钮 */
.scroll-to-bottom {
  position: absolute;
  bottom: 20px;
  right: 24px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  color: white;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  z-index: 10;
}

.scroll-to-bottom:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(74, 140, 255, 0.4);
}

.scroll-to-bottom .unread-count {
  position: absolute;
  top: -4px;
  right: -4px;
  background: #ef4444;
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
}

/* 滚动按钮动画 */
.scroll-btn-enter-active,
.scroll-btn-leave-active {
  transition: all 0.3s ease;
}

.scroll-btn-enter-from,
.scroll-btn-leave-to {
  opacity: 0;
  transform: scale(0.8) translateY(10px);
}

/* 输入区域 */
.chat-input-area {
  background: rgba(255, 255, 255, 0.95);
  border-top: 1px solid rgba(229, 231, 235, 0.8);
  backdrop-filter: blur(20px);
  padding: 16px 24px;
  position: relative;
}

.input-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.tool-btn {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  position: relative;
}

.tool-btn:hover {
  background: rgba(74, 140, 255, 0.1);
  color: #4a8cff;
  transform: scale(1.05);
}

.voice-btn.recording {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  animation: recordingPulse 1s ease-in-out infinite;
}

@keyframes recordingPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.input-container {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  background: rgba(248, 250, 252, 0.8);
  border: 1px solid rgba(229, 231, 235, 0.6);
  border-radius: 20px;
  padding: 8px 12px;
  transition: all 0.3s ease;
}

.input-container:focus-within {
  border-color: #4a8cff;
  box-shadow: 0 0 0 3px rgba(74, 140, 255, 0.1);
  background: rgba(255, 255, 255, 0.9);
}

.message-input {
  flex: 1;
  min-height: 20px;
  max-height: 120px;
  padding: 8px 12px;
  border: none;
  outline: none;
  font-size: 15px;
  line-height: 1.5;
  color: #1e293b;
  background: transparent;
  overflow-y: auto;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.message-input:empty::before {
  content: attr(data-placeholder);
  color: #9ca3af;
  pointer-events: none;
}

.send-btn {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: linear-gradient(135deg, #e5e7eb 0%, #d1d5db 100%);
  color: #9ca3af;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}

.send-btn.has-content {
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(74, 140, 255, 0.3);
}

.send-btn.has-content:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.4);
}

.send-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* 录音状态 */
.recording-status {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.1) 0%, rgba(220, 38, 38, 0.05) 100%);
  border-radius: 12px;
  margin-top: 12px;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.recording-wave {
  display: flex;
  gap: 2px;
  height: 16px;
  align-items: center;
}

.recording-wave .wave {
  width: 3px;
  background: #ef4444;
  border-radius: 2px;
  animation: recordingWave 1s ease-in-out infinite;
}

.recording-wave .wave:nth-child(2) { animation-delay: 0.1s; }
.recording-wave .wave:nth-child(3) { animation-delay: 0.2s; }
.recording-wave .wave:nth-child(4) { animation-delay: 0.3s; }
.recording-wave .wave:nth-child(5) { animation-delay: 0.4s; }

@keyframes recordingWave {
  0%, 100% { height: 4px; }
  50% { height: 16px; }
}

.recording-text {
  flex: 1;
  color: #ef4444;
  font-size: 14px;
  font-weight: 500;
}

.recording-time {
  color: #ef4444;
  font-size: 12px;
  font-weight: 600;
}

/* 录音状态动画 */
.recording-enter-active,
.recording-leave-active {
  transition: all 0.3s ease;
}

.recording-enter-from,
.recording-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 表情选择器 */
.emoji-picker {
  position: absolute;
  bottom: 100%;
  left: 24px;
  right: 24px;
  max-height: 200px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(229, 231, 235, 0.8);
  border-radius: 16px;
  backdrop-filter: blur(20px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  z-index: 20;
}

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 4px;
  padding: 16px;
  max-height: 160px;
  overflow-y: auto;
}

.emoji-item {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 8px;
  background: transparent;
  font-size: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.emoji-item:hover {
  background: rgba(74, 140, 255, 0.1);
  transform: scale(1.2);
}

/* 表情选择器动画 */
.emoji-picker-enter-active,
.emoji-picker-leave-active {
  transition: all 0.3s ease;
}

.emoji-picker-enter-from,
.emoji-picker-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chat-header {
    padding: 16px 20px;
  }
  
  .chat-messages {
    padding: 16px 20px;
  }
  
  .chat-input-area {
    padding: 12px 20px;
  }
  
  .message-item {
    max-width: 85%;
  }
  
  .avatar {
    width: 40px;
    height: 40px;
  }
  
  .name {
    font-size: 16px;
  }
  
  .emoji-picker {
    left: 20px;
    right: 20px;
  }
}
</style>