<template>
  <Transition name="profile-card" appear>
    <div 
      v-if="visible && user" 
      class="user-profile-card"
      :style="cardStyle"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
    >
      <!-- 背景装饰 -->
      <div class="card-background">
        <div class="bg-pattern"></div>
        <div class="bg-glow"></div>
      </div>
      
      <!-- 卡片内容 -->
      <div class="card-content">
        <!-- 用户头像和基本信息 -->
        <div class="user-header">
          <div class="avatar-container">
            <img :src="user.avatar" :alt="user.name" class="user-avatar" />
            <div class="avatar-ring"></div>
            <div class="online-status" v-if="user.isOnline"></div>
          </div>
          
          <div class="user-info">
            <h3 class="user-name">{{ user.name }}</h3>
            <p class="user-description" v-if="user.description">
              {{ user.description }}
            </p>
            <p class="user-description" v-else>
              这个人很懒，什么都没写...
            </p>
          </div>
        </div>
        
        <!-- 用户统计 -->
        <div class="user-stats">
          <div class="stat-item">
            <span class="stat-number">{{ user.feedCount || 0 }}</span>
            <span class="stat-label">动态</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <span class="stat-number">{{ user.friendCount || 0 }}</span>
            <span class="stat-label">好友</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <span class="stat-number">{{ user.likeCount || 0 }}</span>
            <span class="stat-label">获赞</span>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="action-buttons">
          <button class="action-btn primary" @click="viewProfile">
            <i class="icon icon-user"></i>
            <span>个人资料</span>
          </button>
          <button class="action-btn secondary" @click="viewSpace">
            <i class="icon icon-grid"></i>
            <span>个人空间</span>
          </button>
        </div>
      </div>
      
      <!-- 卡片箭头 -->
      <div class="card-arrow" :class="arrowPosition"></div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  user: {
    type: Object,
    default: null
  },
  position: {
    type: Object,
    default: () => ({ x: 0, y: 0 })
  },
  arrowPosition: {
    type: String,
    default: 'top' // top, bottom, left, right
  }
})

const emit = defineEmits(['close', 'view-profile', 'view-space', 'mouse-enter-card', 'mouse-leave-card'])
const router = useRouter()

// 计算卡片位置样式
const cardStyle = computed(() => {
  if (!props.visible || !props.position) return {}
  
  const { x, y } = props.position
  const offset = 10 // 箭头偏移量
  
  let transform = ''
  let top = ''
  let left = ''
  
  switch (props.arrowPosition) {
    case 'top':
      top = `${y + offset}px`
      left = `${x}px`
      transform = 'translateX(-50%)'
      break
    case 'bottom':
      top = `${y - offset}px`
      left = `${x}px`
      transform = 'translateX(-50%) translateY(-100%)'
      break
    case 'left':
      top = `${y}px`
      left = `${x + offset}px`
      transform = 'translateY(-50%)'
      break
    case 'right':
      top = `${y}px`
      left = `${x - offset}px`
      transform = 'translateY(-50%) translateX(-100%)'
      break
  }
  
  return {
    position: 'fixed',
    top,
    left,
    transform,
    zIndex: 1000
  }
})

const handleMouseEnter = () => {
  // 鼠标进入卡片时不关闭，清除关闭定时器
  emit('mouse-enter-card')
}

const handleMouseLeave = () => {
  // 鼠标离开卡片时延迟关闭
  emit('mouse-leave-card')
}

const viewProfile = () => {
  emit('view-profile', props.user)
  emit('close')
}

const viewSpace = () => {
  emit('view-space', props.user)
  emit('close')
}
</script>

<style scoped>
/* 进入/退出动画 */
.profile-card-enter-active {
  transition: all 0.4s cubic-bezier(0.23, 1, 0.320, 1);
}

.profile-card-leave-active {
  transition: all 0.3s cubic-bezier(0.755, 0.05, 0.855, 0.06);
}

.profile-card-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-10px) scale(0.9);
  filter: blur(5px);
}

.profile-card-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(10px) scale(0.95);
  filter: blur(3px);
}

.user-profile-card {
  width: 320px;
  background: linear-gradient(135deg, 
    rgba(255, 255, 255, 0.98) 0%, 
    rgba(252, 253, 255, 0.95) 100%);
  border-radius: 20px;
  padding: 24px;
  box-shadow: 
    0 25px 60px rgba(0, 0, 0, 0.15),
    0 8px 32px rgba(74, 140, 255, 0.1),
    0 0 0 1px rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(229, 231, 235, 0.3);
  backdrop-filter: blur(25px) saturate(1.5);
  position: relative;
  overflow: hidden;
  animation: cardFloat 6s ease-in-out infinite;
}

@keyframes cardFloat {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-2px); }
}

.card-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
  border-radius: 20px;
}

.bg-pattern {
  position: absolute;
  top: -50%;
  left: -50%;
  right: -50%;
  bottom: -50%;
  background: 
    radial-gradient(circle at 20% 20%, rgba(74, 140, 255, 0.08) 0%, transparent 50%),
    radial-gradient(circle at 80% 80%, rgba(138, 105, 255, 0.06) 0%, transparent 50%);
  animation: patternRotate 20s linear infinite;
}

