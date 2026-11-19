"use client";

import { Check, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { CardBack } from "@/features/study/components/CardBack";
import { CardFront } from "@/features/study/components/CardFront";
import { useCardAudio } from "@/features/study/hooks/use-card-audio";
import type { TLibraryWord } from "../types";

type WordPreviewDialogProps = {
  word: TLibraryWord | null;
  accent: "uk" | "us";
  selected: boolean;
  onToggle: () => void;
  onOpenChange: (open: boolean) => void;
};

/** The full card, exactly as it will be studied (the study screen's components). */
export function WordPreviewDialog({ word, accent, selected, onToggle, onOpenChange }: WordPreviewDialogProps) {
  const t = useTranslations();
  const { play, playHeadword } = useCardAudio(accent);
  const content = word?.card;
  return (
    <Dialog open={!!word} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className="sr-only">{content?.headword}</DialogTitle>
          <DialogDescription className="sr-only">{t("library.previewHint")}</DialogDescription>
        </DialogHeader>
        {content && word && (
          <div className="space-y-6">
            <CardFront content={content} accent={accent} showImage={false} onPlay={() => playHeadword(content)} />
            <div className="border-t border-border pt-6">
              <CardBack content={content} onPlay={play} />
            </div>
          </div>
        )}
        {word && (
          <DialogFooter>
            {word.in_deck ? (
              <p className="text-sm text-muted-foreground">{t("library.alreadyInDeck")}</p>
            ) : (
              <Button variant={selected ? "outline" : "default"} onClick={onToggle}>
                {selected ? <Check className="size-4" strokeWidth={1.75} aria-hidden /> : <Plus className="size-4" strokeWidth={1.75} aria-hidden />}
                {selected ? t("library.selected") : t("library.select")}
              </Button>
            )}
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  );
}
