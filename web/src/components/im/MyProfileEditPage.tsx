'use client';

import React, { useState, useRef, useEffect } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useT } from '@/hooks/use-i18n';
import { getAvatarColor, tagColor } from '@/lib/utils';
import {
  ArrowLeft, Camera, Loader2, Copy, ChevronRight, MapPin,
  Briefcase, Mail, Phone, Tag, User, Pen, X, Check,
  Smartphone, AlertCircle, CheckCircle2, QrCode,
} from 'lucide-react';
import { toast } from 'sonner';
import UserQRCodeDialog from './UserQRCodeDialog';

/* ── Reusable edit dialog (Telegram style modal) ── */
function EditDialog({ open, onClose, title, children }: {
  open: boolean; onClose: () => void; title: string; children: React.ReactNode;
}) {
  useEffect(() => {
    if (!open) return;
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed', inset: 0, zIndex: 1000,
        background: 'rgba(0,0,0,0.45)', backdropFilter: 'blur(2px)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        animation: 'fade-in 0.15s ease-out',
      }}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          width: 400, maxWidth: 'calc(100vw - 32px)',
          background: '#fff', borderRadius: 16,
          boxShadow: '0 8px 32px rgba(0,0,0,0.18)',
          overflow: 'hidden',
        }}
      >
        {/* Dialog header */}
        <div style={{
          padding: '16px 20px 12px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          borderBottom: '1px solid rgba(0,0,0,0.06)',
        }}>
          <span style={{ fontSize: 16, fontWeight: 600, color: '#1C2733' }}>{title}</span>
          <button onClick={onClose} style={{
            width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent',
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: '#A2ACB5', transition: 'all 0.15s',
          }}
            onMouseEnter={e => { e.currentTarget.style.background = 'rgba(0,0,0,0.04)'; e.currentTarget.style.color = '#646A73'; }}
            onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = '#A2ACB5'; }}
          ><X size={16} /></button>
        </div>
        {/* Dialog body */}
        <div style={{ padding: '16px 20px 20px' }}>
          {children}
        </div>
      </div>
    </div>
  );
}

/* ── Pill button for gender / tags ── */
function Pill({ active, children, onClick }: { active: boolean; children: React.ReactNode; onClick: () => void }) {
  return (
    <button onClick={onClick} style={{
      padding: '6px 16px', borderRadius: 20, fontSize: 13, fontWeight: 500, border: 'none', cursor: 'pointer',
      background: active ? '#1BB45B' : '#F0F2F5',
      color: active ? '#fff' : '#646A73',
      transition: 'all 0.2s',
    }}>{children}</button>
  );
}

/* ═══════════════════════════════════════════════
   MAIN COMPONENT
   ═══════════════════════════════════════════════ */
