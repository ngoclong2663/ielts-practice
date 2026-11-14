"use client";

import { Check, ImageOff, Search } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import type { TCardImage } from "@/types/card-content";
import { useSearchImages, useUpdateItem, useVocabStatus } from "../hooks/use-vocab";
import type { TJobItem } from "../types";

type ImagePickerDialogProps = {
  jobId: string;
  item: TJobItem | undefined;
  onOpenChange: (open: boolean) => void;
};

const sameImage = (a?: TCardImage, b?: TCardImage) => !!a && !!b && a.provider === b.provider && a.provider_id === b.provider_id;

/** Six suggestions, a custom keyword search, and "No image" (US-3.3). */
export function ImagePickerDialog({ jobId, item, onOpenChange }: ImagePickerDialogProps) {
  const t = useTranslations();
  const status = useVocabStatus().data;
  const search = useSearchImages(jobId);
  const update = useUpdateItem(jobId);
  const [query, setQuery] = useState("");
  const [provider, setProvider] = useState<string | undefined>(undefined);
  const [results, setResults] = useState<TCardImage[] | null>(null);
  const [lastItemId, setLastItemId] = useState<string | undefined>(undefined);

  // Reset when another word is opened.
  if (item?.id !== lastItemId) {
    setLastItemId(item?.id);
    setQuery("");
    setResults(null);
    setProvider(undefined);
  }

  const onError = (error: ApiError) => toast.error(t(`errors.${error.code}`));
  const current = item?.content?.image;
  const images = results ?? item?.image_candidates ?? [];
  const providers = status?.image_providers ?? [];

  const runSearch = (nextProvider = provider) => {
    if (!item) return;
    search.mutate(
      { itemId: item.id, q: query.trim(), provider: nextProvider },
      { onSuccess: (r) => setResults(r.images), onError },
    );
  };

  const choose = (img: TCardImage) => {
    if (!item || !img.provider_id) return;
    update.mutate(
      { itemId: item.id, image: { provider: img.provider, provider_id: img.provider_id } },
      { onSuccess: () => onOpenChange(false), onError },
    );
  };

  const clear = () => {
    if (!item) return;
    update.mutate({ itemId: item.id, clear_image: true }, { onSuccess: () => onOpenChange(false), onError });
  };

  return (
    <Dialog open={!!item} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("vocab.pickImageTitle", { word: item?.content?.headword ?? item?.input_text ?? "" })}</DialogTitle>
          <DialogDescription>{t("vocab.pickImageBody")}</DialogDescription>
        </DialogHeader>

        <form
          className="flex flex-col gap-2 sm:flex-row"
          onSubmit={(e) => {
            e.preventDefault();
            runSearch();
          }}
        >
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("vocab.imageSearchPlaceholder")}
            aria-label={t("vocab.imageSearchPlaceholder")}
            maxLength={100}
          />
          <Button type="submit" variant="secondary" disabled={search.isPending}>
            <Search className="size-4" strokeWidth={1.75} aria-hidden />
            {t("vocab.searchImages")}
          </Button>
        </form>
        {providers.length > 1 && (
          <div className="flex flex-wrap items-center gap-2 text-sm">
            <span className="text-muted-foreground">{t("vocab.imageSource")}</span>
            {[undefined, ...providers].map((p) => (
              <button
                key={p ?? "all"}
                type="button"
                aria-pressed={provider === p}
                onClick={() => {
                  setProvider(p);
                  runSearch(p);
                }}
                className={cn(
                  "rounded-full border px-3 py-0.5 capitalize transition-colors",
                  provider === p ? "border-primary bg-primary/10 text-primary" : "border-border hover:bg-muted",
                )}
              >
                {p ?? t("vocab.imageSourceAuto")}
              </button>
            ))}
          </div>
        )}

        {search.isPending ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {Array.from({ length: 6 }, (_, i) => (
              <div key={i} className="aspect-[4/3] animate-pulse rounded-xl bg-muted" />
            ))}
          </div>
        ) : images.length === 0 ? (
          <p className="rounded-xl border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
            {t("vocab.noImagesFound")}
          </p>
        ) : (
          <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {images.map((img) => {
              const selected = sameImage(img, current);
              return (
                <li key={`${img.provider}-${img.provider_id}`}>
                  <button
                    type="button"
                    onClick={() => choose(img)}
                    disabled={update.isPending}
                    aria-pressed={selected}
                    aria-label={img.alt || t("vocab.imageOption")}
                    className={cn(
                      "relative block aspect-[4/3] w-full overflow-hidden rounded-xl bg-muted ring-offset-2 ring-offset-background transition",
                      selected ? "ring-2 ring-primary" : "hover:opacity-90",
                    )}
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element -- provider CDNs */}
                    <img src={img.thumb_url || img.url} alt="" className="h-full w-full object-cover" loading="lazy" />
                    {selected && (
                      <span className="absolute top-2 right-2 rounded-full bg-primary p-1 text-primary-foreground">
                        <Check className="size-3.5" strokeWidth={2.5} aria-hidden />
                      </span>
                    )}
                  </button>
                  <p className="mt-1 truncate text-[11px] text-muted-foreground">
                    {img.author_name}
                    <span className="capitalize"> · {img.provider}</span>
                  </p>
                </li>
              );
            })}
          </ul>
        )}

        <div className="flex justify-between gap-2 border-t border-border pt-4">
          <Button variant="ghost" onClick={clear} disabled={update.isPending || !current}>
            <ImageOff className="size-4" strokeWidth={1.75} aria-hidden />
            {t("vocab.noImage")}
          </Button>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
