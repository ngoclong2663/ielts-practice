"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useAdvancedSettingsForm } from "../hooks/use-advanced-settings-form";
import type { TStudySettings } from "../types";
import { SettingsSection } from "./SettingsSection";

type AdvancedSettingsFormProps = {
  settings: TStudySettings;
};

export function AdvancedSettingsForm({ settings }: AdvancedSettingsFormProps) {
  const t = useTranslations();
  const { form, onSubmit, isSubmitting } = useAdvancedSettingsForm(settings);
  const errors = form.formState.errors;

  return (
    <SettingsSection title={t("settings.advancedTitle")} description={t("settings.advancedBody")}>
      <form onSubmit={onSubmit} noValidate>
        <FieldGroup>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
            <Field data-invalid={!!errors.learning_steps}>
              <FieldLabel htmlFor="learning_steps">{t("settings.learningSteps")}</FieldLabel>
              <Input id="learning_steps" className="font-mono" {...form.register("learning_steps")} />
              {errors.learning_steps ? (
                <FieldError>{t("errors.USER_INVALID_STEPS")}</FieldError>
              ) : (
                <FieldDescription>{t("settings.learningStepsHelp")}</FieldDescription>
              )}
            </Field>
            <Field data-invalid={!!errors.relearning_steps}>
              <FieldLabel htmlFor="relearning_steps">{t("settings.relearningSteps")}</FieldLabel>
              <Input id="relearning_steps" className="font-mono" {...form.register("relearning_steps")} />
              {errors.relearning_steps ? (
                <FieldError>{t("errors.USER_INVALID_STEPS")}</FieldError>
              ) : (
                <FieldDescription>{t("settings.relearningStepsHelp")}</FieldDescription>
              )}
            </Field>
          </div>
          <Field data-invalid={!!errors.max_interval_days} className="sm:max-w-xs">
            <FieldLabel htmlFor="max_interval_days">{t("settings.maxIntervalDays")}</FieldLabel>
            <Input
              id="max_interval_days"
              type="number"
              inputMode="numeric"
              className="font-mono"
              {...form.register("max_interval_days", { valueAsNumber: true })}
            />
            {errors.max_interval_days && <FieldError>{t("validation.range", { min: 1, max: 36500 })}</FieldError>}
          </Field>
          <div>
            <Button type="submit" disabled={isSubmitting || !form.formState.isDirty}>
              {isSubmitting ? t("common.saving") : t("common.save")}
            </Button>
          </div>
        </FieldGroup>
      </form>
    </SettingsSection>
  );
}
