"use client";

import { Check, Eye, Users, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { PageHeader } from "@/components/shared/PageHeader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { CardBack } from "@/features/study/components/CardBack";
import { CardFront } from "@/features/study/components/CardFront";
import { useCardAudio } from "@/features/study/hooks/use-card-audio";
import { useIsClient } from "@/hooks/use-is-client";
import { useRangeSelect } from "@/hooks/use-range-select";
import { cn } from "@/lib/utils";
import { useAdminCandidates, useDecideCandidates } from "../hooks/use-admin";
import type { TAdminCandidate } from "../types";
import { QueryError } from "./QueryError";

const NO_TOPIC = "";

/**
 * Library candidates (Epic 9): AI-made cards several learners keep. The owner picks a topic and approves a word
 * into the library (its card is then locked), or rejects it (it is not offered again).
 */
export function CandidatesView() {
  const t = useTranslations("admin.library");
  const tv = useTranslations("vocab.topics");
  const isClient = useIsClient();
  const query = useAdminCandidates();
  const data = isClient ? query.data : undefined;
  const { approve, reject, pending } = useDecideCandidates();
  const [topics, setTopics] = useState<Record<string, string>>({});
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [bulkTopic, setBulkTopic] = useState(NO_TOPIC);
  const [preview, setPreview] = useState<TAdminCandidate | null>(null);
  const items = data?.items ?? [];
  const range = useRangeSelect(items.map((c) => c.lexeme_id));

  if (query.error) {
    return (
      <div className="mx-auto max-w-6xl">
        <QueryError onRetry={() => query.refetch()} />
      </div>
    );
  }
  if (!data) {
    return (
      <div className="mx-auto max-w-6xl space-y-3">
        <Skeleton className="h-8 w-1/3" />
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-20 w-full rounded-xl" />
        ))}
      </div>
    );
  }

  const topicItems = data.topics.map((id) => ({ value: id, label: tv(id) }));
  const bulkItems = [{ value: NO_TOPIC, label: t("keepTopics") }, ...topicItems];
  const suggested = new Map(items.map((c) => [c.lexeme_id, c.suggested_topic]));
  // The owner's choice, else the topic the card generator tagged.
  const topicOf = (id: string) => topics[id] ?? suggested.get(id) ?? NO_TOPIC;
  const tick = (id: string) => {
    const on = !selected.has(id);
    const keys = range.affected(id);
    setSelected((prev) => {
      const next = new Set(prev);
      for (const k of keys) {
        if (on) next.add(k);
        else next.delete(k);
      }
      return next;
    });
  };
  const done = (ids: string[]) => {
    setSelected((prev) => new Set([...prev].filter((id) => !ids.includes(id))));
  };

  const approveIds = (ids: string[], topicFor: (id: string) => string) => {
    const missing = ids.filter((id) => !topicFor(id));
    if (missing.length > 0) {
      toast.error(t("topicNeeded", { count: missing.length }));
      return;
    }
    approve.mutate(
      { items: ids.map((id) => ({ lexeme_id: id, topic_id: topicFor(id) })) },
      {
        onSuccess: (r) => {
          done(ids);
          toast.success(t("approved", { count: r.count }));
        },
        onError: (e) => toast.error(t("failed", { code: e.code })),
      },
    );
  };
  const rejectIds = (ids: string[]) =>
    reject.mutate(ids, {
      onSuccess: (r) => {
        done(ids);
        toast.success(t("rejected", { count: r.count }));
      },
      onError: (e) => toast.error(t("failed", { code: e.code })),
    });

  const picked = [...selected].filter((id) => items.some((c) => c.lexeme_id === id));
  const allSelected = items.length > 0 && picked.length === items.length;

  return (
    // Same column as the admin header (logo, tabs, sign out) and the other admin pages.
    <section className="mx-auto max-w-6xl space-y-4">
      <PageHeader title={t("title")} description={t("intro", { min: data.min_users })} />

      {items.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
          {t("empty", { min: data.min_users })}
        </p>
      ) : (
        <>
          <div className="flex flex-wrap items-center gap-3 rounded-2xl border border-border bg-card px-4 py-3">
            <label className="flex cursor-pointer items-center gap-2 text-sm">
              <Checkbox
                checked={allSelected}
                onCheckedChange={() => setSelected(allSelected ? new Set() : new Set(items.map((c) => c.lexeme_id)))}
              />
              {t("selectAll", { count: items.length })}
            </label>
            {picked.length > 0 && (
              <div className="ml-auto flex flex-wrap items-center gap-2">
                <span className="text-sm text-muted-foreground">{t("selected", { count: picked.length })}</span>
                <TopicSelect items={bulkItems} value={bulkTopic} onChange={setBulkTopic} label={t("bulkTopic")} />
                <Button size="sm" disabled={pending} onClick={() => approveIds(picked, (id) => bulkTopic || topicOf(id))}>
                  <Check className="size-4" strokeWidth={1.75} aria-hidden />
                  {t("approveCount", { count: picked.length })}
                </Button>
                <Button size="sm" variant="outline" disabled={pending} onClick={() => rejectIds(picked)}>
                  <X className="size-4" strokeWidth={1.75} aria-hidden />
                  {t("rejectCount", { count: picked.length })}
                </Button>
              </div>
            )}
          </div>

          <ul className="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-card">
            {items.map((c) => {
              const id = `cand-${c.lexeme_id}`;
              const ipa = c.card.ipa?.uk || c.card.ipa?.us;
              return (
                <li key={c.lexeme_id} {...range.rowProps} className={cn("flex flex-wrap items-center gap-3 px-4 py-3", selected.has(c.lexeme_id) && "bg-primary/5")}>
                  <Checkbox id={id} checked={selected.has(c.lexeme_id)} onCheckedChange={() => tick(c.lexeme_id)} aria-label={c.card.headword} />
                  <label htmlFor={id} className="min-w-0 flex-1 cursor-pointer basis-48">
                    <span className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                      <span className="font-medium">{c.card.headword}</span>
                      {ipa && <span className="font-mono text-xs text-muted-foreground">{ipa}</span>}
                      {c.card.part_of_speech && <span className="text-xs italic text-muted-foreground">{c.card.part_of_speech}</span>}
                      {c.card.cefr && <Badge variant="secondary">{c.card.cefr}</Badge>}
                    </span>
                    <span className="block truncate text-sm text-primary">{c.card.meaning_vi}</span>
                    <span className="mt-0.5 flex flex-wrap gap-3 text-xs text-muted-foreground">
                      <span className="inline-flex items-center gap-1">
                        <Users className="size-3.5" strokeWidth={1.75} aria-hidden />
                        {t("users", { count: c.users })}
                      </span>
                      {c.report_count > 0 && <span className="text-destructive">{t("reports", { count: c.report_count })}</span>}
                    </span>
                  </label>
                  <div className="flex w-full flex-wrap items-center gap-2 sm:w-auto">
                    <Button variant="ghost" size="icon-sm" onClick={() => setPreview(c)} aria-label={t("preview", { word: c.card.headword })}>
                      <Eye className="size-4" strokeWidth={1.75} aria-hidden />
                    </Button>
                    <TopicSelect
                      items={topicItems}
                      value={topicOf(c.lexeme_id)}
                      onChange={(v) => setTopics((prev) => ({ ...prev, [c.lexeme_id]: v }))}
                      label={t("topicFor", { word: c.card.headword })}
                    />
                    <Button size="sm" disabled={pending || !topicOf(c.lexeme_id)} onClick={() => approveIds([c.lexeme_id], topicOf)}>
                      {t("approve")}
                    </Button>
                    <Button size="sm" variant="ghost" disabled={pending} onClick={() => rejectIds([c.lexeme_id])}>
                      {t("reject")}
                    </Button>
                  </div>
                </li>
              );
            })}
          </ul>
        </>
      )}
      <CandidatePreview candidate={preview} onClose={() => setPreview(null)} />
    </section>
  );
}

function TopicSelect({ items, value, onChange, label }: { items: { value: string; label: string }[]; value: string; onChange: (v: string) => void; label: string }) {
  return (
    <Select items={items} value={value} onValueChange={(v) => onChange(String(v ?? ""))}>
      <SelectTrigger size="sm" className="w-36 sm:w-44" aria-label={label}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {items.map((it) => (
          <SelectItem key={it.value || "keep"} value={it.value}>
            {it.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

function CandidatePreview({ candidate, onClose }: { candidate: TAdminCandidate | null; onClose: () => void }) {
  const t = useTranslations("admin.library");
  const { play, playHeadword } = useCardAudio("uk");
  const card = candidate?.card;
  return (
    <Dialog open={!!card} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className="sr-only">{card?.headword}</DialogTitle>
          <DialogDescription className="sr-only">{t("previewHint")}</DialogDescription>
        </DialogHeader>
        {card && (
          <div className="space-y-6">
            <CardFront content={card} accent="uk" showImage={false} onPlay={() => playHeadword(card)} />
            <div className="border-t border-border pt-6">
              <CardBack content={card} onPlay={play} />
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
