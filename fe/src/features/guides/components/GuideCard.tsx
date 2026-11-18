"use client";

import { Play, PlayCircle } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useIsClient } from "@/hooks/use-is-client";
import { cn } from "@/lib/utils";
import { formatDuration, useGuide, useLocalized } from "../hooks/use-guides";
import { markGuideSeen, useGuideSeen } from "../lib/guide-state";
import type { TGuide, TGuideTopic } from "../types";
import { GuidePlayerDialog } from "./GuidePlayerDialog";

/**
 * The guide video of a topic at the top of a page. Until the learner has watched it (or chosen "Later"),
 * it is a card with the poster; afterwards a one-line "watch again" link that stays out of the way.
 */
export function GuideCard({ topic, className }: { topic: TGuideTopic; className?: string }) {
  const isClient = useIsClient();
  const guide = useGuide(topic);
  if (!isClient || !guide) return null;
  return <GuideCardBody guide={guide} className={className} />;
}

function GuideCardBody({ guide, className }: { guide: TGuide; className?: string }) {
  const t = useTranslations("guides");
  const pick = useLocalized();
  const seen = useGuideSeen(guide.slug);
  const [open, setOpen] = useState(false);
  const duration = formatDuration(guide.duration_seconds);
  const player = <GuidePlayerDialog guide={guide} open={open} onOpenChange={setOpen} />;

  if (seen) {
    return (
      <div className={className}>
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground"
        >
          <PlayCircle className="size-4 text-primary" strokeWidth={1.75} aria-hidden />
          {t("rewatch", { title: pick(guide.title) })}
          <span className="font-mono text-xs tabular-nums">{duration}</span>
        </button>
        {player}
      </div>
    );
  }

  return (
    <section
      aria-label={t("label")}
      className={cn("flex flex-col gap-4 rounded-2xl border border-primary/25 bg-primary/5 p-3 sm:flex-row sm:items-center sm:p-4", className)}
    >
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="group relative aspect-video w-full shrink-0 overflow-hidden rounded-xl bg-muted outline-none focus-visible:ring-2 focus-visible:ring-ring sm:w-60"
        aria-label={t("play", { title: pick(guide.title) })}
      >
        {guide.poster_url && (
          // eslint-disable-next-line @next/next/no-img-element -- a redirect to a presigned S3 link, not a static asset
          <img src={guide.poster_url} alt="" className="size-full object-cover transition-transform duration-300 group-hover:scale-[1.03]" />
        )}
        {/* Corner play button: the poster's own title stays readable. */}
        <span className="absolute inset-0 bg-gradient-to-t from-black/25 to-transparent transition-colors group-hover:from-black/35" />
        <span className="absolute bottom-2 left-2 grid size-10 place-items-center rounded-full bg-primary text-primary-foreground shadow-lg transition-transform group-hover:scale-110">
          <Play className="ml-0.5 size-4 fill-current" strokeWidth={1.75} aria-hidden />
        </span>
        <span className="absolute right-2 bottom-2 rounded-md bg-black/70 px-1.5 py-0.5 font-mono text-xs text-white tabular-nums">{duration}</span>
      </button>
      <div className="min-w-0 flex-1 sm:py-1">
        <p className="text-xs font-medium tracking-wide text-primary uppercase">{t("kicker")}</p>
        <h2 className="mt-1 font-semibold tracking-tight">{pick(guide.title)}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{pick(guide.description)}</p>
        <div className="mt-3 flex flex-wrap gap-2">
          <Button size="sm" onClick={() => setOpen(true)}>
            <Play className="size-3.5 fill-current" strokeWidth={1.75} aria-hidden />
            {t("watch", { duration })}
          </Button>
          <Button size="sm" variant="ghost" onClick={() => markGuideSeen(guide.slug)}>
            {t("later")}
          </Button>
        </div>
      </div>
      {player}
    </section>
  );
}

/** An inline "watch the video" link, for the place where the learner makes the choice the guide explains. */
export function GuideLink({ topic, label }: { topic: TGuideTopic; label: string }) {
  const isClient = useIsClient();
  const guide = useGuide(topic);
  const [open, setOpen] = useState(false);
  if (!isClient || !guide) return null;
  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="inline-flex items-center gap-1.5 text-sm font-medium text-primary underline-offset-4 hover:underline"
      >
        <PlayCircle className="size-4" strokeWidth={1.75} aria-hidden />
        {label}
        <span className="font-mono text-xs text-muted-foreground tabular-nums">{formatDuration(guide.duration_seconds)}</span>
      </button>
      <GuidePlayerDialog guide={guide} open={open} onOpenChange={setOpen} />
    </>
  );
}
