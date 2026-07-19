'use client';

import React, { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import {
  ArrowLeft,
  Search,
  Bell,
  UserPlus,
  X,
  Shield,
  Crown,
  ShieldCheck,
  MoreHorizontal,
  Send,
  Settings,
  LogOut,
  Trash2,
  Link,
  Megaphone,
  Check,
  Copy,
  Eye,
  EyeOff,
  ChevronRight,
  Users,
  UserMinus,
  UserCheck,
  RefreshCw,
  Pin,
  Loader2,
  AlertCircle,
  CheckCircle,
  XCircle,
  MinusCircle,
  Flag,
  Clock,
} from 'lucide-react';
import { toast } from 'sonner';
import { useT } from '@/hooks/use-i18n';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { getAvatarColor } from '@/lib/utils';
import {
  type GroupInfo,
  type GroupMemberInfo,
  type GroupApplication,
  type GroupInviteLink,
  type GroupAnnouncement,
  type GroupMemberSetting,
  type GroupRoleLevel,
  type GroupAppResult,
  type GroupAppClass,
  type GroupJoinSource,
} from '@/lib/types';

/* ═══════════════════════════════════════
   Helpers
   ═══════════════════════════════════════ */

function fmtTime(date: Date, t: (k: string) => string): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const m = Math.floor(diff / 60000);
  const h = Math.floor(diff / 3600000);
  const d = Math.floor(diff / 86400000);
  if (m < 1) return t('group.time.justNow');
  if (m < 60) return t('group.time.minutesAgo').replace('{m}', String(m));
  if (h < 24) return t('group.time.hoursAgo').replace('{h}', String(h));
  if (d < 30) return t('group.time.monthDay').replace('{month}', String(new Date(date).getMonth() + 1)).replace('{day}', String(new Date(date).getDate()));
  return t('group.time.yearMonthDay').replace('{year}', String(new Date(date).getFullYear())).replace('{month}', String(new Date(date).getMonth() + 1)).replace('{day}', String(new Date(date).getDate()));
}

const roleLabelKey: Record<GroupRoleLevel, string> = { 0: 'group.role.member', 1: 'group.role.admin', 2: 'group.role.owner' };
const roleIcon: Record<GroupRoleLevel, React.ReactNode> = {
  0: <Shield className="w-3 h-3" />,
  1: <ShieldCheck className="w-3 h-3" />,
  2: <Crown className="w-3 h-3" />,
};
const roleColor: Record<GroupRoleLevel, string> = { 0: '#A2ACB5', 1: '#1BB45B', 2: '#F5A623' };
const roleBg: Record<GroupRoleLevel, string> = {
  0: 'rgba(162,172,181,0.1)',
  1: 'rgba(27,180,91,0.1)',
  2: 'rgba(245,166,35,0.1)',
};

const joinSourceLabelKey: Record<GroupJoinSource, string> = { 1: 'group.joinSource.apply', 2: 'group.joinSource.invite', 3: 'group.joinSource.link' };

const appResultConfig: Record<GroupAppResult, { labelKey: string; color: string; bg: string; icon: React.ReactNode }> = {
  0: { labelKey: 'group.appResult.pending', color: '#F5A623', bg: 'rgba(245,166,35,0.1)', icon: <AlertCircle className="w-3.5 h-3.5" /> },
  1: { labelKey: 'group.appResult.approved', color: '#4DCD5E', bg: 'rgba(77,205,94,0.1)', icon: <CheckCircle className="w-3.5 h-3.5" /> },
  2: { labelKey: 'group.appResult.rejected', color: '#FF5252', bg: 'rgba(255,82,82,0.1)', icon: <XCircle className="w-3.5 h-3.5" /> },
  3: { labelKey: 'group.appResult.ignored', color: '#A2ACB5', bg: 'rgba(162,172,181,0.1)', icon: <MinusCircle className="w-3.5 h-3.5" /> },
};

/* ═══════════════════════════════════════
   API helper
   ═══════════════════════════════════════ */

async function apiFetch(path: string, token: string, opts?: RequestInit) {
  const res = await fetch(path, {
    ...opts,
    headers: {
      ...(opts?.headers || {}),
      Authorization: `Bearer ${token}`,
      ...(opts?.body ? { 'Content-Type': 'application/json' } : {}),
    },
  });
  return res.json();
}

/* ═══════════════════════════════════════
   Mappers – API response -> local types
   ═══════════════════════════════════════ */

function mapGroup(g: any): GroupInfo {
  return {
    id: String(g.id),
    name: g.name || '',
    icon: g.icon || '',
    description: g.description || '',
    isVerify: !!g.is_verify,
    notification: g.notification || '',
    createUid: String(g.create_uid || ''),
    groupNickname: g.group_nickname || '',
    groupRemark: g.group_remark || '',
  };
}

function mapMember(m: any, groupId: string): GroupMemberInfo {
  return {
    id: m.id,
    groupId: String(groupId),
    userId: String(m.user_id),
    nickname: m.nickname || (m.user?.nickname) || '',
    avatar: m.user_avatar_url || m.user?.avatar || '',
    roleLevel: (m.role_level ?? 0) as GroupRoleLevel,
    online: false,
    groupNickname: m.group_nickname || '',
    groupRemark: m.group_remark || '',
  };
}

function mapApplication(a: any): GroupApplication {
  return {
    id: a.id,
    userId: String(a.user_id),
    userName: a.user?.nickname || String(a.user_id),
    userAvatar: a.user?.avatar || '',
    groupId: String(a.group_id),
    groupName: a.group?.name || '',
    groupIcon: a.group?.icon || '',
    reqMsg: a.req_msg || '',
    reqTime: tsToDate(a.req_time),
    joinSource: (a.join_source || 1) as GroupJoinSource,
    inviterName: a.inviter_user_id ? String(a.inviter_user_id) : undefined,
    handleResult: (a.handle_result ?? 0) as GroupAppResult,
    // 已读模型：readState 由 receiver_read 决定（进列表即标已读），与好友申请一致
    readState: (a.receiver_read ?? 0) === 1,
  };
}

function tsToDate(ts: any): Date {
  if (!ts) return new Date();
  if (typeof ts === 'number' && ts < 1e12) return new Date(ts * 1000); // Unix秒→毫秒
  return new Date(ts);
}

function mapInviteLink(l: any): GroupInviteLink {
  return {
    token: l.token || l.id || '',
    groupId: String(l.group_id),
    createdBy: String(l.created_by || l.creator_uid || ''),
    createdAt: tsToDate(l.created_at || l.create_time),
    expireAt: (l.expire_at || l.expire_time) ? tsToDate(l.expire_at || l.expire_time) : null,
    maxUses: l.max_uses ?? 0,
    usedCount: l.used_count ?? 0,
    revoked: !!l.revoked,
  };
}

function mapAnnouncement(a: any): GroupAnnouncement {
  return {
    id: a.id,
    groupId: String(a.group_id),
    content: a.content || '',
    createdBy: String(a.created_by || a.creator_uid || ''),
    createdAt: tsToDate(a.created_at || a.create_time),
    pinned: !!a.pinned,
  };
}

function mapMemberSetting(s: any): GroupMemberSetting {
  return {
    groupId: String(s.group_id),
    groupNickname: s.group_nickname || s.nickname || '',
    groupRemark: s.group_remark || s.remark || '',
  };
}

/* ═══════════════════════════════════════
   Reusable Confirm Dialog
   ═══════════════════════════════════════ */

interface ConfirmOpts {
  title: string;
  description: string;
  confirmLabel: string;
  confirmColor: string;
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  onClose: () => void;
}

function ConfirmModal({ open, opts }: { open: boolean; opts: ConfirmOpts | null }) {
  const t = useT();
  if (!open || !opts) return null;
  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) opts.onClose(); }}
    >
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', padding: '24px', width: '90%', maxWidth: '400px', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <button onClick={opts.onClose} className="absolute" style={{ top: 16, right: 16, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <X className="w-4 h-4" />
        </button>
        <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', marginBottom: '8px', paddingRight: 32 }}>{opts.title}</h3>
        <p style={{ fontSize: '14px', color: '#646A73', marginBottom: '20px', lineHeight: '1.5' }}>{opts.description}</p>
        <div className="flex items-center justify-end gap-3">
          <button onClick={opts.onCancel} disabled={opts.loading} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: opts.loading ? 'not-allowed' : 'pointer' }}>{t('common.cancel')}</button>
          <button onClick={opts.onConfirm} disabled={opts.loading} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: opts.confirmColor, color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: opts.loading ? 'not-allowed' : 'pointer', display: 'flex', alignItems: 'center', gap: 6, opacity: opts.loading ? 0.7 : 1 }}>
            {opts.loading && <Loader2 className="w-4 h-4" style={{ animation: 'spin 1s linear infinite' }} />}
            {opts.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Generic Input Modal
   ═══════════════════════════════════════ */

function InputModal({ open, title, onClose, onSubmit, children }: {
  open: boolean;
  title: string;
  onClose: () => void;
  onSubmit: () => void;
  children: React.ReactNode;
}) {
  const t = useT();
  if (!open) return null;
  return (
    <div className="fixed inset-0" style={{ zIndex: 10002, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFF', borderRadius: '16px', padding: '24px', width: '90%', maxWidth: '440px', maxHeight: '80vh', overflow: 'auto', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <button onClick={onClose} className="absolute" style={{ top: 16, right: 16, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <X className="w-4 h-4" />
        </button>
        <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', marginBottom: '16px', paddingRight: 32 }}>{title}</h3>
        {children}
        <div className="flex items-center justify-end gap-3" style={{ marginTop: '16px' }}>
          <button onClick={onClose} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: 'pointer' }}>{t('common.cancel')}</button>
          <button onClick={onSubmit} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: 'pointer' }}>{t('common.confirm')}</button>
        </div>
      </div>
    </div>
  );
}

const inputStyle = { width: '100%', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.1)', padding: '10px 12px', fontSize: '14px', color: '#1C2733', outline: 'none', background: '#F5F7FA', boxSizing: 'border-box' as const };
const focusInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = '#1BB45B'; e.target.style.boxShadow = '0 0 0 3px rgba(27,180,91,0.15)'; };
const blurInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; e.target.style.boxShadow = 'none'; };

/* ═══════════════════════════════════════
   Action Menu (context menu)
   ═══════════════════════════════════════ */

interface ActionMenuItem {
  label: string;
  color?: string;
  onClick: () => void;
}

function ActionMenu({ x, y, items, onClose }: { x: number; y: number; items: ActionMenuItem[]; onClose: () => void }) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const handler = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) onClose(); };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [onClose]);

  return (
    <div ref={ref} className="fixed" style={{ zIndex: 10002, left: Math.min(x, window.innerWidth - 160), top: Math.min(y, window.innerHeight - items.length * 40 - 10) }}>
      <div style={{ background: '#2C3E50', borderRadius: '12px', padding: '4px', boxShadow: '0 4px 24px rgba(0,0,0,0.3)', minWidth: '140px' }}>
        {items.map((item, i) => (
          <button key={i} onClick={() => { item.onClick(); onClose(); }} style={{ display: 'flex', alignItems: 'center', gap: 8, width: '100%', padding: '8px 14px', border: 'none', background: 'transparent', color: item.color || '#FFF', fontSize: '13px', cursor: 'pointer', borderRadius: '8px', transition: 'background 0.15s' }} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.1)'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
            {item.label}
          </button>
        ))}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Main Component
   ═══════════════════════════════════════ */

type View = 'list' | 'app' | 'detail';
type DetailTab = 'members' | 'links' | 'announcements';
type AppStatusFilter = 'all' | GroupAppResult;

