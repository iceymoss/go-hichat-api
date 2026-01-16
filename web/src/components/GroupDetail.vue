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
        <img :src="currentGroupDetail.avatar" class="group-avatar" />
        <div class="group-top-info">
          <div class="group-name">{{ currentGroupDetail.name }}</div>
          <div class="group-id">群号：{{ currentGroupDetail.id }}</div>
          <div class="group-announcement">公告：{{ currentGroupDetail.notice || currentGroupDetail.notification || '暂无公告' }}</div>
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
            <li v-if="canManage" @click="openGroupRequests">加群申请</li>
            <li @click="openInvite">邀请成员</li>
            <li v-if="canManage" @click="openInviteLinks">邀请链接/二维码</li>
            <li v-if="canManage" @click="openAnnouncements">公告管理</li>
            <li v-if="currentUserRole === 2" @click="openAdminManager">管理员管理</li>
            <li v-if="currentUserRole === 2" @click="openTransferOwner">转让群主</li>
            <li v-if="currentUserRole === 2" class="danger" @click="disbandGroup">解散群</li>
            <li class="danger" @click="onQuitGroup">退出群聊</li>
          </ul>
        </div>
        <!-- 群设置悬浮弹窗 -->
        <div v-if="showGroupSettings" class="popover-dialog" @click.self="closeGroupSettings">
          <div class="popover-content">
            <h3>群设置</h3>
            <div class="setting-item">
              <label>群名称</label>
              <input type="text" v-model="editName" :disabled="!canManage" :class="{ disabled: !canManage }" />
            </div>
            <div class="setting-item">
              <label>群公告</label>
              <input type="text" v-model="editNotice" :disabled="!canManage" :class="{ disabled: !canManage }" />
            </div>
            <div class="setting-item">
              <label>入群验证</label>
              <input type="checkbox" v-model="editIsVerify" :disabled="!canManage" />
            </div>
            <div class="setting-item"><label>消息免打扰</label><input type="checkbox" /></div>
            <div class="setting-item"><label>置顶聊天</label><input type="checkbox" /></div>
            <div class="popover-actions">
              <button v-if="canManage" class="popover-confirm" @click="saveGroupSettings">保存</button>
              <button class="popover-close" @click="closeGroupSettings">关闭</button>
            </div>
          </div>
        </div>

        <!-- 转让群主弹窗 -->
        <div v-if="showTransferOwner" class="popover-dialog" @click.self="closeTransferOwner">
          <div class="popover-content invite-content">
            <h3>转让群主</h3>
            <div class="friend-list">
              <div v-for="m in members" :key="m.user_id || m.userId || m.id" class="friend-item">
                <input
                  type="radio"
                  name="newOwner"
                  :value="String(m.user_id || m.userId || m.id)"
                  v-model="transferNewOwnerId"
                  :disabled="(m.user_id || m.userId || m.id) === authStore.user?.id || (m.role_level === 2 || m.roleLevel === 2)"
                />
                <img :src="m.user_avatar_url || m.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${m.user_id}`" class="friend-avatar" />
                <span>{{ m.nickname || m.name || m.user_id }}</span>
                <span v-if="(m.role_level === 2 || m.roleLevel === 2)" class="role owner">群主</span>
                <span v-else-if="(m.role_level === 1 || m.roleLevel === 1)" class="role admin">管理员</span>
              </div>
            </div>
            <div class="setting-item" style="width:100%; justify-content: space-between;">
              <label>我保留为管理员</label>
              <input type="checkbox" v-model="transferKeepOldAsAdmin" />
            </div>
            <div class="popover-actions">
              <button class="popover-confirm" @click="confirmTransferOwner">确认转让</button>
              <button class="popover-close" @click="closeTransferOwner">取消</button>
            </div>
          </div>
        </div>

        <!-- 管理员管理弹窗 -->
        <div v-if="showAdminManager" class="popover-dialog" @click.self="closeAdminManager">
          <div class="popover-content invite-content">
            <h3>管理员管理</h3>
            <div class="friend-list">
              <div v-for="m in members" :key="m.user_id || m.userId || m.id" class="friend-item">
                <img :src="m.user_avatar_url || m.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${m.user_id}`" class="friend-avatar" />
                <span style="flex:1">{{ m.nickname || m.name || m.user_id }}</span>
                <button
                  class="btn-small"
                  @click="toggleAdmin(m)"
                  :disabled="(m.role_level === 2 || m.roleLevel === 2)"
                >
                  {{ (m.role_level === 1 || m.roleLevel === 1) ? '取消管理员' : '设为管理员' }}
                </button>
              </div>
            </div>
            <button class="popover-close" @click="closeAdminManager">关闭</button>
          </div>
        </div>
        <!-- 邀请成员悬浮弹窗 -->
        <div v-if="showInvite" class="popover-dialog invite-animate" @click.self="closeInvite">
          <div class="popover-content invite-content">
            <h3>邀请成员</h3>
            <div class="friend-list">
              <div v-for="group in groupedFriends" :key="group.letter" class="friend-group">
                <div class="group-header">{{ group.letter }}</div>

                <div v-for="friend in group.friends" :key="friend.id" class="friend-item"
                     :class="{ 'disabled': isMemberInGroup(friend.friend_uid) || friend.friend_uid === authStore.user?.id }">

                  <input
                    type="checkbox"
                    v-model="selectedFriends"
                    :value="friend.friend_uid"
                    :disabled="isMemberInGroup(friend.friend_uid) || friend.friend_uid === authStore.user?.id"
                  />
                  <img :src="friend.avatar" class="friend-avatar" />
                  <span>{{ friend.remark || friend.nickname }}</span> <span v-if="isMemberInGroup(friend.friend_uid) || friend.friend_uid === authStore.user?.id" class="in-group-badge">已在群聊</span>
                </div>
              </div>
            </div>
            <button class="popover-confirm" @click="inviteSelected">邀请</button>
            <button class="popover-close" @click="closeInvite">关闭</button>
          </div>
        </div>

        <!-- 邀请链接/二维码弹窗 -->
        <div v-if="showInviteLinks" class="popover-dialog" @click.self="closeInviteLinks">
          <div class="popover-content invite-content">
            <h3>邀请链接/二维码</h3>
            <div class="setting-item" style="width:100%; justify-content: space-between;">
              <label>有效期(秒)</label>
              <input type="number" v-model.number="linkExpireSeconds" style="width:160px;" />
            </div>
            <div class="setting-item" style="width:100%; justify-content: space-between;">
              <label>可用次数(0=无限)</label>
              <input type="number" v-model.number="linkMaxUses" style="width:160px;" />
            </div>
            <button class="popover-confirm" @click="createInviteLink">生成邀请链接</button>

            <div class="friend-list" style="margin-top: 12px;">
              <div v-if="inviteLinksLoading" class="empty">加载中...</div>
              <div v-else-if="inviteLinks.length === 0" class="empty">暂无邀请链接</div>
              <div v-else v-for="l in inviteLinks" :key="l.token" class="friend-item" style="align-items:flex-start;">
                <div style="flex:1; display:flex; flex-direction:column; gap:6px;">
                  <div style="font-weight:600;">Token: {{ l.token }}</div>
                  <div style="font-size:12px; color:#64748b;">
                    used {{ l.used_count }}/{{ l.max_uses === 0 ? '∞' : l.max_uses }}
                    · {{ l.revoked ? '已撤销' : '有效' }}
                  </div>
                </div>
                <button class="btn-small" @click="copyToken(l.token)">复制</button>
                <button class="btn-small" :disabled="l.revoked" @click="revokeLink(l.token)">撤销</button>
              </div>
            </div>
            <button class="popover-close" @click="closeInviteLinks">关闭</button>
          </div>
        </div>

        <!-- 移除成员弹窗 -->
        <div v-if="showKickMembers" class="popover-dialog invite-animate" @click.self="closeKickMembers">
          <div class="popover-content invite-content">
            <h3>移除成员</h3>
            <div class="friend-list">
              <div v-if="kickableMembers.length === 0" class="empty">没有可移除的成员</div>
              <div v-else v-for="m in kickableMembers" :key="m.user_id || m.userId || m.id" class="friend-item">
                <input 
                  type="checkbox" 
                  v-model="selectedMembersToKick" 
                  :value="String(m.user_id || m.userId || m.id)"
                />
                <img :src="m.user_avatar_url || m.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${m.user_id}`" class="friend-avatar" />
                <span>{{ m.nickname || m.name || m.user_id }}</span>
                <span v-if="m.role_level === 1 || m.roleLevel === 1" class="role admin">管理员</span>
              </div>
            </div>
            <button class="popover-confirm" @click="confirmKickMembers" :disabled="selectedMembersToKick.length === 0">确定移除</button>
            <button class="popover-close" @click="closeKickMembers">取消</button>
          </div>
        </div>

        <!-- 公告管理弹窗 -->
        <div v-if="showAnnouncements" class="popover-dialog" @click.self="closeAnnouncements">
          <div class="popover-content invite-content">
            <h3>公告管理</h3>
            <div class="setting-item" style="width:100%;">
              <label>发布新公告</label>
              <input type="text" v-model="announcementContent" placeholder="输入公告内容（最多1024字）" />
            </div>
            <button class="popover-confirm" @click="publishAnnouncement">发布</button>

            <div class="friend-list" style="margin-top: 12px;">
              <div v-if="announcementsLoading" class="empty">加载中...</div>
              <div v-else-if="announcements.length === 0" class="empty">暂无公告</div>
              <div v-else v-for="a in announcements" :key="a.id" class="friend-item" style="align-items:flex-start;">
                <div style="flex:1; display:flex; flex-direction:column; gap:6px;">
                  <div style="font-weight:600;">
                    {{ a.pinned ? '[置顶] ' : '' }}{{ a.content }}
                  </div>
                  <div style="font-size:12px; color:#64748b;">
                    {{ a.creator?.nickname || a.created_by }} · {{ new Date((a.created_at || 0) * 1000).toLocaleString() }}
                  </div>
                </div>
                <button class="btn-small" @click="togglePin(a)">{{ a.pinned ? '取消置顶' : '置顶' }}</button>
              </div>
            </div>
            <button class="popover-close" @click="closeAnnouncements">关闭</button>
          </div>
        </div>
      </div>
      <!-- 群成员区块 -->
      <div class="members-section">
        <div class="members-title">
          <span>群聊成员</span>
          <span class="view-all">查看{{ members.length }}名群成员</span>
        </div>
        <div class="members-list">
          <div 
            v-for="m in members" 
            :key="m.user_id || m.id" 
            class="member-item"
          >
            <div class="avatar-wrapper">
              <img :src="m.user_avatar_url || m.avatar || `https://api.dicebear.com/7.x/personas/svg?seed=${m.user_id}`" class="member-avatar" />
            </div>
            <div class="member-name">{{ m.nickname || m.name || m.user_id }}</div>
            <span v-if="m.role_level === 2 || m.role==='owner'" class="role owner">群主</span>
            <span v-else-if="m.role_level === 1 || m.role==='admin'" class="role admin">管理员</span>
          </div>
          <div class="member-item invite-item">
            <button class="invite-btn" @click="openInvite">+</button>
            <div class="member-name">邀请</div>
          </div>
          <div class="member-item delete-item" v-if="canManage">
            <button class="invite-btn delete-btn" @click="openKickMembers">-</button>
            <div class="member-name">移除</div>
          </div>
        </div>
      </div>
      <!-- 群信息区块 -->
      <div class="group-info-section">
        <div class="info-row"><span class="label">群聊名称</span><span class="value">{{ currentGroupDetail.name }}</span></div>
        <div class="info-row"><span class="label">群号</span><span class="value">{{ currentGroupDetail.id }}</span></div>
        <div class="info-row"><span class="label">群公告</span><span class="value">{{ currentGroupDetail.notice || currentGroupDetail.notification || '暂无公告' }}</span></div>
        <div class="info-row" @click="openMySetting" style="cursor:pointer;">
          <span class="label">我的群昵称</span>
          <span class="value">{{ mySetting?.group_nickname || '未设置' }}</span>
        </div>
        <div class="info-row" @click="openMySetting" style="cursor:pointer;">
          <span class="label">群聊备注</span>
          <span class="value">{{ mySetting?.group_remark || '未设置' }}</span>
        </div>
      </div>

      <!-- 我的群资料弹窗 -->
      <div v-if="showMySetting" class="popover-dialog" @click.self="closeMySetting">
        <div class="popover-content">
          <h3>我的群资料</h3>
          <div class="setting-item">
            <label>群内昵称</label>
            <input type="text" v-model="editGroupNickname" />
          </div>
          <div class="setting-item">
            <label>群备注</label>
            <input type="text" v-model="editGroupRemark" />
          </div>
          <div class="popover-actions">
            <button class="popover-confirm" @click="saveMySetting">保存</button>
            <button class="popover-close" @click="closeMySetting">关闭</button>
          </div>
        </div>
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
    <Toast ref="toastRef" />
  </div>
