'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useT } from '@/hooks/use-i18n';
import {
  ArrowLeft, Search, Plus, FileText, Image as ImageIcon, Link2,
  File, Video, MapPin, StickyNote, Trash2, Loader2, X, Star,
  Filter, ChevronDown,
} from 'lucide-react';
import { toast } from 'sonner';

/* ── Type config ── */
const TYPE_MAP: Record<string, { labelKey: string; icon: React.ReactNode; color: string }> = {
  text:     { labelKey: 'fav.type.text', icon: <FileText size={16} />, color: '#1BB45B' },
  image:    { labelKey: 'fav.type.image', icon: <ImageIcon size={16} />, color: '#FF9500' },
  link:     { labelKey: 'fav.type.link', icon: <Link2 size={16} />, color: '#5856D6' },
  file:     { labelKey: 'fav.type.file', icon: <File size={16} />, color: '#34C759' },
  video:    { labelKey: 'fav.type.video', icon: <Video size={16} />, color: '#FF2D55' },
  location: { labelKey: 'fav.type.location', icon: <MapPin size={16} />, color: '#FF6B35' },
  note:     { labelKey: 'fav.type.note', icon: <StickyNote size={16} />, color: '#AF52DE' },
};

interface FavItem {
  id: number;
  type: string;
  title: string;
  content: string;
  source: string;
  sourceId: string;
  tags: string;
  createdAt: string;
}

/* ── Add Favorite Dialog ── */
const FILE_TYPES = new Set(['image', 'file', 'video']);

