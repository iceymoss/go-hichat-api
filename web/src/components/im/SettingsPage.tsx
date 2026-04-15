'use client';

import React, { useState, useEffect } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useSettingsStore, type ThemeMode } from '@/lib/settings-store';
import { useT } from '@/hooks/use-i18n';
import {
  ArrowLeft, ChevronRight, Lock, Bell, Shield, Moon, BookOpen,
  Key, Mail, Smartphone, UserX, Loader2, Eye, EyeOff,
  CheckCircle2, AlertCircle, X, Sun, Monitor, Globe, Type,
  Trash2, Search, Users,
} from 'lucide-react';
import { toast } from 'sonner';

/* ═══════════════════════════════════════════════
   Shared UI Components
   ═══════════════════════════════════════════════ */

function PageHeader({ title, onBack }: { title: string; onBack: () => void }) {
  return (
    <div style={{
      height: 56, display: 'flex', alignItems: 'center', padding: '0 8px',
      background: '#fff', borderBottom: '1px solid rgba(0,0,0,0.06)', flexShrink: 0,
    }}>
      <button onClick={onBack} style={{
        background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4,
        color: '#3390EC', fontSize: 14, fontWeight: 500, padding: '8px 12px', borderRadius: 8,
        transition: 'background 0.15s',
      }}
        onMouseEnter={e => { e.currentTarget.style.background = 'rgba(51,144,236,0.06)'; }}
        onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; }}
      ><ArrowLeft size={18} /></button>
      <span style={{ flex: 1, textAlign: 'center', fontSize: 17, fontWeight: 600, color: '#1C2733', marginRight: 52 }}>{title}</span>
    </div>
  );
}

function MenuItem({ icon, label, value, onClick, danger }: {
  icon: React.ReactNode; label: string; value?: string; onClick?: () => void; danger?: boolean;
}) {
  return (
    <div className="im-profile-menu-item" onClick={onClick} style={{ padding: '13px 16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1 }}>
        <div style={{
          width: 32, height: 32, borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: danger ? 'rgba(229,57,53,0.08)' : 'rgba(51,144,236,0.08)',
          color: danger ? '#E53935' : '#3390EC', flexShrink: 0,
        }}>{icon}</div>
        <span style={{ fontSize: 14, color: danger ? '#E53935' : '#1C2733' }}>{label}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        {value && <span style={{ fontSize: 13, color: '#A2ACB5' }}>{value}</span>}
        <ChevronRight size={16} style={{ color: '#A2ACB5' }} />
      </div>
    </div>
  );
}

function ToggleRow({ icon, label, desc, checked, onChange }: {
  icon: React.ReactNode; label: string; desc?: string; checked: boolean; onChange: (v: boolean) => void;
}) {
  return (
    <div style={{ padding: '13px 16px', display: 'flex', alignItems: 'center', gap: 12 }}>
      <div style={{
        width: 32, height: 32, borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center',
        background: 'rgba(51,144,236,0.08)', color: '#3390EC', flexShrink: 0,
      }}>{icon}</div>
      <div style={{ flex: 1 }}>
        <div style={{ fontSize: 14, color: '#1C2733' }}>{label}</div>
        {desc && <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 2 }}>{desc}</div>}
      </div>
      <button onClick={() => onChange(!checked)} style={{
        width: 44, height: 24, borderRadius: 12, border: 'none', cursor: 'pointer',
        background: checked ? '#3390EC' : '#D0D5DD', position: 'relative',
        transition: 'background 0.2s',
      }}>
        <div style={{
          width: 20, height: 20, borderRadius: '50%', background: '#fff',
          position: 'absolute', top: 2,
          left: checked ? 22 : 2,
          transition: 'left 0.2s',
          boxShadow: '0 1px 3px rgba(0,0,0,0.15)',
        }} />
      </button>
    </div>
  );
}