export default function GroupList() {
  const t = useT();
  const { setShowGroupPanel, groupAppUnreadCount, currentUser, friends, setActiveTab, setSelectedConversationId, setShowChatDetail, groupRequestsVersion, invalidateGroupRequests, groupsVersion, invalidateGroups, refreshGroupRequestUnread, groupAppNavTab, clearGroupAppNavTab, groupDetailNavId, clearGroupDetailNav } = useIMStore();
  const token = currentUser?.token || '';
  const myUserId = currentUser?.id || '';

  // ── Local State ──
  const [view, setView] = useState<View>('list');
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [detailTab, setDetailTab] = useState<DetailTab>('members');
  const [memberSearch, setMemberSearch] = useState('');

  // App view
  const [appClass, setAppClass] = useState<GroupAppClass>('received');
  const [appStatusFilter, setAppStatusFilter] = useState<AppStatusFilter>('all');
  // 申请是否已真正拉取过——拉取前不要用空 apps 把全局 groupAppUnreadCount 清零（否则进列表视图气泡会闪没）
  const groupsGeneration = useRef(0);
  const appsGeneration = useRef(0);

  // 通知点击带来的跳转意图：进入「群申请」视图并定位子 tab（received=我收到 / sent=我发起）
  useEffect(() => {
    if (!groupAppNavTab) return;
    setView('app');
    setAppClass(groupAppNavTab);
    setAppStatusFilter('all');
    clearGroupAppNavTab();
  }, [groupAppNavTab, clearGroupAppNavTab]);

  // List search
  const [listSearch, setListSearch] = useState('');

  // Local data (fetched from API)
  const [groups, setGroups] = useState<GroupInfo[]>([]);
  const [members, setMembers] = useState<GroupMemberInfo[]>([]);
  const [apps, setApps] = useState<GroupApplication[]>([]);
  const [links, setLinks] = useState<GroupInviteLink[]>([]);
  const [announcements, setAnnouncements] = useState<GroupAnnouncement[]>([]);
  const [settings, setSettings] = useState<GroupMemberSetting[]>([]);

  // Member name cache (userId -> display name from API member data)
  const memberNameCache = useRef<Record<string, string>>({});

  // Modals
  const [confirmOpts, setConfirmOpts] = useState<ConfirmOpts | null>(null);
  const [showCreateGroup, setShowCreateGroup] = useState(false);
  const [showEditGroup, setShowEditGroup] = useState(false);
  const [showInviteFriends, setShowInviteFriends] = useState(false);
  const [showCreateLink, setShowCreateLink] = useState(false);
  const [showCreateAnnouncement, setShowCreateAnnouncement] = useState(false);
  const [showMemberSettings, setShowMemberSettings] = useState(false);
  const [showRevoked, setShowRevoked] = useState(false);

  // Modal form fields
  const [newGroupName, setNewGroupName] = useState('');
  const [newGroupDesc, setNewGroupDesc] = useState('');
  const [newGroupIcon, setNewGroupIcon] = useState('');
  const [newGroupIconFile, setNewGroupIconFile] = useState<File | null>(null);
  const [newGroupIconPreview, setNewGroupIconPreview] = useState('');
  const [createInviteSelected, setCreateInviteSelected] = useState<Set<string>>(new Set());
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [editIcon, setEditIcon] = useState('');
  const [editIconFile, setEditIconFile] = useState<File | null>(null);
  const [editIconPreview, setEditIconPreview] = useState('');
  const [editNotification, setEditNotification] = useState('');
  const [editVerify, setEditVerify] = useState(false);
  const [inviteSelected, setInviteSelected] = useState<Set<string>>(new Set());
  const [sentFriendReqs, setSentFriendReqs] = useState<Set<string>>(new Set());
  const [reportedUsers, setReportedUsers] = useState<Set<string>>(new Set());
  // 加好友 / 举报弹窗
  const [addFriendTarget, setAddFriendTarget] = useState<GroupMemberInfo | null>(null);
  const [addFriendMsg, setAddFriendMsg] = useState('');
  const [addFriendLoading, setAddFriendLoading] = useState(false);
  const [reportTarget, setReportTarget] = useState<GroupMemberInfo | null>(null);
  const [profileTarget, setProfileTarget] = useState<GroupMemberInfo | null>(null); // 成员资料卡
  const [reportReason, setReportReason] = useState('');
  const [reportLoading, setReportLoading] = useState(false);
  const [newLinkExpiry, setNewLinkExpiry] = useState('7d');
  const [newLinkMaxUses, setNewLinkMaxUses] = useState('0');
  const [newAnnContent, setNewAnnContent] = useState('');
  const [settingNickname, setSettingNickname] = useState('');
  const [settingRemark, setSettingRemark] = useState('');

  // Action menu
  const [actionMenu, setActionMenu] = useState<{ x: number; y: number; items: ActionMenuItem[] } | null>(null);

  // Loading
  const [loading, setLoading] = useState(false);

  // ── Helper: get contact name from member data or friends list ──
  const getContactName = useCallback((userId: string): string => {
    if (userId === myUserId) return currentUser?.name || t('group.me');
    if (memberNameCache.current[userId]) return memberNameCache.current[userId];
    const friend = friends.find(f => f.friend_uid === userId) || friends.find(f => f.id === userId);
    if (friend) return friend.remark || friend.name;
    return userId;
  }, [myUserId, currentUser?.name, friends, t]);

  // ── API: Fetch groups list ──
  const fetchGroups = useCallback(async () => {
    if (!token) return;
    const generation = ++groupsGeneration.current;
    try {
      const data = await apiFetch('/api/social/groups', token);
      if (generation !== groupsGeneration.current) return;
      if (data.success && data.data?.list) {
        setGroups(data.data.list.map(mapGroup));
      } else {
        setGroups([]);
      }
    } catch {
      // silently fail, keep empty list
    }
  }, [token]);

  // ── API: Fetch group detail (members) ──
  const fetchGroupDetail = useCallback(async (groupId: string) => {
    if (!token) return;
    try {
      const data = await apiFetch(`/api/social/group/detail?group_id=${groupId}`, token);
      if (data.success && data.data) {
        // Update group info if returned
        if (data.data.group) {
          const updatedGroup = mapGroup(data.data.group);
          // 从成员列表中提取当前用户的群昵称/群备注，合并到群信息中
          const myMember = (data.data.members || []).find((m: any) => String(m.user_id) === myUserId);
          if (myMember) {
            updatedGroup.groupNickname = myMember.group_nickname || '';
            updatedGroup.groupRemark = myMember.group_remark || '';
          }
          setGroups(prev => prev.map(g => g.id === updatedGroup.id ? updatedGroup : g));
        }
        // Update members for this group
        if (data.data.members) {
          const newMembers: GroupMemberInfo[] = data.data.members.map((m: any) => mapMember(m, groupId));
          // Update member name cache
          for (const m of data.data.members) {
            const uid = String(m.user_id);
            const name = m.nickname || m.user?.nickname || '';
            if (name) memberNameCache.current[uid] = name;
          }
          // Replace members for this group, keep others
          setMembers(prev => [
            ...prev.filter(m => m.groupId !== groupId),
            ...newMembers,
          ]);
        }
      }
    } catch {
      // silently fail
    }
  }, [token, myUserId]);

  // ── API: Fetch applications ──
  const fetchApplications = useCallback(async () => {
    if (!token) return;
    const generation = ++appsGeneration.current;
    try {
      const data = await apiFetch('/api/social/group/putInsByUid?class=2', token);
      if (generation !== appsGeneration.current) return;
      if (data.success && data.data?.list) {
        setApps(data.data.list.map(mapApplication));
      } else {
        setApps([]);
      }
    } catch {
      // Preserve the last successful page on refresh failure.
    }
  }, [token]);

  // ── API: Fetch invite links for a group ──
  const fetchInviteLinks = useCallback(async (groupId: string) => {
    if (!token) return;
    try {
      const data = await apiFetch(`/api/social/group/inviteLinks?group_id=${groupId}`, token);
      if (data.success && data.data?.list) {
        const newLinks: GroupInviteLink[] = data.data.list.map(mapInviteLink);
        setLinks(prev => [
          ...prev.filter(l => l.groupId !== groupId),
          ...newLinks,
        ]);
      } else {
        setLinks(prev => prev.filter(l => l.groupId !== groupId));
      }
    } catch {
      // keep existing
    }
  }, [token]);

  // ── API: Fetch announcements for a group ──
  const fetchAnnouncements = useCallback(async (groupId: string) => {
    if (!token) return;
    try {
      const data = await apiFetch(`/api/social/group/announcements?group_id=${groupId}`, token);
      if (data.success && data.data?.list) {
        const newAnns: GroupAnnouncement[] = data.data.list.map(mapAnnouncement);
        setAnnouncements(prev => [
          ...prev.filter(a => a.groupId !== groupId),
          ...newAnns,
        ]);
      } else {
        setAnnouncements(prev => prev.filter(a => a.groupId !== groupId));
      }
    } catch {
      // keep existing
    }
  }, [token]);

  // ── API: Fetch member setting for current user in a group ──
  const fetchMemberSetting = useCallback(async (groupId: string) => {
    if (!token) return;
    try {
      const data = await apiFetch(`/api/social/group/memberSetting?group_id=${groupId}`, token);
      if (data.success && data.data) {
        const setting = mapMemberSetting({ ...data.data, group_id: groupId });
        setSettings(prev => {
          const filtered = prev.filter(s => s.groupId !== groupId);
          return [...filtered, setting];
        });
      }
    } catch {
      // keep existing
    }
  }, [token]);

  // ── On mount: fetch groups ──
  useEffect(() => {
    fetchGroups();
    return () => { groupsGeneration.current += 1; };
  }, [fetchGroups, groupsVersion]);

  // ── 挂载即拉取申请：让「我的群组」列表视图的铃铛徽标也准确，而不只在进入群申请视图后才有 ──
  useEffect(() => {
    if (token) fetchApplications();
    return () => { appsGeneration.current += 1; };
  }, [token, fetchApplications, groupRequestsVersion]);

  // ── When entering app view: fetch applications + 全部标记已读（已读模型，清零气泡） ──
  useEffect(() => {
    if (view !== 'app') return;
    fetchApplications();
    if (!token) return;
    apiFetch('/api/social/group/putIns/read', token, { method: 'PUT' })
      .then(() => {
        // 本地把我收到的申请标已读，气泡立即清零（无需等下次拉取）
        setApps(prev => prev.map(a => (a.userId !== myUserId ? { ...a, readState: true } : a)));
        void refreshGroupRequestUnread();
      })
      .catch(() => {});
  }, [view, token, myUserId, fetchApplications, refreshGroupRequestUnread]);

  // ── When a group is selected: fetch detail, settings, announcements ──
  useEffect(() => {
    if (selectedGroupId && view === 'detail') {
      fetchGroupDetail(selectedGroupId);
      fetchMemberSetting(selectedGroupId);
      fetchAnnouncements(selectedGroupId);
    }
  }, [selectedGroupId, view, fetchGroupDetail, fetchMemberSetting, fetchAnnouncements]);

  // ── When links tab is active: fetch invite links ──
  useEffect(() => {
    if (selectedGroupId && detailTab === 'links') {
      fetchInviteLinks(selectedGroupId);
    }
  }, [selectedGroupId, detailTab, fetchInviteLinks]);

  // ── Computed ──
  const filteredGroups = useMemo(() => {
    if (!listSearch.trim()) return groups;
    const q = listSearch.toLowerCase();
    return groups.filter(g => g.name.toLowerCase().includes(q));
  }, [groups, listSearch]);

  // All groups returned by the API are the user's groups, so myGroupIds = all group ids
  const myGroupIds = useMemo(() => groups.map(g => g.id), [groups]);

  const selectedGroup = useMemo(() => groups.find(g => g.id === selectedGroupId) || null, [groups, selectedGroupId]);

  const myRole = useMemo((): GroupRoleLevel => {
    if (!selectedGroupId) return 0;
    const m = members.find(m => m.groupId === selectedGroupId && m.userId === myUserId);
    return m ? m.roleLevel : 0;
  }, [members, selectedGroupId, myUserId]);

  const isAdmin = myRole >= 1;  // 管理员(1)或群主(2)
  const isOwner = myRole === 2; // 群主

  const groupMembers = useMemo(() => members.filter(m => m.groupId === selectedGroupId), [members, selectedGroupId]);

  const filteredMembers = useMemo(() => {
    let list = groupMembers;
    if (memberSearch.trim()) {
      const q = memberSearch.toLowerCase();
      list = list.filter(m => getContactName(m.userId).toLowerCase().includes(q) || m.nickname.toLowerCase().includes(q));
    }
    return list;
  }, [groupMembers, memberSearch, getContactName]);

  const onlineCount = useMemo(() => groupMembers.filter(m => m.online).length, [groupMembers]);

  const groupLinks = useMemo(() => {
    let list = links.filter(l => l.groupId === selectedGroupId);
    if (!showRevoked) list = list.filter(l => !l.revoked);
    return list;
  }, [links, selectedGroupId, showRevoked]);

  const groupAnns = useMemo(() => {
    const list = announcements.filter(a => a.groupId === selectedGroupId);
    return list.sort((a, b) => (b.pinned ? 1 : 0) - (a.pinned ? 1 : 0) || b.createdAt.getTime() - a.createdAt.getTime());
  }, [announcements, selectedGroupId]);

  const mySetting = useMemo(() => settings.find(s => s.groupId === selectedGroupId), [settings, selectedGroupId]);

  // 群显示名称：有群备注时显示 "群备注（群名称）"，否则直接显示群名称
  const getGroupDisplayName = useCallback((groupId: string, groupName: string) => {
    // 优先从群列表数据中读取（群列表 API 已返回 group_remark）
    const group = groups.find(g => g.id === groupId);
    if (group?.groupRemark) return `${group.groupRemark}（${groupName}）`;
    // 兜底从 settings 读取（点进群详情后加载的）
    const setting = settings.find(s => s.groupId === groupId);
    if (setting?.groupRemark) return `${setting.groupRemark}（${groupName}）`;
    return groupName;
  }, [groups, settings]);

  const filteredApps = useMemo(() => {
    let list = apps.filter(a => appClass === 'received' ? a.userId !== myUserId : a.userId === myUserId);
    if (appStatusFilter !== 'all') list = list.filter(a => a.handleResult === appStatusFilter);
    return list;
  }, [apps, appClass, appStatusFilter, myUserId]);

  const friendsNotInGroup = useMemo(() => {
    const inGroup = new Set(groupMembers.map(m => m.userId));
    return friends.filter(c => !inGroup.has(c.id));
  }, [friends, groupMembers]);

  // ── Handlers ──

  const handleBack = useCallback(() => {
    if (view === 'detail') { setView('list'); setSelectedGroupId(null); setDetailTab('members'); }
    else if (view === 'app') { setView('list'); }
    else { setShowGroupPanel(false); }
  }, [view, setShowGroupPanel]);

  const openGroup = useCallback((gid: string) => { setSelectedGroupId(gid); setView('detail'); setDetailTab('members'); setMemberSearch(''); }, []);

  // 从群聊深链进来：自动打开指定群的详情
  useEffect(() => {
    if (!groupDetailNavId) return;
    openGroup(groupDetailNavId);
    clearGroupDetailNav();
  }, [groupDetailNavId, clearGroupDetailNav, openGroup]);

  // Confirm helper (async-aware)
  const doConfirm = useCallback((title: string, desc: string, confirmLabel: string, confirmColor: string, action: () => Promise<void> | void) => {
    setConfirmOpts({ title, description: desc, confirmLabel, confirmColor, loading: false, onConfirm: async () => {
      setConfirmOpts(prev => prev ? { ...prev, loading: true } : prev);
      try {
        await action();
      } finally {
        setConfirmOpts(null);
      }
    }, onCancel: () => setConfirmOpts(null), onClose: () => setConfirmOpts(null) });
  }, []);

  // Group app actions
  const markAppRead = useCallback((id: number) => { setApps(prev => prev.map(a => a.id === id ? { ...a, readState: true } : a)); }, []);

  const handleAcceptApp = useCallback((app: GroupApplication) => {
    markAppRead(app.id);
    doConfirm(t('group.confirmAgreeTitle'), t('group.confirmAgreeDesc').replace('{user}', app.userName).replace('{group}', app.groupName), t('group.agree'), '#1BB45B', async () => {
      try {
        const data = await apiFetch('/api/social/group/putIn', token, {
          method: 'PUT',
          body: JSON.stringify({ group_req_id: app.id, group_id: app.groupId, handle_result: 1 }),
        });
        if (data.success) {
          appsGeneration.current += 1;
          setApps(prev => prev.map(a => a.id === app.id ? { ...a, handleResult: 1 as GroupAppResult, readState: true } : a));
          invalidateGroupRequests();
          void refreshGroupRequestUnread();
          toast.success(t('group.agreedToast').replace('{user}', app.userName).replace('{group}', app.groupName));
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [doConfirm, markAppRead, token, t, invalidateGroupRequests, refreshGroupRequestUnread]);

  const handleRejectApp = useCallback((app: GroupApplication) => {
    markAppRead(app.id);
    doConfirm(t('group.confirmRejectTitle'), t('group.confirmRejectDesc').replace('{user}', app.userName).replace('{group}', app.groupName), t('group.reject'), '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/putIn', token, {
          method: 'PUT',
          body: JSON.stringify({ group_req_id: app.id, group_id: app.groupId, handle_result: 2 }),
        });
        if (data.success) {
          appsGeneration.current += 1;
          setApps(prev => prev.map(a => a.id === app.id ? { ...a, handleResult: 2 as GroupAppResult, readState: true } : a));
          invalidateGroupRequests();
          void refreshGroupRequestUnread();
          toast.success(t('group.rejectedToast').replace('{user}', app.userName));
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [doConfirm, markAppRead, token, t, invalidateGroupRequests, refreshGroupRequestUnread]);

  const markAllAppRead = useCallback(() => {
    setApps(prev => prev.map(a => a.userId !== myUserId ? { ...a, readState: true } : a));
    toast.success(t('group.allMarkedRead'));
  }, [myUserId, t]);

  // Create group
  const handleCreateGroup = useCallback(async () => {
    if (!newGroupName.trim()) { toast.error(t('group.nameRequiredToast')); return; }
    try {
      setLoading(true);

      // 1. 上传群头像（如果选了图片）
      let iconUrl = newGroupIcon.trim();
      if (newGroupIconFile) {
        const fd = new FormData();
        fd.append('file', newGroupIconFile);
        const uploadRes = await fetch('/api/user/avatar', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
          body: fd,
        }).then(r => r.json());
        if (uploadRes.success && uploadRes.data?.url) {
          iconUrl = uploadRes.data.url;
        } else {
          toast.error(t('group.avatarUploadFailed'));
          return;
        }
      }

      // 2. 创建群
      const body: any = { name: newGroupName.trim() };
      if (iconUrl) body.icon = iconUrl;
      if (newGroupDesc.trim()) body.description = newGroupDesc.trim();
      const data = await apiFetch('/api/social/group', token, {
        method: 'POST',
        body: JSON.stringify(body),
      });
      if (!data.success) {
        toast.error(data.message || t('group.createFailed'));
        return;
      }

      const groupId = data.data?.group_id;

      // 3. 邀请选中的好友
      if (createInviteSelected.size > 0 && groupId) {
        await apiFetch('/api/social/group/invite', token, {
          method: 'POST',
          body: JSON.stringify({ group_id: String(groupId), friend_ids: Array.from(createInviteSelected) }),
        });
      }

      toast.success(t('group.createdToast') + (createInviteSelected.size > 0 ? t('group.createdInvitedSuffix').replace('{count}', String(createInviteSelected.size)) : ''));
      setShowCreateGroup(false);
      setNewGroupName('');
      setNewGroupDesc('');
      setNewGroupIcon('');
      setNewGroupIconFile(null);
      setNewGroupIconPreview('');
      setCreateInviteSelected(new Set());
      await fetchGroups();
    } catch {
      toast.error(t('group.networkError'));
    } finally {
      setLoading(false);
    }
  }, [newGroupName, newGroupIcon, newGroupIconFile, newGroupDesc, createInviteSelected, token, fetchGroups, t]);

  // Edit group
  const openEditGroup = useCallback(() => {
    if (!selectedGroup) return;
    setEditName(selectedGroup.name);
    setEditDesc(selectedGroup.description || '');
    setEditIcon(selectedGroup.icon);
    setEditNotification(selectedGroup.notification);
    setEditVerify(selectedGroup.isVerify);
    setShowEditGroup(true);
  }, [selectedGroup]);

  const handleEditGroup = useCallback(async () => {
    if (!selectedGroupId) return;
    try {
      // 上传新头像（如果选了图片）
      let iconUrl = editIcon;
      if (editIconFile) {
        const fd = new FormData();
        fd.append('file', editIconFile);
        const uploadRes = await fetch('/api/user/avatar', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
          body: fd,
        }).then(r => r.json());
        if (uploadRes.success && uploadRes.data?.url) {
          iconUrl = uploadRes.data.url;
        } else {
          toast.error(t('group.avatarUploadFailed'));
          return;
        }
      }

      const data = await apiFetch('/api/social/group/update', token, {
        method: 'POST',
        body: JSON.stringify({
          group_id: selectedGroupId,
          name: editName,
          icon: iconUrl,
          description: editDesc,
          is_verify: editVerify ? 1 : 0,
        }),
      });
      if (data.success) {
        setGroups(prev => prev.map(g => g.id === selectedGroupId ? { ...g, name: editName, icon: iconUrl, description: editDesc, notification: editNotification, isVerify: editVerify } : g));
        setShowEditGroup(false);
        setEditIconFile(null);
        setEditIconPreview('');
        toast.success(t('group.updatedToast'));
      } else {
        toast.error(data.message || t('group.updateFailed'));
      }
    } catch {
      toast.error(t('group.networkError'));
    }
  }, [selectedGroupId, editName, editDesc, editIcon, editIconFile, editVerify, editNotification, token, t]);

  // Send message to group
  const handleSendMessageToGroup = useCallback(async (group?: { id: string | number; name: string; icon?: string }) => {
    const targetGroup = group || selectedGroup;
    if (!targetGroup || !token || !myUserId) return;
    const groupId = String(targetGroup.id);

    // 确保群会话存在于 chat-store
    const { useChatStore } = await import('@/lib/chat-store');
    const { setupConversation } = await import('@/lib/api-client');
    const store = useChatStore.getState();
    const existing = store.conversations.find(c => c.id === groupId);

    if (!existing) {
      try { await setupConversation(token, myUserId, groupId, 2); } catch { /* may exist */ }
      useChatStore.setState(s => ({
        conversations: [{
          id: groupId,
          type: 'group' as const,
          name: targetGroup.name || groupId,
          avatar: (targetGroup as any).icon || '',
          lastMessage: '',
          lastMessageTime: new Date(),
          unreadCount: 0,
          pinned: false,
          muted: false,
        }, ...s.conversations],
      }));
      store.fetchGroupMembers(token, groupId);
    }

    setShowGroupPanel(false);
    setActiveTab('chats');
    setSelectedConversationId(groupId);
    setShowChatDetail(true);
  }, [selectedGroup, token, myUserId, setShowGroupPanel, setActiveTab, setSelectedConversationId, setShowChatDetail]);

  const handleSendMessage = useCallback(() => {
    handleSendMessageToGroup();
  }, [handleSendMessageToGroup]);

  // Invite friends
  const handleInviteFriends = useCallback(async () => {
    if (inviteSelected.size === 0) { toast.error(t('group.selectFriendsToInvite')); return; }
    if (!selectedGroupId) return;
    try {
      const data = await apiFetch('/api/social/group/invite', token, {
        method: 'POST',
        body: JSON.stringify({ group_id: selectedGroupId, friend_ids: [...inviteSelected] }),
      });
      if (data.success) {
        const names = [...inviteSelected].map(id => getContactName(id)).join('、');
        toast.success(t('group.invitedToast').replace('{names}', names));
        setShowInviteFriends(false);
        setInviteSelected(new Set());
        // Refresh group detail
        fetchGroupDetail(selectedGroupId);
      } else {
        toast.error(data.message || t('group.inviteFailed'));
      }
    } catch {
      toast.error(t('group.networkError'));
    }
  }, [inviteSelected, selectedGroupId, token, getContactName, fetchGroupDetail, t]);

  // Member settings
  const openMemberSettings = useCallback(() => {
    if (!selectedGroupId) return;
    // 优先从成员数据取（已从 detail API 加载），兜底从 settings 取
    const myMember = members.find(m => m.groupId === selectedGroupId && m.userId === myUserId);
    const s = settings.find(s => s.groupId === selectedGroupId);
    setSettingNickname(myMember?.groupNickname || s?.groupNickname || '');
    setSettingRemark(myMember?.groupRemark || s?.groupRemark || '');
    setShowMemberSettings(true);
  }, [selectedGroupId, settings, members, myUserId]);

  const handleSaveMemberSettings = useCallback(async () => {
    if (!selectedGroupId) return;
    try {
      const data = await apiFetch('/api/social/group/memberSetting', token, {
        method: 'POST',
        body: JSON.stringify({
          group_id: selectedGroupId,
          group_nickname: settingNickname,
          group_remark: settingRemark,
        }),
      });
      if (data.success) {
        setSettings(prev => prev.map(s => s.groupId === selectedGroupId ? { ...s, groupNickname: settingNickname, groupRemark: settingRemark } : s));
        // 同步更新群列表数据（让群显示名称实时刷新）
        setGroups(prev => prev.map(g => g.id === selectedGroupId ? { ...g, groupNickname: settingNickname, groupRemark: settingRemark } : g));
        setShowMemberSettings(false);
        toast.success(t('group.settingsSaved'));
      } else {
        toast.error(data.message || t('group.saveFailed'));
      }
    } catch {
      toast.error(t('group.networkError'));
    }
  }, [selectedGroupId, settingNickname, settingRemark, token, t]);

  // Disband / Quit group
  const handleDisband = useCallback(() => {
    if (!selectedGroup) return;
    doConfirm(t('group.confirmDisbandTitle'), t('group.confirmDisbandDesc').replace('{name}', selectedGroup.name), t('group.disband'), '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/disband', token, {
          method: 'POST',
          body: JSON.stringify({ group_id: selectedGroupId }),
        });
        if (data.success) {
          groupsGeneration.current += 1;
          setGroups(prev => prev.filter(g => g.id !== selectedGroupId));
          setMembers(prev => prev.filter(m => m.groupId !== selectedGroupId));
          setSettings(prev => prev.filter(s => s.groupId !== selectedGroupId));
          toast.success(t('group.disbandedToast'));
          setView('list');
          setSelectedGroupId(null);
          invalidateGroups();
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [selectedGroup, selectedGroupId, doConfirm, token, t, invalidateGroups]);

  const handleQuitGroup = useCallback(() => {
    if (!selectedGroup) return;
    doConfirm(t('group.confirmQuitTitle'), t('group.confirmQuitDesc').replace('{name}', selectedGroup.name), t('group.quit'), '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/quit', token, {
          method: 'POST',
          body: JSON.stringify({ group_id: selectedGroupId }),
        });
        if (data.success) {
          groupsGeneration.current += 1;
          setGroups(prev => prev.filter(g => g.id !== selectedGroupId));
          setMembers(prev => prev.filter(m => m.groupId !== selectedGroupId));
          setSettings(prev => prev.filter(s => s.groupId !== selectedGroupId));
          toast.success(t('group.quitToast'));
          setView('list');
          setSelectedGroupId(null);
          invalidateGroups();
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [selectedGroup, selectedGroupId, doConfirm, token, t, invalidateGroups]);

  // Member actions
  const handleSetAdmin = useCallback((m: GroupMemberInfo) => {
    const newRole: GroupRoleLevel = m.roleLevel === 1 ? 0 : 1;
    const label = newRole === 1 ? t('group.setAdmin') : t('group.unsetAdmin');
    doConfirm(label, t('group.confirmSetAdminDesc').replace('{name}', getContactName(m.userId)).replace('{label}', label), label, '#1BB45B', async () => {
      try {
        const data = await apiFetch('/api/social/group/setAdmin', token, {
          method: 'POST',
          body: JSON.stringify({
            group_id: selectedGroupId,
            member_ids: [m.userId],
            is_admin: newRole === 1,
          }),
        });
        if (data.success) {
          setMembers(prev => prev.map(x => x.id === m.id ? { ...x, roleLevel: newRole } : x));
          toast.success(t('group.doneLabel').replace('{label}', label));
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [doConfirm, getContactName, selectedGroupId, token, t]);

  const handleTransferOwner = useCallback((m: GroupMemberInfo) => {
    doConfirm(t('group.confirmTransferTitle'), t('group.confirmTransferDesc').replace('{name}', getContactName(m.userId)), t('group.transfer'), '#F5A623', async () => {
      try {
        const data = await apiFetch('/api/social/group/transferOwner', token, {
          method: 'POST',
          body: JSON.stringify({
            group_id: selectedGroupId,
            new_owner_id: m.userId,
          }),
        });
        if (data.success) {
          setMembers(prev => prev.map(x => {
            if (x.groupId === selectedGroupId && x.userId === myUserId) return { ...x, roleLevel: 1 as GroupRoleLevel };
            if (x.id === m.id) return { ...x, roleLevel: 3 as GroupRoleLevel };
            return x;
          }));
          toast.success(t('group.transferredToast'));
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [doConfirm, getContactName, selectedGroupId, myUserId, token, t]);

  const handleKickMember = useCallback((m: GroupMemberInfo) => {
    doConfirm(t('group.confirmKickTitle'), t('group.confirmKickDesc').replace('{name}', getContactName(m.userId)), t('group.kick'), '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/kick', token, {
          method: 'POST',
          body: JSON.stringify({
            group_id: selectedGroupId,
            member_ids: [m.userId],
          }),
        });
        if (data.success) {
          setMembers(prev => prev.filter(x => x.id !== m.id));
          toast.success(t('group.kickedToast').replace('{name}', getContactName(m.userId)));
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [doConfirm, getContactName, selectedGroupId, token, t]);

  // Invite links
  const handleCreateLink = useCallback(async () => {
    if (!selectedGroupId) return;
    const expireDays: Record<string, number> = { '1d': 86400, '7d': 604800, '30d': 2592000, 'never': 0 };
    const expireSeconds = expireDays[newLinkExpiry] || 604800;
    const maxUses = parseInt(newLinkMaxUses) || 0;
    try {
      const data = await apiFetch('/api/social/group/inviteLink/create', token, {
        method: 'POST',
        body: JSON.stringify({
          group_id: selectedGroupId,
          expire_seconds: expireSeconds > 0 ? expireSeconds : undefined,
          max_uses: maxUses > 0 ? maxUses : undefined,
        }),
      });
      if (data.success) {
        setShowCreateLink(false);
        toast.success(t('group.linkCreatedToast'));
        // Refresh links
        fetchInviteLinks(selectedGroupId);
      } else {
        toast.error(data.message || t('group.createFailed'));
      }
    } catch {
      toast.error(t('group.networkError'));
    }
  }, [selectedGroupId, newLinkExpiry, newLinkMaxUses, token, fetchInviteLinks, t]);

  const handleRevokeLink = useCallback((link: GroupInviteLink) => {
    doConfirm(t('group.confirmRevokeTitle'), t('group.confirmRevokeDesc'), t('group.revoke'), '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/inviteLink/revoke', token, {
          method: 'POST',
          body: JSON.stringify({ token: link.token }),
        });
        if (data.success) {
          setLinks(prev => prev.map(l => l.token === link.token ? { ...l, revoked: true } : l));
          toast.success(t('group.linkRevokedToast'));
        } else {
          toast.error(data.message || t('group.opFailed'));
        }
      } catch {
        toast.error(t('group.networkError'));
      }
    });
  }, [doConfirm, token, t]);

  const handleCopyLink = useCallback((linkToken: string) => {
    navigator.clipboard.writeText(`https://hichat.app/join/${linkToken}`).catch(() => {});
    toast.success(t('group.linkCopiedToast'));
  }, [t]);

  // Announcements
  const handleCreateAnnouncement = useCallback(async () => {
    if (!selectedGroupId || !newAnnContent.trim()) { toast.error(t('group.announcementRequired')); return; }
    try {
      const data = await apiFetch('/api/social/group/announcement', token, {
        method: 'POST',
        body: JSON.stringify({ group_id: selectedGroupId, content: newAnnContent }),
      });
      if (data.success) {
        setShowCreateAnnouncement(false);
        setNewAnnContent('');
        toast.success(t('group.announcementPublished'));
        // Refresh announcements
        fetchAnnouncements(selectedGroupId);
      } else {
        toast.error(data.message || t('group.publishFailed'));
      }
    } catch {
      toast.error(t('group.networkError'));
    }
  }, [selectedGroupId, newAnnContent, token, fetchAnnouncements, t]);

  const handleTogglePin = useCallback(async (ann: GroupAnnouncement) => {
    const newPinned = !ann.pinned;
    try {
      const res = await apiFetch('/api/social/group/announcement/pin', token, {
        method: 'POST',
        body: JSON.stringify({ group_id: ann.groupId, announcement_id: ann.id, pinned: newPinned }),
      });
      if (res.success) {
        setAnnouncements(prev => prev.map(a => a.id === ann.id ? { ...a, pinned: newPinned } : a));
        toast.success(newPinned ? t('group.pinnedToast') : t('group.unpinnedToast'));
      } else {
        toast.error(res.message || t('group.opFailed'));
      }
    } catch { toast.error(t('group.opFailed')); }
  }, [token, t]);

  // Link status
  const getLinkStatus = useCallback((link: GroupInviteLink): { label: string; color: string } => {
    if (link.revoked) return { label: t('group.linkStatus.revoked'), color: '#A2ACB5' };
    if (link.expireAt && new Date() > link.expireAt) return { label: t('group.linkStatus.expired'), color: '#FF5252' };
    if (link.maxUses > 0 && link.usedCount >= link.maxUses) return { label: t('group.linkStatus.usedUp'), color: '#FF5252' };
    return { label: t('group.linkStatus.valid'), color: '#4DCD5E' };
  }, [t]);

  // ── Render ──

  const renderHeader = (title: string, rightContent?: React.ReactNode) => (
    <div className="flex items-center justify-between shrink-0" style={{ height: 56, background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.08)', paddingLeft: 4, paddingRight: 16 }}>
      <button onClick={handleBack} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#1BB45B', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <ArrowLeft className="w-5 h-5" />
      </button>
      <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>{title}</span>
      <div className="flex items-center gap-1">{rightContent}</div>
    </div>
  );

  const pillTab = (active: boolean, onClick: () => void, label: string, badge?: number) => (
    <button onClick={onClick} style={{ padding: '6px 20px', borderRadius: '17px', border: 'none', background: active ? '#1BB45B' : 'transparent', color: active ? '#FFF' : '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', gap: 4 }}>
      {label}
      {badge !== undefined && badge > 0 && !active && (
        <span className="inline-flex items-center justify-center" style={{ width: 16, height: 16, borderRadius: 8, background: '#E53935', color: '#FFF', fontSize: '10px', fontWeight: 700 }}>{badge > 9 ? '9+' : badge}</span>
      )}
    </button>
  );

  const filterPill = (active: boolean, onClick: () => void, label: string, key?: string | number) => (
    <button key={key} onClick={onClick} style={{ padding: '4px 14px', borderRadius: '14px', border: 'none', background: active ? 'rgba(27,180,91,0.1)' : 'rgba(0,0,0,0.04)', color: active ? '#1BB45B' : '#646A73', fontSize: '12px', fontWeight: active ? 500 : 400, cursor: 'pointer', whiteSpace: 'nowrap', transition: 'all 0.15s' }}>{label}</button>
  );

  const avatarCircle = (name: string, size: number, extra?: React.ReactNode, avatar?: string) => (
    <div className="relative shrink-0" style={{ width: size, height: size, borderRadius: '50%', backgroundColor: avatar ? 'transparent' : getAvatarColor(name || '?'), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: size * 0.38, fontWeight: 600, color: '#FFF', overflow: 'hidden' }}>
      {avatar
        ? <img src={avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
        : (name || '?')[0]}
      {extra}
    </div>
  );

  // ═══ VIEW: List ═══
  if (view === 'list') {
    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {renderHeader(t('group.title'),
          <>
            <div className="relative">
              <button onClick={() => { setView('app'); setAppStatusFilter('all'); }} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Bell className="w-5 h-5" />
              </button>
              {groupAppUnreadCount > 0 && (
                <span className="absolute flex items-center justify-center" style={{ top: 6, right: 4, minWidth: 18, height: 18, borderRadius: 9, background: '#E53935', color: '#FFF', fontSize: '10px', fontWeight: 700, padding: '0 5px' }}>
                  {groupAppUnreadCount > 99 ? '99+' : groupAppUnreadCount}
                </span>
              )}
            </div>
            <button onClick={() => setShowCreateGroup(true)} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <UserPlus className="w-5 h-5" />
            </button>
          </>
        )}

        {/* Search */}
        <div className="px-3 py-2.5 shrink-0">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: '#A2ACB5' }} />
            <input value={listSearch} onChange={(e) => setListSearch(e.target.value)} placeholder={t('group.searchPlaceholder')} style={{ ...inputStyle, paddingLeft: '36px', borderRadius: '22px', height: '36px', fontSize: '13px' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </div>

        {/* Group List */}
        <div className="flex-1 overflow-y-auto im-scroll">
          {filteredGroups.filter(g => myGroupIds.includes(g.id)).length > 0 ? (
            <div style={{ background: '#FFF', borderRadius: '12px', margin: '0 8px 8px', overflow: 'hidden' }}>
              {filteredGroups.filter(g => myGroupIds.includes(g.id)).map(group => {
                const gm = members.filter(m => m.groupId === group.id);
                const myM = gm.find(m => m.userId === myUserId);
                const role = myM?.roleLevel ?? 0;
                return (
                  <div key={group.id} className="flex items-center gap-3" style={{ padding: '12px 16px', borderBottom: '1px solid rgba(0,0,0,0.05)', cursor: 'pointer' }} onClick={() => openGroup(group.id)} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
                    {group.icon ? (
                      <img src={group.icon} alt="" className="shrink-0" style={{ width: 48, height: 48, borderRadius: '50%' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
                    ) : (
                      <div className="shrink-0 flex items-center justify-center" style={{ width: 48, height: 48, borderRadius: '50%', background: '#1BB45B', color: '#fff', fontSize: 18, fontWeight: 600 }}>{(group.name || t('group.groupInitial'))[0]}</div>
                    )}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2" style={{ marginBottom: 2 }}>
                        <span style={{ fontSize: '15px', fontWeight: 600, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{getGroupDisplayName(group.id, group.name)}</span>
                        {role >= 1 && (
                          <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[role], backgroundColor: roleBg[role], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                            {roleIcon[role]}{t(roleLabelKey[role])}
                          </span>
                        )}
                        {group.isVerify && (
                          <span style={{ fontSize: '10px', color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 5px' }}>{t('group.needVerify')}</span>
                        )}
                      </div>
                      <div style={{ fontSize: '12px', color: '#A2ACB5' }}>{gm.length > 0 ? t('group.memberCount').replace('{count}', String(gm.length)) : t('group.groupFallback')}</div>
                      {group.notification && (
                        <div style={{ fontSize: '12px', color: '#708499', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginTop: 2 }}>
                          {group.notification}
                        </div>
                      )}
                    </div>
                    <button onClick={(e) => { e.stopPropagation(); handleSendMessageToGroup(group); }} style={{ padding: '5px 14px', borderRadius: '16px', border: '1px solid #1BB45B', background: 'transparent', color: '#1BB45B', fontSize: '12px', fontWeight: 500, cursor: 'pointer', whiteSpace: 'nowrap', display: 'flex', alignItems: 'center', gap: 4 }}>
                      <Send className="w-3 h-3" /> {t('group.sendMessage')}
                    </button>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
              <Users className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: '16px' }} />
              <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: '8px' }}>{listSearch ? t('group.noMatch') : t('group.empty')}</div>
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{listSearch ? t('group.noMatchHint') : t('group.emptyHint')}</div>
            </div>
          )}
        </div>

        {/* Modals */}
        <InputModal open={showCreateGroup} title={t('group.create')} onClose={() => { setShowCreateGroup(false); setNewGroupIconFile(null); setNewGroupIconPreview(''); setCreateInviteSelected(new Set()); }} onSubmit={handleCreateGroup}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.nameRequired')}</label>
            <input value={newGroupName} onChange={(e) => setNewGroupName(e.target.value)} placeholder={t('group.namePlaceholder')} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.desc')}</label>
            <textarea value={newGroupDesc} onChange={(e) => setNewGroupDesc(e.target.value)} rows={2} placeholder={t('group.descPlaceholder')} style={{ ...inputStyle, resize: 'none' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.icon')}</label>
            <div className="flex items-center gap-3">
              {newGroupIconPreview ? (
                <img src={newGroupIconPreview} alt="" style={{ width: 48, height: 48, borderRadius: 8, objectFit: 'cover' }} />
              ) : (
                <div style={{ width: 48, height: 48, borderRadius: 8, background: '#E8EDEF', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, color: '#A2ACB5' }}>+</div>
              )}
              <label style={{ padding: '6px 14px', borderRadius: 8, border: '1px solid rgba(0,0,0,0.1)', background: '#F5F7FA', fontSize: 13, color: '#1BB45B', cursor: 'pointer', fontWeight: 500 }}>
                {t('group.pickImage')}
                <input type="file" accept="image/*" hidden onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    setNewGroupIconFile(file);
                    setNewGroupIconPreview(URL.createObjectURL(file));
                  }
                }} />
              </label>
              {newGroupIconPreview && (
                <button onClick={() => { setNewGroupIconFile(null); setNewGroupIconPreview(''); }} style={{ fontSize: 12, color: '#A2ACB5', background: 'none', border: 'none', cursor: 'pointer' }}>{t('group.remove')}</button>
              )}
            </div>
          </div>
          {friends.length > 0 && (
            <div>
              <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.inviteFriendsLabel')} <span style={{ fontWeight: 400, color: '#A2ACB5' }}>{t('group.selectedCount').replace('{count}', String(createInviteSelected.size))}</span></label>
              <div style={{ maxHeight: '200px', overflow: 'auto', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.06)' }}>
                {friends.map(c => (
                  <label key={c.id} className="flex items-center gap-3" style={{ padding: '8px 12px', borderBottom: '1px solid rgba(0,0,0,0.04)', cursor: 'pointer' }}>
                    <input type="checkbox" checked={createInviteSelected.has(c.id)} onChange={() => {
                      setCreateInviteSelected(prev => { const n = new Set(prev); if (n.has(c.id)) n.delete(c.id); else n.add(c.id); return n; });
                    }} style={{ width: 16, height: 16, accentColor: '#1BB45B' }} />
                    {avatarCircle(c.name, 32, undefined, c.avatar)}
                    <span style={{ fontSize: '14px', color: '#1C2733' }}>{c.name}</span>
                  </label>
                ))}
              </div>
            </div>
          )}
        </InputModal>

        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />
      </div>
    );
  }

  // ═══ VIEW: Applications ═══
  if (view === 'app') {
    const statusFilters: { key: AppStatusFilter; label: string }[] = [
      { key: 'all', label: t('group.filter.all') }, { key: 0, label: t('group.appResult.pending') }, { key: 1, label: t('group.appResult.approved') }, { key: 2, label: t('group.appResult.rejected') },
    ];
    const getEmptyMsg = () => {
      if (appStatusFilter === 'all') return { t: appClass === 'received' ? t('group.emptyReceived') : t('group.emptySent'), d: t('group.emptyHere') };
      return { t: t('group.emptyMatch'), d: t('group.emptyMatchHint') };
    };
    const emptyMsg = getEmptyMsg();

    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {renderHeader(t('group.appTitle'),
          <button onClick={markAllAppRead} style={{ fontSize: '13px', color: '#1BB45B', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 500 }}>{t('group.allRead')}</button>
        )}

        {/* Class tabs */}
        <div className="flex items-center shrink-0" style={{ padding: '12px 16px 8px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
          <div className="flex items-center" style={{ borderRadius: '20px', background: 'rgba(0,0,0,0.04)', padding: '3px' }}>
            {pillTab(appClass === 'received', () => { setAppClass('received'); setAppStatusFilter('all'); }, t('group.app.received'), groupAppUnreadCount)}
            {pillTab(appClass === 'sent', () => { setAppClass('sent'); setAppStatusFilter('all'); }, t('group.app.sent'))}
          </div>
        </div>

        {/* Status filter */}
        <div className="flex items-center gap-2 shrink-0 overflow-x-auto" style={{ padding: '8px 16px 10px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
          {statusFilters.map(f => filterPill(appStatusFilter === f.key, () => setAppStatusFilter(f.key), f.label, f.key))}
        </div>

        {/* App list */}
        <div className="flex-1 overflow-y-auto im-scroll">
          {filteredApps.length > 0 ? (
            <div style={{ background: '#FFF', borderRadius: '12px', margin: '8px', overflow: 'hidden' }}>
              {filteredApps.map(app => {
                const rc = appResultConfig[app.handleResult];
                const isReceived = app.userId !== myUserId;
                const isPending = app.handleResult === 0;
                return (
                  <div key={app.id} className="relative" style={{ background: isReceived && !app.readState ? 'rgba(245,166,35,0.03)' : '#FFF', borderLeft: isReceived && !app.readState ? '3px solid #F5A623' : 'none', borderBottom: '1px solid rgba(0,0,0,0.05)', padding: '12px 16px', transition: 'background 0.15s' }}>
                    <div className="flex gap-3">
                      {/* Avatar */}
                      {isReceived ? (
                        avatarCircle(app.userName, 44, <div className="absolute" style={{ bottom: -2, left: '50%', transform: 'translateX(-50%)', width: 20, height: 3, borderRadius: 2, backgroundColor: rc.color }} />, app.userAvatar)
                      ) : (
                        app.groupIcon ? (
                          <img src={app.groupIcon} alt="" className="shrink-0" style={{ width: 44, height: 44, borderRadius: '50%' }} />
                        ) : (
                          avatarCircle(app.groupName, 44)
                        )
                      )}
                      <div className="flex-1 min-w-0">
                        {/* Row 1: name + source + status + time */}
                        <div className="flex items-center gap-2 flex-wrap" style={{ marginBottom: 4 }}>
                          <span style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733' }}>{isReceived ? app.userName : app.groupName}</span>
                          <span style={{ fontSize: '10px', color: '#708499', backgroundColor: 'rgba(0,0,0,0.04)', borderRadius: '3px', padding: '0 4px' }}>{t(joinSourceLabelKey[app.joinSource])}</span>
                          <span style={{ fontSize: '11px', fontWeight: 500, color: rc.color, backgroundColor: rc.bg, borderRadius: '4px', padding: '1px 6px', display: 'flex', alignItems: 'center', gap: 3 }}>
                            {React.cloneElement(rc.icon as React.ReactElement<any>, { style: { color: rc.color, width: 12, height: 12 } })}{t(rc.labelKey)}
                          </span>
                          <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', whiteSpace: 'nowrap' }}>{fmtTime(app.reqTime, t)}</span>
                        </div>
                        {/* Group info for received */}
                        {isReceived && (
                          <div className="flex items-center gap-1.5" style={{ marginBottom: 4 }}>
                            {app.groupIcon ? (
                              <img src={app.groupIcon} alt="" style={{ width: 16, height: 16, borderRadius: '50%' }} />
                            ) : (
                              <div style={{ width: 16, height: 16, borderRadius: '50%', backgroundColor: getAvatarColor(app.groupName || '?'), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 8, fontWeight: 600, color: '#FFF' }}>{(app.groupName || '?')[0]}</div>
                            )}
                            <span style={{ fontSize: '12px', color: '#708499' }}>{app.groupName}</span>
                          </div>
                        )}
                        {/* Inviter */}
                        {app.inviterName && (
                          <div style={{ fontSize: '12px', color: '#708499', marginBottom: 4 }}>{t('group.inviter').replace('{name}', app.inviterName)}</div>
                        )}
                        {/* Message */}
                        <div style={{ fontSize: '13px', color: '#646A73', lineHeight: '1.5', marginBottom: isReceived && isPending ? '6px' : 0 }}>{app.reqMsg}</div>
                        {/* Action buttons */}
                        {isReceived && isPending && (
                          <div className="flex items-center gap-2" style={{ marginTop: 6 }}>
                            <button onClick={() => handleAcceptApp(app)} style={{ padding: '5px 16px', borderRadius: '6px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '12px', fontWeight: 500, cursor: 'pointer' }}>{t('group.agree')}</button>
                            <button onClick={() => handleRejectApp(app)} style={{ padding: '5px 16px', borderRadius: '6px', border: '1px solid #FF5252', background: '#FFF', color: '#FF5252', fontSize: '12px', fontWeight: 500, cursor: 'pointer' }}>{t('group.reject')}</button>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
              <Users className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: '16px' }} />
              <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: '8px' }}>{emptyMsg.t}</div>
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{emptyMsg.d}</div>
            </div>
          )}
        </div>

        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />
      </div>
    );
  }

  // ═══ VIEW: Detail ═══
  if (view === 'detail' && selectedGroup) {
    const tabItems: { key: DetailTab; label: string; show: boolean }[] = [
      { key: 'members', label: t('group.tab.members').replace('{count}', String(groupMembers.length)), show: true },
      { key: 'links', label: t('group.tab.links'), show: isAdmin },
      { key: 'announcements', label: t('group.tab.announcements'), show: true },
    ];

    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {renderHeader(getGroupDisplayName(selectedGroup.id, selectedGroup.name))}

        <div className="flex-1 overflow-y-auto im-scroll">
          {/* ── Group Info Card ── */}
          <div style={{ background: '#FFF', borderRadius: '12px', margin: '8px', padding: '20px' }}>
            <div className="flex items-start gap-4">
              <img src={selectedGroup.icon} alt="" className="shrink-0" style={{ width: 72, height: 72, borderRadius: '50%' }} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap" style={{ marginBottom: 4 }}>
                  <span style={{ fontSize: '18px', fontWeight: 700, color: '#1C2733' }}>{getGroupDisplayName(selectedGroup.id, selectedGroup.name)}</span>
                  {myRole >= 1 && (
                    <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[myRole], backgroundColor: roleBg[myRole], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                      {roleIcon[myRole]}{t(roleLabelKey[myRole])}
                    </span>
                  )}
                  {selectedGroup.isVerify && (
                    <span style={{ fontSize: '10px', color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 5px' }}>{t('group.needVerify')}</span>
                  )}
                </div>
                <div style={{ fontSize: '13px', color: '#A2ACB5', marginBottom: 4 }}>
                  {t('group.membersOnline').replace('{count}', String(groupMembers.length))}{onlineCount > 0 ? t('group.membersOnlineSuffix').replace('{count}', String(onlineCount)) : ''}
                </div>
                {selectedGroup.description && (
                  <div style={{ fontSize: '12px', color: '#708499', marginBottom: 4, lineHeight: '1.5' }}>
                    {selectedGroup.description}
                  </div>
                )}
                {mySetting?.groupRemark && (
                  <div style={{ fontSize: '12px', color: '#708499', marginBottom: 4 }}>{t('group.remark').replace('{remark}', mySetting.groupRemark)}</div>
                )}
                {/* 置顶公告 */}
                {groupAnns.find(a => a.pinned) && (
                  <div className="flex items-start gap-1.5" style={{ fontSize: '12px', color: '#646A73', padding: '6px 10px', background: 'rgba(120,140,160,0.08)', borderRadius: '8px', marginTop: 6 }}>
                    <Pin className="w-3 h-3 shrink-0" style={{ marginTop: 1 }} />
                    <span>{groupAnns.find(a => a.pinned)!.content}</span>
                  </div>
                )}
                {/* 群描述/公告（非置顶） */}
                {selectedGroup.notification && !groupAnns.find(a => a.pinned) && (
                  <div style={{ fontSize: '12px', color: '#708499', padding: '6px 10px', background: 'rgba(0,0,0,0.03)', borderRadius: '8px', marginTop: 6 }}>
                    {selectedGroup.notification}
                  </div>
                )}
              </div>
            </div>

            {/* Action buttons */}
            <div className="flex items-center gap-2 flex-wrap" style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <button onClick={handleSendMessage} style={{ padding: '7px 16px', borderRadius: '8px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                <Send className="w-3.5 h-3.5" /> {t('group.sendMessage')}
              </button>
              {isAdmin && (
                <button onClick={openEditGroup} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Settings className="w-3.5 h-3.5" /> {t('group.editGroup')}
                </button>
              )}
              <button onClick={() => setShowInviteFriends(true)} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                <UserPlus className="w-3.5 h-3.5" /> {t('group.inviteFriends')}
              </button>
              <button onClick={openMemberSettings} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                <Settings className="w-3.5 h-3.5" /> {t('group.mySettings')}
              </button>
              {isOwner && (
                <button onClick={handleDisband} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid #E53935', background: '#FFF', color: '#E53935', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Trash2 className="w-3.5 h-3.5" /> {t('group.disband')}
                </button>
              )}
              {!isOwner && (
                <button onClick={handleQuitGroup} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid #E53935', background: '#FFF', color: '#E53935', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <LogOut className="w-3.5 h-3.5" /> {t('group.quit')}
                </button>
              )}
            </div>
          </div>

          {/* ── Tab Bar ── */}
          <div className="flex items-center shrink-0" style={{ padding: '8px 8px 0' }}>
            {tabItems.filter(t => t.show).map(tab => (
              <button key={tab.key} onClick={() => setDetailTab(tab.key)} style={{ padding: '8px 16px', borderBottom: detailTab === tab.key ? '2px solid #1BB45B' : '2px solid transparent', background: 'none', borderLeft: 'none', borderRight: 'none', borderTop: 'none', color: detailTab === tab.key ? '#1BB45B' : '#646A73', fontSize: '13px', fontWeight: detailTab === tab.key ? 600 : 400, cursor: 'pointer', transition: 'all 0.2s' }}>
                {tab.label}
              </button>
            ))}
          </div>

          {/* ── Members Tab ── */}
          {detailTab === 'members' && (
            <div style={{ padding: '8px' }}>
              <div className="relative" style={{ marginBottom: '8px' }}>
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: '#A2ACB5' }} />
                <input value={memberSearch} onChange={(e) => setMemberSearch(e.target.value)} placeholder={t('group.searchMember')} style={{ ...inputStyle, paddingLeft: '36px', borderRadius: '22px', height: '36px', fontSize: '13px' }} onFocus={focusInput} onBlur={blurInput} />
              </div>
              <div style={{ background: '#FFF', borderRadius: '12px', overflow: 'hidden' }}>
                {filteredMembers.map(m => {
                  const name = getContactName(m.userId);
                  const isMe = m.userId === myUserId;
                  const displayName = m.nickname || name;
                  return (
                    <div key={m.id} className="flex items-center gap-3" style={{ padding: '10px 16px', borderBottom: '1px solid rgba(0,0,0,0.05)', cursor: isMe ? 'default' : 'pointer' }} onClick={() => { if (!isMe) setProfileTarget(m); }}>
                      <div className="relative shrink-0" style={{ width: 40, height: 40 }}>
                        {m.avatar ? (
                          <img src={m.avatar} alt={name} style={{ width: 40, height: 40, borderRadius: '50%', objectFit: 'cover' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }} />
                        ) : null}
                        <div style={{ width: 40, height: 40, borderRadius: '50%', background: getAvatarColor(name), display: m.avatar ? 'none' : 'flex', alignItems: 'center', justifyContent: 'center', color: '#FFF', fontSize: 16, fontWeight: 600 }}>{name[0]}</div>
                        {m.online && <span className="absolute" style={{ bottom: 0, right: 0, width: 10, height: 10, borderRadius: '50%', backgroundColor: '#4DCD5E', border: '2px solid #FFF' }} />}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span style={{ fontSize: '14px', fontWeight: 500, color: '#1C2733' }}>{displayName}</span>
                          {m.groupNickname && <span style={{ fontSize: '11px', color: '#A2ACB5' }}>({m.groupNickname})</span>}
                          {m.roleLevel >= 1 && (
                            <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[m.roleLevel], backgroundColor: roleBg[m.roleLevel], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                              {roleIcon[m.roleLevel]}{t(roleLabelKey[m.roleLevel])}
                            </span>
                          )}
                          {m.nickname && m.nickname !== name && (
                            <span style={{ fontSize: '11px', color: '#A2ACB5' }}>{name}</span>
                          )}
                        </div>
                      </div>
                      {isMe ? (
                        <button onClick={openMemberSettings} style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                          <Settings className="w-4 h-4" />
                        </button>
                      ) : (
                        <button onClick={(e) => {
                          e.stopPropagation();
                          const items: ActionMenuItem[] = [];
                          // 群主专属操作
                          if (isOwner) {
                            if (m.roleLevel === 1) items.push({ label: t('group.unsetAdmin'), onClick: () => handleSetAdmin(m) });
                            else if (m.roleLevel === 0) items.push({ label: t('group.setAdmin'), onClick: () => handleSetAdmin(m) });
                            items.push({ label: t('group.transferOwner'), color: '#F5A623', onClick: () => handleTransferOwner(m) });
                          }
                          // 管理员/群主可以踢人（管理员不能踢同级或更高）
                          if (isAdmin && (isOwner || m.roleLevel < myRole)) {
                            items.push({ label: t('group.kick'), color: '#FF5252', onClick: () => handleKickMember(m) });
                          }
                          // 所有人都能看到的选项
                          const isFriend = friends.some(f => f.id === m.userId);
                          const alreadySent = sentFriendReqs.has(m.userId);
                          const alreadyReported = reportedUsers.has(m.userId);
                          if (!isFriend) {
                            items.push({
                              label: alreadySent ? t('group.appSent') : t('group.addFriend'),
                              color: alreadySent ? '#A2ACB5' : undefined,
                              onClick: alreadySent ? () => {} : () => {
                                setAddFriendTarget(m);
                                setAddFriendMsg(t('group.applyMsgFromGroup').replace('{name}', selectedGroup?.name || ''));
                                setActionMenu(null);
                              },
                            });
                          }
                          items.push({
                            label: alreadyReported ? t('group.reported') : t('group.report'),
                            color: alreadyReported ? '#A2ACB5' : '#FF5252',
                            onClick: alreadyReported ? () => {} : () => {
                              setReportTarget(m);
                              setReportReason('');
                              setActionMenu(null);
                            },
                          });
                          setActionMenu({ x: e.clientX, y: e.clientY, items });
                        }} style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                          <MoreHorizontal className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* ── Links Tab ── */}
          {detailTab === 'links' && isAdmin && (
            <div style={{ padding: '8px' }}>
              <div className="flex items-center justify-between" style={{ marginBottom: '8px' }}>
                <button onClick={() => setShowCreateLink(true)} style={{ padding: '7px 16px', borderRadius: '8px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Link className="w-3.5 h-3.5" /> {t('group.createLink')}
                </button>
                <button onClick={() => setShowRevoked(v => !v)} style={{ fontSize: '12px', color: '#708499', background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3 }}>
                  {showRevoked ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                  {showRevoked ? t('group.hideRevoked') : t('group.showRevoked')}
                </button>
              </div>
              {groupLinks.length > 0 ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {groupLinks.map(link => {
                    const status = getLinkStatus(link);
                    const progress = link.maxUses > 0 ? Math.min(link.usedCount / link.maxUses * 100, 100) : 0;
                    return (
                      <div key={link.token} style={{ background: '#FFF', borderRadius: '12px', padding: '14px', border: '1px solid rgba(0,0,0,0.06)' }}>
                        <div className="flex items-center justify-between" style={{ marginBottom: 8 }}>
                          <span style={{ fontSize: '11px', fontWeight: 500, color: status.color, backgroundColor: `${status.color}15`, borderRadius: '4px', padding: '2px 8px' }}>{status.label}</span>
                          <span style={{ fontSize: '11px', color: '#A2ACB5' }}>{fmtTime(link.createdAt, t)}</span>
                        </div>
                        <div style={{ fontSize: '12px', color: '#708499', fontFamily: 'monospace', background: '#F5F7FA', borderRadius: '6px', padding: '6px 10px', marginBottom: 8, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{link.token}</div>
                        <div style={{ fontSize: '12px', color: '#A2ACB5', marginBottom: 6 }}>
                          {t('group.linkCreator').replace('{name}', getContactName(link.createdBy))}
                          {link.expireAt && t('group.linkExpire').replace('{date}', new Date(link.expireAt).toLocaleDateString())}
                          {!link.expireAt && t('group.linkNeverExpire')}
                        </div>
                        {link.maxUses > 0 && (
                          <div style={{ marginBottom: 8 }}>
                            <div className="flex items-center justify-between" style={{ marginBottom: 4 }}>
                              <span style={{ fontSize: '11px', color: '#A2ACB5' }}>{t('group.linkUsage')}</span>
                              <span style={{ fontSize: '11px', color: '#646A73' }}>{link.usedCount}/{link.maxUses}</span>
                            </div>
                            <div style={{ height: 4, borderRadius: 2, background: 'rgba(0,0,0,0.06)' }}>
                              <div style={{ height: '100%', borderRadius: 2, background: progress >= 100 ? '#FF5252' : '#1BB45B', width: `${progress}%`, transition: 'width 0.3s' }} />
                            </div>
                          </div>
                        )}
                        {!link.revoked && status.label === t('group.linkStatus.valid') && (
                          <div className="flex items-center gap-2">
                            <button onClick={() => handleCopyLink(link.token)} style={{ padding: '4px 12px', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3 }}>
                              <Copy className="w-3 h-3" /> {t('group.copy')}
                            </button>
                            <button onClick={() => handleRevokeLink(link)} style={{ padding: '4px 12px', borderRadius: '6px', border: '1px solid #FF5252', background: '#FFF', color: '#FF5252', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3 }}>
                              <RefreshCw className="w-3 h-3" /> {t('group.revoke')}
                            </button>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center" style={{ padding: '40px 24px' }}>
                  <Link className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: '12px' }} />
                  <div style={{ fontSize: '14px', color: '#A2ACB5' }}>{showRevoked ? t('group.noLinks') : t('group.noValidLinks')}</div>
                </div>
              )}
            </div>
          )}

          {/* ── Announcements Tab ── */}
          {detailTab === 'announcements' && (
            <div style={{ padding: '8px' }}>
              {isAdmin && (
                <div style={{ marginBottom: '8px' }}>
                  <button onClick={() => setShowCreateAnnouncement(true)} style={{ padding: '7px 16px', borderRadius: '8px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                    <Megaphone className="w-3.5 h-3.5" /> {t('group.publishAnnouncement')}
                  </button>
                </div>
              )}
              {groupAnns.length > 0 ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {groupAnns.map(ann => (
                    <div key={ann.id} style={{ background: '#FFF', borderRadius: '12px', padding: '14px', border: ann.pinned ? '1px solid rgba(27,180,91,0.2)' : '1px solid rgba(0,0,0,0.06)' }}>
                      {ann.pinned && (
                        <div className="flex items-center gap-1" style={{ marginBottom: 8 }}>
                          <Pin className="w-3 h-3" style={{ color: '#1BB45B' }} />
                          <span style={{ fontSize: '11px', fontWeight: 500, color: '#1BB45B' }}>{t('group.pinned')}</span>
                        </div>
                      )}
                      <div style={{ fontSize: '14px', color: '#1C2733', lineHeight: '1.6', marginBottom: 10 }}>{ann.content}</div>
                      <div className="flex items-center justify-between">
                        <div style={{ fontSize: '12px', color: '#A2ACB5' }}>{getContactName(ann.createdBy)} · {fmtTime(ann.createdAt, t)}</div>
                        {isAdmin && (
                          <button onClick={() => handleTogglePin(ann)} style={{ fontSize: '12px', color: '#1BB45B', background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 2 }}>
                            {ann.pinned ? <><XCircle className="w-3 h-3" /> {t('group.unpin')}</> : <><Pin className="w-3 h-3" /> {t('group.pin')}</>}
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center" style={{ padding: '40px 24px' }}>
                  <Megaphone className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: '12px' }} />
                  <div style={{ fontSize: '14px', color: '#A2ACB5' }}>{t('group.noAnnouncements')}</div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* ── Action Menu ── */}
        {actionMenu && <ActionMenu x={actionMenu.x} y={actionMenu.y} items={actionMenu.items} onClose={() => setActionMenu(null)} />}

        {/* ── Edit Group Modal ── */}
        <InputModal open={showEditGroup} title={t('group.editGroup')} onClose={() => { setShowEditGroup(false); setEditIconFile(null); setEditIconPreview(''); }} onSubmit={handleEditGroup}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.name')}</label>
            <input value={editName} onChange={(e) => setEditName(e.target.value)} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.desc')}</label>
            <textarea value={editDesc} onChange={(e) => setEditDesc(e.target.value)} rows={2} placeholder={t('group.descPlaceholderEdit')} style={{ ...inputStyle, resize: 'none' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.icon')}</label>
            <div className="flex items-center gap-3">
              {(editIconPreview || editIcon) ? (
                <img src={editIconPreview || editIcon} alt="" style={{ width: 48, height: 48, borderRadius: 8, objectFit: 'cover' }} />
              ) : (
                <div style={{ width: 48, height: 48, borderRadius: 8, background: '#E8EDEF', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, color: '#A2ACB5' }}>+</div>
              )}
              <label style={{ padding: '6px 14px', borderRadius: 8, border: '1px solid rgba(0,0,0,0.1)', background: '#F5F7FA', fontSize: 13, color: '#1BB45B', cursor: 'pointer', fontWeight: 500 }}>
                {t('group.changeImage')}
                <input type="file" accept="image/*" hidden onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    setEditIconFile(file);
                    setEditIconPreview(URL.createObjectURL(file));
                  }
                }} />
              </label>
            </div>
          </div>
          <div className="flex items-center justify-between" style={{ padding: '8px 0' }}>
            <span style={{ fontSize: '13px', color: '#1C2733' }}>{t('group.joinVerify')}</span>
            <button onClick={() => setEditVerify(v => !v)} style={{ width: 44, height: 24, borderRadius: 12, border: 'none', background: editVerify ? '#1BB45B' : '#D1D5DB', cursor: 'pointer', position: 'relative', transition: 'background 0.2s' }}>
              <div style={{ width: 20, height: 20, borderRadius: '50%', background: '#FFF', position: 'absolute', top: 2, left: editVerify ? 22 : 2, transition: 'left 0.2s', boxShadow: '0 1px 3px rgba(0,0,0,0.2)' }} />
            </button>
          </div>
        </InputModal>

        {/* ── Invite Friends Modal ── */}
        <InputModal open={showInviteFriends} title={t('group.inviteFriends')} onClose={() => { setShowInviteFriends(false); setInviteSelected(new Set()); }} onSubmit={handleInviteFriends}>
          <div style={{ marginBottom: '12px' }}>
            <span style={{ fontSize: '13px', color: '#646A73' }}>{t('group.selectedFriends').replace('{count}', String(inviteSelected.size))}</span>
          </div>
          <div style={{ maxHeight: '300px', overflow: 'auto', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.06)' }}>
            {friendsNotInGroup.map(c => (
              <label key={c.id} className="flex items-center gap-3" style={{ padding: '8px 12px', borderBottom: '1px solid rgba(0,0,0,0.04)', cursor: 'pointer' }}>
                <input type="checkbox" checked={inviteSelected.has(c.id)} onChange={() => {
                  setInviteSelected(prev => { const n = new Set(prev); if (n.has(c.id)) n.delete(c.id); else n.add(c.id); return n; });
                }} style={{ width: 16, height: 16, accentColor: '#1BB45B' }} />
                {avatarCircle(c.name, 32, undefined, c.avatar)}
                <span style={{ fontSize: '14px', color: '#1C2733' }}>{c.name}</span>
              </label>
            ))}
            {friendsNotInGroup.length === 0 && (
              <div style={{ padding: '24px', textAlign: 'center', fontSize: '13px', color: '#A2ACB5' }}>{t('group.allFriendsInGroup')}</div>
            )}
          </div>
        </InputModal>

        {/* ── Create Link Modal ── */}
        <InputModal open={showCreateLink} title={t('group.createLinkTitle')} onClose={() => setShowCreateLink(false)} onSubmit={handleCreateLink}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.validity')}</label>
            <select value={newLinkExpiry} onChange={(e) => setNewLinkExpiry(e.target.value)} style={{ ...inputStyle, cursor: 'pointer' }}>
              <option value="1d">{t('group.validity.1d')}</option>
              <option value="7d">{t('group.validity.7d')}</option>
              <option value="30d">{t('group.validity.30d')}</option>
              <option value="never">{t('group.validity.never')}</option>
            </select>
          </div>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.maxUses')}</label>
            <input type="number" min="0" value={newLinkMaxUses} onChange={(e) => setNewLinkMaxUses(e.target.value)} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        {/* ── Create Announcement Modal ── */}
        <InputModal open={showCreateAnnouncement} title={t('group.publishAnnouncement')} onClose={() => { setShowCreateAnnouncement(false); setNewAnnContent(''); }} onSubmit={handleCreateAnnouncement}>
          <div>
            <div className="flex items-center justify-between" style={{ marginBottom: '6px' }}>
              <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733' }}>{t('group.announcementContent')}</label>
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{newAnnContent.length}/500</span>
            </div>
            <textarea value={newAnnContent} onChange={(e) => setNewAnnContent(e.target.value.slice(0, 500))} rows={5} placeholder={t('group.announcementPlaceholder')} style={{ ...inputStyle, resize: 'none' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        {/* ── Member Settings Modal ── */}
        <InputModal open={showMemberSettings} title={t('group.mySettingsTitle')} onClose={() => setShowMemberSettings(false)} onSubmit={handleSaveMemberSettings}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.groupNickname')}</label>
            <input value={settingNickname} onChange={(e) => setSettingNickname(e.target.value)} placeholder={t('group.groupNicknamePlaceholder')} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.groupRemark')}</label>
            <input value={settingRemark} onChange={(e) => setSettingRemark(e.target.value)} placeholder={t('group.groupRemarkPlaceholder')} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />

        {/* ── 加好友弹窗 ── */}
        {addFriendTarget && (
          <div className="fixed inset-0" style={{ zIndex: 10003, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) setAddFriendTarget(null); }}>
            <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
            <div className="relative" style={{ background: '#FFF', borderRadius: '16px', padding: '24px', width: '90%', maxWidth: '400px', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
              <button onClick={() => setAddFriendTarget(null)} className="absolute" style={{ top: 16, right: 16, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <X className="w-4 h-4" />
              </button>
              <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', marginBottom: '16px' }}>{t('group.addFriendTitle')}</h3>
              <div className="flex items-center gap-3" style={{ marginBottom: '16px', padding: '12px', background: '#F5F7FA', borderRadius: '12px' }}>
                {avatarCircle(getContactName(addFriendTarget.userId), 48, undefined, addFriendTarget.avatar)}
                <div>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#1C2733' }}>{getContactName(addFriendTarget.userId)}</div>
                  <div style={{ fontSize: '12px', color: '#A2ACB5' }}>ID: {addFriendTarget.userId}</div>
                </div>
              </div>
              <div style={{ marginBottom: '16px' }}>
                <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.applyMsg')}</label>
                <textarea value={addFriendMsg} onChange={(e) => setAddFriendMsg(e.target.value)} rows={2} placeholder={t('group.applyMsgPlaceholder')} style={{ width: '100%', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.1)', padding: '10px 12px', fontSize: '14px', color: '#1C2733', outline: 'none', resize: 'none', background: '#F5F7FA', boxSizing: 'border-box' }} onFocus={(e) => { e.target.style.borderColor = '#1BB45B'; e.target.style.boxShadow = '0 0 0 3px rgba(27,180,91,0.15)'; }} onBlur={(e) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; e.target.style.boxShadow = 'none'; }} />
              </div>
              <div className="flex items-center justify-end gap-3">
                <button onClick={() => setAddFriendTarget(null)} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: 'pointer' }}>{t('common.cancel')}</button>
                <button disabled={addFriendLoading} onClick={async () => {
                  setAddFriendLoading(true);
                  try {
                    const res = await apiFetch('/api/social/friend/putIn', token, {
                      method: 'POST',
                      body: JSON.stringify({ user_uid: addFriendTarget.userId, req_msg: addFriendMsg || '' }),
                    });
                    if (res.success) {
                      toast.success(t('group.friendReqSentToast'));
                      setSentFriendReqs(prev => new Set(prev).add(addFriendTarget.userId));
                      setAddFriendTarget(null);
                    } else {
                      toast.error(res.message || t('group.sendFailed'));
                    }
                  } catch { toast.error(t('group.sendFailed')); }
                  finally { setAddFriendLoading(false); }
                }} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: addFriendLoading ? 'not-allowed' : 'pointer', opacity: addFriendLoading ? 0.7 : 1 }}>
                  {addFriendLoading ? t('group.sending') : t('group.sendApply')}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* ── 举报弹窗 ── */}
        {reportTarget && (
          <div className="fixed inset-0" style={{ zIndex: 10003, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) setReportTarget(null); }}>
            <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
            <div className="relative" style={{ background: '#FFF', borderRadius: '16px', padding: '24px', width: '90%', maxWidth: '400px', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
              <button onClick={() => setReportTarget(null)} className="absolute" style={{ top: 16, right: 16, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <X className="w-4 h-4" />
              </button>
              <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#E53935', marginBottom: '16px' }}>{t('group.reportUser')}</h3>
              <div className="flex items-center gap-3" style={{ marginBottom: '16px', padding: '12px', background: '#FFF5F5', borderRadius: '12px' }}>
                {avatarCircle(getContactName(reportTarget.userId), 48, undefined, reportTarget.avatar)}
                <div>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#1C2733' }}>{getContactName(reportTarget.userId)}</div>
                  <div style={{ fontSize: '12px', color: '#A2ACB5' }}>ID: {reportTarget.userId}</div>
                </div>
              </div>
              <div style={{ marginBottom: '16px' }}>
                <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>{t('group.reportReason')}</label>
                <textarea value={reportReason} onChange={(e) => setReportReason(e.target.value)} rows={3} placeholder={t('group.reportReasonPlaceholder')} style={{ width: '100%', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.1)', padding: '10px 12px', fontSize: '14px', color: '#1C2733', outline: 'none', resize: 'none', background: '#F5F7FA', boxSizing: 'border-box' }} onFocus={(e) => { e.target.style.borderColor = '#E53935'; e.target.style.boxShadow = '0 0 0 3px rgba(229,57,53,0.15)'; }} onBlur={(e) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; e.target.style.boxShadow = 'none'; }} />
              </div>
              <div className="flex items-center justify-end gap-3">
                <button onClick={() => setReportTarget(null)} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: 'pointer' }}>{t('common.cancel')}</button>
                <button disabled={reportLoading || !reportReason.trim()} onClick={async () => {
                  setReportLoading(true);
                  try {
                    const res = await apiFetch('/api/social/friend/report', token, {
                      method: 'POST',
                      body: JSON.stringify({ friend_uid: reportTarget.userId, reason: reportReason }),
                    });
                    if (res.success) {
                      toast.success(t('group.reportSubmittedToast'));
                      setReportedUsers(prev => new Set(prev).add(reportTarget.userId));
                      setReportTarget(null);
                    } else {
                      toast.error(res.message || t('group.reportFailed'));
                    }
                  } catch { toast.error(t('group.reportFailed')); }
                  finally { setReportLoading(false); }
                }} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: '#E53935', color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: (reportLoading || !reportReason.trim()) ? 'not-allowed' : 'pointer', opacity: (reportLoading || !reportReason.trim()) ? 0.5 : 1 }}>
                  {reportLoading ? t('group.submitting') : t('group.submitReport')}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* ── 成员资料卡弹窗 ── */}
        {profileTarget && (() => {
          const pm = profileTarget;
          const pmName = getContactName(pm.userId);
          const pmUser = members.find(m => m.id === pm.id);
          const isFriend = friends.some(f => f.id === pm.userId);
          const alreadySent = sentFriendReqs.has(pm.userId);
          return (
            <div className="fixed inset-0" style={{ zIndex: 10003, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) setProfileTarget(null); }}>
              <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
              <div className="relative" style={{ background: '#FFF', borderRadius: '16px', padding: '24px', width: '90%', maxWidth: '400px', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
                <button onClick={() => setProfileTarget(null)} className="absolute" style={{ top: 16, right: 16, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <X className="w-4 h-4" />
                </button>

                {/* 头像 + 基本信息 */}
                <div className="flex items-center gap-4" style={{ marginBottom: '16px' }}>
                  <div className="shrink-0" style={{ width: 64, height: 64 }}>
                    {pm.avatar ? (
                      <img src={pm.avatar} alt={pmName} style={{ width: 64, height: 64, borderRadius: '50%', objectFit: 'cover' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }} />
                    ) : null}
                    <div style={{ width: 64, height: 64, borderRadius: '50%', background: getAvatarColor(pmName), display: pm.avatar ? 'none' : 'flex', alignItems: 'center', justifyContent: 'center', color: '#FFF', fontSize: 24, fontWeight: 600 }}>{pmName[0]}</div>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2" style={{ marginBottom: 4 }}>
                      <span style={{ fontSize: '18px', fontWeight: 600, color: '#1C2733' }}>{pmName}</span>
                      {pm.roleLevel >= 1 && (
                        <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[pm.roleLevel], backgroundColor: roleBg[pm.roleLevel], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                          {roleIcon[pm.roleLevel]}{t(roleLabelKey[pm.roleLevel])}
                        </span>
                      )}
                    </div>
                    <div style={{ fontSize: '12px', color: '#A2ACB5' }}>ID: {pm.userId}</div>
                    {pm.groupNickname && <div style={{ fontSize: '12px', color: '#708499', marginTop: 2 }}>{t('group.groupNicknameLabel').replace('{name}', pm.groupNickname)}</div>}
                  </div>
                </div>

                {/* 好友状态 */}
                {isFriend ? (
                  <div style={{ padding: '10px 14px', background: 'rgba(77,205,94,0.08)', borderRadius: '10px', marginBottom: '16px', fontSize: '13px', color: '#4DCD5E', display: 'flex', alignItems: 'center', gap: 6 }}>
                    <UserCheck className="w-4 h-4" /> {t('group.isFriend')}
                  </div>
                ) : alreadySent ? (
                  <div style={{ padding: '10px 14px', background: 'rgba(27,180,91,0.06)', borderRadius: '10px', marginBottom: '16px', fontSize: '13px', color: '#1BB45B', display: 'flex', alignItems: 'center', gap: 6 }}>
                    <Clock className="w-4 h-4" /> {t('group.friendReqSent')}
                  </div>
                ) : null}

                {/* 底部操作按钮 */}
                <div className="flex items-center gap-2">
                  {isFriend ? (
                    <button onClick={async () => {
                      const memberId = profileTarget?.userId || profileTarget?.id;
                      setProfileTarget(null);
                      if (!currentUser?.token || !currentUser?.id || !memberId) return;
                      try {
                        const conv = await useChatStore.getState().getOrCreateConversation(currentUser.token, currentUser.id, String(memberId));
                        setShowGroupPanel(false);
                        setActiveTab('chats');
                        setSelectedConversationId(conv.id);
                        setShowChatDetail(true);
                      } catch { toast.error(t('group.openConvFailed')); }
                    }} style={{ flex: 1, padding: '10px', borderRadius: '10px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}>
                      <Send className="w-4 h-4" /> {t('group.sendMessage')}
                    </button>
                  ) : !alreadySent ? (
                    <button onClick={() => { setProfileTarget(null); setAddFriendTarget(pm); setAddFriendMsg(t('group.applyMsgFromGroup').replace('{name}', selectedGroup?.name || '')); }} style={{ flex: 1, padding: '10px', borderRadius: '10px', border: 'none', background: '#1BB45B', color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}>
                      <UserPlus className="w-4 h-4" /> {t('group.addFriend')}
                    </button>
                  ) : null}
                  <button onClick={() => { setProfileTarget(null); setReportTarget(pm); setReportReason(''); }} style={{ padding: '10px 16px', borderRadius: '10px', border: '1px solid #E53935', background: '#FFF', color: '#E53935', fontSize: '14px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}>
                    <Flag className="w-4 h-4" /> {t('group.report')}
                  </button>
                </div>
              </div>
            </div>
          );
        })()}
      </div>
    );
  }

  return null;
}