</template>
<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useContactsStore } from '../stores/contacts'
import { useAuthStore } from '../stores/auth'
import { sortFriendsByPinyin, groupFriendsByPinyin } from '../utils/sortUtils'
import Toast from './ui/Toast.vue'

const props = defineProps({
  group: { type: Object, required: true }
})

const contactsStore = useContactsStore()
const authStore = useAuthStore()
const router = useRouter()
const toastRef = ref(null)

const showMenu = ref(false)
const showGroupSettings = ref(false)
const showInvite = ref(false)
const selectedFriends = ref([])
const showTransferOwner = ref(false)
const showAdminManager = ref(false)
const showInviteLinks = ref(false)
const showKickMembers = ref(false)
const selectedMembersToKick = ref([])
const showAnnouncements = ref(false)
const showMySetting = ref(false)

// Editing states
const editName = ref('')
const editNotice = ref('')
const editIsVerify = ref(false)
const initialIsVerify = ref(false)
const transferNewOwnerId = ref('')
const transferKeepOldAsAdmin = ref(true)

// Invite links
const inviteLinksLoading = ref(false)
const inviteLinks = ref([])
const linkExpireSeconds = ref(0)
const linkMaxUses = ref(0)

// Announcements
const announcementsLoading = ref(false)
const announcements = ref([])
const announcementContent = ref('')

