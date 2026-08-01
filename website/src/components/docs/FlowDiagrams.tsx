type DiagramLocale = "en" | "zh";

type NodeKind = "service" | "client" | "queue" | "store" | "memory";

type DiagramNode = {
  id: string;
  x: number;
  y?: number;
  width?: number;
  kind: NodeKind;
  title: string;
  subtitle?: string;
};

type DiagramEdge = {
  from: string;
  to: string;
  label: string;
  async?: boolean;
  path?: string;
  labelX?: number;
  labelY?: number;
};

type Diagram = {
  title: string;
  nodes: DiagramNode[];
  edges: DiagramEdge[];
  notes?: { x: number; y: number; width: number; text: string; tone?: "warn" | "info" }[];
};

const WIDTH = 1160;
const NODE_WIDTH = 132;
const NODE_HEIGHT = 66;

function diagramFor(flowNumber: string, locale: DiagramLocale): Diagram | undefined {
  const t = (en: string, zh: string) => (locale === "zh" ? zh : en);

  const diagrams: Record<string, Diagram> = {
    "01": {
      title: t("Message send pipeline", "消息发送链路"),
      nodes: [
        { id: "client", x: 24, kind: "client", title: t("Client", "客户端"), subtitle: "chat frame" },
        { id: "ws-in", x: 190, kind: "service", title: "IM WebSocket", subtitle: t("chat handler", "聊天处理器") },
        { id: "kafka", x: 356, kind: "queue", title: "Kafka", subtitle: "msgChatTransfer" },
        { id: "task", x: 522, kind: "service", title: "Task MQ", subtitle: t("consumer", "消费者") },
        { id: "mongo", x: 688, kind: "store", title: "MongoDB", subtitle: t("chat history", "聊天记录") },
        { id: "ws-out", x: 854, kind: "service", title: "IM WebSocket", subtitle: t("online fan-out", "在线扇出") },
        { id: "peers", x: 1020, width: 116, kind: "client", title: t("Clients", "收发两端"), subtitle: t("online / pull", "在线 / 拉取") },
      ],
      edges: [
        { from: "client", to: "ws-in", label: t("send", "发送") },
        { from: "ws-in", to: "kafka", label: t("publish", "发布"), async: true },
        { from: "kafka", to: "task", label: t("consume", "消费"), async: true },
        { from: "task", to: "mongo", label: t("persist first", "先持久化"), async: true },
        { from: "mongo", to: "ws-out", label: t("stored ID", "已落库 ID"), async: true },
        { from: "ws-out", to: "peers", label: t("push / echo", "推送 / 回响"), async: true },
      ],
      notes: [{ x: 680, y: 174, width: 298, text: t("No storage-level business deduplication key", "当前没有存储层业务幂等键"), tone: "warn" }],
    },
    "02": {
      title: t("Read receipt sequence", "已读回执链路"),
      nodes: [
        { id: "client", x: 18, width: 122, kind: "client", title: t("Reader", "阅读端"), subtitle: "chat.markChat" },
        { id: "ws-in", x: 178, width: 132, kind: "service", title: "IM WebSocket", subtitle: t("validate", "校验") },
        { id: "kafka", x: 348, width: 132, kind: "queue", title: "Kafka", subtitle: "msgReadTransfer" },
        { id: "task", x: 518, width: 132, kind: "service", title: "Task MQ", subtitle: t("read consumer", "已读消费者") },
        { id: "mongo", x: 688, width: 132, kind: "store", title: "MongoDB", subtitle: t("read records", "已读记录") },
        { id: "ws-out", x: 858, width: 132, kind: "service", title: "IM WebSocket", subtitle: t("receipt push", "回执推送") },
        { id: "sender", x: 1028, width: 114, kind: "client", title: t("Sender", "发送端"), subtitle: t("online", "在线") },
      ],
      edges: [
        { from: "client", to: "ws-in", label: t("mark IDs", "标记 IDs") },
        { from: "ws-in", to: "kafka", label: t("publish", "发布"), async: true },
        { from: "kafka", to: "task", label: t("consume", "消费"), async: true },
        { from: "task", to: "mongo", label: t("update", "更新"), async: true },
        { from: "mongo", to: "ws-out", label: t("persisted", "已持久化"), async: true },
        { from: "ws-out", to: "sender", label: t("best effort", "尽力推送"), async: true },
      ],
      notes: [{ x: 799, y: 174, width: 298, text: t("Persisted state supports recovery after reconnect", "重连后以持久化状态恢复"), tone: "info" }],
    },
    "03": {
      title: t("Message recall sequence", "消息撤回链路"),
      nodes: [
        { id: "api", x: 45, kind: "service", title: "IM API", subtitle: t("authenticate", "认证") },
        { id: "rpc", x: 225, kind: "service", title: "IM RPC", subtitle: t("validate window", "校验时间窗") },
        { id: "mongo", x: 405, kind: "store", title: "MongoDB", subtitle: t("conditional recall", "条件更新撤回态") },
        { id: "kafka", x: 585, kind: "queue", title: "Kafka", subtitle: "msgRecallTransfer" },
        { id: "task", x: 765, kind: "service", title: "Task MQ", subtitle: t("control frame", "控制帧") },
        { id: "ws", x: 945, kind: "service", title: "IM WebSocket", subtitle: t("participants", "会话参与端") },
      ],
      edges: [
        { from: "api", to: "rpc", label: t("recall", "撤回") },
        { from: "rpc", to: "mongo", label: t("update first", "先更新") },
        { from: "mongo", to: "kafka", label: t("then publish", "再发布"), async: true },
        { from: "kafka", to: "task", label: t("consume", "消费"), async: true },
        { from: "task", to: "ws", label: t("best effort", "尽力推送"), async: true },
      ],
      notes: [{ x: 397, y: 174, width: 320, text: t("MongoDB update and Kafka publish are not atomic", "MongoDB 更新与 Kafka 发布不原子"), tone: "warn" }],
    },
    "04": {
      title: t("Reliable social notification", "社交可靠通知链路"),
      nodes: [
        { id: "social", x: 18, width: 140, kind: "store", title: "Social MySQL", subtitle: t("state + outbox", "状态 + outbox") },
        { id: "relay", x: 180, kind: "service", title: t("Outbox relay", "Outbox relay"), subtitle: t("poll + retry", "轮询 + 重试") },
        { id: "kafka", x: 342, width: 144, kind: "queue", title: "Kafka", subtitle: "social.request...v1" },
        { id: "task", x: 508, kind: "service", title: "Task MQ", subtitle: t("consumer group", "消费组") },
        { id: "rpc", x: 670, kind: "service", title: "IM RPC", subtitle: t("create notification", "创建通知") },
        { id: "mysql", x: 832, width: 140, kind: "store", title: "IM MySQL", subtitle: "notifications" },
        { id: "ws", x: 994, width: 146, kind: "service", title: "IM WebSocket", subtitle: "push.notify" },
      ],
      edges: [
        { from: "social", to: "relay", label: t("poll", "轮询"), async: true },
        { from: "relay", to: "kafka", label: t("retry publish", "重试发布"), async: true },
        { from: "kafka", to: "task", label: t("at least once", "至少一次"), async: true },
        { from: "task", to: "rpc", label: "RPC" },
        { from: "rpc", to: "mysql", label: t("persist", "持久化") },
        { from: "mysql", to: "ws", label: t("then push", "再推送"), async: true },
      ],
      notes: [{ x: 20, y: 174, width: 302, text: t("Social state + outbox commit atomically", "社交状态与 outbox 原子提交"), tone: "info" }],
    },
    "05": {
      title: t("Ordered relation change", "关系变更有序链路"),
      nodes: [
        { id: "social", x: 18, width: 140, kind: "store", title: "Social MySQL", subtitle: t("relation + outbox", "关系 + outbox") },
        { id: "relay", x: 180, kind: "service", title: t("Relation relay", "关系 relay"), subtitle: t("aggregate key", "聚合键") },
        { id: "kafka", x: 342, width: 144, kind: "queue", title: "Kafka", subtitle: t("keyed partition", "按 key 分区") },
        { id: "task", x: 508, kind: "service", title: "Task MQ", subtitle: t("version gate", "版本门") },
        { id: "state", x: 670, kind: "store", title: "IM state", subtitle: t("freeze / recover", "冻结 / 恢复") },
        { id: "redis", x: 832, width: 140, kind: "store", title: "Redis", subtitle: t("relation cache", "关系缓存") },
        { id: "ws", x: 994, width: 146, kind: "service", title: "IM WebSocket", subtitle: "push.relation" },
      ],
      edges: [
        { from: "social", to: "relay", label: t("poll", "轮询"), async: true },
        { from: "relay", to: "kafka", label: t("key + version", "key + 版本"), async: true },
        { from: "kafka", to: "task", label: t("ordered / key", "同 key 有序"), async: true },
        { from: "task", to: "state", label: t("reliable first", "先可靠变更"), async: true },
        { from: "state", to: "redis", label: t("repairable", "可修复"), async: true },
        { from: "redis", to: "ws", label: t("invalidate", "失效通知"), async: true },
      ],
      notes: [{ x: 337, y: 174, width: 330, text: t("Ordering is per key/partition, never global", "仅保证 key/partition 内有序，非全局有序"), tone: "info" }],
    },
    "06": {
      title: t("Trend interaction notification", "动态互动通知链路"),
      nodes: [
        { id: "rpc", x: 55, kind: "service", title: "Trend RPC", subtitle: t("like / comment", "点赞 / 评论") },
        { id: "mysql", x: 245, width: 146, kind: "store", title: "Trend MySQL", subtitle: "trend_message" },
        { id: "api", x: 441, kind: "service", title: "Trend API", subtitle: t("after RPC returns", "RPC 返回后") },
        { id: "kafka", x: 631, width: 146, kind: "queue", title: "Kafka", subtitle: "trendNotifyTransfer" },
        { id: "task", x: 827, kind: "service", title: "Task MQ", subtitle: t("translate only", "仅协议转换") },
        { id: "ws", x: 1017, width: 120, kind: "service", title: "IM WS", subtitle: "push.trend" },
      ],
      edges: [
        { from: "rpc", to: "mysql", label: t("persist", "持久化") },
        { from: "mysql", to: "api", label: t("return", "返回") },
        { from: "api", to: "kafka", label: t("publish", "发布"), async: true },
        { from: "kafka", to: "task", label: t("consume", "消费"), async: true },
        { from: "task", to: "ws", label: t("best effort", "尽力推送"), async: true },
      ],
      notes: [{ x: 236, y: 174, width: 404, text: t("trend_message write and Kafka publish are not atomic", "trend_message 写入与 Kafka 发布不原子"), tone: "warn" }],
    },
    "07": {
      title: t("Call signalling and media paths", "音视频信令与媒体链路"),
      nodes: [
        { id: "caller", x: 26, y: 93, kind: "client", title: t("Browsers", "浏览器"), subtitle: "WebRTC" },
        { id: "stream", x: 230, y: 93, width: 150, kind: "service", title: "Streaming WS", subtitle: t("SDP / ICE / control", "SDP / ICE / 控制") },
        { id: "memory", x: 502, y: 18, width: 154, kind: "memory", title: t("Process memory", "进程内存"), subtitle: t("rooms + peers", "房间 + peer") },
        { id: "fallback", x: 502, y: 168, width: 154, kind: "service", title: "IM WebSocket", subtitle: "push.call fallback" },
        { id: "p2p", x: 790, y: 18, width: 146, kind: "client", title: "P2P media", subtitle: t("direct call", "直接通话") },
        { id: "sfu", x: 790, y: 168, width: 146, kind: "memory", title: t("In-process SFU", "进程内 SFU"), subtitle: t("room tracks", "房间媒体轨") },
        { id: "peers", x: 1008, y: 93, width: 126, kind: "client", title: t("Participants", "参与端"), subtitle: t("reconnect", "重连恢复") },
      ],
      edges: [
        { from: "caller", to: "stream", label: t("signal", "信令") },
        { from: "stream", to: "memory", label: t("ephemeral state", "瞬时状态"), path: "M380 112 C430 112 438 51 502 51", labelX: 440, labelY: 66 },
        { from: "stream", to: "fallback", label: t("target absent", "目标不在线"), async: true, path: "M380 140 C430 140 438 201 502 201", labelX: 442, labelY: 194 },
        { from: "memory", to: "p2p", label: t("negotiate", "协商"), path: "M656 51 L790 51" },
        { from: "memory", to: "sfu", label: t("publish tracks", "发布媒体轨"), path: "M579 84 C579 126 720 201 790 201", labelX: 704, labelY: 145 },
        { from: "fallback", to: "peers", label: t("best effort", "尽力回退"), async: true, path: "M656 201 C800 250 930 126 1008 126", labelX: 843, labelY: 226 },
        { from: "p2p", to: "peers", label: t("media", "媒体"), path: "M936 51 C976 51 976 112 1008 112", labelX: 977, labelY: 61 },
        { from: "sfu", to: "peers", label: t("forward", "转发"), path: "M936 201 C976 201 976 140 1008 140", labelX: 978, labelY: 196 },
      ],
      notes: [{ x: 496, y: 254, width: 446, text: t("No durable room state or offline signalling queue", "无持久化房间状态，也无离线信令队列"), tone: "warn" }],
    },
  };

  return diagrams[flowNumber];
}

