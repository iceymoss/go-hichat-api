import type { Metadata } from "next";
import { FlowDiagram } from "@/components/docs/FlowDiagrams";
import { resolveLocale, type Locale } from "@/i18n";

type Marker = "sync" | "async" | "store";

type FlowStep = {
  title: string;
  detail: string;
  markers: Marker[];
};

type Flow = {
  number: string;
  eyebrow: string;
  title: string;
  summary: string;
  accent: "cyan" | "violet" | "amber" | "rose" | "emerald" | "blue" | "lime";
  steps: FlowStep[];
  consistency: string;
};

type PageContent = {
  metadata: { title: string; description: string };
  kicker: string;
  title: string;
  lead: string;
  legend: Record<Marker, string>;
  legendNote: string;
  flowLabel: string;
  consistencyLabel: string;
  flows: Flow[];
  boundaryTitle: string;
  boundaries: { title: string; detail: string }[];
};

const content = {
  en: {
    metadata: {
      title: "Core Data Flows - HiChat Docs",
      description:
        "A bilingual map of HiChat message, notification, relation, trend, and realtime media data flows and their consistency boundaries.",
    },
    kicker: "SYSTEM MAP / 07 FLOWS",
    title: "Core data flows",
    lead:
      "Follow each write from its synchronous entry point through queues, owning storage, and best-effort realtime delivery. The labels describe implementation boundaries, not stronger guarantees than the system currently provides.",
    legend: { sync: "Synchronous", async: "Asynchronous", store: "Storage" },
    legendNote: "A step can cross more than one boundary.",
    flowLabel: "Flow",
    consistencyLabel: "Consistency boundary",
    flows: [
      {
        number: "01",
        eyebrow: "CHAT / WRITE PATH",
        title: "Send a message",
        summary: "The WebSocket accepts the command; durable history appears later in the MQ consumer.",
        accent: "cyan",
        steps: [
          { title: "Client", detail: "Sends a chat frame over its authenticated connection.", markers: ["sync"] },
          { title: "IM WebSocket", detail: "The chat handler publishes the command to msgChatTransfer.", markers: ["sync", "async"] },
          { title: "Task MQ", detail: "Consumes msgChatTransfer and applies the chat write path.", markers: ["async"] },
          { title: "MongoDB", detail: "Persists the chat log before downstream online fan-out.", markers: ["store"] },
          { title: "IM WebSocket", detail: "Pushes to active recipient connections and echoes the persisted message ID to the sender.", markers: ["async"] },
          { title: "Online / offline", detail: "Online clients receive the frame. Offline clients later pull MongoDB-backed history by conversation position.", markers: ["async", "store"] },
        ],
        consistency:
           "Queue publication and MongoDB persistence are separate phases. Delivery is eventually consistent; the sender's persisted-ID echo confirms the stored record, while recipient push remains best-effort. Offline recovery comes from persisted history. The current chat write has no storage-level business deduplication key, so a retry after partial work can create duplicates.",
      },
      {
        number: "02",
        eyebrow: "CHAT / RECEIPT",
        title: "Mark messages read",
        summary: "Read state travels independently from the original message.",
        accent: "violet",
        steps: [
          { title: "Client", detail: "Emits chat.markChat with the conversation and a batch of MongoDB message IDs.", markers: ["sync"] },
          { title: "IM WebSocket", detail: "Validates the frame and publishes msgReadTransfer.", markers: ["sync", "async"] },
          { title: "Task MQ", detail: "Consumes the read event and updates message read records.", markers: ["async"] },
          { title: "MongoDB", detail: "Stores the resulting read state and conversation-level read progress.", markers: ["store"] },
          { title: "IM WebSocket", detail: "Pushes a read receipt to the relevant online sender when policy allows it.", markers: ["async"] },
        ],
        consistency:
          "Read marking is asynchronous and may lag the client action. Persisted read state is the recoverable source; the realtime receipt is best-effort and an offline sender observes the state through a later pull rather than a guaranteed push.",
      },
      {
        number: "03",
        eyebrow: "CHAT / CONTROL EVENT",
        title: "Recall a message",
        summary: "The authoritative state change happens before the recall event is fanned out.",
        accent: "amber",
        steps: [
          { title: "IM API", detail: "Authenticates the caller and sends the recall command to IM RPC.", markers: ["sync"] },
          { title: "IM RPC", detail: "Checks the message, actor, chat context, and recall window.", markers: ["sync"] },
          { title: "MongoDB", detail: "Conditionally changes only a normal message to recalled; repeated or racing recalls do not change it again.", markers: ["sync", "store"] },
          { title: "msgRecallTransfer", detail: "After a successful state change, IM API publishes a dedicated recall event.", markers: ["async"] },
          { title: "Task MQ → IM WebSocket", detail: "Translates the event into a recall control frame and pushes it to currently connected conversation participants.", markers: ["async"] },
        ],
        consistency:
          "The conditional MongoDB update makes the state transition idempotent, but the database update and Kafka publication are not one atomic transaction. Realtime recall frames are best-effort; clients must reconcile against persisted message status when loading history.",
      },
      {
        number: "04",
        eyebrow: "SOCIAL / RELIABLE NOTIFICATION",
        title: "Deliver a social request notification",
        summary: "A transactional outbox protects the hand-off from social state to eventual notification storage.",
        accent: "rose",
        steps: [
          { title: "Social RPC transaction", detail: "Writes the request or decision and social_notification_outbox in the same MySQL transaction.", markers: ["sync", "store"] },
          { title: "Notification relay", detail: "Polls due outbox rows, retries failed publication, and marks successfully published rows.", markers: ["async", "store"] },
          { title: "social.request.notification.v1", detail: "Carries the notification event through Kafka to the Task consumer group.", markers: ["async"] },
          { title: "Task MQ → IM RPC", detail: "Calls the owning IM service to create a notification.", markers: ["async", "sync"] },
          { title: "IM MySQL notifications", detail: "Persists the user-visible notification before attempting realtime delivery.", markers: ["store"] },
          { title: "push.notify", detail: "Sends the new notification through IM WebSocket when the receiver is online.", markers: ["async"] },
        ],
        consistency:
          "Social state and its outbox row commit atomically. Relay and consumer processing are at-least-once in shape, so identifiers and the IM create operation handle duplicates. Notification storage is retried before offset progress; push.notify remains best-effort and is not proof that the client rendered the notification.",
      },
      {
        number: "05",
        eyebrow: "SOCIAL / AUTHORIZATION STATE",
        title: "Apply a relation change",
        summary: "Ordered events move durable social decisions into IM-side cache and conversation state.",
        accent: "emerald",
        steps: [
          { title: "Social MySQL", detail: "Commits the relation mutation together with a relation_outbox row.", markers: ["sync", "store"] },
          { title: "Relation relay", detail: "Publishes the row to relationChangeTransfer using an aggregate key and the outbox ID as a version where applicable.", markers: ["async", "store"] },
          { title: "Keyed Kafka stream", detail: "Preserves order within a key/partition, not as a global order across all relations.", markers: ["async"] },
          { title: "Task MQ", detail: "Applies reliable group conversation state first, including removal freeze or rejoin recovery.", markers: ["async", "store"] },
          { title: "Redis relation cache", detail: "Applies version-gated membership changes or invalidation as best-effort cache maintenance; TTL and read-through repair misses.", markers: ["async", "store"] },
          { title: "push.relation", detail: "Notifies affected online clients to invalidate or disable the conversation UI.", markers: ["async"] },
        ],
        consistency:
          "The social mutation and relation_outbox are atomic. Downstream state converges asynchronously. Version gates protect applicable group events from stale replay, while Redis is still a repairable cache. Conversation freeze is retried by the consumer; websocket invalidation is best-effort and offline clients rely on refreshed server state.",
      },
      {
        number: "06",
        eyebrow: "TREND / INTERACTION",
        title: "Notify a trend interaction",
        summary: "The durable inbox entry is owned by Trend; realtime delivery is only an acceleration path.",
        accent: "blue",
        steps: [
          { title: "Trend RPC", detail: "Processes a like, comment, or other interaction and creates the recipient message.", markers: ["sync"] },
          { title: "MySQL trend_message", detail: "Stores the durable trend notification record in the Trend service database.", markers: ["sync", "store"] },
          { title: "trendNotifyTransfer", detail: "The API publishes each created message to the dedicated Kafka topic after the RPC returns.", markers: ["async"] },
          { title: "Task MQ", detail: "Consumes the event and translates it without another database write.", markers: ["async"] },
          { title: "push.trend", detail: "Pushes to the receiver through IM WebSocket if connected.", markers: ["async"] },
        ],
        consistency:
          "trend_message is the durable source of truth. Its write and topic publication are separate operations, so no atomic database-to-Kafka guarantee is implied. push.trend is best-effort; an offline client recovers by pulling the persisted trend message list.",
      },
      {
        number: "07",
        eyebrow: "STREAMING / CONTROL + MEDIA",
        title: "Run an audio or video call",
        summary: "Control signalling and media transport share a service, but they have different paths and durability.",
        accent: "lime",
        steps: [
          { title: "Streaming WebSocket", detail: "Authenticates participants and carries room, call, SDP, ICE, and media-control signalling.", markers: ["sync"] },
          { title: "IM WebSocket fallback", detail: "If the target is not on streaming WebSocket, control signalling falls back through push.call; offline delivery can still be missed.", markers: ["async"] },
          { title: "P2P media", detail: "Eligible direct calls exchange media between browser peer connections after signalling.", markers: ["sync"] },
          { title: "In-process SFU", detail: "Room calls can publish WebRTC tracks to the process-local SFU, which forwards selected tracks to subscribers.", markers: ["sync"] },
          { title: "In-memory room state", detail: "Rooms, participants, peers, publications, and active-speaker state live in the streaming process.", markers: ["store"] },
        ],
        consistency:
          "Call control and room state are ephemeral, not a durable workflow. push.call is a best-effort fallback, not an offline queue. Process restart or node loss can discard active in-memory rooms and peer state, requiring clients to reconnect or recreate the call; the in-process SFU does not by itself provide cross-node room continuity.",
      },
    ],
    boundaryTitle: "Read the boundaries, not just the arrows",
    boundaries: [
      { title: "Durable does not mean delivered", detail: "MongoDB or MySQL can retain state even when an online WebSocket push fails." },
      { title: "Outbox is scoped", detail: "Social request and relation mutations use transactional outboxes; do not infer the same atomic hand-off for chat, recall, or trend publication." },
      { title: "Realtime is a projection", detail: "Clients reconcile durable lists and history after reconnect; push frames reduce latency but are not universal delivery receipts." },
    ],
  },
  zh: {
    metadata: {
      title: "核心链路与数据流 - HiChat 文档",
      description: "双语梳理 HiChat 消息、通知、关系、动态及实时音视频的数据流与一致性边界。",
    },
    kicker: "系统地图 / 07 条链路",
    title: "核心链路与数据流",
    lead:
      "从同步入口开始，沿队列、归属存储和尽力而为的实时推送追踪每一次写入。页面标记描述的是当前实现边界，不代表系统提供了更强保证。",
    legend: { sync: "同步", async: "异步", store: "存储" },
    legendNote: "一个步骤可能同时跨越多种边界。",
    flowLabel: "链路",
    consistencyLabel: "一致性边界",
    flows: [
      {
        number: "01",
        eyebrow: "聊天 / 写链路",
        title: "发送消息",
        summary: "WebSocket 接收发送命令，持久化聊天记录随后由 MQ 消费者产生。",
        accent: "cyan",
        steps: [
          { title: "客户端", detail: "通过已认证连接发送聊天 frame。", markers: ["sync"] },
          { title: "IM WebSocket", detail: "聊天 handler 将命令发布到 msgChatTransfer。", markers: ["sync", "async"] },
          { title: "Task MQ", detail: "消费 msgChatTransfer，执行聊天写入链路。", markers: ["async"] },
          { title: "MongoDB", detail: "先持久化聊天记录，再进入后续在线扇出。", markers: ["store"] },
          { title: "IM WebSocket", detail: "推送到接收方活动连接，并向发送方回响持久化后的消息 ID。", markers: ["async"] },
          { title: "在线 / 离线", detail: "在线端收到 frame；离线端随后按会话位置拉取 MongoDB 中的历史消息。", markers: ["async", "store"] },
        ],
        consistency:
          "发布队列和写入 MongoDB 是两个阶段，消息投递为最终一致；发送方收到持久化消息 ID 回响后可确认该记录已落库，接收方在线推送仍是尽力而为。离线恢复依赖已落库历史。当前聊天写入没有存储层业务去重键，部分步骤完成后的重试可能产生重复消息。",
      },
      {
        number: "02",
        eyebrow: "聊天 / 回执",
        title: "标记已读",
        summary: "已读状态通过独立于原消息的事件链路传播。",
        accent: "violet",
        steps: [
          { title: "客户端", detail: "发送 chat.markChat，携带会话及一批 MongoDB 消息 ID。", markers: ["sync"] },
          { title: "IM WebSocket", detail: "校验 frame 并发布 msgReadTransfer。", markers: ["sync", "async"] },
          { title: "Task MQ", detail: "消费已读事件并更新消息的已读记录。", markers: ["async"] },
          { title: "MongoDB", detail: "保存已读结果及会话级已读进度。", markers: ["store"] },
          { title: "IM WebSocket", detail: "策略允许时，向相关的在线发送方推送已读回执。", markers: ["async"] },
        ],
        consistency:
          "已读标记异步完成，可能滞后于客户端操作。已持久化状态是可恢复依据；实时回执为尽力推送，发送方离线时通过后续拉取观察状态，而非依赖保证到达的推送。",
      },
      {
        number: "03",
        eyebrow: "聊天 / 控制事件",
        title: "撤回消息",
        summary: "先完成权威状态变更，再扇出撤回事件。",
        accent: "amber",
        steps: [
          { title: "IM API", detail: "认证调用方，并向 IM RPC 发起撤回命令。", markers: ["sync"] },
          { title: "IM RPC", detail: "校验消息、操作者、会话上下文和撤回时间窗。", markers: ["sync"] },
          { title: "MongoDB", detail: "仅将正常态消息条件更新为已撤回；重复或并发撤回不会再次改动。", markers: ["sync", "store"] },
          { title: "msgRecallTransfer", detail: "状态变更成功后，IM API 发布独立撤回事件。", markers: ["async"] },
          { title: "Task MQ → IM WebSocket", detail: "把事件转换为撤回控制帧，推送给会话中当前在线的参与端。", markers: ["async"] },
        ],
        consistency:
          "MongoDB 条件更新使状态迁移具备幂等性，但数据库更新与 Kafka 发布并非同一原子事务。实时撤回帧为尽力推送；客户端加载历史时仍需以持久化消息状态进行校准。",
      },
      {
        number: "04",
        eyebrow: "社交 / 可靠通知",
        title: "投递社交请求通知",
        summary: "事务发件箱保护社交状态到最终通知存储之间的交接。",
        accent: "rose",
        steps: [
          { title: "Social RPC 事务", detail: "在同一 MySQL 事务内写入请求或处理结果及 social_notification_outbox。", markers: ["sync", "store"] },
          { title: "通知 relay", detail: "轮询到期 outbox 行，重试发布失败项，并标记已成功发布记录。", markers: ["async", "store"] },
          { title: "social.request.notification.v1", detail: "通过 Kafka 将通知事件送到 Task 消费组。", markers: ["async"] },
          { title: "Task MQ → IM RPC", detail: "调用数据归属方 IM 服务创建通知。", markers: ["async", "sync"] },
          { title: "IM MySQL notifications", detail: "先持久化用户可见通知，再尝试实时投递。", markers: ["store"] },
          { title: "push.notify", detail: "接收方在线时，通过 IM WebSocket 发送新通知。", markers: ["async"] },
        ],
        consistency:
          "社交状态与 outbox 行原子提交。relay 和消费者呈至少一次处理形态，因此由标识符及 IM 创建操作处理重复。offset 前的通知落库会持续重试；push.notify 仍是尽力推送，也不证明客户端已经展示通知。",
      },
      {
        number: "05",
        eyebrow: "社交 / 鉴权状态",
        title: "应用关系变更",
        summary: "有序事件把持久化社交决策传递到 IM 侧缓存和会话状态。",
        accent: "emerald",
        steps: [
          { title: "Social MySQL", detail: "在关系变更事务中同时提交 relation_outbox 行。", markers: ["sync", "store"] },
          { title: "关系 relay", detail: "按聚合键发布到 relationChangeTransfer，并在适用场景用 outbox ID 作为版本。", markers: ["async", "store"] },
          { title: "按键 Kafka 流", detail: "保证同一 key/partition 内顺序，不代表所有关系事件全局有序。", markers: ["async"] },
          { title: "Task MQ", detail: "先应用需可靠完成的群会话状态，包括移出后的会话冻结或重新入群恢复。", markers: ["async", "store"] },
          { title: "Redis 关系缓存", detail: "通过版本门更新成员关系或尽力失效缓存；失败由 TTL 和读穿透修复。", markers: ["async", "store"] },
          { title: "push.relation", detail: "通知受影响的在线客户端失效会话或禁用输入界面。", markers: ["async"] },
        ],
        consistency:
          "社交关系变更与 relation_outbox 原子提交，下游状态异步收敛。版本门保护适用的群事件不被旧事件覆盖，但 Redis 仍是可修复缓存。消费者会重试会话冻结；WebSocket 失效帧为尽力推送，离线客户端依赖刷新后的服务端状态。",
      },
      {
        number: "06",
        eyebrow: "动态 / 互动",
        title: "通知动态互动",
        summary: "持久化收件记录归 Trend 所有，实时投递只是加速路径。",
        accent: "blue",
        steps: [
          { title: "Trend RPC", detail: "处理点赞、评论等互动，并创建接收方消息。", markers: ["sync"] },
          { title: "MySQL trend_message", detail: "在 Trend 服务数据库中保存持久化动态通知。", markers: ["sync", "store"] },
          { title: "trendNotifyTransfer", detail: "RPC 返回后，API 把各条已创建消息发布到独立 Kafka topic。", markers: ["async"] },
          { title: "Task MQ", detail: "消费事件并完成协议转换，不再次写数据库。", markers: ["async"] },
          { title: "push.trend", detail: "接收方已连接时，通过 IM WebSocket 单推。", markers: ["async"] },
        ],
        consistency:
          "trend_message 是持久化事实来源。其写入与 topic 发布是两个操作，不应推断数据库到 Kafka 具备原子保证。push.trend 为尽力推送；离线客户端通过拉取已存储的动态消息列表恢复。",
      },
      {
        number: "07",
        eyebrow: "STREAMING / 控制 + 媒体",
        title: "进行音视频通话",
        summary: "控制信令与媒体传输共用服务，但路径和持久性不同。",
        accent: "lime",
        steps: [
          { title: "Streaming WebSocket", detail: "认证参与者，承载房间、通话、SDP、ICE 和媒体控制信令。", markers: ["sync"] },
          { title: "IM WebSocket fallback", detail: "目标不在 streaming WebSocket 时，控制信令回退到 push.call；离线时仍可能错过。", markers: ["async"] },
          { title: "P2P 媒体", detail: "满足条件的直接通话在信令协商后，由浏览器 PeerConnection 点对点传输媒体。", markers: ["sync"] },
          { title: "进程内 SFU", detail: "房间通话可将 WebRTC track 发布到本进程 SFU，再选择性转发给订阅者。", markers: ["sync"] },
          { title: "内存房间状态", detail: "房间、参与者、peer、发布轨和活跃说话人状态均位于 streaming 进程内。", markers: ["store"] },
        ],
        consistency:
          "通话控制和房间状态是瞬时状态，不是持久化工作流。push.call 是尽力回退，不是离线队列。进程重启或节点丢失会清除活动房间及 peer 状态，客户端需要重连或重建通话；进程内 SFU 本身不提供跨节点房间连续性。",
      },
    ],
    boundaryTitle: "不要只看箭头，也要看边界",
    boundaries: [
      { title: "已持久化不等于已投递", detail: "即使在线 WebSocket 推送失败，MongoDB 或 MySQL 仍可保留业务状态。" },
      { title: "Outbox 只覆盖特定链路", detail: "社交请求和关系变更使用事务发件箱；不能据此推断聊天、撤回或动态发布也具有同样的原子交接。" },
      { title: "实时推送是状态投影", detail: "客户端重连后通过持久化列表和历史校准；push frame 用于降低延迟，不是通用投递回执。" },
    ],
  },
} satisfies Record<Locale, PageContent>;

