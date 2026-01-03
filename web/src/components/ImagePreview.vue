<template>
  <Teleport to="body">
    <Transition name="image-preview">
      <div v-if="visible" class="image-preview-overlay" @click="handleOverlayClick">
        <div class="image-preview-container">
          <!-- 工具栏 -->
          <div class="preview-toolbar">
            <div class="toolbar-left">
              <span class="image-info">{{ currentIndex + 1 }} / {{ images.length }}</span>
              <span class="image-name">{{ currentImage?.name || '图片' }}</span>
            </div>
            <div class="toolbar-right">
              <button class="tool-btn" @click="zoomOut" title="缩小">
                <i class="icon icon-zoom-out"></i>
              </button>
              <span class="zoom-level">{{ Math.round(scale * 100) }}%</span>
              <button class="tool-btn" @click="zoomIn" title="放大">
                <i class="icon icon-zoom-in"></i>
              </button>
              <button class="tool-btn" @click="rotate" title="旋转">
                <i class="icon icon-rotate"></i>
              </button>
              <button class="tool-btn" @click="resetView" title="重置">
                <i class="icon icon-refresh"></i>
              </button>
              <button class="tool-btn" @click="downloadImage" title="下载">
                <i class="icon icon-download"></i>
              </button>
              <button class="tool-btn close-btn" @click="close" title="关闭">
                <i class="icon icon-x"></i>
              </button>
            </div>
          </div>
          
          <!-- 图片容器 -->
          <div class="image-container" ref="imageContainer">
            <img 
              ref="previewImage"
              :src="currentImage?.src || currentImage"
              :alt="currentImage?.name || '图片'"
              class="preview-image"
              :style="imageStyle"
              @load="handleImageLoad"
              @error="handleImageError"
              @wheel="handleWheel"
              @mousedown="handleMouseDown"
            />
          </div>
          
          <!-- 导航按钮 -->
          <div v-if="images.length > 1" class="navigation">
            <button 
              class="nav-btn prev-btn" 
              @click="previousImage"
              :disabled="currentIndex === 0"
            >
              <i class="icon icon-chevron-left"></i>
            </button>
            <button 
              class="nav-btn next-btn" 
              @click="nextImage"
              :disabled="currentIndex === images.length - 1"
            >
              <i class="icon icon-chevron-right"></i>
            </button>
          </div>
          
          <!-- 缩略图列表 -->
          <div v-if="images.length > 1" class="thumbnail-list">
            <div 
              v-for="(image, index) in images" 
              :key="index"
              class="thumbnail-item"
              :class="{ 'active': index === currentIndex }"
              @click="selectImage(index)"
            >
              <img :src="image.src || image" :alt="`缩略图 ${index + 1}`" />
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

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

const emit = defineEmits(['close', 'change'])

// 响应式数据
const imageContainer = ref(null)
const previewImage = ref(null)
const currentIndex = ref(0)
const scale = ref(1)
const rotation = ref(0)
const translateX = ref(0)
const translateY = ref(0)
const isDragging = ref(false)
const lastMouseX = ref(0)
const lastMouseY = ref(0)
const imageLoaded = ref(false)

// 计算属性
const currentImage = computed(() => {
  return props.images[currentIndex.value]
})

const imageStyle = computed(() => {
  return {
    transform: `translate(${translateX.value}px, ${translateY.value}px) scale(${scale.value}) rotate(${rotation.value}deg)`,
    transition: isDragging.value ? 'none' : 'transform 0.3s ease'
  }
})

// 方法
function close() {
  emit('close')
}

function handleOverlayClick(event) {
  if (event.target === event.currentTarget) {
    close()
  }
}

function zoomIn() {
  scale.value = Math.min(scale.value * 1.2, 5)
}

function zoomOut() {
  scale.value = Math.max(scale.value / 1.2, 0.1)
}

function rotate() {
  rotation.value += 90
  if (rotation.value >= 360) {
    rotation.value = 0
  }
}

function resetView() {
  scale.value = 1
  rotation.value = 0
  translateX.value = 0
  translateY.value = 0
}

function previousImage() {
  if (currentIndex.value > 0) {
    currentIndex.value--
    resetView()
    emit('change', currentIndex.value)
  }
}

function nextImage() {
  if (currentIndex.value < props.images.length - 1) {
    currentIndex.value++
    resetView()
    emit('change', currentIndex.value)
  }
}

function selectImage(index) {
  currentIndex.value = index
  resetView()
  emit('change', index)
}

