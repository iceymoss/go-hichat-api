<template>
  <div class="feed-viewer-overlay" @click.self="close">
    <div class="feed-viewer">
      <button class="close-btn" @click="close">
        <i class="icon icon-close"></i>
      </button>
      
      <div class="image-container">
        <img :src="currentImage" alt="动态图片" class="viewer-image">
        
        <button class="nav-btn prev" v-if="hasPrev" @click="prevImage">
          <i class="icon icon-arrow-left"></i>
        </button>
        
        <button class="nav-btn next" v-if="hasNext" @click="nextImage">
          <i class="icon icon-arrow-right"></i>
        </button>
      </div>
      
      <div class="viewer-info">
        <div class="image-counter">{{ currentIndex + 1 }} / {{ totalImages }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useFeedStore } from '../stores/feed'

const feedStore = useFeedStore()

const totalImages = computed(() => feedStore.viewingImages.length)
const currentIndex = computed(() => feedStore.currentImageIndex)
const currentImage = computed(() => feedStore.viewingImages[currentIndex.value])
const hasPrev = computed(() => currentIndex.value > 0)
const hasNext = computed(() => currentIndex.value < totalImages.value - 1)

const prevImage = () => {
  if (hasPrev.value) {
    feedStore.currentImageIndex--
  }
}

const nextImage = () => {
  if (hasNext.value) {
    feedStore.currentImageIndex++
  }
}

const close = () => {
  feedStore.closeImageViewer()
}
</script>

<style scoped>
.feed-viewer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.feed-viewer {
  position: relative;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
}

.close-btn {
  position: absolute;
  top: -40px;
  right: 0;
  background: none;
  border: none;
  color: var(--fg);
  font-size: 28px;
  cursor: pointer;
  padding: 5px;
  transition: color 0.2s;
}

.close-btn:hover {
  color: var(--accent);
}

.image-container {
  position: relative;
  width: 100%;
  height: 80vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.viewer-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: var(--radius-sm);
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 50px;
  height: 50px;
  border-radius: var(--radius-full);
  background-color: var(--accent-muted);
  border: 1px solid var(--border);
  color: var(--fg);
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.nav-btn:hover {
  background-color: var(--accent);
  color: var(--bg);
  border-color: var(--accent);
}

.prev {
  left: 20px;
}

.next {
  right: 20px;
}

.viewer-info {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  text-align: center;
}

.image-counter {
  color: var(--fg);
  font-size: 16px;
  font-family: var(--font-ui);
  background-color: var(--card);
  border: 1px solid var(--border);
  padding: 5px 15px;
  border-radius: var(--radius-full);
  display: inline-block;
}
</style>