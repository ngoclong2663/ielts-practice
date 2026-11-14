import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { StudyScreen } from "@/features/study/components/StudyScreen";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("study");
  return { title: t("title") };
}

export default function StudyPage() {
  return (
    <Suspense fallback={<Skeleton className="mx-auto mt-24 h-16 w-2/3 max-w-xl" />}>
      <StudyScreen />
    </Suspense>
  );
}
