"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useCardForm } from "../hooks/use-card-form";
import type { TCardFormValues } from "../schemas/deck-schemas";

type CardFormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Set to add a card to this deck. */
  deckId?: string;
  /** Set to edit an existing card. */
  cardId?: string | null;
  onSaved?: () => void;
};

const FIELDS: { name: keyof TCardFormValues; label: string; optional?: boolean; mono?: boolean; max: number }[] = [
  { name: "headword", label: "headword", max: 100 },
  { name: "meaning_vi", label: "meaningVi", max: 300 },
  { name: "ipa", label: "ipa", optional: true, mono: true, max: 100 },
  { name: "part_of_speech", label: "partOfSpeech", optional: true, max: 30 },
  { name: "example_en", label: "exampleEn", optional: true, max: 300 },
  { name: "example_vi", label: "exampleVi", optional: true, max: 300 },
  { name: "image_url", label: "imageUrl", optional: true, max: 1000 },
];

export function CardFormDialog({ open, onOpenChange, deckId, cardId, onSaved }: CardFormDialogProps) {
  const t = useTranslations();
  const { form, onSubmit, addAnyway, duplicate, isEditing, isLoadingCard, isSubmitting } = useCardForm({
    open,
    deckId,
    cardId,
    onDone: () => {
      onSaved?.();
      onOpenChange(false);
    },
  });
  const errors = form.formState.errors;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEditing ? t("cards.editCard") : t("cards.addCard")}</DialogTitle>
        </DialogHeader>
        {isLoadingCard ? (
          <div className="space-y-4">
            <Skeleton className="h-9 w-full" />
            <Skeleton className="h-9 w-full" />
          </div>
        ) : (
          <form id="card-form" onSubmit={onSubmit} noValidate>
            <FieldGroup className="gap-4">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                {FIELDS.map((f) => {
                  const error = errors[f.name];
                  return (
                    <Field
                      key={f.name}
                      data-invalid={!!error}
                      className={f.name === "headword" || f.name === "meaning_vi" || f.name.startsWith("example") || f.name === "image_url" ? "sm:col-span-2" : ""}
                    >
                      <FieldLabel htmlFor={`card-${f.name}`}>
                        {t(`cards.${f.label}`)}
                        {f.optional && <span className="ml-1 font-normal text-muted-foreground">({t("cards.optional")})</span>}
                      </FieldLabel>
                      <Input
                        id={`card-${f.name}`}
                        className={f.mono ? "font-mono" : undefined}
                        autoFocus={f.name === "headword"}
                        {...form.register(f.name)}
                      />
                      {error && (
                        <FieldError>
                          {t(`validation.${error.message}`, { max: f.max })}
                        </FieldError>
                      )}
                    </Field>
                  );
                })}
              </div>
              {duplicate && (
                <div role="alert" className="rounded-lg border border-srs-learn/40 bg-srs-learn/5 px-3 py-2.5 text-sm">
                  <p className="font-medium">{t("cards.duplicateTitle")}</p>
                  <p className="text-muted-foreground">{t("cards.duplicateBody")}</p>
                </div>
              )}
            </FieldGroup>
          </form>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          {duplicate ? (
            <Button onClick={addAnyway} disabled={isSubmitting}>
              {t("cards.addAnyway")}
            </Button>
          ) : (
            <Button type="submit" form="card-form" disabled={isSubmitting || isLoadingCard}>
              {isSubmitting ? t("common.saving") : t("cards.save")}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
