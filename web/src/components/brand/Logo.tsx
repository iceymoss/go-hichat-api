import React from 'react';

// Brand logo. Assets live in web/public/brand (synced from the canonical
// masters in <repo>/assets/brand). Use `lockup` on light surfaces, `lockup-dark`
// on dark surfaces, and `mark` for compact spots (sidebar, footer).
// Note: web/public/brand/lockup-dark.svg is a transparent, light-wordmark
// variant derived from the lockup master (the designer's chip master is kept at
// assets/brand/hichat-green-lockup-dark.svg) so it sits cleanly on dark cards.
export type LogoVariant = 'lockup' | 'lockup-dark' | 'mark';

const SRC: Record<LogoVariant, string> = {
  lockup: '/brand/lockup.svg',
  'lockup-dark': '/brand/lockup-dark.svg',
  mark: '/brand/mark.svg',
};

// Native aspect ratios from each source SVG viewBox.
const RATIO: Record<LogoVariant, number> = {
  lockup: 376 / 96,
  'lockup-dark': 376 / 96,
  mark: 1,
};

export default function Logo({
  variant = 'lockup',
  height = 32,
  className,
  style,
}: {
  variant?: LogoVariant;
  height?: number;
  className?: string;
  style?: React.CSSProperties;
}) {
  return (
    <img
      src={SRC[variant]}
      alt="HiChat"
      width={Math.round(height * RATIO[variant])}
      height={height}
      className={className}
      style={style}
      draggable={false}
    />
  );
}
