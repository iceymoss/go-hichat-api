import type { ReactNode } from "react";
import type { Locale } from "@/i18n";

type DomainDiagramsProps = {
  locale: Locale;
};

const copy = {
  en: {
    context: {
      title: "Domain context map",
      description:
        "The four core domains exchange facts through RPC contracts. Task bridges committed events through Kafka, while Streaming owns realtime signaling and media coordination.",
      hint: "Arrows show the runtime direction of calls or events, not data ownership.",
      core: "CORE BUSINESS DOMAINS",
      async: "ASYNC BRIDGE",
      realtime: "REALTIME DOMAIN",
      user: "Identity & profiles",
      social: "Relations & groups",
      im: "Messages & presence",
      trend: "Posts & feeds",
      task: "Consumers & schedules",
      streaming: "Signaling · P2P / SFU",
      rpc: "RPC",
      kafka: "Kafka",
      ws: "internal WS",
      rpcLegend: "Synchronous RPC",
      kafkaLegend: "Kafka event",
      wsLegend: "Internal WebSocket",
    },
    ownership: {
      title: "Data ownership map",
      description:
        "Solid resources hold domain facts. Dashed resources are caches, delivery state, coordination, or other restartable runtime state.",
      hint: "A Task worker may write for an owning domain, but Task does not acquire ownership of that fact.",
      domain: "DOMAIN",
      resources: "OWNED AND RUNTIME RESOURCES",
      authoritative: "Authoritative",
      transient: "Derived / transient",
      noFacts: "No independent business facts",
      labels: {
        userMysql: "accounts · profiles · settings",
        socialMysql: "graph · groups · outboxes",
        socialRedis: "relationship cache · locks",
        imMongo: "chat · conversations",
        imMysql: "notifications · read state",
        imRedis: "presence · caches",
        imFile: "local attachment bytes",
        trendMysql: "posts · likes · comments",
        trendRedis: "publish-rate state",
        taskKafka: "consumer offsets · coordination",
        taskRedis: "locks · coordination",
        streamMemory: "calls · rooms · peer state",
        streamRedis: "relationship cache",
        streamMedia: "media path · no persistence",
      },
    },
  },
  zh: {
    context: {
      title: "领域上下文地图",
      description: "四个核心领域通过 RPC 契约交换事实；Task 通过 Kafka 衔接已提交事件，Streaming 独立负责实时信令与媒体协调。",
      hint: "箭头表示调用或事件的运行方向，不表示数据归属。",
      core: "核心业务领域",
      async: "异步桥",
      realtime: "实时领域",
      user: "身份与资料",
      social: "关系与群组",
      im: "消息与在线状态",
      trend: "动态与 Feed",
      task: "消费者与定时任务",
      streaming: "信令 · P2P / SFU",
      rpc: "RPC",
      kafka: "Kafka",
      ws: "内部 WS",
      rpcLegend: "同步 RPC",
      kafkaLegend: "Kafka 事件",
      wsLegend: "内部 WebSocket",
    },
    ownership: {
      title: "数据归属地图",
      description: "实线资源保存领域事实；虚线资源表示缓存、投递状态、协调信息或其他可重建的运行时状态。",
      hint: "Task worker 可以代表所属领域写入，但不会因此获得该业务事实的所有权。",
      domain: "领域",
      resources: "权威与运行时资源",
      authoritative: "权威数据",
      transient: "派生 / 临时",
      noFacts: "无独立业务事实",
      labels: {
        userMysql: "账号 · 资料 · 设置",
        socialMysql: "关系图 · 群组 · outbox",
        socialRedis: "关系缓存 · 锁",
        imMongo: "聊天记录 · 会话",
        imMysql: "通知 · 已读状态",
        imRedis: "在线状态 · 缓存",
        imFile: "本地附件文件",
        trendMysql: "动态 · 点赞 · 评论",
        trendRedis: "发布频率状态",
        taskKafka: "消费 offset · 协调",
        taskRedis: "锁 · 协调状态",
        streamMemory: "通话 · 房间 · peer 状态",
        streamRedis: "关系缓存",
        streamMedia: "媒体路径 · 不持久化",
      },
    },
  },
} satisfies Record<Locale, Record<string, unknown>>;

