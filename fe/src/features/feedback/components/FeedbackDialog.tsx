"use client";

import { useLocale, useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useSendFeedback } from "../hooks/use-send-feedback";
import { FEEDBACK_CATEGORIES, FEEDBACK_MAX_LENGTH, FEEDBACK_MIN_LENGTH } from "../types";
import type { TFeedbackCategory } from "../types";

type TFeedbackDialogProps = { open: boolean; onOpenChange: (open: boolean) => void };

/** "Send feedback" from the account menu (Epic 6, US-6.1). The current page is attached. */
export function FeedbackDialog({ open, onOpenChange }: TFeedbackDialogProps) {
  const t = useTranslations();
  const locale = useLocale();
  const send = useSendFeedback();
  const [category, setCategory] = useState<TFeedbackCategory>("bug");
  const [message, setMessage] = useState("");
  const length = message.trim().length;
  const tooShort = length < FEEDBACK_MIN_LENGTH;

  const submit = () => {
    if (tooShort) return;
    send.mutate(
      {
        category,
        message: message.trim(),
        page: `${window.location.pathname}${window.location.search}`.slice(0, 300),
        ui_locale: locale === "vi" ? "vi" : "en",
      },
      {
        onSuccess: () => {
          toast.success(t("feedback.thanks"));
          setMessage("");
          setCategory("bug");
          onOpenChange(false);
        },
        onError: (error) => toast.error(t(`errors.${error instanceof ApiError ? error.code : "UNKNOWN"}`)),
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("feedback.title")}</DialogTitle>
          <DialogDescription>{t("feedback.body")}</DialogDescription>
        </DialogHeader>
        <div role="radiogroup" aria-label={t("feedback.categoryLabel")} className="grid grid-cols-2 gap-2">
          {FEEDBACK_CATEGORIES.map((c) => (
            <button
              key={c}
              type="button"
              role="radio"
              aria-checked={category === c}
              onClick={() => setCategory(c)}
              className={cn(
                "rounded-xl border px-3 py-2 text-left text-sm transition-colors",
                category === c ? "border-primary bg-primary/10" : "border-border hover:bg-muted",
              )}
            >
              {t(`feedback.category.${c}`)}
            </button>
          ))}
        </div>
        <div className="space-y-1.5">
          <Textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            rows={5}
            maxLength={FEEDBACK_MAX_LENGTH}
            placeholder={t(`feedback.placeholder.${category}`)}
            aria-label={t("feedback.messageLabel")}
          />
          <p className="flex justify-between gap-3 text-xs text-muted-foreground">
            <span>
              {length > 0 && tooShort ? t("feedback.tooShort", { min: FEEDBACK_MIN_LENGTH }) : t("feedback.pageNote")}
            </span>
            <span className="shrink-0 font-mono tabular-nums">
              {length}/{FEEDBACK_MAX_LENGTH}
            </span>
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button onClick={submit} disabled={tooShort || send.isPending}>
            {send.isPending ? t("feedback.sending") : t("feedback.send")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
