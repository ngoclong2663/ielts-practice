"use client";

import { Volume2 } from "lucide-react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { sitePhotos } from "@/assets/images";
import { PhotoCredit } from "@/components/shared/PhotoCredit";
import { Button } from "@/components/ui/button";
import { useSpeak } from "@/hooks/use-speak";
import { cn } from "@/lib/utils";

// A working mini version of the study card (sample content, not a screenshot).
const SAMPLE = {
  word: "resilient",
  ipa: "/rɪˈzɪliənt/",
  pos: "adjective",
  meaningVi: "kiên cường, mau phục hồi",
  example: "Children are often more resilient than adults expect.",
  forms: ["resilience (n)", "resiliently (adv)"],
};

const RATINGS = [
  { key: "previewAgain", interval: "<1m" },
  { key: "previewHard", interval: "<6m" },
  { key: "previewGood", interval: "<10m" },
  { key: "previewEasy", interval: "4d" },
] as const;

export function FlashcardPreview() {
  const t = useTranslations("landing");
  const [revealed, setRevealed] = useState(false);
  const { speak, supported } = useSpeak("en-GB");

  return (
    <div className="w-full max-w-md rounded-2xl border border-border bg-card p-5 shadow-[0_24px_60px_-30px_oklch(0.35_0.06_165/0.35)] sm:p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-medium text-muted-foreground">{t("previewWord")}</p>
          <p className="mt-2 text-3xl font-semibold tracking-tight">{SAMPLE.word}</p>
          <p className="mt-1 font-mono text-sm text-muted-foreground">
            {SAMPLE.ipa} <span className="font-sans italic">{SAMPLE.pos}</span>
          </p>
        </div>
        <Button
          variant="outline"
          size="icon"
          onClick={() => speak(SAMPLE.word)}
          disabled={!supported}
          aria-label={`Play ${SAMPLE.word}`}
        >
          <Volume2 className="size-4" strokeWidth={1.75} aria-hidden />
        </Button>
      </div>

      <div
        className={cn(
          "grid transition-[grid-template-rows,opacity] duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]",
          revealed ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0",
        )}
        aria-hidden={!revealed}
      >
        <div className="overflow-hidden">
          <div className="mt-5 space-y-4 border-t border-border pt-5">
            <div className="relative aspect-[12/7] overflow-hidden rounded-xl bg-muted">
              <Image src={sitePhotos.heroResilient.src} alt={sitePhotos.heroResilient.alt} fill placeholder="blur" sizes="(max-width: 640px) 100vw, 420px" className="object-cover" />
            </div>
            <PhotoCredit photo={sitePhotos.heroResilient} className="-mt-2 text-muted-foreground" />
            <p className="text-lg font-medium text-primary">{SAMPLE.meaningVi}</p>
            <p className="text-sm text-muted-foreground">{SAMPLE.forms.join(", ")}</p>
            <button
              type="button"
              onClick={() => speak(SAMPLE.example)}
              className="flex w-full items-start gap-2 text-left text-sm leading-relaxed hover:text-primary"
            >
              <Volume2 className="mt-0.5 size-4 shrink-0" strokeWidth={1.75} aria-hidden />
              <span>{SAMPLE.example}</span>
            </button>
          </div>
        </div>
      </div>

      <div className="mt-6">
        {revealed ? (
          <div className="grid grid-cols-4 gap-2">
            {RATINGS.map(({ key, interval }) => (
              <Button key={key} variant={key === "previewGood" ? "default" : "outline"} className="h-auto flex-col gap-0.5 py-2" onClick={() => setRevealed(false)}>
                <span className="font-mono text-[11px] opacity-80">{interval}</span>
                <span className="text-xs">{t(key)}</span>
              </Button>
            ))}
          </div>
        ) : (
          <Button className="w-full" onClick={() => setRevealed(true)}>
            {t("previewShowAnswer")}
          </Button>
        )}
      </div>
    </div>
  );
}
