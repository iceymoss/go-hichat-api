import type { Metadata } from "next";
import { ArrowUpRight, Box, CalendarDays, Download, Tag } from "lucide-react";
import { localeParams, resolveLocale, type Locale } from "@/i18n";
import { links } from "@/i18n/shared";

type Localized = Record<Locale, string>;
type ReleaseSection = {
  kind: "new" | "improved" | "upgrade";
  title: Localized;
  items: readonly Localized[];
};
type Release = {
  version: string;
  date: string;
  name: Localized;
  summary: Localized;
  sections: readonly ReleaseSection[];
};

const copy = {
  zh: {
    metadataTitle: "发版日志 - HiChat",
    metadataDescription: "查看 HiChat 各正式版本的功能、改进与升级说明。",
    eyebrow: "Release notes",
    title: "HiChat 发版日志",
    description: "按正式发布版本记录新增能力、重要改进与升级注意事项。页面内容与 Git tag 一一对应。",
    versions: "个正式版本",
    latest: "最新版本",
    stable: "正式版",
    releasePage: "查看发布页",
    sourceCode: "下载源码",
    allReleases: "查看全部 GitHub Releases",
    publishedOn: "发布于",
    versionLabel: "版本",
  },
  en: {
    metadataTitle: "Release notes - HiChat",
    metadataDescription: "Explore features, improvements, and upgrade notes for every official HiChat release.",
    eyebrow: "Release notes",
    title: "HiChat release notes",
    description: "New capabilities, important improvements, and upgrade notes organized by official release. Every entry maps directly to a Git tag.",
    versions: "official releases",
    latest: "Latest",
    stable: "Stable",
    releasePage: "View release",
    sourceCode: "Download source",
    allReleases: "View all GitHub Releases",
    publishedOn: "Released",
    versionLabel: "Version",
  },
} as const satisfies Record<Locale, Record<string, string>>;

