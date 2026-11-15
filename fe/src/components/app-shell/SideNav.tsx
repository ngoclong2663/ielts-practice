"use client";

import { ChevronDown } from "lucide-react";
import { useTranslations } from "next-intl";
import { BrandMark } from "@/components/shared/BrandMark";
import { Link, usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";
import { HOME_ITEM, NAV_SECTIONS, SETTINGS_ITEM, isActive, sectionOf } from "./nav-items";
import type { TNavItem } from "./nav-items";
import { setSectionCollapsed, useCollapsedSections } from "./nav-state";

// The path is request data, so the shell renders SideNavView without an active item first.
export function SideNav() {
  return <SideNavView pathname={usePathname()} />;
}

function NavLink({ item, pathname, nested = false }: { item: TNavItem; pathname: string; nested?: boolean }) {
  const t = useTranslations("nav");
  const active = isActive(pathname, item.href);
  const Icon = item.icon;
  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
        nested && "pl-9",
        active
          ? "bg-sidebar-accent text-sidebar-accent-foreground"
          : "text-muted-foreground hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
      )}
    >
      <Icon className="size-4" strokeWidth={1.75} aria-hidden />
      {t(item.labelKey)}
    </Link>
  );
}

export function SideNavView({ pathname }: { pathname: string }) {
  const t = useTranslations("nav");
  const collapsed = useCollapsedSections();
  const current = sectionOf(pathname);

  return (
    <aside className="sticky top-0 hidden h-[100dvh] w-64 shrink-0 flex-col border-r border-sidebar-border bg-sidebar px-4 py-5 lg:flex">
      <Link href="/dashboard" className="px-2">
        <BrandMark />
      </Link>
      <nav aria-label={t("mainNavigation")} className="mt-8 flex flex-1 flex-col gap-1 overflow-y-auto">
        <NavLink item={HOME_ITEM} pathname={pathname} />
        {NAV_SECTIONS.map((section) => {
          // The section holding the current page always stays open.
          const open = section.id === current?.id || !collapsed.has(section.id);
          const Icon = section.icon;
          const listId = `nav-section-${section.id}`;
          return (
            <div key={section.id} className="mt-3">
              <button
                type="button"
                aria-expanded={open}
                aria-controls={listId}
                onClick={() => setSectionCollapsed(section.id, open)}
                disabled={section.id === current?.id}
                className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground transition-colors hover:text-sidebar-foreground disabled:cursor-default disabled:hover:text-muted-foreground"
              >
                <Icon className="size-4" strokeWidth={1.75} aria-hidden />
                <span className="flex-1 text-left">{t(`sections.${section.id}`)}</span>
                <ChevronDown
                  className={cn("size-4 transition-transform", !open && "-rotate-90")}
                  strokeWidth={1.75}
                  aria-hidden
                />
              </button>
              {open && (
                <ul id={listId} className="mt-1 flex flex-col gap-1">
                  {section.items.map((item) => (
                    <li key={item.href}>
                      <NavLink item={item} pathname={pathname} nested />
                    </li>
                  ))}
                </ul>
              )}
            </div>
          );
        })}
        <div className="mt-auto pt-4">
          <NavLink item={SETTINGS_ITEM} pathname={pathname} />
        </div>
      </nav>
    </aside>
  );
}
