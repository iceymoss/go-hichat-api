'use client';

import React, { useRef, useState } from 'react';
import { Play, Pause } from 'lucide-react';

interface Props {
  url: string;
  duration?: number;
  /** 是否对方发来的未播放语音（显示红点） */
  unplayed?: boolean;
  /** 首次播放回调（用于清除未读红点） */
  onPlayed?: () => void;
}

const BAR_HEIGHTS = [10, 16, 22, 16, 10];

/** 微信风格语音条：播放/暂停 + 波形动画 + 时长，未读显示红点 */
export default function VoiceBubble({ url, duration, unplayed, onPlayed }: Props) {
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
    <span style={{ position: 'relative', display: 'inline-flex', alignItems: 'center' }}>
      <span
        onClick={toggle}
        style={{ display: 'inline-flex', alignItems: 'center', gap: 8, width, cursor: 'pointer', userSelect: 'none', padding: '2px 0' }}
      >
        {playing ? <Pause size={18} style={{ flexShrink: 0 }} /> : <Play size={18} style={{ flexShrink: 0 }} />}
        <span style={{ flex: 1, display: 'inline-flex', alignItems: 'center', gap: 3, height: 22 }}>
          {BAR_HEIGHTS.map((h, i) => (
            <span
              key={i}
              className={`voice-bar${playing ? ' playing' : ''}`}
              style={{ height: h, animationDelay: playing ? `${i * 0.12}s` : undefined }}
            />
          ))}
        </span>
        <span style={{ fontSize: 12, opacity: 0.7, flexShrink: 0 }}>{duration ? `${duration}"` : ''}</span>
      </span>

      {unplayed && !playing && (
        <span style={{ position: 'absolute', top: -2, right: -8, width: 8, height: 8, borderRadius: '50%', background: '#FA5151' }} />
      )}

      <audio
        ref={audioRef}
        src={url}
        onPlay={() => { setPlaying(true); onPlayed?.(); }}
        onPause={() => setPlaying(false)}
        onEnded={() => setPlaying(false)}
        style={{ display: 'none' }}
      />
    </span>
  );
}