const releases: readonly Release[] = [
  {
    version: "v0.3.0",
    date: "2026-06-21",
    name: { zh: "一键部署", en: "One-command deployment" },
    summary: {
      zh: "首次提供完整的 Docker Compose 部署方案，将微服务、中间件和 Web 客户端整合为可重复启动的完整技术栈。",
      en: "The first complete Docker Compose deployment, packaging microservices, middleware, and the web client into a reproducible full stack.",
    },
    sections: [
      {
        kind: "new",
        title: { zh: "新增功能", en: "What's new" },
        items: [
          { zh: "一条命令启动 user、social、im、trend、task、streaming 与 Web 客户端。", en: "Start user, social, im, trend, task, streaming, and the web client with one command." },
          { zh: "内置 MySQL、Redis、MongoDB、Etcd 与 Kafka，并按依赖顺序执行健康检查。", en: "Bundled MySQL, Redis, MongoDB, Etcd, and Kafka with dependency-aware health checks." },
          { zh: "新增仅中间件的 Compose 配置，支持在宿主机原生调试 Go 服务。", en: "Added a middleware-only Compose stack for native Go service development." },
          { zh: "补充中英文 Docker 部署文档和完整克隆、启动、清理流程。", en: "Added bilingual Docker deployment guides with clone, startup, and cleanup workflows." },
        ],
      },
      {
        kind: "improved",
        title: { zh: "改进与修复", en: "Improvements and fixes" },
        items: [
          { zh: "每个 Kafka 消费者使用独立消费组，避免不同业务主题互相争抢消息。", en: "Assigned a dedicated consumer group to every Kafka consumer to prevent cross-topic contention." },
          { zh: "统一上传卷挂载，并修复 trend 服务资源及数据库字段初始化。", en: "Unified upload volume mounts and fixed trend resources and database initialization." },
          { zh: "聊天记录游标兼容非 ObjectID 消息 ID，降低历史数据升级风险。", en: "Made chat-history cursors tolerate non-ObjectID message IDs for safer data upgrades." },
        ],
      },
      {
        kind: "upgrade",
        title: { zh: "升级说明", en: "Upgrade notes" },
        items: [
          { zh: "升级前停止旧服务，拉取 v0.3.0 后重新执行 docker compose up -d --build。", en: "Stop the previous stack, check out v0.3.0, then run docker compose up -d --build." },
          { zh: "首次使用 Compose 部署时，请检查本地 2470、8887-8891、10090 和 10093 端口。", en: "For the first Compose deployment, verify that ports 2470, 8887-8891, 10090, and 10093 are available." },
        ],
      },
    ],
  },
  {
    version: "v0.2.0",
    date: "2026-06-20",
    name: { zh: "实时音视频", en: "Realtime audio and video" },
    summary: {
      zh: "HiChat 获得完整 WebRTC 通话能力，包括单人音视频、全网格群组通话、通话记录和丰富的通话交互。",
      en: "HiChat gains complete WebRTC calling with one-on-one audio/video, full-mesh group calls, call history, and a polished call experience.",
    },
    sections: [
      {
        kind: "new",
        title: { zh: "新增功能", en: "What's new" },
        items: [
          { zh: "支持单人语音和视频通话，覆盖邀请、响铃、接听、拒绝和挂断。", en: "One-on-one voice and video calls with invite, ring, accept, reject, and hang-up flows." },
          { zh: "支持最多四人的全网格群组通话，并同步每位参与者的麦克风和摄像头状态。", en: "Full-mesh group calls for up to four participants with synchronized microphone and camera state." },
          { zh: "群聊内展示进行中的通话，可从聊天横幅和通话列表直接加入。", en: "Active calls appear inside group chat and can be joined from the banner or call list." },
          { zh: "新增通话记录、未接来电提示、铃声、悬浮窗口和最小化来电组件。", en: "Added call records, missed-call indicators, ringtones, floating windows, and minimized incoming-call widgets." },
        ],
      },
      {
        kind: "improved",
        title: { zh: "改进与修复", en: "Improvements and fixes" },
        items: [
          { zh: "通话信令通过 streaming WebSocket 投递，并使用系统身份完成服务间鉴权。", en: "Call signaling now travels over the streaming WebSocket with authenticated system identity." },
          { zh: "提供默认 STUN 配置、ICE 诊断和媒体初始化超时处理。", en: "Added default STUN configuration, ICE diagnostics, and media setup timeout handling." },
          { zh: "修复远端音频播放、通话记录方向、群成员名称与已接来电未读状态。", en: "Fixed remote audio, call-record direction, group member names, and unread state for answered calls." },
        ],
      },
      {
        kind: "upgrade",
        title: { zh: "升级说明", en: "Upgrade notes" },
        items: [
          { zh: "需要启动独立的 streaming 服务，并确保客户端可访问其 WebSocket 与 ICE 配置接口。", en: "The standalone streaming service must be running and reachable for WebSocket signaling and ICE configuration." },
          { zh: "生产网络建议配置 TURN 服务，以覆盖严格 NAT 和受限网络环境。", en: "Configure TURN in production to support strict NAT and restricted network environments." },
        ],
      },
    ],
  },
  {
    version: "v0.1.0",
    date: "2026-06-19",
    name: { zh: "首个公开版本", en: "First public release" },
    summary: {
      zh: "HiChat 2.0 的首个正式版本，交付完整的微服务 IM、社交关系、动态空间与 Next.js Web 客户端。",
      en: "The first official HiChat 2.0 release, delivering microservice IM, social relationships, an activity feed, and a Next.js web client.",
    },
    sections: [
      {
        kind: "new",
        title: { zh: "核心能力", en: "Core capabilities" },
        items: [
          { zh: "完整的用户注册、登录、资料、安全设置与中英文界面。", en: "User registration, login, profiles, security settings, and bilingual UI." },
          { zh: "好友申请、好友管理、群组创建、邀请、角色与群公告。", en: "Friend requests and management, plus group creation, invites, roles, and announcements." },
          { zh: "单聊和群聊、富媒体消息、引用回复、撤回、@提醒、已读回执与离线消息。", en: "Single and group chat, rich media, quoted replies, recall, mentions, read receipts, and offline messages." },
          { zh: "动态发布、图片视频、评论回复、点赞、可见性控制与消息通知。", en: "Activity publishing, image/video media, comments, replies, likes, visibility controls, and notifications." },
          { zh: "基于 WebSocket、Kafka、MySQL、MongoDB、Redis 和 Etcd 的完整微服务链路。", en: "A complete microservice pipeline built on WebSocket, Kafka, MySQL, MongoDB, Redis, and Etcd." },
        ],
      },
      {
        kind: "improved",
        title: { zh: "生产化整理", en: "Production readiness" },
        items: [
          { zh: "移除脚手架、沙盒和废弃 mock 数据，收敛为可维护的生产代码结构。", en: "Removed scaffolding, sandboxes, and obsolete mock data for a maintainable production codebase." },
          { zh: "修复字体缩放下的全屏布局、移动端个人页与聊天历史滚动位置。", en: "Fixed full-height layout under font scaling, mobile profile pages, and chat-history scroll position." },
          { zh: "完善品牌标识、主题颜色、用户头像和群聊信息交互。", en: "Polished branding, theme colors, user avatars, and group-chat interactions." },
        ],
      },
      {
        kind: "upgrade",
        title: { zh: "安装说明", en: "Installation notes" },
        items: [
          { zh: "该版本是首个正式 tag，无需从旧版本迁移；请按开发文档准备全部基础设施。", en: "This is the first official tag, so no previous-version migration is required; prepare all infrastructure from the development guide." },
        ],
      },
    ],
  },
];

