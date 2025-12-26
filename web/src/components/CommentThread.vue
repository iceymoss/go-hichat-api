<template>
  <div class="comment-thread" @click.stop>
    <div v-for="comment in getDisplayComments()" :key="comment.id" class="comment-item">
      <img 
        :src="getUserAvatar(comment.user)" 
        alt="头像" 
        class="comment-avatar"
        @mouseenter="(e) => showUserProfile(comment.user, e)"
        @mouseleave="hideUserProfile"
      >
      <div class="comment-content-wrapper">
        <div class="comment-main-row">
          <span 
            class="comment-user"
            @mouseenter="(e) => showUserProfile(comment.user, e)"
            @mouseleave="hideUserProfile"
          >{{ comment.user }}</span>
          <span class="comment-time">{{ comment.time || '刚刚' }}</span>
        </div>
        <div class="comment-text" v-html="formatComment(comment.content)"></div>
        <div class="comment-meta-row">
          <div class="comment-actions">
            <span class="comment-action" @click.stop="startReply(comment)">回复</span>
            <span class="comment-action" @click.stop="likeComment(comment)">点赞</span>
            <button v-if="comment.user === currentUser.name" class="comment-delete-btn" @click.stop="confirmDelete(comment, 'comment')">删除</button>
          </div>
        </div>
        <div class="comment-replies" v-if="comment.replies && comment.replies.length">
          <div v-for="reply in getDisplayReplies(comment)" :key="reply.id" class="reply-item">
            <img 
              :src="getUserAvatar(reply.user)" 
              alt="头像" 
              class="reply-avatar-small"
              @mouseenter="(e) => showUserProfile(reply.user, e)"
              @mouseleave="hideUserProfile"
            >
            <div class="reply-content-wrapper">
              <div class="reply-main-row">
                <span 
                  class="reply-user"
                  @mouseenter="(e) => showUserProfile(reply.user, e)"
                  @mouseleave="hideUserProfile"
                >{{ reply.user }}</span>
                <span class="reply-time">{{ reply.time || '刚刚' }}</span>
              </div>
              <div class="reply-text" v-html="formatComment(reply.content)"></div>
              <div class="reply-meta-row">
                <div class="reply-actions">
                  <span class="reply-action" @click.stop="startReply(comment, reply)">回复</span>
                  <span class="reply-action" @click.stop="likeReply(reply)">点赞</span>
                  <button v-if="reply.user === currentUser.name" class="reply-delete-btn" @click.stop="confirmDelete(reply, 'reply', comment)">删除</button>
                </div>
              </div>
            </div>
          </div>
          <div v-if="(comment.replies.length > (replyDisplayCount[comment.id] ?? replyDefaultCount))" class="more-replies" @click.stop="showMoreReplies(comment)">
            加载更多回复
          </div>
          <div v-else-if="comment.replies.length > replyDefaultCount && (replyDisplayCount[comment.id] ?? replyDefaultCount) > replyDefaultCount" class="more-replies" @click.stop="collapseReplies(comment)">
            收起回复
          </div>
        </div>
      </div>
      <div v-if="replying === comment.id" class="reply-box" @click.stop>
        <img :src="currentUser.avatar" alt="头像" class="reply-avatar">
        <div class="reply-input-wrapper">
          <input 
            v-model="replyContent" 
            :placeholder="`回复 @${comment.user}`" 
            @keyup.enter="submitReply(comment)"
            @click.stop
            class="reply-input"
          />
          <button @click.stop="submitReply(comment)" class="reply-submit-btn">发送</button>
        </div>
      </div>
    </div>
    <!-- 主评论加载更多按钮 -->
    <div v-if="(props.comments || []).length > commentDisplayCount" class="more-comments" @click.stop="showMoreComments">
      加载更多评论 ({{ (props.comments || []).length - commentDisplayCount }}条)
    </div>
    <div v-else-if="(props.comments || []).length > commentDefaultCount && commentDisplayCount >= (props.comments || []).length" class="more-comments" @click.stop="collapseComments">
      收起评论
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useAuthStore } from '../stores/auth'

const props = defineProps({ 
  comments: Array, 
  formatComment: Function,
  showUserProfile: Function,
  hideUserProfile: Function
})
const emit = defineEmits(['reply', 'like-comment', 'like-reply', 'delete-comment', 'delete-reply'])

const authStore = useAuthStore()
const currentUser = computed(() => authStore.user || { 
  avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=You',
  name: '我'
})

