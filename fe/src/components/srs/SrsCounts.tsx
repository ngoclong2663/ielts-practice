import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

type SrsCountsProps = {
  counts: { new: number; learning: number; review: number };
  /** Underlines the group the current card belongs to (study screen). */
  active?: "new" | "learning" | "review";
  className?: string;
};

// Anki's three numbers: new (blue) · learning (orange) · review (green).
export function SrsCounts({ counts, active, className }: SrsCountsProps) {
  const t = useTranslations("srs");
  const items = [
    { key: "new", value: counts.new, color: "text-srs-new" },
    { key: "learning", value: counts.learning, color: "text-srs-learn" },
    { key: "review", value: counts.review, color: "text-srs-review" },
  ] as const;
  return (
    <span className={cn("inline-flex items-center gap-3 font-mono text-sm tabular-nums", className)}>
      {items.map((it) => (
        <span
          key={it.key}
          title={t(it.key)}
          aria-label={`${t(it.key)}: ${it.value}`}
          className={cn(
            "font-semibold",
            it.color,
            active === it.key && "underline decoration-2 underline-offset-4",
            it.value === 0 && "opacity-40",
          )}
        >
          {it.value}
        </span>
      ))}
    </span>
  );
}
