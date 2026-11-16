"use client";

import { useFormatter, useTranslations } from "next-intl";
import { useState } from "react";
import { cn } from "@/lib/utils";
import type { TAdminDaily } from "../types";

const METRICS = ["active_users", "signups", "reviews", "ai_requests"] as const;
type TMetric = (typeof METRICS)[number];

const PLOT_HEIGHT = 140; // px

/**
 * The last 30 days, one metric at a time (a switch above the plot, no legend: the selected
 * metric names the series). Same anatomy as the learner forecast chart: thin bars rounded at
 * the data end, one baseline, a tooltip per day and a visually hidden table.
 */
export function DailyChart({ days }: { days: TAdminDaily[] }) {
  const t = useTranslations("admin.overview.daily");
  const format = useFormatter();
  const [metric, setMetric] = useState<TMetric>("active_users");
  const values = days.map((d) => d[metric]);
  const max = Math.max(1, ...values);
  const peakIndex = values.indexOf(Math.max(...values));
  const total = values.reduce((a, b) => a + b, 0);

  const dayLabel = (date: string) => format.dateTime(new Date(`${date}T12:00:00`), { day: "numeric", month: "short" });

  return (
    <figure>
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <div role="radiogroup" aria-label={t("metric")} className="inline-flex flex-wrap gap-1 rounded-lg bg-muted p-1">
          {METRICS.map((m) => (
            <button
              key={m}
              type="button"
              role="radio"
              aria-checked={metric === m}
              onClick={() => setMetric(m)}
              className={cn(
                "rounded-md px-2.5 py-1 text-xs font-medium transition-colors",
                metric === m ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground",
              )}
            >
              {t(m)}
            </button>
          ))}
        </div>
        <p className="text-sm text-muted-foreground">{t("total", { count: format.number(total) })}</p>
      </div>

      <div className="relative" style={{ height: PLOT_HEIGHT }}>
        <div className="absolute inset-x-0 bottom-0 h-px bg-border" aria-hidden />
        <div className="absolute inset-0 grid gap-[2px]" style={{ gridTemplateColumns: `repeat(${days.length}, minmax(0, 1fr))` }} aria-hidden>
          {days.map((d, i) => {
            const v = values[i];
            const h = v === 0 ? 0 : Math.max(4, Math.round((v / max) * (PLOT_HEIGHT - 22)));
            const label = `${dayLabel(d.date)}: ${format.number(v)}`;
            return (
              <div key={d.date} tabIndex={0} aria-label={label} className="group relative flex h-full items-end justify-center outline-none">
                {v > 0 && <div className="w-full max-w-4 rounded-t-[4px] bg-chart-1 transition-opacity group-hover:opacity-80" style={{ height: h }} />}
                {i === peakIndex && v > 0 && (
                  <span className="absolute font-mono text-xs text-muted-foreground" style={{ bottom: h + 4 }}>
                    {format.number(v)}
                  </span>
                )}
                <span
                  className={cn(
                    "pointer-events-none absolute bottom-full z-10 mb-1 hidden rounded-md border border-border bg-popover px-2 py-1 text-xs whitespace-nowrap text-popover-foreground shadow-sm group-hover:block group-focus-visible:block",
                    i < 4 ? "left-0" : i > days.length - 5 ? "right-0" : "",
                  )}
                >
                  {label}
                </span>
              </div>
            );
          })}
        </div>
      </div>
      <div className="mt-2 flex justify-between text-xs text-muted-foreground" aria-hidden>
        <span>{days[0] && dayLabel(days[0].date)}</span>
        <span>{days[Math.floor(days.length / 2)] && dayLabel(days[Math.floor(days.length / 2)].date)}</span>
        <span>{t("today")}</span>
      </div>
      <div className="sr-only">
        <table>
          <caption>{t(metric)}</caption>
          <tbody>
            {days.map((d, i) => (
              <tr key={d.date}>
                <th scope="row">{dayLabel(d.date)}</th>
                <td>{values[i]}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </figure>
  );
}
