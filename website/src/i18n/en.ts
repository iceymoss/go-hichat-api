/**
 * English copy dictionary.
 *
 * Together with zh.ts this `satisfies SiteContent`, so a missing key fails the
 * build. Language-neutral URLs / sizes / demo account live in shared.ts.
 */
import type { SiteContent } from "./types";
import { links } from "./shared";

const site = {
  name: "HiChat",
  tagline: "Open-Source IM & Social Platform",
  description:
    "A go-zero based microservice IM, social graph, and activity feed — with a WebSocket gateway, WebRTC calls, and a Kafka event pipeline.",
  version: "2.0",
} as const;

const productPages = {
  features: {
    metadataTitle: "Features — HiChat",
    metadataDescription:
      "Instant messaging, social graph, activity feed, WebRTC calls, realtime gateway, and async workers — every subsystem in HiChat 2.0.",
    eyebrow: "Features",
    title: "Six subsystems, one repo",
    description:
      "Each domain owns its own contracts, storage, and async paths. Here is what ships in every layer.",
    source: "Source",
    ctaTitle: "Run it yourself",
    ctaDescription:
      "One Docker Compose command brings the whole stack up, demo data included.",
    quickStart: "Quick Start",
    apiReference: "API Reference",
  },
  quickStart: {
    metadataTitle: "Quick Start — HiChat",
    metadataDescription:
      "Bring up the full HiChat stack with one Docker Compose command, or run the Go services natively for development.",
    eyebrow: "Quick Start",
    title: "Up and running in a minute",
    description:
      "Docker Compose is the fastest path. Prefer running the Go services natively? That path is below too.",
    prerequisites: "Prerequisites",
    dockerRequirement: "Docker with the Compose plugin",
    memoryRequirement: "Roughly 4 GB of free memory for the full stack",
    portsRequirement: "Ports 2470, 8887–8891, 10090, and 10093 free",
    dockerCompose: "Docker Compose",
    demoAccount: "Demo account",
    demoAccountDescription: "Available after seeding the demo dataset in step 3.",
    phone: "Phone",
    password: "Password",
    url: "URL",
    demoModeNote:
      "Verification codes auto-fill in demo mode, so you can also register a fresh account without an SMS provider.",
    exposedPorts: "Exposed ports",
    service: "Service",
    port: "Port",
    purpose: "Purpose",
    localDevelopment: "Local development",
    localDevelopmentDescription:
      "For contributors iterating on the Go services, run the middleware in Docker and the services on the host.",
    secretWarningTitle: "Set a dedicated RPC auth secret.",
    secretWarningDescription:
      "must be at least 32 bytes of random data and must not reuse the JWT secret.",
    nextSteps: "Next steps",
    nextStepLinks: [
      { label: "API Reference", href: links.docsApi, description: "Every REST and gRPC contract" },
      { label: "Developer Guide", href: links.docsDevGuide, description: "Project layout and conventions" },
      { label: "Docker Deploy", href: links.dockerDeploy, description: "Reverse proxy, HTTPS, TURN" },
      { label: "Contributing", href: links.contributing, description: "How to open your first PR" },
    ],
    helpTitle: "Something not working?",
    helpDescription: "Open an issue with your Compose logs and we'll take a look.",
    openIssue: "Open an issue",
  },
} as const;

