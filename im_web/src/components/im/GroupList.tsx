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
} from 'lucide-react';
import { toast } from 'sonner';
import { useIMStore } from '@/lib/im-store';
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
} from '@/lib/mock-data';

/* ═══════════════════════════════════════
   Helpers
   ═══════════════════════════════════════ */

function fmtTime(date: Date): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const m = Math.floor(diff / 60000);
  const h = Math.floor(diff / 3600000);
  const d = Math.floor(diff / 86400000);
  if (m < 1) return '刚刚';
  if (m < 60) return `${m}分钟前`;
  if (h < 24) return `${h}小时前`;
  if (d < 30) return `${new Date(date).getMonth() + 1}月${new Date(date).getDate()}日`;
  return `${new Date(date).getFullYear()}年${new Date(date).getMonth() + 1}月${new Date(date).getDate()}日`;
}

const roleLabel: Record<GroupRoleLevel, string> = { 1: '成员', 2: '管理员', 3: '群主' };
const roleIcon: Record<GroupRoleLevel, React.ReactNode> = {
  1: <Shield className="w-3 h-3" />,
  2: <ShieldCheck className="w-3 h-3" />,
  3: <Crown className="w-3 h-3" />,
};
const roleColor: Record<GroupRoleLevel, string> = { 1: '#A2ACB5', 2: '#3390EC', 3: '#F5A623' };
const roleBg: Record<GroupRoleLevel, string> = {
  1: 'rgba(162,172,181,0.1)',
  2: 'rgba(51,144,236,0.1)',
  3: 'rgba(245,166,35,0.1)',
};

const joinSourceLabel: Record<GroupJoinSource, string> = { 1: '申请', 2: '邀请', 3: '链接' };

const appResultConfig: Record<GroupAppResult, { label: string; color: string; bg: string; icon: React.ReactNode }> = {
  0: { label: '待处理', color: '#F5A623', bg: 'rgba(245,166,35,0.1)', icon: <AlertCircle className="w-3.5 h-3.5" /> },
  1: { label: '已通过', color: '#4DCD5E', bg: 'rgba(77,205,94,0.1)', icon: <CheckCircle className="w-3.5 h-3.5" /> },
  2: { label: '已拒绝', color: '#FF5252', bg: 'rgba(255,82,82,0.1)', icon: <XCircle className="w-3.5 h-3.5" /> },
  3: { label: '已忽略', color: '#A2ACB5', bg: 'rgba(162,172,181,0.1)', icon: <MinusCircle className="w-3.5 h-3.5" /> },
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
    isVerify: !!g.is_verify,
    notification: g.notification || '',
    createUid: String(g.create_uid || ''),
  };
}

function mapMember(m: any, groupId: string): GroupMemberInfo {
  return {
    id: m.id,
    groupId: String(groupId),
    userId: String(m.user_id),
    nickname: m.nickname || (m.user?.nickname) || '',
    roleLevel: (m.role_level || 1) as GroupRoleLevel,
    online: false, // online status not provided by this endpoint
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
    reqTime: new Date(a.req_time || Date.now()),
    joinSource: (a.join_source || 1) as GroupJoinSource,
    inviterName: a.inviter_user_id ? String(a.inviter_user_id) : undefined,
    handleResult: (a.handle_result ?? 0) as GroupAppResult,
    readState: (a.handle_result ?? 0) !== 0,
  };
}

