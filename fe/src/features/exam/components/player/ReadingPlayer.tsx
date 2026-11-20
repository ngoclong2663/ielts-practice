"use client";

import { AArrowUp, Clock, X } from "lucide-react";
import { useParams, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useTest } from "../../hooks/use-exam";
import { TEXT_SIZES, useExamRun } from "../../hooks/use-exam-run";
import { formatClock } from "../../lib/format";
import { runSetup } from "../../lib/run-setup";
import type { TRunConfig } from "../../lib/run-state";
import type { TPart, TTest } from "../../types";
import { QuestionGroup } from "../questions/QuestionGroup";
import { Passage } from "./Passage";
import { QuestionNavigator } from "./QuestionNavigator";
import { SelectionToolbar } from "./SelectionToolbar";
import { SplitView } from "./SplitView";

/**
 * /reading/practice/[test]?parts=1,2&time=2400 (part practice), ?kind=full (full test),
 * ?redo=<attempt>&q=6,9 (the wrong questions of an earlier attempt).
 */
export function ReadingPlayer() {
  const t = useTranslations("exam");
  const { test: slug } = useParams<{ test: string }>();
  const sp = useSearchParams();
  const isClient = useIsClient();
  const query = useTest(slug, "reading");
  const test = isClient ? query.data : undefined;

  const setup = useMemo(() => (test ? runSetup(test, sp, "reading") : null), [test, sp]);

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
  if (!test || !setup) return <PlayerSkeleton />;
  if (setup.parts.length === 0) {
    return (
      <div className="mx-auto mt-24 max-w-md px-4 text-center">
        <p className="text-muted-foreground">{t("noPartsChosen")}</p>
        <Link href="/reading" className={buttonVariants({ variant: "outline", className: "mt-4" })}>
          {t("backToTests")}
        </Link>
      </div>
    );
  }
  return <ReadingRun key={JSON.stringify(setup.config)} test={test} parts={setup.parts} config={setup.config} />;
}

function PlayerSkeleton() {
  return (
    <div className="flex h-dvh flex-col">
      <Skeleton className="h-13 rounded-none" />
      <div className="grid flex-1 gap-6 p-6 md:grid-cols-2">
        <Skeleton className="h-full rounded-xl" />
        <Skeleton className="hidden h-full rounded-xl md:block" />
      </div>
    </div>
  );
}

