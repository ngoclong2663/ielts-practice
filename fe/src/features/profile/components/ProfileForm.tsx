"use client";

import { useTranslations } from "next-intl";
import { Controller } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useProfileForm } from "../hooks/use-profile-form";
import { useTimezones } from "../hooks/use-timezones";
import type { TUserProfile } from "../types";
import { SettingsSection } from "./SettingsSection";

type ProfileFormProps = {
  user: TUserProfile;
};

const HOURS = Array.from({ length: 24 }, (_, h) => ({
  value: h,
  label: `${String(h).padStart(2, "0")}:00`,
}));

export function ProfileForm({ user }: ProfileFormProps) {
  const t = useTranslations();
  const { form, onSubmit, isSubmitting } = useProfileForm(user);
  const timezones = useTimezones(form.watch("timezone"));
  const errors = form.formState.errors;

  const localeItems = [
    { value: "vi", label: t("language.vi") },
    { value: "en", label: t("language.en") },
  ];

  return (
    <SettingsSection title={t("settings.profileTitle")} description={t("settings.profileBody")}>
      <form onSubmit={onSubmit} noValidate>
        <FieldGroup>
          <Field data-invalid={!!errors.display_name}>
            <FieldLabel htmlFor="display_name">{t("settings.displayName")}</FieldLabel>
            <Input id="display_name" autoComplete="name" {...form.register("display_name")} />
            {errors.display_name && (
              <FieldError>{t(`validation.${errors.display_name.message}`, { max: 120 })}</FieldError>
            )}
          </Field>

          <Field>
            <FieldLabel htmlFor="email">{t("settings.email")}</FieldLabel>
            <Input id="email" value={user.email} disabled readOnly />
            <FieldDescription>{t("settings.emailHelp")}</FieldDescription>
          </Field>

          <Field>
            <FieldLabel>{t("settings.uiLocale")}</FieldLabel>
            <Controller
              control={form.control}
              name="ui_locale"
              render={({ field }) => (
                <Select items={localeItems} value={field.value} onValueChange={(v) => v && field.onChange(v)}>
                  <SelectTrigger className="w-full sm:w-64" aria-label={t("settings.uiLocale")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {localeItems.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </Field>

          <Field data-invalid={!!errors.timezone}>
            <FieldLabel>{t("settings.timezone")}</FieldLabel>
            <Controller
              control={form.control}
              name="timezone"
              render={({ field }) => (
                <Select value={field.value} onValueChange={(v) => v && field.onChange(v)}>
                  <SelectTrigger className="w-full sm:w-80" aria-label={t("settings.timezone")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="max-h-80">
                    {timezones.map((tz) => (
                      <SelectItem key={tz} value={tz}>
                        {tz.replaceAll("_", " ")}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
            {errors.timezone ? (
              <FieldError>{t("errors.USER_INVALID_TIMEZONE")}</FieldError>
            ) : (
              <FieldDescription>{t("settings.timezoneHelp")}</FieldDescription>
            )}
          </Field>

          <Field>
            <FieldLabel>{t("settings.dayStartHour")}</FieldLabel>
            <Controller
              control={form.control}
              name="day_start_hour"
              render={({ field }) => (
                <Select items={HOURS} value={field.value} onValueChange={(v) => v !== null && field.onChange(v)}>
                  <SelectTrigger className="w-full sm:w-40 font-mono" aria-label={t("settings.dayStartHour")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="max-h-80">
                    {HOURS.map((h) => (
                      <SelectItem key={h.value} value={h.value} className="font-mono">
                        {h.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
            <FieldDescription>{t("settings.dayStartHourHelp")}</FieldDescription>
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
