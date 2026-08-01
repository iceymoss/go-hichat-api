/**
 * 中文文案字典。
 *
 * 与 en.ts 一同 `satisfies SiteContent`，漏翻译时 TypeScript 会报错。
 * 语言中立的 URL / 尺寸 / 演示账号在 shared.ts。
 */
import type { SiteContent } from "./types";
import { links } from "./shared";

const site = {
  name: "HiChat",
  tagline: "开源即时通讯与社交平台",
  description:
    "基于 go-zero 的微服务 IM、社交关系与动态空间 —— 内置 WebSocket 网关、WebRTC 音视频与 Kafka 事件管道。",
  version: "2.0",
} as const;

const productPages = {
  features: {
    metadataTitle: "功能 — HiChat",
    metadataDescription:
      "全面了解 HiChat 2.0 的即时消息、社交关系、动态空间、WebRTC 音视频、实时网关与异步任务子系统。",
    eyebrow: "功能",
    title: "六个子系统，一个仓库",
    description:
      "每个领域都拥有自己的契约、存储与异步链路。以下是各层已经具备的完整能力。",
    source: "源码",
    ctaTitle: "亲自运行",
    ctaDescription:
      "一条 Docker Compose 命令即可启动完整技术栈，并包含演示数据。",
    quickStart: "快速开始",
    apiReference: "API 参考",
  },
  quickStart: {
    metadataTitle: "快速开始 — HiChat",
    metadataDescription:
      "使用一条 Docker Compose 命令启动完整 HiChat 技术栈，或在本地原生运行 Go 服务进行开发。",
    eyebrow: "快速开始",
    title: "一分钟内运行起来",
    description:
      "Docker Compose 是最快的启动方式。如果更喜欢在本地原生运行 Go 服务，下方也提供了对应步骤。",
    prerequisites: "准备条件",
    dockerRequirement: "安装带 Compose 插件的 Docker",
    memoryRequirement: "为完整技术栈预留约 4 GB 可用内存",
    portsRequirement: "确保 2470、8887–8891、10090 和 10093 端口未被占用",
    dockerCompose: "Docker Compose",
    demoAccount: "演示账号",
    demoAccountDescription: "完成第 3 步演示数据灌入后即可使用。",
    phone: "手机号",
    password: "密码",
    url: "地址",
    demoModeNote:
      "演示模式下验证码会自动填充，因此无需短信服务商也可以注册新账号。",
    exposedPorts: "暴露端口",
    service: "服务",
    port: "端口",
    purpose: "用途",
    localDevelopment: "本地开发",
    localDevelopmentDescription:
      "贡献者开发 Go 服务时，可以在 Docker 中运行中间件，在宿主机上运行服务。",
    secretWarningTitle: "请设置独立的 RPC 鉴权密钥。",
    secretWarningDescription:
      "必须是至少 32 字节的随机数据，且不得复用 JWT secret。",
    nextSteps: "后续步骤",
    nextStepLinks: [
      { label: "API 参考", href: links.docsApi, description: "全部 REST 与 gRPC 契约" },
      { label: "开发指南", href: links.docsDevGuideZh, description: "项目结构与开发约定" },
      { label: "Docker 部署", href: links.dockerDeploy, description: "反向代理、HTTPS 与 TURN" },
      { label: "参与贡献", href: links.contributing, description: "如何提交第一个 PR" },
    ],
    helpTitle: "遇到问题？",
    helpDescription: "请附上 Compose 日志创建 issue，我们会协助排查。",
    openIssue: "创建 issue",
  },
} as const;

/** 导航栏主菜单 */
const navItems = [
  { label: "首页", href: "/" },
  { label: "功能", href: "/features" },
  { label: "快速开始", href: "/quick-start" },
] as const;

/**
 * 首页功能亮点卡片。`icon` 为 lucide-react 的导出名，在 FeatureGrid 中解析，
 * 使本模块保持不依赖组件导入。
 */
