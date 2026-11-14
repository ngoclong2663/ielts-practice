import { ImageIcon, Repeat } from "lucide-react";
import Image from "next/image";
import { getTranslations } from "next-intl/server";
import { sitePhotos } from "@/assets/images";
import { Container } from "@/components/layout/Container";
import { PhotoCredit } from "@/components/shared/PhotoCredit";
import { SpeakWordButton } from "./SpeakWordButton";

// Sample of how review gaps grow when a card keeps being answered "Good" (illustrative).
const INTERVALS = ["10m", "2d", "8d", "24d", "2mo", "5mo"];

const WORD_FORMS = [
  { form: "decide", pos: "v", vi: "quyết định" },
  { form: "decision", pos: "n", vi: "sự quyết định" },
  { form: "decisive", pos: "adj", vi: "quyết đoán" },
];

export async function FeatureBento() {
  const t = await getTranslations("landing");
  return (
    <section id="features" className="scroll-mt-8 py-16 sm:py-24">
      <Container>
        <h2 className="max-w-2xl text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
          {t("featuresTitle")}
        </h2>

        <div className="mt-10 grid grid-cols-1 gap-4 lg:grid-cols-3">
          {/* Spaced repetition: wide cell */}
          <article className="rounded-2xl border border-border bg-accent/60 p-6 sm:p-8 lg:col-span-2">
            <Repeat className="size-5 text-primary" strokeWidth={1.75} aria-hidden />
            <h3 className="mt-4 text-xl font-semibold tracking-tight">{t("featureSrsTitle")}</h3>
            <p className="mt-2 max-w-[55ch] text-sm leading-relaxed text-muted-foreground">{t("featureSrsBody")}</p>
            <ol className="mt-8 flex flex-wrap items-center gap-2">
              {INTERVALS.map((interval, i) => (
                <li key={interval} className="flex items-center gap-2">
                  <span className="rounded-full border border-primary/25 bg-background px-3 py-1 font-mono text-xs tabular-nums">
                    {interval}
                  </span>
                  {i < INTERVALS.length - 1 && <span className="h-px w-4 bg-primary/30" aria-hidden />}
                </li>
              ))}
            </ol>
          </article>

          {/* Audio: interactive */}
          <article className="rounded-2xl border border-border bg-card p-6 sm:p-8">
            <h3 className="text-xl font-semibold tracking-tight">{t("featureAudioTitle")}</h3>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{t("featureAudioBody")}</p>
            <div className="mt-6 space-y-2">
              <SpeakWordButton word="schedule" ipa="/ˈʃedjuːl/" accent="en-GB" label="UK" />
              <SpeakWordButton word="schedule" ipa="/ˈskedʒuːl/" accent="en-US" label="US" />
            </div>
          </article>

          {/* Image: photo cell */}
          <article className="overflow-hidden rounded-2xl border border-border bg-card">
            <div className="relative aspect-[4/3]">
              <Image
                src={sitePhotos.featureLandscape.src}
                alt={sitePhotos.featureLandscape.alt}
                fill
                placeholder="blur"
                sizes="(max-width: 1024px) 100vw, 33vw"
                className="object-cover"
              />
            </div>
            <PhotoCredit photo={sitePhotos.featureLandscape} className="px-6 pt-2 text-muted-foreground sm:px-8" />
            <div className="p-6 pt-4 sm:p-8 sm:pt-4">
              <ImageIcon className="size-5 text-primary" strokeWidth={1.75} aria-hidden />
              <h3 className="mt-4 text-xl font-semibold tracking-tight">{t("featureImageTitle")}</h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{t("featureImageBody")}</p>
            </div>
          </article>

          {/* Generated cards: wide cell */}
          <article className="rounded-2xl border border-border bg-secondary p-6 sm:p-8 lg:col-span-2">
            <h3 className="text-xl font-semibold tracking-tight">{t("featureCardsTitle")}</h3>
            <p className="mt-2 max-w-[55ch] text-sm leading-relaxed text-muted-foreground">{t("featureCardsBody")}</p>
            <dl className="mt-8 grid grid-cols-1 gap-3 sm:grid-cols-3">
              {WORD_FORMS.map((w) => (
                <div key={w.form} className="rounded-xl bg-background p-4">
                  <dt className="font-medium">
                    {w.form} <span className="font-mono text-xs text-muted-foreground">{w.pos}</span>
                  </dt>
                  <dd className="mt-1 text-sm text-primary">{w.vi}</dd>
                </div>
              ))}
            </dl>
          </article>
        </div>
      </Container>
    </section>
  );
}
