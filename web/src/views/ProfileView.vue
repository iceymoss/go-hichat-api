<template>
  <div class="profile-view">
    <!-- 背景图 -->
    <div class="profile-banner">
      <img src="https://picsum.photos/800/200?random=1" alt="背景" class="banner-image">
    </div>
    
    <!-- 个人资料卡 -->
    <div class="profile-card">
      <div class="profile-header">
        <img :src="currentUser.avatar" alt="头像" class="avatar">
        <div class="profile-info">
          <h2 class="name">{{ currentUser.name }}</h2>
          <div class="account">ID: {{ currentUser.account }}</div>
          <div class="signature" v-if="!editing">{{ currentUser.signature || '这个人很懒，什么都没写' }}</div>
          
          <!-- 编辑状态签名 -->
          <div v-if="editing" class="edit-signature">
            <textarea 
              v-model="editSignature" 
              placeholder="编辑个性签名" 
              rows="2"
            ></textarea>
          </div>
        </div>
      </div>
      
      <!-- 资料区域 -->
      <div class="profile-details">
        <div class="section">
          <h3>基础资料</h3>
          <div class="detail-item">
            <span class="label">性别</span>
            <span v-if="!editing" class="value">{{ currentUser.gender === 'male' ? '男' : '女' }}</span>
            <select v-else v-model="editGender" class="value-input">
              <option value="male">男</option>
              <option value="female">女</option>
            </select>
          </div>
          <div class="detail-item">
            <span class="label">地区</span>
            <span v-if="!editing" class="value">{{ currentUser.region || '未设置' }}</span>
            <input v-else v-model="editRegion" type="text" class="value-input" placeholder="输入地区">
          </div>
          <div class="detail-item">
            <span class="label">职业</span>
            <span v-if="!editing" class="value">{{ currentUser.occupation || '未设置' }}</span>
            <input v-else v-model="editOccupation" type="text" class="value-input" placeholder="输入职业">
          </div>
        </div>
        
        <div class="section">
          <h3>联系信息</h3>
          <div class="detail-item">
            <span class="label">邮箱</span>
            <span v-if="!editing" class="value">{{ currentUser.email || '未设置' }}</span>
            <input v-else v-model="editEmail" type="email" class="value-input" placeholder="输入邮箱">
          </div>
          <div class="detail-item">
            <span class="label">手机</span>
            <span v-if="!editing" class="value">{{ currentUser.phone || '未设置' }}</span>
            <input v-else v-model="editPhone" type="tel" class="value-input" placeholder="输入手机号">
          </div>
        </div>
        
        <div class="section tags-section">
          <h3>个人标签</h3>
          <div v-if="!editing" class="tags">
            <span v-for="tag in currentUser.tags" :key="tag" class="tag">{{ tag }}</span>
            <span v-if="currentUser.tags.length === 0">暂无标签</span>
          </div>
          <div v-else class="edit-tags">
            <div class="tag-input-container">
              <input 
                v-model="newTag" 
                type="text" 
                placeholder="添加标签" 
                class="tag-input"
                @keyup.enter="addTag"
              >
              <button class="btn-add-tag" @click="addTag">+</button>
            </div>
            <div class="selected-tags">
              <span v-for="(tag, index) in editTags" :key="index" class="tag">
                {{ tag }}
                <i class="icon icon-close" @click="removeTag(index)"></i>
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 操作按钮 -->
      <div class="profile-actions">
        <template v-if="isMyProfile">
          <button v-if="!editing" class="btn-edit" @click="startEdit">编辑资料</button>
          <div v-else class="edit-actions">
            <button class="btn-cancel" @click="cancelEdit">取消</button>
            <button class="btn-save" @click="saveProfile">保存</button>
          </div>
        </template>
        <template v-else>
          <button 
            class="btn-chat" 
            :class="{ 'btn-gray': currentUser.isFriend }"
            @click="startChat"
          >
            <i class="icon icon-chat"></i>
            {{ currentUser.isFriend ? '发消息' : '添加好友' }}
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const editing = ref(false)
const currentUser = ref({
  id: 1000,
  name: '演示用户',
  account: 'user_demo',
  avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=demo',
  signature: '前端小白，后端大神',
  gender: 'male',
  region: '北京市',
  occupation: 'Go工程师',
  email: 'demo@example.com',
  phone: '138****1234',
  tags: ['Go开发', 'Vue3学习', '后端架构'],
  isFriend: true
})

// 编辑字段
const editName = ref('')
const editSignature = ref('')
const editGender = ref('')
const editRegion = ref('')
const editOccupation = ref('')
const editEmail = ref('')
const editPhone = ref('')
const editTags = ref([])
const newTag = ref('')

const isMyProfile = computed(() => {
  return route.params.userId === authStore.user?.id.toString()
})

