import { useSettingsStore } from '@/lib/settings-store';

/**
 * Hook to check if notifications should be shown for a new message.
 * Used by chat components to decide whether to play sound / show toast.
 */
export function useNotifySettings() {
  const { notifyEnabled, notifySound, notifyVibrate, notifyPreview } = useSettingsStore();
  return {
    shouldNotify: notifyEnabled,
    shouldPlaySound: notifyEnabled && notifySound,
    shouldVibrate: notifyEnabled && notifyVibrate,
    shouldPreview: notifyEnabled && notifyPreview,
  };
}

/**
 * Hook for privacy settings.
 * Used by search/profile components to filter visibility.
 */
export function usePrivacySettings() {
  const { allowSearchByPhone, allowSearchById, momentVisibility } = useSettingsStore();
  return { allowSearchByPhone, allowSearchById, momentVisibility };
}

/**
 * Hook for display settings.
 */
export function useDisplaySettings() {
  const { themeMode, fontSize, language } = useSettingsStore();
  return { themeMode, fontSize, language };
}
