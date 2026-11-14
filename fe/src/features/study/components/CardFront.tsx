"use client";

import { Volume2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import type { TCardContent } from "@/types/card-content";
import { CardImage } from "./CardBack";

type CardFrontProps = {
  content: TCardContent;
  accent: "uk" | "us";
  showImage: boolean;
  onPlay: () => void;
};

export function CardFront({ content, accent, showImage, onPlay }: CardFrontProps) {
  const t = useTranslations("study");
  const ipa = content.ipa?.[accent] || content.ipa?.uk || content.ipa?.us;
  return (
    <div className="space-y-5">
      {showImage && <CardImage content={content} />}
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-4xl font-semibold leading-tight tracking-tight break-words sm:text-5xl">{content.headword}</h1>
          <p className="mt-2 text-muted-foreground">
            {ipa && <span className="font-mono">{ipa}</span>}
            {content.part_of_speech && <span className="ml-2 italic">{content.part_of_speech}</span>}
          </p>
        </div>
        <Button variant="outline" size="icon-lg" onClick={onPlay} aria-label={t("play")} className="shrink-0 rounded-full">
          <Volume2 className="size-5" strokeWidth={1.75} aria-hidden />
        </Button>
      </div>
    </div>
  );
}
