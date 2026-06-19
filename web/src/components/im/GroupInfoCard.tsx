'use client';

import React, { useEffect, useState } from 'react';
import { Users, ChevronRight, LogOut, X } from 'lucide-react';
import { getAvatarColor } from '@/lib/utils';
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { getGroupDetail } from '@/lib/friend-group-api';
import { useT } from '@/hooks/use-i18n';

interface GroupInfoCardProps {
  groupId: string;
  name: string;
  avatar?: string;
  memberCount: number;
  onClose: () => void;
  /** 深链到通讯录群详情（成员管理/公告/邀请等） */
  onOpenManage: () => void;
  /** 退出群聊（二次确认 + 接口在 ChatDetail 处理） */
  onQuit: () => void;
}

/** 展示数量上限：超出折叠为「+N」 */
const PREVIEW_LIMIT = 14;

/**
 * 群信息卡片 —— 群聊会话框点头像 / "..."菜单「群信息」弹出。
 * 轻量展示群头像/名称/群号/成员预览，重操作（成员管理等）深链到通讯录群详情。
 */
export default function GroupInfoCard({ groupId, name, avatar, memberCount, onClose, onOpenManage, onQuit }: GroupInfoCardProps) {
  const t = useT();
  const token = useIMStore(s => s.currentUser?.token) || '';
  const members = useChatStore(s => s.groupMembers[groupId]) || [];
  const [description, setDescription] = useState('');

  useEffect(() => {
    let cancelled = false;
    if (!token) return;
    getGroupDetail(token, groupId).then((detail) => {
      if (cancelled || !detail) return;
      const desc = (detail.group?.description || '').trim();
      if (desc) setDescription(desc);
    });
    return () => { cancelled = true; };
  }, [token, groupId]);

  const count = members.length || memberCount || 0;
  const owner = members.find(m => m.role_level === 2);
  const ownerName = owner ? (owner.group_nickname || owner.nickname || owner.user_id) : '';
  const preview = members.slice(0, PREVIEW_LIMIT);
  const overflow = Math.max(0, count - preview.length);

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10002, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div className="absolute inset-0" style={{ background: 'rgba(0,0,0,0.5)' }} />
      <div className="relative" style={{ background: '#FFFFFF', borderRadius: 16, width: '88%', maxWidth: 380, padding: 20, boxShadow: '0 8px 40px rgba(0,0,0,0.15)' }}>
        <button
          onClick={onClose}
          style={{ position: 'absolute', top: 12, right: 12, width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'rgba(0,0,0,0.04)', color: '#A2ACB5', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <X className="w-4 h-4" />
        </button>

        {/* 群头像 + 名称 */}
        <div className="flex flex-col items-center" style={{ paddingTop: 8 }}>
          <div style={{ width: 72, height: 72, borderRadius: 16, overflow: 'hidden', marginBottom: 12 }}>
            {avatar ? (
              <img src={avatar} alt={name} style={{ width: 72, height: 72, objectFit: 'cover' }} />
            ) : (
              <div className="flex items-center justify-center" style={{ width: 72, height: 72, background: getAvatarColor(name || 'G'), color: '#FFF', fontSize: 28, fontWeight: 600 }}>
                {(name || 'G')[0]}
              </div>
            )}
          </div>
          <div style={{ fontSize: 17, fontWeight: 600, color: '#1C2733' }}>{name || t('group.groupFallback')}</div>
          <div className="flex items-center gap-1" style={{ fontSize: 13, color: '#A2ACB5', marginTop: 4 }}>
            <Users className="w-3.5 h-3.5" />
            {t('group.memberCount').replace('{count}', String(count))}
          </div>
          <div style={{ fontSize: 12, color: '#C4CDD5', marginTop: 2 }}>{t('group.id')}: {groupId}</div>
          {ownerName && <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 2 }}>{t('group.role.owner')}: {ownerName}</div>}
        </div>

        {/* 群简介 */}
        {description && (
          <div style={{ marginTop: 14, padding: '10px 12px', background: '#F5F7FA', borderRadius: 8, fontSize: 13, color: '#646A73', lineHeight: 1.6 }}>
            {description}
          </div>
        )}

        {/* 成员预览 */}
        {preview.length > 0 && (
          <div style={{ marginTop: 16, display: 'flex', flexWrap: 'wrap', gap: 10 }}>
            {preview.map((m) => {
              const mName = m.group_nickname || m.nickname || m.user_id;
              return (
                <div key={m.user_id} style={{ width: 40, textAlign: 'center' }}>
                  {m.user_avatar_url ? (
                    <img src={m.user_avatar_url} alt="" style={{ width: 40, height: 40, borderRadius: 8, objectFit: 'cover' }} />
                  ) : (
                    <div className="flex items-center justify-center" style={{ width: 40, height: 40, borderRadius: 8, background: getAvatarColor(mName), color: '#FFF', fontSize: 15, fontWeight: 600 }}>
                      {mName[0]}
                    </div>
                  )}
                </div>
              );
            })}
            {overflow > 0 && (
              <div className="flex items-center justify-center" style={{ width: 40, height: 40, borderRadius: 8, background: '#F0F2F5', color: '#A2ACB5', fontSize: 13, fontWeight: 600 }}>
                +{overflow}
              </div>
            )}
          </div>
        )}

        {/* 查看全部成员 / 群管理 */}
        <button
          onClick={onOpenManage}
          style={{ marginTop: 16, width: '100%', height: 44, borderRadius: 10, border: '1px solid rgba(0,0,0,0.08)', background: '#FFFFFF', color: '#1C2733', fontSize: 14, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 14px' }}
        >
          <span className="flex items-center gap-2"><Users className="w-4 h-4" style={{ color: '#1BB45B' }} />{t('group.viewAllMembers')}</span>
          <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
        </button>

        {/* 退出群聊 */}
        <button
          onClick={onQuit}
          style={{ marginTop: 10, width: '100%', height: 44, borderRadius: 10, border: 'none', background: 'rgba(229,57,53,0.08)', color: '#E53935', fontSize: 14, fontWeight: 600, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}
        >
          <LogOut className="w-4 h-4" />{t('group.quit')}
        </button>
      </div>
    </div>
  );
}
