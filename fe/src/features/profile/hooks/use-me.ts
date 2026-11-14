import { useQuery } from "@tanstack/react-query";
import { fetchMe } from "../api/profile-api";
import { profileKeys } from "../query-keys";
import type { TMe } from "../types";

export function useMe() {
  return useQuery<TMe>({
    queryKey: profileKeys.me(),
    queryFn: fetchMe,
    staleTime: 5 * 60 * 1000,
  });
}
