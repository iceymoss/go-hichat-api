import { create } from 'zustand';

export type ThemeMode = 'system' | 'light' | 'dark';
export type FontSize = 'small' | 'medium' | 'large';

interface SettingsState {
  loaded: boolean;

  // Theme
  themeMode: ThemeMode;
  // Notifications
  notifyEnabled: boolean;
  notifySound: boolean;
  notifyVibrate: boolean;
  notifyPreview: boolean;
  // Privacy
  allowSearchByPhone: boolean;
  allowSearchById: boolean;
  momentVisibility: 'all' | 'friends' | 'private';
  /** 消息已读回执（发送方能看到对方是否已读） */
  readReceiptEnabled: boolean;
  // General
  language: string;
  fontSize: FontSize;

  // Actions
  loadFromBackend: (token: string) => Promise<void>;
  updateSetting: (key: string, value: any, token: string) => void;
}

/* ═══════════════════════════════════════════════
   Apply settings to DOM — called on load + change
   ═══════════════════════════════════════════════ */

function applyTheme(mode: ThemeMode) {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;
  if (mode === 'dark') {
    root.classList.add('dark');
  } else if (mode === 'light') {
    root.classList.remove('dark');
  } else {
    // system
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }
}

const FONT_SIZE_MAP: Record<FontSize, string> = {
  small: '13px',
  medium: '14px',
  large: '16px',
};

function applyFontSize(size: FontSize) {
  if (typeof document === 'undefined') return;
  document.documentElement.style.setProperty('--hc-font-size', FONT_SIZE_MAP[size] || '14px');
  // Also apply as a data attribute for CSS selectors
  document.documentElement.dataset.fontSize = size;
}

function applyLanguage(lang: string) {
  if (typeof document === 'undefined') return;
  document.documentElement.lang = lang === 'en' ? 'en' : 'zh-CN';
  document.documentElement.dataset.lang = lang;
}

/** Apply all visual settings to DOM */
function applyAll(state: Partial<SettingsState>) {
  if (state.themeMode !== undefined) applyTheme(state.themeMode);
  if (state.fontSize !== undefined) applyFontSize(state.fontSize);
  if (state.language !== undefined) applyLanguage(state.language);
}

/* ═══════════════════════════════════════════════
   System theme change listener
   ═══════════════════════════════════════════════ */

if (typeof window !== 'undefined') {
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    const { themeMode } = useSettingsStore.getState();
    if (themeMode === 'system') applyTheme('system');
  });
}

/* ═══════════════════════════════════════════════
   Debounced save to backend
   ═══════════════════════════════════════════════ */

let saveTimer: any = null;

function saveToBackend(state: SettingsState, token: string) {
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(async () => {
    const payload = {
      themeMode: state.themeMode,
      notifyEnabled: state.notifyEnabled,
      notifySound: state.notifySound,
      notifyVibrate: state.notifyVibrate,
      notifyPreview: state.notifyPreview,
      allowSearchByPhone: state.allowSearchByPhone,
      allowSearchById: state.allowSearchById,
      momentVisibility: state.momentVisibility,
      readReceiptEnabled: state.readReceiptEnabled,
      language: state.language,
      fontSize: state.fontSize,
    };
    try {
      await fetch('/api/user/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(payload),
      });
    } catch { /* silent */ }
  }, 500);
}

/* ═══════════════════════════════════════════════
   Store
   ═══════════════════════════════════════════════ */

export const useSettingsStore = create<SettingsState>((set, get) => ({
  loaded: false,

  // Defaults
  themeMode: 'system',
  notifyEnabled: true,
  notifySound: true,
  notifyVibrate: true,
  notifyPreview: true,
  allowSearchByPhone: true,
  allowSearchById: true,
  momentVisibility: 'all',
  readReceiptEnabled: true,
  language: 'zh-CN',
  fontSize: 'medium',

  // Load from backend on login
  loadFromBackend: async (token: string) => {
    try {
      const res = await fetch('/api/user/settings', {
        headers: { Authorization: `Bearer ${token}` },
      });
      const d = await res.json();
      if (d.success && d.data) {
        const merged = { ...get(), ...d.data, loaded: true };
        set(merged);
        applyAll(merged);
      } else {
        set({ loaded: true });
        applyAll(get());
      }
    } catch {
      set({ loaded: true });
      applyAll(get());
    }
  },

  // Update a single setting → apply to DOM + debounced save to backend
  updateSetting: (key: string, value: any, token: string) => {
    set({ [key]: value } as any);
    // Apply visual effects immediately
    applyAll({ [key]: value } as any);
    // Debounced save
    saveToBackend({ ...get(), [key]: value }, token);
  },
}));