const highlights = [
  {
    icon: "MessageSquare",
    title: "即时消息",
    description:
      "单聊与群聊支持文本、图片、语音、视频消息。引用、@提醒、撤回、已读回执与未读状态齐备。",
    points: ["MongoDB 聊天记录", "置顶与免打扰", "消息撤回"],
  },
  {
    icon: "Users",
    title: "社交关系",
    description:
      "好友申请、备注、拉黑与标签。群组创建、邀请令牌、公告、角色与群主转让。",
    points: ["好友申请", "群管理操作", "邀请链接"],
  },
  {
    icon: "Images",
    title: "动态空间",
    description:
      "发布带媒体资源的动态并控制可见范围。评论、嵌套回复、点赞、草稿与未读通知收件箱。",
    points: ["可见性控制", "评论与点赞", "草稿支持"],
  },
  {
    icon: "Video",
    title: "实时音视频",
    description:
      "WebRTC 单人与群组通话（全网格），以及会议、屏幕共享与直播 —— 由独立的 SFU 服务承载。",
    points: ["单聊与群组通话", "屏幕共享", "Pion SFU"],
  },
  {
    icon: "Workflow",
    title: "Kafka 事件管道",
    description:
      "消息投递、已读事件、撤回与动态通知全部流经 Kafka topic，由幂等的后台消费者处理。",
    points: ["幂等消费", "死信处理", "定时任务"],
  },
  {
    icon: "Boxes",
    title: "微服务架构",
    description:
      "四个领域拆分为 API 与 zRPC 两层，通过 etcd 注册发现。契约定义在 .api 与 .proto，由 goctl 生成。",
    points: ["go-zero + zRPC", "etcd 服务发现", "契约优先"],
  },
] as const;

/**
 * 截图画廊，按分类分组。
 *
 * 原图为 docs/screenshots 下约 3024x1718 的 PNG；`bun run optimize:screenshots`
 * 将其压缩为 public/screenshots 下 1600px 宽的 WebP。SCREENSHOT_W/H 与该输出一致，
 * 传给 next/image 以便浏览器在图片到达前预留正确的盒子尺寸。
 */

const galleryTabs = [
  {
    id: "messaging",
    label: "消息",
    icon: "MessageSquare",
    shots: [
      {
        src: "/screenshots/single-chat.webp",
        title: "单聊",
        caption: "文本、图片、语音、视频消息，支持引用与撤回。",
      },
      {
        src: "/screenshots/group-chat.webp",
        title: "群聊",
        caption: "群会话中的 @提醒、成员角色与群公告。",
      },
      {
        src: "/screenshots/conversation-list.webp",
        title: "会话列表",
        caption: "跨会话的未读计数、置顶与免打扰状态。",
      },
      {
        src: "/screenshots/create-group.webp",
        title: "创建群组",
        caption: "选择成员、设置群名并生成邀请链接。",
      },
    ],
  },
  {
    id: "social",
    label: "好友",
    icon: "Users",
    shots: [
      {
        src: "/screenshots/friend-list-and-settings.webp",
        title: "好友列表与设置",
        caption: "备注、标签、拉黑与逐好友的动态可见权限。",
      },
      {
        src: "/screenshots/friend-requests-received.webp",
        title: "好友申请",
        caption: "收到与发出的申请，附验证消息。",
      },
      {
        src: "/screenshots/profile-home.webp",
        title: "个人主页",
        caption: "头像、资料字段与个人动态历史。",
      },
    ],
  },
  {
    id: "moments",
    label: "动态",
    icon: "Images",
    shots: [
      {
        src: "/screenshots/moments-feed-and-detail.webp",
        title: "动态流与详情",
        caption: "每条动态下的嵌套评论、回复与点赞。",
      },
      {
        src: "/screenshots/publish-moment.webp",
        title: "发布动态",
        caption: "附加媒体资源并选择可见范围。",
      },
      {
        src: "/screenshots/moments.webp",
        title: "动态空间",
        caption: "由 trend 服务支撑的完整动态流。",
      },
    ],
  },
  {
    id: "calls",
    label: "音视频",
    icon: "Video",
    shots: [
      {
        src: "/screenshots/call-incoming.webp",
        title: "来电界面",
        caption: "WebRTC 信令的响铃、接听与拒接状态。",
      },
      {
        src: "/screenshots/group-call-active.webp",
        title: "群组通话",
        caption: "全网格模式下最多四人的群组通话。",
      },
    ],
  },
] as const;

