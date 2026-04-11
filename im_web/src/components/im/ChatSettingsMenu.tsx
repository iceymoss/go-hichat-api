'use client';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Switch } from '@/components/ui/switch';
import {
  User,
  Search,
  BellOff,
  Pin,
  Image as ImageIcon,
  FolderOpen,
  Trash2,
  Ban,
  Flag,
} from 'lucide-react';
import type { ReactNode } from 'react';

interface ChatSettingsMenuProps {
  conversation: {
    id: string;
    name: string;
    muted: boolean;
    pinned: boolean;
  };
  onMutedChange: (muted: boolean) => void;
  onPinnedChange: (pinned: boolean) => void;
  onClearChat: () => void;
  onSearchHistory: () => void;
  children: ReactNode;
}

const DEFAULT_COLOR = '#1C2733';
const DANGER_COLOR = '#E53935';
const ICON_COLOR = '#708499';
const DANGER_ICON_COLOR = '#E53935';

export default function ChatSettingsMenu({
  conversation,
  onMutedChange,
  onPinnedChange,
  onClearChat,
  onSearchHistory,
  children,
}: ChatSettingsMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{children}</DropdownMenuTrigger>
      <DropdownMenuContent
        sideOffset={8}
        align="end"
        style={{
          background: '#FFFFFF',
          borderRadius: '12px',
          boxShadow: '0 4px 24px rgba(0,0,0,0.12)',
          padding: '6px 0',
          minWidth: '220px',
          border: 'none',
          overflow: 'hidden',
        }}
      >
        {/* Group 1 — 会话信息与搜索 */}
        <MenuItem icon={<User />} label="查看资料" />

        <MenuItem
          icon={<Search />}
          label="查找聊天记录"
          onSelect={onSearchHistory}
        />

        {/* Group 2 — 个性化与通知 */}
        <DropdownMenuSeparator
          style={{
            height: '1px',
            background: '#F0F0F0',
            margin: '4px 0',
          }}
        />

        <SwitchMenuItem
          icon={<BellOff />}
          label="消息免打扰"
          checked={conversation.muted}
          onCheckedChange={onMutedChange}
        />

        <SwitchMenuItem
          icon={<Pin />}
          label="置顶会话"
          checked={conversation.pinned}
          onCheckedChange={onPinnedChange}
        />

        <MenuItem icon={<ImageIcon />} label="设置当前聊天背景" />

        {/* Group 3 — 媒体与文件 */}
        <DropdownMenuSeparator
          style={{
            height: '1px',
            background: '#F0F0F0',
            margin: '4px 0',
          }}
        />

        <MenuItem icon={<FolderOpen />} label="共享文件/媒体" />

        {/* Red divider before danger zone */}
        <DropdownMenuSeparator
          style={{
            height: '1px',
            background: '#F0F0F0',
            margin: '4px 0',
          }}
        />

        {/* Group 4 — 危险操作 */}
        <MenuItem
          icon={<Trash2 />}
          label="清空聊天记录"
          danger
          onSelect={onClearChat}
        />
        <MenuItem icon={<Ban />} label="拉黑该用户" danger />
        <MenuItem icon={<Flag />} label="举报..." danger />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/* ---------- helpers ---------- */

interface MenuItemProps {
  icon: ReactNode;
  label: string;
  danger?: boolean;
  onSelect?: () => void;
}

function MenuItem({ icon, label, danger, onSelect }: MenuItemProps) {
  return (
    <DropdownMenuItem
      onSelect={onSelect}
      style={{
        height: '40px',
        padding: '0 16px',
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
        fontSize: '14px',
        color: danger ? DANGER_COLOR : DEFAULT_COLOR,
        cursor: 'pointer',
        borderRadius: '0',
        outline: 'none',
        border: 'none',
      }}
      className="[&_svg]:!size-[18px]"
    >
      <span
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: '18px',
          height: '18px',
          minWidth: '18px',
          color: danger ? DANGER_ICON_COLOR : ICON_COLOR,
        }}
      >
        {icon}
      </span>
      <span style={{ flex: 1 }}>{label}</span>
    </DropdownMenuItem>
  );
}

interface SwitchMenuItemProps {
  icon: ReactNode;
  label: string;
  checked: boolean;
  onCheckedChange: (value: boolean) => void;
}

function SwitchMenuItem({
  icon,
  label,
  checked,
  onCheckedChange,
}: SwitchMenuItemProps) {
  return (
    <DropdownMenuItem
      onSelect={(e) => e.preventDefault()}
      style={{
        height: '40px',
        padding: '0 16px',
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
        fontSize: '14px',
        color: DEFAULT_COLOR,
        cursor: 'pointer',
        borderRadius: '0',
        outline: 'none',
        border: 'none',
      }}
      className="[&_svg]:!size-[18px]"
    >
      <span
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: '18px',
          height: '18px',
          minWidth: '18px',
          color: ICON_COLOR,
        }}
      >
        {icon}
      </span>
      <span style={{ flex: 1 }}>{label}</span>
      <Switch
        checked={checked}
        onCheckedChange={onCheckedChange}
        onClick={(e) => e.stopPropagation()}
      />
    </DropdownMenuItem>
  );
}
