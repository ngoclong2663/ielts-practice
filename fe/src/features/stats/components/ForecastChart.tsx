"use client";

import { useFormatter, useTranslations } from "next-intl";

type ForecastChartProps = {
  days: { date: string; count: number }[];
};

const PLOT_HEIGHT = 112; // px

/**
 * Seven columns, one series (no legend; the title names it). Bars are capped at 24px,
 * rounded only at the data end, grow from one baseline; hover or focus a column for
 * its exact value, and a visually hidden table carries every number.
 */
export function ForecastChart({ days }: ForecastChartProps) {
  const t = useTranslations("stats");
  const format = useFormatter();
  const max = Math.max(1, ...days.map((d) => d.count));
  const peak = days.reduce((best, d) => (d.count > best.count ? d : best), days[0] ?? { date: "", count: 0 });

  const dayLabel = (date: string, i: number) =>
    i === 0 ? t("forecastToday") : format.dateTime(new Date(`${date}T12:00:00`), { weekday: "short" });

  return (
    <figure>
      <div className="relative" style={{ height: PLOT_HEIGHT }}>
        {/* Recessive baseline */}
        <div className="absolute inset-x-0 bottom-0 h-px bg-border" aria-hidden />
        <div className="absolute inset-0 grid grid-cols-7" aria-hidden>
          {days.map((d, i) => {
            const h = d.count === 0 ? 0 : Math.max(4, Math.round((d.count / max) * (PLOT_HEIGHT - 20)));
            const label = `${dayLabel(d.date, i)}: ${t("forecastCards", { count: d.count })}`;
            return (
              <div key={d.date} tabIndex={0} aria-label={label} className="group relative flex h-full items-end justify-center outline-none">
                {d.count > 0 && (
                  <div className="w-full max-w-6 rounded-t-[4px] bg-chart-1 transition-opacity group-hover:opacity-80" style={{ height: h }} />
                )}
                {/* Direct label only on the peak day; the tooltip carries the rest. */}
                {d.date === peak.date && d.count > 0 && (
                  <span className="absolute font-mono text-xs text-muted-foreground" style={{ bottom: h + 4 }}>
                    {d.count}
                  </span>
                )}
                <span className="pointer-events-none absolute bottom-full z-10 mb-1 hidden rounded-md border border-border bg-popover px-2 py-1 text-xs whitespace-nowrap text-popover-foreground shadow-sm group-hover:block group-focus-visible:block">
                  {label}
                </span>
              </div>
            );
          })}
        </div>
      </div>
      <div className="mt-2 grid grid-cols-7 text-center text-xs text-muted-foreground" aria-hidden>
        {days.map((d, i) => (
          <span key={d.date} className={i === 0 ? "font-medium text-foreground" : undefined}>
            {dayLabel(d.date, i)}
          </span>
        ))}
      </div>
      {/* sr-only goes on a wrapper: a <table> ignores the 1px height, so it would still take up its
          full height and make the page scroll past the end of the layout. */}
      <div className="sr-only">
        <table>
          <caption>{t("forecastLabel")}</caption>
          <tbody>
            {days.map((d, i) => (
              <tr key={d.date}>
                <th scope="row">{dayLabel(d.date, i)}</th>
                <td>{d.count}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </figure>
  );
}
