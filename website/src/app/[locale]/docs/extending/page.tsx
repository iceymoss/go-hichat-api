import type { Metadata } from "next";
import Link from "next/link";
import { localeHref, resolveLocale } from "@/i18n";

type PageProps = {
  params: Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const locale = resolveLocale((await params).locale);

  return locale === "zh"
    ? {
        title: "扩展系统 - HiChat 文档",
        description:
          "在不破坏现有契约的前提下，新增 HTTP 接口、RPC 方法、Kafka 消费者或定时任务。",
      }
    : {
        title: "Extending the System - HiChat Docs",
        description:
          "Add a new HTTP endpoint, RPC method, Kafka consumer, or cron task without breaking existing contracts.",
      };
}

export default async function ExtendingPage({ params }: PageProps) {
  const locale = resolveLocale((await params).locale);
  const isZh = locale === "zh";

  return (
    <article className="prose-custom">
      <h1>{isZh ? "扩展系统" : "Extending the System"}</h1>

      <p className="lead">
        {isZh
          ? "HiChat 使用代码生成，而不是手写样板代码。新增路由、方法、消费者或任务之前，请先阅读契约文件。生成器负责代码骨架，你只需补充业务逻辑。"
          : "HiChat uses code generation rather than hand-written boilerplate. Before adding anything - route, method, consumer, or task - read the contract file first. The generator owns the skeleton; you only fill in business logic."}
      </p>

      <h2>{isZh ? "新增 HTTP 接口" : "Add an HTTP endpoint"}</h2>

      <p>
        {isZh ? (
          <>
            每个 HTTP 服务都只有一个契约文件：
            <code>apps/&lt;svc&gt;/api/&lt;svc&gt;.api</code>。它定义路由、请求类型、响应类型和
            JWT 分组。该文件是唯一事实来源，切勿直接编辑生成的文件。
          </>
        ) : (
          <>
            Every HTTP service has a single <code>.api</code> contract file at
            <code>apps/&lt;svc&gt;/api/&lt;svc&gt;.api</code>. It defines routes, request
            types, response types, and JWT groups. This file is the source of truth
            - never edit generated files directly.
          </>
        )}
      </p>

      <h3>{isZh ? "1. 声明路由" : "1. Declare the route"}</h3>

      <p>
        {isZh ? (
          <>
            打开 <code>.api</code> 文件，在合适的分组中添加路由。可选字段使用
            <code>optional</code> 标签；对应的 Go 结构体字段应使用指针和
            <code>omitempty</code>：
          </>
        ) : (
          <>
            Open the <code>.api</code> file and add a route to the appropriate group.
            Optional fields get the <code>optional</code> tag; the corresponding Go
            struct field should use a pointer and <code>omitempty</code>:
          </>
        )}
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

      <h3>{isZh ? "2. 运行 goctl" : "2. Run goctl"}</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`goctl api go \\
  -api apps/social/api/social.api \\
  -dir apps/social/api \\
  -style gozero`}
      </pre>

      <p>
        {isZh
          ? "该命令会生成 handler 和 logic 骨架。handler 已完整实现请求绑定和 logic 调用，你只需编辑 logic 文件。"
          : "This generates a handler stub and a logic stub. The handler is complete - it binds the request and calls the logic. You only edit the logic file."}
      </p>

      <h3>{isZh ? "3. 实现业务逻辑" : "3. Fill in the logic"}</h3>

      <p>
        {isZh ? (
          <>
            打开生成的 <code>apps/social/api/internal/logic/blockuserlogic.go</code>
            并实现对应方法。以下规则适用于所有服务：
          </>
        ) : (
          <>
            Open the generated file at
            <code>apps/social/api/internal/logic/blockuserlogic.go</code> and
            implement the method. A few rules that apply everywhere:
          </>
        )}
      </p>

      <ul>
        <li>
          {isZh ? (
            <>
              通过 <code>internal/svc/servicecontext.go</code> 中其他服务的
              <code>rpc</code> 客户端进行跨服务调用，切勿直接导入其他服务的 model 包。
            </>
          ) : (
            <>
              Call other services through their <code>rpc</code> client in
              <code>internal/svc/servicecontext.go</code> - never import another
              service&apos;s model package directly.
            </>
          )}
        </li>
        <li>
          {isZh ? (
            <>
              使用 <code>pkg/xerr</code> 包装业务错误，不要直接使用
              <code>errors.New</code>。中间件会将 <code>xerr</code> 错误码转换为结构化 JSON 响应。
            </>
          ) : (
            <>
              Wrap business errors with <code>pkg/xerr</code>, not raw
              <code>errors.New</code>. The middleware converts <code>xerr</code>
              codes to structured JSON responses.
            </>
          )}
        </li>
        <li>
          {isZh ? (
            <>
              将 logic 方法收到的 <code>ctx</code> 传给每次数据库调用和 RPC 调用。切勿使用
              <code>context.Background()</code>，否则会丢失链路追踪和超时信息。
            </>
          ) : (
            <>
              Pass <code>ctx</code> from the logic method into every database call
              and RPC call. Never use <code>context.Background()</code> - it
              discards traces and timeouts.
            </>
          )}
        </li>
      </ul>

      <h2>{isZh ? "新增 RPC 方法" : "Add an RPC method"}</h2>

      <p>
        {isZh ? (
          <>
            gRPC 契约位于 <code>apps/&lt;svc&gt;/rpc/&lt;svc&gt;.proto</code>。字段编号一经使用
            就应永久保留，切勿复用。方法只能追加；需要废弃时添加注释，不要直接删除。
          </>
        ) : (
          <>
            gRPC contracts live at <code>apps/&lt;svc&gt;/rpc/&lt;svc&gt;.proto</code>.
            Field numbers are permanent - never reuse a number. Methods are
            append-only; deprecate with a comment rather than removing.
          </>
        )}
      </p>

      <h3>{isZh ? "1. 更新 proto" : "1. Update the proto"}</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`// In apps/social/rpc/social.proto
message BlockUserReq { string targetId = 1; string actorId = 2; }
message BlockUserResp {}

service Social {
  // ... existing methods ...
  rpc BlockUser(BlockUserReq) returns (BlockUserResp);
}`}
      </pre>

      <h3>{isZh ? "2. 运行 goctl" : "2. Run goctl"}</h3>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`goctl rpc protoc apps/social/rpc/social.proto \\
  --zrpc_out=apps/social/rpc \\
  --go_out=apps/social/rpc \\
  --go-grpc_out=apps/social/rpc`}
      </pre>

      <p>
        {isZh ? (
          <>
            goctl 会重新生成服务器骨架和 <code>apps/social/rpc/socialclient/</code> 下的
            类型化客户端。其他服务应导入该客户端包，切勿跨服务复制粘贴 RPC 类型。
          </>
        ) : (
          <>
            goctl regenerates the server stub and the typed client under
            <code>apps/social/rpc/socialclient/</code>. The client package is what
            other services import - do not copy-paste RPC types across service boundaries.
          </>
        )}
      </p>

      <h3>
        {isZh ? "3. 实现业务逻辑并更新客户端配置" : "3. Fill in the logic and update the client config"}
      </h3>

      <p>
        {isZh ? (
          <>
            如果其他服务需要调用新方法，请在其配置结构体中添加
            <code>SocialRpc zrpc.RpcClientConf</code>（尚不存在时），在
            <code>servicecontext.go</code> 中完成注入，并将 etcd key 添加到对应的
            <code>*-sample.yaml</code>。
          </>
        ) : (
          <>
            If another service needs to call the new method, add
            <code>SocialRpc zrpc.RpcClientConf</code> to its config struct (if not
            already present), inject it in <code>servicecontext.go</code>, and add
            the etcd key to its <code>*-sample.yaml</code>.
          </>
        )}
      </p>

      <h2>{isZh ? "新增 Kafka 消费者" : "Add a Kafka consumer"}</h2>

      <p>
        {isZh ? (
          <>
            所有消费者都位于 <code>apps/task/mq/internal/handler/msg_transfer/</code>，
            并在 <code>apps/task/mq/internal/handler/listen.go</code> 中完成装配。
          </>
        ) : (
          <>
            All consumers live in <code>apps/task/mq/internal/handler/msg_transfer/</code>
            and are wired in <code>apps/task/mq/internal/handler/listen.go</code>.
          </>
        )}
      </p>

      <h3>{isZh ? "1. 创建消费者" : "1. Create the consumer"}</h3>

      <p>
        {isZh ? (
          <>
            实现 <code>kq.ConsumeHandler</code>。唯一必需的方法是
            <code>Consume(ctx context.Context, key, value string) error</code>。
            返回非 nil 错误会阻塞 Kafka partition 并触发重试。对于不能阻塞 partition 的
            poison message，应在必要时写入 dead-letter topic 后返回 nil。
          </>
        ) : (
          <>
            Implement <code>kq.ConsumeHandler</code>. The only required method is
            <code>Consume(ctx context.Context, key, value string) error</code>.
            Return a non-nil error to block the Kafka partition and retry. Return nil
            (after writing to the dead-letter topic if needed) for poison messages
            that must not block the partition.
          </>
        )}
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`type MyEventTransfer struct{ svcCtx *svc.ServiceContext }

func NewMyEventTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
  return &MyEventTransfer{svcCtx: svc}
}

func (m *MyEventTransfer) Consume(ctx context.Context, key, value string) error {
  var in mq.MyEvent
  if err := json.Unmarshal([]byte(value), &in); err != nil {
    // malformed - dead-letter and return nil to unblock the partition
    return m.svcCtx.NotificationDLQ.Publish(ctx, []byte(value))
  }
  // idempotent business logic here
  return nil
}`}
      </pre>

      <h3>{isZh ? "2. 在 listen.go 中注册" : "2. Register in listen.go"}</h3>

      <p>
        {isZh ? (
          <>将消费者添加到 <code>listen.go</code> 的 <code>Services()</code> 方法中：</>
        ) : (
          <>Add the consumer to the <code>Services()</code> method in <code>listen.go</code>:</>
        )}
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`myHandler := msgTransfer.NewMyEventTransfer(l.svc)
// ...
return []service.Service{
  // ... existing queues ...
  kq.MustNewQueue(l.svc.Config.MyEventTransfer, myHandler),
}`}
      </pre>

      <h3>{isZh ? "3. 添加 topic 配置" : "3. Add the topic config"}</h3>

      <p>
        {isZh ? (
          <>
            在 <code>apps/task/mq/internal/config/config.go</code> 中添加
            <code>MyEventTransfer kq.KqConf</code> 字段，并在
            <code>apps/task/mq/etc/mq-sample.yaml</code> 中添加对应配置块。topic 名称应遵循
            <code>&lt;domain&gt;.&lt;event&gt;.v&lt;n&gt;</code> 约定。
          </>
        ) : (
          <>
            Add a <code>MyEventTransfer kq.KqConf</code> field to
            <code>apps/task/mq/internal/config/config.go</code> and the corresponding
            block to <code>apps/task/mq/etc/mq-sample.yaml</code>. The topic name
            should follow the convention <code>&lt;domain&gt;.&lt;event&gt;.v&lt;n&gt;</code>.
          </>
        )}
      </p>

      <h3>{isZh ? "幂等性" : "Idempotency"}</h3>

      <p>
        {isZh
          ? "Kafka 保证至少一次投递，因此每个消费者都必须具备幂等性。本代码库常用以下模式："
          : "Kafka guarantees at-least-once delivery, so every consumer must be idempotent. Common patterns used in this codebase:"}
      </p>

      <ul>
        <li>
          {isZh ? "以消息 UUID 为键执行 MongoDB upsert（" : "MongoDB upsert keyed on message UUID (used by "}
          <code>MsgChatTransfer</code>{isZh ? " 使用此模式）。" : ")."}
        </li>
        <li>
          {isZh ? (
            <>
              版本门控：仅当传入事件的版本高于已存储版本时更新状态（
              <code>RelationChangeTransfer</code> 使用该模式处理群成员关系变更）。
            </>
          ) : (
            <>
              Version gate: only apply state if the incoming event&apos;s version is
              higher than the stored version (used by <code>RelationChangeTransfer</code>
              for group membership changes).
            </>
          )}
        </li>
        <li>
          {isZh ? (
            <>
              存在性检查：使用 IM RPC 返回的 <code>resp.Inserted</code>，避免重复投递已持久化
              的通知（<code>CommonNotifyTransfer</code> 使用此模式）。
            </>
          ) : (
            <>
              Existence check: <code>resp.Inserted</code> from the IM RPC prevents
              double-delivering a notification that was already persisted (used by
              <code>CommonNotifyTransfer</code>).
            </>
          )}
        </li>
      </ul>

      <h2>{isZh ? "新增定时任务" : "Add a cron task"}</h2>

      <p>
        {isZh ? (
          <>
            定时任务位于 <code>apps/task/cron/tasks/</code>，并在
            <code>tasks/registry.go</code> 中注册。
          </>
        ) : (
          <>
            Cron tasks live in <code>apps/task/cron/tasks/</code> and are registered
            in <code>tasks/registry.go</code>.
          </>
        )}
      </p>

      <h3>{isZh ? "1. 实现 Task 接口" : "1. Implement the Task interface"}</h3>

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

      <h3>{isZh ? "2. 根据配置有条件地注册" : "2. Register conditionally on config"}</h3>

      <p>
        {isZh ? (
          <>
            在 <code>RegisterAllTasks</code> 中检查 spec 字段是否非空。这样，运维人员只需在
            yaml 中将该字段留空即可禁用任务：
          </>
        ) : (
          <>
            In <code>RegisterAllTasks</code>, guard on the spec field being non-empty
            - this allows operators to disable the task by leaving the field blank
            in their yaml:
          </>
        )}
      </p>

      <pre className="not-prose overflow-x-auto rounded-xl border border-border bg-code-bg p-5 font-mono text-sm text-foreground">
        {`if svc.Config.Cron.MyCleanupSpec != "" {
  registry.RegisterTask(NewMyCleanupTask(svc))
}`}
      </pre>

      <h3>{isZh ? "3. 添加配置字段" : "3. Add the config field"}</h3>

      <p>
        {isZh ? (
          <>
            cron spec 必须始终来自 yaml，切勿硬编码。请在 cron 配置结构体和
            <code>apps/task/cron/etc/cron-sample.yaml</code> 中添加字段。如果任务需要分布式锁
            （避免多个副本同时执行），请使用 Redis 锁，并将其 TTL 设置为短于任务的最坏情况
            运行时间。
          </>
        ) : (
          <>
            The cron spec always comes from yaml - never hardcode it. Add a field to
            the cron config struct and to <code>apps/task/cron/etc/cron-sample.yaml</code>.
            If the task needs a distributed lock (to avoid running on multiple
            replicas simultaneously), use a Redis lock with a TTL shorter than the
            task&apos;s worst-case runtime.
          </>
        )}
      </p>

      <div className="not-prose mt-8">
        <Link
          href={localeHref(locale, "/docs/realtime-gateway")}
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          {isZh ? "← 实时网关" : "← Realtime Gateway"}
        </Link>
      </div>
    </article>
  );
}
