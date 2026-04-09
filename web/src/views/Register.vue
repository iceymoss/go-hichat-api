<template>
  <div class="register-page">
    <div class="register-container">
      <div class="app-info">
        <div class="app-logo">
          <span class="logo-text">HiChat</span>
          <span class="logo-version">2.0</span>
        </div>
      </div>

      <AuthForm mode="register" title="Create Account" />

      <div class="terms">
        By registering, you agree to our
        <a href="#" class="terms-link">Terms of Service</a> and
        <a href="#" class="terms-link">Privacy Policy</a>
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
  background-color: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.register-container {
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
}

.logo-text {
  font-size: 36px;
  font-weight: 700;
  color: #111827;
  letter-spacing: -0.5px;
}

.logo-version {
  font-size: 16px;
  font-weight: 600;
  color: #2563eb;
  background-color: #eff6ff;
  padding: 2px 8px;
  border-radius: 6px;
}

.terms {
  margin-top: 20px;
  text-align: center;
  color: #9ca3af;
  font-size: 13px;
  line-height: 1.6;
}

.terms-link {
  color: #2563eb;
  text-decoration: none;
  font-weight: 500;
}

.terms-link:hover {
  text-decoration: underline;
}
</style>