export function DomainDiagrams({ locale }: DomainDiagramsProps) {
  const text = copy[locale];

  return (
    <div className="mt-10 space-y-8">
      <DiagramFrame
        title={text.context.title}
        description={text.context.description}
        hint={text.context.hint}
        legend={
          <>
            <Legend color="#22d3ee" label={text.context.rpcLegend} />
            <Legend color="#fbbf24" label={text.context.kafkaLegend} dashed />
            <Legend color="#e879f9" label={text.context.wsLegend} dotted />
          </>
        }
      >
        <ContextMap text={text.context} />
      </DiagramFrame>

      <DiagramFrame
        title={text.ownership.title}
        description={text.ownership.description}
        hint={text.ownership.hint}
        legend={
          <>
            <Legend color="#34d399" label={text.ownership.authoritative} />
            <Legend color="#fbbf24" label={text.ownership.transient} dashed />
          </>
        }
      >
        <OwnershipMap text={text.ownership} />
      </DiagramFrame>
    </div>
  );
}

function DiagramFrame({
  title,
  description,
  hint,
  legend,
  children,
}: {
  title: string;
  description: string;
  hint: string;
  legend: ReactNode;
  children: ReactNode;
}) {
  return (
    <figure className="overflow-hidden rounded-2xl border border-slate-700 bg-slate-950 shadow-xl shadow-slate-950/10">
      <figcaption className="border-b border-slate-800 px-5 py-5 sm:px-6">
        <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-end">
          <div className="max-w-3xl">
            <h2 className="text-xl font-semibold tracking-tight text-slate-50 sm:text-2xl">{title}</h2>
            <p className="mt-2 text-sm leading-6 text-slate-400">{description}</p>
          </div>
          <div className="flex shrink-0 flex-wrap gap-x-4 gap-y-2 text-xs text-slate-400">{legend}</div>
        </div>
      </figcaption>
      <div className="overflow-x-auto bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.07),transparent_38%),radial-gradient(circle_at_bottom_right,rgba(232,121,249,0.06),transparent_35%)]">
        {children}
      </div>
      <p className="border-t border-slate-800 px-5 py-3 font-mono text-[11px] leading-5 text-slate-500 sm:px-6">{hint}</p>
    </figure>
  );
}

function Legend({ color, label, dashed = false, dotted = false }: { color: string; label: string; dashed?: boolean; dotted?: boolean }) {
  return (
    <span className="inline-flex items-center gap-2">
      <svg width="28" height="8" viewBox="0 0 28 8" aria-hidden="true">
        <line x1="1" y1="4" x2="27" y2="4" stroke={color} strokeWidth="2" strokeDasharray={dotted ? "2 4" : dashed ? "7 4" : undefined} strokeLinecap="round" />
      </svg>
      {label}
    </span>
  );
}

type ContextText = (typeof copy)[Locale]["context"];

