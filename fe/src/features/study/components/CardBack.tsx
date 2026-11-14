"use client";

import { Volume2 } from "lucide-react";
import { useTranslations } from "next-intl";
import type { TCardContent, TExample } from "@/types/card-content";

type CardBackProps = {
  content: TCardContent;
  onPlay: (text: string, url?: string) => void;
  /** Hide the image here when it is already shown on the front. */
  hideImage?: boolean;
};

function ExampleLine({ example, onPlay, label }: { example: TExample; onPlay: CardBackProps["onPlay"]; label: string }) {
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
      <p className="leading-relaxed">
        {example.en}
        {example.vi && <span className="block text-sm text-muted-foreground">{example.vi}</span>}
      </p>
    </div>
  );
}

export function CardImage({ content }: { content: TCardContent }) {
  const image = content.image;
  if (!image?.url) {
    return null;
  }
  return (
    <figure>
      <div className="relative aspect-[16/9] overflow-hidden rounded-xl bg-muted">
        {/* Card images come from many hosts (Unsplash, Pixabay, user URLs), so next/image is not used here. */}
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={image.url} alt={image.alt || content.headword} className="h-full w-full object-cover" loading="eager" />
      </div>
      {image.author_name && (
        <figcaption className="mt-1 text-[11px] text-muted-foreground">
          {image.author_url ? (
            <a href={image.author_url} target="_blank" rel="noopener noreferrer" className="underline underline-offset-2">
              {image.author_name}
            </a>
          ) : (
            image.author_name
          )}
          {image.provider && image.provider !== "manual" && <span className="capitalize"> / {image.provider}</span>}
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
        <section className="grid gap-4 text-sm sm:grid-cols-2">
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
      )}
    </div>
  );
}