export default function MyProfileEditPage({ onBack }: { onBack: () => void }) {
  const { currentUser: user, updateCurrentUser } = useIMStore();
  const t = useT();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);

  // Edit dialogs
  const [editField, setEditField] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [saving, setSaving] = useState(false);

  // QR code
  const [showQR, setShowQR] = useState(false);

  // Email bind
  const [emailStep, setEmailStep] = useState<'input' | 'verify'>('input');
  const [bindEmail, setBindEmail] = useState('');
  const [emailCode, setEmailCode] = useState('');
  const [emailSending, setEmailSending] = useState(false);
  const [emailCountdown, setEmailCountdown] = useState(0);

  // Tags
  const [editTags, setEditTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');

  useEffect(() => {
    if (emailCountdown <= 0) return;
    const t = setInterval(() => setEmailCountdown(v => v <= 1 ? 0 : v - 1), 1000);
    return () => clearInterval(t);
  }, [emailCountdown > 0]);

  // Refresh user detail on mount
  useEffect(() => {
    if (!user?.token) return;
    fetch('/api/user/detail', { headers: { Authorization: `Bearer ${user.token}` } })
      .then(r => r.json())
      .then(d => { if (d.success) updateCurrentUser(d.data); })
      .catch(() => {});
  }, []);

  if (!user) return null;

  const avatarColor = getAvatarColor(user.name);
  const genderLabel = user.sex === 1 ? t('pe.gender.male') : user.sex === 2 ? t('pe.gender.female') : t('pe.gender.unset');
  const genderEmoji = user.sex === 1 ? '🚹' : user.sex === 2 ? '🚺' : '';

  let parsedTags: string[] = [];
  try { if (user.tags) parsedTags = JSON.parse(user.tags); } catch {}

  /* ── API helpers ── */
  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      // 1. Upload file to get URL
      const fd = new FormData();
      fd.append('file', file);
      const uploadRes = await fetch('/api/user/avatar', { method: 'POST', headers: { Authorization: `Bearer ${user.token}` }, body: fd });
      const uploadData = await uploadRes.json();
      if (!uploadData.success || !uploadData.data?.url) { toast.error(uploadData.message || t('pe.uploadFail')); return; }

      // 2. Save avatar URL to user profile
      const saveRes = await fetch('/api/user/update', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
        body: JSON.stringify({ avatar: uploadData.data.url }),
      });
      const saveData = await saveRes.json();
      if (saveData.success) { updateCurrentUser({ avatar: uploadData.data.url }); toast.success(t('pe.avatarUpdated')); }
      else toast.error(saveData.message || t('pe.saveFail'));
    } catch { toast.error(t('pe.uploadFail')); }
    finally { setUploading(false); if (fileInputRef.current) fileInputRef.current.value = ''; }
  };

  const saveField = async (field: string, value: string | number) => {
    setSaving(true);
    try {
      const res = await fetch('/api/user/update', {
        method: 'PUT', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
        body: JSON.stringify({ [field]: value }),
      });
      const d = await res.json();
      if (d.success) {
        updateCurrentUser({ [field]: value } as any);
        toast.success(t('pe.updated'));
        setEditField(null);
      } else toast.error(d.message || t('pe.updateFail'));
    } catch { toast.error(t('pe.networkError')); }
    finally { setSaving(false); }
  };

  const openEdit = (field: string, currentValue: string) => {
    setEditField(field);
    setEditValue(currentValue);
  };

  const handleSendEmailCode = async () => {
    if (!bindEmail || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(bindEmail)) { toast.error(t('pe.emailFormat')); return; }
    setEmailSending(true);
    try {
      const res = await fetch('/api/auth/send-code', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: bindEmail, type: 'register' }),
      });
      const d = await res.json();
      if (d.success) { setEmailStep('verify'); setEmailCountdown(60); toast.success(t('pe.codeSent')); }
      else toast.error(d.message);
    } catch { toast.error(t('pe.sendFail')); }
    finally { setEmailSending(false); }
  };

  const handleBindEmail = async () => {
    if (!emailCode || emailCode.length !== 6) { toast.error(t('pe.code6')); return; }
    setSaving(true);
    try {
      const res = await fetch('/api/user/email/bind', {
        method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
        body: JSON.stringify({ email: bindEmail, code: emailCode }),
      });
      const d = await res.json();
      if (d.success) {
        updateCurrentUser({ email: bindEmail });
        toast.success(t('pe.emailBound'));
        setEditField(null); setBindEmail(''); setEmailCode(''); setEmailStep('input');
      } else toast.error(d.message);
    } catch { toast.error(t('pe.bindFail')); }
    finally { setSaving(false); }
  };

  const handleCopy = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(t('pe.copied').replace('{label}', label));
  };

  /* ── Info Row (menu-item style) ── */
  const Row = ({ icon, label, value, placeholder, onClick, suffix }: {
    icon: React.ReactNode; label: string; value?: string; placeholder?: string;
    onClick?: () => void; suffix?: React.ReactNode;
  }) => (
    <div
      className="im-profile-menu-item"
      onClick={onClick}
      style={{ padding: '13px 16px' }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1, minWidth: 0 }}>
        <div style={{
          width: 32, height: 32, borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: 'rgba(27,180,91,0.08)', color: '#1BB45B', flexShrink: 0,
        }}>{icon}</div>
        <span style={{ fontSize: 14, color: '#646A73', minWidth: 56, flexShrink: 0, whiteSpace: 'nowrap', marginRight: 8 }}>{label}</span>
        <span style={{
          flex: 1, minWidth: 0, fontSize: 14, fontWeight: 500,
          color: value ? '#1C2733' : '#A2ACB5',
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
        }}>{value || placeholder || t('pe.notSet')}</span>
      </div>
      {suffix || (onClick && <ChevronRight size={16} style={{ color: '#A2ACB5', flexShrink: 0 }} />)}
    </div>
  );

  const Divider = () => <div style={{ height: 1, background: 'rgba(0,0,0,0.06)', marginLeft: 60 }} />;

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>

      {/* ── Top bar ── */}
      <div style={{
        height: 56, display: 'flex', alignItems: 'center', padding: '0 8px', background: '#fff',
        borderBottom: '1px solid rgba(0,0,0,0.06)', flexShrink: 0,
      }}>
        <button onClick={onBack} style={{
          background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4,
          color: '#1BB45B', fontSize: 14, fontWeight: 500, padding: '8px 12px', borderRadius: 8,
          transition: 'background 0.15s',
        }}
          onMouseEnter={e => { e.currentTarget.style.background = 'rgba(27,180,91,0.06)'; }}
          onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; }}
        ><ArrowLeft size={18} /> {t('common.back')}</button>
        <span style={{ flex: 1, textAlign: 'center', fontSize: 17, fontWeight: 600, color: '#1C2733', marginRight: 72 }}>{t('pe.title')}</span>
      </div>

      {/* ── Scroll content ── */}
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">

        {/* ── Hero: Avatar + Name (UserProfileCard style) ── */}
        <div className="mx-4 mt-3">
          <div style={{
            background: '#fff', borderRadius: 16, padding: '24px 20px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}>
            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 20 }}>
              {/* Avatar */}
              <div className="relative shrink-0" style={{ cursor: 'pointer' }} onClick={() => fileInputRef.current?.click()}>
                <div style={{
                  width: 80, height: 80, borderRadius: 16,
                  background: user.avatar ? undefined : avatarColor,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  color: '#fff', fontSize: 32, fontWeight: 600,
                  boxShadow: '0 2px 8px rgba(0,0,0,0.12)',
                  overflow: 'hidden',
                }}>
                  {user.avatar
                    ? <img src={user.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                    : user.name[0]
                  }
                </div>
                <div style={{
                  position: 'absolute', bottom: -2, right: -2,
                  width: 26, height: 26, borderRadius: '50%', background: '#1BB45B',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  border: '2.5px solid #fff', boxShadow: '0 1px 4px rgba(0,0,0,0.15)',
                }}>
                  {uploading ? <Loader2 size={12} className="animate-spin" color="#fff" /> : <Camera size={12} color="#fff" />}
                </div>
                <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={handleAvatarUpload} />
              </div>

              {/* Info */}
              <div style={{ flex: 1, minWidth: 0, paddingTop: 2 }}>
                <div style={{
                  fontSize: 22, fontWeight: 600, color: '#1F2329', lineHeight: 1.3,
                  overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                }}>{user.name}</div>

                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 6 }}>
                  <span style={{ fontSize: 14, color: '#646A73' }}>HiChat ID: {user.id}</span>
                  <button onClick={() => handleCopy(user.id, 'ID')} style={{
                    background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', padding: 2, display: 'flex',
                    transition: 'color 0.15s',
                  }}
                    onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.color = '#1BB45B'; }}
                    onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.color = '#A2ACB5'; }}
                  ><Copy size={13} /></button>
                  <button onClick={() => setShowQR(true)} style={{
                    background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', padding: 2, display: 'flex',
                    transition: 'color 0.15s',
                  }}
                    onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.color = '#1BB45B'; }}
                    onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.color = '#A2ACB5'; }}
                    title={t('pe.qrTitle')}
                  ><QrCode size={14} /></button>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 6, flexWrap: 'wrap' }}>
                  {genderEmoji && <span style={{ fontSize: 14 }}>{genderEmoji} {genderLabel}</span>}
                  {user.region && (
                    <span style={{ display: 'flex', alignItems: 'center', gap: 3, fontSize: 13, color: '#8F959E' }}>
                      <MapPin size={12} /> {user.region}
                    </span>
                  )}
                  {user.occupation && (
                    <span style={{ display: 'flex', alignItems: 'center', gap: 3, fontSize: 13, color: '#8F959E' }}>
                      <Briefcase size={12} /> {user.occupation}
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Signature */}
            {user.introduction && (
              <div style={{
                marginTop: 16, padding: '10px 14px', background: '#F5F7FA', borderRadius: 10,
                fontSize: 14, color: '#646A73', lineHeight: 1.6,
              }}>{user.introduction}</div>
            )}

            {/* Tags */}
            {parsedTags.length > 0 && (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 12 }}>
                {parsedTags.map((t, i) => (
                  <span key={i} style={{
                    padding: '3px 10px', borderRadius: 12, fontSize: 12, fontWeight: 500,
                    background: 'rgba(120,140,160,0.12)', color: '#5b6b7a',
                  }}>{t}</span>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* ── Editable info section ── */}
        <div className="mx-4 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <Row icon={<User size={16} />} label={t('pe.label.nickname')} value={user.name}
              onClick={() => openEdit('name', user.name)} />
            <Divider />
            <Row icon={<span style={{ fontSize: 15 }}>{genderEmoji || '⚪'}</span>} label={t('pe.label.gender')} value={genderLabel}
              onClick={() => { setEditField('sex'); setEditValue(String(user.sex)); }} />
            <Divider />
            <Row icon={<Pen size={16} />} label={t('pe.label.signature')} value={user.introduction}
              placeholder={t('pe.signaturePlaceholder')}
              onClick={() => openEdit('introduction', user.introduction)} />
            <Divider />
            <Row icon={<MapPin size={16} />} label={t('pe.label.region')} value={user.region}
              placeholder={t('pe.regionPlaceholder')}
              onClick={() => openEdit('region', user.region)} />
            <Divider />
            <Row icon={<Briefcase size={16} />} label={t('pe.label.occupation')} value={user.occupation}
              placeholder={t('pe.occupationPlaceholder')}
              onClick={() => openEdit('occupation', user.occupation)} />
            <Divider />
            <Row icon={<Tag size={16} />} label={t('pe.label.tags')}
              value={parsedTags.length > 0 ? parsedTags.join('、') : ''}
              placeholder={t('pe.tagsPlaceholder')}
              onClick={() => { setEditField('tags'); setEditTags([...parsedTags]); setTagInput(''); }} />
          </div>
        </div>

        {/* ── Contact info section ── */}
        <div className="mx-4 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <Row icon={<Smartphone size={16} />} label={t('pe.label.phone')} value={user.phone}
              suffix={<span style={{ fontSize: 12, color: '#4DCD5E', fontWeight: 500 }}>{t('pe.bound')}</span>} />
            <Divider />
            <Row icon={<Mail size={16} />} label={t('pe.label.email')} value={user.email || ''}
              placeholder={t('pe.emailPlaceholder')}
              onClick={() => { setEditField('email'); setBindEmail(user.email || ''); setEmailCode(''); setEmailStep('input'); }}
              suffix={user.email
                ? <span style={{ fontSize: 12, color: '#4DCD5E', fontWeight: 500 }}>{t('pe.bound')}</span>
                : <ChevronRight size={16} style={{ color: '#A2ACB5' }} />
              } />
          </div>
        </div>

        <div style={{ height: 24 }} />
      </div>

      {/* QR Code Dialog */}
      <UserQRCodeDialog
        open={showQR}
        onClose={() => setShowQR(false)}
        userId={user.id}
        userName={user.name}
        userAvatar={user.avatar}
        userIntroduction={user.introduction}
      />

      {/* ═══════════════════════════════════════════════
         EDIT DIALOGS
         ═══════════════════════════════════════════════ */}

      {/* Text field edit (name / introduction / region / occupation) */}
      <EditDialog open={editField === 'name' || editField === 'introduction' || editField === 'region' || editField === 'occupation'}
        onClose={() => setEditField(null)}
        title={editField === 'name' ? t('pe.editNickname') : editField === 'introduction' ? t('pe.editSignature') : editField === 'region' ? t('pe.editRegion') : t('pe.editOccupation')}>
        {editField === 'introduction' ? (
          <textarea
            autoFocus value={editValue} onChange={e => setEditValue(e.target.value)}
            maxLength={100} rows={3} placeholder={t('pe.signatureDialogPlaceholder')}
            style={{
              width: '100%', padding: '10px 14px', borderRadius: 10, border: '1.5px solid #E0E3E8',
              fontSize: 14, color: '#1C2733', background: '#FAFBFC', outline: 'none', resize: 'none',
              fontFamily: 'inherit', lineHeight: 1.6, transition: 'border-color 0.2s',
            }}
            onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
            onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
          />
        ) : (
          <input
            autoFocus value={editValue} onChange={e => setEditValue(e.target.value)}
            maxLength={64}
            placeholder={editField === 'name' ? t('pe.nicknameInputPlaceholder') : editField === 'region' ? t('pe.regionInputPlaceholder') : t('pe.occupationInputPlaceholder')}
            style={{
              width: '100%', height: 44, padding: '0 14px', borderRadius: 10, border: '1.5px solid #E0E3E8',
              fontSize: 14, color: '#1C2733', background: '#FAFBFC', outline: 'none',
              transition: 'border-color 0.2s',
            }}
            onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
            onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
            onKeyDown={e => { if (e.key === 'Enter' && editField) saveField(editField, editValue); }}
          />
        )}
        {editField === 'introduction' && (
          <div style={{ textAlign: 'right', fontSize: 12, color: '#A2ACB5', marginTop: 6 }}>{editValue.length}/100</div>
        )}
        <button className="hc-btn-primary" disabled={saving || !editValue.trim()}
          onClick={() => editField && saveField(editField, editValue.trim())}
          style={{ marginTop: 16, padding: '12px 24px', borderRadius: 22 }}>
          {saving ? <Loader2 size={16} className="animate-spin" /> : t('common.save')}
        </button>
      </EditDialog>

      {/* Gender select */}
      <EditDialog open={editField === 'sex'} onClose={() => setEditField(null)} title={t('pe.selectGender')}>
        <div style={{ display: 'flex', gap: 10, justifyContent: 'center', marginBottom: 20 }}>
          <Pill active={editValue === '0'} onClick={() => setEditValue('0')}>{t('pe.gender.unset')}</Pill>
          <Pill active={editValue === '1'} onClick={() => setEditValue('1')}>🚹 {t('pe.gender.male')}</Pill>
          <Pill active={editValue === '2'} onClick={() => setEditValue('2')}>🚺 {t('pe.gender.female')}</Pill>
        </div>
        <button className="hc-btn-primary" disabled={saving}
          onClick={() => saveField('sex', Number(editValue))}
          style={{ padding: '12px 24px', borderRadius: 22 }}>
          {saving ? <Loader2 size={16} className="animate-spin" /> : t('common.confirm')}
        </button>
      </EditDialog>

      {/* Tags editor */}
      <EditDialog open={editField === 'tags'} onClose={() => setEditField(null)} title={t('pe.editTags')}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, minHeight: 36, marginBottom: 12 }}>
          {editTags.map((t, i) => (
            <span key={i} style={{
              display: 'inline-flex', alignItems: 'center', gap: 4,
              padding: '5px 12px', borderRadius: 16, fontSize: 13, fontWeight: 500,
              background: tagColor(t).b, color: tagColor(t).c,
            }}>
              {t}
              <button onClick={() => setEditTags(editTags.filter((_, j) => j !== i))} style={{
                background: 'none', border: 'none', cursor: 'pointer', color: tagColor(t).c, padding: 0,
                display: 'flex', opacity: 0.6, transition: 'opacity 0.15s',
              }}
                onMouseEnter={e => { e.currentTarget.style.opacity = '1'; }}
                onMouseLeave={e => { e.currentTarget.style.opacity = '0.6'; }}
              ><X size={14} /></button>
            </span>
          ))}
          {editTags.length === 0 && <span style={{ fontSize: 13, color: '#A2ACB5', lineHeight: '28px' }}>{t('pe.noTags')}</span>}
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <input
            value={tagInput} onChange={e => setTagInput(e.target.value)}
            onKeyDown={e => {
              if (e.key === 'Enter' && tagInput.trim() && !editTags.includes(tagInput.trim()) && editTags.length < 10) {
                setEditTags([...editTags, tagInput.trim()]); setTagInput('');
              }
            }}
            placeholder={t('pe.tagInputPlaceholder')} maxLength={20}
            style={{
              flex: 1, height: 40, padding: '0 14px', borderRadius: 10, border: '1.5px solid #E0E3E8',
              fontSize: 14, color: '#1C2733', background: '#FAFBFC', outline: 'none',
              transition: 'border-color 0.2s',
            }}
            onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
            onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
          />
          <button onClick={() => {
            if (tagInput.trim() && !editTags.includes(tagInput.trim()) && editTags.length < 10) {
              setEditTags([...editTags, tagInput.trim()]); setTagInput('');
            }
          }} style={{
            height: 40, padding: '0 16px', borderRadius: 10, fontSize: 13, fontWeight: 600,
            background: tagInput.trim() ? '#1BB45B' : '#F0F2F5',
            color: tagInput.trim() ? '#fff' : '#A2ACB5',
            border: 'none', cursor: 'pointer', transition: 'all 0.2s',
          }}>{t('pe.add')}</button>
        </div>
        <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 8 }}>{t('pe.maxTags').replace('{count}', String(editTags.length))}</div>
        <button className="hc-btn-primary" disabled={saving}
          onClick={async () => {
            await saveField('tags', JSON.stringify(editTags));
          }}
          style={{ marginTop: 12, padding: '12px 24px', borderRadius: 22 }}>
          {saving ? <Loader2 size={16} className="animate-spin" /> : t('common.save')}
        </button>
      </EditDialog>

      {/* Email bind dialog */}
      <EditDialog open={editField === 'email'} onClose={() => { setEditField(null); setEmailStep('input'); }} title={user.email ? t('pe.changeEmail') : t('pe.bindEmail')}>
        {emailStep === 'input' ? (
          <>
            {user.email && (
              <div style={{ fontSize: 13, color: '#646A73', marginBottom: 12, display: 'flex', alignItems: 'center', gap: 6 }}>
                <CheckCircle2 size={14} color="#4DCD5E" /> {t('pe.currentEmail').replace('{email}', user.email)}
              </div>
            )}
            <input
              autoFocus value={bindEmail} onChange={e => setBindEmail(e.target.value)}
              placeholder={t('pe.emailInputPlaceholder')} type="email"
              style={{
                width: '100%', height: 44, padding: '0 14px', borderRadius: 10, border: '1.5px solid #E0E3E8',
                fontSize: 14, color: '#1C2733', background: '#FAFBFC', outline: 'none',
                transition: 'border-color 0.2s',
              }}
              onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
            />
            <button className="hc-btn-primary" disabled={emailSending || !bindEmail || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(bindEmail)}
              onClick={handleSendEmailCode}
              style={{ marginTop: 16, padding: '12px 24px', borderRadius: 22 }}>
              {emailSending ? <Loader2 size={16} className="animate-spin" /> : t('pe.sendCode')}
            </button>
          </>
        ) : (
          <>
            <div style={{ fontSize: 13, color: '#646A73', marginBottom: 12 }}>
              {t('pe.codeSentTo').replace('{email}', bindEmail)}
            </div>
            <input
              autoFocus value={emailCode} onChange={e => setEmailCode(e.target.value.replace(/\D/g, ''))}
              placeholder={t('pe.codePlaceholder')} maxLength={6}
              style={{
                width: '100%', height: 44, padding: '0 14px', borderRadius: 10, border: '1.5px solid #E0E3E8',
                fontSize: 14, color: '#1C2733', background: '#FAFBFC', outline: 'none',
                letterSpacing: 4, textAlign: 'center', fontWeight: 600,
                transition: 'border-color 0.2s',
              }}
              onFocus={e => { e.currentTarget.style.borderColor = '#1BB45B'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
              onKeyDown={e => { if (e.key === 'Enter' && emailCode.length === 6) handleBindEmail(); }}
            />
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
              <button onClick={() => setEmailStep('input')} style={{
                background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, color: '#1BB45B',
              }}>{t('pe.changeEmailLink')}</button>
              <button disabled={emailCountdown > 0} onClick={handleSendEmailCode} style={{
                background: 'none', border: 'none', cursor: emailCountdown > 0 ? 'default' : 'pointer',
                fontSize: 13, color: emailCountdown > 0 ? '#A2ACB5' : '#1BB45B',
              }}>{emailCountdown > 0 ? t('pe.resendIn').replace('{sec}', String(emailCountdown)) : t('pe.resend')}</button>
            </div>
            <button className="hc-btn-primary" disabled={saving || emailCode.length !== 6}
              onClick={handleBindEmail}
              style={{ marginTop: 12, padding: '12px 24px', borderRadius: 22 }}>
              {saving ? <Loader2 size={16} className="animate-spin" /> : t('pe.confirmBind')}
            </button>
          </>
        )}
      </EditDialog>
    </div>
  );
}
