"use client";

import Image from "next/image";
import {
  Images,
  MessageSquare,
  Users,
  Video,
  type LucideIcon,
} from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { SiteContent } from "@/i18n";
import { SCREENSHOT_H, SCREENSHOT_W } from "@/i18n/shared";

const iconMap: Record<string, LucideIcon> = {
  MessageSquare,
  Users,
  Images,
  Video,
};

export function ScreenshotGallery({ content }: { content: SiteContent }) {
  const { galleryTabs, ui } = content;

  return (
    <section
      id="screenshots"
      className="border-t border-border bg-card/30 px-4 py-20 sm:px-6 sm:py-24 lg:px-8"
    >
      <div className="mx-auto max-w-7xl">
        <div className="mx-auto max-w-2xl text-center">
          <p className="text-sm font-medium text-brand">{ui.galleryEyebrow}</p>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            {ui.galleryTitle}
          </h2>
          <p className="mt-4 text-base text-muted-foreground">
            {ui.gallerySubtitle}
          </p>
        </div>

        <Tabs defaultValue={galleryTabs[0].id} className="mt-12 items-center">
          <TabsList>
            {galleryTabs.map((tab) => {
              const Icon = iconMap[tab.icon];
              return (
                <TabsTrigger key={tab.id} value={tab.id}>
                  {Icon ? <Icon /> : null}
                  {tab.label}
                </TabsTrigger>
              );
            })}
          </TabsList>

          {galleryTabs.map((tab) => (
            <TabsContent key={tab.id} value={tab.id} className="w-full">
              <div className="grid gap-5 md:grid-cols-2">
                {tab.shots.map((shot, i) => (
                  <figure
                    key={shot.src}
                    className="group overflow-hidden rounded-xl border border-border bg-card transition-colors hover:border-brand/50"
                  >
                    <div className="overflow-hidden border-b border-border">
                      <Image
                        src={shot.src}
                        alt={shot.title}
                        width={SCREENSHOT_W}
                        height={SCREENSHOT_H}
                        sizes="(min-width: 768px) 50vw, 100vw"
                        /* 首个 tab 的第一张最接近首屏，其余延迟到接近视口再加载 */
                        loading={
                          tab.id === galleryTabs[0].id && i === 0
                            ? "eager"
                            : "lazy"
                        }
                        className="w-full transition-transform duration-300 group-hover:scale-[1.02]"
                      />
                    </div>
                    <figcaption className="p-4">
                      <h3 className="text-sm font-semibold text-foreground">
                        {shot.title}
                      </h3>
                      <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                        {shot.caption}
                      </p>
                    </figcaption>
                  </figure>
                ))}
              </div>
            </TabsContent>
          ))}
        </Tabs>
      </div>
    </section>
  );
}
