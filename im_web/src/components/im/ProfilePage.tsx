'use client';

import React from 'react';
import { currentUser, profileMenuItems } from '@/lib/mock-data';
import { useIMStore } from '@/lib/im-store';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import {
  Star,
  Image as ImageIcon,
  CreditCard,
  Smile,
  Settings,
  ChevronRight,
  QrCode,
  Moon,
  Bell,
  Shield,
  Lock,
  Info,
  HelpCircle,
  MessageSquare,
  Wallet,
  BookOpen,
  FileText,
  LogOut,
  Repeat2,
} from 'lucide-react';

const iconMap: Record<string, React.ReactNode> = {
  Star: <Star className="w-5 h-5" />,
  Image: <ImageIcon className="w-5 h-5" />,
  CreditCard: <CreditCard className="w-5 h-5" />,
  Smile: <Smile className="w-5 h-5" />,
  Settings: <Settings className="w-5 h-5" />,
};

export default function ProfilePage() {
  const { currentUser: authUser, logout, setMeSubPage, meSubPage } = useIMStore();

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
              background: meSubPage === 'profile' ? 'rgba(51,144,236,0.08)' : '#FFFFFF',
              boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
            }}
            onClick={() => setMeSubPage('profile')}
          >
            <Avatar className="w-16 h-16 shrink-0" style={{ ringColor: 'rgba(51,144,236,0.3)' }}>
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
                  style={{ color: '#3390EC' }}
                >
                  HiChat ID: {displayUserId}
                </span>
                <QrCode className="w-3.5 h-3.5" style={{ color: '#3390EC' }} />
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
                style={{ background: 'rgba(51,144,236,0.1)' }}
              >
                <Wallet className="w-4 h-4" style={{ color: '#3390EC' }} />
              </div>
              <span className="text-sm" style={{ color: '#1C2733' }}>服务</span>
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
          {profileMenuItems.map((item, idx) => (
            <React.Fragment key={item.label}>
              {idx > 0 && (
                <div
                  className="mx-3"
                  style={{ height: '1px', background: 'rgba(0,0,0,0.06)' }}
                />
              )}
              <div className="im-profile-menu-item"
                onClick={() => {
                  if (item.label === '收藏') setMeSubPage('favorites');
                  if (item.label === '相册') setMeSubPage('album');
                  if (item.label === '表情') setMeSubPage('emojis');
                }}
              >
                <div className="flex items-center gap-3">
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center"
                    style={{ background: 'rgba(51,144,236,0.1)', color: '#3390EC' }}
                  >
                    {iconMap[item.icon]}
                  </div>
                  <span className="text-sm" style={{ color: '#1C2733' }}>{item.label}</span>
                </div>
                <div className="flex items-center gap-2">
                  {item.value && (
                    <span className="text-xs" style={{ color: '#708499' }}>{item.value}</span>
                  )}
                  <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                </div>
              </div>
            </React.Fragment>
          ))}
        </div>

        {/* Spacer */}
        <div className="h-3" />

        {/* Settings Section */}
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{
            background: '#FFFFFF',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          {[
            { icon: <Settings className="w-4 h-4" />, label: '设置' },
            { icon: <Moon className="w-4 h-4" />, label: '深色模式', value: '跟随系统' },
            { icon: <Bell className="w-4 h-4" />, label: '消息通知' },
            { icon: <Shield className="w-4 h-4" />, label: '隐私' },
            { icon: <Lock className="w-4 h-4" />, label: '账号与安全' },
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
              <div className="im-profile-menu-item">
                <div className="flex items-center gap-3">
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center"
                    style={{ background: 'rgba(51,144,236,0.1)', color: '#3390EC' }}
                  >
                    {item.icon}
                  </div>
                  <span className="text-sm" style={{ color: '#1C2733' }}>{item.label}</span>
                </div>
                <div className="flex items-center gap-2">
                  {item.value && (
                    <span className="text-xs" style={{ color: '#708499' }}>{item.value}</span>
                  )}
                  <ChevronRight className="w-4 h-4" style={{ color: '#A2ACB5' }} />
                </div>
              </div>
            </React.Fragment>
          ))}
        </div>

        {/* Spacer */}
        <div className="h-3" />

        {/* Tools Section */}
        <div
          className="mx-4 rounded-xl overflow-hidden"
          style={{
            background: '#FFFFFF',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
        >
          {[
            { icon: <BookOpen className="w-4 h-4" />, label: '通用' },
            { icon: <FileText className="w-4 h-4" />, label: '聊天记录备份与迁移' },
            { icon: <HelpCircle className="w-4 h-4" />, label: '帮助与反馈' },
            { icon: <MessageSquare className="w-4 h-4" />, label: '关于' },
            { icon: <Info className="w-4 h-4" />, label: '插件' },
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
              <div className="im-profile-menu-item">
                <div className="flex items-center gap-3">
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center"
                    style={{ background: 'rgba(51,144,236,0.1)', color: '#3390EC' }}
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

        {/* Spacer */}
        <div className="h-3" />

        {/* Switch Account / Logout Buttons */}
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
            切换账号
          </button>
          <button
            onClick={logout}
            className="w-full py-3 rounded-xl text-sm text-center flex items-center justify-center gap-2 transition-colors cursor-pointer"
            style={{
              color: '#E53935',
            }}
          >
            <LogOut className="w-4 h-4" />
            退出登录
          </button>
        </div>
      </div>
    </div>
  );
}
