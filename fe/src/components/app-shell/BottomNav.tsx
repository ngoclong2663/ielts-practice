"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { NAV_ITEMS, isActive } from "./nav-items";

// Mobile tab bar: thumb-reachable, respects the iOS home indicator safe area.
export function BottomNav() {
  return <BottomNavView pathname={usePathname()} />;
}

export function BottomNavView({ pathname }: { pathname: string }) {
  const t = useTranslations("nav");

  return (
    <nav
      aria-label={t("mainNavigation")}
      className="fixed inset-x-0 bottom-0 z-30 border-t border-border bg-background/95 pb-[env(safe-area-inset-bottom)] backdrop-blur supports-[backdrop-filter]:bg-background/80 lg:hidden"
    >
      <ul className="grid grid-cols-4">
        {NAV_ITEMS.map(({ href, labelKey, icon: Icon }) => {
          const active = isActive(pathname, href);
          return (
            <li key={href}>
              <Link
                href={href}
                aria-current={active ? "page" : undefined}
                className={cn(
                  "flex flex-col items-center gap-1 px-1 py-2.5 text-[11px] font-medium transition-colors",
                  active ? "text-primary" : "text-muted-foreground",
                )}
              >
                <Icon className="size-5" strokeWidth={1.75} aria-hidden />
                <span className="truncate max-w-full">{t(labelKey)}</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
