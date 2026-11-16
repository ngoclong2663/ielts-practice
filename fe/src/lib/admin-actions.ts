"use server";

import { cookies, headers } from "next/headers";
import { ADMIN_TOKEN_COOKIE } from "@/lib/session-constants";
import { backendPost } from "@/lib/server/backend";
import { isHttpsRequest } from "@/lib/server/request-scheme";

// `username` is sent back so the field keeps it after React resets the form on a failed try.
export type TAdminSignInState = { ok: boolean; code?: string; username?: string };

type TAdminLoginResponse = {
  access_token: string;
  expires_at: string;
  admin: { id: string; username: string };
};

/** Signs in with the admin account seeded in the database (`api seed`) and stores its token. */
export async function adminSignIn(_prev: TAdminSignInState, formData: FormData): Promise<TAdminSignInState> {
  const headerStore = await headers();
  const username = String(formData.get("username") ?? "");
  const result = await backendPost<TAdminLoginResponse>(
    "/api/v1/admin/auth/login",
    { username, password: String(formData.get("password") ?? "") },
    headerStore.get("x-forwarded-for"),
  );
  if (!result.ok) {
    return { ok: false, code: result.code, username };
  }
  const cookieStore = await cookies();
  cookieStore.set(ADMIN_TOKEN_COOKIE, result.data.access_token, {
    httpOnly: true,
    secure: isHttpsRequest(headerStore),
    sameSite: "strict",
    path: "/",
    expires: new Date(result.data.expires_at),
  });
  return { ok: true };
}

export async function adminSignOut(): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.delete(ADMIN_TOKEN_COOKIE);
}
