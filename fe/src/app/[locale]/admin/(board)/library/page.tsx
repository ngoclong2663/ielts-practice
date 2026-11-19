import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { CandidatesView } from "@/features/admin/components/CandidatesView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("admin.library");
  return { title: t("title") };
}

export default function AdminLibraryPage() {
  return <CandidatesView />;
}
