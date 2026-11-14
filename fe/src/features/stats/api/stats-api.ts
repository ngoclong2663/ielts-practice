import { apiClient } from "@/lib/api-client";
import type { TStatsOverview } from "../types";

export const fetchStatsOverview = () => apiClient.get<TStatsOverview>("/stats/overview");
