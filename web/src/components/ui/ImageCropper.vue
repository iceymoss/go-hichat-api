<template>
  <div class="cropper-overlay">
    <div class="cropper-container">
      <div class="cropper-header">
        <h3>裁切头像</h3>
        <button class="close-btn" @click="$emit('cancel')"><i class="icon icon-x"></i></button>
      </div>
      
      <div class="cropper-body">
        <div class="crop-area" 
             @mousedown="startDrag" 
             @touchstart.prevent="startDrag"
             @wheel.prevent="handleWheel">
          <canvas ref="canvasRef" width="300" height="300"></canvas>
          <div class="crop-mask"></div>
          <div class="crop-circle"></div>
        </div>
        
        <div class="controls">
          <i class="icon icon-minus-circle" @click="zoomOut"></i>
          <input type="range" v-model.number="scale" min="0.1" max="3" step="0.01" @input="draw" />
          <i class="icon icon-plus-circle" @click="zoomIn"></i>
        </div>
        
        <p class="hint">滚动缩放，拖动调整位置</p>
      </div>
      
      <div class="cropper-footer">
        <button class="btn-cancel" @click="$emit('cancel')">取消</button>
        <button class="btn-confirm" @click="confirmCrop">确定</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'

const props = defineProps({
  file: { type: File, required: true }
})

const emit = defineEmits(['cancel', 'confirm'])

const canvasRef = ref(null)
const image = new Image()
const scale = ref(1)
const position = ref({ x: 0, y: 0 })
const isDragging = ref(false)
const lastMousePos = ref({ x: 0, y: 0 })

const CANVAS_SIZE = 300

onMounted(() => {
  if (props.file) {
    loadFile(props.file)
  }
})

watch(() => props.file, (newFile) => {
  if (newFile) loadFile(newFile)
})

const loadFile = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    image.src = e.target.result
    image.onload = () => {
      // Init scale to fit
      const minScale = Math.max(CANVAS_SIZE / image.width, CANVAS_SIZE / image.height)
      scale.value = minScale
      // Center image
      position.value = {
        x: (CANVAS_SIZE - image.width * scale.value) / 2,
        y: (CANVAS_SIZE - image.height * scale.value) / 2
      }
      draw()
    }
  }
  reader.readAsDataURL(file)
}

const draw = () => {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  
  // Clear
  ctx.clearRect(0, 0, CANVAS_SIZE, CANVAS_SIZE)
  
  // Draw background (optional)
  ctx.fillStyle = '#000'
  ctx.fillRect(0, 0, CANVAS_SIZE, CANVAS_SIZE)
  
  // Draw Image
  if (image.src) {
    ctx.drawImage(
      image, 
      position.value.x, 
      position.value.y, 
      image.width * scale.value, 
      image.height * scale.value
    )
  }
}

const startDrag = (e) => {
  isDragging.value = true
  const clientX = e.type.startsWith('touch') ? e.touches[0].clientX : e.clientX
  const clientY = e.type.startsWith('touch') ? e.touches[0].clientY : e.clientY
  lastMousePos.value = { x: clientX, y: clientY }
  
  window.addEventListener('mousemove', onDrag)
  window.addEventListener('mouseup', stopDrag)
  window.addEventListener('touchmove', onDrag)
  window.addEventListener('touchend', stopDrag)
}

const onDrag = (e) => {
  if (!isDragging.value) return
  const clientX = e.type.startsWith('touch') ? e.touches[0].clientX : e.clientX
  const clientY = e.type.startsWith('touch') ? e.touches[0].clientY : e.clientY
  
  const deltaX = clientX - lastMousePos.value.x
  const deltaY = clientY - lastMousePos.value.y
  
  position.value.x += deltaX
  position.value.y += deltaY
  
  lastMousePos.value = { x: clientX, y: clientY }
  draw()
}

const stopDrag = () => {
  isDragging.value = false
  window.removeEventListener('mousemove', onDrag)
  window.removeEventListener('mouseup', stopDrag)
  window.removeEventListener('touchmove', onDrag)
  window.removeEventListener('touchend', stopDrag)
}

const handleWheel = (e) => {
  const delta = e.deltaY > 0 ? -0.1 : 0.1
  const newScale = Math.max(0.1, Math.min(3, scale.value + delta))
  // Zoom towards center logic could be added here, simplified for now
  scale.value = newScale
  draw()
}

const zoomIn = () => {
  scale.value = Math.min(3, scale.value + 0.1)
  draw()
}

const zoomOut = () => {
  scale.value = Math.max(0.1, scale.value - 0.1)
  draw()
}

const confirmCrop = () => {
  // Create a new canvas for the result
  const resultCanvas = document.createElement('canvas')
  resultCanvas.width = CANVAS_SIZE
  resultCanvas.height = CANVAS_SIZE
  const ctx = resultCanvas.getContext('2d')
  
  // Draw the visible part
  if (image.src) {
    ctx.drawImage(
      image, 
      position.value.x, 
      position.value.y, 
      image.width * scale.value, 
      image.height * scale.value
    )
  }
  
  resultCanvas.toBlob((blob) => {
    emit('confirm', blob)
  }, 'image/jpeg', 0.9)
}
</script>

<style scoped>
.cropper-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.6);
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cropper-container {
  background: #fff;
  border-radius: 12px;
  width: 340px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0,0,0,0.2);
}
.cropper-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #eee;
}
.cropper-header h3 { margin: 0; font-size: 16px; }
.close-btn { border: none; background: none; cursor: pointer; padding: 4px; }
.cropper-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.crop-area {
  width: 300px;
  height: 300px;
  position: relative;
  background: #333;
  overflow: hidden;
  cursor: move;
}
.crop-mask {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5);
  pointer-events: none;
}
/* Cutout the circle from the mask visually - using radial gradient on mask is better approach but complex. 
   Simple way: use a huge border on the circle. */
.crop-circle {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  border: 1px solid rgba(255,255,255,0.5);
  border-radius: 50%;
  pointer-events: none;
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.5); /* The "mask" */
}

.controls {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 16px;
  width: 100%;
}
.controls input { flex: 1; }
.hint { font-size: 12px; color: #888; margin-top: 8px; }

.cropper-footer {
  padding: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #f9f9f9;
}
.btn-cancel, .btn-confirm {
  padding: 8px 20px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 14px;
}
.btn-cancel { background: #eee; color: #333; }
.btn-confirm { background: #4a8cff; color: #fff; }
</style>

