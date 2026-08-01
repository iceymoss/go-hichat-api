import Link from "next/link";
import { ArrowRight, Terminal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { CodeBlock } from "@/components/ui/code-block";
import { localeHref, type Locale, type SiteContent } from "@/i18n";
import { demoAccount, links } from "@/i18n/shared";

const DOCKER_CMD = `git clone https://github.com/iceymoss/go-hichat-api.git
cd go-hichat-api
docker compose up -d --build`;

interface Props {
  locale: Locale;
  content: SiteContent;
}

export function QuickStartCTA({ locale, content }: Props) {
  const { techStack, ui } = content;

  return (
    <section className="border-t border-border bg-card/30 px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
      <div className="mx-auto max-w-4xl">
        {/* Tech-stack badges */}
        <div className="flex flex-wrap items-center justify-center gap-x-6 gap-y-3">
          {techStack.flatMap((g) =>
            g.items.map((item) => (
              <span
                key={item}
                className="rounded-full border border-border bg-secondary px-3 py-1 text-xs text-muted-foreground"
              >
                {item}
              </span>
            ))
          )}
        </div>

        {/* Heading */}
        <div className="mt-10 text-center">
          <div className="inline-flex items-center gap-2 text-brand">
            <Terminal className="size-5" />
            <span className="text-sm font-medium">{ui.ctaEyebrow}</span>
          </div>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            {ui.ctaTitle}
          </h2>
          <p className="mt-4 text-base text-muted-foreground">{ui.ctaSubtitle}</p>
        </div>

        <CodeBlock code={DOCKER_CMD} lang="bash" className="mt-8" />

        {/* Demo credentials */}
        <div className="mt-4 flex flex-wrap items-center justify-center gap-4 rounded-xl border border-border bg-secondary px-5 py-4">
          <span className="text-xs text-muted-foreground">
            {ui.ctaAfterStartup}{" "}
            <a
              href={demoAccount.url}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-brand hover:underline"
            >
              {demoAccount.url}
            </a>
          </span>
          <span className="hidden text-border sm:block">·</span>
          <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span>{ui.ctaDemoLogin}</span>
            <code className="rounded border border-border bg-code-bg px-1.5 py-0.5 font-mono text-xs text-foreground">
              {demoAccount.phone}
            </code>
            <span>/</span>
            <code className="rounded border border-border bg-code-bg px-1.5 py-0.5 font-mono text-xs text-foreground">
              {demoAccount.password}
            </code>
          </span>
          <span className="hidden text-border sm:block">·</span>
          <span className="text-xs text-muted-foreground">{ui.ctaAutoFill}</span>
        </div>

        {/* CTAs */}
        <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
          <Button size="lg" asChild>
            <Link href={localeHref(locale, "/quick-start")} className="gap-2">
              {ui.fullGuide}
              <ArrowRight className="size-4" />
            </Link>
          </Button>
          <Button variant="outline" size="lg" asChild>
            <a
              href={links.dockerDeploy}
              target="_blank"
              rel="noopener noreferrer"
            >
              {ui.dockerDeployDocs}
            </a>
          </Button>
        </div>
      </div>
    </section>
  );
}
