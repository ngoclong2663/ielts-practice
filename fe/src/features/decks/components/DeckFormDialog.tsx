"use client";

import { useTranslations } from "next-intl";
import { Controller } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { useMe } from "@/features/profile/hooks/use-me";
import { useDeckForm } from "../hooks/use-deck-form";
import type { TDeck } from "../types";

type DeckFormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  deck?: TDeck | null;
  onSaved?: (deck: TDeck) => void;
};

export function DeckFormDialog({ open, onOpenChange, deck = null, onSaved }: DeckFormDialogProps) {
  const t = useTranslations();
  const { data: me } = useMe();
  const defaults = {
    newPerDay: me?.study_settings.new_cards_per_day ?? 20,
    reviewsPerDay: me?.study_settings.max_reviews_per_day ?? 200,
  };
  const { form, onSubmit, isSubmitting } = useDeckForm(open, deck, defaults, (saved) => {
    onSaved?.(saved);
    onOpenChange(false);
  });
  const errors = form.formState.errors;
  const useDefaults = form.watch("use_default_limits");

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{deck ? t("decks.formEditTitle") : t("decks.formCreateTitle")}</DialogTitle>
        </DialogHeader>
        <form id="deck-form" onSubmit={onSubmit} noValidate>
          <FieldGroup className="gap-4">
            <Field data-invalid={!!errors.name}>
              <FieldLabel htmlFor="deck-name">{t("decks.name")}</FieldLabel>
              <Input id="deck-name" placeholder={t("decks.namePlaceholder")} autoFocus {...form.register("name")} />
              {errors.name && (
                <FieldError>
                  {errors.name.message === "taken" ? t("errors.DECK_NAME_TAKEN") : t(`validation.${errors.name.message}`, { max: 80 })}
                </FieldError>
              )}
            </Field>
            <Field data-invalid={!!errors.description}>
              <FieldLabel htmlFor="deck-description">
                {t("decks.description")}
                <span className="ml-1 font-normal text-muted-foreground">({t("cards.optional")})</span>
              </FieldLabel>
              <Textarea id="deck-description" rows={2} {...form.register("description")} />
            </Field>

            <Controller
              control={form.control}
              name="use_default_limits"
              render={({ field }) => (
                <Field orientation="horizontal">
                  <FieldContent>
                    <FieldLabel htmlFor="deck-defaults">
                      {t("decks.useDefaults", { new: defaults.newPerDay, reviews: defaults.reviewsPerDay })}
                    </FieldLabel>
                  </FieldContent>
                  <Switch id="deck-defaults" checked={field.value} onCheckedChange={field.onChange} />
                </Field>
              )}
            />
            {!useDefaults && (
              <div className="grid grid-cols-2 gap-4">
                <Field data-invalid={!!errors.new_per_day}>
                  <FieldLabel htmlFor="deck-new">{t("decks.newPerDay")}</FieldLabel>
                  <Input id="deck-new" type="number" inputMode="numeric" className="font-mono" {...form.register("new_per_day", { valueAsNumber: true })} />
                  {errors.new_per_day && <FieldError>{t("validation.range", { min: 0, max: 500 })}</FieldError>}
                </Field>
                <Field data-invalid={!!errors.reviews_per_day}>
                  <FieldLabel htmlFor="deck-reviews">{t("decks.reviewsPerDay")}</FieldLabel>
                  <Input id="deck-reviews" type="number" inputMode="numeric" className="font-mono" {...form.register("reviews_per_day", { valueAsNumber: true })} />
                  {errors.reviews_per_day && <FieldError>{t("validation.range", { min: 0, max: 9999 })}</FieldError>}
                </Field>
              </div>
            )}
          </FieldGroup>
        </form>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" form="deck-form" disabled={isSubmitting}>
            {isSubmitting ? t("common.saving") : deck ? t("common.save") : t("decks.create")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
