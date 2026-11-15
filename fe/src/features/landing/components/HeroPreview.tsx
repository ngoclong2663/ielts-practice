"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { FlashcardPreview } from "./FlashcardPreview";
import { WritingPreview } from "./WritingPreview";

/** The two halves of the app, side by side: a flashcard and a guided writing step. */
export function HeroPreview() {
  const t = useTranslations("landing");
  const [tab, setTab] = useState<"writing" | "vocab">("writing");
  const tabs = [
    ["writing", t("previewTabWriting")],
    ["vocab", t("previewTabVocab")],
  ] as const;

  return (
    <div className="flex w-full max-w-md flex-col items-stretch gap-3">
      <div role="tablist" aria-label={t("previewTabs")} className="inline-flex w-fit self-center rounded-xl bg-muted p-1 lg:self-start">
        {tabs.map(([id, label]) => (
          <button
            key={id}
            type="button"
            role="tab"
            aria-selected={tab === id}
            onClick={() => setTab(id)}
            className={cn(
              "rounded-lg px-4 py-1.5 text-sm font-medium transition-colors",
              tab === id ? "bg-background shadow-sm" : "text-muted-foreground hover:text-foreground",
            )}
          >
            {label}
          </button>
        ))}
      </div>
      <div role="tabpanel">{tab === "writing" ? <WritingPreview /> : <FlashcardPreview />}</div>
    </div>
  );
}
