import type { TAnswers, TGroup, TGroupType, TPart } from "../types";

export type TAnswerMode = "choice" | "choose_n" | "letter" | "gap";

/** How a group is answered (mirrors entity.AnswerMode on the API). */
export function answerMode(type: TGroupType): TAnswerMode {
  switch (type) {
    case "tfng":
    case "ynng":
    case "mcq":
      return "choice";
    case "choose_n":
      return "choose_n";
    case "match_headings":
    case "match_info":
    case "match_features":
    case "match_endings":
    case "summary_bank":
    case "map":
      return "letter";
    default:
      return "gap";
  }
}

export const range = (from: number, to: number) => Array.from({ length: to - from + 1 }, (_, i) => from + i);

/** Every question number of the parts, in order. */
export function questionsOf(parts: TPart[]): number[] {
  return parts.flatMap((p) => range(p.q_from, p.q_to));
}

export function groupOf(parts: TPart[], n: number): TGroup | undefined {
  for (const p of parts) for (const g of p.content.groups) if (n >= g.from && n <= g.to) return g;
  return undefined;
}

export function partOf(parts: TPart[], n: number): TPart | undefined {
  return parts.find((p) => n >= p.q_from && n <= p.q_to);
}

/** A "Choose TWO" group stores its letters under its first number; question from+i counts once i+1 letters are chosen. */
export function isAnswered(answers: TAnswers, n: number, group: TGroup | undefined): boolean {
  if (group && group.type === "choose_n") {
    const v = answers[String(group.from)];
    return Array.isArray(v) && v.length > n - group.from;
  }
  const v = answers[String(n)];
  return typeof v === "string" ? v.trim() !== "" : false;
}

export function countWords(s: string): number {
  const t = s.trim();
  return t ? t.split(/\s+/).length : 0;
}

/** Client-side copy of rules.WithinLimit, for the live word counter only (the API marks). */
export function overLimit(answer: string, limit = 0, allowNumber = false): boolean {
  if (!limit && !allowNumber) return false;
  const toks = answer.trim().toLowerCase().split(/\s+/).filter(Boolean);
  const isNum = (t: string) => /^[£$€]?\d[\d,.:/]*(%|[a-z]{1,2})?$/.test(t);
  const nums = toks.filter(isNum).length;
  const words = toks.length - nums;
  return allowNumber ? words > limit || nums > 1 : words + nums > limit;
}

/** Splits a text with {n} placeholders into text and gap pieces. */
export function splitGaps(text: string): (string | number)[] {
  const out: (string | number)[] = [];
  const re = /\{(\d+)\}/g;
  let last = 0;
  for (let m = re.exec(text); m; m = re.exec(text)) {
    if (m.index > last) out.push(text.slice(last, m.index));
    out.push(Number(m[1]));
    last = m.index + m[0].length;
  }
  if (last < text.length) out.push(text.slice(last));
  return out;
}
