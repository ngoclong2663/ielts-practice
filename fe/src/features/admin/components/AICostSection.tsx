"use client";

import { useFormatter, useTranslations } from "next-intl";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { TAICost, TAIProvider } from "../types";
import { StatSection, StatTile } from "./StatTile";

const PROVIDERS: TAIProvider[] = ["anthropic", "google"];

/** Epic 7: what the language models cost, by provider, model and feature. */
export function AICostSection({ cost }: { cost: TAICost }) {
  const t = useTranslations("admin.overview.aiCost");
  const format = useFormatter();
  const usd = (v: number) =>
    format.number(v, { style: "currency", currency: "USD", minimumFractionDigits: 2, maximumFractionDigits: v > 0 && v < 1 ? 4 : 2 });
  const compact = (v: number) => format.number(v, { notation: "compact", maximumFractionDigits: 1 });
  const n = (v: number) => format.number(v);
  const byProvider = new Map(cost.providers.map((p) => [p.provider, p]));
  // Ledger features look like "writing.lesson"; message keys can't hold dots.
  const featureLabel = (feature: string) => {
    const key = `features.${feature.replace(".", "_")}`;
    return t.has(key) ? t(key) : feature;
  };

  return (
    <div className="space-y-4">
      <StatSection title={t("title")}>
        <StatTile
          label={t("total30d")}
          value={usd(cost.cost_30d_usd)}
          hint={t("totalHint", {
            total: usd(cost.cost_total_usd),
            since: cost.since ? format.dateTime(new Date(cost.since), { dateStyle: "medium" }) : "-",
          })}
        />
        {PROVIDERS.map((provider) => {
          const p = byProvider.get(provider);
          return (
            <StatTile
              key={provider}
              label={t("provider30d", { provider: t(`providers.${provider}`) })}
              value={usd(p?.cost_30d_usd ?? 0)}
              hint={t("providerHint", { calls: n(p?.calls_30d ?? 0), total: usd(p?.cost_total_usd ?? 0) })}
            />
          );
        })}
      </StatSection>

      <div className="overflow-x-auto rounded-xl border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("model")}</TableHead>
              <TableHead className="text-right">{t("calls30d")}</TableHead>
              <TableHead className="text-right">{t("tokens30d")}</TableHead>
              <TableHead className="text-right">{t("cost30d")}</TableHead>
              <TableHead className="text-right">{t("costTotal")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {cost.models.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="py-6 text-center text-muted-foreground">
                  {t("empty")}
                </TableCell>
              </TableRow>
            )}
            {cost.models.map((m) => (
              <TableRow key={m.model}>
                <TableCell>
                  <span className="font-medium">{m.model}</span>
                  <span className="block text-xs text-muted-foreground">
                    {t(`providers.${m.provider}`)} ·{" "}
                    {m.price
                      ? t("price", { input: usd(m.price.input), output: usd(m.price.output) })
                      : t("noPrice")}
                  </span>
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {n(m.calls_30d)}
                  {m.failed_30d > 0 && (
                    <span className="block text-xs text-destructive">{t("failed", { count: n(m.failed_30d) })}</span>
                  )}
                </TableCell>
                <TableCell className="text-right text-xs tabular-nums text-muted-foreground">
                  {t("tokens", {
                    input: compact(m.input_tokens_30d),
                    output: compact(m.output_tokens_30d),
                    cached: compact(m.cache_read_tokens_30d),
                  })}
                </TableCell>
                <TableCell className="text-right font-medium tabular-nums">{m.price ? usd(m.cost_30d_usd) : "?"}</TableCell>
                <TableCell className="text-right tabular-nums">{m.price ? usd(m.cost_total_usd) : "?"}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {cost.features.some((f) => f.calls_30d > 0) && (
        <ul className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {cost.features
            .filter((f) => f.calls_30d > 0)
            .map((f) => (
              <li key={f.feature} className="flex items-baseline justify-between gap-3 rounded-xl border border-border bg-card px-4 py-2.5 text-sm">
                <span>{featureLabel(f.feature)}</span>
                <span className="tabular-nums text-muted-foreground">
                  {usd(f.cost_30d_usd)} · {t("featureCalls", { count: n(f.calls_30d) })}
                </span>
              </li>
            ))}
        </ul>
      )}
      <p className="text-xs text-muted-foreground">{t("note")}</p>
    </div>
  );
}
