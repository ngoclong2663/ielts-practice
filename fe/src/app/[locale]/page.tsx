import { getTranslations } from "next-intl/server";
import { Container } from "@/components/layout/Container";
import { ClosingCta } from "@/features/landing/components/ClosingCta";
import { FeatureBento } from "@/features/landing/components/FeatureBento";
import { Hero } from "@/features/landing/components/Hero";
import { LandingHeader } from "@/features/landing/components/LandingHeader";

export default async function LandingPage() {
  const t = await getTranslations("landing");
  return (
    <>
      <LandingHeader />
      <main className="flex-1">
        <Hero />
        <FeatureBento />
        <ClosingCta />
      </main>
      <footer className="border-t border-border py-8">
        <Container className="text-sm text-muted-foreground">{t("footer")}</Container>
      </footer>
    </>
  );
}