function RadioRow({ label, icon, selected, onClick }: {
  label: string; icon: React.ReactNode; selected: boolean; onClick: () => void;
}) {
  return (
    <div onClick={onClick} className="im-profile-menu-item" style={{ padding: '13px 16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1 }}>
        <div style={{
          width: 32, height: 32, borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: selected ? 'rgba(51,144,236,0.12)' : 'rgba(0,0,0,0.03)',
          color: selected ? '#3390EC' : '#A2ACB5', flexShrink: 0,
        }}>{icon}</div>
        <span style={{ fontSize: 14, color: '#1C2733' }}>{label}</span>
      </div>
      {selected && <CheckCircle2 size={18} style={{ color: '#3390EC' }} />}
    </div>
  );
}

function Dialog({ open, onClose, title, children }: {
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
    <div onClick={onClose} style={{
      position: 'fixed', inset: 0, zIndex: 1000, background: 'rgba(0,0,0,0.45)',
      backdropFilter: 'blur(2px)', display: 'flex', alignItems: 'center', justifyContent: 'center',
      animation: 'fade-in 0.15s ease-out',
    }}>
      <div onClick={e => e.stopPropagation()} style={{
        width: 400, maxWidth: 'calc(100vw - 32px)', background: '#fff', borderRadius: 16,
        boxShadow: '0 8px 32px rgba(0,0,0,0.18)', overflow: 'hidden',
      }}>
        <div style={{
          padding: '16px 20px 12px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          borderBottom: '1px solid rgba(0,0,0,0.06)',
        }}>
          <span style={{ fontSize: 16, fontWeight: 600, color: '#1C2733' }}>{title}</span>
          <button onClick={onClose} style={{
            width: 28, height: 28, borderRadius: '50%', border: 'none', background: 'transparent',
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#A2ACB5',
          }}><X size={16} /></button>
        </div>
        <div style={{ padding: '16px 20px 20px' }}>{children}</div>
      </div>
    </div>
  );
}

const Divider = () => <div style={{ height: 1, background: 'rgba(0,0,0,0.06)', marginLeft: 60 }} />;
const SectionGap = () => <div style={{ height: 10 }} />;

const inputStyle: React.CSSProperties = {
  width: '100%', height: 44, padding: '0 14px', borderRadius: 10,
  border: '1.5px solid #E0E3E8', fontSize: 14, color: '#1C2733',
  background: '#FAFBFC', outline: 'none',
};

/* ═══════════════════════════════════════════════
   1. Account & Security Page
   ═══════════════════════════════════════════════ */

function AccountSecurityPage({ onBack }: { onBack: () => void }) {
  const { currentUser: user, updateCurrentUser, logout } = useIMStore();
  const [dialog, setDialog] = useState<'password' | 'email' | 'delete' | null>(null);

  // Change password state
  const [oldPwd, setOldPwd] = useState('');
  const [newPwd, setNewPwd] = useState('');
  const [newPwd2, setNewPwd2] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  // Email bind state
  const [email, setEmail] = useState('');
  const [emailCode, setEmailCode] = useState('');
  const [emailStep, setEmailStep] = useState<'input' | 'verify'>('input');
  const [emailSending, setEmailSending] = useState(false);
  const [countdown, setCountdown] = useState(0);

  // Delete account 3-step state
  const [deleteStep, setDeleteStep] = useState<'password' | 'sms' | 'confirm'>('password');
  const [deletePwd, setDeletePwd] = useState('');
  const [deleteSmsCode, setDeleteSmsCode] = useState('');
  const [deleteCountdown, setDeleteCountdown] = useState(0);
  const [deleteError, setDeleteError] = useState('');

  useEffect(() => {
    if (countdown <= 0) return;
    const t = setInterval(() => setCountdown(v => v <= 1 ? 0 : v - 1), 1000);
    return () => clearInterval(t);
  }, [countdown > 0]);

  useEffect(() => {
    if (deleteCountdown <= 0) return;
    const t = setInterval(() => setDeleteCountdown(v => v <= 1 ? 0 : v - 1), 1000);
    return () => clearInterval(t);
  }, [deleteCountdown > 0]);

  if (!user) return null;

  const handleChangePassword = async () => {
    if (!newPwd || newPwd.length < 8) { setError('新密码至少8位'); return; }
    if (newPwd !== newPwd2) { setError('两次密码不一致'); return; }
    setSaving(true); setError('');
    try {
      // Send phone code first
      const codeRes = await fetch('/api/auth/send-code', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: user.phone, type: 'register' }),
      });
      const codeData = await codeRes.json();
      if (!codeData.success) { setError(codeData.message); setSaving(false); return; }

      // For now, use the reset_pwd endpoint (requires phone + code)
      // In production this would be a proper change-password flow
      toast.info('密码修改需要短信验证，请通过"忘记密码"流程重置');
      setDialog(null);
    } catch { setError('操作失败'); }
    finally { setSaving(false); }
  };

  const handleSendEmailCode = async () => {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { toast.error('邮箱格式不正确'); return; }
    setEmailSending(true);
    try {
      const res = await fetch('/api/auth/send-code', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: email, type: 'register' }),
      });
      const d = await res.json();
      if (d.success) { setEmailStep('verify'); setCountdown(60); toast.success('验证码已发送'); }
      else toast.error(d.message);
    } catch { toast.error('发送失败'); }
    finally { setEmailSending(false); }
  };

  const handleBindEmail = async () => {
    if (emailCode.length !== 6) { toast.error('请输入6位验证码'); return; }
    setSaving(true);
    try {
      const res = await fetch('/api/user/email/bind', {
        method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${user.token}` },
        body: JSON.stringify({ email, code: emailCode }),
      });
      const d = await res.json();
      if (d.success) {
        updateCurrentUser({ email });
        toast.success('邮箱绑定成功');
        setDialog(null); setEmail(''); setEmailCode(''); setEmailStep('input');
      } else toast.error(d.message);
    } catch { toast.error('绑定失败'); }
    finally { setSaving(false); }
  };

  // Step 1: Verify password (try login with same phone+password)
  const handleDeleteVerifyPassword = async () => {
    if (!deletePwd) { setDeleteError('请输入密码'); return; }
    setSaving(true); setDeleteError('');
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: user.phone, password: deletePwd }),
      });
      const d = await res.json();
      if (d.success) {
        // Password correct → send SMS code
        const codeRes = await fetch('/api/auth/send-code', {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ target: user.phone, type: 'register' }),
        });
        const codeData = await codeRes.json();
        if (codeData.success) {
          setDeleteStep('sms');
          setDeleteCountdown(60);
          toast.success('验证码已发送');
        } else setDeleteError(codeData.message || '发送验证码失败');
      } else {
        setDeleteError('密码错误');
      }
    } catch { setDeleteError('验证失败'); }
    finally { setSaving(false); }
  };

  // Step 2: Verify SMS code
  const handleDeleteVerifySms = async () => {
    if (deleteSmsCode.length !== 6) { setDeleteError('请输入6位验证码'); return; }
    setSaving(true); setDeleteError('');
    try {
      const res = await fetch('/api/auth/verify-code', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: user.phone, code: deleteSmsCode }),
      });
      const d = await res.json();
      if (d.success) {
        setDeleteStep('confirm');
      } else setDeleteError(d.message || '验证码错误');
    } catch { setDeleteError('验证失败'); }
    finally { setSaving(false); }
  };

  // Step 3: Final confirm delete
  const handleDeleteAccount = async () => {
    setSaving(true);
    try {
      const res = await fetch('/api/auth/logout', {
        method: 'DELETE', headers: { Authorization: `Bearer ${user.token}` },
      });
      const d = await res.json();
      if (d.success) { toast.success('账号已注销'); logout(); }
      else toast.error(d.message);
    } catch { toast.error('注销失败'); }
    finally { setSaving(false); }
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title="账号与安全" onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <MenuItem icon={<Smartphone size={16} />} label="手机号" value={user.phone} />
            <Divider />
            <MenuItem icon={<Mail size={16} />} label="邮箱" value={user.email || '未绑定'}
              onClick={() => { setDialog('email'); setEmail(user.email || ''); setEmailCode(''); setEmailStep('input'); }} />
            <Divider />
            <MenuItem icon={<Key size={16} />} label="修改密码" onClick={() => {
              setDialog('password'); setOldPwd(''); setNewPwd(''); setNewPwd2(''); setError('');
            }} />
          </div>
        </div>

        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <MenuItem icon={<UserX size={16} />} label="注销账号" onClick={() => {
              setDialog('delete'); setDeleteStep('password'); setDeletePwd(''); setDeleteSmsCode(''); setDeleteError('');
            }} danger />
          </div>
        </div>
        <div style={{ height: 24 }} />
      </div>

      {/* Change Password Dialog */}
      <Dialog open={dialog === 'password'} onClose={() => setDialog(null)} title="修改密码">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ position: 'relative' }}>
            <input type={showPwd ? 'text' : 'password'} value={newPwd} onChange={e => { setNewPwd(e.target.value); setError(''); }}
              placeholder="输入新密码（8位以上）" style={inputStyle}
              onFocus={e => { e.currentTarget.style.borderColor = '#3390EC'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
            />
          </div>
          <input type={showPwd ? 'text' : 'password'} value={newPwd2} onChange={e => { setNewPwd2(e.target.value); setError(''); }}
            placeholder="确认新密码" style={inputStyle}
            onFocus={e => { e.currentTarget.style.borderColor = '#3390EC'; }}
            onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
          />
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <button onClick={() => setShowPwd(!showPwd)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', display: 'flex', alignItems: 'center', gap: 4, fontSize: 12 }}>
              {showPwd ? <EyeOff size={14} /> : <Eye size={14} />} {showPwd ? '隐藏' : '显示'}密码
            </button>
          </div>
          {error && <div style={{ fontSize: 13, color: '#E53935', display: 'flex', alignItems: 'center', gap: 4 }}><AlertCircle size={14} />{error}</div>}
          <div style={{ fontSize: 12, color: '#A2ACB5', lineHeight: 1.5 }}>
            修改密码需要短信验证。点击保存后将发送验证码到绑定手机号。
          </div>
          <button className="hc-btn-primary" disabled={saving || !newPwd || !newPwd2}
            onClick={handleChangePassword} style={{ marginTop: 4, padding: '12px 24px', borderRadius: 22 }}>
            {saving ? <Loader2 size={16} className="animate-spin" /> : '保存'}
          </button>
        </div>
      </Dialog>

      {/* Bind Email Dialog */}
      <Dialog open={dialog === 'email'} onClose={() => setDialog(null)} title={user.email ? '更换邮箱' : '绑定邮箱'}>
        {emailStep === 'input' ? (
          <>
            {user.email && (
              <div style={{ fontSize: 13, color: '#646A73', marginBottom: 12, display: 'flex', alignItems: 'center', gap: 6 }}>
                <CheckCircle2 size={14} color="#4DCD5E" /> 当前: {user.email}
              </div>
            )}
            <input autoFocus value={email} onChange={e => setEmail(e.target.value)}
              placeholder="请输入邮箱地址" type="email" style={inputStyle}
              onFocus={e => { e.currentTarget.style.borderColor = '#3390EC'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
            />
            <button className="hc-btn-primary" disabled={emailSending || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)}
              onClick={handleSendEmailCode} style={{ marginTop: 16, padding: '12px 24px', borderRadius: 22 }}>
              {emailSending ? <Loader2 size={16} className="animate-spin" /> : '发送验证码'}
            </button>
          </>
        ) : (
          <>
            <div style={{ fontSize: 13, color: '#646A73', marginBottom: 12 }}>
              验证码已发送至 <span style={{ color: '#3390EC', fontWeight: 500 }}>{email}</span>
            </div>
            <input autoFocus value={emailCode} onChange={e => setEmailCode(e.target.value.replace(/\D/g, ''))}
              placeholder="请输入6位验证码" maxLength={6}
              style={{ ...inputStyle, letterSpacing: 4, textAlign: 'center', fontWeight: 600 }}
              onFocus={e => { e.currentTarget.style.borderColor = '#3390EC'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
              onKeyDown={e => { if (e.key === 'Enter' && emailCode.length === 6) handleBindEmail(); }}
            />
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 8 }}>
              <button onClick={() => setEmailStep('input')} style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 13, color: '#3390EC' }}>更换邮箱</button>
              <button disabled={countdown > 0} onClick={handleSendEmailCode} style={{
                background: 'none', border: 'none', cursor: countdown > 0 ? 'default' : 'pointer',
                fontSize: 13, color: countdown > 0 ? '#A2ACB5' : '#3390EC',
              }}>{countdown > 0 ? `${countdown}s` : '重新发送'}</button>
            </div>
            <button className="hc-btn-primary" disabled={saving || emailCode.length !== 6}
              onClick={handleBindEmail} style={{ marginTop: 12, padding: '12px 24px', borderRadius: 22 }}>
              {saving ? <Loader2 size={16} className="animate-spin" /> : '确认绑定'}
            </button>
          </>
        )}
      </Dialog>

      {/* Delete Account Dialog — 3-step verification */}
      <Dialog open={dialog === 'delete'} onClose={() => setDialog(null)} title="注销账号">
        {/* Step indicator */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginBottom: 20 }}>
          {['验证密码', '短信验证', '确认注销'].map((label, i) => {
            const stepKeys = ['password', 'sms', 'confirm'] as const;
            const current = stepKeys.indexOf(deleteStep);
            const done = i < current;
            const active = i === current;
            return (
              <div key={label} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                {i > 0 && <div style={{ width: 20, height: 1, background: done || active ? '#3390EC' : '#E0E3E8' }} />}
                <div style={{
                  width: 22, height: 22, borderRadius: '50%', fontSize: 11, fontWeight: 600,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  background: done ? '#3390EC' : active ? '#3390EC' : '#F0F2F5',
                  color: done || active ? '#fff' : '#A2ACB5',
                }}>{done ? '✓' : i + 1}</div>
                <span style={{ fontSize: 12, color: active ? '#1C2733' : '#A2ACB5', fontWeight: active ? 600 : 400 }}>{label}</span>
              </div>
            );
          })}
        </div>

        {/* Step 1: Password */}
        {deleteStep === 'password' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div style={{ fontSize: 13, color: '#646A73', lineHeight: 1.5 }}>
              为确保是本人操作，请输入当前账号密码
            </div>
            <input type="password" value={deletePwd} onChange={e => { setDeletePwd(e.target.value); setDeleteError(''); }}
              placeholder="请输入密码" autoFocus style={inputStyle}
              onFocus={e => { e.currentTarget.style.borderColor = '#3390EC'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
              onKeyDown={e => { if (e.key === 'Enter') handleDeleteVerifyPassword(); }}
            />
            {deleteError && <div style={{ fontSize: 13, color: '#E53935', display: 'flex', alignItems: 'center', gap: 4 }}><AlertCircle size={14} />{deleteError}</div>}
            <div style={{ display: 'flex', gap: 10, marginTop: 4 }}>
              <button onClick={() => setDialog(null)} style={{
                flex: 1, height: 44, borderRadius: 22, border: '1.5px solid #E0E3E8',
                background: '#fff', color: '#1C2733', fontSize: 14, fontWeight: 500, cursor: 'pointer',
              }}>取消</button>
              <button onClick={handleDeleteVerifyPassword} disabled={saving || !deletePwd} style={{
                flex: 1, height: 44, borderRadius: 22, border: 'none',
                background: '#E53935', color: '#fff', fontSize: 14, fontWeight: 600, cursor: 'pointer',
                opacity: saving || !deletePwd ? 0.6 : 1,
              }}>{saving ? <Loader2 size={16} className="animate-spin" /> : '下一步'}</button>
            </div>
          </div>
        )}

        {/* Step 2: SMS Code */}
        {deleteStep === 'sms' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div style={{ fontSize: 13, color: '#646A73', lineHeight: 1.5 }}>
              验证码已发送至 <span style={{ color: '#3390EC', fontWeight: 500 }}>{user.phone}</span>
            </div>
            <input value={deleteSmsCode} onChange={e => { setDeleteSmsCode(e.target.value.replace(/\D/g, '')); setDeleteError(''); }}
              placeholder="请输入6位验证码" maxLength={6} autoFocus
              style={{ ...inputStyle, letterSpacing: 4, textAlign: 'center', fontWeight: 600 }}
              onFocus={e => { e.currentTarget.style.borderColor = '#3390EC'; }}
              onBlur={e => { e.currentTarget.style.borderColor = '#E0E3E8'; }}
              onKeyDown={e => { if (e.key === 'Enter' && deleteSmsCode.length === 6) handleDeleteVerifySms(); }}
            />
            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <button disabled={deleteCountdown > 0} onClick={async () => {
                const res = await fetch('/api/auth/send-code', {
                  method: 'POST', headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify({ target: user.phone, type: 'register' }),
                });
                const d = await res.json();
                if (d.success) { setDeleteCountdown(60); toast.success('已重新发送'); }
              }} style={{
                background: 'none', border: 'none', fontSize: 13, cursor: deleteCountdown > 0 ? 'default' : 'pointer',
                color: deleteCountdown > 0 ? '#A2ACB5' : '#3390EC',
              }}>{deleteCountdown > 0 ? `${deleteCountdown}s 后重发` : '重新发送'}</button>
            </div>
            {deleteError && <div style={{ fontSize: 13, color: '#E53935', display: 'flex', alignItems: 'center', gap: 4 }}><AlertCircle size={14} />{deleteError}</div>}
            <div style={{ display: 'flex', gap: 10, marginTop: 4 }}>
              <button onClick={() => setDeleteStep('password')} style={{
                flex: 1, height: 44, borderRadius: 22, border: '1.5px solid #E0E3E8',
                background: '#fff', color: '#1C2733', fontSize: 14, fontWeight: 500, cursor: 'pointer',
              }}>上一步</button>
              <button onClick={handleDeleteVerifySms} disabled={saving || deleteSmsCode.length !== 6} style={{
                flex: 1, height: 44, borderRadius: 22, border: 'none',
                background: '#E53935', color: '#fff', fontSize: 14, fontWeight: 600, cursor: 'pointer',
                opacity: saving || deleteSmsCode.length !== 6 ? 0.6 : 1,
              }}>{saving ? <Loader2 size={16} className="animate-spin" /> : '下一步'}</button>
            </div>
          </div>
        )}

        {/* Step 3: Final Confirm */}
        {deleteStep === 'confirm' && (
          <div>
            <div style={{ textAlign: 'center', padding: '8px 0 20px' }}>
              <UserX size={40} style={{ color: '#E53935', margin: '0 auto 12px' }} />
              <div style={{ fontSize: 15, color: '#1C2733', fontWeight: 600 }}>身份验证通过</div>
              <div style={{ fontSize: 13, color: '#8F959E', marginTop: 8, lineHeight: 1.6 }}>
                注销后你的所有数据将被<strong style={{ color: '#E53935' }}>永久删除</strong>，<br />
                包括好友关系、聊天记录、收藏、表情包等。<br />
                <strong>此操作不可撤销！</strong>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 10 }}>
              <button onClick={() => setDialog(null)} style={{
                flex: 1, height: 44, borderRadius: 22, border: '1.5px solid #E0E3E8',
                background: '#fff', color: '#1C2733', fontSize: 14, fontWeight: 500, cursor: 'pointer',
              }}>取消</button>
              <button onClick={handleDeleteAccount} disabled={saving} style={{
                flex: 1, height: 44, borderRadius: 22, border: 'none',
                background: '#E53935', color: '#fff', fontSize: 14, fontWeight: 600, cursor: 'pointer',
              }}>{saving ? <Loader2 size={16} className="animate-spin" /> : '确认注销'}</button>
            </div>
          </div>
        )}
      </Dialog>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   2. Notification Settings
   ═══════════════════════════════════════════════ */

function NotificationSettingsPage({ onBack }: { onBack: () => void }) {
  const s = useSettingsStore();
  const token = useIMStore(st => st.currentUser?.token) || '';
  const t = useT();
  const u = (key: string, val: any) => s.updateSetting(key, val, token);
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title={t('settings.notification')} onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <ToggleRow icon={<Bell size={16} />} label={t('notify.enable')} checked={s.notifyEnabled}
              onChange={v => u('notifyEnabled', v)} />
            <Divider />
            <ToggleRow icon={<span style={{ fontSize: 16 }}>🔊</span>} label={t('notify.sound')} desc={t('notify.soundDesc')}
              checked={s.notifySound} onChange={v => u('notifySound', v)} />
            <Divider />
            <ToggleRow icon={<span style={{ fontSize: 16 }}>📳</span>} label={t('notify.vibrate')} desc={t('notify.vibrateDesc')}
              checked={s.notifyVibrate} onChange={v => u('notifyVibrate', v)} />
            <Divider />
            <ToggleRow icon={<Eye size={16} />} label={t('notify.preview')} desc={t('notify.previewDesc')}
              checked={s.notifyPreview} onChange={v => u('notifyPreview', v)} />
          </div>
        </div>
        <div style={{ height: 24 }} />
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   3. Privacy Settings
   ═══════════════════════════════════════════════ */

function PrivacySettingsPage({ onBack }: { onBack: () => void }) {
  const s = useSettingsStore();
  const token = useIMStore(st => st.currentUser?.token) || '';
  const t = useT();
  const u = (key: string, val: any) => s.updateSetting(key, val, token);
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title={t('settings.privacy')} onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        <div className="mx-3 mt-3">
          <div style={{ fontSize: 12, color: '#A2ACB5', fontWeight: 500, padding: '0 4px 6px' }}>{t('privacy.addMethods')}</div>
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <ToggleRow icon={<Smartphone size={16} />} label={t('privacy.byPhone')}
              checked={s.allowSearchByPhone} onChange={v => u('allowSearchByPhone', v)} />
            <Divider />
            <ToggleRow icon={<Search size={16} />} label={t('privacy.byId')}
              checked={s.allowSearchById} onChange={v => u('allowSearchById', v)} />
          </div>
        </div>

        <div className="mx-3 mt-3">
          <div style={{ fontSize: 12, color: '#A2ACB5', fontWeight: 500, padding: '0 4px 6px' }}>{t('privacy.moments')}</div>
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <RadioRow icon={<Users size={16} />} label={t('privacy.allVisible')} selected={s.momentVisibility === 'all'}
              onClick={() => u('momentVisibility', 'all')} />
            <Divider />
            <RadioRow icon={<Shield size={16} />} label={t('privacy.friendsOnly')} selected={s.momentVisibility === 'friends'}
              onClick={() => u('momentVisibility', 'friends')} />
            <Divider />
            <RadioRow icon={<Lock size={16} />} label={t('privacy.privateOnly')} selected={s.momentVisibility === 'private'}
              onClick={() => u('momentVisibility', 'private')} />
          </div>
        </div>
        <div style={{ height: 24 }} />
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   4. Dark Mode Settings
   ═══════════════════════════════════════════════ */

function DarkModeSettingsPage({ onBack }: { onBack: () => void }) {
  const { themeMode, updateSetting } = useSettingsStore();
  const token = useIMStore(st => st.currentUser?.token) || '';
  const t = useT();
  const set = (mode: ThemeMode) => { updateSetting('themeMode', mode, token); };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title={t('settings.darkMode')} onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <RadioRow icon={<Monitor size={16} />} label={t('darkMode.system')} selected={themeMode === 'system'}
              onClick={() => set('system')} />
            <Divider />
            <RadioRow icon={<Sun size={16} />} label={t('darkMode.light')} selected={themeMode === 'light'}
              onClick={() => set('light')} />
            <Divider />
            <RadioRow icon={<Moon size={16} />} label={t('darkMode.dark')} selected={themeMode === 'dark'}
              onClick={() => set('dark')} />
          </div>
        </div>
        <div style={{ height: 24 }} />
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   5. General Settings
   ═══════════════════════════════════════════════ */

function GeneralSettingsPage({ onBack }: { onBack: () => void }) {
  const s = useSettingsStore();
  const token = useIMStore(st => st.currentUser?.token) || '';
  const t = useT();
  const u = (key: string, val: any) => s.updateSetting(key, val, token);

  const handleClearCache = () => {
    if (typeof window !== 'undefined') {
      sessionStorage.clear();
      toast.success(s.language === 'en' ? 'Cache cleared' : '缓存已清除');
    }
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title={t('settings.general')} onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        <div className="mx-3 mt-3">
          <div style={{ fontSize: 12, color: '#A2ACB5', fontWeight: 500, padding: '0 4px 6px' }}>{t('general.language')}</div>
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <RadioRow icon={<Globe size={16} />} label="简体中文" selected={s.language === 'zh-CN'}
              onClick={() => u('language', 'zh-CN')} />
            <Divider />
            <RadioRow icon={<Globe size={16} />} label="English" selected={s.language === 'en'}
              onClick={() => u('language', 'en')} />
          </div>
        </div>

        <div className="mx-3 mt-3">
          <div style={{ fontSize: 12, color: '#A2ACB5', fontWeight: 500, padding: '0 4px 6px' }}>{t('general.fontSize')}</div>
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <RadioRow icon={<Type size={16} />} label={t('general.fontSmall')} selected={s.fontSize === 'small'}
              onClick={() => u('fontSize', 'small')} />
            <Divider />
            <RadioRow icon={<Type size={16} />} label={t('general.fontMedium')} selected={s.fontSize === 'medium'}
              onClick={() => u('fontSize', 'medium')} />
            <Divider />
            <RadioRow icon={<Type size={16} />} label={t('general.fontLarge')} selected={s.fontSize === 'large'}
              onClick={() => u('fontSize', 'large')} />
          </div>
        </div>

        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <MenuItem icon={<Trash2 size={16} />} label={t('general.clearCache')} onClick={handleClearCache} danger />
          </div>
        </div>
        <div style={{ height: 24 }} />
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   MAIN: SettingsPage — Hub with sub-navigation
   ═══════════════════════════════════════════════ */

type SettingSub = 'account' | 'notification' | 'privacy' | 'darkmode' | 'general' | null;

export default function SettingsPage({ onBack }: { onBack: () => void }) {
  const [sub, setSub] = useState<SettingSub>(null);
  const { themeMode } = useSettingsStore();

  const t = useT();
  const themeLabel = themeMode === 'dark' ? t('darkMode.dark') : themeMode === 'light' ? t('darkMode.light') : t('darkMode.system');

  if (sub === 'account') return <AccountSecurityPage onBack={() => setSub(null)} />;
  if (sub === 'notification') return <NotificationSettingsPage onBack={() => setSub(null)} />;
  if (sub === 'privacy') return <PrivacySettingsPage onBack={() => setSub(null)} />;
  if (sub === 'darkmode') return <DarkModeSettingsPage onBack={() => setSub(null)} />;
  if (sub === 'general') return <GeneralSettingsPage onBack={() => setSub(null)} />;

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title={t('settings.title')} onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <MenuItem icon={<Lock size={16} />} label={t('settings.accountSecurity')} onClick={() => setSub('account')} />
            <Divider />
            <MenuItem icon={<Bell size={16} />} label={t('settings.notification')} onClick={() => setSub('notification')} />
            <Divider />
            <MenuItem icon={<Shield size={16} />} label={t('settings.privacy')} onClick={() => setSub('privacy')} />
          </div>
        </div>

        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <MenuItem icon={<Moon size={16} />} label={t('settings.darkMode')} value={themeLabel} onClick={() => setSub('darkmode')} />
            <Divider />
            <MenuItem icon={<BookOpen size={16} />} label={t('settings.general')} onClick={() => setSub('general')} />
          </div>
        </div>
        <div style={{ height: 24 }} />
      </div>
    </div>
  );
}
