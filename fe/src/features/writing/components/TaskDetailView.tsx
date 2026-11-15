"use client";

import { ArrowLeft, LoaderCircle, Play } from "lucide-react";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link, useRouter } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useStartAttempt, useTask, useWritingStatus } from "../hooks/use-writing";
import { toastApiError } from "../lib/errors";
import { formatBandRange } from "../lib/format";
import type { TAttemptSummary, TLevel, TPosition, TTaskDetail } from "../types";
import { SourceLabel, TaskPrompt } from "./TaskPrompt";

const LEVELS: TLevel[] = [1, 2, 3, 4];
const LEVEL_KEY = "writing.level";

function savedLevel(): TLevel {
  try {
    const n = Number(localStorage.getItem(LEVEL_KEY));
    return (LEVELS as number[]).includes(n) ? (n as TLevel) : 1;
  } catch {
    return 1;
  }
}

export function TaskDetailView() {
  const t = useTranslations("writing");
  const { taskId } = useParams<{ taskId: string }>();
  const isClient = useIsClient();
  const query = useTask(taskId);
  const task = isClient ? query.data : undefined;

  return (
    <div className="mx-auto max-w-6xl">
      <Link
        href={task?.task_no === 2 ? "/writing/task-2" : "/writing/task-1"}
        className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" strokeWidth={1.75} aria-hidden />
        {t("backToTasks")}
      </Link>
      {isClient && query.isError ? (
        <p className="rounded-xl border border-border p-6 text-sm text-muted-foreground">{t("taskLoadError")}</p>
      ) : !task ? (
        <Skeleton className="h-96 rounded-2xl" />
      ) : (
        <>
          <div className="mb-6">
            <Badge variant="secondary" className="mb-2">
              {t(`taskLabel${task.task_no}`)} · {t(`types.${task.task_type}`)}
            </Badge>
            <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">{task.title}</h1>
            <SourceLabel source={task.source} className="mt-2" />
          </div>
          <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
            <TaskPrompt task={task} />
            <div className="space-y-6">
              <StartPanel task={task} />
              {task.history.length > 0 && <History items={task.history} />}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

function StartPanel({ task }: { task: TTaskDetail }) {
  const t = useTranslations();
  const router = useRouter();
  const status = useWritingStatus().data;
  const start = useStartAttempt(task.id);
  const [level, setLevel] = useState<TLevel>(savedLevel);
  const [position, setPosition] = useState<TPosition | null>(task.positions.length === 1 ? task.positions[0] : null);
  const needsPosition = task.task_no === 2 && task.positions.length > 1;

  const begin = () => {
    try {
      localStorage.setItem(LEVEL_KEY, String(level));
    } catch {
      // storage blocked: the level is not remembered
    }
    start.mutate(
      { level, ...(task.task_no === 2 && position ? { position } : {}) },
      {
        onSuccess: (a) => router.push(`/writing/attempts/${a.id}`),
        onError: (e) => toastApiError(t, e, "writing-start"),
      },
    );
  };

  return (
    <section className="space-y-5 rounded-2xl border border-border bg-card p-5">
      <fieldset>
        <legend className="mb-3 font-medium">{t("writing.levelLabel")}</legend>
        <div className="space-y-2">
          {LEVELS.map((l) => (
            <label
              key={l}
              className={cn(
                "flex cursor-pointer gap-3 rounded-xl border p-3 transition-colors",
                level === l ? "border-primary bg-primary/5" : "border-border hover:bg-muted/50",
              )}
            >
              <input
                type="radio"
                name="level"
                value={l}
                checked={level === l}
                onChange={() => setLevel(l)}
                className="mt-1 accent-[var(--primary)]"
              />
              <span>
                <span className="block text-sm font-medium">{t(`writing.levels.${l}.name`)}</span>
                <span className="block text-xs text-muted-foreground">{t(`writing.levels.${l}.help`)}</span>
              </span>
            </label>
          ))}
        </div>
      </fieldset>

      {needsPosition && (
        <fieldset>
          <legend className="mb-1 font-medium">{t("writing.positionLabel")}</legend>
          <p className="mb-3 text-xs text-muted-foreground">{t("writing.positionHelp")}</p>
          <div className="flex flex-wrap gap-2">
            {task.positions.map((p) => (
              <button
                key={p}
                type="button"
                aria-pressed={position === p}
                onClick={() => setPosition(p)}
                className={cn(
                  "rounded-full border px-3.5 py-1.5 text-sm transition-colors",
                  position === p ? "border-primary bg-primary/10 font-medium text-primary" : "border-border hover:bg-muted",
                )}
              >
                {t(`writing.positions.${p}`)}
              </button>
            ))}
          </div>
        </fieldset>
      )}

      {status && !status.configured && <p className="text-sm text-destructive">{t("writing.notConfigured")}</p>}
      <Button
        size="lg"
        className="w-full"
        onClick={begin}
        disabled={start.isPending || (needsPosition && !position) || status?.configured === false}
      >
        {start.isPending ? (
          <LoaderCircle className="size-4 animate-spin" strokeWidth={1.75} aria-hidden />
        ) : (
          <Play className="size-4" strokeWidth={1.75} aria-hidden />
        )}
        {start.isPending ? t("writing.preparing") : t("writing.start")}
      </Button>
      {start.isPending && <p className="text-xs text-muted-foreground">{t("writing.preparingHelp")}</p>}
    </section>
  );
}

function History({ items }: { items: TAttemptSummary[] }) {
  const t = useTranslations("writing");
  return (
    <section>
      <h2 className="mb-3 font-medium">{t("yourAttempts")}</h2>
      <ul className="divide-y divide-border rounded-2xl border border-border">
        {items.map((a) => (
          <li key={a.id}>
            <AttemptRow attempt={a} />
          </li>
        ))}
      </ul>
    </section>
  );
}

export function AttemptRow({ attempt: a, showTask = false }: { attempt: TAttemptSummary; showTask?: boolean }) {
  const t = useTranslations("writing");
  return (
    <Link href={`/writing/attempts/${a.id}`} className="flex items-center gap-3 px-4 py-3 text-sm hover:bg-muted/50">
      <div className="min-w-0 flex-1">
        {showTask && <p className="truncate font-medium">{a.task_title}</p>}
        <p className="text-muted-foreground">
          {showTask && `${t(`taskLabel${a.task_no}`)} · `}
          {t(`levels.${a.level}.name`)}
          {a.position && a.position !== "default" && ` · ${t(`positions.${a.position}`)}`}
          {" · "}
          {new Date(a.created_at).toLocaleDateString()}
        </p>
      </div>
      {a.band_low != null && a.band_high != null ? (
        <Badge>{t("band", { band: formatBandRange(a.band_low, a.band_high) })}</Badge>
      ) : (
        <Badge variant="outline">{t("inProgress")}</Badge>
      )}
    </Link>
  );
}
