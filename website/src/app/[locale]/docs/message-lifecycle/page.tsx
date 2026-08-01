import type { Metadata } from "next";
import Link from "next/link";
import { localeHref, resolveLocale, type Locale } from "@/i18n";

type Failure = { point: string; effect: string; recovery: string };

type PageContent = {
  metadata: { title: string; description: string };
  title: string;
  lead: string;
  happyPath: string;
  ack: {
    title: string;
    intro: string;
    modesIntro: string;
    modes: [string, string, string];
    deduplication: string;
  };
  kafka: { title: string; paragraphs: [string, string] };
  persistence: {
    title: string;
    intro: string;
    listIntro: string;
    steps: [string, string, string];
    idempotency: string;
  };
  delivery: { title: string; paragraphs: [string, string] };
  readReceipt: { title: string; text: string };
  recall: { title: string; text: string };
  failures: {
    title: string;
    headers: [string, string, string];
    rows: Failure[];
  };
  next: string;
};

const pageContent = {
  en: {
    metadata: {
      title: "Message Lifecycle — HiChat Docs",
      description:
        "The complete path of a chat message from sender to recipient, including ACK, Kafka, persistence, and offline delivery.",
    },
    title: "Message Lifecycle",
    lead:
      "A single chat message crosses four distinct systems before it reaches the recipient. Understanding that path, and the guarantees at each hand-off, is the key to reasoning about delivery failures, latency, and eventual consistency in HiChat.",
    happyPath: "The happy path",
    ack: {
      title: "Step 1 — ACK negotiation (im/ws)",
      intro:
        "The sender types a message and the web client publishes it over the existing WebSocket connection as a JSON frame with method: \"chat.user\" (single chat) or \"chat.group\" (group chat). At this point nothing has been persisted; the frame is in flight. im/ws runs the WebSocket server. When the frame arrives, the server's ACK mode determines what happens before the message is passed to any handler.",
      modesIntro: "The server supports three ACK modes, set at startup via WithAck:",
      modes: [
        "NoAck — the frame goes directly to the handler. Fastest, no delivery guarantee.",
        "OnlyAck — the server immediately sends an ACK frame back with ackSeq = msgSeq + 1, then passes the message to the handler. Two round-trips, server-initiated confirmation.",
        "RigorAck — full three-way handshake. The server sends ACK seq 1; the client must reply with ACK seq 2 before the message is passed to the handler. If the client's reply doesn't arrive within ackTimeout, the message is abandoned. Three round-trips, mutual confirmation.",
      ],
      deduplication:
        "Regardless of mode, every message has a UUID Id. The server deduplicates by Id inside the per-connection readMessageSeq map: if the same ID arrives twice with an equal or lower AckSeq, the second copy is silently dropped.",
    },
    kafka: {
      title: "Step 2 — Kafka publish (im/ws → Kafka)",
      paragraphs: [
         "Once ACK is complete, the chat.user handler publishes the message to the Kafka topic msgChatTransfer. The handler returns after publication succeeds; it does not wait for MongoDB. The transport ACK confirms the WebSocket exchange, not durable message storage.",
        "This is the system's intentional consistency trade-off: if the broker accepts the message but the consumer crashes before writing to MongoDB, the message can be replayed from Kafka (within the topic's retention window) but will not appear in history queries until the consumer catches up.",
      ],
    },
    persistence: {
      title: "Step 3 — Persistence (task/mq)",
      intro:
         "task/mq runs a consumer group against msgChatTransfer. The consumer that handles this topic is MsgChatTransfer, which extends BaseMsgChatTransfer.",
      listIntro: "The consumer does three things in order:",
      steps: [
         "Inserts the message document into the MongoDB chat_logs collection.",
         "Updates MongoDB conversation summaries and each user's conversation state, including unread and @mention projections.",
         "Uses an internal system WebSocket connection to ask im/ws to push the message to the recipient's active connection(s).",
      ],
      idempotency:
         "The current chat write has no storage-level business deduplication key. If the consumer retries after the message insert but before later steps complete, a duplicate chat document can be created.",
    },
    delivery: {
      title: "Step 4 — Push or offline storage",
      paragraphs: [
         "If the recipient has an active WebSocket connection, im/ws writes the message frame directly to every active connection. If not, the message remains in MongoDB and the client later retrieves conversation history through IM API/RPC.",
         "After persistence, task/mq also echoes the real MongoDB message ID to the sender. This lets the client replace its local temporary ID and distinguishes transport ACK from persistence completion.",
      ],
    },
    readReceipt: {
      title: "The read receipt path",
       text: "When the recipient reads messages, the client sends chat.markChat with their MongoDB message IDs. im/ws publishes msgReadTransfer. The task/mq consumer updates read records in MongoDB and, when receipt policy allows it, pushes a read receipt through the WebSocket gateway to online participants.",
    },
    recall: {
      title: "The recall path",
       text: "A recall request goes to im/api and then IM RPC. IM RPC validates the actor, chat context, role, and recall window, then conditionally marks the MongoDB document as recalled. Only after that state change does im/api publish msgRecallTransfer; task/mq only fans out the recall control frame to online participants.",
    },
    failures: {
      title: "What can go wrong, and where",
      headers: ["Failure point", "Effect", "Recovery"],
      rows: [
        {
          point: "im/ws crashes after ACK, before Kafka publish",
          effect: "Message lost. Sender has an ACK but the message never enters the pipeline.",
          recovery: "RigorAck makes this window smaller: the client has confirmed receipt before the handler runs. In practice the window is the time between ACK completion and the Kafka publish call, typically under 1 ms.",
        },
        {
          point: "Kafka broker unavailable",
          effect: "Publish fails; the handler returns an error to the caller.",
          recovery: "The sender's client should surface a delivery failure and offer retry.",
        },
        {
          point: "task/mq consumer crashes mid-write",
          effect: "MongoDB write is incomplete; Kafka offset is not committed.",
           recovery: "On restart, the consumer re-reads the uncommitted offset and retries. Because the chat insert has no business deduplication key, operators and clients must tolerate a possible duplicate after partial completion.",
        },
        {
          point: "Recipient offline at push time",
          effect: "Push silently skipped; no error.",
          recovery: "Message is in MongoDB. Client fetches on next connect via seq pull.",
        },
      ],
    },
    next: "Next: Realtime Gateway →",
  },
  zh: {
    metadata: {
      title: "消息生命周期 — HiChat 文档",
      description: "完整解析聊天消息从发送方到接收方的路径，包括 ACK、Kafka、持久化与离线投递。",
    },
    title: "消息生命周期",
    lead:
      "一条聊天消息在到达接收方之前会经过四个不同的系统。理解这条路径以及每次交接提供的保证，是分析 HiChat 投递失败、延迟和最终一致性的关键。",
    happyPath: "正常链路",
    ack: {
      title: "步骤 1 — ACK 协商（im/ws）",
      intro:
        "发送方输入消息后，Web 客户端通过已有的 WebSocket 连接发布 JSON 帧，其中 method: \"chat.user\" 表示单聊，\"chat.group\" 表示群聊。此时消息尚未持久化，帧仍在传输途中。im/ws 负责运行 WebSocket 服务；帧到达后，服务端的 ACK 模式决定消息交给 handler 之前要执行的流程。",
      modesIntro: "服务端支持三种 ACK 模式，通过 WithAck 在启动时配置：",
      modes: [
        "NoAck — 帧直接交给 handler。速度最快，但不提供投递保证。",
        "OnlyAck — 服务端立即返回 ACK 帧，其中 ackSeq = msgSeq + 1，随后将消息交给 handler。共两次往返，由服务端发起确认。",
        "RigorAck — 完整的三次握手。服务端发送 ACK seq 1，客户端必须回复 ACK seq 2，消息才会交给 handler。如果客户端未在 ackTimeout 内回复，该消息将被放弃。共三次往返，双方互相确认。",
      ],
      deduplication:
        "无论采用哪种模式，每条消息都有 UUID Id。服务端通过每个连接的 readMessageSeq 映射按 Id 去重：如果相同 ID 再次到达，且 AckSeq 小于或等于已记录值，第二份消息会被静默丢弃。",
    },
    kafka: {
      title: "步骤 2 — 发布到 Kafka（im/ws → Kafka）",
      paragraphs: [
         "ACK 完成后，chat.user handler 将消息发布到 Kafka 主题 msgChatTransfer。发布成功后 handler 返回，不等待 MongoDB。此处传输 ACK 只确认 WebSocket 交互，不代表消息已经持久化。",
        "这是系统有意作出的一致性取舍：如果 broker 接受了消息，但消费者在写入 MongoDB 前崩溃，消息仍可在主题保留期内从 Kafka 重放；不过在消费者追上进度前，历史消息查询中不会出现该消息。",
      ],
    },
    persistence: {
      title: "步骤 3 — 持久化（task/mq）",
      intro:
         "task/mq 运行消费组并消费 msgChatTransfer。负责该主题的消费者是 MsgChatTransfer，它扩展自 BaseMsgChatTransfer。",
      listIntro: "消费者依次执行以下三项操作：",
      steps: [
         "将消息文档插入 MongoDB 的 chat_logs 集合。",
         "更新 MongoDB 中的会话摘要和逐用户会话状态，包括未读与 @提醒投影。",
         "通过内部系统 WebSocket 请求 im/ws，把消息推送到接收方的活动连接。",
      ],
      idempotency:
         "当前聊天写入没有存储层业务去重键。如果消费者在插入消息后、后续步骤完成前重试，可能生成重复聊天文档。",
    },
    delivery: {
      title: "步骤 4 — 在线推送或离线存储",
      paragraphs: [
         "如果接收方存在活动的 WebSocket 连接，im/ws 会向其全部活动连接写入消息帧。否则消息保留在 MongoDB，客户端随后通过 IM API/RPC 拉取会话历史。",
         "持久化完成后，task/mq 还会向发送方回响真实的 MongoDB 消息 ID，客户端据此替换本地临时 ID，并区分传输 ACK 与持久化完成。",
      ],
    },
    readReceipt: {
      title: "已读回执链路",
       text: "接收方阅读消息后，客户端通过 chat.markChat 发送 MongoDB 消息 ID。im/ws 发布 msgReadTransfer，task/mq 消费者更新 MongoDB 已读记录；已读回执策略允许时，再通过 WebSocket 网关推送给在线参与者。",
    },
    recall: {
      title: "消息撤回链路",
       text: "撤回请求先到 im/api，再调用 IM RPC。IM RPC 校验操作者、会话、角色与撤回时间窗，并条件更新 MongoDB 文档为已撤回。状态更新成功后，im/api 才发布 msgRecallTransfer；task/mq 只负责向在线参与者扇出撤回控制帧。",
    },
    failures: {
      title: "可能发生的故障及其位置",
      headers: ["故障点", "影响", "恢复方式"],
      rows: [
        {
          point: "im/ws 在发送 ACK 后、发布到 Kafka 前崩溃",
          effect: "消息丢失。发送方已经收到 ACK，但消息从未进入处理管线。",
          recovery: "RigorAck 会缩小该窗口：handler 执行前客户端已确认收到 ACK。实际窗口是 ACK 完成到调用 Kafka 发布之间的时间，通常不足 1 ms。",
        },
        {
          point: "Kafka broker 不可用",
          effect: "发布失败，handler 向调用方返回错误。",
          recovery: "发送方客户端应明确提示投递失败，并提供重试操作。",
        },
        {
          point: "task/mq 消费者在写入过程中崩溃",
          effect: "MongoDB 写入未完成，Kafka offset 未提交。",
           recovery: "重启后，消费者重新读取未提交的 offset 并重试。由于聊天写入没有业务去重键，部分完成后重试可能产生重复，客户端和运维需对此容错。",
        },
        {
          point: "推送时接收方离线",
          effect: "静默跳过推送，不产生错误。",
          recovery: "消息已存入 MongoDB。客户端下次连接时通过 seq 拉取消息。",
        },
      ],
    },
    next: "下一篇：实时通信网关 →",
  },
} satisfies Record<Locale, PageContent>;

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);
  return pageContent[locale].metadata;
}

