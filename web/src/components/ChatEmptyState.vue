<template>
  <div class="chat-empty-state">
    <!-- 背景装饰 -->
    <div class="background-decoration">
      <div class="floating-shape shape-1"></div>
      <div class="floating-shape shape-2"></div>
      <div class="floating-shape shape-3"></div>
      <div class="floating-shape shape-4"></div>
    </div>
    
    <!-- 主要内容 -->
    <div class="empty-content">
      <!-- 主图标 -->
      <div class="main-icon">
        <div class="icon-container">
          <i class="icon icon-message-circle primary-icon"></i>
          <div class="icon-ring ring-1"></div>
          <div class="icon-ring ring-2"></div>
          <div class="icon-ring ring-3"></div>
        </div>
      </div>
      
      <!-- 文字内容 -->
      <div class="text-content">
        <h2 class="main-title">欢迎使用 HiChat</h2>
        <p class="subtitle">选择一个会话开始聊天，或者创建新的对话</p>
        <div class="features">
          <div class="feature-item">
            <i class="icon icon-zap"></i>
            <span>实时消息</span>
          </div>
          <div class="feature-item">
            <i class="icon icon-image"></i>
            <span>多媒体支持</span>
          </div>
          <div class="feature-item">
            <i class="icon icon-shield"></i>
            <span>安全加密</span>
          </div>
        </div>
      </div>
      
      <!-- 快速操作 -->
      <div class="quick-actions">
        <button class="action-btn primary" @click="handleStartNewChat">
          <i class="icon icon-plus"></i>
          <span>开始新对话</span>
          <div class="btn-glow"></div>
        </button>
        <button class="action-btn secondary" @click="handleViewContacts">
          <i class="icon icon-users"></i>
          <span>查看联系人</span>
        </button>
      </div>
      
      <!-- 快捷功能 -->
      <div class="shortcuts">
        <div class="shortcut-item" @click="handleQuickMessage">
          <div class="shortcut-icon">
            <i class="icon icon-send"></i>
          </div>
          <div class="shortcut-text">
            <div class="shortcut-title">快速发送</div>
            <div class="shortcut-desc">发送文字、图片、文件</div>
          </div>
        </div>
        
        <div class="shortcut-item" @click="handleVoiceCall">
          <div class="shortcut-icon">
            <i class="icon icon-phone"></i>
          </div>
          <div class="shortcut-text">
            <div class="shortcut-title">语音通话</div>
            <div class="shortcut-desc">高质量语音通话</div>
          </div>
        </div>
        
        <div class="shortcut-item" @click="handleVideoCall">
          <div class="shortcut-icon">
            <i class="icon icon-video"></i>
          </div>
          <div class="shortcut-text">
            <div class="shortcut-title">视频聊天</div>
            <div class="shortcut-desc">面对面视频交流</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const emit = defineEmits(['start-chat'])

function handleStartNewChat() {
  console.log('开始新对话')
  // TODO: 实现新建对话逻辑
}

function handleViewContacts() {
  router.push('/contacts')
}

function handleQuickMessage() {
  console.log('快速发送消息')
  // TODO: 实现快速发送功能
}

function handleVoiceCall() {
  console.log('发起语音通话')
  // TODO: 实现语音通话功能
}

function handleVideoCall() {
  console.log('发起视频通话')
  // TODO: 实现视频通话功能
}
</script>

<style scoped>
.chat-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 40px;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, 
    rgba(248, 250, 252, 0.9) 0%, 
    rgba(241, 245, 249, 0.8) 50%,
    rgba(235, 239, 245, 0.9) 100%);
}

/* 背景装饰 */
.background-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.floating-shape {
  position: absolute;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.1), rgba(138, 105, 255, 0.05));
  animation: float 6s ease-in-out infinite;
}

.shape-1 {
  width: 120px;
  height: 120px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.shape-2 {
  width: 80px;
  height: 80px;
  top: 20%;
  right: 15%;
  animation-delay: 2s;
}

.shape-3 {
  width: 60px;
  height: 60px;
  bottom: 20%;
  left: 20%;
  animation-delay: 4s;
}

.shape-4 {
  width: 100px;
  height: 100px;
  bottom: 15%;
  right: 10%;
  animation-delay: 1s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
    opacity: 0.3;
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
    opacity: 0.6;
  }
}

/* 主要内容 */
.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 600px;
  width: 100%;
  position: relative;
  z-index: 1;
}

/* 主图标 */
.main-icon {
  margin-bottom: 32px;
  position: relative;
}

.icon-container {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 120px;
}

.primary-icon {
  font-size: 48px;
  color: #4a8cff;
  z-index: 3;
  position: relative;
  filter: drop-shadow(0 8px 16px rgba(74, 140, 255, 0.3));
  animation: iconPulse 3s ease-in-out infinite;
}

