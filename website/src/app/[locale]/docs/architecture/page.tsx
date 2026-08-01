import type { Metadata } from "next";
import {
  ArrowDown,
  ArrowRight,
  Boxes,
  Cable,
  Database,
  Globe2,
  Radio,
  ServerCog,
} from "lucide-react";
import { ArchitectureDiagrams } from "@/components/docs/ArchitectureDiagrams";
import { resolveLocale, type Locale } from "@/i18n";

type EdgeKind = "rpc" | "kafka" | "websocket";
type MatrixRow = { caller: string; targets: string; purpose: string };
type Content = {
  metadata: { title: string; description: string };
  eyebrow: string;
  title: string;
  lead: string;
  diagramTitle: string;
  diagramHint: string;
  layers: {
    client: string;
    edge: string;
    domain: string;
    async: string;
    infrastructure: string;
  };
  clientNote: string;
  gatewayNote: string;
  domainNote: string;
  asyncNote: string;
  sfuNote: string;
  edgeLabels: Record<EdgeKind, string>;
  flows: {
    rpc: string[];
    kafka: string[];
    websocket: string[];
  };
  matrixTitle: string;
  matrixIntro: string;
  matrixHeaders: [string, string, string];
  matrix: MatrixRow[];
  infraTitle: string;
  infraIntro: string;
  infra: { name: string; use: string }[];
  boundaryTitle: string;
  boundaries: string[];
};