/**
 * 五个架构层级，对应 README 中的架构图。
 * 采用堆叠卡片而非原始 ASCII 渲染 —— ASCII 图需约 80 列等宽字符才能对齐，
 * 在手机上必然折行错乱。
 */
const architectureLayers = [
  {
    id: "L0",
    title: "客户端层",
    blurb: "Web 客户端及移动端、第三方接入方。",
    nodes: ["web/（Next.js + React）", "移动端 / 第三方客户端"],
    edge: "REST · WebSocket · WebRTC",
  },
  {
    id: "L1",
    title: "接入层",
    blurb: "HTTP 入口、实时网关与媒体信令。",
    nodes: [
      "user/api · social/api · im/api · trend/api",
      "im/ws —— 鉴权、心跳、ack、在线推送",
      "streaming —— 信令、房间、SFU",
    ],
    edge: "zRPC · 发布/消费 · Redis",
  },
  {
    id: "L2",
    title: "领域服务层",
    blurb: "每个领域一个 zRPC 服务，通过 etcd 发现。",
    nodes: [
      "user/rpc —— 鉴权、资料、验证码",
      "social/rpc —— 好友、群组、申请",
      "im/rpc —— 会话、聊天记录、已读/撤回",
      "trend/rpc —— 动态、评论、点赞、通知",
    ],
    edge: "MySQL · MongoDB · Kafka",
  },
  {
    id: "L3",
    title: "事件与异步层",
    blurb: "Kafka topic 由幂等消费者与定时任务消化。",
    nodes: [
      "chat-transfer · read-transfer",
      "recall-transfer · trend-notify",
      "task/mq —— 落库消息、更新已读、推送事件",
      "task/cron —— 定时统计、清理与扩展任务",
    ],
    edge: "落库 · 更新 · 推送",
  },
  {
    id: "L4",
    title: "数据与运行时基础设施",
    blurb: "每个服务独占自己的库表，禁止跨服务读库。",
    nodes: [
      "MySQL —— 用户、好友、群组、动态、评论",
      "MongoDB —— 聊天记录、已读记录、撤回状态",
      "Redis —— 会话、在线状态、缓存、房间状态",
      "Etcd —— 服务注册与发现",
    ],
    edge: null,
  },
] as const;

/** 服务清单表，对应 README 的服务表格。 */
const services = [
  {
    name: "user",
    layers: "api · rpc",
    responsibility: "账号、鉴权、资料、验证码、用户检索",
  },
  {
    name: "social",
    layers: "api · rpc",
    responsibility: "好友、群组、申请、邀请、公告",
  },
  {
    name: "im",
    layers: "api · rpc · ws",
    responsibility: "会话、聊天记录、已读回执、撤回、WebSocket 网关",
  },
  {
    name: "trend",
    layers: "api · rpc",
    responsibility: "动态、评论、点赞、草稿、媒体、通知",
  },
  {
    name: "task",
    layers: "mq · cron",
    responsibility: "Kafka 消费者与定时任务",
  },
  {
    name: "streaming",
    layers: "sfu · webrtc",
    responsibility: "WebRTC 通话、房间、会议、屏幕共享、直播",
  },
] as const;

