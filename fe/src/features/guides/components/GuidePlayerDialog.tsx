"use client";

import { Captions } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { useLocalized } from "../hooks/use-guides";
import { markGuideSeen } from "../lib/guide-state";
import type { TGuide } from "../types";

type GuidePlayerDialogProps = {
  guide: TGuide;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Plays a guide video. It counts as watched once 80% has played, so the card then shrinks to a link. */
export function GuidePlayerDialog({ guide, open, onOpenChange }: GuidePlayerDialogProps) {
  const t = useTranslations("guides");
  const locale = useLocale();
  const pick = useLocalized();
  const [failed, setFailed] = useState(false);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[calc(100dvh-2rem)] gap-0 overflow-y-auto p-0 sm:max-w-4xl">
        <div className="aspect-video w-full bg-black">
          {open && !failed && (
            <video
              key={guide.slug}
              src={guide.video_url}
              poster={guide.poster_url || undefined}
              controls
              autoPlay
              playsInline
              preload="metadata"
              className="size-full"
              onTimeUpdate={(e) => {
                const v = e.currentTarget;
                if (v.duration && v.currentTime / v.duration >= 0.8) markGuideSeen(guide.slug);
              }}
              onEnded={() => markGuideSeen(guide.slug)}
              onError={() => setFailed(true)}
            />
          )}
          {failed && <p className="grid size-full place-items-center p-6 text-center text-sm text-white/80">{t("playError")}</p>}
        </div>
        <div className="space-y-1 p-4 pr-12 sm:p-5">
          <DialogTitle className="text-base font-semibold">{pick(guide.title)}</DialogTitle>
          <DialogDescription>{pick(guide.description)}</DialogDescription>
          {guide.lang !== locale && (
            <p className="flex items-center gap-1.5 pt-1 text-xs text-muted-foreground">
              <Captions className="size-3.5" strokeWidth={1.75} aria-hidden />
              {t("narratedIn", { lang: guide.lang })}
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
