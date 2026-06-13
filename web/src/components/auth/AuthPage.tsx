'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useIsMobile } from '@/hooks/use-mobile';
import { useT } from '@/hooks/use-i18n';
import {
  Eye, EyeOff, Loader2, CheckCircle2, Send,
  ChevronDown, AlertCircle, Mail, Smartphone, ArrowLeft,
} from 'lucide-react';

/* ═══════════════════════════════════════════════
   Telegram Web Color Palette
   ═══════════════════════════════════════════════ */
const TG = {
  blue: '#2AABEE',           // Primary accent
  blueHover: '#3390EC',      // Primary hover
  blueDark: '#229ED9',       // Primary dark
  blueGlow: 'rgba(42, 171, 238, 0.3)',
  darkBg: '#17212B',         // Dark panel background
  darkSurface: '#242F3D',    // Dark surface
  darkInput: '#242F3D',      // Input bg on dark
  darkBorder: '#2B5278',     // Subtle borders on dark
  darkText: '#FFFFFF',
  darkSub: 'rgba(255,255,255,0.5)',
  formBg: '#FFFFFF',
  inputBg: '#F4F4F5',       // Filled input background
  inputBorder: '#DADCE0',
  inputFocus: '#2AABEE',
  text: '#1C2733',
  textSub: '#708499',
  textLight: '#A2ACB5',
  error: '#E53935',
  errorBg: '#FFF0F0',
  success: '#4CAF50',
};

/* ═══════════════════════════════════════════════
   Helpers
   ═══════════════════════════════════════════════ */

function useCountdown(sec = 60) {
  const [s, setS] = useState(0);
  const running = s > 0;
  useEffect(() => {
    if (s <= 0) return;
    const t = setInterval(() => setS(v => (v <= 1 ? 0 : v - 1)), 1000);
    return () => clearInterval(t);
  }, [s > 0]);
  const start = useCallback(() => setS(sec), [sec]);
  return { seconds: s, running, start };
}

async function sendCode(target: string, type: 'register' | 'login') {
  try {
    const res = await fetch('/api/auth/send-code', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target, type }),
    });
    const d = await res.json();
    return { ok: d.success, msg: d.message };
  } catch { return { ok: false, msg: '网络错误，请稍后重试' }; }
}

/* ═══════════════════════════════════════════════
   Brand Panel — TG dark style with animated particles
   ═══════════════════════════════════════════════ */

function NetworkCanvas() {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const cvs = ref.current;
    if (!cvs) return;
    const ctx = cvs.getContext('2d');
    if (!ctx) return;
    const resize = () => {
      const dpr = window.devicePixelRatio || 1;
      cvs.width = cvs.offsetWidth * dpr;
      cvs.height = cvs.offsetHeight * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };
    resize();
    window.addEventListener('resize', resize);
    const W = cvs.offsetWidth, H = cvs.offsetHeight;
    const pts = Array.from({ length: 45 }, () => ({
      x: Math.random() * W, y: Math.random() * H,
      vx: (Math.random() - 0.5) * 0.25, vy: (Math.random() - 0.5) * 0.25,
      r: Math.random() * 2 + 0.5,
    }));
    let id: number;
    const loop = () => {
      ctx.clearRect(0, 0, W, H);
      for (const p of pts) {
        p.x += p.vx; p.y += p.vy;
        if (p.x < 0 || p.x > W) p.vx *= -1;
        if (p.y < 0 || p.y > H) p.vy *= -1;
      }
      for (let i = 0; i < pts.length; i++) {
        for (let j = i + 1; j < pts.length; j++) {
          const dx = pts[i].x - pts[j].x, dy = pts[i].y - pts[j].y;
          const d = Math.sqrt(dx * dx + dy * dy);
          if (d < 140) {
            const a = (1 - d / 140) * 0.2;
            ctx.strokeStyle = `rgba(42, 171, 238, ${a})`;
            ctx.lineWidth = 0.6;
            ctx.beginPath();
            ctx.moveTo(pts[i].x, pts[i].y);
            ctx.lineTo(pts[j].x, pts[j].y);
            ctx.stroke();
          }
        }
      }
      for (const p of pts) {
        ctx.fillStyle = 'rgba(42, 171, 238, 0.4)';
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fill();
      }
      id = requestAnimationFrame(loop);
    };
    loop();
    return () => { cancelAnimationFrame(id); window.removeEventListener('resize', resize); };
  }, []);
  return <canvas ref={ref} style={{ position: 'absolute', inset: 0, pointerEvents: 'none' }} />;
}

