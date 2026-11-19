"use client";

import { Library, LogOut, MessageSquareText, ShieldCheck, Users, LayoutDashboard } from "lucide-react";
import { useTranslations } from "next-intl";
import { useTransition } from "react";
import type { ReactNode } from "react";
import { BrandMark } from "@/components/shared/BrandMark";
import { LanguageSwitcher } from "@/components/shared/LanguageSwitcher";
import { ThemeToggle } from "@/components/shared/ThemeToggle";
import { Button } from "@/components/ui/button";
import { Link, usePathname, useRouter } from "@/i18n/navigation";
import { adminSignOut } from "@/lib/admin-actions";
import { cn } from "@/lib/utils";

const TABS = [
  { href: "/admin", labelKey: "overview", icon: LayoutDashboard, exact: true },
  { href: "/admin/users", labelKey: "users", icon: Users, exact: false },
  { href: "/admin/feedback", labelKey: "feedback", icon: MessageSquareText, exact: false },
  { href: "/admin/library", labelKey: "library", icon: Library, exact: false },
] as const;

export function AdminTabs({ pathname }: { pathname: string }) {
  const t = useTranslations("admin.nav");
  return (
    <nav aria-label={t("label")} className="-mb-px flex gap-1 overflow-x-auto">
      {TABS.map((tab) => {
        const active = tab.exact ? pathname === tab.href : pathname === tab.href || pathname.startsWith(`${tab.href}/`);
        return (
          <Link
            key={tab.href}
            href={tab.href}
            aria-current={active ? "page" : undefined}
            className={cn(
              "inline-flex items-center gap-2 border-b-2 px-3 py-2.5 text-sm font-medium whitespace-nowrap transition-colors",
              active ? "border-primary text-foreground" : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            <tab.icon className="size-4" strokeWidth={1.75} aria-hidden />
            {t(tab.labelKey)}
          </Link>
        );
      })}
    </nav>
  );
}

export function AdminTabsLive() {
  return <AdminTabs pathname={usePathname()} />;
}

function SignOutButton() {
  const t = useTranslations("admin.nav");
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  return (
    <Button
      variant="ghost"
      size="sm"
      aria-label={t("signOut")}
      disabled={pending}
      onClick={() =>
        startTransition(async () => {
          await adminSignOut();
          router.replace("/admin/login");
        })
      }
    >
      <LogOut className="size-4" strokeWidth={1.75} aria-hidden />
      <span className="hidden sm:inline">{t("signOut")}</span>
    </Button>
  );
}

/** The admin board frame: no learner navigation, four tabs, sign out (Epic 6 §6). */
export function AdminShell({ tabs, children }: { tabs: ReactNode; children: ReactNode }) {
  const t = useTranslations("admin.nav");
  return (
    <div className="flex min-h-[100dvh] flex-col">
      <header className="sticky top-0 z-20 border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
        <div className="px-4 sm:px-6">
          <div className="mx-auto flex h-14 max-w-6xl items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <BrandMark />
              <span className="inline-flex items-center gap-1 rounded-md bg-accent px-2 py-0.5 text-xs font-medium text-accent-foreground">
                <ShieldCheck className="size-3.5" strokeWidth={1.75} aria-hidden />
                {t("badge")}
              </span>
            </div>
            <div className="flex items-center gap-1">
              <LanguageSwitcher />
              <ThemeToggle />
              <SignOutButton />
            </div>
          </div>
          <div className="mx-auto max-w-6xl">{tabs}</div>
        </div>
      </header>
      <main className="flex-1 px-4 pt-6 pb-16 sm:px-6 lg:pt-10">{children}</main>
    </div>
  );
}
