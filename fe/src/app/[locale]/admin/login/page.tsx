import type { Metadata } from "next";
import { ShieldCheck } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { BrandMark } from "@/components/shared/BrandMark";
import { LanguageSwitcher } from "@/components/shared/LanguageSwitcher";
import { ThemeToggle } from "@/components/shared/ThemeToggle";
import { Skeleton } from "@/components/ui/skeleton";
import { AdminLoginForm } from "@/features/admin/components/AdminLoginForm";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("admin.login");
  return { title: t("title") };
}

export default async function AdminLoginPage() {
  const t = await getTranslations("admin.login");
  return (
    <div className="flex min-h-[100dvh] flex-col px-4 py-6 sm:px-8">
      <div className="flex items-center justify-between">
        <BrandMark />
        <div className="flex items-center gap-1">
          <LanguageSwitcher />
          <ThemeToggle />
        </div>
      </div>
      <main className="flex flex-1 items-center justify-center py-12">
        <div className="w-full max-w-sm">
          <span className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
            <ShieldCheck className="size-5" strokeWidth={1.75} aria-hidden />
          </span>
          <h1 className="mt-5 text-3xl font-semibold tracking-tight">{t("title")}</h1>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{t("subtitle")}</p>
          <div className="mt-8">
            {/* The form reads ?next= from the URL. */}
            <Suspense fallback={<Skeleton className="h-56 w-full rounded-lg" />}>
              <AdminLoginForm />
            </Suspense>
          </div>
        </div>
      </main>
    </div>
  );
}