const ui = {
  openMenu: "Open menu",
  closeMenu: "Close menu",
  githubRepo: "GitHub repository",
  github: "GitHub",
  quickStart: "Quick Start",
  changelog: "Changelog",
  apiRef: "API Ref",
  docs: "Docs",
  starOnGithub: "Star on GitHub",
  fullGuide: "Full Guide",
  dockerDeployDocs: "Docker Deploy Docs",
  activeDevelopment: "Active Development",
  heroShotAlt: "HiChat one-on-one chat interface",
  heroTitleLine1: "Open-Source IM &",
  heroTitleLine2: "Social Platform",
  heroSubtitle:
    "A go-zero microservice backend with a WebSocket gateway, WebRTC calls, a Kafka event pipeline, and a full-featured Next.js web client — all in one open-source repo.",
  featuresEyebrow: "Everything included",
  featuresTitle: "A complete IM stack, not a demo",
  featuresSubtitle:
    "Six production-shaped subsystems, each with real contracts, storage, and async paths — ready to read, run, and extend.",
  galleryEyebrow: "See it running",
  galleryTitle: "Real screens, not mockups",
  gallerySubtitle:
    "Every screenshot below comes from the running web client with the demo dataset seeded.",
  archEyebrow: "Under the hood",
  archTitle: "Five layers, clear boundaries",
  archSubtitle:
    "Requests flow down through access, domain, and event layers. Services never read each other's tables — everything crosses via zRPC or Kafka.",
  serviceInventory: "Service inventory",
  serviceInventoryNote: "Every service under apps/.",
  colService: "Service",
  colLayers: "Layers",
  techStackTitle: "Tech stack",
  ctaEyebrow: "One command to start",
  ctaTitle: "Running in under a minute",
  ctaSubtitle:
    "Docker Compose brings up every dependency — MySQL, Redis, MongoDB, Etcd, Kafka, all six microservices, and the web client.",
  ctaAfterStartup: "After startup, open",
  ctaDemoLogin: "Demo login:",
  ctaAutoFill: "Verification codes auto-fill in demo mode",
  footerBlurb:
    "Open-source IM & social platform built on Go microservices, go-zero, Next.js, Kafka, and WebRTC.",
  footerLicense: "Released under the Apache 2.0 License.",
  switchLanguage: "Switch language",
} as const;

const navItems = [
  { label: "Home", href: "/" },
  { label: "Features", href: "/features" },
  { label: "Quick Start", href: "/quick-start" },
] as const;

const highlights = [
  {
    icon: "MessageSquare",
    title: "Instant Messaging",
    description:
      "One-on-one and group conversations with text, image, voice, and video messages. Quotes, mentions, recall, read receipts, and unread state.",
    points: ["MongoDB chat history", "Pin & mute", "Message recall"],
  },
  {
    icon: "Users",
    title: "Social Graph",
    description:
      "Friend requests, remarks, blocking, and tags. Group creation, invite tokens, announcements, roles, and ownership transfer.",
    points: ["Friend requests", "Group admin ops", "Invite links"],
  },
  {
    icon: "Images",
    title: "Activity Feed",
    description:
      "Publish moments with media and visibility control. Comments, threaded replies, likes, drafts, and an unread notification inbox.",
    points: ["Visibility control", "Comments & likes", "Draft support"],
  },
  {
    icon: "Video",
    title: "Realtime Calls",
    description:
      "WebRTC one-on-one and group calls via full mesh, plus meetings, screen sharing, and live streaming through an independent SFU service.",
    points: ["1:1 & group calls", "Screen sharing", "Pion SFU"],
  },
  {
    icon: "Workflow",
    title: "Kafka Event Pipeline",
    description:
      "Chat delivery, read events, recall, and feed notifications all flow through Kafka topics, consumed by idempotent background workers.",
    points: ["Idempotent consumers", "Dead-letter handling", "Cron jobs"],
  },
  {
    icon: "Boxes",
    title: "Microservice Architecture",
    description:
      "Four domains split into API and zRPC layers, registered through etcd. Contracts defined in .api and .proto, generated with goctl.",
    points: ["go-zero + zRPC", "etcd discovery", "Contract-first"],
  },
] as const;

