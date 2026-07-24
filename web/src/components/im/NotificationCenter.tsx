'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { Bell } from 'lucide-react';

import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { useT } from '@/hooks/use-i18n';
import {
  listNotifications,
  markAllNotificationsRead,
  markNotificationIdsRead,
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
  const refreshNotificationUnread = useIMStore(s => s.refreshNotificationUnread);
  const bumpNotificationVersion = useIMStore(s => s.bumpNotificationVersion);
  const navigateToNotificationSource = useIMStore(s => s.navigateToNotificationSource);
  // 通知发起人头像/昵称：好友优先，非好友回退到 chat-store 拉取的用户资料
  const userProfiles = useChatStore(s => s.userProfiles);
  const ensureUserProfiles = useChatStore(s => s.ensureUserProfiles);

  const [open, setOpen] = useState(false);
  const [items, setItems] = useState<NotificationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [listError, setListError] = useState(false);
  const [markAllPending, setMarkAllPending] = useState(false);
  const [markRetryItem, setMarkRetryItem] = useState<NotificationItem | null>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const listGeneration = useRef(0);
  const clickGeneration = useRef(0);
  const markAllGeneration = useRef(0);

  const token = currentUser?.token;

  useEffect(() => {
    listGeneration.current += 1;
    clickGeneration.current += 1;
    markAllGeneration.current += 1;
    setItems([]);
    setLoaded(false);
    setListError(false);
    setMarkRetryItem(null);
    setMarkAllPending(false);
  }, [token]);

  const fetchUnread = useCallback(() => {
    if (!token) return;
    void refreshNotificationUnread();
  }, [token, refreshNotificationUnread]);

  const fetchList = useCallback(() => {
    if (!token) return;
    const generation = ++listGeneration.current;
    setLoading(true);
    setListError(false);
    listNotifications(token, false, 0, 30)
      .then(r => {
        if (generation === listGeneration.current && useIMStore.getState().currentUser?.token === token) {
          setItems(r.list ?? []);
          setLoaded(true);
        }
      })
      .catch(() => { if (generation === listGeneration.current) setListError(true); })
      .finally(() => { if (generation === listGeneration.current) setLoading(false); });
  }, [token]);

  // 未读角标：挂载 + 每次收到新通知（notificationVersion 变化）刷新
  useEffect(() => { fetchUnread(); }, [fetchUnread, notificationVersion]);

  // 面板打开时拉取历史列表；打开期间来新通知也刷新
  useEffect(() => { if (open) fetchList(); }, [open, fetchList, notificationVersion]);

  // 拉取非好友发起人的资料，保证头像/昵称能显示后端真实数据
  useEffect(() => {
    if (!token || items.length === 0) return;
    const ids = Array.from(new Set(
      items.map(n => n.actorId).filter((id): id is string =>
        !!id && !friends.some(f => f.id === id) && !userProfiles[id]),
    ));
    if (ids.length > 0) ensureUserProfiles(token, ids);
  }, [token, items, friends, userProfiles, ensureUserProfiles]);

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
    return f?.name || userProfiles[id]?.nickname || id;
  };
  const actorAvatar = (id?: string) => {
    if (!id) return '';
    return friends.find(x => x.id === id)?.avatar || userProfiles[id]?.avatar || '';
  };

  const handleMarkAllRead = () => {
    if (!token || markAllPending) return;
    listGeneration.current += 1;
    const generation = ++markAllGeneration.current;
    setMarkAllPending(true);
    markAllNotificationsRead(token)
      .then(() => {
        if (useIMStore.getState().currentUser?.token !== token) return;
        setItems(current => current.map(item => ({ ...item, isRead: 1 })));
        void refreshNotificationUnread();
        bumpNotificationVersion();
        toast.success(t('notify.center.allReadDone'));
      })
      .catch(() => {
        if (useIMStore.getState().currentUser?.token === token) toast.error(t('notify.center.markFailed'));
      })
      .finally(() => {
        if (generation === markAllGeneration.current && useIMStore.getState().currentUser?.token === token) setMarkAllPending(false);
      });
  };

  const navigate = (n: NotificationItem) => {
    navigateToNotificationSource(n.notifyType, n.bizId, n.groupId);
    setOpen(false);
  };

  const handleItemClick = (n: NotificationItem) => {
    const click = ++clickGeneration.current;
    if (n.isRead) { navigate(n); return; }
    if (!token) return;
    listGeneration.current += 1;
    setMarkRetryItem(null);
    markNotificationIdsRead(token, [n.id])
      .then(() => {
        if (useIMStore.getState().currentUser?.token !== token) return;
        setItems(current => current.map(item => item.id === n.id ? { ...item, isRead: 1 } : item));
        void refreshNotificationUnread();
        bumpNotificationVersion();
        void refreshNotificationUnread();
        if (click === clickGeneration.current) navigate(n);
      })
      .catch(() => {
        if (click !== clickGeneration.current || useIMStore.getState().currentUser?.token !== token) return;
        setMarkRetryItem(n);
        toast.error(t('notify.center.markFailed'));
      });
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
            <button disabled={markAllPending} onClick={handleMarkAllRead} style={{ fontSize: 13, color: '#1BB45B', background: 'none', border: 'none', cursor: markAllPending ? 'wait' : 'pointer', fontWeight: 500, opacity: markAllPending ? 0.6 : 1 }}>
              {markAllPending ? t('common.loading') : t('notify.center.markAllRead')}
            </button>
          </div>

          <div className="flex-1 overflow-y-auto im-scroll">
            {listError && (
              <div style={{ padding: '8px 12px', background: '#FFF3E8', color: '#AD6800', fontSize: 12 }}>
                {t('notify.center.listFailed')} <button onClick={fetchList} style={{ color: '#1BB45B', background: 'none', border: 0, cursor: 'pointer' }}>{t('common.retry')}</button>
              </div>
            )}
            {markRetryItem && (
              <div style={{ padding: '8px 12px', background: '#FFF3E8', color: '#AD6800', fontSize: 12 }}>
                {t('notify.center.markFailed')} <button onClick={() => handleItemClick(markRetryItem)} style={{ color: '#1BB45B', background: 'none', border: 0, cursor: 'pointer' }}>{t('common.retry')}</button>
              </div>
            )}
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
                    <div style={{ fontSize: 11, color: '#A2ACB5', marginTop: 3 }}>{fmtTime(n.createTime, t)}</div>
                  </div>
                </div>
              ))
            ) : !loaded && listError ? null : (
              <div className="flex flex-col items-center justify-center" style={{ padding: '48px 24px' }}>
                <Bell className="w-12 h-12" style={{ color: '#D1D5DB', marginBottom: 12 }} />
                <div style={{ fontSize: 13, color: '#A2ACB5' }}>{!loaded && loading ? t('common.loading') : t('notify.center.empty')}</div>
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
    return <img src={avatar} alt={name} style={{ width: 36, height: 36, borderRadius: '50%', objectFit: 'cover', flexShrink: 0 }} />;
  }
  return (
    <div style={{ width: 36, height: 36, borderRadius: '50%', background: '#1BB45B', color: '#FFF', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 14, fontWeight: 600, flexShrink: 0 }}>
      {(name || '?').charAt(0).toUpperCase()}
    </div>
  );
}

/** unix 秒 -> 简单相对时间 */
function fmtTime(sec: number, t: (k: string) => string): string {
  if (!sec) return '';
  const ms = sec * 1000;
  const diff = Date.now() - ms;
  const min = Math.floor(diff / 60000);
  if (min < 1) return t('group.time.justNow');
  if (min < 60) return t('group.time.minutesAgo').replace('{m}', String(min));
  const h = Math.floor(min / 60);
  if (h < 24) return t('group.time.hoursAgo').replace('{h}', String(h));
  const d = new Date(ms);
  const pad = (x: number) => String(x).padStart(2, '0');
  return `${d.getMonth() + 1}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
