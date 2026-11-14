"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { formatInterval } from "@/lib/format-interval";
import { cn } from "@/lib/utils";
import type { TRating, TStudyCard } from "../types";

type RatingBarProps = {
  card: TStudyCard;
  revealed: boolean;
  busy: boolean;
  onReveal: () => void;
  onAnswer: (rating: TRating) => void;
};

const BUTTONS = [
  { rating: 1, key: "again", interval: "again", className: "border-destructive/40 text-destructive hover:bg-destructive/10" },
  { rating: 2, key: "hard", interval: "hard", className: "" },
  { rating: 3, key: "good", interval: "good", className: "" },
  { rating: 4, key: "easy", interval: "easy", className: "" },
] as const;

// Bottom action area: "Show answer", then the four Anki buttons with their next intervals.
export function RatingBar({ card, revealed, busy, onReveal, onAnswer }: RatingBarProps) {
  const t = useTranslations("study");
  const tu = useTranslations("units");

  if (!revealed) {
    return (
      <Button size="lg" className="h-14 w-full text-base active:scale-[0.99]" onClick={onReveal} disabled={busy}>
        {t("showAnswer")}
      </Button>
    );
  }
  return (
    <div className="grid grid-cols-4 gap-2">
      {BUTTONS.map((b) => (
        <Button
          key={b.rating}
          variant={b.rating === 3 ? "default" : "outline"}
          disabled={busy}
          onClick={() => onAnswer(b.rating as TRating)}
          className={cn("h-16 flex-col gap-0.5 active:scale-[0.98]", b.className)}
        >
          <span className="font-mono text-xs opacity-80">{formatInterval(card.intervals[b.interval], tu)}</span>
          <span className="text-sm font-medium">{t(b.key)}</span>
          <span className="sr-only">({b.rating})</span>
        </Button>
      ))}
    </div>
  );
}
