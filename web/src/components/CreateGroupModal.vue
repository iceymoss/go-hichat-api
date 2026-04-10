<template>
  <div class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>创建群聊</h3>
        <button class="close-btn" @click="close">
          <i class="icon icon-x"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="avatar-upload">
          <div class="avatar-preview" @click="triggerFileInput">
            <img v-if="groupIcon" :src="groupIcon" />
            <i v-else class="icon icon-camera"></i>
            <div class="upload-hint" v-if="!loadingAvatar">点击上传</div>
            <div class="upload-loading" v-else>上传中...</div>
          </div>
          <input 
            type="file" 
            ref="fileInput" 
            accept="image/*" 
            style="display: none" 
            @change="handleFileSelect"
          />
        </div>

        <div class="input-group">
          <label>群名称</label>
          <input 
            v-model="groupName" 
            type="text" 
            placeholder="请输入群名称" 
            @keyup.enter="handleCreate"
            ref="inputRef"
          />
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="btn-cancel" @click="close">取消</button>
        <button class="btn-create" @click="handleCreate" :disabled="!groupName.trim() || loading || loadingAvatar">
          {{ loading ? '创建中...' : '创建' }}
        </button>
      </div>
    </div>

    <!-- 裁切弹窗 -->
    <ImageCropper 
      v-if="showCropper && selectedFile" 
      :file="selectedFile" 
      @cancel="showCropper = false" 
      @confirm="handleCropConfirm" 
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useContactsStore } from '../stores/contacts'
import { userApi } from '../utils/api'
import ImageCropper from './ui/ImageCropper.vue'

const emit = defineEmits(['close', 'success'])
const contactsStore = useContactsStore()

const groupName = ref('')
const groupIcon = ref('')
const loading = ref(false)
const loadingAvatar = ref(false)
const inputRef = ref(null)
const fileInput = ref(null)

const showCropper = ref(false)
const selectedFile = ref(null)

onMounted(() => {
  inputRef.value?.focus()
})

const close = () => {
  if (showCropper.value) return // 如果正在裁切，不关闭主弹窗
  emit('close')
}

const triggerFileInput = () => {
  if (loadingAvatar.value) return
  fileInput.value?.click()
}

const handleFileSelect = (e) => {
  const file = e.target.files?.[0]
  if (file) {
    selectedFile.value = file
    showCropper.value = true
  }
  e.target.value = '' // reset
}

const handleCropConfirm = async (blob) => {
  showCropper.value = false
  loadingAvatar.value = true
  
  try {
    // Create a File object from Blob
    const file = new File([blob], "avatar.jpg", { type: "image/jpeg" })
    const res = await userApi.uploadAvatar(file)
    if (res && res.url) {
      groupIcon.value = res.url
    }
  } catch (e) {
    console.error(e)
    alert('头像上传失败')
  } finally {
    loadingAvatar.value = false
  }
}

const handleCreate = async () => {
  if (!groupName.value.trim() || loading.value) return
  
  loading.value = true
  try {
    await contactsStore.createGroup(groupName.value.trim(), groupIcon.value.trim())
    emit('success')
    close()
  } catch (error) {
    // Error is logged in store
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: var(--overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  background: var(--card);
  width: 360px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  overflow: hidden;
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-family: var(--font-display);
  color: var(--fg);
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--muted);
  padding: 4px;
  transition: color 0.2s;
}

.close-btn:hover {
  color: var(--fg);
}

.modal-body {
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.avatar-upload {
  margin-bottom: 24px;
  position: relative;
}

.avatar-preview {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--input-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: hidden;
  border: 2px dashed var(--border);
  transition: border-color 0.2s;
}

.avatar-preview:hover {
  border-color: var(--accent);
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-preview .icon {
  font-size: 24px;
  color: var(--muted);
}

.upload-hint {
  position: absolute;
  bottom: -20px;
  width: 100%;
  text-align: center;
  font-size: 12px;
  color: var(--muted);
}

.upload-loading {
  position: absolute;
  bottom: -20px;
  width: 100%;
  text-align: center;
  font-size: 12px;
  color: var(--accent);
}

.input-group {
  width: 100%;
  margin-bottom: 0;
}

.input-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--muted);
}

.input-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-family: var(--font-ui);
  background: var(--input-bg);
  color: var(--fg);
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.input-group input::placeholder {
  color: var(--muted);
}

.input-group input:focus {
  border-color: var(--accent);
}

.modal-footer {
  padding: 16px 20px;
  background: var(--bg);
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  padding: 8px 16px;
  border: 1px solid var(--border);
  background: var(--card);
  border-radius: var(--radius-sm);
  color: var(--muted);
  cursor: pointer;
  font-size: 13px;
  font-family: var(--font-ui);
  transition: color 0.2s;
}

.btn-cancel:hover {
  color: var(--fg);
}

.btn-create {
  padding: 8px 20px;
  border: none;
  background: var(--accent);
  border-radius: var(--radius-sm);
  color: var(--bg);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  font-family: var(--font-ui);
  transition: background-color 0.2s;
}

.btn-create:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-create:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

