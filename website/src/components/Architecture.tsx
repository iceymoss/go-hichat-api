import { ArrowDown } from "lucide-react";
import type { SiteContent } from "@/i18n";

export function Architecture({ content }: { content: SiteContent }) {
  const { architectureLayers, services, techStack, ui } = content;

  return (
    <section
      id="architecture"
      className="mx-auto max-w-7xl px-4 py-20 sm:px-6 sm:py-24 lg:px-8"
    >
      <div className="mx-auto max-w-2xl text-center">
        <p className="text-sm font-medium text-brand">{ui.archEyebrow}</p>
        <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          {ui.archTitle}
        </h2>
        <p className="mt-4 text-base text-muted-foreground">{ui.archSubtitle}</p>
      </div>

      <div className="mt-14 grid gap-10 lg:grid-cols-5">
        {/* Layer stack */}
        <div className="lg:col-span-3">
          <ol className="flex flex-col">
            {architectureLayers.map((layer, i) => (
              <li key={layer.id}>
                <div className="rounded-xl border border-border bg-card p-5">
                  <div className="flex items-baseline gap-3">
                    <span className="rounded-md border border-brand/30 bg-brand/10 px-2 py-0.5 font-mono text-xs font-medium text-brand">
                      {layer.id}
                    </span>
                    <h3 className="text-base font-semibold text-foreground">
                      {layer.title}
                    </h3>
                  </div>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {layer.blurb}
                  </p>
                  <ul className="mt-3 flex flex-col gap-1.5">
                    {layer.nodes.map((node) => (
                      <li
                        key={node}
                        className="rounded-md border border-border bg-secondary px-2.5 py-1.5 font-mono text-xs text-muted-foreground"
                      >
                        {node}
                      </li>
                    ))}
                  </ul>
                </div>

                {layer.edge && i < architectureLayers.length - 1 && (
                  <div className="flex items-center justify-center gap-2 py-2">
                    <ArrowDown className="size-3.5 text-brand" />
                    <span className="font-mono text-[11px] text-muted-foreground">
                      {layer.edge}
                    </span>
                  </div>
                )}
              </li>
            ))}
          </ol>
        </div>

        {/* Side column */}
        <div className="flex flex-col gap-6 lg:col-span-2">
          <div className="rounded-xl border border-border bg-card p-5">
            <h3 className="text-base font-semibold text-foreground">
              {ui.serviceInventory}
            </h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {ui.serviceInventoryNote}
            </p>
            <div className="mt-4 overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="pb-2 pr-3 font-medium text-muted-foreground">
                      {ui.colService}
                    </th>
                    <th className="pb-2 font-medium text-muted-foreground">
                      {ui.colLayers}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {services.map((svc) => (
                    <tr
                      key={svc.name}
                      className="border-b border-border/60 last:border-0"
                    >
                      <td className="py-2.5 pr-3 align-top">
                        <span className="font-mono text-sm text-brand">
                          {svc.name}
                        </span>
                        <p className="mt-0.5 text-xs leading-relaxed text-muted-foreground">
                          {svc.responsibility}
                        </p>
                      </td>
                      <td className="whitespace-nowrap py-2.5 align-top font-mono text-xs text-muted-foreground">
                        {svc.layers}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-5">
            <h3 className="text-base font-semibold text-foreground">
              {ui.techStackTitle}
            </h3>
            <dl className="mt-4 flex flex-col gap-4">
              {techStack.map((group) => (
                <div key={group.group}>
                  <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    {group.group}
                  </dt>
                  <dd className="mt-2 flex flex-wrap gap-1.5">
                    {group.items.map((item) => (
                      <span
                        key={item}
                        className="rounded-md border border-border bg-secondary px-2 py-1 font-mono text-xs text-muted-foreground"
                      >
                        {item}
                      </span>
                    ))}
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        </div>
      </div>
    </section>
  );
}
