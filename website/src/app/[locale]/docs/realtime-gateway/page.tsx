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
        title: "实时网关 - HiChat 文档",
        description:
          "了解 im/ws 如何管理连接生命周期、心跳、ACK 模式，以及基于 Redis 的多节点在线状态。",
      }
    : {
        title: "Realtime Gateway - HiChat Docs",
        description:
          "How im/ws manages connection lifecycle, heartbeats, ACK modes, and Redis-backed presence across multiple nodes.",
      };
}

export default async function RealtimeGatewayPage({ params }: PageProps) {
  const locale = resolveLocale((await params).locale);
  const isZh = locale === "zh";

  return (
    <article className="prose-custom">
      <h1>{isZh ? "实时网关" : "Realtime Gateway"}</h1>

      <p className="lead">
        {isZh ? (
          <>
            <code>im/ws</code> 是 WebSocket 服务器。发送给在线用户的每一条聊天消息、
            已读回执、通知和关系变更事件都会经过它。理解其连接管理和投递保障机制，
            是参与实时链路相关开发的前提。
          </>
        ) : (
          <>
            <code>im/ws</code> is the WebSocket server. Every chat message, read
            receipt, notification, and relation-change event that reaches an online
            user passes through it. Understanding how it manages connections and
            guarantees delivery is prerequisite knowledge for any work that touches
            the live path.
          </>
        )}
      </p>

      <h2>{isZh ? "连接生命周期" : "Connection lifecycle"}</h2>

      <p>
        {isZh
          ? "客户端发起 WebSocket 连接时，服务器会在升级 HTTP 请求之前校验 JWT。校验失败将返回 401，且不会建立 WebSocket 连接。"
          : "When a client opens a WebSocket connection, the server performs a JWT check before upgrading the HTTP request. A failed check returns 401 and no WebSocket connection is created."}
      </p>

      <p>
        {isZh ? (
          <>
            协议升级后，服务器会分配一个 <code>Conn</code> 结构体，其中保存 gorilla
            WebSocket 连接，以及 ACK 跟踪和空闲检测所需的状态。每个连接会立即启动两个
            goroutine：
          </>
        ) : (
          <>
            After the upgrade, a <code>Conn</code> struct is allocated. It holds the
            gorilla WebSocket connection plus the state needed for ACK tracking and
            idle detection. Two goroutines start immediately per connection:
          </>
        )}
      </p>

      <ul>
        <li>
          {isZh ? (
            <>
              <strong>keepalive goroutine</strong>：运行空闲计时器。客户端响应 WebSocket
              ping 时会调用 Pong 处理器，并重置 <code>idle</code> 时间戳。如果超过
              <code>maxConnectionIdle</code> 仍未收到 Pong，连接将被关闭。默认不限制最长
              空闲时间；需要清理失效连接的部署应通过 <code>WithMaxConnectionIdle</code>
              显式设置该值。
            </>
          ) : (
            <>
              <strong>keepalive goroutine</strong> - runs an idle timer. The Pong
              handler (called when the client responds to a WebSocket ping) resets
              the <code>idle</code> timestamp. When <code>maxConnectionIdle</code>
              elapses without a Pong, the connection is closed. The default is no
              maximum - deployments that need ghost connection cleanup should set
              this explicitly via <code>WithMaxConnectionIdle</code>.
            </>
          )}
        </li>
        <li>
          {isZh ? (
            <>
              <strong>handleWrite goroutine</strong>：消费每个连接的 <code>message</code>
              channel。某个 frame 完成 ACK 协商后会被推送到这里，再分发给已注册的路由
              处理器，最后清理 ACK 状态。
            </>
          ) : (
            <>
              <strong>handleWrite goroutine</strong> - drains the per-connection
              <code>message</code> channel. After ACK negotiation completes for a
              frame, the frame is pushed here, dispatched to the registered route
              handler, and the ACK state is cleaned up.
            </>
          )}
        </li>
      </ul>

      <p>
        {isZh ? (
          <>如果 ACK 模式不是 <code>NoAck</code>，还会启动第三个 goroutine：</>
        ) : (
          <>If the ACK mode is not <code>NoAck</code>, a third goroutine starts:</>
        )}
      </p>

      <ul>
        <li>
          {isZh ? (
            <>
              <strong>readAck goroutine</strong>：轮询每个连接的消息队列，向客户端发送
              ACK frame 并等待确认。它使用紧凑的轮询休眠循环，而不是基于 channel 的
              等待机制（队列为空时休眠 100 ms），因此即使空闲也会占用一个 goroutine。
              正常关闭时，<code>done</code> channel 会被关闭，三个 goroutine 随即全部退出。
            </>
          ) : (
            <>
              <strong>readAck goroutine</strong> - spins over the per-connection
              message queue, sending ACK frames to the client and waiting for
              confirmation. It runs a tight spin-sleep loop (100 ms sleep when the
              queue is empty) rather than a channel-based wait, which means it holds
              a goroutine even when idle. On graceful close, the <code>done</code>
              channel is closed and all three goroutines exit.
            </>
          )}
        </li>
      </ul>

      <p>
        {isZh ? (
          <>
            一个用户可以同时建立多个连接，例如同时登录手机端和桌面端。服务器通过两个由
            同一个 <code>sync.RWMutex</code> 保护的 map，维护
            <code>uid → set of *Conn</code> 和 <code>*Conn → uid</code> 的映射。在线状态更新
            还会经过一个按 uid 分片的 256 分片互斥锁（基于 uid 字符串进行哈希），避免同一
            用户的多个连接并发竞争在线状态租约。
          </>
        ) : (
          <>
            One user can have multiple simultaneous connections - a phone and a
            desktop open at the same time, for example. The server maps
            <code>uid → set of *Conn</code> and <code>*Conn → uid</code> in two maps
            protected by a single <code>sync.RWMutex</code>. Presence updates are
            also gated on a 256-shard per-uid mutex (hashed from the uid string) so
            multiple connections for the same user do not race on the presence lease.
          </>
        )}
      </p>

      <h2>{isZh ? "ACK 模式详解" : "ACK modes in detail"}</h2>

      <p>
        {isZh
          ? "ACK 模式是服务器启动时设置的全局选项。同一服务器实例上的所有连接共享相同模式。"
          : "The ACK mode is a server-level option set at startup. All connections on a given server instance share the same mode."}
      </p>

      <table>
        <thead>
          <tr>
            <th>{isZh ? "模式" : "Mode"}</th>
            <th>{isZh ? "往返次数" : "Round trips"}</th>
            <th>{isZh ? "确认内容" : "What it confirms"}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>NoAck</code></td>
            <td>0</td>
            <td>
              {isZh
                ? "frame 已写入服务器的 TCP 缓冲区，不提供应用层确认。"
                : "Frame delivered to the server's TCP buffer. No application-level confirmation."}
            </td>
          </tr>
          <tr>
            <td><code>OnlyAck</code></td>
            <td>2</td>
            <td>
              {isZh
                ? "服务器已接收并解析 frame，随后发送 ACK；客户端无需确认该 ACK。发送方可确认服务器已经收到消息。"
                : "Server received and parsed the frame, then ACKed. Client does not confirm the ACK. Sender knows the server has the message."}
            </td>
          </tr>
          <tr>
            <td><code>RigorAck</code></td>
            <td>3</td>
            <td>
              {isZh ? (
                <>
                  服务器发送 ACK seq 1，客户端发送 ACK seq 2。服务器确认客户端的 ACK 后，
                  才将 frame 交给处理器，实现双向确认。<code>WithAuthentication</code> 设置的
                  就是此模式。
                </>
              ) : (
                <>
                  Server sends ACK seq 1; client sends ACK seq 2. Server confirms
                  the client&apos;s ACK before passing the frame to the handler. Mutual
                  confirmation. This is the mode set by <code>WithAuthentication</code>.
                </>
              )}
            </td>
          </tr>
        </tbody>
      </table>

      <p>
        {isZh ? (
          <>
            在 <code>RigorAck</code> 模式下，如果服务器未在 <code>ackTimeout</code>
            （默认 30 秒）内收到客户端确认，就会放弃该消息：消息会从队列移除，不再交给
            处理器。系统还设有 <code>sendErrCount</code> 上限（默认 5 次）；如果服务器写入
            ACK frame 连续失败达到该次数，无论是否超时，都会放弃消息。
          </>
        ) : (
          <>
            For <code>RigorAck</code>, if the client&apos;s confirmation does not arrive
            within <code>ackTimeout</code> (default 30 s), the server abandons the
            message - it is removed from the queue and not passed to the handler.
            There is also a <code>sendErrCount</code> limit (default 5): if the
            server fails to write the ACK frame that many times, the message is
            abandoned regardless of timeout.
          </>
        )}
      </p>

      <p>
        {isZh ? (
          <>
            当前实现存在一个已知边界情况：如果客户端发送的第一个 frame 使用
            <code>AckSeq = 2</code>（跳过 seq 0 和 1），<code>readAck</code> goroutine
            会进入持续等待一个永远不会到达的确认状态，从而锁住该连接的消息队列。
            <code>handlerConn</code> 中的注释已将其标记为 TODO。实际使用中，行为正确的
            客户端始终从 seq 0 开始。
          </>
        ) : (
          <>
            There is a known edge case in the current implementation: if a client
            sends its first frame with <code>AckSeq = 2</code> (skipping seq 0 and
            1), the <code>readAck</code> goroutine enters a state where it
            perpetually waits for a confirmation that never comes, locking the
            connection&apos;s message queue. The comment in <code>handlerConn</code>
            notes this as a TODO. In practice, correctly behaved clients always
            start at seq 0.
          </>
        )}
      </p>

      <h2>{isZh ? "在线状态" : "Presence"}</h2>

      <p>
        {isZh ? (
          <>
            用户的<em>第一个</em>连接建立时，服务器会通过 <code>presence.Store</code>
            接口在 Redis 中获取在线状态租约。租约除节点 ID 外，还保存一个随机生成的
            16 字节 token，以便节点区分自己持有的租约与崩溃节点遗留的过期租约。
          </>
        ) : (
          <>
            When the <em>first</em> connection for a user is established, the server
            claims a presence lease in Redis via the <code>presence.Store</code>
            interface. The lease stores a randomly generated 16-byte token alongside
            the node ID, so a node can distinguish its own lease from a stale one
            left by a crashed peer.
          </>
        )}
      </p>

      <p>
        {isZh ? (
          <>
            后台 goroutine 每隔 <code>presenceRefresh</code>（默认 2 分钟）刷新租约，
            早于 TTL（默认 5 分钟）到期时间。用户的<em>最后一个</em>连接关闭时，租约
            goroutine 会停止并删除租约，但前提是该节点仍持有租约。如果 Redis key 已被
            其他节点覆盖（用户已在别处重新连接），则跳过删除操作。
          </>
        ) : (
          <>
            A background goroutine refreshes the lease every
            <code>presenceRefresh</code> (default 2 min) before the TTL (default
            5 min) expires. When the <em>last</em> connection for a user closes, the
            lease goroutine stops and the lease is deleted - but only if the node
            still owns it. If the Redis key has already been overwritten by a
            different node (the user reconnected elsewhere), the delete is skipped.
          </>
        )}
      </p>

      <p>
        {isZh ? (
          <>
            需要判断用户是否在线的下游服务会直接查询该 Redis key，而不是通过 RPC
            请求 <code>im/ws</code>。其值为节点 ID，因此需要向该用户推送消息的消费者
            可以据此确定应联系哪个节点。
          </>
        ) : (
          <>
            Consuming services that need to check whether a user is online query this
            Redis key directly rather than asking <code>im/ws</code> over RPC. The
            value is the node ID, so a consumer that needs to push to that user knows
            which node to contact.
          </>
        )}
      </p>

      <h2>{isZh ? "消息路由" : "Message routing"}</h2>

      <p>
        {isZh ? (
          <>
            ACK 完成后（<code>NoAck</code> 模式则立即），frame 会进入连接的
            <code>message</code> channel，该 channel 的缓冲容量为 1。
            <code>handleWrite</code> goroutine 取出消息，并根据 <code>message.Method</code>
            进行分发：
          </>
        ) : (
          <>
            After ACK (or immediately for <code>NoAck</code>), the frame lands on
            the connection&apos;s <code>message</code> channel, buffered at capacity 1.
            The <code>handleWrite</code> goroutine picks it up and dispatches on
            <code>message.Method</code>:
          </>
        )}
      </p>

      <ul>
        <li>
          <code>FramePing</code>
          {isZh ? "：直接回送为 Ping frame，不查询路由。" : " - echoed back as a Ping frame. No route lookup."}
        </li>
        <li>
          <code>FrameData</code> {isZh ? "或" : "or"} <code>FrameNoAck</code>
          {isZh
            ? "：在 routes map 中查找。如果未注册处理器，服务器会返回一个 error frame。"
            : " - looked up in the routes map. If no handler is registered, the server sends an error frame back."}
        </li>
      </ul>

      <p>
        {isZh ? (
          <>
            路由处理器在 <code>handleWrite</code> 中同步执行。处理器执行缓慢会阻塞该连接的
            消息处理。对于数据库写入、RPC 调用等可能耗时的工作，处理器应将任务转交给
            <code>TaskRunner</code> goroutine；可通过 <code>Server.TaskRunner</code> 使用它。
          </>
        ) : (
          <>
            Route handlers run synchronously inside <code>handleWrite</code>. A slow
            handler blocks message processing for that connection. For work that may
            take time (database writes, RPC calls), the handler should hand off to a
            <code>TaskRunner</code> goroutine, available via <code>Server.TaskRunner</code>.
          </>
        )}
      </p>

      <h2>{isZh ? "出站推送" : "Outbound push"}</h2>

      <p>
        {isZh ? (
          <>
            <code>task/mq</code> 的消费者侧以 WebSocket <em>客户端</em>身份连接
            <code>im/ws</code>，而不是作为服务器侧连接。它使用存储在 Redis 中的系统 root
            token 完成认证，并调用 <code>client.Send</code> 推送 frame。<code>Send</code> 会将
            数据序列化为 JSON、获取写互斥锁，然后调用 <code>WriteMessage</code>。写入失败时，
            客户端会尝试重连一次再返回错误，但不会无限重试。因此，<code>task/mq</code> 的
            推送失败后只会记录日志并丢弃，不会在消费者层重试。
          </>
        ) : (
          <>
            The consumer side of <code>task/mq</code> connects to <code>im/ws</code>
            as a WebSocket <em>client</em> (not as a server-side connection). It
            authenticates with a system root token stored in Redis and calls
            <code>client.Send</code> to push frames. <code>Send</code> serialises to
            JSON, acquires a write mutex, and calls <code>WriteMessage</code>. If the
            write fails, the client attempts one reconnect before returning the error
            - but it does not retry indefinitely, so a push from <code>task/mq</code>
            that fails is logged and dropped rather than retried from the consumer level.
          </>
        )}
      </p>

      <div className="not-prose mt-8 flex flex-wrap gap-3">
        <Link
          href={localeHref(locale, "/docs/message-lifecycle")}
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          {isZh ? "← 消息生命周期" : "← Message Lifecycle"}
        </Link>
        <Link
          href={localeHref(locale, "/docs/extending")}
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          {isZh ? "下一篇：扩展系统 →" : "Next: Extending the System →"}
        </Link>
      </div>
    </article>
  );
}
