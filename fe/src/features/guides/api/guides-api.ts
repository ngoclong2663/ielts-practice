import { apiClient } from "@/lib/api-client";
import type { TGuide, TGuideTopic } from "../types";

export const fetchGuides = (topic: TGuideTopic) => apiClient.get<TGuide[]>(`/guides?topic=${topic}`);
