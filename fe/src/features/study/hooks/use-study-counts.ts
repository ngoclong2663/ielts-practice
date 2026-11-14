import { useQuery } from "@tanstack/react-query";
import { fetchCounts } from "../api/study-api";
import { studyKeys } from "../query-keys";

export function useStudyCounts() {
  return useQuery({ queryKey: studyKeys.counts(), queryFn: fetchCounts, staleTime: 30 * 1000 });
}
