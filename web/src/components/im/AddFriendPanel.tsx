'use client';

import React, { useState, useCallback, useRef } from 'react';
import { Search, X, Send, Loader2, UserCircle, UserPlus } from 'lucide-react';
import { toast } from 'sonner';
import { useIMStore } from '@/lib/im-store';
import { getAvatarColor } from '@/lib/utils';

/* ═══════════════════════════════════════
   API helpers（搜索用户 / 发送好友申请）
   ═══════════════════════════════════════ */

async function apiSendFriendRequest(token: string, userUid: string, reqMsg?: string): Promise<boolean> {
  const resp = await fetch('/api/social/friend/putIn', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ user_uid: userUid, ...(reqMsg ? { req_msg: reqMsg } : {}) }),
  });
  const json = await resp.json();
  return json.success === true;
}

async function apiSearchUsers(token: string, query: string): Promise<any[]> {
  const isPhone = /^1[3-9]\d{2,10}$/.test(query);
  const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(query);
  const params = isPhone ? `phone=${query}` : isEmail ? `email=${query}` : `name=${query}`;
  try {
    const resp = await fetch(`/api/user/search?${params}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (json.success && json.data?.users) return json.data.users;
    return [];
  } catch {
    return [];
  }
}

/* ═══════════════════════════════════════
   AddFriendPanel — 搜索并添加好友的浮层卡片
   会话列表 / 新的朋友页 共用
   ═══════════════════════════════════════ */

interface AddFriendPanelProps {
  open: boolean;
  onClose: () => void;
  /** 发送成功后回调（如刷新好友请求列表/未读数） */
  onSent?: () => void;
}

export default function AddFriendPanel({ open, onClose, onSent }: AddFriendPanelProps) {
  const { currentUser } = useIMStore();
  const token = currentUser?.token || '';

  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [searching, setSearching] = useState(false);
  const [sendMsg, setSendMsg] = useState('');
  const [sendingTo, setSendingTo] = useState<string | null>(null);
  const [sendLoading, setSendLoading] = useState(false);
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const reset = useCallback(() => {
    setSearchQuery('');
    setSearchResults([]);
    setSendingTo(null);
    setSendMsg('');
  }, []);

  const handleClose = useCallback(() => {
    reset();
    onClose();
  }, [reset, onClose]);

  const handleSearchUsers = useCallback((q: string) => {
    setSearchQuery(q);
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    if (!q.trim()) { setSearchResults([]); setSearching(false); return; }
    setSearching(true);
    searchTimerRef.current = setTimeout(() => {
      if (!token) return;
      apiSearchUsers(token, q.trim()).then((users) => {
        setSearchResults(users);
        setSearching(false);
      });
    }, 400);
  }, [token]);

  const handleSendRequest = useCallback(async (userId: string) => {
    if (!token) return;
    setSendLoading(true);
    try {
      const ok = await apiSendFriendRequest(token, userId, sendMsg || undefined);
      if (ok) {
        toast.success('好友请求已发送');
        setSendingTo(null);
        setSendMsg('');
        onSent?.();
      } else {
        toast.error('发送失败，请重试');
      }
    } catch {
      toast.error('发送失败，请稍后重试');
    } finally {
      setSendLoading(false);
    }
  }, [token, sendMsg, onSent]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) handleClose(); }}
    >
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div
        className="relative"
        style={{
          background: '#FFFFFF',
          borderRadius: '16px',
          width: '90%',
          maxWidth: '440px',
          maxHeight: '75vh',
          overflow: 'hidden',
          boxShadow: '0 8px 40px rgba(0,0,0,0.15)',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {/* Panel header */}
        <div className="flex items-center justify-between" style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <span style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733' }}>添加好友</span>
          <button
            onClick={handleClose}
            style={{
              width: 28, height: 28,
              borderRadius: '50%',
              border: 'none',
              background: 'rgba(0,0,0,0.04)',
              color: '#A2ACB5',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Search input */}
        <div style={{ padding: '12px 20px' }}>
          <div className="flex items-center gap-2" style={{ background: '#F5F7FA', borderRadius: '10px', padding: '8px 12px' }}>
            <Search className="w-4 h-4" style={{ color: '#A2ACB5' }} />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => handleSearchUsers(e.target.value)}
              placeholder="搜索手机号/邮箱/昵称"
              style={{
                flex: 1,
                border: 'none',
                background: 'transparent',
                outline: 'none',
                fontSize: '14px',
                color: '#1C2733',
              }}
              autoFocus
            />
            {searching && <Loader2 className="w-4 h-4" style={{ color: '#3390EC', animation: 'spin 1s linear infinite' }} />}
          </div>
        </div>

        {/* Search results */}
        <div className="flex-1 overflow-y-auto im-scroll" style={{ padding: '0 20px 16px' }}>
          {searchResults.length > 0 ? (
            searchResults.map((user: any) => (
              <div
                key={user.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '10px 0',
                  borderBottom: '1px solid rgba(0,0,0,0.05)',
                }}
              >
                <div style={{ width: 40, height: 40, flexShrink: 0 }}>
                  {user.avatar ? (
                    <img
                      src={user.avatar}
                      alt={user.nickname || '?'}
                      style={{ width: 40, height: 40, borderRadius: '50%', objectFit: 'cover' }}
                      onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }}
                    />
                  ) : null}
                  <div
                    className="items-center justify-center"
                    style={{
                      width: 40, height: 40,
                      borderRadius: '50%',
                      backgroundColor: getAvatarColor(user.nickname || '?'),
                      fontSize: '16px',
                      fontWeight: 600,
                      color: '#FFFFFF',
                      display: user.avatar ? 'none' : 'flex',
                    }}
                  >
                    {(user.nickname || '?')[0]}
                  </div>
                </div>
                <div className="flex-1 min-w-0">
                  <div style={{ fontSize: '14px', fontWeight: 500, color: '#1C2733' }}>{user.nickname || '未知'}</div>
                  {user.region && <div style={{ fontSize: '12px', color: '#A2ACB5' }}>{user.region}</div>}
                </div>
                {sendingTo === user.id ? (
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      value={sendMsg}
                      onChange={(e) => setSendMsg(e.target.value)}
                      placeholder="附言"
                      style={{
                        width: 100,
                        border: '1px solid rgba(0,0,0,0.1)',
                        borderRadius: '6px',
                        padding: '4px 8px',
                        fontSize: '12px',
                        outline: 'none',
                      }}
                      onFocus={(e) => { e.target.style.borderColor = '#3390EC'; }}
                      onBlur={(e) => { e.target.style.borderColor = 'rgba(0,0,0,0.1)'; }}
                    />
                    <button
                      onClick={() => handleSendRequest(user.id)}
                      disabled={sendLoading}
                      style={{
                        padding: '4px 12px',
                        borderRadius: '6px',
                        border: 'none',
                        background: '#3390EC',
                        color: '#FFFFFF',
                        fontSize: '12px',
                        fontWeight: 500,
                        cursor: sendLoading ? 'not-allowed' : 'pointer',
                        opacity: sendLoading ? 0.7 : 1,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                      }}
                    >
                      {sendLoading ? <Loader2 className="w-3 h-3" style={{ animation: 'spin 1s linear infinite' }} /> : <Send className="w-3 h-3" />}
                      发送
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => { setSendingTo(user.id); setSendMsg(''); }}
                    style={{
                      padding: '5px 14px',
                      borderRadius: '6px',
                      border: 'none',
                      background: '#3390EC',
                      color: '#FFFFFF',
                      fontSize: '12px',
                      fontWeight: 500,
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 4,
                    }}
                  >
                    <UserPlus className="w-3.5 h-3.5" />
                    添加
                  </button>
                )}
              </div>
            ))
          ) : searchQuery.trim() && !searching ? (
            <div className="flex flex-col items-center" style={{ padding: '30px 0' }}>
              <UserCircle className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: '8px' }} />
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>未找到用户</div>
            </div>
          ) : !searchQuery.trim() ? (
            <div className="flex flex-col items-center" style={{ padding: '30px 0' }}>
              <Search className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: '8px' }} />
              <div style={{ fontSize: '13px', color: '#A2ACB5' }}>输入手机号、邮箱或昵称搜索用户</div>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