/** 技术栈徽章，按关注点分组。 */
const techStack = [
  {
    group: "后端",
    items: ["Go 1.25", "go-zero", "zRPC / gRPC", "goctl"],
  },
  {
    group: "实时",
    items: ["WebSocket", "Kafka", "WebRTC", "Pion"],
  },
  {
    group: "存储",
    items: ["MySQL 8", "MongoDB 7", "Redis 7", "Etcd v3.5"],
  },
  {
    group: "前端",
    items: ["Next.js 16", "React 19", "Bun", "Tailwind CSS"],
  },
] as const;

/**
 * /features 页面 —— 每个领域一个分区，图片左右交替。
 * 能力清单对应 README 的核心能力表。
 */
const featureSections = [
  {
    id: "messaging",
    eyebrow: "即时消息",
    title: "像真实 IM 一样运转的会话",
    blurb:
      "单聊与群聊由 MongoDB 聊天记录支撑，投递、已读与撤回三条链路全部经由 Kafka。",
    shot: "/screenshots/single-chat.webp",
    shotAlt: "HiChat 单聊界面",
    capabilities: [
      "文本、图片、文件、语音、视频消息",
      "引用、@提醒与消息撤回",
      "已读回执与会话级未读状态",
      "会话置顶、免打扰与批量操作",
      "离线消息按 conversationId + seq 拉取",
      "媒体上传在各 api 服务间共享",
    ],
    service: "apps/im",
  },
  {
    id: "social",
    eyebrow: "社交关系",
    title: "好友、群组，以及两者之间的一切",
    blurb:
      "social 服务掌管关系图谱 —— 申请、拉黑、标签，以及从建群到群主转让的完整群生命周期。",
    shot: "/screenshots/friend-list-and-settings.webp",
    shotAlt: "HiChat 好友列表与设置",
    capabilities: [
      "带验证消息的好友申请",
      "备注、标签、拉黑与举报",
      "群组创建、搜索与入群申请",
      "邀请与可分享的邀请令牌",
      "成员管理、角色与管理员操作",
      "群公告与群主转让",
    ],
    service: "apps/social",
  },
  {
    id: "moments",
    eyebrow: "动态空间",
    title: "带真实可见性规则的动态",
    blurb:
      "发布带媒体与逐好友可见范围的动态。评论、回复与点赞经 Kafka 扩散进未读通知收件箱。",
    shot: "/screenshots/moments-feed-and-detail.webp",
    shotAlt: "HiChat 动态流与详情",
    capabilities: [
      "发布图片与视频资源",
      "逐条动态的可见性控制",
      "评论、嵌套回复与点赞",
      "发布前的草稿支持",
      "未读计数与通知收件箱",
      "经 WebSocket 网关在线推送",
    ],
    service: "apps/trend",
  },
  {
    id: "calls",
    eyebrow: "实时音视频",
    title: "由独立 SFU 承载的 WebRTC 通话",
    blurb:
      "streaming 服务独立于主链路运行，负责通话、会议与直播的信令、房间与媒体转发。",
    shot: "/screenshots/group-call-active.webp",
    shotAlt: "HiChat 群组通话进行中",
    capabilities: [
      "单人通话的响铃、接听与拒接",
      "全网格群组通话，最多四人",
      "会议与屏幕共享",
      "经 SFU 的直播推流",
      "房间状态保存在 Redis",
      "基于 Pion WebRTC 构建",
    ],
    service: "apps/streaming",
  },
  {
    id: "realtime",
    eyebrow: "实时网关",
    title: "一个 WebSocket 网关，多个节点",
    blurb:
      "im/ws 掌管连接生命周期：鉴权、心跳、ack 追踪与重试。在线状态存于 Redis，跨节点投递经 Kafka 路由。",
    shot: "/screenshots/conversation-list.webp",
    shotAlt: "HiChat 会话列表",
    capabilities: [
      "连接鉴权与逐用户会话",
      "心跳配合幽灵连接清理",
      "ack 追踪、重试与去重",
      "Redis 支撑的逐节点在线状态",
      "经 Kafka topic 跨节点投递",
      "大规模扇出的批处理广播",
    ],
    service: "apps/im/ws",
  },
  {
    id: "async",
    eyebrow: "异步与任务",
    title: "每个事件背后的幂等消费者",
    blurb:
      "消息落库、已读状态、撤回与动态通知全部是消费端职责 —— 可安全重试，且绝不重复生效。",
    shot: "/screenshots/moments.webp",
    shotAlt: "HiChat 动态空间",
    capabilities: [
      "chat-transfer —— 消息落库 MongoDB",
      "read-transfer —— 更新已读记录",
      "recall-transfer —— 应用撤回状态",
      "trend-notify —— 推送动态通知",
      "带去重键的幂等消费者",
      "带 Redis 单实例锁的定时任务",
    ],
    service: "apps/task",
  },
] as const;

