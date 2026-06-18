'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useT } from '@/hooks/use-i18n';
import {
  ArrowLeft, Image as ImageIcon, Play, MapPin, Heart, MessageCircle,
  Loader2, X, ChevronLeft, ChevronRight,
} from 'lucide-react';

interface MediaItem {
  trendId: number;
  url: string;
  type: 'image' | 'video';
  coverUrl: string;
  content: string;
  location: string;
  createTime: number;
  replyCount: number;
  agreeCount: number;
}

/* ── Lightbox: full-screen media viewer ── */
function Lightbox({ items, index, onClose, onPrev, onNext }: {
  items: MediaItem[]; index: number;
  onClose: () => void; onPrev: () => void; onNext: () => void;
}) {
  const item = items[index];

  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowLeft') onPrev();
      if (e.key === 'ArrowRight') onNext();
    };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [onClose, onPrev, onNext]);

  if (!item) return null;

  const date = new Date(item.createTime * 1000);
  const dateStr = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;

  return (
    <div onClick={onClose} style={{
      position: 'fixed', inset: 0, zIndex: 2000,
      background: 'rgba(0,0,0,0.92)', backdropFilter: 'blur(8px)',
      display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
      animation: 'fade-in 0.15s ease-out',
    }}>
      {/* Close */}
      <button onClick={onClose} style={{
        position: 'absolute', top: 16, right: 16, zIndex: 10,
        width: 36, height: 36, borderRadius: '50%', border: 'none',
        background: 'rgba(255,255,255,0.15)', color: '#fff', cursor: 'pointer',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}><X size={18} /></button>

      {/* Counter */}
      <div style={{
        position: 'absolute', top: 18, left: '50%', transform: 'translateX(-50%)',
        fontSize: 14, color: 'rgba(255,255,255,0.7)', fontWeight: 500,
      }}>{index + 1} / {items.length}</div>

      {/* Prev/Next */}
      {index > 0 && (
        <button onClick={e => { e.stopPropagation(); onPrev(); }} style={{
          position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)',
          width: 40, height: 40, borderRadius: '50%', border: 'none',
          background: 'rgba(255,255,255,0.12)', color: '#fff', cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}><ChevronLeft size={20} /></button>
      )}
      {index < items.length - 1 && (
        <button onClick={e => { e.stopPropagation(); onNext(); }} style={{
          position: 'absolute', right: 16, top: '50%', transform: 'translateY(-50%)',
          width: 40, height: 40, borderRadius: '50%', border: 'none',
          background: 'rgba(255,255,255,0.12)', color: '#fff', cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}><ChevronRight size={20} /></button>
      )}

      {/* Media */}
      <div onClick={e => e.stopPropagation()} style={{ maxWidth: '90vw', maxHeight: '80vh' }}>
        {item.type === 'video' ? (
          <video src={item.url} controls autoPlay style={{
            maxWidth: '90vw', maxHeight: '75vh', borderRadius: 8,
          }} />
        ) : (
          <img src={item.url} alt="" style={{
            maxWidth: '90vw', maxHeight: '75vh', borderRadius: 8, objectFit: 'contain',
          }} />
        )}
      </div>

      {/* Info bar */}
      <div onClick={e => e.stopPropagation()} style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        padding: '16px 24px 20px',
        background: 'linear-gradient(transparent, rgba(0,0,0,0.7))',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, justifyContent: 'center' }}>
          <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.7)' }}>{dateStr}</span>
          {item.location && (
            <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)', display: 'flex', alignItems: 'center', gap: 3 }}>
              <MapPin size={12} /> {item.location}
            </span>
          )}
          <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)', display: 'flex', alignItems: 'center', gap: 3 }}>
            <Heart size={12} /> {item.agreeCount}
          </span>
          <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)', display: 'flex', alignItems: 'center', gap: 3 }}>
            <MessageCircle size={12} /> {item.replyCount}
          </span>
        </div>
        {item.content && (
          <div style={{
            fontSize: 13, color: 'rgba(255,255,255,0.8)', marginTop: 8, textAlign: 'center',
            maxWidth: 500, margin: '8px auto 0', lineHeight: 1.5,
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>{item.content}</div>
        )}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   MAIN: AlbumPage
   ═══════════════════════════════════════════════ */
export default function AlbumPage({ onBack }: { onBack: () => void }) {
  const { currentUser: user } = useIMStore();
  const t = useT();
  const [items, setItems] = useState<MediaItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastId, setLastId] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [lightboxIdx, setLightboxIdx] = useState(-1);

  const fetchAlbum = useCallback(async (lid = 0) => {
    if (!user?.token) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/user/album?last_id=${lid}&user_id=${user.id}`, {
        headers: { Authorization: `Bearer ${user.token}` },
      });
      const d = await res.json();
      if (d.success) {
        const newItems = d.data.items || [];
        setItems(lid === 0 ? newItems : [...items, ...newItems]);
        setLastId(d.data.lastId || 0);
        setHasMore(d.data.hasMore);
      }
    } catch { /* ignore */ }
    finally { setLoading(false); }
  }, [user?.token, user?.id, items]);

  useEffect(() => { fetchAlbum(0); }, [user?.token]);

  // Group items by month
  const grouped = items.reduce<Record<string, MediaItem[]>>((acc, item) => {
    const d = new Date(item.createTime * 1000);
    const key = t('album.month').replace('{y}', String(d.getFullYear())).replace('{m}', String(d.getMonth() + 1));
    if (!acc[key]) acc[key] = [];
    acc[key].push(item);
    return acc;
  }, {});

  const totalCount = items.length;
  const imageCount = items.filter(i => i.type === 'image').length;
  const videoCount = items.filter(i => i.type === 'video').length;

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      {/* Top bar */}
      <div style={{
        height: 56, display: 'flex', alignItems: 'center', padding: '0 8px',
        background: '#fff', borderBottom: '1px solid rgba(0,0,0,0.06)', flexShrink: 0,
      }}>
        <button onClick={onBack} style={{
          background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4,
          color: '#1BB45B', fontSize: 14, fontWeight: 500, padding: '8px 12px', borderRadius: 8,
          transition: 'background 0.15s',
        }}
          onMouseEnter={e => { e.currentTarget.style.background = 'rgba(27,180,91,0.06)'; }}
          onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; }}
        ><ArrowLeft size={18} /></button>
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', gap: 6 }}>
          <ImageIcon size={18} style={{ color: '#1BB45B' }} />
          <span style={{ fontSize: 17, fontWeight: 600, color: '#1C2733' }}>{t('album.title')}</span>
        </div>
        {totalCount > 0 && (
          <div style={{ display: 'flex', gap: 10, paddingRight: 8 }}>
            <span style={{ fontSize: 12, color: '#646A73' }}>📷 {imageCount}</span>
            {videoCount > 0 && <span style={{ fontSize: 12, color: '#646A73' }}>🎬 {videoCount}</span>}
          </div>
        )}
      </div>

      {/* Content */}
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        {loading && items.length === 0 ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="animate-spin" style={{ color: '#1BB45B' }} />
          </div>
        ) : items.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 20px' }}>
            <ImageIcon size={48} style={{ color: '#E0E3E8', margin: '0 auto 16px' }} />
            <div style={{ fontSize: 15, color: '#646A73', fontWeight: 500 }}>{t('album.empty')}</div>
            <div style={{ fontSize: 13, color: '#A2ACB5', marginTop: 4 }}>{t('album.emptyHint')}</div>
          </div>
        ) : (
          <div>
            {Object.entries(grouped).map(([month, monthItems]) => (
              <div key={month}>
                {/* Month header */}
                <div style={{
                  padding: '14px 16px 8px', fontSize: 14, fontWeight: 600, color: '#1C2733',
                  position: 'sticky', top: 0, background: '#F0F2F5', zIndex: 5,
                }}>
                  {month}
                  <span style={{ fontSize: 12, color: '#A2ACB5', fontWeight: 400, marginLeft: 8 }}>
                    {t('album.itemCount').replace('{count}', String(monthItems.length))}
                  </span>
                </div>

                {/* Grid */}
                <div style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(3, 1fr)',
                  gap: 3,
                  padding: '0 3px',
                }}>
                  {monthItems.map((item, idx) => {
                    // Find global index for lightbox
                    const globalIdx = items.indexOf(item);
                    return (
                      <div key={`${item.trendId}-${item.url}`}
                        onClick={() => setLightboxIdx(globalIdx)}
                        style={{
                          position: 'relative', aspectRatio: '1', cursor: 'pointer',
                          overflow: 'hidden', background: '#E8EAED',
                        }}
                      >
                        <img
                          src={item.type === 'video' && item.coverUrl ? item.coverUrl : item.url}
                          alt="" loading="lazy"
                          style={{ width: '100%', height: '100%', objectFit: 'cover', transition: 'transform 0.2s' }}
                          onMouseEnter={e => { (e.currentTarget as HTMLImageElement).style.transform = 'scale(1.05)'; }}
                          onMouseLeave={e => { (e.currentTarget as HTMLImageElement).style.transform = 'scale(1)'; }}
                        />
                        {/* Video overlay */}
                        {item.type === 'video' && (
                          <div style={{
                            position: 'absolute', inset: 0,
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            background: 'rgba(0,0,0,0.2)',
                          }}>
                            <div style={{
                              width: 36, height: 36, borderRadius: '50%',
                              background: 'rgba(255,255,255,0.9)',
                              display: 'flex', alignItems: 'center', justifyContent: 'center',
                            }}>
                              <Play size={16} style={{ color: '#1C2733', marginLeft: 2 }} fill="#1C2733" />
                            </div>
                          </div>
                        )}

                        {/* Date badge on first item of group */}
                        {idx === 0 && (
                          <div style={{
                            position: 'absolute', bottom: 4, left: 4,
                            padding: '2px 6px', borderRadius: 4,
                            background: 'rgba(0,0,0,0.5)', fontSize: 10, color: '#fff',
                          }}>
                            {t('album.dayBadge').replace('{d}', String(new Date(item.createTime * 1000).getDate()))}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}

            {/* Load more */}
            {hasMore && (
              <div style={{ textAlign: 'center', padding: '16px 0 24px' }}>
                <button onClick={() => fetchAlbum(lastId)} disabled={loading}
                  style={{
                    background: 'none', border: 'none', cursor: 'pointer',
                    fontSize: 13, color: '#1BB45B', fontWeight: 500,
                  }}>
                  {loading ? <Loader2 size={14} className="animate-spin" /> : t('common.loadMore')}
                </button>
              </div>
            )}

            {!hasMore && items.length > 0 && (
              <div style={{ textAlign: 'center', padding: '12px 0 24px', fontSize: 12, color: '#A2ACB5' }}>
                {t('album.allLoaded').replace('{count}', String(totalCount))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Lightbox */}
      {lightboxIdx >= 0 && (
        <Lightbox
          items={items}
          index={lightboxIdx}
          onClose={() => setLightboxIdx(-1)}
          onPrev={() => setLightboxIdx(Math.max(0, lightboxIdx - 1))}
          onNext={() => setLightboxIdx(Math.min(items.length - 1, lightboxIdx + 1))}
        />
      )}
    </div>
  );
}
