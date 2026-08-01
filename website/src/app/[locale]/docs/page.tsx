import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { localeHref, resolveLocale, type Locale } from "@/i18n";

type PageContent = {
  metadata: { title: string; description: string };
  title: string;
  lead: string;
  microservices: {
    title: string;
    paragraphs: [string, string];
  };
  storage: {
    title: string;
    intro: string;
    detail: string;
  };
  kafka: {
    title: string;
    paragraphs: [string, string, string];
  };
  goZero: {
    title: string;
    text: string;
  };
  next: {
    title: string;
    cards: Array<{ label: string; href: string; desc: string }>;
  };
};

const pageContent = {
  en: {
    metadata: {
      title: "Docs — HiChat",
      description:
        "Understand how HiChat 2.0 is structured, why certain technology choices were made, and where to look when things go wrong.",
    },
    title: "Introduction",
    lead:
      "HiChat 2.0 is a reference implementation of a production-shaped IM system, not a toy chat demo. The codebase makes opinionated choices about storage, delivery guarantees, and service boundaries. This section explains the thinking behind those choices and where to read more.",
    microservices: {
      title: "Why microservices?",
      paragraphs: [
        "A monolith would be simpler to run, but it would obscure the interesting part: the seams between domains. The boundary between im and social exists because chat history and the social graph evolve at different rates, have different consistency requirements, and benefit from independent deployments. Once those seams are explicit, enforced by network calls rather than import paths, the tradeoffs become visible and teachable.",
        "The rule is strict: a service never reads another service's database tables. If im needs to know whether two users are friends before letting them chat, it asks social/rpc, not the social DB. That extra hop has a cost, and that cost is intentional.",
      ],
    },
    storage: {
      title: "Why two storage engines for messages?",
      intro:
        "Chat logs go to MongoDB; user accounts, friendships, and groups go to MySQL. This is not a default. It reflects different access patterns.",
      detail:
        "A chat log is write-once and queried by range: “give me messages in conversation X with seq > N.” That query is cheap on MongoDB because the document model fits a message naturally and the seq field gets a compound index on (conversationId, seq). Updating a message (recall, read state) is a targeted write by ID, also cheap. Social and account data, by contrast, is highly relational: friend lists, group membership, and role checks benefit from foreign keys and joins that MongoDB deliberately omits.",
    },
    kafka: {
      title: "Why Kafka instead of direct RPC for delivery?",
      paragraphs: [
        "When a message arrives at im/ws, it could synchronously call im/rpc to persist it and then push to the recipient. But synchronous persistence on the hot path adds latency proportional to MongoDB write time, and it creates a hard coupling: if the persistence call fails, should the sender get an error? The ack has already been sent.",
        "Kafka separates acknowledgement from persistence. The gateway ACKs the sender as soon as the message hits the topic, then task/mq persists it asynchronously. Delivery to the recipient also happens through the consumer rather than a synchronous push, which means the consumer can retry without bothering the sender and without the gateway needing to track outstanding deliveries per recipient.",
        "The trade-off is eventual consistency: there is a brief window where the sender has an ACK but the message is not yet in MongoDB. The system accepts that.",
      ],
    },
    goZero: {
      title: "Why go-zero?",
      text: "go-zero gives each service a code-generated skeleton from a contract file (.api for HTTP, .proto for gRPC). That contract is the single source of truth for routes, request shapes, and response shapes. The generated types and handler stubs never drift from it because you re-run goctl every time the contract changes. It also provides service discovery through etcd and built-in observability middleware, which would otherwise require custom wiring.",
    },
    next: {
      title: "Where to go from here",
      cards: [
        { label: "Architecture Overview", href: "/docs/architecture", desc: "See the runtime layers and real RPC, Kafka, and internal WebSocket dependencies." },
        { label: "Domain Map", href: "/docs/domains", desc: "Understand ownership, boundaries, processes, and authoritative stores across all six domains." },
        { label: "Core Data Flows", href: "/docs/data-flows", desc: "Trace messaging, social, feed, and media paths with their consistency guarantees." },
        { label: "Message Lifecycle", href: "/docs/message-lifecycle", desc: "Follow a single chat message from the sender's keyboard to the recipient's screen." },
        { label: "Realtime Gateway", href: "/docs/realtime-gateway", desc: "How the WebSocket server manages connections, heartbeats, and delivery guarantees." },
        { label: "Extending the System", href: "/docs/extending", desc: "Add a new service, endpoint, consumer, or cron job without breaking existing contracts." },
        { label: "Quick Start", href: "/quick-start", desc: "Get the full stack running locally in under a minute." },
      ],
    },
  },
  zh: {
    metadata: {
      title: "文档 — HiChat",
      description: "了解 HiChat 2.0 的系统架构、关键技术选型及故障排查入口。",
    },
    title: "简介",
    lead:
      "HiChat 2.0 是面向生产形态的即时通信系统参考实现，而非简单的聊天演示。项目对存储方案、交付保证和服务边界作出了明确取舍。本节将解释这些决策背后的考量，并指引你继续深入阅读。",
    microservices: {
      title: "为什么采用微服务？",
      paragraphs: [
        "单体架构的运行方式更简单，却会掩盖最值得探讨的部分：业务领域之间的边界。im 与 social 之所以分离，是因为聊天记录和社交关系图的演进速度不同、一致性要求不同，并且适合独立部署。当这些边界由网络调用而非代码导入路径明确约束后，各项权衡便清晰可见，也更便于理解。",
        "系统遵循一条严格规则：任何服务都不能读取其他服务的数据库表。如果 im 在允许两名用户聊天前需要确认好友关系，它必须调用 social/rpc，而不能直接查询 social 数据库。额外的网络调用会带来成本，这是维护服务边界的有意取舍。",
      ],
    },
    storage: {
      title: "为什么消息使用两种存储引擎？",
      intro: "聊天记录存入 MongoDB，用户账号、好友关系和群组数据存入 MySQL。这并非默认选择，而是由不同的数据访问模式决定的。",
      detail:
        "聊天记录写入后通常不再改变，主要按范围查询，例如“返回会话 X 中 seq > N 的消息”。MongoDB 的文档模型能够自然表达一条消息，并可在 (conversationId, seq) 上建立复合索引，因此这类查询成本较低。消息更新（撤回、已读状态）也可按 ID 精确执行。相比之下，好友列表、群成员关系和角色校验等社交与账号数据具有高度关系性，更适合利用外键和关联查询，而这些能力正是 MongoDB 有意不强调的。",
    },
    kafka: {
      title: "为什么使用 Kafka 而不是直接 RPC 投递？",
      paragraphs: [
        "消息到达 im/ws 后，网关本可以同步调用 im/rpc 完成持久化，再推送给接收方。但在核心链路中同步持久化会引入与 MongoDB 写入耗时成正比的延迟，也会形成强耦合：如果持久化调用失败，是否应向发送方返回错误？此时 ACK 可能已经发出。",
        "Kafka 将确认与持久化解耦。消息进入主题后，网关即可向发送方返回 ACK，随后由 task/mq 异步持久化。接收方投递同样由消费者完成，而不是同步推送，因此消费者可以独立重试，不必打扰发送方，网关也无须逐一跟踪每个接收方尚未完成的投递。",
        "这种设计的代价是最终一致性：发送方收到 ACK 后，消息可能在短时间内尚未写入 MongoDB。系统明确接受这一时间窗口。",
      ],
    },
    goZero: {
      title: "为什么选择 go-zero？",
      text: "go-zero 根据契约文件为每个服务生成代码骨架：HTTP 使用 .api，gRPC 使用 .proto。契约是路由、请求结构和响应结构的唯一事实来源。每次契约变更后重新运行 goctl，生成的类型与 handler 桩代码便不会偏离契约。go-zero 还提供基于 etcd 的服务发现和内置可观测性中间件，避免项目自行完成这些基础设施接线。",
    },
    next: {
      title: "接下来阅读什么",
      cards: [
        { label: "架构总览", href: "/docs/architecture", desc: "查看运行时分层，以及真实的 RPC、Kafka 和内部 WebSocket 依赖。" },
        { label: "领域模块", href: "/docs/domains", desc: "理解六个领域的职责、边界、运行进程与权威数据存储。" },
        { label: "核心数据流", href: "/docs/data-flows", desc: "追踪消息、社交、动态和音视频链路及其一致性保证。" },
        { label: "消息生命周期", href: "/docs/message-lifecycle", desc: "跟踪一条聊天消息从发送方键盘到接收方屏幕的完整过程。" },
        { label: "实时通信网关", href: "/docs/realtime-gateway", desc: "了解 WebSocket 服务如何管理连接、心跳和投递保证。" },
        { label: "扩展系统", href: "/docs/extending", desc: "在不破坏现有契约的前提下新增服务、接口、消费者或定时任务。" },
        { label: "快速开始", href: "/quick-start", desc: "在一分钟内于本地启动完整技术栈。" },
      ],
    },
  },
} satisfies Record<Locale, PageContent>;

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  return pageContent[locale].metadata;
}

