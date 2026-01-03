<template>
  <div class="auth-form">
    <h2>{{ title }}</h2>
    
    <form @submit.prevent="submitForm">
      <!-- 手机号输入 -->
      <div class="form-group">
        <label for="phone">手机号</label>
        <input 
          type="tel" 
          id="phone" 
          v-model="formData.phone" 
          placeholder="请输入手机号" 
          required
          pattern="[0-9]{11}"
          maxlength="11"
        />
      </div>
      
      <!-- 密码输入 -->
      <div class="form-group">
        <label for="password">密码</label>
        <input 
          type="password" 
          id="password" 
          v-model="formData.password" 
          placeholder="请输入密码" 
          required
          minlength="6"
        />
      </div>
      
      <!-- 注册页面显示手机验证码 -->
      <template v-if="mode === 'register'">
        <div class="form-group">
          <div class="phone-code-container">
            <label for="phoneCode">手机验证码 <span class="required">*</span></label>
            <div class="phone-code-input-group">
              <input 
                type="text" 
                id="phoneCode" 
                v-model="formData.phoneCode" 
                placeholder="请输入6位验证码" 
                required
                maxlength="6"
                pattern="[0-9]{6}"
                :disabled="!formData.phone || phoneCodeSending"
              />
              <button 
                type="button"
                class="send-code-btn"
                @click="sendPhoneCode"
                :disabled="!formData.phone || phoneCodeSending || countdown > 0"
              >
                <span v-if="countdown > 0">{{ countdown }}s</span>
                <span v-else>{{ phoneCodeSending ? '发送中...' : '发送验证码' }}</span>
              </button>
            </div>
            <p v-if="phoneCodeError" class="error-text">{{ phoneCodeError }}</p>
            <p v-if="phoneCodeSuccess" class="success-text">{{ phoneCodeSuccess }}</p>
          </div>
        </div>
      </template>
      
      <!-- 注册页面才显示额外字段 -->
      <template v-if="mode === 'register'">
        <div class="form-group">
          <label for="nickname">昵称</label>
          <input 
            type="text" 
            id="nickname" 
            v-model="formData.nickname" 
            placeholder="设置您的昵称" 
            required
          />
        </div>
      </template>
      
      <!-- 提交按钮 -->
      <button 
        type="submit" 
        class="submit-btn"
        :disabled="loading"
      >
        <span v-if="loading" class="spinner"></span>
        {{ mode === 'login' ? '登录' : '注册' }}
      </button>
      
      <!-- 错误提示 -->
      <div v-if="error" class="error-message">
        {{ error }}
      </div>
    </form>
    
    <!-- 切换链接和重置密码 -->
    <div class="switch-auth">
      <div class="switch-row">
        <span>{{ mode === 'login' ? '还没有账号？' : '已有账号？' }}</span>
        <router-link 
          :to="mode === 'login' ? '/register' : '/login'"
          class="switch-link"
        >
          {{ mode === 'login' ? '立即注册' : '去登录' }}
        </router-link>
      </div>
      <!-- 登录模式下显示重置密码链接 -->
      <div v-if="mode === 'login'" class="reset-password-row">
        <a href="#" class="reset-password-link" @click.prevent="showResetPasswordModal = true">
          忘记密码？
        </a>
      </div>
    </div>
    
    <!-- 重置密码模态框 -->
    <div v-if="showResetPasswordModal" class="modal-overlay" @click.self="closeResetPasswordModal">
      <div class="modal-container reset-password-modal">
        <div class="modal-header">
          <h3>重置密码</h3>
          <button type="button" class="modal-close" @click="closeResetPasswordModal">×</button>
        </div>
        
        <div class="modal-body">
          <!-- 重置方式选择 -->
          <div class="form-group">
            <label>重置方式 <span class="required">*</span></label>
            <div class="reset-method-selector">
              <label class="method-option">
                <input 
                  type="radio" 
                  name="resetMethod" 
                  value="phone"
                  v-model="resetPasswordData.method"
                  :disabled="resetPasswordLoading"
                />
                <span>手机号</span>
              </label>
              <label class="method-option">
                <input 
                  type="radio" 
                  name="resetMethod" 
                  value="email"
                  v-model="resetPasswordData.method"
                  :disabled="resetPasswordLoading"
                />
                <span>邮箱</span>
              </label>
            </div>
          </div>
          
          <!-- 手机号输入（当选择手机号时显示） -->
          <div v-if="resetPasswordData.method === 'phone'" class="form-group">
            <label for="resetPhone">手机号 <span class="required">*</span></label>
            <input 
              type="tel" 
              id="resetPhone"
              v-model="resetPasswordData.phone" 
              placeholder="请输入注册时使用的手机号"
              maxlength="11"
              :disabled="resetPasswordLoading"
              @input="resetPasswordData.phone = resetPasswordData.phone.replace(/\D/g, '')"
            />
          </div>
          
          <!-- 邮箱输入（当选择邮箱时显示） -->
          <div v-if="resetPasswordData.method === 'email'" class="form-group">
            <label for="resetEmail">邮箱地址 <span class="required">*</span></label>
            <input 
              type="email" 
              id="resetEmail"
              v-model="resetPasswordData.email" 
              placeholder="请输入注册时使用的邮箱"
              :disabled="resetPasswordLoading"
            />
          </div>
          
          <!-- 验证码输入 -->
          <div class="form-group">
            <label for="resetCode">
              {{ resetPasswordData.method === 'phone' ? '手机' : '邮箱' }}验证码 
              <span class="required">*</span>
            </label>
            <div class="code-input-group">
              <input 
                type="text" 
                id="resetCode"
                v-model="resetPasswordData.code" 
                placeholder="请输入6位验证码"
                maxlength="6"
                pattern="[0-9]{6}"
                :disabled="resetPasswordLoading"
                @input="resetPasswordData.code = resetPasswordData.code.replace(/\D/g, '')"
              />
              <button 
                type="button"
                class="send-code-btn"
                @click="sendResetPasswordCode"
                :disabled="!canSendResetCode || resetCodeSending || resetCodeCountdown > 0 || resetPasswordLoading"
              >
                <span v-if="resetCodeSending">发送中...</span>
                <span v-else-if="resetCodeCountdown > 0">{{ resetCodeCountdown }}s</span>
                <span v-else>发送验证码</span>
              </button>
            </div>
            <p v-if="resetCodeError" class="error-text">{{ resetCodeError }}</p>
            <p v-if="resetCodeSuccess" class="success-text">{{ resetCodeSuccess }}</p>
          </div>
          
          <div class="form-group">
            <label for="resetPassword">新密码 <span class="required">*</span></label>
            <input 
              type="password" 
              id="resetPassword"
              v-model="resetPasswordData.password" 
              placeholder="请输入新密码（至少6位）"
              minlength="6"
              :disabled="resetPasswordLoading"
            />
          </div>
          
          <div v-if="resetPasswordError" class="error-message">
            {{ resetPasswordError }}
          </div>
        </div>
        
        <div class="modal-footer">
          <button type="button" class="btn-cancel" @click="closeResetPasswordModal">取消</button>
          <button 
            type="button"
            class="btn-confirm"
            @click="handleResetPassword"
            :disabled="resetPasswordLoading || !canResetPassword"
          >
            <span v-if="resetPasswordLoading" class="spinner"></span>
            <span v-else>确认重置</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { userApi } from '../utils/api'

