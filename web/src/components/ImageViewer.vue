<template>
  <Transition name="image-viewer" appear>
    <div v-if="visible" class="image-viewer-overlay" @click.self="closeViewer">
      <!-- 动态背景粒子效果 -->
      <div class="background-particles">
        <div v-for="i in 20" :key="i" class="particle" :style="getParticleStyle(i)"></div>
      </div>
      
      <!-- 光晕效果 -->
      <div class="ambient-glow"></div>
      
      <div class="image-viewer-container">
      <!-- 关闭按钮 -->
      <button class="close-btn" @click="closeViewer">
        <i class="icon icon-x"></i>
      </button>
      
      <!-- 图片计数器 -->
      <div class="image-counter" v-if="images.length > 1">
        {{ currentIndex + 1 }} / {{ images.length }}
      </div>
      
      <!-- 主图片区域 -->
      <div class="image-container" ref="imageContainer">
        <img 
          :src="currentImage" 
          :alt="`图片 ${currentIndex + 1}`"
          class="main-image"
          ref="mainImage"
          @load="handleImageLoad"
          @click="toggleZoom"
          :style="imageStyle"
          @mousedown="startDrag"
          @wheel="handleWheel"
        />
        
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-spinner">
          <div class="spinner"></div>
        </div>
      </div>
      
      <!-- 左右切换按钮 -->
      <button 
        v-if="images.length > 1" 
        class="nav-btn prev-btn" 
        @click="prevImage"
        :disabled="currentIndex === 0"
      >
        <i class="icon icon-chevron-left"></i>
      </button>
      
      <button 
        v-if="images.length > 1" 
        class="nav-btn next-btn" 
        @click="nextImage"
        :disabled="currentIndex === images.length - 1"
      >
        <i class="icon icon-chevron-right"></i>
      </button>
      
      <!-- 缩略图导航 -->
      <div v-if="images.length > 1" class="thumbnail-nav">
        <div 
          v-for="(image, index) in images" 
          :key="index"
          class="thumbnail-item"
          :class="{ active: index === currentIndex }"
          @click="goToImage(index)"
        >
          <img :src="image" :alt="`缩略图 ${index + 1}`" />
        </div>
      </div>
      
      <!-- 工具栏 -->
      <div class="toolbar">
        <button class="tool-btn" @click="zoomOut" :disabled="scale <= 0.5">
          <i class="icon icon-zoom-out"></i>
        </button>
        <span class="zoom-level">{{ Math.round(scale * 100) }}%</span>
        <button class="tool-btn" @click="zoomIn" :disabled="scale >= 3">
          <i class="icon icon-zoom-in"></i>
        </button>
        <button class="tool-btn" @click="resetZoom">
          <i class="icon icon-refresh"></i>
        </button>
        <button class="tool-btn" @click="downloadImage">
          <i class="icon icon-download"></i>
        </button>
      </div>
    </div>
      </div>
    </Transition>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  images: {
    type: Array,
    default: () => []
  },
  initialIndex: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['close'])

// 状态
const currentIndex = ref(0)
const scale = ref(1)
const translateX = ref(0)
const translateY = ref(0)
const loading = ref(false)
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })

// 引用
const imageContainer = ref(null)
const mainImage = ref(null)

// 计算属性
const currentImage = computed(() => {
  return props.images[currentIndex.value] || ''
})

const imageStyle = computed(() => {
  return {
    transform: `translate(${translateX.value}px, ${translateY.value}px) scale(${scale.value})`,
    cursor: isDragging.value ? 'grabbing' : (scale.value > 1 ? 'grab' : 'zoom-in')
  }
})

