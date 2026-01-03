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
  background-color: white;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.post-header {
  display: flex;
  align-items: center;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  margin-right: 12px;
  object-fit: cover;
}

.post-input {
  flex: 1;
  padding: 12px 18px;
  background-color: #f0f2f5;
  border-radius: 24px;
  border: none;
  font-size: 15px;
  cursor: pointer;
}

.post-editor {
  margin-top: 15px;
}

.post-textarea {
  width: 100%;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px;
  font-family: inherit;
  font-size: 15px;
  resize: none;
  margin-bottom: 15px;
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
  border-radius: 8px;
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
  border-radius: 50%;
  background-color: rgba(0, 0, 0, 0.6);
  color: white;
  border: none;
  font-size: 16px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
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
  border-radius: 50%;
  color: #666;
  cursor: pointer;
  position: relative;
}

.upload-btn:hover {
  background-color: #f0f2f5;
}

.file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
}

.post-btn {
  padding: 8px 20px;
  background: linear-gradient(to right, #4a8cff, #8a69ff);
  color: white;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.post-btn:disabled {
  background: #f0f2f5;
  color: #999;
  cursor: not-allowed;
}
</style>