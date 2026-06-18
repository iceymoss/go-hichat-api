import sharp from 'sharp';
import { readFileSync } from 'node:fs';

const src = readFileSync(new URL('../../assets/brand/hichat-green-appicon.svg', import.meta.url));
const out = new URL('../public/brand/', import.meta.url);
const app = new URL('../src/app/', import.meta.url);

// Rounded app-icon tiles (corners kept) — favicons / apple / standard PWA.
const rounded = [
  [new URL('icon.png', app), 256],
  [new URL('apple-icon.png', app), 180],
  [new URL('icon-192.png', out), 192],
  [new URL('icon-512.png', out), 512],
];
for (const [dest, size] of rounded) {
  await sharp(src, { density: 384 }).resize(size, size).png().toFile(dest.pathname);
}

// Maskable: full-bleed green square (OS applies its own mask) — flatten the
// rounded transparent corners onto the brand green so no transparency leaks.
await sharp(src, { density: 384 })
  .resize(512, 512)
  .flatten({ background: '#1BB45B' })
  .png()
  .toFile(new URL('icon-maskable-512.png', out).pathname);

console.log('icons generated');
