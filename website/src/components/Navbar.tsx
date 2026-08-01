"use client";

import Link from "next/link";
import Image from "next/image";
import { useState } from "react";
import { Menu, X, Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { LocaleSwitcher } from "@/components/LocaleSwitcher";
import { localeHref, type Locale, type SiteContent } from "@/i18n";
import { links } from "@/i18n/shared";

interface Props {
  locale: Locale;
  content: SiteContent;
}

export function Navbar({ locale, content }: Props) {
  const [mobileOpen, setMobileOpen] = useState(false);
  const { navItems, ui } = content;
  const href = (path: string) => localeHref(locale, path);

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/80 backdrop-blur-md">
      <nav className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Logo */}
        <Link href={href("/")} className="flex shrink-0 items-center gap-2">
          <Image
            src="/brand/lockup-dark.svg"
            alt={content.site.name}
            width={120}
            height={31}
            priority
            className="h-8 w-auto"
          />
        </Link>

        {/* Desktop nav */}
        <ul className="hidden items-center gap-1 md:flex">
          {navItems.map((item) => (
            <li key={item.href}>
              <Link
                href={href(item.href)}
                className="rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
              >
                {item.label}
              </Link>
            </li>
          ))}
          <li>
            <Link
              href={href("/docs")}
              className="rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
            >
              {ui.docs}
            </Link>
          </li>
          <li>
            <Link
              href={href("/changelog")}
              className="rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
            >
              {ui.changelog}
            </Link>
          </li>
        </ul>

        {/* Desktop CTAs */}
        <div className="hidden items-center gap-2 md:flex">
          <LocaleSwitcher locale={locale} label={ui.switchLanguage} />
          <Button variant="outline" size="sm" asChild>
            <a
              href={links.github}
              target="_blank"
              rel="noopener noreferrer"
              className="gap-1.5"
            >
              <Github className="size-4" />
              {ui.github}
            </a>
          </Button>
          <Button size="sm" asChild>
            <Link href={href("/quick-start")}>{ui.quickStart}</Link>
          </Button>
        </div>

        {/* Mobile hamburger */}
        <div className="flex items-center gap-2 md:hidden">
          <LocaleSwitcher locale={locale} label={ui.switchLanguage} />
          <button
            type="button"
            aria-label={mobileOpen ? ui.closeMenu : ui.openMenu}
            onClick={() => setMobileOpen((v) => !v)}
            className="rounded-md p-2 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          >
            {mobileOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>
        </div>
      </nav>

      {/* Mobile drawer */}
      {mobileOpen && (
        <div className="border-t border-border bg-background/95 px-4 pb-4 pt-2 backdrop-blur-md md:hidden">
          <ul className="flex flex-col gap-1">
            {navItems.map((item) => (
              <li key={item.href}>
                <Link
                  href={href(item.href)}
                  onClick={() => setMobileOpen(false)}
                  className="block rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                >
                  {item.label}
                </Link>
              </li>
            ))}
            <li>
              <Link
                href={href("/docs")}
                onClick={() => setMobileOpen(false)}
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
              >
                {ui.docs}
              </Link>
            </li>
            <li>
              <Link
                href={href("/changelog")}
                onClick={() => setMobileOpen(false)}
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
              >
                {ui.changelog}
              </Link>
            </li>
          </ul>
          <div className="mt-3 flex flex-col gap-2">
            <Button variant="outline" size="sm" asChild className="justify-center">
              <a href={links.github} target="_blank" rel="noopener noreferrer">
                <Github className="size-4" />
                {ui.github}
              </a>
            </Button>
            <Button size="sm" asChild className="justify-center">
              <Link
                href={href("/quick-start")}
                onClick={() => setMobileOpen(false)}
              >
                {ui.quickStart}
              </Link>
            </Button>
          </div>
        </div>
      )}
    </header>
  );
}
