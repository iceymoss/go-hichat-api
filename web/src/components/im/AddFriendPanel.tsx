'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Search, X, UserPlus, Send, Loader2, UserCircle, Users, Clock } from 'lucide-react';
import { toast } from 'sonner';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { getAvatarColor } from '@/lib/utils';
import type { Contact } from '@/lib/mock-data';
import {
  searchUsers,
  searchGroups,
  sendFriendRequest,
  type GroupSearchItem,
} from '@/lib/friend-group-api';
import type { BackendUser } from '@/lib/api-client';
import UserProfileCard from './UserProfileCard';
import GroupProfileCard from './GroupProfileCard';

type Tab = 'user' | 'group';
const PAGE_SIZE = 20;

interface AddFriendPanelProps {
  open: boolean;
  onClose: () => void;
  /** 发送好友/入群申请成功后回调（如刷新请求列表） */
  onSent?: () => void;
}

/** BackendUser -> Contact（用于复用 UserProfileCard） */
function userToContact(u: BackendUser): Contact {
  return {
    id: u.id,
    name: u.nickname || u.id,
    avatar: u.avatar || '',
    pinyin: '',
    letter: '#',
    online: false,
    gender: u.sex === 1 ? 'male' : u.sex === 2 ? 'female' : undefined,
    phone: u.mobile || undefined,
    region: u.region || undefined,
    signature: u.introduction || undefined,
    introduction: u.introduction || undefined,
    occupation: u.occupation || undefined,
    tags: u.tags || undefined,
    account: u.id,
  };
}

