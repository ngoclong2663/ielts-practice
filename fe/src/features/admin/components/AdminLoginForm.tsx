"use client";

import { CircleAlert } from "lucide-react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { useActionState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRouter } from "@/i18n/navigation";
import { adminSignIn } from "@/lib/admin-actions";
import type { TAdminSignInState } from "@/lib/admin-actions";

const INITIAL: TAdminSignInState = { ok: false };

/** Only paths inside the admin board are accepted as `?next=` (no open redirect). */
function safeNext(raw: string | null): string {
  const path = (raw ?? "").replace(/^\/(en|vi)(?=\/|$)/, "");
  return /^\/admin(\/[A-Za-z0-9/_-]*)?(\?[^#]*)?$/.test(path) && !path.startsWith("/admin/login") ? path : "/admin";
}

export function AdminLoginForm() {
  const t = useTranslations("admin.login");
  const tErrors = useTranslations("errors");
  const router = useRouter();
  const next = safeNext(useSearchParams().get("next"));
  const [state, formAction, pending] = useActionState(adminSignIn, INITIAL);

  useEffect(() => {
    if (state.ok) router.replace(next);
  }, [state.ok, next, router]);

  return (
    <form action={formAction} className="space-y-4">
      {state.code && (
        <div role="alert" className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-sm text-destructive">
          <CircleAlert className="mt-0.5 size-4 shrink-0" strokeWidth={1.75} aria-hidden />
          <span>{tErrors.has(state.code) ? tErrors(state.code) : tErrors("UNKNOWN")}</span>
        </div>
      )}
      <div className="space-y-1.5">
        <Label htmlFor="admin-username">{t("username")}</Label>
        <Input id="admin-username" name="username" autoComplete="username" required autoFocus defaultValue={state.username} />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="admin-password">{t("password")}</Label>
        <Input id="admin-password" name="password" type="password" autoComplete="current-password" required />
      </div>
      <Button type="submit" className="h-11 w-full" disabled={pending}>
        {pending ? t("signingIn") : t("submit")}
      </Button>
    </form>
  );
}
