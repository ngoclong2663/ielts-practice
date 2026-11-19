import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { useCreateJob } from "@/features/vocab/hooks/use-vocab";
import { useRangeSelect } from "@/hooks/use-range-select";
import { useRouter } from "@/i18n/navigation";
import { ApiError } from "@/lib/api-client";
import type { TVocabItem } from "../types";

const SENSE_HINT_MAX = 60;

/**
 * Picks words from a task's vocabulary. Seeded vocabulary (Epic 8) has a ready-made card per term, added to a
 * deck at once; vocabulary the AI wrote goes to the card generator (Epic 3), reviewed on the Add words page.
 */
export function useVocabSelection(items: TVocabItem[]) {
  const t = useTranslations();
  const router = useRouter();
  const createJob = useCreateJob();
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const toggle = (term: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(term)) next.delete(term);
      else next.add(term);
      return next;
    });

  // Shift-click ranges: every term between the last tick and this one takes its new state.
  const range = useRangeSelect(items.map((i) => i.term));
  const tick = (term: string) => {
    const on = !selected.has(term);
    const keys = range.affected(term);
    setSelected((prev) => {
      const next = new Set(prev);
      for (const k of keys) {
        if (on) next.add(k);
        else next.delete(k);
      }
      return next;
    });
  };

  const allSelected = items.length > 0 && items.every((i) => selected.has(i.term));
  const toggleAll = () => setSelected(allSelected ? new Set() : new Set(items.map((i) => i.term)));

  const save = () => {
    // The Vietnamese meaning is the sense hint, so the card explains the sense used in this lesson.
    const entries = items
      .filter((i) => selected.has(i.term))
      .map((i) => ({ text: i.term, sense_hint: i.vi.slice(0, SENSE_HINT_MAX) }));
    if (entries.length === 0) return;
    createJob.mutate(
      { entries },
      {
        onSuccess: (job) => router.push(`/add-words?job=${job.id}`),
        onError: (error) => {
          const code = error instanceof ApiError ? error.code : "UNKNOWN";
          toast.error(t(`errors.${code}`, { ...(error.params as Record<string, number>) }), { id: "writing-vocab-save" });
        },
      },
    );
  };

  const clear = () => setSelected(new Set());
  // Seeded vocabulary: every term already has a card, so it goes straight to a deck (no generation job).
  const ready = items.length > 0 && items.every((i) => !!i.lexeme_id);
  const selectedLexemes = items.filter((i) => selected.has(i.term) && i.lexeme_id).map((i) => i.lexeme_id as string);

  return { selected, toggle, tick, rowProps: range.rowProps, allSelected, toggleAll, clear, save, saving: createJob.isPending, ready, selectedLexemes };
}
