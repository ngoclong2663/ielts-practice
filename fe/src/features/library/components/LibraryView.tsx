"use client";

import { ArrowRight, Library } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { PageHeader } from "@/components/shared/PageHeader";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useCollections } from "../hooks/use-library";

const PREVIEW_TOPICS = 6;

/** The library (Epic 8 §3.5): one card per collection that has words. */
export function LibraryView() {
  const t = useTranslations();
  const locale = useLocale() as "vi" | "en";
  const isClient = useIsClient();
  const query = useCollections();
  const collections = isClient ? query.data : undefined;

  return (
    <div className="mx-auto max-w-3xl">
      <PageHeader title={t("library.title")} description={t("library.subtitle")} />
      {!collections ? (
        <Skeleton className="h-44 w-full rounded-2xl" />
      ) : collections.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border p-6 text-center text-muted-foreground">{t("library.empty")}</p>
      ) : (
        <ul className="grid gap-4">
          {collections.map((c) => (
            <li key={c.slug}>
              <Link
                href={`/library/${c.slug}`}
                className="group block rounded-2xl border border-border bg-card p-5 transition-colors hover:border-primary/50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              >
                <div className="flex items-start gap-4">
                  <span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                    <Library className="size-5" strokeWidth={1.75} aria-hidden />
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="text-lg font-semibold tracking-tight">{c.title[locale]}</h2>
                      <Badge variant="secondary" className="uppercase">
                        {c.kind}
                      </Badge>
                    </div>
                    <p className="mt-1 text-sm text-muted-foreground">{c.description[locale]}</p>
                    <p className="mt-3 text-sm font-medium">{t("library.counts", { words: c.words, topics: c.topics.length })}</p>
                    <p className="mt-2 flex flex-wrap gap-1.5">
                      {c.topics.slice(0, PREVIEW_TOPICS).map((tp) => (
                        <span key={tp.id} className="rounded-full bg-muted px-2.5 py-0.5 text-xs text-muted-foreground">
                          {t(`vocab.topics.${tp.id}`)}
                        </span>
                      ))}
                      {c.topics.length > PREVIEW_TOPICS && (
                        <span className="px-1 text-xs text-muted-foreground">+{c.topics.length - PREVIEW_TOPICS}</span>
                      )}
                    </p>
                  </div>
                  <ArrowRight className="mt-3 size-5 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" strokeWidth={1.75} aria-hidden />
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
