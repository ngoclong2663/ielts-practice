import type { Metadata } from "next";
import Image from "next/image";
import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import { sitePhotos } from "@/assets/images";
import { BrandMark } from "@/components/shared/BrandMark";
import { PhotoCredit } from "@/components/shared/PhotoCredit";
import { LanguageSwitcher } from "@/components/shared/LanguageSwitcher";
import { ThemeToggle } from "@/components/shared/ThemeToggle";
import { Skeleton } from "@/components/ui/skeleton";
import { GoogleSignInButton } from "@/features/auth/components/GoogleSignInButton";
import { LoginError } from "@/features/auth/components/LoginError";
import { Link } from "@/i18n/navigation";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("landing");
  return { title: t("navSignIn") };
}

export default async function LoginPage() {
  const t = await getTranslations("auth");
  return (
    <div className="grid min-h-[100dvh] grid-cols-1 lg:grid-cols-2">
      <aside className="relative hidden overflow-hidden bg-muted lg:block">
        <Image
          src={sitePhotos.loginDesk.src}
          alt={sitePhotos.loginDesk.alt}
          fill
          priority
          placeholder="blur"
          sizes="50vw"
          className="object-cover"
        />
        <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent p-10 text-white">
          <p className="max-w-sm text-2xl font-semibold tracking-tight">{t("sideTitle")}</p>
          <p className="mt-2 max-w-sm text-sm opacity-85">{t("sideBody")}</p>
          <PhotoCredit photo={sitePhotos.loginDesk} className="mt-6 opacity-70" />
        </div>
      </aside>

      <div className="flex flex-col px-4 py-6 sm:px-8">
        <div className="flex items-center justify-between">
          <Link href="/">
            <BrandMark />
          </Link>
          <div className="flex items-center gap-1">
            <LanguageSwitcher />
            <ThemeToggle />
          </div>
        </div>

        <main className="flex flex-1 items-center justify-center py-12">
          <div className="w-full max-w-sm">
            <h1 className="text-3xl font-semibold tracking-tight">{t("title")}</h1>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{t("subtitle")}</p>
            <div className="mt-8 space-y-4">
              <Suspense fallback={<Skeleton className="h-12 w-full rounded-lg" />}>
                <LoginError />
                <GoogleSignInButton />
              </Suspense>
            </div>
            <p className="mt-6 text-xs leading-relaxed text-muted-foreground">{t("privacy")}</p>
          </div>
        </main>
      </div>
    </div>
  );
}
