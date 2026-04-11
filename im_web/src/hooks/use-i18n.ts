import { useSettingsStore } from '@/lib/settings-store';
import { t as translate } from '@/lib/i18n';

export function useT() {
  const lang = useSettingsStore(s => s.language);
  return (key: string) => translate(key, lang);
}