export default async function DocsIntroPage({ params }: Props) {
  const locale = resolveLocale((await params).locale);
  const content = pageContent[locale];

  return (
    <article className="prose-custom">
      <h1>{content.title}</h1>
      <p className="lead">{content.lead}</p>

      <h2>{content.microservices.title}</h2>
      <p>{renderTerms(content.microservices.paragraphs[0], ["im", "social"])}</p>
      <p>{renderTerms(content.microservices.paragraphs[1], ["im", "social/rpc"])}</p>

      <h2>{content.storage.title}</h2>
      <p>{content.storage.intro}</p>
      <p>{renderTerms(content.storage.detail, ["seq > N", "(conversationId, seq)", "ID"])}</p>

      <h2>{content.kafka.title}</h2>
      <p>{renderTerms(content.kafka.paragraphs[0], ["im/ws", "im/rpc", "MongoDB", "ack"])}</p>
      <p>{renderTerms(content.kafka.paragraphs[1], ["ACK", "task/mq"])}</p>
      <p>{renderTerms(content.kafka.paragraphs[2], ["ACK", "MongoDB"])}</p>

      <h2>{content.goZero.title}</h2>
      <p>{renderTerms(content.goZero.text, ["code-generated skeleton", ".api", "HTTP", ".proto", "gRPC", "goctl", "etcd", "代码骨架"])}</p>

      <h2>{content.next.title}</h2>
      <div className="not-prose mt-6 grid gap-3 sm:grid-cols-2">
        {content.next.cards.map((card) => (
          <Link
            key={card.href}
            href={localeHref(locale, card.href)}
            className="group flex items-start justify-between gap-3 rounded-xl border border-border bg-card p-4 transition-colors hover:border-brand/50"
          >
            <div>
              <p className="text-sm font-medium text-foreground">{card.label}</p>
              <p className="mt-1 text-xs text-muted-foreground">{card.desc}</p>
            </div>
            <ArrowRight className="mt-0.5 size-4 shrink-0 text-muted-foreground transition-colors group-hover:text-brand" />
          </Link>
        ))}
      </div>
    </article>
  );
}

function renderTerms(text: string, terms: string[]) {
  const pattern = new RegExp(`(${terms.map(escapeRegExp).join("|")})`, "g");
  return text.split(pattern).map((part, index) =>
    terms.includes(part) ? <code key={`${part}-${index}`}>{part}</code> : part,
  );
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
