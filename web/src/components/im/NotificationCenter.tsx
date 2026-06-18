'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { Bell } from 'lucide-react';

import { useIMStore } from '@/lib/im-store';
import { useT } from '@/hooks/use-i18n';
import {
  listNotifications,
  getNotificationUnreadCount,
  markNotificationsRead,
  type NotificationItem,
} from '@/lib/api-client';
import { toast } from 'sonner';

/** 公共通知中心：铃铛 + 未读角标 + 下拉历史面板（读 listNotifications / markNotificationsRead）。
 *  实时性由 chat-store 的 ws.on('notify') 维护 notificationVersion，本组件订阅它刷新。 */
export default function NotificationCenter() {
  const t = useT();
  const currentUser = useIMStore(s => s.currentUser);
  const friends = useIMStore(s => s.friends);
  const notificationVersion = useIMStore(s => s.notificationVersion);
  const unread = useIMStore(s => s.notificationUnreadCount);
  const setUnread = useIMStore(s => s.setNotificationUnreadCount);
  const navigateToNotificationSource = useIMStore(s => s.navigateToNotificationSource);

  const [open, setOpen] = useState(false);
  const [items, setItems] = useState<NotificationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  const token = currentUser?.token;

  const fetchUnread = useCallback(() => {
    if (!token) return;
    getNotificationUnreadCount(token)
      .then(r => setUnread(r.count ?? 0))
      .catch(() => {});
  }, [token, setUnread]);

  const fetchList = useCallback(() => {
    if (!token) return;
    setLoading(true);
    listNotifications(token, false, 0, 30)
      .then(r => setItems(r.list ?? []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [token]);

  // 未读角标：挂载 + 每次收到新通知（notificationVersion 变化）刷新
  useEffect(() => { fetchUnread(); }, [fetchUnread, notificationVersion]);

  // 面板打开时拉取历史列表；打开期间来新通知也刷新
  useEffect(() => { if (open) fetchList(); }, [open, fetchList, notificationVersion]);

  // 点击外部关闭
  useEffect(() => {
    if (!open) return;
    const onDocClick = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onDocClick);
    return () => document.removeEventListener('mousedown', onDocClick);
  }, [open]);

  const actorName = (id?: string) => {
    if (!id) return '';
    const f = friends.find(x => x.id === id);
    return f?.name || id;
  };
  const actorAvatar = (id?: string) => friends.find(x => x.id === id)?.avatar || '';

  const handleMarkAllRead = () => {
    if (!token) return;
    markNotificationsRead(token, [])
      .then(() => {
        setItems(prev => prev.map(n => ({ ...n, isRead: 1 })));
        setUnread(0);
        toast.success(t('notify.center.allReadDone'));
      })
      .catch(() => toast.error(t('notify.center.markFailed')));
  };

  const handleItemClick = (n: NotificationItem) => {
    // 跳到对应来源 + 子 tab（好友→新的朋友/我收到·我发起；群→群申请/我收到·我发起）
    navigateToNotificationSource(n.notifyType);
    setOpen(false);
    if (!token || n.isRead) return;
    markNotificationsRead(token, [n.id])
      .then(() => {
        setItems(prev => prev.map(x => (x.id === n.id ? { ...x, isRead: 1 } : x)));
        setUnread(Math.max(0, unread - 1));
      })
      .catch(() => {});
  };

  return (
    <div ref={rootRef} className="relative">
      <button
        onClick={() => setOpen(o => !o)}
        className="flex items-center justify-center relative"
        title={t('notify.center.title')}
        style={{
          width: 36, height: 36, borderRadius: '50%',
          border: 'none', background: 'transparent',
          color: 'rgba(255,255,255,0.7)', cursor: 'pointer',
        }}
      >
        <Bell className="w-5 h-5" />
        {unread > 0 && (
          <span
            style={{
              position: 'absolute', top: 4, right: 4, minWidth: 16, height: 16,
              padding: '0 4px', borderRadius: 8, background: '#E53935', color: '#FFF',
              fontSize: 10, lineHeight: '16px', textAlign: 'center', fontWeight: 600,
            }}
          >
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div
          style={{
            position: 'absolute', top: 42, right: 0, width: 320, maxHeight: 420,
            background: '#FFF', borderRadius: 12, boxShadow: '0 8px 28px rgba(0,0,0,0.18)',
            zIndex: 10030, display: 'flex', flexDirection: 'column', overflow: 'hidden',
          }}
        >
          <div className="flex items-center justify-between shrink-0" style={{ height: 48, padding: '0 16px', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
            <span style={{ fontSize: 15, fontWeight: 600, color: '#1C2733' }}>{t('notify.center.title')}</span>
            <button onClick={handleMarkAllRead} style={{ fontSize: 13, color: '#1BB45B', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 500 }}>
              {t('notify.center.markAllRead')}
            </button>
          </div>

          <div className="flex-1 overflow-y-auto im-scroll">
            {items.length > 0 ? (
              items.map(n => (
                <div
                  key={n.id}
                  onClick={() => handleItemClick(n)}
                  style={{
                    display: 'flex', gap: 10, padding: '12px 16px',
                    borderBottom: '1px solid rgba(0,0,0,0.05)',
                    background: n.isRead ? '#FFF' : 'rgba(27,180,91,0.04)',
                    borderLeft: n.isRead ? '3px solid transparent' : '3px solid #1BB45B',
                    cursor: 'pointer',
                  }}
                >
                  <Avatar name={actorName(n.actorId)} avatar={actorAvatar(n.actorId)} />
                  <div className="flex-1 min-w-0">
                    <div style={{ fontSize: 13, color: '#1C2733', lineHeight: '1.4' }}>
                      <span style={{ fontWeight: 600 }}>{actorName(n.actorId)}</span>
                      <span style={{ color: '#708499', marginLeft: 4 }}>{t('notify.type.' + n.notifyType)}</span>
                    </div>
                    {n.content ? (
                      <div style={{ fontSize: 12, color: '#646A73', marginTop: 2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{n.content}</div>
                    ) : null}
                    <div style={{ fontSize: 11, color: '#A2ACB5', marginTop: 3 }}>{fmtTime(n.createTime)}</div>
                  </div>
                </div>
              ))
            ) : (
              <div className="flex flex-col items-center justify-center" style={{ padding: '48px 24px' }}>
                <Bell className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: 12 }} />
                <div style={{ fontSize: 13, color: '#A2ACB5' }}>{loading ? '...' : t('notify.center.empty')}</div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function Avatar({ name, avatar }: { name: string; avatar: string }) {
  if (avatar) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={avatar} alt={name} style={{ width: 36, height: 36, borderRadius: '50%', objectFit: 'cover', flexShrink: 0 }} />;
  }
  return (
    <div style={{ width: 36, height: 36, borderRadius: '50%', background: '#1BB45B', color: '#FFF', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 14, fontWeight: 600, flexShrink: 0 }}>
      {(name || '?').charAt(0).toUpperCase()}
    </div>
  );
}

/** unix 秒 -> 简单相对时间 */
function fmtTime(sec: number): string {
  if (!sec) return '';
  const ms = sec * 1000;
  const diff = Date.now() - ms;
  const min = Math.floor(diff / 60000);
  if (min < 1) return '刚刚';
  if (min < 60) return `${min}分钟前`;
  const h = Math.floor(min / 60);
  if (h < 24) return `${h}小时前`;
  const d = new Date(ms);
  const pad = (x: number) => String(x).padStart(2, '0');
  return `${d.getMonth() + 1}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
