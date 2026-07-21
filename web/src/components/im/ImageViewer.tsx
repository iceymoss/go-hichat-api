'use client';

import React, { useCallback, useEffect } from 'react';
import { X, ChevronLeft, ChevronRight } from 'lucide-react';

interface ImageViewerProps {
  images: string[];
  /** Index of the image currently shown. < 0 means the viewer is closed. */
  index: number;
  onClose: () => void;
  onIndexChange: (index: number) => void;
}

/**
 * Fullscreen image lightbox. Enlarges the clicked image and, when a trend has
 * multiple images, supports 上一张 / 下一张 navigation (wrap-around), a counter,
 * and keyboard control (Esc to close, ←/→ to switch).
 */
export default function ImageViewer({ images, index, onClose, onIndexChange }: ImageViewerProps) {
  const open = index >= 0 && index < images.length;
  const multiple = images.length > 1;

  const goPrev = useCallback(() => {
    onIndexChange((index - 1 + images.length) % images.length);
  }, [index, images.length, onIndexChange]);

  const goNext = useCallback(() => {
    onIndexChange((index + 1) % images.length);
  }, [index, images.length, onIndexChange]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      else if (e.key === 'ArrowLeft' && multiple) goPrev();
      else if (e.key === 'ArrowRight' && multiple) goNext();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, multiple, goPrev, goNext, onClose]);

  if (!open) return null;

  const arrowStyle: React.CSSProperties = {
    position: 'absolute',
    top: '50%',
    transform: 'translateY(-50%)',
    width: 44,
    height: 44,
    borderRadius: '50%',
    border: 'none',
    background: 'rgba(0,0,0,0.4)',
    color: '#FFF',
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  };

  return (
    <div
      className="fixed inset-0"
      style={{ zIndex: 10010, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.9)' }}
      onClick={onClose}
    >
      {/* Close */}
      <button
        onClick={(e) => { e.stopPropagation(); onClose(); }}
        style={{ position: 'absolute', top: 20, right: 20, width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'rgba(0,0,0,0.4)', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      >
        <X className="w-5 h-5" />
      </button>

      {/* Counter */}
      {multiple && (
        <div style={{ position: 'absolute', top: 24, left: '50%', transform: 'translateX(-50%)', fontSize: 14, color: '#FFF', background: 'rgba(0,0,0,0.4)', borderRadius: 14, padding: '4px 14px' }}>
          {index + 1} / {images.length}
        </div>
      )}

      {/* Prev */}
      {multiple && (
        <button onClick={(e) => { e.stopPropagation(); goPrev(); }} style={{ ...arrowStyle, left: 20 }}>
          <ChevronLeft className="w-6 h-6" />
        </button>
      )}

      {/* Image */}
      <img
        src={images[index]}
        alt=""
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: '90vw', maxHeight: '90vh', objectFit: 'contain', display: 'block', borderRadius: 4 }}
      />

      {/* Next */}
      {multiple && (
        <button onClick={(e) => { e.stopPropagation(); goNext(); }} style={{ ...arrowStyle, right: 20 }}>
          <ChevronRight className="w-6 h-6" />
        </button>
      )}
    </div>
  );
}
