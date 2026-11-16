"use client";

import { BellOff, CheckCheck, Eye, RotateCcw, StickyNote } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useUpdateFeedback } from "../hooks/use-admin";
import type { TAdminFeedback, TFeedbackStatus } from "../types";

const NOTE_MAX = 1000;

function browserOf(ua: string): string {
  const os = /iPhone|iPad/.test(ua) ? "iOS" : /Android/.test(ua) ? "Android" : /Mac OS X/.test(ua) ? "macOS" : /Windows/.test(ua) ? "Windows" : /Linux/.test(ua) ? "Linux" : "";
  const browser = /Edg\//.test(ua) ? "Edge" : /Chrome\//.test(ua) ? "Chrome" : /Firefox\//.test(ua) ? "Firefox" : /Safari\//.test(ua) ? "Safari" : "";
  return [browser, os].filter(Boolean).join(" · ") || ua.slice(0, 40);
}

export function FeedbackItem({ item, highlighted }: { item: TAdminFeedback; highlighted: boolean }) {
  const t = useTranslations();
  const format = useFormatter();
  const update = useUpdateFeedback();
  const [note, setNote] = useState(item.admin_note);
  const [editingNote, setEditingNote] = useState(false);

  const setStatus = (status: TFeedbackStatus) =>
    update.mutate({ id: item.id, status }, { onError: () => toast.error(t("admin.feedback.updateFailed")) });

  const saveNote = () =>
    update.mutate(
      { id: item.id, admin_note: note.trim() },
      {
        onSuccess: () => {
          setEditingNote(false);
          toast.success(t("admin.feedback.noteSaved"));
        },
        onError: () => toast.error(t("admin.feedback.updateFailed")),
      },
    );

  const name = item.user.display_name || item.user.email;

  return (
    <article
      id={`feedback-${item.id}`}
      className={cn(
        "scroll-mt-24 rounded-2xl border bg-card p-5 transition-shadow",
        highlighted ? "border-primary ring-2 ring-primary/30" : "border-border",
        item.status === "resolved" && "opacity-75",
      )}
    >
      <header className="flex flex-wrap items-center gap-2">
        <Badge variant={item.category === "bug" ? "destructive" : "secondary"}>{t(`feedback.category.${item.category}`)}</Badge>
        <Badge variant="outline">{t(`admin.feedback.status.${item.status}`)}</Badge>
        {!item.notified && (
          <span className="inline-flex items-center gap-1 text-xs text-muted-foreground" title={t("admin.feedback.notDeliveredHint")}>
            <BellOff className="size-3.5" strokeWidth={1.75} aria-hidden />
            {t("admin.feedback.notDelivered")}
          </span>
        )}
        <time dateTime={item.created_at} className="ml-auto text-xs text-muted-foreground" title={format.dateTime(new Date(item.created_at), { dateStyle: "medium", timeStyle: "short" })}>
          {format.relativeTime(new Date(item.created_at))}
        </time>
      </header>

      <p className="mt-3 text-[15px] leading-relaxed whitespace-pre-wrap break-words">{item.message}</p>

      <dl className="mt-4 grid gap-x-6 gap-y-1 text-xs text-muted-foreground sm:grid-cols-[auto_1fr]">
        <dt className="font-medium">{t("admin.feedback.from")}</dt>
        <dd className="min-w-0 truncate">
          {name}
          {item.user.display_name && ` (${item.user.email})`}
        </dd>
        {item.page && (
          <>
            <dt className="font-medium">{t("admin.feedback.page")}</dt>
            <dd className="min-w-0 truncate font-mono">{item.page}</dd>
          </>
        )}
        <dt className="font-medium">{t("admin.feedback.device")}</dt>
        <dd className="min-w-0 truncate" title={item.user_agent}>
          {[browserOf(item.user_agent), item.ui_locale].filter(Boolean).join(" · ")}
        </dd>
      </dl>

      {(editingNote || item.admin_note) && (
        <div className="mt-4 rounded-xl bg-muted/60 p-3">
          <p className="mb-1.5 text-xs font-medium text-muted-foreground">{t("admin.feedback.note")}</p>
          {editingNote ? (
            <div className="space-y-2">
              <Textarea value={note} onChange={(e) => setNote(e.target.value)} rows={3} maxLength={NOTE_MAX} aria-label={t("admin.feedback.note")} autoFocus />
              <div className="flex justify-end gap-2">
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => {
                    setNote(item.admin_note);
                    setEditingNote(false);
                  }}
                >
                  {t("common.cancel")}
                </Button>
                <Button size="sm" onClick={saveNote} disabled={update.isPending}>
                  {t("admin.feedback.saveNote")}
                </Button>
              </div>
            </div>
          ) : (
            <p className="text-sm whitespace-pre-wrap">{item.admin_note}</p>
          )}
        </div>
      )}

      <footer className="mt-4 flex flex-wrap gap-2">
        {item.status === "new" && (
          <Button size="sm" variant="outline" onClick={() => setStatus("seen")} disabled={update.isPending}>
            <Eye className="size-4" strokeWidth={1.75} aria-hidden />
            {t("admin.feedback.markSeen")}
          </Button>
        )}
        {item.status !== "resolved" ? (
          <Button size="sm" variant="outline" onClick={() => setStatus("resolved")} disabled={update.isPending}>
            <CheckCheck className="size-4" strokeWidth={1.75} aria-hidden />
            {t("admin.feedback.resolve")}
          </Button>
        ) : (
          <Button size="sm" variant="outline" onClick={() => setStatus("seen")} disabled={update.isPending}>
            <RotateCcw className="size-4" strokeWidth={1.75} aria-hidden />
            {t("admin.feedback.reopen")}
          </Button>
        )}
        {!editingNote && (
          <Button size="sm" variant="ghost" onClick={() => setEditingNote(true)}>
            <StickyNote className="size-4" strokeWidth={1.75} aria-hidden />
            {item.admin_note ? t("admin.feedback.editNote") : t("admin.feedback.addNote")}
          </Button>
        )}
      </footer>
    </article>
  );
}