// My group setting
const mySetting = computed(() => contactsStore.myGroupMemberSetting)
const editGroupNickname = ref('')
const editGroupRemark = ref('')

// Current group details from store (includes members)
const currentGroupDetail = computed(() => {
  if (contactsStore.currentGroup && contactsStore.currentGroup.id === props.group.id) {
    return contactsStore.currentGroup
  }
  return props.group // Fallback to basic info
})

const members = computed(() => currentGroupDetail.value.members || [])
const sortedFriends = computed(() => sortFriendsByPinyin(contactsStore.friends)) // Use real friends
const groupedFriends = computed(() => groupFriendsByPinyin(sortedFriends.value))

// Check permissions
const currentUserRole = computed(() => {
  if (!authStore.user) return 0
  const myMember = members.value.find(m => m.user_id === authStore.user.id || m.userId === authStore.user.id)
  if (!myMember) return 0
  // Handle inconsistent field naming from backend (role_level vs roleLevel)
  return myMember.role_level !== undefined ? myMember.role_level : (myMember.roleLevel !== undefined ? myMember.roleLevel : 0)
})

const canManage = computed(() => currentUserRole.value >= 1) // 1=Admin, 2=Owner

// Load group detail when group changes
watch(() => props.group.id, async (newId) => {
  if (newId) {
    try {
      await contactsStore.fetchGroupDetail(newId)
      await contactsStore.fetchMyGroupMemberSetting(newId)
    } catch (e) {
      toastRef.value?.show('获取群详情失败', 'error')
    }
  }
  selectedMembersToKick.value = [] // Reset selected members on change
}, { immediate: true })

