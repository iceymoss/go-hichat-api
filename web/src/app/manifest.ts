import type { MetadataRoute } from 'next';

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: 'HiChat - 即时通讯',
    short_name: 'HiChat',
    description: '安全、快速、现代化的即时通讯平台。HiChat — 连接世界，即刻启程。',
    start_url: '/',
    display: 'standalone',
    background_color: '#15161A',
    theme_color: '#1BB45B',
    icons: [
      { src: '/brand/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
      { src: '/brand/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
      { src: '/brand/icon-maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
    ],
  };
}
