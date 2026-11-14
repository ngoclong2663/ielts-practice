import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { DeckDetailView } from "@/features/decks/components/DeckDetailView";

export default function DeckPage() {
  return (
    <Suspense fallback={<Skeleton className="mx-auto h-40 max-w-5xl rounded-2xl" />}>
      <DeckDetailView />
    </Suspense>
  );
}
