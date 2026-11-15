import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { AttemptView } from "@/features/writing/components/AttemptView";

export default function WritingAttemptPage() {
  return (
    <Suspense fallback={<Skeleton className="mx-auto h-96 max-w-6xl rounded-2xl" />}>
      <AttemptView />
    </Suspense>
  );
}
