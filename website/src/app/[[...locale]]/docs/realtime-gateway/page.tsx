import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Realtime Gateway — HiChat Docs",
  description:
    "How im/ws manages connection lifecycle, heartbeats, ACK modes, and Redis-backed presence across multiple nodes.",
};

export default function RealtimeGatewayPage() {
  return (
    <article className="prose-custom">
      <h1>Realtime Gateway</h1>

      <p className="lead">
        <code>im/ws</code> is the WebSocket server. Every chat message, read
        receipt, notification, and relation-change event that reaches an online
        user passes through it. Understanding how it manages connections and
        guarantees delivery is prerequisite knowledge for any work that touches
        the live path.
      </p>

      <h2>Connection lifecycle</h2>

      <p>
        When a client opens a WebSocket connection, the server performs a JWT
        check before upgrading the HTTP request. A failed check returns 401 and
        no WebSocket connection is created.
      </p>

      <p>
        After the upgrade, a <code>Conn</code> struct is allocated. It holds the
        gorilla WebSocket connection plus the state needed for ACK tracking and
        idle detection. Two goroutines start immediately per connection:
      </p>

      <ul>
        <li>
          <strong>keepalive goroutine</strong> — runs an idle timer. The Pong
          handler (called when the client responds to a WebSocket ping) resets
          the{" "}
          <code>idle</code> timestamp. When{" "}
          <code>maxConnectionIdle</code> elapses without a Pong, the connection
          is closed. The default is no maximum — deployments that need ghost
          connection cleanup should set this explicitly via{" "}
          <code>WithMaxConnectionIdle</code>.
        </li>
        <li>
          <strong>handleWrite goroutine</strong> — drains the per-connection{" "}
          <code>message</code> channel. After ACK negotiation completes for a
          frame, the frame is pushed here, dispatched to the registered route
          handler, and the ACK state is cleaned up.
        </li>
      </ul>

      <p>
        If the ACK mode is not <code>NoAck</code>, a third goroutine starts:
      </p>

      <ul>
        <li>
          <strong>readAck goroutine</strong> — spins over the per-connection
          message queue, sending ACK frames to the client and waiting for
          confirmation. It runs a tight spin-sleep loop (100 ms sleep when the
          queue is empty) rather than a channel-based wait, which means it holds
          a goroutine even when idle. On graceful close, the{" "}
          <code>done</code> channel is closed and all three goroutines exit.
        </li>
      </ul>

      <p>
        One user can have multiple simultaneous connections — a phone and a
        desktop open at the same time, for example. The server maps{" "}
        <code>uid → set of *Conn</code> and{" "}
        <code>*Conn → uid</code> in two maps protected by a single{" "}
        <code>sync.RWMutex</code>. Presence updates are also gated on a
        256-shard per-uid mutex (hashed from the uid string) so multiple
        connections for the same user do not race on the presence lease.
      </p>

      <h2>ACK modes in detail</h2>

      <p>
        The ACK mode is a server-level option set at startup. All connections on
        a given server instance share the same mode.
      </p>

      <table>
        <thead>
          <tr>
            <th>Mode</th>
            <th>Round trips</th>
            <th>What it confirms</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <code>NoAck</code>
            </td>
            <td>0</td>
            <td>
              Frame delivered to the server&apos;s TCP buffer. No application-level
              confirmation.
            </td>
          </tr>
          <tr>
            <td>
              <code>OnlyAck</code>
            </td>
            <td>2</td>
            <td>
              Server received and parsed the frame, then ACKed. Client does not
              confirm the ACK. Sender knows the server has the message.
            </td>
          </tr>
          <tr>
            <td>
              <code>RigorAck</code>
            </td>
            <td>3</td>
            <td>
              Server sends ACK seq 1; client sends ACK seq 2. Server confirms
              the client&apos;s ACK before passing the frame to the handler. Mutual
              confirmation. This is the mode set by{" "}
              <code>WithAuthentication</code>.
            </td>
          </tr>
        </tbody>
      </table>

      <p>
        For <code>RigorAck</code>, if the client&apos;s confirmation does not arrive
        within <code>ackTimeout</code> (default 30 s), the server abandons the
        message — it is removed from the queue and not passed to the handler.
        There is also a <code>sendErrCount</code> limit (default 5): if the
        server fails to write the ACK frame that many times, the message is
        abandoned regardless of timeout.
      </p>

      <p>
        There is a known edge case in the current implementation: if a client
        sends its first frame with <code>AckSeq = 2</code> (skipping seq 0 and
        1), the <code>readAck</code> goroutine enters a state where it
        perpetually waits for a confirmation that never comes, locking the
        connection&apos;s message queue. The comment in{" "}
        <code>handlerConn</code> notes this as a TODO. In practice, correctly
        behaved clients always start at seq 0.
      </p>

      <h2>Presence</h2>

      <p>
        When the <em>first</em> connection for a user is established, the server
        claims a presence lease in Redis via the{" "}
        <code>presence.Store</code> interface. The lease stores a randomly
        generated 16-byte token alongside the node ID, so a node can distinguish
        its own lease from a stale one left by a crashed peer.
      </p>

      <p>
        A background goroutine refreshes the lease every{" "}
        <code>presenceRefresh</code> (default 2 min) before the TTL (default
        5 min) expires. When the <em>last</em> connection for a user closes, the
        lease goroutine stops and the lease is deleted — but only if the node
        still owns it. If the Redis key has already been overwritten by a
        different node (the user reconnected elsewhere), the delete is skipped.
      </p>

      <p>
        Consuming services that need to check whether a user is online query this
        Redis key directly rather than asking <code>im/ws</code> over RPC. The
        value is the node ID, so a consumer that needs to push to that user
        knows which node to contact.
      </p>

      <h2>Message routing</h2>

      <p>
        After ACK (or immediately for <code>NoAck</code>), the frame lands on
        the connection&apos;s <code>message</code> channel, buffered at capacity 1.
        The <code>handleWrite</code> goroutine picks it up and dispatches on{" "}
        <code>message.Method</code>:
      </p>

      <ul>
        <li>
          <code>FramePing</code> — echoed back as a Ping frame. No route lookup.
        </li>
        <li>
          <code>FrameData</code> or <code>FrameNoAck</code> — looked up in the
          routes map. If no handler is registered, the server sends an error
          frame back.
        </li>
      </ul>

      <p>
        Route handlers run synchronously inside <code>handleWrite</code>. A
        slow handler blocks message processing for that connection. For work that
        may take time (database writes, RPC calls), the handler should hand off
        to a <code>TaskRunner</code> goroutine, available via{" "}
        <code>Server.TaskRunner</code>.
      </p>

      <h2>Outbound push</h2>

      <p>
        The consumer side of <code>task/mq</code> connects to{" "}
        <code>im/ws</code> as a WebSocket <em>client</em> (not as a server-side
        connection). It authenticates with a system root token stored in Redis
        and calls <code>client.Send</code> to push frames. <code>Send</code>{" "}
        serialises to JSON, acquires a write mutex, and calls{" "}
        <code>WriteMessage</code>. If the write fails, the client attempts one
        reconnect before returning the error — but it does not retry indefinitely,
        so a push from <code>task/mq</code> that fails is logged and dropped
        rather than retried from the consumer level.
      </p>

      <div className="not-prose mt-8 flex flex-wrap gap-3">
        <Link
          href="/docs/message-lifecycle"
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          ← Message Lifecycle
        </Link>
        <Link
          href="/docs/extending"
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          Next: Extending the System →
        </Link>
      </div>
    </article>
  );
}
