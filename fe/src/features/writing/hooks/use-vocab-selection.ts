import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { useCreateJob } from "@/features/vocab/hooks/use-vocab";
import { useRouter } from "@/i18n/navigation";
import { ApiError } from "@/lib/api-client";
import type { TVocabItem } from "../types";

const SENSE_HINT_MAX = 60;

/**
 * Picks words from a lesson's vocabulary and sends them to the card generator (Epic 3). The learner reviews
 * the cards and chooses the deck on the Add words page, as for a pasted list.
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

  return { selected, toggle, allSelected, toggleAll, save, saving: createJob.isPending };
}
