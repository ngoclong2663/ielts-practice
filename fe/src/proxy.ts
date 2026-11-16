import createMiddleware from "next-intl/middleware";
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { routing } from "@/i18n/routing";
import { ACCESS_TOKEN_COOKIE, ADMIN_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/session-constants";

const intlMiddleware = createMiddleware(routing);

const PROTECTED_PREFIXES = ["/dashboard", "/decks", "/add-words", "/settings", "/study", "/writing"];
const GUEST_ONLY_PREFIXES = ["/login"];
// Admin board (Epic 6): its own session. Optimistic check only (cookie presence); the API
// verifies the admin token on every call and the board goes back to sign-in on a 401.
const ADMIN_PREFIX = "/admin";
const ADMIN_LOGIN = "/admin/login";

function stripLocale(pathname: string): { locale: string | null; path: string } {
  const [, maybeLocale, ...rest] = pathname.split("/");
  if ((routing.locales as readonly string[]).includes(maybeLocale)) {
    return { locale: maybeLocale, path: `/${rest.join("/")}` };
  }
  return { locale: null, path: pathname };
}

function matches(path: string, prefixes: string[]): boolean {
  return prefixes.some((p) => path === p || path.startsWith(`${p}/`));
}

// Optimistic check only (cookie presence). The Gin API verifies every request, and the
// API client refreshes an expired access token using the refresh cookie.
export function proxy(request: NextRequest) {
  const response = intlMiddleware(request);
  if (response.headers.get("location")) {
    return response;
  }

  const { locale, path } = stripLocale(request.nextUrl.pathname);
  if (!locale) {
    return response;
  }

  if (matches(path, [ADMIN_PREFIX])) {
    // The sign-in page never redirects away: an expired cookie would otherwise loop.
    if (path !== ADMIN_LOGIN && !request.cookies.has(ADMIN_TOKEN_COOKIE)) {
      const url = new URL(`/${locale}${ADMIN_LOGIN}`, request.url);
      url.searchParams.set("next", `${request.nextUrl.pathname}${request.nextUrl.search}`);
      return NextResponse.redirect(url);
    }
    return response;
  }

  const hasSession =
    request.cookies.has(ACCESS_TOKEN_COOKIE) || request.cookies.has(REFRESH_TOKEN_COOKIE);

  if (!hasSession && matches(path, PROTECTED_PREFIXES)) {
    const url = new URL(`/${locale}/login`, request.url);
    url.searchParams.set("next", `${request.nextUrl.pathname}${request.nextUrl.search}`);
    return NextResponse.redirect(url);
  }

  if (hasSession && matches(path, GUEST_ONLY_PREFIXES)) {
    return NextResponse.redirect(new URL(`/${locale}/dashboard`, request.url));
  }

  return response;
}

export const config = {
  // Skip API routes, Next internals and files with an extension.
  matcher: "/((?!api|_next|_vercel|.*\\..*).*)",
};
