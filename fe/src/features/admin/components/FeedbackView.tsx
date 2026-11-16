"use client";

import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { PageHeader } from "@/components/shared/PageHeader";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { FEEDBACK_CATEGORIES } from "@/features/feedback/types";
import type { TFeedbackCategory } from "@/features/feedback/types";
import { useIsClient } from "@/hooks/use-is-client";
import { cn } from "@/lib/utils";
import { useAdminFeedback } from "../hooks/use-admin";
import { FEEDBACK_STATUSES } from "../types";
import type { TFeedbackStatusFilter } from "../types";
import { FeedbackItem } from "./FeedbackItem";
import { Pager } from "./Pager";
import { QueryError } from "./QueryError";

const FILTERS: TFeedbackStatusFilter[] = [...FEEDBACK_STATUSES, "all"];

export function FeedbackView() {
  const t = useTranslations();
  const isClient = useIsClient();
  // ?id=<id> comes from the Telegram message: show every status so the item is in the list.
  const focusId = useSearchParams().get("id") ?? "";
  const [status, setStatus] = useState<TFeedbackStatusFilter>(focusId ? "all" : "new");
  const [category, setCategory] = useState<TFeedbackCategory | "">("");
  const [page, setPage] = useState(1);
  const query = useAdminFeedback({ status, category, page });
  const data = isClient ? query.data : undefined;
  const focusLoaded = !!focusId && !!data?.items.some((i) => i.id === focusId);

  useEffect(() => {
    if (focusLoaded) {
      document.getElementById(`feedback-${focusId}`)?.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }, [focusLoaded, focusId]);

  const categoryItems = [
    { value: "", label: t("admin.feedback.allCategories") },
    ...FEEDBACK_CATEGORIES.map((c) => ({ value: c, label: t(`feedback.category.${c}`) })),
  ];
  const countOf = (f: TFeedbackStatusFilter) =>
    data ? (f === "all" ? data.counts.new + data.counts.seen + data.counts.resolved : data.counts[f]) : undefined;

  return (
    <div className="mx-auto max-w-3xl">
      <PageHeader title={t("admin.feedback.title")} description={t("admin.feedback.subtitle")} />

      <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div role="tablist" aria-label={t("admin.feedback.statusLabel")} className="inline-flex flex-wrap gap-1 rounded-lg bg-muted p-1">
          {FILTERS.map((f) => {
            const count = countOf(f);
            return (
              <button
                key={f}
                type="button"
                role="tab"
                aria-selected={status === f}
                onClick={() => {
                  setStatus(f);
                  setPage(1);
                }}
                className={cn(
                  "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                  status === f ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground",
                )}
              >
                {t(`admin.feedback.filter.${f}`)}
                {count !== undefined && <span className="ml-1.5 font-mono text-xs tabular-nums text-muted-foreground">{count}</span>}
              </button>
            );
          })}
        </div>
        <Select
          items={categoryItems}
          value={category}
          onValueChange={(v) => {
            setCategory((v ?? "") as TFeedbackCategory | "");
            setPage(1);
          }}
        >
          <SelectTrigger className="w-full sm:w-48" aria-label={t("feedback.categoryLabel")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {categoryItems.map((c) => (
              <SelectItem key={c.value || "all"} value={c.value}>
                {c.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {query.isError && <QueryError onRetry={() => void query.refetch()} />}

      {!data && !query.isError ? (
        <div className="space-y-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-40 rounded-2xl" />
          ))}
        </div>
      ) : data && data.items.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
          {t(status === "new" ? "admin.feedback.emptyNew" : "admin.feedback.empty")}
        </p>
      ) : data ? (
        <div className="space-y-3">
          {data.items.map((item) => (
            <FeedbackItem key={`${item.id}:${item.updated_at}`} item={item} highlighted={item.id === focusId} />
          ))}
        </div>
      ) : null}

      {data && <Pager page={data.page} pageSize={data.page_size} total={data.total} onPage={setPage} />}
    </div>
  );
}
