"use client";

import { ArrowRight, CircleCheck, CircleX, ListChecks } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useAnswerPlan } from "../hooks/use-writing";
import { toastApiError } from "../lib/errors";
import type { TAttempt } from "../types";

/** Step 0 (in Vietnamese): understand the task, then see the plan of ideas. */
export function PlanStep({ attempt, onNext }: { attempt: TAttempt; onNext: () => void }) {
  const t = useTranslations();
  const answer = useAnswerPlan(attempt.id);
  const answered = attempt.plan_answers != null;
  const [choices, setChoices] = useState<(number | null)[]>(
    () => attempt.plan_answers ?? attempt.questions.map(() => null),
  );
  const complete = choices.every((c) => c != null);

  const submit = () =>
    answer.mutate(choices as number[], { onError: (e) => toastApiError(t, e, "writing-plan") });

  return (
    <section className="space-y-6">
      <header>
        <h2 className="text-lg font-semibold">{t("writing.plan.title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t("writing.plan.help")}</p>
      </header>

      {attempt.questions.map((q, qi) => (
        <fieldset key={qi} className="rounded-2xl border border-border p-4">
          <legend className="px-1 font-medium">
            {qi + 1}. {q.question_vi}
          </legend>
          <div className="mt-2 space-y-2">
            {q.options_vi.map((option, oi) => {
              const chosen = choices[qi] === oi;
              const correct = answered && q.answer === oi;
              const wrong = answered && chosen && q.answer !== oi;
              return (
                <label
                  key={oi}
                  className={cn(
                    "flex cursor-pointer items-start gap-3 rounded-xl border p-3 text-sm transition-colors",
                    correct && "border-emerald-500/60 bg-emerald-500/10",
                    wrong && "border-destructive/60 bg-destructive/10",
                    !answered && chosen && "border-primary bg-primary/5",
                    !answered && !chosen && "border-border hover:bg-muted/50",
                    answered && !correct && !wrong && "border-border opacity-70",
                  )}
                >
                  <input
                    type="radio"
                    name={`q${qi}`}
                    checked={chosen}
                    disabled={answered}
                    onChange={() => setChoices((c) => c.map((v, i) => (i === qi ? oi : v)))}
                    className="mt-0.5"
                  />
                  <span className="flex-1">{option}</span>
                  {correct && <CircleCheck className="size-4 text-emerald-600" strokeWidth={2} aria-label={t("writing.correct")} />}
                  {wrong && <CircleX className="size-4 text-destructive" strokeWidth={2} aria-label={t("writing.incorrect")} />}
                </label>
              );
            })}
          </div>
          {answered && q.explain_vi && <p className="mt-3 text-sm text-muted-foreground">{q.explain_vi}</p>}
        </fieldset>
      ))}

      {(answered || attempt.questions.length === 0) && attempt.ideas_vi.length > 0 && (
        <div className="rounded-2xl border border-primary/30 bg-primary/5 p-4">
          <h3 className="mb-2 flex items-center gap-2 font-medium">
            <ListChecks className="size-4" strokeWidth={1.75} aria-hidden />
            {t("writing.plan.ideas")}
          </h3>
          <ul className="list-disc space-y-1 pl-5 text-sm">
            {attempt.ideas_vi.map((idea, i) => (
              <li key={i}>{idea}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="flex justify-end gap-2">
        {!answered && attempt.questions.length > 0 ? (
          <Button onClick={submit} disabled={!complete || answer.isPending}>
            {t("writing.plan.check")}
          </Button>
        ) : (
          <Button onClick={onNext}>
            {t("writing.continue")}
            <ArrowRight className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
        )}
      </div>
    </section>
  );
}
