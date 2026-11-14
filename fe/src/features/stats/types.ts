export type TStatsOverview = {
  today: { reviews: number; new_cards: number; again: number; duration_ms: number };
  streak_days: number;
  studied_today: boolean;
  forecast: { date: string; count: number }[];
  retention_30d: number | null;
  retention_sample: number;
  totals: { new: number; learning: number; review: number; suspended: number };
};
