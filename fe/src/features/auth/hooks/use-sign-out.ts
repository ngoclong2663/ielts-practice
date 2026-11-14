import { useQueryClient } from "@tanstack/react-query";
import { useLocale } from "next-intl";
import { useTransition } from "react";
import { signOut } from "@/lib/auth-actions";

export function useSignOut() {
  const queryClient = useQueryClient();
  const locale = useLocale();
  const [isPending, startTransition] = useTransition();

  const run = () =>
    startTransition(async () => {
      await signOut();
      queryClient.clear();
      // Full navigation so every client cache and preserved route state is dropped.
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination -- intentional hard reload
      window.location.assign(`/${locale}`);
    });

  return { signOut: run, isPending };
}
