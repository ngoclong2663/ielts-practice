import "server-only";

type THeaderReader = { get: (name: string) => string | null };

// True when the browser reached us over HTTPS (directly or via Vercel's proxy).
// Cookies must not be `Secure` on plain http://localhost, or Safari drops them.
export function isHttpsRequest(headers: THeaderReader, url?: URL): boolean {
  const forwarded = headers.get("x-forwarded-proto")?.split(",")[0]?.trim();
  if (forwarded) {
    return forwarded === "https";
  }
  return url?.protocol === "https:";
}
