import type { Metadata } from "next";
import type { ReactNode } from "react";
import { DomainDiagrams } from "@/components/docs/DomainDiagrams";
import { resolveLocale, type Locale } from "@/i18n";

type Domain = {
  name: string;
  layers: string[];
  responsibility: string;
  dependencies: string;
  storage: string;
  boundary: string;
};

type OwnershipRow = {
  domain: string;
  data: string;
  authority: string;
  notes: string;
};

type PageContent = {
  metadata: { title: string; description: string };
  eyebrow: string;
  title: string;
  lead: string;
  legend: string;
  labels: {
    layers: string;
    responsibility: string;
    dependencies: string;
    storage: string;
    boundary: string;
  };
  domains: Domain[];
  ownership: {
    title: string;
    intro: string;
    headers: [string, string, string, string];
    rows: OwnershipRow[];
  };
  rule: {
    title: string;
    text: string;
  };
};

const pageContent = {
  en: {
    metadata: {
      title: "Domain Map - HiChat Docs",
      description:
        "Responsibilities, runtime dependencies, authoritative stores, and boundaries across HiChat's six domains.",
    },
    eyebrow: "Architecture / Domain map",
    title: "Six domains, explicit ownership",
    lead:
      "HiChat separates account identity, the social graph, messaging, activity feeds, background work, and realtime media. This map describes the processes that run today, their synchronous calls, and the data each domain is allowed to own.",
    legend:
      "Dependencies below are synchronous runtime calls. Kafka flows are called out separately as asynchronous bridges.",
    labels: {
      layers: "Layers / processes",
      responsibility: "Responsibility",
      dependencies: "Synchronous dependencies",
      storage: "Authoritative data",
      boundary: "Key boundary",
    },
    domains: [
      {
        name: "User",
        layers: ["user/api", "user/rpc"],
        responsibility:
          "Registration, login, JWT issuance, profiles, user settings, favorites, and custom emoji metadata.",
        dependencies:
          "user/api calls user/rpc. The RPC process has no synchronous dependency on another business domain.",
        storage:
          "MySQL is authoritative for accounts and user-owned settings; Redis is a cache, not the source of truth.",
        boundary:
          "User owns identity and profile facts. Other domains request them through user/rpc rather than reading user tables.",
      },
      {
        name: "Social",
        layers: ["social/api", "social/rpc", "outbox relays (inside social/rpc)"],
        responsibility:
          "Friend relationships, friend requests, groups, membership, invitations, roles, and relationship-change publication.",
        dependencies:
          "social/api calls social/rpc and user/rpc. social/rpc calls user/rpc when workflows need identity validation or profile facts.",
        storage:
          "MySQL owns the social graph, groups, requests, invitations, and transactional relation/notification outboxes. Redis holds derived relationship cache and relay locks.",
        boundary:
          "Mutations and outbox rows commit in the same Social transaction. Relays run inside social/rpc and publish to Kafka after commit; Task consumes those events asynchronously.",
      },
      {
        name: "IM",
        layers: ["im/api", "im/rpc", "im/ws"],
        responsibility:
          "Conversations, message history queries, recalls, read state, WebSocket connections, presence, ACK handling, and online push.",
        dependencies:
          "im/api calls im/rpc, social/rpc, and user/rpc. im/rpc calls social/rpc. im/ws calls social/rpc for live authorization checks such as @all role validation.",
        storage:
          "MongoDB owns chat logs and conversation documents. MySQL owns durable notifications and their read state. Redis provides presence, relationship gates, and caches.",
        boundary:
          "The API/RPC query and command surface is separate from the WebSocket hot path. Message, read, and recall events cross Kafka to Task for asynchronous persistence and delivery work.",
      },
      {
        name: "Trend",
        layers: ["trend/api", "trend/rpc"],
        responsibility:
          "Posts, drafts, likes, comments, feed queries, blocks, and activity notification production.",
        dependencies:
          "trend/api calls trend/rpc, social/rpc, and user/rpc to assemble and authorize HTTP workflows. trend/rpc has no cross-domain synchronous client.",
        storage:
          "MySQL owns posts, drafts, likes, comments, and Trend message records; Redis tracks publish-rate state.",
        boundary:
          "Trend owns feed state, not identity, relationships, or notification delivery. Trend notification events go to Kafka and are handled by Task.",
      },
      {
        name: "Task",
        layers: ["task/mq", "task/cron"],
        responsibility:
          "Kafka consumption, asynchronous persistence and push orchestration, relation-cache maintenance, and scheduled maintenance.",
        dependencies:
          "task/mq calls social/rpc and im/rpc and pushes through im/ws. task/cron currently calls social/rpc to expire pending group invitations in bounded batches.",
        storage:
          "Task has no independent business source of truth. Consumers update the owning domain's stores and use Redis for cache, locks, and coordination.",
        boundary:
          "MQ is the asynchronous bridge, not a public domain API. Cron currently schedules group-invitation expiration; Social performs the authoritative mutation.",
      },
      {
        name: "Streaming",
        layers: ["HTTP + WebSocket signaling", "in-memory call/session managers", "in-process Pion SFU"],
        responsibility:
          "Call setup and control signaling, ICE/TURN information, room/session lifecycle, and realtime audio/video coordination.",
        dependencies:
          "Streaming calls social/rpc for relationship and membership checks. It connects to im/ws as a system user to push call-control signaling.",
        storage:
          "Active call sessions, rooms, and SFU peer state are authoritative only in process memory today. Pion PeerConnections and tracks exist in the Streaming process; Redis supplies relationship-cache data.",
        boundary:
          "Signaling is server-mediated, while media is either client-to-client P2P or routed through the in-process Pion SFU. Restarting the process loses active in-memory sessions; media bytes do not pass through IM.",
      },
    ],
    ownership: {
      title: "Data ownership",
      intro:
        "Ownership identifies the domain that defines invariants and exposes access. A worker may perform a write on that domain's behalf without becoming the owner.",
      headers: ["Domain", "Owned data", "Authority", "Derived / transient"],
      rows: [
        {
          domain: "User",
          data: "Accounts, profiles, settings, favorites, emoji metadata",
          authority: "MySQL",
          notes: "Redis cache; JWTs are signed credentials",
        },
        {
          domain: "Social",
          data: "Friend graph, groups, membership, requests, invitations, outboxes",
          authority: "MySQL",
          notes: "Redis relationship cache and relay locks",
        },
        {
          domain: "IM",
          data: "Chat logs, conversations, notifications, notification read state",
          authority: "MongoDB + MySQL notifications",
          notes: "Redis presence and caches",
        },
        {
          domain: "Trend",
          data: "Posts, drafts, likes, comments, Trend messages",
          authority: "MySQL",
          notes: "Redis cache",
        },
        {
          domain: "Task",
          data: "No independent business records",
          authority: "Owning domain's store",
          notes: "Kafka offsets plus Redis locks/coordination",
        },
        {
          domain: "Streaming",
          data: "Active calls, rooms, peer and track state",
          authority: "Process memory (current implementation)",
          notes: "Ephemeral; lost on process restart",
        },
      ],
    },
    rule: {
      title: "The boundary rule",
      text:
        "A service does not directly read another domain's database. Cross-domain facts come from the owning RPC contract, while Kafka carries asynchronous events. Some background implementation code may currently access shared settings or stores; that is an implementation constraint, not a transfer of data ownership.",
    },
  },
  zh: {
    metadata: {
      title: "领域模块 - HiChat 文档",
      description: "说明 HiChat 六个领域的职责、运行时依赖、权威存储与关键边界。",
    },
    eyebrow: "架构 / 领域地图",
    title: "六个领域，清晰归属",
    lead:
      "HiChat 将账号身份、社交关系、即时消息、动态空间、后台任务和实时音视频拆分为独立领域。本页描述当前实际运行的进程、同步调用，以及每个领域可以拥有的数据。",
    legend: "下文的依赖均指运行时同步调用；Kafka 链路会单独标注为异步桥。",
    labels: {
      layers: "层次 / 进程",
      responsibility: "职责",
      dependencies: "真实同步依赖",
      storage: "权威数据",
      boundary: "关键边界",
    },
    domains: [
      {
        name: "User",
        layers: ["user/api", "user/rpc"],
        responsibility: "注册、登录、JWT 签发、用户资料、用户设置、收藏及自定义表情元数据。",
        dependencies: "user/api 调用 user/rpc；RPC 进程不依赖其他业务域的同步接口。",
        storage: "MySQL 是账号及用户设置的权威存储；Redis 仅为缓存，不是事实来源。",
        boundary: "User 拥有身份和资料事实；其他领域通过 user/rpc 获取，不直接读取 User 表。",
      },
      {
        name: "Social",
        layers: ["social/api", "social/rpc", "outbox relay（运行于 social/rpc 内）"],
        responsibility: "好友关系、好友申请、群组、成员、邀请、角色，以及关系变更事件发布。",
        dependencies:
          "social/api 调用 social/rpc 和 user/rpc；social/rpc 在需要身份校验或用户资料的流程中调用 user/rpc。",
        storage:
          "MySQL 拥有社交图、群组、申请、邀请，以及事务性关系/通知 outbox；Redis 保存派生关系缓存和 relay 锁。",
        boundary:
          "业务变更与 outbox 行在同一个 Social 事务中提交。relay 位于 social/rpc 进程内，提交后发布 Kafka；Task 再异步消费。",
      },
      {
        name: "IM",
        layers: ["im/api", "im/rpc", "im/ws"],
        responsibility: "会话、聊天记录查询、撤回、已读状态、WebSocket 连接、在线状态、ACK 与在线推送。",
        dependencies:
          "im/api 调用 im/rpc、social/rpc 和 user/rpc；im/rpc 调用 social/rpc；im/ws 调用 social/rpc 执行实时鉴权点查，例如 @所有人角色校验。",
        storage:
          "MongoDB 拥有聊天记录和会话文档；MySQL 拥有持久化 notifications 及其已读状态；Redis 提供在线状态、关系闸门和缓存。",
        boundary:
          "API/RPC 查询与命令面和 WebSocket 热路径分离。消息、已读与撤回事件通过 Kafka 交给 Task，异步完成持久化和投递工作。",
      },
      {
        name: "Trend",
        layers: ["trend/api", "trend/rpc"],
        responsibility: "动态、草稿、点赞、评论、Feed 查询、屏蔽，以及动态通知事件生产。",
        dependencies:
          "trend/api 调用 trend/rpc、social/rpc 和 user/rpc 来组装和鉴权 HTTP 流程；trend/rpc 没有跨域同步客户端。",
        storage: "MySQL 拥有动态、草稿、点赞、评论和 Trend 消息记录；Redis 记录发布频率状态。",
        boundary: "Trend 拥有 Feed 状态，不拥有身份、关系或通知投递。动态通知进入 Kafka，由 Task 处理。",
      },
      {
        name: "Task",
        layers: ["task/mq", "task/cron"],
        responsibility: "Kafka 消费、异步持久化与推送编排、关系缓存维护，以及定时维护任务。",
        dependencies:
          "task/mq 调用 social/rpc、im/rpc，并通过 im/ws 推送；task/cron 当前调用 social/rpc，分批将待处理群邀请置为过期。",
        storage: "Task 没有独立的业务事实来源。消费者代表所属领域更新其存储，并使用 Redis 完成缓存、锁和协调。",
        boundary: "MQ 是异步桥，不是公开业务 API。Cron 当前只调度群邀请过期，权威变更仍由 Social 执行。",
      },
      {
        name: "Streaming",
        layers: ["HTTP + WebSocket signaling", "内存 call/session manager", "进程内 Pion SFU"],
        responsibility: "通话建立与控制信令、ICE/TURN 信息、房间/会话生命周期，以及实时音视频协调。",
        dependencies:
          "Streaming 调用 social/rpc 校验好友关系和群成员关系；同时以系统用户连接 im/ws，推送通话控制信令。",
        storage:
          "当前活动通话 session、房间和 SFU peer 状态只以内存为权威；Pion PeerConnection 与 track 位于 Streaming 进程内；Redis 提供关系缓存数据。",
        boundary:
          "信令由服务端协调，媒体则采用客户端间 P2P 或经进程内 Pion SFU 转发。进程重启会丢失活动内存 session；媒体数据不经过 IM。",
      },
    ],
    ownership: {
      title: "数据归属",
      intro: "归属表示由哪个领域定义数据不变量并提供访问接口。Worker 可以代表该领域写入，但不会因此成为数据所有者。",
      headers: ["领域", "拥有的数据", "权威存储", "派生 / 临时数据"],
      rows: [
        {
          domain: "User",
          data: "账号、资料、设置、收藏、表情元数据",
          authority: "MySQL",
          notes: "Redis 缓存；JWT 是签名凭证",
        },
        {
          domain: "Social",
          data: "好友图、群组、成员、申请、邀请、outbox",
          authority: "MySQL",
          notes: "Redis 关系缓存与 relay 锁",
        },
        {
          domain: "IM",
          data: "聊天记录、会话、notifications、通知已读状态",
          authority: "MongoDB + MySQL notifications",
          notes: "Redis 在线状态与缓存",
        },
        {
          domain: "Trend",
          data: "动态、草稿、点赞、评论、Trend 消息",
          authority: "MySQL",
          notes: "Redis 缓存",
        },
        {
          domain: "Task",
          data: "无独立业务记录",
          authority: "对应所属领域的存储",
          notes: "Kafka offset 与 Redis 锁/协调状态",
        },
        {
          domain: "Streaming",
          data: "活动通话、房间、peer 与 track 状态",
          authority: "进程内存（当前实现）",
          notes: "临时数据，进程重启即丢失",
        },
      ],
    },
    rule: {
      title: "边界规则",
      text:
        "服务不跨域直接读取对方数据库。跨域事实来自所有者的 RPC 契约，Kafka 则承载异步事件。部分后台实现目前可能读取共享设置或存储；这是现实现状的约束，不代表数据归属发生转移。",
    },
  },
} satisfies Record<Locale, PageContent>;

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  return pageContent[locale].metadata;
}

