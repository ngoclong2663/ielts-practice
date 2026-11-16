"use client";

import { CircleAlert } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";


export function QueryError({ onRetry }: { onRetry: () => void }) {
  const t = useTranslations();
  return (
    <div role="alert" className="mb-6 flex flex-wrap items-center gap-3 rounded-xl border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
      <CircleAlert className="size-4 shrink-0" strokeWidth={1.75} aria-hidden />
      <span className="flex-1">{t("admin.loadFailed")}</span>
      <Button size="sm" variant="outline" onClick={onRetry}>
        {t("common.retry")}
      </Button>
    </div>
  );
}
