"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useMe } from "../hooks/use-me";
import { ProfileForm } from "./ProfileForm";
import { StudySettingsForm } from "./StudySettingsForm";

function SettingsSkeleton() {
  return (
    <div className="space-y-10" aria-hidden>
      {[0, 1].map((i) => (
        <div key={i} className="grid grid-cols-1 gap-6 lg:grid-cols-[16rem_1fr] lg:gap-12">
          <div className="space-y-2">
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-48" />
          </div>
          <div className="max-w-xl space-y-5">
            <Skeleton className="h-9 w-full" />
            <Skeleton className="h-9 w-full" />
            <Skeleton className="h-9 w-2/3" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function SettingsView() {
  const t = useTranslations();
  const { data, isPending, isError, refetch } = useMe();

  if (isPending) {
    return <SettingsSkeleton />;
  }
  if (isError) {
    return (
      <div className="rounded-xl border border-border p-6">
        <p className="text-sm text-muted-foreground">{t("settings.loadError")}</p>
        <Button variant="outline" className="mt-4" onClick={() => refetch()}>
          {t("common.retry")}
        </Button>
      </div>
    );
  }

  return (
    <div>
      <ProfileForm key={`profile-${data.user.id}`} user={data.user} />
      <StudySettingsForm key={`study-${data.user.id}`} settings={data.study_settings} />
    </div>
  );
}