const sectionStyle: Record<ReleaseSection["kind"], string> = {
  new: "border-brand/25 bg-brand/[0.06]",
  improved: "border-blue-400/20 bg-blue-500/[0.05]",
  upgrade: "border-amber-400/20 bg-amber-500/[0.05]",
};

type Props = { params: Promise<{ locale: string }> };

export function generateStaticParams() {
  return localeParams();
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  return {
    title: copy[locale].metadataTitle,
    description: copy[locale].metadataDescription,
  };
}

export default async function ChangelogPage({ params }: Props) {
  const locale = resolveLocale((await params).locale);
  const text = copy[locale];

  return (
    <div className="mx-auto max-w-5xl px-4 py-14 sm:px-6 sm:py-20 lg:px-8">
      <header className="relative overflow-hidden rounded-2xl border border-border bg-card px-6 py-10 sm:px-10">
        <div aria-hidden className="absolute -right-24 -top-24 size-72 rounded-full bg-brand/10 blur-3xl" />
        <div className="relative max-w-2xl">
          <p className="font-mono text-xs font-medium uppercase tracking-[0.2em] text-brand">{text.eyebrow}</p>
          <h1 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-5xl">{text.title}</h1>
          <p className="mt-5 text-sm leading-7 text-muted-foreground sm:text-base">{text.description}</p>
          <div className="mt-7 flex flex-wrap gap-3">
            <span className="inline-flex items-center gap-2 rounded-full border border-border bg-background/60 px-3 py-1.5 text-xs text-muted-foreground">
              <Box className="size-3.5 text-brand" /> {releases.length} {text.versions}
            </span>
            <span className="inline-flex items-center gap-2 rounded-full border border-brand/25 bg-brand/10 px-3 py-1.5 text-xs text-brand">
              <Tag className="size-3.5" /> {text.latest}: {releases[0].version}
            </span>
          </div>
        </div>
      </header>

      <div className="mt-14 space-y-16">
        {releases.map((release, releaseIndex) => {
          const tagUrl = `${links.github}/releases/tag/${release.version}`;
          return (
            <article key={release.version} id={release.version} className="scroll-mt-24">
              <div className="grid gap-6 border-b border-border pb-7 sm:grid-cols-[180px_1fr]">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-2xl font-bold tracking-tight text-brand">{release.version}</span>
                    <span className="rounded-full border border-brand/25 bg-brand/10 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-brand">
                      {releaseIndex === 0 ? text.latest : text.stable}
                    </span>
                  </div>
                  <p className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
                    <CalendarDays className="size-3.5" /> {text.publishedOn} {release.date}
                  </p>
                </div>
                <div>
                  <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{text.versionLabel}</p>
                  <h2 className="mt-1 text-2xl font-semibold tracking-tight text-foreground">{release.name[locale]}</h2>
                  <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-foreground">{release.summary[locale]}</p>
                  <div className="mt-5 flex flex-wrap gap-3">
                    <a href={tagUrl} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1.5 rounded-lg border border-brand/30 bg-brand/10 px-3 py-2 text-xs font-medium text-brand transition-colors hover:bg-brand/15">
                      {text.releasePage}<ArrowUpRight className="size-3.5" />
                    </a>
                    <a href={`${links.github}/archive/refs/tags/${release.version}.zip`} className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-secondary px-3 py-2 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground">
                      <Download className="size-3.5" />{text.sourceCode}
                    </a>
                  </div>
                </div>
              </div>

              <div className="mt-7 grid gap-4 lg:grid-cols-3">
                {release.sections.map((section) => (
                  <section key={section.kind} className={`rounded-xl border p-5 ${sectionStyle[section.kind]}`}>
                    <h3 className="text-sm font-semibold text-foreground">{section.title[locale]}</h3>
                    <ul className="mt-4 space-y-3">
                      {section.items.map((item, index) => (
                        <li key={index} className="flex gap-2.5 text-sm leading-6 text-muted-foreground">
                          <span aria-hidden className="mt-2 size-1.5 shrink-0 rounded-full bg-brand" />
                          <span>{item[locale]}</span>
                        </li>
                      ))}
                    </ul>
                  </section>
                ))}
              </div>
            </article>
          );
        })}
      </div>

      <a href={links.releases} target="_blank" rel="noopener noreferrer" className="mt-14 flex items-center justify-between rounded-xl border border-border bg-card p-5 text-sm text-muted-foreground transition-colors hover:border-brand/40 hover:text-foreground">
        <span>{text.allReleases}</span><ArrowUpRight className="size-4 text-brand" />
      </a>
    </div>
  );
}
