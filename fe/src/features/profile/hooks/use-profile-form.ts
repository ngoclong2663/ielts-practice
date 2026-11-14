import { zodResolver } from "@hookform/resolvers/zod";
import { useLocale, useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { usePathname, useRouter } from "@/i18n/navigation";
import { ApiError } from "@/lib/api-client";
import { profileFormSchema } from "../schemas/settings-schema";
import type { TProfileFormValues } from "../schemas/settings-schema";
import type { TUserProfile } from "../types";
import { useUpdateProfile } from "./use-update-profile";

export function useProfileForm(user: TUserProfile) {
  const t = useTranslations();
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const mutation = useUpdateProfile();

  const form = useForm<TProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      display_name: user.display_name,
      ui_locale: user.ui_locale,
      timezone: user.timezone,
      day_start_hour: user.day_start_hour,
    },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const me = await mutation.mutateAsync(values);
      form.reset({
        display_name: me.user.display_name,
        ui_locale: me.user.ui_locale,
        timezone: me.user.timezone,
        day_start_hour: me.user.day_start_hour,
      });
      toast.success(t("settings.saved"));
      if (me.user.ui_locale !== locale) {
        router.replace(pathname, { locale: me.user.ui_locale });
      }
    } catch (error) {
      const code = error instanceof ApiError ? error.code : "UNKNOWN";
      if (code === "USER_INVALID_TIMEZONE") {
        form.setError("timezone", { message: "timezone" });
        return;
      }
      toast.error(t.has(`errors.${code}`) ? t(`errors.${code}`) : t("errors.UNKNOWN"));
    }
  });

  return { form, onSubmit, isSubmitting: mutation.isPending };
}
