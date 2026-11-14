import { useSyncExternalStore } from "react";

const subscribe = () => () => {};

/**
 * False on the server and during hydration, true afterwards. Components that show React Query
 * data render their loading state until then, so the first client render matches the server
 * HTML even when another component already filled the cache (avoids hydration mismatches).
 */
export function useIsClient(): boolean {
  return useSyncExternalStore(
    subscribe,
    () => true,
    () => false,
  );
}
