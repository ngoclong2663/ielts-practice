import { getTranslations } from "next-intl/server";
import { BrandMark } from "@/components/shared/BrandMark";
import { LanguageSwitcher } from "@/components/shared/LanguageSwitcher";
import { ThemeToggle } from "@/components/shared/ThemeToggle";
import { Container } from "@/components/layout/Container";
import { buttonVariants } from "@/components/ui/button";
import { Link } from "@/i18n/navigation";

export async function LandingHeader() {
  const t = await getTranslations("landing");
  return (
    <header className="border-b border-border/60">
      <Container className="flex h-16 items-center justify-between">
        <Link href="/">
          <BrandMark />
        </Link>
        <div className="flex items-center gap-1">
          <LanguageSwitcher />
          <ThemeToggle />
          <Link href="/login" className={buttonVariants({ variant: "outline", size: "sm", className: "ml-2" })}>
            {t("navSignIn")}
          </Link>
        </div>
      </Container>
    </header>
  );
}
