'use client';

import React, { useRef, useState } from 'react';
import { Play, Pause } from 'lucide-react';

interface Props {
  url: string;
  duration?: number;
}

/** 微信风格语音条：播放/暂停图标 + 时长，点击播放（不使用原生 controls，避免遮挡右键菜单） */
export default function VoiceBubble({ url, duration }: Props) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);

  const toggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    const a = audioRef.current;
    if (!a) return;
    if (playing) a.pause();
    else a.play().catch(() => {});
  };

  // 气泡宽度随时长增长（微信观感），范围 [70, 200]
  const width = Math.min(200, 70 + (duration || 1) * 7);

  return (
    <span
      onClick={toggle}
      style={{ display: 'inline-flex', alignItems: 'center', gap: 8, width, cursor: 'pointer', userSelect: 'none', padding: '2px 0' }}
    >
      {playing ? <Pause size={18} style={{ flexShrink: 0 }} /> : <Play size={18} style={{ flexShrink: 0 }} />}
      <span style={{ flex: 1, height: 3, background: 'currentColor', opacity: 0.3, borderRadius: 2 }} />
      <span style={{ fontSize: 12, opacity: 0.7, flexShrink: 0 }}>{duration ? `${duration}"` : ''}</span>
      <audio
        ref={audioRef}
        src={url}
        onPlay={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onEnded={() => setPlaying(false)}
        style={{ display: 'none' }}
      />
    </span>
  );
}