const toggleMenu = () => { showMenu.value = !showMenu.value }
const closeMenu = () => { showMenu.value = false }

const openGroupRequests = () => {
  showMenu.value = false
  router.push({ name: 'GroupRequests', query: { group_id: props.group.id } })
}

const openGroupSettings = () => { 
  showMenu.value = false; 
  // Initialize edit values
  editName.value = currentGroupDetail.value.name || ''
  editNotice.value = currentGroupDetail.value.notice || currentGroupDetail.value.notification || ''
  const v = currentGroupDetail.value.is_verify !== undefined
    ? currentGroupDetail.value.is_verify
    : (currentGroupDetail.value.isVerify !== undefined ? currentGroupDetail.value.isVerify : false)
  editIsVerify.value = !!v
  initialIsVerify.value = !!v
  showGroupSettings.value = true 
}

const closeGroupSettings = () => { showGroupSettings.value = false }

const saveGroupSettings = async () => {
  if (!editName.value.trim()) {
    toastRef.value?.show('群名称不能为空', 'warning')
    return
  }
  try {
    const payload = {
      name: editName.value,
      notification: editNotice.value
    }
    // 仅当用户修改了“入群验证”才发送该字段，避免误改
    if (editIsVerify.value !== initialIsVerify.value) {
      payload.is_verify = editIsVerify.value ? 1 : 0
    }

    await contactsStore.updateGroupInfo(props.group.id, payload)
    toastRef.value?.show('群信息已更新')
    closeGroupSettings()
  } catch (e) {
    toastRef.value?.show('更新失败', 'error')
  }
}

