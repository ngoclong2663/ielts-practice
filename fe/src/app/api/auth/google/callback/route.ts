import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { routing } from "@/i18n/routing";
import { OAUTH_STATE_COOKIE } from "@/lib/session-constants";
import { backendPost } from "@/lib/server/backend";
import { isHttpsRequest } from "@/lib/server/request-scheme";
import type { TOAuthState } from "@/lib/server/oauth";
import { writeSessionCookies } from "@/lib/server/session-cookies";
import type { TSession } from "@/types/api";

function readState(request: NextRequest): TOAuthState | null {
  const raw = request.cookies.get(OAUTH_STATE_COOKIE)?.value;
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as TOAuthState;
  } catch {
    return null;
  }
}

// GET /api/auth/google/callback?code=...&state=...  (Google redirects here)
export async function GET(request: NextRequest) {
  const params = request.nextUrl.searchParams;
  const saved = readState(request);
  const locale = saved?.locale ?? routing.defaultLocale;

  const fail = (code: string, detail?: string) => {
    // Never log the code, tokens or cookie contents.
    console.warn(`[auth/google/callback] sign-in failed: ${code}${detail ? ` (${detail})` : ""}`);
    const res = NextResponse.redirect(new URL(`/${locale}/login?error=${code}`, request.url));
    res.cookies.delete({ name: OAUTH_STATE_COOKIE, path: "/api/auth/google" });
    return res;
  };

  const googleError = params.get("error");
  if (googleError) {
    return fail("AUTH_GOOGLE_CANCELLED", `google error=${googleError}`);
  }
  const code = params.get("code");
  if (!saved) {
    return fail("AUTH_STATE_MISMATCH", "state cookie missing: open the app on the same host as OAUTH_REDIRECT_URL");
  }
  if (!code || params.get("state") !== saved.state) {
    return fail("AUTH_STATE_MISMATCH", "state does not match");
  }

  const result = await backendPost<TSession>(
    "/api/v1/auth/google",
    {
      code,
      code_verifier: saved.verifier,
      nonce: saved.nonce,
      ui_locale: locale,
      timezone: saved.tz || undefined,
    },
    request.headers.get("x-forwarded-for"),
  );
  if (!result.ok) {
    return fail(result.code, `api status ${result.status}`);
  }

  // Send the user to their saved interface language (it may differ from the login page).
  const next = saved.next.replace(/^\/(en|vi)(?=\/|$)/, `/${result.data.user.ui_locale}`);
  const response = NextResponse.redirect(new URL(next, request.url));
  const secure = isHttpsRequest(request.headers, request.nextUrl);
  writeSessionCookies(response.cookies, result.data, secure);
  console.info(`[auth/google/callback] signed in user ${result.data.user.id}`);
  response.cookies.delete({ name: OAUTH_STATE_COOKIE, path: "/api/auth/google" });
  response.cookies.set("NEXT_LOCALE", result.data.user.ui_locale, { path: "/", sameSite: "lax", secure });
  return response;
}
