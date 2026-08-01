import type { Metadata } from "next";
import { AlertTriangle, ExternalLink, Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { CodeBlock } from "@/components/ui/code-block";
import {
  demoAccount,
  links,
  localDevSteps,
  quickStartSteps,
  servicePorts,
} from "@/lib/content";

export const metadata: Metadata = {
  title: "Quick Start — HiChat",
  description:
    "Bring up the full HiChat stack with one Docker Compose command, or run the Go services natively for development.",
};

export default function QuickStartPage() {
  return (
    <>
      {/* Page header */}
      <section className="relative overflow-hidden border-b border-border px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-x-0 top-0 flex justify-center"
        >
          <div className="h-[300px] w-[700px] rounded-full bg-brand opacity-[0.06] blur-[120px]" />
        </div>
        <div className="relative mx-auto max-w-3xl text-center">
          <p className="text-sm font-medium text-brand">Quick Start</p>
          <h1 className="mt-3 text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
            Up and running in a minute
          </h1>
          <p className="mt-5 text-base text-muted-foreground sm:text-lg">
            Docker Compose is the fastest path. Prefer running the Go services
            natively? That path is below too.
          </p>
        </div>
      </section>

      <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        {/* Prerequisites */}
        <div className="rounded-xl border border-border bg-card p-5">
          <h2 className="text-base font-semibold text-foreground">
            Prerequisites
          </h2>
          <ul className="mt-3 flex flex-col gap-2 text-sm text-muted-foreground">
            <li>
              Docker with the Compose plugin (<code className="font-mono text-xs">docker compose version</code>)
            </li>
            <li>Roughly 4 GB of free memory for the full stack</li>
            <li>Ports 2470, 8887–8891, 10090, and 10093 free</li>
          </ul>
        </div>

        {/* Compose steps */}
        <h2 className="mt-12 text-2xl font-bold tracking-tight text-foreground">
          Docker Compose
        </h2>
        <ol className="mt-6 flex flex-col gap-8">
          {quickStartSteps.map((step) => (
            <li key={step.n} className="flex gap-4">
              {/* Step number */}
              <span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-brand/30 bg-brand/10 font-mono text-sm font-medium text-brand">
                {step.n}
              </span>
              <div className="min-w-0 flex-1">
                <h3 className="text-base font-semibold text-foreground">
                  {step.title}
                </h3>
                <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                  {step.body}
                </p>
                <CodeBlock
                  code={step.code}
                  lang={step.lang}
                  compact
                  showLineNumbers={false}
                  className="mt-3"
                />
              </div>
            </li>
          ))}
        </ol>

        {/* Demo account */}
        <div className="mt-10 rounded-xl border border-brand/30 bg-brand/[0.06] p-5">
          <h3 className="text-base font-semibold text-foreground">
            Demo account
          </h3>
          <p className="mt-1.5 text-sm text-muted-foreground">
            Available after seeding the demo dataset in step 3.
          </p>
          <dl className="mt-4 grid gap-3 sm:grid-cols-3">
            <div>
              <dt className="text-xs text-muted-foreground">Phone</dt>
              <dd className="mt-1 font-mono text-sm text-foreground">
                {demoAccount.phone}
              </dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">Password</dt>
              <dd className="mt-1 font-mono text-sm text-foreground">
                {demoAccount.password}
              </dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">URL</dt>
              <dd className="mt-1 font-mono text-sm">
                <a
                  href={demoAccount.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-brand hover:underline"
                >
                  localhost:2470
                </a>
              </dd>
            </div>
          </dl>
          <p className="mt-4 text-xs text-muted-foreground">
            Verification codes auto-fill in demo mode, so you can also register
            a fresh account without an SMS provider.
          </p>
        </div>

        {/* Ports table */}
        <h2 className="mt-14 text-2xl font-bold tracking-tight text-foreground">
          Exposed ports
        </h2>
        <div className="mt-5 overflow-x-auto rounded-xl border border-border bg-card">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-4 py-3 font-medium text-muted-foreground">
                  Service
                </th>
                <th className="px-4 py-3 font-medium text-muted-foreground">
                  Port
                </th>
                <th className="hidden px-4 py-3 font-medium text-muted-foreground sm:table-cell">
                  Purpose
                </th>
              </tr>
            </thead>
            <tbody>
              {servicePorts.map((p) => (
                <tr
                  key={p.service}
                  className="border-b border-border/60 last:border-0"
                >
                  <td className="px-4 py-2.5 font-mono text-sm text-brand">
                    {p.service}
                  </td>
                  <td className="px-4 py-2.5 font-mono text-sm text-foreground">
                    {p.port}
                  </td>
                  <td className="hidden px-4 py-2.5 text-sm text-muted-foreground sm:table-cell">
                    {p.note}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Local development */}
        <h2 className="mt-14 text-2xl font-bold tracking-tight text-foreground">
          Local development
        </h2>
        <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
          For contributors iterating on the Go services, run the middleware in
          Docker and the services on the host.
        </p>

        {/* Secret warning */}
        <div className="mt-5 flex gap-3 rounded-xl border border-border bg-secondary p-4">
          <AlertTriangle className="mt-0.5 size-4 shrink-0 text-brand" />
          <p className="text-sm leading-relaxed text-muted-foreground">
            <span className="font-medium text-foreground">
              Set a dedicated RPC auth secret.
            </span>{" "}
            <code className="font-mono text-xs">HICHAT_IM_RPC_AUTH_SECRET</code>{" "}
            must be at least 32 bytes of random data and must not reuse the JWT
            secret.
          </p>
        </div>

        <div className="mt-6 flex flex-col gap-6">
          {localDevSteps.map((step) => (
            <div key={step.title}>
              <h3 className="text-base font-semibold text-foreground">
                {step.title}
              </h3>
              <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                {step.body}
              </p>
              <CodeBlock
                code={step.code}
                lang={step.lang}
                compact
                showLineNumbers={false}
                className="mt-3"
              />
            </div>
          ))}
        </div>

        {/* Next steps */}
        <h2 className="mt-14 text-2xl font-bold tracking-tight text-foreground">
          Next steps
        </h2>
        <div className="mt-5 grid gap-3 sm:grid-cols-2">
          {[
            { label: "API Reference", href: links.docsApi, desc: "Every REST and gRPC contract" },
            { label: "Developer Guide", href: links.docsDevGuide, desc: "Project layout and conventions" },
            { label: "Docker Deploy", href: links.dockerDeploy, desc: "Reverse proxy, HTTPS, TURN" },
            { label: "Contributing", href: links.contributing, desc: "How to open your first PR" },
          ].map((item) => (
            <a
              key={item.label}
              href={item.href}
              target="_blank"
              rel="noopener noreferrer"
              className="group rounded-xl border border-border bg-card p-4 transition-colors hover:border-brand/50"
            >
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium text-foreground">
                  {item.label}
                </span>
                <ExternalLink className="size-3.5 shrink-0 text-muted-foreground transition-colors group-hover:text-brand" />
              </div>
              <p className="mt-1 text-xs text-muted-foreground">{item.desc}</p>
            </a>
          ))}
        </div>

        {/* Help CTA */}
        <div className="mt-12 rounded-xl border border-border bg-card p-6 text-center">
          <h3 className="text-base font-semibold text-foreground">
            Something not working?
          </h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Open an issue with your Compose logs and we&apos;ll take a look.
          </p>
          <Button variant="outline" size="sm" asChild className="mt-4">
            <a
              href={links.githubIssues}
              target="_blank"
              rel="noopener noreferrer"
              className="gap-1.5"
            >
              <Github className="size-4" />
              Open an issue
            </a>
          </Button>
        </div>
      </div>
    </>
  );
}
