"use client";

import { ArrowRight, Library, Lightbulb, Sparkles } from "lucide-react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { PageHeader } from "@/components/shared/PageHeader";
import { Link, usePathname, useRouter } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { JobReview } from "./JobReview";
import { PasteWordsPanel } from "./PasteWordsPanel";
import { SuggestionReview } from "./SuggestionReview";
import { TopicPanel } from "./TopicPanel";

type TMode = "generate" | "topic";

/** Add words: ready-made cards in the library (Epic 8), paste a list (Epic 3), or new words by topic (Epic 4). */
export function AddWordsView() {
  const t = useTranslations("vocab");
  const params = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const jobId = params.get("job");
  const suggestionId = params.get("suggestion");
  const rawMode = params.get("mode");
  const mode: TMode = rawMode === "topic" ? rawMode : "generate";

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
  if (suggestionId) {
    return (
      <SuggestionReview
        suggestionId={suggestionId}
        onBack={() => go({ suggestion: null, mode: "topic" })}
        onGenerated={(id) => go({ suggestion: null, job: id })}
      />
    );
  }

  const tabs: { id: TMode; label: string; icon: typeof Sparkles }[] = [
    { id: "generate", label: t("modeGenerate"), icon: Sparkles },
    { id: "topic", label: t("modeTopic"), icon: Lightbulb },
  ];

  return (
    <div className="mx-auto max-w-3xl">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <Link
        href="/library"
        className="group mb-6 flex items-center gap-4 rounded-2xl border border-border bg-card p-4 transition-colors hover:border-primary/50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      >
        <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <Library className="size-5" strokeWidth={1.75} aria-hidden />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block font-medium">{t("libraryTitle")}</span>
          <span className="block text-sm text-muted-foreground">{t("libraryHint")}</span>
        </span>
        <ArrowRight className="size-5 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" strokeWidth={1.75} aria-hidden />
      </Link>
      <div role="tablist" aria-label={t("modeLabel")} className="mb-6 flex w-full overflow-x-auto rounded-xl bg-muted p-1 sm:inline-flex sm:w-auto">
        {tabs.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            type="button"
            role="tab"
            aria-selected={mode === id}
            onClick={() => go({ mode: id === "generate" ? null : id })}
            className={cn(
              "inline-flex flex-1 items-center justify-center gap-2 rounded-lg px-2 py-1.5 text-sm font-medium whitespace-nowrap transition-colors sm:flex-none sm:px-4",
              mode === id ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground",
            )}
          >
            <Icon className="hidden size-4 sm:block" strokeWidth={1.75} aria-hidden />
            {label}
          </button>
        ))}
      </div>
      {mode === "topic" && <TopicPanel onCreated={(id) => go({ suggestion: id })} />}
      {mode === "generate" && <PasteWordsPanel onCreated={(id) => go({ job: id })} />}
    </div>
  );
}
