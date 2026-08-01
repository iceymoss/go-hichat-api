import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Message Lifecycle — HiChat Docs",
  description:
    "The complete path of a chat message from sender to recipient, including ACK, Kafka, persistence, and offline delivery.",
};

export default function MessageLifecyclePage() {
  return (
    <article className="prose-custom">
      <h1>Message Lifecycle</h1>

      <p className="lead">
        A single chat message crosses four distinct systems before it reaches
        the recipient. Understanding that path — and the guarantees at each
        hand-off — is the key to reasoning about delivery failures, latency, and
        eventual consistency in HiChat.
      </p>

      <h2>The happy path</h2>

      <p>
        The sender types a message and the web client publishes it over the
        existing WebSocket connection as a JSON frame with{" "}
        <code>method: &quot;chat.user&quot;</code> (single chat) or{" "}
        <code>&quot;chat.group&quot;</code> (group chat). At this point nothing
        has been persisted — the frame is in flight.
      </p>

      <h3>Step 1 — ACK negotiation (im/ws)</h3>

      <p>
        <code>im/ws</code> runs the WebSocket server. When the frame arrives,
        the server&apos;s ACK mode determines what happens before the message is
        passed to any handler.
      </p>

      <p>
        The server supports three ACK modes, set at startup via{" "}
        <code>WithAck</code>:
      </p>

      <ul>
        <li>
          <strong>NoAck</strong> — the frame goes directly to the handler.
          Fastest, no delivery guarantee.
        </li>
        <li>
          <strong>OnlyAck</strong> — the server immediately sends an ACK frame
          back with <code>ackSeq = msgSeq + 1</code>, then passes the message
          to the handler. Two round-trips, server-initiated confirmation.
        </li>
        <li>
          <strong>RigorAck</strong> — full three-way handshake. The server
          sends ACK seq 1; the client must reply with ACK seq 2 before the
          message is passed to the handler. If the client&apos;s reply doesn&apos;t
          arrive within <code>ackTimeout</code>, the message is abandoned.
          Three round-trips, mutual confirmation.
        </li>
      </ul>

      <p>
        Regardless of mode, every message has a UUID <code>Id</code>. The server
        deduplicates by <code>Id</code> inside the per-connection{" "}
        <code>readMessageSeq</code> map: if the same ID arrives twice with an
        equal or lower <code>AckSeq</code>, the second copy is silently dropped.
      </p>

      <h3>Step 2 — Kafka publish (im/ws → Kafka)</h3>

      <p>
        Once ACK is complete, the <code>chat.user</code> handler publishes the
        message to the Kafka topic <code>im.message.sent.v1</code>. The
        handler returns immediately after the publish succeeds — it does not
        wait for MongoDB. The sender&apos;s ACK was sent in step 1, so from the
        sender&apos;s perspective the message is &ldquo;delivered&rdquo; at this
        point, even though it is not yet stored.
      </p>

      <p>
        This is the system&apos;s intentional consistency trade-off: if the broker
        accepts the message but the consumer crashes before writing to MongoDB,
        the message can be replayed from Kafka (within the topic&apos;s retention
        window) but will not appear in history queries until the consumer
        catches up.
      </p>

      <h3>Step 3 — Persistence (task/mq)</h3>

      <p>
        <code>task/mq</code> runs a consumer group against{" "}
        <code>im.message.sent.v1</code>. The consumer that handles this topic
        is <code>MsgChatTransfer</code>, which extends{" "}
        <code>BaseMsgChatTransfer</code>.
      </p>

      <p>The consumer does three things in order:</p>

      <ol>
        <li>
          Writes the message document to MongoDB under the{" "}
          <code>chatLog</code> collection, with a compound index on{" "}
          <code>(conversationId, seq)</code>.
        </li>
        <li>
          Updates the conversation&apos;s <code>lastMsg</code> and{" "}
          <code>seq</code> fields in MySQL (the im service owns this table).
        </li>
        <li>
          Calls <code>im/ws</code> to push the message to the recipient&apos;s
          active connection(s), if any exist on this node.
        </li>
      </ol>

      <p>
        The consumer is idempotent: the MongoDB write uses{" "}
        <code>upsert</code> keyed on the message UUID, so replaying a Kafka
        message after a crash does not create a duplicate document.
      </p>

      <h3>Step 4 — Push or offline storage</h3>

      <p>
        If the recipient has an active WebSocket connection,{" "}
        <code>im/ws</code> writes the message frame directly to the connection.
        If not, the message is already in MongoDB — when the recipient connects,
        the client sends its last known <code>seq</code> per conversation and
        the server returns all messages with a higher seq.
      </p>

      <p>
        Cross-node delivery — where the recipient is connected to a different{" "}
        <code>im/ws</code> instance than the one receiving the consumer callback
        — routes through a second Kafka topic so the correct node can push.
      </p>

      <h2>The read receipt path</h2>

      <p>
        When the recipient&apos;s client renders a message, it publishes a read
        event to <code>im.read.v1</code>. The consumer{" "}
        <code>MsgReadTransfer</code> in <code>task/mq</code> updates the read
        record in MongoDB and, if the original sender is online, pushes a read
        receipt back through the WebSocket gateway. The sender&apos;s client then
        updates the checkmark state locally.
      </p>

      <h2>The recall path</h2>

      <p>
        A recall request goes to <code>im/api</code>, which validates
        permissions (only the original sender within a configurable time window)
        and publishes to <code>im.recall.v1</code>. The consumer{" "}
        <code>MsgRecallTransfer</code> marks the MongoDB document as recalled
        and pushes a recall notification to all participants in the conversation
        who are currently online. Clients that receive the notification remove
        or replace the message in their local render state.
      </p>

      <h2>What can go wrong, and where</h2>

      <table>
        <thead>
          <tr>
            <th>Failure point</th>
            <th>Effect</th>
            <th>Recovery</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>im/ws crashes after ACK, before Kafka publish</td>
            <td>Message lost. Sender has an ACK but the message never enters the pipeline.</td>
            <td>
              RigorAck makes this window smaller — the client has confirmed
              receipt before the handler runs. In practice the window is the
              time between ACK completion and the Kafka publish call, typically
              under 1 ms.
            </td>
          </tr>
          <tr>
            <td>Kafka broker unavailable</td>
            <td>
              Publish fails; the handler returns an error to the caller.
            </td>
            <td>
              The sender&apos;s client should surface a delivery failure and offer
              retry.
            </td>
          </tr>
          <tr>
            <td>task/mq consumer crashes mid-write</td>
            <td>
              MongoDB write is incomplete; Kafka offset is not committed.
            </td>
            <td>
              On restart, the consumer re-reads the uncommitted offset and
              retries. The upsert on message UUID ensures no duplicate.
            </td>
          </tr>
          <tr>
            <td>Recipient offline at push time</td>
            <td>Push silently skipped — no error.</td>
            <td>
              Message is in MongoDB. Client fetches on next connect via seq
              pull.
            </td>
          </tr>
        </tbody>
      </table>

      <div className="not-prose mt-8 flex gap-3">
        <Link
          href="/docs/realtime-gateway"
          className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-brand/50 hover:text-foreground"
        >
          Next: Realtime Gateway →
        </Link>
      </div>
    </article>
  );
}
