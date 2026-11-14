import "server-only";
import { ACCESS_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/session-constants";
import type { TSession } from "@/types/api";

type TCookieWriter = {
  set: (name: string, value: string, options: Record<string, unknown>) => unknown;
  delete: (name: string) => unknown;
};

function baseOptions(expiresAt: string, secure: boolean) {
  return {
    httpOnly: true,
    secure,
    sameSite: "lax" as const,
    path: "/",
    expires: new Date(expiresAt),
  };
}

// Works with both `NextResponse.cookies` and `await cookies()` inside Server Actions.
export function writeSessionCookies(store: TCookieWriter, session: TSession, secure: boolean): void {
  store.set(
    ACCESS_TOKEN_COOKIE,
    session.access_token,
    baseOptions(session.access_token_expires_at, secure),
  );
  store.set(
    REFRESH_TOKEN_COOKIE,
    session.refresh_token,
    baseOptions(session.refresh_token_expires_at, secure),
  );
}

export function clearSessionCookies(store: TCookieWriter): void {
  store.delete(ACCESS_TOKEN_COOKIE);
  store.delete(REFRESH_TOKEN_COOKIE);
}
