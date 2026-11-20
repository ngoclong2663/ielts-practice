"use client";

import { CircleCheck } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { PageHeader } from "@/components/shared/PageHeader";
import { GuideCard } from "@/features/guides/components/GuideCard";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { useTasks } from "../hooks/use-writing";
import { formatBand } from "../lib/format";
import type { TTaskNo, TTaskSummary, TTaskType } from "../types";
import { Figure } from "./TaskPrompt";

export function TaskListView({ taskNo }: { taskNo: TTaskNo }) {
  const t = useTranslations("writing");
  const isClient = useIsClient();
  const query = useTasks(taskNo);
  const tasks = isClient ? query.data : undefined;
  const [type, setType] = useState<TTaskType | null>(null);

  const types = [...new Set(tasks?.map((task) => task.task_type) ?? [])];
  const shown = tasks?.filter((task) => !type || task.task_type === type) ?? [];

  return (
    <div className="mx-auto max-w-5xl">
      <PageHeader title={t(`task${taskNo}.title`)} description={t(`task${taskNo}.subtitle`)} />
      <GuideCard topic="writing" className="mb-6" />

      {types.length > 1 && (
        <div className="mb-6 flex flex-wrap gap-2" role="group" aria-label={t("filterByType")}>
          {[null, ...types].map((value) => (
            <button
              key={value ?? "all"}
              type="button"
              aria-pressed={type === value}
              onClick={() => setType(value)}
              className={cn(
                "rounded-full border px-3.5 py-1.5 text-sm transition-colors",
                type === value ? "border-primary bg-primary/10 font-medium text-primary" : "border-border hover:bg-muted",
              )}
            >
              {value ? t(`types.${value}`) : t("allTypes")}
            </button>
          ))}
        </div>
      )}

      {isClient && query.isError ? (
        <div className="rounded-xl border border-border p-6">
          <p className="text-sm text-muted-foreground">{t("loadError")}</p>
          <Button variant="outline" className="mt-4" onClick={() => query.refetch()}>
            {t("retry")}
          </Button>
        </div>
      ) : !tasks ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-64 rounded-2xl" />
          ))}
        </div>
      ) : shown.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
          {t("noTasks")}
        </p>
      ) : (
        <ul className="grid gap-4 sm:grid-cols-2">
          {shown.map((task) => (
            <li key={task.id}>
              <TaskCard task={task} />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function TaskCard({ task }: { task: TTaskSummary }) {
  const t = useTranslations("writing");
  return (
    <Link
      href={`/writing/tasks/${task.id}`}
      className="group flex h-full flex-col overflow-hidden rounded-2xl border border-border bg-card transition-colors hover:border-primary/50"
    >
      {task.figure && (
        <div className="border-b border-border bg-white p-3">
          <Figure figure={task.figure} className="max-h-44 rounded-none border-0 object-contain" />
        </div>
      )}
      <div className="flex flex-1 flex-col gap-2 p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">{t(`types.${task.task_type}`)}</Badge>
          {task.best_band != null ? (
            <Badge variant="outline" className="gap-1">
              <CircleCheck className="size-3.5 text-primary" strokeWidth={2} aria-hidden />
              {t("bestBand", { band: formatBand(task.best_band) })}
            </Badge>
          ) : (
            task.attempts > 0 && <Badge variant="outline">{t("inProgress")}</Badge>
          )}
        </div>
        <h2 className="font-medium leading-snug group-hover:text-primary">{task.title}</h2>
      </div>
    </Link>
  );
}
