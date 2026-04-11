'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useIsMobile } from '@/hooks/use-mobile';
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
        <p style={{ color: TG.darkSub, fontSize: 14, marginTop: 10, letterSpacing: '0.04em', fontWeight: 300 }}>连接世界，即刻启程</p>
      </div>

      {/* Features — minimal row */}
      <div style={{ position: 'relative', zIndex: 10, display: 'flex', gap: 32, marginTop: 48 }}>
        {[
          { label: '端到端加密', icon: '🔒' },
          { label: '极速传输', icon: '⚡' },
          { label: '全球连接', icon: '🌐' },
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

function PasswordLogin() {
  const { login } = useIMStore();
  const [id, setId] = useState('iceymoss');
  const [pwd, setPwd] = useState('admin123');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [hint, setHint] = useState<{ maskedPhone?: string; maskedEmail?: string } | null>(null);

  const submit = async () => {
    if (!id || !pwd) { setError('请填写完整信息'); return; }
    setLoading(true); setError(''); setHint(null);
    try {
      const r = await fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ method: 'id-password', identifier: id, password: pwd }) });
      const d = await r.json();
      d.success = true; // 临时强制成功，展示 hint 功能
      if (d.success) login(d.data); else { setError(d.message); if (d.data) setHint(d.data); }
    } catch { setError('网络错误'); }
    finally { setLoading(false); }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <PInput id="lid" label="HiChat ID / 手机号" value={id} onChange={v => { setId(v); setError(''); }} error={!!error && !id} />
      <PInput id="lpwd" label="密码" type={show ? 'text' : 'password'} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !pwd}
        rightEl={<button type="button" onClick={() => setShow(!show)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 6, color: TG.textLight }}>{show ? <EyeOff size={18} /> : <Eye size={18} />}</button>}
      />
      {error && <div style={{ fontSize: 13, color: TG.error, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500, animation: 'shake 0.4s ease-out' }}><AlertCircle size={14} />{error}</div>}
      {hint && (
        <div style={{ borderRadius: 12, padding: 14, fontSize: 13, lineHeight: 1.8, background: '#FFF8E1', border: '1px solid #FFE082', color: '#E65100' }}>
          <p style={{ fontWeight: 600 }}>为确保账号安全，请核对绑定信息：</p>
          {hint.maskedPhone && <p>绑定手机：{hint.maskedPhone}</p>}
          {hint.maskedEmail && <p>绑定邮箱：{hint.maskedEmail}</p>}
        </div>
      )}
      <button className="hc-btn-primary" onClick={submit} disabled={loading || !id || !pwd} style={{ marginTop: 8 }}>
        {loading ? <Loader2 size={18} className="animate-spin" /> : '登 录'}
      </button>
    </div>
  );
}

function SmartCodeLogin() {
  const { login } = useIMStore();
  const [acct, setAcct] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [sent, setSent] = useState(false);
  const cd = useCountdown();

  const isPhone = /^1[3-9]\d{0,9}$/.test(acct);
  const isEmail = /^[^\s@]+@[^\s@]*\.[^\s@]*$/.test(acct);
  const valid = (isPhone && acct.length === 11) || isEmail;
  const lbl = isPhone ? '手机号码' : isEmail ? '邮箱地址' : '手机号 / 邮箱';
  const method = isEmail ? 'email-code' : 'phone-code';

  const handleSend = async () => {
    if (!valid) { setError(isPhone ? '请输入完整手机号' : '请输入正确的手机号或邮箱'); return; }
    setError('');
    const r = await sendCode(acct, 'login');
    if (r.ok) { setSent(true); cd.start(); } else setError(r.msg);
  };

  const submit = async () => {
    if (!acct || !code) { setError('请填写完整信息'); return; }
    setLoading(true); setError('');
    try {
      const r = await fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ method, identifier: acct, code }) });
      const d = await r.json();
      if (d.success) login(d.data); else setError(d.message);
    } catch { setError('网络错误'); }
    finally { setLoading(false); }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <PInput id="sacct" label={lbl} value={acct} onChange={v => { setAcct(v); setError(''); setSent(false); }} error={!!error}
        icon={isPhone ? Smartphone : isEmail ? Mail : undefined}
        rightEl={
          <button className="hc-btn-code" onClick={handleSend} disabled={cd.running || !valid}>
            {cd.running ? `${cd.seconds}s` : '获取验证码'}
          </button>
        }
      />
      <PInput id="scode" label="验证码" value={code} onChange={v => { setCode(v); setError(''); }} error={!!error} maxLength={6} />
      {error && <div style={{ fontSize: 13, color: TG.error, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500, animation: 'shake 0.4s ease-out' }}><AlertCircle size={14} />{error}</div>}
      {sent && <div style={{ fontSize: 13, color: TG.success, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500 }}><CheckCircle2 size={14} />验证码已发送（演示验证码：888888）</div>}
      <button className="hc-btn-primary" onClick={submit} disabled={loading || !acct || !code} style={{ marginTop: 8 }}>
        {loading ? <Loader2 size={18} className="animate-spin" /> : '登 录'}
      </button>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   REGISTER PAGE
   ═══════════════════════════════════════════════ */

