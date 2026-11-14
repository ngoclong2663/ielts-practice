"use server";

import { cookies, headers } from "next/headers";
import { REFRESH_TOKEN_COOKIE } from "@/lib/session-constants";
import { backendPost } from "@/lib/server/backend";
import { isHttpsRequest } from "@/lib/server/request-scheme";
import { clearSessionCookies, writeSessionCookies } from "@/lib/server/session-cookies";
import type { TSession } from "@/types/api";

/**
 * Rotates the session using the httpOnly refresh cookie.
 * Returns `ok: false` when the user must sign in again.
 */
export async function refreshSession(): Promise<{ ok: boolean; code?: string }> {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(REFRESH_TOKEN_COOKIE)?.value;
  if (!refreshToken) {
    return { ok: false, code: "AUTH_UNAUTHORIZED" };
  }
  const headerStore = await headers();
  const result = await backendPost<TSession>(
    "/api/v1/auth/refresh",
    { refresh_token: refreshToken },
    headerStore.get("x-forwarded-for"),
  );
  if (!result.ok) {
    // AUTH_REFRESH_INVALID can be a harmless race with another tab that already rotated the
    // token and stored new cookies, so cookies are only cleared here on token reuse. The API
    // client retries once and calls signOut() when the session is really gone.
    if (result.code === "AUTH_REFRESH_REUSED") {
      clearSessionCookies(cookieStore);
    }
    return { ok: false, code: result.code };
  }
  writeSessionCookies(cookieStore, result.data, isHttpsRequest(headerStore));
  return { ok: true };
}

/** Revokes the refresh token on the API and clears the session cookies. */
export async function signOut(): Promise<void> {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(REFRESH_TOKEN_COOKIE)?.value;
  if (refreshToken) {
    await backendPost("/api/v1/auth/logout", { refresh_token: refreshToken });
  }
  clearSessionCookies(cookieStore);
}
