import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { ExamHistoryView } from "@/features/exam/components/history/ExamHistoryView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("history.title.reading") };
}

export default function ReadingHistoryPage() {
  return <ExamHistoryView skill="reading" />;
}
