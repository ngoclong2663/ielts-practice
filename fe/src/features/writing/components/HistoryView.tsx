"use client";

import { useTranslations } from "next-intl";
import { PageHeader } from "@/components/shared/PageHeader";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useAttempts } from "../hooks/use-writing";
import { AttemptRow } from "./TaskDetailView";

export function HistoryView() {
  const t = useTranslations("writing");
  const isClient = useIsClient();
  const query = useAttempts();
  const attempts = isClient ? query.data : undefined;

  return (
    <div className="mx-auto max-w-3xl">
      <PageHeader title={t("history.title")} description={t("history.subtitle")} />
      {!attempts ? (
        <Skeleton className="h-48 rounded-2xl" />
      ) : attempts.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border p-8 text-center">
          <p className="text-sm text-muted-foreground">{t("history.empty")}</p>
          <Link href="/writing/task-1" className={buttonVariants({ className: "mt-4" })}>
            {t("history.start")}
          </Link>
        </div>
      ) : (
        <ul className="divide-y divide-border rounded-2xl border border-border">
          {attempts.map((a) => (
            <li key={a.id}>
              <AttemptRow attempt={a} showTask />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
