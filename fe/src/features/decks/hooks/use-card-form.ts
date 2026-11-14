import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";
import { applyServerFieldErrors } from "@/lib/form-errors";
import { cardFormSchema } from "../schemas/deck-schemas";
import type { TCardFormValues } from "../schemas/deck-schemas";
import type { TCard } from "../types";
import { useCard, useCreateCard, useUpdateCard } from "./use-decks";

const EMPTY: TCardFormValues = {
  headword: "",
  meaning_vi: "",
  ipa: "",
  part_of_speech: "",
  example_en: "",
  example_vi: "",
  image_url: "",
};

function toFormValues(card: TCard): TCardFormValues {
  const c = card.content;
  return {
    headword: c.headword ?? "",
    meaning_vi: c.meaning_vi ?? "",
    ipa: c.ipa?.uk ?? c.ipa?.us ?? "",
    part_of_speech: c.part_of_speech ?? "",
    example_en: c.extra_examples?.[0]?.en ?? "",
    example_vi: c.extra_examples?.[0]?.vi ?? "",
    image_url: c.image?.url ?? "",
  };
}

type TUseCardFormArgs = {
  open: boolean;
  deckId?: string;
  cardId?: string | null;
  onDone: () => void;
};

/** Add (deckId) or edit (cardId) a card; handles the duplicate-word warning. */
export function useCardForm({ open, deckId, cardId, onDone }: TUseCardFormArgs) {
  const t = useTranslations();
  const { data: card, isPending: isLoadingCard } = useCard(open && cardId ? cardId : null);
  const create = useCreateCard(deckId ?? "");
  const update = useUpdateCard();
  const [duplicateOf, setDuplicateOf] = useState<TCardFormValues | null>(null);

  const form = useForm<TCardFormValues>({ resolver: zodResolver(cardFormSchema), defaultValues: EMPTY });

  // Load the card into the form when editing, clear it when adding.
  useEffect(() => {
    if (!open) {
      return;
    }
    form.reset(card && cardId ? toFormValues(card) : EMPTY);
  }, [open, card, cardId, form]);

  const save = async (values: TCardFormValues, allowDuplicate = false) => {
    try {
      if (cardId) {
        await update.mutateAsync({ id: cardId, fields: values });
        toast.success(t("cards.saved"));
      } else {
        await create.mutateAsync({ ...values, allow_duplicate: allowDuplicate });
        toast.success(t("cards.added"));
      }
      setDuplicateOf(null);
      onDone();
    } catch (error) {
      const code = error instanceof ApiError ? error.code : "UNKNOWN";
      if (code === "CARD_DUPLICATE_HEADWORD") {
        setDuplicateOf(values);
        return;
      }
      if (applyServerFieldErrors(error, form.setError, Object.keys(EMPTY))) {
        return;
      }
      toast.error(t.has(`errors.${code}`) ? t(`errors.${code}`) : t("errors.UNKNOWN"));
    }
  };

  return {
    form,
    onSubmit: form.handleSubmit((values) => save(values)),
    addAnyway: () => duplicateOf && save(duplicateOf, true),
    duplicate: !!duplicateOf,
    isEditing: !!cardId,
    isLoadingCard: !!cardId && isLoadingCard,
    isSubmitting: create.isPending || update.isPending,
  };
}
