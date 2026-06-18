'use client';

import React, { useEffect, useState } from 'react';
import data from '@emoji-mart/data';
import Picker from '@emoji-mart/react';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { useT } from '@/hooks/use-i18n';

export interface StickerItem {
  id: number;
  url: string;
  name?: string;
  thumbnail?: string;
  width?: number;
  height?: number;
}

interface Props {
  token?: string;
  /** 选中 emoji（插入到输入框） */
  onPickEmoji: (native: string) => void;
  /** 选中收藏表情（直接发送到会话） */
  onPickSticker: (s: StickerItem) => void;
  children: React.ReactNode;
}

/** 输入框左侧的表情面板：Emoji（全量开源数据）+ 我的表情收藏 */
export default function ChatEmojiPanel({ token, onPickEmoji, onPickSticker, children }: Props) {
  const t = useT();
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<'emoji' | 'sticker'>('emoji');
  const [stickers, setStickers] = useState<StickerItem[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    if (!open || tab !== 'sticker' || loaded || !token) return;
    (async () => {
      try {
        const res = await fetch('/api/user/emojis?page=1&pageSize=50', {
          headers: { Authorization: `Bearer ${token}` },
        });
        const d = await res.json();
        if (d.success) setStickers(d.data.list || []);
      } catch { /* ignore */ }
      finally { setLoaded(true); }
    })();
  }, [open, tab, loaded, token]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent align="start" side="top" className="w-auto p-0 overflow-hidden" style={{ width: 352 }}>
        <div className="flex" style={{ borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
          <button
            onClick={() => setTab('emoji')}
            style={{ flex: 1, padding: '8px 0', border: 'none', background: 'transparent', cursor: 'pointer', fontSize: 13, fontWeight: tab === 'emoji' ? 600 : 400, color: tab === 'emoji' ? '#1BB45B' : '#708499' }}
          >
            Emoji
          </button>
          <button
            onClick={() => setTab('sticker')}
            style={{ flex: 1, padding: '8px 0', border: 'none', background: 'transparent', cursor: 'pointer', fontSize: 13, fontWeight: tab === 'sticker' ? 600 : 400, color: tab === 'sticker' ? '#1BB45B' : '#708499' }}
          >
            {t('emojiPanel.title')}
          </button>
        </div>

        {tab === 'emoji' ? (
          <Picker
            data={data}
            onEmojiSelect={(e: { native: string }) => onPickEmoji(e.native)}
            theme="light"
            previewPosition="none"
            skinTonePosition="none"
            locale="zh"
          />
        ) : (
          <div style={{ height: 320, overflowY: 'auto', padding: 12 }}>
            {stickers.length === 0 ? (
              <div style={{ textAlign: 'center', color: '#999', fontSize: 13, paddingTop: 40 }}>
                {loaded ? t('emojiPanel.empty') : t('emojiPanel.loading')}
              </div>
            ) : (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 8 }}>
                {stickers.map((s) => (
                  <img
                    key={s.id}
                    src={s.thumbnail || s.url}
                    alt={s.name || ''}
                    onClick={() => { onPickSticker(s); setOpen(false); }}
                    style={{ width: '100%', aspectRatio: '1', objectFit: 'cover', borderRadius: 8, cursor: 'pointer' }}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
