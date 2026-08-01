import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Navbar } from "@/components/Navbar";
import { Footer } from "@/components/Footer";
import { getContent, localeParams, resolveLocale } from "@/i18n";

export function generateStaticParams() {
  return localeParams();
}

type Props = {
  children: ReactNode;
  params: Promise<{ locale: string }>;
};

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  const { site } = getContent(locale);

  return {
    title: `${site.name} — ${site.tagline}`,
    description: site.description,
    keywords: ["hichat", "go-zero", "IM", "instant messaging", "open source", "webrtc"],
    openGraph: {
      title: `${site.name} — ${site.tagline}`,
      description: site.description,
      type: "website",
      locale: locale === "zh" ? "zh_CN" : "en_US",
    },
    alternates: {
      languages: {
        "zh-CN": "/zh",
        en: "/en",
      },
    },
  };
}

export default async function LocaleLayout({ children, params }: Props) {
  const locale = resolveLocale((await params).locale);
  const content = getContent(locale);

  return (
    <>
      {/* lang 由 script 在客户端补到 <html> 上，根 layout 无法感知动态段。 */}
      <script
        dangerouslySetInnerHTML={{
          __html: `document.documentElement.lang=${JSON.stringify(
            locale === "zh" ? "zh-CN" : "en"
          )}`,
        }}
      />
      <Navbar locale={locale} content={content} />
      <main>{children}</main>
      <Footer locale={locale} content={content} />
    </>
  );
}
