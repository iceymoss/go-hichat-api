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
    
    <!-- 切换链接 -->
    <div class="switch-auth">
      <span>{{ mode === 'login' ? '还没有账号？' : '已有账号？' }}</span>
      <router-link 
        :to="mode === 'login' ? '/register' : '/login'"
        class="switch-link"
      >
        {{ mode === 'login' ? '立即注册' : '去登录' }}
      </router-link>
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

@media (max-width: 480px) {
  .auth-form {
    padding: 20px 15px;
    margin: 0 15px;
  }
}
</style>