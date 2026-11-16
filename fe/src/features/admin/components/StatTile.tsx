import type { ReactNode } from "react";
import { Link } from "@/i18n/navigation";
import { cn } from "@/lib/utils";

type TStatTileProps = { label: string; value: string; hint?: ReactNode; href?: string; highlight?: boolean };

/** One number with its label; a hint line carries the breakdown. */
export function StatTile({ label, value, hint, href, highlight }: TStatTileProps) {
  const body = (
    <>
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-semibold tracking-tight tabular-nums">{value}</p>
      {hint && <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{hint}</p>}
    </>
  );
  const className = cn(
    "block rounded-xl border bg-card p-4",
    highlight ? "border-primary/40 bg-primary/5" : "border-border",
    href && "transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
  );
  return href ? (
    <Link href={href} className={className}>
      {body}
    </Link>
  ) : (
    <div className={className}>{body}</div>
  );
}

export function StatSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <h2 className="mb-3 text-sm font-semibold tracking-tight text-muted-foreground">{title}</h2>
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">{children}</div>
    </section>
  );
}