function handleImageLoad() {
  imageLoaded.value = true
}

function handleImageError() {
  console.error('图片加载失败')
}

function handleWheel(event) {
  event.preventDefault()
  const delta = event.deltaY
  if (delta > 0) {
    zoomOut()
  } else {
    zoomIn()
  }
}

function handleMouseDown(event) {
  if (event.button !== 0) return // 只处理左键
  
  isDragging.value = true
  lastMouseX.value = event.clientX
  lastMouseY.value = event.clientY
  
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
  event.preventDefault()
}

function handleMouseMove(event) {
  if (!isDragging.value) return
  
  const deltaX = event.clientX - lastMouseX.value
  const deltaY = event.clientY - lastMouseY.value
  
  translateX.value += deltaX
  translateY.value += deltaY
  
  lastMouseX.value = event.clientX
  lastMouseY.value = event.clientY
}

function handleMouseUp() {
  isDragging.value = false
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
}

function downloadImage() {
  const currentImg = currentImage.value
  if (!currentImg) return
  
  const link = document.createElement('a')
  link.href = currentImg.src || currentImg
  link.download = currentImg.name || `image_${Date.now()}.jpg`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// 键盘事件处理
function handleKeydown(event) {
  if (!props.visible) return
  
  switch (event.key) {
    case 'Escape':
      close()
      break
    case 'ArrowLeft':
      previousImage()
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
    case 'r':
    case 'R':
      rotate()
      break
    case '0':
      resetView()
      break
  }
}

// 生命周期
onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
})

// 监听初始索引变化
watch(() => props.initialIndex, (newIndex) => {
  currentIndex.value = newIndex
  resetView()
}, { immediate: true })

// 监听可见性变化
watch(() => props.visible, (visible) => {
  if (visible) {
    resetView()
    // 禁止页面滚动
    document.body.style.overflow = 'hidden'
  } else {
    // 恢复页面滚动
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.9);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  backdrop-filter: blur(10px);
}

.image-preview-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  position: relative;
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
  color: white;
}

.image-info {
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
}

.image-name {
  font-size: 16px;
  font-weight: 500;
  color: white;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-btn {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.tool-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.05);
}

.close-btn:hover {
  background: rgba(239, 68, 68, 0.8);
}

.zoom-level {
  color: white;
  font-size: 14px;
  font-weight: 500;
  min-width: 50px;
  text-align: center;
}

.image-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
  cursor: grab;
}

.image-container:active {
  cursor: grabbing;
}

.preview-image {
  max-width: 90vw;
  max-height: 80vh;
  object-fit: contain;
  user-select: none;
  pointer-events: auto;
}

.navigation {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  left: 20px;
  right: 20px;
  display: flex;
  justify-content: space-between;
  pointer-events: none;
}

.nav-btn {
  width: 50px;
  height: 50px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  pointer-events: auto;
  backdrop-filter: blur(10px);
}

.nav-btn:hover:not(:disabled) {
  background: rgba(0, 0, 0, 0.8);
  transform: scale(1.1);
}

.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.thumbnail-list {
  display: flex;
  gap: 8px;
  padding: 16px 20px;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(20px);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  overflow-x: auto;
  justify-content: center;
}

.thumbnail-item {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 2px solid transparent;
  flex-shrink: 0;
}

.thumbnail-item:hover {
  transform: scale(1.05);
  border-color: rgba(255, 255, 255, 0.3);
}

.thumbnail-item.active {
  border-color: #4a8cff;
  box-shadow: 0 0 0 2px rgba(74, 140, 255, 0.3);
}

.thumbnail-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 过渡动画 */
.image-preview-enter-active,
.image-preview-leave-active {
  transition: all 0.3s ease;
}

.image-preview-enter-from,
.image-preview-leave-to {
  opacity: 0;
  transform: scale(0.9);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .preview-toolbar {
    padding: 12px 16px;
  }
  
  .toolbar-left {
    gap: 12px;
  }
  
  .image-name {
    max-width: 150px;
    font-size: 14px;
  }
  
  .tool-btn {
    width: 36px;
    height: 36px;
    font-size: 16px;
  }
  
  .zoom-level {
    font-size: 12px;
    min-width: 40px;
  }
  
  .nav-btn {
    width: 40px;
    height: 40px;
    font-size: 20px;
  }
  
  .navigation {
    left: 10px;
    right: 10px;
  }
  
  .thumbnail-list {
    padding: 12px 16px;
  }
  
  .thumbnail-item {
    width: 50px;
    height: 50px;
  }
}
</style> 