function AddFavoriteDialog({ open, onClose, onAdded, token }: {
  open: boolean; onClose: () => void; onAdded: () => void; token: string;
}) {
  const t = useT();
  const [type, setType] = useState('text');
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [saving, setSaving] = useState(false);
  // File upload state
  const [uploadedFile, setUploadedFile] = useState<{ url: string; name: string; size: number; fileType: string } | null>(null);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) { setType('text'); setTitle(''); setContent(''); setUploadedFile(null); }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [open, onClose]);

  const isFileType = FILE_TYPES.has(type);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append('file', file);
      const res = await fetch('/api/user/favorites/upload', {
        method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: fd,
      });
      const d = await res.json();
      if (d.success) {
        setUploadedFile(d.data);
        if (!title) setTitle(d.data.name || file.name);
        toast.success(t('fav.uploadOk'));
      } else toast.error(d.message || t('fav.uploadFail'));
    } catch { toast.error(t('fav.uploadFail')); }
    finally { setUploading(false); if (fileInputRef.current) fileInputRef.current.value = ''; }
  };

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (!file) return;
    // Trigger same logic as file select
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append('file', file);
      const res = await fetch('/api/user/favorites/upload', {
        method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: fd,
      });
      const d = await res.json();
      if (d.success) {
        setUploadedFile(d.data);
        if (!title) setTitle(d.data.name || file.name);
        toast.success(t('fav.uploadOk'));
      } else toast.error(d.message || t('fav.uploadFail'));
    } catch { toast.error(t('fav.uploadFail')); }
    finally { setUploading(false); }
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + 'B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + 'KB';
    return (bytes / (1024 * 1024)).toFixed(1) + 'MB';
  };

  const fileAccept = type === 'image' ? 'image/*' : type === 'video' ? 'video/*' : '*';

  const submit = async () => {
    // For file types, need uploaded file; for text types, need content
    if (isFileType && !uploadedFile) { toast.error(t('fav.needFile')); return; }
    if (!isFileType && !content.trim()) { toast.error(t('fav.needContent')); return; }
    setSaving(true);
    try {
      let body: any = { type, title: title.trim(), source: 'manual' };

      if (isFileType && uploadedFile) {
        if (type === 'image') {
          body.content = JSON.stringify({ url: uploadedFile.url, name: uploadedFile.name, size: uploadedFile.size });
        } else if (type === 'video') {
          body.content = JSON.stringify({ url: uploadedFile.url, name: uploadedFile.name, size: uploadedFile.size });
        } else {
          body.content = JSON.stringify({ url: uploadedFile.url, name: uploadedFile.name, size: uploadedFile.size });
        }
        if (!body.title) body.title = uploadedFile.name;
      } else if (type === 'link') {
        body.content = JSON.stringify({ url: content.trim(), title: title.trim() });
        if (!body.title) body.title = content.trim();
      } else {
        body.content = JSON.stringify({ text: content.trim() });
        if (!body.title) body.title = content.trim().slice(0, 50);
      }

      const res = await fetch('/api/user/favorites', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(body),
      });
      const d = await res.json();
      if (d.success) { toast.success(t('fav.addOk')); onAdded(); onClose(); }
      else toast.error(d.message);
    } catch { toast.error(t('fav.addFail')); }
    finally { setSaving(false); }
  };

  if (!open) return null;

  return (
    <div onClick={onClose} style={{
      position: 'fixed', inset: 0, zIndex: 1000,
      background: 'rgba(0,0,0,0.45)', backdropFilter: 'blur(2px)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      animation: 'fade-in 0.15s ease-out',
    }}>
      <div onClick={e => e.stopPropagation()} style={{
        width: 440, maxWidth: 'calc(100vw - 32px)',
        background: '#fff', borderRadius: 16, boxShadow: '0 8px 32px rgba(0,0,0,0.18)',
        overflow: 'hidden',
      }}>
        <div style={{
          padding: '16px 20px 12px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          borderBottom: '1px solid rgba(0,0,0,0.06)',
        }}>
          <span style={{ fontSize: 16, fontWeight: 600, color: '#1C2733' }}>{t('fav.add')}</span>
          <button onClick={onClose} style={{
            width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent',
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#A2ACB5',
          }}><X size={16} /></button>
        </div>

        <div style={{ padding: '16px 20px 20px' }}>
          {/* Type selector */}
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 16 }}>
            {Object.entries(TYPE_MAP).map(([k, v]) => (
              <button key={k} onClick={() => { setType(k); setUploadedFile(null); setContent(''); }} style={{
                display: 'flex', alignItems: 'center', gap: 4, padding: '6px 12px', borderRadius: 16,
                fontSize: 12, fontWeight: 500, border: 'none', cursor: 'pointer',
                background: type === k ? v.color : '#F0F2F5',
                color: type === k ? '#fff' : '#646A73',
                transition: 'all 0.2s',
              }}>{v.icon} {t(v.labelKey)}</button>
            ))}
          </div>

          {/* Title */}
          <input value={title} onChange={e => setTitle(e.target.value)}
            placeholder={t('fav.titlePlaceholder')} maxLength={100}
            style={{
              width: '100%', height: 40, padding: '0 14px', borderRadius: 10,
              border: '1.5px solid #E0E3E8', fontSize: 14, color: '#1C2733',
              background: '#FAFBFC', outline: 'none', marginBottom: 12,
            }}
            onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
            onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
          />

          {/* File upload zone for image/file/video */}
          {isFileType ? (
            <div>
              {uploadedFile ? (
                <div style={{
                  padding: 12, borderRadius: 10, border: '1.5px solid #4DCD5E', background: '#F6FFF6',
                  display: 'flex', alignItems: 'center', gap: 10,
                }}>
                  {uploadedFile.fileType === 'image' ? (
                    <img src={uploadedFile.url} alt="" style={{ width: 48, height: 48, borderRadius: 8, objectFit: 'cover' }} />
                  ) : (
                    <div style={{ width: 48, height: 48, borderRadius: 8, background: '#F0F2F5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      {uploadedFile.fileType === 'video' ? <Video size={20} style={{ color: '#FF2D55' }} /> : <File size={20} style={{ color: '#34C759' }} />}
                    </div>
                  )}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13, fontWeight: 500, color: '#1C2733', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{uploadedFile.name}</div>
                    <div style={{ fontSize: 12, color: '#8F959E', marginTop: 2 }}>{formatSize(uploadedFile.size)}</div>
                  </div>
                  <button onClick={() => setUploadedFile(null)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', padding: 4 }}><X size={16} /></button>
                </div>
              ) : (
                <div
                  onDragOver={e => e.preventDefault()}
                  onDrop={handleDrop}
                  onClick={() => fileInputRef.current?.click()}
                  style={{
                    padding: '28px 20px', borderRadius: 10, border: '2px dashed #D0D5DD',
                    background: '#FAFBFC', textAlign: 'center', cursor: 'pointer',
                    transition: 'all 0.2s',
                  }}
                  onMouseEnter={e => { e.currentTarget.style.borderColor = '#1BB45B'; e.currentTarget.style.background = 'rgba(27,180,91,0.02)'; }}
                  onMouseLeave={e => { e.currentTarget.style.borderColor = '#D0D5DD'; e.currentTarget.style.background = '#FAFBFC'; }}
                >
                  {uploading ? (
                    <Loader2 size={28} className="animate-spin" style={{ color: '#1BB45B', margin: '0 auto 8px' }} />
                  ) : (
                    <div style={{ color: '#A2ACB5', marginBottom: 8 }}>
                      {type === 'image' ? <ImageIcon size={28} style={{ margin: '0 auto' }} /> : type === 'video' ? <Video size={28} style={{ margin: '0 auto' }} /> : <File size={28} style={{ margin: '0 auto' }} />}
                    </div>
                  )}
                  <div style={{ fontSize: 14, color: '#646A73', fontWeight: 500 }}>
                    {uploading ? t('fav.uploading') : t('fav.dropHint')}
                  </div>
                  <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 4 }}>
                    {type === 'image' ? t('fav.supportImage') : type === 'video' ? t('fav.supportVideo') : t('fav.supportFile')}
                  </div>
                  <input ref={fileInputRef} type="file" accept={fileAccept} className="hidden" onChange={handleFileSelect} />
                </div>
              )}
            </div>
          ) : (
            /* Text input for text/note/link/location */
            <textarea value={content} onChange={e => setContent(e.target.value)}
              placeholder={type === 'link' ? t('fav.placeholderLink') : type === 'location' ? t('fav.placeholderLocation') : t('fav.placeholderContent')}
              rows={4}
              style={{
                width: '100%', padding: '10px 14px', borderRadius: 10,
                border: '1.5px solid #E0E3E8', fontSize: 14, color: '#1C2733',
                background: '#FAFBFC', outline: 'none', resize: 'none',
                fontFamily: 'inherit', lineHeight: 1.6,
              }}
              onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
            />
          )}

          <button className="hc-btn-primary"
            disabled={saving || (isFileType ? !uploadedFile : !content.trim())}
            onClick={submit}
            style={{ marginTop: 16, padding: '12px 24px', borderRadius: 22 }}>
            {saving ? <Loader2 size={16} className="animate-spin" /> : t('fav.saveBtn')}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ── Favorite Card ── */
function FavCard({ item, onDelete }: { item: FavItem; onDelete: () => void }) {
  const t = useT();
  const cfg = TYPE_MAP[item.type] || TYPE_MAP.text;
  const [deleting, setDeleting] = useState(false);

  let preview = '';
  try {
    const c = JSON.parse(item.content);
    preview = c.text || c.url || c.title || c.name || '';
  } catch {
    preview = item.content;
  }

  // For image type, try to show thumbnail
  let imageUrl = '';
  try {
    const c = JSON.parse(item.content);
    if (item.type === 'image' && c.url) imageUrl = c.url;
  } catch {}

  const handleDelete = async () => {
    setDeleting(true);
    onDelete();
  };

  const date = item.createdAt?.replace(/T.*/, '') || '';
  const sourceLabel = item.source === 'chat' ? t('fav.fromChat') : item.source === 'moment' ? t('fav.fromMoment') : '';

  return (
    <div className="im-profile-menu-item" style={{
      padding: '14px 16px', alignItems: 'flex-start', gap: 12,
      display: 'flex',
    }}>
      {/* Type icon */}
      <div style={{
        width: 36, height: 36, borderRadius: 10, flexShrink: 0,
        background: `${cfg.color}14`, color: cfg.color,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}>{cfg.icon}</div>

      {/* Content */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{
          fontSize: 14, fontWeight: 500, color: '#1C2733', lineHeight: 1.4,
          overflow: 'hidden', textOverflow: 'ellipsis',
          display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
        }}>{item.title || preview}</div>

        {/* Image preview */}
        {imageUrl && (
          <div style={{ marginTop: 8 }}>
            <img src={imageUrl} alt="" style={{
              maxWidth: 200, maxHeight: 120, borderRadius: 8, objectFit: 'cover',
            }} />
          </div>
        )}

        {/* Preview text (if title differs from preview) */}
        {item.title && preview && item.title !== preview && item.type !== 'image' && (
          <div style={{
            fontSize: 13, color: '#8F959E', marginTop: 4, lineHeight: 1.4,
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>{preview}</div>
        )}

        {/* Meta row */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 6 }}>
          <span style={{
            fontSize: 11, padding: '1px 6px', borderRadius: 4,
            background: `${cfg.color}14`, color: cfg.color, fontWeight: 500,
          }}>{t(cfg.labelKey)}</span>
          <span style={{ fontSize: 11, color: '#A2ACB5' }}>{date}</span>
          {sourceLabel && <span style={{ fontSize: 11, color: '#A2ACB5' }}>{sourceLabel}</span>}
        </div>
      </div>

      {/* Delete button */}
      <button onClick={handleDelete} disabled={deleting} style={{
        width: 28, height: 28, borderRadius: 6, border: 'none', background: 'transparent',
        cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
        color: '#A2ACB5', flexShrink: 0, transition: 'all 0.15s',
      }}
        onMouseEnter={e => { e.currentTarget.style.background = '#FFF0F0'; e.currentTarget.style.color = '#E53935'; }}
        onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = '#A2ACB5'; }}
      ><Trash2 size={14} /></button>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   MAIN: FavoritesPage
   ═══════════════════════════════════════════════ */
export default function FavoritesPage({ onBack }: { onBack: () => void }) {
  const { currentUser: user } = useIMStore();
  const t = useT();
  const [items, setItems] = useState<FavItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [filterType, setFilterType] = useState('');
  const [showFilter, setShowFilter] = useState(false);
  const [showAdd, setShowAdd] = useState(false);

  const fetchList = useCallback(async (p = 1, type = filterType, keyword = search) => {
    if (!user?.token) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({ page: String(p), pageSize: '20' });
      if (type) params.set('type', type);
      if (keyword) params.set('keyword', keyword);

      const res = await fetch(`/api/user/favorites?${params}`, {
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
  }, [user?.token, filterType, search, items]);

  useEffect(() => { fetchList(1); }, [user?.token, filterType]);

  // Debounced search
  useEffect(() => {
    const t = setTimeout(() => fetchList(1, filterType, search), 300);
    return () => clearTimeout(t);
  }, [search]);

  const handleDelete = async (id: number) => {
    if (!user?.token) return;
    try {
      const res = await fetch('/api/user/favorites', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
        body: JSON.stringify({ id }),
      });
      const d = await res.json();
      if (d.success) {
        setItems(items.filter(i => i.id !== id));
        setTotal(t => t - 1);
        toast.success(t('fav.removed'));
      } else toast.error(d.message);
    } catch { toast.error(t('fav.deleteFail')); }
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
          <Star size={18} style={{ color: '#1BB45B' }} />
          <span style={{ fontSize: 17, fontWeight: 600, color: '#1C2733' }}>{t('fav.title')}</span>
          <span style={{ fontSize: 12, color: '#A2ACB5', marginLeft: 4 }}>{total}</span>
        </div>

        <button onClick={() => setShowAdd(true)} style={{
          width: 32, height: 32, borderRadius: 8, border: 'none',
          background: '#1BB45B', color: '#fff', cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          transition: 'background 0.15s',
        }}
          onMouseEnter={e => { e.currentTarget.style.background = '#149A4C'; }}
          onMouseLeave={e => { e.currentTarget.style.background = '#1BB45B'; }}
        ><Plus size={16} /></button>
      </div>

      {/* Search + Filter bar */}
      <div style={{ padding: '10px 12px 6px', background: '#fff', display: 'flex', gap: 8 }}>
        <div style={{ flex: 1, position: 'relative' }}>
          <Search size={14} style={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', color: '#A2ACB5' }} />
          <input value={search} onChange={e => setSearch(e.target.value)}
            placeholder={t('fav.searchPlaceholder')}
            style={{
              width: '100%', height: 36, paddingLeft: 32, paddingRight: 10, borderRadius: 22,
              border: 'none', background: '#F0F2F5', fontSize: 13, color: '#1C2733', outline: 'none',
            }}
          />
        </div>
        <div style={{ position: 'relative' }}>
          <button onClick={() => setShowFilter(!showFilter)} style={{
            height: 36, padding: '0 12px', borderRadius: 22, border: 'none',
            background: filterType ? '#1BB45B' : '#F0F2F5',
            color: filterType ? '#fff' : '#646A73',
            fontSize: 12, fontWeight: 500, cursor: 'pointer',
            display: 'flex', alignItems: 'center', gap: 4,
          }}>
            <Filter size={13} />
            {filterType ? t(TYPE_MAP[filterType]?.labelKey) : t('fav.filter')}
            <ChevronDown size={12} />
          </button>

          {/* Filter dropdown */}
          {showFilter && (
            <div style={{
              position: 'absolute', top: 42, right: 0, zIndex: 100,
              background: '#fff', borderRadius: 12, boxShadow: '0 4px 20px rgba(0,0,0,0.12)',
              padding: 6, minWidth: 120,
            }}>
              <button onClick={() => { setFilterType(''); setShowFilter(false); }}
                style={{
                  width: '100%', padding: '8px 12px', borderRadius: 8, border: 'none',
                  background: !filterType ? 'rgba(27,180,91,0.08)' : 'transparent',
                  color: !filterType ? '#1BB45B' : '#1C2733',
                  fontSize: 13, textAlign: 'left', cursor: 'pointer',
                }}>{t('fav.filterAll')}</button>
              {Object.entries(TYPE_MAP).map(([k, v]) => (
                <button key={k} onClick={() => { setFilterType(k); setShowFilter(false); }}
                  style={{
                    width: '100%', padding: '8px 12px', borderRadius: 8, border: 'none',
                    background: filterType === k ? `${v.color}14` : 'transparent',
                    color: filterType === k ? v.color : '#1C2733',
                    fontSize: 13, textAlign: 'left', cursor: 'pointer',
                    display: 'flex', alignItems: 'center', gap: 6,
                  }}>{v.icon} {t(v.labelKey)}</button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* List */}
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        {loading && items.length === 0 ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Loader2 size={24} className="animate-spin" style={{ color: '#1BB45B' }} />
          </div>
        ) : items.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 20px' }}>
            <Star size={48} style={{ color: '#E0E3E8', margin: '0 auto 16px' }} />
            <div style={{ fontSize: 15, color: '#646A73', fontWeight: 500 }}>{t('fav.empty')}</div>
            <div style={{ fontSize: 13, color: '#A2ACB5', marginTop: 4 }}>{t('fav.emptyHint')}</div>
          </div>
        ) : (
          <div className="mx-3 mt-2">
            <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
              {items.map((item, idx) => (
                <React.Fragment key={item.id}>
                  {idx > 0 && <div style={{ height: 1, background: 'rgba(0,0,0,0.06)', marginLeft: 64 }} />}
                  <FavCard item={item} onDelete={() => handleDelete(item.id)} />
                </React.Fragment>
              ))}
            </div>

            {/* Load more */}
            {items.length < total && (
              <div style={{ textAlign: 'center', padding: '16px 0' }}>
                <button onClick={() => fetchList(page + 1)} disabled={loading}
                  style={{
                    background: 'none', border: 'none', cursor: 'pointer',
                    fontSize: 13, color: '#1BB45B', fontWeight: 500,
                  }}>
                  {loading ? <Loader2 size={14} className="animate-spin" /> : t('common.loadMore')}
                </button>
              </div>
            )}
          </div>
        )}

        <div style={{ height: 20 }} />
      </div>

      {/* Add dialog */}
      <AddFavoriteDialog open={showAdd} onClose={() => setShowAdd(false)}
        onAdded={() => fetchList(1)} token={user.token} />
    </div>
  );
}
