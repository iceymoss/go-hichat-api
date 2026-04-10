<template>
  <div class="create-post-card">
    <div class="post-header">
      <img :src="user.avatar" alt="头像" class="avatar">
      <input 
        type="text" 
        v-model="postContent" 
        placeholder="分享新鲜事..." 
        class="post-input"
        @click="openEditor"
      >
    </div>
    
    <div v-if="showEditor" class="post-editor">
      <textarea 
        v-model="postContent" 
        placeholder="分享你的想法..."
        rows="3"
        class="post-textarea"
      ></textarea>
      
      <div v-if="images.length > 0" class="preview-images">
        <div v-for="(image, index) in images" :key="index" class="preview-item">
          <img :src="image" alt="预览" class="preview-image">
          <button class="btn-remove" @click="removeImage(index)">&times;</button>
        </div>
      </div>
      
      <div class="post-actions">
        <div class="action-buttons">
          <label class="upload-btn">
            <i class="icon icon-image"></i>
            <input 
              type="file" 
              accept="image/*" 
              multiple 
              class="file-input" 
              @change="handleImageUpload"
            >
          </label>
        </div>
        <button class="post-btn" @click="publishPost" :disabled="!canPost">发布</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const emit = defineEmits(['post-created'])

const user = computed(() => authStore.user || { 
  avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=You'
})

const postContent = ref('')
const images = ref([])
const showEditor = ref(false)

const canPost = computed(() => {
  return postContent.value.trim() || images.value.length > 0
})

const openEditor = () => {
  showEditor.value = true
}

const handleImageUpload = (e) => {
  const files = Array.from(e.target.files)
  
  // 限制最多9张图片
  if (images.value.length + files.length > 9) {
    alert('最多只能上传9张图片')
    return
  }
  
  files.forEach(file => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      images.value.push(e.target.result)
    }
    
    reader.readAsDataURL(file)
  })
}

const removeImage = (index) => {
  images.value.splice(index, 1)
}

const publishPost = () => {
  if (!canPost.value) return
  
  emit('post-created', {
    content: postContent.value,
    images: images.value
  })
  
  // 重置表单
  postContent.value = ''
  images.value = []
  showEditor.value = false
}
</script>

<style scoped>
.create-post-card {
  background-color: var(--card);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid var(--border);
  font-family: var(--font-ui);
}

.post-header {
  display: flex;
  align-items: center;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  margin-right: 12px;
  object-fit: cover;
  border: 2px solid var(--border);
}

.post-input {
  flex: 1;
  padding: 12px 18px;
  background-color: var(--input-bg);
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  font-size: 15px;
  cursor: pointer;
  color: var(--fg);
  font-family: var(--font-ui);
}

.post-input::placeholder {
  color: var(--muted);
}

.post-input:focus {
  outline: none;
  border-color: var(--accent);
}

.post-editor {
  margin-top: 15px;
}

.post-textarea {
  width: 100%;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px;
  font-family: var(--font-ui);
  font-size: 15px;
  resize: none;
  margin-bottom: 15px;
  background-color: var(--input-bg);
  color: var(--fg);
}

.post-textarea::placeholder {
  color: var(--muted);
}

.post-textarea:focus {
  outline: none;
  border-color: var(--accent);
}

.preview-images {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 15px;
}

.preview-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.btn-remove {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background-color: var(--overlay);
  color: var(--fg);
  border: none;
  font-size: 16px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-remove:hover {
  background-color: var(--danger);
}

.post-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-buttons {
  display: flex;
  gap: 10px;
}

.upload-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  color: var(--muted);
  cursor: pointer;
  position: relative;
  transition: all 0.2s;
}

.upload-btn:hover {
  background-color: var(--accent-muted);
  color: var(--accent);
}

.file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
}

.post-btn {
  padding: 8px 20px;
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-full);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  font-family: var(--font-ui);
  transition: background-color 0.2s;
}

.post-btn:hover {
  background: var(--accent-hover);
}

.post-btn:disabled {
  background: var(--border);
  color: var(--muted);
  cursor: not-allowed;
}
</style>