const galleryTabs = [
  {
    id: "messaging",
    label: "Messaging",
    icon: "MessageSquare",
    shots: [
      {
        src: "/screenshots/single-chat.webp",
        title: "One-on-one chat",
        caption: "Text, image, voice, and video messages with quotes and recall.",
      },
      {
        src: "/screenshots/group-chat.webp",
        title: "Group chat",
        caption: "Mentions, member roles, and announcements in group threads.",
      },
      {
        src: "/screenshots/conversation-list.webp",
        title: "Conversation list",
        caption: "Unread counts, pinning, and mute state across conversations.",
      },
      {
        src: "/screenshots/create-group.webp",
        title: "Create a group",
        caption: "Pick members, set a name, and generate an invite link.",
      },
    ],
  },
  {
    id: "social",
    label: "Friends",
    icon: "Users",
    shots: [
      {
        src: "/screenshots/friend-list-and-settings.webp",
        title: "Friend list & settings",
        caption: "Remarks, tags, blocking, and per-friend moment permissions.",
      },
      {
        src: "/screenshots/friend-requests-received.webp",
        title: "Friend requests",
        caption: "Incoming and outgoing requests with verification messages.",
      },
      {
        src: "/screenshots/profile-home.webp",
        title: "Profile",
        caption: "Avatar, profile fields, and personal moment history.",
      },
    ],
  },
  {
    id: "moments",
    label: "Moments",
    icon: "Images",
    shots: [
      {
        src: "/screenshots/moments-feed-and-detail.webp",
        title: "Feed & detail",
        caption: "Threaded comments, replies, and likes on each moment.",
      },
      {
        src: "/screenshots/publish-moment.webp",
        title: "Publish a moment",
        caption: "Attach media and choose who is allowed to see the post.",
      },
      {
        src: "/screenshots/moments.webp",
        title: "Moments space",
        caption: "A full activity feed backed by the trend service.",
      },
    ],
  },
  {
    id: "calls",
    label: "Calls",
    icon: "Video",
    shots: [
      {
        src: "/screenshots/call-incoming.webp",
        title: "Incoming call",
        caption: "WebRTC signalling with ringing, accept, and reject states.",
      },
      {
        src: "/screenshots/group-call-active.webp",
        title: "Group call",
        caption: "Full-mesh group calls for up to four participants.",
      },
    ],
  },
] as const;

const architectureLayers = [
  {
    id: "L0",
    title: "Client Layer",
    blurb: "Web client plus mobile and third-party consumers.",
    nodes: ["web/ (Next.js + React)", "Mobile / third-party clients"],
    edge: "REST · WebSocket · WebRTC",
  },
  {
    id: "L1",
    title: "Access Layer",
    blurb: "HTTP entry points, the realtime gateway, and media signalling.",
    nodes: [
      "user/api · social/api · im/api · trend/api",
      "im/ws — auth, heartbeat, ack, online push",
      "streaming — signalling, rooms, SFU",
    ],
    edge: "zRPC · publish/consume · Redis",
  },
  {
    id: "L2",
    title: "Domain Service Layer",
    blurb: "One zRPC service per domain, discovered through etcd.",
    nodes: [
      "user/rpc — auth, profile, verification",
      "social/rpc — friends, groups, requests",
      "im/rpc — conversations, chat logs, read/recall",
      "trend/rpc — feed, comments, likes, notify",
    ],
    edge: "MySQL · MongoDB · Kafka",
  },
  {
    id: "L3",
    title: "Event & Async Layer",
    blurb: "Kafka topics drained by idempotent workers and cron jobs.",
    nodes: [
      "chat-transfer · read-transfer",
      "recall-transfer · trend-notify",
      "task/mq — persist chat, update read state, push events",
      "task/cron — scheduled stats, cleanup, extensions",
    ],
    edge: "persist · update · push",
  },
  {
    id: "L4",
    title: "Data & Runtime Infrastructure",
    blurb: "Each service owns its own tables; cross-service DB reads are banned.",
    nodes: [
      "MySQL — users, friends, groups, trends, comments",
      "MongoDB — chat logs, read records, recall state",
      "Redis — sessions, online presence, cache, room state",
      "Etcd — service registration and discovery",
    ],
    edge: null,
  },
] as const;

const services = [
  {
    name: "user",
    layers: "api · rpc",
    responsibility: "Accounts, auth, profile, verification codes, user lookup",
  },
  {
    name: "social",
    layers: "api · rpc",
    responsibility: "Friends, groups, requests, invites, announcements",
  },
  {
    name: "im",
    layers: "api · rpc · ws",
    responsibility:
      "Conversations, chat logs, read receipts, recall, WebSocket gateway",
  },
  {
    name: "trend",
    layers: "api · rpc",
    responsibility: "Activity feed, comments, likes, drafts, media, notifications",
  },
  {
    name: "task",
    layers: "mq · cron",
    responsibility: "Kafka consumers and scheduled jobs",
  },
  {
    name: "streaming",
    layers: "sfu · webrtc",
    responsibility: "WebRTC calls, rooms, meetings, screen sharing, live",
  },
] as const;

const techStack = [
  { group: "Backend", items: ["Go 1.25", "go-zero", "zRPC / gRPC", "goctl"] },
  { group: "Realtime", items: ["WebSocket", "Kafka", "WebRTC", "Pion"] },
  { group: "Storage", items: ["MySQL 8", "MongoDB 7", "Redis 7", "Etcd v3.5"] },
  { group: "Frontend", items: ["Next.js 16", "React 19", "Bun", "Tailwind CSS"] },
] as const;

