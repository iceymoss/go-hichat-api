import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Extending the System — HiChat Docs",
  description:
    "Add a new HTTP endpoint, RPC method, Kafka consumer, or cron task without breaking existing contracts.",
};

export default function ExtendingPage() {
  return (
    <article className="prose-custom">
      <h1>Extending the System</h1>

      <p className="lead">
        HiChat uses code generation rather than hand-written boilerplate.
        Before adding anything — route, method, consumer, or task — read the
        contract file first. The generator owns the skeleton; you only fill in
        business logic.
      </p>

      <h2>Add an HTTP endpoint</h2>

      <p>
        Every HTTP service has a single <code>.api</code> contract file at{" "}
        <code>apps/&lt;svc&gt;/api/&lt;svc&gt;.api</code>. It defines routes,
        request types, response types, and JWT groups. This file is the source
        of truth — never edit generated files directly.
      </p>

      <h3>1. Declare the route</h3>

      <p>
        Open the <code>.api</code> file and add a route to the appropriate
        group. Optional fields get the <code>optional</code> tag; the
        corresponding Go struct field should use a pointer and{" "}
        <code>omitempty</code>:
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`// In apps/social/api/social.api
type (
  BlockUserReq {
    TargetId string \`json:"targetId"\`
    Note     string \`json:"note,optional"\`
  }
  BlockUserResp {}
)

@server (
  prefix: v1/social
  jwt:    Auth
)
service social-api {
  @handler blockUser
  post /block (BlockUserReq) returns (BlockUserResp)
}`}
      </pre>

      <h3>2. Run goctl</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`goctl api go \\
  -api apps/social/api/social.api \\
  -dir apps/social/api \\
  -style gozero`}
      </pre>

      <p>
        This generates a handler stub and a logic stub. The handler is
        complete — it binds the request and calls the logic. You only edit the
        logic file.
      </p>

      <h3>3. Fill in the logic</h3>

      <p>
        Open the generated file at{" "}
        <code>apps/social/api/internal/logic/blockuserlogic.go</code> and
        implement the method. A few rules that apply everywhere:
      </p>

      <ul>
        <li>
          Call other services through their <code>rpc</code> client in{" "}
          <code>internal/svc/servicecontext.go</code> — never import another
          service&apos;s model package directly.
        </li>
        <li>
          Wrap business errors with <code>pkg/xerr</code>, not raw{" "}
          <code>errors.New</code>. The middleware converts <code>xerr</code>{" "}
          codes to structured JSON responses.
        </li>
        <li>
          Pass <code>ctx</code> from the logic method into every database call
          and RPC call. Never use <code>context.Background()</code> — it
          discards traces and timeouts.
        </li>
      </ul>

      <h2>Add an RPC method</h2>

      <p>
        gRPC contracts live at{" "}
        <code>apps/&lt;svc&gt;/rpc/&lt;svc&gt;.proto</code>. Field numbers are
        permanent — never reuse a number. Methods are append-only; deprecate
        with a comment rather than removing.
      </p>

      <h3>1. Update the proto</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`// In apps/social/rpc/social.proto
message BlockUserReq { string targetId = 1; string actorId = 2; }
message BlockUserResp {}

service Social {
  // … existing methods …
  rpc BlockUser(BlockUserReq) returns (BlockUserResp);
}`}
      </pre>

      <h3>2. Run goctl</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`goctl rpc protoc apps/social/rpc/social.proto \\
  --zrpc_out=apps/social/rpc \\
  --go_out=apps/social/rpc \\
  --go-grpc_out=apps/social/rpc`}
      </pre>

      <p>
        goctl regenerates the server stub and the typed client under{" "}
        <code>apps/social/rpc/socialclient/</code>. The client package is what
        other services import — do not copy-paste RPC types across service
        boundaries.
      </p>

      <h3>3. Fill in the logic and update the client config</h3>

      <p>
        If another service needs to call the new method, add{" "}
        <code>SocialRpc zrpc.RpcClientConf</code> to its config struct (if not
        already present), inject it in <code>servicecontext.go</code>, and add
        the etcd key to its <code>*-sample.yaml</code>.
      </p>

      <h2>Add a Kafka consumer</h2>

      <p>
        All consumers live in{" "}
        <code>apps/task/mq/internal/handler/msg_transfer/</code> and are wired
        in <code>apps/task/mq/internal/handler/listen.go</code>.
      </p>

      <h3>1. Create the consumer</h3>

      <p>
        Implement <code>kq.ConsumeHandler</code>. The only required method is{" "}
        <code>Consume(ctx context.Context, key, value string) error</code>.
        Return a non-nil error to block the Kafka partition and retry. Return
        nil (after writing to the dead-letter topic if needed) for poison
        messages that must not block the partition.
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`type MyEventTransfer struct{ svcCtx *svc.ServiceContext }

func NewMyEventTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
  return &MyEventTransfer{svcCtx: svc}
}

func (m *MyEventTransfer) Consume(ctx context.Context, key, value string) error {
  var in mq.MyEvent
  if err := json.Unmarshal([]byte(value), &in); err != nil {
    // malformed — dead-letter and return nil to unblock the partition
    return m.svcCtx.NotificationDLQ.Publish(ctx, []byte(value))
  }
  // idempotent business logic here
  return nil
}`}
      </pre>

      <h3>2. Register in listen.go</h3>

      <p>
        Add the consumer to the <code>Services()</code> method in{" "}
        <code>listen.go</code>:
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`myHandler := msgTransfer.NewMyEventTransfer(l.svc)
// …
return []service.Service{
  // … existing queues …
  kq.MustNewQueue(l.svc.Config.MyEventTransfer, myHandler),
}`}
      </pre>

      <h3>3. Add the topic config</h3>

      <p>
        Add a <code>MyEventTransfer kq.KqConf</code> field to{" "}
        <code>apps/task/mq/internal/config/config.go</code> and the
        corresponding block to <code>apps/task/mq/etc/mq-sample.yaml</code>.
        The topic name should follow the convention{" "}
        <code>&lt;domain&gt;.&lt;event&gt;.v&lt;n&gt;</code>.
      </p>

      <h3>Idempotency</h3>

      <p>
        Kafka guarantees at-least-once delivery, so every consumer must be
        idempotent. Common patterns used in this codebase:
      </p>

      <ul>
        <li>
          MongoDB upsert keyed on message UUID (used by{" "}
          <code>MsgChatTransfer</code>).
        </li>
        <li>
          Version gate: only apply state if the incoming event&apos;s version
          is higher than the stored version (used by{" "}
          <code>RelationChangeTransfer</code> for group membership changes).
        </li>
        <li>
          Existence check: <code>resp.Inserted</code> from the IM RPC prevents
          double-delivering a notification that was already persisted (used by{" "}
          <code>CommonNotifyTransfer</code>).
        </li>
      </ul>

      <h2>Add a cron task</h2>

      <p>
        Cron tasks live in <code>apps/task/cron/tasks/</code> and are
        registered in <code>tasks/registry.go</code>.
      </p>

      <h3>1. Implement the Task interface</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`type MyCleanupTask struct{ svcCtx *svc.ServiceContext }

func (t *MyCleanupTask) GetName() string         { return "my_cleanup" }
func (t *MyCleanupTask) GetSpec() string         { return t.svcCtx.Config.Cron.MyCleanupSpec }
func (t *MyCleanupTask) GetTimeout() time.Duration { return 5 * time.Minute }
func (t *MyCleanupTask) Execute(ctx context.Context) error {
  // Check ctx.Done() periodically for graceful shutdown
  return nil
}`}
      </pre>

      <h3>2. Register conditionally on config</h3>

      <p>
        In <code>RegisterAllTasks</code>, guard on the spec field being
        non-empty — this allows operators to disable the task by leaving the
        field blank in their yaml:
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`if svc.Config.Cron.MyCleanupSpec != "" {
  registry.RegisterTask(NewMyCleanupTask(svc))
}`}
      </pre>

      <h3>3. Add the config field</h3>

      <p>
        The cron spec always comes from yaml — never hardcode it. Add a field
        to the cron config struct and to <code>apps/task/cron/etc/cron-sample.yaml</code>.
        If the task needs a distributed lock (to avoid running on multiple
        replicas simultaneously), use a Redis lock with a TTL shorter than the
        task&apos;s worst-case runtime.
      </p>

      <div className="not-prose mt-8">
        <Link
          href="/docs/realtime-gateway"
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          ← Realtime Gateway
        </Link>
      </div>
    </article>
  );
}
