import type { Locale } from "@/i18n";

type Point = { x: number; y: number };
type NodeProps = Point & {
  width?: number;
  height?: number;
  label: string;
  detail?: string;
  tone?: "default" | "rpc" | "worker" | "media" | "infra";
};

type EdgeProps = {
  d: string;
  marker: string;
  kind?: "rpc" | "kafka" | "ws" | "media" | "ingress";
};

const copy = {
  en: {
    topologyTitle: "System topology",
    topologyDescription: "Runtime entry points, domain services, workers, and shared infrastructure with directed traffic.",
    dependencyTitle: "Service dependency graph",
    dependencyDescription: "Only dependencies constructed by the running services are shown.",
    clients: "Clients",
    ingress: "Ingress",
    domain: "Domain RPC",
    workers: "Workers & media",
    infrastructure: "Infrastructure",
    relay: "Social Relay",
    relayDetail: "outbox publisher",
    sfuDetail: "embedded in Streaming",
    rpc: "Synchronous RPC",
    kafka: "Kafka event",
    ws: "Internal WebSocket",
    media: "Media",
    http: "HTTP / WS / WebRTC",
  },
  zh: {
    topologyTitle: "系统拓扑",
    topologyDescription: "运行时接入点、领域服务、任务与共享基础设施之间的真实有向流量。",
    dependencyTitle: "服务依赖图",
    dependencyDescription: "仅展示运行服务在代码中实际构造的依赖。",
    clients: "客户端",
    ingress: "接入层",
    domain: "领域 RPC",
    workers: "任务与媒体",
    infrastructure: "基础设施",
    relay: "Social Relay",
    relayDetail: "outbox 投递器",
    sfuDetail: "内置于 Streaming",
    rpc: "同步 RPC",
    kafka: "Kafka 事件",
    ws: "内部 WebSocket",
    media: "媒体流",
    http: "HTTP / WS / WebRTC",
  },
} satisfies Record<Locale, Record<string, string>>;

const edgeClasses = {
  rpc: "stroke-cyan-400",
  kafka: "stroke-amber-400 [stroke-dasharray:8_6]",
  ws: "stroke-fuchsia-400 [stroke-dasharray:2_7]",
  media: "stroke-emerald-400",
  ingress: "stroke-slate-500 dark:stroke-slate-400",
};

function DiagramNode({ x, y, width = 150, height = 54, label, detail, tone = "default" }: NodeProps) {
  const tones = {
    default: "stroke-violet-400/60 fill-violet-400/[0.08]",
    rpc: "stroke-cyan-400/70 fill-cyan-400/[0.08]",
    worker: "stroke-amber-400/70 fill-amber-400/[0.08]",
    media: "stroke-emerald-400/70 fill-emerald-400/[0.08]",
    infra: "stroke-slate-500/60 fill-slate-400/[0.06]",
  };

  return (
    <g>
      <rect x={x} y={y} width={width} height={height} rx="10" className={tones[tone]} strokeWidth="1.5" />
      <text x={x + width / 2} y={y + (detail ? 23 : 32)} textAnchor="middle" className="fill-slate-900 text-[13px] font-semibold dark:fill-slate-100">
        {label}
      </text>
      {detail && (
        <text x={x + width / 2} y={y + 40} textAnchor="middle" className="fill-slate-500 text-[10px] dark:fill-slate-400">
          {detail}
        </text>
      )}
    </g>
  );
}