const openTransferOwner = () => {
  showMenu.value = false
  transferNewOwnerId.value = ''
  transferKeepOldAsAdmin.value = true
  showTransferOwner.value = true
}
const closeTransferOwner = () => { showTransferOwner.value = false }

const confirmTransferOwner = async () => {
  if (!transferNewOwnerId.value) {
    toastRef.value?.show('请选择新群主', 'warning')
    return
  }
  try {
    await contactsStore.transferGroupOwner(props.group.id, transferNewOwnerId.value, transferKeepOldAsAdmin.value)
    toastRef.value?.show('已转让群主')
    closeTransferOwner()
  } catch (e) {
    toastRef.value?.show('转让失败', 'error')
  }
}

const openAdminManager = () => {
  showMenu.value = false
  showAdminManager.value = true
}
const closeAdminManager = () => { showAdminManager.value = false }

const toggleAdmin = async (member) => {
  if (!member) return
  const memberId = member.user_id || member.userId || member.id
  if (!memberId) return
  // 不操作群主
  const role = member.role_level !== undefined ? member.role_level : (member.roleLevel || 0)
  if (role === 2) return
  const nextIsAdmin = role !== 1
  try {
    await contactsStore.setGroupAdmin(props.group.id, [String(memberId)], nextIsAdmin)
    toastRef.value?.show(nextIsAdmin ? '已设为管理员' : '已取消管理员')
  } catch (e) {
    toastRef.value?.show('操作失败', 'error')
  }
}

const disbandGroup = async () => {
  showMenu.value = false
  if (!confirm('确定要解散该群吗？解散后无法恢复。')) return
  try {
    await contactsStore.disbandGroup(props.group.id)
    toastRef.value?.show('群已解散')
  } catch (e) {
    toastRef.value?.show('解散失败', 'error')
  }
}

const openInvite = () => { showMenu.value = false; showInvite.value = true }
const closeInvite = () => { showInvite.value = false }

const openInviteLinks = async () => {
  showMenu.value = false
  showInviteLinks.value = true
  inviteLinksLoading.value = true
  try {
    const list = await contactsStore.fetchInviteLinks(props.group.id, true)
    inviteLinks.value = list
  } catch (e) {
    toastRef.value?.show('获取邀请链接失败', 'error')
  } finally {
    inviteLinksLoading.value = false
  }
}

const closeInviteLinks = () => { showInviteLinks.value = false }

const createInviteLink = async () => {
  try {
    const link = await contactsStore.createInviteLink(props.group.id, Number(linkExpireSeconds.value || 0), Number(linkMaxUses.value || 0))
    if (link?.token) {
      await contactsStore.fetchInviteLinks(props.group.id, true)
      inviteLinks.value = contactsStore.inviteLinks
      await copyToken(link.token)
      toastRef.value?.show('已生成邀请链接（token已复制）')
    } else {
      toastRef.value?.show('生成失败', 'error')
    }
  } catch (e) {
    toastRef.value?.show('生成邀请链接失败', 'error')
  }
}

