"use client";

import { X } from "lucide-react";
import { useParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useAttempt } from "../../hooks/use-exam";
import { formatClock } from "../../lib/format";
import { groupOf } from "../../lib/questions";
import type { TAttempt, TGroupType, TResult } from "../../types";

/** The answer key's accepted forms for one question, as shown to the learner. */
export function correctText(attempt: TAttempt, n: number): string {
  const part = attempt.test.parts.find((p) => n >= p.q_from && n <= p.q_to);
  const item = part?.key?.[String(n)];
  return item ? item.answer.join(" / ") : "";
}

export function typeStats(results: TResult[]) {
  const map = new Map<TGroupType, { type: TGroupType; correct: number; total: number }>();
  for (const r of results) {
    const s = map.get(r.type) ?? { type: r.type, correct: 0, total: 0 };
    s.total++;
    if (r.correct) s.correct++;
    map.set(r.type, s);
  }
  return [...map.values()];
}

/** /<skill>/attempts/[attemptId]: score, accuracy per question type, the answer table and next steps (Epic 10 §7.1). */
export function ResultView() {
  const t = useTranslations("exam");
  const locale = useLocale() as "vi" | "en";
  const { attemptId } = useParams<{ attemptId: string }>();
  const isClient = useIsClient();
  const query = useAttempt(attemptId);
  const a = isClient ? query.data : undefined;

  if (query.isError) {
    return (
      <div className="mx-auto mt-24 max-w-md px-4 text-center">
        <p className="text-muted-foreground">{t("loadError")}</p>
        <Link href="/reading" className={buttonVariants({ variant: "outline", className: "mt-4" })}>
          {t("backToTests")}
        </Link>
      </div>
    );
  }
  if (!a) return <Skeleton className="mx-auto mt-16 h-96 max-w-3xl rounded-xl" />;

  const skill = a.skill;
  const wrong = a.results.filter((r) => !r.correct).map((r) => r.n);
  const stats = typeStats(a.results);
  const weakest = [...stats].sort((x, y) => x.correct / x.total - y.correct / y.total)[0];
  const scope =
    a.kind === "full"
      ? t("fullTest")
      : a.kind === "redo"
        ? t("result.redoScope")
        : a.part_nos.map((n) => t(`partLabel.${skill}`, { n })).join(", ");
  const bookTitle = a.book_title[locale] || a.book_title.en;

  return (
    <div className="min-h-dvh bg-background">
      <header className="flex h-13 items-center gap-2 border-b border-border bg-card px-3">
        <Link href={`/${skill}`} className={buttonVariants({ variant: "ghost", size: "sm" })} aria-label={t("exit")}>
          <X aria-hidden />
          <span className="hidden sm:inline">{t("backToTests")}</span>
        </Link>
        <p className="min-w-0 flex-1 truncate text-sm font-semibold">
          {bookTitle} · {t("testNo", { n: a.test_no })}
        </p>
      </header>

      <main className="mx-auto max-w-4xl px-4 py-8">
        <h1 className="text-2xl font-semibold tracking-tight">{t("result.title", { scope })}</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {a.time_limit_sec
            ? t("result.timeVsLimit", { spent: formatClock(a.duration_sec), limit: formatClock(a.time_limit_sec) })
            : t("result.time", { spent: formatClock(a.duration_sec) })}
        </p>

        <div className="mt-6 grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)]">
          <section className="rounded-xl border border-border bg-card p-5" aria-labelledby="score-h">
            <h2 id="score-h" className="mb-3 text-[15px] font-semibold">
              {t("result.score")}
            </h2>
            <p className="text-4xl font-bold tracking-tight tabular-nums">
              {a.correct}
              <span className="ml-1 text-lg font-medium text-muted-foreground">{t("result.ofCorrect", { total: a.total })}</span>
            </p>
            {a.kind === "full" ? (
              <p className="mt-3 text-sm">
                {a.band !== null ? (
                  <>
                    <span className="font-semibold">{t("result.band", { band: a.band.toFixed(1) })}</span>{" "}
                    <span className="text-muted-foreground">{t("result.bandNote")}</span>
                  </>
                ) : (
                  t("result.belowBand")
                )}
              </p>
            ) : (
              <p className="mt-3 text-sm text-muted-foreground">{t("result.noBand")}</p>
            )}
            {weakest && weakest.correct < weakest.total && (
              <div className="mt-4 border-t border-border pt-4 text-sm">
                <p className="font-semibold">{t("result.weakest", { type: t(`types.${weakest.type}`) })}</p>
                <p className="text-muted-foreground">{t("result.weakestDetail", { correct: weakest.correct, total: weakest.total })}</p>
              </div>
            )}
          </section>

          <section className="rounded-xl border border-border bg-card p-5" aria-labelledby="types-h">
            <h2 id="types-h" className="mb-2 text-[15px] font-semibold">
              {t("result.byType")}
            </h2>
            <ul>
              {stats.map((s) => (
                <li key={s.type} className="grid grid-cols-[minmax(0,1fr)_7rem_3rem] items-center gap-3 border-t border-border py-2 first:border-t-0">
                  <span className="truncate text-sm">{t(`types.${s.type}`)}</span>
                  <span className="h-2 overflow-hidden rounded-full bg-muted" aria-hidden>
                    <span
                      className={cn("block h-full rounded-full", s.correct / s.total < 0.5 ? "bg-destructive" : "bg-primary")}
                      style={{ width: `${(s.correct / s.total) * 100}%` }}
                    />
                  </span>
                  <span className="text-right text-sm tabular-nums">
                    {s.correct}/{s.total}
                  </span>
                </li>
              ))}
            </ul>
          </section>
        </div>

        <div className="mt-6 flex flex-wrap gap-2">
          <Link href={`/${skill}/attempts/${a.id}/review`} className={buttonVariants({ size: "lg" })}>
            {t("result.review")}
          </Link>
          {wrong.length > 0 && (
            <Link
              href={`/${skill}/practice/${a.test_slug}?redo=${a.id}&q=${wrong.join(",")}`}
              className={buttonVariants({ variant: "outline", size: "lg" })}
            >
              {t("result.redo", { count: wrong.length })}
            </Link>
          )}
          <Link href={`/${skill}`} className={buttonVariants({ variant: "outline", size: "lg" })}>
            {t("result.another")}
          </Link>
        </div>

        <section className="mt-6 rounded-xl border border-border bg-card p-5" aria-labelledby="answers-h">
          <h2 id="answers-h" className="mb-2 text-[15px] font-semibold">
            {t("result.answers")}
          </h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-muted-foreground">
                  <th className="py-2 pr-2 font-medium">{t("result.colQuestion")}</th>
                  <th className="py-2 pr-2 font-medium">{t("result.colYours")}</th>
                  <th className="py-2 pr-2 font-medium">{t("result.colAnswer")}</th>
                  <th className="hidden py-2 pr-2 font-medium sm:table-cell">{t("result.colType")}</th>
                  <th className="py-2" />
                </tr>
              </thead>
              <tbody>
                {a.results.map((r) => {
                  const g = groupOf(a.test.parts, r.n);
                  return (
                    <tr key={r.n} className="border-t border-border align-top">
                      <td className="py-2.5 pr-2 tabular-nums">{r.n}</td>
                      <td className={cn("py-2.5 pr-2 font-semibold", r.correct ? "text-primary" : "text-destructive")}>
                        {r.given || "—"}
                        {r.over_limit && <span className="ml-1 text-xs font-normal text-muted-foreground">({t("result.overLimit")})</span>}
                      </td>
                      <td className="py-2.5 pr-2">
                        {correctText(a, r.n)}
                        {g?.type === "choose_n" && <span className="ml-1 text-xs text-muted-foreground">({t("result.anyOrder")})</span>}
                      </td>
                      <td className="hidden py-2.5 pr-2 text-muted-foreground sm:table-cell">{t(`types.${r.type}`)}</td>
                      <td className="py-1.5 text-right">
                        <Link href={`/${skill}/attempts/${a.id}/review?q=${r.n}`} className={buttonVariants({ variant: "ghost", size: "sm" })}>
                          {t("result.explain")}
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>
      </main>
    </div>
  );
}
