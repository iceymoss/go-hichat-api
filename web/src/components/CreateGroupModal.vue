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
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: #fff;
  width: 360px;
  border-radius: 16px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 { margin: 0; font-size: 16px; color: #1e293b; }

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 4px;
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
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: hidden;
  border: 2px solid #e2e8f0;
  transition: all 0.2s;
}

.avatar-preview:hover { border-color: #4a8cff; }

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-preview .icon {
  font-size: 24px;
  color: #94a3b8;
}

.upload-hint {
  position: absolute;
  bottom: -20px;
  width: 100%;
  text-align: center;
  font-size: 12px;
  color: #64748b;
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
  color: #475569;
}

.input-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.input-group input:focus { border-color: #4a8cff; }

.modal-footer {
  padding: 16px 20px;
  background: #f8fafc;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  padding: 8px 16px;
  border: 1px solid #cbd5e1;
  background: #fff;
  border-radius: 6px;
  color: #475569;
  cursor: pointer;
  font-size: 13px;
}

.btn-create {
  padding: 8px 20px;
  border: none;
  background: #4a8cff;
  border-radius: 6px;
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}

.btn-create:disabled { background: #94a3b8; cursor: not-allowed; }
</style>

