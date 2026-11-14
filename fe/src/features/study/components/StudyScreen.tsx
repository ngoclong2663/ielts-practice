"use client";

import { EllipsisVertical, Pencil, PauseCircle, Undo2, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { SrsCounts } from "@/components/srs/SrsCounts";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { CardFormDialog } from "@/features/decks/components/CardFormDialog";
import { useBulkCards } from "@/features/decks/hooks/use-decks";
import { useMe } from "@/features/profile/hooks/use-me";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useCardAudio } from "../hooks/use-card-audio";
import { useStudySession } from "../hooks/use-study-session";
import { CardBack } from "./CardBack";
import { CardFront } from "./CardFront";
import { RatingBar } from "./RatingBar";
import { SessionDone } from "./SessionDone";

function activeGroup(state: string): "new" | "learning" | "review" {
  return state === "new" ? "new" : state === "review" ? "review" : "learning";
}

export function StudyScreen() {
  const t = useTranslations("study");
  const tc = useTranslations("common");
  const deckId = useSearchParams().get("deck") ?? "";
  const isClient = useIsClient();
  const liveSession = useStudySession(deckId);
  // Until hydration finishes, render the loading state the server rendered.
  const session = isClient ? liveSession : { ...liveSession, queue: null, card: null, isLoading: true, isError: false };
  const { data: me } = useMe();
  const settings = me?.study_settings;
  const accent = settings?.accent ?? "uk";
  const { play, playHeadword } = useCardAudio(accent);
  const bulk = useBulkCards();
  const [editingCardId, setEditingCardId] = useState<string | null>(null);

  const { card, revealed } = session;

  // Toasts from the previous page would cover the study header.
  useEffect(() => {
    toast.dismiss();
  }, []);

  // Auto-play the word when a card appears and again when the answer is revealed.
  useEffect(() => {
    if (card && settings?.audio_autoplay) {
      playHeadword(card.content);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- play once per card / reveal
  }, [card?.id, card?.reps, revealed, settings?.audio_autoplay]);

  const exitHref = deckId ? `/decks/${deckId}` : "/decks";

  const suspendCurrent = () => {
    if (!card) return;
    bulk.mutate(
      { card_ids: [card.id], action: "suspend" },
      {
        onSuccess: () => {
          toast(t("suspended"));
          void session.reload();
        },
      },
    );
  };

  return (
    <div className="mx-auto flex min-h-[100dvh] w-full max-w-2xl flex-col px-4 sm:px-6">
      <header className="flex h-14 items-center justify-between gap-3">
        <Link href={exitHref} aria-label={t("exit")} className={buttonVariants({ variant: "ghost", size: "icon-sm" })}>
          <X className="size-5" strokeWidth={1.75} aria-hidden />
        </Link>
        <div className="min-w-0 flex-1 text-center">
          {card && <p className="truncate text-sm text-muted-foreground">{card.deck_name}</p>}
        </div>
        <div className="flex items-center gap-1">
          {session.queue && card && (
            <SrsCounts counts={session.queue.counts} active={activeGroup(card.state)} className="mr-2" />
          )}
          <Button variant="ghost" size="icon-sm" onClick={session.undo} disabled={!session.canUndo || session.busy} aria-label={t("undo")}>
            <Undo2 className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
          {card && (
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="ghost" size="icon-sm" aria-label={t("cardMenu")} />}>
                <EllipsisVertical className="size-4" strokeWidth={1.75} aria-hidden />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuItem onClick={() => setEditingCardId(card.id)}>
                  <Pencil className="size-4" strokeWidth={1.75} aria-hidden /> {t("editCard")}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={suspendCurrent}>
                  <PauseCircle className="size-4" strokeWidth={1.75} aria-hidden /> {t("suspendCard")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </header>

      <main className="flex flex-1 flex-col">
        {session.isLoading && (
          <div className="space-y-4 pt-10" aria-busy>
            <Skeleton className="h-12 w-2/3" />
            <Skeleton className="h-5 w-1/3" />
          </div>
        )}

        {session.isError && (
          <div className="pt-10">
            <p className="text-muted-foreground">{t("loadError")}</p>
            <Button variant="outline" className="mt-4" onClick={() => session.refetch()}>
              {tc("retry")}
            </Button>
          </div>
        )}

        {session.queue && !card && (
          <SessionDone
            answered={session.summary.answered}
            again={session.summary.again}
            durationMs={session.summary.durationMs}
            nextDueAt={session.queue.next_due_at}
            deckId={deckId}
          />
        )}

        {card && (
          <>
            <article key={`${card.id}-${card.reps}`} className="flex-1 py-6">
              <CardFront
                content={card.content}
                accent={accent}
                showImage={!!settings?.show_image_on_front}
                onPlay={() => playHeadword(card.content)}
              />
              {card.is_leech && <p className="mt-4 rounded-lg bg-secondary px-3 py-2 text-sm text-muted-foreground">{t("leechHint")}</p>}
              {revealed && (
                <div className="mt-8 border-t border-border pt-8 duration-300 animate-in fade-in-0 slide-in-from-bottom-2">
                  <CardBack content={card.content} onPlay={play} hideImage={!!settings?.show_image_on_front} />
                </div>
              )}
            </article>

            <div className="sticky bottom-0 -mx-4 bg-background/95 px-4 pt-3 pb-[max(1rem,env(safe-area-inset-bottom))] backdrop-blur sm:-mx-6 sm:px-6">
              <RatingBar card={card} revealed={revealed} busy={session.busy} onReveal={session.reveal} onAnswer={session.answer} />
              <p className="mt-2 hidden text-center text-xs text-muted-foreground sm:block">{t("shortcuts")}</p>
            </div>
          </>
        )}
      </main>

      <CardFormDialog
        open={!!editingCardId}
        onOpenChange={(open) => !open && setEditingCardId(null)}
        cardId={editingCardId}
        onSaved={() => void session.reload()}
      />
    </div>
  );
}