const replying = ref(null)
const replyContent = ref('')
const format = computed(() => props.formatComment || ((c) => c))

// 主评论分页
const commentDefaultCount = 10
const commentPageCount = 10
const commentDisplayCount = ref(commentDefaultCount)
watch(() => props.comments, () => { commentDisplayCount.value = commentDefaultCount })
function getDisplayComments() {
  return (props.comments || []).slice(0, commentDisplayCount.value)
}
function showMoreComments() {
  commentDisplayCount.value = Math.min(commentDisplayCount.value + commentPageCount, (props.comments || []).length)
}
function collapseComments() {
  commentDisplayCount.value = commentDefaultCount
}

// 子评论分页状态
const replyDisplayCount = ref({})
const replyDefaultCount = 2
const replyPageCount = 2

function getDisplayReplies(comment) {
  const count = replyDisplayCount.value[comment.id] ?? replyDefaultCount
  return (comment.replies || []).slice(0, count)
}

function showMoreReplies(comment) {
  if (!comment.replies || comment.replies.length === 0) return
  
  const current = replyDisplayCount.value[comment.id] ?? replyDefaultCount
  const next = Math.min(current + replyPageCount, comment.replies.length)
  
  if (next > current) {
    replyDisplayCount.value[comment.id] = next
  }
}

function collapseReplies(comment) {
  replyDisplayCount.value[comment.id] = replyDefaultCount
}

function getUserAvatar(userName) {
  return `https://api.dicebear.com/7.x/personas/svg?seed=${encodeURIComponent(userName)}`
}

function formatComment(content) {
  return format.value(content)
}

function showUserProfile(userName, event) {
  if (props.showUserProfile) {
    // 构建用户对象
    const user = {
      name: userName,
      avatar: getUserAvatar(userName)
    }
    props.showUserProfile(user, event)
  }
}

function hideUserProfile() {
  if (props.hideUserProfile) {
    props.hideUserProfile()
  }
}

function startReply(comment, reply = null) {
  replying.value = comment.id
  replyContent.value = reply ? `@${reply.user} ` : ''
}

function submitReply(comment) {
  if (!replyContent.value.trim()) return
  
  const newReply = {
    id: Date.now(),
    user: currentUser.value.name,
    content: replyContent.value,
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    avatar: currentUser.value.avatar
  }
  
  emit('reply', { comment, reply: newReply })
  replyContent.value = ''
  replying.value = null
  
  if (comment.replies && comment.replies.length > (replyDisplayCount.value[comment.id] ?? replyDefaultCount)) {
    replyDisplayCount.value[comment.id] = Math.min(
      (replyDisplayCount.value[comment.id] ?? replyDefaultCount) + 1,
      comment.replies.length
    )
  }
}

function likeComment(comment) {
  emit('like-comment', comment)
}

function likeReply(reply) {
  emit('like-reply', reply)
}

function confirmDelete(commentOrReply, type, parentComment = null) {
  if (window.confirm('确定要删除这条评论吗？')) {
    if (type === 'comment') {
      emit('delete-comment', commentOrReply)
      if (commentOrReply.replies && commentOrReply.replies.length > 0) {
        replyDisplayCount.value[commentOrReply.id] = replyDefaultCount
      }
    } else if (type === 'reply') {
      emit('delete-reply', { comment: parentComment, reply: commentOrReply })
    }
  }
}
</script>

<style scoped>
.comment-thread {
  margin-top: 8px;
}

.comment-item {
  display: flex;
  align-items: flex-start;
  padding: 10px 0 6px 0;
  border-bottom: 1px solid #f3f3f3;
  background: none;
  border-radius: 0;
  box-shadow: none;
  margin-bottom: 0;
  position: relative;
  min-height: 36px;
}
.comment-item:last-child {
  border-bottom: none;
}
.comment-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  margin: 0 10px 0 0;
  flex-shrink: 0;
  box-shadow: none;
  align-self: flex-start;
  cursor: pointer;
  transition: all 0.3s ease;
}

.comment-avatar:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(74, 140, 255, 0.3);
  border: 2px solid rgba(74, 140, 255, 0.5);
}
.comment-content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  padding: 0;
}
.comment-main-row {
  display: flex;
  align-items: flex-start;
  gap: 6px;
}
.comment-user {
  color: #222;
  font-weight: 600;
  font-size: 14px;
  margin-right: 2px;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 2px 4px;
  border-radius: 4px;
  margin: -2px 2px -2px -4px;
}

