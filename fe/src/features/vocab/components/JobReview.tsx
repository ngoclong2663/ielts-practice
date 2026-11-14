"use client";

import { ArrowLeft } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { ApiError } from "@/lib/api-client";
import { useJob } from "../hooks/use-vocab";
import type { TJobItem } from "../types";
import { ImagePickerDialog } from "./ImagePickerDialog";
import { ItemCard } from "./ItemCard";
import { ItemDialog } from "./ItemDialog";
import { SaveBar } from "./SaveBar";

/** Live progress and the review grid for one generation (Epic 3 §3 steps 4–6). */
export function JobReview({ jobId, onNewList }: { jobId: string; onNewList: () => void }) {
  const t = useTranslations();
  const isClient = useIsClient();
  const query = useJob(jobId);
  const job = isClient ? query.data : undefined;
  const [openItemId, setOpenItemId] = useState<string | null>(null);
  const [pickingItemId, setPickingItemId] = useState<string | null>(null);

  const back = (
    <Button variant="ghost" size="sm" onClick={onNewList} className="-ml-2 mb-4">
      <ArrowLeft className="size-4" strokeWidth={1.75} aria-hidden />
      {t("vocab.newList")}
    </Button>
  );

  if (query.error) {
    const code = query.error instanceof ApiError ? query.error.code : "UNKNOWN";
    return (
      <div className="mx-auto max-w-5xl">
        {back}
        <p className="rounded-2xl border border-border bg-card p-6 text-muted-foreground">{t(`errors.${code}`)}</p>
      </div>
    );
  }
  if (!job) {
    return (
      <div className="mx-auto max-w-5xl space-y-4">
        {back}
        <Skeleton className="h-16 w-full rounded-2xl" />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }, (_, i) => (
            <Skeleton key={i} className="h-64 rounded-2xl" />
          ))}
        </div>
      </div>
    );
  }

  const visible = job.items.filter((i) => !i.removed);
  const ready = visible.filter((i) => i.status === "ready");
  const settled = visible.filter((i) => ["ready", "failed", "invalid"].includes(i.status)).length;
  const working = settled < visible.length;
  const byId = (id: string | null): TJobItem | undefined => job.items.find((i) => i.id === id);

  return (
    <div className="mx-auto max-w-5xl">
      {back}
      <header className="mb-6 space-y-3">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <h1 className="text-2xl font-semibold tracking-tight">{t("vocab.reviewTitle")}</h1>
          <p className="text-sm text-muted-foreground" aria-live="polite">
            {working
              ? t("vocab.progress", { done: settled, total: visible.length })
              : t("vocab.progressDone", { ready: ready.length, total: visible.length })}
          </p>
        </div>
        {working && <Progress value={visible.length ? (settled / visible.length) * 100 : 0} aria-label={t("vocab.progressLabel")} />}
        <p className="text-sm text-muted-foreground">{t("vocab.reviewHint")}</p>
      </header>

      <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {job.items.map((item) => (
          <li key={item.id}>
            <ItemCard
              jobId={job.id}
              item={item}
              onOpen={() => setOpenItemId(item.id)}
              onPickImage={() => setPickingItemId(item.id)}
            />
          </li>
        ))}
      </ul>

      <SaveBar job={job} readyItems={ready} />

      <ItemDialog jobId={job.id} item={byId(openItemId)} onOpenChange={(open) => !open && setOpenItemId(null)} onPickImage={setPickingItemId} />
      <ImagePickerDialog jobId={job.id} item={byId(pickingItemId)} onOpenChange={(open) => !open && setPickingItemId(null)} />
    </div>
  );
}
