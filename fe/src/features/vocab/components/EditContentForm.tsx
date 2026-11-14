"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { type FieldError as TFieldError, useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api-client";
import { applyServerFieldErrors } from "@/lib/form-errors";
import type { TCardContent } from "@/types/card-content";
import { useUpdateItem } from "../hooks/use-vocab";
import { contentFormSchema, contentToForm, formToContent, type TContentFormValues } from "../schemas";
import type { TJobItem } from "../types";

type EditContentFormProps = {
  jobId: string;
  item: TJobItem;
  content: TCardContent;
  onSaved: () => void;
};

function Section({ title, onAdd, addLabel, children }: { title: string; onAdd: () => void; addLabel: string; children: React.ReactNode }) {
  return (
    <fieldset className="space-y-3 border-t border-border pt-4">
      <div className="flex items-center justify-between">
        <legend className="text-xs font-medium tracking-wide text-muted-foreground uppercase">{title}</legend>
        <Button type="button" variant="ghost" size="sm" onClick={onAdd}>
          <Plus className="size-4" strokeWidth={1.75} aria-hidden />
          {addLabel}
        </Button>
      </div>
      {children}
    </fieldset>
  );
}

function RemoveButton({ onClick, label }: { onClick: () => void; label: string }) {
  return (
    <Button type="button" variant="ghost" size="icon-sm" onClick={onClick} aria-label={label} className="shrink-0">
      <X className="size-4" strokeWidth={1.75} aria-hidden />
    </Button>
  );
}

