<template>
  <div class="group-detail-container">
    <div v-if="!group" class="empty-state">
      <div class="empty-content">
        <i class="icon icon-users"></i>
        <h3>选择群聊</h3>
        <p>从左侧列表中选择一个群聊查看详情</p>
      </div>
    </div>
    <div v-else class="group-detail">
      <!-- 顶部群信息 -->
      <div class="group-top">
        <img :src="group.avatar" class="group-avatar" />
        <div class="group-top-info">
          <div class="group-name">{{ group.name }}</div>
          <div class="group-id">群号：{{ group.id }}</div>
          <div class="group-announcement">公告：{{ group.notice }}</div>
        </div>
        <div class="group-qrcode">
          <i class="icon icon-qrcode"></i>
        </div>
        <!-- 设置按钮 -->
        <button class="btn-settings" @click="toggleMenu" title="设置">
          <i class="icon icon-settings"></i>
        </button>
        <!-- 悬浮设置弹窗 -->
        <div v-if="showMenu" class="popover-menu" @click.self="closeMenu">
          <ul>
            <li @click="openGroupSettings">群设置</li>
            <li @click="openInvite">邀请成员</li>
            <li class="danger" @click="onQuitGroup">退出群聊</li>
          </ul>
        </div>
        <!-- 群设置悬浮弹窗 -->
        <div v-if="showGroupSettings" class="popover-dialog" @click.self="closeGroupSettings">
          <div class="popover-content">
            <h3>群设置</h3>
            <div class="setting-item"><label>群名称</label><input type="text" :value="group.name" disabled /></div>
            <div class="setting-item"><label>群公告</label><input type="text" :value="group.notice" disabled /></div>
            <div class="setting-item"><label>消息免打扰</label><input type="checkbox" /></div>
            <div class="setting-item"><label>置顶聊天</label><input type="checkbox" /></div>
            <button class="popover-close" @click="closeGroupSettings">关闭</button>
          </div>
        </div>
        <!-- 邀请成员悬浮弹窗 -->
        <div v-if="showInvite" class="popover-dialog invite-animate" @click.self="closeInvite">
          <div class="popover-content invite-content">
            <h3>邀请成员</h3>
            <div class="friend-list">
              <div v-for="group in groupedFriends" :key="group.letter" class="friend-group">
                <div class="group-header">{{ group.letter }}</div>
                <div v-for="friend in group.friends" :key="friend.id" class="friend-item">
                  <input type="checkbox" v-model="selectedFriends" :value="friend.id" />
                  <img :src="friend.avatar" class="friend-avatar" />
                  <span>{{ friend.remark || friend.name }}</span>
                </div>
              </div>
            </div>
            <button class="popover-confirm" @click="inviteSelected">邀请</button>
            <button class="popover-close" @click="closeInvite">关闭</button>
          </div>
        </div>
      </div>
      <!-- 群成员区块 -->
      <div class="members-section">
        <div class="members-title">
          <span>群聊成员</span>
          <span class="view-all">查看{{ group.members.length }}名群成员</span>
        </div>
        <div class="members-list">
          <div v-for="m in group.members.slice(0, 12)" :key="m.id" class="member-item">
            <img :src="m.avatar" class="member-avatar" />
            <div class="member-name">{{ m.name }}</div>
            <span v-if="m.role==='owner'" class="role owner">群主</span>
            <span v-else-if="m.role==='admin'" class="role admin">管理员</span>
          </div>
          <div class="member-item invite-item">
            <button class="invite-btn" @click="openInvite">+</button>
            <div class="member-name">邀请</div>
          </div>
        </div>
      </div>
      <!-- 群信息区块 -->
      <div class="group-info-section">
        <div class="info-row"><span class="label">群聊名称</span><span class="value">{{ group.name }}</span></div>
        <div class="info-row"><span class="label">群号</span><span class="value">{{ group.id }}</span></div>
        <div class="info-row"><span class="label">群公告</span><span class="value">{{ group.notice }}</span></div>
        <div class="info-row"><span class="label">我的群昵称</span><span class="value">未设置</span></div>
        <div class="info-row"><span class="label">群聊备注</span><span class="value">未设置</span></div>
      </div>
      <div class="group-apps-section">
        <h4>群应用中心</h4>
        <div class="apps-list">
          <div class="app-item"><i class="icon icon-folder"></i> 文件</div>
          <div class="app-item"><i class="icon icon-image"></i> 相册</div>
          <div class="app-item"><i class="icon icon-star"></i> 精华消息</div>
          <div class="app-item"><i class="icon icon-music"></i> 一起听歌</div>
          <div class="app-item more">更多</div>
        </div>
      </div>
      <div class="group-bot-section">
        <h4>群机器人</h4>
        <div class="bot-list">
          <div class="bot-item">
            <i class="icon icon-robot"></i> 群助手
          </div>
        </div>
      </div>
      <div class="group-chat-settings">
        <div class="setting-row">
          <span>查找聊天记录</span>
          <span class="setting-desc">图片、视频、文件等</span>
        </div>
        <div class="setting-row">
          <span>设为置顶</span>
          <input type="checkbox" />
        </div>
        <div class="setting-row">
          <span>隐藏会话</span>
          <input type="checkbox" />
        </div>
        <div class="setting-row">
          <span>消息免打扰</span>
          <input type="checkbox" />
        </div>
        <div class="setting-row">
          <span>群通知设置</span>
          <span class="setting-desc">消息预览、提示音等</span>
        </div>
      </div>
      <div class="group-personal-section">
        <div class="setting-row">个性设置</div>
        <div class="setting-row">删除聊天记录</div>
      </div>
      <div class="group-actions">
        <button class="action-btn message" @click="sendMessage">
          <i class="icon icon-message"></i> 发消息
        </button>
        <button class="action-btn audio" @click="startAudioCall">
          <i class="icon icon-phone"></i> 发起群音频
        </button>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed } from 'vue'
