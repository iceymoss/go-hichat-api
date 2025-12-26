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
        
        <div class="form-group">
          <label for="email">邮箱（可选）</label>
          <input 
            type="email" 
            id="email" 
            v-model="formData.email" 
            placeholder="请输入邮箱地址" 
          />
        </div>
        
        <div class="form-group">
          <label for="avatar">头像URL（可选）</label>
          <input 
            type="url" 
            id="avatar" 
            v-model="formData.avatar" 
            placeholder="头像图片链接" 
          />
        </div>
        
        <div class="form-group">
          <label for="sex">性别</label>
          <select id="sex" v-model="formData.sex" class="select-input">
            <option :value="0">未知</option>
            <option :value="1">男</option>
            <option :value="2">女</option>
          </select>
        </div>
      </template>
      
      <!-- 验证码 -->
      <div v-if="mode === 'login'" class="form-group">
        <div class="captcha-container">
          <label for="captcha">验证码</label>
          <input 
            type="text" 
            id="captcha" 
            v-model="formData.captcha" 
            placeholder="验证码" 
            required
          />
          <div class="captcha-img" @click="refreshCaptcha">
            <img :src="captchaImage" alt="captcha" />
          </div>
        </div>
      </div>
      
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

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
  email: '',
  avatar: '',
  sex: 0,
  captcha: ''
})

// 加载状态
const loading = ref(false)
// 错误信息
const error = ref('')
// 验证码图片
const captchaImage = ref('')
// 记住我
const rememberMe = ref(false)

// 根据模式设置标题
const formTitle = computed(() => {
  return props.mode === 'login' ? '登录 HiChat2' : '创建新账号'
})

// 生成验证码（模拟）
const generateCaptcha = () => {
  // 实际项目中应由后端生成验证码
  return 'https://api.dicebear.com/7.x/bottts-neutral/svg?seed=HiChat2'
}

// 刷新验证码
const refreshCaptcha = () => {
  captchaImage.value = generateCaptcha()
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
      
      await authStore.register({
        phone: formData.value.phone,
        password: formData.value.password,
        nickname: formData.value.nickname,
        email: formData.value.email,
        avatar: formData.value.avatar,
        sex: formData.value.sex
      })
      // 注册成功后自动登录，跳转到主界面
      router.push('/app')
    }
  } catch (err) {
    error.value = err.message || '请求失败，请稍后重试'
    // 刷新验证码
    if (props.mode === 'login') refreshCaptcha()
  } finally {
    loading.value = false
  }
}

// 组件挂载时初始化验证码
onMounted(() => {
  if (props.mode === 'login') {
    refreshCaptcha()
  }
})
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

.captcha-container {
  display: flex;
  gap: 10px;
}

.captcha-container input {
  flex: 1;
}

.captcha-img {
  width: 100px;
  height: 48px;
  background-color: #f5f7fa;
  border-radius: 6px;
  border: 1px solid #ddd;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.captcha-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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