function ReadingRun({ test, parts, config }: { test: TTest; parts: TPart[]; config: TRunConfig }) {
  const t = useTranslations("exam");
  const { root, ...r } = useExamRun(test, parts, config);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [timeHidden, setTimeHidden] = useState(false);
  const { state, run } = r;
  const flagged = state.flags.filter(r.inRun).length;
  const timed = run.remainingSec !== null;
  const clock = timed ? Math.max(0, run.remainingSec ?? 0) : run.elapsedSec;
  const low = timed && clock <= 600;

  const handlers = {
    onAnswer: run.setAnswer,
    onToggleLetter: run.toggleLetter,
    onFlag: run.toggleFlag,
    onFocus: run.setCurrent,
  };

  return (
    <div ref={root} className="flex h-dvh flex-col bg-background">
      <header className="flex h-13 shrink-0 items-center gap-2 border-b border-border bg-card px-3">
        <Link href="/reading" className={buttonVariants({ variant: "ghost", size: "sm" })} aria-label={t("exit")}>
          <X aria-hidden />
          <span className="hidden sm:inline">{t("exit")}</span>
        </Link>
        <p className="min-w-0 flex-1 truncate text-sm font-semibold">{r.title}</p>
        <button
          type="button"
          onClick={() => setTimeHidden((v) => !v)}
          aria-label={timed ? t("timeLeftHint") : t("timeSpentHint")}
          className={cn(
            "flex h-8 items-center gap-1.5 rounded-md bg-muted px-2.5 text-sm font-semibold tabular-nums",
            low && "bg-destructive/10 text-destructive",
          )}
        >
          <Clock className="size-4" aria-hidden />
          {!timeHidden && <span>{formatClock(clock)}</span>}
        </button>
        <span className="sr-only" aria-live="polite">
          {timed && (clock === 600 || clock === 300) ? t("minutesLeft", { count: clock / 60 }) : ""}
        </span>
        <DropdownMenu>
          <DropdownMenuTrigger aria-label={t("display")} className={buttonVariants({ variant: "ghost", size: "icon" })}>
            <AArrowUp aria-hidden />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuGroup>
              <DropdownMenuLabel>{t("textSize")}</DropdownMenuLabel>
              <DropdownMenuRadioGroup value={String(r.size)} onValueChange={(v) => r.setSize(Number(v))}>
                {TEXT_SIZES.map((s, i) => (
                  <DropdownMenuRadioItem key={s} value={String(s)}>
                    {t(`size.${["small", "medium", "large"][i]}`)}
                  </DropdownMenuRadioItem>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => run.setHighlights([])} disabled={state.highlights.length === 0}>
              {t("highlight.clearAll")}
            </DropdownMenuItem>
            <p className="px-2 pt-1 pb-1.5 text-xs text-muted-foreground">
              {r.hl.supported ? t("highlight.hint") : t("highlight.unsupported")}
            </p>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button onClick={() => setConfirmOpen(true)} disabled={r.submitting}>
          {t("submit")}
        </Button>
      </header>

      <div className="shrink-0 border-b border-border bg-card px-4 py-2 md:hidden">
        <div className="grid grid-cols-2 gap-1 rounded-lg bg-muted p-1" role="tablist">
          {(["left", "right"] as const).map((side) => (
            <button
              key={side}
              type="button"
              role="tab"
              aria-selected={r.show === side}
              onClick={() => r.setShow(side)}
              className={cn("h-8 rounded-md text-sm font-medium text-muted-foreground", r.show === side && "bg-background text-foreground shadow-xs")}
            >
              {side === "left"
                ? t("passageTab")
                : t("questionsTab", { done: r.questions.filter(r.answered).length, total: r.questions.length })}
            </button>
          ))}
        </div>
      </div>

      {run.timeUp && config.kind !== "full" && (
        <p role="status" className="shrink-0 bg-amber-50 px-4 py-2 text-sm text-amber-900 dark:bg-amber-500/15 dark:text-amber-200">
          {t("timeUpPractice")}
        </p>
      )}

      <SplitView
        show={r.show}
        leftLabel={t("passageTab")}
        rightLabel={t("questionsTabPlain")}
        onResize={r.onResize}
        left={<Passage part={r.currentPart} size={r.size} />}
        right={r.currentPart.content.groups
          .filter((g) => r.questions.some((n) => n >= g.from && n <= g.to))
          .map((g) => (
            <QuestionGroup
              key={g.from}
              group={g}
              answers={state.answers}
              current={state.current}
              flags={state.flags}
              locked={(n) => !r.inRun(n)}
              handlers={handlers}
            />
          ))}
      />

      <QuestionNavigator
        skill="reading"
        parts={parts}
        current={state.current}
        inRun={r.inRun}
        answered={r.answered}
        flagged={(n) => state.flags.includes(n)}
        onGo={r.go}
        onStep={r.step}
      />

      <SelectionToolbar
        toolbar={r.hl.toolbar}
        onAdd={() => r.hl.toolbar?.kind === "add" && r.hl.add(r.hl.toolbar.pieces)}
        onRemove={() => r.hl.toolbar?.kind === "remove" && r.hl.removeOne(r.hl.toolbar.target)}
        onDismiss={r.hl.dismiss}
      />

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("confirm.title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {r.unanswered ? t("confirm.unanswered", { count: r.unanswered }) : t("confirm.allAnswered")}
              {flagged > 0 && ` ${t("confirm.flagged", { count: flagged })}`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("confirm.keepGoing")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setConfirmOpen(false);
                r.submit();
              }}
              disabled={r.submitting}
            >
              {t("submit")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
