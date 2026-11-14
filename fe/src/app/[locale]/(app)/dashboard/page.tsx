import { CalendarCheck } from "lucide-react";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { buttonVariants } from "@/components/ui/button";
import { DashboardGreeting } from "@/features/dashboard/components/DashboardGreeting";
import { Link } from "@/i18n/navigation";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("nav");
  return { title: t("dashboard") };
}

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");
  return (
    <div className="mx-auto max-w-5xl">
      <DashboardGreeting />
      <section className="flex flex-col items-start gap-4 rounded-2xl border border-border bg-card p-8 sm:p-10">
        <span className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
          <CalendarCheck className="size-5" strokeWidth={1.75} aria-hidden />
        </span>
        <div>
          <h2 className="text-lg font-semibold tracking-tight">{t("emptyTitle")}</h2>
          <p className="mt-1.5 max-w-[60ch] text-sm leading-relaxed text-muted-foreground">{t("emptyBody")}</p>
        </div>
        <Link href="/settings" className={buttonVariants({ variant: "outline" })}>
          {t("emptyCta")}
        </Link>
      </section>
    </div>
  );
}