type PageProps = { params: Promise<{ locale: string }> };

const accentClasses: Record<Flow["accent"], { border: string; number: string; dot: string }> = {
  cyan: { border: "border-cyan-400/25", number: "text-cyan-300", dot: "bg-cyan-300" },
  violet: { border: "border-violet-400/25", number: "text-violet-300", dot: "bg-violet-300" },
  amber: { border: "border-amber-400/25", number: "text-amber-300", dot: "bg-amber-300" },
  rose: { border: "border-rose-400/25", number: "text-rose-300", dot: "bg-rose-300" },
  emerald: { border: "border-emerald-400/25", number: "text-emerald-300", dot: "bg-emerald-300" },
  blue: { border: "border-blue-400/25", number: "text-blue-300", dot: "bg-blue-300" },
  lime: { border: "border-lime-400/25", number: "text-lime-300", dot: "bg-lime-300" },
};

const markerClasses: Record<Marker, string> = {
  sync: "border-sky-400/25 bg-sky-400/10 text-sky-300",
  async: "border-fuchsia-400/25 bg-fuchsia-400/10 text-fuchsia-300",
  store: "border-amber-400/25 bg-amber-400/10 text-amber-300",
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  return content[locale].metadata;
}

export default async function DataFlowsPage({ params }: PageProps) {
  const locale = resolveLocale((await params).locale);
  const page = content[locale];

  return (
    <article className="not-prose mx-auto max-w-6xl pb-16 text-foreground">
      <header className="relative overflow-hidden rounded-3xl border border-white/10 bg-[#090d16] px-5 py-10 shadow-2xl shadow-black/20 sm:px-9 sm:py-14">
        <div className="pointer-events-none absolute -right-16 -top-24 h-72 w-72 rounded-full bg-cyan-400/10 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-24 left-1/4 h-64 w-64 rounded-full bg-violet-500/10 blur-3xl" />
        <div className="relative">
          <p className="font-mono text-xs font-semibold tracking-[0.24em] text-cyan-300">{page.kicker}</p>
          <h1 className="mt-4 max-w-3xl text-4xl font-semibold tracking-tight text-white sm:text-5xl lg:text-6xl">
            {page.title}
          </h1>
          <p className="mt-5 max-w-3xl text-base leading-7 text-slate-300 sm:text-lg">{page.lead}</p>
          <div className="mt-8 flex flex-wrap items-center gap-2">
            {(Object.keys(page.legend) as Marker[]).map((marker) => (
              <MarkerBadge key={marker} marker={marker} label={page.legend[marker]} />
            ))}
            <span className="ml-1 text-xs text-slate-500">{page.legendNote}</span>
          </div>
        </div>
      </header>

      <div className="mt-8 grid gap-6">
        {page.flows.map((flow) => (
          <FlowCard
            key={flow.number}
            flow={flow}
            legend={page.legend}
             flowLabel={page.flowLabel}
             consistencyLabel={page.consistencyLabel}
             locale={locale}
           />
        ))}
      </div>

      <section className="mt-10 rounded-3xl border border-border bg-card/60 p-5 sm:p-8">
        <div className="flex items-center gap-3">
          <span className="h-px w-8 bg-cyan-400" />
          <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">{page.boundaryTitle}</h2>
        </div>
        <div className="mt-6 grid gap-4 md:grid-cols-3">
          {page.boundaries.map((boundary, index) => (
            <div key={boundary.title} className="rounded-2xl border border-border bg-background/60 p-5">
              <span className="font-mono text-xs text-muted-foreground">0{index + 1}</span>
              <h3 className="mt-3 font-semibold">{boundary.title}</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">{boundary.detail}</p>
            </div>
          ))}
        </div>
      </section>
    </article>
  );
}

