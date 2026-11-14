import "server-only";
import type { TApiErrorBody } from "@/types/api";
import { serverEnv } from "./env";

export type TBackendResult<TData> =
  | { ok: true; data: TData }
  | { ok: false; status: number; code: string };

// Server-to-server call to the Gin API. Forwards the client IP so per-IP rate limits
// apply to the real user instead of the Next.js server.
export async function backendPost<TData>(
  path: string,
  body: unknown,
  forwardedFor?: string | null,
): Promise<TBackendResult<TData>> {
  try {
    const res = await fetch(`${serverEnv.apiUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
        ...(forwardedFor ? { "X-Forwarded-For": forwardedFor } : {}),
      },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    if (res.status === 204) {
      return { ok: true, data: undefined as TData };
    }
    const json = (await res.json().catch(() => null)) as TData | TApiErrorBody | null;
    if (!res.ok) {
      const code = (json as TApiErrorBody | null)?.error?.code ?? "UNKNOWN";
      return { ok: false, status: res.status, code };
    }
    return { ok: true, data: json as TData };
  } catch {
    return { ok: false, status: 503, code: "NETWORK_ERROR" };
  }
}