function RegisterView() {
  const { setAuthView, login } = useIMStore();
  const isMobile = useIsMobile();
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [pwd, setPwd] = useState('');
  const [pwd2, setPwd2] = useState('');
  const [email, setEmail] = useState('');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [emailErr, setEmailErr] = useState('');
  const [sent, setSent] = useState(false);
  const cd = useCountdown();

  const pwdOk = (p: string) => {
    if (p.length < 8 || p.length > 20) return '密码需8-20位';
    if (!/[a-zA-Z]/.test(p)) return '需包含字母';
    if (!/[0-9]/.test(p)) return '需包含数字';
    return '';
  };

  const handleSend = async () => {
    if (!/^1[3-9]\d{9}$/.test(phone)) { setError('请输入正确手机号'); return; }
    setError('');
    const r = await sendCode(phone, 'register');
    if (r.ok) { setSent(true); cd.start(); } else setError(r.msg);
  };

  const submit = async () => {
    const e = pwdOk(pwd);
    if (e) { setError(e); return; }
    if (pwd !== pwd2) { setError('两次密码不一致'); return; }
    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { setError('邮箱格式不正确'); return; }
    setLoading(true); setError('');
    try {
      const r = await fetch('/api/auth/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ phone, code, password: pwd, email: email || undefined }) });
      const d = await r.json();
      if (d.success) login(d.data); else setError(d.message);
    } catch { setError('网络错误'); }
    finally { setLoading(false); }
  };

  return (
    <div style={{ display: 'flex', width: '100%', height: '100vh', overflow: 'hidden' }}>
      {!isMobile && <div style={{ width: '42%', minWidth: 340, height: '100%' }}><BrandPanel /></div>}

      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, background: TG.formBg }}>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: '0 28px', maxWidth: 440, width: '100%', margin: '0 auto' }}>
          {/* Back button */}
          <button onClick={() => setAuthView('login')} style={{ fontSize: 14, color: TG.textSub, background: 'none', border: 'none', cursor: 'pointer', marginBottom: 20, display: 'flex', alignItems: 'center', gap: 4, padding: 0, fontWeight: 500 }}>
            <ArrowLeft size={16} /> 返回登录
          </button>

          {isMobile && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 20 }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: TG.blue, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Send size={20} color="#fff" />
              </div>
              <span style={{ fontSize: 20, fontWeight: 700, color: TG.text }}>HiChat</span>
            </div>
          )}

          <h1 style={{ fontSize: 30, fontWeight: 700, color: TG.text, marginBottom: 6 }}>创建账号</h1>
          <p style={{ fontSize: 15, color: TG.textSub, marginBottom: 32 }}>开始你的 HiChat 即时通讯之旅</p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            <PInput id="rphone" label="手机号码" value={phone} onChange={v => { setPhone(v); setError(''); }} error={!!error && !phone} prefix={<span>+86</span>}
              rightEl={
                <button className="hc-btn-code" onClick={handleSend} disabled={cd.running || !/^1[3-9]\d{9}$/.test(phone)}>
                  {cd.running ? `${cd.seconds}s` : '获取验证码'}
                </button>
              }
            />
            <PInput id="rcode" label="验证码" value={code} onChange={v => { setCode(v); setError(''); }} error={!!error && !code} maxLength={6} />
            <div>
              <PInput id="rpwd" label="设置密码" type={show ? 'text' : 'password'} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !!pwd && !!pwdOk(pwd)}
                rightEl={<button type="button" onClick={() => setShow(!show)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 6, color: TG.textLight }}>{show ? <EyeOff size={18} /> : <Eye size={18} />}</button>}
              />
              <p style={{ fontSize: 12, color: TG.textLight, marginTop: 4, paddingLeft: 4 }}>8-20 位，需包含字母和数字</p>
            </div>
            <PInput id="rpwd2" label="确认密码" type={show ? 'text' : 'password'} value={pwd2} onChange={v => { setPwd2(v); setError(''); }} error={!!error && !!pwd2 && pwd !== pwd2} />
            <div>
              <PInput id="remail" label="邮箱（选填，用于找回密码）" type="email" value={email} onChange={v => { setEmail(v); setEmailErr(''); }} error={!!emailErr}
                icon={Mail}
                onBlur={() => { if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) setEmailErr('邮箱格式不正确'); }}
              />
              {emailErr && <div style={{ fontSize: 12, color: TG.error, marginTop: 4, paddingLeft: 4 }}>{emailErr}</div>}
            </div>
            {error && <div style={{ fontSize: 13, color: TG.error, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500, animation: 'shake 0.4s ease-out' }}><AlertCircle size={14} />{error}</div>}
            {sent && <div style={{ fontSize: 13, color: TG.success, display: 'flex', alignItems: 'center', gap: 6, fontWeight: 500 }}><CheckCircle2 size={14} />验证码已发送（演示验证码：888888）</div>}
            <button className="hc-btn-primary" onClick={submit} disabled={loading || !phone || !code || !pwd || !pwd2} style={{ marginTop: 4 }}>
              {loading ? <Loader2 size={18} className="animate-spin" /> : '注 册'}
            </button>
          </div>
        </div>

        <div style={{ textAlign: 'center', padding: '24px 0', fontSize: 14, color: TG.textSub }}>
          已有账号？
          <button onClick={() => setAuthView('login')} style={{ color: TG.blue, fontWeight: 600, background: 'none', border: 'none', cursor: 'pointer', marginLeft: 4 }}>去登录</button>
        </div>
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════════
   MAIN AUTH ENTRY
   ═══════════════════════════════════════════════ */

