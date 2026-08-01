import {
  Boxes,
  Images,
  MessageSquare,
  Users,
  Video,
  Workflow,
  type LucideIcon,
} from "lucide-react";
import type { SiteContent } from "@/i18n";

const iconMap: Record<string, LucideIcon> = {
  MessageSquare,
  Users,
  Images,
  Video,
  Workflow,
  Boxes,
};

export function FeatureGrid({ content }: { content: SiteContent }) {
  const { highlights, ui } = content;

  return (
    <section
      id="features"
      className="mx-auto max-w-7xl px-4 py-20 sm:px-6 sm:py-24 lg:px-8"
    >
      <div className="mx-auto max-w-2xl text-center">
        <p className="text-sm font-medium text-brand">{ui.featuresEyebrow}</p>
        <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          {ui.featuresTitle}
        </h2>
        <p className="mt-4 text-base text-muted-foreground">
          {ui.featuresSubtitle}
        </p>
      </div>

      <div className="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {highlights.map((item) => {
          const Icon = iconMap[item.icon];

          return (
            <article
              key={item.title}
              className="group rounded-xl border border-border bg-card p-6 transition-colors hover:border-brand/50"
            >
              <div className="inline-flex size-10 items-center justify-center rounded-lg border border-border bg-secondary transition-colors group-hover:border-brand/40">
                {Icon ? <Icon className="size-5 text-brand" /> : null}
              </div>

              <h3 className="mt-4 text-lg font-semibold text-foreground">
                {item.title}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {item.description}
              </p>

              <ul className="mt-4 flex flex-wrap gap-2">
                {item.points.map((point) => (
                  <li
                    key={point}
                    className="rounded-md border border-border bg-secondary px-2 py-1 text-xs text-muted-foreground"
                  >
                    {point}
                  </li>
                ))}
              </ul>
            </article>
          );
        })}
      </div>
    </section>
  );
}
