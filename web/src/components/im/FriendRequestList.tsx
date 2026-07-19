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
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { toast } from 'sonner';
import { useIMStore } from '@/lib/im-store';
import { getAvatarColor, tagColor } from '@/lib/utils';
import { useT } from '@/hooks/use-i18n';
import AddFriendPanel from './AddFriendPanel';
import { deleteFriendRequest, handleFriendRequest, listFriendRequests, markFriendRequestsRead, type FriendRequest, type FriendRequestClass, type FriendRequestStatus, type FriendRequestStatusFilter } from '@/lib/social-request-api';

/* ═══════════════════════════════════════
   Helpers
   ═══════════════════════════════════════ */

function formatRelativeTime(date: Date, t: (k: string) => string): string {
  const now = Date.now();
  const diff = now - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return t('group.time.justNow');
  if (minutes < 60) return t('group.time.minutesAgo').replace('{m}', String(minutes));
  if (hours < 24) return t('group.time.hoursAgo').replace('{h}', String(hours));
  if (days < 30) {
    const d = new Date(date);
    return t('group.time.monthDay').replace('{month}', String(d.getMonth() + 1)).replace('{day}', String(d.getDate()));
  }
  const d = new Date(date);
  return t('group.time.yearMonthDay').replace('{year}', String(d.getFullYear())).replace('{month}', String(d.getMonth() + 1)).replace('{day}', String(d.getDate()));
}

