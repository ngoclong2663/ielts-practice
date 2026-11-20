import { useLocale, useTranslations } from "next-intl";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { useRouter } from "@/i18n/navigation";
import { keepHighlights, type TRunConfig } from "../lib/run-state";
import { groupOf, isAnswered, partOf, questionsOf } from "../lib/questions";
import type { TPart, TTest } from "../types";
import { useSubmitAttempt } from "./use-exam";
import { useHighlights } from "./use-highlights";
import { useRun } from "./use-run";

const SIZE_KEY = "exam.size";
export const TEXT_SIZES = [15, 17, 20] as const;

function readSize(): number {
  try {
    const v = Number(localStorage.getItem(SIZE_KEY));
    return (TEXT_SIZES as readonly number[]).includes(v) ? v : 17;
  } catch {
    return 17;
  }
}

/** Everything a player does between opening a test and submitting it (Epic 10 §5, §6), for either skill. */
export function useExamRun(test: TTest, parts: TPart[], config: TRunConfig) {
  const t = useTranslations("exam");
  const locale = useLocale() as "vi" | "en";
  const router = useRouter();
  const questions = useMemo(() => {
    const all = questionsOf(parts);
    return config.redo ? all.filter((n) => config.redo!.questions.includes(n)) : all;
  }, [parts, config.redo]);
  const inRun = useCallback((n: number) => questions.includes(n), [questions]);
  const run = useRun(config, questions[0] ?? parts[0].q_from);
  const { state } = run;

  const root = useRef<HTMLDivElement>(null);
  const [show, setShow] = useState<"left" | "right">("left");
  const [size, setSizeState] = useState(readSize);
  const [layoutTick, setLayoutTick] = useState(0);
  const hl = useHighlights(root, state.highlights, run.setHighlights, `${show}-${size}-${layoutTick}-${state.current}`);

  const setSize = (v: number) => {
    setSizeState(v);
    try {
      localStorage.setItem(SIZE_KEY, String(v));
    } catch {
      // storage blocked
    }
  };

  const currentPart = partOf(parts, state.current) ?? parts[0];
  const answered = useCallback((n: number) => isAnswered(state.answers, n, groupOf(parts, n)), [state.answers, parts]);
  const unanswered = questions.filter((n) => !answered(n)).length;

  const go = useCallback(
    (n: number) => {
      run.setCurrent(n);
      setShow("right");
      window.requestAnimationFrame(() => {
        const el = document.querySelector<HTMLElement>(`[data-q="${n}"]`);
        if (!el) return;
        el.scrollIntoView({ block: "center", behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth" });
        el.querySelector<HTMLElement>("input:not([type=checkbox]):not([type=radio]), button[role=radio]")?.focus({ preventScroll: true });
      });
    },
    [run],
  );
  const step = (delta: number) => {
    const i = questions.indexOf(state.current);
    const next = questions[Math.min(questions.length - 1, Math.max(0, (i < 0 ? 0 : i) + delta))];
    if (next !== undefined) go(next);
  };

  const submit = useSubmitAttempt();
  const doSubmit = useCallback(() => {
    if (submit.isPending) return;
    submit.mutate(
      {
        test: test.slug,
        skill: config.skill,
        mode: config.skill === "listening" ? config.mode : undefined,
        kind: config.kind,
        parts: config.kind === "redo" ? undefined : parts.map((p) => p.part_no),
        questions: config.redo?.questions,
        parent_attempt_id: config.redo?.attemptId,
        time_limit_sec: config.timeLimitSec,
        duration_sec: run.elapsedSec,
        answers: state.answers,
      },
      {
        onSuccess: ({ id }) => {
          keepHighlights(id, state.highlights);
          run.discard();
          router.replace(`/${config.skill}/attempts/${id}`);
        },
        onError: (e) => toast.error(t.has(`submitError.${e.code}`) ? t(`submitError.${e.code}`) : t("submitError.default")),
      },
    );
  }, [submit, test.slug, config, parts, run, state.answers, state.highlights, router, t]);

  // A full test ends when its time runs out, as in the exam. Part practice only warns.
  const autoSubmitted = useRef(false);
  useEffect(() => {
    if (run.timeUp && config.kind === "full" && !autoSubmitted.current) {
      autoSubmitted.current = true;
      doSubmit();
    }
  }, [run.timeUp, config.kind, doSubmit]);

  const title = [
    test.book_title[locale] || test.book_title.en,
    t("testNo", { n: test.test_no }),
    config.kind === "full" ? t("fullTest") : t(`partLabel.${config.skill}`, { n: currentPart.part_no }),
  ].join(" · ");

  return {
    root,
    run,
    state,
    questions,
    inRun,
    answered,
    unanswered,
    currentPart,
    show,
    setShow,
    size,
    setSize,
    hl,
    go,
    step,
    submit: doSubmit,
    submitting: submit.isPending,
    title,
    onResize: () => setLayoutTick((x) => x + 1),
  };
}