function FlowCard({
  flow,
  legend,
  flowLabel,
  consistencyLabel,
  locale,
}: {
  flow: Flow;
  legend: Record<Marker, string>;
  flowLabel: string;
  consistencyLabel: string;
  locale: Locale;
}) {
  const accent = accentClasses[flow.accent];

  return (
    <section className={`overflow-hidden rounded-3xl border bg-card/70 ${accent.border}`}>
      <div className="grid gap-5 border-b border-border px-5 py-6 sm:px-7 lg:grid-cols-[7rem_1fr] lg:py-7">
        <div>
          <p className="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{flowLabel}</p>
          <p className={`mt-1 font-mono text-4xl font-light ${accent.number}`}>{flow.number}</p>
        </div>
        <div>
          <p className={`font-mono text-[11px] font-semibold tracking-[0.18em] ${accent.number}`}>{flow.eyebrow}</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-tight sm:text-3xl">{flow.title}</h2>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground sm:text-base">{flow.summary}</p>
        </div>
      </div>

      <div className="px-5 py-6 sm:px-7">
        <FlowDiagram flowNumber={flow.number} locale={locale} />

        <ol className="mt-5 grid gap-3 lg:grid-cols-2">
          {flow.steps.map((step, index) => (
            <li key={`${flow.number}-${step.title}`} className="relative flex gap-4 rounded-2xl border border-border bg-background/55 p-4 sm:p-5">
              <div className="flex shrink-0 flex-col items-center">
                <span className={`mt-1 h-2.5 w-2.5 rounded-full ${accent.dot}`} />
                {index < flow.steps.length - 1 ? <span className="mt-2 h-full w-px bg-border lg:hidden" /> : null}
              </div>
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-mono text-[10px] text-muted-foreground">{String(index + 1).padStart(2, "0")}</span>
                  <h3 className="text-sm font-semibold sm:text-base">{step.title}</h3>
                </div>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">{step.detail}</p>
                <div className="mt-3 flex flex-wrap gap-1.5">
                  {step.markers.map((marker) => (
                    <MarkerBadge key={marker} marker={marker} label={legend[marker]} />
                  ))}
                </div>
              </div>
            </li>
          ))}
        </ol>

        <div className={`mt-5 rounded-2xl border ${accent.border} bg-background/35 p-4 sm:p-5`}>
          <p className={`font-mono text-[10px] font-semibold uppercase tracking-[0.18em] ${accent.number}`}>
            {consistencyLabel}
          </p>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{flow.consistency}</p>
        </div>
      </div>
    </section>
  );
}

function MarkerBadge({ marker, label }: { marker: Marker; label: string }) {
  return (
    <span className={`inline-flex rounded-full border px-2.5 py-1 font-mono text-[10px] font-semibold uppercase tracking-wider ${markerClasses[marker]}`}>
      {label}
    </span>
  );
}
