"use client";

import { LanguageSwitcher } from "@/components/shared/LanguageSwitcher";
import { ThemeToggle } from "@/components/shared/ThemeToggle";
import { BrandMark } from "@/components/shared/BrandMark";
import { useUpdateProfile } from "@/features/profile/hooks/use-update-profile";
import { Link } from "@/i18n/navigation";
import { UserMenu } from "./UserMenu";

export function TopBar() {
  const updateProfile = useUpdateProfile();

  return (
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between gap-3 border-b border-border bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80 sm:px-6 lg:justify-end">
      <Link href="/dashboard" className="lg:hidden">
        <BrandMark />
      </Link>
      <div className="flex items-center gap-1">
        {/* Signed-in users also get the language saved to their profile. */}
        <LanguageSwitcher onChange={(ui_locale) => updateProfile.mutate({ ui_locale })} />
        <ThemeToggle />
        <div className="ml-2">
          <UserMenu />
        </div>
      </div>
    </header>
  );
}
