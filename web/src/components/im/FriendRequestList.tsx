'use client';

import React, { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import {
  ArrowLeft,
  Bell,
  UserPlus,
  Trash2,
  X,
  Loader2,
  UserCircle,
  Mail,
  Phone,
  MapPin,
  MessageSquare,
  Clock,
  CheckCircle,
  XCircle,
  MinusCircle,
  AlertCircle,
  Search,
  Send,
} from 'lucide-react';
import { toast } from 'sonner';
import { useIMStore } from '@/lib/im-store';
import { getAvatarColor } from '@/lib/utils';

/* ═══════════════════════════════════════
   Types (previously from mock-data)
   ═══════════════════════════════════════ */

export type FriendRequestStatus = 'pending' | 'accepted' | 'rejected' | 'ignored';
export type FriendRequestClass = 'received' | 'sent';

export interface FriendRequest {
  id: string;
  class: FriendRequestClass;
  nickname: string;
  avatar: string;
  sex: 'male' | 'female' | 'unknown';
  region: string;
  occupation: string;
  introduction: string;
  tags: string[];
  reqMsg: string;
  handleMsg: string;
  status: FriendRequestStatus;
  readState: boolean;
  reqTime: Date;
  hiChatId?: string;
  email?: string;
  phone?: string;
}

/* ═══════════════════════════════════════
   API helpers
   ═══════════════════════════════════════ */

/** Map backend sex int to string */
function mapSex(sex?: number): 'male' | 'female' | 'unknown' {
  if (sex === 1) return 'male';
  if (sex === 2) return 'female';
  return 'unknown';
}

/** Map backend handle_result int to status string */
function mapHandleResult(hr: number): FriendRequestStatus {
  switch (hr) {
    case 1: return 'accepted';
    case 2: return 'rejected';
    case 3: return 'ignored';
    default: return 'pending';
  }
}

/** Map a single API record to our FriendRequest shape */
function mapApiRequest(item: any, reqClass: FriendRequestClass): FriendRequest {
  const tagsRaw = item.tags;
  let tags: string[] = [];
  if (Array.isArray(tagsRaw)) {
    tags = tagsRaw;
  } else if (typeof tagsRaw === 'string' && tagsRaw) {
    try { tags = JSON.parse(tagsRaw); } catch { tags = tagsRaw.split(',').filter(Boolean); }
  }

  return {
    id: String(item.id),
    class: reqClass,
    nickname: item.nickname || '未知用户',
    avatar: item.avatar || '',
    sex: mapSex(item.sex),
    region: item.region || '',
    occupation: item.occupation || '',
    introduction: item.introduction || '',
    tags,
    reqMsg: item.req_msg || '',
    handleMsg: item.handle_msg || '',
    status: (item.status as FriendRequestStatus) || mapHandleResult(item.handle_result ?? 0),
    readState: item.read_state === 1,
    reqTime: new Date((item.req_time ?? 0) * 1000),
    hiChatId: item.user_id ? String(item.user_id) : undefined,
    email: item.email || undefined,
    phone: item.phone || undefined,
  };
}

async function fetchRequests(token: string, cls: '0' | '1', type: number): Promise<FriendRequest[]> {
  const reqClass: FriendRequestClass = cls === '1' ? 'received' : 'sent';
  const resp = await fetch(`/api/social/friend/putIns?class=${cls}&type=${type}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const json = await resp.json();
  if (!json.success || !json.data?.list) return [];
  return (json.data.list as any[]).map((item) => mapApiRequest(item, reqClass));
}

async function fetchAllRequests(token: string): Promise<FriendRequest[]> {
  // Fetch both received (class=1) and sent (class=0) for all status types (type=0,1,2,3)
  const promises: Promise<FriendRequest[]>[] = [];
  for (const cls of ['0', '1'] as const) {
    for (const type of [0, 1, 2, 3]) {
      promises.push(fetchRequests(token, cls, type));
    }
  }
  const results = await Promise.all(promises);
  const all = results.flat();
  // Deduplicate by id
  const seen = new Set<string>();
  return all.filter((r) => {
    if (seen.has(r.id)) return false;
    seen.add(r.id);
    return true;
  });
}

async function fetchUnreadCount(token: string): Promise<number> {
  try {
    const resp = await fetch('/api/social/friend/putIn/messageCount', {
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = await resp.json();
    if (json.success && json.data?.count !== undefined) return json.data.count;
    return 0;
  } catch {
    return 0;
  }
}

async function apiHandleRequest(token: string, friendReqId: number, handleResult: number, handleMsg?: string): Promise<boolean> {
  const resp = await fetch('/api/social/friend/putIn', {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      friend_req_id: friendReqId,
      handle_result: handleResult,
      ...(handleMsg ? { handle_msg: handleMsg } : {}),
    }),
  });
  const json = await resp.json();
  return json.success === true;
}

async function apiDeleteRequest(token: string, friendReqId: number): Promise<boolean> {
  const resp = await fetch('/api/social/friend/putIn/delete', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ friend_req_id: friendReqId }),
  });
  const json = await resp.json();
  return json.success === true;
}

async function apiMarkAsRead(token: string, friendReqId: number): Promise<boolean> {
  try {
    const resp = await fetch('/api/social/friend/putIn/read', {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ friend_req_id: friendReqId }),
    });
    const json = await resp.json();
    return json.success === true;
  } catch {
    return false;
  }
}

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
   Helpers
   ═══════════════════════════════════════ */

function formatRelativeTime(date: Date): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 30) {
    const d = new Date(date);
    return `${d.getMonth() + 1}月${d.getDate()}日`;
  }
  const d = new Date(date);
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`;
}

const statusConfig: Record<FriendRequestStatus, { label: string; color: string; bg: string; icon: React.ReactNode }> = {
  pending: {
    label: '待处理',
    color: '#F5A623',
    bg: 'rgba(245,166,35,0.1)',
    icon: <AlertCircle className="w-3.5 h-3.5" />,
  },
  accepted: {
    label: '已同意',
    color: '#4DCD5E',
    bg: 'rgba(77,205,94,0.1)',
    icon: <CheckCircle className="w-3.5 h-3.5" />,
  },
  rejected: {
    label: '已拒绝',
    color: '#FF5252',
    bg: 'rgba(255,82,82,0.1)',
    icon: <XCircle className="w-3.5 h-3.5" />,
  },
  ignored: {
    label: '已忽略',
    color: '#A2ACB5',
    bg: 'rgba(162,172,181,0.1)',
    icon: <MinusCircle className="w-3.5 h-3.5" />,
  },
};

const statusStripeColors: Record<FriendRequestStatus, string> = {
  pending: '#F5A623',
  accepted: '#4DCD5E',
  rejected: '#FF5252',
  ignored: '#A2ACB5',
};

type StatusFilter = 'all' | FriendRequestStatus;

/* ═══════════════════════════════════════
   Confirm Dialog (Accept / Reject / Delete)
   ═══════════════════════════════════════ */

interface ConfirmDialogProps {
  open: boolean;
  type: 'accept' | 'reject' | 'delete';
  nickname: string;
  loading: boolean;
  onClose: () => void;
  onConfirm: (msg: string) => void;
}

function ConfirmDialog({ open, type, nickname, loading, onClose, onConfirm }: ConfirmDialogProps) {
  const [message, setMessage] = useState('');

  if (!open) return null;

  const isDelete = type === 'delete';
  const title = isDelete
    ? '删除记录'
    : type === 'accept'
      ? '同意好友请求'
      : '拒绝好友请求';
  const description = isDelete
    ? `确定要删除与 ${nickname} 的好友请求记录吗？`
    : type === 'accept'
      ? `确定同意 ${nickname} 的好友请求吗？`
      : `确定拒绝 ${nickname} 的好友请求吗？`;
  const confirmLabel = isDelete ? '删除' : type === 'accept' ? '同意' : '拒绝';
  const confirmColor = isDelete ? '#E53935' : type === 'accept' ? '#3390EC' : '#E53935';

  const handleConfirm = () => {
    onConfirm(message);
    setMessage('');
  };

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0"
        style={{ background: 'rgba(0,0,0,0.5)' }}
      />

      {/* Card */}
      <div
        className="relative"
        style={{
          background: '#FFFFFF',
          borderRadius: '16px',
          padding: '24px',
          width: '90%',
          maxWidth: '400px',
          boxShadow: '0 8px 40px rgba(0,0,0,0.15)',
        }}
      >
        {/* Close */}
        <button
          onClick={onClose}
          className="absolute"
          style={{
            top: 16, right: 16,
            width: 28, height: 28,
            borderRadius: '50%',
            border: 'none',
            background: 'transparent',
            color: '#A2ACB5',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <X className="w-4 h-4" />
        </button>

        {/* Title */}
        <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#1C2733', marginBottom: '8px', paddingRight: 32 }}>
          {title}
        </h3>

        {/* Description */}
        <p style={{ fontSize: '14px', color: '#646A73', marginBottom: '16px', lineHeight: '1.5' }}>
          {description}
        </p>

        {/* Textarea (only for accept/reject) */}
        {!isDelete && (
          <div style={{ marginBottom: '20px' }}>
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder={type === 'accept' ? '附言（选填）' : '拒绝原因（选填）'}
              rows={3}
              style={{
                width: '100%',
                borderRadius: '10px',
                border: '1px solid rgba(0,0,0,0.1)',
                padding: '10px 12px',
                fontSize: '14px',
                color: '#1C2733',
                outline: 'none',
                resize: 'none',
                background: '#F5F7FA',
                lineHeight: '1.5',
              }}
              onFocus={(e) => {
                e.target.style.borderColor = '#3390EC';
                e.target.style.boxShadow = '0 0 0 3px rgba(51,144,236,0.15)';
              }}
              onBlur={(e) => {
                e.target.style.borderColor = 'rgba(0,0,0,0.1)';
                e.target.style.boxShadow = 'none';
              }}
            />
          </div>
        )}

        {/* Buttons */}
        <div className="flex items-center justify-end gap-3">
          <button
            onClick={onClose}
            disabled={loading}
            style={{
              padding: '8px 20px',
              borderRadius: '8px',
              border: '1px solid rgba(0,0,0,0.1)',
              background: '#FFFFFF',
              color: '#646A73',
              fontSize: '14px',
              fontWeight: 500,
              cursor: loading ? 'not-allowed' : 'pointer',
            }}
          >
            取消
          </button>
          <button
            onClick={handleConfirm}
            disabled={loading}
            style={{
              padding: '8px 20px',
              borderRadius: '8px',
              border: 'none',
              background: confirmColor,
              color: '#FFFFFF',
              fontSize: '14px',
              fontWeight: 500,
              cursor: loading ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              opacity: loading ? 0.7 : 1,
            }}
          >
            {loading && <Loader2 className="w-4 h-4" style={{ animation: 'spin 1s linear infinite' }} />}
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Detail Modal
   ═══════════════════════════════════════ */

interface DetailModalProps {
  request: FriendRequest | null;
  onClose: () => void;
  onAccept: (id: string) => void;
  onReject: (id: string) => void;
  onDelete: (id: string) => void;
}

function DetailModal({ request, onClose, onAccept, onReject, onDelete }: DetailModalProps) {
  if (!request) return null;

  const sc = statusConfig[request.status];
  const canAction = request.status === 'pending' && request.class === 'received';

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      {/* Backdrop */}
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />

      {/* Card */}
      <div
        className="relative"
        style={{
          background: '#FFFFFF',
          borderRadius: '16px',
          width: '90%',
          maxWidth: '480px',
          maxHeight: '85vh',
          overflow: 'auto',
          boxShadow: '0 8px 40px rgba(0,0,0,0.15)',
        }}
      >
        {/* Close */}
        <button
          onClick={onClose}
          className="absolute"
          style={{
            top: 16, right: 16,
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

        {/* Header section with avatar */}
        <div
          className="flex flex-col items-center"
          style={{ paddingTop: '28px', paddingBottom: '20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}
        >
          {/* Avatar */}
          <div className="shrink-0" style={{ marginBottom: '12px' }}>
            {request.avatar ? (
              <img
                src={request.avatar}
                alt={request.nickname}
                style={{ width: 80, height: 80, borderRadius: '50%', objectFit: 'cover' }}
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }}
              />
            ) : null}
            <div
              className="items-center justify-center"
              style={{
                width: 80,
                height: 80,
                borderRadius: '50%',
                backgroundColor: getAvatarColor(request.nickname),
                fontSize: '32px',
                fontWeight: 600,
                color: '#FFFFFF',
                display: request.avatar ? 'none' : 'flex',
              }}
            >
              {request.nickname[0]}
            </div>
          </div>

          {/* Name + sex badge */}
          <div className="flex items-center gap-2" style={{ marginBottom: '4px' }}>
            <span style={{ fontSize: '18px', fontWeight: 600, color: '#1C2733' }}>
              {request.nickname}
            </span>
            {request.sex === 'male' && (
              <span style={{ fontSize: '11px', color: '#3390EC', backgroundColor: 'rgba(51,144,236,0.1)', borderRadius: '4px', padding: '1px 5px' }}>♂</span>
            )}
            {request.sex === 'female' && (
              <span style={{ fontSize: '11px', color: '#E91E63', backgroundColor: 'rgba(233,30,99,0.1)', borderRadius: '4px', padding: '1px 5px' }}>♀</span>
            )}
          </div>

          {/* Occupation */}
          <span style={{ fontSize: '13px', color: '#646A73', marginBottom: '8px' }}>
            {request.occupation}
          </span>

          {/* Introduction */}
          {request.introduction && (
            <span style={{ fontSize: '13px', color: '#8F959E', fontStyle: 'italic', marginBottom: '10px' }}>
              {request.introduction}
            </span>
          )}

          {/* Tags */}
          {request.tags.length > 0 && (
            <div className="flex items-center gap-1.5 flex-wrap justify-center">
              {request.tags.map((tag, i) => (
                <span
                  key={i}
                  style={{
                    fontSize: '11px',
                    color: '#3390EC',
                    backgroundColor: 'rgba(51,144,236,0.08)',
                    borderRadius: '4px',
                    padding: '2px 8px',
                  }}
                >
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Request info section */}
        <div style={{ padding: '16px 20px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <div style={{ fontSize: '13px', fontWeight: 600, color: '#1C2733', marginBottom: '12px' }}>
            申请信息
          </div>

          {/* Request message */}
          <div className="flex gap-2" style={{ marginBottom: '10px' }}>
            <MessageSquare className="w-4 h-4 shrink-0" style={{ color: '#A2ACB5', marginTop: 2 }} />
            <span style={{ fontSize: '13px', color: '#646A73', lineHeight: '1.5' }}>
              {request.reqMsg}
            </span>
          </div>

          {/* Time */}
          <div className="flex items-center gap-2" style={{ marginBottom: '10px' }}>
            <Clock className="w-3.5 h-3.5" style={{ color: '#A2ACB5' }} />
            <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{formatRelativeTime(request.reqTime)}</span>
          </div>

          {/* Status */}
          <div className="flex items-center gap-2" style={{ marginBottom: request.handleMsg ? '10px' : 0 }}>
            {React.cloneElement(sc.icon as React.ReactElement<any>, { style: { color: sc.color, width: 14, height: 14 } })}
            <span
              style={{
                fontSize: '12px',
                fontWeight: 500,
                color: sc.color,
                backgroundColor: sc.bg,
                borderRadius: '4px',
                padding: '2px 8px',
              }}
            >
              {sc.label}
            </span>
            <span style={{ fontSize: '12px', color: '#A2ACB5' }}>
              ({request.class === 'received' ? '我收到的' : '我发起的'})
            </span>
          </div>

          {/* Handle message */}
          {request.handleMsg && (
            <div className="flex gap-2" style={{ marginTop: '10px' }}>
              <CheckCircle className="w-4 h-4 shrink-0" style={{ color: '#4DCD5E', marginTop: 2 }} />
              <span style={{ fontSize: '13px', color: '#646A73' }}>
                回复：{request.handleMsg}
              </span>
            </div>
          )}
        </div>

        {/* Personal info section */}
        <div style={{ padding: '16px 20px' }}>
          <div style={{ fontSize: '13px', fontWeight: 600, color: '#1C2733', marginBottom: '12px' }}>
            个人信息
          </div>
          <div className="flex flex-col gap-3">
            {request.hiChatId && (
              <div className="flex items-center gap-2">
                <UserCircle className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                <span style={{ fontSize: '13px', color: '#646A73' }}>
                  HiChat: <span style={{ color: '#3390EC' }}>{request.hiChatId}</span>
                </span>
              </div>
            )}
            {request.region && (
              <div className="flex items-center gap-2">
                <MapPin className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                <span style={{ fontSize: '13px', color: '#646A73' }}>{request.region}</span>
              </div>
            )}
            {request.email && (
              <div className="flex items-center gap-2">
                <Mail className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                <span style={{ fontSize: '13px', color: '#646A73' }}>{request.email}</span>
              </div>
            )}
            {request.phone && (
              <div className="flex items-center gap-2">
                <Phone className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                <span style={{ fontSize: '13px', color: '#646A73' }}>{request.phone}</span>
              </div>
            )}
          </div>
        </div>

        {/* Action buttons */}
        <div
          className="flex items-center justify-end gap-3"
          style={{ padding: '16px 20px', borderTop: '1px solid rgba(0,0,0,0.06)' }}
        >
          {canAction ? (
            <>
              <button
                onClick={() => onReject(request.id)}
                style={{
                  padding: '8px 20px',
                  borderRadius: '8px',
                  border: '1px solid #FF5252',
                  background: '#FFFFFF',
                  color: '#FF5252',
                  fontSize: '14px',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                拒绝
              </button>
              <button
                onClick={() => onAccept(request.id)}
                style={{
                  padding: '8px 20px',
                  borderRadius: '8px',
                  border: 'none',
                  background: '#3390EC',
                  color: '#FFFFFF',
                  fontSize: '14px',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                同意
              </button>
            </>
          ) : (
            <button
              onClick={() => onDelete(request.id)}
              style={{
                padding: '8px 20px',
                borderRadius: '8px',
                border: '1px solid rgba(0,0,0,0.1)',
                background: '#FFFFFF',
                color: '#646A73',
                fontSize: '14px',
                fontWeight: 500,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 4,
              }}
            >
              <Trash2 className="w-4 h-4" />
              删除
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Request Card
   ═══════════════════════════════════════ */

interface RequestCardProps {
  request: FriendRequest;
  onClick: (req: FriendRequest) => void;
  onAccept: (id: string) => void;
  onReject: (id: string) => void;
  onDelete: (id: string) => void;
}

function RequestCard({ request, onClick, onAccept, onReject, onDelete }: RequestCardProps) {
  const sc = statusConfig[request.status];
  const canAction = request.status === 'pending' && request.class === 'received';

  return (
    <div
      className="relative"
      style={{
        background: request.readState ? '#FFFFFF' : 'rgba(51,144,236,0.03)',
        borderLeft: request.readState ? 'none' : '3px solid #3390EC',
        borderBottom: '1px solid rgba(0,0,0,0.05)',
        padding: '12px 16px 12px 16px',
        cursor: 'pointer',
        transition: 'background 0.15s',
        boxShadow: request.readState ? 'none' : 'inset 3px 0 8px -4px rgba(51,144,236,0.15)',
      }}
      onClick={() => onClick(request)}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.background = request.readState ? '#FFFFFF' : 'rgba(51,144,236,0.03)';
      }}
    >
      <div className="flex gap-3">
        {/* Avatar */}
        <div
          className="relative shrink-0"
          style={{ width: 44, height: 44 }}
        >
          {request.avatar ? (
            <img
              src={request.avatar}
              alt={request.nickname}
              style={{ width: 44, height: 44, borderRadius: '50%', objectFit: 'cover' }}
              onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }}
            />
          ) : null}
          <div
            className="items-center justify-center"
            style={{
              width: 44,
              height: 44,
              borderRadius: '50%',
              backgroundColor: getAvatarColor(request.nickname),
              fontSize: '18px',
              fontWeight: 600,
              color: '#FFFFFF',
              display: request.avatar ? 'none' : 'flex',
            }}
          >
            {request.nickname[0]}
          </div>
          {/* Status stripe at bottom */}
          <div
            className="absolute"
            style={{
              bottom: -2,
              left: '50%',
              transform: 'translateX(-50%)',
              width: 20,
              height: 3,
              borderRadius: 2,
              backgroundColor: statusStripeColors[request.status],
            }}
          />
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Top row: name + sex + status */}
          <div className="flex items-center gap-2" style={{ marginBottom: '4px' }}>
            <span style={{ fontSize: '14px', fontWeight: 600, color: '#1C2733' }}>
              {request.nickname}
            </span>
            {request.sex === 'male' && (
              <span style={{ fontSize: '10px', color: '#3390EC', backgroundColor: 'rgba(51,144,236,0.1)', borderRadius: '3px', padding: '0px 4px', lineHeight: '16px' }}>♂</span>
            )}
            {request.sex === 'female' && (
              <span style={{ fontSize: '10px', color: '#E91E63', backgroundColor: 'rgba(233,30,99,0.1)', borderRadius: '3px', padding: '0px 4px', lineHeight: '16px' }}>♀</span>
            )}
            <span
              style={{
                fontSize: '11px',
                fontWeight: 500,
                color: sc.color,
                backgroundColor: sc.bg,
                borderRadius: '4px',
                padding: '1px 6px',
                display: 'flex',
                alignItems: 'center',
                gap: 3,
              }}
            >
              {React.cloneElement(sc.icon as React.ReactElement<any>, { style: { color: sc.color, width: 12, height: 12 } })}
              {sc.label}
            </span>
            <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', whiteSpace: 'nowrap' }}>
              {formatRelativeTime(request.reqTime)}
            </span>
          </div>

          {/* Region */}
          <div style={{ fontSize: '12px', color: '#A2ACB5', marginBottom: '4px' }}>
            {request.region} · {request.occupation}
          </div>

          {/* Request message */}
          <div
            style={{
              fontSize: '13px',
              color: '#646A73',
              lineHeight: '1.5',
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
              marginBottom: '6px',
            }}
          >
            {request.reqMsg}
          </div>

          {/* Tags */}
          {request.tags.length > 0 && (
            <div className="flex items-center gap-1.5 flex-wrap" style={{ marginBottom: '8px' }}>
              {request.tags.map((tag, i) => (
                <span
                  key={i}
                  style={{
                    fontSize: '10px',
                    color: '#3390EC',
                    backgroundColor: 'rgba(51,144,236,0.08)',
                    borderRadius: '3px',
                    padding: '1px 6px',
                  }}
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Action buttons */}
          {canAction ? (
            <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
              <button
                onClick={() => onAccept(request.id)}
                style={{
                  padding: '5px 16px',
                  borderRadius: '6px',
                  border: 'none',
                  background: '#3390EC',
                  color: '#FFFFFF',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                同意
              </button>
              <button
                onClick={() => onReject(request.id)}
                style={{
                  padding: '5px 16px',
                  borderRadius: '6px',
                  border: '1px solid #FF5252',
                  background: '#FFFFFF',
                  color: '#FF5252',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                拒绝
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
              <button
                onClick={() => onDelete(request.id)}
                style={{
                  padding: '5px 12px',
                  borderRadius: '6px',
                  border: '1px solid rgba(0,0,0,0.08)',
                  background: '#FFFFFF',
                  color: '#A2ACB5',
                  fontSize: '12px',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 3,
                }}
              >
                <Trash2 className="w-3 h-3" />
                删除
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════
   Main FriendRequestList Component
   ═══════════════════════════════════════ */

export default function FriendRequestList() {
  const { currentUser, setShowFriendRequests, friendRequestUnreadCount, setFriendRequestUnreadCount, invalidateFriends } = useIMStore();
  const token = currentUser?.token || '';

  // Local state
  const [activeTab, setActiveTab] = useState<FriendRequestClass>('received');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [requests, setRequests] = useState<FriendRequest[]>([]);
  const [detailRequest, setDetailRequest] = useState<FriendRequest | null>(null);
  const [loading, setLoading] = useState(true);

  // Send request panel state
  const [showSendPanel, setShowSendPanel] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [searching, setSearching] = useState(false);
  const [sendMsg, setSendMsg] = useState('');
  const [sendingTo, setSendingTo] = useState<string | null>(null);
  const [sendLoading, setSendLoading] = useState(false);
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Confirm dialog state
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmType, setConfirmType] = useState<'accept' | 'reject' | 'delete'>('accept');
  const [confirmTargetId, setConfirmTargetId] = useState<string>('');
  const [confirmNickname, setConfirmNickname] = useState('');
  const [confirmLoading, setConfirmLoading] = useState(false);

  // Fetch all requests on mount, then mark all as read
  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    setLoading(true);
    fetchAllRequests(token)
      .then((data) => {
        if (!cancelled) {
          setRequests(data);
          // 进入列表后自动全部标记已读，清除 badge
          const hasUnread = data.some(r => !r.readState);
          if (hasUnread) {
            apiMarkAsRead(token, 0).then(() => {
              setFriendRequestUnreadCount(0);
            });
          }
        }
      })
      .catch(() => {
        if (!cancelled) toast.error('加载好友请求失败');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [token, setFriendRequestUnreadCount]);

  // Fetch unread count on mount and sync with store
  useEffect(() => {
    if (!token) return;
    fetchUnreadCount(token).then((count) => {
      setFriendRequestUnreadCount(count);
    });
  }, [token, setFriendRequestUnreadCount]);

  // Filter requests
  const filteredRequests = useMemo(() => {
    let filtered = requests.filter(r => r.class === activeTab);
    if (statusFilter !== 'all') {
      filtered = filtered.filter(r => r.status === statusFilter);
    }
    return filtered;
  }, [requests, activeTab, statusFilter]);

  // Handlers
  const handleBack = useCallback(() => {
    setShowFriendRequests(false);
  }, [setShowFriendRequests]);

  const markAsRead = useCallback((id: string) => {
    setRequests(prev => prev.map(r => r.id === id ? { ...r, readState: true } : r));
    if (token) {
      apiMarkAsRead(token, Number(id));
    }
  }, [token]);

  const openAcceptDialog = useCallback((id: string, nickname: string) => {
    markAsRead(id);
    setConfirmType('accept');
    setConfirmTargetId(id);
    setConfirmNickname(nickname);
    setConfirmOpen(true);
  }, [markAsRead]);

  const openRejectDialog = useCallback((id: string, nickname: string) => {
    markAsRead(id);
    setConfirmType('reject');
    setConfirmTargetId(id);
    setConfirmNickname(nickname);
    setConfirmOpen(true);
  }, [markAsRead]);

  const openDeleteDialog = useCallback((id: string, nickname: string) => {
    setConfirmType('delete');
    setConfirmTargetId(id);
    setConfirmNickname(nickname);
    setConfirmOpen(true);
  }, []);

  const handleConfirm = useCallback(async (msg: string) => {
    if (!token) return;
    setConfirmLoading(true);
    try {
      const reqId = Number(confirmTargetId);

      if (confirmType === 'accept') {
        const ok = await apiHandleRequest(token, reqId, 1, msg || undefined);
        if (!ok) { toast.error('操作失败，请重试'); return; }
        setRequests(prev => prev.map(r =>
          r.id === confirmTargetId
            ? { ...r, status: 'accepted' as const, handleMsg: msg, readState: true }
            : r
        ));
        const req = requests.find(r => r.id === confirmTargetId);
        toast.success(`已同意 ${req?.nickname || ''} 的好友请求`);
        invalidateFriends();
      } else if (confirmType === 'reject') {
        const ok = await apiHandleRequest(token, reqId, 2, msg || undefined);
        if (!ok) { toast.error('操作失败，请重试'); return; }
        setRequests(prev => prev.map(r =>
          r.id === confirmTargetId
            ? { ...r, status: 'rejected' as const, handleMsg: msg, readState: true }
            : r
        ));
        const req = requests.find(r => r.id === confirmTargetId);
        toast.success(`已拒绝 ${req?.nickname || ''} 的好友请求`);
      } else {
        // delete
        const ok = await apiDeleteRequest(token, reqId);
        if (!ok) { toast.error('删除失败，请重试'); return; }
        setRequests(prev => prev.filter(r => r.id !== confirmTargetId));
        toast.success('已删除记录');
        if (detailRequest?.id === confirmTargetId) {
          setDetailRequest(null);
        }
      }

      // Refresh unread count
      fetchUnreadCount(token).then((count) => setFriendRequestUnreadCount(count));
      setConfirmOpen(false);
    } catch {
      toast.error('操作失败，请稍后重试');
    } finally {
      setConfirmLoading(false);
    }
  }, [confirmType, confirmTargetId, detailRequest, token, requests, setFriendRequestUnreadCount]);

  const handleCardClick = useCallback((req: FriendRequest) => {
    markAsRead(req.id);
    setDetailRequest(req);
  }, [markAsRead]);

  const handleDetailAccept = useCallback((id: string) => {
    const req = requests.find(r => r.id === id);
    if (req) openAcceptDialog(id, req.nickname);
  }, [requests, openAcceptDialog]);

  const handleDetailReject = useCallback((id: string) => {
    const req = requests.find(r => r.id === id);
    if (req) openRejectDialog(id, req.nickname);
  }, [requests, openRejectDialog]);

  const handleDetailDelete = useCallback((id: string) => {
    const req = requests.find(r => r.id === id);
    if (req) openDeleteDialog(id, req.nickname);
  }, [requests, openDeleteDialog]);

  // Status filter options
  const filterOptions: { key: StatusFilter; label: string }[] = [
    { key: 'all', label: '全部' },
    { key: 'pending', label: '待处理' },
    { key: 'accepted', label: '已同意' },
    { key: 'rejected', label: '已拒绝' },
    { key: 'ignored', label: '已忽略' },
  ];

  // Empty state messages
  const getEmptyMessage = () => {
    if (statusFilter === 'all') {
      return activeTab === 'received'
        ? { title: '暂无收到的请求', desc: '当有人添加你为好友时，会显示在这里' }
        : { title: '暂无发出的请求', desc: '当你添加好友时，会记录在这里' };
    }
    const filterLabels: Record<FriendRequestStatus, string> = {
      pending: '待处理', accepted: '已同意', rejected: '已拒绝', ignored: '已忽略',
    };
    return {
      title: `暂无${filterLabels[statusFilter]}的请求`,
      desc: '换个筛选条件看看',
    };
  };

  const emptyMsg = getEmptyMessage();

  // Search users for send panel
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
        // Refresh requests list and count from backend
        fetchAllRequests(token).then(setRequests);
        fetchUnreadCount(token).then(setFriendRequestUnreadCount);
      } else {
        toast.error('发送失败，请重试');
      }
    } catch {
      toast.error('发送失败，请稍后重试');
    } finally {
      setSendLoading(false);
    }
  }, [token, sendMsg]);

  return (
    <div className="h-full flex flex-col" style={{ background: '#F5F7FA' }}>
      {/* ── Header ── */}
      <div
        className="flex items-center justify-between shrink-0"
        style={{
          height: 56,
          background: '#FFFFFF',
          borderBottom: '1px solid rgba(0,0,0,0.08)',
          paddingLeft: 4,
          paddingRight: 16,
        }}
      >
        <button
          onClick={handleBack}
          style={{
            width: 40, height: 40,
            borderRadius: '50%',
            border: 'none',
            background: 'transparent',
            color: '#3390EC',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <ArrowLeft className="w-5 h-5" />
        </button>

        <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>
          新的朋友
        </span>

        <div className="flex items-center gap-1">
          {/* Add friend button */}
          <button
            onClick={() => setShowSendPanel(true)}
            style={{
              width: 40, height: 40,
              borderRadius: '50%',
              border: 'none',
              background: 'transparent',
              color: '#3390EC',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <UserPlus className="w-5 h-5" />
          </button>
          <div className="relative">
            <button
              style={{
                width: 40, height: 40,
                borderRadius: '50%',
                border: 'none',
                background: 'transparent',
                color: '#708499',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Bell className="w-5 h-5" />
            </button>
            {friendRequestUnreadCount > 0 && (
              <span
                className="absolute flex items-center justify-center"
                style={{
                  top: 6, right: 4,
                  minWidth: 18, height: 18,
                  borderRadius: 9,
                  background: '#E53935',
                  color: '#FFFFFF',
                  fontSize: '10px',
                  fontWeight: 700,
                  padding: '0 5px',
                }}
              >
                {friendRequestUnreadCount > 99 ? '99+' : friendRequestUnreadCount}
              </span>
            )}
          </div>
        </div>
      </div>

      {/* ── Tab Switch (pill style) ── */}
      <div
        className="flex items-center shrink-0"
        style={{
          padding: '12px 16px 8px',
          background: '#FFFFFF',
          borderBottom: '1px solid rgba(0,0,0,0.05)',
        }}
      >
        <div
          className="flex items-center"
          style={{
            borderRadius: '20px',
            background: 'rgba(0,0,0,0.04)',
            padding: '3px',
          }}
        >
          {(['received', 'sent'] as FriendRequestClass[]).map((tab) => (
            <button
              key={tab}
              onClick={() => { setActiveTab(tab); setStatusFilter('all'); }}
              style={{
                padding: '6px 20px',
                borderRadius: '17px',
                border: 'none',
                background: activeTab === tab ? '#3390EC' : 'transparent',
                color: activeTab === tab ? '#FFFFFF' : '#646A73',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {tab === 'received' ? '我收到的' : '我发起的'}
              {tab === 'received' && friendRequestUnreadCount > 0 && activeTab !== tab && (
                <span
                  className="inline-flex items-center justify-center"
                  style={{
                    marginLeft: 4,
                    width: 16, height: 16,
                    borderRadius: 8,
                    background: '#E53935',
                    color: '#FFFFFF',
                    fontSize: '10px',
                    fontWeight: 700,
                  }}
                >
                  {friendRequestUnreadCount > 9 ? '9+' : friendRequestUnreadCount}
                </span>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* ── Status Filter ── */}
      <div
        className="flex items-center gap-2 shrink-0 overflow-x-auto"
        style={{
          padding: '8px 16px 10px',
          background: '#FFFFFF',
          borderBottom: '1px solid rgba(0,0,0,0.05)',
        }}
      >
        {filterOptions.map((opt) => (
          <button
            key={opt.key}
            onClick={() => setStatusFilter(opt.key)}
            style={{
              padding: '4px 14px',
              borderRadius: '14px',
              border: 'none',
              background: statusFilter === opt.key ? 'rgba(51,144,236,0.1)' : 'rgba(0,0,0,0.04)',
              color: statusFilter === opt.key ? '#3390EC' : '#646A73',
              fontSize: '12px',
              fontWeight: statusFilter === opt.key ? 500 : 400,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              transition: 'all 0.15s',
            }}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* ── Request List ── */}
      <div className="flex-1 overflow-y-auto im-scroll">
        {loading ? (
          <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
            <Loader2 className="w-8 h-8" style={{ color: '#3390EC', animation: 'spin 1s linear infinite', marginBottom: '12px' }} />
            <div style={{ fontSize: '13px', color: '#A2ACB5' }}>加载中...</div>
          </div>
        ) : filteredRequests.length > 0 ? (
          <div style={{ background: '#FFFFFF', borderRadius: '12px', margin: '8px', overflow: 'hidden' }}>
            {filteredRequests.map((req) => (
              <RequestCard
                key={req.id}
                request={req}
                onClick={handleCardClick}
                onAccept={(id) => openAcceptDialog(id, req.nickname)}
                onReject={(id) => openRejectDialog(id, req.nickname)}
                onDelete={(id) => openDeleteDialog(id, req.nickname)}
              />
            ))}
          </div>
        ) : (
          <div
            className="flex flex-col items-center justify-center"
            style={{
              padding: '60px 24px',
            }}
          >
            <UserCircle className="w-16 h-16" style={{ color: '#D1D5DB', marginBottom: '16px' }} />
            <div style={{ fontSize: '15px', fontWeight: 500, color: '#646A73', marginBottom: '8px' }}>
              {emptyMsg.title}
            </div>
            <div style={{ fontSize: '13px', color: '#A2ACB5' }}>
              {emptyMsg.desc}
            </div>
          </div>
        )}
      </div>

      {/* ── Confirm Dialog ── */}
      <ConfirmDialog
        open={confirmOpen}
        type={confirmType}
        nickname={confirmNickname}
        loading={confirmLoading}
        onClose={() => { if (!confirmLoading) setConfirmOpen(false); }}
        onConfirm={handleConfirm}
      />

      {/* ── Detail Modal ── */}
      <DetailModal
        request={detailRequest}
        onClose={() => setDetailRequest(null)}
        onAccept={handleDetailAccept}
        onReject={handleDetailReject}
        onDelete={handleDetailDelete}
      />

      {/* ── Send Friend Request Panel ── */}
      {showSendPanel && (
        <div
          className="fixed inset-0"
          style={{ zIndex: 10001, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
          onClick={(e) => { if (e.target === e.currentTarget) { setShowSendPanel(false); setSearchQuery(''); setSearchResults([]); setSendingTo(null); setSendMsg(''); } }}
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
                onClick={() => { setShowSendPanel(false); setSearchQuery(''); setSearchResults([]); setSendingTo(null); setSendMsg(''); }}
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
      )}
    </div>
  );
}
