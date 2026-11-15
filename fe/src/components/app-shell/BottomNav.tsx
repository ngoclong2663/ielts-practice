"use client";

import { useTranslations } from "next-intl";
import { useEffect } from "react";
import { Link, usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { HOME_ITEM, NAV_SECTIONS, isActive, sectionOf } from "./nav-items";
import { rememberSection, useLastSection } from "./nav-state";

// Mobile tab bar: Home, the current section's pages, and a switch to the other section.
// Thumb-reachable; respects the iOS home indicator safe area. Settings is in the user menu.
export function BottomNav() {
  return <BottomNavView pathname={usePathname()} />;
}

export function BottomNavView({ pathname }: { pathname: string }) {
  const t = useTranslations("nav");
  const last = useLastSection();
  const current = sectionOf(pathname);
  const section = current ?? NAV_SECTIONS.find((s) => s.id === last) ?? NAV_SECTIONS[0];
  const other = NAV_SECTIONS.find((s) => s.id !== section.id) ?? NAV_SECTIONS[0];

  useEffect(() => {
    if (current) rememberSection(current.id);
  }, [current]);

  const cells = [
    { href: HOME_ITEM.href, label: t(HOME_ITEM.labelKey), icon: HOME_ITEM.icon, active: isActive(pathname, HOME_ITEM.href) },
    ...section.items.map((item) => ({
      href: item.href,
      label: t(item.shortLabelKey ?? item.labelKey),
      icon: item.icon,
      active: isActive(pathname, item.href),
    })),
  ];
  const SwitchIcon = other.icon;

  return (
    <nav
      aria-label={t("mainNavigation")}
      className="fixed inset-x-0 bottom-0 z-30 border-t border-border bg-background/95 pb-[env(safe-area-inset-bottom)] backdrop-blur supports-[backdrop-filter]:bg-background/80 lg:hidden"
    >
      <ul className="grid" style={{ gridTemplateColumns: `repeat(${cells.length + 1}, minmax(0, 1fr))` }}>
        {cells.map(({ href, label, icon: Icon, active }) => (
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
              <span className="max-w-full truncate">{label}</span>
            </Link>
          </li>
        ))}
        <li className="border-l border-border">
          <Link
            href={other.items[0].href}
            aria-label={t("switchTo", { section: t(`sections.${other.id}`) })}
            className="flex flex-col items-center gap-1 px-1 py-2.5 text-[11px] font-medium text-muted-foreground transition-colors"
          >
            <SwitchIcon className="size-5" strokeWidth={1.75} aria-hidden />
            <span className="max-w-full truncate">{t(`sections.${other.id}`)}</span>
          </Link>
        </li>
      </ul>
    </nav>
  );
}
