import Link from "next/link";
import Image from "next/image";
import { ArrowRight, Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { localeHref, type Locale, type SiteContent } from "@/i18n";
import { links, SCREENSHOT_H, SCREENSHOT_W } from "@/i18n/shared";

interface Props {
  locale: Locale;
  content: SiteContent;
}

export function Hero({ locale, content }: Props) {
  const { site, ui } = content;

  return (
    <section className="relative overflow-hidden px-4 pb-0 pt-20 sm:px-6 sm:pt-28 lg:px-8">
      {/* Background glow */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 top-0 flex justify-center"
      >
        <div className="h-[420px] w-[900px] rounded-full bg-brand opacity-[0.07] blur-[120px]" />
      </div>

      <div className="relative mx-auto max-w-5xl">
        {/* Badge */}
        <div className="flex justify-center">
          <a
            href={links.github}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 rounded-full border border-border bg-secondary px-3 py-1 text-xs text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
          >
            <span className="inline-block size-1.5 rounded-full bg-brand" />
            {site.name} {site.version} · {ui.activeDevelopment}
            <ArrowRight className="size-3" />
          </a>
        </div>

        {/* Heading */}
        <h1 className="mt-6 text-center text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl xl:text-7xl">
          <span className="bg-gradient-to-b from-foreground to-muted-foreground bg-clip-text text-transparent">
            {ui.heroTitleLine1}
          </span>
          <br />
          <span className="bg-gradient-to-b from-foreground to-muted-foreground bg-clip-text text-transparent">
            {ui.heroTitleLine2}
          </span>
        </h1>

        <p className="mx-auto mt-5 max-w-2xl text-center text-base text-muted-foreground sm:text-lg">
          {ui.heroSubtitle}
        </p>

        {/* CTAs */}
        <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
          <Button size="lg" asChild>
            <Link href={localeHref(locale, "/quick-start")} className="gap-2">
              {ui.quickStart}
              <ArrowRight className="size-4" />
            </Link>
          </Button>
          <Button variant="outline" size="lg" asChild>
            <a
              href={links.github}
              target="_blank"
              rel="noopener noreferrer"
              className="gap-2"
            >
              <Github className="size-4" />
              {ui.starOnGithub}
            </a>
          </Button>
        </div>

        {/* Screenshot */}
        <div className="mt-14 sm:mt-16">
          <div className="overflow-hidden rounded-t-xl border border-b-0 border-border bg-secondary shadow-2xl shadow-black/60">
            <div className="flex items-center gap-2 border-b border-border bg-secondary px-4 py-3">
              <span className="size-3 rounded-full bg-border" />
              <span className="size-3 rounded-full bg-border" />
              <span className="size-3 rounded-full bg-border" />
              <div className="mx-4 flex h-5 max-w-[280px] flex-1 items-center justify-center rounded-md bg-border/40 px-3">
                <span className="truncate text-[11px] text-muted-foreground">
                  localhost:2470
                </span>
              </div>
            </div>
            <Image
              src="/screenshots/single-chat.webp"
              alt={ui.heroShotAlt}
              width={SCREENSHOT_W}
              height={SCREENSHOT_H}
              priority
              sizes="(min-width: 1024px) 1024px, 100vw"
              className="w-full"
            />
          </div>
        </div>
      </div>
    </section>
  );
}
