"use client";

import { RefreshCw } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { useAdminOverview } from "../hooks/use-admin";
import { AICostSection } from "./AICostSection";
import { DailyChart } from "./DailyChart";
import { QueryError } from "./QueryError";
import { StatSection, StatTile } from "./StatTile";

export function OverviewView() {
  const t = useTranslations("admin.overview");
  const tReason = useTranslations("vocab.reportReason");
  const format = useFormatter();
  const isClient = useIsClient();
  const query = useAdminOverview();
  const o = isClient ? query.data : undefined;

  const n = (value: number) => format.number(value);
  const compact = (value: number) => format.number(value, { notation: "compact", maximumFractionDigits: 1 });

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader
        title={t("title")}
        description={
          o
            ? t("updated", { time: format.dateTime(new Date(o.generated_at), { hour: "2-digit", minute: "2-digit" }), timezone: o.timezone })
            : t("subtitle")
        }
        actions={
          <Button variant="outline" size="sm" onClick={() => void query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={query.isFetching ? "size-4 animate-spin" : "size-4"} strokeWidth={1.75} aria-hidden />
            {t("refresh")}
          </Button>
        }
      />

      {query.isError && <QueryError onRetry={() => void query.refetch()} />}

      {!o && !query.isError && (
        <div className="space-y-6">
          {[0, 1, 2].map((i) => (
            <div key={i} className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              {[0, 1, 2, 3].map((j) => (
                <Skeleton key={j} className="h-24 rounded-xl" />
              ))}
            </div>
          ))}
        </div>
      )}

      {o && (
        <div className="space-y-8">
          <StatSection title={t("sections.users")}>
            <StatTile label={t("users.total")} value={n(o.users.total)} hint={t("users.newHint", { week: n(o.users.new_7d), month: n(o.users.new_30d) })} href="/admin/users" />
            <StatTile label={t("users.active1d")} value={n(o.users.active_1d)} />
            <StatTile label={t("users.active7d")} value={n(o.users.active_7d)} />
            <StatTile label={t("users.active30d")} value={n(o.users.active_30d)} />
          </StatSection>

          <section className="rounded-2xl border border-border bg-card p-6">
            <h2 className="mb-4 text-base font-semibold tracking-tight">{t("daily.title")}</h2>
            <DailyChart days={o.daily} />
          </section>

          <StatSection title={t("sections.feedback")}>
            <StatTile
              label={t("feedback.new")}
              value={n(o.feedback.new)}
              hint={t("feedback.totalHint", { count: n(o.feedback.total) })}
              href="/admin/feedback"
              highlight={o.feedback.new > 0}
            />
            <StatTile label={t("reports.total")} value={n(o.content_reports.total)} hint={t("reports.weekHint", { count: n(o.content_reports.last_7d) })} />
            <StatTile label={t("reports.needsReview")} value={n(o.content_reports.needs_review)} hint={t("reports.needsReviewHint")} highlight={o.content_reports.needs_review > 0} />
          </StatSection>

          <StatSection title={t("sections.vocab")}>
            <StatTile label={t("vocab.decks")} value={n(o.vocab.decks)} />
            <StatTile
              label={t("vocab.cards")}
              value={n(o.vocab.cards)}
              hint={t("vocab.cardsHint", {
                new: n(o.vocab.cards_by_state.new),
                learning: n(o.vocab.cards_by_state.learning + o.vocab.cards_by_state.relearning),
                review: n(o.vocab.cards_by_state.review),
              })}
            />
            <StatTile label={t("vocab.reviews7d")} value={n(o.vocab.reviews_7d)} hint={t("vocab.reviewsTotalHint", { count: n(o.vocab.reviews_total) })} />
            <StatTile label={t("vocab.lexemes")} value={n(o.vocab.lexemes_cached)} hint={t("vocab.lexemesHint")} />
          </StatSection>

          <StatSection title={t("sections.ai")}>
            <StatTile
              label={t("ai.wordsGenerated")}
              value={n(o.ai.words_generated_total)}
              hint={t("ai.wordsHint", { week: n(o.ai.words_generated_7d), cached: n(o.ai.words_cached_total) })}
            />
            <StatTile label={t("ai.wordsFailed7d")} value={n(o.ai.words_failed_7d)} highlight={o.ai.words_failed_7d > 0} />
            <StatTile label={t("ai.topicLists")} value={n(o.ai.topic_lists_total)} hint={t("ai.weekHint", { count: n(o.ai.topic_lists_7d) })} />
            <StatTile
              label={t("ai.tokens30d")}
              value={compact(o.ai.input_tokens_30d + o.ai.output_tokens_30d)}
              hint={t("ai.tokensHint", { input: compact(o.ai.input_tokens_30d), output: compact(o.ai.output_tokens_30d) })}
            />
            <StatTile label={t("ai.writingChecks")} value={n(o.ai.writing_checks_total)} hint={t("ai.weekHint", { count: n(o.ai.writing_checks_7d) })} />
            <StatTile label={t("ai.writingReviews")} value={n(o.ai.writing_reviews_total)} hint={t("ai.weekHint", { count: n(o.ai.writing_reviews_7d) })} />
            <StatTile label={t("ai.writingLessons")} value={n(o.ai.writing_lessons_total)} hint={t("ai.lessonsHint")} />
          </StatSection>

          <AICostSection cost={o.ai_cost} />

          <StatSection title={t("sections.writing")}>
            <StatTile label={t("writing.attempts")} value={n(o.writing.attempts_total)} />
            <StatTile
              label={t("writing.submitted")}
              value={n(o.writing.attempts_submitted)}
              hint={
                o.writing.attempts_total > 0
                  ? t("writing.submittedHint", {
                      percent: format.number(o.writing.attempts_submitted / o.writing.attempts_total, { style: "percent", maximumFractionDigits: 0 }),
                    })
                  : undefined
              }
            />
            <StatTile
              label={t("writing.avgBand")}
              value={o.writing.avg_band_30d === null ? "-" : format.number(o.writing.avg_band_30d, { minimumFractionDigits: 1, maximumFractionDigits: 1 })}
              hint={t("writing.avgBandHint")}
            />
          </StatSection>

          <section>
            <h2 className="mb-3 text-sm font-semibold tracking-tight text-muted-foreground">{t("reports.recentTitle")}</h2>
            {o.recent_reports.length === 0 ? (
              <p className="rounded-xl border border-dashed border-border p-6 text-center text-sm text-muted-foreground">{t("reports.empty")}</p>
            ) : (
              <ul className="divide-y divide-border rounded-xl border border-border bg-card">
                {o.recent_reports.map((r) => (
                  <li key={r.id} className="flex flex-col gap-1 px-4 py-3 sm:flex-row sm:items-baseline sm:gap-4">
                    <span className="font-medium">{r.headword}</span>
                    <span className="text-sm text-muted-foreground">{tReason.has(r.reason) ? tReason(r.reason) : r.reason}</span>
                    {r.note && <span className="min-w-0 flex-1 truncate text-sm">“{r.note}”</span>}
                    <span className="text-xs text-muted-foreground sm:ml-auto">
                      {r.user_email} · {format.relativeTime(new Date(r.created_at))}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </div>
      )}
    </div>
  );
}
