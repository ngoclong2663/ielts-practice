import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";
import { studySettingsFormSchema } from "../schemas/settings-schema";
import type { TStudySettingsFormValues } from "../schemas/settings-schema";
import type { TStudySettings } from "../types";
import { useUpdateStudySettings } from "./use-update-profile";

function toFormValues(s: TStudySettings): TStudySettingsFormValues {
  return {
    new_cards_per_day: s.new_cards_per_day,
    max_reviews_per_day: s.max_reviews_per_day,
    desired_retention: Math.min(0.97, Math.max(0.8, s.desired_retention)),
    audio_autoplay: s.audio_autoplay,
    accent: s.accent,
    show_image_on_front: s.show_image_on_front,
  };
}

export function useStudySettingsForm(settings: TStudySettings) {
  const t = useTranslations();
  const mutation = useUpdateStudySettings();

  const form = useForm<TStudySettingsFormValues>({
    resolver: zodResolver(studySettingsFormSchema),
    defaultValues: toFormValues(settings),
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const saved = await mutation.mutateAsync({
        ...values,
        desired_retention: Math.round(values.desired_retention * 100) / 100,
      });
      form.reset(toFormValues(saved));
      toast.success(t("settings.saved"));
    } catch (error) {
      const code = error instanceof ApiError ? error.code : "UNKNOWN";
      toast.error(t.has(`errors.${code}`) ? t(`errors.${code}`) : t("errors.UNKNOWN"));
    }
  });

  return { form, onSubmit, isSubmitting: mutation.isPending };
}
