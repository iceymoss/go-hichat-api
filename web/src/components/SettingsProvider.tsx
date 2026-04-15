'use client';

import React, { useEffect } from 'react';
import { useSettingsStore } from '@/lib/settings-store';
import { useIMStore } from '@/lib/im-store';

/**
 * SettingsProvider — wraps the app and applies visual settings.
 * - Dark mode: CSS filter invert + hue-rotate on .hc-app container
 * - Font size: CSS variable --hc-scale on .hc-app
 * - Language: data-lang attribute
 */
export default function SettingsProvider({ children }: { children: React.ReactNode }) {
  const { themeMode, fontSize, language, loaded, loadFromBackend } = useSettingsStore();
  const token = useIMStore(s => s.currentUser?.token);

  // Load settings when user is authenticated
  useEffect(() => {
    if (token && !loaded) {
      loadFromBackend(token);
    }
  }, [token, loaded, loadFromBackend]);

  // Listen for system theme changes
  useEffect(() => {
    if (themeMode !== 'system') return;
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => {
      // Force re-render by touching state
      useSettingsStore.setState({});
    };
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, [themeMode]);

  const isDark = themeMode === 'dark' || (themeMode === 'system' && typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches);

  const fontScale = fontSize === 'small' ? 0.9 : fontSize === 'large' ? 1.15 : 1;

  return (
    <div
      className="hc-app"
      data-theme={isDark ? 'dark' : 'light'}
      data-font-size={fontSize}
      data-lang={language}
      style={{
        height: '100%',
        // Font size scaling via CSS zoom (affects everything including inline styles)
        zoom: fontScale !== 1 ? fontScale : undefined,
        // Dark mode via CSS filter (affects ALL colors including inline styles)
        filter: isDark ? 'invert(1) hue-rotate(180deg)' : undefined,
        // Prevent color shift on specific elements
      }}
    >
      {children}
    </div>
  );
}
