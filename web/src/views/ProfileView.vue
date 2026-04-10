<template>
  <div class="profile-view">
    <!-- 加载状态 -->
    <div v-if="loadingUserInfo && !isMyProfile" class="loading-container">
      <div class="loading-spinner-large"></div>
      <p>加载用户信息中...</p>
    </div>
    
    <template v-else>
      <!-- 背景图 -->
      <div class="profile-banner">
        <img src="https://picsum.photos/800/200?random=1" alt="背景" class="banner-image">
      </div>
      
      <!-- 个人资料卡 -->
      <div class="profile-card">
      <div class="profile-header">
        <div class="avatar-container">
          <img :src="currentUser.avatar" alt="头像" class="avatar">
          <div v-if="editing" class="avatar-overlay" @click="triggerAvatarUpload">
            <i class="icon icon-camera"></i>
            <span>更换头像</span>
          </div>
          <input 
            v-if="editing"
            ref="avatarInput"
            type="file" 
            accept="image/*" 
            style="display: none"
            @change="handleAvatarChange"
          >
        </div>
        <div class="profile-info">
          <div v-if="!editing">
            <h2 class="name">{{ currentUser.name }}</h2>
          </div>
          <div v-else class="edit-name">
            <input 
              v-model="editName" 
              type="text" 
              class="name-input" 
              placeholder="输入昵称"
              maxlength="20"
            >
          </div>
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
            <span class="label">手机号</span>
            <span class="value">{{ currentUser.phone || '未设置' }}</span>
            <span class="value-note">（手机号不可修改）</span>
          </div>
          <div class="detail-item">
            <span class="label">邮箱</span>
            <div class="email-display">
              <span class="value">{{ currentUser.email || '未绑定' }}</span>
              <button 
                v-if="isMyProfile"
                type="button"
                class="btn-edit-email"
                @click="openEmailModal"
              >
                {{ currentUser.email ? '修改' : '绑定' }}
              </button>
            </div>
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
          <div class="profile-actions">
            <button v-if="!editing" class="btn-edit" @click="startEdit">编辑资料</button>
            <div v-else class="edit-actions">
              <button class="btn-cancel" @click="cancelEdit">取消</button>
              <button class="btn-save" @click="saveProfile">保存</button>
            </div>
            <button v-if="!editing" class="btn-logout" @click="handleLogout">
              <i class="icon icon-logout"></i>
              退出登录
            </button>
          </div>
        </template>
        <template v-else>
          <div class="profile-actions-buttons">
            <button 
              v-if="currentUser.isFriend"
              class="btn-chat" 
              @click="startChat"
            >
              <i class="icon icon-chat"></i>
              发消息
            </button>
            <button 
              v-else-if="!friendRequestSent"
              class="btn-add-friend" 
              @click="showAddFriendDialog = true"
              :disabled="sendingFriendRequest"
            >
              <i v-if="sendingFriendRequest" class="icon icon-spinner"></i>
              <i v-else class="icon icon-add"></i>
              {{ sendingFriendRequest ? '发送中...' : '添加好友' }}
            </button>
            <button 
              v-else
              class="btn-request-sent" 
              disabled
            >
              <i class="icon icon-check"></i>
              已发送申请
            </button>
          </div>
        </template>
      </div>
    </div>
    </template>

  <!-- 头像裁切弹窗 -->
  <ImageCropper 
    v-if="showCropper && selectedFile" 
    :file="selectedFile" 
    @cancel="showCropper = false" 
    @confirm="handleCropConfirm" 
  />

  <!-- 邮箱绑定模态框 -->
  <div v-if="showEmailModal" class="modal-overlay" @click.self="closeEmailModal">
    <div class="modal-container email-modal">
      <div class="modal-header">
        <h3>{{ currentUser.email ? '修改邮箱' : '绑定邮箱' }}</h3>
        <button type="button" class="modal-close" @click="closeEmailModal">×</button>
      </div>
      
      <div class="modal-body">
        <!-- 当前邮箱显示 -->
        <div v-if="currentUser.email" class="current-email-info">
          <span class="current-email-label">当前邮箱：</span>
          <span class="current-email-value">{{ currentUser.email }}</span>
        </div>
        
        <!-- 邮箱绑定表单 -->
        <div class="email-bind-form">
          <div class="form-step">
            <div class="step-number">1</div>
            <div class="step-content">
              <label class="step-label">输入新邮箱地址</label>
              <div class="email-input-wrapper">
                <input 
                  v-model="modalEmail" 
                  type="email" 
                  class="email-input" 
                  placeholder="请输入邮箱地址，如：example@email.com"
                  :disabled="emailBinding || emailCodeCountdown > 0"
                  :class="{ 'input-error': emailCodeError && !isValidModalEmail }"
                />
                <button 
                  type="button"
                  class="btn-send-code"
                  @click="sendEmailCode"
                  :disabled="!isValidModalEmail || emailCodeSending || emailCodeCountdown > 0 || emailBinding || modalEmail === currentUser.email"
                >
                  <span v-if="emailCodeSending" class="loading-spinner"></span>
                  <span v-if="emailCodeCountdown > 0">{{ emailCodeCountdown }}秒后重试</span>
                  <span v-else-if="emailCodeSending">发送中...</span>
                  <span v-else>发送验证码</span>
                </button>
              </div>
              <p v-if="emailCodeError && !isValidModalEmail" class="error-message">{{ emailCodeError }}</p>
            </div>
          </div>

          <div v-if="modalEmail && modalEmail !== currentUser.email && emailCodeSuccess" class="form-step">
            <div class="step-number">2</div>
            <div class="step-content">
              <label class="step-label">输入验证码</label>
              <div class="code-input-wrapper">
                <input 
                  v-model="modalEmailCode" 
                  type="text" 
                  class="code-input" 
                  placeholder="请输入6位验证码"
                  maxlength="6"
                  pattern="[0-9]{6}"
                  :disabled="emailBinding"
                  :class="{ 'input-error': emailCodeError && modalEmailCode }"
                  @input="modalEmailCode = modalEmailCode.replace(/\D/g, '')"
                />
                <div class="code-hint">
                  <span v-if="modalEmailCode.length > 0 && modalEmailCode.length < 6" class="hint-text">
                    还需输入 {{ 6 - modalEmailCode.length }} 位
                  </span>
                  <span v-else-if="modalEmailCode.length === 6" class="hint-success">✓ 验证码格式正确</span>
                </div>
              </div>
              <p v-if="emailCodeError && modalEmailCode" class="error-message">{{ emailCodeError }}</p>
              <p v-if="emailCodeSuccess && !emailCodeError" class="success-message">
                ✓ {{ emailCodeSuccess }}
              </p>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button type="button" class="btn-cancel-modal" @click="closeEmailModal">取消</button>
        <button 
          type="button"
          class="btn-confirm-bind"
          @click="bindEmail"
          :disabled="emailBinding || !modalEmail || !modalEmailCode || modalEmailCode.length !== 6 || modalEmail === currentUser.email"
        >
          <span v-if="emailBinding" class="loading-spinner"></span>
          <span v-else>确认{{ currentUser.email ? '修改' : '绑定' }}</span>
        </button>
      </div>
    </div>
  </div>

  <!-- 申请好友对话框 -->
  <div v-if="showAddFriendDialog" class="modal-overlay" @click.self="showAddFriendDialog = false">
    <div class="modal-container add-friend-modal">
      <div class="modal-header">
        <h3>发送好友申请</h3>
        <button type="button" class="modal-close" @click="showAddFriendDialog = false">×</button>
      </div>
      
      <div class="modal-body">
        <div class="user-preview">
          <img :src="currentUser.avatar" alt="头像" class="avatar" />
          <div class="user-name">{{ currentUser.name }}</div>
        </div>
        <div class="message-input">
          <label>申请消息（可选）</label>
          <textarea 
            v-model="requestMessage" 
            placeholder="请输入申请消息..."
            rows="3"
            maxlength="100"
          ></textarea>
          <div class="char-count">{{ requestMessage.length }}/100</div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button type="button" class="btn-cancel-modal" @click="showAddFriendDialog = false">取消</button>
        <button 
          type="button"
          class="btn-confirm-bind"
          @click="sendFriendRequest"
          :disabled="sendingFriendRequest"
        >
          <span v-if="sendingFriendRequest" class="loading-spinner"></span>
          <span v-else>发送</span>
        </button>
      </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useContactsStore } from '../stores/contacts'
