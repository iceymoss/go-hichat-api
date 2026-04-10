<template>
  <div v-if="visible" class="create-post-modal-overlay" @click.self="closeModal">
    <div class="create-post-modal">
      <!-- 头部 -->
      <div class="modal-header">
        <div class="header-left">
          <div class="header-icon">
            <i class="icon icon-edit"></i>
          </div>
          <div class="header-text">
            <h2 class="header-title">发布动态</h2>
            <p class="header-subtitle">分享你的想法和生活瞬间</p>
          </div>
        </div>
        <div class="header-actions">
          <button class="close-button" @click="closeModal" title="关闭">
            <i class="icon icon-x"></i>
          </button>
        </div>
      </div>

      <!-- 内容区 -->
      <div class="modal-content">
        <!-- 用户信息 -->
        <div class="user-info">
          <img :src="currentUser.avatar" alt="头像" class="user-avatar">
          <div class="user-details">
            <div class="user-name">{{ currentUser.name }}</div>
            <div class="post-type">
              <i class="icon icon-globe"></i>
              <span>公开</span>
            </div>
          </div>
        </div>

        <!-- 文本输入区 -->
        <div class="text-area">
          <textarea
            ref="textareaRef"
            v-model="postContent"
            placeholder="分享你的想法..."
            class="post-textarea"
            :style="{ minHeight: minTextareaHeight + 'px' }"
            @input="adjustTextareaHeight"
          ></textarea>
        </div>

        <!-- 图片预览区 -->
        <div v-if="images.length > 0" class="image-preview-area">
          <div class="preview-grid" :class="getGridClass()">
            <div
              v-for="(image, index) in images"
              :key="index"
              class="preview-item"
              @click="viewImage(index)"
            >
              <img :src="image.url" alt="预览" class="preview-image">
              <button class="remove-image-btn" @click.stop="removeImage(index)">
                <i class="icon icon-x"></i>
              </button>
            </div>
          </div>
        </div>

        <!-- 工具栏 -->
        <div class="toolbar">
          <div class="toolbar-left">
            <label class="tool-button" title="添加图片">
              <i class="icon icon-image"></i>
              <span>图片</span>
              <input
                type="file"
                accept="image/*"
                multiple
                class="file-input"
                @change="handleImageUpload"
              >
            </label>
            <button class="tool-button" title="添加位置" @click="addLocation">
              <i class="icon icon-map-pin"></i>
              <span>位置</span>
            </button>
            <button class="tool-button" title="添加话题" @click="addTopic">
              <i class="icon icon-hash"></i>
              <span>话题</span>
            </button>
          </div>
          <div class="toolbar-right">
            <div class="char-count" :class="{ warning: charCount > 500 }">
              {{ charCount }}/500
            </div>
          </div>
        </div>
      </div>

      <!-- 底部操作区 -->
      <div class="modal-footer">
        <button class="cancel-btn" @click="closeModal">
          取消
        </button>
        <button
          class="publish-btn"
          :disabled="!canPublish"
          @click="publishPost"
        >
          <i class="icon icon-send"></i>
          <span>发布</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFeedStore } from '../stores/feed'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'published'])

const authStore = useAuthStore()
const feedStore = useFeedStore()

const currentUser = computed(() => authStore.user || { 
  avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=You',
  name: '我'
})

// 表单数据
const postContent = ref('')
const images = ref([])
const textareaRef = ref(null)
const minTextareaHeight = 120

// 字符计数
const charCount = computed(() => postContent.value.length)

// 是否可以发布
const canPublish = computed(() => {
  return (postContent.value.trim().length > 0 || images.value.length > 0) && charCount.value <= 500
})

// 图片网格类
const getGridClass = () => {
  const count = images.value.length
  if (count === 1) return 'grid-single'
  if (count === 2) return 'grid-double'
  if (count === 3) return 'grid-triple'
  if (count === 4) return 'grid-quad'
  return 'grid-multi'
}

// 自动调整文本框高度
const adjustTextareaHeight = () => {
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.style.height = 'auto'
      const newHeight = Math.max(minTextareaHeight, textareaRef.value.scrollHeight)
      textareaRef.value.style.height = newHeight + 'px'
    }
  })
}

// 处理图片上传
const handleImageUpload = (event) => {
  const files = Array.from(event.target.files)
  
  // 限制最多9张图片
  if (images.value.length + files.length > 9) {
    alert('最多只能上传9张图片')
    return
  }
  
  files.forEach(file => {
    // 检查文件大小 (限制10MB)
    if (file.size > 10 * 1024 * 1024) {
      alert('图片大小不能超过10MB')
      return
    }
    
    const reader = new FileReader()
    reader.onload = (e) => {
      images.value.push({
        file,
        url: e.target.result,
        name: file.name
      })
    }
    reader.readAsDataURL(file)
  })
  
  // 清空input
  event.target.value = ''
}

// 移除图片
const removeImage = (index) => {
  images.value.splice(index, 1)
}

// 查看图片
const viewImage = (index) => {
  // TODO: 打开图片查看器
  console.log('查看图片:', index)
}

// 添加位置
const addLocation = () => {
  // TODO: 打开位置选择器
  console.log('添加位置')
}

