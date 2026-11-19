import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { CollectionView } from "@/features/library/components/CollectionView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("library");
  return { title: t("title") };
}

export default function LibraryCollectionPage() {
  return (
    // CollectionView reads the slug and ?topic= from the URL.
    <Suspense fallback={<Skeleton className="mx-auto h-72 max-w-3xl rounded-2xl" />}>
      <CollectionView />
    </Suspense>
  );
}
