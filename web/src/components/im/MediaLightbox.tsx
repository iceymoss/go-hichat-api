'use client';

import React, { useCallback, useEffect } from 'react';
import { X, ChevronLeft, ChevronRight } from 'lucide-react';

export interface LightboxItem {
  type: 'image' | 'video';
  url: string;
}

interface Props {
  items: LightboxItem[];
  index: number;
  onClose: () => void;
  onIndex: (i: number) => void;
}

/** 全屏图片/视频预览，支持多张切换（← → / 点击箭头），视频可播放 */
export default function MediaLightbox({ items, index, onClose, onIndex }: Props) {
  const go = useCallback((delta: number) => {
    const n = index + delta;
    if (n >= 0 && n < items.length) onIndex(n);
  }, [index, items.length, onIndex]);

  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      else if (e.key === 'ArrowLeft') go(-1);
      else if (e.key === 'ArrowRight') go(1);
    };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [onClose, go]);

  const cur = items[index];
  if (!cur) return null;

  const navBtn: React.CSSProperties = {
    position: 'absolute', top: '50%', transform: 'translateY(-50%)',
    width: 44, height: 44, borderRadius: '50%', border: 'none',
    background: 'rgba(255,255,255,0.15)', color: '#fff', cursor: 'pointer',
    display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1,
  };

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed', inset: 0, zIndex: 3000,
        background: 'rgba(0,0,0,0.9)', backdropFilter: 'blur(6px)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        animation: 'fade-in 0.15s ease-out',
      }}
    >
      <button
        onClick={onClose}
        style={{ position: 'absolute', top: 16, right: 16, width: 36, height: 36, borderRadius: '50%', border: 'none', background: 'rgba(255,255,255,0.15)', color: '#fff', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      >
        <X size={18} />
      </button>

      {items.length > 1 && index > 0 && (
        <button onClick={(e) => { e.stopPropagation(); go(-1); }} style={{ ...navBtn, left: 16 }}>
          <ChevronLeft size={24} />
        </button>
      )}

      <div onClick={(e) => e.stopPropagation()} style={{ maxWidth: '90vw', maxHeight: '85vh' }}>
        {cur.type === 'video' ? (
          <video src={cur.url} controls autoPlay style={{ maxWidth: '90vw', maxHeight: '85vh', borderRadius: 8 }} />
        ) : (
          <img src={cur.url} alt="" style={{ maxWidth: '90vw', maxHeight: '85vh', objectFit: 'contain', borderRadius: 8 }} />
        )}
      </div>

      {items.length > 1 && index < items.length - 1 && (
        <button onClick={(e) => { e.stopPropagation(); go(1); }} style={{ ...navBtn, right: 16 }}>
          <ChevronRight size={24} />
        </button>
      )}

      {items.length > 1 && (
        <div style={{ position: 'absolute', bottom: 20, color: 'rgba(255,255,255,0.7)', fontSize: 13 }}>
          {index + 1} / {items.length}
        </div>
      )}
    </div>
  );
}
