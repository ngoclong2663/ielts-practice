"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useReportLexeme } from "../hooks/use-vocab";
import type { TReportReason } from "../types";

const REASONS: TReportReason[] = ["wrong_meaning", "bad_example", "wrong_image", "other"];

/** "Report wrong content" for a generated card (US-3.5). */
export function ReportContentDialog({ lexemeId, onOpenChange }: { lexemeId: string | null; onOpenChange: (open: boolean) => void }) {
  const t = useTranslations();
  const report = useReportLexeme();
  const [reason, setReason] = useState<TReportReason>("wrong_meaning");
  const [note, setNote] = useState("");

  const submit = () => {
    if (!lexemeId) return;
    report.mutate(
      { lexemeId, reason, note: note.trim() || undefined },
      {
        onSuccess: () => {
          toast.success(t("vocab.reportThanks"));
          setNote("");
          onOpenChange(false);
        },
        onError: (error) => toast.error(t(`errors.${error instanceof ApiError ? error.code : "UNKNOWN"}`)),
      },
    );
  };

  return (
    <Dialog open={!!lexemeId} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("vocab.reportTitle")}</DialogTitle>
          <DialogDescription>{t("vocab.reportBody")}</DialogDescription>
        </DialogHeader>
        <div role="radiogroup" aria-label={t("vocab.reportTitle")} className="grid gap-2">
          {REASONS.map((r) => (
            <button
              key={r}
              type="button"
              role="radio"
              aria-checked={reason === r}
              onClick={() => setReason(r)}
              className={cn(
                "rounded-xl border px-3 py-2 text-left text-sm transition-colors",
                reason === r ? "border-primary bg-primary/10" : "border-border hover:bg-muted",
              )}
            >
              {t(`vocab.reportReason.${r}`)}
            </button>
          ))}
        </div>
        <Textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={3}
          maxLength={500}
          placeholder={t("vocab.reportNote")}
          aria-label={t("vocab.reportNote")}
        />
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button onClick={submit} disabled={report.isPending}>
            {t("vocab.reportSend")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
