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
