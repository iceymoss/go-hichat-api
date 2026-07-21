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

  /** Language picked on the login screen before auth; persisted on next login. */
  pendingLoginLanguage: string | null;

  // Actions
  loadFromBackend: (token: string) => Promise<void>;
  updateSetting: (key: string, value: any, token: string) => void;
  /** Set language locally (no token, e.g. on the auth page) + remember it for login. */
  setLanguage: (lang: string) => void;
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
  language: 'en',
  fontSize: 'medium',
  pendingLoginLanguage: null,

  // Load from backend on login
  loadFromBackend: async (token: string) => {
    const pending = get().pendingLoginLanguage;
    try {
      const res = await fetch('/api/user/settings', {
        headers: { Authorization: `Bearer ${token}` },
      });
      const d = await res.json();
      const base = d.success && d.data ? { ...get(), ...d.data } : { ...get() };
      const backendLang = d.success && d.data ? d.data.language : undefined;
      // The language chosen on the login screen wins over the stored one.
      const merged: any = { ...base, loaded: true, pendingLoginLanguage: null };
      if (pending) merged.language = pending;
      set(merged);
      applyAll(merged);
      // Persist the login-screen choice to the user's settings (full payload,
      // so other settings loaded above are preserved).
      if (pending && pending !== backendLang) {
        saveToBackend(merged as SettingsState, token);
      }
    } catch {
      set({ loaded: true, pendingLoginLanguage: null });
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

  // Set language locally (pre-auth, no token) and remember it for login.
  setLanguage: (lang: string) => {
    set({ language: lang, pendingLoginLanguage: lang });
    applyLanguage(lang);
  },
}));
