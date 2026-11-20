"use client";

import { MapPin, Play, X } from "lucide-react";
import { useParams, useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useAttempt } from "../../hooks/use-exam";
import { useHighlights } from "../../hooks/use-highlights";
import { HL_EVIDENCE, paint } from "../../lib/highlights";
import { answerMode, partOf } from "../../lib/questions";
import { loadAttemptHighlights, saveAttemptHighlights } from "../../lib/run-state";
import type { TAttempt, TGroup, THighlight, TPart, TResult } from "../../types";
import { Passage } from "../player/Passage";
import { QuestionNavigator } from "../player/QuestionNavigator";
import { SelectionToolbar } from "../player/SelectionToolbar";
import { SplitView } from "../player/SplitView";
import { QNum, QuestionGroup } from "../questions/QuestionGroup";
import { correctText } from "../result/ResultView";
import { ListeningReview } from "./ListeningReview";

/** /reading/attempts/[attemptId]/review?q=5: the passage beside the explanations (Epic 10 §7.2). */
export function ReviewView() {
  const t = useTranslations("exam");
  const { attemptId } = useParams<{ attemptId: string }>();
  const isClient = useIsClient();
  const query = useAttempt(attemptId);
  const a = isClient ? query.data : undefined;
  if (query.isError) {
    return (
      <div className="mx-auto mt-24 max-w-md px-4 text-center">
        <p className="text-muted-foreground">{t("loadError")}</p>
      </div>
    );
  }
  if (!a) return <Skeleton className="mx-auto mt-16 h-96 max-w-5xl rounded-xl" />;
  return a.skill === "listening" ? <ListeningReview attempt={a} /> : <Review attempt={a} />;
}

