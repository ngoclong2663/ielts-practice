import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { FeedbackView } from "@/features/admin/components/FeedbackView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("admin.feedback");
  return { title: t("title") };
}

export default function AdminFeedbackPage() {
  return (
    // FeedbackView reads ?id= (the link in the Telegram message).
    <Suspense fallback={<Skeleton className="mx-auto h-72 max-w-3xl rounded-2xl" />}>
      <FeedbackView />
    </Suspense>
  );
}