export function FlowDiagram({ flowNumber, locale }: { flowNumber: string; locale: DiagramLocale }) {
  const diagram = diagramFor(flowNumber, locale);
  if (!diagram) return null;

  const height = flowNumber === "07" ? 310 : 238;
  const nodeMap = new Map(diagram.nodes.map((node) => [node.id, node]));
  const markerId = `flow-arrow-${flowNumber}`;

  return (
    <div className="mt-5 overflow-x-auto rounded-2xl border border-white/10 bg-[#070b13]">
      <svg
        aria-label={diagram.title}
        className="block min-w-[72rem]"
        role="img"
        viewBox={`0 0 ${WIDTH} ${height}`}
        xmlns="http://www.w3.org/2000/svg"
      >
        <title>{diagram.title}</title>
        <defs>
          <marker id={markerId} markerHeight="7" markerWidth="8" orient="auto" refX="7" refY="3.5">
            <path d="M0 0L8 3.5L0 7Z" fill="#94a3b8" />
          </marker>
          <linearGradient id={`flow-bg-${flowNumber}`} x1="0" x2="1">
            <stop stopColor="#0b1220" />
            <stop offset="1" stopColor="#080d16" />
          </linearGradient>
        </defs>
        <rect width={WIDTH} height={height} rx="16" fill={`url(#flow-bg-${flowNumber})`} />
        <text x="22" y="25" fill="#64748b" fontFamily="ui-monospace, monospace" fontSize="10" letterSpacing="1.5">
          {diagram.title.toUpperCase()}
        </text>

        {diagram.edges.map((edge, index) => {
          const from = nodeMap.get(edge.from);
          const to = nodeMap.get(edge.to);
          if (!from || !to) return null;
          const fromWidth = from.width ?? NODE_WIDTH;
          const fromY = from.y ?? 68;
          const toY = to.y ?? 68;
          const defaultPath = `M${from.x + fromWidth} ${fromY + NODE_HEIGHT / 2} L${to.x} ${toY + NODE_HEIGHT / 2}`;
          const labelX = edge.labelX ?? (from.x + fromWidth + to.x) / 2;
          const labelY = edge.labelY ?? Math.min(fromY, toY) + NODE_HEIGHT / 2 - 10;

          return (
            <g key={`${edge.from}-${edge.to}-${index}`}>
              <path
                d={edge.path ?? defaultPath}
                fill="none"
                markerEnd={`url(#${markerId})`}
                stroke={edge.async ? "#c084fc" : "#38bdf8"}
                strokeDasharray={edge.async ? "6 6" : undefined}
                strokeLinecap="round"
                strokeWidth="1.5"
              />
              <text fill={edge.async ? "#d8b4fe" : "#7dd3fc"} fontFamily="ui-monospace, monospace" fontSize="9" textAnchor="middle" x={labelX} y={labelY}>
                {edge.label}
              </text>
            </g>
          );
        })}

        {diagram.nodes.map((node) => (
          <DiagramNodeView key={node.id} node={node} />
        ))}

        {diagram.notes?.map((note) => (
          <g key={note.text}>
            <rect
              fill={note.tone === "warn" ? "#2a170d" : "#0b1d2b"}
              height="30"
              rx="8"
              stroke={note.tone === "warn" ? "#f59e0b" : "#0ea5e9"}
              strokeOpacity=".45"
              width={note.width}
              x={note.x}
              y={note.y}
            />
            <text fill={note.tone === "warn" ? "#fbbf24" : "#7dd3fc"} fontFamily="ui-monospace, monospace" fontSize="10" textAnchor="middle" x={note.x + note.width / 2} y={note.y + 19}>
              {note.text}
            </text>
          </g>
        ))}

        <g transform={`translate(22 ${height - 17})`}>
          <line stroke="#38bdf8" strokeWidth="1.5" x2="28" />
          <text fill="#64748b" fontFamily="ui-monospace, monospace" fontSize="9" x="35" y="3">{locale === "zh" ? "同步" : "SYNC"}</text>
          <line stroke="#c084fc" strokeDasharray="5 5" strokeWidth="1.5" x1="82" x2="110" />
          <text fill="#64748b" fontFamily="ui-monospace, monospace" fontSize="9" x="117" y="3">{locale === "zh" ? "异步" : "ASYNC"}</text>
        </g>
      </svg>
    </div>
  );
}

