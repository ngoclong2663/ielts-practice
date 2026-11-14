import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";
import { advancedSettingsFormSchema } from "../schemas/settings-schema";
import type { TAdvancedSettingsFormValues } from "../schemas/settings-schema";
import type { TStudySettings } from "../types";
import { useUpdateStudySettings } from "./use-update-profile";

const toValues = (s: TStudySettings): TAdvancedSettingsFormValues => ({
  learning_steps: s.learning_steps,
  relearning_steps: s.relearning_steps,
  max_interval_days: s.max_interval_days,
});

export function useAdvancedSettingsForm(settings: TStudySettings) {
  const t = useTranslations();
  const mutation = useUpdateStudySettings();
  const form = useForm<TAdvancedSettingsFormValues>({
    resolver: zodResolver(advancedSettingsFormSchema),
    defaultValues: toValues(settings),
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const saved = await mutation.mutateAsync(values);
      form.reset(toValues(saved));
      toast.success(t("settings.saved"));
    } catch (error) {
      const apiError = error instanceof ApiError ? error : null;
      if (apiError?.code === "USER_INVALID_STEPS") {
        for (const field of Object.keys(apiError.fields)) {
          form.setError(field as keyof TAdvancedSettingsFormValues, { message: "steps" });
        }
        return;
      }
      const code = apiError?.code ?? "UNKNOWN";
      toast.error(t.has(`errors.${code}`) ? t(`errors.${code}`) : t("errors.UNKNOWN"));
    }
  });

  return { form, onSubmit, isSubmitting: mutation.isPending };
}
