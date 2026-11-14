import { BookOpenText } from "lucide-react";

export function BrandMark() {
  return (
    <span className="inline-flex items-center gap-2 font-semibold tracking-tight">
      <span className="grid size-8 place-items-center rounded-lg bg-primary text-primary-foreground">
        <BookOpenText className="size-4.5" strokeWidth={1.75} aria-hidden />
      </span>
      <span>IELTS Practice</span>
    </span>
  );
}