const content = {
  en: {
    metadata: {
      title: "Architecture Overview - HiChat Docs",
      description:
        "A code-aligned overview of HiChat's HTTP, RPC, WebSocket, Kafka, SFU, and infrastructure boundaries.",
    },
    eyebrow: "SYSTEM MAP / HICHAT 2.0",
    title: "Architecture overview",
    lead:
      "HiChat separates request ingress, domain ownership, asynchronous delivery, and realtime media. This map follows clients that are actually constructed by the running modules, not every field that may appear in a configuration file.",
    diagramTitle: "Runtime topology",
    diagramHint: "Read top to bottom. The detailed edge register below names every cross-module call shown here.",
    layers: {
      client: "Clients",
      edge: "Ingress",
      domain: "Domain RPC",
      async: "Async & realtime workers",
      infrastructure: "Infrastructure",
    },
    clientNote: "Web, mobile, and service consumers",
    gatewayNote: "HTTP request/response, persistent sockets, and WebRTC signaling",
    domainNote: "Business ownership and service-to-service contracts",
    asyncNote: "Eventually consistent work, schedules, and media forwarding",
    sfuNote: "Pion WebRTC SFU is embedded in Streaming; it is not a separate microservice.",
    edgeLabels: {
      rpc: "Synchronous RPC",
      kafka: "Kafka event",
      websocket: "Internal WebSocket",
    },
    flows: {
      rpc: [
        "User API → User RPC",
        "Social API → Social / User RPC",
        "IM API → IM / User / Social RPC",
        "Trend API → Trend / User / Social RPC",
        "IM WebSocket → Social RPC",
        "Streaming → Social RPC",
        "Social RPC → User RPC",
        "IM RPC → Social RPC",
        "Task MQ → IM / Social RPC",
        "Task Cron → Social RPC",
      ],
      kafka: [
        "IM WebSocket → Kafka → Task MQ",
        "IM API → Kafka → Task MQ",
        "Trend API → Kafka → Task MQ",
        "Social outbox relays → Kafka → Task MQ",
      ],
      websocket: [
        "Task MQ → IM WebSocket",
        "Streaming → IM WebSocket",
      ],
    },
    matrixTitle: "Synchronous call matrix",
    matrixIntro:
      "Only constructed RPC clients are listed. Etcd-backed discovery and RPC authentication are transport concerns and do not add domain edges.",
    matrixHeaders: ["Caller", "RPC target", "Responsibility"],
    matrix: [
      { caller: "User API", targets: "User", purpose: "Accounts, authentication, and profiles" },
      { caller: "Social API", targets: "Social · User", purpose: "Relations and groups, plus profile enrichment" },
      { caller: "IM API", targets: "IM · User · Social", purpose: "Conversations/history, participants, and relation checks" },
      { caller: "Trend API", targets: "Trend · User · Social", purpose: "Posts, actor profiles, and social visibility" },
      { caller: "IM WebSocket", targets: "Social", purpose: "Friend/group authorization, including @all role checks" },
      { caller: "Streaming", targets: "Social", purpose: "Friend and group membership checks" },
      { caller: "Social RPC", targets: "User", purpose: "User lookups owned by the User domain" },
      { caller: "IM RPC", targets: "Social", purpose: "Membership and relation checks" },
      { caller: "Task MQ", targets: "IM · Social", purpose: "Message-side updates and group membership fallback" },
      { caller: "Task Cron", targets: "Social", purpose: "Scheduled invitation expiration" },
    ],
    infraTitle: "Infrastructure ownership",
    infraIntro: "The lower layer is shared infrastructure, while schemas and records remain owned by their business modules.",
    infra: [
      { name: "MySQL", use: "User, Social, IM metadata, Trend, and worker-side updates" },
      { name: "MongoDB", use: "Chat logs, read records, and conversation documents" },
      { name: "Redis", use: "Caches, replay protection, presence, relation cache, and distributed locks" },
      { name: "Etcd", use: "Registration and discovery for go-zero RPC services" },
      { name: "Kafka", use: "Messages, reads, recalls, notifications, and relation-change delivery" },
      { name: "File", use: "Local attachment storage exposed by IM API" },
      { name: "STUN / TURN", use: "ICE traversal for Streaming WebRTC sessions" },
    ],
    boundaryTitle: "Boundaries that matter",
    boundaries: [
      "HTTP APIs orchestrate requests; domain state is exposed through RPC rather than another service's database.",
      "Social writes relation and notification outboxes transactionally, then its in-process relays publish them to Kafka.",
      "Task MQ consumes Kafka and uses an authenticated internal WebSocket client for online push through IM WebSocket.",
      "Streaming owns signaling, rooms, and the embedded Pion SFU; it uses IM WebSocket only for call-control notification push.",
    ],
  },
  zh: {
    metadata: {
      title: "架构总览 - HiChat 文档",
      description: "与代码一致的 HiChat HTTP、RPC、WebSocket、Kafka、SFU 与基础设施边界总览。",
    },
    eyebrow: "SYSTEM MAP / HICHAT 2.0",
    title: "架构总览",
    lead:
      "HiChat 将请求接入、领域所有权、异步投递和实时音视频分层组织。本图只依据运行模块中真实创建的客户端，不会把配置文件里可能出现但未使用的字段画成依赖。",
    diagramTitle: "运行时拓扑",
    diagramHint: "从上向下阅读；图下方的边清单逐条列出这里展示的跨模块调用。",
    layers: {
      client: "客户端",
      edge: "接入层",
      domain: "领域 RPC",
      async: "异步与实时处理",
      infrastructure: "基础设施",
    },
    clientNote: "Web、移动端与服务消费者",
    gatewayNote: "HTTP 请求响应、长连接与 WebRTC 信令",
    domainNote: "业务数据所有权与服务间契约",
    asyncNote: "最终一致任务、定时调度与媒体转发",
    sfuNote: "Pion WebRTC SFU 内置于 Streaming，不是独立微服务。",
    edgeLabels: {
      rpc: "同步 RPC",
      kafka: "Kafka 事件",
      websocket: "内部 WebSocket",
    },
    flows: {
      rpc: [
        "User API → User RPC",
        "Social API → Social / User RPC",
        "IM API → IM / User / Social RPC",
        "Trend API → Trend / User / Social RPC",
        "IM WebSocket → Social RPC",
        "Streaming → Social RPC",
        "Social RPC → User RPC",
        "IM RPC → Social RPC",
        "Task MQ → IM / Social RPC",
        "Task Cron → Social RPC",
      ],
      kafka: [
        "IM WebSocket → Kafka → Task MQ",
        "IM API → Kafka → Task MQ",
        "Trend API → Kafka → Task MQ",
        "Social outbox relays → Kafka → Task MQ",
      ],
      websocket: [
        "Task MQ → IM WebSocket",
        "Streaming → IM WebSocket",
      ],
    },
    matrixTitle: "同步调用矩阵",
    matrixIntro: "这里只列出代码中实际构造的 RPC 客户端。基于 Etcd 的服务发现与 RPC 鉴权属于传输机制，不会形成额外领域边。",
    matrixHeaders: ["调用方", "RPC 目标", "职责"],
    matrix: [
      { caller: "User API", targets: "User", purpose: "账号、认证与用户资料" },
      { caller: "Social API", targets: "Social · User", purpose: "关系和群组，以及用户资料补全" },
      { caller: "IM API", targets: "IM · User · Social", purpose: "会话与历史、参与者资料、关系校验" },
      { caller: "Trend API", targets: "Trend · User · Social", purpose: "动态、发布者资料与社交可见性" },
      { caller: "IM WebSocket", targets: "Social", purpose: "好友/群成员鉴权，包括 @所有人角色点查" },
      { caller: "Streaming", targets: "Social", purpose: "好友与群成员关系校验" },
      { caller: "Social RPC", targets: "User", purpose: "通过 User 领域查询用户" },
      { caller: "IM RPC", targets: "Social", purpose: "成员与关系校验" },
      { caller: "Task MQ", targets: "IM · Social", purpose: "消息侧更新与群成员缓存回源" },
      { caller: "Task Cron", targets: "Social", purpose: "定时处理邀请过期" },
    ],
    infraTitle: "基础设施归属",
    infraIntro: "底层设施由多个服务共享，但 schema 与业务记录仍由对应业务模块负责。",
    infra: [
      { name: "MySQL", use: "User、Social、IM 元数据、Trend 及 worker 侧更新" },
      { name: "MongoDB", use: "聊天记录、已读记录与会话文档" },
      { name: "Redis", use: "缓存、防重放、在线状态、关系缓存与分布式锁" },
      { name: "Etcd", use: "go-zero RPC 服务注册与发现" },
      { name: "Kafka", use: "消息、已读、撤回、通知与关系变更投递" },
      { name: "File", use: "由 IM API 暴露的本地附件存储" },
      { name: "STUN / TURN", use: "Streaming WebRTC 会话的 ICE 穿透" },
    ],
    boundaryTitle: "关键边界",
    boundaries: [
      "HTTP API 负责编排请求；跨领域数据通过 RPC 获取，而不是读取其他服务的数据库。",
      "Social 在业务事务内写关系与通知 outbox，再由进程内 relay 发布到 Kafka。",
      "Task MQ 消费 Kafka，并使用已鉴权的内部 WebSocket 客户端经 IM WebSocket 完成在线推送。",
      "Streaming 负责信令、房间和内置 Pion SFU；它只通过 IM WebSocket 推送通话控制通知。",
    ],
  },
} satisfies Record<Locale, Content>;

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  return content[locale].metadata;
}

