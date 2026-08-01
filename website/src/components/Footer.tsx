import Link from "next/link";
import Image from "next/image";
import { Github } from "lucide-react";
import { localeHref, type Locale, type SiteContent } from "@/i18n";
import { links } from "@/i18n/shared";

interface Props {
  locale: Locale;
  content: SiteContent;
}

export function Footer({ locale, content }: Props) {
  const currentYear = new Date().getFullYear();
  const { footerGroups, site, ui } = content;
  const href = (path: string) => localeHref(locale, path);

  return (
    <footer className="border-t border-border bg-background">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          {/* Brand column */}
          <div className="col-span-2 md:col-span-1">
            <Link href={href("/")} className="inline-flex items-center gap-2">
              <Image
                src="/brand/lockup-dark.svg"
                alt={site.name}
                width={108}
                height={28}
                className="h-7 w-auto"
              />
            </Link>
            <p className="mt-3 max-w-xs text-sm leading-relaxed text-muted-foreground">
              {ui.footerBlurb}
            </p>
            <a
              href={links.github}
              target="_blank"
              rel="noopener noreferrer"
              aria-label={ui.githubRepo}
              className="mt-4 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              <Github className="size-4" />
              iceymoss/go-hichat-api
            </a>
          </div>

          {/* Link groups */}
          {footerGroups.map((group) => (
            <div key={group.title}>
              <h3 className="text-sm font-semibold text-foreground">
                {group.title}
              </h3>
              <ul className="mt-3 flex flex-col gap-2">
                {group.items.map((item) => (
                  <li key={item.label}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                      >
                        {item.label}
                      </a>
                    ) : (
                      <Link
                        href={href(item.href)}
                        className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                      >
                        {item.label}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom row */}
        <div className="mt-10 flex flex-col items-start justify-between gap-3 border-t border-border pt-6 sm:flex-row sm:items-center">
          <p className="text-xs text-muted-foreground">
            © {currentYear} {site.name}.{" "}
            <a
              href={links.license}
              target="_blank"
              rel="noopener noreferrer"
              className="underline underline-offset-2 transition-colors hover:text-foreground"
            >
              {ui.footerLicense}
            </a>
          </p>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <span className="inline-block size-1.5 rounded-full bg-brand" />
            {site.version} · {ui.activeDevelopment}
          </div>
        </div>
      </div>
    </footer>
  );
}
