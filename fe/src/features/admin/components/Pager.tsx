"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";

type TPagerProps = { page: number; pageSize: number; total: number; onPage: (page: number) => void };

export function Pager({ page, pageSize, total, onPage }: TPagerProps) {
  const t = useTranslations("admin");
  const pages = Math.max(1, Math.ceil(total / pageSize));
  if (pages <= 1) return null;
  return (
    <div className="mt-4 flex items-center justify-end gap-2 text-sm text-muted-foreground">
      <span>{t("page", { page, pages })}</span>
      <Button size="icon-sm" variant="outline" disabled={page <= 1} onClick={() => onPage(page - 1)} aria-label={t("previous")}>
        <ChevronLeft className="size-4" aria-hidden />
      </Button>
      <Button size="icon-sm" variant="outline" disabled={page >= pages} onClick={() => onPage(page + 1)} aria-label={t("next")}>
        <ChevronRight className="size-4" aria-hidden />
      </Button>
    </div>
  );
}