/** /quick-start 页面 —— Docker Compose 路径。 */
const quickStartSteps = [
  {
    n: 1,
    title: "克隆仓库",
    body: "拉取源码。以下所有命令均在仓库根目录执行。",
    code: `git clone https://github.com/iceymoss/go-hichat-api.git
cd go-hichat-api`,
    lang: "bash",
  },
  {
    n: 2,
    title: "启动整套服务",
    body: "Compose 会按依赖顺序并带健康检查地启动 MySQL、Redis、MongoDB、Etcd、Kafka、六个微服务以及 Web 客户端。",
    code: `docker compose up -d --build`,
    lang: "bash",
  },
  {
    n: 3,
    title: "灌入演示数据",
    body: "可选。注册 14 个演示用户，预填会话、好友与动态 —— 即截图中所用的数据集。",
    code: `docker compose --profile mock up mockdata`,
    lang: "bash",
  },
  {
    n: 4,
    title: "打开客户端",
    body: "Web 客户端监听 2470 端口。演示模式下验证码自动填充，可自由注册新账号。",
    code: `open http://localhost:2470`,
    lang: "bash",
  },
] as const;

/** docker-compose.yaml 暴露的端口。 */
const servicePorts = [
  { service: "web", port: "2470", note: "Next.js Web 客户端" },
  { service: "user-api", port: "8887", note: "账号、鉴权、资料" },
  { service: "social-api", port: "8889", note: "好友与群组" },
  { service: "im-api", port: "8890", note: "会话与聊天记录" },
  { service: "trend-api", port: "8891", note: "动态空间" },
  { service: "im-ws", port: "10090", note: "WebSocket 网关" },
  {
    service: "streaming",
    port: "10093",
    note: "WebRTC SFU（另需 50000-50200/udp）",
  },
] as const;

/** 本地开发路径，供不想用 Compose 的贡献者使用。 */
const localDevSteps = [
  {
    title: "仅启动中间件",
    body: "以宿主机可访问的端口启动 MySQL、Redis、MongoDB、Etcd 与 Kafka，Go 服务留给本机原生运行。",
    code: `docker compose -f docker-compose.dependencies.yaml up -d`,
    lang: "bash",
  },
  {
    title: "启动所有 Go 服务",
    body: "该脚本按依赖顺序启动全部服务。RPC 鉴权密钥至少 32 字节，且不得复用 JWT secret。",
    code: `HICHAT_IM_RPC_AUTH_SECRET=<32字节随机值> ./hichat2.sh`,
    lang: "bash",
  },
  {
    title: "或只启动单个服务",
    body: "迭代某一层时更方便。每个服务读取自己的 sample 配置。",
    code: `go run apps/user/api/user.go -f apps/user/api/etc/user-sample.yaml`,
    lang: "bash",
  },
  {
    title: "启动 Web 客户端",
    body: "开发服务器监听 3001 端口。",
    code: `cd web && bun install && bun dev`,
    lang: "bash",
  },
] as const;

