'use client';

import React from 'react';
import { ArrowLeft, ChevronRight } from 'lucide-react';
import { toast } from 'sonner';
import Logo from '@/components/brand/Logo';
import { useT } from '@/hooks/use-i18n';

/** Hardcoded app version (no backend update check yet). */
const APP_VERSION = '2.0.0';

function PageHeader({ title, onBack }: { title: string; onBack: () => void }) {
  return (
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
      <span style={{ flex: 1, textAlign: 'center', fontSize: 17, fontWeight: 600, color: '#1C2733', marginRight: 52 }}>{title}</span>
    </div>
  );
}

function LinkRow({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <div className="im-profile-menu-item" onClick={onClick} style={{ padding: '13px 16px' }}>
      <span style={{ flex: 1, fontSize: 14, color: '#1C2733' }}>{label}</span>
      <ChevronRight size={16} style={{ color: '#A2ACB5' }} />
    </div>
  );
}

const Divider = () => <div style={{ height: 1, background: 'rgba(0,0,0,0.06)', marginLeft: 16 }} />;

export default function AboutPage({ onBack }: { onBack: () => void }) {
  const t = useT();

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#F0F2F5' }}>
      <PageHeader title={t('about.title')} onBack={onBack} />
      <div style={{ flex: 1, overflowY: 'auto' }} className="im-scroll">
        {/* Brand block */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '40px 16px 28px' }}>
          <Logo variant="mark" height={72} style={{ borderRadius: 18 }} />
          <div style={{ fontSize: 20, fontWeight: 700, color: '#1C2733', marginTop: 16 }}>HiChat</div>
          <div style={{ fontSize: 13, color: '#8F959E', marginTop: 6 }}>{t('about.slogan')}</div>
          <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 10 }}>{t('about.version')} {APP_VERSION}</div>
        </div>

        {/* Check for updates */}
        <div className="mx-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <LinkRow label={t('about.checkUpdate')} onClick={() => toast.success(t('about.upToDate'))} />
          </div>
        </div>

        {/* Legal links */}
        <div className="mx-3 mt-3">
          <div style={{ background: '#fff', borderRadius: 16, overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}>
            <LinkRow label={t('about.terms')} onClick={() => toast(t('common.featureWip'))} />
            <Divider />
            <LinkRow label={t('about.privacy')} onClick={() => toast(t('common.featureWip'))} />
          </div>
        </div>

        <div style={{ textAlign: 'center', fontSize: 12, color: '#A2ACB5', padding: '28px 16px 24px' }}>
          {t('about.copyright')}
        </div>
      </div>
    </div>
  );
}
