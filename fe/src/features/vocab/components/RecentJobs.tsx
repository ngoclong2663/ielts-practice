"use client";

import { ChevronRight } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { useIsClient } from "@/hooks/use-is-client";
import { useRecentJobs } from "../hooks/use-vocab";
import { FINAL_JOB_STATUSES } from "../types";

/** "Recent generations": leaving the page never loses a job (US-3.2). */
export function RecentJobs({ onOpen }: { onOpen: (jobId: string) => void }) {
  const t = useTranslations("vocab");
  const format = useFormatter();
  const isClient = useIsClient();
  const query = useRecentJobs();
  const jobs = isClient ? query.data : undefined;
  if (!jobs?.length) return null;

  return (
    <section aria-labelledby="recent-jobs">
      <h2 id="recent-jobs" className="mb-3 text-sm font-medium text-muted-foreground">
        {t("recentTitle")}
      </h2>
      <ul className="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-card">
        {jobs.map((job) => {
          const working = !FINAL_JOB_STATUSES.includes(job.status);
          return (
            <li key={job.id}>
              <button
                type="button"
                onClick={() => onOpen(job.id)}
                className="flex w-full items-center gap-4 px-4 py-3 text-left transition-colors hover:bg-muted/50"
              >
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">
                    {job.words.join(", ")}
                    {job.total > job.words.length && <span className="text-muted-foreground"> …</span>}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {format.dateTime(new Date(job.created_at), { dateStyle: "medium", timeStyle: "short" })} ·{" "}
                    {working
                      ? t("recentWorking", { done: job.done, total: job.total })
                      : job.committed_deck_id
                        ? t("recentSaved", { count: job.total })
                        : t("recentReady", { count: job.total })}
                  </p>
                </div>
                {working && <span className="size-2 shrink-0 animate-pulse rounded-full bg-primary" aria-hidden />}
                <ChevronRight className="size-4 shrink-0 text-muted-foreground" strokeWidth={1.75} aria-hidden />
              </button>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
