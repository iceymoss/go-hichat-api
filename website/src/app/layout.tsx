import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Geist, Geist_Mono, Noto_Sans_SC } from "next/font/google";
import "./globals.css";

export const metadata: Metadata = {
  icons: {
    icon: [
      { url: "/brand/appicon.svg", type: "image/svg+xml" },
      { url: "/brand/icon-192.png", type: "image/png", sizes: "192x192" },
    ],
    apple: [{ url: "/brand/icon-192.png", sizes: "192x192" }],
  },
};

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
 * lang 属性与 Navbar/Footer 由 [locale]/layout.tsx 按语言决定，根 layout
 * 拿不到动态语言段参数。
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