/** 文档侧边栏导航。 */
const docsNav = [
  {
    title: "概览",
    items: [
      { label: "简介", href: "/docs" },
      { label: "架构总览", href: "/docs/architecture" },
      { label: "领域模块", href: "/docs/domains" },
    ],
  },
  {
    title: "核心概念",
    items: [
      { label: "核心数据流", href: "/docs/data-flows" },
      { label: "消息生命周期", href: "/docs/message-lifecycle" },
      { label: "实时网关", href: "/docs/realtime-gateway" },
    ],
  },
  {
    title: "指南",
    items: [{ label: "扩展系统", href: "/docs/extending" }],
  },
  {
    title: "项目",
    items: [{ label: "更新日志", href: "/changelog" }],
  },
] as const;

/** 页脚链接分组。`external: true` 渲染为 target="_blank"。 */
const footerGroups = [
  {
    title: "产品",
    items: [
      { label: "功能", href: "/features", external: false },
      { label: "快速开始", href: "/quick-start", external: false },
    ],
  },
  {
    title: "文档",
    items: [
      { label: "设计文档", href: "/docs", external: false },
      { label: "API 参考", href: links.docsApi, external: true },
      { label: "开发指南", href: links.docsDevGuide, external: true },
      { label: "Docker 部署", href: links.dockerDeploy, external: true },
    ],
  },
  {
    title: "社区",
    items: [
      { label: "GitHub", href: links.github, external: true },
      { label: "Issues", href: links.githubIssues, external: true },
      { label: "参与贡献", href: links.contributing, external: true },
    ],
  },
] as const;

/** 组件内的零散 UI 字符串。 */
const ui = {
  openMenu: "打开菜单",
  closeMenu: "关闭菜单",
  githubRepo: "GitHub 仓库",
  github: "GitHub",
  quickStart: "快速开始",
  changelog: "更新日志",
  apiRef: "API 参考",
  docs: "文档",
  starOnGithub: "在 GitHub 上 Star",
  fullGuide: "完整指南",
  dockerDeployDocs: "Docker 部署文档",
  activeDevelopment: "持续开发中",
  heroShotAlt: "HiChat 单聊界面",
  heroTitleLine1: "开源即时通讯",
  heroTitleLine2: "与社交平台",
  heroSubtitle:
    "基于 go-zero 的微服务后端，内置 WebSocket 网关、WebRTC 音视频、Kafka 事件管道，以及功能完整的 Next.js Web 客户端 —— 全部在同一个开源仓库里。",
  featuresEyebrow: "开箱即全",
  featuresTitle: "一套完整的 IM 技术栈，不是演示 demo",
  featuresSubtitle:
    "六个具备生产形态的子系统，各自有真实的契约、存储与异步链路 —— 可读、可跑、可扩展。",
  galleryEyebrow: "真实运行效果",
  galleryTitle: "真实界面，不是设计稿",
  gallerySubtitle: "以下每张截图都来自灌入演示数据后实际运行的 Web 客户端。",
  archEyebrow: "深入内部",
  archTitle: "五个层级，边界清晰",
  archSubtitle:
    "请求自上而下穿过接入层、领域层与事件层。服务之间从不读对方的库表 —— 一切跨越都经由 zRPC 或 Kafka。",
  serviceInventory: "服务清单",
  serviceInventoryNote: "apps/ 目录下的每个服务。",
  colService: "服务",
  colLayers: "层级",
  techStackTitle: "技术栈",
  ctaEyebrow: "一条命令启动",
  ctaTitle: "一分钟内跑起来",
  ctaSubtitle:
    "Docker Compose 会拉起全部依赖 —— MySQL、Redis、MongoDB、Etcd、Kafka、六个微服务，以及 Web 客户端。",
  ctaAfterStartup: "启动完成后，打开",
  ctaDemoLogin: "演示账号：",
  ctaAutoFill: "演示模式下验证码自动填充",
  footerBlurb:
    "基于 Go 微服务、go-zero、Next.js、Kafka 与 WebRTC 构建的开源 IM 与社交平台。",
  footerLicense: "基于 Apache 2.0 许可证发布。",
  switchLanguage: "切换语言",
} as const;

export const zh = {
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
