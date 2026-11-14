import type { BrowserContext } from "@playwright/test";

const API_URL = process.env.E2E_API_URL ?? "http://localhost:8081";

/** Signs in through the test-only endpoint and stores the session cookies in the browser. */
export async function signIn(context: BrowserContext, baseURL: string, locale: "en" | "vi" = "en") {
  const email = `e2e-${Date.now()}-${Math.random().toString(36).slice(2, 8)}@example.com`;
  const res = await fetch(`${API_URL}/api/v1/auth/test-login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, display_name: "Minh Anh", ui_locale: locale }),
  });
  if (!res.ok) {
    throw new Error(`test-login failed: ${res.status} (is the API running with APP_ENV=test?)`);
  }
  const session = await res.json();
  const url = new URL(baseURL);
  await context.addCookies([
    { name: "access_token", value: session.access_token, domain: url.hostname, path: "/", httpOnly: true, sameSite: "Lax" },
    { name: "refresh_token", value: session.refresh_token, domain: url.hostname, path: "/", httpOnly: true, sameSite: "Lax" },
    { name: "NEXT_LOCALE", value: locale, domain: url.hostname, path: "/", sameSite: "Lax" },
  ]);
  return session;
}
