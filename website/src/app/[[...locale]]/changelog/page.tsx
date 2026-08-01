import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Changelog — HiChat",
  description: "What changed and when in go-hichat-api.",
};

type Entry = { label: string; text: string; kind: "feat" | "fix" | "refactor" | "docs" | "chore" };
type Release = { date: string; pr?: string; branch?: string; title: string; summary: string; entries: Entry[] };

const releases: Release[] = [
  {
    date: "2026-07-24",
    pr: "253",
    branch: "feat-social-request-reliability",
    title: "Social request reliability",
    summary:
      "End-to-end delivery guarantees for friend and group requests. Requests now survive network interruptions, server restarts, and concurrent state changes.",
    entries: [
      { kind: "feat", label: "social/rpc", text: "Transactional outbox for request notifications; events survive restarts and are never double-delivered." },
      { kind: "feat", label: "social/rpc", text: "Request state machine with explicit terminal states — accepted, declined, expired, and withdrawn — preventing invalid transitions." },
      { kind: "feat", label: "task/mq", text: "Dead-letter topic for notification events that exhaust retries, with structured classification." },
      { kind: "fix",  label: "social/rpc", text: "Race on concurrent accept+decline resolved; only the first actor wins." },
      { kind: "fix",  label: "social/rpc", text: "Group invitation expiration now runs atomically in the cron task." },
      { kind: "fix",  label: "web", text: "Friend request list paginates correctly; stale data is discarded on route change." },
    ],
  },
  {
    date: "2026-07-17",
    pr: "250",
    branch: "feat-streaming-group-sfu-pion",
    title: "Group video calls via Pion SFU",
    summary:
      "The streaming service now routes group video through a server-side selective forwarding unit rather than a full mesh, removing the four-peer hard cap and making bandwidth consumption proportional to the number of subscribers.",
    entries: [
      { kind: "feat", label: "streaming/sfu", text: "Pion-based SFU with a published-track registry and late-joiner backfill." },
      { kind: "feat", label: "streaming/sfu", text: "Active-speaker top-N selection drives audio and video subscription allocation." },
      { kind: "feat", label: "streaming/sfu", text: "Simulcast layer selection: server forwards the appropriate spatial layer per subscriber bandwidth." },
      { kind: "feat", label: "streaming/sfu", text: "Paginated video subscriptions; participants beyond the visible grid do not consume downlink." },
      { kind: "feat", label: "streaming/sfu", text: "Multi-client diagnostic timeline persisted for post-call analysis." },
      { kind: "fix",  label: "streaming/sfu", text: "ICE Lite + mDNS candidate filtering prevent connection failures behind symmetric NAT." },
      { kind: "fix",  label: "streaming/sfu", text: "Renegotiation debounced to coalesce rapid track bursts into a single offer/answer cycle." },
      { kind: "fix",  label: "web/call", text: "Camera track rebuilt correctly after toggle; VP8 forced for concurrent group video." },
    ],
  },
  {
    date: "2026-07-01",
    pr: "248",
    branch: "feat-demo-mockdata-seeder",
    title: "Demo data seeder",
    summary:
      "A one-command seeder registers 14 demo users with pre-populated conversations, friend graphs, and moments — the exact dataset shown in the screenshots.",
    entries: [
      { kind: "feat", label: "scripts/mockdata", text: "Docker Compose mock profile: docker compose --profile mock up seeds demo state idempotently." },
      { kind: "feat", label: "scripts/mockdata", text: "Verification codes auto-fill in demo mode; no SMS provider required." },
    ],
  },
  {
    date: "2026-06-21",
    pr: "244–246",
    branch: "feat-deploy-docker-compose",
    title: "Production-ready Docker Compose stack",
    summary:
      "The full stack — all six microservices, all middleware, and the web client — now starts with a single command. An optional coturn profile adds TURN support for NAT traversal in production.",
    entries: [
      { kind: "feat", label: "deploy", text: "docker-compose.yaml: all services in dependency order with health checks." },
      { kind: "feat", label: "deploy", text: "docker-compose.dependencies.yaml: middleware-only stack for native Go development." },
      { kind: "feat", label: "deploy", text: "coturn service profile with public IP/UDP range config for WebRTC behind NAT." },
      { kind: "feat", label: "deploy", text: "Caddy reverse proxy profile with automatic HTTPS." },
      { kind: "fix",  label: "deploy", text: "Phase 1 SFU TURN config is deployable without manual certificate provisioning." },
    ],
  },
  {
    date: "2026-06-20",
    pr: "243",
    branch: "feat-streaming-call",
    title: "1:1 and full-mesh group calls",
    summary:
      "WebRTC one-on-one calls and full-mesh group calls for up to four participants, with screen sharing and signaling through the streaming service.",
    entries: [
      { kind: "feat", label: "streaming", text: "Call signaling: room creation, join, leave, and ICE candidate exchange." },
      { kind: "feat", label: "streaming", text: "Screen sharing track support in both 1:1 and group calls." },
      { kind: "feat", label: "web/call", text: "Incoming call dialog, call overlay, and mute/camera toggle." },
      { kind: "fix",  label: "web/call", text: "Mute and camera state synchronized correctly across all group participants." },
    ],
  },
  {
    date: "2026-06-12",
    pr: "229–232",
    branch: "feat-im-send-authz-relation-cache",
    title: "IM send authorization and relation cache",
    summary:
      "Kicked or withdrawn members can no longer send messages to groups they have left. Group membership is now cached in Redis for both authorization checks and group fan-out.",
    entries: [
      { kind: "feat", label: "task/mq", text: "Common notification channel: a single consumer handles all system notifications with a dead-letter fallback." },
      { kind: "feat", label: "task/mq", text: "Group send authorization gate backed by Redis relation cache (fail-open on cache miss)." },
      { kind: "feat", label: "task/mq", text: "Relation change events maintain group and friend caches with version gates to prevent stale writes." },
      { kind: "feat", label: "im/ws",  text: "Presence lifecycle hardened: lease refresh, token-based ownership, and 256-shard per-uid lock." },
      { kind: "fix",  label: "im/ws",  text: "WebSocket push broadcast reaches all sessions of a multi-connection user." },
      { kind: "fix",  label: "social", text: "Friend request actor boundaries enforced; one user cannot act on behalf of another." },
    ],
  },
  {
    date: "2026-06-10",
    pr: "224–227",
    branch: "feat-trend-*",
    title: "Activity feed: notifications, image preview, user circles",
    summary:
      "Moments notifications (likes, comments, replies) are now pushed in real time. Images in the feed open in a full-screen lightbox. Friends can be organized into visibility circles.",
    entries: [
      { kind: "feat", label: "trend/rpc", text: "Like and comment events published to Kafka and consumed by TrendNotifyTransfer for real-time push." },
      { kind: "feat", label: "web",       text: "Full-screen image lightbox for moment media." },
      { kind: "feat", label: "trend/rpc", text: "User circle management: create, name, and assign friends for per-circle visibility." },
      { kind: "feat", label: "web",       text: "Moments notifications inbox surfaces unread like and comment events." },
    ],
  },
  {
    date: "2026-06-05",
    pr: "223",
    branch: "feat-trend-api-frontend-integration",
    title: "Trend API / frontend integration",
    summary:
      "The activity feed API and the Next.js client are fully wired together: publish, browse, comment, and react to moments end-to-end.",
    entries: [
      { kind: "feat", label: "trend/api", text: "REST endpoints for publish, list, detail, comment, reply, and like." },
      { kind: "feat", label: "web",       text: "Moments feed, post detail panel, publish flow, and draft support." },
      { kind: "feat", label: "trend/rpc", text: "Visibility filter applied server-side based on the viewer's friend relationship." },
    ],
  },
];

