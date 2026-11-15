"use client";

import { CircleAlert, CircleCheck, CircleDashed, CircleX } from "lucide-react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import type { TSentenceFeedback, TVerdict } from "../types";

const VERDICT_STYLE: Record<TVerdict, { icon: typeof CircleCheck; className: string }> = {
  correct: { icon: CircleCheck, className: "border-emerald-500/40 bg-emerald-500/10 text-emerald-800 dark:text-emerald-300" },
  understandable: { icon: CircleAlert, className: "border-amber-500/40 bg-amber-500/10 text-amber-900 dark:text-amber-300" },
  wrong: { icon: CircleX, className: "border-destructive/40 bg-destructive/10 text-destructive" },
  empty: { icon: CircleDashed, className: "border-border bg-muted text-muted-foreground" },
};

export function SentenceFeedback({
  feedback,
  answer,
  model,
}: {
  feedback: TSentenceFeedback;
  answer: string;
  model?: string;
}) {
  const t = useTranslations("writing");
  const style = VERDICT_STYLE[feedback.verdict];
  const Icon = style.icon;
  const corrected = feedback.corrected_en && feedback.corrected_en !== answer && feedback.verdict !== "empty";
  return (
    <div className={cn("space-y-2 rounded-xl border p-3 text-sm", style.className)} data-verdict={feedback.verdict}>
      <p className="flex flex-wrap items-center gap-2 font-medium">
        <Icon className="size-4" strokeWidth={2} aria-hidden />
        {t(`verdicts.${feedback.verdict}`)}
        {feedback.error_types.map((e) => (
          <span key={e} className="rounded-full border border-current/30 px-2 py-0.5 text-xs font-normal">
            {t(`errorTypes.${e}`)}
          </span>
        ))}
      </p>
      {feedback.explain_vi && <p className="text-foreground">{feedback.explain_vi}</p>}
      {corrected && (
        <p className="text-foreground">
          <span className="text-muted-foreground">{t("corrected")} </span>
          {feedback.corrected_en}
        </p>
      )}
      {model && (
        <p className="text-foreground">
          <span className="text-muted-foreground">{t("model")} </span>
          {model}
        </p>
      )}
    </div>
  );
}
