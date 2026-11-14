"use client";

import { AlertCircle, ImageIcon, MoreHorizontal, RefreshCw, Trash2, Undo2, Wand2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useRetryItem, useUpdateItem } from "../hooks/use-vocab";
import type { TJobItem } from "../types";

type ItemCardProps = {
  jobId: string;
  item: TJobItem;
  onOpen: () => void;
  onPickImage: () => void;
};

const STEP_KEYS = { queued: "stepQueued", content: "stepContent", image: "stepImage" } as const;

/** Mini card preview in the review grid (Epic 3 §3 step 5). */
export function ItemCard({ jobId, item, onOpen, onPickImage }: ItemCardProps) {
  const t = useTranslations();
  const retry = useRetryItem(jobId);
  const update = useUpdateItem(jobId);
  const [failedSrc, setFailedSrc] = useState<string | null>(null);

  const onError = (error: ApiError) => toast.error(t(`errors.${error.code}`, { ...(error.params as Record<string, number>) }));
  const doRetry = (p: { input_text?: string; regenerate?: boolean } = {}) =>
    retry.mutate({ itemId: item.id, ...p }, { onError });
  const setRemoved = (removed: boolean) => update.mutate({ itemId: item.id, removed }, { onError });

  if (item.removed) {
    return (
      <article className="flex h-full min-h-24 items-center justify-between gap-3 rounded-2xl border border-dashed border-border px-4 py-3 text-muted-foreground">
        <span className="truncate line-through">{item.input_text}</span>
        <Button variant="ghost" size="sm" onClick={() => setRemoved(false)} disabled={update.isPending}>
          <Undo2 className="size-4" strokeWidth={1.75} aria-hidden />
          {t("vocab.undoRemove")}
        </Button>
      </article>
    );
  }

  if (item.status === "queued" || item.status === "content" || item.status === "image") {
    return (
      <article aria-busy className="flex h-full flex-col overflow-hidden rounded-2xl border border-border bg-card">
        <div className="aspect-[16/9] animate-pulse bg-muted" />
        <div className="space-y-2 p-4">
          <p className="text-lg font-semibold">{item.input_text}</p>
          <p className="flex items-center gap-2 text-sm text-muted-foreground">
            <span className="size-1.5 animate-pulse rounded-full bg-primary" aria-hidden />
            {t(`vocab.${STEP_KEYS[item.status]}`)}
          </p>
        </div>
      </article>
    );
  }

  if (item.status === "invalid" || item.status === "failed") {
    const invalid = item.status === "invalid";
    return (
      <article className="flex h-full flex-col gap-3 rounded-2xl border border-destructive/30 bg-card p-4">
        <div className="flex items-start gap-2">
          <AlertCircle className="mt-0.5 size-4 shrink-0 text-destructive" strokeWidth={1.75} aria-hidden />
          <div className="min-w-0">
            <p className="font-semibold break-words">{item.input_text}</p>
            <p className="text-sm text-muted-foreground">
              {invalid ? t("vocab.invalidWord") : t(`errors.${item.error_code || "VOCAB_GENERATION_FAILED"}`)}
            </p>
          </div>
        </div>
        <div className="mt-auto flex flex-wrap gap-2">
          {invalid && item.suggestion && (
            <Button size="sm" onClick={() => doRetry({ input_text: item.suggestion })} disabled={retry.isPending}>
              <Wand2 className="size-4" strokeWidth={1.75} aria-hidden />
              {t("vocab.useSuggestion", { word: item.suggestion })}
            </Button>
          )}
          {!invalid && (
            <Button size="sm" variant="outline" onClick={() => doRetry()} disabled={retry.isPending}>
              <RefreshCw className="size-4" strokeWidth={1.75} aria-hidden />
              {t("common.retry")}
            </Button>
          )}
          <Button size="sm" variant="ghost" onClick={() => setRemoved(true)} disabled={update.isPending}>
            <Trash2 className="size-4" strokeWidth={1.75} aria-hidden />
            {t("vocab.remove")}
          </Button>
        </div>
      </article>
    );
  }

  const c = item.content;
  const image = c?.image;
  const src = image ? image.thumb_url || image.url : "";
  const ipa = c?.ipa?.uk || c?.ipa?.us;
  return (
    <article className="group flex h-full flex-col overflow-hidden rounded-2xl border border-border bg-card transition-shadow hover:shadow-sm">
      <button
        type="button"
        onClick={onPickImage}
        className="relative block aspect-[16/9] overflow-hidden bg-muted"
        aria-label={t("vocab.changeImageFor", { word: c?.headword ?? item.input_text })}
      >
        {image && src !== failedSrc ? (
          // eslint-disable-next-line @next/next/no-img-element -- provider CDNs and our redirect endpoint
          <img
            src={src}
            alt={image.alt || c?.headword || ""}
            className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
            loading="lazy"
            onError={() => setFailedSrc(src)}
          />
        ) : (
          <span className="flex h-full items-center justify-center text-sm text-muted-foreground">
            <ImageIcon className="mr-2 size-4" strokeWidth={1.75} aria-hidden />
            {t("vocab.addImage")}
          </span>
        )}
        <span className="absolute right-2 bottom-2 rounded-md bg-background/85 px-2 py-0.5 text-xs font-medium opacity-0 backdrop-blur transition-opacity group-hover:opacity-100 focus-visible:opacity-100">
          {t("vocab.changeImage")}
        </span>
      </button>

      <div className="flex flex-1 flex-col gap-2 p-4">
        <div className="flex items-start justify-between gap-2">
          <button type="button" onClick={onOpen} className="min-w-0 text-left">
            <p className="text-lg leading-tight font-semibold break-words hover:underline hover:underline-offset-4">
              {c?.headword ?? item.input_text}
            </p>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {ipa && <span className="font-mono">{ipa}</span>}
              {c?.part_of_speech && <span className={cn("italic", ipa && "ml-2")}>{c.part_of_speech}</span>}
            </p>
          </button>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="ghost" size="icon-sm" aria-label={t("vocab.itemMenu")} />}>
              <MoreHorizontal className="size-4" strokeWidth={1.75} aria-hidden />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem onClick={onOpen}>{t("vocab.previewEdit")}</DropdownMenuItem>
              <DropdownMenuItem onClick={onPickImage}>{t("vocab.changeImage")}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => doRetry({ regenerate: true })}>{t("vocab.regenerate")}</DropdownMenuItem>
              <DropdownMenuItem variant="destructive" onClick={() => setRemoved(true)}>
                {t("vocab.remove")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <p className="font-medium text-primary">{c?.meaning_vi}</p>
        {c?.word_forms && c.word_forms.length > 1 && (
          <p className="truncate text-sm text-muted-foreground">{c.word_forms.map((w) => w.form).join(" · ")}</p>
        )}
        {item.input_text !== c?.headword && c?.headword && (
          <p className="text-xs text-muted-foreground">{t("vocab.lemmaNote", { input: item.input_text, word: c.headword })}</p>
        )}
        <div className="mt-auto pt-2">
          <Button variant="outline" size="sm" className="w-full" onClick={onOpen}>
            {t("vocab.previewEdit")}
          </Button>
        </div>
      </div>
    </article>
  );
}
