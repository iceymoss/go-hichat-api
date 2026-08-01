import { zh } from "./zh";
import { en } from "./en";
import type { SiteContent } from "./types";

export const locales = ["zh", "en"] as const;
export type Locale = (typeof locales)[number];

/** 官网使用显式语言段：/zh 与 /en。 */
export const defaultLocale: Locale = "zh";

const dictionaries: Record<Locale, SiteContent> = { zh, en };

/**
 * 把动态路由段解析为 Locale，未知值回退中文。
 */
export function resolveLocale(segment?: string): Locale {
  return segment === "en" ? "en" : "zh";
}

/** 同步字典查找 —— 服务端组件直接调用，不破坏静态预渲染。 */
export function getContent(locale: Locale): SiteContent {
  return dictionaries[locale];
}

/**
 * 生成带语言前缀的站内链接。
 * 统一走这里，避免在每个文件里手写前缀拼接。
 */
export function localeHref(locale: Locale, path: string): string {
  return path === "/" ? `/${locale}` : `/${locale}${path}`;
}

/** 每个页面的 generateStaticParams 都返回这个：产出根路径与 /en 两套。 */
export function localeParams() {
  return locales.map((locale) => ({ locale }));
}

export type { SiteContent };
