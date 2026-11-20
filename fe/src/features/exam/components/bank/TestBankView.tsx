"use client";

import { ChevronDown } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { PageHeader } from "@/components/shared/PageHeader";
import { buttonVariants } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useBooks } from "../../hooks/use-exam";
import { READING_FULL_MINUTES, READING_MINUTES_PER_PART, partsParam } from "../../lib/format";
import { clearRun, describeRun, runsInProgress, type TRunInProgress } from "../../lib/run-state";
import type { TBook, TListenMode, TSkill, TTestSummary } from "../../types";

type Selection = { test: string; parts: number[] } | null;

// Which test blocks are open is a per-viewer preference, remembered per skill once the learner toggles one.
const OPEN_KEY = "exam.bank.open.";

function readOpen(skill: TSkill): string[] | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(OPEN_KEY + skill);
    return raw ? (JSON.parse(raw) as string[]) : null;
  } catch {
    return null;
  }
}

function writeOpen(skill: TSkill, slugs: string[]) {
  try {
    localStorage.setItem(OPEN_KEY + skill, JSON.stringify(slugs));
  } catch {
    // storage blocked: the choice lasts for this visit only
  }
}

/** Test bank for one skill (Epic 10 §11): pick parts of a test, a timer, and start. */
export function TestBankView({ skill }: { skill: TSkill }) {
  const t = useTranslations("exam");
  const locale = useLocale() as "vi" | "en";
  const isClient = useIsClient();
  const query = useBooks(skill);
  const books = isClient ? query.data : undefined;
  const [bookSlug, setBookSlug] = useState<string | null>(null);
  const [sel, setSel] = useState<Selection>(null);
  const [timer, setTimer] = useState<"suggested" | "none">("suggested");
  const [mode, setMode] = useState<TListenMode>("practice");
  const [runsVersion, setRunsVersion] = useState(0);
  // eslint-disable-next-line react-hooks/exhaustive-deps -- runsVersion re-reads storage after a run is discarded
  const runs = useMemo(() => (isClient ? runsInProgress(skill) : {}), [isClient, skill, runsVersion]);

  const book = books?.find((b) => b.slug === bookSlug) ?? books?.[0];
  const [opened, setOpened] = useState<string[] | null>(() => readOpen(skill));
  // Until the learner chooses, a test is open when it has a run in progress or ticked parts, or when it is the
  // first test not yet finished (Test 1 for a new learner).
  const firstUnfinished = book?.tests.find((x) => x.parts.some((p) => !p.best))?.slug ?? book?.tests[0]?.slug;
  const isOpen = (slug: string) =>
    opened ? opened.includes(slug) : slug === firstUnfinished || slug === sel?.test || (runs[slug]?.length ?? 0) > 0;
  const toggleOpen = (slug: string) => {
    const all = book?.tests.map((x) => x.slug) ?? [];
    const now = all.filter(isOpen);
    const next = now.includes(slug) ? now.filter((x) => x !== slug) : [...now, slug];
    setOpened(next);
    writeOpen(skill, next);
  };
  const toggle = (test: string, part: number) =>
    setSel((s) => {
      const parts = s?.test === test ? s.parts : [];
      const next = parts.includes(part) ? parts.filter((p) => p !== part) : [...parts, part];
      return next.length ? { test, parts: next } : null;
    });

  const minutes = (sel?.parts.length ?? 0) * READING_MINUTES_PER_PART;
  const startHref = !sel
    ? ""
    : skill === "listening"
      ? `/listening/practice/${sel.test}?parts=${partsParam(sel.parts)}&mode=${mode}`
      : `/reading/practice/${sel.test}?parts=${partsParam(sel.parts)}${timer === "suggested" ? `&time=${minutes * 60}` : ""}`;
  const questionCount =
    sel && book
      ? book.tests
          .find((x) => x.slug === sel.test)
          ?.parts.filter((p) => sel.parts.includes(p.part_no))
          .reduce((n, p) => n + p.q_to - p.q_from + 1, 0) ?? 0
      : 0;

  return (
    <div className="mx-auto max-w-4xl pb-28">
      <PageHeader title={t(`bank.title.${skill}`)} description={t(`bank.subtitle.${skill}`)} />

      {!books ? (
        <div className="grid gap-4">
          <Skeleton className="h-10 w-72 rounded-lg" />
          <Skeleton className="h-56 rounded-xl" />
          <Skeleton className="h-56 rounded-xl" />
        </div>
      ) : books.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">{t("bank.empty")}</div>
      ) : (
        <>
          {books.length > 1 && (
            <div role="tablist" aria-label={t("bank.books")} className="mb-2 flex gap-5 overflow-x-auto border-b border-border">
              {books.map((b) => (
                <button
                  key={b.slug}
                  type="button"
                  role="tab"
                  aria-selected={b.slug === book?.slug}
                  onClick={() => {
                    setBookSlug(b.slug);
                    setSel(null);
                  }}
                  className={cn(
                    "-mb-px shrink-0 border-b-2 border-transparent py-2.5 text-sm font-medium text-muted-foreground",
                    b.slug === book?.slug && "border-foreground text-foreground",
                  )}
                >
                  {b.title[locale] || b.title.en}
                </button>
              ))}
            </div>
          )}
          {book && (
            <BookProgress book={book} />
          )}
          <div className="grid gap-4">
            {book?.tests.map((test) => (
              <TestBlock
                key={test.slug}
                skill={skill}
                test={test}
                open={isOpen(test.slug)}
                onToggleOpen={() => toggleOpen(test.slug)}
                selected={sel?.test === test.slug ? sel.parts : []}
                onToggle={(p) => toggle(test.slug, p)}
                runs={(runs[test.slug] ?? []).map(describeRun).filter((x): x is TRunInProgress => !!x)}
                onDiscard={(key) => {
                  clearRun(key);
                  setRunsVersion((v) => v + 1);
                }}
              />
            ))}
          </div>
        </>
      )}

      {sel && (
        <div className="fixed inset-x-0 bottom-16 z-20 px-4 lg:bottom-4 lg:pl-64">
          <div
            role="region"
            aria-label={t("bank.start")}
            className="mx-auto flex max-w-4xl flex-wrap items-center gap-3 rounded-xl border border-border bg-card px-4 py-3 shadow-lg"
          >
            <p className="text-sm font-semibold">
              {t("bank.selected", { parts: sel.parts.length, questions: questionCount })}
            </p>
            {skill === "listening" ? (
              <Select
                items={[
                  { value: "practice", label: t("bank.modePractice") },
                  { value: "exam", label: t("bank.modeExam") },
                ]}
                value={mode}
                onValueChange={(v) => v && setMode(v as TListenMode)}
              >
                <SelectTrigger className="h-8 w-64" aria-label={t("bank.mode")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="practice">{t("bank.modePractice")}</SelectItem>
                  <SelectItem value="exam">{t("bank.modeExam")}</SelectItem>
                </SelectContent>
              </Select>
            ) : (
              <Select
                items={[
                  { value: "suggested", label: t("bank.timerSuggested", { minutes }) },
                  { value: "none", label: t("bank.timerNone") },
                ]}
                value={timer}
                onValueChange={(v) => v && setTimer(v as "suggested" | "none")}
              >
                <SelectTrigger className="h-8 w-48" aria-label={t("bank.timer")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="suggested">{t("bank.timerSuggested", { minutes })}</SelectItem>
                  <SelectItem value="none">{t("bank.timerNone")}</SelectItem>
                </SelectContent>
              </Select>
            )}
            <span className="flex-1" />
            <Link href={startHref} className={buttonVariants({ size: "lg" })}>
              {t("bank.startCta")}
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}

function BookProgress({ book }: { book: TBook }) {
  const t = useTranslations("exam");
  const all = book.tests.flatMap((x) => x.parts);
  const done = all.filter((p) => p.best).length;
  return (
    <p className="mb-4 text-sm text-muted-foreground">
      {t("bank.progress", { tests: book.tests.length, parts: all.length, done })}
    </p>
  );
}

type TestBlockProps = {
  skill: TSkill;
  test: TTestSummary;
  open: boolean;
  onToggleOpen: () => void;
  selected: number[];
  onToggle: (part: number) => void;
  runs: TRunInProgress[];
  onDiscard: (key: string) => void;
};

function TestBlock({ skill, test, open, onToggleOpen, selected, onToggle, runs, onDiscard }: TestBlockProps) {
  const t = useTranslations("exam");
  const questions = test.parts.reduce((n, p) => n + p.q_to - p.q_from + 1, 0);
  const done = test.parts.filter((p) => p.best);
  const listId = `parts-${test.slug}`;
  return (
    <section aria-labelledby={`test-${test.slug}`} className="rounded-xl border border-border bg-card">
      <div className="flex items-center justify-between gap-3 py-2 pr-4 pl-2">
        <h2 className="min-w-0 flex-1">
          <button
            type="button"
            onClick={onToggleOpen}
            aria-expanded={open}
            aria-controls={listId}
            className="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left outline-none hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring"
          >
            <ChevronDown className={cn("size-4 shrink-0 text-muted-foreground transition-transform", !open && "-rotate-90")} aria-hidden />
            <span className="min-w-0">
              <span id={`test-${test.slug}`} className="block text-[15px] font-semibold">
                {t("testNo", { n: test.test_no })}
              </span>
              <span className="block truncate text-[13px] font-normal text-muted-foreground">
                {t(`bank.testMeta.${skill}`, { parts: test.parts.length, questions, minutes: READING_FULL_MINUTES })}
                {done.length > 0 && ` · ${t("bank.testDone", { done: done.length, parts: test.parts.length })}`}
              </span>
            </span>
          </button>
        </h2>
        <Link href={`/${skill}/practice/${test.slug}?kind=full`} className={buttonVariants({ variant: "outline", size: "sm" })}>
          {t("bank.fullTest")}
        </Link>
      </div>

      {runs.map((run) => (
        <div key={run.key} className="flex flex-wrap items-center gap-3 border-t border-border bg-accent/50 px-4 py-2.5 text-sm">
          <span className="flex-1 text-accent-foreground">
            {run.kind === "full"
              ? t("bank.inProgressFull", { answered: run.answered })
              : run.kind === "redo"
                ? t("bank.inProgressRedo", { answered: run.answered })
                : t("bank.inProgress", { parts: run.parts.join(", "), answered: run.answered })}
          </span>
          <button type="button" onClick={() => onDiscard(run.key)} className="text-[13px] text-muted-foreground hover:text-foreground">
            {t("bank.discard")}
          </button>
          <Link href={continueHref(skill, test.slug, run)} className={buttonVariants({ size: "sm" })}>
            {t("bank.continue")}
          </Link>
        </div>
      ))}

      <ul id={listId} hidden={!open}>
        {test.parts.map((p) => {
          const id = `pick-${test.slug}-${p.part_no}`;
          return (
            <li key={p.part_no} className="border-t border-border">
              <label htmlFor={id} className="grid cursor-pointer grid-cols-[1.5rem_minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 hover:bg-muted/50 sm:grid-cols-[1.5rem_5.5rem_minmax(0,1fr)_auto]">
                <Checkbox id={id} checked={selected.includes(p.part_no)} onCheckedChange={() => onToggle(p.part_no)} />
                <span className="hidden text-[13px] text-muted-foreground sm:block">{t(`partLabel.${skill}`, { n: p.part_no })}</span>
                <span className="min-w-0">
                  <span className="line-clamp-2 block font-medium sm:truncate">{p.title}</span>
                  <span className="block truncate text-[13px] text-muted-foreground">
                    <span className="sm:hidden">{t(`partLabel.${skill}`, { n: p.part_no })} · </span>
                    {p.question_types.map((ty) => t(`types.${ty}`)).join(" · ")}
                  </span>
                </span>
                <span className="text-right text-[13px] tabular-nums">
                  {p.best ? (
                    <span className="inline-flex h-6 items-center rounded-full bg-accent px-2 font-medium text-accent-foreground">
                      {p.best.correct}/{p.best.total}
                    </span>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </span>
              </label>
            </li>
          );
        })}
      </ul>
    </section>
  );
}

function continueHref(skill: TSkill, test: string, run: TRunInProgress): string {
  const mode = skill === "listening" && run.mode ? `&mode=${run.mode}` : "";
  if (run.kind === "full") return `/${skill}/practice/${test}?kind=full${mode}`;
  if (run.kind === "redo") return `/${skill}/attempts/${run.redo}`;
  return `/${skill}/practice/${test}?parts=${partsParam(run.parts)}${run.timeSec ? `&time=${run.timeSec}` : ""}${mode}`;
}
