'use client';

import React, { useState, useMemo, useCallback } from 'react';
import {
  ArrowLeft,
  Bell,
  UserPlus,
  UserMinus,
  UserCheck,
  UserX,
  Trash2,
  X,
  Loader2,
  UserCircle,
  Mail,
  Phone,
  MapPin,
  Briefcase,
  MessageSquare,
  Tag,
  Clock,
  CheckCircle,
  XCircle,
  MinusCircle,
  AlertCircle,
} from 'lucide-react';
import { toast } from 'sonner';
import { useIMStore } from '@/lib/im-store';
import {
  friendRequests as initialRequests,
  type FriendRequest,
  type FriendRequestClass,
  type FriendRequestStatus,
} from '@/lib/mock-data';
import { getAvatarColor } from '@/lib/utils';

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
          <div
            className="flex items-center justify-center shrink-0"
            style={{
              width: 80,
              height: 80,
              borderRadius: '50%',
              backgroundColor: getAvatarColor(request.nickname),
              fontSize: '32px',
              fontWeight: 600,
              color: '#FFFFFF',
              marginBottom: '12px',
            }}
          >
            {request.nickname[0]}
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
            {React.cloneElement(sc.icon as React.ReactElement, { style: { color: sc.color, width: 14, height: 14 } })}
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
          style={{
            width: 44,
            height: 44,
            borderRadius: '50%',
            backgroundColor: getAvatarColor(request.nickname),
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '18px',
            fontWeight: 600,
            color: '#FFFFFF',
          }}
        >
          {request.nickname[0]}
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
              {React.cloneElement(sc.icon as React.ReactElement, { style: { color: sc.color, width: 12, height: 12 } })}
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
  const { setShowFriendRequests, friendRequestUnreadCount, setFriendRequestUnreadCount } = useIMStore();

  // Local state
  const [activeTab, setActiveTab] = useState<FriendRequestClass>('received');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [requests, setRequests] = useState<FriendRequest[]>(initialRequests);
  const [detailRequest, setDetailRequest] = useState<FriendRequest | null>(null);

  // Confirm dialog state
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmType, setConfirmType] = useState<'accept' | 'reject' | 'delete'>('accept');
  const [confirmTargetId, setConfirmTargetId] = useState<string>('');
  const [confirmNickname, setConfirmNickname] = useState('');
  const [confirmLoading, setConfirmLoading] = useState(false);

  // Compute unread count and sync with store
  const unreadCount = useMemo(() => {
    return requests.filter(r => !r.readState).length;
  }, [requests]);

  React.useEffect(() => {
    setFriendRequestUnreadCount(unreadCount);
  }, [unreadCount, setFriendRequestUnreadCount]);

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
  }, []);

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

  const handleConfirm = useCallback((msg: string) => {
    setConfirmLoading(true);
    // Simulate async operation
    setTimeout(() => {
      setRequests(prev => {
        const newReqs = [...prev];
        const idx = newReqs.findIndex(r => r.id === confirmTargetId);
        if (idx === -1) return prev;
        if (confirmType === 'accept') {
          newReqs[idx] = { ...newReqs[idx], status: 'accepted' as const, handleMsg: msg, readState: true };
          toast.success(`已同意 ${newReqs[idx].nickname} 的好友请求`);
        } else if (confirmType === 'reject') {
          newReqs[idx] = { ...newReqs[idx], status: 'rejected' as const, handleMsg: msg, readState: true };
          toast.success(`已拒绝 ${newReqs[idx].nickname} 的好友请求`);
        } else {
          return prev.filter(r => r.id !== confirmTargetId);
        }
        return newReqs;
      });
      if (confirmType === 'delete') {
        toast.success('已删除记录');
        if (detailRequest?.id === confirmTargetId) {
          setDetailRequest(null);
        }
      }
      setConfirmLoading(false);
      setConfirmOpen(false);
    }, 600);
  }, [confirmType, confirmTargetId, detailRequest]);

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
        {filteredRequests.length > 0 ? (
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
    </div>
  );
}
