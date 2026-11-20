import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { ResultView } from "@/features/exam/components/result/ResultView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("result.pageTitle") };
}

export default function ReadingResultPage() {
  return (
    <Suspense fallback={<Skeleton className="mx-auto mt-16 h-96 max-w-3xl rounded-xl" />}>
      <ResultView />
    </Suspense>
  );
}
