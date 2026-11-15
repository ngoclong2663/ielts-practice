import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { HistoryView } from "@/features/writing/components/HistoryView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("writing");
  return { title: t("history.title") };
}

export default function WritingHistoryPage() {
  return <HistoryView />;
}
