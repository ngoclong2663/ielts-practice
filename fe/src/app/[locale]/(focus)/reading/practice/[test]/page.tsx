import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { ReadingPlayer } from "@/features/exam/components/player/ReadingPlayer";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("bank.title.reading") };
}

export default function ReadingPracticePage() {
  return (
    <Suspense fallback={<Skeleton className="h-dvh rounded-none" />}>
      <ReadingPlayer />
    </Suspense>
  );
}
