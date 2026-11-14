"use client";

import { CircleCheck } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { buttonVariants } from "@/components/ui/button";
import { Link } from "@/i18n/navigation";

type SessionDoneProps = {
  answered: number;
  again: number;
  durationMs: number;
  nextDueAt: string | null;
  deckId: string;
};

export function SessionDone({ answered, again, durationMs, nextDueAt, deckId }: SessionDoneProps) {
  const t = useTranslations("study");
  const format = useFormatter();
  const accuracy = answered > 0 ? Math.round(((answered - again) / answered) * 100) : null;

  return (
    <div className="flex flex-col items-start gap-6 py-10">
      <span className="grid size-12 place-items-center rounded-2xl bg-accent text-accent-foreground">
        <CircleCheck className="size-6" strokeWidth={1.75} aria-hidden />
      </span>
      <div>
        <h1 className="text-3xl font-semibold tracking-tight">{answered > 0 ? t("doneTitle") : t("emptyDeck")}</h1>
        <p className="mt-2 text-muted-foreground">
          {nextDueAt
            ? t("comeBackAt", { time: format.dateTime(new Date(nextDueAt), { hour: "2-digit", minute: "2-digit" }) })
            : answered > 0 && t("doneBody")}
        </p>
      </div>
      {answered > 0 && (
        <dl className="grid w-full max-w-md grid-cols-3 gap-3">
          {[
            { label: t("summaryCardsLabel"), value: String(answered) },
            { label: t("summaryMinutesLabel"), value: durationMs < 60_000 ? "<1" : String(Math.round(durationMs / 60_000)) },
            { label: t("summaryRememberedLabel"), value: `${accuracy ?? 0}%` },
          ].map((item) => (
            <div key={item.label} className="flex flex-col-reverse rounded-xl bg-secondary p-4">
              <dt className="mt-1 text-xs text-muted-foreground">{item.label}</dt>
              <dd className="text-2xl font-semibold tracking-tight">{item.value}</dd>
            </div>
          ))}
        </dl>
      )}
      <div className="flex flex-wrap gap-3">
        <Link href="/decks" className={buttonVariants({ variant: answered > 0 ? "default" : "outline" })}>
          {t("backToDecks")}
        </Link>
        {answered === 0 && deckId && (
          <Link href={`/decks/${deckId}`} className={buttonVariants()}>
            {t("addCards")}
          </Link>
        )}
      </div>
    </div>
  );
}
