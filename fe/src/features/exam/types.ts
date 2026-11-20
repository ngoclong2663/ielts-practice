// Epic 10: IELTS Reading and Listening practice. Mirrors be/app/modules/exam/dto and entity/content.go.

export type TSkill = "reading" | "listening";
export type TAttemptKind = "part" | "full" | "redo";
export type TListenMode = "practice" | "exam";

export type TGroupType =
  | "tfng"
  | "ynng"
  | "mcq"
  | "choose_n"
  | "match_headings"
  | "match_info"
  | "match_features"
  | "match_endings"
  | "summary_bank"
  | "sentence"
  | "notes"
  | "table"
  | "flowchart"
  | "form"
  | "summary"
  | "short_answer"
  | "diagram"
  | "map";

export type TLocalized = { vi: string; en: string };

export type TOption = { key: string; text?: string };
export type TItem = { n: number; text: string; options?: TOption[] };

export type TGroup = {
  type: TGroupType;
  from: number;
  to: number;
  instructions: string;
  title?: string;
  word_limit?: number;
  allow_number?: boolean;
  pick?: number;
  stem?: string;
  options?: TOption[];
  reuse_options?: boolean;
  items?: TItem[];
  lines?: string[];
  header?: string[];
  rows?: string[][];
  figure?: string;
};

export type TParagraph = { id: string; label?: string; text: string };
export type TLine = { id: string; speaker?: string; text: string; at: number };

export type TContent = {
  subtitle?: string;
  paragraphs?: TParagraph[];
  lines?: TLine[];
  groups: TGroup[];
};

export type TKeyItem = {
  answer: string[];
  evidence?: { block: string; quote: string };
  pairs?: [string, string][];
  why: string;
  trap?: string;
};
export type TKey = Record<string, TKeyItem>;

export type TCue = { at: number; seconds: number; label: string };
export type TAudio = { url: string; duration: number; cues: TCue[] };

export type TScore = { correct: number; total: number };

export type TPartSummary = {
  part_no: number;
  title: string;
  q_from: number;
  q_to: number;
  question_types: TGroupType[];
  best: TScore | null;
};

export type TTestSummary = { slug: string; test_no: number; parts: TPartSummary[] };
export type TBook = { slug: string; title: TLocalized; tests: TTestSummary[] };

export type TPart = {
  part_no: number;
  title: string;
  q_from: number;
  q_to: number;
  question_types: TGroupType[];
  content: TContent;
  audio: TAudio | null;
  key?: TKey;
};

export type TTest = {
  slug: string;
  test_no: number;
  skill: TSkill;
  book_slug: string;
  book_title: TLocalized;
  parts: TPart[];
};

/** A question's answer: a string, or for a "Choose TWO" group (keyed by its first number) a list of letters. */
export type TAnswers = Record<string, string | string[]>;

export type TResult = {
  n: number;
  part_no: number;
  type: TGroupType;
  given: string;
  correct: boolean;
  over_limit?: boolean;
};

export type TAttemptSummary = {
  id: string;
  skill: TSkill;
  test_slug: string;
  test_no: number;
  book_title: TLocalized;
  part_nos: number[];
  kind: TAttemptKind;
  mode: TListenMode;
  correct: number;
  total: number;
  band: number | null;
  duration_sec: number;
  time_limit_sec: number;
  created_at: string;
};

export type TAttempt = TAttemptSummary & {
  parent_attempt_id?: string;
  questions: number[];
  answers: TAnswers;
  results: TResult[];
  test: TTest;
};

export type TAttemptList = { items: TAttemptSummary[]; next_before?: string };

export type TStats = {
  attempts: number;
  by_type: { type: TGroupType; correct: number; total: number }[];
  bands: { attempt_id: string; band: number | null; correct: number; at: string }[];
};

export type TSubmitAttempt = {
  test: string;
  skill: TSkill;
  kind: TAttemptKind;
  parts?: number[];
  questions?: number[];
  parent_attempt_id?: string;
  mode?: TListenMode;
  time_limit_sec: number;
  duration_sec: number;
  answers: TAnswers;
};

/** A highlight in one text block (paragraph "p-B", question stem "q7"): character offsets in its text. */
export type THighlight = { block: string; start: number; end: number };