@keyframes patternRotate {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.bg-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 200px;
  height: 200px;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, 
    rgba(74, 140, 255, 0.1) 0%, 
    transparent 70%);
  filter: blur(20px);
  animation: glowPulse 4s ease-in-out infinite alternate;
}

@keyframes glowPulse {
  0% { opacity: 0.3; transform: translate(-50%, -50%) scale(0.8); }
  100% { opacity: 0.6; transform: translate(-50%, -50%) scale(1.2); }
}

.card-content {
  position: relative;
  z-index: 1;
}

.user-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.avatar-container {
  position: relative;
  flex-shrink: 0;
}

.user-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: 3px solid rgba(255, 255, 255, 0.9);
  box-shadow: 
    0 8px 24px rgba(0, 0, 0, 0.12),
    0 0 0 2px rgba(74, 140, 255, 0.2);
  transition: all 0.3s ease;
  cursor: pointer;
}

.user-avatar:hover {
  transform: scale(1.05);
  box-shadow: 
    0 12px 32px rgba(0, 0, 0, 0.16),
    0 0 0 3px rgba(74, 140, 255, 0.3);
}

.avatar-ring {
  position: absolute;
  inset: -3px;
  border-radius: 50%;
  background: linear-gradient(45deg, #4a8cff, #8a69ff, #4a8cff);
  background-size: 300% 300%;
  opacity: 0;
  animation: ringPulse 3s ease-in-out infinite;
  z-index: -1;
}

@keyframes ringPulse {
  0%, 100% { opacity: 0; background-position: 0% 50%; }
  50% { opacity: 0.6; background-position: 100% 50%; }
}

.online-status {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 16px;
  height: 16px;
  background: linear-gradient(135deg, #10b981, #059669);
  border: 3px solid white;
  border-radius: 50%;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
  animation: onlinePulse 2s ease-in-out infinite;
}

@keyframes onlinePulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 6px 0;
  color: #1e293b;
  background: linear-gradient(135deg, #1e293b 0%, #4a8cff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1.3;
}

.user-description {
  font-size: 14px;
  color: #64748b;
  margin: 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.user-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding: 16px;
  background: linear-gradient(135deg, 
    rgba(248, 250, 252, 0.8) 0%, 
    rgba(241, 245, 249, 0.6) 100%);
  border-radius: 12px;
  border: 1px solid rgba(229, 231, 235, 0.4);
  backdrop-filter: blur(10px);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.stat-number {
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.stat-divider {
  width: 1px;
  height: 24px;
  background: linear-gradient(to bottom, 
    transparent 0%, 
    rgba(229, 231, 235, 0.6) 50%, 
    transparent 100%);
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.action-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, 
    rgba(255, 255, 255, 0.2) 0%, 
    transparent 50%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.action-btn:hover::before {
  opacity: 1;
}

.action-btn.primary {
  background: linear-gradient(135deg, #4a8cff 0%, #8a69ff 100%);
  color: white;
  box-shadow: 
    0 4px 15px rgba(74, 140, 255, 0.3),
    0 2px 8px rgba(138, 105, 255, 0.2);
}

.action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 
    0 6px 20px rgba(74, 140, 255, 0.4),
    0 4px 12px rgba(138, 105, 255, 0.3);
}

.action-btn.secondary {
  background: rgba(248, 250, 252, 0.8);
  color: #4a8cff;
  border: 1px solid rgba(74, 140, 255, 0.2);
  backdrop-filter: blur(10px);
}

.action-btn.secondary:hover {
  background: rgba(74, 140, 255, 0.1);
  border-color: rgba(74, 140, 255, 0.3);
  transform: translateY(-1px);
  box-shadow: 
    0 4px 12px rgba(74, 140, 255, 0.2),
    0 2px 6px rgba(0, 0, 0, 0.1);
}

.action-btn .icon {
  font-size: 16px;
}

/* 卡片箭头 */
.card-arrow {
  position: absolute;
  width: 12px;
  height: 12px;
  background: linear-gradient(135deg, 
    rgba(255, 255, 255, 0.98) 0%, 
    rgba(252, 253, 255, 0.95) 100%);
  border: 1px solid rgba(229, 231, 235, 0.3);
  transform: rotate(45deg);
  z-index: -1;
}

.card-arrow.top {
  top: -6px;
  left: 50%;
  transform: translateX(-50%) rotate(45deg);
  border-bottom: none;
  border-right: none;
}

.card-arrow.bottom {
  bottom: -6px;
  left: 50%;
  transform: translateX(-50%) rotate(45deg);
  border-top: none;
  border-left: none;
}

.card-arrow.left {
  left: -6px;
  top: 50%;
  transform: translateY(-50%) rotate(45deg);
  border-top: none;
  border-right: none;
}

.card-arrow.right {
  right: -6px;
  top: 50%;
  transform: translateY(-50%) rotate(45deg);
  border-bottom: none;
  border-left: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .user-profile-card {
    width: 280px;
    padding: 20px;
  }
  
  .user-header {
    gap: 12px;
    margin-bottom: 16px;
  }
  
  .user-avatar {
    width: 56px;
    height: 56px;
  }
  
  .user-name {
    font-size: 16px;
  }
  
  .user-description {
    font-size: 13px;
  }
  
  .action-buttons {
    flex-direction: column;
    gap: 8px;
  }
  
  .action-btn {
    padding: 10px 14px;
    font-size: 13px;
  }
}
</style> 