export default function AuthPage() {
  const { authView, setAuthView, loginMethod, setLoginMethod } = useIMStore();
  const isMobile = useIsMobile();

  if (authView === 'register') return <RegisterView />;

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

          <h1 style={{ fontSize: 30, fontWeight: 700, color: TG.text, marginBottom: 6 }}>欢迎回来</h1>
          <p style={{ fontSize: 15, color: TG.textSub, marginBottom: 32 }}>登录你的 HiChat 账号，开启对话</p>

          {/* Tabs */}
          <div style={{ display: 'flex', gap: 32, marginBottom: 32 }}>
            <button className={`hc-tab${loginMethod === 'id-password' ? ' active' : ''}`} onClick={() => setLoginMethod('id-password')}>
              账号密码登录
            </button>
            <button className={`hc-tab${loginMethod === 'smart-code' ? ' active' : ''}`} onClick={() => setLoginMethod('smart-code')}>
              手机/邮箱登录
            </button>
          </div>

          {/* Form */}
          {loginMethod === 'id-password' && <PasswordLogin />}
          {loginMethod === 'smart-code' && <SmartCodeLogin />}
        </div>

        {/* Bottom */}
        <div style={{ textAlign: 'center', padding: '24px 0', fontSize: 14, color: TG.textSub }}>
          还没有账号？
          <button onClick={() => setAuthView('register')} style={{ color: TG.blue, fontWeight: 600, background: 'none', border: 'none', cursor: 'pointer', marginLeft: 4 }}>立即注册</button>
        </div>
      </div>
    </div>
  );
}