function Review({ attempt: a }: { attempt: TAttempt }) {
  const t = useTranslations("exam");
  const locale = useLocale() as "vi" | "en";
  const sp = useSearchParams();
  const parts = a.test.parts;
  const results = useMemo(() => Object.fromEntries(a.results.map((r) => [r.n, r])) as Record<number, TResult>, [a.results]);
  const inRun = useCallback((n: number) => n in results, [results]);
  const first = Number(sp.get("q")) || a.results[0]?.n || parts[0].q_from;
  const [current, setCurrent] = useState(inRun(first) ? first : (a.results[0]?.n ?? first));
  const [show, setShow] = useState<"left" | "right">("right");
  const [evidence, setEvidence] = useState<THighlight[]>([]);
  const [mine, setMine] = useState<THighlight[]>(() => loadAttemptHighlights(a.id));
  const root = useRef<HTMLDivElement>(null);
  const part = partOf(parts, current) ?? parts[0];

  const saveMine = useCallback(
    (next: THighlight[]) => {
      setMine(next);
      saveAttemptHighlights(a.id, next);
    },
    [a.id],
  );
  const hl = useHighlights(root, mine, saveMine, `${show}-${part.part_no}`);

  useEffect(() => {
    paint(HL_EVIDENCE, root.current, evidence);
  }, [evidence, show, part.part_no]);
  useEffect(() => () => paint(HL_EVIDENCE, null, []), []);

  const locate = useCallback(
    (n: number) => {
      const p = partOf(parts, n);
      const ev = p?.key?.[String(n)]?.evidence;
      if (!p || !ev) return;
      const para = p.content.paragraphs?.find((x) => x.id === ev.block);
      const start = para?.text.indexOf(ev.quote) ?? -1;
      setCurrent(n);
      setShow("left");
      if (start < 0) return;
      setEvidence([{ block: `p-${ev.block}`, start, end: start + ev.quote.length }]);
      window.requestAnimationFrame(() =>
        document.getElementById(`para-${ev.block}`)?.scrollIntoView({
          block: "center",
          behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth",
        }),
      );
    },
    [parts],
  );

  const go = (n: number) => {
    setCurrent(n);
    setShow("right");
    window.requestAnimationFrame(() => document.getElementById(`rv-${n}`)?.scrollIntoView({ block: "start" }));
  };

  // Opening ?q=5 scrolls straight to that explanation (current already starts there).
  useEffect(() => {
    if (!sp.get("q")) return;
    const id = window.requestAnimationFrame(() => document.getElementById(`rv-${first}`)?.scrollIntoView({ block: "start" }));
    return () => window.cancelAnimationFrame(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- once, on open
  }, []);

  const order = a.results.map((r) => r.n);
  const step = (d: number) => {
    const i = order.indexOf(current);
    const next = order[Math.min(order.length - 1, Math.max(0, i + d))];
    if (next !== undefined) go(next);
  };

  return (
    <div ref={root} className="flex h-dvh flex-col bg-background">
      <header className="flex h-13 shrink-0 items-center gap-2 border-b border-border bg-card px-3">
        <Link href={`/${a.skill}/attempts/${a.id}`} className={buttonVariants({ variant: "ghost", size: "sm" })} aria-label={t("result.backToScore")}>
          <X aria-hidden />
          <span className="hidden sm:inline">{t("result.backToScore")}</span>
        </Link>
        <p className="min-w-0 flex-1 truncate text-sm font-semibold">
          {(a.book_title[locale] || a.book_title.en)} · {t("testNo", { n: a.test_no })} · {t("review.title")}
        </p>
        <span className="inline-flex h-6 items-center rounded-full bg-accent px-2 text-xs font-medium text-accent-foreground tabular-nums">
          {t("review.score", { correct: a.correct, total: a.total })}
        </span>
      </header>

      <div className="shrink-0 border-b border-border bg-card px-4 py-2 md:hidden">
        <div className="grid grid-cols-2 gap-1 rounded-lg bg-muted p-1" role="tablist">
          {(["left", "right"] as const).map((side) => (
            <button
              key={side}
              type="button"
              role="tab"
              aria-selected={show === side}
              onClick={() => setShow(side)}
              className={cn("h-8 rounded-md text-sm font-medium text-muted-foreground", show === side && "bg-background text-foreground shadow-xs")}
            >
              {side === "left" ? t("passageTab") : t("review.explanationsTab")}
            </button>
          ))}
        </div>
      </div>

      <SplitView
        show={show}
        leftLabel={t("passageTab")}
        rightLabel={t("review.explanationsTab")}
        left={<Passage part={part} size={17} />}
        right={
          <PartExplanations
            part={part}
            attempt={a}
            results={results}
            current={current}
            onLocate={locate}
            locateLabel={(p, block) => t("review.locate", { block: blockLabel(p, block) })}
          />
        }
      />

      <QuestionNavigator
        skill={a.skill}
        parts={parts}
        current={current}
        inRun={inRun}
        answered={() => true}
        flagged={() => false}
        correct={(n) => results[n]?.correct}
        onGo={go}
        onStep={step}
      />

      <SelectionToolbar
        toolbar={hl.toolbar}
        onAdd={() => hl.toolbar?.kind === "add" && hl.add(hl.toolbar.pieces)}
        onRemove={() => hl.toolbar?.kind === "remove" && hl.removeOne(hl.toolbar.target)}
        onDismiss={hl.dismiss}
      />
    </div>
  );
}

export function PartExplanations({ part, attempt, results, current, onLocate, locateLabel }: {
  part: TPart;
  attempt: TAttempt;
  results: Record<number, TResult>;
  current: number;
  onLocate: (n: number) => void;
  locateLabel: (part: TPart, block: string) => string;
}) {
  return (
    <>
      {part.content.groups
        .filter((g) => Object.keys(results).some((n) => Number(n) >= g.from && Number(n) <= g.to))
        .map((g) => (
          <section key={g.from} className="mb-8">
            {(answerMode(g.type) === "gap" || g.type === "summary_bank") ? (
              <QuestionGroup group={g} answers={attempt.answers} results={results} />
            ) : (
              <GroupHeader group={g} />
            )}
            {Array.from({ length: g.to - g.from + 1 }, (_, i) => g.from + i)
              .filter((n) => n in results)
              .map((n) => (
                <Explanation key={n} n={n} group={g} part={part} attempt={attempt} result={results[n]} current={current === n} onLocate={onLocate} locateLabel={locateLabel} />
              ))}
          </section>
        ))}
    </>
  );
}

function GroupHeader({ group }: { group: TGroup }) {
  const t = useTranslations("exam");
  return (
    <>
      <h3 className="text-[15px] font-semibold">
        {group.from === group.to ? t("question", { n: group.from }) : t("questions", { from: group.from, to: group.to })}
      </h3>
      <p className="mt-1 mb-3 text-sm text-muted-foreground">{group.instructions}</p>
      {(group.options ?? []).some((o) => o.text) && group.type !== "choose_n" && (
        <dl className="mb-3 grid grid-cols-[1.75rem_minmax(0,1fr)] gap-x-2 gap-y-1 rounded-xl border border-border bg-card p-3 text-sm">
          {(group.options ?? []).map((o) => (
            <div key={o.key} className="contents">
              <dt className="font-semibold">{o.key}</dt>
              <dd>{o.text}</dd>
            </div>
          ))}
        </dl>
      )}
    </>
  );
}

/** The question's own text: its statement, or for a gap the line (or table row) it sits in, with this gap
 * as a blank and other gaps as "…". */
function stemOf(group: TGroup, n: number): string {
  const item = group.items?.find((i) => i.n === n);
  if (item) return item.text;
  if (group.type === "choose_n") return group.stem ?? "";
  const mark = `{${n}}`;
  const line = group.lines?.find((l) => l.includes(mark)) ?? group.rows?.find((r) => r.some((c) => c.includes(mark)))?.join(" · ");
  if (!line) return "";
  // A summary is one long paragraph: keep only the sentence around the gap.
  const sentence = line.split(/(?<=[.!?])\s+/).find((x) => x.includes(mark)) ?? line;
  return sentence.replace(mark, "______").replace(/\{\d+\}/g, "…");
}

function optionText(group: TGroup, n: number, key: string): string {
  const opts = group.items?.find((i) => i.n === n)?.options ?? group.options ?? [];
  const o = opts.find((x) => x.key === key);
  return o?.text ? `${key} · ${o.text}` : key;
}

function Explanation({ n, group, part, attempt, result, current, onLocate, locateLabel }: {
  n: number;
  group: TGroup;
  part: TPart;
  attempt: TAttempt;
  result: TResult;
  current: boolean;
  onLocate: (n: number) => void;
  locateLabel: (part: TPart, block: string) => string;
}) {
  const t = useTranslations("exam");
  const k = part.key?.[String(n)];
  const stem = stemOf(group, n);
  const showOption = answerMode(group.type) !== "gap" && group.type !== "tfng" && group.type !== "ynng";
  return (
    <div id={`rv-${n}`} className={cn("scroll-mt-4 border-t border-border py-3", current && "bg-accent/30 -mx-2 rounded-lg px-2")}>
      <div className="grid grid-cols-[2.25rem_minmax(0,1fr)] gap-2">
        <QNum n={n} result={result} />
        <p className="pt-0.5 leading-relaxed" data-hl={`rq${n}`}>
          {stem || t("review.gap", { n })}
        </p>
      </div>
      <div className="mt-2 ml-11 grid gap-2 text-sm">
        <div className="flex flex-wrap gap-x-5 gap-y-1">
          <p>
            <span className="text-muted-foreground">{t("review.yours")} </span>
            <b className={result.correct ? "text-primary" : "text-destructive"}>
              {result.given ? (showOption ? optionText(group, n, result.given) : result.given) : "—"}
            </b>
            {result.over_limit && <span className="ml-1 text-xs text-muted-foreground">({t("result.overLimit")})</span>}
          </p>
          <p>
            <span className="text-muted-foreground">{t("review.answer")} </span>
            <b>{showOption && k?.answer.length === 1 ? optionText(group, n, k.answer[0]) : correctText(attempt, n)}</b>
          </p>
        </div>
        {k && (
          <div className="rounded-lg bg-muted px-3 py-2.5">
            <p lang="en">{k.why}</p>
            {k.pairs && k.pairs.length > 0 && (
              <div className="mt-2 grid gap-1">
                <p className="text-xs font-medium text-muted-foreground">{t("review.pairs")}</p>
                {k.pairs.map(([q, p], i) => (
                  <p key={i} lang="en" className="grid grid-cols-[minmax(0,1fr)_1.25rem_minmax(0,1fr)] gap-1.5 text-[13px]">
                    <span className="text-muted-foreground">{q}</span>
                    <span aria-hidden className="text-center">≈</span>
                    <span>{p}</span>
                  </p>
                ))}
              </div>
            )}
          </div>
        )}
        {k?.trap && (
          <p lang="en" className="text-[13px] text-amber-800 dark:text-amber-300">
            <b>{t("review.trap")}</b> {k.trap}
          </p>
        )}
        {k?.evidence && (
          <button
            type="button"
            onClick={() => onLocate(n)}
            className={buttonVariants({ variant: "outline", size: "sm", className: "justify-self-start" })}
          >
            {attempt.skill === "listening" ? <Play aria-hidden /> : <MapPin aria-hidden />}
            {locateLabel(part, k.evidence.block)}
          </button>
        )}
      </div>
    </div>
  );
}

function blockLabel(part: TPart, block: string): string {
  const paras = part.content.paragraphs ?? [];
  const p = paras.find((x) => x.id === block);
  return p?.label || String(paras.indexOf(p!) + 1);
}
