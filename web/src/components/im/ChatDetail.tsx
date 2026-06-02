'use client';

import React, { useState, useRef, useEffect, useMemo, useCallback } from 'react';
import { formatTime, currentUser as mockCurrentUser, contacts, type Message, type Contact, type Conversation } from '@/lib/mock-data';

/** 气泡时间：今天 HH:mm，昨天 昨天 HH:mm，今年 MM/DD HH:mm，跨年 YYYY/MM/DD HH:mm */
function formatBubbleTime(date: Date): string {
  const now = new Date();
  const hm = `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  const md = `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`;
  const isToday = date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth() && date.getDate() === now.getDate();
  if (isToday) return hm;
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  const isYesterday = date.getFullYear() === yesterday.getFullYear() && date.getMonth() === yesterday.getMonth() && date.getDate() === yesterday.getDate();
  if (isYesterday) return `昨天 ${hm}`;
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
import { buildMediaContent, parseMediaContent } from '@/lib/media-message';
import ChatEmojiPanel, { type StickerItem } from './ChatEmojiPanel';
import { useIsMobile } from '@/hooks/use-mobile';
import { CallDialog } from './CallDialog';
import ChatSettingsMenu from './ChatSettingsMenu';
import MessageContextMenu from './MessageContextMenu';
import ForwardDialog from './ForwardDialog';
import FloatingProfileCard from './FloatingProfileCard';
import ReadStatusDialog from './ReadStatusDialog';

function formatBytes(n?: number): string {
  if (!n || n <= 0) return '';
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}

/** 按消息类型渲染气泡内容（文本 / 图片 / 视频 / 文件 / 语音 / 表情包） */
function MessageContent({ message }: { message: Message }) {
  const { type, content } = message;

  if (type === 'image' || type === 'memes') {
    const meta = parseMediaContent(content);
    if (!meta) return <span>{content}</span>;
    return (
      <img
        src={meta.thumbUrl || meta.url}
        alt=""
        onClick={() => window.open(meta.url, '_blank')}
        style={{ maxWidth: type === 'memes' ? 120 : 220, maxHeight: 220, borderRadius: 8, cursor: 'pointer', display: 'block' }}
      />
    );
  }

  if (type === 'video') {
    const meta = parseMediaContent(content);
    if (!meta) return <span>{content}</span>;
    return (
      <video
        src={meta.url}
        poster={meta.coverUrl}
        controls
        style={{ maxWidth: 240, maxHeight: 240, borderRadius: 8, display: 'block' }}
      >
        <Play size={16} />
      </video>
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
        <FileText size={32} style={{ flexShrink: 0, color: '#3390EC' }} />
        <span style={{ flex: 1, minWidth: 0 }}>
          <span style={{ display: 'block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: 14 }}>
            {meta.name || '文件'}
          </span>
          <span style={{ display: 'block', fontSize: 12, opacity: 0.6 }}>{formatBytes(meta.size)}</span>
        </span>
        <Download size={18} style={{ flexShrink: 0, opacity: 0.6 }} />
      </a>
    );
  }

  if (type === 'voice') {
    const meta = parseMediaContent(content);
    if (meta) return <audio src={meta.url} controls style={{ maxWidth: 240, height: 36 }} />;
    return <span>{content}</span>;
  }

  return <span>{content}</span>;
}

export default function ChatDetail() {
  const { selectedConversationId, setSelectedConversationId, setShowChatDetail } = useIMStore();
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

  // Forward dialog state
  const [forwardMsg, setForwardMsg] = useState<Message | null>(null);

  // Deleted messages tracking
  const [deletedIds, setDeletedIds] = useState<Set<string>>(new Set());

  // Floating profile card state
  const [showProfileCard, setShowProfileCard] = useState(false);

  // Recalled messages tracking
  const [recalledIds, setRecalledIds] = useState<Set<string>>(new Set());

  // 群聊已读详情弹窗（存的是 mongoID）
  const [readDetailMsgId, setReadDetailMsgId] = useState<string | null>(null);

  // 优先使用 chat-store 真实数据，fallback 到 mock 数据
  const chatConversations = useChatStore(s => s.conversations);
  const chatMessages = useChatStore(s => s.messagesMap);
  const fetchMessages = useChatStore(s => s.fetchMessages);
  const storeSendMessage = useChatStore(s => s.sendMessage);
  const clearUnread = useChatStore(s => s.clearUnread);
  const storeMarkRead = useChatStore(s => s.markRead);
  const storeGroupMembers = useChatStore(s => s.groupMembers);
  const fetchGroupMembers = useChatStore(s => s.fetchGroupMembers);
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
    const friend = friends.find(c => c.id === peerId || c.friend_uid === peerId);
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
      fetchMessages(currentUser.token, selectedConversationId);
      clearUnread(selectedConversationId);
      // 群聊自动加载群成员
      if (conversation?.type === 'group') {
        fetchGroupMembers(currentUser.token, selectedConversationId);
      }
    }
  }, [selectedConversationId, currentUser?.token, resetNoMoreHistory, fetchMessages, clearUnread, conversation?.type, fetchGroupMembers]);

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

  // Scroll to bottom when messages change (new message sent/loaded)
  const prevMsgCount = useRef(0);
  useEffect(() => {
    if (messages.length === 0) return;
    // If switching conversation (count changed drastically) → instant, otherwise smooth
    const isNewConv = Math.abs(messages.length - prevMsgCount.current) > 1;
    scrollToBottom(isNewConv);
    prevMsgCount.current = messages.length;
  }, [messages, scrollToBottom]);

  // Also scroll when conversation switches (even if messages haven't changed yet)
  useEffect(() => {
    if (selectedConversationId) {
      scrollToBottom(true);
    }
  }, [selectedConversationId, scrollToBottom]);

  // 滚动到顶部时加载更早的聊天记录
  const handleMessagesScroll = useCallback(() => {
    const el = messagesContainerRef.current;
    if (!el || loadingMore || noMoreHistory) return;
    if (el.scrollTop < 50) {
      const msgs = chatMessages[selectedConversationId!] || [];
      if (msgs.length === 0) return;
      const oldestId = msgs[0]?.id;
      if (!oldestId || oldestId.startsWith('local_') || oldestId.startsWith('push_')) return;
      setLoadingMore(true);
      const prevCount = msgs.length;
      const prevHeight = el.scrollHeight;
      fetchMessages(currentUser!.token, selectedConversationId!, oldestId).then(() => {
        requestAnimationFrame(() => {
          const newMsgs = useChatStore.getState().messagesMap[selectedConversationId!] || [];
          // 没有加载到新消息 → 已到头
          if (newMsgs.length <= prevCount) {
            setNoMoreHistory(true);
          }
          el.scrollTop = el.scrollHeight - prevHeight;
          setLoadingMore(false);
        });
      }).catch(() => setLoadingMore(false));
    }
  }, [selectedConversationId, loadingMore, noMoreHistory, chatMessages, currentUser, fetchMessages]);

  const handleBack = () => {
    if (isMobile) {
      setShowChatDetail(false);
    } else {
      setSelectedConversationId(null);
    }
  };

  const handleSend = () => {
    if (!input.trim() || !selectedConversationId) return;

    // 通过 chat-store 发送（走 WebSocket RigorAck）
    if (currentUser?.token && currentUser?.id) {
      storeSendMessage(currentUser.token, currentUser.id, selectedConversationId, input.trim());
    } else {
      // Fallback: 本地 mock 发送
      const newMsg: Message = {
        id: `msg-${Date.now()}`,
        senderId: 'me',
        content: input.trim(),
        timestamp: new Date(),
        type: 'text',
        replyTo: replyTo ? { senderName: replyTo.senderName, content: replyTo.message.content } : undefined,
      };
      setSentMap(prev => ({
        ...prev,
        [selectedConversationId]: [...(prev[selectedConversationId] || []), newMsg],
      }));
    }

    setInput('');
    setReplyTo(null);
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
        toast.error(`${file.name} 上传失败`);
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
        const token = currentUser.token!;
        const userId = currentUser.id!;
        try {
          const up = await imUpload(token, blob, `voice_${Date.now()}.${ext}`);
          const content = buildMediaContent({ url: up.url, size: up.size, duration });
          storeSendMessage(token, userId, selectedConversationId, content, 'voice');
        } catch (err) {
          console.error('[ChatDetail] voice upload failed:', err);
          toast.error('语音发送失败');
        }
      };
      recorderRef.current = rec;
      recordStartRef.current = Date.now();
      rec.start();
      setRecording(true);
    } catch (err) {
      console.error('[ChatDetail] mic permission denied:', err);
      toast.error('无法访问麦克风，请检查授权');
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
  const handleCopyMessage = useCallback((msg: Message) => {
    navigator.clipboard.writeText(msg.content);
    setContextMenu(null);
    toast.success('已复制');
  }, []);

  const handleForwardMessage = useCallback((msg: Message) => {
    setForwardMsg(msg);
    setContextMenu(null);
  }, []);

  const handleReplyMessage = useCallback((msg: Message, senderName: string) => {
    setReplyTo({ message: msg, senderName });
    setContextMenu(null);
  }, []);

  const handleRecallMessage = useCallback((msgId: string) => {
    setRecalledIds(prev => new Set(prev).add(msgId));
    setContextMenu(null);
    toast.success('消息已撤回');
  }, []);

  const handleDeleteMessage = useCallback((msgId: string) => {
    setDeletedIds(prev => new Set(prev).add(msgId));
    setContextMenu(null);
    toast.success('消息已删除');
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
          <p style={{ fontSize: 14, fontWeight: 500 }}>选择一个会话开始聊天</p>
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
              color: '#3390EC', cursor: 'pointer',
            }}
          >
            <ArrowLeft style={{ width: 20, height: 20 }} />
          </button>
          <div
            onClick={contactMatch ? () => {
              setShowProfileCard(true);
            } : undefined}
            style={{
              width: 40, height: 40, borderRadius: '50%', flexShrink: 0,
              background: peerAvatar ? 'transparent' : getAvatarColor(peerName),
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#FFFFFF', fontSize: 16, fontWeight: 600,
              cursor: contactMatch ? 'pointer' : 'default',
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
                {(storeGroupMembers[selectedConversationId!] || []).length || conv.members || ''}位成员
              </p>
            ) : conv.online ? (
              <p style={{ fontSize: 12, color: '#4DCD5E', lineHeight: 1.3, marginTop: 1 }}>在线</p>
            ) : (
              <p style={{ fontSize: 12, color: '#A2ACB5', lineHeight: 1.3, marginTop: 1 }}>离线</p>
            )}
          </div>
        </div>
        <div className="flex items-center shrink-0">
          <button
            onClick={() => handleOpenCall('voice')}
            className="hc-header-btn"
            style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#3390EC'; (e.currentTarget as HTMLButtonElement).style.background = 'rgba(51,144,236,0.08)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#708499'; (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; }}
          >
            <Phone style={{ width: 20, height: 20 }} />
          </button>
          <button
            onClick={() => handleOpenCall('video')}
            className="hc-header-btn"
            style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#3390EC'; (e.currentTarget as HTMLButtonElement).style.background = 'rgba(51,144,236,0.08)'; }}
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
          >
            <button
              className="hc-header-btn"
              style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'transparent', color: '#708499', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.color = '#3390EC'; (e.currentTarget as HTMLButtonElement).style.background = 'rgba(51,144,236,0.08)'; }}
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
            placeholder="搜索聊天记录..."
            autoFocus
            className="flex-1 outline-none"
            style={{ height: 36, padding: '0 14px', borderRadius: 18, border: '1px solid rgba(0,0,0,0.08)', background: '#F0F2F5', fontSize: 13, color: '#1C2733' }}
          />
          <button
            onClick={() => { setSearchBarOpen(false); setSearchKeyword(''); }}
            style={{ fontSize: 13, color: '#3390EC', fontWeight: 500, background: 'transparent', border: 'none', cursor: 'pointer', whiteSpace: 'nowrap', padding: '4px 8px' }}
          >
            取消
          </button>
        </div>
      )}

      {/* ── Messages Area: light gray background ── */}
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
              加载中...
            </div>
          )}
          {noMoreHistory && !loadingMore && (
            <div style={{ textAlign: 'center', padding: '8px 0', color: '#A2ACB5', fontSize: 12 }}>
              没有更多消息了
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
          />
          {searchBarOpen && searchKeyword && displayMessages.length === 0 && (
            <div style={{ textAlign: 'center', padding: '24px 0', color: '#A2ACB5', fontSize: 13 }}>
              没有找到相关消息
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>

      {/* ── Input Area: clean white, single row ── */}
      <div
        className="shrink-0"
        style={{
          background: '#FFFFFF',
          borderTop: '1px solid rgba(0,0,0,0.06)',
        }}
      >
        {/* Reply indicator */}
        {replyTo && (
          <div style={{
            display: 'flex', alignItems: 'center', gap: 8, padding: '8px 5%',
            background: '#FFFFFF', borderTop: '1px solid rgba(0,0,0,0.06)',
          }}>
            <div style={{ width: 3, height: 32, borderRadius: 2, background: '#3390EC', flexShrink: 0 }} />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 12, fontWeight: 600, color: '#3390EC' }}>{replyTo.senderName}</div>
              <div style={{ fontSize: 13, color: '#708499', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {replyTo.message.content}
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
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            placeholder={replyTo ? `回复 ${replyTo.senderName}...` : '输入消息...'}
            type="text"
            className="flex-1 outline-none"
            style={{
              height: 40,
              padding: '0 16px',
              borderRadius: 20,
              border: '1px solid rgba(0,0,0,0.08)',
              background: '#F0F2F5',
              fontSize: 14,
              color: '#1C2733',
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
            title="发送图片 / 视频 / 文件"
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
              style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: '#3390EC', color: '#FFFFFF', cursor: 'pointer' }}
            >
              <Send style={{ width: 20, height: 20, marginLeft: 1 }} />
            </button>
          ) : (
            <button
              onClick={recording ? stopRecording : startRecording}
              title={recording ? '点击停止并发送' : '按下录音'}
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
          position={contextMenu.position}
          onClose={() => setContextMenu(null)}
          onCopy={handleCopyMessage}
          onForward={handleForwardMessage}
          onReply={handleReplyMessage}
          onRecall={handleRecallMessage}
          onDelete={handleDeleteMessage}
        />
      )}

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
            toast.success('转发成功');
          }}
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
}: {
  messages: Message[];
  conversation: any;
  recalledIds: Set<string>;
  onBubbleContext: (e: React.MouseEvent | React.TouchEvent, message: Message, senderName: string, isOwn: boolean) => void;
  peerName: string;
  peerAvatar: string;
  readReceiptEnabled: boolean;
  onShowReadDetail: (msgId: string) => void;
}) {
  const { currentUser } = useIMStore();
  const { selectedConversationId } = useIMStore();
  const userProfiles = useChatStore(s => s.userProfiles);
  const storeGroupMembers = useChatStore(s => s.groupMembers);
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

  const GROUP_INTERVAL = 3 * 60 * 1000;

  type TimeGroup = { type: 'time'; time: string; key: string };
  type MsgGroup = { type: 'msg'; msgs: Message[]; key: string };

  const groups: (TimeGroup | MsgGroup)[] = [];
  let lastTime = 0;

  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i];
    const ts = msg.timestamp.getTime();

    // Time separator: first message or >5 min gap
    if (i === 0 || ts - lastTime > 5 * 60 * 1000) {
      groups.push({ type: 'time', time: formatTime(msg.timestamp), key: `t-${msg.id}` });
    }

    // New message group: different sender or >3 min gap
    const last = groups[groups.length - 1];
    if (!last || last.type === 'time' ||
      (last.type === 'msg' && (
        last.msgs[last.msgs.length - 1].senderId !== msg.senderId ||
        ts - last.msgs[last.msgs.length - 1].timestamp.getTime() > GROUP_INTERVAL
      ))
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
    if (conversation?.type === 'group') return groupMemberNames[senderId] || senderId;
    return conversation?.name || senderId;
  }, [conversation, currentUser?.name, isOwnMessage]);

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
                ? (userProfiles[msgs[0].senderId]?.avatar || '')
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
                <div style={{ fontSize: 12, color: '#3390EC', fontWeight: 600, marginBottom: 2, textAlign: 'left' }}>
                  {groupMemberNames[msgs[0].senderId] || msgs[0].senderId}
                </div>
              )}
              {msgs.map((m) => {
                const msgIsSent = isOwnMessage(m.senderId);
                const senderName = getSenderName(m.senderId);

                return (
                  <div key={m.id} style={{ textAlign: msgIsSent ? 'right' : 'left', marginBottom: 3 }}>
                    <div
                      className={`im-chat-bubble ${msgIsSent ? 'sent' : 'received'}`}
                      onContextMenu={(e) => onBubbleContext(e, m, senderName, msgIsSent)}
                      onTouchStart={(e) => handleTouchStart(e, m, senderName, msgIsSent)}
                      onTouchEnd={handleTouchEnd}
                      onTouchMove={handleTouchMove}
                      style={{ cursor: 'context-menu', userSelect: 'none', textAlign: 'left' }}
                    >
                    {recalledIds.has(m.id) ? (
                      <span style={{ fontStyle: 'italic', opacity: 0.6 }}>
                        {isOwnMessage(m.senderId) ? '你撤回了一条消息' : '对方撤回了一条消息'}
                      </span>
                    ) : (
                      <>
                        {/* Reply quote preview */}
                        {m.replyTo && (
                          <div style={{
                            padding: '4px 8px',
                            marginBottom: 4,
                            borderLeft: '3px solid #3390EC',
                            background: 'rgba(51,144,236,0.06)',
                            borderRadius: '0 6px 6px 0',
                            fontSize: 12,
                            lineHeight: 1.4,
                          }}>
                            <div style={{ fontWeight: 600, color: '#3390EC', marginBottom: 1 }}>
                              {m.replyTo.senderName}
                            </div>
                            <div style={{ opacity: 0.7, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {m.replyTo.content}
                            </div>
                          </div>
                        )}
                        <MessageContent message={m} />
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
                const checkColor = showRead ? '#3390EC' : '#C8CCD0';
                let footerText = formatBubbleTime(lastMsg.timestamp);
                if (lastMsg.status === 'failed') {
                  footerText = '发送失败';
                } else if (isSent && isGroup && readReceiptEnabled && totalMembers > 0) {
                  // 群聊自己发的消息：总是展示 X/N，哪怕 X=0，方便点开看谁没读
                  const rc = Math.min(lastMsg.readCount || 0, totalMembers);
                  footerText = `${rc}/${totalMembers} 人已读 · ${formatBubbleTime(lastMsg.timestamp)}`;
                } else if (showRead) {
                  footerText = `已读 · ${formatBubbleTime(lastMsg.timestamp)}`;
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
                    color: lastMsg.status === 'failed' ? '#e74c3c' : (showRead ? '#3390EC' : '#A2ACB5'),
                    lineHeight: 1,
                    cursor: 'pointer',
                    textDecoration: 'underline dotted',
                    textUnderlineOffset: 2,
                  } as React.CSSProperties,
                } : {
                  style: {
                    fontSize: 11,
                    color: lastMsg.status === 'failed' ? '#e74c3c' : (showRead ? '#3390EC' : '#A2ACB5'),
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
                        title="发送失败，点击重试"
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
                    <span {...footerSpanProps} title={canShowDetail ? '查看已读详情' : undefined}>
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
