"use client";

import { Volume2 } from "lucide-react";
import { useSpeak } from "@/hooks/use-speak";

type SpeakWordButtonProps = {
  word: string;
  ipa: string;
  accent: "en-GB" | "en-US";
  label: string;
};

export function SpeakWordButton({ word, ipa, accent, label }: SpeakWordButtonProps) {
  const { speak } = useSpeak(accent);
  return (
    <button
      type="button"
      onClick={() => speak(word)}
      className="flex w-full items-center justify-between gap-3 rounded-xl border border-border bg-background px-4 py-3 text-left transition-transform active:scale-[0.98] hover:border-primary/40"
    >
      <span>
        <span className="block font-medium">{word}</span>
        <span className="font-mono text-xs text-muted-foreground">
          {label} {ipa}
        </span>
      </span>
      <Volume2 className="size-4 text-primary" strokeWidth={1.75} aria-hidden />
    </button>
  );
}
