"use client";

import { Volume2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import type { TCardContent, TExample } from "@/types/card-content";

type CardBackProps = {
  content: TCardContent;
  onPlay: (text: string, url?: string) => void;
  /** Hide the image here when it is already shown on the front. */
  hideImage?: boolean;
};

function ExampleLine({ example, onPlay, label }: { example: TExample; onPlay: CardBackProps["onPlay"]; label: string }) {
  const t = useTranslations("study");
  const [showVi, setShowVi] = useState(false);
  return (
    <div className="flex items-start gap-2">
      <button
        type="button"
        onClick={() => onPlay(example.en)}
        aria-label={label}
        className="mt-0.5 shrink-0 rounded-md p-0.5 text-muted-foreground transition-colors hover:text-primary"
      >
        <Volume2 className="size-4" strokeWidth={1.75} aria-hidden />
      </button>
      <div className="leading-relaxed">
        <p>{example.en}</p>
        {/* The translation stays hidden until tapped: recalling the meaning first helps learning. */}
        {example.vi &&
          (showVi ? (
            <p className="text-sm text-muted-foreground">{example.vi}</p>
          ) : (
            <button
              type="button"
              onClick={() => setShowVi(true)}
              className="text-xs text-muted-foreground underline decoration-dotted underline-offset-2 hover:text-foreground"
            >
              {t("showTranslation")}
            </button>
          ))}
      </div>
    </div>
  );
}

function PhotoCredit({ image }: { image: NonNullable<TCardContent["image"]> }) {
  const t = useTranslations("study");
  const link = (href: string | undefined, text: string) =>
    href ? (
      <a href={href} target="_blank" rel="noopener noreferrer" className="underline underline-offset-2">
        {text}
      </a>
    ) : (
      text
    );
  const author = image.author_name ?? "";
  if (image.provider === "unsplash") {
    return <>{t.rich("photoByOn", { author: () => link(image.author_url, author), source: () => link(image.source_url, "Unsplash") })}</>;
  }
  if (image.provider === "pixabay") {
    return <>{t.rich("imageByFrom", { author: () => link(image.author_url, author), source: () => link(image.source_url, "Pixabay") })}</>;
  }
  return <>{link(image.author_url, author)}</>;
}

export function CardImage({ content }: { content: TCardContent }) {
  const image = content.image;
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  // A photo removed by its provider: hide it instead of showing a broken image.
  if (!image?.url || image.url === failedSrc) {
    return null;
  }
  return (
    <figure>
      <div className="relative aspect-[16/9] overflow-hidden rounded-xl bg-muted">
        {/* Card images are hotlinked from provider CDNs (Unsplash, Pixabay via our refresh endpoint) or user URLs. */}
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={image.url}
          alt={image.alt || content.headword}
          className="h-full w-full object-cover"
          loading="eager"
          onError={() => setFailedSrc(image.url)}
        />
      </div>
      {image.author_name && (
        <figcaption className="mt-1 text-[11px] text-muted-foreground">
          <PhotoCredit image={image} />
        </figcaption>
      )}
    </figure>
  );
}

// The answer side of a card. Shared with card previews (Epic 3).
export function CardBack({ content, onPlay, hideImage }: CardBackProps) {
  const t = useTranslations("study");
  return (
    <div className="space-y-6">
      {!hideImage && <CardImage content={content} />}

      <div>
        <p className="text-2xl font-semibold tracking-tight text-primary">{content.meaning_vi}</p>
        {content.definition_en && <p className="mt-1 text-muted-foreground">{content.definition_en}</p>}
      </div>

      {!!content.word_forms?.length && (
        <section className="space-y-3">
          <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("wordForms")}</h3>
          <ul className="space-y-4">
            {content.word_forms.map((wf) => (
              <li key={`${wf.form}-${wf.pos}`} className="space-y-1">
                <p>
                  <span className="font-semibold">{wf.form}</span>
                  {wf.pos && <span className="ml-1.5 text-sm italic text-muted-foreground">{wf.pos}</span>}
                  {wf.ipa && <span className="ml-1.5 font-mono text-sm text-muted-foreground">{wf.ipa}</span>}
                  {wf.meaning_vi && <span className="ml-2 text-primary">{wf.meaning_vi}</span>}
                </p>
                {wf.example && <ExampleLine example={wf.example} onPlay={onPlay} label={t("playExample")} />}
              </li>
            ))}
          </ul>
        </section>
      )}

      {!!content.collocations?.length && (
        <section className="space-y-3">
          <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("collocations")}</h3>
          <ul className="flex flex-wrap gap-2">
            {content.collocations.map((c) => (
              <li key={c.phrase} className="rounded-lg bg-secondary px-3 py-1.5 text-sm">
                <span className="font-medium">{c.phrase}</span>
                {c.meaning_vi && <span className="ml-1.5 text-muted-foreground">{c.meaning_vi}</span>}
              </li>
            ))}
          </ul>
        </section>
      )}

      {!!content.extra_examples?.length && (
        <section className="space-y-3">
          <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("examples")}</h3>
          <div className="space-y-3">
            {content.extra_examples.map((ex) => (
              <ExampleLine key={ex.en} example={ex} onPlay={onPlay} label={t("playExample")} />
            ))}
          </div>
        </section>
      )}

      {(!!content.synonyms?.length || !!content.other_senses?.length) && (
        <details className="group text-sm">
          <summary className="cursor-pointer list-none text-xs font-medium uppercase tracking-wide text-muted-foreground hover:text-foreground [&::-webkit-details-marker]:hidden">
            <span className="mr-1 inline-block transition-transform group-open:rotate-90" aria-hidden>
              ›
            </span>
            {t("more")}
          </summary>
        <section className="mt-3 grid gap-4 sm:grid-cols-2">
          {!!content.synonyms?.length && (
            <div>
              <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("synonyms")}</h3>
              <p className="mt-1">{content.synonyms.join(", ")}</p>
            </div>
          )}
          {!!content.other_senses?.length && (
            <div>
              <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("otherMeanings")}</h3>
              <ul className="mt-1 space-y-0.5">
                {content.other_senses.map((s) => (
                  <li key={s.meaning_vi}>
                    {s.pos && <span className="italic text-muted-foreground">{s.pos} </span>}
                    {s.meaning_vi}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </section>
        </details>
      )}
    </div>
  );
}
