<template>
  <div class="modal-overlay" @click.self="close">
    <div class="add-friend-modal">
      <div class="modal-header">
        <h3>添加好友</h3>
        <button class="btn-close" @click="close">
          <i class="icon icon-close"></i>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="search-container">
          <i class="icon icon-search"></i>
          <input 
            type="text" 
            v-model="searchKeyword" 
            placeholder="输入用户账号或昵称" 
            @input="handleSearch"
          >
          <button v-if="searching" class="btn-loading">
            <i class="icon icon-spinner"></i>
          </button>
        </div>
        
        <div v-if="searchResults.length > 0" class="search-results">
          <div class="result-item" v-for="user in searchResults" :key="user.id">
            <img :src="user.avatar" alt="头像" class="avatar">
            <div class="user-info">
              <div class="name">{{ user.name }} 
                <span class="account">@{{ user.account }}</span>
              </div>
              <div class="tags">
                <span v-for="tag in user.tags" :key="tag" class="tag">{{ tag }}</span>
              </div>
            </div>
            <button class="btn-add" @click="sendRequest(user)">
              <i class="icon icon-add"></i>
            </button>
          </div>
        </div>
        
        <div v-else-if="searchKeyword && !searching" class="no-results">
          <p>没有找到用户</p>
          <p class="tip">请检查用户名拼写是否正确</p>
        </div>
        
        <div v-else class="instruction">
          <p>请输入用户账号或昵称进行搜索</p>
          <div class="tips">
            <div class="tip-item">
              <i class="icon icon-tip"></i>
              <span>支持搜索用户名、账号或ID</span>
            </div>
            <div class="tip-item">
              <i class="icon icon-tip"></i>
              <span>对方接受后你们将成为好友</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps(['searchUser'])
const emit = defineEmits(['close', 'search', 'send-request'])

const searchKeyword = ref('')
const searchResults = ref([])
const searching = ref(false)

const close = () => {
  emit('close')
}

const handleSearch = () => {
  if (!searchKeyword.value.trim()) {
    searchResults.value = []
    return
  }
  
  searching.value = true
  props.searchUser(searchKeyword.value).then(results => {
    searchResults.value = results
    searching.value = false
  })
}

const sendRequest = (user) => {
  emit('send-request', user.id)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.add-friend-modal {
  background-color: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #f0f2f5;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
}

.btn-close {
  background: none;
  border: none;
  font-size: 20px;
  color: #999;
  cursor: pointer;
}

.modal-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
  background-color: #f0f2f5;
  border-radius: 10px;
  padding: 10px 15px;
  margin-bottom: 20px;
}

.search-container .icon {
  color: #999;
  margin-right: 10px;
}

.search-container input {
  flex: 1;
  border: none;
  background: none;
  font-size: 14px;
  outline: none;
}

.btn-loading {
  background: none;
  border: none;
  padding: 5px;
  cursor: pointer;
}

.icon-spinner {
  animation: spin 1s infinite linear;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.search-results {
  margin-top: 10px;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid #f0f2f5;
}

.result-item .avatar {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  margin-right: 15px;
}

.user-info {
  flex: 1;
}

.name {
  font-weight: 500;
  margin-bottom: 5px;
}

.account {
  font-weight: normal;
  font-size: 13px;
  color: #999;
}

.tags {
  display: flex;
  gap: 5px;
}

.tag {
  font-size: 12px;
  padding: 3px 8px;
  background-color: #f0f2f5;
  border-radius: 4px;
  color: #666;
}

.btn-add {
  background: none;
  border: none;
  color: #4a8cff;
  font-size: 20px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-add:hover {
  background-color: #eef5ff;
}

.no-results, .instruction {
  text-align: center;
  padding: 30px 20px;
  color: #999;
}

.tip {
  margin-top: 10px;
  font-size: 13px;
}

.tips {
  margin-top: 20px;
  text-align: left;
}

.tip-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.tip-item i {
  margin-right: 8px;
  color: #4a8cff;
}
</style>