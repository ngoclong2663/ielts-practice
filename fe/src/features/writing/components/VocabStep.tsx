"use client";

import { ArrowRight } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import type { TAttempt } from "../types";
import { UsefulVocab } from "./UsefulVocab";

/** The step between planning and writing: learn the words, and the other ways to say them, before using them. */
export function VocabStep({ attempt, onNext }: { attempt: TAttempt; onNext: () => void }) {
  const t = useTranslations("writing");
  return (
    <section className="space-y-5">
      <header>
        <h2 className="text-lg font-semibold">{t("vocab.title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t("vocab.help")}</p>
      </header>
      <UsefulVocab attempt={attempt} />
      {!attempt.submission && (
        <div className="flex justify-end border-t border-border pt-4">
          <Button onClick={onNext}>
            {t("continue")}
            <ArrowRight className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
        </div>
      )}
    </section>
  );
}
