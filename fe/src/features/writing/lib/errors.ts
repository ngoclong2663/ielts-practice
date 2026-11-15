import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";

type TTranslate = (key: string, values?: Record<string, string | number>) => string;

export function toastApiError(t: TTranslate, error: unknown, id?: string) {
  const code = error instanceof ApiError ? error.code : "UNKNOWN";
  const params = error instanceof ApiError ? (error.params as Record<string, number>) : {};
  toast.error(t(`errors.${code}`, params), id ? { id } : undefined);
}
