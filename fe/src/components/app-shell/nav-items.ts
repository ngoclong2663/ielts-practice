import { CirclePlus, Layers, LayoutDashboard, Settings } from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type TNavItem = {
  href: "/dashboard" | "/decks" | "/add-words" | "/settings";
  labelKey: "dashboard" | "decks" | "addWords" | "settings";
  icon: LucideIcon;
};

export const NAV_ITEMS: TNavItem[] = [
  { href: "/dashboard", labelKey: "dashboard", icon: LayoutDashboard },
  { href: "/decks", labelKey: "decks", icon: Layers },
  { href: "/add-words", labelKey: "addWords", icon: CirclePlus },
  { href: "/settings", labelKey: "settings", icon: Settings },
];

export function isActive(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(`${href}/`);
}
