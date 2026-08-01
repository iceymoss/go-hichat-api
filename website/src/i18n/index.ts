import { zh } from "./zh";
import { en } from "./en";
import type { SiteContent } from "./types";

export const locales = ["zh", "en"] as const;
export type Locale = (typeof locales)[number];

/** 中文在根路径，无前缀；英文走 /en 前缀。 */
export const defaultLocale: Locale = "zh";

const dictionaries: Record<Locale, SiteContent> = { zh, en };

/**
 * 把 [[...locale]] 的 catch-all 段解析为 Locale。
 * undefined / [] → 中文（根路径）；["en"] → 英文。
 */
export function resolveLocale(segment?: string[]): Locale {
  const first = segment?.[0];
  return first === "en" ? "en" : "zh";
}

/** 同步字典查找 —— 服务端组件直接调用，不破坏静态预渲染。 */
export function getContent(locale: Locale): SiteContent {
  return dictionaries[locale];
}

/**
 * 生成带语言前缀的站内链接。中文返回原路径，英文加 /en。
 * 统一走这里，避免在每个文件里手写前缀拼接。
 */
export function localeHref(locale: Locale, path: string): string {
  if (locale === "zh") return path;
  return path === "/" ? "/en" : `/en${path}`;
}

/** 每个页面的 generateStaticParams 都返回这个：产出根路径与 /en 两套。 */
export function localeParams() {
  return [{ locale: [] as string[] }, { locale: ["en"] }];
}

export type { SiteContent };
