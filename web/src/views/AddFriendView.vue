<template>
  <div class="add-friend-view">
    <div class="search-bar">
      <input v-model="keyword" @keyup.enter="search" placeholder="输入QQ号/手机号/邮箱/昵称搜索用户" />
      <button @click="search">搜索</button>
    </div>
    <div v-if="searching" class="searching">正在搜索...</div>
    <div v-if="result" class="search-result">
      <img :src="result.avatar" class="avatar" />
      <div class="info">
        <div class="name">{{ result.name }}</div>
        <div class="gender">{{ getGenderText(result.gender) }}</div>
        <div class="email" v-if="result.email">{{ result.email }}</div>
        <div class="phone" v-if="result.phone">{{ result.phone }}</div>
        <div class="region" v-if="result.location">{{ result.location }}</div>
      </div>
      <button @click="showAddModal = true">加好友</button>
    </div>
    <AddFriendModal v-if="showAddModal" :user="result" @close="showAddModal = false" />
  </div>
</template>
<script setup>
import { ref } from 'vue'
import AddFriendModal from '../components/AddFriendModal.vue'
const keyword = ref('')
const searching = ref(false)
const result = ref(null)
const showAddModal = ref(false)
const search = () => {
  searching.value = true
  setTimeout(() => {
    // mock search
    result.value = {
      name: '小明',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ming',
      gender: 'male',
      email: 'xiaoming@example.com',
      phone: '13800000005',
      location: '北京'
    }
    searching.value = false
  }, 1000)
}
const getGenderText = g => g === 'male' ? '男' : g === 'female' ? '女' : '其他'
</script>
<style scoped>
.add-friend-view { max-width: 480px; margin: 40px auto; }
.search-bar { display: flex; gap: 8px; margin-bottom: 24px; }
.search-bar input { flex: 1; padding: 8px; border-radius: 6px; border: 1px solid #ddd; }
.search-bar button { padding: 8px 16px; border-radius: 6px; background: #4a8cff; color: #fff; border: none; }
.searching { color: #888; margin-bottom: 16px; }
.search-result { display: flex; align-items: center; gap: 16px; background: #f8fafc; border-radius: 12px; padding: 16px; }
.avatar { width: 56px; height: 56px; border-radius: 12px; }
.info { flex: 1; }
.name { font-size: 18px; font-weight: 600; }
.gender, .email, .phone, .region { font-size: 14px; color: #666; margin-top: 2px; }
.search-result button { background: #4a8cff; color: #fff; border: none; border-radius: 8px; padding: 8px 20px; }
</style> 