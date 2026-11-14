import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";
import { answerApi, fetchQueue, QUEUE_SIZE, undoApi } from "../api/study-api";
import { studyKeys } from "../query-keys";
import type { TRating, TStudyQueue } from "../types";

type TSessionSummary = { answered: number; again: number; durationMs: number };

const EMPTY_SUMMARY: TSessionSummary = { answered: 0, again: 0, durationMs: 0 };

function isTypingTarget(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null;
  return !!el && (el.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(el.tagName));
}

/**
 * Drives one study session: the server owns scheduling and returns the next cards after
 * every answer, so this hook only tracks what is on screen and the session summary.
 */
export function useStudySession(deckId: string) {
  const t = useTranslations();
  const qc = useQueryClient();
  const initial = useQuery({
    queryKey: studyKeys.queue(deckId),
    queryFn: () => fetchQueue(deckId || undefined),
    staleTime: 0,
    gcTime: 0,
  });

  const [override, setOverride] = useState<TStudyQueue | null>(null);
  const queue = override ?? initial.data ?? null;
  const card = queue?.cards[0] ?? null;

  const [revealed, setRevealed] = useState(false);
  const [summary, setSummary] = useState<TSessionSummary>(EMPTY_SUMMARY);
  const shownAt = useRef(0);

  // Start timing whenever a different card is shown.
  useEffect(() => {
    shownAt.current = Date.now();
  }, [card?.id, card?.reps]);

  // Preload the next cards' images so flipping never waits on the network.
  useEffect(() => {
    queue?.cards.slice(1).forEach((c) => {
      if (c.content.image?.url) {
        const img = new Image();
        img.src = c.content.image.url;
      }
    });
  }, [queue]);

  // Due counts on other pages are stale once the session ends.
  useEffect(
    () => () => {
      qc.invalidateQueries({ queryKey: studyKeys.counts() });
      qc.invalidateQueries({ queryKey: studyKeys.stats() });
    },
    [qc],
  );

  const showError = useCallback(
    (error: unknown) => {
      const code = error instanceof ApiError ? error.code : "UNKNOWN";
      toast.error(t.has(`errors.${code}`) ? t(`errors.${code}`) : t("errors.UNKNOWN"));
    },
    [t],
  );

  const reload = useCallback(async () => {
    const fresh = await fetchQueue(deckId || undefined);
    setOverride(fresh);
    setRevealed(false);
  }, [deckId]);

  const answerMutation = useMutation({
    mutationFn: answerApi,
    onSuccess: (res, vars) => {
      setOverride(res);
      setRevealed(false);
      setSummary((s) => ({
        answered: s.answered + 1,
        again: s.again + (vars.rating === 1 ? 1 : 0),
        durationMs: s.durationMs + vars.duration_ms,
      }));
      if (res.answered?.became_leech) {
        toast.warning(t("study.leechHint"));
      }
    },
    onError: (error) => {
      showError(error);
      if (error instanceof ApiError && error.code === "STUDY_CARD_STATE_CHANGED") {
        void reload();
      }
    },
  });

  const undoMutation = useMutation({
    mutationFn: () => undoApi(deckId || undefined),
    onSuccess: (res) => {
      setOverride(res);
      setRevealed(false);
      setSummary((s) => ({ ...s, answered: Math.max(0, s.answered - 1) }));
      toast(t("study.undone"));
    },
    onError: showError,
  });

  const busy = answerMutation.isPending || undoMutation.isPending;

  const reveal = useCallback(() => {
    if (card && !busy) {
      setRevealed(true);
    }
  }, [card, busy]);

  const answer = useCallback(
    (rating: TRating) => {
      if (!card || !revealed || busy) {
        return;
      }
      answerMutation.mutate({
        card_id: card.id,
        rating,
        duration_ms: Math.min(Date.now() - shownAt.current, 60_000),
        client_review_id: crypto.randomUUID(),
        expected_reps: card.reps,
        deck_id: deckId || undefined,
        size: QUEUE_SIZE,
      });
    },
    [answerMutation, busy, card, deckId, revealed],
  );

  const undo = useCallback(() => {
    if (!busy) {
      undoMutation.mutate();
    }
  }, [busy, undoMutation]);

  // Keyboard: Space/Enter reveal, 1-4 answer, Ctrl/Cmd+Z undo.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (isTypingTarget(e.target) || document.querySelector("[role=dialog]")) {
        return;
      }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "z") {
        e.preventDefault();
        undo();
        return;
      }
      if (!revealed && (e.key === " " || e.key === "Enter")) {
        e.preventDefault();
        reveal();
        return;
      }
      if (revealed && ["1", "2", "3", "4"].includes(e.key)) {
        e.preventDefault();
        answer(Number(e.key) as TRating);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [answer, reveal, revealed, undo]);

  return {
    queue,
    card,
    revealed,
    reveal,
    answer,
    undo,
    reload,
    busy,
    canUndo: summary.answered > 0,
    summary,
    isLoading: initial.isPending && !override,
    isError: initial.isError && !override,
    refetch: initial.refetch,
  };
}
