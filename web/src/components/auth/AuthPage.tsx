'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useIMStore } from '@/lib/im-store';
import { useSettingsStore } from '@/lib/settings-store';
import { useT } from '@/hooks/use-i18n';
import { Eye, EyeOff, Loader2, CheckCircle2, ChevronDown, AlertCircle } from 'lucide-react';
import Logo from '@/components/brand/Logo';

const C = {
  error: '#E53935',
  errorDark: '#FF9A9A',
  success: '#4CAF50',
  successDark: '#7BE0A0',
};

const PHONE_RE = /^1[3-9]\d{9}$/;
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

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

const pwdOk = (p: string, t: (k: string) => string) => {
  if (p.length < 8 || p.length > 20) return t('auth.err.pwdLen');
  if (!/[a-zA-Z]/.test(p)) return t('auth.err.pwdLetter');
  if (!/[0-9]/.test(p)) return t('auth.err.pwdDigit');
  return '';
};

/* ═══════════════════════════════════════════════
   Pill input (translucent, demo-style)
   ═══════════════════════════════════════════════ */

function Pill({ value, onChange, placeholder, type = 'text', error, maxLength, prefix, right }: {
  value: string; onChange: (v: string) => void; placeholder: string;
  type?: string; error?: boolean; maxLength?: number;
  prefix?: React.ReactNode; right?: React.ReactNode;
}) {
  const [focused, setFocused] = useState(false);
  return (
    <div className={`pill${focused ? ' focused' : ''}${error ? ' err' : ''}`}>
      {prefix && <span className="pill-prefix">{prefix}<ChevronDown size={13} style={{ opacity: 0.6 }} /></span>}
      <input
        type={type}
        value={value}
        placeholder={placeholder}
        onChange={e => onChange(e.target.value)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        maxLength={maxLength}
        autoComplete="off"
      />
      {right}
    </div>
  );
}

function EyeBtn({ show, toggle }: { show: boolean; toggle: () => void }) {
  return (
    <button type="button" className="pill-icon" onClick={toggle} aria-label="toggle password">
      {show ? <EyeOff size={18} /> : <Eye size={18} />}
    </button>
  );
}

function Mac() {
  return <div className="auth-mac"><span className="dot r" /><span className="dot y" /><span className="dot g" /></div>;
}

/* Top-right CN/EN language switcher (persisted on login via settings store) */
function LangSwitch() {
  const lang = useSettingsStore(s => s.language);
  const setLanguage = useSettingsStore(s => s.setLanguage);
  return (
    <div className="auth-lang">
      <button className={lang === 'zh-CN' ? 'on' : ''} onClick={() => setLanguage('zh-CN')}>中</button>
      <button className={lang === 'en' ? 'on' : ''} onClick={() => setLanguage('en')}>EN</button>
    </div>
  );
}

function Err({ dark, children }: { dark?: boolean; children: React.ReactNode }) {
  return <div className="auth-msg" style={{ color: dark ? C.errorDark : C.error, animation: 'shake 0.4s ease-out' }}><AlertCircle size={14} />{children}</div>;
}

function Sent({ dark, children }: { dark?: boolean; children: React.ReactNode }) {
  return <div className="auth-msg" style={{ color: dark ? C.successDark : C.success }}><CheckCircle2 size={14} />{children}</div>;
}

/* ═══════════════════════════════════════════════
   LOGIN box (dark)
   ═══════════════════════════════════════════════ */

function LoginBox() {
  const { login, setAuthView } = useIMStore();
  const t = useT();
  const [account, setAccount] = useState('');
  const [pwd, setPwd] = useState('');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const submit = async () => {
    if (!account || !pwd) { setError(t('auth.err.incomplete')); return; }
    if (!PHONE_RE.test(account) && !EMAIL_RE.test(account)) { setError(t('auth.err.account')); return; }
    setLoading(true); setError('');
    try {
      const r = await fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ phone: account, password: pwd }) });
      const d = await r.json();
      if (d.success) login(d.data); else setError(d.message);
    } catch { setError(t('auth.err.network')); }
    finally { setLoading(false); }
  };

  return (
    <>
      <Mac />
      <Logo variant="lockup-dark" height={34} className="auth-brand" />
      <h1 className="auth-h">{t('auth.login')}</h1>
      <div className="auth-fields">
        <Pill placeholder={t('auth.accountPh')} value={account} onChange={v => { setAccount(v); setError(''); }} error={!!error && !account} />
        <Pill type={show ? 'text' : 'password'} placeholder={t('auth.password')} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !pwd}
          right={<EyeBtn show={show} toggle={() => setShow(!show)} />} />
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: -4 }}>
          <button className="auth-link" onClick={() => setAuthView('forgot-password')}>{t('auth.forgotPwd')}</button>
        </div>
        {error && <Err dark>{error}</Err>}
        <button className="auth-go" onClick={submit} disabled={loading || !account || !pwd}>
          {loading ? <Loader2 size={18} className="animate-spin" /> : t('auth.login')}
        </button>
      </div>
      <div className="auth-corner" />
      <button className="auth-change" onClick={() => setAuthView('register')}>{t('auth.toRegister')}</button>
    </>
  );
}

