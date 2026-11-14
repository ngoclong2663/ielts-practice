import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { DecksView } from "@/features/decks/components/DecksView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("decks");
  return { title: t("title") };
}

export default function DecksPage() {
  return <DecksView />;
}
