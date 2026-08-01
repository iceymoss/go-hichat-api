import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight } from "lucide-react";

export const metadata: Metadata = {
  title: "Docs — HiChat",
  description:
    "Understand how HiChat 2.0 is structured, why certain technology choices were made, and where to look when things go wrong.",
};

export default function DocsIntroPage() {
  return (
    <article className="prose-custom">
      <h1>Introduction</h1>

      <p className="lead">
        HiChat 2.0 is a reference implementation of a production-shaped IM
        system, not a toy chat demo. The codebase makes opinionated choices
        about storage, delivery guarantees, and service boundaries — this
        section explains the thinking behind those choices and where to read
        more.
      </p>

      <h2>Why microservices?</h2>

      <p>
        A monolith would be simpler to run, but it would obscure the interesting
        part: the seams between domains. The boundary between{" "}
        <code>im</code> and <code>social</code> exists because chat history and
        the social graph evolve at different rates, have different consistency
        requirements, and benefit from independent deployments. Once those seams
        are explicit — enforced by network calls rather than import paths — the
        tradeoffs become visible and teachable.
      </p>

      <p>
        The rule is strict: a service never reads another service&apos;s
        database tables. If <code>im</code> needs to know whether two users are
        friends before letting them chat, it asks <code>social/rpc</code>, not
        the social DB. That extra hop has a cost, and that cost is intentional.
      </p>

      <h2>Why two storage engines for messages?</h2>

      <p>
        Chat logs go to MongoDB; user accounts, friendships, and groups go to
        MySQL. This is not a default — it reflects different access patterns.
      </p>

      <p>
        A chat log is write-once and queried by <em>range</em>:{" "}
        &ldquo;give me messages in conversation X with seq &gt; N.&rdquo; That
        query is cheap on MongoDB because the document model fits a message
        naturally and the seq field gets a compound index on{" "}
        <code>(conversationId, seq)</code>. Updating a message (recall, read
        state) is a targeted write by ID, also cheap. Social and account data,
        by contrast, is highly relational — friend lists, group membership, role
        checks — and benefits from foreign keys and joins that MongoDB
        deliberately omits.
      </p>

      <h2>Why Kafka instead of direct RPC for delivery?</h2>

      <p>
        When a message arrives at <code>im/ws</code>, it could synchronously
        call <code>im/rpc</code> to persist it and then push to the recipient.
        But synchronous persistence on the hot path adds latency proportional to
        MongoDB write time, and it creates a hard coupling: if the persistence
        call fails, should the sender get an error? The ack has already been
        sent.
      </p>

      <p>
        Kafka separates acknowledgement from persistence. The gateway ACKs the
        sender as soon as the message hits the topic, then{" "}
        <code>task/mq</code> persists it asynchronously. Delivery to the
        recipient also happens through the consumer rather than a synchronous
        push, which means the consumer can retry without bothering the sender
        and without the gateway needing to track outstanding deliveries per
        recipient.
      </p>

      <p>
        The trade-off is eventual consistency: there is a brief window where the
        sender has an ACK but the message is not yet in MongoDB. The system
        accepts that.
      </p>

      <h2>Why go-zero?</h2>

      <p>
        go-zero gives each service a{" "}
        <strong>code-generated skeleton</strong> from a contract file (
        <code>.api</code> for HTTP, <code>.proto</code> for gRPC). That contract
        is the single source of truth for routes, request shapes, and response
        shapes — the generated types and handler stubs never drift from it
        because you re-run <code>goctl</code> every time the contract changes.
        It also provides service discovery through etcd and built-in
        observability middleware, which would otherwise require custom wiring.
      </p>

      <h2>Where to go from here</h2>

      <div className="not-prose mt-6 grid gap-3 sm:grid-cols-2">
        {[
          {
            label: "Message Lifecycle",
            href: "/docs/message-lifecycle",
            desc: "Follow a single chat message from the sender's keyboard to the recipient's screen.",
          },
          {
            label: "Realtime Gateway",
            href: "/docs/realtime-gateway",
            desc: "How the WebSocket server manages connections, heartbeats, and delivery guarantees.",
          },
          {
            label: "Extending the System",
            href: "/docs/extending",
            desc: "Add a new service, endpoint, consumer, or cron job without breaking existing contracts.",
          },
          {
            label: "Quick Start",
            href: "/quick-start",
            desc: "Get the full stack running locally in under a minute.",
          },
        ].map((card) => (
          <Link
            key={card.href}
            href={card.href}
            className="group flex items-start justify-between gap-3 rounded-xl border border-border bg-card p-4 transition-colors hover:border-brand/50"
          >
            <div>
              <p className="text-sm font-medium text-foreground">{card.label}</p>
              <p className="mt-1 text-xs text-muted-foreground">{card.desc}</p>
            </div>
            <ArrowRight className="mt-0.5 size-4 shrink-0 text-muted-foreground transition-colors group-hover:text-brand" />
          </Link>
        ))}
      </div>
    </article>
  );
}