const props = defineProps({
  mode: { // 'login' 或 'register'
    type: String,
    default: 'login'
  },
  title: {
    type: String,
    default: '欢迎登录'
  }
})

const authStore = useAuthStore()
const router = useRouter()

// 表单数据
const formData = ref({
  phone: '',
  password: '',
  nickname: '',
  phoneCode: ''  // 手机验证码（6位数）
})

// 加载状态
const loading = ref(false)
// 错误信息
const error = ref('')
// 手机验证码相关
const phoneCodeSending = ref(false)
const phoneCodeError = ref('')
const phoneCodeSuccess = ref('')
const countdown = ref(0)

// 重置密码相关
const showResetPasswordModal = ref(false)
const resetPasswordData = ref({
  method: 'phone', // 'phone' 或 'email'
  phone: '',
  email: '',
  code: '',
  password: ''
})
const resetPasswordLoading = ref(false)
const resetPasswordError = ref('')
const resetCodeSending = ref(false)
const resetCodeError = ref('')
const resetCodeSuccess = ref('')
const resetCodeCountdown = ref(0)

// 验证手机号格式
const isValidResetPhone = computed(() => {
  if (!resetPasswordData.value.phone) return false
  return /^1[3-9]\d{9}$/.test(resetPasswordData.value.phone)
})