import { userApi, socialApi } from '../utils/api'
import ImageCropper from '../components/ui/ImageCropper.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const contactsStore = useContactsStore()

// 申请好友相关状态
const showAddFriendDialog = ref(false)
const requestMessage = ref('')
const sendingFriendRequest = ref(false)
const friendRequestSent = ref(false)
const loadingUserInfo = ref(false)

const editing = ref(false)
const currentUser = ref({
  id: '',
  name: '',
  account: '',
  avatar: '',
  signature: '',
  gender: 'male',
  region: '',
  occupation: '',
  email: '',
  phone: '',
  tags: [],
  isFriend: false
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
const avatarInput = ref(null)
const avatarUploading = ref(false)
const showCropper = ref(false)
const selectedFile = ref(null)

// 邮箱绑定相关状态
const showEmailModal = ref(false)
const modalEmail = ref('')
const modalEmailCode = ref('')
const emailCodeError = ref('')
const emailCodeSuccess = ref('')
const emailCodeCountdown = ref(0)
const emailCodeSending = ref(false)
const emailBinding = ref(false)

// 验证邮箱格式
const isValidModalEmail = computed(() => {
  if (!modalEmail.value) return false
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(modalEmail.value)
})

const isMyProfile = computed(() => {
  const userId = route.params.userId
  if (!userId || userId === 'current') {
    return true
  }
  return userId === authStore.user?.id?.toString()
})

onMounted(async () => {
  // 如果是查看自己的资料，使用 authStore 中的用户信息
  if (isMyProfile.value && authStore.user) {
    const user = authStore.user
    // 解析tags（如果是JSON字符串）
    let tags = []
    if (user.tags) {
      try {
        tags = typeof user.tags === 'string' ? JSON.parse(user.tags) : user.tags
      } catch (e) {
        tags = []
      }
    }
    
    currentUser.value = {
      id: user.id,
      name: user.nickname || user.name || '未设置',
      account: user.id,
      avatar: user.avatar || 'https://api.dicebear.com/7.x/personas/svg?seed=' + user.id,
      signature: user.introduction || '这个人很懒，什么都没写',
      gender: user.sex === 1 ? 'male' : user.sex === 2 ? 'female' : 'unknown',
      region: user.region || '',
      occupation: user.occupation || '',
      email: user.email || '',
      phone: user.phone || user.mobile || '', // 兼容mobile和phone字段
      tags: tags,
      isFriend: false
    }
  } else {
    // 查看其他用户资料，通过搜索 API 获取用户信息
    await loadOtherUserInfo()
  }
  
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

// 监听路由参数变化，重新加载用户信息
watch(() => route.params.userId, async (newUserId, oldUserId) => {
  if (newUserId && newUserId !== oldUserId && !isMyProfile.value) {
    await loadOtherUserInfo()
  }
}, { immediate: false })

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

// 打开邮箱绑定模态框
const openEmailModal = () => {
  showEmailModal.value = true
  modalEmail.value = currentUser.value.email || ''
  modalEmailCode.value = ''
  emailCodeError.value = ''
  emailCodeSuccess.value = ''
  emailCodeCountdown.value = 0
}

// 关闭邮箱绑定模态框
const closeEmailModal = () => {
  showEmailModal.value = false
  modalEmail.value = ''
  modalEmailCode.value = ''
  emailCodeError.value = ''
  emailCodeSuccess.value = ''
  emailCodeCountdown.value = 0
}

// 发送邮箱验证码
const sendEmailCode = async () => {
  if (!modalEmail.value) {
    emailCodeError.value = '请先输入邮箱地址'
    return
  }
  
  if (!isValidModalEmail.value) {
    emailCodeError.value = '请输入正确的邮箱地址'
    return
  }
  
  // 如果邮箱没有变化，不需要验证
  if (modalEmail.value === currentUser.value.email) {
    emailCodeError.value = ''
    emailCodeSuccess.value = '邮箱未变更，无需验证'
    return
  }
  
  emailCodeError.value = ''
  emailCodeSuccess.value = ''
  emailCodeSending.value = true
  
  try {
    await userApi.sendEmailCode(modalEmail.value)
    emailCodeSuccess.value = '验证码已发送，请查收邮箱（测试时查看控制台）'
    
    // 开始倒计时
    emailCodeCountdown.value = 60
    const timer = setInterval(() => {
      emailCodeCountdown.value--
      if (emailCodeCountdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (err) {
    emailCodeError.value = err.message || '发送验证码失败，请稍后重试'
  } finally {
    emailCodeSending.value = false
  }
}

// 触发头像上传
const triggerAvatarUpload = () => {
  if (avatarInput.value) {
    avatarInput.value.click()
  }
}

// 处理头像文件选择
const handleAvatarChange = (event) => {
  const file = event.target.files?.[0]
  if (!file) return
  
  // 验证文件类型
  if (!file.type.startsWith('image/')) {
    alert('请选择图片文件')
    return
  }
  
  // 验证文件大小（限制5MB）
  if (file.size > 5 * 1024 * 1024) {
    alert('图片大小不能超过5MB')
    return
  }

  // 打开裁切弹窗
  selectedFile.value = file
  showCropper.value = true
  event.target.value = '' // 重置 input，允许重复选择同一文件
}

const handleCropConfirm = async (blob) => {
  showCropper.value = false
  avatarUploading.value = true
  
  try {
    // 临时预览裁切后的图片
    const previewUrl = URL.createObjectURL(blob)
    currentUser.value.avatar = previewUrl

    // 创建 File 对象
    const file = new File([blob], "avatar.jpg", { type: "image/jpeg" })
    
    // 上传到服务器
    const response = await userApi.uploadAvatar(file)
    if (response && response.url) {
      // 使用服务器返回的URL
      currentUser.value.avatar = response.url
    }
  } catch (error) {
    console.error('上传头像失败:', error)
    alert('上传头像失败：' + (error.message || '请稍后重试'))
    // 恢复原头像
    if (authStore.user && authStore.user.avatar) {
      currentUser.value.avatar = authStore.user.avatar
    }
  } finally {
    avatarUploading.value = false
  }
}

// 绑定邮箱
const bindEmail = async () => {
  if (!modalEmail.value) {
    emailCodeError.value = '请先输入邮箱地址'
    return
  }
  
  if (!isValidModalEmail.value) {
    emailCodeError.value = '请输入正确的邮箱地址'
    return
  }
  
  if (!modalEmailCode.value || modalEmailCode.value.length !== 6) {
    emailCodeError.value = '请输入6位验证码'
    return
  }
  
  emailCodeError.value = ''
  emailCodeSuccess.value = ''
  emailBinding.value = true
  
  try {
    await userApi.bindEmail(modalEmail.value, modalEmailCode.value)
    
    // 刷新用户信息
    await authStore.fetchUserInfo()
    
    // 更新本地显示
    if (authStore.user) {
      currentUser.value.email = authStore.user.email || ''
    }
    
    // 关闭模态框
    closeEmailModal()
    
    // 显示成功提示
    alert('邮箱绑定成功！')
  } catch (err) {
    emailCodeError.value = err.message || '邮箱绑定失败，请稍后重试'
  } finally {
    emailBinding.value = false
  }
}

const saveProfile = async () => {
  try {
    // 验证昵称
    if (!editName.value || editName.value.trim() === '') {
      alert('昵称不能为空')
      return
    }
    
    // 将tags数组转换为JSON字符串
    const tagsJson = JSON.stringify(editTags.value)
    
    // 调用更新用户信息API（不再处理邮箱，邮箱绑定使用独立的 API）
    await authStore.updateUser({
      name: editName.value.trim(),
      avatar: currentUser.value.avatar, // 如果头像已更新，使用新头像
      introduction: editSignature.value,
      sex: editGender.value === 'male' ? 1 : editGender.value === 'female' ? 2 : 0,
      region: editRegion.value,
      occupation: editOccupation.value,
      tags: tagsJson
    })
    
    // 更新本地显示
    currentUser.value = {
      ...currentUser.value,
      name: editName.value.trim(),
      avatar: currentUser.value.avatar, // 保持当前头像（可能已更新）
      signature: editSignature.value,
      gender: editGender.value,
      region: editRegion.value,
      occupation: editOccupation.value,
      // 注意：邮箱和手机号不能通过此API更新
      // 邮箱使用独立的绑定API，手机号不可修改
      tags: [...editTags.value]
    }
    
    editing.value = false
  } catch (error) {
    console.error('保存用户信息失败:', error)
    alert('保存失败：' + error.message)
  }
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

// 加载其他用户信息
const loadOtherUserInfo = async () => {
  const userId = route.params.userId
  if (!userId) return
  
  loadingUserInfo.value = true
  
  try {
    // 使用 getUserById 方法通过用户ID获取用户信息
    const user = await userApi.getUserById(userId)
      
      if (!user) {
      // 用户不存在
        currentUser.value = {
          id: userId,
        name: '用户不存在',
          account: userId,
          avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${userId}`,
        signature: '用户不存在',
          gender: 'unknown',
          region: '',
          occupation: '',
          email: '',
          phone: '',
          tags: [],
          isFriend: false
        }
        return
      }
      
      // 解析标签
      let tags = []
      if (user.tags) {
        try {
          tags = typeof user.tags === 'string' ? JSON.parse(user.tags) : user.tags
        } catch (e) {
          tags = []
        }
      }
      
      currentUser.value = {
        id: user.id,
        name: user.nickname || user.name || '未设置昵称',
        account: user.id,
        avatar: user.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${user.id}`,
        signature: user.introduction || '这个人很懒，什么都没写',
        gender: user.sex === 1 ? 'male' : user.sex === 2 ? 'female' : 'unknown',
        region: user.region || '',
        occupation: user.occupation || '',
        email: user.email || '',
        phone: user.mobile || '',
        tags: tags,
        isFriend: false // 稍后检查好友关系
      }
      
      // 检查是否是好友
      await checkFriendStatus(userId)
      
      // 检查是否已发送好友申请
      await checkFriendRequestStatus(userId)
  } catch (error) {
    console.error('加载用户信息失败:', error)
    // 使用默认数据
    currentUser.value = {
      id: userId,
      name: error.message?.includes('不存在') ? '用户不存在' : '加载失败',
      account: userId,
      avatar: `https://api.dicebear.com/7.x/personas/svg?seed=${userId}`,
      signature: error.message?.includes('不存在') ? '用户不存在' : '加载用户信息失败',
      gender: 'unknown',
      region: '',
      occupation: '',
      email: '',
      phone: '',
      tags: [],
      isFriend: false
    }
  } finally {
    loadingUserInfo.value = false
  }
}

// 检查好友关系
const checkFriendStatus = async (userId) => {
  try {
    // 获取好友列表
    await contactsStore.fetchFriends()
    
    // 检查当前用户是否在好友列表中
    const isFriend = contactsStore.friends.some(
      friend => friend.friend_uid === userId || friend.id === userId || friend.friend_uid === currentUser.value.id || friend.id === currentUser.value.id
    )
    
    currentUser.value.isFriend = isFriend
  } catch (error) {
    console.error('检查好友关系失败:', error)
  }
}

// 检查好友申请状态
const checkFriendRequestStatus = async (userId) => {
  try {
    // 获取我发起的申请列表
    const sentRequests = await contactsStore.fetchFriendRequests(0, '0')
    
    // 检查是否已发送申请给该用户
    const hasSentRequest = sentRequests.some(
      req => req.req_uid === userId || req.user_id === userId
    )
    
    friendRequestSent.value = hasSentRequest
  } catch (error) {
    console.error('检查申请状态失败:', error)
  }
}

// 发送好友申请
const sendFriendRequest = async () => {
  if (!currentUser.value.id || sendingFriendRequest.value) return
  
  sendingFriendRequest.value = true
  
  try {
    await socialApi.friendPutIn(currentUser.value.id, requestMessage.value.trim())
    
    friendRequestSent.value = true
    showAddFriendDialog.value = false
    requestMessage.value = ''
    
    // 显示成功提示
    alert('好友申请已发送')
  } catch (error) {
    console.error('发送好友申请失败:', error)
    alert(error.message || '发送失败，请稍后重试')
  } finally {
    sendingFriendRequest.value = false
  }
}

const startChat = () => {
  // 跳转到聊天
  router.push(`/app/chat?userId=${currentUser.value.id}`)
}

const handleLogout = async () => {
  if (!confirm('确定要退出登录吗？')) {
    return
  }
  
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('登出失败:', error)
    // 即使后端登出失败，也清除本地数据并跳转
    router.push('/login')
  }
}
</script>

<style scoped>
.profile-view {
  height: 100%;
  overflow-y: auto;
  position: relative;
  background: var(--bg);
}

.profile-banner {
  height: 160px;
  overflow: hidden;
  background-color: var(--card);
  position: relative;
}

.profile-banner::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
}

.banner-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-card {
  position: relative;
  margin: -48px 20px 20px;
  background-color: var(--card);
  border-radius: var(--radius-lg);
  padding: 20px;
  border: 1px solid var(--border);
}

.profile-header {
  display: flex;
  margin-bottom: 20px;
}

.avatar-container {
  position: relative;
  margin-top: -44px;
}

.avatar {
  width: 88px;
  height: 88px;
  border-radius: var(--radius-md);
  border: 3px solid var(--card);
  object-fit: cover;
  display: block;
}

.avatar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.55);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.15s;
  color: var(--fg);
  font-size: 12px;
}

.avatar-container:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay i {
  font-size: 20px;
  margin-bottom: 4px;
}

.edit-name {
  margin: 4px 0;
}

.name-input {
  width: 100%;
  font-size: 20px;
  font-weight: 600;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 8px 12px;
  font-family: var(--font-ui);
  color: var(--fg);
  background: var(--input-bg);
}

.name-input:focus {
  outline: none;
  border-color: var(--accent);
}

.profile-info {
  flex: 1;
  padding: 0 16px;
}

.name {
  margin: 4px 0 4px;
  font-size: 20px;
  font-weight: 600;
  color: var(--fg);
  font-family: var(--font-display);
}

.account {
  color: var(--muted);
  font-size: 13px;
  margin-bottom: 8px;
  font-family: var(--font-ui);
}

.signature {
  color: var(--muted);
  font-size: 14px;
  line-height: 1.5;
  background-color: var(--input-bg);
  padding: 10px 12px;
  border-radius: var(--radius-md);
  font-family: var(--font-ui);
}

.edit-signature textarea {
  width: 100%;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 10px 12px;
  font-size: 14px;
  font-family: var(--font-ui);
  resize: none;
  color: var(--fg);
  background: var(--input-bg);
}

.edit-signature textarea:focus {
  outline: none;
  border-color: var(--accent);
}

.profile-details {
  border-top: 1px solid var(--border);
  padding-top: 16px;
}

.section {
  margin-bottom: 20px;
}

.section h3 {
  padding-bottom: 8px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border);
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-family: var(--font-display);
}

.detail-item {
  display: flex;
  margin-bottom: 12px;
  align-items: flex-start;
}

.label {
  flex: 0 0 72px;
  font-size: 13px;
  color: var(--muted);
}

.value {
  font-size: 14px;
  color: var(--fg);
  word-break: break-all;
}

.value-input {
  flex: 1;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 8px 12px;
  font-size: 14px;
  color: var(--fg);
  background: var(--input-bg);
}

.value-input:focus {
  outline: none;
  border-color: var(--accent);
}

.value-note {
  font-size: 12px;
  color: var(--muted);
  margin-left: 8px;
}

/* Email */
.email-item {
  flex-direction: column;
  align-items: flex-start;
}

.email-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bind-hint {
  font-size: 12px;
  color: var(--muted);
}

.email-bind-wrapper {
  width: 100%;
  margin-top: 8px;
}

.current-email-info {
  padding: 10px 12px;
  background-color: var(--input-bg);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.current-email-label {
  font-size: 13px;
  color: var(--muted);
}

.current-email-value {
  font-size: 14px;
  color: var(--fg);
  font-weight: 500;
}

.email-bind-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-step {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.step-number {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background-color: var(--accent);
  color: var(--bg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}

.step-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.step-label {
  font-size: 13px;
  color: var(--muted);
  font-weight: 500;
  margin-bottom: 2px;
}

.email-input-wrapper {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}

.email-input {
  flex: 1;
  padding: 9px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-size: 14px;
  transition: border-color 0.15s;
  background-color: var(--input-bg);
  color: var(--fg);
}

.email-input:focus {
  outline: none;
  border-color: var(--accent);
}

.email-input:disabled {
  background-color: var(--card);
  cursor: not-allowed;
  opacity: 0.6;
}

.email-input.input-error {
  border-color: var(--danger);
}

.btn-send-code {
  flex-shrink: 0;
  padding: 9px 16px;
  background-color: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 110px;
  justify-content: center;
}

.btn-send-code:hover:not(:disabled) {
  background-color: var(--accent-hover);
}

.btn-send-code:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.code-input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.code-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-size: 16px;
  text-align: center;
  letter-spacing: 6px;
  font-weight: 600;
  transition: border-color 0.15s;
  background-color: var(--input-bg);
  color: var(--fg);
}

.code-input:focus {
  outline: none;
  border-color: var(--accent);
}

.code-input:disabled {
  background-color: var(--card);
  cursor: not-allowed;
  opacity: 0.6;
}

.code-input.input-error {
  border-color: var(--danger);
}

.code-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.hint-text {
  color: var(--muted);
}

.hint-success {
  color: var(--success);
  font-weight: 500;
}

.error-message {
  color: var(--danger);
  font-size: 12px;
  margin-top: 4px;
}

.success-message {
  color: var(--success);
  font-size: 13px;
  margin-top: 4px;
  font-weight: 500;
}

.btn-bind-email {
  width: 100%;
  padding: 10px 20px;
  background-color: var(--success);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-bind-email:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-bind-email:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.loading-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Email edit button */
.email-display {
  display: flex;
  align-items: center;
  gap: 10px;
}

.btn-edit-email {
  padding: 3px 10px;
  background-color: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-edit-email:hover {
  background-color: var(--accent-hover);
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background: var(--card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  max-width: 460px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.email-modal {
  width: 100%;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  border-bottom: 1px solid var(--border);
}

.modal-header h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: var(--fg);
  font-family: var(--font-display);
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 20px;
  color: var(--muted);
  cursor: pointer;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.15s;
}

.modal-close:hover {
  background-color: var(--border);
  color: var(--fg);
}

.modal-body {
  padding: 20px 24px;
}

.modal-footer {
  display: flex;
  gap: 8px;
  padding: 16px 24px;
  border-top: 1px solid var(--border);
  justify-content: flex-end;
}

.btn-cancel-modal {
  padding: 8px 16px;
  background: var(--input-bg);
  color: var(--fg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-cancel-modal:hover {
  background-color: var(--card-hover);
}

.btn-confirm-bind {
  padding: 8px 20px;
  background-color: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-confirm-bind:hover:not(:disabled) {
  background-color: var(--accent-hover);
}

.btn-confirm-bind:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  background-color: var(--accent-muted);
  color: var(--accent);
  border-radius: var(--radius-sm);
  font-size: 13px;
}

.edit-tags {
  padding: 4px 0;
}

.tag-input-container {
  display: flex;
  margin-bottom: 12px;
}

.tag-input {
  flex: 1;
  border: 1px solid var(--border);
  border-radius: var(--radius-md) 0 0 var(--radius-md);
  padding: 8px 12px;
  font-size: 14px;
  border-right: none;
  color: var(--fg);
  background: var(--input-bg);
}

.tag-input:focus {
  outline: none;
  border-color: var(--accent);
}

.btn-add-tag {
  padding: 0 14px;
  background-color: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.15s;
}

.btn-add-tag:hover {
  background-color: var(--accent-hover);
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.selected-tags .tag {
  position: relative;
  padding-right: 26px;
}

.selected-tags .icon-close {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 12px;
  cursor: pointer;
  opacity: 0.5;
}

.selected-tags .icon-close:hover {
  opacity: 1;
}

.profile-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: center;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}

.btn-edit, .btn-chat {
  padding: 9px 24px;
  background-color: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: background-color 0.15s;
}

.btn-edit:hover, .btn-chat:hover {
  background-color: var(--accent-hover);
}

.btn-chat.btn-gray {
  background-color: var(--input-bg);
  color: var(--muted);
}

.edit-actions {
  display: flex;
  gap: 10px;
}

.btn-cancel, .btn-save {
  padding: 9px 20px;
  border-radius: var(--radius-md);
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-cancel {
  background-color: var(--input-bg);
  color: var(--fg);
  border: 1px solid var(--border);
}

.btn-cancel:hover {
  background-color: var(--card-hover);
}

.btn-save {
  background-color: var(--accent);
  color: var(--bg);
}

.btn-save:hover {
  background-color: var(--accent-hover);
}

.profile-actions-buttons {
  display: flex;
  gap: 10px;
  width: 100%;
  justify-content: center;
}

.btn-add-friend,
.btn-request-sent {
  padding: 9px 24px;
  background-color: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: background-color 0.15s;
}

.btn-add-friend:hover:not(:disabled) {
  background-color: var(--accent-hover);
}

.btn-add-friend:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-request-sent {
  background-color: var(--input-bg);
  color: var(--muted);
  cursor: default;
}

.add-friend-modal {
  max-width: 380px;
}

.add-friend-modal .user-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding: 14px;
  background: var(--input-bg);
  border-radius: var(--radius-md);
}

.add-friend-modal .user-preview .avatar {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-md);
}

.add-friend-modal .user-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--fg);
}

.add-friend-modal .message-input {
  margin-top: 12px;
}

.add-friend-modal .message-input label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: var(--muted);
  font-weight: 500;
}

.add-friend-modal .message-input textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: var(--font-ui);
  resize: vertical;
  outline: none;
  transition: border-color 0.15s;
  background: var(--input-bg);
  color: var(--fg);
}

.add-friend-modal .message-input textarea:focus {
  border-color: var(--accent);
}

.add-friend-modal .char-count {
  text-align: right;
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
}

.btn-logout {
  padding: 9px 24px;
  background: transparent;
  color: var(--danger);
  border: 1px solid var(--danger);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  justify-content: center;
}

.btn-logout:hover {
  background-color: rgba(192, 57, 43, 0.1);
}

.btn-logout i {
  font-size: 14px;
}

/* Loading */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 12px;
  color: var(--muted);
}

.loading-spinner-large {
  width: 36px;
  height: 36px;
  border: 3px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-container p {
  margin: 0;
  font-size: 14px;
}
</style>