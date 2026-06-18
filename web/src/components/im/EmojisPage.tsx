'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useIMStore } from '@/lib/im-store';
import {
  ArrowLeft, Plus, Trash2, Loader2, X, Smile, Upload,
} from 'lucide-react';
import { toast } from 'sonner';

interface EmojiItem {
  id: number;
  url: string;
  name: string;
  thumbnail: string;
  width: number;
  height: number;
  size: number;
  fileType: string;
  createdAt: string;
}

/* ── Lightbox preview ── */
function EmojiPreview({ emoji, onClose, onDelete }: {
  emoji: EmojiItem; onClose: () => void; onDelete: () => void;
}) {
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [onClose]);

  return (
    <div onClick={onClose} style={{
      position: 'fixed', inset: 0, zIndex: 2000,
      background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(8px)',
      display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
      animation: 'fade-in 0.15s ease-out',
    }}>
      <button onClick={onClose} style={{
        position: 'absolute', top: 16, right: 16,
        width: 36, height: 36, borderRadius: '50%', border: 'none',
        background: 'rgba(255,255,255,0.15)', color: '#fff', cursor: 'pointer',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}><X size={18} /></button>

      <div onClick={e => e.stopPropagation()}>
        <img src={emoji.url} alt={emoji.name}
          style={{ maxWidth: '80vw', maxHeight: '70vh', borderRadius: 8, objectFit: 'contain' }} />
      </div>

      <div onClick={e => e.stopPropagation()} style={{
        display: 'flex', alignItems: 'center', gap: 16, marginTop: 20,
      }}>
        {emoji.name && <span style={{ fontSize: 14, color: 'rgba(255,255,255,0.7)' }}>{emoji.name}</span>}
        <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.4)' }}>
          {emoji.fileType === 'gif' ? 'GIF' : '静态'} · {emoji.width > 0 ? `${emoji.width}×${emoji.height}` : ''} · {formatSize(emoji.size)}
        </span>
        <button onClick={onDelete} style={{
          padding: '6px 14px', borderRadius: 16, border: '1px solid rgba(255,255,255,0.25)',
          background: 'transparent', color: '#E53935', fontSize: 13, cursor: 'pointer',
          display: 'flex', alignItems: 'center', gap: 4,
        }}><Trash2 size={13} /> 删除</button>
      </div>
    </div>
  );
}

function formatSize(bytes: number) {
  if (!bytes) return '';
  if (bytes < 1024) return bytes + 'B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + 'KB';
  return (bytes / (1024 * 1024)).toFixed(1) + 'MB';
}

/* ═══════════════════════════════════════════════
   MAIN: EmojisPage
   ═══════════════════════════════════════════════ */
