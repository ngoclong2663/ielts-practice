"use client";

import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useDeleteDeck } from "../hooks/use-decks";
import type { TDeck } from "../types";

type DeleteDeckDialogProps = {
  deck: TDeck | null;
  onOpenChange: (open: boolean) => void;
  onDeleted?: () => void;
};

export function DeleteDeckDialog({ deck, onOpenChange, onDeleted }: DeleteDeckDialogProps) {
  const t = useTranslations();
  const del = useDeleteDeck();
  const confirm = () => {
    if (!deck) return;
    del.mutate(deck.id, {
      onSuccess: () => {
        toast.success(t("decks.deleted"));
        onOpenChange(false);
        onDeleted?.();
      },
      onError: () => toast.error(t("errors.UNKNOWN")),
    });
  };
  return (
    <AlertDialog open={!!deck} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t("decks.deleteTitle", { name: deck?.name ?? "" })}</AlertDialogTitle>
          <AlertDialogDescription>{t("decks.deleteBody", { count: deck?.card_count ?? 0 })}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={confirm} disabled={del.isPending}>
            {t("decks.delete")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