const statusConfig: Record<FriendRequestStatus, { labelKey: string; color: string; bg: string; icon: React.ReactNode }> = {
  pending: {
    labelKey: 'friend.status.pending',
    color: '#F5A623',
    bg: 'rgba(245,166,35,0.1)',
    icon: <AlertCircle className="w-3.5 h-3.5" />,
  },
  accepted: {
    labelKey: 'friend.status.accepted',
    color: '#4DCD5E',
    bg: 'rgba(77,205,94,0.1)',
    icon: <CheckCircle className="w-3.5 h-3.5" />,
  },
  rejected: {
    labelKey: 'friend.status.rejected',
    color: '#FF5252',
    bg: 'rgba(255,82,82,0.1)',
    icon: <XCircle className="w-3.5 h-3.5" />,
  },
  ignored: {
    labelKey: 'friend.status.ignored',
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

type StatusFilter = FriendRequestStatusFilter;

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
  const t = useT();

  if (!open) return null;

  const isDelete = type === 'delete';
  const title = isDelete
    ? t('friend.confirm.deleteTitle')
    : type === 'accept'
      ? t('friend.confirm.acceptTitle')
      : t('friend.confirm.rejectTitle');
  const description = isDelete
    ? t('friend.confirm.deleteDesc').replace('{name}', nickname)
    : type === 'accept'
      ? t('friend.confirm.acceptDesc').replace('{name}', nickname)
      : t('friend.confirm.rejectDesc').replace('{name}', nickname);
  const confirmLabel = isDelete ? t('friend.delete') : type === 'accept' ? t('group.agree') : t('group.reject');
  const confirmColor = isDelete ? '#E53935' : type === 'accept' ? '#1BB45B' : '#E53935';

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
              placeholder={type === 'accept' ? t('friend.acceptMsgPh') : t('friend.rejectMsgPh')}
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
                e.target.style.borderColor = '#1BB45B';
                e.target.style.boxShadow = '0 0 0 3px rgba(27,180,91,0.15)';
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
            {t('common.cancel')}
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
  const t = useT();
  if (!request) return null;

  const sc = statusConfig[request.status];
  const canAction = request.actionable;

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
              <span style={{ fontSize: '11px', color: '#2D7FF9', backgroundColor: 'rgba(45,127,249,0.1)', borderRadius: '4px', padding: '1px 5px' }}>♂</span>
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
                    color: tagColor(tag).c,
                    backgroundColor: tagColor(tag).b,
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
            {t('friend.reqInfo')}
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
            <span style={{ fontSize: '12px', color: '#A2ACB5' }}>{formatRelativeTime(request.reqTime, t)}</span>
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
              {t(sc.labelKey)}
            </span>
            <span style={{ fontSize: '12px', color: '#A2ACB5' }}>
              ({request.class === 'received' ? t('group.app.received') : t('group.app.sent')})
            </span>
          </div>

          {/* Handle message */}
          {request.handleMsg && (
            <div className="flex gap-2" style={{ marginTop: '10px' }}>
              <CheckCircle className="w-4 h-4 shrink-0" style={{ color: '#4DCD5E', marginTop: 2 }} />
              <span style={{ fontSize: '13px', color: '#646A73' }}>
                {t('friend.replyPrefix').replace('{msg}', request.handleMsg)}
              </span>
            </div>
          )}
        </div>

        {/* Personal info section */}
        <div style={{ padding: '16px 20px' }}>
          <div style={{ fontSize: '13px', fontWeight: 600, color: '#1C2733', marginBottom: '12px' }}>
            {t('friend.personalInfo')}
          </div>
          <div className="flex flex-col gap-3">
            {request.hiChatId && (
              <div className="flex items-center gap-2">
                <UserCircle className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                <span style={{ fontSize: '13px', color: '#646A73' }}>
                  HiChat: <span style={{ color: '#1C2733' }}>{request.hiChatId}</span>
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
                {t('group.reject')}
              </button>
              <button
                onClick={() => onAccept(request.id)}
                style={{
                  padding: '8px 20px',
                  borderRadius: '8px',
                  border: 'none',
                  background: '#1BB45B',
                  color: '#FFFFFF',
                  fontSize: '14px',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                {t('group.agree')}
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
              {t('friend.delete')}
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
  const t = useT();
  const sc = statusConfig[request.status];
  const canAction = request.actionable;

  return (
    <div
      className="relative"
      style={{
        background: request.readState ? '#FFFFFF' : 'rgba(27,180,91,0.03)',
        borderLeft: request.readState ? 'none' : '3px solid #1BB45B',
        borderBottom: '1px solid rgba(0,0,0,0.05)',
        padding: '12px 16px 12px 16px',
        cursor: 'pointer',
        transition: 'background 0.15s',
        boxShadow: request.readState ? 'none' : 'inset 3px 0 8px -4px rgba(27,180,91,0.15)',
      }}
      onClick={() => onClick(request)}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.02)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.background = request.readState ? '#FFFFFF' : 'rgba(27,180,91,0.03)';
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
              <span style={{ fontSize: '10px', color: '#2D7FF9', backgroundColor: 'rgba(45,127,249,0.1)', borderRadius: '3px', padding: '0px 4px', lineHeight: '16px' }}>♂</span>
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
              {t(sc.labelKey)}
            </span>
            <span style={{ fontSize: '11px', color: '#A2ACB5', marginLeft: 'auto', whiteSpace: 'nowrap' }}>
              {formatRelativeTime(request.reqTime, t)}
            </span>
          </div>

          {request.handledAt && (
            <div style={{ fontSize: '11px', color: '#A2ACB5', marginBottom: '4px' }}>
              {t('friend.requests.handledAt').replace('{time}', formatRelativeTime(request.handledAt, t))}
            </div>
          )}

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
                    color: tagColor(tag).c,
                    backgroundColor: tagColor(tag).b,
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
                  background: '#1BB45B',
                  color: '#FFFFFF',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                {t('group.agree')}
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
                {t('group.reject')}
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
                {t('friend.delete')}
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
  const { currentUser, setShowFriendRequests, friendRequestUnreadCount, friendRequestUnread, setFriendRequestUnread, invalidateFriends, friendRequestsVersion, invalidateFriendRequests, refreshFriendRequestUnread, friendReqNavTarget, clearFriendReqNavTarget } = useIMStore();
  const t = useT();
  const loadRequestFailureText = t('friend.loadReqFail');
  const locationNotFoundText = t('friend.requests.locationNotFound');
  const token = currentUser?.token || '';

  // Local state
  const [activeTab, setActiveTab] = useState<FriendRequestClass>('received');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [page, setPage] = useState(1);
  const [navigationVersion, setNavigationVersion] = useState(0);
  const [total, setTotal] = useState(0);
  const [committedQuery, setCommittedQuery] = useState('');
  const [loadError, setLoadError] = useState(false);
  const [readError, setReadError] = useState(false);
  const pageSize = 20;

  // 通知点击带来的子 tab 跳转意图（received=我收到 / sent=我发起），消费后清除
  useEffect(() => {
    if (!friendReqNavTarget) return;
    locatorTarget.current = friendReqNavTarget;
    setActiveTab(friendReqNavTarget.tab);
    setStatusFilter('all');
    setPage(1);
    setNavigationVersion(version => version + 1);
  }, [friendReqNavTarget]);

  const [requests, setRequests] = useState<FriendRequest[]>([]);
  const [detailRequest, setDetailRequest] = useState<FriendRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const requestGeneration = useRef(0);
  const mutationGeneration = useRef(0);
  const locatorTarget = useRef<{ tab: FriendRequestClass; requestId?: string } | null>(null);
  const currentQuery = `${token}:${activeTab}:${statusFilter}:${page}`;
  const visibleRequests = committedQuery === currentQuery ? requests : [];
  const visibleDetailRequest = committedQuery === currentQuery ? detailRequest : null;

  // Send request panel state
  const [showSendPanel, setShowSendPanel] = useState(false);

  // Confirm dialog state
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmType, setConfirmType] = useState<'accept' | 'reject' | 'delete'>('accept');
  const [confirmTargetId, setConfirmTargetId] = useState<string>('');
  const [confirmNickname, setConfirmNickname] = useState('');
  const [confirmLoading, setConfirmLoading] = useState(false);

  // Fetch one active page, then mark only visible unread receipts and refresh the authoritative count.
  useEffect(() => {
    if (!token) return;
    const generation = ++requestGeneration.current;
    setLoading(true);
    setLoadError(false);
    setReadError(false);
    listFriendRequests(token, activeTab, statusFilter, page, pageSize)
      .then(async ({ list, total: nextTotal }) => {
        if (generation === requestGeneration.current) {
          const maxPage = Math.max(1, Math.ceil(nextTotal / pageSize));
          if (page > maxPage) {
            setPage(maxPage);
            return;
          }
          const locator = locatorTarget.current;
          if (locator?.requestId) {
            const target = list.find(request => request.id === locator.requestId);
            if (!target && page < maxPage) {
              setPage(current => current + 1);
              return;
            }
            if (!target) {
              toast.error(locationNotFoundText);
            } else {
              setDetailRequest(target);
            }
          }
          setRequests(list);
          setTotal(nextTotal);
          setCommittedQuery(currentQuery);
          const unreadIds = list.filter(request => !request.readState).map(request => request.id);
          if (unreadIds.length > 0) {
            try {
              const unread = await markFriendRequestsRead(token, unreadIds);
              if (generation !== requestGeneration.current) return;
              setRequests(current => current.map(request => unreadIds.includes(request.id) ? { ...request, readState: true } : request));
              setFriendRequestUnread(unread);
            } catch {
              if (generation === requestGeneration.current) setReadError(true);
            }
          } else {
            await refreshFriendRequestUnread();
          }
          if (generation === requestGeneration.current && locator) {
            locatorTarget.current = null;
            clearFriendReqNavTarget();
          }
        }
      })
      .catch(() => {
        if (generation === requestGeneration.current) {
          setLoadError(true);
          toast.error(loadRequestFailureText);
        }
      })
      .finally(() => {
        if (generation === requestGeneration.current) setLoading(false);
      });
    return () => { requestGeneration.current += 1; };
  }, [token, activeTab, statusFilter, page, currentQuery, friendRequestsVersion, navigationVersion, refreshFriendRequestUnread, setFriendRequestUnread, clearFriendReqNavTarget, loadRequestFailureText, locationNotFoundText]);

  // Handlers
  const handleBack = useCallback(() => {
    setShowFriendRequests(false);
  }, [setShowFriendRequests]);

  const openAcceptDialog = useCallback((id: string, nickname: string) => {
    setConfirmType('accept');
    setConfirmTargetId(id);
    setConfirmNickname(nickname);
    setConfirmOpen(true);
  }, []);

  const openRejectDialog = useCallback((id: string, nickname: string) => {
    setConfirmType('reject');
    setConfirmTargetId(id);
    setConfirmNickname(nickname);
    setConfirmOpen(true);
  }, []);

  const openDeleteDialog = useCallback((id: string, nickname: string) => {
    setConfirmType('delete');
    setConfirmTargetId(id);
    setConfirmNickname(nickname);
    setConfirmOpen(true);
  }, []);

  const handleConfirm = useCallback(async (msg: string) => {
    if (!token) return;
    const mutation = ++mutationGeneration.current;
    setConfirmLoading(true);
    try {
      if (confirmType === 'accept') {
        await handleFriendRequest(token, confirmTargetId, 1, msg || undefined);
        if (mutation !== mutationGeneration.current || useIMStore.getState().currentUser?.token !== token) return;
        setRequests(prev => prev.map(r =>
          r.id === confirmTargetId
            ? { ...r, status: 'accepted' as const, handleMsg: msg, readState: true }
            : r
        ));
        const req = requests.find(r => r.id === confirmTargetId);
        toast.success(t('friend.acceptedToast').replace('{name}', req?.nickname || ''));
        invalidateFriends();
      } else if (confirmType === 'reject') {
        await handleFriendRequest(token, confirmTargetId, 2, msg || undefined);
        if (mutation !== mutationGeneration.current || useIMStore.getState().currentUser?.token !== token) return;
        setRequests(prev => prev.map(r =>
          r.id === confirmTargetId
            ? { ...r, status: 'rejected' as const, handleMsg: msg, readState: true }
            : r
        ));
        const req = requests.find(r => r.id === confirmTargetId);
        toast.success(t('friend.rejectedToast').replace('{name}', req?.nickname || ''));
      } else {
        // delete
        await deleteFriendRequest(token, confirmTargetId);
        if (mutation !== mutationGeneration.current || useIMStore.getState().currentUser?.token !== token) return;
        setRequests(prev => prev.filter(r => r.id !== confirmTargetId));
        toast.success(t('friend.recordDeleted'));
        if (detailRequest?.id === confirmTargetId) {
          setDetailRequest(null);
        }
      }

      if (mutation !== mutationGeneration.current || useIMStore.getState().currentUser?.token !== token) return;

      requestGeneration.current += 1;
      invalidateFriendRequests();
      void refreshFriendRequestUnread();
      setConfirmOpen(false);
    } catch {
      toast.error(t('friend.opFailLater'));
    } finally {
      if (mutation === mutationGeneration.current) setConfirmLoading(false);
    }
  }, [confirmType, confirmTargetId, detailRequest, token, requests, invalidateFriendRequests, refreshFriendRequestUnread]);

  const handleCardClick = useCallback((req: FriendRequest) => {
    setDetailRequest(req);
  }, []);

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
    { key: 'all', label: t('group.filter.all') },
    { key: 'pending', label: t('friend.status.pending') },
    { key: 'accepted', label: t('friend.status.accepted') },
    { key: 'rejected', label: t('friend.status.rejected') },
    { key: 'ignored', label: t('friend.status.ignored') },
  ];

  // Empty state messages
  const getEmptyMessage = () => {
    if (statusFilter === 'all') {
      return activeTab === 'received'
        ? { title: t('friend.emptyReceivedTitle'), desc: t('friend.emptyReceivedDesc') }
        : { title: t('friend.emptySentTitle'), desc: t('friend.emptySentDesc') };
    }
    const filterLabels: Record<FriendRequestStatus, string> = {
      pending: t('friend.status.pending'), accepted: t('friend.status.accepted'), rejected: t('friend.status.rejected'), ignored: t('friend.status.ignored'),
    };
    return {
      title: t('friend.emptyFilterTitle').replace('{status}', filterLabels[statusFilter]),
      desc: t('group.emptyMatchHint'),
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
            color: '#1BB45B',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <ArrowLeft className="w-5 h-5" />
        </button>

        <span style={{ fontSize: '17px', fontWeight: 600, color: '#1C2733' }}>
          {t('contact.newFriends')}
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
              color: '#1BB45B',
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
              onClick={() => { setActiveTab(tab); setStatusFilter('all'); setPage(1); }}
              style={{
                padding: '6px 20px',
                borderRadius: '17px',
                border: 'none',
                background: activeTab === tab ? '#1BB45B' : 'transparent',
                color: activeTab === tab ? '#FFFFFF' : '#646A73',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {tab === 'received' ? t('group.app.received') : t('group.app.sent')}
              {friendRequestUnread[tab === 'received' ? 'apply' : 'result'] > 0 && activeTab !== tab && (
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
                  {friendRequestUnread[tab === 'received' ? 'apply' : 'result'] > 9 ? '9+' : friendRequestUnread[tab === 'received' ? 'apply' : 'result']}
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
            onClick={() => { setStatusFilter(opt.key); setPage(1); }}
            style={{
              padding: '4px 14px',
              borderRadius: '14px',
              border: 'none',
              background: statusFilter === opt.key ? 'rgba(27,180,91,0.1)' : 'rgba(0,0,0,0.04)',
              color: statusFilter === opt.key ? '#1BB45B' : '#646A73',
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
        {(loadError || readError) && (
          <div style={{ margin: '8px', padding: '10px 12px', borderRadius: 8, background: '#FFF3E8', color: '#AD6800', fontSize: 12 }}>
            {loadError ? t('friend.requests.staleData') : t('friend.requests.markReadFailed')}
          </div>
        )}
        {loading ? (
          <div className="flex flex-col items-center justify-center" style={{ padding: '60px 24px' }}>
            <Loader2 className="w-8 h-8" style={{ color: '#1BB45B', animation: 'spin 1s linear infinite', marginBottom: '12px' }} />
            <div style={{ fontSize: '13px', color: '#A2ACB5' }}>{t('common.loading')}</div>
          </div>
        ) : visibleRequests.length > 0 ? (
          <div style={{ background: '#FFFFFF', borderRadius: '12px', margin: '8px', overflow: 'hidden' }}>
            {visibleRequests.map((req) => (
              <RequestCard
                key={req.id}
                request={req}
                onClick={handleCardClick}
                onAccept={(id) => openAcceptDialog(id, req.nickname)}
                onReject={(id) => openRejectDialog(id, req.nickname)}
                onDelete={(id) => openDeleteDialog(id, req.nickname)}
              />
            ))}
            {Math.ceil(total / pageSize) > 1 && (
              <div className="flex items-center justify-center gap-3" style={{ padding: '12px' }}>
                <button disabled={loading || page <= 1} onClick={() => setPage(current => Math.max(1, current - 1))} className="flex items-center gap-1" style={{ opacity: page <= 1 ? 0.4 : 1 }}>
                  <ChevronLeft className="w-4 h-4" />{t('friend.requests.pagination.previous')}
                </button>
                <span style={{ fontSize: 12, color: '#646A73' }}>{t('friend.requests.pagination.page').replace('{page}', String(page)).replace('{total}', String(Math.ceil(total / pageSize)))}</span>
                <button disabled={loading || page >= Math.ceil(total / pageSize)} onClick={() => setPage(current => Math.min(Math.ceil(total / pageSize), current + 1))} className="flex items-center gap-1" style={{ opacity: page >= Math.ceil(total / pageSize) ? 0.4 : 1 }}>
                  {t('friend.requests.pagination.next')}<ChevronRight className="w-4 h-4" />
                </button>
              </div>
            )}
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
        request={visibleDetailRequest}
        onClose={() => setDetailRequest(null)}
        onAccept={handleDetailAccept}
        onReject={handleDetailReject}
        onDelete={handleDetailDelete}
      />

      {/* ── Send Friend Request Panel ── */}
      <AddFriendPanel
        open={showSendPanel}
        onClose={() => setShowSendPanel(false)}
        onSent={() => {
          requestGeneration.current += 1;
          invalidateFriendRequests();
          void refreshFriendRequestUnread();
        }}
      />
    </div>
  );
}