function BrandPanel() {
  const t = useT();
  return (
    <div style={{
      width: '100%', height: '100%',
      background: `linear-gradient(160deg, ${TG.darkBg} 0%, #1C2733 60%, #0E1621 100%)`,
      position: 'relative', overflow: 'hidden',
      display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
    }}>
      <NetworkCanvas />

      {/* Decorative circles */}
      <div style={{ position: 'absolute', width: 500, height: 500, top: -180, right: -120, borderRadius: '50%', background: 'rgba(42,171,238,0.05)', border: '1px solid rgba(42,171,238,0.08)', pointerEvents: 'none' }} />
      <div style={{ position: 'absolute', width: 350, height: 350, bottom: -100, left: -80, borderRadius: '50%', background: 'rgba(42,171,238,0.04)', border: '1px solid rgba(42,171,238,0.06)', pointerEvents: 'none' }} />

      {/* Glow */}
      <div style={{ position: 'absolute', inset: 0, background: 'radial-gradient(ellipse 50% 45% at 50% 45%, rgba(42,171,238,0.1) 0%, transparent 70%)', pointerEvents: 'none' }} />

      {/* Logo + Title */}
      <div style={{ position: 'relative', zIndex: 10, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <div style={{
          width: 80, height: 80, borderRadius: 22,
          background: 'rgba(42,171,238,0.12)',
          border: '1px solid rgba(42,171,238,0.2)',
          boxShadow: '0 0 60px rgba(42,171,238,0.2)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <Send size={38} color={TG.blue} />
        </div>
        <h1 style={{ color: TG.darkText, fontSize: 38, fontWeight: 700, marginTop: 24, letterSpacing: '-0.02em' }}>HiChat</h1>
        <p style={{ color: TG.darkSub, fontSize: 14, marginTop: 10, letterSpacing: '0.04em', fontWeight: 300 }}>{t('auth.slogan')}</p>
      </div>

      {/* Features — minimal row */}
      <div style={{ position: 'relative', zIndex: 10, display: 'flex', gap: 32, marginTop: 48 }}>
        {[
          { label: t('auth.feature.e2e'), icon: '🔒' },
          { label: t('auth.feature.fast'), icon: '⚡' },
          { label: t('auth.feature.global'), icon: '🌐' },
        ].map(f => (
          <div key={f.label} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6 }}>
            <span style={{ fontSize: 20 }}>{f.icon}</span>
            <span style={{ color: TG.darkSub, fontSize: 11, fontWeight: 400 }}>{f.label}</span>
          </div>
        ))}
      </div>

      <div style={{ position: 'absolute', bottom: 24, left: 0, right: 0, textAlign: 'center', zIndex: 10 }}>
        <p style={{ color: 'rgba(255,255,255,0.15)', fontSize: 11 }}>&copy; {new Date().getFullYear()} HiChat</p>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   PROMINENT INPUT — filled bg, rounded, clear
   ═══════════════════════════════════════════════ */

function PInput({ id, label, type = 'text', value, onChange, error, maxLength, rightEl, icon: Icon, prefix }: {
  id: string; label: string; type?: string; value: string; onChange: (v: string) => void;
  error?: boolean; maxLength?: number; rightEl?: React.ReactNode;
  icon?: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  prefix?: React.ReactNode;
}) {
  const [focused, setFocused] = useState(false);
  const active = focused || value.length > 0;

  return (
    <div style={{ position: 'relative' }}>
      {/* Label */}
      <label htmlFor={id} style={{
        position: 'absolute', left: (Icon ? 42 : 0) + (prefix ? 56 : 16), top: active ? -9 : 16,
        fontSize: active ? 12 : 15,
        color: error ? TG.error : focused ? TG.blue : TG.textLight,
        fontWeight: active ? 600 : 400,
        background: active ? '#fff' : 'transparent',
        padding: active ? '0 6px' : '0',
        transition: 'all 0.2s cubic-bezier(0.4,0,0.2,1)',
        pointerEvents: 'none', zIndex: 2,
        lineHeight: '16px',
      }}>
        {label}
      </label>

      {/* Input wrapper */}
      <div style={{
        display: 'flex', alignItems: 'center',
        background: focused ? '#fff' : TG.inputBg,
        border: `2px solid ${error ? TG.error : focused ? TG.blue : 'transparent'}`,
        borderRadius: 12,
        transition: 'all 0.2s ease',
        boxShadow: focused ? `0 0 0 3px ${TG.blueGlow}` : 'none',
      }}>
        {/* Left icon */}
        {Icon && (
          <div style={{ paddingLeft: 14, paddingRight: 4, color: focused ? TG.blue : TG.textLight, display: 'flex', transition: 'color 0.2s' }}>
            <Icon size={18} style={{}} />
          </div>
        )}

        {/* Prefix (like +86) */}
        {prefix && (
          <div style={{ paddingLeft: 14, paddingRight: 2, color: TG.text, fontSize: 15, fontWeight: 500, display: 'flex', alignItems: 'center', gap: 1 }}>
            {prefix}
            <ChevronDown size={14} color={TG.textLight} />
          </div>
        )}

        {/* Input */}
        <input
          id={id}
          type={type}
          value={value}
          onChange={e => onChange(e.target.value)}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          maxLength={maxLength}
          className="auth-input-reset"
          style={{
            flex: 1, background: 'transparent',
            border: 'none', outline: 'none',
            fontSize: 16, color: TG.text,
            padding: '15px 14px',
            borderRadius: 0,
          }}
          autoComplete="off"
        />

        {/* Right action */}
        {rightEl && (
          <div style={{ paddingRight: 12, display: 'flex', alignItems: 'center' }}>{rightEl}</div>
        )}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   LOGIN FORMS
   ═══════════════════════════════════════════════ */

function PhonePasswordLogin() {
  const { login, setAuthView } = useIMStore();
  const t = useT();
  const [phone, setPhone] = useState('');
  const [pwd, setPwd] = useState('');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const submit = async () => {
    if (!phone || !pwd) { setError(t('auth.err.incomplete')); return; }
    if (!/^1[3-9]\d{9}$/.test(phone)) { setError(t('auth.err.phone')); return; }
    setLoading(true); setError('');
    try {
      const r = await fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ phone, password: pwd }) });
      const d = await r.json();
      if (d.success) login(d.data); else setError(d.message);
    } catch { setError(t('auth.err.network')); }
    finally { setLoading(false); }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <PInput id="lphone" label={t('auth.phone')} value={phone} onChange={v => { setPhone(v); setError(''); }} error={!!error && !phone} icon={Smartphone} prefix={<span>+86</span>} />
      <PInput id="lpwd" label={t('auth.password')} type={show ? 'text' : 'password'} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !pwd}
        rightEl={<button type="button" onClick={() => setShow(!show)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 6, color: TG.textLight }}>{show ? <EyeOff size={18} /> : <Eye size={18} />}</button>}
      />
      {error && <div style={{ fontSize: 13, color: TG.error, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500, animation: 'shake 0.4s ease-out' }}><AlertCircle size={14} />{error}</div>}
      <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
        <button onClick={() => setAuthView('forgot-password')} style={{ fontSize: 13, color: TG.blue, background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontWeight: 500 }}>{t('auth.forgotPwd')}</button>
      </div>
      <button className="hc-btn-primary" onClick={submit} disabled={loading || !phone || !pwd} style={{ marginTop: 4 }}>
        {loading ? <Loader2 size={18} className="animate-spin" /> : t('auth.login')}
      </button>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   REGISTER PAGE
   ═══════════════════════════════════════════════ */

function RegisterView() {
  const { setAuthView, login } = useIMStore();
  const t = useT();
  const isMobile = useIsMobile();
  const [phone, setPhone] = useState('');
  const [nickname, setNickname] = useState('');
  const [code, setCode] = useState('');
  const [pwd, setPwd] = useState('');
  const [pwd2, setPwd2] = useState('');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [sent, setSent] = useState(false);
  const cd = useCountdown();

  const pwdOk = (p: string) => {
    if (p.length < 8 || p.length > 20) return t('auth.err.pwdLen');
    if (!/[a-zA-Z]/.test(p)) return t('auth.err.pwdLetter');
    if (!/[0-9]/.test(p)) return t('auth.err.pwdDigit');
    return '';
  };

  const handleSend = async () => {
    if (!/^1[3-9]\d{9}$/.test(phone)) { setError(t('auth.err.phone')); return; }
    setError('');
    const r = await sendCode(phone, 'register');
    if (r.ok) { setSent(true); cd.start(); } else setError(r.msg);
  };

  const submit = async () => {
    if (!nickname.trim()) { setError(t('auth.err.nickname')); return; }
    const e = pwdOk(pwd);
    if (e) { setError(e); return; }
    if (pwd !== pwd2) { setError(t('auth.err.pwdMismatch')); return; }
    setLoading(true); setError('');
    try {
      const r = await fetch('/api/auth/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ phone, password: pwd, nickname: nickname.trim(), phoneCode: code }) });
      const d = await r.json();
      if (d.success) login(d.data); else setError(d.message);
    } catch { setError(t('auth.err.network')); }
    finally { setLoading(false); }
  };

  return (
    <div style={{ display: 'flex', width: '100%', height: '100vh', overflow: 'hidden' }}>
      {!isMobile && <div style={{ width: '42%', minWidth: 340, height: '100%' }}><BrandPanel /></div>}

      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, background: TG.formBg }}>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: '0 28px', maxWidth: 440, width: '100%', margin: '0 auto' }}>
          {/* Back button */}
          <button onClick={() => setAuthView('login')} style={{ fontSize: 14, color: TG.textSub, background: 'none', border: 'none', cursor: 'pointer', marginBottom: 20, display: 'flex', alignItems: 'center', gap: 4, padding: 0, fontWeight: 500 }}>
            <ArrowLeft size={16} /> {t('auth.backToLogin')}
          </button>

          {isMobile && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 20 }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: TG.blue, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Send size={20} color="#fff" />
              </div>
              <span style={{ fontSize: 20, fontWeight: 700, color: TG.text }}>HiChat</span>
            </div>
          )}

          <h1 style={{ fontSize: 30, fontWeight: 700, color: TG.text, marginBottom: 6 }}>{t('auth.createAccount')}</h1>
          <p style={{ fontSize: 15, color: TG.textSub, marginBottom: 32 }}>{t('auth.registerSub')}</p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            <PInput id="rphone" label={t('auth.phone')} value={phone} onChange={v => { setPhone(v); setError(''); }} error={!!error && !phone} prefix={<span>+86</span>}
              rightEl={
                <button className="hc-btn-code" onClick={handleSend} disabled={cd.running || !/^1[3-9]\d{9}$/.test(phone)}>
                  {cd.running ? `${cd.seconds}s` : t('auth.getCode')}
                </button>
              }
            />
            <PInput id="rcode" label={t('auth.code')} value={code} onChange={v => { setCode(v); setError(''); }} error={!!error && !code} maxLength={6} />
            <PInput id="rnick" label={t('auth.nickname')} value={nickname} onChange={v => { setNickname(v); setError(''); }} error={!!error && !nickname} maxLength={20} />
            <div>
              <PInput id="rpwd" label={t('auth.setPwd')} type={show ? 'text' : 'password'} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !!pwd && !!pwdOk(pwd)}
                rightEl={<button type="button" onClick={() => setShow(!show)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 6, color: TG.textLight }}>{show ? <EyeOff size={18} /> : <Eye size={18} />}</button>}
              />
              <p style={{ fontSize: 12, color: TG.textLight, marginTop: 4, paddingLeft: 4 }}>{t('auth.pwdHint')}</p>
            </div>
            <PInput id="rpwd2" label={t('auth.confirmPwd')} type={show ? 'text' : 'password'} value={pwd2} onChange={v => { setPwd2(v); setError(''); }} error={!!error && !!pwd2 && pwd !== pwd2} />
            {error && <div style={{ fontSize: 13, color: TG.error, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500, animation: 'shake 0.4s ease-out' }}><AlertCircle size={14} />{error}</div>}
            {sent && <div style={{ fontSize: 13, color: TG.success, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500 }}><CheckCircle2 size={14} />{t('auth.codeSent')}</div>}
            <button className="hc-btn-primary" onClick={submit} disabled={loading || !phone || !code || !nickname || !pwd || !pwd2} style={{ marginTop: 4 }}>
              {loading ? <Loader2 size={18} className="animate-spin" /> : t('auth.register')}
            </button>
          </div>
        </div>

        <div style={{ textAlign: 'center', padding: '24px 0', fontSize: 14, color: TG.textSub }}>
          {t('auth.haveAccount')}
          <button onClick={() => setAuthView('login')} style={{ color: TG.blue, fontWeight: 600, background: 'none', border: 'none', cursor: 'pointer', marginLeft: 4 }}>{t('auth.goLogin')}</button>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   FORGOT PASSWORD PAGE
   ═══════════════════════════════════════════════ */