function mapInviteLink(l: any): GroupInviteLink {
  return {
    token: l.token || l.id || '',
    groupId: String(l.group_id),
    createdBy: String(l.created_by || l.creator_uid || ''),
    createdAt: new Date(l.created_at || l.create_time || Date.now()),
    expireAt: l.expire_at || l.expire_time ? new Date(l.expire_at || l.expire_time) : null,
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
    createdAt: new Date(a.created_at || a.create_time || Date.now()),
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
          <button onClick={opts.onCancel} disabled={opts.loading} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: opts.loading ? 'not-allowed' : 'pointer' }}>取消</button>
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
          <button onClick={onClose} style={{ padding: '8px 20px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '14px', fontWeight: 500, cursor: 'pointer' }}>取消</button>
          <button onClick={onSubmit} style={{ padding: '8px 20px', borderRadius: '8px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '14px', fontWeight: 500, cursor: 'pointer' }}>确定</button>
        </div>
      </div>
    </div>
  );
}

const inputStyle = { width: '100%', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.1)', padding: '10px 12px', fontSize: '14px', color: '#1C2733', outline: 'none', background: '#F5F7FA', boxSizing: 'border-box' as const };
const focusInput = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => { e.target.style.borderColor = '#3390EC'; e.target.style.boxShadow = '0 0 0 3px rgba(51,144,236,0.15)'; };
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
  const { setShowGroupPanel, groupAppUnreadCount, setGroupAppUnreadCount, currentUser, friends } = useIMStore();
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
  const [newGroupIcon, setNewGroupIcon] = useState('');
  const [editName, setEditName] = useState('');
  const [editIcon, setEditIcon] = useState('');
  const [editNotification, setEditNotification] = useState('');
  const [editVerify, setEditVerify] = useState(false);
  const [inviteSelected, setInviteSelected] = useState<Set<string>>(new Set());
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
    if (userId === myUserId) return currentUser?.name || '我';
    if (memberNameCache.current[userId]) return memberNameCache.current[userId];
    const friend = friends.find(f => f.id === userId);
    if (friend) return friend.name;
    return userId;
  }, [myUserId, currentUser?.name, friends]);

  // ── API: Fetch groups list ──
  const fetchGroups = useCallback(async () => {
    if (!token) return;
    try {
      const data = await apiFetch('/api/social/groups', token);
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
  }, [token]);

  // ── API: Fetch applications ──
  const fetchApplications = useCallback(async () => {
    if (!token) return;
    try {
      const data = await apiFetch('/api/social/group/putInsByUid?class=2', token);
      if (data.success && data.data?.list) {
        setApps(data.data.list.map(mapApplication));
      } else {
        setApps([]);
      }
    } catch {
      setApps([]);
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
  }, [fetchGroups]);

  // ── When entering app view: fetch applications ──
  useEffect(() => {
    if (view === 'app') {
      fetchApplications();
    }
  }, [view, fetchApplications]);

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
  const unreadCount = useMemo(() => apps.filter(a => a.userId !== myUserId && a.handleResult === 0 && !a.readState).length, [apps, myUserId]);
  useEffect(() => { setGroupAppUnreadCount(unreadCount); }, [unreadCount, setGroupAppUnreadCount]);

  const filteredGroups = useMemo(() => {
    if (!listSearch.trim()) return groups;
    const q = listSearch.toLowerCase();
    return groups.filter(g => g.name.toLowerCase().includes(q));
  }, [groups, listSearch]);

  // All groups returned by the API are the user's groups, so myGroupIds = all group ids
  const myGroupIds = useMemo(() => groups.map(g => g.id), [groups]);

  const selectedGroup = useMemo(() => groups.find(g => g.id === selectedGroupId) || null, [groups, selectedGroupId]);

  const myRole = useMemo((): GroupRoleLevel => {
    if (!selectedGroupId) return 1;
    const m = members.find(m => m.groupId === selectedGroupId && m.userId === myUserId);
    return m ? m.roleLevel : 1;
  }, [members, selectedGroupId, myUserId]);

  const isAdmin = myRole >= 2;
  const isOwner = myRole === 3;

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
    doConfirm('同意入群', `确定同意 ${app.userName} 加入 ${app.groupName} 吗？`, '同意', '#3390EC', async () => {
      try {
        const data = await apiFetch('/api/social/group/putIn', token, {
          method: 'PUT',
          body: JSON.stringify({ group_req_id: app.id, group_id: app.groupId, handle_result: 1 }),
        });
        if (data.success) {
          setApps(prev => prev.map(a => a.id === app.id ? { ...a, handleResult: 1 as GroupAppResult, readState: true } : a));
          toast.success(`已同意 ${app.userName} 加入 ${app.groupName}`);
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [doConfirm, markAppRead, token]);

  const handleRejectApp = useCallback((app: GroupApplication) => {
    markAppRead(app.id);
    doConfirm('拒绝入群', `确定拒绝 ${app.userName} 加入 ${app.groupName} 吗？`, '拒绝', '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/putIn', token, {
          method: 'PUT',
          body: JSON.stringify({ group_req_id: app.id, group_id: app.groupId, handle_result: 2 }),
        });
        if (data.success) {
          setApps(prev => prev.map(a => a.id === app.id ? { ...a, handleResult: 2 as GroupAppResult, readState: true } : a));
          toast.success(`已拒绝 ${app.userName} 的申请`);
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [doConfirm, markAppRead, token]);

  const markAllAppRead = useCallback(() => {
    setApps(prev => prev.map(a => a.userId !== myUserId ? { ...a, readState: true } : a));
    toast.success('已全部标记为已读');
  }, [myUserId]);

  // Create group
  const handleCreateGroup = useCallback(async () => {
    if (!newGroupName.trim()) { toast.error('请输入群名称'); return; }
    try {
      setLoading(true);
      const body: any = { name: newGroupName.trim() };
      if (newGroupIcon.trim()) body.icon = newGroupIcon.trim();
      const data = await apiFetch('/api/social/group', token, {
        method: 'POST',
        body: JSON.stringify(body),
      });
      if (data.success) {
        toast.success('群组创建成功');
        setShowCreateGroup(false);
        setNewGroupName('');
        setNewGroupIcon('');
        // Refresh groups list
        await fetchGroups();
      } else {
        toast.error(data.message || '创建失败');
      }
    } catch {
      toast.error('网络错误');
    } finally {
      setLoading(false);
    }
  }, [newGroupName, newGroupIcon, token, fetchGroups]);

  // Edit group
  const openEditGroup = useCallback(() => {
    if (!selectedGroup) return;
    setEditName(selectedGroup.name);
    setEditIcon(selectedGroup.icon);
    setEditNotification(selectedGroup.notification);
    setEditVerify(selectedGroup.isVerify);
    setShowEditGroup(true);
  }, [selectedGroup]);

  const handleEditGroup = useCallback(async () => {
    if (!selectedGroupId) return;
    try {
      const data = await apiFetch('/api/social/group/update', token, {
        method: 'POST',
        body: JSON.stringify({
          group_id: selectedGroupId,
          name: editName,
          icon: editIcon,
          notification: editNotification,
          is_verify: editVerify ? 1 : 0,
        }),
      });
      if (data.success) {
        setGroups(prev => prev.map(g => g.id === selectedGroupId ? { ...g, name: editName, icon: editIcon, notification: editNotification, isVerify: editVerify } : g));
        setShowEditGroup(false);
        toast.success('群信息已更新');
      } else {
        toast.error(data.message || '更新失败');
      }
    } catch {
      toast.error('网络错误');
    }
  }, [selectedGroupId, editName, editIcon, editNotification, editVerify, token]);

  // Send message
  const handleSendMessage = useCallback(() => { toast.success('正在打开聊天...'); }, []);

  // Invite friends
  const handleInviteFriends = useCallback(async () => {
    if (inviteSelected.size === 0) { toast.error('请选择要邀请的好友'); return; }
    if (!selectedGroupId) return;
    try {
      const data = await apiFetch('/api/social/group/invite', token, {
        method: 'POST',
        body: JSON.stringify({ group_id: selectedGroupId, friend_ids: [...inviteSelected] }),
      });
      if (data.success) {
        const names = [...inviteSelected].map(id => getContactName(id)).join('、');
        toast.success(`已邀请 ${names} 加入群组`);
        setShowInviteFriends(false);
        setInviteSelected(new Set());
        // Refresh group detail
        fetchGroupDetail(selectedGroupId);
      } else {
        toast.error(data.message || '邀请失败');
      }
    } catch {
      toast.error('网络错误');
    }
  }, [inviteSelected, selectedGroupId, token, getContactName, fetchGroupDetail]);

  // Member settings
  const openMemberSettings = useCallback(() => {
    if (!selectedGroupId) return;
    const s = settings.find(s => s.groupId === selectedGroupId);
    setSettingNickname(s?.groupNickname || '');
    setSettingRemark(s?.groupRemark || '');
    setShowMemberSettings(true);
  }, [selectedGroupId, settings]);

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
        setShowMemberSettings(false);
        toast.success('设置已保存');
      } else {
        toast.error(data.message || '保存失败');
      }
    } catch {
      toast.error('网络错误');
    }
  }, [selectedGroupId, settingNickname, settingRemark, token]);

  // Disband / Quit group
  const handleDisband = useCallback(() => {
    if (!selectedGroup) return;
    doConfirm('解散群组', `确定解散「${selectedGroup.name}」吗？此操作不可撤销！`, '解散群组', '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/disband', token, {
          method: 'POST',
          body: JSON.stringify({ group_id: selectedGroupId }),
        });
        if (data.success) {
          setGroups(prev => prev.filter(g => g.id !== selectedGroupId));
          setMembers(prev => prev.filter(m => m.groupId !== selectedGroupId));
          setSettings(prev => prev.filter(s => s.groupId !== selectedGroupId));
          toast.success('群组已解散');
          setView('list');
          setSelectedGroupId(null);
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [selectedGroup, selectedGroupId, doConfirm, token]);

  const handleQuitGroup = useCallback(() => {
    if (!selectedGroup) return;
    doConfirm('退出群组', `确定退出「${selectedGroup.name}」吗？`, '退出群组', '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/quit', token, {
          method: 'POST',
          body: JSON.stringify({ group_id: selectedGroupId }),
        });
        if (data.success) {
          setGroups(prev => prev.filter(g => g.id !== selectedGroupId));
          setMembers(prev => prev.filter(m => m.groupId !== selectedGroupId));
          setSettings(prev => prev.filter(s => s.groupId !== selectedGroupId));
          toast.success('已退出群组');
          setView('list');
          setSelectedGroupId(null);
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [selectedGroup, selectedGroupId, doConfirm, token]);

  // Member actions
  const handleSetAdmin = useCallback((m: GroupMemberInfo) => {
    const newRole: GroupRoleLevel = m.roleLevel === 2 ? 1 : 2;
    const label = newRole === 2 ? '设为管理员' : '取消管理员';
    doConfirm(label, `确定将 ${getContactName(m.userId)} ${label}吗？`, label, '#3390EC', async () => {
      try {
        const data = await apiFetch('/api/social/group/setAdmin', token, {
          method: 'POST',
          body: JSON.stringify({
            group_id: selectedGroupId,
            member_ids: [m.userId],
            is_admin: newRole === 2,
          }),
        });
        if (data.success) {
          setMembers(prev => prev.map(x => x.id === m.id ? { ...x, roleLevel: newRole } : x));
          toast.success(`已${label}`);
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [doConfirm, getContactName, selectedGroupId, token]);

  const handleTransferOwner = useCallback((m: GroupMemberInfo) => {
    doConfirm('转让群主', `确定将群主转让给 ${getContactName(m.userId)} 吗？转让后你将变为普通成员。`, '转让', '#F5A623', async () => {
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
          toast.success('群主已转让');
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [doConfirm, getContactName, selectedGroupId, myUserId, token]);

  const handleKickMember = useCallback((m: GroupMemberInfo) => {
    doConfirm('踢出群组', `确定将 ${getContactName(m.userId)} 移出群组吗？`, '踢出', '#E53935', async () => {
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
          toast.success(`已将 ${getContactName(m.userId)} 移出群组`);
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [doConfirm, getContactName, selectedGroupId, token]);

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
        toast.success('邀请链接已创建');
        // Refresh links
        fetchInviteLinks(selectedGroupId);
      } else {
        toast.error(data.message || '创建失败');
      }
    } catch {
      toast.error('网络错误');
    }
  }, [selectedGroupId, newLinkExpiry, newLinkMaxUses, token, fetchInviteLinks]);

  const handleRevokeLink = useCallback((link: GroupInviteLink) => {
    doConfirm('撤销链接', '确定撤销此邀请链接吗？', '撤销', '#E53935', async () => {
      try {
        const data = await apiFetch('/api/social/group/inviteLink/revoke', token, {
          method: 'POST',
          body: JSON.stringify({ token: link.token }),
        });
        if (data.success) {
          setLinks(prev => prev.map(l => l.token === link.token ? { ...l, revoked: true } : l));
          toast.success('链接已撤销');
        } else {
          toast.error(data.message || '操作失败');
        }
      } catch {
        toast.error('网络错误');
      }
    });
  }, [doConfirm, token]);

  const handleCopyLink = useCallback((linkToken: string) => {
    navigator.clipboard.writeText(`https://hichat.app/join/${linkToken}`).catch(() => {});
    toast.success('链接已复制');
  }, []);

  // Announcements
  const handleCreateAnnouncement = useCallback(async () => {
    if (!selectedGroupId || !newAnnContent.trim()) { toast.error('请输入公告内容'); return; }
    try {
      const data = await apiFetch('/api/social/group/announcement', token, {
        method: 'POST',
        body: JSON.stringify({ group_id: selectedGroupId, content: newAnnContent }),
      });
      if (data.success) {
        setShowCreateAnnouncement(false);
        setNewAnnContent('');
        toast.success('公告已发布');
        // Refresh announcements
        fetchAnnouncements(selectedGroupId);
      } else {
        toast.error(data.message || '发布失败');
      }
    } catch {
      toast.error('网络错误');
    }
  }, [selectedGroupId, newAnnContent, token, fetchAnnouncements]);

  const handleTogglePin = useCallback((ann: GroupAnnouncement) => {
    setAnnouncements(prev => prev.map(a => a.id === ann.id ? { ...a, pinned: !a.pinned } : a));
    toast.success(ann.pinned ? '已取消置顶' : '已置顶');
  }, []);

  // Link status
  const getLinkStatus = useCallback((link: GroupInviteLink): { label: string; color: string } => {
    if (link.revoked) return { label: '已撤销', color: '#A2ACB5' };
    if (link.expireAt && new Date() > link.expireAt) return { label: '已过期', color: '#FF5252' };
    if (link.maxUses > 0 && link.usedCount >= link.maxUses) return { label: '已用完', color: '#FF5252' };
    return { label: '有效', color: '#4DCD5E' };
  }, []);

  // ── Render ──

  const renderHeader = (title: string, rightContent?: React.ReactNode) => (
    <div className="flex items-center justify-between shrink-0" style={{ height: 56, background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.08)', paddingLeft: 4, paddingRight: 16 }}>
      <button onClick={handleBack} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#3390EC', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <ArrowLeft className="w-5 h-5" />
      </button>
      <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>{title}</span>
      <div className="flex items-center gap-1">{rightContent}</div>
    </div>
  );

  const pillTab = (active: boolean, onClick: () => void, label: string, badge?: number) => (
    <button onClick={onClick} style={{ padding: '6px 20px', borderRadius: '17px', border: 'none', background: active ? '#3390EC' : 'transparent', color: active ? '#FFF' : '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', gap: 4 }}>
      {label}
      {badge !== undefined && badge > 0 && !active && (
        <span className="inline-flex items-center justify-center" style={{ width: 16, height: 16, borderRadius: 8, background: '#E53935', color: '#FFF', fontSize: '10px', fontWeight: 700 }}>{badge > 9 ? '9+' : badge}</span>
      )}
    </button>
  );

  const filterPill = (active: boolean, onClick: () => void, label: string) => (
    <button onClick={onClick} style={{ padding: '4px 14px', borderRadius: '14px', border: 'none', background: active ? 'rgba(51,144,236,0.1)' : 'rgba(0,0,0,0.04)', color: active ? '#3390EC' : '#646A73', fontSize: '12px', fontWeight: active ? 500 : 400, cursor: 'pointer', whiteSpace: 'nowrap', transition: 'all 0.15s' }}>{label}</button>
  );

  const avatarCircle = (name: string, size: number, extra?: React.ReactNode) => (
    <div className="relative shrink-0" style={{ width: size, height: size, borderRadius: '50%', backgroundColor: getAvatarColor(name || '?'), display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: size * 0.38, fontWeight: 600, color: '#FFF' }}>
      {(name || '?')[0]}
      {extra}
    </div>
  );

  // ═══ VIEW: List ═══
  if (view === 'list') {
    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {renderHeader('我的群组',
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
            <input value={listSearch} onChange={(e) => setListSearch(e.target.value)} placeholder="搜索群组" style={{ ...inputStyle, paddingLeft: '36px', borderRadius: '22px', height: '36px', fontSize: '13px' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </div>

        {/* Group List */}
        <div className="flex-1 overflow-y-auto im-scroll">
          {filteredGroups.filter(g => myGroupIds.includes(g.id)).length > 0 ? (
            <div style={{ background: '#FFF', borderRadius: '12px', margin: '0 8px 8px', overflow: 'hidden' }}>
              {filteredGroups.filter(g => myGroupIds.includes(g.id)).map(group => {
                const gm = members.filter(m => m.groupId === group.id);
                const myM = gm.find(m => m.userId === myUserId);
                const role = myM?.roleLevel || 1;
                return (
                  <div key={group.id} className="flex items-center gap-3" style={{ padding: '12px 16px', borderBottom: '1px solid rgba(0,0,0,0.05)', cursor: 'pointer' }} onClick={() => openGroup(group.id)} onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)'; }} onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}>
                    <img src={group.icon} alt="" className="shrink-0" style={{ width: 48, height: 48, borderRadius: '50%' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2" style={{ marginBottom: 2 }}>
                        <span style={{ fontSize: '15px', fontWeight: 600, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{group.name}</span>
                        {role >= 2 && (
                          <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[role], backgroundColor: roleBg[role], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                            {roleIcon[role]}{roleLabel[role]}
                          </span>
                        )}
                        {group.isVerify && (
                          <span style={{ fontSize: '10px', color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 5px' }}>需验证</span>
                        )}
                      </div>
                      <div style={{ fontSize: '12px', color: '#A2ACB5' }}>{gm.length > 0 ? `${gm.length} 位成员` : '群组'}</div>
                      {group.notification && (
                        <div style={{ fontSize: '12px', color: '#708499', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginTop: 2 }}>
                          {group.notification}
                        </div>
                      )}
                    </div>
                    <button onClick={(e) => { e.stopPropagation(); handleSendMessage(); }} style={{ padding: '5px 14px', borderRadius: '16px', border: '1px solid #3390EC', background: 'transparent', color: '#3390EC', fontSize: '12px', fontWeight: 500, cursor: 'pointer', whiteSpace: 'nowrap', display: 'flex', alignItems: 'center', gap: 4 }}>
                      <Send className="w-3 h-3" /> 发消息
                    </button>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
              <Users className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: '16px' }} />
              <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: '8px' }}>{listSearch ? '没有匹配的群组' : '暂无群组'}</div>
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{listSearch ? '换个关键词试试' : '点击右上角创建一个群组吧'}</div>
            </div>
          )}
        </div>

        {/* Modals */}
        <InputModal open={showCreateGroup} title="创建群组" onClose={() => setShowCreateGroup(false)} onSubmit={handleCreateGroup}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群名称 *</label>
            <input value={newGroupName} onChange={(e) => setNewGroupName(e.target.value)} placeholder="输入群名称" style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群图标 (URL, 可选)</label>
            <input value={newGroupIcon} onChange={(e) => setNewGroupIcon(e.target.value)} placeholder="留空自动生成" style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />
      </div>
    );
  }

  // ═══ VIEW: Applications ═══
  if (view === 'app') {
    const statusFilters: { key: AppStatusFilter; label: string }[] = [
      { key: 'all', label: '全部' }, { key: 0, label: '待处理' }, { key: 1, label: '已通过' }, { key: 2, label: '已拒绝' },
    ];
    const getEmptyMsg = () => {
      if (appStatusFilter === 'all') return { t: appClass === 'received' ? '暂无收到的申请' : '暂无发出的申请', d: '这里空空如也' };
      return { t: '暂无匹配的申请', d: '换个筛选条件看看' };
    };
    const emptyMsg = getEmptyMsg();

    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {renderHeader('群申请',
          <button onClick={markAllAppRead} style={{ fontSize: '13px', color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 500 }}>全部已读</button>
        )}

        {/* Class tabs */}
        <div className="flex items-center shrink-0" style={{ padding: '12px 16px 8px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
          <div className="flex items-center" style={{ borderRadius: '20px', background: 'rgba(0,0,0,0.04)', padding: '3px' }}>
            {pillTab(appClass === 'received', () => { setAppClass('received'); setAppStatusFilter('all'); }, '我收到的', unreadCount)}
            {pillTab(appClass === 'sent', () => { setAppClass('sent'); setAppStatusFilter('all'); }, '我发起的')}
          </div>
        </div>

        {/* Status filter */}
        <div className="flex items-center gap-2 shrink-0 overflow-x-auto" style={{ padding: '8px 16px 10px', background: '#FFF', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
          {statusFilters.map(f => filterPill(appStatusFilter === f.key, () => setAppStatusFilter(f.key), f.label))}
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
                        avatarCircle(app.userName, 44, <div className="absolute" style={{ bottom: -2, left: '50%', transform: 'translateX(-50%)', width: 20, height: 3, borderRadius: 2, backgroundColor: rc.color }} />)
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
                          <span style={{ fontSize: '10px', color: '#708499', backgroundColor: 'rgba(0,0,0,0.04)', borderRadius: '3px', padding: '0 4px' }}>{joinSourceLabel[app.joinSource]}</span>
                          <span style={{ fontSize: '11px', fontWeight: 500, color: rc.color, backgroundColor: rc.bg, borderRadius: '4px', padding: '1px 6px', display: 'flex', alignItems: 'center', gap: 3 }}>
                            {React.cloneElement(rc.icon as React.ReactElement<any>, { style: { color: rc.color, width: 12, height: 12 } })}{rc.label}
                          </span>
                          <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', whiteSpace: 'nowrap' }}>{fmtTime(app.reqTime)}</span>
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
                          <div style={{ fontSize: '12px', color: '#708499', marginBottom: 4 }}>邀请人: {app.inviterName}</div>
                        )}
                        {/* Message */}
                        <div style={{ fontSize: '13px', color: '#646A73', lineHeight: '1.5', marginBottom: isReceived && isPending ? '6px' : 0 }}>{app.reqMsg}</div>
                        {/* Action buttons */}
                        {isReceived && isPending && (
                          <div className="flex items-center gap-2" style={{ marginTop: 6 }}>
                            <button onClick={() => handleAcceptApp(app)} style={{ padding: '5px 16px', borderRadius: '6px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '12px', fontWeight: 500, cursor: 'pointer' }}>同意</button>
                            <button onClick={() => handleRejectApp(app)} style={{ padding: '5px 16px', borderRadius: '6px', border: '1px solid #FF5252', background: '#FFF', color: '#FF5252', fontSize: '12px', fontWeight: 500, cursor: 'pointer' }}>拒绝</button>
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
      { key: 'members', label: `成员 (${groupMembers.length})`, show: true },
      { key: 'links', label: '邀请链接', show: isAdmin },
      { key: 'announcements', label: '公告', show: true },
    ];

    return (
      <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
        {renderHeader(selectedGroup.name)}

        <div className="flex-1 overflow-y-auto im-scroll">
          {/* ── Group Info Card ── */}
          <div style={{ background: '#FFF', borderRadius: '12px', margin: '8px', padding: '20px' }}>
            <div className="flex items-start gap-4">
              <img src={selectedGroup.icon} alt="" className="shrink-0" style={{ width: 72, height: 72, borderRadius: '50%' }} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap" style={{ marginBottom: 4 }}>
                  <span style={{ fontSize: '18px', fontWeight: 700, color: '#1C2733' }}>{selectedGroup.name}</span>
                  {myRole >= 2 && (
                    <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[myRole], backgroundColor: roleBg[myRole], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                      {roleIcon[myRole]}{roleLabel[myRole]}
                    </span>
                  )}
                  {selectedGroup.isVerify && (
                    <span style={{ fontSize: '10px', color: '#F5A623', backgroundColor: 'rgba(245,166,35,0.1)', borderRadius: '4px', padding: '1px 5px' }}>需验证</span>
                  )}
                </div>
                <div style={{ fontSize: '13px', color: '#A2ACB5', marginBottom: 4 }}>
                  {groupMembers.length} 位成员{onlineCount > 0 ? ` · ${onlineCount} 人在线` : ''}
                </div>
                {mySetting?.groupRemark && (
                  <div style={{ fontSize: '12px', color: '#708499', marginBottom: 4 }}>备注: {mySetting.groupRemark}</div>
                )}
                {selectedGroup.notification && (
                  <div style={{ fontSize: '12px', color: '#708499', padding: '6px 10px', background: 'rgba(0,0,0,0.03)', borderRadius: '8px', marginTop: 6 }}>
                    {selectedGroup.notification}
                  </div>
                )}
              </div>
            </div>

            {/* Action buttons */}
            <div className="flex items-center gap-2 flex-wrap" style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid rgba(0,0,0,0.06)' }}>
              <button onClick={handleSendMessage} style={{ padding: '7px 16px', borderRadius: '8px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                <Send className="w-3.5 h-3.5" /> 发消息
              </button>
              {isAdmin && (
                <button onClick={openEditGroup} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Settings className="w-3.5 h-3.5" /> 编辑群信息
                </button>
              )}
              {isAdmin && (
                <button onClick={() => setShowInviteFriends(true)} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <UserPlus className="w-3.5 h-3.5" /> 邀请好友
                </button>
              )}
              <button onClick={openMemberSettings} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                <Settings className="w-3.5 h-3.5" /> 我的设置
              </button>
              {isOwner && (
                <button onClick={handleDisband} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid #E53935', background: '#FFF', color: '#E53935', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Trash2 className="w-3.5 h-3.5" /> 解散群组
                </button>
              )}
              {!isOwner && (
                <button onClick={handleQuitGroup} style={{ padding: '7px 16px', borderRadius: '8px', border: '1px solid #E53935', background: '#FFF', color: '#E53935', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <LogOut className="w-3.5 h-3.5" /> 退出群组
                </button>
              )}
            </div>
          </div>

          {/* ── Tab Bar ── */}
          <div className="flex items-center shrink-0" style={{ padding: '8px 8px 0' }}>
            {tabItems.filter(t => t.show).map(tab => (
              <button key={tab.key} onClick={() => setDetailTab(tab.key)} style={{ padding: '8px 16px', borderBottom: detailTab === tab.key ? '2px solid #3390EC' : '2px solid transparent', background: 'none', borderLeft: 'none', borderRight: 'none', borderTop: 'none', color: detailTab === tab.key ? '#3390EC' : '#646A73', fontSize: '13px', fontWeight: detailTab === tab.key ? 600 : 400, cursor: 'pointer', transition: 'all 0.2s' }}>
                {tab.label}
              </button>
            ))}
          </div>

          {/* ── Members Tab ── */}
          {detailTab === 'members' && (
            <div style={{ padding: '8px' }}>
              <div className="relative" style={{ marginBottom: '8px' }}>
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: '#A2ACB5' }} />
                <input value={memberSearch} onChange={(e) => setMemberSearch(e.target.value)} placeholder="搜索成员" style={{ ...inputStyle, paddingLeft: '36px', borderRadius: '22px', height: '36px', fontSize: '13px' }} onFocus={focusInput} onBlur={blurInput} />
              </div>
              <div style={{ background: '#FFF', borderRadius: '12px', overflow: 'hidden' }}>
                {filteredMembers.map(m => {
                  const name = getContactName(m.userId);
                  const isMe = m.userId === myUserId;
                  const displayName = m.nickname || name;
                  return (
                    <div key={m.id} className="flex items-center gap-3" style={{ padding: '10px 16px', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
                      {avatarCircle(name, 40,
                        m.online ? <span className="absolute" style={{ bottom: 0, right: 0, width: 10, height: 10, borderRadius: '50%', backgroundColor: '#4DCD5E', border: '2px solid #FFF' }} /> : null
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span style={{ fontSize: '14px', fontWeight: 500, color: '#1C2733' }}>{displayName}</span>
                          {m.roleLevel >= 2 && (
                            <span style={{ fontSize: '10px', fontWeight: 500, color: roleColor[m.roleLevel], backgroundColor: roleBg[m.roleLevel], borderRadius: '4px', padding: '1px 5px', display: 'flex', alignItems: 'center', gap: 2 }}>
                              {roleIcon[m.roleLevel]}{roleLabel[m.roleLevel]}
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
                      ) : isAdmin ? (
                        <button onClick={(e) => {
                          e.stopPropagation();
                          const items: ActionMenuItem[] = [];
                          if (isOwner) {
                            if (m.roleLevel === 2) items.push({ label: '取消管理员', onClick: () => handleSetAdmin(m) });
                            else if (m.roleLevel === 1) items.push({ label: '设为管理员', onClick: () => handleSetAdmin(m) });
                            items.push({ label: '转让群主', color: '#F5A623', onClick: () => handleTransferOwner(m) });
                          }
                          items.push({ label: '踢出群组', color: '#FF5252', onClick: () => handleKickMember(m) });
                          setActionMenu({ x: e.clientX, y: e.clientY, items });
                        }} style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                          <MoreHorizontal className="w-4 h-4" />
                        </button>
                      ) : null}
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
                <button onClick={() => setShowCreateLink(true)} style={{ padding: '7px 16px', borderRadius: '8px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Link className="w-3.5 h-3.5" /> 创建链接
                </button>
                <button onClick={() => setShowRevoked(v => !v)} style={{ fontSize: '12px', color: '#708499', background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3 }}>
                  {showRevoked ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                  {showRevoked ? '隐藏已撤销' : '显示已撤销'}
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
                          <span style={{ fontSize: '11px', color: '#A2ACB5' }}>{fmtTime(link.createdAt)}</span>
                        </div>
                        <div style={{ fontSize: '12px', color: '#708499', fontFamily: 'monospace', background: '#F5F7FA', borderRadius: '6px', padding: '6px 10px', marginBottom: 8, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{link.token}</div>
                        <div style={{ fontSize: '12px', color: '#A2ACB5', marginBottom: 6 }}>
                          创建者: {getContactName(link.createdBy)}
                          {link.expireAt && ` · 过期: ${new Date(link.expireAt).toLocaleDateString()}`}
                          {!link.expireAt && ' · 永久有效'}
                        </div>
                        {link.maxUses > 0 && (
                          <div style={{ marginBottom: 8 }}>
                            <div className="flex items-center justify-between" style={{ marginBottom: 4 }}>
                              <span style={{ fontSize: '11px', color: '#A2ACB5' }}>使用次数</span>
                              <span style={{ fontSize: '11px', color: '#646A73' }}>{link.usedCount}/{link.maxUses}</span>
                            </div>
                            <div style={{ height: 4, borderRadius: 2, background: 'rgba(0,0,0,0.06)' }}>
                              <div style={{ height: '100%', borderRadius: 2, background: progress >= 100 ? '#FF5252' : '#3390EC', width: `${progress}%`, transition: 'width 0.3s' }} />
                            </div>
                          </div>
                        )}
                        {!link.revoked && status.label === '有效' && (
                          <div className="flex items-center gap-2">
                            <button onClick={() => handleCopyLink(link.token)} style={{ padding: '4px 12px', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)', background: '#FFF', color: '#646A73', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3 }}>
                              <Copy className="w-3 h-3" /> 复制
                            </button>
                            <button onClick={() => handleRevokeLink(link)} style={{ padding: '4px 12px', borderRadius: '6px', border: '1px solid #FF5252', background: '#FFF', color: '#FF5252', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3 }}>
                              <RefreshCw className="w-3 h-3" /> 撤销
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
                  <div style={{ fontSize: '14px', color: '#A2ACB5' }}>{showRevoked ? '暂无链接' : '暂无有效链接'}</div>
                </div>
              )}
            </div>
          )}

          {/* ── Announcements Tab ── */}
          {detailTab === 'announcements' && (
            <div style={{ padding: '8px' }}>
              {isAdmin && (
                <div style={{ marginBottom: '8px' }}>
                  <button onClick={() => setShowCreateAnnouncement(true)} style={{ padding: '7px 16px', borderRadius: '8px', border: 'none', background: '#3390EC', color: '#FFF', fontSize: '13px', fontWeight: 500, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}>
                    <Megaphone className="w-3.5 h-3.5" /> 发布公告
                  </button>
                </div>
              )}
              {groupAnns.length > 0 ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {groupAnns.map(ann => (
                    <div key={ann.id} style={{ background: '#FFF', borderRadius: '12px', padding: '14px', border: ann.pinned ? '1px solid rgba(51,144,236,0.2)' : '1px solid rgba(0,0,0,0.06)' }}>
                      {ann.pinned && (
                        <div className="flex items-center gap-1" style={{ marginBottom: 8 }}>
                          <Pin className="w-3 h-3" style={{ color: '#3390EC' }} />
                          <span style={{ fontSize: '11px', fontWeight: 500, color: '#3390EC' }}>置顶</span>
                        </div>
                      )}
                      <div style={{ fontSize: '14px', color: '#1C2733', lineHeight: '1.6', marginBottom: 10 }}>{ann.content}</div>
                      <div className="flex items-center justify-between">
                        <div style={{ fontSize: '12px', color: '#A2ACB5' }}>{getContactName(ann.createdBy)} · {fmtTime(ann.createdAt)}</div>
                        {isAdmin && (
                          <button onClick={() => handleTogglePin(ann)} style={{ fontSize: '12px', color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 2 }}>
                            {ann.pinned ? <><XCircle className="w-3 h-3" /> 取消置顶</> : <><Pin className="w-3 h-3" /> 置顶</>}
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center" style={{ padding: '40px 24px' }}>
                  <Megaphone className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: '12px' }} />
                  <div style={{ fontSize: '14px', color: '#A2ACB5' }}>暂无公告</div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* ── Action Menu ── */}
        {actionMenu && <ActionMenu x={actionMenu.x} y={actionMenu.y} items={actionMenu.items} onClose={() => setActionMenu(null)} />}

        {/* ── Edit Group Modal ── */}
        <InputModal open={showEditGroup} title="编辑群信息" onClose={() => setShowEditGroup(false)} onSubmit={handleEditGroup}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群名称</label>
            <input value={editName} onChange={(e) => setEditName(e.target.value)} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群图标 (URL)</label>
            <input value={editIcon} onChange={(e) => setEditIcon(e.target.value)} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群公告</label>
            <textarea value={editNotification} onChange={(e) => setEditNotification(e.target.value)} rows={3} style={{ ...inputStyle, resize: 'none' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div className="flex items-center justify-between" style={{ padding: '8px 0' }}>
            <span style={{ fontSize: '13px', color: '#1C2733' }}>入群验证</span>
            <button onClick={() => setEditVerify(v => !v)} style={{ width: 44, height: 24, borderRadius: 12, border: 'none', background: editVerify ? '#3390EC' : '#D1D5DB', cursor: 'pointer', position: 'relative', transition: 'background 0.2s' }}>
              <div style={{ width: 20, height: 20, borderRadius: '50%', background: '#FFF', position: 'absolute', top: 2, left: editVerify ? 22 : 2, transition: 'left 0.2s', boxShadow: '0 1px 3px rgba(0,0,0,0.2)' }} />
            </button>
          </div>
        </InputModal>

        {/* ── Invite Friends Modal ── */}
        <InputModal open={showInviteFriends} title="邀请好友" onClose={() => { setShowInviteFriends(false); setInviteSelected(new Set()); }} onSubmit={handleInviteFriends}>
          <div style={{ marginBottom: '12px' }}>
            <span style={{ fontSize: '13px', color: '#646A73' }}>已选择 {inviteSelected.size} 人</span>
          </div>
          <div style={{ maxHeight: '300px', overflow: 'auto', borderRadius: '10px', border: '1px solid rgba(0,0,0,0.06)' }}>
            {friendsNotInGroup.map(c => (
              <label key={c.id} className="flex items-center gap-3" style={{ padding: '8px 12px', borderBottom: '1px solid rgba(0,0,0,0.04)', cursor: 'pointer' }}>
                <input type="checkbox" checked={inviteSelected.has(c.id)} onChange={() => {
                  setInviteSelected(prev => { const n = new Set(prev); if (n.has(c.id)) n.delete(c.id); else n.add(c.id); return n; });
                }} style={{ width: 16, height: 16, accentColor: '#3390EC' }} />
                {avatarCircle(c.name, 32)}
                <span style={{ fontSize: '14px', color: '#1C2733' }}>{c.name}</span>
              </label>
            ))}
            {friendsNotInGroup.length === 0 && (
              <div style={{ padding: '24px', textAlign: 'center', fontSize: '13px', color: '#A2ACB5' }}>所有好友都已在群中</div>
            )}
          </div>
        </InputModal>

        {/* ── Create Link Modal ── */}
        <InputModal open={showCreateLink} title="创建邀请链接" onClose={() => setShowCreateLink(false)} onSubmit={handleCreateLink}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>有效期</label>
            <select value={newLinkExpiry} onChange={(e) => setNewLinkExpiry(e.target.value)} style={{ ...inputStyle, cursor: 'pointer' }}>
              <option value="1d">1天</option>
              <option value="7d">7天</option>
              <option value="30d">30天</option>
              <option value="never">永久</option>
            </select>
          </div>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>最大使用次数 (0 = 无限制)</label>
            <input type="number" min="0" value={newLinkMaxUses} onChange={(e) => setNewLinkMaxUses(e.target.value)} style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        {/* ── Create Announcement Modal ── */}
        <InputModal open={showCreateAnnouncement} title="发布公告" onClose={() => { setShowCreateAnnouncement(false); setNewAnnContent(''); }} onSubmit={handleCreateAnnouncement}>
          <div>
            <div className="flex items-center justify-between" style={{ marginBottom: '6px' }}>
              <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733' }}>公告内容</label>
              <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{newAnnContent.length}/500</span>
            </div>
            <textarea value={newAnnContent} onChange={(e) => setNewAnnContent(e.target.value.slice(0, 500))} rows={5} placeholder="输入公告内容..." style={{ ...inputStyle, resize: 'none' }} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        {/* ── Member Settings Modal ── */}
        <InputModal open={showMemberSettings} title="我的群设置" onClose={() => setShowMemberSettings(false)} onSubmit={handleSaveMemberSettings}>
          <div style={{ marginBottom: '12px' }}>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群昵称</label>
            <input value={settingNickname} onChange={(e) => setSettingNickname(e.target.value)} placeholder="在群中的显示名称" style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
          <div>
            <label style={{ fontSize: '13px', fontWeight: 500, color: '#1C2733', marginBottom: '6px', display: 'block' }}>群备注</label>
            <input value={settingRemark} onChange={(e) => setSettingRemark(e.target.value)} placeholder="仅自己可见" style={inputStyle} onFocus={focusInput} onBlur={blurInput} />
          </div>
        </InputModal>

        <ConfirmModal open={!!confirmOpts} opts={confirmOpts} />
      </div>
    );
  }

  return null;
}
