<template>
  <div class="group-settings-modal" @click.self="$emit('close')">
    <div class="settings-card">
      <button class="settings-btn" @click="$emit('group-settings')">群设置</button>
      <button class="settings-btn" @click="$emit('invite-member')">邀请成员</button>
      <button class="settings-btn danger" @click="$emit('quit-group')">退出群聊</button>
    </div>
  </div>
</template>
<script setup>
const emit = defineEmits(['close', 'group-settings', 'invite-member', 'quit-group'])
</script>
<style scoped>
.group-settings-modal {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.15);
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  z-index: 1000;
}
.settings-card {
  margin: 32px 48px 0 0;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.settings-btn {
  padding: 10px 0;
  border: none;
  border-radius: 8px;
  background: #f1f5f9;
  color: #333;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}
.settings-btn:hover {
  background: #e0e7ff;
}
.settings-btn.danger {
  background: #ffe4e6;
  color: #e53e3e;
}
.settings-btn.danger:hover {
  background: #fecaca;
}
</style> 