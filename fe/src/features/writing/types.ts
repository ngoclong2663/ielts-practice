export type TTaskNo = 1 | 2;

export const TASK1_TYPES = ["line_graph", "bar_chart", "pie_chart", "table", "mixed", "map", "process"] as const;
export const TASK2_TYPES = ["opinion", "discussion", "advantages_disadvantages", "problem_solution", "two_part"] as const;
export type TTaskType = (typeof TASK1_TYPES)[number] | (typeof TASK2_TYPES)[number];

export type TPosition =
  | "agree"
  | "disagree"
  | "partly"
  | "favour_first"
  | "favour_second"
  | "advantages"
  | "disadvantages"
  | "default";

export type TLevel = 1 | 2 | 3 | 4;

export type TParagraphKey = "introduction" | "overview" | "body1" | "body2" | "conclusion";

export type TFigure = { url: string; alt: string; width: number; height: number };

export type TTaskSummary = {
  id: string;
  task_no: TTaskNo;
  task_type: TTaskType;
  topic: string;
  title: string;
  source: { book: number; test: number } | null;
  figure: TFigure | null;
  attempts: number;
  best_band: number | null;
};

export type TTask = TTaskSummary & {
  prompt: string;
  figures: TFigure[];
  positions: TPosition[];
  min_words: number;
  minutes: number;
};

export type TAttemptSummary = {
  id: string;
  task_id: string;
  task_title: string;
  task_no: TTaskNo;
  task_type: TTaskType;
  level: TLevel;
  position: TPosition | "";
  status: "in_progress" | "submitted";
  band_low: number | null;
  band_high: number | null;
  created_at: string;
  submitted_at: string | null;
};

export type TTaskDetail = TTask & { history: TAttemptSummary[] };

export type THint = { en: string; vi: string };

export type TPlanQuestion = {
  question_vi: string;
  options_vi: string[];
  answer: number | null;
  explain_vi?: string;
};

export type TVerdict = "correct" | "understandable" | "wrong" | "empty";

export type TErrorType = "literal_translation" | "grammar" | "word_choice" | "spelling" | "missing_idea" | "punctuation";

export type TSentenceFeedback = {
  verdict: TVerdict;
  error_types: TErrorType[];
  explain_vi: string;
  corrected_en: string;
};

export type TParagraph = {
  key: TParagraphKey;
  goal_vi: string;
  sentences: { prompt: string; hints: THint[]; model_en?: string }[];
  answers: string[];
  feedback: TSentenceFeedback[];
  checked_count: number;
};

export type TCriterionKey =
  | "task_achievement"
  | "task_response"
  | "coherence_cohesion"
  | "lexical_resource"
  | "grammatical_range";

export type TEssayFeedback = {
  criteria: { key: TCriterionKey; band: number; comment_vi: string }[];
  summary_vi: string;
  fixes: { quote: string; better: string; explain_vi: string }[];
  model: string;
};

export type TSubmission = {
  text: string;
  word_count: number;
  seconds_spent: number;
  band_low: number;
  band_high: number;
  feedback: TEssayFeedback;
  created_at: string;
};

export type TAttempt = {
  id: string;
  task: TTask;
  level: TLevel;
  position: TPosition | "";
  status: "in_progress" | "submitted";
  ideas_vi: string[];
  questions: TPlanQuestion[];
  plan_answers: number[] | null;
  paragraphs: TParagraph[];
  useful_language: THint[];
  model_answer?: string;
  submission: TSubmission | null;
  created_at: string;
  submitted_at: string | null;
};

export type TQuota = { limit: number; used: number; remaining: number };
export type TWritingStatus = { configured: boolean; checks: TQuota; submissions: TQuota };