const revokeLink = async (token) => {
  try {
    await contactsStore.revokeInviteLink(props.group.id, token)
    inviteLinks.value = contactsStore.inviteLinks
    toastRef.value?.show('已撤销')
  } catch (e) {
    toastRef.value?.show('撤销失败', 'error')
  }
}

const copyToken = async (token) => {
  try {
    // Clipboard API only works in secure contexts; provide a fallback.
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(token)
    } else {
      const ta = document.createElement('textarea')
      ta.value = token
      ta.setAttribute('readonly', '')
      ta.style.position = 'fixed'
      ta.style.top = '-9999px'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    toastRef.value?.show('已复制')
  } catch (e) {
    toastRef.value?.show('复制失败', 'error')
  }
}

const openAnnouncements = async () => {
  showMenu.value = false
  showAnnouncements.value = true
  announcementsLoading.value = true
  try {
    const list = await contactsStore.fetchGroupAnnouncements(props.group.id, false)
    announcements.value = list
  } catch (e) {
    toastRef.value?.show('获取公告失败', 'error')
  } finally {
    announcementsLoading.value = false
  }
}

const closeAnnouncements = () => { showAnnouncements.value = false }

const publishAnnouncement = async () => {
  if (!announcementContent.value.trim()) {
    toastRef.value?.show('请输入公告内容', 'warning')
    return
  }
  try {
    await contactsStore.createGroupAnnouncement(props.group.id, announcementContent.value.trim())
    announcementContent.value = ''
    announcements.value = await contactsStore.fetchGroupAnnouncements(props.group.id, false)
    toastRef.value?.show('已发布')
  } catch (e) {
    toastRef.value?.show('发布失败', 'error')
  }
}

const togglePin = async (a) => {
  try {
    await contactsStore.pinGroupAnnouncement(props.group.id, a.id, !a.pinned)
    announcements.value = contactsStore.groupAnnouncements
    toastRef.value?.show(a.pinned ? '已取消置顶' : '已置顶')
  } catch (e) {
    toastRef.value?.show('操作失败', 'error')
  }
}

const openMySetting = async () => {
  showMySetting.value = true
  try {
    const setting = await contactsStore.fetchMyGroupMemberSetting(props.group.id)
    editGroupNickname.value = setting?.group_nickname || setting?.groupNickname || ''
    editGroupRemark.value = setting?.group_remark || setting?.groupRemark || ''
  } catch (e) {
    editGroupNickname.value = ''
    editGroupRemark.value = ''
  }
}

const closeMySetting = () => { showMySetting.value = false }

const saveMySetting = async () => {
  try {
    await contactsStore.updateMyGroupMemberSetting(props.group.id, {
      group_nickname: editGroupNickname.value,
      group_remark: editGroupRemark.value
    })
    toastRef.value?.show('已保存')
    closeMySetting()
  } catch (e) {
    toastRef.value?.show('保存失败', 'error')
  }
}

// 可移除的成员列表（排除自己和群主）
const kickableMembers = computed(() => {
  if (!members.value || !authStore.user) return []
  return members.value.filter(m => {
    const memberId = m.user_id || m.userId || m.id
    const memberRole = m.role_level !== undefined ? m.role_level : (m.roleLevel !== undefined ? m.roleLevel : 0)
    // 排除自己、群主、以及同级或更高级的成员
    return memberId !== authStore.user.id && 
           memberRole < currentUserRole.value && 
           memberRole !== 2 // 不能移除群主
  })
})

const openKickMembers = () => {
  showKickMembers.value = true
  selectedMembersToKick.value = []
}

const closeKickMembers = () => {
  showKickMembers.value = false
  selectedMembersToKick.value = []
}

