import { getTranslations } from "next-intl/server";
import { Container } from "@/components/layout/Container";
import { buttonVariants } from "@/components/ui/button";
import { Link } from "@/i18n/navigation";
import { HeroPreview } from "./HeroPreview";

export async function Hero() {
  const t = await getTranslations("landing");
  return (
    <section className="py-12 sm:py-16 lg:py-20">
      <Container className="grid grid-cols-1 items-center gap-12 lg:grid-cols-[1.1fr_1fr] lg:gap-16">
        <div className="max-w-xl">
          <h1 className="text-4xl font-semibold leading-[1.1] tracking-tight text-balance sm:text-5xl lg:text-[3.4rem]">
            {t("heroTitle")}
          </h1>
          <p className="mt-5 max-w-[52ch] text-base leading-relaxed text-muted-foreground sm:text-lg">
            {t("heroSubtitle")}
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Link href="/login" className={buttonVariants({ size: "lg", className: "active:translate-y-px" })}>
              {t("cta")}
            </Link>
            <a href="#features" className={buttonVariants({ variant: "ghost", size: "lg" })}>
              {t("secondaryCta")}
            </a>
          </div>
        </div>
        <div className="flex justify-center lg:justify-end">
          <HeroPreview />
        </div>
      </Container>
    </section>
  );
}
