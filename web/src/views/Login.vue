<template>
  <div class="login-page">
    <div class="login-container">
      <div class="app-info">
        <div class="app-logo">
          <span class="logo-text">HiChat</span>
          <span class="logo-version">2.0</span>
        </div>
        <p class="app-description">Connect with friends, anytime, anywhere</p>
      </div>

      <AuthForm mode="login" />

      <div class="login-footer">
        <div class="divider">
          <span class="divider-line"></span>
          <span class="divider-text">or continue with</span>
          <span class="divider-line"></span>
        </div>
        <div class="social-login">
          <button class="social-btn">
            <img src="https://static.hotcdn.cn/admin/img/common/v8/icon_wechat_square.png" alt="WeChat">
          </button>
          <button class="social-btn">
            <img src="https://static.hotcdn.cn/admin/img/common/v8/icon_qq_square.png" alt="QQ">
          </button>
        </div>
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
  background-color: var(--bg);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  font-family: var(--font-ui);
}

.login-container {
  width: 100%;
  max-width: 420px;
}

.app-info {
  text-align: center;
  margin-bottom: 32px;
}

.app-logo {
  display: flex;
  justify-content: center;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 8px;
}

.logo-text {
  font-size: 36px;
  font-weight: 700;
  color: var(--fg);
  letter-spacing: -0.5px;
  font-family: var(--font-display);
}

.logo-version {
  font-size: 16px;
  font-weight: 600;
  color: var(--accent);
  background-color: var(--accent-muted);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.app-description {
  margin: 0;
  font-size: 15px;
  color: var(--muted);
}

.login-footer {
  margin-top: 24px;
}

.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.divider-line {
  flex: 1;
  height: 1px;
  background-color: var(--border);
}

.divider-text {
  font-size: 13px;
  color: var(--muted);
  white-space: nowrap;
}

.social-login {
  display: flex;
  justify-content: center;
  gap: 12px;
}

.social-btn {
  width: 48px;
  height: 48px;
  padding: 0;
  background-color: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s, background-color 0.2s;
}

.social-btn:hover {
  border-color: var(--accent);
  background-color: var(--card-hover);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.social-btn img {
  width: 24px;
  height: 24px;
}
</style>