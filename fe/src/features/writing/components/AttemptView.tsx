"use client";

import { ArrowLeft, Check } from "lucide-react";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useAttempt } from "../hooks/use-writing";
import type { TAttempt } from "../types";
import { FinalStep } from "./FinalStep";
import { ParagraphStep } from "./ParagraphStep";
import { PlanStep } from "./PlanStep";
import { ResultView } from "./ResultView";
import { SourceLabel, TaskPrompt } from "./TaskPrompt";
import { VocabStep } from "./VocabStep";

/** Step ids: "plan", "vocab", a paragraph key, "final" or "result". */
export type TStep = string;

export function stepsOf(a: TAttempt): TStep[] {
  const steps = ["plan", ...(a.vocab_available ? ["vocab"] : []), ...a.paragraphs.map((p) => p.key), "final"];
  return a.submission ? [...steps, "result"] : steps;
}

/** Where to resume: the result once submitted, else the first step that has not been done. */
function firstStep(a: TAttempt): TStep {
  if (a.submission) return "result";
  if (a.plan_answers == null && a.questions.length > 0) return "plan";
  return a.paragraphs.find((p) => p.checked_count === 0)?.key ?? "final";
}

function stepDone(a: TAttempt, step: TStep): boolean {
  if (step === "plan") return a.plan_answers != null || a.questions.length === 0;
  if (step === "final" || step === "result") return !!a.submission;
  // Vocabulary is reading, not a task: done once the learner has moved on to writing.
  if (step === "vocab") return !!a.submission || a.paragraphs.some((p) => p.checked_count > 0);
  return (a.paragraphs.find((p) => p.key === step)?.checked_count ?? 0) > 0;
}

export function AttemptView() {
  const t = useTranslations("writing");
  const { attemptId } = useParams<{ attemptId: string }>();
  const isClient = useIsClient();
  const query = useAttempt(attemptId);
  const attempt = isClient ? query.data : undefined;

  if (isClient && query.isError) {
    return <p className="mx-auto max-w-3xl rounded-xl border border-border p-6 text-sm text-muted-foreground">{t("attemptLoadError")}</p>;
  }
  if (!attempt) {
    return <Skeleton className="mx-auto h-96 max-w-6xl rounded-2xl" />;
  }
  return <AttemptScreen key={attempt.id} attempt={attempt} />;
}

function AttemptScreen({ attempt }: { attempt: TAttempt }) {
  const t = useTranslations("writing");
  const [step, setStep] = useState<TStep>(() => firstStep(attempt));
  const steps = stepsOf(attempt);
  const next = () => setStep(steps[Math.min(steps.indexOf(step) + 1, steps.length - 1)]);
  const task = attempt.task;

  return (
    <div className="mx-auto max-w-6xl">
      <Link
        href={`/writing/tasks/${task.id}`}
        className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" strokeWidth={1.75} aria-hidden />
        {t("backToTask")}
      </Link>
      <div className="mb-5 flex flex-wrap items-center gap-2">
        <h1 className="mr-2 text-xl font-semibold tracking-tight sm:text-2xl">{task.title}</h1>
        <Badge variant="secondary">{t(`levels.${attempt.level}.name`)}</Badge>
        {attempt.position && attempt.position !== "default" && (
          <Badge variant="outline">{t(`positions.${attempt.position}`)}</Badge>
        )}
      </div>

      <nav aria-label={t("steps")} className="-mx-4 mb-6 overflow-x-auto px-4">
        <ol className="flex min-w-max gap-2">
          {steps.map((s, i) => {
            const done = stepDone(attempt, s);
            return (
              <li key={s}>
                <button
                  type="button"
                  onClick={() => setStep(s)}
                  aria-current={step === s ? "step" : undefined}
                  className={cn(
                    "inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm transition-colors",
                    step === s
                      ? "border-primary bg-primary text-primary-foreground"
                      : done
                        ? "border-primary/40 text-primary hover:bg-primary/5"
                        : "border-border text-muted-foreground hover:bg-muted",
                  )}
                >
                  {done && step !== s ? (
                    <Check className="size-3.5" strokeWidth={2.5} aria-hidden />
                  ) : (
                    <span className="tabular-nums">{i + 1}.</span>
                  )}
                  {t(`stepNames.${s}`)}
                </button>
              </li>
            );
          })}
        </ol>
      </nav>

      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_24rem]">
        <div className="min-w-0">
          {/* On small screens the question sits above the work, folded. */}
          <details className="mb-6 rounded-2xl border border-border p-4 lg:hidden">
            <summary className="cursor-pointer text-sm font-medium">{t("showQuestion")}</summary>
            <div className="mt-4">
              <TaskPrompt task={task} compact />
            </div>
          </details>
          {step === "plan" && <PlanStep attempt={attempt} onNext={next} />}
          {step === "vocab" && <VocabStep attempt={attempt} onNext={next} />}
          {attempt.paragraphs.map(
            (p) => step === p.key && <ParagraphStep key={p.key} attempt={attempt} paragraph={p} onNext={next} />,
          )}
          {step === "final" && <FinalStep attempt={attempt} onSubmitted={() => setStep("result")} />}
          {step === "result" && <ResultView attempt={attempt} />}
        </div>
        <aside className="hidden lg:block">
          <div className="sticky top-6 max-h-[calc(100dvh-3rem)] space-y-3 overflow-y-auto rounded-2xl border border-border p-4">
            <TaskPrompt task={task} compact />
            <SourceLabel source={task.source} />
          </div>
        </aside>
      </div>
    </div>
  );
}