// 添加话题
const addTopic = () => {
  const topic = prompt('请输入话题名称:')
  if (topic) {
    postContent.value += ` #${topic} `
    adjustTextareaHeight()
  }
}

// 发布动态
const publishPost = async () => {
  if (!canPublish.value) return
  
  try {
    // 处理图片 (这里简化处理，实际应用中应该上传到服务器)
    const imageUrls = images.value.map(img => img.url)
    
    // 调用store的createPost方法
    const newPost = feedStore.createPost(postContent.value, imageUrls, 'friend')
    
    // 重置表单
    resetForm()
    
    // 发出事件
    emit('published', newPost)
    emit('close')
    
    // 提示成功
    console.log('动态发布成功！')
  } catch (error) {
    console.error('发布失败:', error)
    alert('发布失败，请重试')
  }
}

// 重置表单
const resetForm = () => {
  postContent.value = ''
  images.value = []
  if (textareaRef.value) {
    textareaRef.value.style.height = minTextareaHeight + 'px'
  }
}

// 关闭弹窗
const closeModal = () => {
  emit('close')
}

// 监听弹窗显示状态
watch(() => props.visible, (visible) => {
  if (visible) {
    nextTick(() => {
      // 聚焦到文本框
      if (textareaRef.value) {
        textareaRef.value.focus()
      }
    })
  }
})
</script>

<style scoped>
.create-post-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--overlay);
  backdrop-filter: blur(8px);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.create-post-modal {
  background: var(--card);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  border: 1px solid var(--border);
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
  font-family: var(--font-ui);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 28px;
  background: var(--input-bg);
  border-bottom: 1px solid var(--border);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: var(--accent);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--bg);
  font-size: 20px;
}

.header-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.header-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: var(--accent);
  font-family: var(--font-display);
}

.header-subtitle {
  font-size: 13px;
  color: var(--muted);
  margin: 0;
  font-weight: 500;
}

.close-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  background: var(--card);
  color: var(--muted);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 18px;
}

.close-button:hover {
  background: var(--accent-muted);
  color: var(--accent);
  transform: scale(1.1);
}

.modal-content {
  flex: 1;
  padding: 24px 28px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  border: 2px solid var(--border);
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--fg);
}

.post-type {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--muted);
  font-weight: 500;
}

.post-type .icon {
  font-size: 14px;
}

.text-area {
  position: relative;
}

.post-textarea {
  width: 100%;
  border: none;
  outline: none;
  font-size: 16px;
  line-height: 1.6;
  color: var(--fg);
  background: transparent;
  resize: none;
  font-family: var(--font-ui);
  min-height: 120px;
}

.post-textarea::placeholder {
  color: var(--muted);
}

.image-preview-area {
  margin-top: 8px;
}

.preview-grid {
  display: grid;
  gap: 8px;
  border-radius: var(--radius-md);
  overflow: hidden;
}

.preview-grid.grid-single {
  grid-template-columns: 1fr;
}

.preview-grid.grid-double {
  grid-template-columns: 1fr 1fr;
}

.preview-grid.grid-triple {
  grid-template-columns: 2fr 1fr;
  grid-template-rows: 1fr 1fr;
}

.preview-grid.grid-triple .preview-item:first-child {
  grid-row: 1 / 3;
}

.preview-grid.grid-quad {
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
}

.preview-grid.grid-multi {
  grid-template-columns: repeat(3, 1fr);
}

.preview-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.preview-item:hover {
  transform: scale(1.02);
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-image-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background: var(--overlay);
  color: var(--fg);
  border: none;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.remove-image-btn:hover {
  background: var(--danger);
  transform: scale(1.1);
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--input-bg);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.tool-button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--muted);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  font-family: var(--font-ui);
}

.tool-button:hover {
  background: var(--accent-muted);
  color: var(--accent);
}

.tool-button .icon {
  font-size: 18px;
}

.file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
}

.char-count {
  font-size: 14px;
  color: var(--muted);
  font-weight: 500;
}

.char-count.warning {
  color: var(--danger);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 28px;
  background: var(--input-bg);
  border-top: 1px solid var(--border);
}

.cancel-btn {
  padding: 10px 20px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--muted);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: var(--font-ui);
}

.cancel-btn:hover {
  background: var(--accent-muted);
  color: var(--fg);
}

.publish-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: var(--bg);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: var(--font-ui);
}

.publish-btn:hover:not(:disabled) {
  background: var(--accent-hover);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(200,149,108,0.3);
}

.publish-btn:disabled {
  background: var(--border);
  color: var(--muted);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.publish-btn .icon {
  font-size: 16px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .create-post-modal-overlay {
    padding: 8px;
  }

  .create-post-modal {
    max-width: 100%;
    max-height: 95vh;
    border-radius: var(--radius-lg);
  }

  .modal-header {
    padding: 20px 20px;
  }

  .modal-content {
    padding: 20px 20px;
  }

  .modal-footer {
    padding: 16px 20px;
  }

  .toolbar {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .toolbar-left {
    justify-content: center;
  }

  .toolbar-right {
    text-align: center;
  }
}
</style> 