"use client";

import { useTranslations } from "next-intl";
import { Skeleton } from "@/components/ui/skeleton";
import { useMe } from "@/features/profile/hooks/use-me";
import { useIsClient } from "@/hooks/use-is-client";

export function DashboardGreeting() {
  const t = useTranslations("dashboard");
  const me = useMe();
  const isClient = useIsClient();
  const isPending = !isClient || me.isPending;
  const data = isClient ? me.data : undefined;

  return (
    <div className="mb-8">
      {isPending ? (
        <Skeleton className="h-9 w-56" />
      ) : (
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          {data?.user.display_name ? t("greeting", { name: data.user.display_name.split(" ")[0] }) : t("greetingFallback")}
        </h1>
      )}
      <p className="mt-1.5 text-sm text-muted-foreground">{t("subtitle")}</p>
    </div>
  );
}
