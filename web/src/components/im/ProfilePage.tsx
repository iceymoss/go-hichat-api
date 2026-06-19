'use client';

import React from 'react';
import { currentUser } from '@/lib/types';
import { useIMStore } from '@/lib/im-store';
import { useT } from '@/hooks/use-i18n';
import { useIsMobile } from '@/hooks/use-mobile';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import {
  Star,
  Image as ImageIcon,
  CreditCard,
  Smile,
  Settings,
  ChevronRight,
  QrCode,
  Info,
  HelpCircle,
  MessageSquare,
  Wallet,
  FileText,
  LogOut,
  Repeat2,
  Aperture,
} from 'lucide-react';

const iconMap: Record<string, React.ReactNode> = {
  Star: <Star className="w-5 h-5" />,
  Image: <ImageIcon className="w-5 h-5" />,
  CreditCard: <CreditCard className="w-5 h-5" />,
  Smile: <Smile className="w-5 h-5" />,
  Settings: <Settings className="w-5 h-5" />,
};

export default function ProfilePage() {
  const { currentUser: authUser, logout, setMeSubPage, meSubPage, openUserTrends, setSystemPanel } = useIMStore();
  const t = useT();
  const isMobile = useIsMobile();

  const displayName = authUser?.name || currentUser.name;
  const displayAvatar = authUser?.avatar || currentUser.avatar;
  const displayUserId = authUser?.id || '未登录';
  const displaySignature = authUser?.introduction || currentUser.signature;

  return (
    <div className="h-full flex flex-col">
      {/* Profile Content */}
      <div className="flex-1 overflow-y-auto im-scroll">
        {/* Spacer top */}
        <div className="h-3" />

        {/* User Profile Card — click to open edit page */}
        <div className="mx-4">
          <div
            className="flex items-center gap-4 p-4 rounded-xl cursor-pointer transition-colors"
            style={{
              background: meSubPage === 'profile' ? 'rgba(27,180,91,0.08)' : '#FFFFFF',
              boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
            }}
            onClick={() => setMeSubPage('profile')}
          >
            <Avatar className="w-16 h-16 shrink-0" style={{ ringColor: 'rgba(27,180,91,0.3)' }}>
              <AvatarImage src={displayAvatar} alt={displayName} />
              <AvatarFallback className="text-lg">{displayName[0]}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <h3
                className="text-lg font-bold truncate"
                style={{ color: '#1C2733' }}
              >
                {displayName}
              </h3>
              <div className="flex items-center gap-1.5 mt-1">
                <span
                  className="text-sm"
                  style={{ color: '#1BB45B' }}
                >
                  HiChat ID: {displayUserId}
                </span>
                <QrCode className="w-3.5 h-3.5" style={{ color: '#1BB45B' }} />
              </div>
              <p
                className="text-xs mt-1 truncate"
                style={{ color: '#708499' }}
              >
                {displaySignature}
              </p>
            </div>
            <ChevronRight className="w-4 h-4 shrink-0" style={{ color: '#A2ACB5' }} />
          </div>
        </div>

        {/* Spacer between sections */}
        <div className="h-3" />

        {/* Service Item */}
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{
            background: '#FFFFFF',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          <div className="im-profile-menu-item">
            <div className="flex items-center gap-3">
              <div
                className="w-8 h-8 rounded-lg flex items-center justify-center"
                style={{ background: 'rgba(27,180,91,0.1)' }}
              >
                <Wallet className="w-4 h-4" style={{ color: '#1BB45B' }} />
              </div>
              <span className="text-sm" style={{ color: '#1C2733' }}>{t('profile.service')}</span>
            </div>
            <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
          </div>
        </div>

        {/* Spacer */}
        <div className="h-3" />

        {/* My Moments (我的朋友圈) */}
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{ background: '#FFFFFF', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}
        >
          <div
            className="im-profile-menu-item"
            onClick={() => { if (authUser?.id) openUserTrends(authUser.id, displayName); }}
          >
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg flex items-center justify-center"
                style={{ background: 'rgba(27,180,91,0.1)', color: '#1BB45B' }}>
                <Aperture className="w-5 h-5" />
              </div>
              <span className="text-sm" style={{ color: '#1C2733' }}>{t('profile.moments')}</span>
            </div>
            <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
          </div>
        </div>

        {/* Spacer */}
        <div className="h-3" />

        {/* Favorites, Albums, Cards, Stickers */}
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{
            background: '#FFFFFF',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          {[
            { icon: <Star className="w-5 h-5" />, label: t('profile.favorites'), action: 'favorites' as const },
            { icon: <ImageIcon className="w-5 h-5" />, label: t('profile.album'), action: 'album' as const },
            { icon: <CreditCard className="w-5 h-5" />, label: t('profile.cards'), action: null },
            { icon: <Smile className="w-5 h-5" />, label: t('profile.emojis'), action: 'emojis' as const },
          ].map((item, idx) => (
            <React.Fragment key={idx}>
              {idx > 0 && <div className="mx-3" style={{ height: '1px', background: 'rgba(0,0,0,0.06)' }} />}
              <div className="im-profile-menu-item"
                onClick={() => { if (item.action) setMeSubPage(item.action); }}
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg flex items-center justify-center"
                    style={{ background: 'rgba(27,180,91,0.1)', color: '#1BB45B' }}>
                    {item.icon}
                  </div>
                  <span className="text-sm" style={{ color: '#1C2733' }}>{item.label}</span>
                </div>
                <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
              </div>
            </React.Fragment>
          ))}
        </div>

        {/* Spacer */}
        <div className="h-3" />

        {/* Settings — single entry (mobile only; desktop uses the sidebar gear) */}
        {isMobile && (
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{ background: '#FFFFFF', boxShadow: '0 1px 3px rgba(0,0,0,0.06)' }}
        >
          <div className="im-profile-menu-item" onClick={() => setSystemPanel('settings')}>
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg flex items-center justify-center"
                style={{ background: 'rgba(27,180,91,0.1)', color: '#1BB45B' }}>
                <Settings className="w-4 h-4" />
              </div>
              <span className="text-sm" style={{ color: '#1C2733' }}>{t('profile.settings')}</span>
            </div>
            <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
          </div>
        </div>
        )}

        {/* Spacer */}
        {isMobile && <div className="h-3" />}

        {/* Tools Section (mobile only; desktop uses the sidebar gear) */}
        {isMobile && (
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{
            background: '#FFFFFF',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          {[
            { icon: <FileText className="w-4 h-4" />, label: t('profile.backup'), onClick: undefined },
            { icon: <HelpCircle className="w-4 h-4" />, label: t('profile.help'), onClick: undefined },
            { icon: <MessageSquare className="w-4 h-4" />, label: t('profile.about'), onClick: () => setSystemPanel('about') },
            { icon: <Info className="w-4 h-4" />, label: t('profile.plugins'), onClick: undefined },
          ].map((item, idx) => (
            <React.Fragment key={item.label}>
              {idx > 0 && (
                <div
                  className="mx-3"
                  style={{
                    height: '1px',
                    background: 'rgba(0,0,0,0.06)',
                  }}
                />
              )}
              <div className="im-profile-menu-item" onClick={item.onClick}>
                <div className="flex items-center gap-3">
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center"
                    style={{ background: 'rgba(27,180,91,0.1)', color: '#1BB45B' }}
                  >
                    {item.icon}
                  </div>
                  <span className="text-sm" style={{ color: '#1C2733' }}>{item.label}</span>
                </div>
                <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
              </div>
            </React.Fragment>
          ))}
        </div>
        )}

        {/* Spacer */}
        {isMobile && <div className="h-3" />}

        {/* Switch Account / Logout Buttons (mobile only; desktop uses the sidebar gear) */}
        {isMobile && (
        <div className="mx-4 mb-6 space-y-2">
          <button
            className="w-full py-3 rounded-xl text-sm text-center flex items-center justify-center gap-2 transition-colors cursor-pointer"
            style={{
              border: '1px solid rgba(0,0,0,0.1)',
              background: '#FFFFFF',
              color: '#1C2733',
              boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
            }}
          >
            <Repeat2 className="w-4 h-4" style={{ color: '#708499' }} />
            {t('profile.switchAccount')}
          </button>
          <button
            onClick={logout}
            className="w-full py-3 rounded-xl text-sm text-center flex items-center justify-center gap-2 transition-colors cursor-pointer"
            style={{
              color: '#E53935',
            }}
          >
            <LogOut className="w-4 h-4" />
            {t('profile.logout')}
          </button>
        </div>
        )}
      </div>
    </div>
  );
}
