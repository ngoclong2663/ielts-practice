"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { BrandMark } from "@/components/shared/BrandMark";
import { NAV_ITEMS, isActive } from "./nav-items";

// The path is request data, so the shell renders SideNavView without an active item first.
export function SideNav() {
  return <SideNavView pathname={usePathname()} />;
}

export function SideNavView({ pathname }: { pathname: string }) {
  const t = useTranslations("nav");

  return (
    <aside className="sticky top-0 hidden h-[100dvh] w-64 shrink-0 flex-col border-r border-sidebar-border bg-sidebar px-4 py-5 lg:flex">
      <Link href="/dashboard" className="px-2">
        <BrandMark />
      </Link>
      <nav aria-label={t("mainNavigation")} className="mt-8 flex flex-col gap-1">
        {NAV_ITEMS.map(({ href, labelKey, icon: Icon }) => {
          const active = isActive(pathname, href);
          return (
            <Link
              key={href}
              href={href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                active
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-muted-foreground hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
              )}
            >
              <Icon className="size-4" strokeWidth={1.75} aria-hidden />
              {t(labelKey)}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
