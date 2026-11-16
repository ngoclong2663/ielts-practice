import { apiClient } from "@/lib/api-client";
import type { TSendFeedbackRequest, TSendFeedbackResponse } from "../types";

export const sendFeedbackApi = (body: TSendFeedbackRequest) => apiClient.post<TSendFeedbackResponse>("/feedback", body);
