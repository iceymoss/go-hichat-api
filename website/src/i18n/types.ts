/**
 * 站点文案的唯一 shape 来源。
 *
 * zh.ts 与 en.ts 都用 `satisfies SiteContent`，这样漏翻译某个键时
 * TypeScript 会在构建期报错 —— 不引第三方 i18n 库也能防止两份字典漂移。
 */

export interface NavItem {
  label: string;
  href: string;
}

export interface Highlight {
  icon: string;
  title: string;
  description: string;
  points: readonly string[];
}

export interface Shot {
  src: string;
  title: string;
  caption: string;
}

export interface GalleryTab {
  id: string;
  label: string;
  icon: string;
  shots: readonly Shot[];
}

export interface ArchitectureLayer {
  id: string;
  title: string;
  blurb: string;
  nodes: readonly string[];
  edge: string | null;
}

export interface ServiceRow {
  name: string;
  layers: string;
  responsibility: string;
}

export interface TechGroup {
  group: string;
  items: readonly string[];
}

export interface FeatureSection {
  id: string;
  eyebrow: string;
  title: string;
  blurb: string;
  shot: string;
  shotAlt: string;
  capabilities: readonly string[];
  service: string;
}

export interface QuickStartStep {
  n: number;
  title: string;
  body: string;
  code: string;
  lang: string;
}

export interface LocalDevStep {
  title: string;
  body: string;
  code: string;
  lang: string;
}

export interface PortRow {
  service: string;
  port: string;
  note: string;
}

export interface DocsNavGroup {
  title: string;
  items: readonly NavItem[];
}

export interface FooterItem {
  label: string;
  href: string;
  external: boolean;
}

export interface FooterGroup {
  title: string;
  items: readonly FooterItem[];
}

/** 组件内的零散 UI 字符串（按钮、aria-label、分区小标题）。 */
export interface UIStrings {
  openMenu: string;
  closeMenu: string;
  githubRepo: string;
  github: string;
  quickStart: string;
  changelog: string;
  apiRef: string;
  docs: string;
  starOnGithub: string;
  fullGuide: string;
  dockerDeployDocs: string;
  activeDevelopment: string;
  heroShotAlt: string;
  heroTitleLine1: string;
  heroTitleLine2: string;
  heroSubtitle: string;
  featuresEyebrow: string;
  featuresTitle: string;
  featuresSubtitle: string;
  galleryEyebrow: string;
  galleryTitle: string;
  gallerySubtitle: string;
  archEyebrow: string;
  archTitle: string;
  archSubtitle: string;
  serviceInventory: string;
  serviceInventoryNote: string;
  colService: string;
  colLayers: string;
  techStackTitle: string;
  ctaEyebrow: string;
  ctaTitle: string;
  ctaSubtitle: string;
  ctaAfterStartup: string;
  ctaDemoLogin: string;
  ctaAutoFill: string;
  footerBlurb: string;
  footerLicense: string;
  switchLanguage: string;
}

export interface SiteContent {
  site: {
    name: string;
    tagline: string;
    description: string;
    version: string;
  };
  ui: UIStrings;
  navItems: readonly NavItem[];
  highlights: readonly Highlight[];
  galleryTabs: readonly GalleryTab[];
  architectureLayers: readonly ArchitectureLayer[];
  services: readonly ServiceRow[];
  techStack: readonly TechGroup[];
  featureSections: readonly FeatureSection[];
  quickStartSteps: readonly QuickStartStep[];
  localDevSteps: readonly LocalDevStep[];
  servicePorts: readonly PortRow[];
  docsNav: readonly DocsNavGroup[];
  footerGroups: readonly FooterGroup[];
}
