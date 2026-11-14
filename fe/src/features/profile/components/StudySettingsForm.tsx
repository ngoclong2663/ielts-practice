"use client";

import { useTranslations } from "next-intl";
import { Controller } from "react-hook-form";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { useStudySettingsForm } from "../hooks/use-study-settings-form";
import type { TStudySettings } from "../types";
import { SettingsSection } from "./SettingsSection";

type StudySettingsFormProps = {
  settings: TStudySettings;
};

export function StudySettingsForm({ settings }: StudySettingsFormProps) {
  const t = useTranslations();
  const { form, onSubmit, isSubmitting } = useStudySettingsForm(settings);
  const errors = form.formState.errors;
  const retention = form.watch("desired_retention");

  const accentItems = [
    { value: "uk", label: t("settings.accentUk") },
    { value: "us", label: t("settings.accentUs") },
  ];

  return (
    <SettingsSection title={t("settings.studyTitle")} description={t("settings.studyBody")}>
      <form onSubmit={onSubmit} noValidate>
        <FieldGroup>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
            <Field data-invalid={!!errors.new_cards_per_day}>
              <FieldLabel htmlFor="new_cards_per_day">{t("settings.newCardsPerDay")}</FieldLabel>
              <Input
                id="new_cards_per_day"
                type="number"
                inputMode="numeric"
                min={0}
                max={500}
                className="font-mono"
                {...form.register("new_cards_per_day", { valueAsNumber: true })}
              />
              {errors.new_cards_per_day && (
                <FieldError>{t("validation.range", { min: 0, max: 500 })}</FieldError>
              )}
            </Field>
            <Field data-invalid={!!errors.max_reviews_per_day}>
              <FieldLabel htmlFor="max_reviews_per_day">{t("settings.maxReviewsPerDay")}</FieldLabel>
              <Input
                id="max_reviews_per_day"
                type="number"
                inputMode="numeric"
                min={0}
                max={9999}
                className="font-mono"
                {...form.register("max_reviews_per_day", { valueAsNumber: true })}
              />
              {errors.max_reviews_per_day && (
                <FieldError>{t("validation.range", { min: 0, max: 9999 })}</FieldError>
              )}
            </Field>
          </div>
          <FieldDescription className="-mt-2">{t("settings.newCardsPerDayHelp")}</FieldDescription>

          <Field>
            <div className="flex items-baseline justify-between gap-4">
              <FieldLabel>{t("settings.desiredRetention")}</FieldLabel>
              <span className="font-mono text-sm tabular-nums">{Math.round(retention * 100)}%</span>
            </div>
            <Controller
              control={form.control}
              name="desired_retention"
              render={({ field }) => (
                <Slider
                  min={0.8}
                  max={0.97}
                  step={0.01}
                  value={field.value}
                  onValueChange={(v) => field.onChange(Array.isArray(v) ? v[0] : v)}
                  aria-label={t("settings.desiredRetention")}
                />
              )}
            />
            <FieldDescription>{t("settings.desiredRetentionHelp")}</FieldDescription>
          </Field>

          <Field>
            <FieldLabel>{t("settings.accent")}</FieldLabel>
            <Controller
              control={form.control}
              name="accent"
              render={({ field }) => (
                <Select items={accentItems} value={field.value} onValueChange={(v) => v && field.onChange(v)}>
                  <SelectTrigger className="w-full sm:w-64" aria-label={t("settings.accent")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {accentItems.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </Field>

          <Controller
            control={form.control}
            name="audio_autoplay"
            render={({ field }) => (
              <Field orientation="horizontal">
                <FieldContent>
                  <FieldLabel htmlFor="audio_autoplay">{t("settings.audioAutoplay")}</FieldLabel>
                </FieldContent>
                <Switch id="audio_autoplay" checked={field.value} onCheckedChange={field.onChange} />
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="show_image_on_front"
            render={({ field }) => (
              <Field orientation="horizontal">
                <FieldContent>
                  <FieldLabel htmlFor="show_image_on_front">{t("settings.showImageOnFront")}</FieldLabel>
                </FieldContent>
                <Switch id="show_image_on_front" checked={field.value} onCheckedChange={field.onChange} />
              </Field>
            )}
          />

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
