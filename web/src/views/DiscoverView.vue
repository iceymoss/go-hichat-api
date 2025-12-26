<template>
  <div class="discover-view">
    <div class="discover-container">
      <!-- 左侧：动态分类列表 -->
      <FeedList v-model="activeFeedType" />
      
      <!-- 中间：动态列表 -->
      <FeedTypeList 
        :feeds="feedData[activeFeedType]" 
        @select="handleFeedSelect" 
        @like="handleLike"
        @comment="handleComment"
        v-if="!selectedFeed" 
        class="feed-list-section"
      />
      
      <!-- 右侧：动态详情 -->
      <FeedDetail 
        :selected="selectedFeed" 
        @back="handleFeedBack" 
        @like="handleDetailLike"
        @comment="handleDetailComment"
        v-else 
        class="feed-detail-section"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import FeedList from '../components/FeedList.vue'
import FeedTypeList from '../components/FeedTypeList.vue'
import FeedDetail from '../components/FeedDetail.vue'
import { useFeedStore } from '../stores/feed'

const feedStore = useFeedStore()

// 使用feed store的数据，改为computed，保证响应式和不为undefined
const feedData = computed(() => ({
  friend: feedStore.getFriendFeeds,
  community: feedStore.getCommunityFeeds
}))

const activeFeedType = ref('friend')
const selectedFeed = ref(null)

// 事件处理函数 - 完全按照feat-login-api分支的逻辑
function handleFeedSelect(feed) {
  selectedFeed.value = feed
}

function handleFeedBack() {
  selectedFeed.value = null
}

function handleLike(feedId) {
  feedStore.toggleLike(feedId, activeFeedType.value)
}

function handleComment(commentData) {
  feedStore.addComment(commentData.feedId, commentData.content, activeFeedType.value)
}

function handleDetailLike(feedId) {
  feedStore.toggleLike(feedId, activeFeedType.value)
}

function handleDetailComment(commentData) {
  feedStore.addComment(commentData.feedId, commentData.content, activeFeedType.value)
}

onMounted(() => {
  // 确保有初始数据
  if (feedStore.posts.length === 0) {
    feedStore.createPost("欢迎使用HiChat 2.0动态功能！", [
      'https://picsum.photos/400/300?random=1'
    ])
  }
})
</script>

<style scoped>
.discover-view {
  height: 100vh;
  min-height: 100vh;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.discover-container {
  display: flex;
  height: 100%;
  width: 100%;
  min-height: 0;
  flex: 1;
}

.feed-list-section {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.feed-detail-section {
  flex: 1;
  height: 100%;
  min-height: 0;
}
</style>