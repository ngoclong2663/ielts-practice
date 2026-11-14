"use client";

import { CheckCircle2, GraduationCap } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useDecks } from "@/features/decks/hooks/use-decks";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { ApiError } from "@/lib/api-client";
import { useCommitJob } from "../hooks/use-vocab";
import { readLastDeck, rememberDeck } from "../lib/last-deck";
import type { TJob, TJobItem } from "../types";

const NEW_DECK = "__new__";

/** Sticky footer: choose a deck (default: last used, else a new dated deck) and save the ready cards. */
export function SaveBar({ job, readyItems }: { job: TJob; readyItems: TJobItem[] }) {
  const t = useTranslations();
  const format = useFormatter();
  const isClient = useIsClient();
  const decksQuery = useDecks();
  const decks = isClient ? decksQuery.data : undefined;
  const commit = useCommitJob(job.id);
  const [picked, setPicked] = useState<string | null>(null);
  const [newName, setNewName] = useState<string | null>(null);
  const [savedDeck, setSavedDeck] = useState<{ id: string; name: string } | null>(null);

  const lastDeck = isClient ? readLastDeck() : null;
  const fallback =
    [job.target_deck_id, lastDeck].find((id) => id && decks?.some((d) => d.id === id)) ?? NEW_DECK;
  const deckId = picked ?? fallback;
  const defaultName = t("vocab.defaultDeckName", { date: format.dateTime(new Date(job.created_at), { dateStyle: "short" }) });
  const name = newName ?? defaultName;
  const items = [{ value: NEW_DECK, label: t("vocab.newDeckOption") }, ...(decks ?? []).map((d) => ({ value: d.id, label: d.name }))];
  const committedId = savedDeck?.id ?? job.committed_deck_id;

  const save = () => {
    commit.mutate(
      {
        item_ids: readyItems.map((i) => i.id),
        ...(deckId === NEW_DECK ? { new_deck_name: name.trim() } : { deck_id: deckId }),
      },
      {
        onSuccess: (res) => {
          rememberDeck(res.deck_id);
          setSavedDeck({ id: res.deck_id, name: res.deck_name });
          toast.success(t("vocab.saved", { count: res.created, deck: res.deck_name }));
          if (res.skipped.length > 0) {
            toast.info(t("vocab.skipped", { count: res.skipped.length }), { description: res.skipped.join(", ") });
          }
        },
        onError: (error) => {
          const code = error instanceof ApiError ? error.code : "UNKNOWN";
          toast.error(t(`errors.${code}`));
        },
      },
    );
  };

  return (
    // Sticks above the mobile bottom navigation, and to the bottom of the content column on desktop.
    <div className="sticky bottom-[calc(4.25rem+env(safe-area-inset-bottom))] z-20 mt-8 lg:bottom-6">
      <div className="flex flex-col gap-3 rounded-2xl border border-border bg-background/95 px-4 py-3 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/85 sm:flex-row sm:items-center sm:justify-between">
        {committedId && !commit.isPending ? (
          <>
            <p className="flex items-center gap-2 text-sm">
              <CheckCircle2 className="size-4 text-primary" strokeWidth={1.75} aria-hidden />
              {savedDeck ? t("vocab.savedTo", { deck: savedDeck.name }) : t("vocab.savedEarlier")}
            </p>
            <div className="flex gap-2">
              <Button variant="outline" render={<Link href={`/decks/${committedId}`} />}>
                {t("vocab.openDeck")}
              </Button>
              <Button render={<Link href={`/study?deck=${committedId}`} />}>
                <GraduationCap className="size-4" strokeWidth={1.75} aria-hidden />
                {t("vocab.studyNow")}
              </Button>
            </div>
          </>
        ) : (
          <>
            <div className="flex flex-1 flex-col gap-2 sm:flex-row sm:items-center">
              <span className="text-sm font-medium">{t("vocab.saveTo")}</span>
              <Select items={items} value={deckId} onValueChange={(v) => v && setPicked(String(v))}>
                <SelectTrigger className="w-full sm:w-56" aria-label={t("vocab.saveTo")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {items.map((d) => (
                    <SelectItem key={d.value} value={d.value}>
                      {d.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {deckId === NEW_DECK && (
                <Input
                  value={name}
                  onChange={(e) => setNewName(e.target.value)}
                  aria-label={t("decks.name")}
                  maxLength={80}
                  className="sm:w-56"
                />
              )}
            </div>
            <Button
              size="lg"
              onClick={save}
              disabled={readyItems.length === 0 || commit.isPending || (deckId === NEW_DECK && !name.trim())}
            >
              {commit.isPending ? t("common.saving") : t("vocab.saveCards", { count: readyItems.length })}
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
