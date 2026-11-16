import { useMutation } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import { sendFeedbackApi } from "../api/feedback-api";
import type { TSendFeedbackRequest, TSendFeedbackResponse } from "../types";

export function useSendFeedback() {
  return useMutation<TSendFeedbackResponse, ApiError, TSendFeedbackRequest>({ mutationFn: sendFeedbackApi });
}
