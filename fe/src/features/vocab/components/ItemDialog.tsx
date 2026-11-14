"use client";

import { Eye, PenLine } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useMe } from "@/features/profile/hooks/use-me";
import { CardBack } from "@/features/study/components/CardBack";
import { CardFront } from "@/features/study/components/CardFront";
import { useCardAudio } from "@/features/study/hooks/use-card-audio";
import { cn } from "@/lib/utils";
import type { TJobItem } from "../types";
import { EditContentForm } from "./EditContentForm";

type ItemDialogProps = {
  jobId: string;
  item: TJobItem | undefined;
  onOpenChange: (open: boolean) => void;
  onPickImage: (itemId: string) => void;
};

/** Preview (the exact study card components) and edit one generated word before saving. */
export function ItemDialog({ jobId, item, onOpenChange, onPickImage }: ItemDialogProps) {
  const t = useTranslations("vocab");
  const accent = useMe().data?.study_settings.accent ?? "uk";
  const { play, playHeadword } = useCardAudio(accent);
  const [tab, setTab] = useState<"preview" | "edit">("preview");
  const [lastItemId, setLastItemId] = useState<string | undefined>(undefined);
  if (item?.id !== lastItemId) {
    setLastItemId(item?.id);
    setTab("preview");
  }
  const content = item?.status === "ready" ? item.content : null;

  return (
    <Dialog open={!!content} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{content?.headword}</DialogTitle>
          <DialogDescription className="sr-only">{t("previewEdit")}</DialogDescription>
        </DialogHeader>
        <div role="tablist" aria-label={t("previewEdit")} className="inline-flex w-fit rounded-xl bg-muted p-1">
          {(
            [
              ["preview", t("preview"), Eye],
              ["edit", t("edit"), PenLine],
            ] as const
          ).map(([id, label, Icon]) => (
            <button
              key={id}
              type="button"
              role="tab"
              aria-selected={tab === id}
              onClick={() => setTab(id)}
              className={cn(
                "inline-flex items-center gap-2 rounded-lg px-3 py-1 text-sm font-medium transition-colors",
                tab === id ? "bg-background shadow-sm" : "text-muted-foreground hover:text-foreground",
              )}
            >
              <Icon className="size-4" strokeWidth={1.75} aria-hidden />
              {label}
            </button>
          ))}
        </div>

        {content && item && tab === "preview" && (
          <div className="space-y-6">
            <CardFront content={content} accent={accent} showImage={false} onPlay={() => playHeadword(content)} />
            <div className="border-t border-border pt-6">
              <CardBack content={content} onPlay={play} />
            </div>
            <button
              type="button"
              onClick={() => onPickImage(item.id)}
              className="text-sm text-primary underline underline-offset-2"
            >
              {t("changeImage")}
            </button>
          </div>
        )}
        {content && item && tab === "edit" && (
          <EditContentForm jobId={jobId} item={item} content={content} onSaved={() => setTab("preview")} />
        )}
      </DialogContent>
    </Dialog>
  );
}