function ForgotPasswordView() {
  const { setAuthView } = useIMStore();
  const t = useT();
  const isMobile = useIsMobile();
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [pwd, setPwd] = useState('');
  const [pwd2, setPwd2] = useState('');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [sent, setSent] = useState(false);
  const cd = useCountdown();

  const handleSend = async () => {
    if (!/^1[3-9]\d{9}$/.test(phone)) { setError(t('auth.err.phone')); return; }
    setError('');
    const r = await sendCode(phone, 'register');
    if (r.ok) { setSent(true); cd.start(); } else setError(r.msg);
  };

  const submit = async () => {
    if (!phone || !code || !pwd) { setError(t('auth.err.incomplete')); return; }
    if (pwd.length < 8 || pwd.length > 20) { setError(t('auth.err.pwdLen')); return; }
    if (pwd !== pwd2) { setError(t('auth.err.pwdMismatch')); return; }
    setLoading(true); setError('');
    try {
      const r = await fetch('/api/auth/reset-pwd', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ phone, code, password: pwd }) });
      const d = await r.json();
      if (d.success) setSuccess(true); else setError(d.message);
    } catch { setError(t('auth.err.network')); }
    finally { setLoading(false); }
  };

  if (success) {
    return (
      <div style={{ display: 'flex', width: '100%', height: '100vh', overflow: 'hidden' }}>
        {!isMobile && <div style={{ width: '42%', minWidth: 340, height: '100%' }}><BrandPanel /></div>}
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', background: TG.formBg, padding: '0 28px' }}>
          <CheckCircle2 size={48} color={TG.success} />
          <h2 style={{ fontSize: 22, fontWeight: 700, color: TG.text, marginTop: 16 }}>{t('auth.resetOk')}</h2>
          <p style={{ fontSize: 14, color: TG.textSub, marginTop: 8 }}>{t('auth.resetOkSub')}</p>
          <button className="hc-btn-primary" onClick={() => setAuthView('login')} style={{ marginTop: 24, width: 200 }}>{t('auth.goLogin')}</button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', width: '100%', height: '100vh', overflow: 'hidden' }}>
      {!isMobile && <div style={{ width: '42%', minWidth: 340, height: '100%' }}><BrandPanel /></div>}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, background: TG.formBg }}>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: '0 28px', maxWidth: 440, width: '100%', margin: '0 auto' }}>
          <button onClick={() => setAuthView('login')} style={{ fontSize: 14, color: TG.textSub, background: 'none', border: 'none', cursor: 'pointer', marginBottom: 20, display: 'flex', alignItems: 'center', gap: 4, padding: 0, fontWeight: 500 }}>
            <ArrowLeft size={16} /> {t('auth.backToLogin')}
          </button>
          <h1 style={{ fontSize: 30, fontWeight: 700, color: TG.text, marginBottom: 6 }}>{t('auth.resetPwd')}</h1>
          <p style={{ fontSize: 15, color: TG.textSub, marginBottom: 32 }}>{t('auth.resetSub')}</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            <PInput id="fpphone" label={t('auth.phone')} value={phone} onChange={v => { setPhone(v); setError(''); }} error={!!error && !phone} icon={Smartphone} prefix={<span>+86</span>}
              rightEl={
                <button className="hc-btn-code" onClick={handleSend} disabled={cd.running || !/^1[3-9]\d{9}$/.test(phone)}>
                  {cd.running ? `${cd.seconds}s` : t('auth.getCode')}
                </button>
              }
            />
            <PInput id="fpcode" label={t('auth.code')} value={code} onChange={v => { setCode(v); setError(''); }} error={!!error && !code} maxLength={6} />
            <div>
              <PInput id="fppwd" label={t('auth.newPwd')} type={show ? 'text' : 'password'} value={pwd} onChange={v => { setPwd(v); setError(''); }}
                rightEl={<button type="button" onClick={() => setShow(!show)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 6, color: TG.textLight }}>{show ? <EyeOff size={18} /> : <Eye size={18} />}</button>}
              />
              <p style={{ fontSize: 12, color: TG.textLight, marginTop: 4, paddingLeft: 4 }}>{t('auth.pwdHint')}</p>
            </div>
            <PInput id="fppwd2" label={t('auth.confirmNewPwd')} type={show ? 'text' : 'password'} value={pwd2} onChange={v => { setPwd2(v); setError(''); }} error={!!error && !!pwd2 && pwd !== pwd2} />
            {error && <div style={{ fontSize: 13, color: TG.error, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500 }}><AlertCircle size={14} />{error}</div>}
            {sent && <div style={{ fontSize: 13, color: TG.success, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500 }}><CheckCircle2 size={14} />{t('auth.codeSent')}</div>}
            <button className="hc-btn-primary" onClick={submit} disabled={loading || !phone || !code || !pwd || !pwd2} style={{ marginTop: 4 }}>
              {loading ? <Loader2 size={18} className="animate-spin" /> : t('auth.resetPwd')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   MAIN AUTH ENTRY
   ═══════════════════════════════════════════════ */

export default function AuthPage() {
  const { authView, setAuthView } = useIMStore();
  const t = useT();
  const isMobile = useIsMobile();

  if (authView === 'register') return <RegisterView />;
  if (authView === 'forgot-password') return <ForgotPasswordView />;

  return (
    <div style={{ display: 'flex', width: '100%', height: '100vh', overflow: 'hidden' }}>
      {/* Left brand panel */}
      {!isMobile && <div style={{ width: '42%', minWidth: 340, height: '100%' }}><BrandPanel /></div>}

      {/* Right form panel */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, background: TG.formBg }}>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: '0 28px', maxWidth: 440, width: '100%', margin: '0 auto' }}>
          {/* Header */}
          {isMobile && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 28 }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: TG.blue, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Send size={20} color="#fff" />
              </div>
              <span style={{ fontSize: 20, fontWeight: 700, color: TG.text }}>HiChat</span>
            </div>
          )}

          <h1 style={{ fontSize: 30, fontWeight: 700, color: TG.text, marginBottom: 6 }}>{t('auth.welcomeBack')}</h1>
          <p style={{ fontSize: 15, color: TG.textSub, marginBottom: 32 }}>{t('auth.loginSub')}</p>

          {/* Form */}
          <PhonePasswordLogin />
        </div>

        {/* Bottom */}
        <div style={{ textAlign: 'center', padding: '24px 0', fontSize: 14, color: TG.textSub }}>
          {t('auth.noAccount')}
          <button onClick={() => setAuthView('register')} style={{ color: TG.blue, fontWeight: 600, background: 'none', border: 'none', cursor: 'pointer', marginLeft: 4 }}>{t('auth.goRegister')}</button>
        </div>
      </div>
    </div>
  );
}