function Edge({ d, marker, kind = "rpc" }: EdgeProps) {
  return <path d={d} fill="none" className={edgeClasses[kind]} strokeWidth={kind === "ws" ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round" markerEnd={`url(#${marker})`} />;
}

function MarkerDefs({ prefix }: { prefix: "topology" | "dependency" }) {
  return (
    <defs>
      {(["rpc", "kafka", "ws", "media", "ingress"] as const).map((kind) => (
        <marker key={kind} id={`${prefix}-${kind}-arrow`} viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M 0 0 L 10 5 L 0 10 z" className={kind === "ingress" ? "fill-slate-500 dark:fill-slate-400" : edgeClasses[kind].split(" ")[0].replace("stroke-", "fill-")} />
        </marker>
      ))}
    </defs>
  );
}

function LayerLabel({ x, y, children }: Point & { children: React.ReactNode }) {
  return <text x={x} y={y} className="fill-slate-500 text-[10px] font-semibold uppercase tracking-[0.16em] dark:fill-slate-400">{children}</text>;
}

function Legend({ labels, y }: { labels: typeof copy.en; y: number }) {
  const entries = [
    ["rpc", labels.rpc],
    ["kafka", labels.kafka],
    ["ws", labels.ws],
    ["media", labels.media],
  ] as const;

  return (
    <g transform={`translate(52 ${y})`}>
      {entries.map(([kind, label], index) => {
        const x = index * 270;
        return (
          <g key={kind} transform={`translate(${x} 0)`}>
            <line x1="0" y1="0" x2="34" y2="0" className={edgeClasses[kind]} strokeWidth="2.5" />
            <text x="44" y="4" className="fill-slate-600 text-[11px] dark:fill-slate-300">{label}</text>
          </g>
        );
      })}
    </g>
  );
}

function DiagramFrame({ title, description, children, minWidth = 1040 }: { title: string; description: string; children: React.ReactNode; minWidth?: number }) {
  return (
    <figure className="overflow-hidden rounded-2xl border border-border bg-slate-50 shadow-sm dark:bg-slate-950/70">
      <figcaption className="border-b border-border px-5 py-4 sm:px-6">
        <h3 className="text-base font-semibold text-foreground">{title}</h3>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">{description}</p>
      </figcaption>
      <div className="overflow-x-auto">
        <div style={{ minWidth }}>
          {children}
        </div>
      </div>
    </figure>
  );
}

function SystemTopology({ locale }: { locale: Locale }) {
  const labels = copy[locale];
  const ingress = ["User API", "Social API", "IM API", "Trend API", "IM WS", "Streaming"];
  const rpc = ["User RPC", "Social RPC", "IM RPC", "Trend RPC"];
  const infra = ["Kafka", "MySQL", "MongoDB", "Redis", "Etcd"];
  const topologyMarker = (kind: keyof typeof edgeClasses) => `topology-${kind}-arrow`;

  return (
    <DiagramFrame title={labels.topologyTitle} description={labels.topologyDescription}>
      <svg viewBox="0 0 1200 900" role="img" aria-labelledby="system-topology-title system-topology-description" className="block h-auto w-full min-w-[1040px]">
        <title id="system-topology-title">{labels.topologyTitle}</title>
        <desc id="system-topology-description">{labels.topologyDescription}</desc>
        <MarkerDefs prefix="topology" />
        <g opacity="0.65" className="stroke-slate-300 dark:stroke-slate-700">
          {[130, 300, 470, 650].map((y) => <line key={y} x1="42" y1={y} x2="1158" y2={y} strokeDasharray="3 7" />)}
        </g>

        <LayerLabel x={48} y={30}>{labels.clients}</LayerLabel>
        <LayerLabel x={48} y={146}>{labels.ingress}</LayerLabel>
        <LayerLabel x={48} y={316}>{labels.domain}</LayerLabel>
        <LayerLabel x={48} y={486}>{labels.workers}</LayerLabel>
        <LayerLabel x={48} y={666}>{labels.infrastructure}</LayerLabel>

        <DiagramNode x={500} y={48} width={200} label={labels.clients} detail="Web / Mobile / Desktop" />
        <text x="600" y="124" textAnchor="middle" className="fill-slate-500 text-[10px] dark:fill-slate-400">{labels.http}</text>

        {ingress.map((label, index) => <DiagramNode key={label} x={48 + index * 188} y={166} width={164} label={label} tone={label === "Streaming" ? "media" : "default"} />)}
        {rpc.map((label, index) => <DiagramNode key={label} x={156 + index * 248} y={350} width={180} label={label} tone="rpc" />)}
        <DiagramNode x={98} y={520} width={190} label={labels.relay} detail={labels.relayDetail} tone="worker" />
        <DiagramNode x={370} y={520} width={170} label="Task MQ" tone="worker" />
        <DiagramNode x={640} y={520} width={170} label="Task Cron" tone="worker" />
        <DiagramNode x={912} y={520} width={190} label="SFU" detail={labels.sfuDetail} tone="media" />
        {infra.map((label, index) => <DiagramNode key={label} x={48 + index * 225} y={700} width={180} label={label} tone="infra" />)}

        {ingress.map((_, index) => <Edge key={index} d={`M 600 102 C 600 138, ${130 + index * 188} 128, ${130 + index * 188} 164`} marker={topologyMarker("ingress")} kind="ingress" />)}

        <Edge d="M 130 220 C 130 290, 246 280, 246 348" marker={topologyMarker("rpc")} />
        <Edge d="M 318 220 C 318 280, 494 288, 494 348" marker={topologyMarker("rpc")} />
        <Edge d="M 318 220 C 318 270, 246 286, 246 348" marker={topologyMarker("rpc")} />
        <Edge d="M 506 220 C 506 278, 742 280, 742 348" marker={topologyMarker("rpc")} />
        <Edge d="M 506 220 C 506 265, 494 290, 494 348" marker={topologyMarker("rpc")} />
        <Edge d="M 506 220 C 506 250, 246 265, 246 348" marker={topologyMarker("rpc")} />
        <Edge d="M 694 220 C 694 275, 990 278, 990 348" marker={topologyMarker("rpc")} />
        <Edge d="M 694 220 C 694 250, 494 265, 494 348" marker={topologyMarker("rpc")} />
        <Edge d="M 694 220 C 694 242, 246 248, 246 348" marker={topologyMarker("rpc")} />
        <Edge d="M 882 220 C 882 286, 494 292, 494 348" marker={topologyMarker("rpc")} />
        <Edge d="M 1070 220 C 1070 278, 494 278, 494 348" marker={topologyMarker("rpc")} />
        <Edge d="M 404 377 L 338 377" marker={topologyMarker("rpc")} />
        <Edge d="M 652 392 C 620 430, 536 430, 516 406" marker={topologyMarker("rpc")} />

        <Edge d="M 494 406 C 494 455, 193 458, 193 518" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 193 574 C 193 632, 138 640, 138 698" marker={topologyMarker("kafka")} kind="kafka" />
        <Edge d="M 506 220 C 506 610, 138 605, 138 698" marker={topologyMarker("kafka")} kind="kafka" />
        <Edge d="M 694 220 C 694 625, 138 620, 138 698" marker={topologyMarker("kafka")} kind="kafka" />
        <Edge d="M 882 220 C 882 640, 138 635, 138 698" marker={topologyMarker("kafka")} kind="kafka" />
        <Edge d="M 228 727 C 285 727, 295 547, 368 547" marker={topologyMarker("kafka")} kind="kafka" />
        <Edge d="M 455 520 C 455 468, 494 455, 494 406" marker={topologyMarker("rpc")} />
        <Edge d="M 485 520 C 485 448, 742 460, 742 406" marker={topologyMarker("rpc")} />
        <Edge d="M 725 520 C 725 450, 494 462, 494 406" marker={topologyMarker("rpc")} />
        <Edge d="M 455 520 C 455 300, 882 300, 882 222" marker={topologyMarker("ws")} kind="ws" />
        <Edge d="M 1070 220 C 1070 270, 900 270, 900 220" marker={topologyMarker("ws")} kind="ws" />
        <Edge d="M 1070 220 L 1010 518" marker={topologyMarker("media")} kind="media" />

        <Edge d="M 246 406 C 246 630, 363 625, 363 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 494 406 C 494 640, 363 635, 363 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 742 406 C 742 625, 588 625, 588 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 990 406 C 990 640, 363 650, 363 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 882 220 C 882 675, 813 652, 813 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 246 406 C 246 675, 1038 610, 1038 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 494 406 C 494 650, 1038 625, 1038 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 742 406 C 742 625, 1038 640, 1038 698" marker={topologyMarker("ingress")} kind="ingress" />
        <Edge d="M 990 406 L 1038 698" marker={topologyMarker("ingress")} kind="ingress" />

        <Legend labels={labels} y={850} />
      </svg>
    </DiagramFrame>
  );
}

function ServiceDependencies({ locale }: { locale: Locale }) {
  const labels = copy[locale];
  const dependencyMarker = (kind: keyof typeof edgeClasses) => `dependency-${kind}-arrow`;

  return (
    <DiagramFrame title={labels.dependencyTitle} description={labels.dependencyDescription} minWidth={980}>
      <svg viewBox="0 0 1200 760" role="img" aria-labelledby="dependency-title dependency-description" className="block h-auto w-full min-w-[980px]">
        <title id="dependency-title">{labels.dependencyTitle}</title>
        <desc id="dependency-description">{labels.dependencyDescription}</desc>
        <MarkerDefs prefix="dependency" />

        <LayerLabel x={48} y={38}>{labels.ingress}</LayerLabel>
        <LayerLabel x={48} y={318}>{labels.domain}</LayerLabel>
        <LayerLabel x={48} y={610}>{labels.workers}</LayerLabel>

        {[
          [45, "User API"], [230, "Social API"], [415, "IM API"], [600, "Trend API"], [785, "IM WS"], [970, "Streaming"],
        ].map(([x, label]) => <DiagramNode key={label} x={Number(x)} y={62} width={150} label={String(label)} tone={label === "Streaming" ? "media" : "default"} />)}
        {[[145, "User RPC"], [385, "Social RPC"], [625, "IM RPC"], [865, "Trend RPC"]].map(([x, label]) => <DiagramNode key={label} x={Number(x)} y={350} width={170} label={String(label)} tone="rpc" />)}
        <DiagramNode x={60} y={640} width={170} label={labels.relay} tone="worker" />
        <DiagramNode x={315} y={640} width={170} label="Kafka" tone="worker" />
        <DiagramNode x={570} y={640} width={170} label="Task MQ" tone="worker" />
        <DiagramNode x={825} y={640} width={170} label="Task Cron" tone="worker" />
        <DiagramNode x={1030} y={350} width={130} label="SFU" tone="media" />

        <Edge d="M 120 116 C 120 245, 230 245, 230 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 305 116 C 305 235, 470 235, 470 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 305 116 C 305 210, 230 240, 230 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 490 116 C 490 220, 710 220, 710 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 490 116 C 490 245, 470 245, 470 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 490 116 C 490 190, 230 205, 230 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 675 116 C 675 200, 950 210, 950 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 675 116 C 675 235, 470 225, 470 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 675 116 C 675 180, 230 180, 230 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 860 116 C 860 250, 470 260, 470 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 1045 116 C 1045 260, 470 275, 470 348" marker={dependencyMarker("rpc")} />
        <Edge d="M 385 377 L 317 377" marker={dependencyMarker("rpc")} />
        <Edge d="M 625 392 C 590 445, 520 445, 490 406" marker={dependencyMarker("rpc")} />
        <Edge d="M 655 640 C 655 535, 710 510, 710 406" marker={dependencyMarker("rpc")} />
        <Edge d="M 625 640 C 625 515, 470 520, 470 406" marker={dependencyMarker("rpc")} />
        <Edge d="M 910 640 C 910 500, 470 510, 470 406" marker={dependencyMarker("rpc")} />

        <Edge d="M 470 406 C 470 545, 145 535, 145 638" marker={dependencyMarker("ingress")} kind="ingress" />
        <Edge d="M 230 694 L 313 694" marker={dependencyMarker("kafka")} kind="kafka" />
        <Edge d="M 490 116 C 490 590, 400 555, 400 638" marker={dependencyMarker("kafka")} kind="kafka" />
        <Edge d="M 675 116 C 675 570, 400 560, 400 638" marker={dependencyMarker("kafka")} kind="kafka" />
        <Edge d="M 860 116 C 860 550, 400 545, 400 638" marker={dependencyMarker("kafka")} kind="kafka" />
        <Edge d="M 485 694 L 568 694" marker={dependencyMarker("kafka")} kind="kafka" />

        <Edge d="M 655 640 C 655 560, 860 540, 860 118" marker={dependencyMarker("ws")} kind="ws" />
        <Edge d="M 1045 116 C 1045 180, 880 185, 880 118" marker={dependencyMarker("ws")} kind="ws" />
        <Edge d="M 1045 116 L 1095 348" marker={dependencyMarker("media")} kind="media" />

        <Legend labels={labels} y={735} />
      </svg>
    </DiagramFrame>
  );
}

export function ArchitectureDiagrams({ locale }: { locale: Locale }) {
  return (
    <div className="space-y-5">
      <SystemTopology locale={locale} />
      <ServiceDependencies locale={locale} />
    </div>
  );
}
