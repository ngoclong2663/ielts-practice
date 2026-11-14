import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { AddWordsView } from "@/features/decks/components/AddWordsView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("addWords");
  return { title: t("title") };
}

export default function AddWordsPage() {
  return <AddWordsView />;
}
