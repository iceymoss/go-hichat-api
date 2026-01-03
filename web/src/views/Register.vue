<template>
  <div class="register-page">
    <div class="form-wrapper">
      <h1 class="app-title">HiChat<span>2.0</span></h1>
      
      <AuthForm mode="register" title="创建新账号" />
      
      <div class="terms">
        注册即表示您同意我们的
        <a href="#" class="terms-link">用户协议</a>和
        <a href="#" class="terms-link">隐私政策</a>
      </div>
    </div>
  </div>
</template>

<script setup>
import AuthForm from '../components/AuthForm.vue'
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #ff7e5f, #feb47b);
  padding: 50px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.form-wrapper {
  width: 100%;
  max-width: 450px;
  background-color: #fff;
  border-radius: 12px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
}

.app-title {
  text-align: center;
  font-size: 32px;
  color: #ff7e5f;
  margin-bottom: 30px;
  font-weight: bold;
}

.app-title span {
  font-size: 24px;
  position: relative;
  top: -3px;
  margin-left: 3px;
  color: #feb47b;
}

.terms {
  margin-top: 25px;
  text-align: center;
  color: #666;
  font-size: 13px;
}

.terms-link {
  color: #ff7e5f;
  text-decoration: none;
  font-weight: 500;
}

.terms-link:hover {
  text-decoration: underline;
}
</style>