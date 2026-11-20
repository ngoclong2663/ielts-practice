import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { ExamHistoryView } from "@/features/exam/components/history/ExamHistoryView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("history.title.listening") };
}

export default function ListeningHistoryPage() {
  return <ExamHistoryView skill="listening" />;
}
