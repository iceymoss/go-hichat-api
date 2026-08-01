import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "HiChat",
  description: "Open-source instant messaging and social platform.",
  alternates: {
    canonical: "/zh",
    languages: { "zh-CN": "/zh", en: "/en" },
  },
};

export default function RootPage() {
  return (
    <main className="flex min-h-screen items-center justify-center px-6 text-center">
      <meta httpEquiv="refresh" content="0; url=/zh" />
      <script dangerouslySetInnerHTML={{ __html: 'location.replace("/zh")' }} />
      <div>
        <p className="text-sm text-muted-foreground">正在进入 HiChat / Entering HiChat</p>
        <div className="mt-4 flex justify-center gap-4 text-sm">
          <Link className="text-brand hover:underline" href="/zh">中文</Link>
          <Link className="text-brand hover:underline" href="/en">English</Link>
        </div>
      </div>
    </main>
  );
}
