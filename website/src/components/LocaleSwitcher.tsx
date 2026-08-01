"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Languages } from "lucide-react";
import type { Locale } from "@/i18n";
import { cn } from "@/lib/utils";

/**
 * 替换当前路径的语言段，保留页面位置。
 */
function swapLocale(pathname: string, target: Locale): string {
  const stripped = pathname.replace(/^\/(?:zh|en)(?=\/|$)/, "") || "/";
  return stripped === "/" ? `/${target}` : `/${target}${stripped}`;
}

interface Props {
  locale: Locale;
  label: string;
  className?: string;
}

export function LocaleSwitcher({ locale, label, className }: Props) {
  const pathname = usePathname() || "/";
  const other: Locale = locale === "zh" ? "en" : "zh";

  return (
    <Link
      href={swapLocale(pathname, other)}
      aria-label={label}
      title={label}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-lg border border-border px-2.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground",
        className
      )}
    >
      <Languages className="size-4" />
      <span className="font-medium">{other === "zh" ? "中文" : "EN"}</span>
    </Link>
  );
}