export default function AddFriendPanel({ open, onClose, onSent }: AddFriendPanelProps) {
  const { currentUser, friends, setActiveTab, setSelectedConversationId } = useIMStore();
  const token = currentUser?.token || '';
  const myId = currentUser?.id || '';

  const [tab, setTab] = useState<Tab>('user');
  const [keyword, setKeyword] = useState('');
  const [users, setUsers] = useState<BackendUser[]>([]);
  const [groups, setGroups] = useState<GroupSearchItem[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  // 选中的资料卡片
  const [selectedUser, setSelectedUser] = useState<BackendUser | null>(null);
  const [selectedGroup, setSelectedGroup] = useState<GroupSearchItem | null>(null);

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reqIdRef = useRef(0); // 在途请求作废
  const scrollRef = useRef<HTMLDivElement>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);

  const loaded = tab === 'user' ? users.length : groups.length;
  const hasMore = loaded < total;

  const reset = useCallback(() => {
    setKeyword('');
    setUsers([]); setGroups([]);
    setPage(1); setTotal(0); setError(false);
    setSelectedUser(null); setSelectedGroup(null);
  }, []);

  const handleClose = useCallback(() => { reset(); onClose(); }, [reset, onClose]);

  // 执行搜索（pageToLoad>1 表示加载更多）
  const runSearch = useCallback(async (kw: string, pageToLoad: number, curTab: Tab) => {
    const q = kw.trim();
    if (!q) { setUsers([]); setGroups([]); setTotal(0); setLoading(false); return; }
    const reqId = ++reqIdRef.current;
    setLoading(true); setError(false);
    try {
      if (curTab === 'user') {
        const { users: list, total: t } = await searchUsers(token, q, pageToLoad, PAGE_SIZE);
        if (reqId !== reqIdRef.current) return; // 已过期
        setUsers((prev) => (pageToLoad === 1 ? list : [...prev, ...list]));
        setTotal(t);
      } else {
        const { list, total: t } = await searchGroups(token, q, pageToLoad, PAGE_SIZE);
        if (reqId !== reqIdRef.current) return;
        setGroups((prev) => (pageToLoad === 1 ? list : [...prev, ...list]));
        setTotal(t);
      }
      setPage(pageToLoad);
    } catch {
      if (reqId === reqIdRef.current) setError(true);
    } finally {
      if (reqId === reqIdRef.current) setLoading(false);
    }
  }, [token]);

  // 关键词/标签变化 → 防抖重置搜索
  const onKeywordChange = useCallback((v: string) => {
    setKeyword(v);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    reqIdRef.current++; // 作废在途
    setPage(1); setTotal(0);
    if (!v.trim()) { setUsers([]); setGroups([]); setLoading(false); return; }
    setLoading(true);
    debounceRef.current = setTimeout(() => runSearch(v, 1, tab), 400);
  }, [runSearch, tab]);

  // 切 tab：重置并按当前关键词搜
  const switchTab = useCallback((t: Tab) => {
    if (t === tab) return;
    reqIdRef.current++;
    setTab(t); setUsers([]); setGroups([]); setPage(1); setTotal(0); setError(false);
    if (keyword.trim()) { setLoading(true); runSearch(keyword, 1, t); }
  }, [tab, keyword, runSearch]);

  // 无限滚动：触底加载下一页
  useEffect(() => {
    if (!open) return;
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const io = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && hasMore && !loading && keyword.trim()) {
        runSearch(keyword, page + 1, tab);
      }
    }, { root: scrollRef.current, rootMargin: '120px' });
    io.observe(sentinel);
    return () => io.disconnect();
  }, [open, hasMore, loading, keyword, page, tab, runSearch]);

  if (!open) return null;

  const isFriend = (uid: string) => friends.some((f) => f.id === uid || f.friend_uid === uid);

  const openUserChat = async (uid: string) => {
    if (!token || !myId) return;
    try { await useChatStore.getState().getOrCreateConversation(token, myId, uid); } catch { /* ignore */ }
    const parts = [myId, uid].sort();
    setActiveTab('chats');
    setSelectedConversationId(`${parts[0]}_${parts[1]}`);
    handleClose();
  };

  const enterGroupChat = (gid: string) => {
    setActiveTab('chats');
    setSelectedConversationId(gid);
    handleClose();
  };

  const handleSendFriend = async (uid: string, msg?: string) => {
    if (!token) return false;
    const ok = await sendFriendRequest(token, uid, msg);
    if (ok) { toast.success('好友请求已发送'); onSent?.(); }
    else toast.error('发送失败，请重试');
    return ok;
  };

  const tabBtn = (t: Tab, label: string) => (
    <button
      onClick={() => switchTab(t)}
      style={{
        flex: 1, height: 36, border: 'none', background: 'transparent', cursor: 'pointer',
        fontSize: 14, fontWeight: tab === t ? 600 : 400,
        color: tab === t ? '#3390EC' : '#646A73',
        borderBottom: tab === t ? '2px solid #3390EC' : '2px solid transparent',
      }}
    >
      {label}
    </button>
  );

  const placeholder = tab === 'user' ? '搜索手机号/邮箱/昵称' : '搜索群号/群名';

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) handleClose(); }}
    >
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div
        className="relative"
        style={{ background: '#FFFFFF', borderRadius: 16, width: '90%', maxWidth: 440, height: '75vh', maxHeight: 640, overflow: 'hidden', boxShadow: '0 8px 40px rgba(0,0,0,0.15)', display: 'flex', flexDirection: 'column' }}
      >
        {/* Header */}
        <div className="flex items-center justify-between" style={{ padding: '16px 20px 0' }}>
          <span style={{ fontSize: 16, fontWeight: 600, color: '#1C2733' }}>添加好友 / 群</span>
          <button onClick={handleClose} style={{ width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'rgba(0,0,0,0.04)', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex" style={{ padding: '8px 20px 0', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          {tabBtn('user', '用户')}
          {tabBtn('group', '群聊')}
        </div>

        {/* Search input */}
        <div style={{ padding: '12px 20px' }}>
          <div className="flex items-center gap-2" style={{ background: '#F5F7FA', borderRadius: 10, padding: '8px 12px' }}>
            <Search className="w-4 h-4" style={{ color: '#A2ACB5' }} />
            <input
              type="text"
              value={keyword}
              onChange={(e) => onKeywordChange(e.target.value)}
              placeholder={placeholder}
              autoFocus
              style={{ flex: 1, border: 'none', background: 'transparent', outline: 'none', fontSize: 14, color: '#1C2733' }}
            />
            {loading && <Loader2 className="w-4 h-4" style={{ color: '#3390EC', animation: 'spin 1s linear infinite' }} />}
          </div>
        </div>

        {/* Results */}
        <div ref={scrollRef} className="flex-1 overflow-y-auto im-scroll" style={{ padding: '0 20px 16px' }}>
          {!keyword.trim() ? (
            <Empty icon={<Search className="w-12 h-12" />} text={tab === 'user' ? '输入手机号、邮箱或昵称搜索用户' : '输入群号或群名搜索群聊'} />
          ) : tab === 'user' ? (
            users.length > 0 ? (
              <>
                {users.map((u) => (
                  <UserRow key={u.id} user={u} self={u.id === myId} onClick={() => setSelectedUser(u)} />
                ))}
                <ListFooter loading={loading} hasMore={hasMore} error={error} onRetry={() => runSearch(keyword, page, tab)} sentinelRef={sentinelRef} count={users.length} total={total} />
              </>
            ) : !loading ? (
              error ? <ErrorState onRetry={() => runSearch(keyword, 1, tab)} /> : <Empty icon={<UserCircle className="w-12 h-12" />} text="未找到用户" />
            ) : null
          ) : (
            groups.length > 0 ? (
              <>
                {groups.map((g) => (
                  <GroupRow key={g.id} group={g} onClick={() => setSelectedGroup(g)} />
                ))}
                <ListFooter loading={loading} hasMore={hasMore} error={error} onRetry={() => runSearch(keyword, page, tab)} sentinelRef={sentinelRef} count={groups.length} total={total} />
              </>
            ) : !loading ? (
              error ? <ErrorState onRetry={() => runSearch(keyword, 1, tab)} /> : <Empty icon={<Users className="w-12 h-12" />} text="未找到群聊" />
            ) : null
          )}
        </div>
      </div>

      {/* 用户资料卡片 */}
      {selectedUser && (
        <div className="fixed inset-0" style={{ zIndex: 10002, display: 'flex', alignItems: 'center', justifyContent: 'center' }} onClick={(e) => { if (e.target === e.currentTarget) setSelectedUser(null); }}>
          <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
          <div
            className="relative im-scroll"
            style={{ width: '90%', maxWidth: 380, maxHeight: '82vh', overflowY: 'auto', background: '#FFFFFF', borderRadius: 16, padding: '24px 20px 20px', boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}
          >
            <button
              onClick={() => setSelectedUser(null)}
              style={{ position: 'absolute', top: 12, right: 12, zIndex: 2, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'rgba(0,0,0,0.04)', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            >
              <X className="w-4 h-4" />
            </button>
            <UserProfileCard
              contact={userToContact(selectedUser)}
              isStranger={selectedUser.id !== myId && !isFriend(selectedUser.id)}
              onSendMessage={() => openUserChat(selectedUser.id)}
              onAddFriend={async () => { const ok = await handleSendFriend(selectedUser.id); if (ok) setSelectedUser(null); }}
            />
          </div>
        </div>
      )}

      {/* 群资料卡片 */}
      {selectedGroup && (
        <GroupProfileCard group={selectedGroup} onClose={() => setSelectedGroup(null)} onEnterGroup={enterGroupChat} />
      )}
    </div>
  );
}

/* ── 列表行 ── */

function UserRow({ user, self, onClick }: { user: BackendUser; self: boolean; onClick: () => void }) {
  const sub = user.introduction || user.region || (user.id ? `ID: ${user.id}` : '');
  return (
    <div onClick={onClick} className="flex items-center" style={{ gap: 12, padding: '10px 0', borderBottom: '1px solid rgba(0,0,0,0.05)', cursor: 'pointer' }}>
      <Avatar src={user.avatar} name={user.nickname} />
      <div className="flex-1 min-w-0">
        <div style={{ fontSize: 14, fontWeight: 500, color: '#1C2733' }}>
          {user.nickname || '未知'}{self && <span style={{ fontSize: 12, color: '#A2ACB5', marginLeft: 6 }}>（我）</span>}
        </div>
        {sub && <div className="truncate" style={{ fontSize: 12, color: '#A2ACB5' }}>{sub}</div>}
      </div>
    </div>
  );
}

function GroupRow({ group, onClick }: { group: GroupSearchItem; onClick: () => void }) {
  return (
    <div onClick={onClick} className="flex items-center" style={{ gap: 12, padding: '10px 0', borderBottom: '1px solid rgba(0,0,0,0.05)', cursor: 'pointer' }}>
      <Avatar src={group.icon} name={group.name} square />
      <div className="flex-1 min-w-0">
        <div style={{ fontSize: 14, fontWeight: 500, color: '#1C2733' }}>{group.name || '未命名群聊'}</div>
        <div className="flex items-center gap-1" style={{ fontSize: 12, color: '#A2ACB5' }}>
          <Users className="w-3 h-3" />{group.member_count} 人 · 群号 {group.id}
        </div>
      </div>
    </div>
  );
}

function Avatar({ src, name, square }: { src?: string; name?: string; square?: boolean }) {
  const radius = square ? 10 : '50%';
  return (
    <div style={{ width: 40, height: 40, flexShrink: 0 }}>
      {src ? (
        <img src={src} alt={name || '?'} style={{ width: 40, height: 40, borderRadius: radius, objectFit: 'cover' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const n = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (n) n.style.display = 'flex'; }} />
      ) : null}
      <div className="items-center justify-center" style={{ width: 40, height: 40, borderRadius: radius, background: getAvatarColor(name || '?'), color: '#FFF', fontSize: 16, fontWeight: 600, display: src ? 'none' : 'flex' }}>
        {(name || '?')[0]}
      </div>
    </div>
  );
}

function ListFooter({ loading, hasMore, error, onRetry, sentinelRef, count, total }: { loading: boolean; hasMore: boolean; error: boolean; onRetry: () => void; sentinelRef: React.RefObject<HTMLDivElement | null>; count: number; total: number }) {
  return (
    <div ref={sentinelRef} className="flex items-center justify-center" style={{ padding: '12px 0', fontSize: 12, color: '#A2ACB5' }}>
      {error ? (
        <button onClick={onRetry} style={{ color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer' }}>加载失败，点击重试</button>
      ) : loading ? (
        <Loader2 className="w-4 h-4" style={{ animation: 'spin 1s linear infinite' }} />
      ) : hasMore ? (
        '上滑加载更多'
      ) : (
        `共 ${total} 条`
      )}
    </div>
  );
}

function Empty({ icon, text }: { icon: React.ReactNode; text: string }) {
  return (
    <div className="flex flex-col items-center" style={{ padding: '40px 0', color: '#D1D5DB' }}>
      {icon}
      <div style={{ fontSize: 13, color: '#A2ACB5', marginTop: 8 }}>{text}</div>
    </div>
  );
}

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center" style={{ padding: '40px 0' }}>
      <Clock className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: 8 }} />
      <div style={{ fontSize: 13, color: '#A2ACB5', marginBottom: 8 }}>加载失败</div>
      <button onClick={onRetry} style={{ color: '#3390EC', background: 'none', border: 'none', cursor: 'pointer', fontSize: 13 }}>重试</button>
    </div>
  );
}
