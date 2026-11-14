import type { FieldValues, Path, UseFormSetError } from "react-hook-form";
import { ApiError } from "./api-client";

// Backend field codes (VALIDATION_<TAG>) -> keys under `validation.*` in the message catalogues.
const FIELD_CODE_TO_MESSAGE: Record<string, string> = {
  VALIDATION_REQUIRED: "required",
  VALIDATION_MIN: "required",
  VALIDATION_MAX: "tooLong",
  VALIDATION_URL: "url",
};

/** Puts server-side field errors on the matching form inputs. Returns true when any were applied. */
export function applyServerFieldErrors<T extends FieldValues>(
  error: unknown,
  setError: UseFormSetError<T>,
  fields: readonly string[],
): boolean {
  if (!(error instanceof ApiError) || error.code !== "VALIDATION_FAILED") {
    return false;
  }
  let applied = false;
  for (const [field, code] of Object.entries(error.fields)) {
    if (fields.includes(field)) {
      setError(field as Path<T>, { message: FIELD_CODE_TO_MESSAGE[code] ?? "invalid" });
      applied = true;
    }
  }
  return applied;
}
