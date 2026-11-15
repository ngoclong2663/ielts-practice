import { BookOpenText, CircleAlert, CircleCheck, CircleX, ImageIcon, PenLine, Repeat } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import Image from "next/image";
import { getTranslations } from "next-intl/server";
import type { ReactNode } from "react";
import { sitePhotos } from "@/assets/images";
import { Container } from "@/components/layout/Container";
import { PhotoCredit } from "@/components/shared/PhotoCredit";
import { SpeakWordButton } from "./SpeakWordButton";

// Review gaps when a new card keeps being answered "Good" (default 97% retention).
const INTERVALS = ["10m", "1d", "2d", "4d", "7d", "13d", "22d"];

const WORD_FORMS = [
  { form: "decide", pos: "v", vi: "quyết định" },
  { form: "decision", pos: "n", vi: "sự quyết định" },
  { form: "decisive", pos: "adj", vi: "quyết đoán" },
];

// Writing samples (Vietnamese, as learners see them in the app).
const PLAN_QUESTION = {
  q: "Phần tổng quan (overview) nên viết gì?",
  options: ["Số liệu cụ thể của từng năm", "Các xu hướng chính: ngành nào tăng, ngành nào giảm", "Ý kiến cá nhân về biểu đồ"],
  answer: 1,
};

const SENTENCES = [
  { verdict: "correct", en: "The line graph shows how many people worked in four industries." },
  { verdict: "understandable", en: "Fishing workers decrease strongly to 2 million.", fix: "fell sharply" },
  { verdict: "wrong", en: "Tourism is the biggest in all years." },
] as const;

const VERDICT_ICON: Record<(typeof SENTENCES)[number]["verdict"], { icon: LucideIcon; className: string }> = {
  correct: { icon: CircleCheck, className: "text-emerald-600 dark:text-emerald-400" },
  understandable: { icon: CircleAlert, className: "text-amber-600 dark:text-amber-400" },
  wrong: { icon: CircleX, className: "text-destructive" },
};

const CRITERIA = [
  { label: "Task Achievement", band: 6 },
  { label: "Coherence & Cohesion", band: 6 },
  { label: "Lexical Resource", band: 5 },
  { label: "Grammar", band: 6 },
];

const LEVELS = [1, 2, 3, 4] as const;

const TASK_TYPES = ["line_graph", "bar_chart", "pie_chart", "table", "map", "process", "opinion", "discussion", "advantages_disadvantages", "two_part"] as const;