// 监听器
watch(() => props.visible, (visible) => {
  if (visible) {
    currentIndex.value = props.initialIndex
    resetZoom()
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

watch(() => props.initialIndex, (newIndex) => {
  currentIndex.value = newIndex
})

watch(currentIndex, () => {
  resetZoom()
})

// 生命周期
onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
  document.body.style.overflow = ''
})

// 方法
const closeViewer = () => {
  emit('close')
}

const prevImage = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

const nextImage = () => {
  if (currentIndex.value < props.images.length - 1) {
    currentIndex.value++
  }
}

const goToImage = (index) => {
  currentIndex.value = index
}

const handleImageLoad = () => {
  loading.value = false
}

const toggleZoom = () => {
  if (scale.value === 1) {
    zoomIn()
  } else {
    resetZoom()
  }
}

const zoomIn = () => {
  scale.value = Math.min(scale.value * 1.5, 3)
}

const zoomOut = () => {
  scale.value = Math.max(scale.value / 1.5, 0.5)
}

const resetZoom = () => {
  scale.value = 1
  translateX.value = 0
  translateY.value = 0
}

const handleWheel = (e) => {
  e.preventDefault()
  const delta = e.deltaY > 0 ? -0.1 : 0.1
  const newScale = Math.max(0.5, Math.min(3, scale.value + delta))
  scale.value = newScale
}

const startDrag = (e) => {
  if (scale.value <= 1) return
  
  isDragging.value = true
  dragStart.value = { x: e.clientX, y: e.clientY }
  dragOffset.value = { x: translateX.value, y: translateY.value }
}

const handleMouseMove = (e) => {
  if (!isDragging.value) return
  
  const deltaX = e.clientX - dragStart.value.x
  const deltaY = e.clientY - dragStart.value.y
  
  translateX.value = dragOffset.value.x + deltaX
  translateY.value = dragOffset.value.y + deltaY
}

const handleMouseUp = () => {
  isDragging.value = false
}

const handleKeydown = (e) => {
  if (!props.visible) return
  
  switch (e.key) {
    case 'Escape':
      closeViewer()
      break
    case 'ArrowLeft':
      prevImage()
      break
    case 'ArrowRight':
      nextImage()
      break
    case '+':
    case '=':
      zoomIn()
      break
    case '-':
      zoomOut()
      break
    case '0':
      resetZoom()
      break
  }
}

const downloadImage = () => {
  const link = document.createElement('a')
  link.href = currentImage.value
  link.download = `image_${currentIndex.value + 1}.jpg`
  link.click()
}

// 粒子效果样式生成
const getParticleStyle = (index) => {
  const randomSize = Math.random() * 3 + 1
  const randomX = Math.random() * 100
  const randomY = Math.random() * 100
  const randomDelay = Math.random() * 5
  const randomDuration = Math.random() * 10 + 5
  
  return {
    width: `${randomSize}px`,
    height: `${randomSize}px`,
    left: `${randomX}%`,
    top: `${randomY}%`,
    animationDelay: `${randomDelay}s`,
    animationDuration: `${randomDuration}s`
  }
}
</script>

<style scoped>
/* 图片查看器进入/退出动画 */
.image-viewer-enter-active {
  transition: all 0.6s cubic-bezier(0.23, 1, 0.320, 1);
}

.image-viewer-leave-active {
  transition: all 0.4s cubic-bezier(0.755, 0.05, 0.855, 0.06);
}

.image-viewer-enter-from {
  opacity: 0;
  transform: scale(0.8) rotateY(-15deg);
  filter: blur(10px);
}

.image-viewer-leave-to {
  opacity: 0;
  transform: scale(1.2) rotateY(15deg);
  filter: blur(5px);
}

.image-viewer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, 
    rgba(0, 0, 0, 0.95) 0%, 
    rgba(17, 24, 39, 0.98) 50%, 
    rgba(0, 0, 0, 0.95) 100%);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(20px) saturate(1.5);
}

/* 背景粒子效果 */
.background-particles {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
}

.particle {
  position: absolute;
  background: linear-gradient(45deg, 
    rgba(74, 140, 255, 0.6), 
    rgba(138, 105, 255, 0.4), 
    rgba(255, 255, 255, 0.3));
  border-radius: 50%;
  filter: blur(1px);
  animation: float linear infinite;
}

