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
/* Transition */
.image-viewer-enter-active {
  transition: all 0.3s ease;
}

.image-viewer-leave-active {
  transition: all 0.25s ease;
}

.image-viewer-enter-from {
  opacity: 0;
  transform: scale(0.95);
}

.image-viewer-leave-to {
  opacity: 0;
  transform: scale(1.02);
}

.image-viewer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.95);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Particles and glow removed for flat dark theme */
.background-particles,
.ambient-glow {
  display: none;
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
  width: 44px;
  height: 44px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  color: var(--muted);
  font-size: 22px;
  cursor: pointer;
  z-index: 10;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: rgba(192, 57, 43, 0.15);
  border-color: var(--danger);
  color: var(--danger);
}

.image-counter {
  position: absolute;
  top: 20px;
  left: 20px;
  background: var(--card);
  color: var(--muted);
  padding: 8px 16px;
  border-radius: var(--radius-full);
  font-size: 14px;
  font-family: var(--font-ui);
  font-weight: 600;
  z-index: 10;
  border: 1px solid var(--border);
}

.image-container {
  position: relative;
  max-width: 90%;
  max-height: 90%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: var(--radius-md);
}

.main-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  transition: transform 0.3s ease;
  user-select: none;
  -webkit-user-drag: none;
  border-radius: var(--radius-md);
}

.loading-spinner {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.spinner {
  width: 48px;
  height: 48px;
  border: 3px solid var(--border);
  border-top: 3px solid var(--accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 50px;
  height: 50px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  color: var(--fg);
  font-size: 22px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-btn:hover:not(:disabled) {
  background: var(--accent-muted);
  border-color: var(--accent);
  color: var(--accent);
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
  gap: 10px;
  background: var(--card);
  padding: 10px 16px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  max-width: 90%;
  overflow-x: auto;
}

.thumbnail-item {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 2px solid transparent;
  flex-shrink: 0;
}

.thumbnail-item::before {
  display: none;
}

.thumbnail-item.active {
  border-color: var(--accent);
}

.thumbnail-item:hover {
  border-color: var(--border);
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
  gap: 10px;
  background: var(--card);
  padding: 10px 16px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
}

.tool-btn {
  width: 36px;
  height: 36px;
  background: var(--input-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  color: var(--muted);
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tool-btn:hover:not(:disabled) {
  background: var(--accent-muted);
  border-color: var(--accent);
  color: var(--accent);
}

.tool-btn:disabled {
  opacity: 0.2;
  cursor: not-allowed;
}

.zoom-level {
  color: var(--muted);
  font-size: 13px;
  font-family: var(--font-ui);
  font-weight: 600;
  min-width: 50px;
  text-align: center;
}

/* Responsive */
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
    width: 44px;
    height: 44px;
    font-size: 20px;
  }

  .prev-btn {
    left: 16px;
  }

  .next-btn {
    right: 16px;
  }

  .thumbnail-item {
    width: 48px;
    height: 48px;
  }
}
</style> 