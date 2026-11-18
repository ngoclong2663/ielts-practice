export const writingKeys = {
  all: ["writing"] as const,
  status: () => [...writingKeys.all, "status"] as const,
  tasks: (taskNo: number) => [...writingKeys.all, "tasks", taskNo] as const,
  task: (id: string) => [...writingKeys.all, "task", id] as const,
  attempts: () => [...writingKeys.all, "attempts"] as const,
  attempt: (id: string) => [...writingKeys.all, "attempt", id] as const,
  vocab: (attemptId: string) => [...writingKeys.all, "vocab", attemptId] as const,
};
