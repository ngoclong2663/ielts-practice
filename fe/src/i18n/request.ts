import { hasLocale } from "next-intl";
import { getRequestConfig } from "next-intl/server";
import { locale as rootLocale } from "next/root-params";
import { routing } from "./routing";

// The locale comes from the `[locale]` root param (not request headers), so pages stay
// statically prerenderable with Cache Components.
export default getRequestConfig(async ({ locale }) => {
  const candidate = locale ?? (await rootLocale());
  const resolved = hasLocale(routing.locales, candidate) ? candidate : routing.defaultLocale;

  return {
    locale: resolved,
    timeZone: "Asia/Ho_Chi_Minh",
    messages: (await import(`../../messages/${resolved}.json`)).default,
  };
});
