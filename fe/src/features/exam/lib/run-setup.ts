import type { TPart, TSkill, TTest } from "../types";
import { parseParts } from "./format";
import type { TRunConfig } from "./run-state";

function parseNumbers(raw: string | null): number[] {
  return (raw ?? "")
    .split(",")
    .map(Number)
    .filter((n) => Number.isInteger(n) && n >= 1 && n <= 40);
}

/**
 * Reads a player URL: ?parts=1,2&time=2400 (part practice), ?kind=full (full test),
 * ?redo=<attempt>&q=6,9 (the wrong questions of an earlier attempt); listening adds &mode=exam|practice
 * (a full test defaults to exam mode, everything else to practice).
 */
export function runSetup(test: TTest, sp: { get(name: string): string | null }, skill: TSkill): { parts: TPart[]; config: TRunConfig } {
  const redoId = sp.get("redo");
  const redoQs = parseNumbers(sp.get("q"));
  let kind: TRunConfig["kind"] = sp.get("kind") === "full" ? "full" : "part";
  let parts: TPart[];
  if (redoId && redoQs.length) {
    kind = "redo";
    parts = test.parts.filter((p) => redoQs.some((n) => n >= p.q_from && n <= p.q_to));
  } else if (kind === "full") {
    parts = test.parts;
  } else {
    const chosen = parseParts(sp.get("parts"));
    parts = test.parts.filter((p) => chosen.includes(p.part_no));
  }
  const time = Number(sp.get("time") ?? 0);
  const timeLimitSec =
    skill === "listening" || kind === "redo" ? 0 : kind === "full" ? 3600 : Number.isFinite(time) && time > 0 ? Math.min(time, 7200) : 0;
  const modeParam = sp.get("mode");
  const mode = kind === "redo" ? "practice" : modeParam === "exam" || modeParam === "practice" ? modeParam : kind === "full" ? "exam" : "practice";
  return {
    parts,
    config: {
      skill,
      test: test.slug,
      kind,
      parts: parts.map((p) => p.part_no),
      timeLimitSec,
      mode: skill === "listening" ? mode : "practice",
      redo: kind === "redo" && redoId ? { attemptId: redoId, questions: redoQs } : undefined,
    },
  };
}