@keyframes iconPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.icon-ring {
  position: absolute;
  border: 2px solid;
  border-radius: 50%;
  opacity: 0.6;
}

.ring-1 {
  width: 100px;
  height: 100px;
  border-color: rgba(74, 140, 255, 0.3);
  animation: ringExpand 4s ease-in-out infinite;
  animation-delay: 0s;
}

.ring-2 {
  width: 80px;
  height: 80px;
  border-color: rgba(138, 105, 255, 0.3);
  animation: ringExpand 4s ease-in-out infinite;
  animation-delay: 1s;
}

.ring-3 {
  width: 60px;
  height: 60px;
  border-color: rgba(74, 140, 255, 0.4);
  animation: ringExpand 4s ease-in-out infinite;
  animation-delay: 2s;
}

@keyframes ringExpand {
  0%, 100% {
    transform: scale(1);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.2;
  }
}

/* 文字内容 */
.text-content {
  margin-bottom: 40px;
}

.main-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 12px 0;
  background: linear-gradient(135deg, #1e293b 0%, #4a8cff 50%, #8a69ff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: titleShine 3s ease-in-out infinite;
}

@keyframes titleShine {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.subtitle {
  font-size: 16px;
  color: #64748b;
  margin: 0 0 32px 0;
  line-height: 1.6;
  max-width: 400px;
}

.features {
  display: flex;
  gap: 32px;
  justify-content: center;
  flex-wrap: wrap;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 20px;
  border: 1px solid rgba(74, 140, 255, 0.1);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(74, 140, 255, 0.1);
  border-color: rgba(74, 140, 255, 0.3);
  transform: translateY(-2px);
}

.feature-item .icon {
  font-size: 16px;
  color: #4a8cff;
}

.feature-item span {
  font-size: 14px;
  color: #1e293b;
  font-weight: 500;
}

/* 快速操作 */
.quick-actions {
  display: flex;
  gap: 16px;
  margin-bottom: 48px;
  flex-wrap: wrap;
  justify-content: center;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 24px;
  border: none;
  border-radius: 16px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  min-width: 160px;
  justify-content: center;
}

.action-btn.primary {
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  color: white;
  box-shadow: 
    0 8px 25px rgba(74, 140, 255, 0.3),
    0 4px 12px rgba(138, 105, 255, 0.2);
}

.action-btn.primary:hover {
  transform: translateY(-3px) scale(1.02);
  box-shadow: 
    0 12px 35px rgba(74, 140, 255, 0.4),
    0 6px 16px rgba(138, 105, 255, 0.3);
}

.action-btn.secondary {
  background: rgba(255, 255, 255, 0.8);
  color: #4a8cff;
  border: 1px solid rgba(74, 140, 255, 0.2);
  backdrop-filter: blur(10px);
}

.action-btn.secondary:hover {
  background: rgba(74, 140, 255, 0.1);
  border-color: rgba(74, 140, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(74, 140, 255, 0.2);
}

.btn-glow {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.2), transparent);
  border-radius: 16px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.action-btn.primary:hover .btn-glow {
  opacity: 1;
}

.action-btn .icon {
  font-size: 18px;
}

/* 快捷功能 */
.shortcuts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  width: 100%;
  max-width: 720px;
}

.shortcut-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(229, 231, 235, 0.5);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(20px);
}

.shortcut-item:hover {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(74, 140, 255, 0.3);
  transform: translateY(-4px);
  box-shadow: 0 12px 28px rgba(74, 140, 255, 0.15);
}

.shortcut-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.1), rgba(138, 105, 255, 0.05));
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.3s ease;
}

.shortcut-item:hover .shortcut-icon {
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.2), rgba(138, 105, 255, 0.1));
  transform: scale(1.1);
}

.shortcut-icon .icon {
  font-size: 20px;
  color: #4a8cff;
}

.shortcut-text {
  flex: 1;
  text-align: left;
}

.shortcut-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
}

.shortcut-desc {
  font-size: 14px;
  color: #64748b;
  line-height: 1.4;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chat-empty-state {
    padding: 20px;
  }
  
  .main-title {
    font-size: 28px;
  }
  
  .subtitle {
    font-size: 15px;
  }
  
  .features {
    gap: 16px;
  }
  
  .quick-actions {
    flex-direction: column;
    align-items: center;
  }
  
  .action-btn {
    min-width: 200px;
  }
  
  .shortcuts {
    grid-template-columns: 1fr;
  }
  
  .shortcut-item {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .icon-container {
    width: 100px;
    height: 100px;
  }
  
  .primary-icon {
    font-size: 40px;
  }
  
  .main-title {
    font-size: 24px;
  }
  
  .features {
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }
}
</style> 