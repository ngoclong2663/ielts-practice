"use client";

import { CalendarCheck, Flame } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { SrsCounts } from "@/components/srs/SrsCounts";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { DashboardGreeting } from "@/features/dashboard/components/DashboardGreeting";
import { useStudyCounts } from "@/features/study/hooks/use-study-counts";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useStatsOverview } from "../hooks/use-stats";
import { ForecastChart } from "./ForecastChart";

function StatTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-semibold tracking-tight">{value}</p>
    </div>
  );
}

export function DashboardView() {
  const t = useTranslations();
  const format = useFormatter();
  const isClient = useIsClient();
  const countsQuery = useStudyCounts();
  const statsQuery = useStatsOverview();
  // Until hydration finishes, render exactly what the server rendered (loading state).
  const counts = isClient ? countsQuery : { data: undefined, isPending: true };
  const stats = isClient ? statsQuery : { data: undefined, isPending: true };

  const total = counts.data?.total;
  const due = total ? total.new + total.learning + total.review : 0;
  const s = stats.data;
  const hasCards = !!s && s.totals.new + s.totals.learning + s.totals.review + s.totals.suspended > 0;

  return (
    <div className="mx-auto max-w-5xl">
      <DashboardGreeting />

      {(counts.isPending || stats.isPending) && (
        <div className="space-y-4">
          <Skeleton className="h-40 w-full rounded-2xl" />
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            {[0, 1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-20 rounded-xl" />
            ))}
          </div>
        </div>
      )}

      {s && total && !hasCards && (
        <section className="flex flex-col items-start gap-4 rounded-2xl border border-border bg-card p-8 sm:p-10">
          <span className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
            <CalendarCheck className="size-5" strokeWidth={1.75} aria-hidden />
          </span>
          <div>
            <h2 className="text-lg font-semibold tracking-tight">{t("dashboard.emptyTitle")}</h2>
            <p className="mt-1.5 max-w-[60ch] text-sm leading-relaxed text-muted-foreground">{t("dashboard.emptyBody")}</p>
          </div>
          <Link href="/decks" className={buttonVariants()}>
            {t("dashboard.emptyCta")}
          </Link>
        </section>
      )}

      {s && total && hasCards && (
        <div className="space-y-6">
          <section className="grid grid-cols-1 gap-6 rounded-2xl border border-border bg-card p-6 sm:p-8 lg:grid-cols-[1fr_auto] lg:items-center">
            <div>
              <p className="text-5xl font-semibold tracking-tight sm:text-6xl">{format.number(due)}</p>
              <p className="mt-2 text-lg font-medium">{t("stats.dueTitle", { count: due })}</p>
              <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                <SrsCounts counts={total} />
                <span className="inline-flex items-center gap-1.5">
                  <Flame className="size-4" strokeWidth={1.75} aria-hidden />
                  {t("stats.streak", { count: s.streak_days })}
                </span>
              </div>
            </div>
            <Link
              href={due > 0 ? "/study" : "/decks"}
              className={buttonVariants({ size: "lg", variant: due > 0 ? "default" : "outline", className: "h-12 px-6 active:translate-y-px" })}
            >
              {due > 0 ? t("stats.startStudying") : t("dashboard.emptyCta")}
            </Link>
          </section>

          <section aria-labelledby="today-heading">
            <h2 id="today-heading" className="sr-only">
              {t("dashboard.statsTitle")}
            </h2>
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              <StatTile label={t("stats.todayReviews")} value={format.number(s.today.reviews)} />
              <StatTile label={t("stats.todayNew")} value={format.number(s.today.new_cards)} />
              <StatTile label={t("stats.todayTime")} value={t("stats.minutes", { n: Math.round(s.today.duration_ms / 60000) })} />
              <StatTile
                label={t("stats.retention")}
                value={s.retention_30d === null ? "-" : format.number(s.retention_30d, { style: "percent", maximumFractionDigits: 0 })}
              />
            </div>
            {s.retention_30d === null && <p className="mt-2 text-xs text-muted-foreground">{t("stats.retentionEmpty")}</p>}
          </section>

          <section className="rounded-2xl border border-border bg-card p-6">
            <h2 className="mb-6 text-base font-semibold tracking-tight">{t("stats.forecastTitle")}</h2>
            <ForecastChart days={s.forecast} />
          </section>
        </div>
      )}
    </div>
  );
}