/* ═══════════════════════════════════════════════
   REGISTER box (light)
   ═══════════════════════════════════════════════ */

function RegisterBox() {
  const { setAuthView, login } = useIMStore();
  const t = useT();
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

  const handleSend = async () => {
    if (!PHONE_RE.test(phone)) { setError(t('auth.err.phone')); return; }
    setError('');
    const r = await sendCode(phone, 'register');
    if (r.ok) { setSent(true); cd.start(); } else setError(r.msg);
  };

  const submit = async () => {
    if (!nickname.trim()) { setError(t('auth.err.nickname')); return; }
    const e = pwdOk(pwd, t);
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
    <>
      <Mac />
      <Logo variant="lockup" height={34} className="auth-brand" />
      <h1 className="auth-h">{t('auth.register')}</h1>
      <div className="auth-fields">
        <Pill placeholder={t('auth.phone')} value={phone} onChange={v => { setPhone(v); setError(''); }} error={!!error && !phone} prefix={<span>+86</span>}
          right={<button className="hc-btn-code" style={{ marginRight: 2 }} onClick={handleSend} disabled={cd.running || !PHONE_RE.test(phone)}>{cd.running ? `${cd.seconds}s` : t('auth.getCode')}</button>} />
        <Pill placeholder={t('auth.code')} value={code} onChange={v => { setCode(v); setError(''); }} error={!!error && !code} maxLength={6} />
        <Pill placeholder={t('auth.nickname')} value={nickname} onChange={v => { setNickname(v); setError(''); }} error={!!error && !nickname} maxLength={20} />
        <Pill type={show ? 'text' : 'password'} placeholder={t('auth.setPwd')} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !!pwd && !!pwdOk(pwd, t)}
          right={<EyeBtn show={show} toggle={() => setShow(!show)} />} />
        <Pill type={show ? 'text' : 'password'} placeholder={t('auth.confirmPwd')} value={pwd2} onChange={v => { setPwd2(v); setError(''); }} error={!!error && !!pwd2 && pwd !== pwd2} />
        {error && <Err>{error}</Err>}
        {sent && !error && <Sent>{t('auth.codeSent')}</Sent>}
        <button className="auth-go" onClick={submit} disabled={loading || !phone || !code || !nickname || !pwd || !pwd2}>
          {loading ? <Loader2 size={18} className="animate-spin" /> : t('auth.register')}
        </button>
      </div>
      <div className="auth-corner" />
      <button className="auth-change" onClick={() => setAuthView('login')}>{t('auth.toLogin')}</button>
    </>
  );
}

/* ═══════════════════════════════════════════════
   RESET box (light) — phone or email
   ═══════════════════════════════════════════════ */

