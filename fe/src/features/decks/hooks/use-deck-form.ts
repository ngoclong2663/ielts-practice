import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";
import { deckFormSchema } from "../schemas/deck-schemas";
import type { TDeckFormValues } from "../schemas/deck-schemas";
import type { TDeck } from "../types";
import { useCreateDeck, useUpdateDeck } from "./use-decks";

type TDefaults = { newPerDay: number; reviewsPerDay: number };

function toValues(deck: TDeck | null, d: TDefaults): TDeckFormValues {
  return {
    name: deck?.name ?? "",
    description: deck?.description ?? "",
    use_default_limits: !deck || (deck.new_per_day === null && deck.reviews_per_day === null),
    new_per_day: deck?.new_per_day ?? d.newPerDay,
    reviews_per_day: deck?.reviews_per_day ?? d.reviewsPerDay,
  };
}

/** Create a deck (deck = null) or edit one, including optional per-deck limits. */
export function useDeckForm(open: boolean, deck: TDeck | null, defaults: TDefaults, onDone: (deck: TDeck) => void) {
  const t = useTranslations();
  const create = useCreateDeck();
  const update = useUpdateDeck(deck?.id ?? "");
  const form = useForm<TDeckFormValues>({ resolver: zodResolver(deckFormSchema), defaultValues: toValues(deck, defaults) });

  useEffect(() => {
    if (open) {
      form.reset(toValues(deck, defaults));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- reset only when the dialog opens
  }, [open, deck?.id]);

  const onSubmit = form.handleSubmit(async (v) => {
    try {
      let saved: TDeck;
      if (deck) {
        saved = await update.mutateAsync({
          name: v.name,
          description: v.description,
          ...(v.use_default_limits
            ? { use_default_limits: true }
            : { new_per_day: v.new_per_day, reviews_per_day: v.reviews_per_day }),
        });
        toast.success(t("decks.saved"));
      } else {
        saved = await create.mutateAsync({ name: v.name, description: v.description });
        if (!v.use_default_limits) {
          // Limits are set with a follow-up update (create only takes name/description).
          const { updateDeckApi } = await import("../api/decks-api");
          saved = await updateDeckApi(saved.id, { new_per_day: v.new_per_day, reviews_per_day: v.reviews_per_day });
        }
        toast.success(t("decks.created"));
      }
      onDone(saved);
    } catch (error) {
      const code = error instanceof ApiError ? error.code : "UNKNOWN";
      if (code === "DECK_NAME_TAKEN") {
        form.setError("name", { message: "taken" });
        return;
      }
      toast.error(t.has(`errors.${code}`) ? t(`errors.${code}`) : t("errors.UNKNOWN"));
    }
  });

  return { form, onSubmit, isSubmitting: create.isPending || update.isPending };
}
