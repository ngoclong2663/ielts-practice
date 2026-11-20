import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { ListeningPlayer } from "@/features/exam/components/player/ListeningPlayer";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("bank.title.listening") };
}

export default function ListeningPracticePage() {
  return (
    <Suspense fallback={<Skeleton className="h-dvh rounded-none" />}>
      <ListeningPlayer />
    </Suspense>
  );
}
