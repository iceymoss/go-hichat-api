import type { Metadata } from "next";
import Link from "next/link";
import Image from "next/image";
import { ArrowRight, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { getContent, localeHref, localeParams, resolveLocale } from "@/i18n";
import { links, SCREENSHOT_H, SCREENSHOT_W } from "@/i18n/shared";

export function generateStaticParams() {
  return localeParams();
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  const page = getContent(locale).productPages.features;

  return { title: page.metadataTitle, description: page.metadataDescription };
}

export default async function FeaturesPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const locale = resolveLocale((await params).locale);
  const { featureSections, productPages } = getContent(locale);
  const page = productPages.features;

  return (
    <>
      {/* Page header */}
      <section className="relative overflow-hidden border-b border-border px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-x-0 top-0 flex justify-center"
        >
          <div className="h-[300px] w-[700px] rounded-full bg-brand opacity-[0.06] blur-[120px]" />
        </div>
        <div className="relative mx-auto max-w-3xl text-center">
          <p className="text-sm font-medium text-brand">{page.eyebrow}</p>
          <h1 className="mt-3 text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
            {page.title}
          </h1>
          <p className="mt-5 text-base text-muted-foreground sm:text-lg">
            {page.description}
          </p>
        </div>
      </section>

      {/* Alternating feature sections */}
      {featureSections.map((section, i) => {
        const imageFirst = i % 2 === 1;

        return (
          <section
            key={section.id}
            id={section.id}
            className={
              i % 2 === 1
                ? "border-b border-border bg-card/30 px-4 py-16 sm:px-6 sm:py-20 lg:px-8"
                : "border-b border-border px-4 py-16 sm:px-6 sm:py-20 lg:px-8"
            }
          >
            <div className="mx-auto grid max-w-7xl items-center gap-10 lg:grid-cols-2 lg:gap-14">
              {/* Copy column */}
              <div className={imageFirst ? "lg:order-2" : undefined}>
                <p className="text-sm font-medium text-brand">
                  {section.eyebrow}
                </p>
                <h2 className="mt-2 text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                  {section.title}
                </h2>
                <p className="mt-4 text-base leading-relaxed text-muted-foreground">
                  {section.blurb}
                </p>

                <ul className="mt-6 grid gap-2.5 sm:grid-cols-2">
                  {section.capabilities.map((cap) => (
                    <li key={cap} className="flex items-start gap-2">
                      <Check className="mt-0.5 size-4 shrink-0 text-brand" />
                      <span className="text-sm text-muted-foreground">
                        {cap}
                      </span>
                    </li>
                  ))}
                </ul>

                <p className="mt-6 inline-flex items-center gap-2 rounded-md border border-border bg-secondary px-2.5 py-1.5">
                   <span className="text-xs text-muted-foreground">{page.source}</span>
                  <code className="font-mono text-xs text-brand">
                    {section.service}
                  </code>
                </p>
              </div>

              {/* Screenshot column */}
              <div className={imageFirst ? "lg:order-1" : undefined}>
                <div className="overflow-hidden rounded-xl border border-border bg-card shadow-xl shadow-black/40">
                  <Image
                    src={section.shot}
                    alt={section.shotAlt}
                    width={SCREENSHOT_W}
                    height={SCREENSHOT_H}
                    sizes="(min-width: 1024px) 50vw, 100vw"
                    loading={i === 0 ? "eager" : "lazy"}
                    className="w-full"
                  />
                </div>
              </div>
            </div>
          </section>
        );
      })}

      {/* Closing CTA */}
      <section className="px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
            {page.ctaTitle}
          </h2>
          <p className="mt-4 text-base text-muted-foreground">
            {page.ctaDescription}
          </p>
          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Button size="lg" asChild>
              <Link href={localeHref(locale, "/quick-start")} className="gap-2">
                {page.quickStart}
                <ArrowRight className="size-4" />
              </Link>
            </Button>
            <Button variant="outline" size="lg" asChild>
              <a href={links.docsApi} target="_blank" rel="noopener noreferrer">
                {page.apiReference}
              </a>
            </Button>
          </div>
        </div>
      </section>
    </>
  );
}
