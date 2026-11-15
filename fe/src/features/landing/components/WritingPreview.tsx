"use client";

import { CircleAlert, PenLine } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// A working mini version of a guided writing step (sample content). The learner's sentence and the
// feedback are fixed; the feedback is in Vietnamese, as in the app.
const SAMPLE = {
  promptVi: "Nhiều người cao tuổi phải uống thuốc mỗi ngày.",
  answer: "Many old people must drink medicine every day.",
  explainVi: "“Uống thuốc” là take medicine, không dùng drink. “Phải” ở đây nên là have to.",
  corrected: "Many elderly people have to take medicine every day.",
};

export function WritingPreview() {
  const t = useTranslations("landing");
  const [checked, setChecked] = useState(false);

  return (
    <div className="w-full max-w-md rounded-2xl border border-border bg-card p-5 shadow-[0_24px_60px_-30px_oklch(0.35_0.06_165/0.35)] sm:p-6">
      <p className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
        <PenLine className="size-3.5" strokeWidth={1.75} aria-hidden />
        {t("previewWritingLabel")}
      </p>
      <p className="mt-3 text-lg leading-snug font-medium" lang="vi">
        {SAMPLE.promptVi}
      </p>
      <p className="mt-5 text-xs font-medium text-muted-foreground">{t("previewYourSentence")}</p>
      <p className="mt-1.5 rounded-lg border border-input px-3 py-2 text-[15px]" lang="en">
        {SAMPLE.answer}
      </p>

      <div
        className={cn(
          "grid transition-[grid-template-rows,opacity] duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]",
          checked ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0",
        )}
        aria-hidden={!checked}
      >
        <div className="overflow-hidden">
          <div className="mt-4 space-y-2 rounded-xl border border-amber-500/40 bg-amber-500/10 p-3 text-sm">
            <p className="flex flex-wrap items-center gap-2 font-medium text-amber-900 dark:text-amber-300">
              <CircleAlert className="size-4" strokeWidth={2} aria-hidden />
              {t("previewVerdict")}
              <span className="rounded-full border border-current/30 px-2 py-0.5 text-xs font-normal">
                {t("previewErrorType")}
              </span>
            </p>
            <p lang="vi">{SAMPLE.explainVi}</p>
            <p>
              <span className="text-muted-foreground">{t("previewCorrected")} </span>
              <span lang="en">{SAMPLE.corrected}</span>
            </p>
          </div>
        </div>
      </div>

      <Button className="mt-6 w-full" variant={checked ? "outline" : "default"} onClick={() => setChecked((c) => !c)}>
        {checked ? t("previewTryAgain") : t("previewCheck")}
      </Button>
    </div>
  );
}
