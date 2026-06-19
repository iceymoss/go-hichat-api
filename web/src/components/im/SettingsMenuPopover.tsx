'use client';

import type { ReactNode } from 'react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Settings, Info, HelpCircle, Repeat2, LogOut } from 'lucide-react';
import { useIMStore } from '@/lib/im-store';
import { useT } from '@/hooks/use-i18n';
import { toast } from 'sonner';

const DEFAULT_COLOR = '#1C2733';
const DANGER_COLOR = '#E53935';
const ICON_COLOR = '#708499';

interface SettingsMenuPopoverProps {
  /** The sidebar gear button — used as the dropdown trigger. */
  children: ReactNode;
}

/**
 * Popover menu anchored to the sidebar settings gear. Entry point for
 * system/app settings (settings panel, about, help, account actions).
 */
export default function SettingsMenuPopover({ children }: SettingsMenuPopoverProps) {
  const { setSystemPanel, logout } = useIMStore();
  const t = useT();
  const wip = () => toast(t('common.featureWip'));

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{children}</DropdownMenuTrigger>
      <DropdownMenuContent
        side="right"
        align="end"
        sideOffset={8}
        style={{
          background: '#FFFFFF',
          borderRadius: '12px',
          boxShadow: '0 4px 24px rgba(0,0,0,0.12)',
          padding: '6px 0',
          minWidth: '200px',
          border: 'none',
          overflow: 'hidden',
        }}
      >
        <MenuItem icon={<Settings />} label={t('profile.settings')} onSelect={() => setSystemPanel('settings')} />
        <MenuItem icon={<Info />} label={t('profile.about')} onSelect={() => setSystemPanel('about')} />
        <MenuItem icon={<HelpCircle />} label={t('profile.help')} onSelect={wip} />

        <DropdownMenuSeparator style={{ height: '1px', background: '#F0F0F0', margin: '4px 0' }} />

        <MenuItem icon={<Repeat2 />} label={t('profile.switchAccount')} onSelect={wip} />
        <MenuItem icon={<LogOut />} label={t('profile.logout')} danger onSelect={logout} />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/* ---------- helpers ---------- */

function MenuItem({ icon, label, danger, onSelect }: {
  icon: ReactNode; label: string; danger?: boolean; onSelect?: () => void;
}) {
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
          color: danger ? DANGER_COLOR : ICON_COLOR,
        }}
      >
        {icon}
      </span>
      <span style={{ flex: 1 }}>{label}</span>
    </DropdownMenuItem>
  );
}
