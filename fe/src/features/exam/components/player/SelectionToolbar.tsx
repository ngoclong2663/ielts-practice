"use client";

import { Highlighter, X } from "lucide-react";
import { useTranslations } from "next-intl";
import type { TToolbar } from "../../hooks/use-highlights";

type SelectionToolbarProps = {
  toolbar: TToolbar;
  onAdd: () => void;
  onRemove: () => void;
  onDismiss: () => void;
};

/** The small floating bar shown over a selection (add) or a highlight (remove). Floats, so it has a shadow. */
export function SelectionToolbar({ toolbar, onAdd, onRemove, onDismiss }: SelectionToolbarProps) {
  const t = useTranslations("exam.highlight");
  if (!toolbar) return null;
  return (
    <div
      role="toolbar"
      aria-label={t("toolbar")}
      // Keep the selection alive while the toolbar is pressed.
      onMouseDown={(e) => e.preventDefault()}
      style={{ left: toolbar.x, top: toolbar.y }}
      className={`fixed z-50 flex -translate-x-1/2 gap-0.5 rounded-lg bg-foreground p-1 text-background shadow-lg ${toolbar.below ? "" : "-translate-y-full"}`}
    >
      {toolbar.kind === "add" ? (
        <>
          <button type="button" onClick={onAdd} className="flex h-8 items-center gap-1.5 rounded-md px-2.5 text-[13px] font-medium hover:bg-background/15">
            <span aria-hidden className="size-3 rounded-sm bg-(--exam-highlight)" />
            {t("add")}
          </button>
          <button type="button" onClick={onDismiss} aria-label={t("cancel")} className="grid size-8 place-items-center rounded-md hover:bg-background/15">
            <X className="size-4" aria-hidden />
          </button>
        </>
      ) : (
        <button type="button" onClick={onRemove} className="flex h-8 items-center gap-1.5 rounded-md px-2.5 text-[13px] font-medium hover:bg-background/15">
          <Highlighter className="size-4" aria-hidden />
          {t("remove")}
        </button>
      )}
    </div>
  );
}