function DiagramNodeView({ node }: { node: DiagramNode }) {
  const x = node.x;
  const y = node.y ?? 68;
  const width = node.width ?? NODE_WIDTH;
  const palette = {
    client: { fill: "#111827", stroke: "#64748b" },
    service: { fill: "#0c1b2b", stroke: "#0ea5e9" },
    queue: { fill: "#21142d", stroke: "#c084fc" },
    store: { fill: "#281b0c", stroke: "#f59e0b" },
    memory: { fill: "#17220d", stroke: "#84cc16" },
  }[node.kind];

  return (
    <g>
      {node.kind === "store" || node.kind === "memory" ? (
        <g>
          <rect fill={palette.fill} height={NODE_HEIGHT - 14} stroke={palette.stroke} strokeOpacity=".7" width={width} x={x} y={y + 7} />
          <ellipse cx={x + width / 2} cy={y + 7} fill={palette.fill} rx={width / 2} ry="7" stroke={palette.stroke} strokeOpacity=".7" />
          <path d={`M${x} ${y + NODE_HEIGHT - 7} A${width / 2} 7 0 0 0 ${x + width} ${y + NODE_HEIGHT - 7}`} fill="none" stroke={palette.stroke} strokeOpacity=".7" />
        </g>
      ) : node.kind === "queue" ? (
        <path d={`M${x + 10} ${y}H${x + width - 10}L${x + width} ${y + NODE_HEIGHT / 2}L${x + width - 10} ${y + NODE_HEIGHT}H${x + 10}L${x} ${y + NODE_HEIGHT / 2}Z`} fill={palette.fill} stroke={palette.stroke} strokeOpacity=".75" />
      ) : (
        <rect fill={palette.fill} height={NODE_HEIGHT} rx={node.kind === "client" ? 14 : 8} stroke={palette.stroke} strokeOpacity=".75" width={width} x={x} y={y} />
      )}
      <text fill="#f8fafc" fontFamily="ui-sans-serif, system-ui, sans-serif" fontSize="12" fontWeight="600" textAnchor="middle" x={x + width / 2} y={y + 29}>
        {node.title}
      </text>
      {node.subtitle ? (
        <text fill="#94a3b8" fontFamily="ui-monospace, monospace" fontSize="8.5" textAnchor="middle" x={x + width / 2} y={y + 46}>
          {node.subtitle}
        </text>
      ) : null}
    </g>
  );
}