.comment-user:hover {
  color: #4a8cff;
  background: rgba(74, 140, 255, 0.1);
  transform: scale(1.05);
}
.comment-text {
  color: #333;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-all;
  flex: 1;
}
.comment-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 2px;
}
.comment-time {
  color: #b0b0b0;
  font-size: 12px;
}
.comment-actions {
  display: flex;
  gap: 10px;
}
.comment-action {
  color: #b0b0b0;
  font-size: 12px;
  cursor: pointer;
  border-radius: 4px;
  padding: 2px 6px;
  transition: color 0.2s;
}
.comment-action:hover {
  color: #2563eb;
  background: none;
}
.comment-delete-btn {
  margin-left: 8px;
  background: none;
  border: none;
  color: #f87171;
  font-size: 12px;
  cursor: pointer;
  border-radius: 4px;
  padding: 2px 6px;
  transition: color 0.2s;
}
.comment-delete-btn:hover {
  color: #d32f2f;
  background: none;
}
.comment-replies {
  margin-left: 36px;
  margin-top: 0;
  border-left: 1.5px solid #f0f0f0;
  padding-left: 10px;
}
.reply-item {
  display: flex;
  align-items: flex-start;
  background: none;
  border-radius: 0;
  margin-bottom: 0;
  padding: 2px 0 2px 0;
  border-bottom: 1px solid #f3f3f3;
  position: relative;
}
.reply-item:last-child {
  border-bottom: none;
}
.reply-avatar-small {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  margin: 2px 8px 0 0;
  flex-shrink: 0;
  box-shadow: none;
  cursor: pointer;
  transition: all 0.3s ease;
}

.reply-avatar-small:hover {
  transform: scale(1.15);
  box-shadow: 0 3px 10px rgba(74, 140, 255, 0.3);
  border: 2px solid rgba(74, 140, 255, 0.5);
}
.reply-content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.reply-main-row {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}
.reply-user {
  color: #2563eb;
  font-weight: 600;
  font-size: 13px;
  margin-right: 2px;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 2px 4px;
  border-radius: 4px;
  margin: -2px 2px -2px -4px;
}

.reply-user:hover {
  color: #8a69ff;
  background: rgba(138, 105, 255, 0.1);
  transform: scale(1.05);
}
.reply-text {
  color: #444;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-all;
  flex: 1;
  margin-left: 0;
}
.reply-meta-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: 1px;
  gap: 8px;
}
.reply-time {
  color: #b0b0b0;
  font-size: 11px;
  margin-left: 8px;
}
.reply-actions {
  display: flex;
  gap: 8px;
}
.reply-action {
  color: #b0b0b0;
  font-size: 11px;
  cursor: pointer;
  border-radius: 4px;
  padding: 2px 6px;
  transition: color 0.2s;
}
.reply-action:hover {
  color: #2563eb;
  background: none;
}
.reply-delete-btn {
  margin-left: 8px;
  background: none;
  border: none;
  color: #f87171;
  font-size: 11px;
  cursor: pointer;
  border-radius: 4px;
  padding: 2px 6px;
  transition: color 0.2s;
}
.reply-delete-btn:hover {
  color: #d32f2f;
  background: none;
}

.reply-box {
  margin: 12px 0 0 0;
  display: flex;
  gap: 8px;
}
.reply-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}
.reply-input-wrapper {
  flex: 1;
  display: flex;
  gap: 8px;
}
.reply-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 16px;
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s ease;
}
.reply-input:focus {
  border-color: #2563eb;
}
.reply-submit-btn {
  padding: 6px 12px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  border: none;
  border-radius: 16px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.reply-submit-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(37,99,235,0.3);
}

.at-user {
  color: #2563eb;
  font-weight: 600;
}

.more-replies {
  color: #2563eb;
  font-size: 12px;
  cursor: pointer;
  margin: 4px 0 4px 36px;
  user-select: none;
  transition: color 0.2s;
  display: inline-block;
}
.more-replies:hover {
  text-decoration: underline;
  color: #1d4ed8;
}

.more-comments {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 20px;
  margin: 16px 0;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  color: #2563eb;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}

.more-comments:hover {
  background: #e0e7ef;
  color: #1d4ed8;
  border-color: #d1d5db;
}

.more-comments:active {
  transform: translateY(1px);
}
</style> 