import { useQuery } from "@tanstack/react-query";
import { studyKeys } from "@/features/study/query-keys";
import { fetchStatsOverview } from "../api/stats-api";

export function useStatsOverview() {
  return useQuery({ queryKey: studyKeys.stats(), queryFn: fetchStatsOverview, staleTime: 30 * 1000 });
}
