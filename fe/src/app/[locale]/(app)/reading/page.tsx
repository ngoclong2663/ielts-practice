import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { TestBankView } from "@/features/exam/components/bank/TestBankView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("exam");
  return { title: t("bank.title.reading") };
}

export default function ReadingPage() {
  return <TestBankView skill="reading" />;
}
