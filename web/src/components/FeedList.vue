<template>
  <div class="feed-type-list">
    <div class="feed-type-list-header">
      <h2 class="feed-type-title">动态分类</h2>
      <p class="feed-type-subtitle">选择您想查看的动态类型</p>
    </div>
    
    <div class="feed-type-list-inner">
      <div
        v-for="type in types"
        :key="type.key"
        :class="['feed-type-item', { active: activeType === type.key }]"
        @click="selectType(type.key)"
      >
        <div class="feed-type-icon">
          <i :class="['icon', type.icon]"></i>
        </div>
        <div class="feed-type-content">
          <span class="feed-type-label">{{ type.label }}</span>
          <span class="feed-type-desc">{{ type.description }}</span>
        </div>
        <div class="feed-type-indicator">
          <div class="indicator-dot"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const types = [
  { 
    key: 'friend', 
    label: '好友动态', 
    icon: 'icon-friend-feed',
    description: '查看好友的最新动态'
  },
  { 
    key: 'community', 
    label: '社区动态', 
    icon: 'icon-community-feed',
    description: '发现更多精彩内容'
  }
]

const props = defineProps({ modelValue: String })
const emit = defineEmits(['update:modelValue'])
const activeType = ref(props.modelValue || 'friend')

watch(() => props.modelValue, v => { 
  if (v) activeType.value = v 
})

function selectType(key) {
  activeType.value = key
  emit('update:modelValue', key)
}
</script>

<style scoped>
.feed-type-list {
  display: flex;
  flex-direction: column;
  width: 450px;
  min-width: 450px;
  max-width: 450px;
  height: 100%;
  background: var(--bg);
  border-right: 1px solid var(--border);
  overflow: hidden;
  flex: 1 0 0;
  position: relative;
}

.feed-type-list-header {
  padding: 32px 36px 24px 36px;
  border-bottom: 1px solid var(--border);
}

.feed-type-title {
  color: var(--accent);
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 6px 0;
  font-family: var(--font-display);
}

.feed-type-subtitle {
  color: var(--muted);
  font-size: 14px;
  margin: 0;
  font-family: var(--font-ui);
}

.feed-type-list-inner {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 24px 36px;
  flex: 1;
}

.feed-type-item {
  display: flex;
  align-items: center;
  padding: 20px 24px;
  cursor: pointer;
  color: var(--muted);
  font-size: 16px;
  border-radius: var(--radius-md);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 0;
  position: relative;
  background: var(--card);
  border: 1px solid var(--border-light);
  font-family: var(--font-ui);
}

.feed-type-item:hover {
  background: var(--card-hover);
  border-color: var(--border);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0,0,0,0.3);
}

.feed-type-item.active {
  background: var(--accent-muted);
  border-color: var(--accent);
  color: var(--accent);
  font-weight: 600;
  box-shadow: 0 6px 20px rgba(200,149,108,0.15);
  transform: translateY(-2px);
}

.feed-type-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--accent);
  border-radius: 0 2px 2px 0;
}

.feed-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  background: var(--accent-muted);
  margin-right: 16px;
  transition: all 0.3s ease;
}

.feed-type-item.active .feed-type-icon {
  background: var(--accent-muted);
  transform: scale(1.1);
}

.feed-type-item .icon {
  font-size: 20px;
  color: inherit;
}

.feed-type-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.feed-type-label {
  font-size: 16px;
  font-weight: 600;
  color: inherit;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.feed-type-desc {
  font-size: 12px;
  color: var(--muted);
  font-weight: 400;
}

.feed-type-item.active .feed-type-desc {
  color: var(--accent);
  opacity: 0.8;
}

.feed-type-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
}

.indicator-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--muted);
  transition: all 0.3s ease;
}

.feed-type-item.active .indicator-dot {
  background: var(--accent);
  box-shadow: 0 0 10px rgba(200,149,108,0.4);
  transform: scale(1.2);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .feed-type-list {
    width: 400px;
    min-width: 400px;
    max-width: 400px;
  }
}

@media (max-width: 768px) {
  .feed-type-list {
    width: 100%;
    min-width: 100%;
    max-width: 100%;
  }

  .feed-type-list-header,
  .feed-type-list-inner {
    padding-left: 20px;
    padding-right: 20px;
  }
}
</style> 