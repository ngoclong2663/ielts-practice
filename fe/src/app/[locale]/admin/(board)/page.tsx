import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { OverviewView } from "@/features/admin/components/OverviewView";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("admin.overview");
  return { title: t("title") };
}

export default function AdminOverviewPage() {
  return <OverviewView />;
}