@keyframes float {
  0% {
    transform: translateY(100vh) rotate(0deg);
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    transform: translateY(-100vh) rotate(360deg);
    opacity: 0;
  }
}

/* 环境光晕效果 */
.ambient-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 800px;
  height: 800px;
  transform: translate(-50%, -50%) scale(1);
  background: radial-gradient(circle, 
    rgba(74, 140, 255, 0.08) 0%, 
    rgba(138, 105, 255, 0.05) 30%, 
    transparent 70%);
  border-radius: 50%;
  pointer-events: none;
  opacity: 0.4;
}

.image-viewer-container {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 50px;
  height: 50px;
  background: linear-gradient(135deg, 
    rgba(255, 255, 255, 0.2), 
    rgba(255, 255, 255, 0.1));
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  color: white;
  font-size: 24px;
  cursor: pointer;
  z-index: 10;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(15px) saturate(1.5);
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.close-btn:hover {
  background: linear-gradient(135deg, 
    rgba(239, 68, 68, 0.8), 
    rgba(220, 38, 38, 0.6));
  border-color: rgba(239, 68, 68, 0.6);
  transform: scale(1.15) rotate(90deg);
  box-shadow: 
    0 12px 48px rgba(239, 68, 68, 0.4),
    0 0 20px rgba(239, 68, 68, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.image-counter {
  position: absolute;
  top: 20px;
  left: 20px;
  background: linear-gradient(135deg, 
    rgba(0, 0, 0, 0.8), 
    rgba(17, 24, 39, 0.7));
  color: white;
  padding: 10px 18px;
  border-radius: 25px;
  font-size: 14px;
  font-weight: 600;
  z-index: 10;
  backdrop-filter: blur(15px) saturate(1.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.image-container {
  position: relative;
  max-width: 90%;
  max-height: 90%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 20px;
  box-shadow: 
    0 25px 80px rgba(0, 0, 0, 0.5),
    0 0 50px rgba(74, 140, 255, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.main-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
  -webkit-user-drag: none;
  border-radius: 20px;
  filter: saturate(1.1) contrast(1.05);
  animation: imageReveal 0.8s cubic-bezier(0.23, 1, 0.320, 1);
}

@keyframes imageReveal {
  0% {
    opacity: 0;
    transform: scale(0.8) rotateY(-20deg);
    filter: blur(20px) saturate(0.5);
  }
  100% {
    opacity: 1;
    transform: scale(1) rotateY(0deg);
    filter: saturate(1.1) contrast(1.05);
  }
}

.loading-spinner {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.spinner {
  width: 60px;
  height: 60px;
  border: 4px solid transparent;
  border-top: 4px solid #4a8cff;
  border-right: 4px solid #8a69ff;
  border-radius: 50%;
  animation: spin 1.5s cubic-bezier(0.68, -0.55, 0.265, 1.55) infinite;
  box-shadow: 
    0 0 30px rgba(74, 140, 255, 0.3),
    inset 0 0 20px rgba(138, 105, 255, 0.2);
}

@keyframes spin {
  0% { 
    transform: rotate(0deg) scale(1);
    box-shadow: 
      0 0 30px rgba(74, 140, 255, 0.3),
      inset 0 0 20px rgba(138, 105, 255, 0.2);
  }
  50% { 
    transform: rotate(180deg) scale(1.1);
    box-shadow: 
      0 0 40px rgba(74, 140, 255, 0.5),
      inset 0 0 30px rgba(138, 105, 255, 0.4);
  }
  100% { 
    transform: rotate(360deg) scale(1);
    box-shadow: 
      0 0 30px rgba(74, 140, 255, 0.3),
      inset 0 0 20px rgba(138, 105, 255, 0.2);
  }
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, 
    rgba(255, 255, 255, 0.2), 
    rgba(255, 255, 255, 0.1));
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  color: white;
  font-size: 24px;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(15px) saturate(1.5);
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.nav-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, 
    rgba(74, 140, 255, 0.8), 
    rgba(138, 105, 255, 0.6));
  border-color: rgba(74, 140, 255, 0.6);
  transform: translateY(-50%) scale(1.2);
  box-shadow: 
    0 15px 50px rgba(74, 140, 255, 0.4),
    0 0 30px rgba(74, 140, 255, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.nav-btn:disabled {
  opacity: 0.2;
  cursor: not-allowed;
}

.prev-btn {
  left: 30px;
}

.next-btn {
  right: 30px;
}

.thumbnail-nav {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 12px;
  background: linear-gradient(135deg, 
    rgba(0, 0, 0, 0.8), 
    rgba(17, 24, 39, 0.7));
  padding: 12px 20px;
  border-radius: 35px;
  backdrop-filter: blur(20px) saturate(1.5);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 
    0 15px 50px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  max-width: 90%;
  overflow-x: auto;
}

.thumbnail-item {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid transparent;
  flex-shrink: 0;
  position: relative;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
}

.thumbnail-item::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 12px;
  background: linear-gradient(135deg, 
    rgba(74, 140, 255, 0.2), 
    rgba(138, 105, 255, 0.2));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.thumbnail-item.active {
  border-color: #4a8cff;
  transform: scale(1.15);
  box-shadow: 
    0 8px 25px rgba(74, 140, 255, 0.4),
    0 0 20px rgba(74, 140, 255, 0.3);
}

.thumbnail-item.active::before {
  opacity: 1;
}

.thumbnail-item:hover {
  transform: scale(1.08);
  border-color: rgba(74, 140, 255, 0.5);
  box-shadow: 
    0 6px 20px rgba(74, 140, 255, 0.3),
    0 0 15px rgba(74, 140, 255, 0.2);
}

.thumbnail-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.toolbar {
  position: absolute;
  bottom: 20px;
  right: 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(135deg, 
    rgba(0, 0, 0, 0.8), 
    rgba(17, 24, 39, 0.7));
  padding: 12px 18px;
  border-radius: 35px;
  backdrop-filter: blur(20px) saturate(1.5);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 
    0 15px 50px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.tool-btn {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, 
    rgba(255, 255, 255, 0.2), 
    rgba(255, 255, 255, 0.1));
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  color: white;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 
    0 4px 15px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.tool-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, 
    rgba(74, 140, 255, 0.8), 
    rgba(138, 105, 255, 0.6));
  border-color: rgba(74, 140, 255, 0.6);
  transform: scale(1.15) rotate(5deg);
  box-shadow: 
    0 8px 25px rgba(74, 140, 255, 0.4),
    0 0 20px rgba(74, 140, 255, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.tool-btn:disabled {
  opacity: 0.2;
  cursor: not-allowed;
  transform: none;
}

.zoom-level {
  color: white;
  font-size: 14px;
  font-weight: 600;
  min-width: 55px;
  text-align: center;
  background: linear-gradient(135deg, 
    rgba(74, 140, 255, 0.3), 
    rgba(138, 105, 255, 0.2));
  border-radius: 20px;
  padding: 6px 12px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  backdrop-filter: blur(10px);
  box-shadow: 
    0 2px 10px rgba(0, 0, 0, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .thumbnail-nav {
    bottom: 100px;
    max-width: 95%;
  }
  
  .toolbar {
    bottom: 20px;
    right: 50%;
    transform: translateX(50%);
  }
  
  .nav-btn {
    width: 50px;
    height: 50px;
    font-size: 20px;
  }
  
  .prev-btn {
    left: 20px;
  }
  
  .next-btn {
    right: 20px;
  }
  
  .thumbnail-item {
    width: 50px;
    height: 50px;
  }
}
</style> 