onMounted(() => {
  // 初始化编辑字段
  editName.value = currentUser.value.name
  editSignature.value = currentUser.value.signature
  editGender.value = currentUser.value.gender
  editRegion.value = currentUser.value.region
  editOccupation.value = currentUser.value.occupation
  editEmail.value = currentUser.value.email
  editPhone.value = currentUser.value.phone
  editTags.value = [...currentUser.value.tags]
})

const startEdit = () => {
  editing.value = true
}

const cancelEdit = () => {
  editing.value = false
  resetEditFields()
}

const resetEditFields = () => {
  editName.value = currentUser.value.name
  editSignature.value = currentUser.value.signature
  editGender.value = currentUser.value.gender
  editRegion.value = currentUser.value.region
  editOccupation.value = currentUser.value.occupation
  editEmail.value = currentUser.value.email
  editPhone.value = currentUser.value.phone
  editTags.value = [...currentUser.value.tags]
}

const saveProfile = () => {
  // 模拟保存
  currentUser.value = {
    ...currentUser.value,
    name: editName.value,
    signature: editSignature.value,
    gender: editGender.value,
    region: editRegion.value,
    occupation: editOccupation.value,
    email: editEmail.value,
    phone: editPhone.value,
    tags: [...editTags.value]
  }
  
  editing.value = false
}

const addTag = () => {
  if (newTag.value.trim() && !editTags.value.includes(newTag.value.trim())) {
    editTags.value.push(newTag.value.trim())
    newTag.value = ''
  }
}

const removeTag = (index) => {
  editTags.value.splice(index, 1)
}

const startChat = () => {
  // 跳转到聊天
}
</script>

<style scoped>
.profile-view {
  height: 100%;
  overflow-y: auto;
  position: relative;
}

.profile-banner {
  height: 180px;
  overflow: hidden;
}

.banner-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-card {
  position: relative;
  margin: -60px 20px 20px;
  background-color: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 5px 20px rgba(0, 0, 0, 0.1);
}

.profile-header {
  display: flex;
  margin-bottom: 25px;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 12px;
  border: 4px solid white;
  margin-top: -50px;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.1);
  object-fit: cover;
}

.profile-info {
  flex: 1;
  padding: 0 20px;
}

.name {
  margin: 5px 0 5px;
}

.account {
  color: #999;
  font-size: 14px;
  margin-bottom: 10px;
}

.signature {
  color: #666;
  font-size: 14px;
  line-height: 1.5;
  background-color: #f9f9f9;
  padding: 10px;
  border-radius: 8px;
}

.edit-signature textarea {
  width: 100%;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 10px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
}

.profile-details {
  border-top: 1px solid #f0f2f5;
  padding-top: 20px;
}

.section {
  margin-bottom: 25px;
}

.section h3 {
  padding-bottom: 10px;
  margin-bottom: 15px;
  border-bottom: 1px solid #f0f2f5;
  font-size: 16px;
}

.detail-item {
  display: flex;
  margin-bottom: 15px;
  align-items: flex-start;
}

.label {
  flex: 0 0 80px;
  font-size: 14px;
  color: #999;
}

.value {
  font-size: 14px;
  color: #333;
  word-break: break-all;
}

.value-input {
  flex: 1;
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 14px;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  display: inline-flex;
  align-items: center;
  padding: 5px 12px;
  background-color: #eef5ff;
  color: #4a8cff;
  border-radius: 16px;
  font-size: 13px;
}

.edit-tags {
  padding: 5px 0;
}

.tag-input-container {
  display: flex;
  margin-bottom: 15px;
}

.tag-input {
  flex: 1;
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 14px;
  border-right: none;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.btn-add-tag {
  padding: 0 15px;
  background-color: #4a8cff;
  color: white;
  border: none;
  border-radius: 0 6px 6px 0;
  cursor: pointer;
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.selected-tags .tag {
  position: relative;
  padding-right: 28px;
}

.selected-tags .icon-close {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 12px;
  cursor: pointer;
  opacity: 0.7;
}

.selected-tags .icon-close:hover {
  opacity: 1;
}

.profile-actions {
  display: flex;
  justify-content: center;
  padding-top: 15px;
  border-top: 1px solid #f0f2f5;
}

.btn-edit, .btn-chat {
  padding: 10px 30px;
  background: linear-gradient(to right, #4a8cff, #8a69ff);
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-chat.btn-gray {
  background: #f0f2f5;
  color: #666;
}

.edit-actions {
  display: flex;
  gap: 15px;
}

.btn-cancel, .btn-save {
  padding: 10px 25px;
  border-radius: 6px;
  border: none;
  font-weight: 500;
  cursor: pointer;
}

.btn-cancel {
  background: #f5f5f5;
  color: #666;
}

.btn-save {
  background: linear-gradient(to right, #4a8cff, #8a69ff);
  color: white;
}
</style>