import "server-only";
import { createHash, randomBytes } from "node:crypto";

export type TOAuthState = {
  state: string;
  nonce: string;
  verifier: string;
  next: string;
  locale: string;
  tz: string;
};

const base64url = (buf: Buffer) => buf.toString("base64url");

export function createPkce(): { verifier: string; challenge: string } {
  const verifier = base64url(randomBytes(48)); // 64 chars, within the 43-128 PKCE range
  const challenge = base64url(createHash("sha256").update(verifier).digest());
  return { verifier, challenge };
}

export function randomToken(): string {
  return base64url(randomBytes(24));
}

// Only allow same-site relative paths to avoid open redirects after login.
export function safeNextPath(next: string | null | undefined, fallback: string): string {
  if (!next || !next.startsWith("/") || next.startsWith("//") || next.includes("\\")) {
    return fallback;
  }
  return next;
}
