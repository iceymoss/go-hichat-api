import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/* ═══════════════════════════════════════
   Shared Avatar Color Utility
   ═══════════════════════════════════════ */

const avatarColors = [
  '#E67E22', '#E74C3C', '#3498DB', '#2ECC71', '#9B59B6',
  '#1ABC9C', '#F39C12', '#E91E63', '#00BCD4', '#FF7043',
];

/** Returns a deterministic color for a given name (used for avatar backgrounds). */
export function getAvatarColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
  return avatarColors[Math.abs(hash) % avatarColors.length];
}

/* ═══════════════════════════════════════
   Shared Tag Chip Color Utility
   ═══════════════════════════════════════ */

const tagColors: { c: string; b: string }[] = [
  { c: '#2D7FF9', b: 'rgba(45,127,249,0.10)' },   // blue
  { c: '#1BB45B', b: 'rgba(27,180,91,0.10)' },    // green
  { c: '#F59E0B', b: 'rgba(245,158,11,0.12)' },   // amber
  { c: '#9B59B6', b: 'rgba(155,89,182,0.10)' },   // purple
  { c: '#E84393', b: 'rgba(232,67,147,0.10)' },   // pink
  { c: '#14B8A6', b: 'rgba(20,184,166,0.12)' },   // teal
  { c: '#FA5151', b: 'rgba(250,81,81,0.10)' },    // red
  { c: '#EC6F1A', b: 'rgba(236,111,26,0.10)' },   // orange
];

/** Returns a deterministic {text, background} color pair for a tag string, so the
    same tag always renders the same hue while a tag list looks varied. */
export function tagColor(tag: string): { c: string; b: string } {
  let hash = 0;
  for (let i = 0; i < tag.length; i++) hash = tag.charCodeAt(i) + ((hash << 5) - hash);
  return tagColors[Math.abs(hash) % tagColors.length];
}
