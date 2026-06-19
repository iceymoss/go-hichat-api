'use client';

import React, { useState, useRef, useEffect, useMemo, useCallback } from 'react';
import { formatTime, currentUser as mockCurrentUser, contacts, type Message, type Contact, type Conversation } from '@/lib/mock-data';

// 本人撤回的时间窗（秒），需与后端 im-rpc 配置 RecallWindowSeconds 保持一致（0=不限）。
// 仅用于普通用户撤回自己消息时的前端预校验；管理员/群主撤回不受此限。
const RECALL_WINDOW_SECONDS = 120;

/** 气泡时间：今天 HH:mm，昨天 昨天 HH:mm，今年 MM/DD HH:mm，跨年 YYYY/MM/DD HH:mm */
function formatBubbleTime(date: Date, t: (k: string) => string): string {
  const now = new Date();
  const hm = `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  const md = `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`;
  const isToday = date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth() && date.getDate() === now.getDate();
  if (isToday) return hm;
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  const isYesterday = date.getFullYear() === yesterday.getFullYear() && date.getMonth() === yesterday.getMonth() && date.getDate() === yesterday.getDate();
  if (isYesterday) return `${t('chat.yesterday')} ${hm}`;
  if (date.getFullYear() !== now.getFullYear()) return `${date.getFullYear()}/${md} ${hm}`;
  return `${md} ${hm}`;
}
import { useIMStore } from '@/lib/im-store';
import { useChatStore } from '@/lib/chat-store';
import { useSettingsStore } from '@/lib/settings-store';
import { getAvatarColor } from '@/lib/utils';
import { toast } from 'sonner';
import {
  ArrowLeft,
  Phone,
  Video,
  MoreVertical,
  Send,
  Smile,
  Paperclip,
  Mic,
  CheckCheck,
  MessageCircle,
  X,
  FileText,
  Download,
  Play,
} from 'lucide-react';
import { imUpload } from '@/lib/api-client';
import { buildMediaContent, parseMediaContent, mediaPreview } from '@/lib/media-message';
import ChatEmojiPanel, { type StickerItem } from './ChatEmojiPanel';
import MediaLightbox, { type LightboxItem } from './MediaLightbox';
import VoiceBubble from './VoiceBubble';
import { useIsMobile } from '@/hooks/use-mobile';
import { useT } from '@/hooks/use-i18n';
import { CallDialog } from './CallDialog';
import ChatSettingsMenu from './ChatSettingsMenu';
import MessageContextMenu from './MessageContextMenu';
import ConfirmDialog from './ConfirmDialog';
import ForwardDialog from './ForwardDialog';
import FloatingProfileCard from './FloatingProfileCard';
import GroupInfoCard from './GroupInfoCard';
import ReadStatusDialog from './ReadStatusDialog';

/** 稳定的空数组引用，避免 zustand selector 每次返回新数组导致无限重渲染 */
const EMPTY_STR_ARR: string[] = [];

function formatBytes(n?: number): string {
  if (!n || n <= 0) return '';
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}

/** 按消息类型渲染气泡内容（文本 / 图片 / 视频 / 文件 / 语音 / 表情包） */
function MessageContent({ message, onOpenMedia, isOwn, voiceUnplayed, onVoicePlayed }: {
  message: Message;
  onOpenMedia?: (m: Message) => void;
  isOwn?: boolean;
  voiceUnplayed?: boolean;
  onVoicePlayed?: () => void;
}) {
  const { type, content } = message;
  const t = useT();

  if (type === 'image' || type === 'memes') {
    const meta = parseMediaContent(content);
    if (!meta) return <span>{content}</span>;
    return (
      <img
        src={meta.thumbUrl || meta.url}
        alt=""
        onClick={() => onOpenMedia?.(message)}
        style={{ maxWidth: type === 'memes' ? 120 : 220, maxHeight: 220, borderRadius: 8, cursor: 'pointer', display: 'block' }}
      />
    );
  }

  if (type === 'video') {
    const meta = parseMediaContent(content);
    if (!meta) return <span>{content}</span>;
    return (
      <div
        onClick={() => onOpenMedia?.(message)}
        style={{ position: 'relative', display: 'inline-block', cursor: 'pointer', lineHeight: 0 }}
      >
        {meta.coverUrl ? (
          <img src={meta.coverUrl} alt="" style={{ maxWidth: 240, maxHeight: 240, borderRadius: 8, display: 'block' }} />
        ) : (
          <video src={meta.url} preload="metadata" style={{ maxWidth: 240, maxHeight: 240, borderRadius: 8, display: 'block', pointerEvents: 'none' }} />
        )}
        <span style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <span style={{ width: 48, height: 48, borderRadius: '50%', background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Play size={24} color="#fff" style={{ marginLeft: 2 }} />
          </span>
        </span>
      </div>
    );
  }

  if (type === 'file') {
    const meta = parseMediaContent(content);
    if (!meta) return <span>{content}</span>;
    return (
      <a
        href={meta.url}
        target="_blank"
        rel="noreferrer"
        download={meta.name}
        style={{ display: 'flex', alignItems: 'center', gap: 10, textDecoration: 'none', color: 'inherit', minWidth: 180 }}
      >
        <FileText size={32} style={{ flexShrink: 0, color: '#1BB45B' }} />
        <span style={{ flex: 1, minWidth: 0 }}>
          <span style={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: 14 }}>
            {meta.name || t('chat.fileFallback')}
          </span>
          <span style={{ display: 'block', fontSize: 12, opacity: 0.6 }}>{formatBytes(meta.size)}</span>
        </span>
        <Download size={18} style={{ flexShrink: 0, opacity: 0.6 }} />
      </a>
    );
  }

  if (type === 'voice') {
    const meta = parseMediaContent(content);
    if (meta) return <VoiceBubble url={meta.url} duration={meta.duration} unplayed={!isOwn && voiceUnplayed} onPlayed={onVoicePlayed} />;
    return <span>{content}</span>;
  }

  return <span>{renderTextWithMentions(content, isOwn)}</span>;
}

/** 文本中的 @所有人 / @某人 高亮渲染。发送方气泡是蓝底，@ 用浅金色才看得清；接收方白底用蓝色。 */
function renderTextWithMentions(text: string, isOwn?: boolean) {
  if (!text || text.indexOf('@') < 0) return text;
  const color = isOwn ? '#FFD666' : '#1BB45B';
  const parts = text.split(/(@所有人|@\S+)/g);
  return parts.map((part, i) =>
    part.startsWith('@')
      ? <span key={i} style={{ color, fontWeight: 600 }}>{part}</span>
      : part,
  );
}

/** 引用块：渲染被引用消息（图片/视频显示缩略图，否则文字预览），可点击跳转 */
function QuoteBlock({ reply, onJump, recalled }: { reply: NonNullable<Message['replyTo']>; onJump?: () => void; recalled?: boolean }) {
  const t = useT();
  const hasThumb = !recalled && (reply.mType === 'image' || reply.mType === 'video' || reply.mType === 'memes') && !!reply.thumbUrl;
  return (
    <div
      onClick={onJump ? (e) => { e.stopPropagation(); onJump(); } : undefined}
      style={{
        display: 'flex', gap: 8, alignItems: 'center', padding: '4px 8px', marginBottom: 4,
        borderLeft: '3px solid #1BB45B', background: 'rgba(27,180,91,0.06)', borderRadius: '0 6px 6px 0',
        fontSize: 12, lineHeight: 1.4, cursor: onJump ? 'pointer' : 'default', maxWidth: 240,
      }}
    >
      {hasThumb && (
        <span style={{ position: 'relative', flexShrink: 0, lineHeight: 0 }}>
          <img src={reply.thumbUrl} alt="" style={{ width: 32, height: 32, borderRadius: 4, objectFit: 'cover', display: 'block' }} />
          {reply.mType === 'video' && (
            <Play size={14} color="#fff" style={{ position: 'absolute', inset: 0, margin: 'auto' }} />
          )}
        </span>
      )}
      <span style={{ minWidth: 0 }}>
        <span style={{ display: 'block', fontWeight: 600, color: '#576b95' }}>{reply.senderName}</span>
        <span style={{ display: 'block', opacity: 0.7, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontStyle: recalled ? 'italic' : 'normal' }}>
          {recalled ? t('chat.recalledShort') : reply.content}
        </span>
      </span>
    </div>
  );
}

export default function ChatDetail() {
  const { selectedConversationId, setSelectedConversationId, setShowChatDetail, openUserProfile, invalidateFriends, openGroupDetail } = useIMStore();
  const t = useT();
  const [input, setInput] = useState('');
  // Track sent messages per conversation so they persist when switching back
  const [sentMap, setSentMap] = useState<Record<string, Message[]>>({});
  // Header dialog states
  const [callDialogOpen, setCallDialogOpen] = useState(false);
  const [callType, setCallType] = useState<'voice' | 'video'>('voice');
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  // Local mutable conversation overrides (muted, pinned)
  const [convOverrides, setConvOverrides] = useState<Record<string, { muted?: boolean; pinned?: boolean }>>({});
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const [loadingMore, setLoadingMore] = useState(false);
  const [noMoreHistory, setNoMoreHistory] = useState(false);
  const isMobile = useIsMobile();

  // Context menu state
  const [contextMenu, setContextMenu] = useState<{
    message: Message;
    senderName: string;
    isOwn: boolean;
    position: { x: number; y: number };
  } | null>(null);

  // Reply/Quote state
  const [replyTo, setReplyTo] = useState<{ message: Message; senderName: string } | null>(null);

  // 群 @ 状态
  const [atPicker, setAtPicker] = useState(false);
  const [atQuery, setAtQuery] = useState('');
  // 已选 @ 成员（uid→展示名）；发送时按 input 里是否仍含 @名字 过滤
  const [mentions, setMentions] = useState<{ uid: string; name: string }[]>([]);
  const [atAll, setAtAll] = useState(false);

  // Forward dialog state
  const [forwardMsg, setForwardMsg] = useState<Message | null>(null);

  // 图片/视频全屏预览
  const [lightbox, setLightbox] = useState<{ items: LightboxItem[]; index: number } | null>(null);

  // Deleted messages tracking
  const [deletedIds, setDeletedIds] = useState<Set<string>>(new Set());

  // Floating profile card state
  const [showProfileCard, setShowProfileCard] = useState(false);

  // 群信息卡片 + 退出群聊二次确认
  const [showGroupCard, setShowGroupCard] = useState(false);
  const [quitConfirmOpen, setQuitConfirmOpen] = useState(false);

  // Recalled messages tracking
  const [recalledIds, setRecalledIds] = useState<Set<string>>(new Set());

  // 撤回二次确认弹窗：保存待撤回的 msgId（null 表示关闭）
  const [recallConfirmId, setRecallConfirmId] = useState<string | null>(null);
  // 撤回被拦截/失败时的居中提示（如超时、无权限），null 表示关闭
  const [recallBlockedReason, setRecallBlockedReason] = useState<string | null>(null);

  // 群聊已读详情弹窗（存的是 mongoID）
  const [readDetailMsgId, setReadDetailMsgId] = useState<string | null>(null);

  // 优先使用 chat-store 真实数据，fallback 到 mock 数据
  const chatConversations = useChatStore(s => s.conversations);
  const chatMessages = useChatStore(s => s.messagesMap);
  const fetchMessages = useChatStore(s => s.fetchMessages);
  const fetchNewer = useChatStore(s => s.fetchNewer);
  const backToLatest = useChatStore(s => s.backToLatest);
  const anchored = useChatStore(s => !!s.anchoredConvs[selectedConversationId || '']);
  const jumpToContext = useChatStore(s => s.jumpToContext);
  const atMeIds = useChatStore(s => s.atMeMap[selectedConversationId || ''] || EMPTY_STR_ARR);
  const disabledInfo = useChatStore(s => s.disabledConversations[selectedConversationId || '']);
  const fetchAtMe = useChatStore(s => s.fetchAtMe);
  const consumeAtMe = useChatStore(s => s.consumeAtMe);
  const clearAtMe = useChatStore(s => s.clearAtMe);
  const [loadingNewer, setLoadingNewer] = useState(false);
  const storeSendMessage = useChatStore(s => s.sendMessage);
  const storeRecallMessage = useChatStore(s => s.recallMessage);
  const clearUnread = useChatStore(s => s.clearUnread);
  const storeMarkRead = useChatStore(s => s.markRead);
  const storeGroupMembers = useChatStore(s => s.groupMembers);
  const fetchGroupMembers = useChatStore(s => s.fetchGroupMembers);
  const markGroupRemoved = useChatStore(s => s.markGroupRemoved);
  const ensureUserProfiles = useChatStore(s => s.ensureUserProfiles);
  const readReceiptEnabled = useSettingsStore(s => s.readReceiptEnabled);
  const { currentUser } = useIMStore();

  // 去重：记录已对哪些 msgId 发过 markRead，避免每次 re-render 重复投递
  const markedReadRef = useRef<Set<string>>(new Set());

  const conversation = chatConversations.find(c => c.id === selectedConversationId);

  // 群成员名称/头像映射 (从 chat-store 真实数据)
  const groupMemberNames = useMemo<Record<string, string>>(() => {
    if (!selectedConversationId || conversation?.type !== 'group') return {};
    const members = storeGroupMembers[selectedConversationId] || [];
    const map: Record<string, string> = {};
    for (const m of members) {
      map[m.user_id] = m.group_nickname || m.nickname || m.user_id;
    }
    return map;
  }, [selectedConversationId, conversation?.type, storeGroupMembers]);

  const groupMembersMap = useMemo<Record<string, any[]>>(() => {
    if (!selectedConversationId || conversation?.type !== 'group') return {};
    const members = storeGroupMembers[selectedConversationId] || [];
    return { [selectedConversationId]: members.map(m => ({
      id: m.user_id,
      name: m.group_nickname || m.nickname || m.user_id,
      avatar: m.user_avatar_url || '',
      roleLevel: m.role_level,
    })) };
  }, [selectedConversationId, conversation?.type, storeGroupMembers]);

  // 当前用户在本群是否为管理员/群主（决定能否 @所有人）
  const isGroupAdmin = useMemo(() => {
    if (!selectedConversationId || conversation?.type !== 'group' || !currentUser?.id) return false;
    const me = (storeGroupMembers[selectedConversationId] || []).find(m => m.user_id === currentUser.id);
    return (me?.role_level ?? 0) >= 1;
  }, [selectedConversationId, conversation?.type, storeGroupMembers, currentUser?.id]);

  // 刷新后从服务端成员名单重新推导"被移出群"：成员已加载且自己不在其中 → 持久化禁用 + 系统消息
  useEffect(() => {
    if (!selectedConversationId || conversation?.type !== 'group' || !currentUser?.id) return;
    const members = storeGroupMembers[selectedConversationId] || [];
    if (members.length === 0) return; // 尚未拉取，避免误判
    if (!members.some(m => m.user_id === currentUser.id)) {
      markGroupRemoved(selectedConversationId, 'group.member.removed');
    }
  }, [selectedConversationId, conversation?.type, storeGroupMembers, currentUser?.id, markGroupRemoved]);

  // @ 候选成员（排除自己，按 atQuery 过滤）
  const atCandidates = useMemo(() => {
    if (!selectedConversationId || conversation?.type !== 'group') return [];
    const members = (storeGroupMembers[selectedConversationId] || []).filter(m => m.user_id !== currentUser?.id);
    const list = members.map(m => ({ uid: m.user_id, name: m.group_nickname || m.nickname || m.user_id, avatar: m.user_avatar_url || '' }));
    const q = atQuery.trim().toLowerCase();
    return q ? list.filter(m => m.name.toLowerCase().includes(q)) : list;
  }, [selectedConversationId, conversation?.type, storeGroupMembers, currentUser?.id, atQuery]);

  // Merged conversation data with local overrides
  const conv = useMemo(() => {
    if (!conversation) return null;
    const override = convOverrides[conversation.id] || {};
    return { ...conversation, ...override };
  }, [conversation, convOverrides]);

  // Find the contact matching the conversation partner (by ID, not by name)
  const { friends } = useIMStore();
  const userProfiles = useChatStore(s => s.userProfiles);
  const contactMatch = useMemo<Contact | null>(() => {
    if (!conv || conv.type !== 'private' || !currentUser?.id) return null;
    // 从 conversationId 提取对方 userId
    const parts = conv.id.split('_');
    const peerId = parts.find(p => p !== currentUser.id);
    if (!peerId) return null;
    // 优先从 friends 列表查找
    // 优先按 friend_uid（对方真实 userId）匹配；friend_uid 是好友关系行号，
    // 与他人的 userId 可能撞号，故不能用 c.id === peerId 先匹配
    const friend = friends.find(c => c.friend_uid === peerId) || friends.find(c => c.id === peerId);
    if (friend) return friend;
    // fallback: 从 userProfiles 构造最小 Contact
    const profile = userProfiles[peerId];
    if (profile) {
      return { id: peerId, name: profile.nickname, avatar: profile.avatar, pinyin: '', letter: '' } as Contact;
    }
    // fallback: mock contacts (by name)
    return contacts.find(c => c.name === conv.name) || null;
  }, [conv, currentUser, friends, userProfiles]);

  // 对方显示名称和头像（优先 contactMatch，fallback conv）
  const peerName = contactMatch?.remark || contactMatch?.name || conv?.name || '';
  const peerAvatar = contactMatch?.avatar || conv?.avatar || '';

  // 加载聊天记录 & 清除未读
  const resetNoMoreHistory = useCallback(() => setNoMoreHistory(false), []);

  useEffect(() => {
    if (selectedConversationId && currentUser?.token) {
      resetNoMoreHistory();
      markedReadRef.current = new Set();
      // 先拉取"@我未读"列表：必须早于 markRead（markRead 走 Kafka 异步会清未读位），
      // 仅群聊有 @；拉到的 id 用于顶部"有人@我"横幅逐条跳转
      if (conversation?.type === 'group') {
        fetchAtMe(currentUser.token, selectedConversationId);
      }
      fetchMessages(currentUser.token, selectedConversationId);
      clearUnread(selectedConversationId);
      // 群聊自动加载群成员
      if (conversation?.type === 'group') {
        fetchGroupMembers(currentUser.token, selectedConversationId);
      } else {
        // 私聊预加载对方资料，保证标题/引用归属/回复名稳定可解析
        const peerId = selectedConversationId.split('_').find(p => p !== currentUser.id);
        if (peerId) ensureUserProfiles(currentUser.token, [peerId]);
      }
    }
  }, [selectedConversationId, currentUser?.token, currentUser?.id, resetNoMoreHistory, fetchMessages, clearUnread, conversation?.type, fetchGroupMembers, ensureUserProfiles, fetchAtMe]);

  // 进入会话/新消息到达时自动上报已读：对自己收到的消息 ID 批量调用 markRead。
  // - 只处理 MongoDB 真实 id（排除 local_/push_ 临时 id）
  // - 通过 markedReadRef 去重，避免重复投递
  // - readReceiptEnabled=false 时本地仍上报，便于服务端统计未读数；仅显示层根据开关决定是否渲染
  useEffect(() => {
    if (!selectedConversationId || !currentUser?.id) return;
    const msgs = chatMessages[selectedConversationId] || [];
    if (msgs.length === 0) return;
    const unreadIds: string[] = [];
    for (const m of msgs) {
      if (m.senderId === currentUser.id) continue;
      if (!m.id || m.id.startsWith('local_') || m.id.startsWith('push_')) continue;
      if (markedReadRef.current.has(m.id)) continue;
      markedReadRef.current.add(m.id);
      unreadIds.push(m.id);
    }
    if (unreadIds.length > 0) {
      storeMarkRead(currentUser.id, selectedConversationId, unreadIds);
    }
  }, [selectedConversationId, currentUser?.id, chatMessages, storeMarkRead]);

  // Derive the current message list: chat-store 真实消息 + 本地 sentMap (兼容 mock)
  const messages = useMemo<Message[]>(() => {
    if (!selectedConversationId) return [];
    const storeMsgs = chatMessages[selectedConversationId] || [];
    const extra = sentMap[selectedConversationId] || [];
    return [...storeMsgs, ...extra];
  }, [selectedConversationId, chatMessages, sentMap]);

  // Reliable scroll-to-bottom: retries until the container is fully rendered
  const scrollToBottom = useCallback((instant = false) => {
    const el = messagesContainerRef.current;
    if (!el) return;
    const doScroll = () => { el.scrollTop = el.scrollHeight; };
    if (instant) {
      // Retry a few times to catch async render
      doScroll();
      requestAnimationFrame(doScroll);
      setTimeout(doScroll, 100);
      setTimeout(doScroll, 300);
    } else {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, []);

  // 跳转到"@我"列表里最早的一条：在当前列表则直接滚动高亮，否则按 around 加载其上下文窗口；
  // 跳转后从列表里移除该条（横幅计数递减），列表清空后横幅自动消失。
  const jumpToAtMe = useCallback(async () => {
    if (!currentUser?.token || !selectedConversationId) return;
    const msgId = atMeIds[0];
    if (!msgId) return;
    const scrollTo = () => {
      const el = document.querySelector(`[data-msgid="${msgId}"]`);
      if (!el) return false;
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      el.classList.add('msg-flash');
      setTimeout(() => el.classList.remove('msg-flash'), 1600);
      return true;
    };
    if (scrollTo()) {
      consumeAtMe(selectedConversationId, msgId);
      return;
    }
    const ok = await jumpToContext(currentUser.token, selectedConversationId, msgId);
    if (!ok) {
      toast.error(t('chat.atMeNotFound'));
      consumeAtMe(selectedConversationId, msgId);
      return;
    }
    setTimeout(() => { if (!scrollTo()) setTimeout(scrollTo, 250); }, 80);
    consumeAtMe(selectedConversationId, msgId);
  }, [currentUser?.token, selectedConversationId, atMeIds, jumpToContext, consumeAtMe]);

  // Scroll to bottom when messages change (new message sent/loaded)
  // 用首/尾消息 id 区分「向上分页（头部变化、尾部不变）」与「新消息追加 / 切换会话」：
  // 分页加载更早消息时不能自动滚到底（否则会把用户从历史位置拽回底部）。
  const prevFirstId = useRef<string | null>(null);
  const prevLastId = useRef<string | null>(null);
  const prevScrollConvId = useRef<string | null>(null);
  useEffect(() => {
    if (messages.length === 0) return;
    const firstId = messages[0].id;
    const lastId = messages[messages.length - 1].id;
    const convChanged = prevScrollConvId.current !== selectedConversationId;
    // 浏览历史（跳转到非最新窗口）时不要自动滚到底，保持在目标位置
    if (anchored) {
      prevFirstId.current = firstId;
      prevLastId.current = lastId;
      prevScrollConvId.current = selectedConversationId;
      return;
    }
    // 向上分页：尾部消息不变、头部消息变了 → 是 prepend，保持当前阅读位置，不滚动
    const prepended = !convChanged && prevLastId.current === lastId && prevFirstId.current !== firstId;
    if (!prepended) {
      scrollToBottom(convChanged);
    }
    prevFirstId.current = firstId;
    prevLastId.current = lastId;
    prevScrollConvId.current = selectedConversationId;
  }, [messages, scrollToBottom, anchored, selectedConversationId]);

  // Also scroll when conversation switches (even if messages haven't changed yet)
  useEffect(() => {
    if (selectedConversationId) {
      scrollToBottom(true);
    }
  }, [selectedConversationId, scrollToBottom]);

  // 滚动到顶部时加载更早的聊天记录
  const handleMessagesScroll = useCallback(() => {
    const el = messagesContainerRef.current;
    if (!el) return;

    // 浏览历史态：滚到接近底部 → 增量加载更新的消息
    if (anchored && !loadingNewer) {
      const distToBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
      if (distToBottom < 80) {
        setLoadingNewer(true);
        // 新内容追加在下方，不影响已有内容位置，无需校正 scrollTop
        fetchNewer(currentUser!.token, selectedConversationId!)
          .finally(() => setLoadingNewer(false));
      }
    }

    if (loadingMore || noMoreHistory) return;
    if (el.scrollTop < 50) {
      const msgs = chatMessages[selectedConversationId!] || [];
      if (msgs.length === 0) return;
      const oldestId = msgs[0]?.id;
      if (!oldestId || oldestId.startsWith('local_') || oldestId.startsWith('push_')) return;
      setLoadingMore(true);
      const prevCount = msgs.length;
      const prevHeight = el.scrollHeight;
      const prevScrollTop = el.scrollTop;
      fetchMessages(currentUser!.token, selectedConversationId!, oldestId).then(() => {
        requestAnimationFrame(() => {
          const newMsgs = useChatStore.getState().messagesMap[selectedConversationId!] || [];
          // 没有加载到新消息 → 已到头
          if (newMsgs.length <= prevCount) {
            setNoMoreHistory(true);
          }
          el.scrollTop = el.scrollHeight - prevHeight + prevScrollTop;
          setLoadingMore(false);
        });
      }).catch(() => setLoadingMore(false));
    }
  }, [selectedConversationId, loadingMore, noMoreHistory, chatMessages, currentUser, fetchMessages, anchored, loadingNewer, fetchNewer]);

  const handleBack = () => {
    if (isMobile) {
      setShowChatDetail(false);
    } else {
      setSelectedConversationId(null);
    }
  };

  const handleSend = () => {
    if (!input.trim() || !selectedConversationId) return;
    // 会话已失效（被踢/退群/删好友）：拦截发送，服务端也会兜底拒绝
    if (disabledInfo) return;

    // 引用消息：把被引用消息打包成 quote JSON（含类型与缩略图）随消息发送
    let quote: string | undefined;
    if (replyTo) {
      const qm = replyTo.message;
      const meta = parseMediaContent(qm.content);
      let thumb: string | undefined;
      if (meta) {
        if (qm.type === 'image' || qm.type === 'memes') thumb = meta.thumbUrl || meta.url;
        else if (qm.type === 'video') thumb = meta.coverUrl;
      }
      quote = JSON.stringify({ id: qm.id, uid: qm.senderId, name: replyTo.senderName, preview: mediaPreview(qm.type, qm.content), mType: qm.type, thumb });
    }

    // 群 @：按 input 是否仍含 "@名字" 过滤已选成员，得到最终 atUsers / atAll
    const text = input.trim();
    let mentionPayload: { atUsers?: string[]; atAll?: boolean } | undefined;
    if (conversation?.type === 'group') {
      const liveUsers = mentions.filter(m => text.includes('@' + m.name)).map(m => m.uid);
      const liveAtAll = atAll && text.includes('@' + t('chat.everyone'));
      if (liveUsers.length > 0 || liveAtAll) {
        mentionPayload = { atUsers: liveUsers.length > 0 ? Array.from(new Set(liveUsers)) : undefined, atAll: liveAtAll };
      }
    }

    // 通过 chat-store 发送（走 WebSocket RigorAck）
    if (currentUser?.token && currentUser?.id) {
      storeSendMessage(currentUser.token, currentUser.id, selectedConversationId, input.trim(), 'text', quote, mentionPayload);
    } else {
      // Fallback: 本地 mock 发送
      const newMsg: Message = {
        id: `msg-${Date.now()}`,
        senderId: 'me',
        content: input.trim(),
        timestamp: new Date(),
        type: 'text',
        replyTo: replyTo ? { senderName: replyTo.senderName, content: mediaPreview(replyTo.message.type, replyTo.message.content) } : undefined,
      };
      setSentMap(prev => ({
        ...prev,
        [selectedConversationId]: [...(prev[selectedConversationId] || []), newMsg],
      }));
    }

    setInput('');
    setReplyTo(null);
    setMentions([]);
    setAtAll(false);
    setAtPicker(false);
  };

  // 输入变化：群聊里检测光标处的 "@查询串" 决定是否弹出成员候选
  const handleInputChange = (val: string) => {
    setInput(val);
    if (conversation?.type !== 'group') return;
    const at = val.lastIndexOf('@');
    if (at >= 0) {
      const after = val.slice(at + 1);
      if (!/\s/.test(after)) {
        setAtQuery(after);
        setAtPicker(true);
        if (currentUser?.token && selectedConversationId && (storeGroupMembers[selectedConversationId] || []).length === 0) {
          fetchGroupMembers(currentUser.token, selectedConversationId);
        }
        return;
      }
    }
    setAtPicker(false);
  };

  // 选中某成员：把光标处的 "@查询" 替换为 "@昵称 "
  const pickMention = (m: { uid: string; name: string }) => {
    const at = input.lastIndexOf('@');
    const base = at >= 0 ? input.slice(0, at) : input;
    setInput(base + '@' + m.name + ' ');
    setMentions(prev => (prev.some(x => x.uid === m.uid) ? prev : [...prev, m]));
    setAtPicker(false);
    setAtQuery('');
  };

  const pickAtAll = () => {
    const at = input.lastIndexOf('@');
    const base = at >= 0 ? input.slice(0, at) : input;
    setInput(base + '@' + t('chat.everyone') + ' ');
    setAtAll(true);
    setAtPicker(false);
    setAtQuery('');
  };

  // 富媒体文件选择 → 上传 → 发送
  const fileInputRef = useRef<HTMLInputElement>(null);

  const buildVideoCover = (file: File): Promise<{ coverUrl?: string; width?: number; height?: number; duration?: number }> =>
    new Promise((resolve) => {
      try {
        const video = document.createElement('video');
        video.preload = 'metadata';
        video.muted = true;
        video.src = URL.createObjectURL(file);
        video.onloadeddata = () => {
          const canvas = document.createElement('canvas');
          canvas.width = video.videoWidth;
          canvas.height = video.videoHeight;
          canvas.getContext('2d')?.drawImage(video, 0, 0);
          const coverUrl = canvas.toDataURL('image/jpeg', 0.7);
          resolve({ coverUrl, width: video.videoWidth, height: video.videoHeight, duration: Math.round(video.duration) });
          URL.revokeObjectURL(video.src);
        };
        video.onerror = () => resolve({});
      } catch {
        resolve({});
      }
    });

  const handleFilesSelected = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || !files.length || !selectedConversationId || !currentUser?.token || !currentUser?.id) return;
    const token = currentUser.token;
    const userId = currentUser.id;
    const convId = selectedConversationId;

    for (const file of Array.from(files)) {
      try {
        const up = await imUpload(token, file);
        const kind = up.fileType === 'image' ? 'image' : up.fileType === 'video' ? 'video' : 'file';
        let extra: Record<string, unknown> = {};
        if (kind === 'image') {
          const dim = await new Promise<{ width?: number; height?: number }>((resolve) => {
            const img = new Image();
            img.onload = () => resolve({ width: img.naturalWidth, height: img.naturalHeight });
            img.onerror = () => resolve({});
            img.src = up.url;
          });
          extra = dim;
        } else if (kind === 'video') {
          extra = await buildVideoCover(file);
        }
        const content = buildMediaContent({ url: up.url, name: up.name, size: up.size, ...extra });
        storeSendMessage(token, userId, convId, content, kind);
      } catch (err) {
        console.error('[ChatDetail] upload failed:', err);
        toast.error(t('chat.uploadFail').replace('{name}', file.name));
      }
    }
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  // 语音录制（Safari 用 m4a，其余用 webm/opus）
  const [recording, setRecording] = useState(false);
  const recorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const recordStartRef = useRef<number>(0);

  const pickAudioMime = (): { mime: string; ext: string } => {
    const cands = [
      { mime: 'audio/webm;codecs=opus', ext: 'webm' },
      { mime: 'audio/mp4', ext: 'm4a' },
      { mime: 'audio/webm', ext: 'webm' },
    ];
    for (const c of cands) {
      if (typeof MediaRecorder !== 'undefined' && MediaRecorder.isTypeSupported(c.mime)) return c;
    }
    return { mime: '', ext: 'webm' };
  };

  const startRecording = async () => {
    if (!selectedConversationId || !currentUser?.token || !currentUser?.id) return;
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const { mime, ext } = pickAudioMime();
      const rec = new MediaRecorder(stream, mime ? { mimeType: mime } : undefined);
      chunksRef.current = [];
      rec.ondataavailable = (ev) => { if (ev.data.size > 0) chunksRef.current.push(ev.data); };
      rec.onstop = async () => {
        stream.getTracks().forEach(t => t.stop());
        const duration = Math.max(1, Math.round((Date.now() - recordStartRef.current) / 1000));
        const blob = new Blob(chunksRef.current, { type: rec.mimeType || mime || 'audio/webm' });
        // 录音太短 / 空录音（仅有容器头）直接丢弃，避免发出无法播放的空语音
        if (blob.size < 1024) {
          toast.error(t('chat.voiceTooShort'));
          return;
        }
        const token = currentUser.token!;
        const userId = currentUser.id!;
        try {
          const up = await imUpload(token, blob, `voice_${Date.now()}.${ext}`);
          const content = buildMediaContent({ url: up.url, size: up.size, duration });
          storeSendMessage(token, userId, selectedConversationId, content, 'voice');
        } catch (err) {
          console.error('[ChatDetail] voice upload failed:', err);
          toast.error(t('chat.voiceSendFail'));
        }
      };
      recorderRef.current = rec;
      recordStartRef.current = Date.now();
      rec.start(100); // 100ms 时间片，确保增量收集音频数据（避免空录音）
      setRecording(true);
    } catch (err) {
      console.error('[ChatDetail] mic permission denied:', err);
      toast.error(t('chat.micDenied'));
    }
  };

  const stopRecording = () => {
    recorderRef.current?.stop();
    recorderRef.current = null;
    setRecording(false);
  };

  // 发送收藏表情（表情包，memes）
  const handleSendSticker = (s: StickerItem) => {
    if (!selectedConversationId || !currentUser?.token || !currentUser?.id) return;
    const content = buildMediaContent({ url: s.url, thumbUrl: s.thumbnail, width: s.width, height: s.height });
    storeSendMessage(currentUser.token, currentUser.id, selectedConversationId, content, 'memes');
  };

  // 把聊天里的图片/表情消息收藏到「我的表情」
  const handleSaveSticker = async (m: Message) => {
    if (!currentUser?.token) return;
    const meta = parseMediaContent(m.content);
    if (!meta?.url) return;
    try {
      const res = await fetch('/api/user/emojis', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${currentUser.token}` },
        body: JSON.stringify({
          url: meta.url,
          name: meta.name || '',
          thumbnail: meta.thumbUrl || '',
          width: meta.width || 0,
          height: meta.height || 0,
          size: meta.size || 0,
          fileType: 'image',
        }),
      });
      const d = await res.json();
      if (d.success) toast.success(t('chat.stickerAdded'));
      else toast.error(d.message || t('chat.addFail'));
    } catch {
      toast.error(t('chat.addFail'));
    }
    setContextMenu(null);
  };

  const handleOpenCall = (type: 'voice' | 'video') => {
    setCallType(type);
    setCallDialogOpen(true);
  };

  const handleMutedChange = (muted: boolean) => {
    if (!selectedConversationId) return;
    setConvOverrides(prev => ({
      ...prev,
      [selectedConversationId]: { ...(prev[selectedConversationId] || {}), muted },
    }));
    if (currentUser?.token) {
      useChatStore.getState().setConversationSettings(currentUser.token, selectedConversationId, { muted });
    }
  };

  const handlePinnedChange = (pinned: boolean) => {
    if (!selectedConversationId) return;
    setConvOverrides(prev => ({
      ...prev,
      [selectedConversationId]: { ...(prev[selectedConversationId] || {}), pinned },
    }));
    if (currentUser?.token) {
      useChatStore.getState().setConversationSettings(currentUser.token, selectedConversationId, { pinned });
    }
  };

  const handleClearChat = () => {
    if (!selectedConversationId) return;
    setSentMap(prev => ({
      ...prev,
      [selectedConversationId]: [],
    }));
  };

  const handleSearchHistory = () => {
    setSearchBarOpen(true);
  };

  // ── More-menu actions (private chat): view profile / block / report ──
  const peerId = contactMatch?.friend_uid || contactMatch?.id || '';

  const handleViewProfile = () => {
    if (peerId) openUserProfile(peerId);
    else toast(t('common.featureWip'));
  };

  const handleBlockPeer = () => {
    if (!currentUser?.token || !peerId) return;
    fetch('/api/social/friend/block', {
      method: 'POST',
      headers: { Authorization: `Bearer ${currentUser.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ friend_uid: peerId, block: true }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) { toast.success(t('contact.blockedOk')); invalidateFriends(); }
        else toast.error(d.message || t('group.opFailed'));
      })
      .catch(() => toast.error(t('group.opFailed')));
  };

  const handleReportPeer = () => {
    if (!currentUser?.token || !peerId) return;
    fetch('/api/social/friend/report', {
      method: 'POST',
      headers: { Authorization: `Bearer ${currentUser.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ friend_uid: peerId, reason: '' }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) toast.success(t('contact.reportOk'));
        else toast.error(d.message || t('group.reportFailed'));
      })
      .catch(() => toast.error(t('group.reportFailed')));
  };

  // ── 群聊：打开群信息卡片 / 退出群聊 ──
  const handleOpenGroupInfo = () => setShowGroupCard(true);

  // 退出群聊：群会话的 conversationId 即后端 group_id
  const doQuitGroup = () => {
    if (!currentUser?.token || !selectedConversationId) return;
    const gid = selectedConversationId;
    fetch('/api/social/group/quit', {
      method: 'POST',
      headers: { Authorization: `Bearer ${currentUser.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ group_id: gid }),
    })
      .then(r => r.json())
      .then(d => {
        if (d.success) {
          toast.success(t('group.quitToast'));
          setShowGroupCard(false);
          // 复用被移出群的禁用态：输入框禁用 + "你已不在该群"横幅
          markGroupRemoved(gid, 'group.member.removed');
        } else {
          toast.error(d.message || t('group.opFailed'));
        }
      })
      .catch(() => toast.error(t('group.networkError')));
  };

  // Close context menu when clicking elsewhere
  useEffect(() => {
    const handleClick = () => setContextMenu(null);
    const handleContextMenu = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest('.im-chat-bubble') || target.closest('.msg-context-menu-backdrop')) {
        // Let bubble handlers deal with it
      } else {
        setContextMenu(null);
      }
    };
    window.addEventListener('click', handleClick);
    window.addEventListener('contextmenu', handleContextMenu);
    return () => {
      window.removeEventListener('click', handleClick);
      window.removeEventListener('contextmenu', handleContextMenu);
    };
  }, []);

  // Context menu action handlers
  const handleCopyMessage = useCallback(async (msg: Message) => {
    setContextMenu(null);
    // 文本直接复制；图片尝试复制图片本身，失败则复制链接；其他富媒体复制链接
    if (msg.type === 'text' || msg.type === 'system') {
      await navigator.clipboard.writeText(msg.content);
      toast.success(t('chat.copied'));
      return;
    }
    const meta = parseMediaContent(msg.content);
    if (!meta?.url) { toast.error(t('chat.copyFail')); return; }
    if (msg.type === 'image' || msg.type === 'memes') {
      try {
        const blob = await (await fetch(meta.url)).blob();
        await navigator.clipboard.write([new ClipboardItem({ [blob.type || 'image/png']: blob })]);
        toast.success(t('chat.imgCopied'));
        return;
      } catch {
        // 跨域或不支持 → 退化为复制链接
      }
    }
    await navigator.clipboard.writeText(meta.url);
    toast.success(t('chat.linkCopied'));
  }, []);

  // 打开图片/视频全屏预览（同会话内的图片/视频可左右切换；表情包单独预览）
  const openMedia = useCallback((m: Message) => {
    if (m.type === 'memes') {
      const meta = parseMediaContent(m.content);
      if (meta?.url) setLightbox({ items: [{ type: 'image', url: meta.url }], index: 0 });
      return;
    }
    const gallery: { id: string; item: LightboxItem }[] = [];
    for (const mm of messages) {
      if (mm.type === 'image' || mm.type === 'video') {
        const meta = parseMediaContent(mm.content);
        if (meta?.url) gallery.push({ id: mm.id, item: { type: mm.type, url: meta.url } });
      }
    }
    if (!gallery.length) return;
    const idx = Math.max(0, gallery.findIndex(g => g.id === m.id));
    setLightbox({ items: gallery.map(g => g.item), index: idx });
  }, [messages]);

  const handleForwardMessage = useCallback((msg: Message) => {
    setForwardMsg(msg);
    setContextMenu(null);
  }, []);

  const handleReplyMessage = useCallback((msg: Message, senderName: string) => {
    setReplyTo({ message: msg, senderName });
    setContextMenu(null);
  }, []);

  // 点击「撤回」：先做本地前置校验，通过后弹二次确认框，确认才真正撤回
  const handleRecallMessage = useCallback((msgId: string) => {
    setContextMenu(null);
    if (!currentUser?.token || !selectedConversationId) return;
    // 本地/未落库的占位消息不能撤回（没有真实 MongoID）
    if (msgId.startsWith('local_') || msgId.startsWith('push_')) {
      toast.error(t('chat.notSentYet'));
      return;
    }
    // 本地时间窗预校验：仅拦"普通用户撤回自己消息超时"这一场景，省掉一次必然失败的请求。
    // 管理员/群主撤回（自己超时的消息或他人消息）不受时间窗限制，交由后端放行，这里不拦。
    const target = messages.find(m => m.id === msgId);
    const isSelf = !!target && (target.senderId === currentUser.id || target.senderId === 'me');
    const adminBypass = conversation?.type === 'group' && isGroupAdmin;
    if (target && isSelf && !adminBypass && RECALL_WINDOW_SECONDS > 0) {
      const elapsedMs = Date.now() - target.timestamp.getTime();
      if (elapsedMs > RECALL_WINDOW_SECONDS * 1000) {
        // 超时不发请求，直接弹居中提示框告知用户原因
        setRecallBlockedReason(t('chat.recallTimeout').replace('{sec}', String(RECALL_WINDOW_SECONDS)));
        return;
      }
    }
    // 弹出二次确认框，等用户确认
    setRecallConfirmId(msgId);
  }, [currentUser?.token, currentUser?.id, selectedConversationId, messages, conversation?.type, isGroupAdmin]);

  // 二次确认后真正执行撤回（乐观置态，失败回滚）
  const doRecall = useCallback((msgId: string) => {
    if (!currentUser?.token || !selectedConversationId) return;
    setRecalledIds(prev => new Set(prev).add(msgId));
    storeRecallMessage(currentUser.token, selectedConversationId, msgId)
      .then(() => toast.success(t('chat.recalledOk')))
      .catch(() => {
        setRecalledIds(prev => { const n = new Set(prev); n.delete(msgId); return n; });
        // 失败原因（多为超时或无权限）同样用居中提示框，避免 toast 一闪而过用户没察觉
        setRecallBlockedReason(t('chat.recallFail'));
      });
  }, [currentUser?.token, selectedConversationId, storeRecallMessage]);

  const handleDeleteMessage = useCallback((msgId: string) => {
    setDeletedIds(prev => new Set(prev).add(msgId));
    setContextMenu(null);
    toast.success(t('chat.msgDeleted'));
  }, []);

  // Bubble context menu trigger
  const handleBubbleContext = useCallback((
    e: React.MouseEvent | React.TouchEvent,
    message: Message,
    senderName: string,
    isOwn: boolean,
  ) => {
    e.preventDefault();
    e.stopPropagation();
    let clientX: number;
    let clientY: number;
    if ('touches' in e) {
      clientX = e.changedTouches[0].clientX;
      clientY = e.changedTouches[0].clientY;
    } else {
      clientX = e.clientX;
      clientY = e.clientY;
    }
    setContextMenu({ message, senderName, isOwn, position: { x: clientX, y: clientY } });
  }, []);

  // Filter messages when search is active, exclude deleted
  const displayMessages = useMemo(() => {
    let result = messages;
    if (searchKeyword.trim()) {
      const kw = searchKeyword.toLowerCase();
      result = result.filter(m => m.content.toLowerCase().includes(kw));
    }
    return result.filter(m => !deletedIds.has(m.id));
  }, [messages, searchKeyword, deletedIds]);

  if (!conv) {
    return (
      <div
        className="h-full flex items-center justify-center"
        style={{ backgroundColor: '#F5F7FA' }}
      >
        <div className="text-center" style={{ color: '#A2ACB5' }}>
          <MessageCircle
            style={{ width: 64, height: 64, margin: '0 auto 16px', opacity: 0.3 }}
            strokeWidth={1.5}
          />
          <p style={{ fontSize: 14, fontWeight: 500 }}>{t('chat.empty')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* ── Chat Header: clean white ── */}
      <header
        className="flex items-center justify-between shrink-0"
        style={{
          height: 56,
          paddingLeft: 12,
          paddingRight: 8,
          background: '#FFFFFF',
          borderBottom: '1px solid rgba(0,0,0,0.08)',
          borderLeft: 'none',
          borderRight: 'none',
          borderTop: 'none',
        }}
      >
        <div className="flex items-center gap-3 min-w-0">
          <button
            onClick={handleBack}
            className="shrink-0 flex items-center justify-center"
            style={{
              width: 36, height: 36, borderRadius: 10,
              border: 'none', background: 'transparent',
              color: '#1BB45B', cursor: 'pointer',
            }}
          >
            <ArrowLeft style={{ width: 20, height: 20 }} />
          </button>
          <div
            onClick={conv.type === 'group'
              ? () => setShowGroupCard(true)
              : (contactMatch ? () => setShowProfileCard(true) : undefined)}
            style={{
              width: 40, height: 40, borderRadius: '50%', flexShrink: 0,
              background: peerAvatar ? 'transparent' : getAvatarColor(peerName),
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#FFFFFF', fontSize: 16, fontWeight: 600,
              cursor: (conv.type === 'group' || contactMatch) ? 'pointer' : 'default',
              overflow: 'hidden',
            }}
          >
            {peerAvatar ? <img src={peerAvatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : peerName[0]}
          </div>
          <div className="min-w-0">
            <h2 className="truncate" style={{ fontSize: 15, fontWeight: 600, color: '#1C2733', lineHeight: 1.3 }}>
              {peerName}
            </h2>
            {conv.type === 'group' ? (
              <p className="truncate" style={{ fontSize: 12, color: '#708499', lineHeight: 1.3, marginTop: 1 }}>
                {(() => {
                  const n = (storeGroupMembers[selectedConversationId!] || []).length || conv.members || 0;
                  return n ? t('chat.memberCount').replace('{count}', String(n)) : '';
                })()}
              </p>
            ) : conv.online ? (
              <p style={{ fontSize: 12, color: '#4DCD5E', lineHeight: 1.3, marginTop: 1 }}>{t('chat.online')}</p>
            ) : (
              <p style={{ fontSize: 12, color: '#A2ACB5', lineHeight: 1.3, marginTop: 1 }}>{t('chat.offline')}</p>
            )}
          </div>
        </div>
        <div className="flex items-center shrink-0">
          <button
            onClick={() => handleOpenCall('voice')}
            className="hc-header-btn"
            style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#1BB45B'; (e.currentTarget as HTMLButtonElement).style.background = 'rgba(27,180,91,0.08)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#708499'; (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; }}
          >
            <Phone style={{ width: 20, height: 20 }} />
          </button>
          <button
            onClick={() => handleOpenCall('video')}
            className="hc-header-btn"
            style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#1BB45B'; (e.currentTarget as HTMLButtonElement).style.background = 'rgba(27,180,91,0.08)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#708499'; (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; }}
          >
            <Video style={{ width: 20, height: 20 }} />
          </button>
          <ChatSettingsMenu
            conversation={conv}
            onMutedChange={handleMutedChange}
            onPinnedChange={handlePinnedChange}
            onClearChat={handleClearChat}
            onSearchHistory={handleSearchHistory}
            onViewProfile={conv.type === 'private' ? handleViewProfile : undefined}
            onBlock={conv.type === 'private' ? handleBlockPeer : undefined}
            onReport={conv.type === 'private' ? handleReportPeer : undefined}
            isGroup={conv.type === 'group'}
            onGroupInfo={conv.type === 'group' ? handleOpenGroupInfo : undefined}
            onQuitGroup={conv.type === 'group' ? () => setQuitConfirmOpen(true) : undefined}
          >
            <button
              className="hc-header-btn"
              style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#1BB45B'; (e.currentTarget as HTMLButtonElement).style.background = 'rgba(27,180,91,0.08)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#708499'; (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; }}
            >
              <MoreVertical style={{ width: 20, height: 20 }} />
            </button>
          </ChatSettingsMenu>
        </div>
      </header>

      {/* ── Search Bar (toggle from settings menu) ── */}
      {searchBarOpen && (
        <div
          className="shrink-0"
          style={{ background: '#FFFFFF', borderBottom: '1px solid rgba(0,0,0,0.06)', padding: '8px 14px', display: 'flex', alignItems: 'center', gap: 8 }}
        >
          <input
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            placeholder={t('chat.searchPlaceholder')}
            autoFocus
            className="flex-1 outline-none"
            style={{ height: 36, padding: '0 14px', borderRadius: 18, border: '1px solid rgba(0,0,0,0.08)', background: '#F0F2F5', fontSize: 13, color: '#1C2733' }}
          />
          <button
            onClick={() => { setSearchBarOpen(false); setSearchKeyword(''); }}
            style={{ fontSize: 13, color: '#1BB45B', fontWeight: 500, background: 'transparent', border: 'none', cursor: 'pointer', whiteSpace: 'nowrap', padding: '4px 8px' }}
          >
            {t('common.cancel')}
          </button>
        </div>
      )}

      {/* ── Messages Area: light gray background ── */}
      <div style={{ position: 'relative', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      <div
        ref={messagesContainerRef}
        onScroll={handleMessagesScroll}
        className="flex-1 overflow-y-auto im-scroll"
        style={{
          padding: '8px 0',
          backgroundColor: '#F0F2F5',
        }}
      >
        <div style={{ display: 'flex', flexDirection: 'column', padding: '0 2%', minHeight: '100%' }}>
          {loadingMore && (
            <div style={{ textAlign: 'center', padding: '8px 0', color: '#A2ACB5', fontSize: 12 }}>
              {t('common.loading')}
            </div>
          )}
          {noMoreHistory && !loadingMore && (
            <div style={{ textAlign: 'center', padding: '8px 0', color: '#A2ACB5', fontSize: 12 }}>
              {t('chat.noMore')}
            </div>
          )}
          <MessageList
            messages={displayMessages}
            conversation={conv}
            recalledIds={recalledIds}
            onBubbleContext={handleBubbleContext}
            peerName={peerName}
            peerAvatar={peerAvatar}
            readReceiptEnabled={readReceiptEnabled}
            onShowReadDetail={(mid) => setReadDetailMsgId(mid)}
            onOpenMedia={openMedia}
          />
          {searchBarOpen && searchKeyword && displayMessages.length === 0 && (
            <div style={{ textAlign: 'center', padding: '24px 0', color: '#A2ACB5', fontSize: 13 }}>
              {t('chat.noMatchMsg')}
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>
        {atMeIds.length > 0 && (
          <button
            onClick={jumpToAtMe}
            title={t('chat.jumpAtMe')}
            style={{ position: 'absolute', right: 16, top: 16, padding: '8px 14px', borderRadius: 18, border: 'none', background: '#FA5151', color: '#fff', fontSize: 13, fontWeight: 600, cursor: 'pointer', boxShadow: '0 2px 8px rgba(250,81,81,0.35)', display: 'flex', alignItems: 'center', gap: 6, zIndex: 6 }}
          >
            <span>{atMeIds.length > 1 ? t('chat.atMeCount').replace('{count}', String(atMeIds.length)) : t('chat.atMeOne')}</span>
            <span style={{ fontSize: 12 }}>↓</span>
            <span
              role="button"
              tabIndex={0}
              onClick={(e) => { e.stopPropagation(); if (selectedConversationId) clearAtMe(selectedConversationId); }}
              style={{ display: 'inline-flex', alignItems: 'center', marginLeft: 2, opacity: 0.85 }}
            >
              <X size={13} />
            </span>
          </button>
        )}
        {anchored && (
          <button
            onClick={() => { if (currentUser?.token && selectedConversationId) backToLatest(currentUser.token, selectedConversationId); }}
            style={{ position: 'absolute', right: 16, bottom: 16, padding: '8px 14px', borderRadius: 18, border: 'none', background: '#1BB45B', color: '#fff', fontSize: 13, cursor: 'pointer', boxShadow: '0 2px 8px rgba(0,0,0,0.25)', display: 'flex', alignItems: 'center', gap: 4, zIndex: 5 }}
          >
            {t('chat.backToLatest')} ↓
          </button>
        )}
      </div>

      {/* ── Input Area: clean white, single row ── */}
      <div
        className="shrink-0"
        style={{
          background: '#FFFFFF',
          borderTop: '1px solid rgba(0,0,0,0.06)',
        }}
      >
        {/* 会话失效横幅：被踢/退群/解散/删好友后禁用输入 */}
        {disabledInfo && (
          <div style={{
            padding: '10px 5%', textAlign: 'center', fontSize: 13, color: '#A0291F',
            background: '#FDECEA', borderTop: '1px solid rgba(0,0,0,0.06)',
          }}>
            {disabledInfo.eventType === 'friend.deleted'
              ? t('chat.disabled.friend')
              : t('chat.disabled.group')}
          </div>
        )}
        {/* Reply indicator */}
        {replyTo && (
          <div style={{
            display: 'flex', alignItems: 'center', gap: 8, padding: '8px 5%',
            background: '#FFFFFF', borderTop: '1px solid rgba(0,0,0,0.06)',
          }}>
            <div style={{ width: 3, height: 32, borderRadius: 2, background: '#1BB45B', flexShrink: 0 }} />
            {(() => {
              const meta = parseMediaContent(replyTo.message.content);
              const thumb = meta
                ? (replyTo.message.type === 'video' ? meta.coverUrl
                  : (replyTo.message.type === 'image' || replyTo.message.type === 'memes') ? (meta.thumbUrl || meta.url)
                  : undefined)
                : undefined;
              return thumb ? <img src={thumb} alt="" style={{ width: 32, height: 32, borderRadius: 4, objectFit: 'cover', flexShrink: 0 }} /> : null;
            })()}
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 12, fontWeight: 600, color: '#576b95' }}>{replyTo.senderName}</div>
              <div style={{ fontSize: 13, color: '#708499', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {mediaPreview(replyTo.message.type, replyTo.message.content)}
              </div>
            </div>
            <button
              onClick={() => setReplyTo(null)}
              style={{
                width: 28, height: 28, borderRadius: '50%', border: 'none',
                background: 'transparent', color: '#A2ACB5', cursor: 'pointer',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                transition: 'all 0.15s', flexShrink: 0,
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.color = '#708499';
                (e.currentTarget as HTMLButtonElement).style.background = 'rgba(0,0,0,0.04)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.color = '#A2ACB5';
                (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
              }}
            >
              <X size={18} />
            </button>
          </div>
        )}

        {/* 群 @ 成员候选 */}
        {atPicker && conversation?.type === 'group' && (
          <div style={{
            maxHeight: 220, overflowY: 'auto', margin: '0 5%',
            background: '#FFFFFF', border: '1px solid rgba(0,0,0,0.08)',
            borderRadius: 8, boxShadow: '0 4px 16px rgba(0,0,0,0.12)',
          }}>
            {isGroupAdmin && (atQuery === '' || t('chat.everyone').toLowerCase().includes(atQuery.toLowerCase())) && (
              <div
                onClick={pickAtAll}
                style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 12px', cursor: 'pointer' }}
                onMouseEnter={(e) => ((e.currentTarget as HTMLDivElement).style.background = '#F0F2F5')}
                onMouseLeave={(e) => ((e.currentTarget as HTMLDivElement).style.background = 'transparent')}
              >
                <span style={{ width: 32, height: 32, borderRadius: '50%', background: '#1BB45B', color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16 }}>@</span>
                <span style={{ fontSize: 14, color: '#1C2733', fontWeight: 500 }}>{t('chat.everyone')}</span>
              </div>
            )}
            {atCandidates.length === 0 && !(isGroupAdmin && (atQuery === '' || t('chat.everyone').toLowerCase().includes(atQuery.toLowerCase()))) ? (
              <div style={{ padding: '12px', fontSize: 13, color: '#A2ACB5', textAlign: 'center' }}>{t('chat.noMatchMember')}</div>
            ) : atCandidates.map(m => (
              <div
                key={m.uid}
                onClick={() => pickMention(m)}
                style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 12px', cursor: 'pointer' }}
                onMouseEnter={(e) => ((e.currentTarget as HTMLDivElement).style.background = '#F0F2F5')}
                onMouseLeave={(e) => ((e.currentTarget as HTMLDivElement).style.background = 'transparent')}
              >
                {m.avatar
                  ? <img src={m.avatar} alt="" style={{ width: 32, height: 32, borderRadius: '50%', objectFit: 'cover' }} />
                  : <span style={{ width: 32, height: 32, borderRadius: '50%', background: getAvatarColor(m.name), color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 14 }}>{m.name.slice(0, 1)}</span>}
                <span style={{ fontSize: 14, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{m.name}</span>
              </div>
            ))}
          </div>
        )}

        <div className="flex items-center" style={{ gap: 8, padding: '10px 5%', minWidth: 0 }}>
          {/* Emoji */}
          <ChatEmojiPanel
            token={currentUser?.token}
            onPickEmoji={(native) => setInput((v) => v + native)}
            onPickSticker={handleSendSticker}
          >
            <button
              className="flex items-center justify-center shrink-0"
              style={{ width: 38, height: 38, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer' }}
            >
              <Smile style={{ width: 22, height: 22 }} />
            </button>
          </ChatEmojiPanel>

          {/* Input */}
          <input
            value={input}
            onChange={(e) => handleInputChange(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            disabled={!!disabledInfo}
            placeholder={disabledInfo
              ? (disabledInfo.eventType === 'friend.deleted' ? t('chat.disabled.friendShort') : t('chat.disabled.groupShort'))
              : (replyTo
                ? t('chat.replyTo').replace('{name}', conversation?.type === 'private' ? (peerName || replyTo.senderName) : replyTo.senderName)
                : t('chat.inputPlaceholder'))}
            type="text"
            className="flex-1 outline-none"
            style={{
              height: 40,
              padding: '0 16px',
              borderRadius: 20,
              border: '1px solid rgba(0,0,0,0.08)',
              background: disabledInfo ? '#E8EBED' : '#F0F2F5',
              fontSize: 14,
              color: '#1C2733',
              cursor: disabledInfo ? 'not-allowed' : 'text',
            }}
          />

          {/* Attach */}
          <input
            ref={fileInputRef}
            type="file"
            multiple
            accept="image/*,video/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar,.7z"
            style={{ display: 'none' }}
            onChange={handleFilesSelected}
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            title={t('chat.attachTitle')}
            className="flex items-center justify-center shrink-0"
            style={{ width: 38, height: 38, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer' }}
          >
            <Paperclip style={{ width: 20, height: 20 }} />
          </button>

          {/* Send / Mic */}
          {input.trim() ? (
            <button
              onClick={handleSend}
              className="flex items-center justify-center shrink-0"
              style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: '#1BB45B', color: '#FFFFFF', cursor: 'pointer' }}
            >
              <Send style={{ width: 20, height: 20, marginLeft: 1 }} />
            </button>
          ) : (
            <button
              onClick={recording ? stopRecording : startRecording}
              title={recording ? t('chat.recordStop') : t('chat.recordStart')}
              className="flex items-center justify-center shrink-0"
              style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: recording ? '#FA5151' : 'transparent', color: recording ? '#FFFFFF' : '#708499', cursor: 'pointer' }}
            >
              <Mic style={{ width: 22, height: 22 }} />
            </button>
          )}
        </div>
      </div>

      {/* ── Call Dialog ── */}
      {conv && (
        <CallDialog
          open={callDialogOpen}
          onOpenChange={setCallDialogOpen}
          type={callType}
          contactName={peerName}
          contactAvatar={conv.type === 'private' ? peerAvatar : undefined}
          isGroup={conv.type === 'group'}
          members={conv.type === 'group' ? (groupMembersMap[conv.id] || []) : []}
        />
      )}

      {/* ── Context Menu ── */}
      {contextMenu && (
        <MessageContextMenu
          message={contextMenu.message}
          senderName={contextMenu.senderName}
          isOwn={contextMenu.isOwn}
          canRecall={contextMenu.isOwn || (conversation?.type === 'group' && isGroupAdmin)}
          position={contextMenu.position}
          onClose={() => setContextMenu(null)}
          onCopy={handleCopyMessage}
          onForward={handleForwardMessage}
          onReply={handleReplyMessage}
          onRecall={handleRecallMessage}
          onDelete={handleDeleteMessage}
          onSaveSticker={handleSaveSticker}
        />
      )}

      {/* ── 撤回二次确认 ── */}
      <ConfirmDialog
        open={recallConfirmId !== null}
        onClose={() => setRecallConfirmId(null)}
        title={t('chat.recall.title')}
        description={t('chat.recall.desc')}
        confirmText={t('chat.recall.confirm')}
        confirmVariant="danger"
        onConfirm={() => { if (recallConfirmId) doRecall(recallConfirmId); }}
      />

      {/* ── 撤回被拦截/失败提示（单按钮） ── */}
      <ConfirmDialog
        open={recallBlockedReason !== null}
        onClose={() => setRecallBlockedReason(null)}
        title={t('chat.recall.blockedTitle')}
        description={recallBlockedReason || ''}
        confirmText={t('chat.recall.gotIt')}
        confirmVariant="default"
        hideCancel
        onConfirm={() => setRecallBlockedReason(null)}
      />

      {/* ── Floating Profile Card ── */}
      {showProfileCard && contactMatch && (
        <FloatingProfileCard
          contact={contactMatch}
          isStranger={false}
          onClose={() => setShowProfileCard(false)}
          onVoiceCall={() => handleOpenCall('voice')}
          onVideoCall={() => handleOpenCall('video')}
        />
      )}

      {/* ── Group Info Card ── */}
      {showGroupCard && conv.type === 'group' && (
        <GroupInfoCard
          groupId={selectedConversationId!}
          name={peerName}
          avatar={peerAvatar}
          memberCount={(storeGroupMembers[selectedConversationId!] || []).length || conv.members || 0}
          onClose={() => setShowGroupCard(false)}
          onOpenManage={() => { openGroupDetail(selectedConversationId!); setShowGroupCard(false); }}
          onQuit={() => setQuitConfirmOpen(true)}
        />
      )}

      {/* ── 退出群聊二次确认 ── */}
      <ConfirmDialog
        open={quitConfirmOpen}
        onClose={() => setQuitConfirmOpen(false)}
        title={t('group.confirmQuitTitle')}
        description={t('group.confirmQuitDesc').replace('{name}', peerName)}
        confirmText={t('group.quit')}
        confirmVariant="danger"
        onConfirm={doQuitGroup}
      />

      {/* ── Forward Dialog ── */}
      {forwardMsg && (
        <ForwardDialog
          message={forwardMsg}
          onClose={() => setForwardMsg(null)}
          onForward={(targetConvId) => {
            if (currentUser?.token && currentUser?.id) {
              storeSendMessage(currentUser.token, currentUser.id, targetConvId, forwardMsg.content, forwardMsg.type);
            }
            setForwardMsg(null);
            toast.success(t('chat.forwardOk'));
          }}
        />
      )}

      {/* ── 图片/视频全屏预览 ── */}
      {lightbox && (
        <MediaLightbox
          items={lightbox.items}
          index={lightbox.index}
          onClose={() => setLightbox(null)}
          onIndex={(i) => setLightbox(lb => (lb ? { ...lb, index: i } : lb))}
        />
      )}

      {/* ── Group Read Status Dialog ── */}
      {readDetailMsgId && (
        <ReadStatusDialog msgId={readDetailMsgId} onClose={() => setReadDetailMsgId(null)} />
      )}
    </div>
  );
}

/* ═══════════════════════════════════════════════
   Grouped message list with time separators
   ═══════════════════════════════════════════════ */

function MessageList({
  messages,
  conversation,
  recalledIds,
  onBubbleContext,
  peerName,
  peerAvatar,
  readReceiptEnabled,
  onShowReadDetail,
  onOpenMedia,
}: {
  messages: Message[];
  conversation: any;
  recalledIds: Set<string>;
  onBubbleContext: (e: React.MouseEvent | React.TouchEvent, message: Message, senderName: string, isOwn: boolean) => void;
  peerName: string;
  peerAvatar: string;
  readReceiptEnabled: boolean;
  onShowReadDetail: (msgId: string) => void;
  onOpenMedia: (m: Message) => void;
}) {
  const { currentUser } = useIMStore();
  const { selectedConversationId } = useIMStore();
  const { friends } = useIMStore();
  const t = useT();
  const userProfiles = useChatStore(s => s.userProfiles);
  const storeGroupMembers = useChatStore(s => s.groupMembers);
  const playedVoices = useChatStore(s => s.playedVoices);
  const markVoicePlayed = useChatStore(s => s.markVoicePlayed);
  const jumpToContext = useChatStore(s => s.jumpToContext);

  // 点击引用块跳转到原消息并高亮；若不在当前已加载列表，则按 msgId 加载其上下文窗口
  const jumpToMessage = async (msgId?: string) => {
    if (!msgId) return;
    const scrollTo = () => {
      const el = document.querySelector(`[data-msgid="${msgId}"]`);
      if (!el) return false;
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      el.classList.add('msg-flash');
      setTimeout(() => el.classList.remove('msg-flash'), 1600);
      return true;
    };
    if (scrollTo()) return;
    // 不在当前列表 → 加载目标上下文窗口（进入浏览历史态）
    if (!currentUser?.token || !selectedConversationId) return;
    const ok = await jumpToContext(currentUser.token, selectedConversationId, msgId);
    if (!ok) { toast.error(t('chat.origNotFound')); return; }
    // 等待新列表渲染后再滚动
    setTimeout(() => { if (!scrollTo()) setTimeout(scrollTo, 250); }, 80);
  };
  const longPressTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isLongPressRef = useRef(false);

  // 群成员名称映射
  const groupMemberNames = useMemo<Record<string, string>>(() => {
    if (!selectedConversationId || conversation?.type !== 'group') return {};
    const members = storeGroupMembers[selectedConversationId] || [];
    const map: Record<string, string> = {};
    for (const m of members) {
      map[m.user_id] = m.group_nickname || m.nickname || m.user_id;
    }
    return map;
  }, [selectedConversationId, conversation?.type, storeGroupMembers]);

  // 群成员头像映射（后端 user_avatar_url 是群内权威头像，优先于 userProfiles）
  const groupMemberAvatars = useMemo<Record<string, string>>(() => {
    if (!selectedConversationId || conversation?.type !== 'group') return {};
    const members = storeGroupMembers[selectedConversationId] || [];
    const map: Record<string, string> = {};
    for (const m of members) {
      if (m.user_avatar_url) map[m.user_id] = m.user_avatar_url;
    }
    return map;
  }, [selectedConversationId, conversation?.type, storeGroupMembers]);

  const GROUP_INTERVAL = 3 * 60 * 1000;

  type TimeGroup = { type: 'time'; time: string; key: string };
  type MsgGroup = { type: 'msg'; msgs: Message[]; key: string };
  // 撤回提示作为独立的居中系统行展示（类似微信），不进入气泡分组
  type SystemGroup = { type: 'system'; msg: Message; key: string };

  const groups: (TimeGroup | MsgGroup | SystemGroup)[] = [];
  let lastTime = 0;

  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i];
    const ts = msg.timestamp.getTime();

    // Time separator: first message or >5 min gap
    if (i === 0 || ts - lastTime > 5 * 60 * 1000) {
      groups.push({ type: 'time', time: formatTime(msg.timestamp), key: `t-${msg.id}` });
    }

    // 已撤回消息：单独作为居中系统行，既不并入上一组也不让下一条并入它
    const recalled = msg.recalled || recalledIds.has(msg.id);
    if (recalled) {
      groups.push({ type: 'system', msg, key: `s-${msg.id}` });
      lastTime = ts;
      continue;
    }

    // New message group: different sender or >3 min gap
    const last = groups[groups.length - 1];
    if (!last || last.type !== 'msg' ||
      last.msgs[last.msgs.length - 1].senderId !== msg.senderId ||
      ts - last.msgs[last.msgs.length - 1].timestamp.getTime() > GROUP_INTERVAL
    ) {
      groups.push({ type: 'msg', msgs: [msg], key: `g-${msg.id}` });
    } else {
      last.msgs.push(msg);
    }

    lastTime = ts;
  }

  // Long-press handlers for mobile
  const handleTouchStart = useCallback((
    e: React.TouchEvent,
    message: Message,
    senderName: string,
    isOwn: boolean,
  ) => {
    isLongPressRef.current = false;
    longPressTimerRef.current = setTimeout(() => {
      isLongPressRef.current = true;
      onBubbleContext(e, message, senderName, isOwn);
    }, 500);
  }, [onBubbleContext]);

  const handleTouchEnd = useCallback((e: React.TouchEvent) => {
    if (longPressTimerRef.current) {
      clearTimeout(longPressTimerRef.current);
      longPressTimerRef.current = null;
    }
    if (isLongPressRef.current) {
      e.preventDefault();
    }
  }, []);

  const handleTouchMove = useCallback(() => {
    if (longPressTimerRef.current) {
      clearTimeout(longPressTimerRef.current);
      longPressTimerRef.current = null;
    }
  }, []);

  // 判断消息是否是自己发的 (兼容 mock 的 'me' 和真实 userId)
  const isOwnMessage = useCallback((senderId: string) => {
    return senderId === 'me' || senderId === currentUser?.id;
  }, [currentUser?.id]);

  // Get sender name for a message
  const getSenderName = useCallback((senderId: string) => {
    if (isOwnMessage(senderId)) return currentUser?.name || mockCurrentUser.name;
    // 群聊：群昵称 > 用户昵称 > id（groupMemberNames 已按此优先级构建）
    if (conversation?.type === 'group') return groupMemberNames[senderId] || userProfiles[senderId]?.nickname || senderId;
    // 私聊：好友备注 > 用户昵称 > 会话名 > id
    const friend = friends.find(c => c.friend_uid === senderId) || friends.find(c => c.id === senderId);
    return friend?.remark || friend?.name || userProfiles[senderId]?.nickname || conversation?.name || senderId;
  }, [conversation, currentUser?.name, isOwnMessage, friends, userProfiles, groupMemberNames]);

  // 撤回提示文案：你 / 对方 / 某成员 / 管理员撤回了一条消息
  const recalledText = useCallback((m: Message) => {
    const by = m.recalledBy;
    if (by && by === currentUser?.id) return t('chat.recalled.you');
    if (by && by !== m.senderId) return t('chat.recalled.admin');
    if (isOwnMessage(m.senderId)) return t('chat.recalled.you');
    return conversation?.type === 'group'
      ? t('chat.recalled.member').replace('{name}', getSenderName(m.senderId))
      : t('chat.recalled.peer');
  }, [currentUser?.id, conversation?.type, isOwnMessage, getSenderName]);

  return (
    <>
      {groups.map((group) => {
        if (group.type === 'time') {
          return (
            <div key={group.key} style={{ textAlign: 'center', padding: '10px 0 12px' }}>
              <span style={{
                fontSize: 12,
                color: '#A2ACB5',
                background: 'rgba(0,0,0,0.04)',
                padding: '3px 10px',
                borderRadius: 12,
                display: 'inline-block',
              }}>
                {group.time}
              </span>
            </div>
          );
        }

        if (group.type === 'system') {
          return (
            <div key={group.key} data-msgid={group.msg.id} style={{ textAlign: 'center', padding: '4px 0' }}>
              <span style={{ fontSize: 12, color: '#A2ACB5', display: 'inline-block', maxWidth: '80%' }}>
                {recalledText(group.msg)}
              </span>
            </div>
          );
        }

        const { msgs } = group;
        const lastMsg = msgs[msgs.length - 1];
        const isSent = isOwnMessage(msgs[0].senderId);

        return (
          <div key={group.key} style={{
            display: 'flex',
            justifyContent: isSent ? 'flex-end' : 'flex-start',
            marginBottom: 14,
          }}>
            {/* Received avatar */}
            {!isSent && (() => {
              const senderName = conversation.type === 'group'
                ? (groupMemberNames[msgs[0].senderId] || msgs[0].senderId)
                : peerName;
              const senderAvatar = conversation.type === 'group'
                ? (groupMemberAvatars[msgs[0].senderId] || userProfiles[msgs[0].senderId]?.avatar || '')
                : peerAvatar;
              return (
                <div style={{ width: 36, flexShrink: 0, marginRight: 8 }}>
                  {senderAvatar ? (
                    <img src={senderAvatar} alt="" style={{ width: 36, height: 36, borderRadius: '50%', objectFit: 'cover' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }} />
                  ) : null}
                  <div style={{
                    width: 36, height: 36, borderRadius: '50%',
                    background: getAvatarColor(senderName),
                    display: senderAvatar ? 'none' : 'flex', alignItems: 'center', justifyContent: 'center',
                    color: '#FFFFFF', fontSize: 13, fontWeight: 600,
                  }}>
                    {senderName[0]}
                  </div>
                </div>
              );
            })()}

            {/* Bubbles */}
            <div style={{ flex: 1, minWidth: 0 }}>
              {/* Group sender name (only for first message in group) */}
              {conversation.type === 'group' && !isSent && (
                <div style={{ fontSize: 12, color: '#576b95', fontWeight: 600, marginBottom: 2, textAlign: 'left' }}>
                  {groupMemberNames[msgs[0].senderId] || msgs[0].senderId}
                </div>
              )}
              {msgs.map((m) => {
                const msgIsSent = isOwnMessage(m.senderId);
                const senderName = getSenderName(m.senderId);
                const mediaMeta = parseMediaContent(m.content);
                // 已撤回：服务端历史(m.recalled) 或 本地乐观态(recalledIds) 任一命中
                const isRecalled = m.recalled || recalledIds.has(m.id);
                // 图片/视频/文件/表情包：裸露展示，不套消息气泡
                const bareMedia = !isRecalled && !!mediaMeta &&
                  (m.type === 'image' || m.type === 'video' || m.type === 'file' || m.type === 'memes');

                return (
                  <div key={m.id} data-msgid={m.id} style={{ textAlign: msgIsSent ? 'right' : 'left', marginBottom: 3 }}>
                    <div
                      className={bareMedia ? `im-chat-media ${msgIsSent ? 'sent' : 'received'}` : `im-chat-bubble ${msgIsSent ? 'sent' : 'received'}`}
                      onContextMenu={(e) => onBubbleContext(e, m, senderName, msgIsSent)}
                      onTouchStart={(e) => handleTouchStart(e, m, senderName, msgIsSent)}
                      onTouchEnd={handleTouchEnd}
                      onTouchMove={handleTouchMove}
                      style={bareMedia
                        ? { display: 'inline-block', cursor: 'context-menu', userSelect: 'none', textAlign: 'left', background: 'transparent', padding: 0, boxShadow: 'none' }
                        : { cursor: 'context-menu', userSelect: 'none', textAlign: 'left' }}
                    >
                    {isRecalled ? (
                      <span style={{ fontStyle: 'italic', opacity: 0.6 }}>
                        {recalledText(m)}
                      </span>
                    ) : (
                      <>
                        {/* Reply quote preview */}
                        {m.replyTo && (() => {
                          // 用固定的 senderId 实时解析当前昵称/备注（名字会变，id 不变）。
                          // 旧引用没存 senderId 时，回退用 msgId 在已加载列表里找原消息取其 senderId。
                          const sid = m.replyTo.senderId
                            || (m.replyTo.msgId ? messages.find(x => x.id === m.replyTo!.msgId)?.senderId : undefined);
                          const live = sid ? getSenderName(sid) : '';
                          const name = (live && live !== sid) ? live : (m.replyTo.senderName || sid || '');
                          // 被引用的原消息若已撤回，引用预览降级为"消息已撤回"
                          const orig = m.replyTo.msgId ? messages.find(x => x.id === m.replyTo!.msgId) : undefined;
                          const quotedRecalled = !!(orig && (orig.recalled || recalledIds.has(orig.id)));
                          return <QuoteBlock reply={{ ...m.replyTo, senderName: name }} recalled={quotedRecalled} onJump={() => jumpToMessage(m.replyTo!.msgId)} />;
                        })()}
                        <MessageContent
                          message={m}
                          onOpenMedia={onOpenMedia}
                          isOwn={msgIsSent}
                          voiceUnplayed={m.type === 'voice' && !playedVoices[m.id]}
                          onVoicePlayed={() => markVoicePlayed(m.id)}
                        />
                      </>
                    )}
                  </div>
                  </div>
                );
              })}

              {/* Timestamp + Status */}
              {(() => {
                // 已读回执展示规则：只在自己发送的消息、且未发送失败/发送中时启用
                // - readReceiptEnabled=false 时退化到「灰色 ✓✓ + 时间」行为（旧视觉）
                // - 私聊：isRead=true → 蓝色 ✓✓ + 「已读 · 时间」
                // - 群聊：readCount>0 → 蓝色 ✓✓ + 「X/N 人已读」
                const isGroup = conversation?.type === 'group';
                // 群总人数优先用 storeGroupMembers，缺失时用 conversation.members（会话列表里的 memberCount）
                // 避免群成员异步加载期间 totalMembers=0 导致 footer 不展示
                const liveMembersLen = (storeGroupMembers[selectedConversationId!] || []).length;
                const fallbackTotal = conversation?.members || 0;
                const totalMembers = isGroup
                  ? Math.max(0, (liveMembersLen || fallbackTotal) - 1)
                  : 1;
                const showRead = isSent
                  && lastMsg.status !== 'failed' && lastMsg.status !== 'sending'
                  && readReceiptEnabled
                  && (isGroup ? (lastMsg.readCount || 0) > 0 : !!lastMsg.isRead);

                // 未读用中性灰，已读用品牌蓝，两者肉眼可辨
                const checkColor = showRead ? '#1BB45B' : '#C8CCD0';
                let footerText = formatBubbleTime(lastMsg.timestamp, t);
                if (lastMsg.status === 'failed') {
                  footerText = t('chat.sendFail');
                } else if (isSent && isGroup && readReceiptEnabled && totalMembers > 0) {
                  // 群聊自己发的消息：总是展示 X/N，哪怕 X=0，方便点开看谁没读
                  const rc = Math.min(lastMsg.readCount || 0, totalMembers);
                  footerText = t('chat.readCount')
                    .replace('{read}', String(rc))
                    .replace('{total}', String(totalMembers))
                    .replace('{time}', formatBubbleTime(lastMsg.timestamp, t));
                } else if (showRead) {
                  footerText = t('chat.readAt').replace('{time}', formatBubbleTime(lastMsg.timestamp, t));
                }

                // 仅自己发出 + 群聊 + 有真实 mongoID 才能点开已读详情
                const canShowDetail = isSent && isGroup
                  && !!lastMsg.id
                  && !lastMsg.id.startsWith('local_')
                  && !lastMsg.id.startsWith('push_');

                const footerSpanProps = canShowDetail ? {
                  onClick: () => onShowReadDetail(lastMsg.id),
                  style: {
                    fontSize: 11,
                    color: lastMsg.status === 'failed' ? '#e74c3c' : (showRead ? '#1BB45B' : '#A2ACB5'),
                    lineHeight: 1,
                    cursor: 'pointer',
                    textDecoration: 'underline dotted',
                    textUnderlineOffset: 2,
                  } as React.CSSProperties,
                } : {
                  style: {
                    fontSize: 11,
                    color: lastMsg.status === 'failed' ? '#e74c3c' : (showRead ? '#1BB45B' : '#A2ACB5'),
                    lineHeight: 1,
                  } as React.CSSProperties,
                };

                return (
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                    marginTop: 4,
                    flexDirection: isSent ? 'row-reverse' : 'row',
                  }}>
                    {isSent && lastMsg.status === 'failed' && (
                      <span
                        title={t('chat.sendFailRetry')}
                        onClick={() => {
                          if (currentUser?.token && currentUser?.id && selectedConversationId) {
                            useChatStore.getState().resendMessage(currentUser.token, currentUser.id, selectedConversationId, lastMsg.id);
                          }
                        }}
                        style={{ cursor: 'pointer', color: '#e74c3c', fontSize: 14, lineHeight: 1 }}
                      >!</span>
                    )}
                    {isSent && lastMsg.status === 'sending' && (
                      <span style={{ fontSize: 11, color: '#A2ACB5' }}>...</span>
                    )}
                    {isSent && lastMsg.status !== 'failed' && lastMsg.status !== 'sending' && (
                      <CheckCheck style={{ width: 14, height: 14, color: checkColor }} />
                    )}
                    <span {...footerSpanProps} title={canShowDetail ? t('chat.readDetail') : undefined}>
                      {footerText}
                    </span>
                  </div>
                );
              })()}
            </div>

            {/* Sent avatar */}
            {isSent && (() => {
              const myName = currentUser?.name || mockCurrentUser.name;
              const myAvatar = currentUser?.avatar || '';
              return (
                <div style={{ width: 36, flexShrink: 0, marginLeft: 8 }}>
                  {myAvatar ? (
                    <img src={myAvatar} alt="" style={{ width: 36, height: 36, borderRadius: '50%', objectFit: 'cover' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; const next = (e.target as HTMLImageElement).nextElementSibling as HTMLElement; if (next) next.style.display = 'flex'; }} />
                  ) : null}
                  <div style={{
                    width: 36, height: 36, borderRadius: '50%',
                    background: getAvatarColor(myName),
                    display: myAvatar ? 'none' : 'flex', alignItems: 'center', justifyContent: 'center',
                    color: '#FFFFFF', fontSize: 13, fontWeight: 600,
                  }}>
                    {myName[0]}
                  </div>
                </div>
              );
            })()}
          </div>
        );
      })}
    </>
  );
}
