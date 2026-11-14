import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { AddWordsView } from "@/features/vocab/components/AddWordsView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("addWords");
  return { title: t("title") };
}

export default function AddWordsPage() {
  return (
    // AddWordsView reads ?job= and ?mode= from the URL.
    <Suspense fallback={<Skeleton className="mx-auto h-72 max-w-3xl rounded-2xl" />}>
      <AddWordsView />
    </Suspense>
  );
}
