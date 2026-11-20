import { useCallback, useEffect, useMemo, useState } from "react";
import { type TRunConfig, type TRunState, clearRun, loadRun, newRun, runKey, saveRun } from "../lib/run-state";
import type { THighlight } from "../types";

/** Re-renders every second while `active`; returns the current time in ms. */
export function useNow(active: boolean): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!active) return;
    const id = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(id);
  }, [active]);
  return now;
}

/**
 * The state of one run (Epic 10 §5.4), restored from and saved to localStorage on every change. Only the
 * submit call sends anything to the API.
 */
export function useRun(config: TRunConfig, firstQuestion: number) {
  const key = runKey(config);
  const [state, setState] = useState<TRunState>(() => loadRun(key) ?? newRun(config, firstQuestion));

  useEffect(() => {
    saveRun(key, state);
  }, [key, state]);

  const setAnswer = useCallback((n: number, value: string | string[]) => {
    setState((s) => ({ ...s, answers: { ...s.answers, [String(n)]: value }, current: n }));
  }, []);

  /** "Choose TWO": toggles a letter; choosing one more than allowed drops the oldest. */
  const toggleLetter = useCallback((groupFrom: number, letter: string, pick: number) => {
    setState((s) => {
      const cur = s.answers[String(groupFrom)];
      let list = Array.isArray(cur) ? cur.filter((l) => l !== letter) : [];
      if (!(Array.isArray(cur) && cur.includes(letter))) {
        list = [...list, letter];
        if (list.length > pick) list = list.slice(list.length - pick);
      }
      return { ...s, answers: { ...s.answers, [String(groupFrom)]: list }, current: groupFrom };
    });
  }, []);

  const toggleFlag = useCallback((n: number) => {
    setState((s) => ({ ...s, flags: s.flags.includes(n) ? s.flags.filter((f) => f !== n) : [...s.flags, n] }));
  }, []);

  const setCurrent = useCallback((n: number) => setState((s) => (s.current === n ? s : { ...s, current: n })), []);
  const setHighlights = useCallback((highlights: THighlight[]) => setState((s) => ({ ...s, highlights })), []);
  const discard = useCallback(() => clearRun(key), [key]);
  const setAudio = useCallback((audio: NonNullable<TRunState["audio"]>) => setState((s) => ({ ...s, audio })), []);

  const now = useNow(true);
  const timing = useMemo(() => {
    const elapsedSec = Math.max(0, Math.round((now - state.startedAt) / 1000));
    const remainingSec = state.deadline === null ? null : Math.round((state.deadline - now) / 1000);
    return { elapsedSec, remainingSec, timeUp: remainingSec !== null && remainingSec <= 0 };
  }, [now, state.startedAt, state.deadline]);

  return { key, state, setAnswer, toggleLetter, toggleFlag, setCurrent, setHighlights, setAudio, discard, ...timing };
}