const featureSections = [
  {
    id: "messaging",
    eyebrow: "Instant Messaging",
    title: "Conversations that behave like a real IM",
    blurb:
      "Single and group conversations backed by MongoDB chat logs, with the delivery, read, and recall paths all running through Kafka.",
    shot: "/screenshots/single-chat.webp",
    shotAlt: "HiChat one-on-one chat view",
    capabilities: [
      "Text, image, file, voice, and video messages",
      "Quotes, @mentions, and message recall",
      "Read receipts and per-conversation unread state",
      "Pin and mute conversations, batch actions",
      "Offline messages pulled by conversationId + seq",
      "Media upload shared across api services",
    ],
    service: "apps/im",
  },
  {
    id: "social",
    eyebrow: "Social Graph",
    title: "Friends, groups, and everything in between",
    blurb:
      "The social service owns the relationship graph — requests, blocking, tags, and the full group lifecycle from creation to ownership transfer.",
    shot: "/screenshots/friend-list-and-settings.webp",
    shotAlt: "HiChat friend list and settings",
    capabilities: [
      "Friend requests with verification messages",
      "Remarks, tags, blocking, and reports",
      "Group creation, search, and join requests",
      "Invitations and shareable invite tokens",
      "Member management, roles, and admin operations",
      "Announcements and ownership transfer",
    ],
    service: "apps/social",
  },
  {
    id: "moments",
    eyebrow: "Activity Feed",
    title: "A moments space with real visibility rules",
    blurb:
      "Publish posts with media and per-friend visibility. Comments, replies, and likes fan out through Kafka into an unread notification inbox.",
    shot: "/screenshots/moments-feed-and-detail.webp",
    shotAlt: "HiChat moments feed and detail view",
    capabilities: [
      "Publish with images and video resources",
      "Visibility control per post",
      "Comments, threaded replies, and likes",
      "Draft support before publishing",
      "Unread counters and notification inbox",
      "Online push via the WebSocket gateway",
    ],
    service: "apps/trend",
  },
  {
    id: "calls",
    eyebrow: "Realtime Calls",
    title: "WebRTC calls from an independent SFU",
    blurb:
      "The streaming service runs separately from the main stack, handling signalling, rooms, and media for calls, meetings, and live sessions.",
    shot: "/screenshots/group-call-active.webp",
    shotAlt: "HiChat active group call",
    capabilities: [
      "One-on-one calls with ring, accept, reject",
      "Group calls via full mesh, up to four peers",
      "Meetings and screen sharing",
      "Live streaming through the SFU",
      "Room state kept in Redis",
      "Built on Pion WebRTC",
    ],
    service: "apps/streaming",
  },
  {
    id: "realtime",
    eyebrow: "Realtime Gateway",
    title: "One WebSocket gateway, many nodes",
    blurb:
      "im/ws owns connection lifecycle: auth, heartbeat, ack tracking, and retry. Presence lives in Redis, and cross-node delivery routes through Kafka.",
    shot: "/screenshots/conversation-list.webp",
    shotAlt: "HiChat conversation list",
    capabilities: [
      "Connection auth and per-user sessions",
      "Heartbeat with ghost-connection cleanup",
      "ACK tracking, retry, and dedup",
      "Redis-backed online presence per node",
      "Cross-node delivery over Kafka topics",
      "Batched broadcast for large fan-out",
    ],
    service: "apps/im/ws",
  },
  {
    id: "async",
    eyebrow: "Async & Tasks",
    title: "Idempotent workers behind every event",
    blurb:
      "Chat persistence, read state, recall, and feed notifications are all consumer-side concerns — retried safely and never double-applied.",
    shot: "/screenshots/moments.webp",
    shotAlt: "HiChat moments space",
    capabilities: [
      "chat-transfer — persist chat to MongoDB",
      "read-transfer — update read records",
      "recall-transfer — apply recall state",
      "trend-notify — push feed notifications",
      "Idempotent consumers with dedup keys",
      "Cron jobs with Redis single-instance locks",
    ],
    service: "apps/task",
  },
] as const;