// 验证邮箱格式
const isValidResetEmail = computed(() => {
  if (!resetPasswordData.value.email) return false
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(resetPasswordData.value.email)
})

// 检查是否可以发送验证码
const canSendResetCode = computed(() => {
  if (resetPasswordData.value.method === 'phone') {
    return isValidResetPhone.value
  } else {
    return isValidResetEmail.value
  }
})

// 检查是否可以重置密码
const canResetPassword = computed(() => {
  const hasValidInput = resetPasswordData.value.method === 'phone' 
    ? isValidResetPhone.value 
    : isValidResetEmail.value
  return hasValidInput && 
         resetPasswordData.value.code.length === 6 && 
         resetPasswordData.value.password.length >= 6
})

// 根据模式设置标题
const formTitle = computed(() => {
  return props.mode === 'login' ? '登录 HiChat2' : '创建新账号'
})

// 发送手机验证码
const sendPhoneCode = async () => {
  if (!formData.value.phone) {
    phoneCodeError.value = '请先输入手机号'
    return
  }
  
  // 验证手机号格式
  if (!/^1[3-9]\d{9}$/.test(formData.value.phone)) {
    phoneCodeError.value = '请输入正确的手机号'
    return
  }
  
  phoneCodeError.value = ''
  phoneCodeSuccess.value = ''
  phoneCodeSending.value = true
  
  try {
    await userApi.sendPhoneCode(formData.value.phone)
    phoneCodeSuccess.value = '验证码已发送，请查收短信（测试时查看控制台）'
    
    // 开始倒计时
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (err) {
    phoneCodeError.value = err.message || '发送验证码失败，请稍后重试'
  } finally {
    phoneCodeSending.value = false
  }
}

// 表单提交
const submitForm = async () => {
  error.value = ''
  loading.value = true
  
  try {
    if (props.mode === 'login') {
      // 验证手机号格式
      if (!/^1[3-9]\d{9}$/.test(formData.value.phone)) {
        error.value = '请输入正确的手机号'
        return
      }
      
      await authStore.login(
        formData.value.phone, 
        formData.value.password
      )
      // 登录成功后跳转到主界面
      router.push('/app')
    } else {
      // 验证必填字段
      if (!/^1[3-9]\d{9}$/.test(formData.value.phone)) {
        error.value = '请输入正确的手机号'
        return
      }
      if (!formData.value.nickname || formData.value.nickname.trim() === '') {
        error.value = '请输入昵称'
        return
      }
      if (formData.value.password.length < 6) {
        error.value = '密码长度至少6位'
        return
      }
      
      // 验证手机验证码（必填）
      if (!formData.value.phoneCode || formData.value.phoneCode.trim() === '') {
        error.value = '请输入手机验证码'
        return
      }
      
      // 验证验证码格式（必须是6位数字）
      if (!/^\d{6}$/.test(formData.value.phoneCode)) {
        error.value = '验证码必须是6位数字'
        return
      }
      
      await authStore.register({
        phone: formData.value.phone,
        password: formData.value.password,
        nickname: formData.value.nickname,
        phoneCode: formData.value.phoneCode
      })
      // 注册成功后自动登录，跳转到主界面
      router.push('/app')
    }
  } catch (err) {
    error.value = err.message || '请求失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

// 关闭重置密码模态框
const closeResetPasswordModal = () => {
  showResetPasswordModal.value = false
  resetPasswordData.value = {
    method: 'phone',
    phone: '',
    email: '',
    code: '',
    password: ''
  }
  resetPasswordError.value = ''
  resetCodeError.value = ''
  resetCodeSuccess.value = ''
  resetCodeCountdown.value = 0
}

// 发送重置密码验证码
const sendResetPasswordCode = async () => {
  resetCodeError.value = ''
  resetCodeSuccess.value = ''
  
  if (resetPasswordData.value.method === 'phone') {
    // 手机号方式
    if (!resetPasswordData.value.phone) {
      resetCodeError.value = '请先输入手机号'
      return
    }
    
    if (!isValidResetPhone.value) {
      resetCodeError.value = '请输入正确的手机号'
      return
    }
    
    resetCodeSending.value = true
    
    try {
      await userApi.sendPhoneCode(resetPasswordData.value.phone)
      resetCodeSuccess.value = '验证码已发送，请查收短信（测试时查看控制台）'
      
      // 开始倒计时
      resetCodeCountdown.value = 60
      const timer = setInterval(() => {
        resetCodeCountdown.value--
        if (resetCodeCountdown.value <= 0) {
          clearInterval(timer)
        }
      }, 1000)
    } catch (err) {
      resetCodeError.value = err.message || '发送验证码失败，请稍后重试'
    } finally {
      resetCodeSending.value = false
    }
  } else {
    // 邮箱方式
    if (!resetPasswordData.value.email) {
      resetCodeError.value = '请先输入邮箱地址'
      return
    }
    
    if (!isValidResetEmail.value) {
      resetCodeError.value = '请输入正确的邮箱地址'
      return
    }
    
    resetCodeSending.value = true
    
    try {
      await userApi.sendEmailCode(resetPasswordData.value.email)
      resetCodeSuccess.value = '验证码已发送，请查收邮箱（测试时查看控制台）'
      
      // 开始倒计时
      resetCodeCountdown.value = 60
      const timer = setInterval(() => {
        resetCodeCountdown.value--
        if (resetCodeCountdown.value <= 0) {
          clearInterval(timer)
        }
      }, 1000)
    } catch (err) {
      resetCodeError.value = err.message || '发送验证码失败，请稍后重试'
    } finally {
      resetCodeSending.value = false
    }
  }
}

// 重置密码
const handleResetPassword = async () => {
  // 验证输入
  if (resetPasswordData.value.method === 'phone') {
    if (!isValidResetPhone.value) {
      resetPasswordError.value = '请输入正确的手机号'
      return
    }
  } else {
    if (!isValidResetEmail.value) {
      resetPasswordError.value = '请输入正确的邮箱地址'
      return
    }
  }
  
  if (resetPasswordData.value.code.length !== 6) {
    resetPasswordError.value = '请输入6位验证码'
    return
  }
  
  if (resetPasswordData.value.password.length < 6) {
    resetPasswordError.value = '密码长度至少6位'
    return
  }
  
  resetPasswordError.value = ''
  resetPasswordLoading.value = true
  
  try {
    // 根据选择的方式传递不同的参数
    const params = {
      password: resetPasswordData.value.password,
      code: resetPasswordData.value.code
    }
    
    if (resetPasswordData.value.method === 'phone') {
      params.phone = resetPasswordData.value.phone
    } else {
      params.email = resetPasswordData.value.email
    }
    
    await userApi.resetPassword(params)
    
    // 重置成功
    alert('密码重置成功，请使用新密码登录')
    closeResetPasswordModal()
  } catch (err) {
    resetPasswordError.value = err.message || '重置密码失败，请稍后重试'
  } finally {
    resetPasswordLoading.value = false
  }
}
</script>

<style scoped>
.auth-form {
  max-width: 400px;
  margin: 0 auto;
  padding: 30px;
  background-color: #fff;
  border-radius: 10px;
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
}

h2 {
  text-align: center;
  margin-bottom: 25px;
  color: #333;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #555;
}

input {
  width: 100%;
  padding: 12px 15px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  transition: border-color 0.3s;
}

input:focus {
  border-color: #4a8cff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(74, 140, 255, 0.2);
}

.select-input {
  width: 100%;
  padding: 12px 15px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  background-color: white;
  cursor: pointer;
  transition: border-color 0.3s;
}

.select-input:focus {
  border-color: #4a8cff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(74, 140, 255, 0.2);
}

.phone-code-container {
  width: 100%;
}

.phone-code-input-group {
  display: flex;
  gap: 10px;
}

.phone-code-input-group input {
  flex: 1;
}

.send-code-btn {
  padding: 12px 20px;
  background-color: #4a8cff;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 0.3s;
}

.send-code-btn:hover:not(:disabled) {
  background-color: #3a7beb;
}

.send-code-btn:disabled {
  background-color: #a0c0ff;
  cursor: not-allowed;
}

.error-text {
  margin-top: 5px;
  color: #ff4d4d;
  font-size: 12px;
}

.success-text {
  margin-top: 5px;
  color: #52c41a;
  font-size: 12px;
}

.required {
  color: #ff4d4d;
}

.submit-btn {
  width: 100%;
  padding: 13px;
  background-color: #4a8cff;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.3s;
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.submit-btn:hover {
  background-color: #3a7beb;
}

.submit-btn:disabled {
  background-color: #a0c0ff;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  animation: spin 1s infinite linear;
  margin-right: 8px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-message {
  margin-top: 15px;
  padding: 10px;
  background-color: #ffeded;
  border: 1px solid #ffc7c7;
  border-radius: 6px;
  color: #ff4d4d;
  text-align: center;
  font-size: 14px;
}

.switch-auth {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: #666;
}

.switch-link {
  color: #4a8cff;
  text-decoration: none;
  margin-left: 5px;
  font-weight: 500;
}

.switch-link:hover {
  text-decoration: underline;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: center;
}

.reset-password-row {
  margin-top: 12px;
  text-align: center;
}

.reset-password-link {
  color: #4a8cff;
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
}

.reset-password-link:hover {
  text-decoration: underline;
}

/* 重置密码模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.reset-password-modal {
  width: 100%;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 24px;
  color: #999;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.modal-close:hover {
  background-color: #f5f5f5;
  color: #333;
}

.modal-body {
  padding: 24px;
}

.code-input-group {
  display: flex;
  gap: 10px;
}

.code-input-group input {
  flex: 1;
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #f0f0f0;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 10px 20px;
  background: white;
  color: #666;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  border-color: #999;
  color: #333;
}

.btn-confirm {
  padding: 10px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-confirm:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-confirm:disabled {
  background: #d0d0d0;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* 重置方式选择器 */
.reset-method-selector {
  display: flex;
  gap: 16px;
  margin-top: 8px;
}

.method-option {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  transition: all 0.2s;
  user-select: none;
}

.method-option:hover {
  border-color: #4a8cff;
  background-color: #f0f7ff;
}

.method-option input[type="radio"] {
  margin: 0;
  cursor: pointer;
}

.method-option input[type="radio"]:checked + span {
  color: #4a8cff;
  font-weight: 600;
}

.method-option:has(input[type="radio"]:checked) {
  border-color: #4a8cff;
  background-color: #f0f7ff;
}

.method-option span {
  font-size: 14px;
  color: #666;
  transition: all 0.2s;
}

@media (max-width: 480px) {
  .auth-form {
    padding: 20px 15px;
    margin: 0 15px;
  }
  
  .modal-container {
    width: 95%;
    max-height: 85vh;
  }
}
</style>