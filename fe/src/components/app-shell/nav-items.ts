import { BookOpenText, ChartLine, CirclePlus, FileText, History, Layers, LayoutDashboard, Library, PenLine, Settings } from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type TNavLabel =
  | "dashboard"
  | "decks"
  | "addWords"
  | "library"
  | "settings"
  | "writingTask1"
  | "writingTask2"
  | "writingHistory"
  | "writingTask1Short"
  | "writingTask2Short";

export type TNavItem = {
  href: string;
  labelKey: TNavLabel;
  /** Label for the narrow mobile tab bar. */
  shortLabelKey?: TNavLabel;
  icon: LucideIcon;
};

export type TSectionId = "vocabulary" | "writing";

export type TNavSection = {
  id: TSectionId;
  icon: LucideIcon;
  items: TNavItem[];
  /** Paths that belong to the section (its pages and their sub-pages). */
  prefixes: string[];
};

export const HOME_ITEM: TNavItem = { href: "/dashboard", labelKey: "dashboard", icon: LayoutDashboard };
export const SETTINGS_ITEM: TNavItem = { href: "/settings", labelKey: "settings", icon: Settings };

export const NAV_SECTIONS: TNavSection[] = [
  {
    id: "vocabulary",
    icon: BookOpenText,
    prefixes: ["/decks", "/add-words", "/library"],
    items: [
      { href: "/decks", labelKey: "decks", icon: Layers },
      { href: "/add-words", labelKey: "addWords", icon: CirclePlus },
      { href: "/library", labelKey: "library", icon: Library },
    ],
  },
  {
    id: "writing",
    icon: PenLine,
    prefixes: ["/writing"],
    items: [
      { href: "/writing/task-1", labelKey: "writingTask1", shortLabelKey: "writingTask1Short", icon: ChartLine },
      { href: "/writing/task-2", labelKey: "writingTask2", shortLabelKey: "writingTask2Short", icon: FileText },
      { href: "/writing/history", labelKey: "writingHistory", icon: History },
    ],
  },
];

export function isActive(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function sectionOf(pathname: string): TNavSection | undefined {
  return NAV_SECTIONS.find((s) => s.prefixes.some((p) => isActive(pathname, p)));
}
