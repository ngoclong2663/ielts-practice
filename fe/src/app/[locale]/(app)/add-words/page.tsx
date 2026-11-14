import { CirclePlus } from "lucide-react";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { ComingSoon } from "@/components/shared/ComingSoon";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("placeholder");
  return { title: t("addWordsTitle") };
}

export default async function Page() {
  const t = await getTranslations("placeholder");
  return <ComingSoon icon={CirclePlus} title={t("addWordsTitle")} body={t("addWordsBody")} />;
}
