"use client";

import { Pause, Play, RotateCcw } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import { SPEEDS, type useListeningAudio } from "../../hooks/use-listening-audio";
import { formatClock } from "../../lib/format";
import type { TListenMode } from "../../types";

type AudioBarProps = {
  audio: ReturnType<typeof useListeningAudio>;
  mode: TListenMode;
  /** Review screens allow every control, whatever mode the attempt used. */
  showPausesToggle?: boolean;
};

/** Play controls under the top bar (Epic 10 §6.1). Exam mode: one start button, no pause or seek. */
export function AudioBar({ audio, mode, showPausesToggle = true }: AudioBarProps) {
  const t = useTranslations("exam.audio");
  const exam = mode === "exam";
  const playing = audio.phase === "playing";
  const pct = audio.duration ? Math.min(100, (audio.time / audio.duration) * 100) : 0;

  return (
    <div className="shrink-0 border-b border-border bg-card">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-2 px-4 py-2.5">
        {exam ? (
          <Button size="icon-lg" className="rounded-full" onClick={audio.play} disabled={!audio.canPlay} aria-label={audio.started ? t("resume") : t("start")}>
            <Play aria-hidden />
          </Button>
        ) : (
          <>
            <Button size="icon-lg" className="rounded-full" onClick={playing ? audio.pause : audio.play} aria-label={playing ? t("pause") : t("play")}>
              {playing ? <Pause aria-hidden /> : <Play aria-hidden />}
            </Button>
            <Button variant="ghost" size="sm" onClick={() => audio.seek(audio.time - 5)} aria-label={t("back5")}>
              <RotateCcw aria-hidden />
              5s
            </Button>
          </>
        )}
        <div className="flex min-w-40 flex-1 items-center gap-3">
          {exam ? (
            <span className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted" aria-hidden>
              <span className="block h-full rounded-full bg-primary" style={{ width: `${pct}%` }} />
            </span>
          ) : (
            <input
              type="range"
              min={0}
              max={audio.duration || 0}
              step={0.1}
              value={audio.time}
              onChange={(e) => audio.seek(Number(e.target.value))}
              aria-label={t("position")}
              className="flex-1 accent-primary"
            />
          )}
          <span className="text-sm text-muted-foreground tabular-nums">
            {formatClock(audio.time)} / {formatClock(audio.duration)}
          </span>
        </div>
        {exam ? (
          <span className="text-[13px] text-muted-foreground">{t("examNote")}</span>
        ) : (
          <div className="flex items-center gap-3">
            <div className="flex rounded-lg bg-muted p-0.5" role="radiogroup" aria-label={t("speed")}>
              {SPEEDS.map((s) => (
                <button
                  key={s}
                  type="button"
                  role="radio"
                  aria-checked={audio.speed === s}
                  onClick={() => audio.setSpeed(s)}
                  className={cn(
                    "h-7 rounded-md px-2 text-[13px] text-muted-foreground tabular-nums",
                    audio.speed === s && "bg-background font-semibold text-foreground shadow-xs",
                  )}
                >
                  {s}×
                </button>
              ))}
            </div>
            {showPausesToggle && (
              <label className="flex items-center gap-2 text-[13px] text-muted-foreground">
                <Switch checked={audio.keepPauses} onCheckedChange={audio.setKeepPauses} />
                {t("keepPauses")}
              </label>
            )}
          </div>
        )}
      </div>
      {audio.error && (
        <p role="alert" className="flex items-center gap-3 border-t border-border bg-destructive/10 px-4 py-2 text-sm text-destructive">
          {t("error")}
          <Button variant="outline" size="sm" onClick={audio.retry}>
            {t("retry")}
          </Button>
        </p>
      )}
      {audio.cue && (
        <p role="status" className="flex items-center gap-3 border-t border-border bg-amber-50 px-4 py-2 text-sm font-medium text-amber-900 dark:bg-amber-500/15 dark:text-amber-200">
          <span className="flex-1">{t("cue", { label: audio.cue.label })}</span>
          <span className="tabular-nums">{formatClock(audio.cue.left)}</span>
          {!exam && (
            <Button variant="ghost" size="sm" onClick={audio.skipCue}>
              {t("skip")}
            </Button>
          )}
        </p>
      )}
      {audio.checkLeft !== null && (
        <p role="status" className="flex items-center gap-3 border-t border-border bg-accent px-4 py-2 text-sm font-medium text-accent-foreground">
          <span className="flex-1">{t("check")}</span>
          <span className="tabular-nums">{formatClock(audio.checkLeft)}</span>
        </p>
      )}
    </div>
  );
}
