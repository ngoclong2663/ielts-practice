"use client";

import { X } from "lucide-react";
import { useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { buttonVariants } from "@/components/ui/button";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useListeningAudio } from "../../hooks/use-listening-audio";
import { formatClock } from "../../lib/format";
import { partOf } from "../../lib/questions";
import type { TAttempt, TLine, TResult } from "../../types";
import { AudioBar } from "../player/AudioBar";
import { QuestionNavigator } from "../player/QuestionNavigator";
import { PartExplanations } from "./ReviewView";

/** Listening review (Epic 10 §6.3): explanations beside the transcript, which follows the audio. */
export function ListeningReview({ attempt: a }: { attempt: TAttempt }) {
  const t = useTranslations("exam");
  const locale = useLocale() as "vi" | "en";
  const sp = useSearchParams();
  const parts = a.test.parts;
  const results = useMemo(() => Object.fromEntries(a.results.map((r) => [r.n, r])) as Record<number, TResult>, [a.results]);
  const inRun = useCallback((n: number) => n in results, [results]);
  const first = Number(sp.get("q")) || a.results[0]?.n || parts[0].q_from;
  const [current, setCurrent] = useState(inRun(first) ? first : (a.results[0]?.n ?? first));
  const [tab, setTab] = useState<"left" | "right">("left");
  const audioRef = useRef<HTMLAudioElement>(null);
  const part = partOf(parts, current) ?? parts[0];
  const audio = useListeningAudio({
    audioRef,
    parts,
    mode: "practice",
    fullTest: false,
    currentPartNo: part.part_no,
    saved: undefined,
    onSave: () => {},
    onPartChange: () => {},
    onFinished: () => {},
  });
  const lines = useMemo(() => part.content.lines ?? [], [part]);
  const nowLine = audio.part.part_no === part.part_no ? lineAt(lines, audio.time) : -1;
  const pendingPlay = useRef<number | null>(null);

  // Questions marked on each transcript line, and the answer words to underline.
  const marks = useMemo(() => {
    const byLine: Record<string, { n: number; quote: string }[]> = {};
    for (const [n, k] of Object.entries(part.key ?? {})) {
      if (k.evidence) (byLine[k.evidence.block] ??= []).push({ n: Number(n), quote: k.evidence.quote });
    }
    return byLine;
  }, [part]);

  const playFrom = useCallback(
    (at: number) => {
      audio.seek(Math.max(0, at - 1.5));
      audio.play();
    },
    [audio],
  );

  const listenAgain = (n: number) => {
    const p = partOf(parts, n);
    const ev = p?.key?.[String(n)]?.evidence;
    const line = p?.content.lines?.find((l) => l.id === ev?.block);
    if (!p || !line) return;
    setCurrent(n);
    setTab("right");
    if (audio.part.part_no !== p.part_no) {
      // The other part's audio loads first; play once it is ready.
      pendingPlay.current = line.at;
      return;
    }
    playFrom(line.at);
    document.getElementById(`line-${line.id}`)?.scrollIntoView({ block: "center", behavior: "smooth" });
  };

  useEffect(() => {
    if (pendingPlay.current === null || audio.part.part_no !== part.part_no || audio.duration === 0) return;
    const at = pendingPlay.current;
    pendingPlay.current = null;
    playFrom(at);
  }, [audio.part.part_no, audio.duration, part.part_no, playFrom]);

  // Keep the line being spoken in view.
  useEffect(() => {
    if (nowLine >= 0 && audio.phase === "playing") {
      document.getElementById(`line-${lines[nowLine].id}`)?.scrollIntoView({ block: "nearest", behavior: "smooth" });
    }
  }, [nowLine, audio.phase, lines]);

  const go = (n: number) => {
    setCurrent(n);
    setTab("left");
    window.requestAnimationFrame(() => document.getElementById(`rv-${n}`)?.scrollIntoView({ block: "start" }));
  };
  const order = a.results.map((r) => r.n);
  const step = (d: number) => {
    const next = order[Math.min(order.length - 1, Math.max(0, order.indexOf(current) + d))];
    if (next !== undefined) go(next);
  };

  return (
    <div className="flex h-dvh flex-col bg-background">
      <header className="flex h-13 shrink-0 items-center gap-2 border-b border-border bg-card px-3">
        <Link href={`/listening/attempts/${a.id}`} className={buttonVariants({ variant: "ghost", size: "sm" })} aria-label={t("result.backToScore")}>
          <X aria-hidden />
          <span className="hidden sm:inline">{t("result.backToScore")}</span>
        </Link>
        <p className="min-w-0 flex-1 truncate text-sm font-semibold">
          {a.book_title[locale] || a.book_title.en} · {t("testNo", { n: a.test_no })} · {t("review.title")}
        </p>
        <span className="inline-flex h-6 items-center rounded-full bg-accent px-2 text-xs font-medium text-accent-foreground tabular-nums">
          {t("review.score", { correct: a.correct, total: a.total })}
        </span>
      </header>

      {audio.part.audio && (
        <audio key={audio.part.part_no} ref={audioRef} src={audio.part.audio.url} preload="auto" {...audio.handlers}>
          <track kind="captions" />
        </audio>
      )}
      <AudioBar audio={audio} mode="practice" showPausesToggle={false} />

      <div className="shrink-0 border-b border-border bg-card px-4 py-2 md:hidden">
        <div className="grid grid-cols-2 gap-1 rounded-lg bg-muted p-1" role="tablist">
          {(["left", "right"] as const).map((side) => (
            <button
              key={side}
              type="button"
              role="tab"
              aria-selected={tab === side}
              onClick={() => setTab(side)}
              className={cn("h-8 rounded-md text-sm font-medium text-muted-foreground", tab === side && "bg-background text-foreground shadow-xs")}
            >
              {side === "left" ? t("review.explanationsTab") : t("review.transcript")}
            </button>
          ))}
        </div>
      </div>

      <div className="relative min-h-0 flex-1 md:grid md:grid-cols-2">
        <section
          aria-label={t("review.explanationsTab")}
          className={cn("absolute inset-0 overflow-y-auto px-4 pt-4 pb-20 md:static md:px-7 md:pt-6", tab === "right" && "hidden md:block")}
        >
          <PartExplanations
            part={part}
            attempt={a}
            results={results}
            current={current}
            onLocate={listenAgain}
            locateLabel={(p, block) => t("review.listenAgain", { time: formatClock(p.content.lines?.find((l) => l.id === block)?.at ?? 0) })}
          />
        </section>
        <section
          aria-label={t("review.transcript")}
          className={cn(
            "absolute inset-0 overflow-y-auto border-border bg-card px-3 pt-4 pb-20 md:static md:border-l md:px-5 md:pt-6",
            tab === "left" && "hidden md:block",
          )}
        >
          <h2 className="mb-3 px-2 text-[15px] font-semibold">
            {t("review.transcript")} · {t("partLabel.listening", { n: part.part_no })}
          </h2>
          <ol className="grid gap-0.5">
            {lines.map((l, i) => (
              <li key={l.id}>
                <button
                  type="button"
                  id={`line-${l.id}`}
                  onClick={() => (audio.part.part_no === part.part_no ? playFrom(l.at + 1.5) : undefined)}
                  className={cn(
                    "grid w-full grid-cols-[4.5rem_minmax(0,1fr)_2.75rem] gap-2 rounded-lg px-2 py-1.5 text-left text-[15px] leading-relaxed hover:bg-muted",
                    i === nowLine && "bg-accent hover:bg-accent",
                  )}
                >
                  <span className="pt-0.5 text-xs font-semibold text-muted-foreground">{l.speaker === "NARRATOR" ? "" : l.speaker}</span>
                  <span>{underline(l, marks[l.id] ?? [])}</span>
                  <span className="pt-0.5 text-right text-xs font-semibold text-primary tabular-nums">
                    {(marks[l.id] ?? []).map((m) => `Q${m.n}`).join(" ")}
                  </span>
                </button>
              </li>
            ))}
          </ol>
        </section>
      </div>

      <QuestionNavigator
        skill="listening"
        parts={parts}
        current={current}
        inRun={inRun}
        answered={() => true}
        flagged={() => false}
        correct={(n) => results[n]?.correct}
        onGo={go}
        onStep={step}
      />
    </div>
  );
}

function lineAt(lines: TLine[], t: number): number {
  let idx = -1;
  lines.forEach((l, i) => {
    if (t >= l.at) idx = i;
  });
  return idx;
}

/** Underlines the evidence quotes inside a transcript line. */
function underline(line: TLine, marks: { quote: string }[]): ReactNode {
  const spans: { start: number; end: number }[] = [];
  for (const m of marks) {
    const start = line.text.indexOf(m.quote);
    if (start >= 0) spans.push({ start, end: start + m.quote.length });
  }
  if (!spans.length) return line.text;
  spans.sort((x, y) => x.start - y.start);
  const out: ReactNode[] = [];
  let pos = 0;
  spans.forEach((s, i) => {
    if (s.start < pos) return;
    out.push(line.text.slice(pos, s.start));
    out.push(
      <u key={i} className="decoration-primary decoration-2 underline-offset-4">
        {line.text.slice(s.start, s.end)}
      </u>,
    );
    pos = s.end;
  });
  out.push(line.text.slice(pos));
  return out;
}
