'use client';

import { useIMStore } from '@/lib/im-store';
import IMLayout from '@/components/im/IMLayout';
import AuthPage from '@/components/auth/AuthPage';

export default function Home() {
  const { isAuthenticated } = useIMStore();

  if (!isAuthenticated) {
    return <AuthPage />;
  }

  return (
    <main className="h-dvh overflow-hidden">
      <IMLayout />
    </main>
  );
}