export default async function MessageLifecyclePage({ params }: Props) {
  const locale = resolveLocale((await params).locale);
  const content = pageContent[locale];

  return (
    <article className="prose-custom">
      <h1>{content.title}</h1>
      <p className="lead">{content.lead}</p>

      <h2>{content.happyPath}</h2>
      <h3>{content.ack.title}</h3>
      <p>{renderTerms(content.ack.intro, ["WebSocket", "JSON", "method: \"chat.user\"", "\"chat.group\"", "im/ws", "ACK", "handler"])}</p>
      <p>{renderTerms(content.ack.modesIntro, ["ACK", "WithAck"])}</p>
      <ul>
        {content.ack.modes.map((mode) => <li key={mode}>{renderTerms(mode, ackTerms)}</li>)}
      </ul>
      <p>{renderTerms(content.ack.deduplication, ["UUID", "Id", "readMessageSeq", "ID", "AckSeq"])}</p>

      <h3>{content.kafka.title}</h3>
      {content.kafka.paragraphs.map((paragraph) => (
         <p key={paragraph}>{renderTerms(paragraph, ["ACK", "chat.user", "Kafka", "msgChatTransfer", "MongoDB", "broker"])}</p>
      ))}

      <h3>{content.persistence.title}</h3>
       <p>{renderTerms(content.persistence.intro, ["task/mq", "msgChatTransfer", "MsgChatTransfer", "BaseMsgChatTransfer"])}</p>
      <p>{content.persistence.listIntro}</p>
      <ol>
        {content.persistence.steps.map((step) => (
           <li key={step}>{renderTerms(step, ["MongoDB", "chat_logs", "WebSocket", "im/ws"])}</li>
        ))}
      </ol>
       <p>{renderTerms(content.persistence.idempotency, ["MongoDB", "Kafka"])}</p>

      <h3>{content.delivery.title}</h3>
      {content.delivery.paragraphs.map((paragraph) => (
        <p key={paragraph}>{renderTerms(paragraph, ["WebSocket", "im/ws", "MongoDB", "seq", "Kafka"])}</p>
      ))}

      <h2>{content.readReceipt.title}</h2>
       <p>{renderTerms(content.readReceipt.text, ["chat.markChat", "msgReadTransfer", "task/mq", "MongoDB", "WebSocket", "im/ws"])}</p>

      <h2>{content.recall.title}</h2>
       <p>{renderTerms(content.recall.text, ["im/api", "IM RPC", "msgRecallTransfer", "task/mq", "MongoDB"])}</p>

      <h2>{content.failures.title}</h2>
      <table>
        <thead>
          <tr>{content.failures.headers.map((header) => <th key={header}>{header}</th>)}</tr>
        </thead>
        <tbody>
          {content.failures.rows.map((row) => (
            <tr key={row.point}>
              <td>{renderTerms(row.point, failureTerms)}</td>
              <td>{renderTerms(row.effect, failureTerms)}</td>
              <td>{renderTerms(row.recovery, failureTerms)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="not-prose mt-8 flex gap-3">
        <Link
          href={localeHref(locale, "/docs/realtime-gateway")}
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          {content.next}
        </Link>
      </div>
    </article>
  );
}

const ackTerms = ["NoAck", "OnlyAck", "RigorAck", "ACK", "ackSeq = msgSeq + 1", "ACK seq 1", "ACK seq 2", "ackTimeout", "handler"];
const failureTerms = ["im/ws", "ACK", "Kafka", "RigorAck", "handler", "1 ms", "broker", "task/mq", "MongoDB", "offset", "UUID", "upsert", "seq"];

function renderTerms(text: string, terms: string[]) {
  const pattern = new RegExp(`(${terms.map(escapeRegExp).join("|")})`, "g");
  return text.split(pattern).map((part, index) =>
    terms.includes(part) ? <code key={`${part}-${index}`}>{part}</code> : part,
  );
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
