'use client';

import React, { useEffect, useState } from 'react';
import { Users, UserPlus, Send, Loader2, X, Clock } from 'lucide-react';
import { getAvatarColor } from '@/lib/utils';
import { useIMStore } from '@/lib/im-store';
import { getGroupDetail, sendGroupRequest, type GroupSearchItem } from '@/lib/friend-group-api';
import { toast } from 'sonner';

interface GroupProfileCardProps {
  group: GroupSearchItem;
  onClose: () => void;
  /** 进入群聊（已是成员时） */
  onEnterGroup?: (groupId: string) => void;
}

/**
 * 群资料卡片 —— 搜索结果点击后展示群信息并据成员状态发起加群申请。
 * 复用 GET /api/social/group/detail 拉取成员以判断「是否已在群内」。
 */
export default function GroupProfileCard({ group, onClose, onEnterGroup }: GroupProfileCardProps) {
  const { currentUser } = useIMStore();
  const token = currentUser?.token || '';
  const myId = currentUser?.id || '';

  const [isMember, setIsMember] = useState(false);
  const [memberCount, setMemberCount] = useState(group.member_count);
  const [ownerName, setOwnerName] = useState('');
  const [checking, setChecking] = useState(true);
  const [sending, setSending] = useState(false);
  const [applied, setApplied] = useState(false);
  const [showMsgInput, setShowMsgInput] = useState(false);
  const [reqMsg, setReqMsg] = useState('');

  useEffect(() => {
    let cancelled = false;
    if (!token) { setChecking(false); return; }
    getGroupDetail(token, group.id).then((detail) => {
      if (cancelled || !detail) { setChecking(false); return; }
      const members = (detail.members || []) as Array<{ user_id?: string; UserId?: string; nickname?: string; role_level?: number }>;
      setIsMember(members.some((m) => (m.user_id || m.UserId) === myId));
      if (members.length) setMemberCount(members.length);
      // 群主：creator_uid 对应成员，或 role_level 最高者
      const owner = members.find((m) => (m.user_id || m.UserId) === group.creator_uid)
        || members.find((m) => m.role_level === 3 || m.role_level === 2);
      if (owner?.nickname) setOwnerName(owner.nickname);
      setChecking(false);
    });
    return () => { cancelled = true; };
  }, [token, group.id, myId, group.creator_uid]);

  const handleApply = async () => {
    if (!token) return;
    setSending(true);
    try {
      const ok = await sendGroupRequest(token, group.id, reqMsg || undefined);
      if (ok) {
        toast.success('入群申请已发送');
        setApplied(true);
        setShowMsgInput(false);
        setReqMsg('');
      } else {
        toast.error('发送失败，请重试');
      }
    } catch {
      toast.error('发送失败，请稍后重试');
    } finally {
      setSending(false);
    }
  };

  const btn = (label: string, onClick: () => void, icon: React.ReactNode, disabled = false) => (
    <button
      onClick={onClick}
      disabled={disabled}
      style={{
        width: '100%', height: 40, borderRadius: 8, border: 'none',
        background: disabled ? '#B9C4CE' : '#3390EC', color: '#FFFFFF',
        fontSize: 14, fontWeight: 600, cursor: disabled ? 'not-allowed' : 'pointer',
        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
      }}
    >
      {icon}{label}
    </button>
  );

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10002, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFFFFF', borderRadius: 16, width: '88%', maxWidth: 360, padding: 20, boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <button
          onClick={onClose}
          style={{ position: 'absolute', top: 12, right: 12, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'rgba(0,0,0,0.04)', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <X className="w-4 h-4" />
        </button>

        {/* 群头像 + 名称 */}
        <div className="flex flex-col items-center" style={{ paddingTop: 8 }}>
          <div style={{ width: 72, height: 72, borderRadius: 16, overflow: 'hidden', marginBottom: 12 }}>
            {group.icon ? (
              <img src={group.icon} alt={group.name} style={{ width: 72, height: 72, objectFit: 'cover' }} />
            ) : (
              <div className="flex items-center justify-center" style={{ width: 72, height: 72, background: getAvatarColor(group.name || '群'), color: '#FFF', fontSize: 28, fontWeight: 600 }}>
                {(group.name || '群')[0]}
              </div>
            )}
          </div>
          <div style={{ fontSize: 17, fontWeight: 600, color: '#1C2733' }}>{group.name || '未命名群聊'}</div>
          <div className="flex items-center gap-1" style={{ fontSize: 13, color: '#A2ACB5', marginTop: 4 }}>
            <Users className="w-3.5 h-3.5" />
            {memberCount} 名成员
          </div>
          <div style={{ fontSize: 12, color: '#C4CDD5', marginTop: 2 }}>群号: {group.id}</div>
          {ownerName && <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 2 }}>群主: {ownerName}</div>}
        </div>

        {/* 群简介 */}
        {group.description && (
          <div style={{ marginTop: 14, padding: '10px 12px', background: '#F5F7FA', borderRadius: 8, fontSize: 13, color: '#646A73', lineHeight: 1.6 }}>
            {group.description}
          </div>
        )}

        {/* 操作 */}
        <div style={{ marginTop: 18 }}>
          {checking ? (
            <div className="flex items-center justify-center" style={{ height: 40, color: '#A2ACB5' }}>
              <Loader2 className="w-4 h-4" style={{ animation: 'spin 1s linear infinite' }} />
            </div>
          ) : isMember ? (
            btn('进入群聊', () => onEnterGroup?.(group.id), <Send size={15} />)
          ) : applied ? (
            btn('已申请', () => {}, <Clock size={15} />, true)
          ) : showMsgInput ? (
            <div className="flex flex-col gap-2">
              <input
                type="text"
                value={reqMsg}
                onChange={(e) => setReqMsg(e.target.value)}
                placeholder="申请附言（选填）"
                autoFocus
                style={{ width: '100%', height: 38, border: '1px solid rgba(0,0,0,0.1)', borderRadius: 8, padding: '0 12px', fontSize: 13, outline: 'none' }}
              />
              {btn(sending ? '发送中…' : '发送申请', handleApply, sending ? <Loader2 className="w-4 h-4" style={{ animation: 'spin 1s linear infinite' }} /> : <UserPlus size={15} />, sending)}
            </div>
          ) : (
            btn('申请加入', () => setShowMsgInput(true), <UserPlus size={15} />)
          )}
        </div>
      </div>
    </div>
  );
}