import { sortFriendsByPinyin, groupFriendsByPinyin } from '../utils/sortUtils'
// mock更多好友
const friends = ref([
  { id: 1, name: '小明', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ming', remark: '阿明' },
  { id: 2, name: '小芳', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Fang', remark: '' },
  { id: 3, name: '小军', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Jun', remark: 'Ben' },
  { id: 4, name: '小李', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Li', remark: 'Cathy' },
  { id: 5, name: '小王', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Wang', remark: '' },
  { id: 6, name: '小赵', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhao', remark: '赵六' },
  { id: 7, name: '小孙', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Sun', remark: '' },
  { id: 8, name: '小周', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhou', remark: '周星驰' },
  { id: 9, name: '小吴', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Wu', remark: '' },
  { id: 10, name: '小郑', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zheng', remark: '郑爽' },
  { id: 11, name: '阿丽', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ali', remark: '' },
  { id: 12, name: '安安', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Anan', remark: '' },
  { id: 13, name: '安南', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Annan', remark: '' },
  { id: 14, name: '兵哥', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Bingge', remark: '' },
  { id: 15, name: '鲍姐', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Baojie', remark: '' },
  { id: 16, name: '八爷', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Bayi', remark: '' },
  { id: 17, name: '张三', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=ZhangSan', remark: '' },
  { id: 18, name: '赵高', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=ZhaoGao', remark: '' },
  { id: 19, name: 'Alice', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Alice', remark: '' },
  { id: 20, name: 'Bob', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Bob', remark: '' },
  { id: 21, name: 'Charlie', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Charlie', remark: '' }
])
const selectedFriends = ref([])
const props = defineProps({
  group: { type: Object, required: true }
})
const showMenu = ref(false)
const showGroupSettings = ref(false)
const showInvite = ref(false)
const toggleMenu = () => { showMenu.value = !showMenu.value }
const closeMenu = () => { showMenu.value = false }
const openGroupSettings = () => { showMenu.value = false; showGroupSettings.value = true }
const closeGroupSettings = () => { showGroupSettings.value = false }
const openInvite = () => { showMenu.value = false; showInvite.value = true }
const closeInvite = () => { showInvite.value = false }
const onQuitGroup = () => { showMenu.value = false; alert('已退出群聊（mock）') }
const sendMessage = () => { alert('发消息（mock）') }
const startAudioCall = () => { alert('发起群音频（mock）') }
const inviteSelected = () => { alert('邀请好友ID: ' + selectedFriends.value.join(',')); showInvite.value = false }
// 好友A-Z排序
const sortedFriends = computed(() => {
  return sortFriendsByPinyin(friends.value)
})

// 按字母分组的好友
const groupedFriends = computed(() => {
  return groupFriendsByPinyin(sortedFriends.value)
})
</script>
<style scoped>
.group-detail-container {
  width: 90%;
  min-width: 700px;
  max-width: 1100px;
  margin: 96px auto 0 auto;
  padding: 32px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
  height: 100%;
  display: flex;
  flex-direction: column;
}
.group-top {
  display: flex;
  align-items: center;
  gap: 32px;
  margin-bottom: 32px;
  position: relative;
}
.group-avatar {
  width: 80px;
  height: 80px;
  border-radius: 16px;
  background: #f8fafc;
}
.group-top-info { flex: 1; }
.group-name { font-size: 28px; font-weight: 700; margin-bottom: 8px; }
.group-id, .group-announcement { font-size: 16px; color: #64748b; margin-bottom: 4px; }
.group-qrcode { font-size: 32px; color: #4a8cff; }
.btn-settings {
  position: absolute;
  top: 0;
  right: 0;
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 12px;
  background: #f1f5f9;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  z-index: 2;
}
.btn-settings:hover { background: #e2e8f0; color: #475569; }
.popover-menu {
  position: absolute;
  top: 54px;
  right: 0;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  min-width: 160px;
  z-index: 10;
  padding: 8px 0;
  animation: popin 0.18s cubic-bezier(.4,1.3,.6,1) both;
}
@keyframes popin {
  0% { opacity: 0; transform: translateY(-10px) scale(0.98); }
  100% { opacity: 1; transform: none; }
}
.popover-menu ul { list-style: none; margin: 0; padding: 0; }
.popover-menu li {
  padding: 12px 24px;
  cursor: pointer;
  font-size: 15px;
  color: #222;
  transition: background 0.2s;
}
.popover-menu li:hover { background: #f1f5f9; }
.popover-menu li.danger { color: #e53e3e; }
.popover-dialog {
  position: fixed;
  left: 0; top: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.18);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}
.invite-animate .popover-content {
  animation: popInvite 0.28s cubic-bezier(.4,1.3,.6,1) both;
  box-shadow: 0 12px 48px rgba(74,140,255,0.18), 0 2px 8px rgba(0,0,0,0.08);
  border: 2.5px solid #4a8cff;
}
@keyframes popInvite {
  0% { opacity: 0; transform: scale(0.92) translateY(40px); }
  100% { opacity: 1; transform: none; }
}
.popover-content {
  background: #fff;
  border-radius: 18px;
  min-width: 340px;
  max-width: 90vw;
  padding: 32px 24px 24px 24px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.popover-content h3 { margin-top: 0; font-size: 22px; font-weight: 700; margin-bottom: 18px; }
.setting-item { margin-bottom: 18px; display: flex; align-items: center; gap: 12px; }
.setting-item label { min-width: 80px; color: #64748b; font-size: 15px; }
.setting-item input[type="text"] { flex: 1; padding: 8px 12px; border-radius: 8px; border: 1px solid #e5e7eb; background: #f8fafc; }
.setting-item input[type="checkbox"] { width: 18px; height: 18px; }
.popover-close, .popover-confirm {
  margin-top: 12px;
  padding: 10px 24px;
  border: none;
  border-radius: 10px;
  background: #4a8cff;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  margin-right: 12px;
}
.popover-close { background: #e5e7eb; color: #222; }
.invite-content { min-width: 400px; }
.friend-list { max-height: 320px; overflow-y: auto; margin-bottom: 16px; width: 100%; }
.friend-group { margin-bottom: 16px; }
.group-header {
  font-size: 14px;
  color: #888;
  padding: 8px 12px;
  background: #f1f5f9;
  border-radius: 8px;
  margin-bottom: 8px;
  font-weight: 600;
}
.friend-item { display: flex; align-items: center; gap: 12px; padding: 8px 0; }
.friend-avatar { width: 36px; height: 36px; border-radius: 8px; }
.group-detail-container {
  width: 90%;
  min-width: 700px;
  max-width: 1100px;
  margin: 48px auto 0 auto;
  padding: 32px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
  height: 100%;
  display: flex;
  flex-direction: column;
}
.group-top {
  display: flex;
  align-items: center;
  gap: 32px;
  margin-bottom: 32px;
  position: relative;
}
.group-avatar {
  width: 80px;
  height: 80px;
  border-radius: 16px;
  background: #f8fafc;
}
.group-top-info { flex: 1; }
.group-name { font-size: 28px; font-weight: 700; margin-bottom: 8px; }
.group-id, .group-announcement { font-size: 16px; color: #64748b; margin-bottom: 4px; }
.group-qrcode { font-size: 32px; color: #4a8cff; }
.btn-settings {
  position: absolute;
  top: 0;
  right: 0;
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 12px;
  background: #f1f5f9;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  z-index: 2;
}
.btn-settings:hover { background: #e2e8f0; color: #475569; }
.settings-popover {
  position: absolute;
  top: 54px;
  right: 0;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  min-width: 160px;
  z-index: 10;
  padding: 8px 0;
}
.settings-popover ul { list-style: none; margin: 0; padding: 0; }
.settings-popover li {
  padding: 12px 24px;
  cursor: pointer;
  font-size: 15px;
  color: #222;
  transition: background 0.2s;
}
.settings-popover li:hover { background: #f1f5f9; }
.settings-popover li.danger { color: #e53e3e; }
.members-section { margin-bottom: 32px; }
.members-title { display: flex; justify-content: space-between; align-items: center; font-size: 16px; font-weight: 600; margin-bottom: 8px; }
.view-all { color: #4a8cff; font-size: 14px; cursor: pointer; }
.members-list {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 18px 8px;
  margin-top: 8px;
}
.member-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.member-avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  object-fit: cover;
  margin-bottom: 4px;
}
.member-name {
  font-size: 13px;
  margin-top: 2px;
  text-align: center;
  max-width: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.role {
  font-size: 11px;
  margin-top: 2px;
  padding: 2px 8px;
  border-radius: 8px;
  background: #f1f5f9;
  color: #4a8cff;
}
.owner { background: #4a8cff; color: #fff; }
.admin { background: #10b981; color: #fff; }
.invite-item { display: flex; flex-direction: column; align-items: center; justify-content: center; }
.invite-btn { width: 48px; height: 48px; border-radius: 12px; background: #f1f5f9; color: #4a8cff; font-size: 28px; border: none; display: flex; align-items: center; justify-content: center; cursor: pointer; margin-bottom: 2px; }
.invite-btn:hover { background: #e0e7ff; }
.group-info-section { background: #f8fafc; border-radius: 12px; padding: 18px 18px 8px 18px; margin-bottom: 24px; }
.info-row { display: flex; justify-content: space-between; align-items: center; font-size: 15px; padding: 6px 0; border-bottom: 1px solid #f0f2f5; }
.info-row:last-child { border-bottom: none; }
.label { color: #888; }
.value { color: #222; font-weight: 500; }
.group-apps-section { background: #f8fafc; border-radius: 12px; padding: 18px; margin-bottom: 24px; }
.group-apps-section h4 { font-size: 18px; font-weight: 600; color: #1e293b; margin: 0 0 12px 0; }
.apps-list { display: flex; gap: 24px; flex-wrap: wrap; }
.app-item { display: flex; align-items: center; gap: 8px; font-size: 15px; color: #334155; background: #fff; border-radius: 10px; padding: 10px 18px; box-shadow: 0 2px 8px rgba(0,0,0,0.04); cursor: pointer; }
.app-item.more { color: #4a8cff; background: #f1f5f9; }
.group-bot-section { background: #f8fafc; border-radius: 12px; padding: 18px; margin-bottom: 24px; }
.group-bot-section h4 { font-size: 18px; font-weight: 600; color: #1e293b; margin: 0 0 12px 0; }
.bot-list { display: flex; gap: 24px; }
.bot-item { display: flex; align-items: center; gap: 8px; font-size: 15px; color: #334155; background: #fff; border-radius: 10px; padding: 10px 18px; box-shadow: 0 2px 8px rgba(0,0,0,0.04); }
.icon-robot { font-size: 20px; color: #4a8cff; }
.group-chat-settings { background: #f8fafc; border-radius: 12px; padding: 18px; margin-bottom: 24px; }
.setting-row { display: flex; justify-content: space-between; align-items: center; font-size: 15px; padding: 10px 0; border-bottom: 1px solid #f0f2f5; }
.setting-row:last-child { border-bottom: none; }
.setting-desc { color: #888; font-size: 13px; margin-left: 12px; }
.group-personal-section { background: #f8fafc; border-radius: 12px; padding: 18px; margin-bottom: 24px; }
.group-actions { display: flex; flex-direction: column; gap: 8px; margin-bottom: 32px; }
.action-btn {
  width: 100%;
  padding: 20px 24px;
  border: none;
  border-radius: 16px;
  font-size: 18px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.action-btn.message {
  background: linear-gradient(135deg, #4a8cff, #8a69ff);
  color: #fff;
}
.action-btn.audio {
  background: linear-gradient(135deg, #10b981, #059669);
  color: #fff;
}
</style> 