export default async function DomainsPage({ params }: Props) {
  const locale = resolveLocale((await params).locale);
  const content = pageContent[locale];

  return (
    <article className="not-prose">
      <header className="border-b border-border pb-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-brand">
          {content.eyebrow}
        </p>
        <h1 className="mt-3 max-w-3xl text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
          {content.title}
        </h1>
        <p className="mt-4 max-w-3xl text-base leading-7 text-muted-foreground sm:text-lg">
          {content.lead}
        </p>
        <p className="mt-5 max-w-3xl border-l-2 border-brand/60 pl-4 text-sm leading-6 text-muted-foreground">
          {content.legend}
        </p>
      </header>

      <DomainDiagrams locale={locale} />

      <section className="mt-8 grid gap-5 xl:grid-cols-2" aria-label={content.title}>
        {content.domains.map((domain, index) => (
          <section
            key={domain.name}
            className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm"
          >
            <div className="flex items-center justify-between border-b border-border bg-muted/30 px-5 py-4 sm:px-6">
              <h2 className="text-xl font-semibold text-foreground">{domain.name}</h2>
              <span className="font-mono text-xs text-muted-foreground">
                {String(index + 1).padStart(2, "0")}
              </span>
            </div>
            <div className="space-y-5 p-5 sm:p-6">
              <DomainField label={content.labels.layers}>
                <div className="flex flex-wrap gap-2">
                  {domain.layers.map((layer) => (
                    <code
                      key={layer}
                      className="rounded-md border border-border bg-muted/50 px-2 py-1 text-xs text-foreground"
                    >
                      {layer}
                    </code>
                  ))}
                </div>
              </DomainField>
              <DomainField label={content.labels.responsibility} text={domain.responsibility} />
              <DomainField label={content.labels.dependencies} text={domain.dependencies} />
              <DomainField label={content.labels.storage} text={domain.storage} />
              <div className="rounded-xl border border-brand/20 bg-brand/5 p-4">
                <DomainField label={content.labels.boundary} text={domain.boundary} />
              </div>
            </div>
          </section>
        ))}
      </section>

      <section className="mt-12">
        <h2 className="text-2xl font-semibold tracking-tight text-foreground">
          {content.ownership.title}
        </h2>
        <p className="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">
          {content.ownership.intro}
        </p>
        <div className="mt-5 overflow-x-auto rounded-2xl border border-border bg-card">
          <table className="w-full min-w-[760px] border-collapse text-left text-sm">
            <thead className="bg-muted/40 text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                {content.ownership.headers.map((header) => (
                  <th key={header} className="border-b border-border px-4 py-3 font-semibold">
                    {header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {content.ownership.rows.map((row) => (
                <tr key={row.domain} className="align-top">
                  <td className="px-4 py-4 font-semibold text-foreground">{row.domain}</td>
                  <td className="px-4 py-4 leading-6 text-muted-foreground">{row.data}</td>
                  <td className="px-4 py-4 font-mono text-xs leading-6 text-foreground">
                    {row.authority}
                  </td>
                  <td className="px-4 py-4 leading-6 text-muted-foreground">{row.notes}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <aside className="mt-10 rounded-2xl border border-brand/30 bg-brand/5 p-5 sm:p-6">
        <h2 className="text-lg font-semibold text-foreground">{content.rule.title}</h2>
        <p className="mt-2 max-w-4xl text-sm leading-6 text-muted-foreground">
          {content.rule.text}
        </p>
      </aside>
    </article>
  );
}

function DomainField({
  label,
  text,
  children,
}: {
  label: string;
  text?: string;
  children?: ReactNode;
}) {
  return (
    <div>
      <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        {label}
      </h3>
      {text ? <p className="mt-1.5 text-sm leading-6 text-foreground/80">{text}</p> : null}
      {children ? <div className="mt-2">{children}</div> : null}
    </div>
  );
}
