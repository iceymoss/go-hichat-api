import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "HiChat - 即时通讯",
  description: "安全、快速、现代化的即时通讯平台。HiChat — 连接世界，即刻启程。",
  keywords: ["IM", "Chat", "Messaging", "Next.js", "TypeScript"],
  authors: [{ name: "HiChat" }],
  // Favicon / apple-touch / manifest are auto-injected from the app-router
  // file conventions: app/icon.svg, app/apple-icon.png, app/manifest.ts.
};

export const viewport: Viewport = {
  themeColor: "#1BB45B",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background text-foreground`}
      >
        {children}
        <Toaster position="top-center" richColors closeButton />
      </body>
    </html>
  );
}
