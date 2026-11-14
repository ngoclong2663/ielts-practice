import { getTranslations } from "next-intl/server";
import { Container } from "@/components/layout/Container";
import { buttonVariants } from "@/components/ui/button";
import { Link } from "@/i18n/navigation";

export async function ClosingCta() {
  const t = await getTranslations("landing");
  return (
    <section className="pb-20 sm:pb-28">
      <Container size="md">
        <div className="flex flex-col items-start gap-6 rounded-2xl bg-primary px-6 py-10 text-primary-foreground sm:flex-row sm:items-center sm:justify-between sm:px-10">
          <div>
            <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t("closingTitle")}</h2>
            <p className="mt-2 max-w-[48ch] text-sm leading-relaxed opacity-85">{t("closingBody")}</p>
          </div>
          <Link
            href="/login"
            className={buttonVariants({ variant: "secondary", size: "lg", className: "shrink-0 active:translate-y-px" })}
          >
            {t("cta")}
          </Link>
        </div>
      </Container>
    </section>
  );
}
