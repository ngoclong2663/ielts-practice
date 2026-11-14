import type { ReactNode } from "react";

type SettingsSectionProps = {
  title: string;
  description: string;
  children: ReactNode;
};

// Two-column on desktop (heading left, form right), stacked on mobile.
export function SettingsSection({ title, description, children }: SettingsSectionProps) {
  return (
    <section className="grid grid-cols-1 gap-6 border-t border-border py-8 first:border-t-0 first:pt-0 lg:grid-cols-[minmax(0,16rem)_minmax(0,1fr)] lg:gap-12">
      <div className="space-y-1">
        <h2 className="text-base font-semibold tracking-tight">{title}</h2>
        <p className="text-sm text-muted-foreground leading-relaxed">{description}</p>
      </div>
      <div className="max-w-xl">{children}</div>
    </section>
  );
}
