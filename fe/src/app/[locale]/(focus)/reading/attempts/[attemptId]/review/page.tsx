import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { ReviewView } from "@/features/exam/components/review/ReviewView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("review.title") };
}

export default function ReadingReviewPage() {
  return (
    <Suspense fallback={<Skeleton className="h-dvh rounded-none" />}>
      <ReviewView />
    </Suspense>
  );
}
