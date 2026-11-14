"use client";

import { CircleAlert } from "lucide-react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";

export function LoginError() {
  const t = useTranslations("errors");
  const code = useSearchParams().get("error");
  if (!code) {
    return null;
  }
  const message = t.has(code) ? t(code) : t("UNKNOWN");
  return (
    <div role="alert" className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-sm text-destructive">
      <CircleAlert className="mt-0.5 size-4 shrink-0" strokeWidth={1.75} aria-hidden />
      <span>{message}</span>
    </div>
  );
}
