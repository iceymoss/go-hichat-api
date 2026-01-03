<template>
  <div class="login-page">
    <div class="app-info">
      <div class="app-logo">
        <h1>HiChat</h1>
        <span>2.0</span>
      </div>
      <p class="app-description">连接你我，随时随地畅聊</p>
    </div>
    
    <AuthForm mode="login" />
    
    <div class="login-footer">
      <div class="divider">其他登录方式</div>
      <div class="social-login">
        <button class="social-btn wechat">
          <img src="https://static.hotcdn.cn/admin/img/common/v8/icon_wechat_square.png" alt="微信登录">
          微信
        </button>
        <button class="social-btn qq">
          <img src="https://static.hotcdn.cn/admin/img/common/v8/icon_qq_square.png" alt="QQ登录">
          QQ
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import AuthForm from '../components/AuthForm.vue'
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  padding: 50px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.app-info {
  text-align: center;
  color: white;
  margin-bottom: 40px;
}

.app-logo {
  display: flex;
  justify-content: center;
  align-items: baseline;
  font-size: 42px;
  font-weight: bold;
  line-height: 1;
}

.app-logo h1 {
  margin: 0;
  font-size: 48px;
}

.app-logo span {
  font-size: 24px;
  margin-left: 5px;
  position: relative;
  top: 3px;
}

.app-description {
  margin-top: 10px;
  font-size: 16px;
  opacity: 0.9;
}

.login-footer {
  margin-top: 40px;
  width: 100%;
  max-width: 400px;
}

.divider {
  position: relative;
  color: rgba(255, 255, 255, 0.8);
  text-align: center;
  margin-bottom: 20px;
}

.divider::before,
.divider::after {
  content: '';
  position: absolute;
  top: 50%;
  width: calc(50% - 30px);
  height: 1px;
  background-color: rgba(255, 255, 255, 0.3);
}

.divider::before {
  left: 0;
}

.divider::after {
  right: 0;
}

.social-login {
  display: flex;
  justify-content: center;
  gap: 20px;
}

.social-btn {
  width: 100px;
  padding: 10px 0;
  background-color: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  color: white;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.social-btn:hover {
  background-color: rgba(255, 255, 255, 0.25);
}

.social-btn img {
  width: 32px;
  height: 32px;
  margin-bottom: 8px;
}

.social-btn.wechat img {
  filter: drop-shadow(0 0 1px rgba(0, 0, 0, 0.2));
}
</style>