function GroupHeader({ icon: Icon, title, body, className }: { icon: LucideIcon; title: string; body: string; className: string }) {
  return (
    <div className={`flex items-start gap-3 ${className}`}>
      <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
        <Icon className="size-5" strokeWidth={1.75} aria-hidden />
      </span>
      <div>
        <h3 className="text-2xl font-semibold tracking-tight">{title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{body}</p>
      </div>
    </div>
  );
}

function Cell({ title, body, className, children }: { title: string; body: string; className?: string; children?: ReactNode }) {
  return (
    <article className={`rounded-2xl border border-border p-6 sm:p-8 ${className ?? "bg-card"}`}>
      <h4 className="text-xl font-semibold tracking-tight">{title}</h4>
      <p className="mt-2 max-w-[55ch] text-sm leading-relaxed text-muted-foreground">{body}</p>
      {children}
    </article>
  );
}

export async function FeatureBento() {
  const t = await getTranslations("landing");
  const tw = await getTranslations("writing");
  return (
    <section id="features" className="scroll-mt-8 py-16 sm:py-24">
      <Container>
        <h2 className="max-w-2xl text-3xl font-semibold tracking-tight text-balance sm:text-4xl">{t("featuresTitle")}</h2>

        {/* ---- Vocabulary ---- */}
        <GroupHeader icon={BookOpenText} title={t("vocabGroupTitle")} body={t("vocabGroupBody")} className="mt-10" />
        <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-3">
          <article className="rounded-2xl border border-border bg-accent/60 p-6 sm:p-8 lg:col-span-2">
            <Repeat className="size-5 text-primary" strokeWidth={1.75} aria-hidden />
            <h4 className="mt-4 text-xl font-semibold tracking-tight">{t("featureSrsTitle")}</h4>
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

          <Cell title={t("featureAudioTitle")} body={t("featureAudioBody")}>
            <div className="mt-6 space-y-2">
              <SpeakWordButton word="schedule" ipa="/ˈʃedjuːl/" accent="en-GB" label="UK" />
              <SpeakWordButton word="schedule" ipa="/ˈskedʒuːl/" accent="en-US" label="US" />
            </div>
          </Cell>

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
              <h4 className="mt-4 text-xl font-semibold tracking-tight">{t("featureImageTitle")}</h4>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{t("featureImageBody")}</p>
            </div>
          </article>

          <Cell title={t("featureCardsTitle")} body={t("featureCardsBody")} className="bg-secondary lg:col-span-2">
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
          </Cell>
        </div>

        {/* ---- Writing ---- */}
        <GroupHeader icon={PenLine} title={t("writingGroupTitle")} body={t("writingGroupBody")} className="mt-16" />
        <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-3">
          <Cell title={t("featurePlanTitle")} body={t("featurePlanBody")}>
            <div className="mt-6 rounded-xl bg-background p-4 text-sm" lang="vi">
              <p className="font-medium">{PLAN_QUESTION.q}</p>
              <ul className="mt-3 space-y-2">
                {PLAN_QUESTION.options.map((o, i) => (
                  <li
                    key={o}
                    className={`rounded-lg border px-3 py-2 ${i === PLAN_QUESTION.answer ? "border-emerald-500/60 bg-emerald-500/10" : "border-border"}`}
                  >
                    {o}
                  </li>
                ))}
              </ul>
            </div>
          </Cell>

          <Cell title={t("featureSentencesTitle")} body={t("featureSentencesBody")} className="bg-accent/60 lg:col-span-2">
            <ul className="mt-6 space-y-2">
              {SENTENCES.map((s) => {
                const v = VERDICT_ICON[s.verdict];
                const Icon = v.icon;
                return (
                  <li key={s.en} className="flex items-start gap-3 rounded-xl bg-background px-4 py-3 text-sm" lang="en">
                    <Icon className={`mt-0.5 size-4 shrink-0 ${v.className}`} strokeWidth={2} aria-hidden />
                    <span>
                      {s.en}
                      {"fix" in s && <span className="ml-2 font-medium text-emerald-700 dark:text-emerald-400">→ {s.fix}</span>}
                    </span>
                  </li>
                );
              })}
            </ul>
          </Cell>

          <Cell title={t("featureTasksTitle")} body={t("featureTasksBody")} className="bg-secondary lg:col-span-2">
            <ul className="mt-6 flex flex-wrap gap-2">
              {TASK_TYPES.map((type) => (
                <li key={type} className="rounded-full border border-border bg-background px-3 py-1 text-xs">
                  {tw(`types.${type}`)}
                </li>
              ))}
            </ul>
            <ol className="mt-6 grid grid-cols-2 gap-2 sm:grid-cols-4">
              {LEVELS.map((level) => (
                <li key={level} className="rounded-xl bg-background p-3">
                  <span className="font-mono text-xs text-muted-foreground">{level}</span>
                  <p className="mt-1 text-sm font-medium">{tw(`levels.${level}.name`)}</p>
                  <p className="mt-0.5 text-xs leading-snug text-muted-foreground">{tw(`levels.${level}.help`)}</p>
                </li>
              ))}
            </ol>
          </Cell>

          <Cell title={t("featureBandTitle")} body={t("featureBandBody")}>
            <div className="mt-6 rounded-xl bg-background p-4">
              <p className="text-xs text-muted-foreground">{t("sampleBand")}</p>
              <p className="text-3xl font-semibold tabular-nums">5.5–6.0</p>
              <ul className="mt-3 space-y-2">
                {CRITERIA.map((c) => (
                  <li key={c.label} className="text-xs">
                    <div className="flex justify-between">
                      <span>{c.label}</span>
                      <span className="tabular-nums">{c.band.toFixed(1)}</span>
                    </div>
                    <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-muted" aria-hidden>
                      <div className="h-full rounded-full bg-primary" style={{ width: `${(c.band / 9) * 100}%` }} />
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </Cell>
        </div>
      </Container>
    </section>
  );
}
