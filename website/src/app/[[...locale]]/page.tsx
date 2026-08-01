import { Hero } from "@/components/Hero";
import { FeatureGrid } from "@/components/FeatureGrid";
import { ScreenshotGallery } from "@/components/ScreenshotGallery";
import { Architecture } from "@/components/Architecture";
import { QuickStartCTA } from "@/components/QuickStartCTA";
import { getContent, localeParams, resolveLocale } from "@/i18n";

export function generateStaticParams() {
  return localeParams();
}

export default async function Home({
  params,
}: {
  params: Promise<{ locale?: string[] }>;
}) {
  const locale = resolveLocale((await params).locale);
  const content = getContent(locale);

  return (
    <>
      <Hero locale={locale} content={content} />
      <FeatureGrid content={content} />
      <ScreenshotGallery content={content} />
      <Architecture content={content} />
      <QuickStartCTA locale={locale} content={content} />
    </>
  );
}