const confirmKickMembers = async () => {
  if (selectedMembersToKick.value.length === 0) {
    toastRef.value?.show('请选择要移除的成员', 'warning')
    return
  }
  
  const memberNames = selectedMembersToKick.value.map(memberId => {
    const member = members.value.find(m => (m.user_id || m.userId || m.id) === memberId)
    return member ? (member.nickname || member.name || memberId) : memberId
  }).join('、')
  
  if (confirm(`确定要将 ${memberNames} 移出群聊吗？`)) {
    try {
      await contactsStore.kickGroupMember(props.group.id, selectedMembersToKick.value)
      toastRef.value?.show('已移除 ' + selectedMembersToKick.value.length + ' 位成员')
      closeKickMembers()
      // 刷新群详情
      await contactsStore.fetchGroupDetail(props.group.id)
    } catch (e) {
      toastRef.value?.show('移除成员失败', 'error')
    }
  }
}

const onQuitGroup = async () => { 
  showMenu.value = false; 
  if (confirm('确定要退出该群聊吗？')) {
    try {
      await contactsStore.quitGroup(props.group.id)
      toastRef.value?.show('已退出群聊')
    } catch (e) {
      toastRef.value?.show('退出群聊失败', 'error')
    }
  }
}

// 检查好友是否已在群里
const isMemberInGroup = (friendId) => {
  if (!friendId || !members.value || members.value.length === 0) return false
  return members.value.some(m => {
    const memberUserId = m.user_id || m.userId || m.id
    return String(memberUserId) === String(friendId)
  })
}

const sendMessage = () => { toastRef.value?.show('发消息功能开发中...') }
const startAudioCall = () => { toastRef.value?.show('群音频功能开发中...') }
const inviteSelected = async () => { 
  // 过滤掉已在群里的好友和自己
  const validFriends = selectedFriends.value.filter(friendId => {
    return !isMemberInGroup(friendId) && friendId !== authStore.user?.friend_uid
  })
  
  if (validFriends.length === 0) {
    toastRef.value?.show('请选择未在群里的好友', 'warning')
    return
  }
  
  try {
    await contactsStore.inviteGroupMembers(props.group.id, validFriends)
    toastRef.value?.show('已邀请 ' + validFriends.length + ' 位好友')
    showInvite.value = false 
    selectedFriends.value = []
  } catch (e) {
    toastRef.value?.show('邀请失败', 'error')
  }
}
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
.setting-item input[type="text"] { flex: 1; padding: 8px 12px; border-radius: 8px; border: 1px solid #e5e7eb; background: #fff; }
.setting-item input[type="text"].disabled { background: #f3f4f6; color: #9ca3af; }
.setting-item input[type="checkbox"] { width: 18px; height: 18px; }
.popover-actions { display: flex; gap: 12px; margin-top: 12px; width: 100%; }
.popover-close, .popover-confirm {
  margin-top: 0;
  flex: 1;
  padding: 10px 24px;
  border: none;
  border-radius: 10px;
  background: #4a8cff;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}
.popover-close { background: #e5e7eb; color: #222; margin-right: 0; }
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
.friend-item.disabled { opacity: 0.6; cursor: not-allowed; }
.friend-item.disabled input[type="checkbox"] { cursor: not-allowed; }
.in-group-badge {
  margin-left: auto;
  padding: 2px 8px;
  background: #e0e7ff;
  color: #4f46e5;
  border-radius: 4px;
  font-size: 12px;
}
.friend-avatar { width: 36px; height: 36px; border-radius: 8px; }
.btn-small {
  padding: 6px 12px;
  border: none;
  border-radius: 10px;
  background: #4a8cff;
  color: #fff;
  cursor: pointer;
  font-size: 13px;
}
.btn-small:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
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
.delete-btn { color: #ef4444; font-size: 32px; font-weight: 300; }
.delete-btn.active { background: #fee2e2; color: #ef4444; border: 1px solid #fca5a5; }
.avatar-wrapper { position: relative; width: 48px; height: 48px; margin-bottom: 4px; }
.delete-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 18px;
  height: 18px;
  background: #ef4444;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 12px;
  border: 2px solid #fff;
  z-index: 2;
}
.delete-badge .icon { font-size: 10px; }
.member-item.shake .member-avatar {
  animation: shake 0.3s infinite;
}
@keyframes shake {
  0% { transform: rotate(0deg); }
  25% { transform: rotate(1deg); }
  50% { transform: rotate(0deg); }
  75% { transform: rotate(-1deg); }
  100% { transform: rotate(0deg); }
}
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