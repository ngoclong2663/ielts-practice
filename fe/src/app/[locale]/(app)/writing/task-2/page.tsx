import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { TaskListView } from "@/features/writing/components/TaskListView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("writing");
  return { title: t("task2.title") };
}

export default function WritingTask2Page() {
  return <TaskListView taskNo={2} />;
}
