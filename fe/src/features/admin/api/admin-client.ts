import { ApiError } from "@/lib/api-client";
import type { TApiErrorBody } from "@/types/api";

// The admin board calls /api/v1/admin/* through the same-origin rewrite; the httpOnly
// admin_token cookie goes with it and the API checks it on every call (Epic 6 §5).
async function request<TData>(path: string, init: RequestInit = {}): Promise<TData> {
  let res: Response;
  try {
    res = await fetch(`/api/v1/admin${path}`, {
      ...init,
      credentials: "same-origin",
      headers: { Accept: "application/json", ...(init.body ? { "Content-Type": "application/json" } : {}) },
    });
  } catch {
    throw new ApiError(0, "NETWORK_ERROR");
  }
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as TApiErrorBody | null;
    const code = body?.error?.code ?? "UNKNOWN";
    if (res.status === 401) {
      // The admin session expired: back to the admin sign-in, then here again.
      const locale = document.documentElement.lang || "en";
      const next = `${window.location.pathname}${window.location.search}`;
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination -- runs outside React; hard reload drops stale state
      window.location.assign(`/${locale}/admin/login?next=${encodeURIComponent(next)}`);
    }
    throw new ApiError(res.status, code, body?.error?.fields ?? {}, body?.error?.params ?? {});
  }
  return (await res.json()) as TData;
}

export const adminClient = {
  get: <TData>(path: string) => request<TData>(path),
  patch: <TData>(path: string, body: unknown) => request<TData>(path, { method: "PATCH", body: JSON.stringify(body) }),
  post: <TData>(path: string, body: unknown) => request<TData>(path, { method: "POST", body: JSON.stringify(body) }),
};
