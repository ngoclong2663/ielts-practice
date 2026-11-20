"use client";

import { Clock, X } from "lucide-react";
import { useParams, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useMemo, useRef, useState } from "react";
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
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useTest } from "../../hooks/use-exam";
import { useExamRun } from "../../hooks/use-exam-run";
import { useListeningAudio } from "../../hooks/use-listening-audio";
import { formatClock } from "../../lib/format";
import { runSetup } from "../../lib/run-setup";
import type { TRunConfig } from "../../lib/run-state";
import type { TPart, TTest } from "../../types";
import { QuestionGroup } from "../questions/QuestionGroup";
import { AudioBar } from "./AudioBar";
import { QuestionNavigator } from "./QuestionNavigator";
import { SelectionToolbar } from "./SelectionToolbar";

/** /listening/practice/[test]?parts=1,2&mode=practice|exam, ?kind=full, ?redo=<attempt>&q=6,9 (Epic 10 §6). */
export function ListeningPlayer() {
  const t = useTranslations("exam");
  const { test: slug } = useParams<{ test: string }>();
  const sp = useSearchParams();
  const isClient = useIsClient();
  const query = useTest(slug, "listening");
  const test = isClient ? query.data : undefined;
  const setup = useMemo(() => (test ? runSetup(test, sp, "listening") : null), [test, sp]);

  if (query.isError || (setup && setup.parts.length === 0)) {
    return (
      <div className="mx-auto mt-24 max-w-md px-4 text-center">
        <p className="text-muted-foreground">{query.isError ? t("loadError") : t("noPartsChosen")}</p>
        <Link href="/listening" className={buttonVariants({ variant: "outline", className: "mt-4" })}>
          {t("backToTests")}
        </Link>
      </div>
    );
  }
  if (!test || !setup) {
    return (
      <div className="flex h-dvh flex-col">
        <Skeleton className="h-13 rounded-none" />
        <Skeleton className="h-16 rounded-none" />
        <Skeleton className="mx-auto mt-6 h-96 w-full max-w-3xl rounded-xl" />
      </div>
    );
  }
  return <ListeningRun key={JSON.stringify(setup.config)} test={test} parts={setup.parts} config={setup.config} />;
}

function ListeningRun({ test, parts, config }: { test: TTest; parts: TPart[]; config: TRunConfig }) {
  const t = useTranslations("exam");
  const { root, ...r } = useExamRun(test, parts, config);
  const { state, run } = r;
  const audioRef = useRef<HTMLAudioElement>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const exam = config.mode === "exam";

  // Exam mode moved on to the next part: show its first question (the React Compiler memoises this).
  const onPartChange = (p: TPart) => run.setCurrent(r.questions.find((n) => n >= p.q_from && n <= p.q_to) ?? p.q_from);
  const submit = r.submit;
  const audio = useListeningAudio({
    audioRef,
    parts,
    mode: config.mode,
    fullTest: config.kind === "full",
    currentPartNo: r.currentPart.part_no,
    saved: state.audio,
    onSave: run.setAudio,
    onPartChange,
    onFinished: submit,
  });
  // The navigator decides which part's questions show; in exam mode the audio moves it on to each new part.
  const shownPart = r.currentPart;
  const flagged = state.flags.filter(r.inRun).length;
  const handlers = { onAnswer: run.setAnswer, onToggleLetter: run.toggleLetter, onFlag: run.toggleFlag, onFocus: run.setCurrent };

  return (
    <div ref={root} className="flex h-dvh flex-col bg-background">
      <header className="flex h-13 shrink-0 items-center gap-2 border-b border-border bg-card px-3">
        <Link href="/listening" className={buttonVariants({ variant: "ghost", size: "sm" })} aria-label={t("exit")}>
          <X aria-hidden />
          <span className="hidden sm:inline">{t("exit")}</span>
        </Link>
        <p className="min-w-0 flex-1 truncate text-sm font-semibold">{r.title}</p>
        <span className="hidden h-6 items-center rounded-full bg-accent px-2 text-xs font-medium text-accent-foreground sm:inline-flex">
          {exam ? t("audio.modeExam") : t("audio.modePractice")}
        </span>
        <span className="flex h-8 items-center gap-1.5 rounded-md bg-muted px-2.5 text-sm font-semibold tabular-nums" aria-label={t("timeSpentHint")}>
          <Clock className="size-4" aria-hidden />
          {formatClock(run.elapsedSec)}
        </span>
        <Button onClick={() => setConfirmOpen(true)} disabled={r.submitting}>
          {t("submit")}
        </Button>
      </header>

      {/* One audio element per part (key), so a new part starts from a clean state. */}
      {audio.part.audio && (
        <audio key={audio.part.part_no} ref={audioRef} src={audio.part.audio.url} preload="auto" {...audio.handlers}>
          <track kind="captions" />
        </audio>
      )}
      <AudioBar audio={audio} mode={config.mode} />

      <main className="min-h-0 flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl px-4 pt-5 pb-24">
          <h1 className="mb-4 text-lg font-semibold">
            {t(`partLabel.listening`, { n: shownPart.part_no })}
            <span className="ml-2 font-normal text-muted-foreground">{shownPart.title}</span>
          </h1>
          {shownPart.content.groups
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
        </div>
      </main>

      <QuestionNavigator
        skill="listening"
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
              {audio.phase !== "ended" && audio.phase !== "check" && ` ${t("audio.stillPlaying")}`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("confirm.keepGoing")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setConfirmOpen(false);
                audioRef.current?.pause();
                submit();
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
