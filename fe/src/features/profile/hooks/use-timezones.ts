import { useMemo } from "react";

// IANA time zones supported by the browser, always including the current value.
export function useTimezones(current: string): string[] {
  return useMemo(() => {
    const list =
      typeof Intl.supportedValuesOf === "function" ? Intl.supportedValuesOf("timeZone") : [];
    return list.includes(current) || !current ? list : [current, ...list];
  }, [current]);
}
