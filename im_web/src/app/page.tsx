'use client';

import { useState, useEffect } from 'react';
import { useIMStore } from '@/lib/im-store';
import IMLayout from '@/components/im/IMLayout';
import AuthPage from '@/components/auth/AuthPage';
import SettingsProvider from '@/components/SettingsProvider';

export default function Home() {
  const { isAuthenticated } = useIMStore();
  const [hydrated, setHydrated] = useState(false);

  // Wait for zustand persist to rehydrate from localStorage before rendering
  useEffect(() => {
    setHydrated(true);
  }, []);

  if (!hydrated) {
    return null;
  }

  if (!isAuthenticated) {
    return <AuthPage />;
  }

  return (
    <SettingsProvider>
      <main className="h-dvh overflow-hidden">
        <IMLayout />
      </main>
    </SettingsProvider>
  );
}