const kindStyle: Record<Entry["kind"], string> = {
  feat:     "bg-brand/10 text-brand border-brand/20",
  fix:      "bg-blue-500/10 text-blue-400 border-blue-400/20",
  refactor: "bg-purple-500/10 text-purple-400 border-purple-400/20",
  docs:     "bg-muted text-muted-foreground border-border",
  chore:    "bg-muted text-muted-foreground border-border",
};

export default function ChangelogPage() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="mb-12">
        <p className="text-sm font-medium text-brand">Changelog</p>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          What&apos;s changed
        </h1>
        <p className="mt-4 text-base text-muted-foreground">
          Major milestones from the git history — each entry corresponds to a
          merged feature branch.
        </p>
      </div>

      <ol className="relative border-l border-border pl-8">
        {releases.map((release, i) => (
          <li key={i} className="mb-12 last:mb-0">
            {/* Timeline dot */}
            <span className="absolute -left-[7px] flex size-3.5 items-center justify-center rounded-full border border-brand/40 bg-brand/20" />

            {/* Date + PR */}
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <time dateTime={release.date}>{release.date}</time>
              {release.pr && (
                <a
                  href={`https://github.com/iceymoss/go-hichat-api/pull/${release.pr.split("–")[0]}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="rounded border border-border bg-secondary px-1.5 py-0.5 font-mono text-xs text-muted-foreground transition-colors hover:text-foreground"
                >
                  #{release.pr}
                </a>
              )}
              {release.branch && (
                <code className="rounded border border-border bg-secondary px-1.5 py-0.5 text-xs text-muted-foreground">
                  {release.branch}
                </code>
              )}
            </div>

            <h2 className="mt-2 text-lg font-semibold text-foreground">
              {release.title}
            </h2>
            <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
              {release.summary}
            </p>

            <ul className="mt-5 flex flex-col gap-2">
              {release.entries.map((entry, j) => (
                <li key={j} className="flex items-start gap-2.5">
                  <span
                    className={`mt-0.5 inline-flex shrink-0 items-center rounded border px-1.5 py-0.5 font-mono text-[10px] font-medium ${kindStyle[entry.kind]}`}
                  >
                    {entry.kind}
                  </span>
                  <span className="rounded border border-border bg-secondary px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground shrink-0">
                    {entry.label}
                  </span>
                  <span className="text-sm leading-relaxed text-muted-foreground">
                    {entry.text}
                  </span>
                </li>
              ))}
            </ul>
          </li>
        ))}
      </ol>

      <div className="mt-12 rounded-xl border border-border bg-card p-5 text-sm text-muted-foreground">
        Full commit history available on{" "}
        <a
          href="https://github.com/iceymoss/go-hichat-api/commits/main"
          target="_blank"
          rel="noopener noreferrer"
          className="text-brand hover:underline"
        >
          GitHub
        </a>
        .
      </div>
    </div>
  );
}
