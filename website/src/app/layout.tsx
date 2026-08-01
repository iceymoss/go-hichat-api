import type { ReactNode } from "react";
import { Geist, Geist_Mono, Noto_Sans_SC } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

/**
 * Geist 只含 latin subset，中文会回落到系统字体导致中英页面观感不一致。
 * Noto Sans SC 作为 CJK 回落挂在同一个 font-sans 栈里。
 */
const notoSansSC = Noto_Sans_SC({
  variable: "--font-noto-sc",
  subsets: ["latin"],
  weight: ["400", "500", "700"],
  display: "swap",
});

/**
 * 根 layout 只提供 <html> / <body> 骨架。
 * lang 属性与 Navbar/Footer 由 [[...locale]]/layout.tsx 按语言决定 —— 静态导出下
 * 根 layout 拿不到 locale 参数（它在 catch-all 段之外）。
 */
export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html className="dark" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} ${notoSansSC.variable} font-sans antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
