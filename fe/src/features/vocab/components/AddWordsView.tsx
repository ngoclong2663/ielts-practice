"use client";

import { PenLine, Sparkles } from "lucide-react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { PageHeader } from "@/components/shared/PageHeader";
import { ManualAddPanel } from "@/features/decks/components/ManualAddPanel";
import { usePathname, useRouter } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { JobReview } from "./JobReview";
import { PasteWordsPanel } from "./PasteWordsPanel";

type TMode = "generate" | "manual";

/** Add words: paste a list and get generated cards (Epic 3), or type one card by hand. */
export function AddWordsView() {
  const t = useTranslations("vocab");
  const params = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const jobId = params.get("job");
  const mode: TMode = params.get("mode") === "manual" ? "manual" : "generate";

  const go = (next: Record<string, string | null>) => {
    const q = new URLSearchParams(params.toString());
    for (const [k, v] of Object.entries(next)) {
      if (v === null) q.delete(k);
      else q.set(k, v);
    }
    const s = q.toString();
    router.replace(s ? `${pathname}?${s}` : pathname, { scroll: false });
  };

  if (jobId) {
    return <JobReview jobId={jobId} onNewList={() => go({ job: null })} />;
  }

  const tabs: { id: TMode; label: string; icon: typeof Sparkles }[] = [
    { id: "generate", label: t("modeGenerate"), icon: Sparkles },
    { id: "manual", label: t("modeManual"), icon: PenLine },
  ];

  return (
    <div className="mx-auto max-w-3xl">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <div role="tablist" aria-label={t("modeLabel")} className="mb-6 inline-flex rounded-xl bg-muted p-1">
        {tabs.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            type="button"
            role="tab"
            aria-selected={mode === id}
            onClick={() => go({ mode: id === "manual" ? "manual" : null })}
            className={cn(
              "inline-flex items-center gap-2 rounded-lg px-4 py-1.5 text-sm font-medium transition-colors",
              mode === id ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground",
            )}
          >
            <Icon className="size-4" strokeWidth={1.75} aria-hidden />
            {label}
          </button>
        ))}
      </div>
      {mode === "manual" ? <ManualAddPanel /> : <PasteWordsPanel onCreated={(id) => go({ job: id })} />}
    </div>
  );
}