export default function EmojisPage({ onBack }: { onBack: () => void }) {
  const { currentUser: user } = useIMStore();
  const [items, setItems] = useState<EmojiItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [preview, setPreview] = useState<EmojiItem | null>(null);
  const [uploading, setUploading] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchList = useCallback(async (p = 1) => {
    if (!user?.token) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/user/emojis?page=${p}&pageSize=50`, {
        headers: { Authorization: `Bearer ${user.token}` },
      });
      const d = await res.json();
      if (d.success) {
        setItems(p === 1 ? d.data.list || [] : [...items, ...(d.data.list || [])]);
        setTotal(d.data.total || 0);
        setPage(p);
      }
    } catch { /* ignore */ }
    finally { setLoading(false); }
  }, [user?.token, items]);

  useEffect(() => { fetchList(1); }, [user?.token]);

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || !user?.token) return;
    setUploading(true);

    for (const file of Array.from(files)) {
      try {
        // 1. Upload file
        const fd = new FormData();
        fd.append('file', file);
        const uploadRes = await fetch('/api/user/emojis/upload', {
          method: 'POST', headers: { Authorization: `Bearer ${user.token}` }, body: fd,
        });
        const uploadData = await uploadRes.json();
        if (!uploadData.success) { toast.error(`${file.name}: ${uploadData.message}`); continue; }

        // 2. Add to collection
        const addRes = await fetch('/api/user/emojis', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
          body: JSON.stringify({
            url: uploadData.data.url,
            name: file.name.replace(/\.[^.]+$/, ''),
            size: uploadData.data.size,
            fileType: uploadData.data.fileType,
          }),
        });
        const addData = await addRes.json();
        if (addData.success) toast.success(`已添加: ${file.name}`);
        else toast.error(addData.message);
      } catch { toast.error(`${file.name} 上传失败`); }
    }

    setUploading(false);
    if (fileInputRef.current) fileInputRef.current.value = '';
    fetchList(1);
  };

  const handleDelete = async (id: number) => {
    if (!user?.token) return;
    try {
      const res = await fetch('/api/user/emojis', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
        body: JSON.stringify({ id }),
      });
      const d = await res.json();
      if (d.success) {
        setItems(items.filter(i => i.id !== id));
        setTotal(t => t - 1);
        setPreview(null);
        toast.success('已删除');
      } else toast.error(d.message);
    } catch { toast.error('删除失败'); }
  };

  if (!user) return null;

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      {/* Top bar */}
      <div style={{
        height: 56, display: 'flex', alignItems: 'center', padding: '0 8px', gap: 8,
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
          <Smile size={18} style={{ color: '#1BB45B' }} />
          <span style={{ fontSize: 17, fontWeight: 600, color: '#1C2733' }}>我的表情</span>
          <span style={{ fontSize: 12, color: '#A2ACB5', marginLeft: 4 }}>{total}</span>
        </div>

        <button onClick={() => setEditMode(!editMode)} style={{
          padding: '6px 12px', borderRadius: 16, border: 'none', fontSize: 12, fontWeight: 500, cursor: 'pointer',
          background: editMode ? '#E53935' : '#F0F2F5',
          color: editMode ? '#fff' : '#646A73',
          transition: 'all 0.2s',
        }}>{editMode ? '完成' : '管理'}</button>

        <button onClick={() => fileInputRef.current?.click()} disabled={uploading} style={{
          width: 32, height: 32, borderRadius: 8, border: 'none',
          background: '#1BB45B', color: '#fff', cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          {uploading ? <Loader2 size={16} className="animate-spin" /> : <Plus size={16} />}
        </button>
        <input ref={fileInputRef} type="file" accept="image/*,.gif" multiple className="hidden" onChange={handleUpload} />
      </div>

      {/* Grid */}
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        {loading && items.length === 0 ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="animate-spin" style={{ color: '#1BB45B' }} />
          </div>
        ) : items.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 20px' }}>
            <Smile size={48} style={{ color: '#E0E3E8', margin: '0 auto 16px' }} />
            <div style={{ fontSize: 15, color: '#646A73', fontWeight: 500 }}>暂无自定义表情</div>
            <div style={{ fontSize: 13, color: '#A2ACB5', marginTop: 4 }}>点击右上角 + 上传你的表情包</div>
            <button onClick={() => fileInputRef.current?.click()} style={{
              marginTop: 20, padding: '10px 24px', borderRadius: 20,
              background: '#1BB45B', color: '#fff', border: 'none',
              fontSize: 14, fontWeight: 500, cursor: 'pointer',
              display: 'inline-flex', alignItems: 'center', gap: 6,
            }}><Upload size={16} /> 上传表情</button>
          </div>
        ) : (
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(4, 1fr)',
            gap: 8,
            padding: 12,
          }}>
            {items.map(item => (
              <div key={item.id}
                onClick={() => editMode ? handleDelete(item.id) : setPreview(item)}
                style={{
                  position: 'relative', aspectRatio: '1',
                  borderRadius: 12, overflow: 'hidden', cursor: 'pointer',
                  background: '#fff', border: '1px solid rgba(0,0,0,0.04)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  transition: 'transform 0.15s, box-shadow 0.15s',
                }}
                onMouseEnter={e => {
                  e.currentTarget.style.transform = 'scale(1.04)';
                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.1)';
                }}
                onMouseLeave={e => {
                  e.currentTarget.style.transform = 'scale(1)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                <img src={item.url} alt={item.name}
                  loading="lazy"
                  style={{ maxWidth: '88%', maxHeight: '88%', objectFit: 'contain' }} />

                {/* GIF badge */}
                {item.fileType === 'gif' && (
                  <span style={{
                    position: 'absolute', top: 4, right: 4,
                    padding: '1px 5px', borderRadius: 4,
                    background: 'rgba(0,0,0,0.5)', fontSize: 10, color: '#fff', fontWeight: 600,
                  }}>GIF</span>
                )}

                {/* Delete overlay in edit mode */}
                {editMode && (
                  <div style={{
                    position: 'absolute', inset: 0,
                    background: 'rgba(229,57,53,0.12)',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                  }}>
                    <div style={{
                      width: 28, height: 28, borderRadius: '50%',
                      background: '#E53935', display: 'flex', alignItems: 'center', justifyContent: 'center',
                    }}>
                      <X size={14} color="#fff" />
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Load more */}
        {items.length < total && items.length > 0 && (
          <div style={{ textAlign: 'center', padding: '12px 0 20px' }}>
            <button onClick={() => fetchList(page + 1)} disabled={loading}
              style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, color: '#1BB45B', fontWeight: 500 }}>
              {loading ? <Loader2 size={14} className="animate-spin" /> : '加载更多'}
            </button>
          </div>
        )}
      </div>

      {/* Preview lightbox */}
      {preview && (
        <EmojiPreview
          emoji={preview}
          onClose={() => setPreview(null)}
          onDelete={() => handleDelete(preview.id)}
        />
      )}
    </div>
  );
}
