import type { LucideIcon } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";

type ComingSoonProps = {
  icon: LucideIcon;
  title: string;
  body: string;
};

export async function ComingSoon({ icon: Icon, title, body }: ComingSoonProps) {
  const t = await getTranslations("common");
  return (
    <div className="flex max-w-2xl flex-col items-start gap-4 rounded-2xl border border-dashed border-border p-8 sm:p-10">
      <span className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
        <Icon className="size-5" strokeWidth={1.75} aria-hidden />
      </span>
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
          <Badge variant="secondary">{t("comingSoon")}</Badge>
        </div>
        <p className="mt-2 max-w-[55ch] text-sm leading-relaxed text-muted-foreground">{body}</p>
      </div>
    </div>
  );
}