function ContextMap({ text }: { text: ContextText }) {
  return (
    <svg className="block h-auto min-w-[920px]" viewBox="0 0 1120 610" role="img" aria-labelledby="domain-context-svg-title domain-context-svg-desc">
      <title id="domain-context-svg-title">{text.title}</title>
      <desc id="domain-context-svg-desc">{text.description}</desc>
      <defs>
        <marker id="context-rpc-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#22d3ee" />
        </marker>
        <marker id="context-kafka-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#fbbf24" />
        </marker>
        <marker id="context-ws-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#e879f9" />
        </marker>
      </defs>

      <rect x="32" y="48" width="610" height="514" rx="24" fill="#0f172a" fillOpacity="0.58" stroke="#334155" />
      <text x="58" y="82" fill="#94a3b8" fontSize="12" fontWeight="700" letterSpacing="2.2">{text.core}</text>
      <text x="726" y="82" fill="#fbbf24" fontSize="12" fontWeight="700" letterSpacing="2.2">{text.async}</text>
      <text x="816" y="390" fill="#e879f9" fontSize="12" fontWeight="700" letterSpacing="2.2">{text.realtime}</text>

      <ContextEdge d="M 405 164 H 279" label={text.rpc} labelX={342} labelY={151} kind="rpc" />
      <ContextEdge d="M 482 288 V 235" label={text.rpc} labelX={500} labelY={265} kind="rpc" />
      <ContextEdge d="M 279 447 C 350 447 370 211 405 211" label={text.rpc} labelX={355} labelY={337} kind="rpc" />
      <ContextEdge d="M 805 180 C 700 180 704 188 602 188" label={text.rpc} labelX={694} labelY={165} kind="rpc" />
      <ContextEdge d="M 805 220 C 700 220 703 342 602 342" label={text.rpc} labelX={700} labelY={296} kind="rpc" />
      <ContextEdge d="M 602 430 C 704 430 700 270 805 270" label={text.kafka} labelX={697} labelY={391} kind="kafka" />
      <ContextEdge d="M 279 447 C 620 520 700 244 805 244" label={text.kafka} labelX={690} labelY={332} kind="kafka" />
      <ContextEdge d="M 602 210 C 703 210 704 202 805 202" label={text.kafka} labelX={698} labelY={222} kind="kafka" />
      <ContextEdge d="M 805 292 C 724 292 700 365 602 365" label={text.ws} labelX={704} labelY={350} kind="ws" />
      <ContextEdge d="M 898 458 C 760 458 747 365 602 365" label={text.ws} labelX={731} labelY={435} kind="ws" />
      <ContextEdge d="M 905 414 C 795 414 763 216 602 216" label={text.rpc} labelX={746} labelY={308} kind="rpc" />

      <DomainNode x={82} y={112} name="User" subtitle={text.user} accent="#38bdf8" />
      <DomainNode x={405} y={112} name="Social" subtitle={text.social} accent="#2dd4bf" />
      <DomainNode x={405} y={288} name="IM" subtitle={text.im} accent="#a78bfa" />
      <DomainNode x={82} y={400} name="Trend" subtitle={text.trend} accent="#fb7185" />
      <DomainNode x={805} y={154} name="Task" subtitle={text.task} accent="#fbbf24" width={250} />
      <DomainNode x={816} y={414} name="Streaming" subtitle={text.streaming} accent="#e879f9" width={250} />
    </svg>
  );
}

type EdgeKind = "rpc" | "kafka" | "ws";

function ContextEdge({ d, label, labelX, labelY, kind }: { d: string; label: string; labelX: number; labelY: number; kind: EdgeKind }) {
  const styles = {
    rpc: { color: "#22d3ee", dash: undefined, marker: "url(#context-rpc-arrow)" },
    kafka: { color: "#fbbf24", dash: "8 6", marker: "url(#context-kafka-arrow)" },
    ws: { color: "#e879f9", dash: "2 6", marker: "url(#context-ws-arrow)" },
  }[kind];

  return (
    <g>
      <path d={d} fill="none" stroke={styles.color} strokeWidth="2" strokeDasharray={styles.dash} strokeLinecap="round" strokeLinejoin="round" markerEnd={styles.marker} opacity="0.9" />
      <rect x={labelX - 29} y={labelY - 12} width="58" height="18" rx="9" fill="#020617" stroke={styles.color} strokeOpacity="0.45" />
      <text x={labelX} y={labelY + 1} textAnchor="middle" fill={styles.color} fontSize="10" fontWeight="700">{label}</text>
    </g>
  );
}

function DomainNode({ x, y, name, subtitle, accent, width = 197 }: { x: number; y: number; name: string; subtitle: string; accent: string; width?: number }) {
  return (
    <g>
      <rect x={x} y={y} width={width} height="123" rx="18" fill="#111c30" stroke={accent} strokeOpacity="0.7" />
      <rect x={x} y={y} width="6" height="123" rx="3" fill={accent} />
      <circle cx={x + 31} cy={y + 34} r="7" fill={accent} fillOpacity="0.22" stroke={accent} />
      <text x={x + 51} y={y + 40} fill="#f8fafc" fontSize="21" fontWeight="700">{name}</text>
      <text x={x + 30} y={y + 79} fill="#94a3b8" fontSize="13">{subtitle}</text>
      <path d={`M ${x + 30} ${y + 96} H ${x + width - 24}`} stroke="#334155" />
      <text x={x + 30} y={y + 112} fill={accent} fontFamily="monospace" fontSize="10">domain boundary</text>
    </g>
  );
}

type OwnershipText = (typeof copy)[Locale]["ownership"];

