import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { routing } from "@/i18n/routing";
import { OAUTH_STATE_COOKIE } from "@/lib/session-constants";
import { serverEnv } from "@/lib/server/env";
import { createPkce, randomToken, safeNextPath } from "@/lib/server/oauth";
import type { TOAuthState } from "@/lib/server/oauth";
import { isHttpsRequest } from "@/lib/server/request-scheme";

const GOOGLE_AUTHORIZE_URL = "https://accounts.google.com/o/oauth2/v2/auth";

// GET /api/auth/google/start?next=/vi/dashboard&locale=vi&tz=Asia/Ho_Chi_Minh
export function GET(request: NextRequest) {
  const params = request.nextUrl.searchParams;
  const locale = (routing.locales as readonly string[]).includes(params.get("locale") ?? "")
    ? (params.get("locale") as string)
    : routing.defaultLocale;

  if (!serverEnv.googleClientId) {
    console.error("[auth/google/start] OAUTH_CLIENT_ID is not set");
    return NextResponse.redirect(
      new URL(`/${locale}/login?error=AUTH_GOOGLE_NOT_CONFIGURED`, request.url),
    );
  }

  const { verifier, challenge } = createPkce();
  const oauthState: TOAuthState = {
    state: randomToken(),
    nonce: randomToken(),
    verifier,
    locale,
    next: safeNextPath(params.get("next"), `/${locale}/dashboard`),
    tz: (params.get("tz") ?? "").slice(0, 64),
  };

  const authorizeUrl = new URL(GOOGLE_AUTHORIZE_URL);
  authorizeUrl.search = new URLSearchParams({
    client_id: serverEnv.googleClientId,
    redirect_uri: serverEnv.googleRedirectUrl,
    response_type: "code",
    scope: "openid email profile",
    state: oauthState.state,
    nonce: oauthState.nonce,
    code_challenge: challenge,
    code_challenge_method: "S256",
    prompt: "select_account",
  }).toString();

  const response = NextResponse.redirect(authorizeUrl);
  response.cookies.set(OAUTH_STATE_COOKIE, JSON.stringify(oauthState), {
    httpOnly: true,
    secure: isHttpsRequest(request.headers, request.nextUrl),
    sameSite: "lax",
    path: "/api/auth/google",
    maxAge: 600,
  });
  return response;
}
