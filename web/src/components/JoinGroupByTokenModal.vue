<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal-card">
      <h3>Join Group by Invite Token</h3>

      <div class="form-item">
        <label>Token</label>
        <input v-model="token" placeholder="Paste invite token" />
      </div>

      <div class="form-item">
        <label>Message (optional)</label>
        <input v-model="reqMsg" placeholder="Say something to admins" />
      </div>

      <div class="actions">
        <button class="btn" @click="$emit('close')">Cancel</button>
        <button class="btn primary" :disabled="!token || loading" @click="submit">
          {{ loading ? 'Joining...' : 'Join' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useContactsStore } from '../stores/contacts'

const emit = defineEmits(['close', 'success'])

const contactsStore = useContactsStore()
const token = ref('')
const reqMsg = ref('')
const loading = ref(false)

const submit = async () => {
  if (!token.value) return
  loading.value = true
  try {
    const resp = await contactsStore.joinGroupByToken(token.value.trim(), reqMsg.value)
    emit('success', resp)
    emit('close')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1200;
}
.modal-card {
  width: 420px;
  max-width: calc(100vw - 32px);
  background: #fff;
  border-radius: 14px;
  padding: 20px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.18);
}
h3 { margin: 0 0 14px; font-size: 16px; }
.form-item { display: flex; flex-direction: column; gap: 8px; margin-bottom: 12px; }
label { font-size: 13px; color: #64748b; }
input {
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  outline: none;
}
.actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 8px; }
.btn { border: none; padding: 10px 14px; border-radius: 10px; cursor: pointer; background: #f1f5f9; }
.btn.primary { background: #4a8cff; color: #fff; }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>