function OwnershipMap({ text }: { text: OwnershipText }) {
  const rows = [
    { domain: "User", note: "user/api · user/rpc", resources: [{ name: "MySQL", detail: text.labels.userMysql, authority: true }] },
    { domain: "Social", note: "social/api · social/rpc", resources: [{ name: "MySQL", detail: text.labels.socialMysql, authority: true }, { name: "Redis", detail: text.labels.socialRedis, authority: false }] },
    { domain: "IM", note: "im/api · im/rpc · im/ws", resources: [{ name: "MongoDB", detail: text.labels.imMongo, authority: true }, { name: "MySQL", detail: text.labels.imMysql, authority: true }, { name: "Redis", detail: text.labels.imRedis, authority: false }, { name: "File", detail: text.labels.imFile, authority: true }] },
    { domain: "Trend", note: "trend/api · trend/rpc", resources: [{ name: "MySQL", detail: text.labels.trendMysql, authority: true }, { name: "Redis", detail: text.labels.trendRedis, authority: false }] },
    { domain: "Task", note: text.noFacts, resources: [{ name: "Kafka", detail: text.labels.taskKafka, authority: false }, { name: "Redis", detail: text.labels.taskRedis, authority: false }] },
    { domain: "Streaming", note: "signaling · Pion SFU", resources: [{ name: "Process Memory", detail: text.labels.streamMemory, authority: true }, { name: "Redis", detail: text.labels.streamRedis, authority: false }, { name: "P2P / SFU", detail: text.labels.streamMedia, authority: false }] },
  ];

  return (
    <svg className="block h-auto min-w-[980px]" viewBox="0 0 1180 770" role="img" aria-labelledby="data-ownership-svg-title data-ownership-svg-desc">
      <title id="data-ownership-svg-title">{text.title}</title>
      <desc id="data-ownership-svg-desc">{text.description}</desc>
      <defs>
        <marker id="ownership-solid-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#34d399" />
        </marker>
        <marker id="ownership-dashed-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#fbbf24" />
        </marker>
      </defs>
      <text x="42" y="43" fill="#94a3b8" fontSize="11" fontWeight="700" letterSpacing="2">{text.domain}</text>
      <text x="350" y="43" fill="#94a3b8" fontSize="11" fontWeight="700" letterSpacing="2">{text.resources}</text>
      {rows.map((row, rowIndex) => {
        const y = 72 + rowIndex * 115;
        const chipWidth = row.resources.length === 4 ? 190 : 240;
        const gap = row.resources.length === 4 ? 14 : 22;

        return (
          <g key={row.domain}>
            <rect x="28" y={y} width="264" height="82" rx="14" fill="#111c30" stroke="#475569" />
            <text x="50" y={y + 32} fill="#f8fafc" fontSize="18" fontWeight="700">{row.domain}</text>
            <text x="50" y={y + 57} fill={row.domain === "Task" ? "#fbbf24" : "#94a3b8"} fontSize="11">{row.note}</text>
            {row.resources.map((resource, index) => {
              const x = 350 + index * (chipWidth + gap);
              const color = resource.authority ? "#34d399" : "#fbbf24";
              return (
                <g key={`${row.domain}-${resource.name}`}>
                  <path d={`M 292 ${y + 41} H ${x - 18} V ${y + 41} H ${x}`} fill="none" stroke={color} strokeWidth="2" strokeDasharray={resource.authority ? undefined : "7 6"} markerEnd={resource.authority ? "url(#ownership-solid-arrow)" : "url(#ownership-dashed-arrow)"} opacity="0.82" />
                  <rect x={x} y={y} width={chipWidth} height="82" rx="14" fill={resource.authority ? "#09251f" : "#241d0b"} stroke={color} strokeDasharray={resource.authority ? undefined : "7 5"} strokeOpacity="0.85" />
                  <circle cx={x + 20} cy={y + 25} r="5" fill={color} />
                  <text x={x + 33} y={y + 30} fill="#f8fafc" fontFamily="monospace" fontSize={resource.name.length > 12 ? "12" : "14"} fontWeight="700">{resource.name}</text>
                  <text x={x + 18} y={y + 57} fill="#94a3b8" fontSize={row.resources.length === 4 ? "10" : "11"}>{resource.detail}</text>
                </g>
              );
            })}
            {rowIndex < rows.length - 1 ? <line x1="28" y1={y + 98} x2="1152" y2={y + 98} stroke="#1e293b" /> : null}
          </g>
        );
      })}
    </svg>
  );
}
