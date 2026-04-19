'use client';

import React, { useEffect, useState } from 'react';
import { X, CheckCheck, Clock } from 'lucide-react';
import { useIMStore } from '@/lib/im-store';
import { getChatLogReadRecords, type ReadRecordUser } from '@/lib/api-client';
import { getAvatarColor } from '@/lib/utils';

/** 群聊消息已读详情：展示谁读了 / 谁没读 */
export default function ReadStatusDialog({
  msgId,
  onClose,
}: {
  msgId: string;
  onClose: () => void;
}) {
  const token = useIMStore(s => s.currentUser?.token) || '';
  const [tab, setTab] = useState<'read' | 'unread'>('read');
  const [reads, setReads] = useState<ReadRecordUser[]>([]);
  const [unreads, setUnreads] = useState<ReadRecordUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!msgId || !token) return;
    let cancelled = false;
    setLoading(true); setError('');
    getChatLogReadRecords(token, msgId)
      .then(r => {
        if (cancelled) return;
        setReads(r?.reads || []);
        setUnreads(r?.unReads || []);
      })
      .catch(() => { if (!cancelled) setError('网络错误'); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [msgId, token]);

  useEffect(() => {
    const onEsc = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', onEsc);
    return () => window.removeEventListener('keydown', onEsc);
  }, [onClose]);

  const list = tab === 'read' ? reads : unreads;

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed', inset: 0, zIndex: 1100, background: 'rgba(0,0,0,0.45)',
        backdropFilter: 'blur(2px)', display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          width: 380, maxWidth: 'calc(100vw - 32px)', maxHeight: '70vh',
          background: '#fff', borderRadius: 16, boxShadow: '0 8px 32px rgba(0,0,0,0.18)',
          display: 'flex', flexDirection: 'column', overflow: 'hidden',
        }}
      >
        {/* Header */}
        <div style={{
          padding: '14px 18px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        }}>
          <span style={{ fontSize: 16, fontWeight: 600, color: '#1C2733' }}>消息已读情况</span>
          <button onClick={onClose} style={{
            width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent',
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#A2ACB5',
          }}><X size={16} /></button>
        </div>

        {/* Tabs */}
        <div style={{ display: 'flex', gap: 8, padding: '12px 18px 0' }}>
          <TabBtn active={tab === 'read'} onClick={() => setTab('read')} icon={<CheckCheck size={14} />}
            label={`已读 ${reads.length}`} activeColor="#3390EC" />
          <TabBtn active={tab === 'unread'} onClick={() => setTab('unread')} icon={<Clock size={14} />}
            label={`未读 ${unreads.length}`} activeColor="#A2ACB5" />
        </div>
        <div style={{ height: 1, background: 'rgba(0,0,0,0.06)', marginTop: 10 }} />

        {/* Body */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '6px 0' }}>
          {loading && (
            <div style={{ padding: 24, textAlign: 'center', color: '#A2ACB5', fontSize: 13 }}>加载中…</div>
          )}
          {!loading && error && (
            <div style={{ padding: 24, textAlign: 'center', color: '#E53935', fontSize: 13 }}>{error}</div>
          )}
          {!loading && !error && list.length === 0 && (
            <div style={{ padding: 24, textAlign: 'center', color: '#A2ACB5', fontSize: 13 }}>
              {tab === 'read' ? '还没有人已读' : '全员已读'}
            </div>
          )}
          {!loading && !error && list.map(u => (
            <div key={u.id} style={{
              display: 'flex', alignItems: 'center', gap: 12, padding: '10px 18px',
            }}>
              <div style={{
                width: 36, height: 36, borderRadius: '50%', flexShrink: 0, overflow: 'hidden',
                background: u.avatar ? 'transparent' : getAvatarColor(u.nickname || u.id),
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: '#fff', fontSize: 14, fontWeight: 600,
              }}>
                {u.avatar
                  ? <img src={u.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  : (u.nickname || u.id).slice(0, 1)}
              </div>
              <span style={{
                fontSize: 14, color: '#1C2733', flex: 1, minWidth: 0,
                overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
              }}>
                {u.nickname || u.id}
              </span>
              {tab === 'read' && u.readAt ? (
                <span style={{ fontSize: 12, color: '#A2ACB5', flexShrink: 0 }}>
                  {formatReadTime(u.readAt)}
                </span>
              ) : null}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/** 服务端 readAt 是 unix nano，转为可读的相对/绝对时间 */
function formatReadTime(readAt: number): string {
  // 过大（纳秒）→ 毫秒
  const ms = readAt > 1e15 ? Math.floor(readAt / 1e6) : readAt;
  const d = new Date(ms);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  if (diffMs < 60 * 1000) return '刚刚';
  if (diffMs < 60 * 60 * 1000) return `${Math.floor(diffMs / 60000)}分钟前`;
  const sameDay = d.toDateString() === now.toDateString();
  const hm = `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`;
  if (sameDay) return hm;
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  if (d.toDateString() === yesterday.toDateString()) return `昨天 ${hm}`;
  const md = `${(d.getMonth() + 1).toString().padStart(2, '0')}/${d.getDate().toString().padStart(2, '0')}`;
  if (d.getFullYear() !== now.getFullYear()) return `${d.getFullYear()}/${md} ${hm}`;
  return `${md} ${hm}`;
}

function TabBtn({ active, onClick, icon, label, activeColor }: {
  active: boolean; onClick: () => void; icon: React.ReactNode; label: string; activeColor: string;
}) {
  return (
    <button onClick={onClick} style={{
      display: 'flex', alignItems: 'center', gap: 4, padding: '6px 10px', borderRadius: 14,
      border: 'none', cursor: 'pointer', fontSize: 13, fontWeight: 500,
      background: active ? `${activeColor}14` : 'transparent',
      color: active ? activeColor : '#708499',
      transition: 'background 0.15s, color 0.15s',
    }}>
      {icon}{label}
    </button>
  );
}
