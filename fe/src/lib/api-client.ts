import { refreshSession, signOut } from "@/lib/auth-actions";
import type { TApiErrorBody } from "@/types/api";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    public readonly fields: Record<string, string> = {},
  ) {
    super(code);
    this.name = "ApiError";
  }
}

// One refresh at a time per tab; concurrent 401s wait for the same promise.
let refreshing: Promise<boolean> | null = null;

function refreshOnce(): Promise<boolean> {
  refreshing ??= refreshSession()
    .then((r) => r.ok)
    .catch(() => false)
    .finally(() => {
      refreshing = null;
    });
  return refreshing;
}

async function handleSessionExpired(code: string): Promise<never> {
  await signOut().catch(() => undefined);
  const locale = document.documentElement.lang || "en";
  const next = `${window.location.pathname}${window.location.search}`;
  // eslint-disable-next-line @next/next/no-location-assign-relative-destination -- runs outside React; hard reload drops stale state
  window.location.assign(
    `/${locale}/login?error=${encodeURIComponent(code)}&next=${encodeURIComponent(next)}`,
  );
  throw new ApiError(401, code);
}

async function toApiError(res: Response): Promise<ApiError> {
  const body = (await res.json().catch(() => null)) as TApiErrorBody | null;
  return new ApiError(res.status, body?.error?.code ?? "UNKNOWN", body?.error?.fields ?? {});
}

async function request<TData>(path: string, init: RequestInit, isRetry = false): Promise<TData> {
  let res: Response;
  try {
    res = await fetch(`/api/v1${path}`, {
      ...init,
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
        ...(init.body ? { "Content-Type": "application/json" } : {}),
        ...init.headers,
      },
    });
  } catch {
    throw new ApiError(0, "NETWORK_ERROR");
  }

  if (res.status === 401) {
    const error = await toApiError(res);
    if (!isRetry) {
      // Retry even if this tab's refresh lost a race: another tab may have stored new cookies.
      await refreshOnce();
      return request<TData>(path, init, true);
    }
    return handleSessionExpired(error.code);
  }
  if (!res.ok) {
    throw await toApiError(res);
  }
  if (res.status === 204) {
    return undefined as TData;
  }
  return (await res.json()) as TData;
}

export const apiClient = {
  get: <TData>(path: string) => request<TData>(path, { method: "GET" }),
  post: <TData>(path: string, body?: unknown) =>
    request<TData>(path, { method: "POST", body: body === undefined ? undefined : JSON.stringify(body) }),
  patch: <TData>(path: string, body: unknown) =>
    request<TData>(path, { method: "PATCH", body: JSON.stringify(body) }),
  delete: <TData>(path: string) => request<TData>(path, { method: "DELETE" }),
};