const edgeStyles: Record<EdgeKind, { dot: string; line: string; panel: string }> = {
  rpc: { dot: "bg-cyan-400", line: "border-cyan-400/50", panel: "border-cyan-400/25 bg-cyan-400/[0.04]" },
  kafka: { dot: "bg-amber-400", line: "border-amber-400/50 border-dashed", panel: "border-amber-400/25 bg-amber-400/[0.04]" },
  websocket: { dot: "bg-fuchsia-400", line: "border-fuchsia-400/50 border-dotted", panel: "border-fuchsia-400/25 bg-fuchsia-400/[0.04]" },
};

const gatewayModules = ["User HTTP API", "Social HTTP API", "IM HTTP API", "Trend HTTP API", "IM WebSocket", "Streaming Signaling"];
const rpcModules = ["User RPC", "Social RPC", "IM RPC", "Trend RPC"];
const asyncModules = ["Social outbox relays", "Task MQ", "Task Cron", "Streaming · Pion SFU"];
const infrastructureModules = ["MySQL", "MongoDB", "Redis", "Etcd", "Kafka", "File", "STUN / TURN"];

export default async function ArchitecturePage({ params }: Props) {
  const locale = resolveLocale((await params).locale);
  const page = content[locale];

  return (
    <article className="not-prose -mx-1 pb-12 sm:mx-0">
      <header className="relative overflow-hidden rounded-2xl border border-border bg-card px-5 py-8 sm:px-8 sm:py-10">
        <div className="absolute -right-16 -top-20 h-56 w-56 rounded-full bg-cyan-400/10 blur-3xl" />
        <div className="absolute -bottom-24 left-1/3 h-48 w-48 rounded-full bg-fuchsia-400/10 blur-3xl" />
        <div className="relative max-w-3xl">
          <p className="mb-3 font-mono text-[11px] font-semibold tracking-[0.22em] text-cyan-500 dark:text-cyan-300">{page.eyebrow}</p>
          <h1 className="text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">{page.title}</h1>
          <p className="mt-4 text-base leading-7 text-muted-foreground sm:text-lg">{page.lead}</p>
        </div>
      </header>

      <section className="mt-8" aria-labelledby="runtime-topology">
        <div className="mb-5 flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
          <div>
            <h2 id="runtime-topology" className="text-2xl font-semibold tracking-tight text-foreground">{page.diagramTitle}</h2>
            <p className="mt-1 max-w-2xl text-sm leading-6 text-muted-foreground">{page.diagramHint}</p>
          </div>
          <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
            {(Object.keys(page.edgeLabels) as EdgeKind[]).map((kind) => (
              <span key={kind} className="inline-flex items-center gap-1.5"><span className={`h-2 w-2 rounded-full ${edgeStyles[kind].dot}`} />{page.edgeLabels[kind]}</span>
            ))}
          </div>
        </div>

        <ArchitectureDiagrams locale={locale} />

        <div className="mt-8 overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
          <ArchitectureLayer icon={<Globe2 className="size-4" />} label={page.layers.client} note={page.clientNote} tone="cyan">
            <ModuleCard name="Web / Mobile / Desktop" accent="cyan" />
          </ArchitectureLayer>
          <LayerConnector />
          <ArchitectureLayer icon={<Cable className="size-4" />} label={page.layers.edge} note={page.gatewayNote} tone="violet">
            {gatewayModules.map((name) => (
              <ModuleCard key={name} name={name} accent={name.includes("WebSocket") ? "fuchsia" : name.includes("Streaming") ? "emerald" : "violet"} />
            ))}
          </ArchitectureLayer>
          <LayerConnector />
          <ArchitectureLayer icon={<ServerCog className="size-4" />} label={page.layers.domain} note={page.domainNote} tone="cyan">
            {rpcModules.map((name) => <ModuleCard key={name} name={name} accent="cyan" />)}
          </ArchitectureLayer>
          <LayerConnector />
          <ArchitectureLayer icon={<Radio className="size-4" />} label={page.layers.async} note={page.asyncNote} tone="amber">
            {asyncModules.map((name) => <ModuleCard key={name} name={name} accent={name.includes("SFU") ? "emerald" : "amber"} />)}
            <p className="col-span-full mt-1 text-xs leading-5 text-muted-foreground">{page.sfuNote}</p>
          </ArchitectureLayer>
          <LayerConnector />
          <ArchitectureLayer icon={<Database className="size-4" />} label={page.layers.infrastructure} tone="slate">
            {infrastructureModules.map((name) => <ModuleCard key={name} name={name} accent="slate" compact />)}
          </ArchitectureLayer>
        </div>

        <div className="mt-5 grid gap-3 lg:grid-cols-3">
          {(Object.keys(page.flows) as EdgeKind[]).map((kind) => (
            <EdgeRegister key={kind} kind={kind} label={page.edgeLabels[kind]} rows={page.flows[kind]} />
          ))}
        </div>
      </section>

      <section className="mt-12" aria-labelledby="sync-matrix">
        <div className="max-w-3xl">
          <h2 id="sync-matrix" className="text-2xl font-semibold tracking-tight text-foreground">{page.matrixTitle}</h2>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{page.matrixIntro}</p>
        </div>
        <div className="mt-5 overflow-x-auto rounded-2xl border border-border bg-card">
          <table className="w-full min-w-[680px] border-collapse text-left text-sm">
            <thead className="bg-muted/50 font-mono text-[11px] uppercase tracking-wider text-muted-foreground">
              <tr>{page.matrixHeaders.map((header) => <th key={header} className="border-b border-border px-5 py-3 font-medium">{header}</th>)}</tr>
            </thead>
            <tbody>
              {page.matrix.map((row) => (
                <tr key={row.caller} className="border-b border-border/70 last:border-0 hover:bg-muted/30">
                  <td className="whitespace-nowrap px-5 py-4 font-mono text-xs font-semibold text-foreground">{row.caller}</td>
                  <td className="whitespace-nowrap px-5 py-4 text-cyan-600 dark:text-cyan-300">{row.targets}</td>
                  <td className="px-5 py-4 leading-6 text-muted-foreground">{row.purpose}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="mt-12 grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
        <div className="rounded-2xl border border-border bg-card p-5 sm:p-6">
          <div className="flex items-center gap-2"><Database className="h-5 w-5 text-cyan-500" /><h2 className="text-xl font-semibold text-foreground">{page.infraTitle}</h2></div>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{page.infraIntro}</p>
          <div className="mt-5 grid gap-3 sm:grid-cols-2">
            {page.infra.map((item) => (
              <div key={item.name} className="rounded-xl border border-border bg-background/60 p-4">
                <p className="font-mono text-xs font-semibold text-foreground">{item.name}</p>
                <p className="mt-1.5 text-xs leading-5 text-muted-foreground">{item.use}</p>
              </div>
            ))}
          </div>
        </div>
        <div className="rounded-2xl border border-border bg-card p-5 sm:p-6">
          <div className="flex items-center gap-2"><Boxes className="h-5 w-5 text-fuchsia-500" /><h2 className="text-xl font-semibold text-foreground">{page.boundaryTitle}</h2></div>
          <div className="mt-5 space-y-5">
            {page.boundaries.map((boundary, index) => (
              <div key={boundary} className="flex gap-3">
                <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-fuchsia-400/30 bg-fuchsia-400/10 font-mono text-[10px] text-fuchsia-600 dark:text-fuchsia-300">{String(index + 1).padStart(2, "0")}</span>
                <p className="text-sm leading-6 text-muted-foreground">{boundary}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </article>
  );
}

function ArchitectureLayer({ icon, label, note, tone, children }: { icon: React.ReactNode; label: string; note?: string; tone: "cyan" | "violet" | "amber" | "slate"; children: React.ReactNode }) {
  const tones = {
    cyan: "text-cyan-600 dark:text-cyan-300",
    violet: "text-violet-600 dark:text-violet-300",
    amber: "text-amber-600 dark:text-amber-300",
    slate: "text-muted-foreground",
  };

  return (
    <div className="grid gap-4 border-b border-border/60 p-4 last:border-0 sm:p-5 lg:grid-cols-[180px_1fr]">
      <div>
        <div className={`flex items-center gap-2 ${tones[tone]}`}>
          {icon}
          <h3 className="font-mono text-xs font-semibold uppercase tracking-wider">{label}</h3>
        </div>
        {note && <p className="mt-2 text-xs leading-5 text-muted-foreground">{note}</p>}
      </div>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-6">{children}</div>
    </div>
  );
}

function ModuleCard({ name, accent, compact = false }: { name: string; accent: "cyan" | "violet" | "fuchsia" | "amber" | "emerald" | "slate"; compact?: boolean }) {
  const accents = {
    cyan: "border-l-cyan-400",
    violet: "border-l-violet-400",
    fuchsia: "border-l-fuchsia-400",
    amber: "border-l-amber-400",
    emerald: "border-l-emerald-400",
    slate: "border-l-slate-400",
  };

  return (
    <div className={`flex items-center rounded-lg border border-border border-l-2 bg-background/70 px-3 ${compact ? "min-h-10 py-2" : "min-h-14 py-3"} ${accents[accent]}`}>
      <span className="font-mono text-[11px] font-medium leading-4 text-foreground">{name}</span>
    </div>
  );
}

function LayerConnector() {
  return (
    <div className="flex h-7 items-center justify-center border-b border-border/60 bg-muted/20" aria-hidden="true">
      <ArrowDown className="size-3.5 text-muted-foreground/70" />
    </div>
  );
}

function EdgeRegister({ kind, label, rows }: { kind: EdgeKind; label: string; rows: string[] }) {
  const style = edgeStyles[kind];
  return (
    <div className={`rounded-xl border p-4 ${style.panel}`}>
      <div className="mb-3 flex items-center gap-2"><span className={`h-2 w-2 rounded-full ${style.dot}`} /><h3 className="font-mono text-xs font-semibold uppercase tracking-wider text-foreground">{label}</h3></div>
      <div className="space-y-2.5">
        {rows.map((row) => <div key={row} className={`flex gap-2 border-l pl-2.5 ${style.line}`}><ArrowRight className="mt-0.5 h-3 w-3 shrink-0 text-muted-foreground" /><code className="text-[10px] leading-4 text-muted-foreground">{row}</code></div>)}
      </div>
    </div>
  );
}
