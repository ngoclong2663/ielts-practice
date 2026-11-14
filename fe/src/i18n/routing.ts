import { defineRouting } from "next-intl/routing";

export const routing = defineRouting({
  locales: ["en", "vi"],
  // Used only when the browser prefers neither vi nor en and no NEXT_LOCALE cookie exists.
  defaultLocale: "en",
  localePrefix: "always",
  localeDetection: true,
});

export type TLocale = (typeof routing.locales)[number];
