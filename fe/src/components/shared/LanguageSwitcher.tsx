"use client";

import { Languages } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { Suspense, useTransition } from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { usePathname, useRouter } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";
import type { TLocale } from "@/i18n/routing";

type LanguageSwitcherProps = {
  /** Called after switching, e.g. to save the choice to the user profile. */
  onChange?: (locale: TLocale) => void;
};

export function LanguageSwitcher(props: LanguageSwitcherProps) {
  return (
    <Suspense fallback={<LanguageButton disabled />}>
      <LanguageMenu {...props} />
    </Suspense>
  );
}

function LanguageButton({ disabled }: { disabled?: boolean }) {
  const t = useTranslations("language");
  const locale = useLocale();
  return (
    <Button variant="ghost" size="sm" disabled={disabled} aria-label={t("label")}>
      <Languages className="size-4" strokeWidth={1.75} aria-hidden />
      <span className="uppercase">{locale}</span>
    </Button>
  );
}

function LanguageMenu({ onChange }: LanguageSwitcherProps) {
  const t = useTranslations("language");
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const [isPending, startTransition] = useTransition();

  const change = (next: string) => {
    if (next === locale) {
      return;
    }
    startTransition(() => {
      router.replace(pathname, { locale: next });
    });
    onChange?.(next as TLocale);
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button variant="ghost" size="sm" disabled={isPending} aria-label={t("label")}>
            <Languages className="size-4" strokeWidth={1.75} aria-hidden />
            <span className="uppercase">{locale}</span>
          </Button>
        }
      />
      <DropdownMenuContent align="end" className="w-40">
        <DropdownMenuRadioGroup value={locale} onValueChange={(v) => change(String(v))}>
          {routing.locales.map((l) => (
            <DropdownMenuRadioItem key={l} value={l}>
              {t(l)}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