function ResetBox() {
  const { setAuthView } = useIMStore();
  const t = useT();
  const [account, setAccount] = useState('');
  const [code, setCode] = useState('');
  const [pwd, setPwd] = useState('');
  const [pwd2, setPwd2] = useState('');
  const [show, setShow] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [sent, setSent] = useState(false);
  const cd = useCountdown();

  const isPhone = PHONE_RE.test(account);
  const isEmail = EMAIL_RE.test(account);

  const handleSend = async () => {
    if (!isPhone && !isEmail) { setError(t('auth.err.account')); return; }
    setError('');
    const r = await sendCode(account, 'register');
    if (r.ok) { setSent(true); cd.start(); } else setError(r.msg);
  };

  const submit = async () => {
    if (!account || !code || !pwd) { setError(t('auth.err.incomplete')); return; }
    if (!isPhone && !isEmail) { setError(t('auth.err.account')); return; }
    const e = pwdOk(pwd, t);
    if (e) { setError(e); return; }
    if (pwd !== pwd2) { setError(t('auth.err.pwdMismatch')); return; }
    setLoading(true); setError('');
    try {
      const body = isEmail ? { email: account, code, password: pwd } : { phone: account, code, password: pwd };
      const r = await fetch('/api/auth/reset-pwd', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      const d = await r.json();
      if (d.success) setSuccess(true); else setError(d.message);
    } catch { setError(t('auth.err.network')); }
    finally { setLoading(false); }
  };

  if (success) {
    return (
      <>
        <Mac />
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', textAlign: 'center', padding: '10px 20px 30px' }}>
          <CheckCircle2 size={48} color={C.success} />
          <h2 style={{ fontSize: 22, fontWeight: 700, color: '#17273b', marginTop: 16 }}>{t('auth.resetOk')}</h2>
          <p style={{ fontSize: 14, color: '#708499', marginTop: 8 }}>{t('auth.resetOkSub')}</p>
          <button className="auth-go" onClick={() => setAuthView('login')} style={{ marginTop: 22 }}>{t('auth.goLogin')}</button>
        </div>
      </>
    );
  }

  return (
    <>
      <Mac />
      <Logo variant="lockup" height={34} className="auth-brand" />
      <h1 className="auth-h">{t('auth.resetPwd')}</h1>
      <div className="auth-fields">
        <Pill placeholder={t('auth.accountPh')} value={account} onChange={v => { setAccount(v); setError(''); }} error={!!error && !account}
          right={<button className="hc-btn-code" style={{ marginRight: 2 }} onClick={handleSend} disabled={cd.running || (!isPhone && !isEmail)}>{cd.running ? `${cd.seconds}s` : t('auth.getCode')}</button>} />
        <Pill placeholder={t('auth.code')} value={code} onChange={v => { setCode(v); setError(''); }} error={!!error && !code} maxLength={6} />
        <Pill type={show ? 'text' : 'password'} placeholder={t('auth.newPwd')} value={pwd} onChange={v => { setPwd(v); setError(''); }} error={!!error && !!pwd && !!pwdOk(pwd, t)}
          right={<EyeBtn show={show} toggle={() => setShow(!show)} />} />
        <Pill type={show ? 'text' : 'password'} placeholder={t('auth.confirmNewPwd')} value={pwd2} onChange={v => { setPwd2(v); setError(''); }} error={!!error && !!pwd2 && pwd !== pwd2} />
        {error && <Err>{error}</Err>}
        {sent && !error && <Sent>{t('auth.codeSent')}</Sent>}
        <button className="auth-go" onClick={submit} disabled={loading || !account || !code || !pwd || !pwd2}>
          {loading ? <Loader2 size={18} className="animate-spin" /> : t('auth.resetPwd')}
        </button>
      </div>
      <div className="auth-corner" />
      <button className="auth-change" onClick={() => setAuthView('login')}>{t('auth.toLogin')}</button>
    </>
  );
}

/* ═══════════════════════════════════════════════
   MAIN — full-screen gradient stage + rotating card
   ═══════════════════════════════════════════════ */

const ORDER = ['login', 'register', 'forgot-password'] as const;

export default function AuthPage() {
  const { authView } = useIMStore();
  const t = useT();
  const [mounted, setMounted] = useState(false);
  const [cardH, setCardH] = useState<number | undefined>(undefined);
  const boxRefs = useRef<Record<string, HTMLDivElement | null>>({});

  useEffect(() => { const id = requestAnimationFrame(() => setMounted(true)); return () => cancelAnimationFrame(id); }, []);

  // Animate the shell height to fit whichever box is active. The box is
  // content-height (not stretched to the shell), so measuring it never feeds
  // back into the height we set — no resize loop.
  useEffect(() => {
    const el = boxRefs.current[authView];
    if (!el) return;
    const update = () => setCardH(el.scrollHeight);
    update();
    const ro = new ResizeObserver(update);
    ro.observe(el);
    return () => ro.disconnect();
  }, [authView]);

  const activeIdx = ORDER.indexOf(authView as typeof ORDER[number]);

  const boxes: { key: typeof ORDER[number]; dark?: boolean; node: React.ReactNode }[] = [
    { key: 'login', dark: true, node: <LoginBox /> },
    { key: 'register', node: <RegisterBox /> },
    { key: 'forgot-password', node: <ResetBox /> },
  ];

  return (
    <div className="auth-stage">
      <LangSwitch />
      <div className={`auth-shell${mounted ? ' auth-shell-show' : ''}`} style={{ height: cardH }}>
        {boxes.map((b, i) => {
          const pos = i < activeIdx ? 'is-before' : i > activeIdx ? 'is-after' : 'is-active';
          return (
            <div
              key={b.key}
              ref={el => { boxRefs.current[b.key] = el; }}
              className={`auth-box ${b.dark ? 'auth-box-dark' : 'auth-box-light'} ${pos}`}
              aria-hidden={pos !== 'is-active'}
            >
              {b.node}
            </div>
          );
        })}
      </div>
      <footer className="auth-footer">
        <Logo variant="mark" height={16} />
        <span>{t('auth.copyright')}</span>
      </footer>
    </div>
  );
}