/** Edits the user's copy of a generated card (the shared word in the library is never changed). */
export function EditContentForm({ jobId, item, content, onSaved }: EditContentFormProps) {
  const t = useTranslations();
  const update = useUpdateItem(jobId);
  const form = useForm<TContentFormValues>({
    resolver: zodResolver(contentFormSchema),
    defaultValues: contentToForm(content),
  });
  const forms = useFieldArray({ control: form.control, name: "word_forms" });
  const collocations = useFieldArray({ control: form.control, name: "collocations" });
  const examples = useFieldArray({ control: form.control, name: "extra_examples" });
  const errors = form.formState.errors;
  const msg = (e?: TFieldError) => (e?.message ? t(`validation.${e.message}`, { max: 300 }) : null);

  const onSubmit = form.handleSubmit((values) => {
    update.mutate(
      { itemId: item.id, content: formToContent(values, content) },
      {
        onSuccess: () => {
          toast.success(t("vocab.edited"));
          onSaved();
        },
        onError: (error) => {
          if (!applyServerFieldErrors(error, form.setError, ["headword", "meaning_vi"])) {
            toast.error(t(`errors.${error instanceof ApiError ? error.code : "UNKNOWN"}`));
          }
        },
      },
    );
  });

  return (
    <form onSubmit={onSubmit} noValidate className="space-y-4">
      <FieldGroup className="gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field data-invalid={!!errors.headword}>
            <FieldLabel htmlFor="edit-headword">{t("cards.headword")}</FieldLabel>
            <Input id="edit-headword" {...form.register("headword")} />
            {errors.headword && <FieldError>{msg(errors.headword)}</FieldError>}
          </Field>
          <Field>
            <FieldLabel htmlFor="edit-pos">{t("cards.partOfSpeech")}</FieldLabel>
            <Input id="edit-pos" {...form.register("part_of_speech")} />
          </Field>
          <Field>
            <FieldLabel htmlFor="edit-ipa-uk">{t("vocab.ipaUk")}</FieldLabel>
            <Input id="edit-ipa-uk" className="font-mono" {...form.register("ipa_uk")} />
          </Field>
          <Field>
            <FieldLabel htmlFor="edit-ipa-us">{t("vocab.ipaUs")}</FieldLabel>
            <Input id="edit-ipa-us" className="font-mono" {...form.register("ipa_us")} />
          </Field>
        </div>
        <Field data-invalid={!!errors.meaning_vi}>
          <FieldLabel htmlFor="edit-meaning">{t("cards.meaningVi")}</FieldLabel>
          <Input id="edit-meaning" {...form.register("meaning_vi")} />
          {errors.meaning_vi && <FieldError>{msg(errors.meaning_vi)}</FieldError>}
        </Field>
        <Field>
          <FieldLabel htmlFor="edit-definition">{t("vocab.definition")}</FieldLabel>
          <Textarea id="edit-definition" rows={2} {...form.register("definition_en")} />
        </Field>
      </FieldGroup>

      <Section
        title={t("study.wordForms")}
        addLabel={t("vocab.add")}
        onAdd={() => forms.append({ form: "", pos: "", ipa: "", meaning_vi: "", example_en: "", example_vi: "" })}
      >
        {forms.fields.map((f, i) => (
          <div key={f.id} className="space-y-2 rounded-xl bg-muted/50 p-3">
            <div className="flex gap-2">
              <Input aria-label={t("vocab.form")} placeholder={t("vocab.form")} {...form.register(`word_forms.${i}.form`)} />
              <Input aria-label={t("cards.partOfSpeech")} placeholder={t("cards.partOfSpeech")} className="w-32" {...form.register(`word_forms.${i}.pos`)} />
              <RemoveButton onClick={() => forms.remove(i)} label={t("vocab.removeRow")} />
            </div>
            <Input aria-label={t("cards.meaning")} placeholder={t("cards.meaning")} {...form.register(`word_forms.${i}.meaning_vi`)} />
            <Input aria-label={t("vocab.exampleEn")} placeholder={t("vocab.exampleEn")} {...form.register(`word_forms.${i}.example_en`)} />
            <Input aria-label={t("vocab.exampleVi")} placeholder={t("vocab.exampleVi")} {...form.register(`word_forms.${i}.example_vi`)} />
            {errors.word_forms?.[i]?.form && <p className="text-sm text-destructive">{msg(errors.word_forms[i]?.form)}</p>}
          </div>
        ))}
      </Section>

      <Section title={t("study.collocations")} addLabel={t("vocab.add")} onAdd={() => collocations.append({ phrase: "", meaning_vi: "" })}>
        {collocations.fields.map((f, i) => (
          <div key={f.id} className="flex gap-2">
            <Input aria-label={t("vocab.phrase")} placeholder={t("vocab.phrase")} {...form.register(`collocations.${i}.phrase`)} />
            <Input aria-label={t("cards.meaning")} placeholder={t("cards.meaning")} {...form.register(`collocations.${i}.meaning_vi`)} />
            <RemoveButton onClick={() => collocations.remove(i)} label={t("vocab.removeRow")} />
          </div>
        ))}
      </Section>

      <Section title={t("study.examples")} addLabel={t("vocab.add")} onAdd={() => examples.append({ en: "", vi: "" })}>
        {examples.fields.map((f, i) => (
          <div key={f.id} className="flex gap-2">
            <div className="flex-1 space-y-2">
              <Input aria-label={t("vocab.exampleEn")} placeholder={t("vocab.exampleEn")} {...form.register(`extra_examples.${i}.en`)} />
              <Input aria-label={t("vocab.exampleVi")} placeholder={t("vocab.exampleVi")} {...form.register(`extra_examples.${i}.vi`)} />
            </div>
            <RemoveButton onClick={() => examples.remove(i)} label={t("vocab.removeRow")} />
          </div>
        ))}
      </Section>

      <Field className="border-t border-border pt-4">
        <FieldLabel htmlFor="edit-synonyms">{t("study.synonyms")}</FieldLabel>
        <Input id="edit-synonyms" placeholder={t("vocab.synonymsHelp")} {...form.register("synonyms")} />
      </Field>

      <div className="sticky bottom-0 -mx-4 -mb-4 flex justify-end gap-2 rounded-b-xl border-t border-border bg-popover p-4">
        <Button type="button" variant="outline" onClick={onSaved}>
          {t("common.cancel")}
        </Button>
        <Button type="submit" disabled={update.isPending}>
          {update.isPending ? t("common.saving") : t("common.save")}
        </Button>
      </div>
    </form>
  );
}
