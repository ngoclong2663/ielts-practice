"use client";

import { useFormatter, useLocale, useTranslations } from "next-intl";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button, buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useAttempts, useStats } from "../../hooks/use-exam";
import { formatClock } from "../../lib/format";
import type { TAttemptSummary, TSkill } from "../../types";

/** /<skill>/history: accuracy per question type (weakest first), full-test bands and every attempt. */
export function ExamHistoryView({ skill }: { skill: TSkill }) {
  const t = useTranslations("exam");
  const isClient = useIsClient();
  const attemptsQ = useAttempts(skill);
  const statsQ = useStats(skill);
  const items = isClient ? attemptsQ.data?.pages.flatMap((p) => p.items) : undefined;
  const stats = isClient ? statsQ.data : undefined;

  return (
    <div className="mx-auto max-w-3xl">
      <PageHeader title={t(`history.title.${skill}`)} description={t("history.subtitle")} />
      {!items || !stats ? (
        <div className="grid gap-4">
          <Skeleton className="h-40 rounded-xl" />
          <Skeleton className="h-64 rounded-xl" />
        </div>
      ) : items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border p-8 text-center">
          <p className="text-sm text-muted-foreground">{t("history.empty")}</p>
          <Link href={`/${skill}`} className={buttonVariants({ className: "mt-4" })}>
            {t("history.start")}
          </Link>
        </div>
      ) : (
        <div className="grid gap-4">
          {stats.by_type.length > 0 && (
            <section aria-labelledby="by-type" className="rounded-xl border border-border bg-card p-5">
              <h2 id="by-type" className="text-[15px] font-semibold">
                {t("history.byType")}
              </h2>
              <p className="mb-2 text-[13px] text-muted-foreground">{t("history.byTypeNote", { count: stats.attempts })}</p>
              <ul>
                {stats.by_type.map((s) => (
                  <li key={s.type} className="grid grid-cols-[minmax(0,1fr)_8rem_3.5rem] items-center gap-3 border-t border-border py-2 first:border-t-0">
                    <span className="truncate text-sm">{t(`types.${s.type}`)}</span>
                    <span className="h-2 overflow-hidden rounded-full bg-muted" aria-hidden>
                      <span
                        className={cn("block h-full rounded-full", s.correct / s.total < 0.5 ? "bg-destructive" : "bg-primary")}
                        style={{ width: `${(s.correct / s.total) * 100}%` }}
                      />
                    </span>
                    <span className="text-right text-sm tabular-nums">{Math.round((s.correct / s.total) * 100)}%</span>
                  </li>
                ))}
              </ul>
            </section>
          )}

          <section aria-labelledby="attempts-h" className="rounded-xl border border-border bg-card">
            <h2 id="attempts-h" className="px-5 pt-4 pb-2 text-[15px] font-semibold">
              {t("history.attempts")}
            </h2>
            <ul>
              {items.map((a) => (
                <li key={a.id} className="border-t border-border">
                  <AttemptRow attempt={a} />
                </li>
              ))}
            </ul>
            {attemptsQ.hasNextPage && (
              <div className="border-t border-border p-3 text-center">
                <Button variant="ghost" onClick={() => attemptsQ.fetchNextPage()} disabled={attemptsQ.isFetchingNextPage}>
                  {t("history.more")}
                </Button>
              </div>
            )}
          </section>
        </div>
      )}
    </div>
  );
}

function AttemptRow({ attempt: a }: { attempt: TAttemptSummary }) {
  const t = useTranslations("exam");
  const locale = useLocale() as "vi" | "en";
  const format = useFormatter();
  const scope =
    a.kind === "full"
      ? t("fullTest")
      : a.kind === "redo"
        ? t("result.redoScope")
        : a.part_nos.map((n) => t(`partLabel.${a.skill}`, { n })).join(", ");
  return (
    <Link href={`/${a.skill}/attempts/${a.id}`} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-5 py-3 hover:bg-muted/50">
      <span className="min-w-0">
        <span className="block truncate text-sm font-medium">
          {(a.book_title[locale] || a.book_title.en)} · {t("testNo", { n: a.test_no })} · {scope}
        </span>
        <span className="block text-[13px] text-muted-foreground">
          {format.dateTime(new Date(a.created_at), { dateStyle: "medium", timeStyle: "short" })} · {formatClock(a.duration_sec)}
        </span>
      </span>
      <span className="text-right">
        <span className="block text-sm font-semibold tabular-nums">
          {a.correct}/{a.total}
        </span>
        {a.band !== null && <span className="block text-[13px] text-muted-foreground">{t("result.band", { band: a.band.toFixed(1) })}</span>}
      </span>
    </Link>
  );
}
