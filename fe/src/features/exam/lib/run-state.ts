import type { TAnswers, TAttemptKind, THighlight, TListenMode, TSkill } from "../types";

// A run in progress lives only in this browser (Epic 10 §5.4): answers, flags, highlights, the current
// question and the timer's deadline. Nothing reaches the API until the learner submits. Storage may be
// blocked (private window); every access is wrapped so the run still works in memory.

export type TRunConfig = {
  skill: TSkill;
  test: string;
  kind: TAttemptKind;
  parts: number[];
  /** 0 = no limit (the clock counts up). */
  timeLimitSec: number;
  mode: TListenMode;
  redo?: { attemptId: string; questions: number[] };
};

export type TRunState = {
  answers: TAnswers;
  flags: number[];
  highlights: THighlight[];
  current: number;
  startedAt: number;
  /** Epoch ms when the time limit runs out; null when untimed. A reload keeps the same deadline. */
  deadline: number | null;
  savedAt: number;
  /** Listening: where the audio was (part number and seconds), the reading pauses already taken, and
   * whether an exam-mode run has started (it cannot be restarted). */
  audio?: { part: number; t: number; cuesDone: string[]; started: boolean };
  /** Listening: practice or exam mode, so "Continue" reopens the run as it was. */
  mode?: TListenMode;
};

const RUN_PREFIX = "exam.run.";
const HL_PREFIX = "exam.hl.";
const RUN_TTL = 7 * 24 * 3600 * 1000;
const HL_TTL = 30 * 24 * 3600 * 1000;

export function runKey(c: TRunConfig): string {
  // Listening practice and exam runs of the same parts are different runs.
  const mode = c.skill === "listening" ? `-${c.mode}` : "";
  const scope = c.redo ? `redo-${c.redo.attemptId}` : `${c.kind}-${[...c.parts].sort().join("")}${mode}`;
  return `${RUN_PREFIX}${c.skill}.${c.test}.${scope}`;
}

function read<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key);
    return raw ? (JSON.parse(raw) as T) : null;
  } catch {
    return null;
  }
}

function write(key: string, value: unknown) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // storage blocked or full: the run lives in memory only
  }
}

function remove(key: string) {
  try {
    localStorage.removeItem(key);
  } catch {
    // nothing to remove
  }
}

export function loadRun(key: string): TRunState | null {
  const s = read<TRunState>(key);
  if (!s || Date.now() - s.savedAt > RUN_TTL) return null;
  return s;
}

export function saveRun(key: string, s: TRunState) {
  write(key, { ...s, savedAt: Date.now() });
}

export function clearRun(key: string) {
  remove(key);
}

export function newRun(c: TRunConfig, firstQuestion: number): TRunState {
  const now = Date.now();
  return {
    answers: {},
    flags: [],
    highlights: [],
    current: firstQuestion,
    startedAt: now,
    deadline: c.timeLimitSec > 0 ? now + c.timeLimitSec * 1000 : null,
    savedAt: now,
    mode: c.mode,
  };
}

/** The tests with a run in progress for a skill, keyed by test slug, with the run's key. */
export function runsInProgress(skill: TSkill): Record<string, string[]> {
  const out: Record<string, string[]> = {};
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const k = localStorage.key(i);
      if (!k?.startsWith(`${RUN_PREFIX}${skill}.`)) continue;
      const s = loadRun(k);
      if (!s) continue;
      const test = k.slice(`${RUN_PREFIX}${skill}.`.length).split(".")[0];
      (out[test] ??= []).push(k);
    }
  } catch {
    // storage blocked
  }
  return out;
}

export type TRunInProgress = {
  key: string;
  kind: TAttemptKind;
  parts: number[];
  timeSec: number;
  answered: number;
  redo?: string;
  mode?: TListenMode;
};

/** Reads back what a stored run is, from its key and state (for "Continue" links on the test bank). */
export function describeRun(key: string): TRunInProgress | null {
  const s = loadRun(key);
  const scope = key.split(".").pop() ?? "";
  if (!s) return null;
  const answered = Object.values(s.answers).filter((v) => (Array.isArray(v) ? v.length > 0 : v.trim() !== "")).length;
  const timeSec = s.deadline ? Math.round((s.deadline - s.startedAt) / 1000) : 0;
  if (scope.startsWith("redo-")) return { key, kind: "redo", parts: [], timeSec: 0, answered, redo: scope.slice(5) };
  const [kind, digits = ""] = scope.split("-");
  if (kind !== "part" && kind !== "full") return null;
  return { key, kind, parts: digits.split("").map(Number), timeSec, answered, mode: s.mode };
}

/** Keeps the run's highlights for the review screen of the attempt it became. */
export function keepHighlights(attemptId: string, highlights: THighlight[]) {
  if (highlights.length) write(HL_PREFIX + attemptId, { highlights, savedAt: Date.now() });
  pruneHighlights();
}

export function loadAttemptHighlights(attemptId: string): THighlight[] {
  return read<{ highlights: THighlight[] }>(HL_PREFIX + attemptId)?.highlights ?? [];
}

export function saveAttemptHighlights(attemptId: string, highlights: THighlight[]) {
  write(HL_PREFIX + attemptId, { highlights, savedAt: Date.now() });
}

function pruneHighlights() {
  try {
    const stale: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const k = localStorage.key(i);
      if (!k?.startsWith(HL_PREFIX) && !k?.startsWith(RUN_PREFIX)) continue;
      const v = read<{ savedAt?: number }>(k);
      const ttl = k.startsWith(HL_PREFIX) ? HL_TTL : RUN_TTL;
      if (!v?.savedAt || Date.now() - v.savedAt > ttl) stale.push(k);
    }
    stale.forEach(remove);
  } catch {
    // storage blocked
  }
}
