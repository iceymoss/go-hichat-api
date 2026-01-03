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
  background: linear-gradient(135deg, rgba(30,32,40,0.98), rgba(45,47,55,0.95));
  border-right: 1px solid rgba(74,140,255,0.15);
  overflow: hidden;
  flex: 1 0 0;
  position: relative;
}

.feed-type-list::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(74,140,255,0.3), transparent);
}

.feed-type-list-header {
  padding: 32px 36px 24px 36px;
  border-bottom: 1px solid rgba(74,140,255,0.1);
}

.feed-type-title {
  color: #ffffff;
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 6px 0;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.feed-type-subtitle {
  color: #b3cfff;
  font-size: 14px;
  margin: 0;
  opacity: 0.8;
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
  color: #b3cfff;
  font-size: 16px;
  border-radius: 14px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 0;
  position: relative;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(74,140,255,0.05);
  backdrop-filter: blur(10px);
}

.feed-type-item:hover {
  background: rgba(74,140,255,0.08);
  border-color: rgba(74,140,255,0.2);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(74,140,255,0.15);
}

.feed-type-item.active {
  background: linear-gradient(135deg, rgba(74,140,255,0.15), rgba(138,105,255,0.1));
  border-color: rgba(74,140,255,0.3);
  color: #4a8cff;
  font-weight: 600;
  box-shadow: 0 6px 20px rgba(74,140,255,0.2);
  transform: translateY(-2px);
}

.feed-type-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  border-radius: 0 2px 2px 0;
}

.feed-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(74,140,255,0.1);
  margin-right: 16px;
  transition: all 0.3s ease;
}

.feed-type-item.active .feed-type-icon {
  background: linear-gradient(135deg, rgba(74,140,255,0.2), rgba(138,105,255,0.15));
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
  color: rgba(179,207,255,0.7);
  font-weight: 400;
}

.feed-type-item.active .feed-type-desc {
  color: rgba(74,140,255,0.8);
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
  background: rgba(74,140,255,0.3);
  transition: all 0.3s ease;
}

.feed-type-item.active .indicator-dot {
  background: #4a8cff;
  box-shadow: 0 0 10px rgba(74,140,255,0.5);
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