const quickStartSteps = [
  {
    n: 1,
    title: "Clone the repository",
    body: "Grab the source. Everything below runs from the repo root.",
    code: `git clone https://github.com/iceymoss/go-hichat-api.git
cd go-hichat-api`,
    lang: "bash",
  },
  {
    n: 2,
    title: "Bring up the stack",
    body: "Compose starts MySQL, Redis, MongoDB, Etcd, Kafka, all six services, and the web client — in dependency order with health checks.",
    code: `docker compose up -d --build`,
    lang: "bash",
  },
  {
    n: 3,
    title: "Seed the demo dataset",
    body: "Optional. Registers 14 demo users with pre-filled conversations, friends, and moments — the exact dataset in the screenshots.",
    code: `docker compose --profile mock up mockdata`,
    lang: "bash",
  },
  {
    n: 4,
    title: "Open the client",
    body: "The web client is served on port 2470. Verification codes auto-fill in demo mode, so you can register freely.",
    code: `open http://localhost:2470`,
    lang: "bash",
  },
] as const;

const servicePorts = [
  { service: "web", port: "2470", note: "Next.js web client" },
  { service: "user-api", port: "8887", note: "Accounts, auth, profile" },
  { service: "social-api", port: "8889", note: "Friends and groups" },
  { service: "im-api", port: "8890", note: "Conversations and chat logs" },
  { service: "trend-api", port: "8891", note: "Activity feed" },
  { service: "im-ws", port: "10090", note: "WebSocket gateway" },
  {
    service: "streaming",
    port: "10093",
    note: "WebRTC SFU (plus 50000-50200/udp)",
  },
] as const;

const localDevSteps = [
  {
    title: "Start middleware only",
    body: "Runs MySQL, Redis, MongoDB, Etcd, and Kafka with host-accessible ports, leaving the Go services to run natively.",
    code: `docker compose -f docker-compose.dependencies.yaml up -d`,
    lang: "bash",
  },
  {
    title: "Run every Go service",
    body: "The helper script starts all services in dependency order. The RPC auth secret must be at least 32 bytes and must not reuse the JWT secret.",
    code: `HICHAT_IM_RPC_AUTH_SECRET=<random-32-byte-value> ./hichat2.sh`,
    lang: "bash",
  },
  {
    title: "Or run a single service",
    body: "Useful when iterating on one layer. Each service takes its own sample config.",
    code: `go run apps/user/api/user.go -f apps/user/api/etc/user-sample.yaml`,
    lang: "bash",
  },
  {
    title: "Start the web client",
    body: "The dev server listens on port 3001.",
    code: `cd web && bun install && bun dev`,
    lang: "bash",
  },
] as const;

const docsNav = [
  {
    title: "Overview",
    items: [
      { label: "Introduction", href: "/docs" },
      { label: "Architecture", href: "/docs/architecture" },
      { label: "Domain Map", href: "/docs/domains" },
    ],
  },
  {
    title: "Core Concepts",
    items: [
      { label: "Core Data Flows", href: "/docs/data-flows" },
      { label: "Message Lifecycle", href: "/docs/message-lifecycle" },
      { label: "Realtime Gateway", href: "/docs/realtime-gateway" },
    ],
  },
  {
    title: "Guides",
    items: [{ label: "Extending the System", href: "/docs/extending" }],
  },
  {
    title: "Project",
    items: [{ label: "Changelog", href: "/changelog" }],
  },
] as const;

const footerGroups = [
  {
    title: "Product",
    items: [
      { label: "Features", href: "/features", external: false },
      { label: "Quick Start", href: "/quick-start", external: false },
    ],
  },
  {
    title: "Documentation",
    items: [
      { label: "Design Docs", href: "/docs", external: false },
      { label: "API Reference", href: links.docsApi, external: true },
      { label: "Developer Guide", href: links.docsDevGuide, external: true },
      { label: "Docker Deploy", href: links.dockerDeploy, external: true },
    ],
  },
  {
    title: "Community",
    items: [
      { label: "GitHub", href: links.github, external: true },
      { label: "Issues", href: links.githubIssues, external: true },
      { label: "Contributing", href: links.contributing, external: true },
    ],
  },
] as const;

export const en = {
  site,
  ui,
  productPages,
  navItems,
  highlights,
  galleryTabs,
  architectureLayers,
  services,
  techStack,
  featureSections,
  quickStartSteps,
  localDevSteps,
  servicePorts,
  docsNav,
  footerGroups,
} satisfies SiteContent;
